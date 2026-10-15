// Package pdf2md provides a high-performance PDF-to-Markdown conversion engine.
package pdf2md

import (
	"context"
	"io"
)

// Options defines configuration parameters for PDF to Markdown conversion.
type Options struct {
	// ExtractImages extracts embedded raster images to ImageDir.
	ExtractImages bool
	// ImageDir specifies where extracted images are stored.
	ImageDir string
	// ExtractDiagrams attempts vector bounding box extraction for charts.
	ExtractDiagrams bool
	// CleanWatermarks filters out repeating headers, footers, and watermark text.
	CleanWatermarks bool
	// PreserveStyles preserves font weight (bold), slant (italic), and monospace.
	PreserveStyles bool
	// FormatTables identifies tabular column alignments and generates markdown tables.
	FormatTables bool
	// StripRegex specifies an optional regex pattern to scrub matching text.
	StripRegex string
}

// DefaultOptions returns conservative, high-fidelity defaults.
func DefaultOptions() Options {
	return Options{
		ExtractImages:   true,
		ImageDir:        "assets",
		ExtractDiagrams: false,
		CleanWatermarks: true,
		PreserveStyles:  true,
		FormatTables:    true,
		StripRegex:      "",
	}
}

// Result holds the conversion artifacts and telemetry.
type Result struct {
	Markdown    string
	ImagePaths  []string
	PageCount   int
	CleanedRows int
}

// Converter defines the conversion engine contract.
type Converter interface {
	Convert(ctx context.Context, r io.ReaderAt, size int64, opts Options) (*Result, error)
	ConvertFile(ctx context.Context, filePath string, opts Options) (*Result, error)
}
