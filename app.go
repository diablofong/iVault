package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ivault/internal/backup"
	"ivault/internal/config"
	"ivault/internal/device"
	"ivault/internal/heic"
	"ivault/internal/platform"
	"ivault/internal/updater"

	"github.com/danielpaulus/go-ios/ios"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	configMgr    *config.Manager
	platformInfo *platform.Info

	// 裝置狀態
	mu              sync.RWMutex
	connectedDevice *device.DeviceInfo

	// 備份取消函式
	backupCancel context.CancelFunc

	// 信任輪詢取消
	trustCancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.configMgr = config.NewManager()
	a.platformInfo = platform.Detect()
	go a.watchDevices()
	go a.checkForUpdate()
}

// checkForUpdate runs once at startup: fetches the latest GitHub release and
// emits update:available if a newer version exists. Errors are silently dropped.
func (a *App) checkForUpdate() {
	// Small delay so the UI can fully mount before we emit.
	time.Sleep(3 * time.Second)

	info, err := updater.Check(Version)
	if err != nil || !info.Available {
		return
	}
	wailsRuntime.EventsEmit(a.ctx, "update:available", map[string]any{
		"version": info.Version,
		"url":     info.URL,
	})
}

// shutdown is called when the app is about to quit
func (a *App) shutdown(ctx context.Context) {
	if a.backupCancel != nil {
		a.backupCancel()
	}
	if a.trustCancel != nil {
		a.trustCancel()
	}
	platform.AllowSleep()
}

// ============================================================
// 裝置相關 API
// ============================================================

// GetPlatformInfo 取得平台資訊
func (a *App) GetPlatformInfo() platform.Info {
	if a.platformInfo == nil {
		return platform.Info{}
	}
	return *a.platformInfo
}

// GetConnectedDevice 取得當前連接的裝置（nil 代表無裝置）
func (a *App) GetConnectedDevice() *device.DeviceInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connectedDevice
}

// GetDeviceDetail 取得裝置詳細資訊（含照片數、空間）
func (a *App) GetDeviceDetail(udid string) (*device.DeviceDetail, error) {
	return device.GetDeviceDetail(udid)
}

// CheckTrustStatus 檢查裝置配對狀態
// 嘗試 GetValues：成功 = 已信任，失敗 = 未信任
func (a *App) CheckTrustStatus(udid string) (bool, error) {
	deviceList, err := ios.ListDevices()
	if err != nil {
		return false, err
	}
	for _, d := range deviceList.DeviceList {
		if d.Properties.SerialNumber == udid {
			_, err := ios.GetValues(d)
			return err == nil, nil
		}
	}
	return false, fmt.Errorf("device not found: %s", udid)
}

// TriggerTrustCheck 由前端「我已信任，繼續」按鈕觸發，立即查一次信任狀態，
// 不等 polling tick。若已信任就同步 emit device:trust-changed，
// 讓前端直接切到 READY；否則回傳 false 讓使用者得到明確回饋。
func (a *App) TriggerTrustCheck(udid string) bool {
	trusted, err := a.CheckTrustStatus(udid)
	if err != nil || !trusted {
		return false
	}
	a.mu.Lock()
	if a.connectedDevice != nil {
		a.connectedDevice.Trusted = true
	}
	a.mu.Unlock()
	wailsRuntime.EventsEmit(a.ctx, "device:trust-changed", map[string]any{
		"udid": udid, "trusted": true,
	})
	return true
}

// CheckAppleDevicesInstalled Windows 專用：每次都做即時偵測（不用快取）
func (a *App) CheckAppleDevicesInstalled() bool {
	return platform.RecheckAppleDevices()
}

// SetAutostart Windows：設定登錄開機自動啟動
func (a *App) SetAutostart(enabled bool) error {
	return platform.SetAutostart(enabled)
}

// GetAutostart Windows：查詢是否已設定開機自動啟動
func (a *App) GetAutostart() bool {
	return platform.GetAutostart()
}

// LaunchAppleDevices Windows：直接啟動 Apple Devices App（不透過 Store）
func (a *App) LaunchAppleDevices() {
	go platform.LaunchAppleDevices()
}

// GetBackupEstimate 取得備份估算（總大小、最大單檔、新檔數）
func (a *App) GetBackupEstimate(udid string, backupPath string) (backup.BackupEstimate, error) {
	photos, err := device.ScanDCIM(udid)
	if err != nil {
		return backup.BackupEstimate{}, err
	}
	m := backup.LoadOrCreateManifest(backupPath, udid, "")
	var est backup.BackupEstimate
	for _, p := range photos {
		if !m.IsBackedUp(p) {
			est.TotalBytes += p.Size
			est.FileCount++
			if p.Size > est.MaxBytes {
				est.MaxBytes = p.Size
			}
		}
	}
	return est, nil
}

