package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ivault/internal/config"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "iVault",
		Width:            600,
		Height:           700,
		MinWidth:         520,
		MinHeight:        600,
		DisableResize:    false,
		Fullscreen:       false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// 透明背景（讓 Mica / Vibrancy 顯示）
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},

		// macOS：毛玻璃效果
		Mac: &mac.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           mac.DefaultAppearance,
			About: &mac.AboutInfo{
				Title:   "iVault",
				Message: "iPhone 照片備份工具",
			},
		},

		// Windows 11：Mica 材質
		Windows: &windows.Options{
			BackdropType:                      windows.Mica,
			DisableFramelessWindowDecorations: false,
			WebviewIsTransparent:              true,
			Theme:                             windows.SystemDefault,
		},
	})

	if err != nil {
		logStartupError(err)
	}
}

// logStartupError 將啟動失敗寫入使用者資料夾下的 startup-error.log，
// 以 0600 附加模式（避免敏感路徑/UDID 被其他帳號讀取）。
// 同時 fallback 到 stderr 供開發模式使用。
func logStartupError(err error) {
	logDir := filepath.Dir(config.ConfigPath())
	if mkErr := os.MkdirAll(logDir, 0755); mkErr == nil {
		logPath := filepath.Join(logDir, "startup-error.log")
		entry := fmt.Sprintf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), err.Error())
		if f, openErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600); openErr == nil {
			_, _ = f.WriteString(entry)
			_ = f.Close()
		}
	}
	println("Error:", err.Error())
}
