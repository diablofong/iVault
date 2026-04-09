package backup

import (
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// ReadShootDate 從本地已複製的照片讀取拍攝日期（EXIF DateTimeOriginal）。
// 支援 JPEG/JPG；HEIC 暫不支援（goexif 無法解析），回傳 (zero, false)。
// 呼叫端應在 false 時使用 time.Now() 作為 fallback。
func ReadShootDate(localPath string) (time.Time, bool) {
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

	// 防止異常日期（如 0001-01-01）
	if t.Year() < 2000 || t.Year() > 2100 {
		return time.Time{}, false
	}

	return t, true
}
