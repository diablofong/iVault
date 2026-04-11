//go:build !windows || cgo

package heic

import (
	"context"
	"image/jpeg"
	"os"
)

// ConvertAll 掃描 backupPath 中所有 HEIC/HEIF，轉為 JPEG 副本（非 Windows / CGo 版）
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

// convertOne 單一 HEIC → JPEG 轉檔（使用 Go image pipeline）
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
