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
)

// App struct
type App struct {
	ctx context.Context
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
func (a *App) shutdown(ctx context.Context) {}

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
