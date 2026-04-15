# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Development (hot-reload, opens app window)
wails dev

# Production build
wails build

# Run tests
go test ./...

# Run a single package's tests
go test ./internal/backup/...
```

CI builds via GitHub Actions (`.github/workflows/release.yml`):
- `build-windows` on `windows-latest` — requires MSYS2 UCRT64 GCC for CGO
- `build-macos` on `macos-latest` — `darwin/universal` (amd64 + arm64)
- Triggered on push to `v1.0.0` branch or any `v*` tag

## Architecture

**Stack:** Go backend + Wails v2 (UI shell) + vanilla HTML/CSS/JS (no framework, go:embed)

```
main.go          — Wails bootstrap, window options (Mica on Win11, TitleBarHiddenInset on Mac)
app.go           — All Wails-exported backend methods; watchDevices() goroutine drives state
internal/
  device/        — go-ios wrapper: list devices, poll trust state via AFC ping
  backup/        — Core AFC copy loop, resume logic, EXIF date routing
    exif_date.go     ReadShootDate(): JPEG (goexif) + HEIC (byte scan) + video router
    video_date.go    ReadVideoShootDate(): QuickTime mvhd atom parser
  heic/          — HEIC→JPEG thumbnail (goheif); 4-worker pool on Windows, goroutines on Mac
  config/        — JSON persistence (~/.ivault/config.json)
  platform/      — OS detection; split across platform.go / platform_darwin.go / platform_windows.go
frontend/
  index.html     — All UI views in one file, toggled by JS state machine
  src/main.js    — State machine + all Wails event listeners
  src/style.css  — Design system (OKLCH, CSS vars, dark/light auto)
  src/i18n.js    — zh-TW / en runtime switch
```

## Critical Constraints

**AFC worker = 1.** usbmuxd serialises AFC sessions; multiple parallel workers produced no throughput gain and added instability. Do not change this.

**Platform split — always add darwin stubs.** `platform_windows.go` contains Windows-only functions (`WMIDetectIPhone`, `IsAMDSReady`, `EnsureAMDSRunning`, etc.). Any new Windows-only function must have a no-op stub in `platform_darwin.go` or the macOS build will fail at compile time — even if the call site is guarded by `if OS == "windows"`.

**Windows requires Apple Devices App** (MS Store). The `watchDevices()` goroutine checks `AppleDevicesInstalled` and emits `driver:required` before attempting AFC. On macOS this check is skipped entirely.

**`hiddenCmd()` on Windows.** All `exec.Command` calls in `platform_windows.go` must go through `hiddenCmd()` (sets `CREATE_NO_WINDOW`) to suppress the black console flash in the Wails GUI process.

## UI State Machine

```
IDLE → DEVICE_FOUND → TRUST_GUIDE → READY → BACKING_UP → DONE
                                                        ↘ ERROR
```

**IDLE has three variants** (same view, switched by JS):
- `idle-variant-first` — `history.length === 0`
- `idle-variant-returning` — `history.length > 0 && !lastInterrupted`
- `idle-variant-interrupted` — `lastInterrupted === true`

**Config fields that drive UI state:**
- `LastInterrupted bool` — selects IDLE variant
- `InterruptedDone / InterruptedTotal int` — shown in interrupted variant
- `FirstBackupDone bool` — controls DONE page first-time confetti
- `BackupRecord.PhotosCount / VideosCount` — shown in returning variant

## i18n

All user-visible strings live in `frontend/src/locales/zh-TW.js` and `en.js`. HTML elements use `data-i18n="key"` attributes; `i18n.js` does the swap at runtime. When adding new UI text, add the key to both locale files.
