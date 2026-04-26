# iVault

> Windows 的免費 iPhone 照片備份工具 — USB 直傳、開源、不碰 iCloud。

**繁體中文** | [English](README.md)

![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Release](https://img.shields.io/badge/release-v1.0.0-brightgreen)

因為 Microsoft Photos 老是失敗，iCloud 5GB 又不夠用。
iVault 透過 USB 直接從 iPhone 傳輸照片，無需 iCloud、無需 iTunes 同步、無訂閱費。

## 功能特色

- **USB 直接傳輸** — 透過 Apple AFC 協定，不需雲端、不需帳號
- **依拍攝日期自動按月份分類** — 例如 `2024-07/`
- **支援中斷後繼續備份** — 從中斷點無縫繼續
- **開機自動啟動** — 插上 iPhone 自動開始備份，免手動開啟
- **備份完成 Toast 通知** — 備份結束時系統推播通知
- 永久免費、完全開源

## 應用程式截圖

<table>
  <tr>
    <td align="center"><img src="docs/screenshots/idle-first-zh.png" width="100%"/><br/><sub>首次啟動</sub></td>
    <td align="center"><img src="docs/screenshots/idle-returning-zh.png" width="100%"/><br/><sub>再次使用</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/screenshots/ready-zh.png" width="100%"/><br/><sub>裝置就緒</sub></td>
    <td align="center"><img src="docs/screenshots/backing-up-zh.png" width="100%"/><br/><sub>備份中</sub></td>
  </tr>
</table>

**[→ 下載與使用說明（官網）](https://diablofong.github.io/iVault)**

---

## 系統需求

| 需求 | 說明 |
|---|---|
| 作業系統 | Windows 10 / 11（64 位元） |
| 驅動程式 | [Apple Devices App](https://apps.microsoft.com/detail/9NP83LWLPZ9K)（免費，Microsoft Store） |
| 執行環境 | WebView2（Windows 11 內建；Windows 10 需[另行安裝](https://developer.microsoft.com/microsoft-edge/webview2/)） |

> **備注：** 首次啟動時，iVault 會引導你完成 Apple Devices 安裝（若尚未安裝）。

> **SmartScreen 警告：** 由於 iVault 尚未申請付費程式碼簽章，Windows 可能顯示藍色安全提示。點「更多資訊」→「仍要執行」即可。這是開源軟體的正常現象。

## 安裝方式

1. 從 [Releases](https://github.com/diablofong/iVault/releases/latest) 下載 `iVault-v1.0.0-windows-amd64.zip`
2. 解壓縮後執行 `iVault.exe`
3. 首次啟動時，依照三步引導完成設定：
   - 安裝 **Apple Devices**（Microsoft Store，若尚未安裝）
   - 選擇備份目標資料夾
   - 選擇是否開機自動啟動

## 開發者建置

### 前置需求

- [Go 1.23+](https://golang.org/dl/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- C 編譯器：[TDM-GCC](https://jmeubank.github.io/tdm-gcc/) 或 MSYS2 UCRT64 GCC
- WebView2（Windows 11 內建）

### 建置步驟

```bash
git clone https://github.com/diablofong/iVault.git
cd iVault

# 安裝 Go 依賴
go mod tidy

# 開發模式（含 hot-reload）
wails dev

# 正式建置
wails build -platform windows/amd64
```

## 技術架構

```
Go + Wails v2（UI Shell）
├── go-ios        → iPhone USB 通訊（AFC 協定）
├── goheif        → HEIC 格式縮圖處理
├── goexif        → 讀取 EXIF 拍攝日期，按月份自動分類
└── Wails Events  → WebSocket 即時進度推送（後端 push）
```

## 常見問題

**Q：需要安裝 iTunes 嗎？**
A：不需要。iVault 使用 Apple Devices（Apple 的新版驅動 App）而非 iTunes。iTunes 反而可能造成衝突，若有開啟請先關閉。

**Q：照片存到哪裡？**
A：存到你設定的資料夾（預設為最大的非系統磁碟），並按拍攝月份自動整理成 `YYYY-MM/` 子資料夾。

**Q：需要 iCloud 嗎？**
A：完全不需要。備份完全透過 USB 在本機進行，iCloud 不參與。

**Q：如果我開了 iCloud 最佳化儲存怎麼辦？**
A：開啟最佳化儲存後，iPhone 本機只保留縮圖，iVault 會備份這些縮圖。若需要完整原始檔，請先到 iPhone「設定 → 照片」關閉最佳化儲存後再備份。

**Q：備份中途拔掉 USB 線會怎樣？**
A：iVault 會記錄進度，下次插上後從中斷點繼續，不會重複備份已完成的檔案。

**Q：我的資料安全嗎？**
A：iVault 不收集、不傳送、不儲存任何個人資料。所有傳輸均在本機 USB 進行。詳見 [PRIVACY.md](PRIVACY.md)。

**Q：支援 iOS 18 嗎？**
A：支援。iVault 支援 iOS 14 及以上版本。

**Q：為什麼 Windows 顯示安全警告？**
A：iVault 沒有付費程式碼簽章。點「更多資訊」→「仍要執行」即可。原始碼完全開放，可在 GitHub 上查閱。

## 回報問題

遇到 Bug 或有功能建議？請至 [GitHub Issues](https://github.com/diablofong/iVault/issues) 開票。

## 貢獻

歡迎提交 Pull Request。重大變更請先開 Issue 討論。

## 隱私

iVault 不收集、不傳送、不儲存任何個人資料。所有傳輸均在本機 USB 進行。詳見 [PRIVACY.md](PRIVACY.md)。

## 授權

[Apache License 2.0](LICENSE)
