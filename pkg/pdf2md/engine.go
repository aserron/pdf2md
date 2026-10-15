package pdf2md

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aserron/pdf2md/internal/parser"
	"github.com/aserron/pdf2md/internal/render"
)

// Engine is the default implementation of Converter.
type Engine struct {
	parser  *parser.PDFParser
	emitter *render.MarkdownEmitter
}

// NewEngine creates an initialized Engine ready for PDF conversions.
func NewEngine() *Engine {
	return &Engine{
		parser:  parser.NewPDFParser(),
		emitter: render.NewMarkdownEmitter(),
	}
}

// Convert processes an io.ReaderAt into Markdown with the specified options.
func (e *Engine) Convert(ctx context.Context, r io.ReaderAt, size int64, opts Options) (*Result, error) {
	parserOpts := parser.Options{
		PreserveStyles: opts.PreserveStyles,
	}
	doc, pageCount, err := e.parser.Parse(ctx, r, size, parserOpts)
	if err != nil {
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	markdown := e.emitter.Render(doc)

	return &Result{
		Markdown:    markdown,
		PageCount:   pageCount,
		CleanedRows: 0,
	}, nil
}

// ConvertFile reads a PDF from disk and executes conversion.
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

	return e.Convert(ctx, file, stat.Size(), opts)
}
