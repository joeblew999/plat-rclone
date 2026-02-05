# plat-rclone

Cross-platform GUI for [rclone](https://rclone.org) using Datastar + Gio.

## Features

- **Cross-platform** - macOS, iOS, Android, Windows
- **Native webview** - Single Go codebase via Gio + webviewer
- **Real-time UI** - Datastar SSE (~11KB)
- **Web server** - Headless mode for browsers

## Quick Start

```bash
make download                      # Get rclone
./.bin/rclone rcd --rc-no-auth     # Start rclone (other terminal)
make run-macos                     # Run app
```

## Build

Uses [goup-util](https://github.com/joeblew999/goup-util) for cross-platform builds.

```bash
make all        # Build all platforms
make macos      # macOS app       -> .bin/plat-rclone.app
make ios        # iOS app         -> .bin/plat-rclone.app
make android    # Android APK     -> .bin/plat-rclone.apk
make windows    # Windows exe     -> .bin/plat-rclone.exe
make web        # Web server      -> .bin/plat-rclone-web
```

## Run

```bash
make run-macos  # Run macOS app
make ios-sim    # Install to iOS simulator
make run-web    # Web server http://localhost:8080
make dev        # Dev mode
```

## Project Structure

```
plat-rclone/
├── main.go             # Gio + webviewer app
├── cmd/plat-rclone/    # Web server CLI
├── pkg/
│   ├── datastar/       # SSE helpers
│   ├── rclone/         # RC API client
│   └── router/         # Chi + Datastar
├── templates/          # templ templates
├── static/             # CSS, JS (embedded)
├── icon-source.png     # App icon
├── Makefile
├── .bin/               # Build output
└── .build/             # Intermediate
```

## Datastar Versions

The Go SDK and JS library must be compatible:

| Go SDK (datastar-go) | JS Library | Status |
|---------------------|------------|--------|
| v1.1.0 | v1.0.0-RC.7 | Current |

Update JS library:
```bash
make datastar    # Downloads all bundles to static/js/
```

Available bundles:
- `datastar.js` - Full bundle (30KB)
- `datastar-core.js` - Core only (9KB)
- `datastar-aliased.js` - Aliased exports

## AI Debugging

Since plat-rclone uses standard web technologies (HTML, CSS, JavaScript), AI tools can easily debug the UI using browser automation like Playwright:

```bash
make dev                         # Start dev server
# AI can use Playwright MCP to:
# - Take snapshots of the DOM
# - Click buttons and test interactions
# - Inspect network requests (SSE streams)
# - Verify Datastar reactivity
```

## Requirements

- Go 1.24+
- [goup-util](https://github.com/joeblew999/goup-util)
- [templ](https://github.com/a-h/templ)
- Xcode (iOS)
- Android NDK (Android)

## References

- [Datastar](https://data-star.dev/)
- [rclone RC API](https://rclone.org/rc/)
- [Gio](https://gioui.org/)
- [goup-util](https://github.com/joeblew999/goup-util)
