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
- Triggered on push to `v1.0.0` branch or any `v*` tag
- macOS removed (Windows-only since v1.0, 2026-04-16)

## Architecture

**Stack:** Go backend + Wails v2 (UI shell) + vanilla HTML/CSS/JS (no framework, go:embed)

```
main.go          — Wails bootstrap, window options (Mica on Win11)
app.go           — All Wails-exported backend methods; watchDevices() goroutine drives state
internal/
  device/        — go-ios wrapper: list devices, poll trust state via AFC ping
  backup/        — Core AFC copy loop, resume logic, EXIF date routing
    exif_date.go     ReadShootDate(): JPEG (goexif) + HEIC (byte scan) + video router
    video_date.go    ReadVideoShootDate(): QuickTime mvhd atom parser
  heic/          — HEIC→JPEG thumbnail (goheif); 4-worker pool on Windows
  config/        — JSON persistence (~/.ivault/config.json)
  platform/      — Windows-only: platform.go + platform_windows.go (darwin removed)
frontend/
  index.html     — All UI views in one file, toggled by JS state machine
  src/main.js    — State machine + all Wails event listeners
  src/style.css  — Design system (OKLCH, CSS vars, dark/light auto)
  src/i18n.js    — zh-TW / en runtime switch
```

## Critical Constraints

**AFC worker = 1.** usbmuxd serialises AFC sessions; multiple parallel workers produced no throughput gain and added instability. Do not change this.

**Windows-only codebase.** `platform_darwin.go` has been removed. `GOOS=darwin go build` is expected to fail. All platform functions live in `platform_windows.go`.

**Windows requires Apple Devices App** (MS Store). The `watchDevices()` goroutine checks `AppleDevicesInstalled` and emits `driver:required` before attempting AFC.

**`hiddenCmd()` on Windows.** All `exec.Command` calls in `platform_windows.go` must go through `hiddenCmd()` (sets `CREATE_NO_WINDOW`) to suppress the black console flash in the Wails GUI process.

## UI State Machine

```
IDLE → DEVICE_FOUND → TRUST_GUIDE → READY → BACKING_UP → DONE
                                                        ↘ ERROR
```

**IDLE has three variants** (same view, switched by JS):
- `idle-variant-first` — `devices[udid]` is nil / no backup record
- `idle-variant-returning` — `devices[udid].lastInterrupted === false`
- `idle-variant-interrupted` — `devices[udid].lastInterrupted === true`

**Config architecture (per-device, v1.0.0+):**
- `AppConfig.DefaultBackupPath` — global default backup path
- `AppConfig.Devices map[UDID]*DeviceConfig` — per-device state
- `DeviceConfig.FolderName` — immutable folder name, set at first connection: `"{Name} [{UDID[:8]}]"`
- `DeviceConfig.Name` — editable display name shown in UI
- `DeviceConfig.LastInterrupted bool` — selects IDLE variant
- `DeviceConfig.InterruptedDone / InterruptedTotal int` — shown in interrupted variant
- `DeviceConfig.FirstBackupDone bool` — controls DONE page first-time confetti
- `DeviceConfig.PhotosCount / VideosCount` — shown in returning variant (delta from last session)

**Legacy fields (deprecated, kept for migration only):**
- `History []BackupRecord`, `LastInterrupted bool`, `FirstBackupDone bool`, `FirstBackupDoneDevices []string`

## i18n

All user-visible strings live in `frontend/src/locales/zh-TW.js` and `en.js`. HTML elements use `data-i18n="key"` attributes; `i18n.js` does the swap at runtime. When adding new UI text, add the key to both locale files.
