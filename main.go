package main

import (
	"embed"

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
		Width:            480,
		Height:           640,
		MinWidth:         400,
		MinHeight:        500,
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
		println("Error:", err.Error())
	}
}
