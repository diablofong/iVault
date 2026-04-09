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

// ResolveLocalPath 決定檔案的暫存本機路徑（尚未讀 EXIF）
// 結構：{backupPath}/{deviceName}/{filename}（flat，之後由 ResolveByDate 移動）
func (o *Organizer) ResolveLocalPath(file device.PhotoFile) string {
	dir := filepath.Join(o.backupPath, o.deviceName, ".staging")
	candidate := filepath.Join(dir, file.FileName)
	return o.resolveConflict(candidate)
}

// ResolveByDate 根據 EXIF 拍攝日期決定最終路徑並移動檔案。
// 若 organizeByDate 為 false，直接移動到 {backupPath}/{deviceName}/ 下。
// 若移動失敗，保留在 staging 路徑（不影響備份正確性）。
// 回傳最終實際路徑。
func (o *Organizer) ResolveByDate(file device.PhotoFile, stagingPath string, shootDate time.Time) string {
	var dir string
	if o.organizeByDate {
		yearMonth := shootDate.Format("2006-01")
		dir = filepath.Join(o.backupPath, o.deviceName, yearMonth)
	} else {
		dir = filepath.Join(o.backupPath, o.deviceName)
	}

	finalPath := o.resolveConflictInDir(dir, file.FileName)

	if err := moveFile(stagingPath, finalPath); err != nil {
		// 移動失敗，保留 staging 路徑（仍算備份成功）
		return stagingPath
	}
	return finalPath
}

// resolveConflictInDir 在指定目錄內尋找不衝突的路徑
func (o *Organizer) resolveConflictInDir(dir, fileName string) string {
	candidate := filepath.Join(dir, fileName)
	return o.resolveConflict(candidate)
}

// moveFile 建立目標目錄並移動檔案
func moveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.Rename(src, dst)
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
