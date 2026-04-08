package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ivault/internal/device"
)

// Organizer 負責決定備份檔案的本機路徑
type Organizer struct {
	backupPath     string
	deviceName     string
	organizeByDate bool
}

// NewOrganizer 建立 Organizer
func NewOrganizer(backupPath, deviceName string, organizeByDate bool) *Organizer {
	return &Organizer{
		backupPath:     backupPath,
		deviceName:     sanitizeFilename(deviceName),
		organizeByDate: organizeByDate,
	}
}

// ResolveLocalPath 決定檔案的本機完整路徑（含衝突處理）
// 結構：{backupPath}/{deviceName}/{year-month}/{filename}
func (o *Organizer) ResolveLocalPath(file device.PhotoFile) string {
	var dir string
	if o.organizeByDate && file.ModTime > 0 {
		yearMonth := time.Unix(file.ModTime, 0).Format("2006-01")
		dir = filepath.Join(o.backupPath, o.deviceName, yearMonth)
	} else if o.organizeByDate {
		dir = filepath.Join(o.backupPath, o.deviceName, "Unknown Date")
	} else {
		dir = filepath.Join(o.backupPath, o.deviceName)
	}
	candidate := filepath.Join(dir, file.FileName)
	return o.resolveConflict(candidate)
}

// RelativeLocalPath 回傳相對於 backupPath 的路徑（存入 manifest）
func (o *Organizer) RelativeLocalPath(localPath string) string {
	rel, err := filepath.Rel(o.backupPath, localPath)
	if err != nil {
		return localPath
	}
	return rel
}

func (o *Organizer) resolveConflict(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; i < 10000; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	// 極端情況：使用時間戳後綴
	return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
}

// sanitizeFilename 清除 Windows 不允許的檔名字元
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		`\`, "_", `/`, "_", `:`, "_",
		`*`, "_", `?`, "_", `"`, "_",
		`<`, "_", `>`, "_", `|`, "_",
	)
	return strings.TrimSpace(replacer.Replace(name))
}
