package heic

import (
	"context"
	"image"
	"image/jpeg"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ConvertResult 轉檔結果
type ConvertResult struct {
	Converted int
	Failed    int
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

// ConvertAll 掃描 backupPath 中所有 HEIC/HEIF，轉為 JPEG 副本
func (c *Converter) ConvertAll(ctx context.Context, backupPath string) (*ConvertResult, error) {
	heicFiles := scanHeicFiles(backupPath)
	toConvert := filterAlreadyConverted(heicFiles)

	result := &ConvertResult{}
	total := len(toConvert)

	if total == 0 {
		c.emitFn("heic:complete", map[string]any{"converted": 0, "failed": 0})
		return result, nil
	}

	for i, heicPath := range toConvert {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		jpgPath := jpegPath(heicPath)
		if err := c.convertOne(heicPath, jpgPath); err != nil {
			result.Failed++
			continue
		}
		result.Converted++

		c.emitFn("heic:progress", map[string]any{
			"total":   total,
			"done":    i + 1,
			"percent": float64(i+1) / float64(total) * 100,
		})
	}

	c.emitFn("heic:complete", map[string]any{
		"converted": result.Converted,
		"failed":    result.Failed,
	})
	return result, nil
}

// convertOne 單一 HEIC → JPEG 轉檔
func (c *Converter) convertOne(heicPath, jpgPath string) error {
	img, exifData, err := decodeHEIC(heicPath)
	if err != nil {
		return err
	}

	if exifData != nil {
		img = applyExifOrientation(img, exifData)
	}

	out, err := os.Create(jpgPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, img, &jpeg.Options{Quality: c.quality})
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

// filterAlreadyConverted 過濾掉已有 JPEG 副本的檔案
func filterAlreadyConverted(heicFiles []string) []string {
	var result []string
	for _, f := range heicFiles {
		if _, err := os.Stat(jpegPath(f)); os.IsNotExist(err) {
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
