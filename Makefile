.PHONY: all clean templ macos ios android windows web run-macos ios-sim run-web dev download icons tools test fmt help

# goup-util for cross-platform builds
GOUP := goup-util

# Build all platforms
all: templ macos ios android windows web

# Generate templ files
templ:
	@templ generate

# === Platform Builds ===

macos: templ
	@$(GOUP) build macos .

ios: templ
	@$(GOUP) build ios .

android: templ
	@$(GOUP) build android .

windows: templ
	@$(GOUP) build windows .

# === Run ===

run-macos: macos
	@open .bin/*.app

ios-sim: ios
	@xcrun simctl install booted .bin/*.app
	@xcrun simctl launch booted com.github.plat_rclone

# === Web Server ===

web: templ
	@go build -o .bin/plat-rclone-web ./cmd/plat-rclone

run-web: web
	@.bin/plat-rclone-web serve

dev: templ
	@go run ./cmd/plat-rclone serve

# === Utils ===

download:
	@go run ./cmd/plat-rclone download .bin

icons:
	@$(GOUP) icons .

clean:
	@rm -rf .bin .build
	@echo "Cleaned"

tools:
	@go install github.com/a-h/templ/cmd/templ@latest

test:
	@go test ./...

fmt:
	@go fmt ./...
	@templ fmt templates/

# === Help ===

help:
	@echo "plat-rclone - Cross-platform GUI for rclone"
	@echo ""
	@echo "Build:"
	@echo "  make all        Build all platforms"
	@echo "  make macos      macOS app"
	@echo "  make ios        iOS app"
	@echo "  make android    Android APK"
	@echo "  make windows    Windows exe"
	@echo "  make web        Web server"
	@echo ""
	@echo "Run:"
	@echo "  make run-macos  Run macOS"
	@echo "  make ios-sim    iOS simulator"
	@echo "  make run-web    Web :8080"
	@echo "  make dev        Dev mode"
	@echo ""
	@echo "Utils:"
	@echo "  make download   Get rclone"
	@echo "  make clean      Clean"
