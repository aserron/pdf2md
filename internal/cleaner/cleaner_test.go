package cleaner

import (
	"testing"

	"github.com/aserron/pdf2md/internal/ast"
	"github.com/aserron/pdf2md/internal/parser"
)

func TestCleanerWatermarks(t *testing.T) {
	c, err := NewCleaner("")
	if err != nil {
		t.Fatalf("failed to create cleaner: %v", err)
	}

	lines := []parser.Line{
		{Y: 780, Spans: []ast.TextSpan{{Text: "Page 1 of 12"}}},
		{Y: 500, Spans: []ast.TextSpan{{Text: "CONFIDENTIAL DRAFT"}}},
		{Y: 400, Spans: []ast.TextSpan{{Text: "Normal document content."}}},
		{Y: 30, Spans: []ast.TextSpan{{Text: "12"}}},
	}

	filtered, removed := c.FilterLines(lines, true)
	if removed != 3 {
		t.Fatalf("expected 3 removed lines, got %d", removed)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 remaining line, got %d", len(filtered))
	}

	if filtered[0].Spans[0].Text != "Normal document content." {
		t.Errorf("unexpected content: %s", filtered[0].Spans[0].Text)
	}
}

func TestCleanerCustomRegex(t *testing.T) {
	// Strip UUID-like patterns
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

	filtered, _ := c.FilterLines(lines, false)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 line, got %d", len(filtered))
	}

	if filtered[0].Spans[0].Text != "Audit Log entry  processed." {
		t.Errorf("regex replacement failed, got: %s", filtered[0].Spans[0].Text)
	}
}
