.PHONY: all clean templ macos ios android windows web run-macos ios-sim run-web dev download datastar icons tools test fmt help screenshot install-goup

# Datastar JS version (must match Go SDK)
DATASTAR_VERSION := v1.0.0-RC.7

# goup-util for cross-platform builds
GOUP := goup-util
GOUP_VERSION := v2.0.0

# Build all platforms
all: templ macos ios android windows web

# Generate templ files
templ:
	@templ generate

# === Platform Builds ===
# Uses goup-util to build native Gio apps with embedded webview
# Output: .bin/plat-rclone.{app,apk,exe}

macos: templ  # Build macOS .app bundle
	@$(GOUP) build macos .

ios: templ  # Build iOS .app for device/simulator
	@$(GOUP) build ios .

android: templ  # Build Android .apk
	@$(GOUP) build android .

windows: templ  # Build Windows .exe with embedded icon
	@$(GOUP) build windows .

# === Run ===

run-macos: macos
	@open .bin/*.app

ios-sim: ios
	@xcrun simctl install booted .bin/*.app
	@xcrun simctl launch booted com.github.plat_rclone

# === Web Server ===
# Headless mode - serve via HTTP for browsers (no native UI)

web: templ  # Build standalone web server binary
	@go build -o .bin/plat-rclone-web ./cmd/plat-rclone

run-web: web  # Run web server on http://localhost:8080
	@.bin/plat-rclone-web serve

dev: templ  # Development mode with hot reload (go run)
	@go run ./cmd/plat-rclone serve

# === Utils ===

download:  # Download rclone binary for current platform
	@go run ./cmd/plat-rclone download .bin

datastar:  # Download Datastar JS bundles (must match Go SDK version)
	@mkdir -p static/js
	@echo "Downloading Datastar $(DATASTAR_VERSION) bundles..."
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar.js" -o static/js/datastar.js
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar.js.map" -o static/js/datastar.js.map
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar-core.js" -o static/js/datastar-core.js
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar-core.js.map" -o static/js/datastar-core.js.map
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar-aliased.js" -o static/js/datastar-aliased.js
	@curl -sL "https://raw.githubusercontent.com/starfederation/datastar/$(DATASTAR_VERSION)/bundles/datastar-aliased.js.map" -o static/js/datastar-aliased.js.map
	@echo "Downloaded all bundles:"
	@ls -lh static/js/datastar*.js

icons:  # Generate app icons from icon-source.png
	@$(GOUP) icons .

screenshot:  # Take iOS simulator screenshot
	@xcrun simctl io booted screenshot docs/ios-screenshot.png
	@echo "Screenshot saved to docs/ios-screenshot.png"

clean:  # Remove build outputs
	@rm -rf .bin .build
	@echo "Cleaned"

tools:  # Install required Go tools
	@go install github.com/a-h/templ/cmd/templ@latest

install-goup:  # Install goup-util $(GOUP_VERSION) for cross-platform builds
	@echo "Installing goup-util $(GOUP_VERSION)..."
	@curl -fsSL "https://github.com/joeblew999/goup-util/releases/download/$(GOUP_VERSION)/goup-util_$(shell uname -s)_$(shell uname -m).tar.gz" | tar -xz -C /usr/local/bin goup-util
	@echo "Installed goup-util $(GOUP_VERSION) to /usr/local/bin"
	@goup-util --help | head -1

test:  # Run all tests
	@go test ./...

fmt:  # Format Go and templ files
	@go fmt ./...
	@templ fmt templates/

# === Help ===

help:  # Show this help
	@echo "plat-rclone - Cross-platform GUI for rclone"
	@echo ""
	@echo "Build:                              Run:"
	@echo "  make all       All platforms        make run-macos  Run macOS app"
	@echo "  make macos     macOS .app           make run-web    Web :8080"
	@echo "  make ios       iOS .app             make dev        Dev mode"
	@echo "  make android   Android .apk         make ios-sim    iOS simulator"
	@echo "  make windows   Windows .exe"
	@echo "  make web       Web server"
	@echo ""
	@echo "Utils:                              Setup:"
	@echo "  make download  Get rclone binary    make install-goup  Install goup-util"
	@echo "  make datastar  Update Datastar JS   make tools         Install templ"
	@echo "  make icons     Generate app icons   make clean         Remove build files"
	@echo ""
	@echo "Versions: goup-util $(GOUP_VERSION) | Datastar JS $(DATASTAR_VERSION)"
