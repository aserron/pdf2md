# ADR 003: Standalone Windows App-Mode GUI Architecture

## Status
Accepted

## Context
A Windows graphical user interface was required alongside the CLI. Traditional options for desktop applications include:
1. **Electron**: Adds 150MB+ of Chromium/Node.js bundle overhead and high memory consumption.
2. **Fyne / Gio**: Pure Go GUI frameworks that require OpenGL, Cgo toolchains, and graphics drivers.
3. **Embedded Webview / Native App Mode**: Embedding static web assets into the Go binary (`embed.FS`), serving via a loopback port, and launching Microsoft Edge in standalone desktop application mode (`msedge.exe --app=http://127.0.0.1:<port>`).

## Decision
We adopted the **Embedded App-Mode Architecture** (`cmd/pdf2md-gui`):
- HTML5, CSS3, and JavaScript are bundled directly into the compiled executable via `embed.FS`.
- The application launches on an ephemeral loopback port (`127.0.0.1:0`).
- The native Windows window is instantiated using Microsoft Edge's built-in application mode (`--app=...`).

## Consequences

### Positive
- **Zero Overhead**: Binary size is under 25MB with zero external runtime dependencies.
- **Native Look & Feel**: Launches as an independent, frameless desktop application window with standard window borders and taskbar grouping (no browser navigation bar, tabs, or bookmarks).
- **Universal Availability**: Microsoft Edge is pre-installed on all Windows 10 and 11 distributions; falls back gracefully to standard browser if unavailable.
- **Full Platform Interactivity**: Supports drag-and-drop file ingestion, clipboard access, and local file downloads.

### Negative
- Requires a local network socket for IPC between the frontend UI and the Go engine backend.
