package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"ivault/internal/backup"
	"ivault/internal/device"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	backupCancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is about to quit
func (a *App) shutdown(ctx context.Context) {
	if a.backupCancel != nil {
		a.backupCancel()
	}
}

// ListDevices 回傳所有已連接的 iOS 裝置
func (a *App) ListDevices() ([]device.DeviceInfo, error) {
	return device.ListDevices()
}

// GetDeviceDetail 取得裝置詳細資訊（含照片數、儲存空間）
func (a *App) GetDeviceDetail(udid string) (*device.DeviceDetail, error) {
	return device.GetDeviceDetail(udid)
}

// ScanDCIM 掃描裝置 DCIM 目錄，回傳照片清單
func (a *App) ScanDCIM(udid string) ([]device.PhotoFile, error) {
	return device.ScanDCIM(udid)
}

// StartBackup 開始備份（非同步，進度透過 Event 推送）
func (a *App) StartBackup(cfg backup.BackupConfig) error {
	// 若已有備份在執行，先取消
	if a.backupCancel != nil {
		a.backupCancel()
	}

	ctx, cancel := context.WithCancel(a.ctx)
	a.backupCancel = cancel

	emitFn := func(event string, data any) {
		wailsRuntime.EventsEmit(a.ctx, event, data)
	}

	engine := backup.NewEngine(cfg, emitFn)

	go func() {
		result, err := engine.Run(ctx)
		if err != nil {
			// 非取消導致的錯誤
			if err != context.Canceled {
				be, ok := err.(*backup.BackupError)
				if ok {
					wailsRuntime.EventsEmit(a.ctx, "backup:error", map[string]any{
						"code":        be.Code,
						"message":     be.Message,
						"recoverable": be.Recoverable,
					})
				} else {
					wailsRuntime.EventsEmit(a.ctx, "backup:error", map[string]any{
						"code":        "UNKNOWN_ERROR",
						"message":     "發生未預期的錯誤，請重試",
						"recoverable": true,
					})
				}
			}
			return
		}
		wailsRuntime.EventsEmit(a.ctx, "backup:complete", result)
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

// CopyFirstPhoto PoC：複製 DCIM 第一張照片到本機 Pictures\iVault_test\
func (a *App) CopyFirstPhoto(udid string) (*backup.CopyResult, error) {
	photos, err := device.ScanDCIM(udid)
	if err != nil {
		return nil, err
	}
	if len(photos) == 0 {
		return nil, fmt.Errorf("no photos found")
	}

	deviceList, err := ios.ListDevices()
	if err != nil {
		return nil, err
	}
	var d ios.DeviceEntry
	for _, dev := range deviceList.DeviceList {
		if dev.Properties.SerialNumber == udid {
			d = dev
			break
		}
	}

	afcClient, err := afc.New(d)
	if err != nil {
		return nil, fmt.Errorf("afc connect: %w", err)
	}
	defer afcClient.Close()

	home, _ := os.UserHomeDir()
	destDir := filepath.Join(home, "Pictures", "iVault_test")

	return backup.CopyFile(afcClient, photos[0].RemotePath, destDir)
}
