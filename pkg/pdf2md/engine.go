package pdf2md

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserron/pdf2md/internal/cleaner"
	"github.com/aserron/pdf2md/internal/extractor"
	"github.com/aserron/pdf2md/internal/parser"
	"github.com/aserron/pdf2md/internal/render"
	"github.com/ledongthuc/pdf"
)

// Engine is the default implementation of Converter.
type Engine struct {
	parser    *parser.PDFParser
	emitter   *render.MarkdownEmitter
	extractor *extractor.AssetExtractor
}

// NewEngine creates an initialized Engine ready for PDF conversions.
func NewEngine() *Engine {
	return &Engine{
		parser:    parser.NewPDFParser(),
		emitter:   render.NewMarkdownEmitter(),
		extractor: extractor.NewAssetExtractor(),
	}
}

// Convert processes an io.ReaderAt into Markdown with the specified options.
func (e *Engine) Convert(ctx context.Context, r io.ReaderAt, size int64, opts Options) (*Result, error) {
	cl, err := cleaner.NewCleaner(opts.StripRegex)
	if err != nil {
		return nil, fmt.Errorf("invalid strip regex: %w", err)
	}

	parserOpts := parser.Options{
		PreserveStyles:  opts.PreserveStyles,
		FormatTables:    opts.FormatTables,
		CleanWatermarks: opts.CleanWatermarks,
		CleanMargins:    opts.CleanMargins,
		TextFilter: func(texts []pdf.Text) ([]pdf.Text, int) {
			return cl.FilterWatermarkTexts(texts, opts.CleanWatermarks)
		},
		LineFilter: func(lines []parser.Line) ([]parser.Line, int) {
			return cl.FilterMarginLines(lines, opts.CleanMargins)
		},
	}

	doc, pageCount, watermarksCleaned, marginsCleaned, detectedEncoding, err := e.parser.Parse(ctx, r, size, parserOpts)
	if err != nil {
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	e.emitter.MergeFormatting = opts.MergeFormatting
	markdown := e.emitter.Render(doc)

	return &Result{
		Markdown:          markdown,
		PageCount:         pageCount,
		CleanedWatermarks: watermarksCleaned,
		CleanedMargins:    marginsCleaned,
		CleanedRows:       watermarksCleaned + marginsCleaned,
		DetectedEncoding:  detectedEncoding,
	}, nil
}

// ConvertFile reads a PDF from disk, executes conversion, and extracts assets if requested.
func (e *Engine) ConvertFile(ctx context.Context, filePath string, opts Options) (*Result, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open source pdf: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to stat source pdf: %w", err)
	}

	result, err := e.Convert(ctx, file, stat.Size(), opts)
	if err != nil {
		return nil, err
	}

	// Extract embedded images if enabled
	if opts.ExtractImages {
		images, err := e.extractor.ExtractImages(ctx, filePath, opts.ImageDir)
		if err == nil && len(images) > 0 {
			result.ImagePaths = images

			// Append image gallery links to Markdown output
			var b strings.Builder
			b.WriteString(result.Markdown)
			b.WriteString("\n## Extracted Images\n\n")
			for _, img := range images {
				relPath := filepath.ToSlash(img)
				name := filepath.Base(img)
				b.WriteString(fmt.Sprintf("![%s](%s)\n\n", name, relPath))
			}
			result.Markdown = b.String()
		}
	}

	return result, nil
}
