package device

import (
	"fmt"
	"path"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
)

// supportedExtensions 備份的檔案格式白名單
var supportedExtensions = map[string]bool{
	".heic": true, ".heif": true,
	".jpg": true, ".jpeg": true,
	".png":  true,
	".dng":  true, ".cr2": true, ".cr3": true,
	".mov": true, ".mp4": true,
	".hevc": true,
}

// ListDevices 回傳所有已連接的 iOS 裝置
func ListDevices() ([]DeviceInfo, error) {
	deviceList, err := ios.ListDevices()
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	var result []DeviceInfo
	for _, d := range deviceList.DeviceList {
		udid := d.Properties.SerialNumber
		info := DeviceInfo{UDID: udid}

		values, err := ios.GetValues(d)
		if err == nil {
			info.Name = values.Value.DeviceName
			info.Model = values.Value.ProductType
			info.IOSVersion = values.Value.ProductVersion
		}

		result = append(result, info)
	}
	return result, nil
}

// ConnectAFC 建立 AFC 連線（供 Engine 使用）
func ConnectAFC(udid string) (*afc.Client, error) {
	d, err := getIOSDevice(udid)
	if err != nil {
		return nil, err
	}
	client, err := afc.New(d)
	if err != nil {
		return nil, fmt.Errorf("afc connect: %w", err)
	}
	return client, nil
}

// IsSupportedExtension 檢查副檔名是否為備份支援格式（小寫，含點，如 ".heic"）
func IsSupportedExtension(ext string) bool {
	return supportedExtensions[ext]
}

// getIOSDevice 依 UDID 找出對應的 ios.DeviceEntry
func getIOSDevice(udid string) (ios.DeviceEntry, error) {
	deviceList, err := ios.ListDevices()
	if err != nil {
		return ios.DeviceEntry{}, fmt.Errorf("list devices: %w", err)
	}
	for _, d := range deviceList.DeviceList {
		if d.Properties.SerialNumber == udid {
			return d, nil
		}
	}
	return ios.DeviceEntry{}, fmt.Errorf("device not found: %s", udid)
}

// ScanDCIM 透過 AFC 列出裝置 /DCIM/ 下所有照片與影片
func ScanDCIM(udid string) ([]PhotoFile, error) {
	d, err := getIOSDevice(udid)
	if err != nil {
		return nil, err
	}

	afcClient, err := afc.New(d)
	if err != nil {
		return nil, fmt.Errorf("afc connect failed: %w", err)
	}
	defer afcClient.Close()

	// 列出 /DCIM 下的子目錄（100APPLE, 101APPLE, ...）
	dirs, err := afcClient.List("/DCIM")
	if err != nil {
		return nil, fmt.Errorf("list /DCIM: %w", err)
	}

	var photos []PhotoFile
	for _, dir := range dirs {
		dirPath := "/DCIM/" + dir

		dirInfo, err := afcClient.Stat(dirPath)
		if err != nil || dirInfo.Type != afc.S_IFDIR {
			continue
		}

		files, err := afcClient.List(dirPath)
		if err != nil {
			continue
		}

		for _, fileName := range files {
			ext := strings.ToLower(path.Ext(fileName))
			if !supportedExtensions[ext] {
				continue
			}

			remotePath := dirPath + "/" + fileName
			fileInfo, err := afcClient.Stat(remotePath)
			if err != nil {
				continue
			}

			photos = append(photos, PhotoFile{
				RemotePath: remotePath,
				FileName:   fileName,
				Size:       fileInfo.Size,
				ModTime:    0, // TODO: go-ios afc.FileInfo 無 mtime，待 Step 1.3 研究
			})
		}
	}
	return photos, nil
}

// GetDeviceDetail 取得裝置詳細資訊（含照片數、儲存空間）
func GetDeviceDetail(udid string) (*DeviceDetail, error) {
	d, err := getIOSDevice(udid)
	if err != nil {
		return nil, err
	}

	info := DeviceInfo{UDID: udid}
	values, err := ios.GetValues(d)
	if err == nil {
		info.Name = values.Value.DeviceName
		info.Model = values.Value.ProductType
		info.IOSVersion = values.Value.ProductVersion
	}

	afcClient, err := afc.New(d)
	if err != nil {
		return nil, fmt.Errorf("afc connect failed: %w", err)
	}
	defer afcClient.Close()

	detail := &DeviceDetail{DeviceInfo: info}

	// 取得裝置儲存空間
	deviceInfo, err := afcClient.DeviceInfo()
	if err == nil {
		detail.TotalSpace = int64(deviceInfo.TotalBytes)
		detail.UsedSpace = int64(deviceInfo.TotalBytes - deviceInfo.FreeBytes)
	}

	// 估算照片數（掃描 DCIM）
	photos, err := ScanDCIM(udid)
	if err == nil {
		detail.PhotoCount = int64(len(photos))
	}

	return detail, nil
}
