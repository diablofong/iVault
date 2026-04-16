# iVault

> Free iPhone photo backup for Windows — USB direct, open source, no iCloud.

[繁體中文](README.zh-TW.md) | **English**

![Platform](https://img.shields.io/badge/platform-Windows-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Release](https://img.shields.io/badge/release-pre--release-orange)

Built because Microsoft Photos keeps failing and iCloud runs out of space.
iVault transfers photos directly from your iPhone via USB — no iCloud, no iTunes sync, no subscription.

## Features

- USB direct transfer via AFC protocol — no cloud, no account required
- Auto monthly folder sorting by EXIF shoot date
- Resume interrupted backups
- Windows native UI (Wails + WebView2)
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

- Windows 10 / 11 (64-bit)
- [Apple Devices App](https://apps.microsoft.com/detail/9NP83LWLPZ9K) (free, Microsoft Store) — required for iPhone USB driver

## Building from Source

### Prerequisites

- [Go 1.23+](https://golang.org/dl/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- C compiler: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or [MSYS2](https://www.msys2.org/)
- WebView2 (built into Windows 11; Windows 10 requires separate install)

### Steps

```bash
git clone https://github.com/diablofong/iVault.git
cd iVault

# Install Go dependencies
go mod tidy

# Development mode (with hot-reload)
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

## Reporting Issues

Found a bug or have a feature request? Please open an issue on [GitHub Issues](https://github.com/diablofong/iVault/issues).

## Contributing

Pull requests are welcome. Please open an issue first to discuss significant changes.

## Privacy

iVault does not collect, transmit, or store any personal data. All transfers happen locally over USB. See [PRIVACY.md](PRIVACY.md) for details.

## License

[Apache License 2.0](LICENSE)
