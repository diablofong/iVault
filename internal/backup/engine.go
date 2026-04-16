package backup

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ivault/internal/device"
	"ivault/internal/platform"

	"github.com/danielpaulus/go-ios/ios/afc"
)

const copyBufferSize = 256 * 1024 // 256KB
const afcCallTimeout = 30 * time.Second
const afcWorkerCount = 1 // usbmuxd 序列化限制：worker=2 僅提升 8% 但 p95 延遲爆增 3.4×，維持 1

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
		organizer: NewOrganizer(config.BackupPath, config.DeviceName, config.DeviceUDID, config.OrganizeByDate),
		speed:     NewSpeedTracker(),
		emitFn:    emitFn,
	}
}

// isPhotoExt 判斷副檔名是否為靜態照片
func isPhotoExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".heic", ".heif", ".png", ".tiff", ".tif":
		return true
	}
	return false
}

// isVideoExt 判斷副檔名是否為影片
func isVideoExt(ext string) bool {
	switch ext {
	case ".mov", ".mp4", ".m4v":
		return true
	}
	return false
}

// afcList 執行 AFC List，帶 30 秒 hard timeout，避免 AFC 卡死讓整個 app hang。
func afcList(ctx context.Context, client *afc.Client, dirPath string) ([]string, error) {
	type result struct {
		files []string
		err   error
	}
	tctx, cancel := context.WithTimeout(ctx, afcCallTimeout)
	defer cancel()
	ch := make(chan result, 1)
	go func() {
		f, e := client.List(dirPath)
		// 若 caller 已超時離開，select 避免在無人接收的 channel 上永久 block（1 容量其實能寫入，
		// 但改成雙路更清楚地表達「caller 已放棄」的語意，也方便未來擴充為 unbuffered channel）。
		select {
		case ch <- result{f, e}:
		case <-tctx.Done():
		}
	}()
	select {
	case r := <-ch:
		return r.files, r.err
	case <-tctx.Done():
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrAFCTimeout
	}
}