// InstallAppleDevices Windows 專用：開啟 Microsoft Store Apple Devices 頁面
// 安裝由使用者在 Store 完成後手動點「重新偵測」，不依賴自動偵測推進 UX。
func (a *App) InstallAppleDevices() {
	_ = platform.OpenURL("ms-windows-store://pdp/?productId=9NP83LWLPZ9K")
}

// InstallHeicCodec Windows 專用：開啟 Microsoft Store HEIC 影像擴充頁面
func (a *App) InstallHeicCodec() {
	_ = platform.OpenURL("ms-windows-store://pdp?productId=9PMMSR1CGPWG")
}

// ============================================================
// 備份相關 API
// ============================================================

// SelectBackupFolder 開啟原生資料夾選擇對話框
func (a *App) SelectBackupFolder() string {
	path, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "選擇備份目標資料夾",
	})
	if err != nil {
		return ""
	}
	return path
}

// GetDefaultBackupPath 取得智慧預設備份路徑
func (a *App) GetDefaultBackupPath() string {
	// 優先使用上次備份路徑
	if last := a.configMgr.Get().LastBackupPath; last != "" {
		return last
	}
	return platform.GetDefaultBackupPath()
}

// CheckBackupPath 進入 READY 頁前呼叫：確認備份路徑是否仍有效。
// path 為前端當前 selectedPath；空字串時 fallback 到 config 的 LastBackupPath。
// 若路徑存在但磁碟上找不到（外接硬碟未插），emit backup:path-missing。
func (a *App) CheckBackupPath(path string) {
	checkPath := path
	if checkPath == "" {
		checkPath = a.configMgr.Get().LastBackupPath
	}
	if checkPath == "" {
		return
	}
	if _, err := os.Stat(checkPath); err != nil {
		wailsRuntime.EventsEmit(a.ctx, "backup:path-missing", map[string]any{
			"path": checkPath,
		})
	}
}

// ManifestExists 檢查備份路徑下是否存有指定裝置的 manifest（斷點記錄）
// 前端用來判斷「全部已是最新」是否還有效：若備份資料夾被刪，manifest 消失，
// 就不該再顯示「全部已是最新」，避免誤導使用者。
func (a *App) ManifestExists(backupPath, udid string) bool {
	return backup.ManifestExists(backupPath, udid)
}

// GetDiskInfo 取得指定路徑的磁碟空間資訊
func (a *App) GetDiskInfo(path string) platform.DiskInfo {
	return platform.GetDiskInfo(path)
}

// EstimateBackupSize 估算需要備份的大小（排除已備份檔案）
func (a *App) EstimateBackupSize(udid string, backupPath string) (int64, error) {
	photos, err := device.ScanDCIM(udid)
	if err != nil {
		return 0, err
	}
	manifest := backup.LoadOrCreateManifest(backupPath, udid, "")
	var total int64
	for _, p := range photos {
		if !manifest.IsBackedUp(p) {
			total += p.Size
		}
	}
	return total, nil
}

