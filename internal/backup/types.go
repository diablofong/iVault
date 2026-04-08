package backup

// BackupConfig 備份設定（前端傳入）
type BackupConfig struct {
	DeviceUDID     string `json:"deviceUdid"`
	DeviceName     string `json:"deviceName"`
	BackupPath     string `json:"backupPath"`
	ConvertHeic    bool   `json:"convertHeic"`
	OrganizeByDate bool   `json:"organizeByDate"`
}

// BackupProgress 備份進度（每個檔案完成時推送）
type BackupProgress struct {
	Phase        string  `json:"phase"`        // "scanning" | "copying" | "converting"
	CurrentFile  string  `json:"currentFile"`
	TotalFiles   int     `json:"totalFiles"`
	DoneFiles    int     `json:"doneFiles"`
	SkippedFiles int     `json:"skippedFiles"`
	TotalBytes   int64   `json:"totalBytes"`
	DoneBytes    int64   `json:"doneBytes"`
	SpeedBps     int64   `json:"speedBps"`
	ETA          string  `json:"eta"` // "2m 30s"
	Percent      float64 `json:"percent"`
}

// BackupResult 備份完成結果
type BackupResult struct {
	Success      bool         `json:"success"`
	NewFiles     int          `json:"newFiles"`
	SkippedFiles int          `json:"skippedFiles"`
	FailedFiles  int          `json:"failedFiles"`
	FailedList   []FailedFile `json:"failedList,omitempty"`
	TotalBytes   int64        `json:"totalBytes"`
	Duration     string       `json:"duration"`
	BackupPath   string       `json:"backupPath"`
	HasHeic      bool         `json:"hasHeic"`
	Interrupted  bool         `json:"interrupted"` // 是否為中途取消
}

// FailedFile 單一失敗檔案
type FailedFile struct {
	FileName string `json:"fileName"`
	Reason   string `json:"reason"` // 人類可讀錯誤原因（中文）
}
