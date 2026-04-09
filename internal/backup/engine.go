package backup

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"ivault/internal/device"
	"ivault/internal/platform"

	"github.com/danielpaulus/go-ios/ios/afc"
)

const copyBufferSize = 256 * 1024 // 256KB

// Engine 核心備份引擎
type Engine struct {
	config    BackupConfig
	manifest  *Manifest
	organizer *Organizer
	speed     *SpeedTracker
	emitFn    func(event string, data any)
	startTime time.Time
}

// NewEngine 建立備份引擎
func NewEngine(config BackupConfig, emitFn func(string, any)) *Engine {
	return &Engine{
		config:    config,
		organizer: NewOrganizer(config.BackupPath, config.DeviceName, config.OrganizeByDate),
		speed:     NewSpeedTracker(),
		emitFn:    emitFn,
	}
}

// Run 執行完整備份流程，支援 context 取消
func (e *Engine) Run(ctx context.Context) (*BackupResult, error) {
	e.startTime = time.Now()
	result := &BackupResult{BackupPath: e.config.BackupPath}

	// === Phase 1: Scanning ===
	e.emit("backup:progress", BackupProgress{Phase: "scanning"})

	afcClient, err := device.ConnectAFC(e.config.DeviceUDID)
	if err != nil {
		return nil, ErrAFCConnectFailed
	}
	defer afcClient.Close()

	e.manifest = LoadOrCreateManifest(e.config.BackupPath, e.config.DeviceUDID, e.config.DeviceName)

	newFiles, skippedCount, err := e.scanDCIM(ctx, afcClient)
	if err != nil {
		if isDeviceDisconnected(err) {
			return nil, ErrDeviceDisconnected
		}
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// 磁碟空間檢查
	var totalNewBytes int64
	for _, f := range newFiles {
		totalNewBytes += f.Size
	}
	diskInfo := platform.GetDiskInfo(e.config.BackupPath)
	if diskInfo.FreeSpace > 0 && totalNewBytes > diskInfo.FreeSpace {
		return nil, ErrDiskFull
	}

	e.emit("backup:progress", BackupProgress{
		Phase:        "scanning",
		TotalFiles:   len(newFiles) + skippedCount,
		SkippedFiles: skippedCount,
		TotalBytes:   totalNewBytes,
	})

	// === Phase 2: Copying ===
	buf := make([]byte, copyBufferSize)
	var doneBytes int64

	for i, file := range newFiles {
		// 檢查取消
		select {
		case <-ctx.Done():
			_ = e.manifest.Save()
			result.Interrupted = true
			result.NewFiles = i
			result.SkippedFiles = skippedCount
			result.FailedFiles = len(result.FailedList)
			result.TotalBytes = doneBytes
			result.Duration = formatDuration(time.Since(e.startTime))
			return result, nil
		default:
		}

		localPath := e.organizer.ResolveLocalPath(file)
		localRelPath := e.organizer.RelativeLocalPath(localPath)

		start := time.Now()
		n, copyErr := CopyFileBuffered(afcClient, file.RemotePath, localPath, buf)
		elapsed := time.Since(start)

		if copyErr != nil {
			if isDeviceDisconnected(copyErr) {
				_ = e.manifest.Save()
				result.Interrupted = true
				result.NewFiles = i
				result.SkippedFiles = skippedCount
				result.FailedFiles = len(result.FailedList)
				result.TotalBytes = doneBytes
				result.Duration = formatDuration(time.Since(e.startTime))
				return result, ErrDeviceDisconnected
			}
			result.FailedList = append(result.FailedList, FailedFile{
				FileName: file.FileName,
				Reason:   humanizeError(copyErr),
			})
			continue
		}

		// 驗證檔案大小
		if fi, statErr := os.Stat(localPath); statErr == nil && fi.Size() != file.Size {
			_ = os.Remove(localPath)
			result.FailedList = append(result.FailedList, FailedFile{
				FileName: file.FileName,
				Reason:   "檔案大小不符，可能傳輸中斷",
			})
			continue
		}

		// 讀 EXIF 取得拍攝日期，移動到正確的 YYYY-MM 目錄
		shootDate, ok := ReadShootDate(localPath)
		if !ok {
			shootDate = time.Now() // HEIC / 無 EXIF → 用備份當天日期
		}
		localPath = e.organizer.ResolveByDate(file, localPath, shootDate)
		localRelPath = e.organizer.RelativeLocalPath(localPath)

		e.manifest.MarkDone(file, localRelPath)
		e.speed.Add(n, elapsed)
		doneBytes += n

		// HEIC 偵測
		ext := strings.ToLower(path.Ext(file.FileName))
		if ext == ".heic" || ext == ".heif" {
			result.HasHeic = true
		}

		// 進度推送
		remaining := totalNewBytes - doneBytes
		e.emit("backup:progress", BackupProgress{
			Phase:        "copying",
			CurrentFile:  file.FileName,
			TotalFiles:   len(newFiles),
			DoneFiles:    i + 1,
			SkippedFiles: skippedCount,
			TotalBytes:   totalNewBytes,
			DoneBytes:    doneBytes,
			SpeedBps:     int64(e.speed.Average()),
			ETA:          formatDuration(e.speed.ETA(remaining)),
			Percent:      float64(i+1) / float64(len(newFiles)) * 100,
		})

		// 每 50 個檔案存一次 manifest
		if (i+1)%50 == 0 {
			_ = e.manifest.Save()
		}
	}

	// === Phase 3: Finalizing ===
	_ = e.manifest.Save()

	result.Success = true
	result.NewFiles = len(newFiles) - len(result.FailedList)
	result.SkippedFiles = skippedCount
	result.FailedFiles = len(result.FailedList)
	result.TotalBytes = doneBytes
	result.Duration = formatDuration(time.Since(e.startTime))

	return result, nil
}

// scanDCIM 兩階段掃描：
// Phase 1a — 只用 List() 拿檔名，與 manifest 比對，已備份的直接跳過
// Phase 1b — 只對新檔案 Stat() 拿 size
func (e *Engine) scanDCIM(ctx context.Context, afcClient *afc.Client) (newFiles []device.PhotoFile, skippedCount int, err error) {
	dirs, err := afcClient.List("/DCIM")
	if err != nil {
		return nil, 0, fmt.Errorf("list /DCIM: %w", err)
	}

	for _, dir := range dirs {
		// 只處理 *APPLE 目錄
		if !strings.HasSuffix(dir, "APPLE") {
			continue
		}

		select {
		case <-ctx.Done():
			return newFiles, skippedCount, ctx.Err()
		default:
		}

		dirPath := "/DCIM/" + dir
		files, listErr := afcClient.List(dirPath)
		if listErr != nil {
			continue
		}

		for _, fileName := range files {
			ext := strings.ToLower(path.Ext(fileName))
			if !device.IsSupportedExtension(ext) {
				continue
			}

			relativePath := dir + "/" + fileName

			// Phase 1a：檢查 manifest，已備份的直接跳過（不 Stat）
			if _, exists := e.manifest.Files[relativePath]; exists {
				skippedCount++
				continue
			}

			// Phase 1b：新檔案才 Stat 拿 size
			remotePath := dirPath + "/" + fileName
			fileInfo, statErr := afcClient.Stat(remotePath)
			if statErr != nil {
				continue
			}

			newFiles = append(newFiles, device.PhotoFile{
				RemotePath: remotePath,
				FileName:   fileName,
				Size:       fileInfo.Size,
				ModTime:    0, // go-ios afc.FileInfo 無 mtime
			})
		}
	}

	return newFiles, skippedCount, nil
}

// emit 推送事件給前端
func (e *Engine) emit(event string, data any) {
	if e.emitFn != nil {
		e.emitFn(event, data)
	}
}

// formatDuration 格式化時間為人類可讀格式
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
