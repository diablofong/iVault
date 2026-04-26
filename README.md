# iVault

> Free iPhone photo backup for Windows — USB direct, open source, no iCloud.

[繁體中文](README.zh-TW.md) | **English**

![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Release](https://img.shields.io/badge/release-v1.0.0-brightgreen)

Built because Microsoft Photos keeps failing and iCloud runs out of space.
iVault transfers photos directly from your iPhone via USB — no iCloud, no iTunes sync, no subscription.

## Features

- **USB direct transfer** via Apple's AFC protocol — no cloud, no account required
- **Auto monthly folder sorting** by EXIF shoot date (e.g. `2024-07/`)
- **Resume interrupted backups** — picks up exactly where it left off
- **Auto-start at login** — plug in iPhone and backup begins automatically
- **Windows Toast notifications** — get notified when backup completes
- Free & open source — always

## Screenshots

<table>
  <tr>
    <td align="center"><img src="docs/screenshots/idle-first-en.png" width="100%"/><br/><sub>First launch</sub></td>
    <td align="center"><img src="docs/screenshots/idle-returning-en.png" width="100%"/><br/><sub>Returning user</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/screenshots/ready-en.png" width="100%"/><br/><sub>Device ready</sub></td>
    <td align="center"><img src="docs/screenshots/backing-up-en.png" width="100%"/><br/><sub>Backing up</sub></td>
  </tr>
</table>

**[→ Download & user guide (website)](https://diablofong.github.io/iVault)**

---

## System Requirements

| Requirement | Details |
|---|---|
| OS | Windows 10 / 11 (64-bit) |
| Driver | [Apple Devices App](https://apps.microsoft.com/detail/9NP83LWLPZ9K) — free, Microsoft Store |
| Runtime | WebView2 (built into Windows 11; [download](https://developer.microsoft.com/microsoft-edge/webview2/) for Windows 10) |

> **Note:** On first launch, iVault will guide you through installing Apple Devices if it's not already installed.

> **SmartScreen warning:** Because iVault is unsigned, Windows may show a blue SmartScreen dialog. Click "More info" → "Run anyway" — this is expected for open-source apps without code signing.

## Installation

1. Download `iVault-v1.0.0-windows-amd64.zip` from [Releases](https://github.com/diablofong/iVault/releases/latest)
2. Unzip and run `iVault.exe`
3. On first launch, follow the 3-step setup guide:
   - Install **Apple Devices** from Microsoft Store (if not already installed)
   - Choose a backup folder
   - Optionally enable auto-start at login

## Building from Source

### Prerequisites

- [Go 1.23+](https://golang.org/dl/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- C compiler: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MSYS2 UCRT64 GCC
- WebView2 (built into Windows 11)

### Steps

```bash
git clone https://github.com/diablofong/iVault.git
cd iVault

# Install Go dependencies
go mod tidy

# Development mode (hot-reload)
wails dev

# Production build
wails build -platform windows/amd64
```

## Architecture

```
Go + Wails v2 (UI shell)
├── go-ios        → iPhone USB communication (AFC protocol)
├── goheif        → HEIC thumbnail processing
├── goexif        → EXIF shoot-date extraction for monthly folder sorting
└── Wails Events  → WebSocket real-time progress (server push)
```

## FAQ

**Q: Do I need iTunes?**
A: No. iVault uses Apple Devices (a newer, lighter Apple app) instead of iTunes. iTunes can actually conflict — close it if it's running.

**Q: Where are my photos saved?**
A: To whatever folder you choose during setup. Default is the largest non-system drive. Files are organized into `YYYY-MM/` subfolders by shoot date.

**Q: Is iCloud required?**
A: No. Transfer is entirely local over USB. iCloud never comes into play.

**Q: What if I use iCloud Optimize Storage?**
A: If Optimize Storage is on, your iPhone only stores thumbnails locally. iVault will back up those thumbnails. To get full-resolution originals, disable Optimize Storage in iPhone Settings → Photos first.

**Q: What happens if I unplug the cable during backup?**
A: iVault saves your progress. Next time you connect, it resumes from where it left off.

**Q: Is my data safe?**
A: iVault does not collect, transmit, or store any personal data. All transfers happen locally over USB. See [PRIVACY.md](PRIVACY.md).

**Q: Does it work with iOS 18?**
A: Yes. iVault supports iOS 14 and later.

**Q: Why does Windows show a security warning?**
A: iVault doesn't have paid code signing. Click "More info" → "Run anyway" to proceed. The source code is fully open and auditable on GitHub.

## Reporting Issues

Found a bug or have a feature request? Open an issue on [GitHub Issues](https://github.com/diablofong/iVault/issues).

## Contributing

Pull requests are welcome. Please open an issue first to discuss significant changes.

## Privacy

iVault does not collect, transmit, or store any personal data. All transfers happen locally over USB. See [PRIVACY.md](PRIVACY.md) for details.

## License

[Apache License 2.0](LICENSE)
