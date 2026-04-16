package backup

// BackupError 備份錯誤（可序列化給前端）
type BackupError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
}

func (e *BackupError) Error() string { return e.Message }

// 預定義錯誤碼
var (
	ErrDeviceDisconnected = &BackupError{Code: "DEVICE_DISCONNECTED", Message: "iPhone 已斷開連線", Recoverable: false}
	ErrDiskFull           = &BackupError{Code: "DISK_FULL", Message: "磁碟空間不足，請釋放空間後繼續", Recoverable: true}
	ErrPermissionDenied   = &BackupError{Code: "PERMISSION_DENIED", Message: "沒有權限寫入此資料夾，請選擇其他位置", Recoverable: true}
	ErrAFCConnectFailed   = &BackupError{Code: "AFC_CONNECT_FAILED", Message: "無法存取 iPhone 照片，請確認 iPhone 已解鎖", Recoverable: true}
	ErrAFCTimeout         = &BackupError{Code: "AFC_TIMEOUT", Message: "iPhone 連線不穩，請換一條 USB 線或重新插拔", Recoverable: true}
	ErrTrustTimeout       = &BackupError{Code: "TRUST_TIMEOUT", Message: "等待超時，請確認 iPhone 已解鎖並點信任", Recoverable: true}
	ErrUnsupportedIOS     = &BackupError{Code: "IOS_UNSUPPORTED", Message: "你的 iOS 版本暫不支援，需 iOS 14 以上", Recoverable: false}
)

// isDeviceDisconnected 判斷錯誤是否為裝置斷線
func isDeviceDisconnected(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// go-ios 斷線時常見的錯誤字串
	for _, keyword := range []string{"connection reset", "broken pipe", "EOF", "device not found"} {
		for i := 0; i < len(msg)-len(keyword)+1; i++ {
			if msg[i:i+len(keyword)] == keyword {
				return true
			}
		}
	}
	return false
}

// humanizeError 將技術錯誤轉為人類可讀的中文訊息
func humanizeError(err error) string {
	if err == nil {
		return ""
	}
	if be, ok := err.(*BackupError); ok {
		return be.Message
	}
	if isDeviceDisconnected(err) {
		return "傳輸中途 iPhone 斷開連線"
	}
	return "寫入失敗，請確認磁碟有足夠空間"
}
