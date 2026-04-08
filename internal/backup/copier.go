package backup

import (
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
