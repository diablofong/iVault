# iVault Build Makefile
# ─────────────────────────────────────────────────────────────
# make dev        — 開發模式（熱重載）
# make build-win  — 建置 Windows .exe（在 Windows 執行）
# make build-mac  — 建置 macOS .app（在 macOS 執行）
# make release    — 建置並打包 GitHub Releases 產出物
# make test       — 執行 Go 測試
# make clean      — 清除建置產出
# ─────────────────────────────────────────────────────────────

VERSION  := 1.0.0
APP_NAME := iVault

.PHONY: dev
dev:
	wails dev

.PHONY: test
test:
	go test ./...

# ── Windows（在 Windows 執行） ─────────────────────────────────

.PHONY: build-win
build-win:
	wails build --platform windows/amd64 --clean
	@echo ""
	@echo "✓ build/bin/$(APP_NAME).exe"

# ── macOS（在 macOS 執行） ────────────────────────────────────

.PHONY: build-mac
build-mac:
	wails build --platform darwin/universal --clean
	@echo ""
	@echo "✓ build/bin/$(APP_NAME).app"

# ── GitHub Releases 打包 ──────────────────────────────────────
# Windows：直接上傳 .exe
# macOS：壓縮 .app 後上傳（使用者解壓後右鍵→開啟即可）

.PHONY: release
release:
	@echo "請在對應平台上分別執行 build-win / build-mac，再用 GitHub UI 上傳："
	@echo "  Windows → build/bin/$(APP_NAME).exe"
	@echo "  macOS   → zip -r iVault-$(VERSION)-macos.zip build/bin/$(APP_NAME).app"

# ── 清除 ─────────────────────────────────────────────────────

.PHONY: clean
clean:
	rm -rf build/bin/
	@echo "✓ 清除完成"
