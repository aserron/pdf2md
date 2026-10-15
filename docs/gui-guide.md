# Windows GUI Guide: `pdf2md-gui`

The `pdf2md-gui` desktop application provides an interactive, distraction-free graphical interface for PDF to Markdown conversions on Windows.

---

## 1. Launching the GUI

From PowerShell or Terminal:

```powershell
# Using Make
make run-gui

# Or directly with Go
go run ./cmd/pdf2md-gui
```

When launched, `pdf2md-gui` initializes an embedded local service and automatically opens a standalone Windows desktop app window.

---

## 2. Features & Workflow

### 1. Document Ingestion
- **Drag & Drop**: Drag any `.pdf` document directly into the designated drop zone.
- **File Picker**: Click on the drop zone to open the native Windows file selector dialog.

### 2. Conversion Configuration
- **Clean Watermarks & Margins**: Automatically removes recurring margin headers, footers, pagination indicators, and standard watermark phrases ("CONFIDENTIAL", "DRAFT", "SAMPLE").
- **Extract Embedded Images**: Exports all embedded raster images into an assets folder and generates corresponding Markdown image tags.
- **Detect & Format Tables**: Analyzes column horizontal alignments to format multi-column text into GitHub-flavored Markdown tables.
- **Preserve Styles**: Preserves font weights (bold), font slants (italic), and monospace code fragments.
- **Regex Scrub Filter**: Provide custom regular expressions (e.g. `\[UUID:[^\]]+\]`) to redact proprietary tokens or tracking IDs.

### 3. Real-Time Conversion & Live Preview
- **Interactive Markdown Viewer**: Instant split-pane view displaying generated Markdown text.
- **One-Click Copy**: Copies generated Markdown directly to the Windows clipboard.
- **Save File**: Exports the resulting `.md` file to disk.
- **Telemetry Card**: Displays execution duration in milliseconds, total pages processed, cleaned artifact count, and extracted image count.
