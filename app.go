package main

import (
	"context"

	"ivault/internal/device"
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
