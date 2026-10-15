package parser

import (
	"strings"
	"testing"

	"github.com/aserron/pdf2md/internal/ast"
	"github.com/ledongthuc/pdf"
)

func TestDetectStyle(t *testing.T) {
	cases := []struct {
		font     string
		expected ast.InlineStyle
	}{
		{"Helvetica-Bold", ast.StyleBold},
		{"Times-Italic", ast.StyleItalic},
		{"Courier-BoldOblique", ast.StyleBold | ast.StyleItalic | ast.StyleCode},
		{"ArialMT", ast.StyleNone},
		{"Consolas", ast.StyleCode},
	}

	for _, c := range cases {
		style := detectStyle(c.font)
		if style != c.expected {
			t.Errorf("for font %s, expected %v, got %v", c.font, c.expected, style)
		}
	}
}

func TestComputeFontHierarchy(t *testing.T) {
	// Sample font sizes across document
	sizes := []float64{12, 12, 12, 12, 12, 12, 24, 18, 12, 12, 18, 12}
	h := computeFontHierarchy(sizes)

	if h.bodyFontSize != 12 {
		t.Errorf("expected body size 12, got %v", h.bodyFontSize)
	}

	if len(h.headingSizes) < 2 {
		t.Fatalf("expected at least 2 heading sizes, got %d", len(h.headingSizes))
	}

	if h.headingSizes[0] != 24 {
		t.Errorf("expected H1 size 24, got %v", h.headingSizes[0])
	}
	if h.headingSizes[1] != 18 {
		t.Errorf("expected H2 size 18, got %v", h.headingSizes[1])
	}
}

func TestGroupIntoLinesAndBlocks(t *testing.T) {
	texts := []pdf.Text{
		{Font: "Helvetica-Bold", FontSize: 24, X: 50, Y: 700, W: 100, S: "Document Title"},
		{Font: "Helvetica", FontSize: 12, X: 50, Y: 650, W: 30, S: "Hello"},
		{Font: "Helvetica-Bold", FontSize: 12, X: 85, Y: 650, W: 30, S: "World"},
		{Font: "Helvetica", FontSize: 12, X: 50, Y: 600, W: 80, S: "- First bullet item"},
		{Font: "Helvetica", FontSize: 12, X: 50, Y: 580, W: 80, S: "- Second bullet item"},
	}

	lines := groupIntoLines(texts, 1)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	hierarchy := fontHierarchy{
		bodyFontSize: 12,
		headingSizes: []float64{24},
	}

	opts := Options{PreserveStyles: true}
	nodes := processLinesToBlocks(lines, hierarchy, opts)

	if len(nodes) < 3 {
		t.Fatalf("expected at least 3 nodes, got %d", len(nodes))
	}

	// First node: Heading H1
	if nodes[0].Type != ast.NodeHeading || nodes[0].Level != 1 {
		t.Errorf("expected node 0 to be Heading 1, got type %v, level %d", nodes[0].Type, nodes[0].Level)
	}

	// Second node: Paragraph with styled span
	if nodes[1].Type != ast.NodeParagraph {
		t.Errorf("expected node 1 to be Paragraph, got %v", nodes[1].Type)
	}

	// Third node: List with 2 items
	if nodes[2].Type != ast.NodeList || len(nodes[2].Children) != 2 {
		t.Errorf("expected node 2 to be List with 2 items, got type %v, children %d", nodes[2].Type, len(nodes[2].Children))
	}
}

func TestHorizontalGapSpacing(t *testing.T) {
	// Two words on the same line with a gap
	elements := []pdf.Text{
		{Font: "Helvetica", FontSize: 12, X: 50, Y: 500, W: 25, S: "Hello"},
		{Font: "Helvetica", FontSize: 12, X: 80, Y: 500, W: 25, S: "World"}, // gap = 80 - 75 = 5pt
	}

	line := assembleLine(elements, 500, 1)
	fullText := lineText(line)

	if !strings.Contains(fullText, "Hello World") {
		t.Errorf("expected inserted space between words, got: %q", fullText)
	}
}