// StartBackup 開始備份（非同步，進度透過 Event 推送）
func (a *App) StartBackup(cfg backup.BackupConfig) error {
	if a.backupCancel != nil {
		a.backupCancel()
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.backupCancel = cancel

	// 儲存備份路徑到 config
	current := a.configMgr.Get()
	current.LastBackupPath = cfg.BackupPath
	_ = a.configMgr.Save(current)

	platform.PreventSleep() // AD: 備份中阻止系統睡眠

	emitFn := func(event string, data any) {
		wailsRuntime.EventsEmit(a.ctx, event, data)
	}
	engine := backup.NewEngine(cfg, emitFn)

	go func() {
		defer platform.AllowSleep() // AD: 備份結束恢復睡眠
		result, err := engine.Run(ctx)
		if err != nil && err != context.Canceled {
			// 記錄中斷狀態
			cur := a.configMgr.Get()
			cur.LastInterrupted = true
			cur.InterruptedDeviceUDID = cfg.DeviceUDID
			_ = a.configMgr.Save(cur)

			if be, ok := err.(*backup.BackupError); ok {
				wailsRuntime.EventsEmit(a.ctx, "backup:error", map[string]any{
					"code": be.Code, "message": be.Message, "recoverable": be.Recoverable,
				})
			} else {
				wailsRuntime.EventsEmit(a.ctx, "backup:error", map[string]any{
					"code": "UNKNOWN_ERROR", "message": "發生未預期的錯誤，請重試", "recoverable": true,
				})
			}
			return
		}
		if result != nil {
			if result.Interrupted {
				// 中斷：更新 config 中斷旗標與進度
				cur := a.configMgr.Get()
				cur.LastInterrupted = true
				cur.InterruptedDone = result.InterruptedDone
				cur.InterruptedTotal = result.InterruptedTotal
				cur.InterruptedDeviceUDID = cfg.DeviceUDID
				_ = a.configMgr.Save(cur)
				wailsRuntime.EventsEmit(a.ctx, "backup:interrupted", result)
				return
			}

			// 成功完成：清除中斷旗標
			cur := a.configMgr.Get()
			cur.LastInterrupted = false
			cur.InterruptedDone = 0
			cur.InterruptedTotal = 0
			cur.InterruptedDeviceUDID = ""
			if !cur.FirstBackupDone {
				cur.FirstBackupDone = true
			}
			alreadyDone := false
			for _, uid := range cur.FirstBackupDoneDevices {
				if uid == cfg.DeviceUDID {
					alreadyDone = true
					break
				}
			}
			if !alreadyDone && cfg.DeviceUDID != "" {
				cur.FirstBackupDoneDevices = append(cur.FirstBackupDoneDevices, cfg.DeviceUDID)
			}
			_ = a.configMgr.Save(cur)

			_ = a.configMgr.AddRecord(config.BackupRecord{
				Date:        time.Now().Format(time.RFC3339),
				DeviceName:  cfg.DeviceName,
				DeviceUDID:  cfg.DeviceUDID,
				NewFiles:    result.NewFiles,
				PhotosCount: result.PhotosCount,
				VideosCount: result.VideosCount,
				Skipped:     result.SkippedFiles,
				Failed:      result.FailedFiles,
				TotalBytes:  result.TotalBytes,
				Duration:    result.Duration,
			})
			// P: Toast 通知（非同步，不阻塞 event emit）
			go platform.ShowToast(
				"備份完成",
				fmt.Sprintf("已備份 %d 張照片、%d 段影片", result.PhotosCount, result.VideosCount),
			)
			wailsRuntime.EventsEmit(a.ctx, "backup:complete", result)

			// 若勾選了「轉存 JPEG」且備份結果含 HEIC，啟動轉檔（綁定 cancel context）
			if cfg.ConvertHeic && result.HasHeic {
				converter := heic.NewConverter(92, emitFn)
				go converter.ConvertAll(ctx, cfg.BackupPath)
			}
		}
	}()
	return nil
}

// CancelBackup 取消正在進行的備份
func (a *App) CancelBackup() error {
	if a.backupCancel != nil {
		a.backupCancel()
		a.backupCancel = nil
	}
	return nil
}

// ============================================================
// 設定與工具 API
// ============================================================

// LoadConfig 載入設定
func (a *App) LoadConfig() config.AppConfig {
	return a.configMgr.Get()
}

// SaveConfig 儲存設定
func (a *App) SaveConfig(cfg config.AppConfig) error {
	return a.configMgr.Save(cfg)
}

// GetBackupHistory 取得備份歷史紀錄
func (a *App) GetBackupHistory() []config.BackupRecord {
	return a.configMgr.Get().History
}

// OpenFolder 用系統檔案管理員開啟指定資料夾。
// 安全：必須為絕對本機路徑、不得為 UNC（\\host\share），且必須存在。
// 阻擋的主要威脅是前端被注入後傳入 UNC 路徑觸發 SMB 連線而洩漏 NTLM hash。
func (a *App) OpenFolder(path string) error {
	if path == "" {
		return fmt.Errorf("invalid path")
	}
	if strings.HasPrefix(path, `\\`) || strings.HasPrefix(path, "//") {
		return fmt.Errorf("UNC path not allowed")
	}
	// Defense-in-depth：Windows explorer.exe 把逗號解析為 flag 分隔字元，
	// 即便路徑實際存在也可能被用來附加隱藏 flag。阻擋含逗號的路徑。
	if strings.ContainsRune(path, ',') {
		return fmt.Errorf("path contains invalid character")
	}
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute")
	}
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return fmt.Errorf("path not accessible")
	}
	return platform.OpenFolder(path)
}

// OpenURL 用系統瀏覽器開啟 URL。
// 安全：只允許 http(s) scheme，避免前端傳入 file://、ms-settings:、javascript: 等
// 會被 Windows FileProtocolHandler 解析成本機操作的 URI。
func (a *App) OpenURL(url string) error {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return fmt.Errorf("only http/https URLs are allowed")
	}
	return platform.OpenURL(url)
}

// ============================================================
// 後台 Goroutine
// ============================================================

