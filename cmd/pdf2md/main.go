package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aserron/pdf2md/pkg/pdf2md"
	"github.com/ledongthuc/pdf"
)

var (
	version = "1.0.0"
)

// InspectReport holds diagnostic structure data for a PDF file.
type InspectReport struct {
	FilePath    string   `json:"file_path"`
	Pages       int      `json:"pages"`
	FontsFound  []string `json:"fonts_found"`
	HasImages   bool     `json:"has_images"`
	FileSize    int64    `json:"file_size_bytes"`
	DurationMs  int64    `json:"analysis_duration_ms"`
}

func main() {
	opts := pdf2md.DefaultOptions()

	input := flag.String("i", "", "Path to input PDF file or folder (required)")
	output := flag.String("o", "", "Path to output Markdown file or directory (default: stdout)")
	flag.BoolVar(&opts.ExtractImages, "images", opts.ExtractImages, "Extract embedded images into image directory")
	flag.StringVar(&opts.ImageDir, "image-dir", opts.ImageDir, "Directory to store extracted images")
	flag.BoolVar(&opts.ExtractDiagrams, "diagrams", opts.ExtractDiagrams, "Detect and extract diagram bounding regions")
	flag.BoolVar(&opts.CleanWatermarks, "clean", opts.CleanWatermarks, "Strip watermarks and margin headers/footers")
	flag.BoolVar(&opts.PreserveStyles, "styles", opts.PreserveStyles, "Preserve bold, italic, and code typography")
	flag.BoolVar(&opts.FormatTables, "tables", opts.FormatTables, "Detect and format markdown tables")
	flag.StringVar(&opts.StripRegex, "strip", "", "Regex pattern to strip from output")
	inspect := flag.Bool("inspect", false, "Output PDF structure diagnostics in JSON without converting")
	quiet := flag.Bool("quiet", false, "Suppress progress and telemetry output")
	showVersion := flag.Bool("version", false, "Print version information and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "pdf2md - High-Performance PDF to Markdown Engine (v%s)\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: pdf2md -i <input.pdf|folder> [-o <output.md|dir>] [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  pdf2md -i document.pdf -o document.md\n")
		fmt.Fprintf(os.Stderr, "  pdf2md -i document.pdf --clean --images --image-dir assets/\n")
		fmt.Fprintf(os.Stderr, "  pdf2md -i document.pdf --strip \"\\[UUID:[^\\]]+\\]\"\n")
		fmt.Fprintf(os.Stderr, "  pdf2md -i ./manuals/ -o ./markdown/ (batch conversion)\n")
		fmt.Fprintf(os.Stderr, "  pdf2md -i document.pdf --inspect\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("pdf2md v%s\n", version)
		os.Exit(0)
	}

	if *input == "" {
		flag.Usage()
		os.Exit(1)
	}

	stat, err := os.Stat(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot access input '%s': %v\n", *input, err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Inspect mode
	if *inspect {
		if stat.IsDir() {
			fmt.Fprintf(os.Stderr, "error: --inspect requires a single PDF file, not a directory\n")
			os.Exit(1)
		}
		if err := runInspect(*input); err != nil {
			fmt.Fprintf(os.Stderr, "inspect error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	engine := pdf2md.NewEngine()

	// Batch folder conversion
	if stat.IsDir() {
		if err := runBatch(ctx, engine, *input, *output, opts, *quiet); err != nil {
			fmt.Fprintf(os.Stderr, "batch error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Single file conversion
	start := time.Now()
	if !*quiet {
		fmt.Fprintf(os.Stderr, "pdf2md: converting '%s'...\n", *input)
	}

	res, err := engine.ConvertFile(ctx, *input, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error converting file: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(start)

	if *output != "" {
		outDir := filepath.Dir(*output)
		if outDir != "." && outDir != "" {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "error creating output directory: %v\n", err)
				os.Exit(1)
			}
		}
		if err := os.WriteFile(*output, []byte(res.Markdown), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing markdown file '%s': %v\n", *output, err)
			os.Exit(1)
		}
		if !*quiet {
			fmt.Fprintf(os.Stderr, "pdf2md: wrote %d bytes to '%s'\n", len(res.Markdown), *output)
		}
	} else {
		// Output to stdout
		io.WriteString(os.Stdout, res.Markdown)
	}

	if !*quiet {
		fmt.Fprintf(os.Stderr, "pdf2md: finished in %v | Pages: %d | Cleaned Artifacts: %d | Extracted Images: %d\n",
			duration.Round(time.Millisecond), res.PageCount, res.CleanedRows, len(res.ImagePaths))
	}
}

func runBatch(ctx context.Context, engine *pdf2md.Engine, inDir, outDir string, opts pdf2md.Options, quiet bool) error {
	entries, err := os.ReadDir(inDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var pdfFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.ToLower(filepath.Ext(entry.Name())) == ".pdf" {
			pdfFiles = append(pdfFiles, filepath.Join(inDir, entry.Name()))
		}
	}

	if len(pdfFiles) == 0 {
		return fmt.Errorf("no .pdf files found in directory '%s'", inDir)
	}

	if outDir == "" {
		outDir = filepath.Join(inDir, "markdown")
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "pdf2md: found %d PDF files to convert into '%s'\n", len(pdfFiles), outDir)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 4) // max 4 concurrent conversions
	var errCount int
	var mu sync.Mutex

	start := time.Now()

	for _, file := range pdfFiles {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(pdfPath string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
			outFile := filepath.Join(outDir, baseName+".md")

			fileOpts := opts
			if opts.ExtractImages {
				fileOpts.ImageDir = filepath.Join(outDir, "assets", baseName)
			}

			res, err := engine.ConvertFile(ctx, pdfPath, fileOpts)
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
				fmt.Fprintf(os.Stderr, "failed to convert %s: %v\n", filepath.Base(pdfPath), err)
				return
			}

			if err := os.WriteFile(outFile, []byte(res.Markdown), 0644); err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
				fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", outFile, err)
				return
			}

			if !quiet {
				fmt.Fprintf(os.Stderr, "converted: %s -> %s (pages: %d)\n", filepath.Base(pdfPath), filepath.Base(outFile), res.PageCount)
			}
		}(file)
	}

	wg.Wait()
	totalTime := time.Since(start)

	if !quiet {
		fmt.Fprintf(os.Stderr, "pdf2md: batch completed in %v. Converted: %d, Failed: %d\n",
			totalTime.Round(time.Millisecond), len(pdfFiles)-errCount, errCount)
	}

	if errCount > 0 {
		return fmt.Errorf("%d conversions failed", errCount)
	}
	return nil
}

func runInspect(filePath string) error {
	start := time.Now()
	f, reader, err := pdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open pdf: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	numPages := reader.NumPage()
	fontMap := make(map[string]bool)

	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		for _, fontName := range page.Fonts() {
			fontMap[fontName] = true
		}
	}

	var fonts []string
	for font := range fontMap {
		fonts = append(fonts, font)
	}

	report := InspectReport{
		FilePath:   filePath,
		Pages:      numPages,
		FontsFound: fonts,
		FileSize:   stat.Size(),
		DurationMs: time.Since(start).Milliseconds(),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
