package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ivault/internal/device"
)

// Manifest 斷點續傳記錄
type Manifest struct {
	Version    int                      `json:"version"`
	DeviceUDID string                   `json:"deviceUDID"`
	DeviceName string                   `json:"deviceName"`
	CreatedAt  string                   `json:"createdAt"`
	UpdatedAt  string                   `json:"updatedAt"`
	Files      map[string]ManifestEntry `json:"files"`

	filePath string // 不序列化
}

// ManifestEntry 單一檔案的備份紀錄
type ManifestEntry struct {
	Size       int64  `json:"size"`
	ModTime    int64  `json:"modTime"`
	LocalPath  string `json:"localPath"`
	BackedUpAt string `json:"backedUpAt"`
}

// LoadOrCreateManifest 載入既有 manifest 或建立新的
// 存放於 {backupPath}/.ivault/manifest-{udid}.json
func LoadOrCreateManifest(backupPath, udid, deviceName string) *Manifest {
	filePath := manifestPath(backupPath, udid)

	m := &Manifest{
		Version:    1,
		DeviceUDID: udid,
		DeviceName: deviceName,
		CreatedAt:  time.Now().Format(time.RFC3339),
		Files:      make(map[string]ManifestEntry),
		filePath:   filePath,
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return m // 新建
	}

	existing := &Manifest{}
	if err := json.Unmarshal(data, existing); err != nil {
		return m // 損毀，重新建立
	}
	existing.filePath = filePath
	if existing.Files == nil {
		existing.Files = make(map[string]ManifestEntry)
	}
	return existing
}

// IsBackedUp 檢查檔案是否已備份（比對 key + size + modTime）
func (m *Manifest) IsBackedUp(file device.PhotoFile) bool {
	entry, exists := m.Files[file.RelativePath()]
	if !exists {
		return false
	}
	return entry.Size == file.Size && entry.ModTime == file.ModTime
}

// MarkDone 標記檔案已備份
func (m *Manifest) MarkDone(file device.PhotoFile, localPath string) {
	m.Files[file.RelativePath()] = ManifestEntry{
		Size:       file.Size,
		ModTime:    file.ModTime,
		LocalPath:  localPath,
		BackedUpAt: time.Now().Format(time.RFC3339),
	}
	m.UpdatedAt = time.Now().Format(time.RFC3339)
}

// Save 寫入磁碟
func (m *Manifest) Save() error {
	if err := os.MkdirAll(filepath.Dir(m.filePath), 0755); err != nil {
		return err
	}
	m.UpdatedAt = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0644)
}

func manifestPath(backupPath, udid string) string {
	// 清理 UDID 中的特殊字元避免路徑問題
	safeUDID := strings.ReplaceAll(udid, ":", "-")
	return filepath.Join(backupPath, ".ivault", "manifest-"+safeUDID+".json")
}
