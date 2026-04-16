package backup

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"time"
)

// qtEpoch 是 QuickTime/MP4 時間戳的紀元（1904-01-01 00:00:00 UTC）。
var qtEpoch = time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)

// ReadVideoShootDate 從 QuickTime / MP4 容器讀取拍攝時間。
//
// iPhone 錄製的 .MOV / .MP4 / .M4V 均為 ISOBMFF 容器，
// 時間存在 moov/mvhd atom 的 creation_time 欄位。
// iPhone 會把 moov box 放在檔案開頭（fast-start），
// 讀前 4MB 足以涵蓋所有正常情況。
func ReadVideoShootDate(localPath string) (time.Time, bool) {
	f, err := os.Open(localPath)
	if err != nil {
		return time.Time{}, false
	}
	defer f.Close()

	lr := &io.LimitedReader{R: f, N: 4 * 1024 * 1024}
	data, err := io.ReadAll(lr)
	if err != nil || len(data) < 16 {
		return time.Time{}, false
	}

	return parseMvhd(data)
}

// parseMvhd 在原始 bytes 中找 mvhd box 並解析 creation_time。
//
// mvhd box 結構：
//   [size:4][type:"mvhd":4][version:1][flags:3][creation_time:4or8]...
//
// version 0 → creation_time 是 uint32（秒，自 1904-01-01）
// version 1 → creation_time 是 uint64（秒，自 1904-01-01）
func parseMvhd(data []byte) (time.Time, bool) {
	mvhdType := []byte("mvhd")
	idx := bytes.Index(data, mvhdType)
	if idx < 0 {
		return time.Time{}, false
	}

	// content 從 type 後面開始：[version:1][flags:3][creation_time:4or8]...
	content := data[idx+4:]
	if len(content) < 8 {
		return time.Time{}, false
	}

	version := content[0]
	// content[1:4] 是 flags，略過

	var creationSecs uint64
	switch version {
	case 0:
		// version/flags = 4 bytes，creation_time = uint32 at offset 4
		if len(content) < 8 {
			return time.Time{}, false
		}
		creationSecs = uint64(binary.BigEndian.Uint32(content[4:8]))
	case 1:
		// version/flags = 4 bytes，creation_time = uint64 at offset 4
		if len(content) < 12 {
			return time.Time{}, false
		}
		creationSecs = binary.BigEndian.Uint64(content[4:12])
	default:
		return time.Time{}, false
	}

	if creationSecs == 0 {
		return time.Time{}, false
	}

	t := qtEpoch.Add(time.Duration(creationSecs) * time.Second)
	if t.Year() < 2000 || t.Year() > 2100 {
		return time.Time{}, false
	}
	return t, true
}
