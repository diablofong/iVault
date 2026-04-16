# iVault 端對端測試清單 v1.0.0

## 發布方式

| 平台 | 檔案 | 說明 |
|---|---|---|
| Windows | `iVault.exe` | 直接執行，SmartScreen 警告點「仍要執行」 |
| macOS | `iVault.app.zip` | 解壓後右鍵→開啟（第一次需確認） |

## 測試環境

| 項目 | 規格 |
|---|---|
| 測試裝置 | iPhone（iOS 14 / 16 / 17 各一台） |
| Windows 測試機 | Windows 11 22H2 + Windows 11 23H2 |
| macOS 測試機 | macOS Monterey 12 + macOS Sonoma 14 |
| USB 線 | Lightning 和 USB-C 各一條 |

---

## 首次啟動

- [ ] Windows：雙擊 .exe，SmartScreen 警告可略過（「更多資訊」→「仍要執行」）
- [ ] macOS：右鍵 .app → 開啟 → 確認，後續直接雙擊即可
- [ ] App 視窗正常彈出（Windows：Mica 材質；macOS：Vibrancy 毛玻璃）
- [ ] IDLE 畫面顯示正確（iVault logo、隱私聲明）
- [ ] 深色/淺色模式跟隨系統
- [ ] 語言切換（中/EN）正常，localStorage 持久化

---

## Windows — Apple Devices 安裝引導

- [ ] 未安裝 Apple Devices 時啟動 → 顯示 DRIVER_MISSING 頁
- [ ] WMI 偵測到 iPhone 時顯示裝置名稱
- [ ] 點「一鍵安裝」→ winget 啟動（有 winget 時）
- [ ] 無 winget 時 → 開啟 Microsoft Store
- [ ] 安裝進度顯示三個階段（下載中 / 安裝驅動 / 啟動服務）
- [ ] 安裝完成後顯示「安裝完成！請重新插拔 iPhone」
- [ ] FAQ 展開/收合正常

---

## 裝置連線流程

- [ ] 插入已配對 iPhone → 自動進入 READY
- [ ] 插入未配對 iPhone → 顯示 TRUST_GUIDE
- [ ] 在 iPhone 點「信任」→ 自動進入 READY
- [ ] READY 頁顯示裝置名稱、照片數、磁碟空間、預估大小
- [ ] 上次備份資訊（初次顯示「尚未備份過」）
- [ ] 拔除 iPhone → 回到 IDLE
- [ ] 備份中拔除 iPhone → ERROR（1.5s 後自動回 IDLE）
- [ ] App 啟動時 iPhone 已連線 → 正常偵測

---

## 備份功能

- [ ] 選擇備份路徑（原生 Folder Picker）
- [ ] 進度條、速度、ETA、目前檔名顯示正確
- [ ] 完成後 DONE 頁統計數字正確
- [ ] 「開啟備份資料夾」按鈕正常
- [ ] 照片按 `YYYY-MM/` 分類
- [ ] 中途取消 → 重新備份時已完成的被跳過

---

## HEIC 轉檔

- [ ] 勾選「備份後同時轉存 JPEG 副本」後完成備份
- [ ] DONE 頁顯示 HEIC 轉檔進度條和完成張數
- [ ] 轉換後 JPEG 可正常開啟，方向正確
- [ ] Windows 有 HEIC 照片時顯示「安裝 HEIF 擴充」提示

---

## 錯誤處理

- [ ] 備份路徑無寫入權限 → 顯示錯誤訊息
- [ ] 磁碟空間不足 → 顯示錯誤 + 重試按鈕
- [ ] UNKNOWN_ERROR → 顯示「回報問題 →」連結

---

## 非技術用戶測試（5 人）

給安裝檔 + USB 線 + iPhone，不提供任何說明，計時能否在 5 分鐘內完成備份：

| 測試者 | 完成時間 | 卡住的步驟 |
|---|---|---|
| 1 | | |
| 2 | | |
| 3 | | |
| 4 | | |
| 5 | | |

---

## 版本資訊

| 項目 | 值 |
|---|---|
| 版本 | 1.0.0 |
| 測試日期 | |
| Windows .exe SHA256 | |
| macOS .app.zip SHA256 | |
