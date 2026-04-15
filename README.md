# iVault

> iPhone photo backup tool — plug in USB, back up all your photos in minutes. Free, open source, no iCloud needed.

[繁體中文](README.zh-TW.md) | **English**

![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Release](https://img.shields.io/badge/release-pre--release-orange)

<!-- TODO: add hero GIF (docs/screenshots/demo.gif) -->

## Features

- **Fast**: Direct transfer via Apple AFC protocol — no iTunes sync required
- **Free & open source**: No subscription, no privacy concerns, fully auditable code
- **Cross-platform**: Native macOS support; Windows with the free Apple Devices App
- **Simple**: First-time backup in minutes, no technical knowledge needed
- **Resume support**: Interrupted backups continue where they left off
- **Auto-organized**: Reads EXIF data and sorts photos into `YYYY-MM` folders by shoot date
- **HEIC conversion**: Optionally save a JPEG copy alongside the original HEIC
- **Bilingual UI**: Supports Traditional Chinese / English, preference saved locally

## Requirements

| Platform | Requirements |
|---|---|
| macOS | macOS 12 Monterey or later |
| Windows | Windows 10/11 + [Apple Devices App](https://apps.microsoft.com/store/detail/apple-devices/9NP83LWLPZ9N) (free) |
| iPhone | iOS 14 or later, USB cable |

## Installation

Download the latest release from [GitHub Releases](https://github.com/diablofong/iVault/releases):

- **Windows**: Download `iVault-Setup.exe` and run the installer
- **macOS**: Download `iVault.dmg` and drag to Applications

> No official release is available yet — stay tuned.

### First-launch Security Warnings

**Windows — SmartScreen warning**

On first run, Windows may show a "Windows protected your PC" dialog. This is because iVault has not yet obtained a commercial code-signing certificate.

Fix: click **More info** → **Run anyway**.

> iVault is fully open source — feel free to audit the code in this repo.

**macOS — Gatekeeper warning**

On first open, macOS may show "cannot be opened because the developer cannot be verified". This is because iVault has not yet completed Apple notarization.

Fix (either option):
- In Finder, **right-click** iVault.app → **Open** → click **Open** again
- Or: System Settings → Privacy & Security → find the iVault block → click **Open Anyway**

## Building from Source

### Prerequisites

- [Go 1.23+](https://golang.org/dl/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- **macOS**: Xcode Command Line Tools (`xcode-select --install`)
- **Windows**:
  - WebView2 (built into Windows 11; Windows 10 requires separate install)
  - C compiler: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or [MSYS2](https://www.msys2.org/) (required by Wails)

### Steps

```bash
git clone https://github.com/diablofong/iVault.git
cd iVault

# Install Go dependencies
go mod tidy

# Development mode (with hot-reload)
wails dev

# Production build
wails build
```

## Architecture

```
Go + Wails v2 (UI shell)
├── go-ios        → iPhone USB communication (AFC protocol)
├── goheif        → HEIC thumbnail processing
├── goexif        → EXIF shoot-date extraction for monthly sorting
└── Wails Events  → WebSocket real-time progress (server push)
```

## Comparison

| | iCloud Photos | iMazing | **iVault** |
|---|---|---|---|
| **Cost** | $0.99–$9.99/mo (subscription) | $29.99+/yr (subscription) | **Free** |
| **Photo storage** | Apple cloud | Local | **Local** |
| **Open source** | ✗ | ✗ | **✓** |
| **Windows** | Web only | ✓ | **✓** |
| **macOS** | ✓ (built-in) | ✓ | **✓** |
| **Ease of setup** | Easy | Medium | **Easy** |

## Sponsorship

iVault is free and open source. If it saved you an iMazing subscription, consider buying me a coffee ☕

[![Buy Me a Coffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-support-yellow)](https://buymeacoffee.com/ivault)

## Reporting Issues

Found a bug or have a feature request? Please open an issue on [GitHub Issues](https://github.com/diablofong/iVault/issues).

## License

[Apache License 2.0](LICENSE)
