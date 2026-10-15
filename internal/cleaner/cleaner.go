// Package cleaner handles sanitization of watermarks, running headers/footers, and custom patterns.
package cleaner

import (
	"regexp"
	"strings"

	"github.com/aserron/pdf2md/internal/ast"
	"github.com/aserron/pdf2md/internal/parser"
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
	}

	// Pagination patterns for headers/footers (e.g. "Page 1", "Page 1 of 5", "1 / 10", "- 1 -")
	pageNumberRegex = regexp.MustCompile(`^(?i)(page\s+\d+(\s+of\s+\d+)?|\d+\s*/\s*\d+|[-–—]\s*\d+\s*[-–—]|\d+)$`)
)

// Cleaner provides methods to sanitize text lines.
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

// FilterLines cleans invasive elements, watermarks, and margin artifacts from parsed lines.
// Returns the filtered lines and the count of removed lines.
func (c *Cleaner) FilterLines(lines []parser.Line, cleanWatermarks bool) ([]parser.Line, int) {
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

		if cleanWatermarks {
			// 1. Watermark phrase check
			if isWatermarkText(trimmed) {
				removedCount++
				continue
			}

			// 2. Margin header/footer check
			// In standard 72-dpi letter/A4 pages:
			// Top margin is typically Y > 750, bottom margin is Y < 55.
			if (line.Y > 750 || line.Y < 55) && isMarginArtifact(trimmed) {
				removedCount++
				continue
			}
		}

		// 3. Custom regex stripping
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
	// Check if matches page number pattern
	if pageNumberRegex.MatchString(text) {
		return true
	}
	// Also strip isolated single/double digits in margins
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
