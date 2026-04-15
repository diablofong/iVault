# iVault

> iPhone 照片備份工具 — 插上 USB，幾分鐘備份所有照片。免費、開源、無需 iCloud。

![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Release](https://img.shields.io/badge/release-pre--release-orange)

<!-- TODO: 加入 hero GIF（docs/screenshots/demo.gif） -->

## 特色

- **快速**：使用 Apple AFC 協定直接傳輸，無需透過 iTunes 同步
- **免費開源**：無訂閱、無隱私疑慮、程式碼完全透明
- **跨平台**：macOS 原生支援；Windows 搭配免費的 Apple Devices App
- **簡單**：非技術用戶也能在幾分鐘內完成首次備份
- **斷點續傳**：中途中斷可繼續，跳過已備份檔案
- **按日期整理**：自動讀取 EXIF，照片依拍攝月份分類至 `YYYY-MM` 資料夾
- **HEIC 轉檔**：可選在備份同時轉存 JPEG 副本
- **中英文介面**：支援繁體中文 / English，語言偏好本機儲存

## 系統需求

| 平台 | 需求 |
|---|---|
| macOS | macOS 12 Monterey 以上 |
| Windows | Windows 10/11 + [Apple Devices App](https://apps.microsoft.com/store/detail/apple-devices/9NP83LWLPZ9N)（免費）|
| iPhone | iOS 14 以上，USB 連接線 |

## 安裝

前往 [GitHub Releases](https://github.com/diablofong/iVault/releases) 下載最新版本：

- **Windows**：下載 `iVault-Setup.exe`，執行安裝精靈
- **macOS**：下載 `iVault.dmg`，拖曳至 Applications 資料夾

> 目前尚無正式發布版本，敬請期待。

### 首次啟動安全性警告

**Windows — SmartScreen 警告**

首次執行時，Windows 可能顯示「Windows 已保護您的電腦」對話框。這是因為 iVault 尚未取得商業程式碼簽署憑證。

解決方式：點擊「**更多資訊**」→「**仍要執行**」即可。

> iVault 為完全開源專案，程式碼可在此 repo 自行審閱驗證。

**macOS — Gatekeeper 警告**

首次開啟時，macOS 可能顯示「無法開啟，因為開發者無法驗證」。這是因為 iVault 尚未完成 Apple 公證流程。

解決方式（擇一）：
- 在 Finder 中**右鍵點擊** iVault.app → **開啟** → 再點一次「開啟」
- 或：系統設定 → 隱私與安全性 → 找到 iVault 的封鎖提示 → 點「**仍要開啟**」

## 開發者建置

### 前置需求

- [Go 1.23+](https://golang.org/dl/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- **macOS**：Xcode Command Line Tools（`xcode-select --install`）
- **Windows**：
  - WebView2（Windows 11 內建，Windows 10 需另行安裝）
  - C 編譯器：[TDM-GCC](https://jmeubank.github.io/tdm-gcc/) 或 [MSYS2](https://www.msys2.org/)（Wails 建置需要）

### 建置步驟

```bash
git clone https://github.com/diablofong/iVault.git
cd iVault

# 安裝 Go 依賴
go mod tidy

# 開發模式（含 hot-reload）
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

| | iCloud 照片 | iMazing | **iVault** |
|---|---|---|---|
| **費用** | $0.99–$9.99/月（訂閱）| $29.99+/年（訂閱）| **完全免費** |
| **照片存放位置** | Apple 雲端 | 本機 | **本機** |
| **開源** | ✗ | ✗ | **✓** |
| **Windows** | 僅網頁版 | ✓ | **✓** |
| **macOS** | ✓（內建）| ✓ | **✓** |
| **設定難度** | 低 | 中 | **低** |

## 回報問題

遇到 Bug 或有功能建議？請至 [GitHub Issues](https://github.com/diablofong/iVault/issues) 開票。

## 授權

[Apache License 2.0](LICENSE)
