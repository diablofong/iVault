//go:build windows && !cgo

package heic

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ConvertAll Windows 版：以單一 PowerShell 程序批次轉換所有 HEIC。
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

	// 建立批次 PowerShell 腳本：每個檔案輸出 "OK" 或 "FAIL"
	var sb strings.Builder
	sb.WriteString("[void][System.Reflection.Assembly]::LoadWithPartialName('PresentationCore')\n")
	sb.WriteString("$q = 92\n")

	for _, src := range toConvert {
		// 拒絕含控制字元的路徑（\r \n \x00），防止 PS 腳本被注入額外指令。
		// 雖然上游 organizer.safeFileName 已擋過一次，這裡再做一層深度防禦。
		if strings.ContainsAny(src, "\r\n\x00") {
			result.Failed++
			continue
		}
		dst := jpegPath(src)
		// PowerShell 單引號字串中的單引號需以 '' 跳脫
		srcEsc := strings.ReplaceAll(src, "'", "''")
		dstEsc := strings.ReplaceAll(dst, "'", "''")
		fmt.Fprintf(&sb, `
try {
    $s = [System.IO.File]::OpenRead('%s')
    $bi = New-Object System.Windows.Media.Imaging.BitmapImage
    $bi.BeginInit()
    $bi.StreamSource = $s
    $bi.CacheOption = [System.Windows.Media.Imaging.BitmapCacheOption]::OnLoad
    $bi.EndInit()
    $bi.Freeze()
    $s.Close()
    $enc = New-Object System.Windows.Media.Imaging.JpegBitmapEncoder
    $enc.QualityLevel = $q
    $enc.Frames.Add([System.Windows.Media.Imaging.BitmapFrame]::Create($bi))
    $fs = [System.IO.File]::OpenWrite('%s')
    $enc.Save($fs)
    $fs.Close()
    Write-Output 'OK'
} catch {
    Write-Output 'FAIL'
}
`, srcEsc, dstEsc)
	}

	// 寫入暫存 .ps1 檔（避免 cmd.exe 命令列長度限制）
	tmpFile, err := os.CreateTemp("", "ivault_heic_*.ps1")
	if err != nil {
		return result, fmt.Errorf("建立暫存腳本失敗: %w", err)
	}
	scriptPath := tmpFile.Name()
	defer os.Remove(scriptPath)

	if _, err := tmpFile.WriteString(sb.String()); err != nil {
		tmpFile.Close()
		return result, err
	}
	tmpFile.Close()

	// 執行單一 PowerShell 程序，串流讀取逐行輸出
	cmd := exec.CommandContext(ctx,
		"powershell", "-NoProfile", "-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-File", scriptPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return result, err
	}
	if err := cmd.Start(); err != nil {
		return result, fmt.Errorf("PowerShell 啟動失敗: %w", err)
	}

	fileIdx := 0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "OK" {
			result.Converted++
		} else {
			result.Failed++
		}
		fileIdx++
		c.emitFn("heic:progress", map[string]any{
			"total":   total,
			"done":    fileIdx,
			"percent": float64(fileIdx) / float64(total) * 100,
		})
	}
	_ = cmd.Wait()

	c.emitFn("heic:complete", map[string]any{
		"converted": result.Converted,
		"failed":    result.Failed,
	})
	return result, nil
}
