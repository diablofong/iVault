# iVault

> 透過 USB 備份 iPhone 照片 — 快速、免費、開源。

**繁體中文** | [English](README.md)

![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Release](https://img.shields.io/badge/release-pre--release-orange)

iVault 透過 Apple AFC 協定直接從 iPhone 傳輸照片，無需 iTunes、無需 iCloud、無訂閱費。

<!-- TODO: 加入 hero GIF（docs/screenshots/demo.gif） -->

**[→ 下載與使用說明（官網）](https://diablofong.github.io/iVault)**

---

## 開發者建置

### 前置需求

- [Go 1.23+](https://golang.org/dl/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- **macOS**：Xcode Command Line Tools（`xcode-select --install`）
- **Windows**：
  - WebView2（Windows 11 內建，Windows 10 需另行安裝）
  - C 編譯器：[TDM-GCC](https://jmeubank.github.io/tdm-gcc/) 或 [MSYS2](https://www.msys2.org/)

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

## 回報問題

遇到 Bug 或有功能建議？請至 [GitHub Issues](https://github.com/diablofong/iVault/issues) 開票。

## 貢獻

歡迎提交 Pull Request。重大變更請先開 Issue 討論。

## 授權

[Apache License 2.0](LICENSE)
