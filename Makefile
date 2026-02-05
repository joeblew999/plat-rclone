.PHONY: all clean templ macos ios android windows web run-macos ios-sim run-web dev dev-embedded dev-embedded-full download datastar icons tools test fmt help screenshot install-goup

# Versions
DATASTAR_VERSION := v1.0.0-RC.7
GOUP := goup-util
GOUP_VERSION := v2.0.0

# === Build ===

all: templ macos ios android windows web

templ:
	@templ generate

# Native apps (Gio + webview) via goup-util
macos: templ
	@$(GOUP) build macos .

ios: templ
	@$(GOUP) build ios .

android: templ
	@$(GOUP) build android .

windows: templ
	@$(GOUP) build windows .

# Web server binary
web: templ
	@go build -o .bin/plat-rclone-web ./cmd/plat-rclone

web-full: templ  # Web server with all rclone backends
	@go build -tags=rclone_full -o .bin/plat-rclone-web ./cmd/plat-rclone

# === Run ===

run-macos: macos
	@open .bin/*.app

run-web: web
	@.bin/plat-rclone-web serve

ios-sim: ios
	@xcrun simctl install booted .bin/*.app
	@xcrun simctl launch booted com.github.plat_rclone

# Development modes
dev: templ  # HTTP mode - requires: rclone rcd --rc-no-auth
	@go run ./cmd/plat-rclone serve

dev-embedded: templ  # Embedded rclone - local backend only (fast build)
	@go run ./cmd/plat-rclone serve -embedded

dev-embedded-full: templ  # Embedded rclone - all backends (slow build, large binary)
	@go run -tags=rclone_full ./cmd/plat-rclone serve -embedded

# === Utils ===

download:
	@go run ./cmd/plat-rclone download .bin

datastar:
	@mkdir -p static/js
	@echo "Downloading Datastar $(DATASTAR_VERSION)..."
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar.js" -o static/js/datastar.js
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar.js.map" -o static/js/datastar.js.map
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar-core.js" -o static/js/datastar-core.js
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar-core.js.map" -o static/js/datastar-core.js.map
	@echo "Done: static/js/datastar*.js"

icons:
	@$(GOUP) icons .

screenshot:
	@xcrun simctl io booted screenshot docs/ios-screenshot.png
	@echo "Saved: docs/ios-screenshot.png"

clean:
	@rm -rf .bin .build
	@echo "Cleaned"

# === Setup ===

tools:
	@go install github.com/a-h/templ/cmd/templ@latest

install-goup:
	@echo "Installing goup-util $(GOUP_VERSION)..."
	@curl -fsSL "https://github.com/joeblew999/goup-util/releases/download/$(GOUP_VERSION)/goup-util_$(shell uname -s)_$(shell uname -m).tar.gz" | tar -xz -C /usr/local/bin goup-util
	@echo "Installed: /usr/local/bin/goup-util"
	@goup-util --help | head -1

test:
	@go test ./...

fmt:
	@go fmt ./...
	@templ fmt templates/

# === Help ===

help:
	@echo "plat-rclone - Cross-platform GUI for rclone"
	@echo ""
	@echo "Development:"
	@echo "  make dev               HTTP mode (needs: rclone rcd --rc-no-auth)"
	@echo "  make dev-embedded      Embedded mode, local backend only (fast)"
	@echo "  make dev-embedded-full Embedded mode, all backends (slow, ~100MB)"
	@echo ""
	@echo "Build:                   Run:"
	@echo "  make all               make run-macos    Run macOS app"
	@echo "  make macos             make run-web      Web server :8080"
	@echo "  make ios               make ios-sim      iOS simulator"
	@echo "  make android"
	@echo "  make windows"
	@echo "  make web               make web-full     With all backends"
	@echo ""
	@echo "Setup:                   Utils:"
	@echo "  make tools             make download     Get rclone binary"
	@echo "  make install-goup      make datastar     Update JS bundles"
	@echo "  make clean             make icons        Generate app icons"
	@echo ""
	@echo "Versions: goup-util $(GOUP_VERSION) | Datastar $(DATASTAR_VERSION)"
