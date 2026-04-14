//go:build !windows || cgo

package heic

import (
	"context"
	"image/jpeg"
	"os"
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

	workers := heicWorkerCount
	if total < workers {
		workers = total
	}
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
