package device

import "strings"

// DeviceInfo 裝置基本資訊（偵測到裝置時立即可用）
type DeviceInfo struct {
	UDID       string `json:"udid"`
	Name       string `json:"name"`       // "iPhone 15 Pro"
	Model      string `json:"model"`      // "iPhone16,1"
	IOSVersion string `json:"iosVersion"` // "17.4.1"
	Trusted    bool   `json:"trusted"`
}

// DeviceDetail 裝置詳細資訊（需要 AFC 連線後才能取得）
type DeviceDetail struct {
	DeviceInfo
	PhotoCount int64 `json:"photoCount"` // DCIM 下檔案數估算
	UsedSpace  int64 `json:"usedSpace"`  // bytes
	TotalSpace int64 `json:"totalSpace"` // bytes
}

// PhotoFile 單一照片/影片檔案資訊
type PhotoFile struct {
	RemotePath string `json:"remotePath"` // "/DCIM/100APPLE/IMG_0001.HEIC"
	FileName   string `json:"fileName"`   // "IMG_0001.HEIC"
	Size       int64  `json:"size"`       // bytes
	ModTime    int64  `json:"modTime"`    // Unix timestamp（秒）
}

// RelativePath 回傳相對於 /DCIM/ 的路徑，作為 manifest key
// 例："/DCIM/100APPLE/IMG_0001.HEIC" → "100APPLE/IMG_0001.HEIC"
func (f PhotoFile) RelativePath() string {
	parts := strings.SplitN(f.RemotePath, "/DCIM/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return f.FileName
}
