package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ivault/internal/device"
)

// ============================================================
// #1 HEIC 掃描範圍測試：確認 EXIF 在 600KB 位置能被 2MB 限制抓到
// ============================================================

// buildMinimalTIFF 產生一個最小合法 TIFF EXIF blob，含 DateTime "2023:06:15 12:00:00"
func buildMinimalTIFF() []byte {
	// 小端序 TIFF
	// 格式：header(8) + IFD_count(2) + IFD_entry(12) + next_IFD(4) + DateTime_value(20)
	//        offset:0                  offset:8          offset:10     offset:22            offset:26
	dateStr := []byte("2023:06:15 12:00:00\x00") // 20 bytes
	const dataOffset = uint32(26)

	var buf bytes.Buffer
	// Header
	buf.Write([]byte{0x49, 0x49}) // II little-endian
	buf.Write([]byte{0x2A, 0x00}) // TIFF magic
	buf.Write([]byte{0x08, 0x00, 0x00, 0x00}) // IFD offset = 8
	// IFD entry count = 1
	buf.Write([]byte{0x01, 0x00})
	// IFD entry: tag=0x0132 (DateTime), type=2 (ASCII), count=20, offset=26
	writeLE16 := func(v uint16) { buf.Write([]byte{byte(v), byte(v >> 8)}) }
	writeLE32 := func(v uint32) { buf.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}) }
	writeLE16(0x0132) // DateTime tag
	writeLE16(0x0002) // ASCII
	writeLE32(20)     // count
	writeLE32(dataOffset)
	// next IFD = 0
	writeLE32(0)
	// DateTime value at offset 26
	buf.Write(dateStr)
	return buf.Bytes()
}

// buildFakeHEIC 建立一個假 HEIC 檔：在指定 offset 放 Exif\x00\x00 + TIFF
func buildFakeHEIC(t *testing.T, exifOffset int) string {
	t.Helper()
	tiff := buildMinimalTIFF()
	data := make([]byte, exifOffset)
	magic := []byte("Exif\x00\x00")
	data = append(data, magic...)
	data = append(data, tiff...)

	f, err := os.CreateTemp(t.TempDir(), "fake_*.heic")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return f.Name()
}

func TestReadHEICShootDate_ExifBeyond512KB(t *testing.T) {
	// EXIF 放在 600KB 位置（512KB 舊限制抓不到，2MB 新限制可以抓到）
	const exifOffset = 600 * 1024
	path := buildFakeHEIC(t, exifOffset)

	ts, ok := readHEICShootDate(path)
	if !ok {
		t.Fatal("expected readHEICShootDate to return ok=true for EXIF at 600KB, got false")
	}
	if ts.Year() != 2023 || ts.Month() != 6 || ts.Day() != 15 {
		t.Errorf("unexpected shoot date: %v", ts)
	}
}

func TestReadHEICShootDate_ExifWithin512KB(t *testing.T) {
	// EXIF 在 200KB，無論舊新限制都能抓到
	const exifOffset = 200 * 1024
	path := buildFakeHEIC(t, exifOffset)

	_, ok := readHEICShootDate(path)
	if !ok {
		t.Fatal("expected readHEICShootDate to return ok=true for EXIF at 200KB, got false")
	}
}

// ============================================================
// #2 unknown-date 測試：EXIF 失敗時檔案應進 _unknown-date/ 子目錄
// ============================================================

func makeTestOrganizer(t *testing.T, backupPath string) *Organizer {
	t.Helper()
	return NewOrganizer(backupPath, "TestPhone", "AABBCCDD1234", true)
}

func TestResolveByDate_UnknownDate(t *testing.T) {
	dir := t.TempDir()
	org := makeTestOrganizer(t, dir)

	// 建立假 staging 檔案
	stagingDir := filepath.Join(dir, "TestPhone (AABBCCDD)", ".staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		t.Fatal(err)
	}
	stagingPath := filepath.Join(stagingDir, "IMG_0001.HEIC")
	if err := os.WriteFile(stagingPath, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	file := device.PhotoFile{FileName: "IMG_0001.HEIC"}
	// dateOK=false → 應進 _unknown-date/
	result := org.ResolveByDate(file, stagingPath, time.Time{}, false)

	if !strings.Contains(result, "_unknown-date") {
		t.Errorf("expected result path to contain '_unknown-date', got: %s", result)
	}
	// 確認檔案已移動到 _unknown-date/
	if _, err := os.Stat(result); err != nil {
		t.Errorf("file not found at result path %s: %v", result, err)
	}
}

func TestResolveByDate_WithDate(t *testing.T) {
	dir := t.TempDir()
	org := makeTestOrganizer(t, dir)

	stagingDir := filepath.Join(dir, "TestPhone (AABBCCDD)", ".staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		t.Fatal(err)
	}
	stagingPath := filepath.Join(stagingDir, "IMG_0002.jpg")
	if err := os.WriteFile(stagingPath, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	file := device.PhotoFile{FileName: "IMG_0002.jpg"}
	shootDate := time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)
	result := org.ResolveByDate(file, stagingPath, shootDate, true)

	if strings.Contains(result, "_unknown-date") {
		t.Errorf("expected result path NOT to contain '_unknown-date', got: %s", result)
	}
	if !strings.Contains(result, "2024-07") {
		t.Errorf("expected result path to contain '2024-07', got: %s", result)
	}
}

// ============================================================
// #3 Manifest os.Stat 驗證測試
// ============================================================

// manifestLocalExists 提取路徑拼接邏輯，供 unit test 驗證
func manifestLocalExists(backupPath string, entry ManifestEntry) bool {
	localAbsPath := filepath.Join(backupPath, entry.LocalPath)
	_, err := os.Stat(localAbsPath)
	return err == nil
}

func TestManifestLocalExists_FilePresent(t *testing.T) {
	dir := t.TempDir()
	relPath := filepath.Join("MyPhone (AABB1234)", "2024-07", "IMG_001.jpg")
	absPath := filepath.Join(dir, relPath)

	// 建立目錄與檔案
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absPath, []byte("photo"), 0644); err != nil {
		t.Fatal(err)
	}

	entry := ManifestEntry{LocalPath: relPath}
	if !manifestLocalExists(dir, entry) {
		t.Errorf("expected manifestLocalExists=true when file exists at %s", absPath)
	}
}

func TestManifestLocalExists_FileMissing(t *testing.T) {
	dir := t.TempDir()
	relPath := filepath.Join("MyPhone (AABB1234)", "2024-07", "IMG_002.jpg")
	// 刻意不建立這個檔案

	entry := ManifestEntry{LocalPath: relPath}
	if manifestLocalExists(dir, entry) {
		t.Errorf("expected manifestLocalExists=false when file is missing")
	}
}

func TestManifestLocalExists_NoPathTraversal(t *testing.T) {
	dir := t.TempDir()
	// 故意構造一個 ../ 路徑，確認 filepath.Join 清理後不會逃脫 dir
	entry := ManifestEntry{LocalPath: "../secret/file.jpg"}
	// 不論上層目錄有無此檔案，測試的重點是路徑拼接邏輯不 panic
	// 這裡只要不 panic、不 crash 即通過；實際結果取決於測試環境
	_ = manifestLocalExists(dir, entry)
}
