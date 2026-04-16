package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// ConfigPath 回傳各平台的設定檔路徑
// Windows: %APPDATA%\iVault\config.json
// macOS:   ~/Library/Application Support/iVault/config.json
func ConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "iVault", "config.json")
	default: // darwin, linux
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "iVault", "config.json")
	}
}
