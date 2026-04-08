package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig 應用程式設定（持久化到 JSON）
type AppConfig struct {
	LastBackupPath string         `json:"lastBackupPath"`
	ConvertHeic    bool           `json:"convertHeic"`
	OrganizeByDate bool           `json:"organizeByDate"`
	History        []BackupRecord `json:"history"`
}

// BackupRecord 歷史備份紀錄
type BackupRecord struct {
	Date       string `json:"date"`       // ISO 8601
	DeviceName string `json:"deviceName"`
	DeviceUDID string `json:"deviceUdid"`
	NewFiles   int    `json:"newFiles"`
	Skipped    int    `json:"skipped"`
	Failed     int    `json:"failed"`
	TotalBytes int64  `json:"totalBytes"`
	Duration   string `json:"duration"`
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
			OrganizeByDate: true, // 預設開啟按日期組織
		},
	}
	m.load()
	return m
}

// Get 回傳當前設定（副本）
func (m *Manager) Get() AppConfig {
	return m.cfg
}

// Save 儲存設定到磁碟
func (m *Manager) Save(cfg AppConfig) error {
	m.cfg = cfg

	if err := os.MkdirAll(filepath.Dir(m.filePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// AddRecord 新增備份紀錄並儲存
func (m *Manager) AddRecord(record BackupRecord) error {
	m.cfg.History = append([]BackupRecord{record}, m.cfg.History...)
	// 只保留最近 50 筆
	if len(m.cfg.History) > 50 {
		m.cfg.History = m.cfg.History[:50]
	}
	return m.Save(m.cfg)
}

func (m *Manager) load() {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return // 檔案不存在，使用預設值
	}
	_ = json.Unmarshal(data, &m.cfg)
}
