//go:build windows

package platform

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// hiddenCmdTimeout 是所有偵測類 hidden cmd 的預設 timeout。
// watchDevices 每 3 秒輪詢 checkAppleDevices，任何 hidden cmd 卡住
// （PowerShell 首次啟動、WMI 服務異常、reg 鎖住等）都會凍結 UI。
const hiddenCmdTimeout = 8 * time.Second

// runHiddenOutput 執行 hidden cmd 並取得 stdout，帶 timeout 保護。
// timeout 命中會 kill process、回 context.DeadlineExceeded。
func runHiddenOutput(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Output()
}

// aumidRegex 限制 Windows AppUserModelId 只能含安全字元，防止 Get-AppxPackage 輸出
// 被竄改後導致 shell:AppsFolder\{AUMID} 拼接出意外的路徑或指令。
// 格式：{PackageName}_{PublisherHash}!{EntryPoint}
var aumidRegex = regexp.MustCompile(`^[A-Za-z0-9._-]+_[A-Za-z0-9]+![A-Za-z0-9]+$`)

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

// hiddenCmd 建立一個帶有 CREATE_NO_WINDOW 旗標的 exec.Cmd。
// 從 Wails GUI 程序啟動 console 程式時，若不設此旗標會閃出黑色視窗。
func hiddenCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

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

	// 方法 2：MS Store Apple Devices — 查 AppxPackage
	// USBAAPL64 kernel driver 解安裝後會殘留為 STOPPED，不能作為判斷依據
	// PackageFamilyName 有值才代表真正安裝中
	out, err := runHiddenOutput(hiddenCmdTimeout, "powershell", "-NoProfile", "-Command",
		"(Get-AppxPackage AppleInc.AppleDevices).PackageFamilyName")
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return true
	}

	// 方法 3：iTunes legacy service（sc.exe 避免 PowerShell 別名問題）
	out2, err2 := runHiddenOutput(hiddenCmdTimeout, "sc.exe", "query", "Apple Mobile Device Service")
	if err2 == nil && strings.Contains(string(out2), "SERVICE_NAME") {
		return true
	}

	return false
}

func checkHeicSupported() bool {
	// 檢查是否有 HEIC 副檔名關聯（Windows 11 通常有）
	out, err := runHiddenOutput(hiddenCmdTimeout, "reg", "query", `HKEY_CLASSES_ROOT\.heic`)
	return err == nil && len(out) > 0
}

