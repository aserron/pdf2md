package cleaner

import (
	"testing"

	"github.com/aserron/pdf2md/internal/ast"
	"github.com/aserron/pdf2md/internal/parser"
	"github.com/ledongthuc/pdf"
)

func TestCleanerWatermarkTexts(t *testing.T) {
	c, err := NewCleaner("")
	if err != nil {
		t.Fatalf("failed to create cleaner: %v", err)
	}

	texts := []pdf.Text{
		{FontSize: 4.2, X: -100, Y: 0, S: "CONFIDENTIAL tracking watermark"},
		{FontSize: 12.0, X: 50, Y: 500, S: "Normal human readable text"},
		{FontSize: 3.0, X: 50, Y: 500, S: "microprint"},
	}

	filtered, removed := c.FilterWatermarkTexts(texts, true)
	if removed != 2 {
		t.Fatalf("expected 2 removed watermarks, got %d", removed)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 remaining text, got %d", len(filtered))
	}
	if filtered[0].S != "Normal human readable text" {
		t.Errorf("unexpected remaining text: %s", filtered[0].S)
	}
}

func TestCleanerMargins(t *testing.T) {
	c, err := NewCleaner("")
	if err != nil {
		t.Fatalf("failed to create cleaner: %v", err)
	}

	lines := []parser.Line{
		{Y: 780, Spans: []ast.TextSpan{{Text: "Page 1 of 12"}}},
		{Y: 400, Spans: []ast.TextSpan{{Text: "Normal document content."}}},
		{Y: 30, Spans: []ast.TextSpan{{Text: "12"}}},
	}

	filtered, removed := c.FilterMarginLines(lines, true)
	if removed != 2 {
		t.Fatalf("expected 2 removed margin lines, got %d", removed)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 remaining line, got %d", len(filtered))
	}

	if filtered[0].Spans[0].Text != "Normal document content." {
		t.Errorf("unexpected content: %s", filtered[0].Spans[0].Text)
	}
}

func TestCleanerCustomRegex(t *testing.T) {
	c, err := NewCleaner(`\[UUID:[a-f0-9\-]+\]`)
	if err != nil {
		t.Fatalf("failed to compile regex: %v", err)
	}

	lines := []parser.Line{
		{
			Y: 400,
			Spans: []ast.TextSpan{
				{Text: "Audit Log entry [UUID:1234-abcd] processed."},
			},
		},
	}

	filtered, _ := c.FilterMarginLines(lines, false)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 line, got %d", len(filtered))
	}

	if filtered[0].Spans[0].Text != "Audit Log entry  processed." {
		t.Errorf("regex replacement failed, got: %s", filtered[0].Spans[0].Text)
	}
}
