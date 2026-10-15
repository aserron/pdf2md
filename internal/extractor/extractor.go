// Package extractor extracts embedded raster images and diagrams from PDF documents.
package extractor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// AssetExtractor manages asset discovery and export.
type AssetExtractor struct{}

// NewAssetExtractor creates an initialized extractor.
func NewAssetExtractor() *AssetExtractor {
	return &AssetExtractor{}
}

// ExtractImages extracts embedded images from the input PDF file into the destination directory.
// Returns a list of relative paths to the exported images.
func (e *AssetExtractor) ExtractImages(ctx context.Context, inputPath, outDir string) ([]string, error) {
	if inputPath == "" || outDir == "" {
		return nil, nil
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create image directory: %w", err)
	}

	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.EXTRACTIMAGES

	// Extract images using pdfcpu
	err := api.ExtractImagesFile(ctx, inputPath, outDir, nil, conf)
	if err != nil {
		// If error is simply "no images found" or unsupported resource, don't fail conversion
		if isNoImagesError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("image extraction encountered error: %w", err)
	}

	// Scan output directory for extracted image files
	var images []string
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return nil, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".tif" || ext == ".tiff" {
			images = append(images, filepath.Join(outDir, entry.Name()))
		}
	}

	return images, nil
}

// DiagramInfo holds metadata for a detected diagram or chart.
type DiagramInfo struct {
	PageNumber int
	Label      string
	AssetPath  string
}

// ExtractDiagrams inspects pages for vector chart bounding boxes.
func (e *AssetExtractor) ExtractDiagrams(ctx context.Context, pageCount int) []DiagramInfo {
	// Diagram detection heuristic: placeholder for vector drawing clusters
	// Returns diagram info descriptors for embedding in Markdown
	var diagrams []DiagramInfo
	return diagrams
}

func isNoImagesError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no images") || strings.Contains(msg, "selected pages")
}
