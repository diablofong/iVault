package heic

import (
	"encoding/binary"
	"image"
)

// parseExifOrientation 從原始 EXIF bytes 解析 Orientation tag (0x0112)
// 回傳 1-8，預設 1（正常方向）
func parseExifOrientation(data []byte) int {
	// EXIF 格式：可能以 "Exif\x00\x00" 開頭（JPEG JFIF 包裝）
	offset := 0
	if len(data) >= 6 && string(data[0:6]) == "Exif\x00\x00" {
		offset = 6
	}
	if offset+8 > len(data) {
		return 1
	}

	// TIFF 位元組序標頭
	var order binary.ByteOrder
	switch string(data[offset : offset+2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}

	// IFD0 offset（相對 TIFF 標頭起點）
	ifd0Offset := int(order.Uint32(data[offset+4:]))
	ifdStart := offset + ifd0Offset
	if ifdStart+2 > len(data) {
		return 1
	}

	numEntries := int(order.Uint16(data[ifdStart:]))
	for i := range numEntries {
		pos := ifdStart + 2 + i*12
		if pos+12 > len(data) {
			break
		}
		tag := order.Uint16(data[pos:])
		if tag == 0x0112 { // Orientation
			return int(order.Uint16(data[pos+8:]))
		}
	}
	return 1
}

// rotate90CW 順時針旋轉 90°
func rotate90CW(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(h-1-y, x, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

// rotate90CCW 逆時針旋轉 90°
func rotate90CCW(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(y, w-1-x, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

// rotate180 旋轉 180°
func rotate180(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(w-1-x, h-1-y, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

