package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/danielpaulus/go-ios/ios/afc"
)

// CopyResult 單次複製結果
type CopyResult struct {
	RemotePath string `json:"remotePath"`
	LocalPath  string `json:"localPath"`
	BytesCopied int64 `json:"bytesCopied"`
}

// CopyFile 從 AFC client 複製單一檔案到本機
// destDir：目標目錄（自動建立）
// remotePath：裝置上的完整路徑，例如 /DCIM/100APPLE/IMG_0001.HEIC
func CopyFile(afcClient *afc.Client, remotePath, destDir string) (*CopyResult, error) {
	// 確保目標目錄存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	// 開啟裝置端檔案
	src, err := afcClient.Open(remotePath, afc.READ_ONLY)
	if err != nil {
		return nil, fmt.Errorf("afc open %s: %w", remotePath, err)
	}
	defer src.Close()

	// 建立本機檔案
	fileName := filepath.Base(remotePath)
	localPath := filepath.Join(destDir, fileName)
	dst, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", localPath, err)
	}
	defer dst.Close()

	// 複製
	n, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("copy %s: %w", remotePath, err)
	}

	return &CopyResult{
		RemotePath:  remotePath,
		LocalPath:   localPath,
		BytesCopied: n,
	}, nil
}

// CopyFileBuffered 使用自訂 buffer 複製檔案（Engine 使用），支援 context 取消。
// remotePath：裝置完整路徑
// localPath：本機完整路徑（呼叫端負責解析，含衝突處理）
// buf：預先分配的 buffer（建議 256KB），在多次呼叫間重複使用
// 若 ctx 在過程中被取消，複製會中止並回傳 ctx.Err()，已寫出的部分檔會被刪除避免殘留。
func CopyFileBuffered(ctx context.Context, afcClient *afc.Client, remotePath, localPath string, buf []byte) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return 0, fmt.Errorf("mkdir: %w", err)
	}

	src, err := afcClient.Open(remotePath, afc.READ_ONLY)
	if err != nil {
		return 0, fmt.Errorf("afc open %s: %w", remotePath, err)
	}
	defer src.Close()

	dst, err := os.Create(localPath)
	if err != nil {
		return 0, fmt.Errorf("create %s: %w", localPath, err)
	}

	n, err := copyWithContext(ctx, dst, src, buf)

	// 先 close，再視情況刪除部分寫入的檔案
	_ = dst.Close()

	if err != nil {
		// 取消或失敗時刪掉殘缺檔，避免下次啟動以為已備份
		_ = os.Remove(localPath)
		if err == context.Canceled || err == context.DeadlineExceeded {
			return n, err
		}
		return n, fmt.Errorf("copy %s: %w", remotePath, err)
	}
	return n, nil
}

// copyWithContext 是 io.CopyBuffer 的 context-aware 版本。
// 每次 Read/Write 迴圈前檢查 ctx，被取消後立即回傳，平均中止延遲 = 單次 Read 時間。
func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if len(buf) == 0 {
		buf = make([]byte, 32*1024)
	}
	var written int64
	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = io.ErrShortWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				return written, nil
			}
			return written, er
		}
	}
}
