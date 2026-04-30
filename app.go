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
	info := *a.platformInfo
	info.IsDevMode = (Version == "dev")
	return info
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

// TriggerTrustCheck 由前端「我已信任，繼續」按鈕觸發，立即查一次信任狀態。
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

// CheckAppleDevicesInstalled Windows 專用：每次都做即時偵測
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

// LaunchAppleDevices Windows：直接啟動 Apple Devices App
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
	if last := a.configMgr.Get().DefaultBackupPath; last != "" {
		return last
	}
	return platform.GetDefaultBackupPath()
}

// GetBackupFolderSize 計算指定路徑（含子目錄）的總佔用空間（bytes）
func (a *App) GetBackupFolderSize(path string) int64 {
	if path == "" {
		return 0
	}
	var total int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

// CheckBackupPath 進入 READY 頁前呼叫：確認備份路徑是否仍有效。
func (a *App) CheckBackupPath(path string) {
	checkPath := path
	if checkPath == "" {
		checkPath = a.configMgr.Get().DefaultBackupPath
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

// ManifestExists 檢查備份路徑下是否存有指定裝置的 manifest
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

	// 從 DeviceConfig 取得 FolderName（確保換名後資料夾路徑不變）
	appCfg := a.configMgr.Get()
	if appCfg.Devices != nil {
		if devCfg, ok := appCfg.Devices[cfg.DeviceUDID]; ok && devCfg.FolderName != "" {
			cfg.FolderName = devCfg.FolderName
		}
	}
	if cfg.FolderName == "" {
		cfg.FolderName = config.BuildFolderName(cfg.DeviceName, cfg.DeviceUDID)
	}

	// 更新 DefaultBackupPath 與 per-device BackupPath
	if appCfg.Devices == nil {
		appCfg.Devices = make(map[string]*config.DeviceConfig)
	}
	if _, ok := appCfg.Devices[cfg.DeviceUDID]; !ok {
		appCfg.Devices[cfg.DeviceUDID] = &config.DeviceConfig{
			Name:       cfg.DeviceName,
			FolderName: cfg.FolderName,
		}
	}
	appCfg.Devices[cfg.DeviceUDID].BackupPath = cfg.BackupPath
	appCfg.DefaultBackupPath = cfg.BackupPath
	_ = a.configMgr.Save(appCfg)

	platform.PreventSleep() // AD: 備份中阻止系統睡眠

	emitFn := func(event string, data any) {
		wailsRuntime.EventsEmit(a.ctx, event, data)
	}
	engine := backup.NewEngine(cfg, emitFn)

	go func() {
		defer platform.AllowSleep()
		result, err := engine.Run(ctx)
		if err != nil && err != context.Canceled {
			cur := a.configMgr.Get()
			if cur.Devices != nil {
				if dev, ok := cur.Devices[cfg.DeviceUDID]; ok {
					dev.LastInterrupted = true
				}
			}
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
				cur := a.configMgr.Get()
				if cur.Devices != nil {
					if dev, ok := cur.Devices[cfg.DeviceUDID]; ok {
						dev.LastInterrupted = true
						dev.InterruptedDone = result.InterruptedDone
						dev.InterruptedTotal = result.InterruptedTotal
					}
				}
				_ = a.configMgr.Save(cur)
				wailsRuntime.EventsEmit(a.ctx, "backup:interrupted", result)
				return
			}

			// 成功完成：更新 DeviceConfig
			cur := a.configMgr.Get()
			if cur.Devices == nil {
				cur.Devices = make(map[string]*config.DeviceConfig)
			}
			if _, ok := cur.Devices[cfg.DeviceUDID]; !ok {
				cur.Devices[cfg.DeviceUDID] = &config.DeviceConfig{
					Name:       cfg.DeviceName,
					FolderName: cfg.FolderName,
					BackupPath: cfg.BackupPath,
				}
			}
			dev := cur.Devices[cfg.DeviceUDID]
			dev.LastInterrupted = false
			dev.InterruptedDone = 0
			dev.InterruptedTotal = 0
			dev.FirstBackupDone = true
			dev.PhotosCount = result.PhotosCount
			dev.VideosCount = result.VideosCount
			dev.LastBackupDate = time.Now().Format(time.RFC3339)
			_ = a.configMgr.Save(cur)

			go platform.ShowToast(
				"備份完成",
				fmt.Sprintf("已備份 %d 張照片、%d 段影片", result.PhotosCount, result.VideosCount),
			)
			wailsRuntime.EventsEmit(a.ctx, "backup:complete", result)

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

// GetBackupHistory 回傳備份歷史（legacy API，遷移後回傳空陣列）
func (a *App) GetBackupHistory() []config.BackupRecord {
	return a.configMgr.Get().History
}

// OpenFolder 用系統檔案管理員開啟指定資料夾。
func (a *App) OpenFolder(path string) error {
	if path == "" {
		return fmt.Errorf("invalid path")
	}
	if strings.HasPrefix(path, `\\`) || strings.HasPrefix(path, "//") {
		return fmt.Errorf("UNC path not allowed")
	}
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

// OpenURL 用系統瀏覽器開啟 URL（只允許 http/https）。
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
func (a *App) watchDevices() {
	driverWasMissing := false
	amdsFailNotified := false

	for {
		if a.platformInfo != nil && a.platformInfo.OS == "windows" {
			if !platform.RecheckAppleDevices() {
				if !driverWasMissing {
					time.Sleep(600 * time.Millisecond)
					deviceName := platform.WMIDetectIPhone()
					wailsRuntime.EventsEmit(a.ctx, "driver:required", map[string]any{
						"deviceName": deviceName,
					})
					driverWasMissing = true
				}
				time.Sleep(3 * time.Second)
				continue
			}
			if driverWasMissing {
				a.mu.Lock()
				if a.platformInfo != nil {
					a.platformInfo.AppleDevicesInstalled = true
				}
				a.mu.Unlock()
				wailsRuntime.EventsEmit(a.ctx, "driver:installed", nil)
				driverWasMissing = false
			}

			if !platform.IsAMDSReady() {
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
			amdsFailNotified = false
		}

		a.runDeviceListener()
		time.Sleep(3 * time.Second)
	}
}

func (a *App) runDeviceListener() {
	receiver, closeFunc, err := ios.Listen()
	if err != nil {
		return
	}
	defer closeFunc()

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

			// 確保 DeviceConfig 存在（建立 FolderName）
			a.ensureDeviceConfig(udid, info.Name)

			wailsRuntime.EventsEmit(a.ctx, "device:connected", info)

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

// ensureDeviceConfig 確保指定裝置有 DeviceConfig；若尚無則建立並儲存。
// 已存在時不覆蓋（保護不可變的 FolderName）。
func (a *App) ensureDeviceConfig(udid, name string) {
	if udid == "" {
		return
	}
	cur := a.configMgr.Get()
	if cur.Devices == nil {
		cur.Devices = make(map[string]*config.DeviceConfig)
	}
	if _, exists := cur.Devices[udid]; !exists {
		cur.Devices[udid] = &config.DeviceConfig{
			Name:       name,
			FolderName: config.BuildFolderName(name, udid),
		}
		_ = a.configMgr.Save(cur)
	}
}

// startTrustPolling 每 1 秒輪詢信任狀態，60 秒超時後 emit trust:timeout。
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
					return
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
