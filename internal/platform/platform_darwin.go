//go:build darwin

package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func detectPlatformSpecific(info *Info) {
	info.AppleDevicesInstalled = true // macOS 原生支援
	info.HeicSupported = true         // macOS 原生支援 HEIC
	info.DarkMode = detectDarkModeMac()
}

func detectDarkModeMac() bool {
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "Dark"
}

// GetDefaultBackupPath macOS：~/Pictures/iVault Backup
func GetDefaultBackupPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Pictures", "iVault Backup")
}

// GetNonSystemDisks 列出 /Volumes/ 下的磁碟
func GetNonSystemDisks() []DiskInfo {
	entries, _ := os.ReadDir("/Volumes")
	var disks []DiskInfo
	for _, entry := range entries {
		path := filepath.Join("/Volumes", entry.Name())
		disk := getDiskSpace(path)
		disk.IsSystem = (entry.Name() == "Macintosh HD")
		disks = append(disks, disk)
	}
	return disks
}

// GetDiskInfo 取得指定路徑的磁碟空間
func GetDiskInfo(path string) DiskInfo {
	return getDiskSpace(path)
}

// OpenFolder 用 Finder 開啟資料夾
func OpenFolder(path string) error {
	return exec.Command("open", path).Start()
}

// OpenURL 用系統瀏覽器開啟 URL
func OpenURL(url string) error {
	return exec.Command("open", url).Start()
}

// HasWinget macOS 無 winget
func HasWinget() bool { return false }

// InstallAppleDevicesViaWinget macOS 不需要（原生支援）
func InstallAppleDevicesViaWinget() error { return nil }

// RecheckAppleDevices macOS 永遠回傳 true
func RecheckAppleDevices() bool { return true }

func getDiskSpace(path string) DiskInfo {
	disk := DiskInfo{Path: path}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return disk
	}
	disk.TotalSpace = int64(stat.Blocks) * stat.Bsize
	disk.FreeSpace = int64(stat.Bavail) * stat.Bsize
	return disk
}
