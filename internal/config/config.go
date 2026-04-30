package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DeviceConfig 每台裝置的獨立備份狀態（per-device 架構，v1.0.0+）
type DeviceConfig struct {
	Name             string `json:"name"`
	FolderName       string `json:"folderName"`
	BackupPath       string `json:"backupPath,omitempty"`
	LastBackupDate   string `json:"lastBackupDate,omitempty"`
	PhotosCount      int    `json:"photosCount"`
	VideosCount      int    `json:"videosCount"`
	FirstBackupDone  bool   `json:"firstBackupDone"`
	LastInterrupted  bool   `json:"lastInterrupted"`
	InterruptedDone  int    `json:"interruptedDone"`
	InterruptedTotal int    `json:"interruptedTotal"`
}

// AppConfig 應用程式設定（持久化到 JSON）
type AppConfig struct {
	// 當前欄位（per-device 架構）
	DefaultBackupPath string                   `json:"defaultBackupPath,omitempty"`
	ConvertHeic       bool                     `json:"convertHeic"`
	OrganizeByDate    bool                     `json:"organizeByDate"`
	OnboardingDone    bool                     `json:"onboardingDone"`
	Devices           map[string]*DeviceConfig `json:"devices,omitempty"`

	// Legacy 欄位：僅供 migration 讀取，遷移完成後清除
	LastBackupPath         string         `json:"lastBackupPath,omitempty"`
	History                []BackupRecord `json:"history,omitempty"`
	LastInterrupted        bool           `json:"lastInterrupted,omitempty"`
	InterruptedDone        int            `json:"interruptedDone,omitempty"`
	InterruptedTotal       int            `json:"interruptedTotal,omitempty"`
	InterruptedDeviceUDID  string         `json:"interruptedDeviceUdid,omitempty"`
	FirstBackupDone        bool           `json:"firstBackupDone,omitempty"`
	FirstBackupDoneDevices []string       `json:"firstBackupDoneDevices,omitempty"`
}

// BackupRecord 歷史備份紀錄（legacy，僅供 migration 讀取）
type BackupRecord struct {
	Date        string `json:"date"`
	DeviceName  string `json:"deviceName"`
	DeviceUDID  string `json:"deviceUdid"`
	NewFiles    int    `json:"newFiles"`
	PhotosCount int    `json:"photosCount"`
	VideosCount int    `json:"videosCount"`
	Skipped     int    `json:"skipped"`
	Failed      int    `json:"failed"`
	TotalBytes  int64  `json:"totalBytes"`
	Duration    string `json:"duration"`
}

// BuildFolderName 產生裝置備份資料夾名稱，格式："{sanitized name} ({udid[:8]})"
// 與 backup/organizer.go folderName() 保持一致。
func BuildFolderName(name, udid string) string {
	r := strings.NewReplacer(
		`\`, "_", `/`, "_", `:`, "_",
		`*`, "_", `?`, "_", `"`, "_",
		`<`, "_", `>`, "_", `|`, "_",
	)
	cleaned := strings.TrimSpace(r.Replace(name))
	if cleaned == "" {
		cleaned = "iPhone"
	}
	if len(udid) >= 8 {
		return fmt.Sprintf("%s (%s)", cleaned, udid[:8])
	}
	return cleaned
}

// Manager 設定管理器
type Manager struct {
	cfg      AppConfig
	filePath string
}

// NewManager 建立設定管理器並載入設定
func NewManager() *Manager {
	m := &Manager{
		filePath: ConfigPath(),
		cfg: AppConfig{
			OrganizeByDate: true,
		},
	}
	m.load()
	return m
}

// Get 回傳當前設定（副本）
func (m *Manager) Get() AppConfig {
	return m.cfg
}

// Save 儲存設定到磁碟（atomic：先寫 .tmp 再 rename，避免 crash 時 config 損毀）
func (m *Manager) Save(cfg AppConfig) error {
	m.cfg = cfg

	if err := os.MkdirAll(filepath.Dir(m.filePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.cfg, "", "  ")
	if err != nil {
		return err
	}

	tmp := m.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, m.filePath); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (m *Manager) load() {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return // 檔案不存在，使用預設值
	}
	if err := json.Unmarshal(data, &m.cfg); err != nil {
		// 損毀：備份壞檔供使用者事後檢查，並重置為預設值
		backup := m.filePath + ".corrupt-" + time.Now().UTC().Format("20060102-150405")
		_ = os.WriteFile(backup, data, 0600)
		m.cfg = AppConfig{OrganizeByDate: true}
		return
	}
	m.migrate()
}

// migrate 偵測舊格式 config（有 History 無 Devices），一次性遷移為 per-device 架構。
// 遷移完成後清除 legacy 欄位並存檔，後續啟動不再觸發。
func (m *Manager) migrate() {
	cfg := &m.cfg
	if len(cfg.Devices) > 0 || len(cfg.History) == 0 {
		return // 已是新格式或無資料
	}

	cfg.Devices = make(map[string]*DeviceConfig)

	seen := make(map[string]bool)
	for _, rec := range cfg.History {
		if rec.DeviceUDID == "" || seen[rec.DeviceUDID] {
			continue
		}
		seen[rec.DeviceUDID] = true

		isInterrupted := cfg.LastInterrupted &&
			(cfg.InterruptedDeviceUDID == "" || cfg.InterruptedDeviceUDID == rec.DeviceUDID)

		firstDone := cfg.FirstBackupDone
		if !firstDone {
			for _, uid := range cfg.FirstBackupDoneDevices {
				if uid == rec.DeviceUDID {
					firstDone = true
					break
				}
			}
		}

		cfg.Devices[rec.DeviceUDID] = &DeviceConfig{
			Name:             rec.DeviceName,
			FolderName:       BuildFolderName(rec.DeviceName, rec.DeviceUDID),
			LastBackupDate:   rec.Date,
			PhotosCount:      rec.PhotosCount,
			VideosCount:      rec.VideosCount,
			FirstBackupDone:  firstDone,
			LastInterrupted:  isInterrupted,
			InterruptedDone:  cfg.InterruptedDone,
			InterruptedTotal: cfg.InterruptedTotal,
		}
	}

	if cfg.DefaultBackupPath == "" && cfg.LastBackupPath != "" {
		cfg.DefaultBackupPath = cfg.LastBackupPath
	}

	// 清除 legacy 欄位（omitempty 確保不再寫入 JSON）
	cfg.History = nil
	cfg.LastBackupPath = ""
	cfg.LastInterrupted = false
	cfg.InterruptedDone = 0
	cfg.InterruptedTotal = 0
	cfg.InterruptedDeviceUDID = ""
	cfg.FirstBackupDone = false
	cfg.FirstBackupDoneDevices = nil

	_ = m.Save(*cfg)
}
