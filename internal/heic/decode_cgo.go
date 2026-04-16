//go:build cgo

package heic

import (
	"image"
	"os"

	"github.com/adrium/goheif"
)

// decodeHEIC CGo 版本：使用 goheif 解碼
func decodeHEIC(path string) (image.Image, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	img, err := goheif.Decode(f)
	if err != nil {
		return nil, nil, err
	}

	// ExtractExif 接受 io.ReaderAt，可直接傳 *os.File
	exifData, _ := goheif.ExtractExif(f)
	return img, exifData, nil
}
