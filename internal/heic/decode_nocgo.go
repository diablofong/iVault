//go:build !cgo

package heic

import (
	"errors"
	"image"
)

// decodeHEIC non-CGo 版本：不支援 HEIC 解碼
// Windows 無 CGo 環境時回傳此錯誤，ConvertAll 會將每個檔案計入 Failed
func decodeHEIC(_ string) (image.Image, []byte, error) {
	return nil, nil, errors.New("HEIC 轉檔需要 CGo 支援（目前平台不支援）")
}
