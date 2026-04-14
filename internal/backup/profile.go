package backup

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ProfileEntry 單一檔案各階段耗時紀錄
type ProfileEntry struct {
	FileName   string
	SizeBytes  int64
	Ext        string
	AFCCopyMs  int64 // CopyFileBuffered（afcOpen + read loop + afcClose 合計）
	ExifMs     int64 // ReadShootDate
	RenameMs   int64 // ResolveByDate（os.Rename）
	ManifestMs int64 // manifest.MarkDone
	TotalMs    int64 // 上述全部合計
}

// Profiler 備份效能紀錄器
type Profiler struct {
	entries   []ProfileEntry
	startTime time.Time
	csvPath   string
	mu        sync.Mutex // 保護並行 worker 同時 Add
}

// NewProfiler 建立 Profiler；CSV 存在 {backupPath}/.ivault/ 下
func NewProfiler(backupPath string) *Profiler {
	ts := time.Now().Format("20060102_150405")
	return &Profiler{
		startTime: time.Now(),
		csvPath:   filepath.Join(backupPath, ".ivault", "ivault_profile_"+ts+".csv"),
	}
}

// Add 新增一筆 per-file 紀錄（thread-safe）
func (p *Profiler) Add(e ProfileEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries = append(p.entries, e)
}

// Save 將所有紀錄寫成 CSV（0600 權限）
func (p *Profiler) Save() error {
	if err := os.MkdirAll(filepath.Dir(p.csvPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(p.csvPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{
		"filename", "size_bytes", "ext",
		"afc_copy_ms", "exif_ms", "rename_ms", "manifest_ms", "total_ms",
	})
	for _, e := range p.entries {
		_ = w.Write([]string{
			e.FileName,
			fmt.Sprintf("%d", e.SizeBytes),
			e.Ext,
			fmt.Sprintf("%d", e.AFCCopyMs),
			fmt.Sprintf("%d", e.ExifMs),
			fmt.Sprintf("%d", e.RenameMs),
			fmt.Sprintf("%d", e.ManifestMs),
			fmt.Sprintf("%d", e.TotalMs),
		})
	}
	return nil
}

// PrintSummary 備份結束後印出統計摘要（wails dev 模式的 stdout 可見）
func (p *Profiler) PrintSummary() {
	n := len(p.entries)
	if n == 0 {
		return
	}

	var sumCopy, sumExif, sumRename, sumManifest, sumTotal int64
	var sumBytes int64
	totals := make([]int64, n)

	for i, e := range p.entries {
		sumCopy += e.AFCCopyMs
		sumExif += e.ExifMs
		sumRename += e.RenameMs
		sumManifest += e.ManifestMs
		sumTotal += e.TotalMs
		sumBytes += e.SizeBytes
		totals[i] = e.TotalMs
	}

	sort.Slice(totals, func(i, j int) bool { return totals[i] < totals[j] })
	p50 := totals[n*50/100]
	p95 := totals[n*95/100]

	elapsed := time.Since(p.startTime).Round(time.Second)
	var avgMBs float64
	if s := elapsed.Seconds(); s > 0 {
		avgMBs = float64(sumBytes) / 1024 / 1024 / s
	}

	avg := func(v int64) int64 {
		if n == 0 {
			return 0
		}
		return v / int64(n)
	}
	pct := func(v int64) float64 {
		if sumTotal == 0 {
			return 0
		}
		return float64(v) / float64(sumTotal) * 100
	}

	line := strings.Repeat("─", 56)
	fmt.Printf("\n%s\n", line)
	fmt.Printf("  iVault Profile Summary\n")
	fmt.Printf("%s\n", line)
	fmt.Printf("  Files: %-5d  Size: %-8.1f MB  Time: %-8s  Speed: %.1f MB/s\n\n",
		n, float64(sumBytes)/1024/1024, elapsed, avgMBs)

	fmt.Printf("  %-14s  %9s  %8s  %6s\n", "Phase", "total(ms)", "avg(ms)", "%")
	fmt.Printf("  %-14s  %9d  %8d  %5.1f%%\n", "AFC Copy",    sumCopy,     avg(sumCopy),     pct(sumCopy))
	fmt.Printf("  %-14s  %9d  %8d  %5.1f%%\n", "EXIF Parse",  sumExif,     avg(sumExif),     pct(sumExif))
	fmt.Printf("  %-14s  %9d  %8d  %5.1f%%\n", "Rename",      sumRename,   avg(sumRename),   pct(sumRename))
	fmt.Printf("  %-14s  %9d  %8d  %5.1f%%\n", "Manifest",    sumManifest, avg(sumManifest), pct(sumManifest))
	fmt.Printf("  %-14s  %9d  %8d\n\n",         "TOTAL",       sumTotal,    avg(sumTotal))

	fmt.Printf("  Per-file latency:  p50=%dms  p95=%dms\n\n", p50, p95)

	// 檔案大小分佈
	type szBucket struct {
		label    string
		maxBytes int64
	}
	buckets := []szBucket{
		{"< 100 KB",   100 * 1024},
		{"100KB – 1MB", 1024 * 1024},
		{"1MB – 5MB",   5 * 1024 * 1024},
		{"5MB – 20MB",  20 * 1024 * 1024},
		{"> 20 MB",     1<<62},
	}
	fmt.Printf("  File size distribution:\n")
	prev := int64(0)
	for _, b := range buckets {
		var cnt int
		var sz int64
		for _, e := range p.entries {
			if e.SizeBytes > prev && e.SizeBytes <= b.maxBytes {
				cnt++
				sz += e.SizeBytes
			}
		}
		if cnt > 0 {
			fmt.Printf("    %-14s  %4d files  %7.1f MB\n", b.label, cnt, float64(sz)/1024/1024)
		}
		prev = b.maxBytes
	}

	fmt.Printf("\n  CSV → %s\n", p.csvPath)
	fmt.Printf("%s\n\n", line)
}
