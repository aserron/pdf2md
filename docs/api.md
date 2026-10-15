# Programmatic Go API Reference: `pkg/pdf2md`

The `pkg/pdf2md` package enables programmatic integration of the PDF to Markdown conversion engine directly into Go applications, microservices, and ETL pipelines.

---

## 1. Installation

```bash
go get github.com/aserron/pdf2md@latest
```

---

## 2. API Contract

### Engine Instantiation
```go
import "github.com/aserron/pdf2md/pkg/pdf2md"

engine := pdf2md.NewEngine()
```

### Options Configuration
```go
opts := pdf2md.DefaultOptions()
opts.CleanWatermarks = true            // Strip watermarks and margin headers/footers
opts.ExtractImages   = true            // Extract embedded images
opts.ImageDir        = "./output/imgs" // Target directory for extracted assets
opts.FormatTables    = true            // Detect and emit GitHub-flavored Markdown tables
opts.PreserveStyles  = true            // Preserve bold, italic, and monospace code
opts.StripRegex      = `\[ID:\d+\]`   // Optional regex pattern to redact
```

### Converting a File on Disk
```go
ctx := context.Background()
result, err := engine.ConvertFile(ctx, "path/to/document.pdf", opts)
if err != nil {
    log.Fatalf("conversion failed: %v", err)
}

fmt.Printf("Generated Markdown (%d bytes):\n%s\n", len(result.Markdown), result.Markdown)
fmt.Printf("Pages processed: %d\n", result.PageCount)
fmt.Printf("Artifacts cleaned: %d\n", result.CleanedRows)
fmt.Printf("Extracted image paths: %v\n", result.ImagePaths)
```

### Converting an In-Memory Stream (`io.ReaderAt`)
```go
var reader io.ReaderAt = getPDFStream()
var size int64 = 1048576

result, err := engine.Convert(ctx, reader, size, opts)
if err != nil {
    log.Fatalf("stream conversion failed: %v", err)
}
```

---

## 3. Data Structures

### `Options`
```go
type Options struct {
    ExtractImages   bool   // Extract embedded raster images
    ImageDir        string // Target directory for images
    ExtractDiagrams bool   // Attempt diagram bounding region extraction
    CleanWatermarks bool   // Strip watermarks and margin pagination
    PreserveStyles  bool   // Preserve bold, italic, and monospace styles
    FormatTables    bool   // Format detected tabular data into Markdown tables
    StripRegex      string // Custom regex pattern to redact
}
```

### `Result`
```go
type Result struct {
    Markdown    string   // Generated Markdown document
    ImagePaths  []string // Paths to exported images
    PageCount   int      // Total pages processed
    CleanedRows int      // Number of watermark/header lines removed
}
```
