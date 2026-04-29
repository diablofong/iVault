package platform

import "runtime"

// Info 平台資訊
type Info struct {
	OS                    string `json:"os"`                    // "darwin" | "windows"
	Arch                  string `json:"arch"`                  // "amd64" | "arm64"
	AppleDevicesInstalled bool   `json:"appleDevicesInstalled"` // Windows only
	HeicSupported         bool   `json:"heicSupported"`
	DarkMode              bool   `json:"darkMode"`
	IsDevMode             bool   `json:"isDevMode"` // true when Version == "dev" (wails dev)
}

// DiskInfo 磁碟空間資訊
type DiskInfo struct {
	Path       string `json:"path"`
	TotalSpace int64  `json:"totalSpace"` // bytes
	FreeSpace  int64  `json:"freeSpace"`  // bytes
	IsSystem   bool   `json:"isSystem"`
}

// Detect 偵測當前平台資訊
func Detect() *Info {
	info := &Info{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
	detectPlatformSpecific(info)
	return info
}
