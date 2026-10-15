package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aserron/pdf2md/pkg/pdf2md"
)

var (
	version = "0.1.0-dev"
)

func main() {
	opts := pdf2md.DefaultOptions()

	input := flag.String("i", "", "Path to input PDF file")
	output := flag.String("o", "", "Path to output Markdown file (default: stdout)")
	flag.BoolVar(&opts.ExtractImages, "images", opts.ExtractImages, "Extract embedded images")
	flag.StringVar(&opts.ImageDir, "image-dir", opts.ImageDir, "Directory to store extracted images")
	flag.BoolVar(&opts.CleanWatermarks, "clean", opts.CleanWatermarks, "Strip watermarks and margin headers/footers")
	flag.BoolVar(&opts.PreserveStyles, "styles", opts.PreserveStyles, "Preserve bold, italic, and code typography")
	flag.BoolVar(&opts.FormatTables, "tables", opts.FormatTables, "Detect and format markdown tables")
	flag.StringVar(&opts.StripRegex, "strip", "", "Regex pattern to strip from output")
	showVersion := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "pdf2md - High-Performance PDF to Markdown Engine (v%s)\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: pdf2md -i <input.pdf> [-o <output.md>] [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
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

	fmt.Fprintf(os.Stderr, "pdf2md: converting %s...\n", *input)
	if *output != "" {
		fmt.Fprintf(os.Stderr, "pdf2md: target output %s\n", *output)
	}
}
