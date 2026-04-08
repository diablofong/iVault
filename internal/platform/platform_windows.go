//go:build windows

package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceExW     = kernel32.NewProc("GetDiskFreeSpaceExW")
	getDriveTypeW           = kernel32.NewProc("GetDriveTypeW")
)

const (
	DRIVE_UNKNOWN   = 0
	DRIVE_NO_ROOT   = 1
	DRIVE_REMOVABLE = 2
	DRIVE_FIXED     = 3
	DRIVE_REMOTE    = 4
	DRIVE_CDROM     = 5
	DRIVE_RAMDISK   = 6
)

func detectPlatformSpecific(info *Info) {
	info.AppleDevicesInstalled = checkAppleDevices()
	info.HeicSupported = checkHeicSupported()
	info.DarkMode = detectDarkModeWindows()
}

func checkAppleDevices() bool {
	paths := []string{
		`C:\Program Files\Common Files\Apple\Mobile Device Support`,
		`C:\Program Files (x86)\Common Files\Apple\Mobile Device Support`,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func checkHeicSupported() bool {
	// 檢查是否有 HEIC 副檔名關聯（Windows 11 通常有）
	out, err := exec.Command("reg", "query", `HKEY_CLASSES_ROOT\.heic`).Output()
	return err == nil && len(out) > 0
}

func detectDarkModeWindows() bool {
	out, err := exec.Command("reg", "query",
		`HKEY_CURRENT_USER\SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		"/v", "AppsUseLightTheme").Output()
	if err != nil {
		return false
	}
	// 回傳值 0x0 表示深色模式
	return strings.Contains(string(out), "0x0")
}

// GetDefaultBackupPath Windows：選最大的非系統磁碟
func GetDefaultBackupPath() string {
	disks := GetNonSystemDisks()
	var best *DiskInfo
	for i := range disks {
		if !disks[i].IsSystem && (best == nil || disks[i].FreeSpace > best.FreeSpace) {
			best = &disks[i]
		}
	}
	if best != nil {
		return filepath.Join(best.Path, "iVault Backup")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Pictures", "iVault Backup")
}

// GetNonSystemDisks 掃描 C-Z 磁碟機
func GetNonSystemDisks() []DiskInfo {
	var disks []DiskInfo
	for letter := 'C'; letter <= 'Z'; letter++ {
		root := string(letter) + `:\`
		if !driveExists(root) {
			continue
		}
		disk := getDiskSpace(root)
		disk.IsSystem = (letter == 'C')
		disks = append(disks, disk)
	}
	return disks
}

// GetDiskInfo 取得指定路徑的磁碟空間
func GetDiskInfo(path string) DiskInfo {
	return getDiskSpace(path)
}

// OpenFolder 用 Explorer 開啟資料夾
func OpenFolder(path string) error {
	return exec.Command("explorer", path).Start()
}

// OpenURL 用系統瀏覽器開啟 URL
func OpenURL(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func driveExists(root string) bool {
	ptr, _ := syscall.UTF16PtrFromString(root)
	driveType, _, _ := getDriveTypeW.Call(uintptr(unsafe.Pointer(ptr)))
	return driveType == DRIVE_FIXED || driveType == DRIVE_REMOVABLE
}

func getDiskSpace(path string) DiskInfo {
	disk := DiskInfo{Path: path}
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return disk
	}
	var freeBytesAvailable, totalBytes, totalFree uint64
	r, _, _ := getDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if r != 0 {
		disk.TotalSpace = int64(totalBytes)
		disk.FreeSpace = int64(freeBytesAvailable)
	}
	return disk
}