// afcStat 執行 AFC Stat，帶 30 秒 hard timeout。
func afcStat(ctx context.Context, client *afc.Client, remotePath string) (afc.FileInfo, error) {
	type result struct {
		info afc.FileInfo
		err  error
	}
	tctx, cancel := context.WithTimeout(ctx, afcCallTimeout)
	defer cancel()
	ch := make(chan result, 1)
	go func() {
		info, e := client.Stat(remotePath)
		select {
		case ch <- result{info, e}:
		case <-tctx.Done():
		}
	}()
	select {
	case r := <-ch:
		return r.info, r.err
	case <-tctx.Done():
		if ctx.Err() != nil {
			return afc.FileInfo{}, ctx.Err()
		}
		return afc.FileInfo{}, ErrAFCTimeout
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
		if err == ErrAFCTimeout {
			return nil, ErrAFCTimeout
		}
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	totalFiles := len(newFiles)

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
		TotalFiles:   totalFiles + skippedCount,
		SkippedFiles: skippedCount,
		TotalBytes:   totalNewBytes,
	})

	// === Phase 2: Copying（worker pool，afcWorkerCount 個並行 AFC session）===
	type copyJob struct {
		file  device.PhotoFile
		index int
	}
	jobs := make(chan copyJob, totalFiles)
	for i, f := range newFiles {
		jobs <- copyJob{f, i}
	}
	close(jobs)

	var (
		doneBytes   atomic.Int64
		doneCount   atomic.Int32
		resultMu    sync.Mutex
		fatalErr    error
		fatalOnce   sync.Once
		wg          sync.WaitGroup
	)
	innerCtx, innerCancel := context.WithCancel(ctx)
	defer innerCancel()

	saveInterrupted := func() *BackupResult {
		_ = e.manifest.SaveInterrupted()
		result.Interrupted = true
		done := int(doneCount.Load())
		result.InterruptedDone = done
		result.InterruptedTotal = totalFiles + skippedCount
		result.NewFiles = done
		result.SkippedFiles = skippedCount
		result.FailedFiles = len(result.FailedList)
		result.TotalBytes = doneBytes.Load()
		result.Duration = formatDuration(time.Since(e.startTime))
		return result
	}

	workerFn := func(workerAFC *afc.Client) {
		buf := make([]byte, copyBufferSize)
		for job := range jobs {
			select {
			case <-innerCtx.Done():
				return
			default:
			}

			file := job.file
			localPath := e.organizer.ResolveLocalPath(file)
			localRelPath := e.organizer.RelativeLocalPath(localPath)

			tFile := time.Now()

			n, copyErr := CopyFileBuffered(innerCtx, workerAFC, file.RemotePath, localPath, buf)

			if copyErr != nil {
				if isDeviceDisconnected(copyErr) {
					fatalOnce.Do(func() {
						fatalErr = ErrDeviceDisconnected
						innerCancel()
					})
					return
				}
				if copyErr == ErrAFCTimeout {
					fatalOnce.Do(func() {
						fatalErr = ErrAFCTimeout
						innerCancel()
					})
					return
				}
				resultMu.Lock()
				result.FailedList = append(result.FailedList, FailedFile{
					FileName: file.FileName,
					Reason:   humanizeError(copyErr),
				})
				resultMu.Unlock()
				continue
			}

			// 驗證檔案大小
			if fi, statErr := os.Stat(localPath); statErr == nil && fi.Size() != file.Size {
				_ = os.Remove(localPath)
				resultMu.Lock()
				result.FailedList = append(result.FailedList, FailedFile{
					FileName: file.FileName,
					Reason:   "檔案大小不符，可能傳輸中斷",
				})
				resultMu.Unlock()
				continue
			}

			// 讀取拍攝日期
			shootDate, ok := ReadShootDate(localPath)
			if !ok {
				shootDate = time.Now()
			}

			localPath = e.organizer.ResolveByDate(file, localPath, shootDate)
			localRelPath = e.organizer.RelativeLocalPath(localPath)

			e.manifest.MarkDone(file, localRelPath)

			e.speed.Add(n, time.Since(tFile))
			totalDone := doneBytes.Add(n)

			ext := strings.ToLower(path.Ext(file.FileName))
			resultMu.Lock()
			if isPhotoExt(ext) {
				result.PhotosCount++
				if ext == ".heic" || ext == ".heif" {
					result.HasHeic = true
				}
			} else if isVideoExt(ext) {
				result.VideosCount++
			}
			resultMu.Unlock()

			currentMonth := ""
			if ok {
				currentMonth = shootDate.Format("2006-01")
			}
			n32 := doneCount.Add(1)
			remaining := totalNewBytes - totalDone
			e.emit("backup:progress", BackupProgress{
				Phase:        "copying",
				CurrentFile:  file.FileName,
				CurrentMonth: currentMonth,
				TotalFiles:   totalFiles,
				DoneFiles:    int(n32),
				SkippedFiles: skippedCount,
				TotalBytes:   totalNewBytes,
				DoneBytes:    totalDone,
				SpeedBps:     int64(e.speed.Average()),
				ETA:          formatDuration(e.speed.ETA(remaining)),
				Percent:      float64(n32) / float64(totalFiles) * 100,
			})

			if int(n32)%50 == 0 {
				_ = e.manifest.Save()
			}
		}
	}

	// 啟動 worker pool：每個 worker 各自獨立的 AFC session
	var launchedWorkers int
	for w := 0; w < afcWorkerCount; w++ {
		workerAFC, err := device.ConnectAFC(e.config.DeviceUDID)
		if err != nil {
			continue
		}
		launchedWorkers++
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer workerAFC.Close()
			workerFn(workerAFC)
		}()
	}
	if launchedWorkers == 0 {
		return nil, ErrAFCConnectFailed
	}
	wg.Wait()
	innerCancel()

	// 判斷是否中斷
	if fatalErr == ErrDeviceDisconnected {
		return saveInterrupted(), ErrDeviceDisconnected
	}
	if fatalErr != nil {
		return saveInterrupted(), fatalErr
	}
	if ctx.Err() != nil {
		return saveInterrupted(), nil
	}

	// === Phase 3: Finalizing ===
	_ = e.manifest.SaveCompleted()

	result.Success = true
	result.NewFiles = totalFiles - len(result.FailedList)
	result.SkippedFiles = skippedCount
	result.FailedFiles = len(result.FailedList)
	result.TotalBytes = doneBytes.Load()
	result.Duration = formatDuration(time.Since(e.startTime))

	return result, nil
}

// scanDCIM 兩階段掃描（含 AFC call timeout）：
// Phase 1a — 只用 List() 拿檔名，與 manifest 比對，已備份的直接跳過
// Phase 1b — 只對新檔案 Stat() 拿 size
func (e *Engine) scanDCIM(ctx context.Context, afcClient *afc.Client) (newFiles []device.PhotoFile, skippedCount int, err error) {
	dirs, err := afcList(ctx, afcClient, "/DCIM")
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
		files, listErr := afcList(ctx, afcClient, dirPath)
		if listErr != nil {
			if listErr == ErrAFCTimeout {
				return nil, 0, ErrAFCTimeout
			}
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
			fileInfo, statErr := afcStat(ctx, afcClient, remotePath)
			if statErr != nil {
				if statErr == ErrAFCTimeout {
					return nil, 0, ErrAFCTimeout
				}
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
