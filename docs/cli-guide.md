# CLI Guide: `pdf2md`

The `pdf2md` command-line utility provides high-throughput conversion from PDF documents into standard Markdown.

---

## 1. Installation & Compilation

Ensure Go 1.22+ and Make are installed:

```bash
# Build binary
make build-cli

# Run directly
go run ./cmd/pdf2md --help
```

---

## 2. Command-Line Reference

```text
Usage: pdf2md -i <input.pdf|folder> [-o <output.md|dir>] [flags]
```

### Available Flags

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-i` | `string` | *required* | Path to input PDF file or directory. |
| `-o` | `string` | stdout | Destination Markdown file or output directory. |
| `--images` | `bool` | `true` | Extract embedded raster images into the assets directory. |
| `--image-dir` | `string` | `assets` | Directory where extracted images are saved. |
| `--diagrams` | `bool` | `false` | Detect and export diagram bounding regions. |
| `--clean` | `bool` | `true` | Strip watermarks (e.g. "CONFIDENTIAL", "DRAFT") and margin headers/footers. |
| `--styles` | `bool` | `true` | Preserve bold, italic, and monospace code typography. |
| `--tables` | `bool` | `true` | Automatically detect multi-column table data and format as GFM tables. |
| `--strip` | `string` | `""` | Optional regex pattern to scrub matching tokens from output. |
| `--inspect` | `bool` | `false` | Dump structural diagnostic information as formatted JSON without converting. |
| `--quiet` | `bool` | `false` | Suppress progress telemetry from stderr. |
| `--version` | `bool` | `false` | Print version and exit. |

---

## 3. Usage Examples

### Convert Single File to Markdown
```bash
pdf2md -i report.pdf -o report.md
```

### Stream Markdown to Standard Output
```bash
pdf2md -i report.pdf --quiet > report.md
```

### Extract Embedded Images to Custom Folder
```bash
pdf2md -i manual.pdf -o manual.md --images --image-dir ./figures/
```

### Strip Confidential Watermarks & Custom UUID Tokens
```bash
pdf2md -i draft.pdf -o clean.md --clean --strip "\[UUID:[a-f0-9\-]+\]"
```

### Batch Convert an Entire Directory
```bash
pdf2md -i ./input_pdfs/ -o ./converted_md/
```

### Inspect Document Structure & Fonts
```bash
pdf2md -i report.pdf --inspect
```

Output:
```json
{
  "file_path": "report.pdf",
  "pages": 14,
  "fonts_found": [
    "Helvetica-Bold",
    "Helvetica",
    "Courier"
  ],
  "has_images": true,
  "file_size_bytes": 1048576,
  "analysis_duration_ms": 35
}
```
