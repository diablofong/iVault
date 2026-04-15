//go:build windows && !cgo

package heic

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

// ConvertAll Windows 版：以 heicWorkerCount 個 PowerShell 程序並行轉換所有 HEIC。
// 需要 Apple Devices App 已安裝（提供 Windows HEIC WIC codec）。
func (c *Converter) ConvertAll(ctx context.Context, backupPath string) (*ConvertResult, error) {
	heicFiles := scanHeicFiles(backupPath)
	toConvert := filterAlreadyConverted(heicFiles)

	result := &ConvertResult{}
	total := len(toConvert)

	if total == 0 {
		c.emitFn("heic:complete", map[string]any{"converted": 0, "failed": 0})
		return result, nil
	}

	chunks := splitChunks(toConvert, heicWorkerCount)

	var wg sync.WaitGroup
	var converted, failed, done atomic.Int64

	for _, chunk := range chunks {
		wg.Add(1)
		chunk := chunk
		go func() {
			defer wg.Done()
			c.runPSWorker(ctx, chunk, total, &converted, &failed, &done)
		}()
	}
	wg.Wait()

	result.Converted = int(converted.Load())
	result.Failed = int(failed.Load())
	c.emitFn("heic:complete", map[string]any{
		"converted": result.Converted,
		"failed":    result.Failed,
	})
	return result, nil
}

// runPSWorker 以單一 PowerShell 程序轉換一個 chunk 的 HEIC 檔案。
func (c *Converter) runPSWorker(ctx context.Context, files []string, total int, converted, failed, done *atomic.Int64) {
	var validFiles []string
	var sb strings.Builder
	// Add-Type 比 LoadWithPartialName 更可靠（後者在部分 Windows 11 環境失效）
	sb.WriteString("Add-Type -AssemblyName PresentationCore\n")
	sb.WriteString("$q = 92\n")

	for _, src := range files {
		// 拒絕含控制字元的路徑，防止 PS 腳本注入
		if strings.ContainsAny(src, "\r\n\x00") {
			failed.Add(1)
			continue
		}
		dst := jpegPath(src)
		srcEsc := strings.ReplaceAll(src, "'", "''")
		dstEsc := strings.ReplaceAll(dst, "'", "''")
		validFiles = append(validFiles, src)
		fmt.Fprintf(&sb, `
$s = $null; $fs = $null
try {
    $s = [System.IO.File]::OpenRead('%s')
    $bi = New-Object System.Windows.Media.Imaging.BitmapImage
    $bi.BeginInit()
    $bi.StreamSource = $s
    $bi.CacheOption = [System.Windows.Media.Imaging.BitmapCacheOption]::OnLoad
    $bi.EndInit()
    $bi.Freeze()
    $s.Close(); $s = $null
    $enc = New-Object System.Windows.Media.Imaging.JpegBitmapEncoder
    $enc.QualityLevel = $q
    $enc.Frames.Add([System.Windows.Media.Imaging.BitmapFrame]::Create($bi))
    $fs = [System.IO.File]::OpenWrite('%s')
    $enc.Save($fs)
    $fs.Close(); $fs = $null
    Write-Output 'OK'
} catch {
    if ($s)  { try { $s.Close()  } catch {} }
    if ($fs) { try { $fs.Close() } catch {} }
    Remove-Item '%s' -ErrorAction SilentlyContinue
    Write-Output 'FAIL'
}
`, srcEsc, dstEsc, dstEsc)
	}

	if len(validFiles) == 0 {
		return
	}

	tmpFile, err := os.CreateTemp("", "ivault_heic_*.ps1")
	if err != nil {
		failed.Add(int64(len(validFiles)))
		return
	}
	scriptPath := tmpFile.Name()
	defer os.Remove(scriptPath)

	if _, err := tmpFile.WriteString(sb.String()); err != nil {
		tmpFile.Close()
		failed.Add(int64(len(validFiles)))
		return
	}
	tmpFile.Close()

	cmd := exec.CommandContext(ctx,
		"powershell", "-NoProfile", "-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-File", scriptPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		failed.Add(int64(len(validFiles)))
		return
	}
	if err := cmd.Start(); err != nil {
		failed.Add(int64(len(validFiles)))
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "OK" {
			converted.Add(1)
		} else {
			failed.Add(1)
		}
		n := done.Add(1)
		c.emitFn("heic:progress", map[string]any{
			"total":   total,
			"done":    int(n),
			"percent": float64(n) / float64(total) * 100,
		})
	}
	_ = cmd.Wait()
}
