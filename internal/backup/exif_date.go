package backup

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// ReadShootDate 根據副檔名分派正確的日期讀取策略。
// 支援：JPEG/JPG（goexif）、HEIC/HEIF（bytes scan）、MOV/MP4/M4V（video_date.go）。
// 全部失敗回傳 (zero, false)，呼叫端用 time.Now() fallback。
func ReadShootDate(localPath string) (time.Time, bool) {
	ext := strings.ToLower(filepath.Ext(localPath))
	switch ext {
	case ".jpg", ".jpeg":
		return readJPEGShootDate(localPath)
	case ".heic", ".heif":
		return readHEICShootDate(localPath)
	case ".mov", ".mp4", ".m4v":
		return ReadVideoShootDate(localPath)
	default:
		return time.Time{}, false
	}
}

// readJPEGShootDate 從 JPEG 讀取 EXIF DateTimeOriginal。
func readJPEGShootDate(localPath string) (time.Time, bool) {
	f, err := os.Open(localPath)
	if err != nil {
		return time.Time{}, false
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return time.Time{}, false
	}

	t, err := x.DateTime()
	if err != nil {
		return time.Time{}, false
	}

	if t.Year() < 2000 || t.Year() > 2100 {
		return time.Time{}, false
	}
	return t, true
}

// readHEICShootDate 從 HEIC/HEIF 讀取拍攝日期。
//
// HEIC 是 ISOBMFF 容器，EXIF metadata 藏在 meta/iinf/iloc box 中，
// 以 "Exif\x00\x00" 起頭，後面接標準 TIFF header（II 或 MM）。
// 做法：讀取前 512KB，找到 magic bytes，把後面的 TIFF 資料交給 goexif 解析。
func readHEICShootDate(localPath string) (time.Time, bool) {
	f, err := os.Open(localPath)
	if err != nil {
		return time.Time{}, false
	}
	defer f.Close()

	// HEIC meta box 通常在檔案開頭 512KB 內
	lr := &io.LimitedReader{R: f, N: 512 * 1024}
	data, err := io.ReadAll(lr)
	if err != nil || len(data) < 16 {
		return time.Time{}, false
	}

	magic := []byte("Exif\x00\x00")
	idx := bytes.Index(data, magic)
	if idx < 0 {
		return time.Time{}, false
	}

	// TIFF header 緊接在 magic 後面（II 或 MM 開頭）
	tiffData := data[idx+len(magic):]
	if len(tiffData) < 8 {
		return time.Time{}, false
	}

	x, err := exif.Decode(bytes.NewReader(tiffData))
	if err != nil {
		return time.Time{}, false
	}

	t, err := x.DateTime()
	if err != nil {
		return time.Time{}, false
	}

	if t.Year() < 2000 || t.Year() > 2100 {
		return time.Time{}, false
	}
	return t, true
}
