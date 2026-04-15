package heic

import (
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// heicWorkerCount 並行轉檔 worker 數量（Windows: 4 PowerShell；macOS/Linux: 4 goroutine）
const heicWorkerCount = 4

// ConvertResult 轉檔結果
type ConvertResult struct {
	Converted int
	Failed    int
}

// splitChunks 將切片分成最多 n 個大致均等的子切片
func splitChunks[T any](items []T, n int) [][]T {
	if n <= 0 {
		n = 1
	}
	if len(items) < n {
		n = len(items)
	}
	chunks := make([][]T, 0, n)
	size := (len(items) + n - 1) / n
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// Converter HEIC → JPEG 批次轉檔器
type Converter struct {
	quality int
	emitFn  func(string, any)
}

// NewConverter 建立轉檔器，quality 1-100（建議 92）
func NewConverter(quality int, emitFn func(string, any)) *Converter {
	if quality <= 0 || quality > 100 {
		quality = 92
	}
	return &Converter{quality: quality, emitFn: emitFn}
}

// scanHeicFiles 遞迴掃描目錄中所有 .heic / .heif 檔案
func scanHeicFiles(root string) []string {
	var files []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".heic" || ext == ".heif" {
			files = append(files, path)
		}
		return nil
	})
	return files
}

// filterAlreadyConverted 過濾掉已有有效 JPEG 副本的檔案。
// 若 .jpg 存在但大小為 0（上次轉檔失敗的殘留），視為未轉換、納入重試。
func filterAlreadyConverted(heicFiles []string) []string {
	var result []string
	for _, f := range heicFiles {
		info, err := os.Stat(jpegPath(f))
		if os.IsNotExist(err) || (err == nil && info.Size() == 0) {
			result = append(result, f)
		}
	}
	return result
}

// jpegPath IMG_0001.HEIC → IMG_0001.jpg
func jpegPath(heicPath string) string {
	ext := filepath.Ext(heicPath)
	return strings.TrimSuffix(heicPath, ext) + ".jpg"
}

// applyExifOrientation 根據 EXIF 方向 tag 旋轉圖片
func applyExifOrientation(img image.Image, exifData []byte) image.Image {
	switch parseExifOrientation(exifData) {
	case 3:
		return rotate180(img)
	case 6:
		return rotate90CW(img)
	case 8:
		return rotate90CCW(img)
	default:
		return img
	}
}
