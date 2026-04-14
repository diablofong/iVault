// testseeder：透過 AFC 向 iPhone 的 /DCIM/ 資料夾寫入大量假媒體檔案，
// 供 iVault 備份速度測試（特別是多 AFC session POC）使用。
//
// 用法：
//
//	go run ./cmd/testseeder                       # 預設：100 個 50MB 混合檔（~5 GB）
//	go run ./cmd/testseeder -count 20 -size 100   # 快速小量 2 GB
//	go run ./cmd/testseeder -clean                # 刪除已寫入的測試檔案
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"ivault/internal/device"

	"github.com/danielpaulus/go-ios/ios/afc"
)

const (
	chunkSize  = 256 * 1024 // 256 KB write buffer
	seedPrefix = "SEED_"    // 測試檔案名前綴（供 -clean 識別）
)

var (
	udidFlag  = flag.String("udid", "", "iPhone UDID（空白時自動取第一台）")
	countFlag = flag.Int("count", 100, "寫入檔案數量")
	sizeFlag  = flag.Int("size", 50, "每個檔案大小（MB）")
	mixFlag   = flag.Bool("mix", true, "混合副檔名：70% .MOV、20% .MP4、10% .HEIC")
	extFlag   = flag.String("ext", "", "固定副檔名（如 .MOV）；設定後 -mix 無效")
	dirFlag   = flag.String("dir", "999APPLE", "DCIM 子目錄（需以 APPLE 結尾，iVault 才能掃到）")
	cleanFlag = flag.Bool("clean", false, "刪除 /DCIM/{dir}/SEED_* 測試檔案後退出")
)

func main() {
	flag.Parse()

	// 取得 UDID
	udid := *udidFlag
	if udid == "" {
		devices, err := device.ListDevices()
		if err != nil || len(devices) == 0 {
			log.Fatal("找不到連接的 iPhone，請確認 USB 已連接且已信任此電腦")
		}
		udid = devices[0].UDID
		name := devices[0].Name
		if name == "" {
			name = udid
		}
		fmt.Printf("偵測到裝置：%s (%s)\n", name, udid[:8])
	}

	// 建立 AFC 連線
	client, err := device.ConnectAFC(udid)
	if err != nil {
		log.Fatalf("AFC 連線失敗：%v", err)
	}
	defer client.Close()

	dcimDir := "/DCIM/" + *dirFlag

	// -clean 模式：刪除測試檔案
	if *cleanFlag {
		cleanTestFiles(client, dcimDir)
		return
	}

	// 確保目標目錄存在
	_ = client.MkDir(dcimDir)

	count := *countFlag
	sizeMB := *sizeFlag
	sizeBytes := int64(sizeMB) * 1024 * 1024

	totalBytes := sizeBytes * int64(count)
	fmt.Printf("準備寫入 %d 個 %d MB 檔案，共 %.1f GB → /DCIM/%s/\n\n",
		count, sizeMB, float64(totalBytes)/1024/1024/1024, *dirFlag)

	// 固定 pattern buffer（比 crypto/rand 快，足夠測速用）
	buf := make([]byte, chunkSize)
	for i := range buf {
		buf[i] = byte(i & 0xFF)
	}

	var totalWritten int64
	startAll := time.Now()

	for i := 1; i <= count; i++ {
		ext := pickExt(i, *mixFlag, *extFlag)
		filename := fmt.Sprintf("%s/%s%04d%s", dcimDir, seedPrefix, i, ext)

		tStart := time.Now()
		written, err := writeFile(client, filename, sizeBytes, buf)
		elapsed := time.Since(tStart)

		if err != nil {
			fmt.Printf("[%3d/%d] %-20s  FAIL: %v\n", i, count, seedPrefix+fmt.Sprintf("%04d", i)+ext, err)
			continue
		}

		totalWritten += written
		speedMBs := float64(written) / 1024 / 1024 / elapsed.Seconds()

		// ETA 估算（依目前平均速度）
		overallElapsed := time.Since(startAll)
		var eta string
		if totalWritten > 0 && totalWritten < totalBytes {
			remaining := totalBytes - totalWritten
			avgBps := float64(totalWritten) / overallElapsed.Seconds()
			etaDur := time.Duration(float64(remaining)/avgBps) * time.Second
			eta = "ETA " + fmtDuration(etaDur)
		} else if i == count {
			eta = "完成"
		}

		fmt.Printf("[%3d/%d] %-24s %4d MB  ✓  %5.1f MB/s  %s\n",
			i, count,
			seedPrefix+fmt.Sprintf("%04d", i)+ext,
			sizeMB, speedMBs, eta)
	}

	totalElapsed := time.Since(startAll)
	avgSpeed := float64(totalWritten) / 1024 / 1024 / totalElapsed.Seconds()
	fmt.Printf("\n寫入完成：%.1f GB  總耗時 %s  平均 %.1f MB/s\n",
		float64(totalWritten)/1024/1024/1024,
		fmtDuration(totalElapsed),
		avgSpeed)
	fmt.Printf("清除指令：go run ./cmd/testseeder -clean -dir %s\n", *dirFlag)
}

// pickExt 依混合模式輪流選副檔名（10 個一輪：7 MOV / 2 MP4 / 1 HEIC）
func pickExt(i int, mix bool, override string) string {
	if override != "" {
		// 確保有點
		if !strings.HasPrefix(override, ".") {
			return "." + override
		}
		return override
	}
	if !mix {
		return ".MOV"
	}
	switch (i - 1) % 10 {
	case 9:
		return ".HEIC"
	case 7, 8:
		return ".MP4"
	default:
		return ".MOV"
	}
}

// writeFile 以 chunkSize 為單位寫入 sizeBytes 大小的檔案
func writeFile(client *afc.Client, path string, sizeBytes int64, chunk []byte) (int64, error) {
	f, err := client.Open(path, afc.WRITE_ONLY_CREATE_TRUNC)
	if err != nil {
		return 0, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	var written int64
	for written < sizeBytes {
		toWrite := int64(len(chunk))
		if written+toWrite > sizeBytes {
			toWrite = sizeBytes - written
		}
		n, err := f.Write(chunk[:toWrite])
		if err != nil {
			return written, fmt.Errorf("write at offset %d: %w", written, err)
		}
		written += int64(n)
	}
	return written, nil
}

// cleanTestFiles 刪除 dcimDir 下所有 SEED_ 前綴的測試檔案
func cleanTestFiles(client *afc.Client, dcimDir string) {
	files, err := client.List(dcimDir)
	if err != nil {
		fmt.Printf("無法列出 %s：%v\n", dcimDir, err)
		return
	}

	deleted := 0
	for _, f := range files {
		if !strings.HasPrefix(f, seedPrefix) {
			continue
		}
		p := dcimDir + "/" + f
		if err := client.Remove(p); err != nil {
			fmt.Printf("刪除失敗 %s：%v\n", f, err)
		} else {
			fmt.Printf("已刪除 %s\n", f)
			deleted++
		}
	}
	fmt.Printf("\n共刪除 %d 個測試檔案\n", deleted)
}

// fmtDuration 格式化為 Xm Ys 或 Xs
func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