// watchDevices 使用 ios.Listen() 監聽裝置熱插拔事件。
// Windows 上若 Apple Devices 未安裝，ios.Listen() 會靜默失敗；
// 此函式主動偵測並透過事件通知前端，同時持續輪詢直到驅動裝好。
func (a *App) watchDevices() {
	driverWasMissing := false
	amdsFailNotified := false // 避免重複發同一個錯誤事件

	for {
		// Windows 專用：Apple Devices 前置條件檢查
		if a.platformInfo != nil && a.platformInfo.OS == "windows" {
			// [1] Driver 檢查（原有）
			if !platform.RecheckAppleDevices() {
				if !driverWasMissing {
					// 等前端 ready 後再發事件（避免前端尚未掛載監聽器）
					time.Sleep(600 * time.Millisecond)
					// 嘗試 WMI 偵測 iPhone（驅動未裝仍可取得裝置名稱）
					deviceName := platform.WMIDetectIPhone()
					wailsRuntime.EventsEmit(a.ctx, "driver:required", map[string]any{
						"deviceName": deviceName,
					})
					driverWasMissing = true
				}
				time.Sleep(3 * time.Second)
				continue
			}
			// 驅動剛裝好（missing → available）
			if driverWasMissing {
				a.mu.Lock()
				if a.platformInfo != nil {
					a.platformInfo.AppleDevicesInstalled = true
				}
				a.mu.Unlock()
				wailsRuntime.EventsEmit(a.ctx, "driver:installed", nil)
				driverWasMissing = false
			}

			// [2] AMDS 確認：確保 AppleMobileDeviceProcess.exe 已在 listening port 27015
			// MS Store 版 Apple Devices 不像 iTunes 開機自動常駐，需手動喚起
			if !platform.IsAMDSReady() {
				// 通知前端：即將喚起 Apple Devices UI，讓使用者有心理準備
				wailsRuntime.EventsEmit(a.ctx, "amds:starting", nil)
			}
			if err := platform.EnsureAMDSRunning(); err != nil {
				if !amdsFailNotified {
					errMsg := err.Error()
					eventName := "amds:start_failed"
					if errMsg == "AMDS_TIMEOUT" {
						eventName = "amds:timeout"
					}
					wailsRuntime.EventsEmit(a.ctx, eventName, map[string]any{
						"error": errMsg,
					})
					amdsFailNotified = true
				}
				time.Sleep(5 * time.Second)
				continue
			}
			amdsFailNotified = false // 成功後 reset，未來若再失敗可以重新通知
		}

		// [3] 裝置監聽（原有）
		a.runDeviceListener()
		// listener 斷掉（usbmuxd 重啟、裝置拔除等）→ 等 3 秒後重試
		time.Sleep(3 * time.Second)
	}
}

func (a *App) runDeviceListener() {
	receiver, closeFunc, err := ios.Listen()
	if err != nil {
		return
	}
	defer closeFunc()

	// 建立 DeviceID → UDID 映射（Detached 訊息只有 DeviceID）
	idToUDID := make(map[int]string)

	for {
		msg, err := receiver()
		if err != nil {
			return
		}

		if msg.DeviceAttached() {
			udid := msg.Properties.SerialNumber
			if udid == "" {
				continue
			}
			idToUDID[msg.DeviceID] = udid

			// 取得裝置名稱
			info := device.DeviceInfo{UDID: udid}
			if values, err := ios.GetValues(msg.DeviceEntry()); err == nil {
				info.Name = values.Value.DeviceName
				info.Model = values.Value.ProductType
				info.IOSVersion = values.Value.ProductVersion
				info.Trusted = true
			}

			a.mu.Lock()
			a.connectedDevice = &info
			a.mu.Unlock()

			wailsRuntime.EventsEmit(a.ctx, "device:connected", info)

			// 若未信任，啟動信任輪詢
			if !info.Trusted {
				a.startTrustPolling(udid)
			}

		} else if msg.DeviceDetached() {
			udid := idToUDID[msg.DeviceID]

			a.mu.Lock()
			a.connectedDevice = nil
			a.mu.Unlock()

			if a.trustCancel != nil {
				a.trustCancel()
				a.trustCancel = nil
			}

			wailsRuntime.EventsEmit(a.ctx, "device:disconnected", map[string]string{
				"udid": udid,
			})
		}
	}
}

// startTrustPolling 每 1 秒輪詢信任狀態，60 秒超時後 emit trust:timeout。
// 使用者按「我已信任」按鈕會另外走 TriggerTrustCheck 強制立即檢查，不等 tick。
func (a *App) startTrustPolling(udid string) {
	if a.trustCancel != nil {
		a.trustCancel()
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.trustCancel = cancel

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		timeout := time.After(60 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-timeout:
				wailsRuntime.EventsEmit(a.ctx, "trust:timeout", nil)
				return
			case <-ticker.C:
				trusted, err := a.CheckTrustStatus(udid)
				if err != nil {
					return // 裝置已斷線
				}
				if trusted {
					a.mu.Lock()
					if a.connectedDevice != nil {
						a.connectedDevice.Trusted = true
					}
					a.mu.Unlock()
					wailsRuntime.EventsEmit(a.ctx, "device:trust-changed", map[string]any{
						"udid": udid, "trusted": true,
					})
					return
				}
			}
		}
	}()
}
