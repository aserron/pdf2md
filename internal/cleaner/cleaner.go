// Package cleaner handles sanitization of watermarks, running headers/footers, and custom patterns.
package cleaner

import (
	"regexp"
	"strings"

	"github.com/aserron/pdf2md/internal/ast"
	"github.com/aserron/pdf2md/internal/parser"
	"github.com/ledongthuc/pdf"
)

var (
	// Default common watermark patterns
	watermarkWords = []string{
		"confidential",
		"draft",
		"sample",
		"watermark",
		"do not duplicate",
		"do not copy",
		"internal use only",
		"preliminary",
		"unapproved",
		"proprietary",
		"unlicensed",
	}

	// Pagination patterns for headers/footers (e.g. "Page 1", "Page 1 of 5", "1 / 10", "- 1 -")
	pageNumberRegex = regexp.MustCompile(`^(?i)(page\s+\d+(\s+of\s+\d+)?|\d+\s*/\s*\d+|[-–—]\s*\d+\s*[-–—]|\d+)$`)
	emailWatermark  = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}(\+[a-z0-9]+)?`)
)

// Cleaner provides methods to sanitize text lines and raw PDF elements.
type Cleaner struct {
	customRegex *regexp.Regexp
}

// NewCleaner creates a Cleaner configured with an optional user regex.
func NewCleaner(stripPattern string) (*Cleaner, error) {
	var re *regexp.Regexp
	var err error
	if stripPattern != "" {
		re, err = regexp.Compile(stripPattern)
		if err != nil {
			return nil, err
		}
	}
	return &Cleaner{customRegex: re}, nil
}

// FilterWatermarkTexts purges microprint, off-canvas, and explicit watermark strings from raw page text.
// Executing this before font analysis prevents microscopic watermarks from skewing font hierarchies.
func (c *Cleaner) FilterWatermarkTexts(texts []pdf.Text, cleanWatermarks bool) ([]pdf.Text, int) {
	if !cleanWatermarks || len(texts) == 0 {
		return texts, 0
	}

	var filtered []pdf.Text
	removed := 0

	for _, t := range texts {
		// 1. Off-canvas coordinates
		if t.X < -50 || t.Y < -50 {
			removed++
			continue
		}

		// 2. Microprint watermarks (human readable text in documents is rarely < 6.0pt)
		if t.FontSize > 0 && t.FontSize < 6.0 {
			removed++
			continue
		}

		// 3. Explicit watermark keywords
		if isWatermarkText(t.S) || emailWatermark.MatchString(t.S) {
			removed++
			continue
		}

		filtered = append(filtered, t)
	}

	return filtered, removed
}

// FilterMarginLines cleans running headers, footers, and pagination artifacts from assembled lines.
func (c *Cleaner) FilterMarginLines(lines []parser.Line, cleanMargins bool) ([]parser.Line, int) {
	if len(lines) == 0 {
		return nil, 0
	}

	var filtered []parser.Line
	removedCount := 0

	for _, line := range lines {
		raw := lineContent(line)
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			removedCount++
			continue
		}

		if cleanMargins {
			// In standard 72-dpi letter/A4 pages:
			// Top margin is typically Y > 750, bottom margin is Y < 55.
			if (line.Y > 750 || line.Y < 55) && isMarginArtifact(trimmed) {
				removedCount++
				continue
			}
		}

		// Custom regex stripping
		if c.customRegex != nil {
			cleanedSpans := c.stripSpans(line.Spans)
			if len(cleanedSpans) == 0 || isEmptySpans(cleanedSpans) {
				removedCount++
				continue
			}
			line.Spans = cleanedSpans
		}

		filtered = append(filtered, line)
	}

	return filtered, removedCount
}

func isWatermarkText(text string) bool {
	lower := strings.ToLower(text)
	for _, word := range watermarkWords {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

func isMarginArtifact(text string) bool {
	if pageNumberRegex.MatchString(text) {
		return true
	}
	if len(text) <= 3 && strings.Trim(text, "0123456789- ") == "" {
		return true
	}
	return false
}

func (c *Cleaner) stripSpans(spans []ast.TextSpan) []ast.TextSpan {
	var result []ast.TextSpan
	for _, span := range spans {
		stripped := c.customRegex.ReplaceAllString(span.Text, "")
		if stripped != "" {
			span.Text = stripped
			result = append(result, span)
		}
	}
	return result
}

func isEmptySpans(spans []ast.TextSpan) bool {
	for _, s := range spans {
		if strings.TrimSpace(s.Text) != "" {
			return false
		}
	}
	return true
}

func lineContent(line parser.Line) string {
	var b strings.Builder
	for _, s := range line.Spans {
		b.WriteString(s.Text)
	}
	return b.String()
}
