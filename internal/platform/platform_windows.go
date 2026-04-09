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
	// 方法 1：傳統 iTunes Win32 安裝路徑（最快速，路徑存在即可用）
	paths := []string{
		`C:\Program Files\Common Files\Apple\Mobile Device Support`,
		`C:\Program Files (x86)\Common Files\Apple\Mobile Device Support`,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}

	// 方法 2：Apple Mobile Device Service 必須是 RUNNING 狀態。
	// ⚠️  不能只看服務名稱是否存在（安裝進行中就會寫入）或 Registry（更早），
	//     必須等到 STATE : 4 RUNNING 才代表 Apple Devices / iTunes 真正可用。
	out, err := exec.Command("sc", "query", "Apple Mobile Device Service").Output()
	if err == nil && strings.Contains(string(out), "RUNNING") {
		return true
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

// HasWinget 檢查系統是否有 winget CLI
func HasWinget() bool {
	_, err := exec.LookPath("winget")
	return err == nil
}

// InstallAppleDevicesViaWinget 執行安裝或升級（blocking，應在 goroutine 中呼叫）
// 策略：先 install；失敗時 fallback 到 upgrade（處理已安裝舊版的情況）
func InstallAppleDevicesViaWinget() error {
	const id = "9NP83LWLPZ9K"
	commonArgs := []string{
		"--source", "msstore",
		"--accept-package-agreements",
		"--accept-source-agreements",
	}

	// Step 1：全新安裝
	installArgs := append([]string{"install", id}, commonArgs...)
	if err := exec.Command("winget", installArgs...).Run(); err == nil {
		return nil
	}

	// Step 2：install 失敗（可能已安裝舊版）→ 嘗試升級
	upgradeArgs := append([]string{"upgrade", id}, commonArgs...)
	return exec.Command("winget", upgradeArgs...).Run()
	// 若 upgrade 也失敗（已是最新版），pollDriverInstall() 仍會透過路徑偵測判斷
}

// RecheckAppleDevices 重新偵測 Apple Devices 安裝狀態（不使用快取）
func RecheckAppleDevices() bool {
	return checkAppleDevices()
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