func detectDarkModeWindows() bool {
	out, err := runHiddenOutput(hiddenCmdTimeout, "reg", "query",
		`HKEY_CURRENT_USER\SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		"/v", "AppsUseLightTheme")
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

// GetDiskInfo 取得指定路徑的磁碟空間。
// 若目標路徑不存在（例如備份資料夾被刪），沿路徑往上找到第一個存在的目錄，
// 確保磁碟可用空間永遠能正確顯示（至少顯示磁碟根目錄的剩餘空間）。
func GetDiskInfo(path string) DiskInfo {
	p := path
	for {
		if _, err := os.Stat(p); err == nil {
			break
		}
		parent := filepath.Dir(p)
		if parent == p {
			break // 到達根目錄仍找不到，直接用原路徑
		}
		p = parent
	}
	return getDiskSpace(p)
}

// OpenFolder 用 Explorer 開啟資料夾
// 不用 hiddenCmd：explorer.exe 是 GUI 程式，HideWindow 旗標會讓 Explorer
// 視窗以 SW_HIDE 建立，導致資料夾視窗開啟後看不見。
func OpenFolder(path string) error {
	return exec.Command("explorer", path).Start()
}

// OpenURL 用系統瀏覽器開啟 URL
func OpenURL(url string) error {
	return hiddenCmd("rundll32", "url.dll,FileProtocolHandler", url).Start()
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
	if err := hiddenCmd("winget", installArgs...).Run(); err == nil {
		return nil
	}

	// Step 2：install 失敗（可能已安裝舊版）→ 嘗試升級
	upgradeArgs := append([]string{"upgrade", id}, commonArgs...)
	return hiddenCmd("winget", upgradeArgs...).Run()
	// 若 upgrade 也失敗（已是最新版），pollDriverInstall() 仍會透過路徑偵測判斷
}

// RecheckAppleDevices 重新偵測 Apple Devices 安裝狀態（不使用快取）
func RecheckAppleDevices() bool {
	return checkAppleDevices()
}

// DetectInstallStage 偵測 Apple Devices 安裝進度階段
// 回傳："downloading" | "installing" | "starting"
func DetectInstallStage() string {
	out, err := runHiddenOutput(hiddenCmdTimeout, "sc", "query", "Apple Mobile Device Service")
	if err != nil {
		// 服務尚未寫入 = 下載中
		return "downloading"
	}
	s := string(out)
	if strings.Contains(s, "RUNNING") || strings.Contains(s, "START_PENDING") {
		return "starting"
	}
	if strings.Contains(s, "SERVICE_NAME") {
		// 服務已存在但尚未啟動
		return "installing"
	}
	return "downloading"
}

// WMIDetectIPhone 透過 WMI 偵測已連接的 iPhone（不需要 Apple Devices 驅動）
// 回傳裝置名稱（如 "Apple iPhone"），未偵測到則回傳空字串
func WMIDetectIPhone() string {
	out, err := runHiddenOutput(hiddenCmdTimeout, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		"(Get-WmiObject Win32_PnPEntity | Where-Object { $_.Name -like '*iPhone*' } | Select-Object -First 1 -ExpandProperty Name)")
	if err != nil {
		return ""
	}
	name := strings.TrimSpace(string(out))
	if strings.Contains(name, "iPhone") {
		return name
	}
	return ""
}

// IsAMDSReady 檢查 Apple Mobile Device Service 是否正在 listening port 27015。
// go-ios 透過此 port 與 iOS 裝置通訊。
// MS Store 版 Apple Devices 的 AppleMobileDeviceProcess.exe 提供此服務，
// 但只有 Apple Devices UI 被啟動過後才會運行（見 BUG-002 文件）。
func IsAMDSReady() bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:27015", 300*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// EnsureAMDSRunning 確保 AMDS 背景服務運行中。
// 流程：
//  1. 若 port 27015 已 listening → 回 nil
//  2. 找 Apple Devices AUMID
//  3. explorer.exe shell:AppsFolder\{AUMID} 喚起
//  4. 輪詢 port 27015 最多 8s
//  5. 偵測到就緒 → 背景 goroutine kill UI → 回 nil
//  6. 超時 → 回 error
func EnsureAMDSRunning() error {
	if IsAMDSReady() {
		return nil
	}

	aumid, err := findAppleDevicesAUMID()
	if err != nil {
		return fmt.Errorf("find Apple Devices AUMID: %w", err)
	}

	// 用 explorer.exe 走 UWP activation path
	// 不能直接 exec AppleMobileDeviceLauncher.exe（WindowsApps ACL 會拒絕）
	if err := hiddenCmd("explorer.exe", "shell:AppsFolder\\"+aumid).Start(); err != nil {
		return fmt.Errorf("launch Apple Devices: %w", err)
	}

	// 輪詢最多 20 秒（MS Store Apple Devices 冷啟動需要較長時間）
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		if IsAMDSReady() {
			// 背景 goroutine 延遲 kill UI，不阻塞呼叫端
			go killAppleDevicesUI()
			return nil
		}
	}

	return fmt.Errorf("AMDS_TIMEOUT")
}

// findAppleDevicesAUMID 動態取得 Apple Devices 的 AppUserModelId。
// 不寫死 package family name，避免 Apple 改版後失效。
// 格式：{PackageFamilyName}!App，例如 AppleInc.AppleDevices_nzyj5cx40ttqa!App
func findAppleDevicesAUMID() (string, error) {
	out, err := runHiddenOutput(hiddenCmdTimeout, "powershell", "-NoProfile", "-Command",
		"(Get-AppxPackage AppleInc.AppleDevices).PackageFamilyName")
	if err != nil {
		return "", err
	}
	familyName := strings.TrimSpace(string(out))
	if familyName == "" {
		return "", fmt.Errorf("Apple Devices package not installed")
	}
	aumid := familyName + "!App"
	if !aumidRegex.MatchString(aumid) {
		return "", fmt.Errorf("unexpected AUMID format: %q", aumid)
	}
	return aumid, nil
}

// killAppleDevicesUI 關閉 Apple Devices UI 窗，保留背景 AppleMobileDeviceProcess.exe。
// 延遲 1 秒執行，避免 UI 還在初始化時就被 kill。
func killAppleDevicesUI() {
	time.Sleep(1 * time.Second)
	_ = hiddenCmd("taskkill", "/F", "/IM", "AppleDevices.exe").Run()
}

// ── 系統睡眠控制（AD：備份中阻止睡眠）────────────────────────
var setThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")

const (
	esSystemRequired = 0x00000001
	esContinuous     = 0x80000000
)

// PreventSleep 備份中阻止系統進入睡眠
func PreventSleep() {
	setThreadExecutionState.Call(uintptr(esContinuous | esSystemRequired))
}

// AllowSleep 恢復系統睡眠策略
func AllowSleep() {
	setThreadExecutionState.Call(uintptr(esContinuous))
}

// ── 開機自動啟動（O）────────────────────────────────────────
const autostartRegKey = `HKEY_CURRENT_USER\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`

// SetAutostart 設定或移除開機自動啟動登錄值
func SetAutostart(enabled bool) error {
	if !enabled {
		_ = hiddenCmd("reg", "delete", autostartRegKey, "/v", "iVault", "/f").Run()
		return nil
	}
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	return hiddenCmd("reg", "add", autostartRegKey, "/v", "iVault", "/t", "REG_SZ", "/d", exePath, "/f").Run()
}

// GetAutostart 查詢是否已設定開機自動啟動
func GetAutostart() bool {
	out, err := runHiddenOutput(hiddenCmdTimeout, "reg", "query", autostartRegKey, "/v", "iVault")
	return err == nil && strings.Contains(string(out), "iVault")
}

// ── Toast 通知（P）──────────────────────────────────────────

// ShowToast 發送 Windows Toast 通知
func ShowToast(title, body string) {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	t, b := r.Replace(title), r.Replace(body)
	script := `[void][Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime];` +
		`[void][Windows.Data.Xml.Dom.XmlDocument,Windows.Data.Xml.Dom.XmlDocument,ContentType=WindowsRuntime];` +
		`$x=[Windows.Data.Xml.Dom.XmlDocument]::new();` +
		`$x.LoadXml('<toast><visual><binding template="ToastGeneric"><text>` + t + `</text><text>` + b + `</text></binding></visual></toast>');` +
		`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('iVault').Show([Windows.UI.Notifications.ToastNotification]::new($x))`
	_ = hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Run()
}

// ── iTunes 衝突偵測（AA）────────────────────────────────────

// IsITunesRunning 偵測 iTunes 是否正在執行（可能干擾 AFC 通訊）
func IsITunesRunning() bool {
	out, err := runHiddenOutput(hiddenCmdTimeout, "tasklist",
		"/FI", "IMAGENAME eq iTunes.exe", "/NH", "/FO", "CSV")
	return err == nil && strings.Contains(string(out), "iTunes.exe")
}

// ── Apple Devices 啟動（I）──────────────────────────────────

// LaunchAppleDevices 直接啟動 Apple Devices App（不透過 Store）
func LaunchAppleDevices() {
	aumid, err := findAppleDevicesAUMID()
	if err != nil {
		return
	}
	_ = hiddenCmd("explorer.exe", "shell:AppsFolder\\"+aumid).Start()
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
