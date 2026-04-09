# iVault

> iPhone 照片備份工具 — 插上 USB，3 分鐘備份所有照片。免費、開源、無需 iCloud。

![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Status](https://img.shields.io/badge/status-development-orange)

## 特色

- **快速**：AFC 協定傳輸，速度比 iTunes/MTP 快 3–5 倍
- **免費開源**：無訂閱、無隱私疑慮、程式碼完全透明
- **跨平台**：macOS 原生支援，Windows 需搭配免費的 Apple Devices App
- **簡單**：非技術用戶也能在 5 分鐘內完成首次備份
- **斷點續傳**：中途中斷可繼續，跳過已備份檔案
- **按日期整理**：自動讀取 EXIF，照片依拍攝月份分類至 `YYYY-MM` 資料夾
- **中英文切換**：介面支援繁體中文 / English，語言偏好本機儲存

## 系統需求

| 平台 | 需求 |
|---|---|
| macOS | macOS 12 Monterey 以上 |
| Windows | Windows 10/11 + [Apple Devices App](https://apps.microsoft.com/store/detail/apple-devices/9NP83LWLPZ9N)（免費）|
| iPhone | iOS 14 以上，USB 連接線 |

## 安裝

> 目前專案處於開發階段，尚無正式發布版本。

```bash
# 從 GitHub Releases 下載（即將推出）
# Windows：iVault-Setup.exe
# macOS：iVault.dmg
```

## 開發者建置

### 前置需求

- [Go 1.21+](https://golang.org/dl/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- macOS：Xcode Command Line Tools
- Windows：WebView2（Windows 11 內建）

### 建置步驟

```bash
git clone https://github.com/diablofong/iVault.git
cd iVault

# 安裝依賴
go mod tidy

# 開發模式
wails dev

# 正式建置
wails build
```

## 技術架構

```
Go + Wails v2（UI Shell）
├── go-ios        → iPhone USB 通訊（AFC 協定）
├── goheif        → HEIC 格式縮圖處理
├── goexif        → 讀取 EXIF 拍攝日期，按月份自動分類
└── Wails Events  → WebSocket 即時進度推送（後端 push）
```

## 競品比較

| | iMazing | FoneTool | iTunes | **iVault** |
|---|---|---|---|---|
| 價格 | $40/年 | 免費（有限制）| 免費 | **免費/開源** |
| 傳輸速度 | 慢（MTP）| 慢（MTP）| 慢 | **快（AFC）3–5×** |
| Linux 支援 | ✗ | ✗ | ✗ | ✓（規劃中）|
| 隱私 | 低疑慮 | 高（中國公司）| 低疑慮 | **無（開源）**|

## 路線圖

- [x] 專案初始化
- [x] M0：go-ios AFC 技術驗證（Windows + Mac）
- [x] M1：核心備份功能（AFC 複製、斷點續傳、即時進度）
- [x] M2：使用者體驗優化（信任引導、HEIC 轉檔、錯誤處理、Linear 設計、i18n、EXIF 日期分類）
- [ ] M3：打包發布（Windows .exe + Mac .dmg）

## 贊助

如果這個工具幫你省下了 iMazing 的訂閱費，歡迎請我喝杯咖啡 ☕

## 授權

[Apache License 2.0](LICENSE)
