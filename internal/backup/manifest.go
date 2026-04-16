package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ivault/internal/device"
)

// Manifest 斷點續傳記錄
type Manifest struct {
	Version     int                      `json:"version"`
	DeviceUDID  string                   `json:"deviceUDID"`
	DeviceName  string                   `json:"deviceName"`
	CreatedAt   string                   `json:"createdAt"`
	UpdatedAt   string                   `json:"updatedAt"`
	Interrupted bool                     `json:"interrupted"` // 上次備份是否被中斷
	Files       map[string]ManifestEntry `json:"files"`

	filePath string    // 不序列化
	mu       sync.Mutex // 保護並行 worker 同時 MarkDone / Save
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

// MarkDone 標記檔案已備份（thread-safe）
func (m *Manifest) MarkDone(file device.PhotoFile, localPath string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Files[file.RelativePath()] = ManifestEntry{
		Size:       file.Size,
		ModTime:    file.ModTime,
		LocalPath:  localPath,
		BackedUpAt: time.Now().Format(time.RFC3339),
	}
	m.UpdatedAt = time.Now().Format(time.RFC3339)
}

// Save 寫入磁碟（thread-safe；中斷時 interrupted=true，完成時 interrupted=false）
func (m *Manifest) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveLocked()
}

// saveLocked 寫入磁碟（呼叫方須持有 m.mu）
// atomic：先寫 .tmp 再 rename，避免 crash 時 manifest 損毀
// （損毀會讓斷點續傳誤判「全部已是最新」而跳過真實未備份檔）。
func (m *Manifest) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(m.filePath), 0755); err != nil {
		return err
	}
	m.UpdatedAt = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(m, "", "  ")
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

// SaveInterrupted 標記中斷狀態並寫入磁碟（thread-safe）
func (m *Manifest) SaveInterrupted() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Interrupted = true
	return m.saveLocked()
}

// SaveCompleted 標記完成狀態並寫入磁碟（thread-safe）
func (m *Manifest) SaveCompleted() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Interrupted = false
	return m.saveLocked()
}

// ManifestExists 檢查備份路徑下是否有該裝置的 manifest 檔案
func ManifestExists(backupPath, udid string) bool {
	_, err := os.Stat(manifestPath(backupPath, udid))
	return err == nil
}

func manifestPath(backupPath, udid string) string {
	return filepath.Join(backupPath, ".ivault", "manifest-"+sanitizeUDID(udid)+".json")
}

// sanitizeUDID 把 UDID 轉成安全的檔名組件：只保留 [A-Za-z0-9]，其餘一律轉 `-`。
// 阻止 `..`、路徑分隔符、控制字元等被當成檔名組件。
func sanitizeUDID(udid string) string {
	var b strings.Builder
	for _, r := range udid {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	s := b.String()
	if s == "" {
		return "unknown"
	}
	return s
}
