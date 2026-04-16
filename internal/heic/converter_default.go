//go:build !windows || cgo

package heic

import (
	"context"
	"image/jpeg"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// ConvertAll 掃描 backupPath 中所有 HEIC/HEIF，以 heicWorkerCount 個 goroutine 並行轉為 JPEG 副本。
func (c *Converter) ConvertAll(ctx context.Context, backupPath string) (*ConvertResult, error) {
	heicFiles := scanHeicFiles(backupPath)
	toConvert := filterAlreadyConverted(heicFiles)

	result := &ConvertResult{}
	total := len(toConvert)

	if total == 0 {
		c.emitFn("heic:complete", map[string]any{"converted": 0, "failed": 0})
		return result, nil
	}

	fileCh := make(chan string, len(toConvert))
	for _, f := range toConvert {
		fileCh <- f
	}
	close(fileCh)

	var converted, failed, done atomic.Int64
	var wg sync.WaitGroup

	workers := min(heicWorkerCount, total)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for heicPath := range fileCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				jpgPath := jpegPath(heicPath)
				if err := c.convertOne(heicPath, jpgPath); err != nil {
					failed.Add(1)
				} else {
					converted.Add(1)
				}
				n := done.Add(1)
				c.emitFn("heic:progress", map[string]any{
					"total":   total,
					"done":    int(n),
					"percent": float64(n) / float64(total) * 100,
				})
			}
		}()
	}
	wg.Wait()

	result.Converted = int(converted.Load())
	result.Failed = int(failed.Load())
	c.emitFn("heic:complete", map[string]any{
		"converted": result.Converted,
		"failed":    result.Failed,
	})
	return result, nil
}

// convertOne 單一 HEIC → JPEG 轉檔（使用 Go image pipeline）
// 先寫入暫存檔，成功後再 rename，避免轉檔失敗留下 0 bytes 殘留。
func (c *Converter) convertOne(heicPath, jpgPath string) error {
	img, exifData, err := decodeHEIC(heicPath)
	if err != nil {
		return err
	}

	if exifData != nil {
		img = applyExifOrientation(img, exifData)
	}

	tmp, err := os.CreateTemp(filepath.Dir(jpgPath), ".ivault_heic_*.jpg")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if err := jpeg.Encode(tmp, img, &jpeg.Options{Quality: c.quality}); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, jpgPath); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
