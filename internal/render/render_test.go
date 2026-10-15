package render

import (
	"strings"
	"testing"

	"github.com/aserron/pdf2md/internal/ast"
)

func TestMarkdownEmitter(t *testing.T) {
	doc := ast.NewDocument()
	doc.Append(&ast.Node{
		Type:  ast.NodeHeading,
		Level: 1,
		Spans: []ast.TextSpan{{Text: "Hello World"}},
	})
	doc.Append(&ast.Node{
		Type: ast.NodeParagraph,
		Spans: []ast.TextSpan{
			{Text: "This is "},
			{Text: "bold", Style: ast.StyleBold},
			{Text: " and "},
			{Text: "code", Style: ast.StyleCode},
			{Text: "."},
		},
	})

	emitter := NewMarkdownEmitter()
	out := emitter.Render(doc)

	expected := "# Hello World\n\nThis is **bold** and `code`.\n"
	if out != expected {
		t.Fatalf("unexpected markdown output:\nGot:\n%q\nExpected:\n%q", out, expected)
	}

	if !strings.HasPrefix(out, "# Hello World") {
		t.Errorf("heading not rendered correctly: %s", out)
	}
}

func TestCoalesceSpans(t *testing.T) {
	// Character-by-character bold letters should merge into single token
	spans := []ast.TextSpan{
		{Text: "P", Style: ast.StyleBold},
		{Text: "r", Style: ast.StyleBold},
		{Text: "o", Style: ast.StyleBold},
		{Text: "j", Style: ast.StyleBold},
		{Text: "e", Style: ast.StyleBold},
		{Text: "c", Style: ast.StyleBold},
		{Text: "t", Style: ast.StyleBold},
		{Text: " "},
		{Text: "N", Style: ast.StyleItalic},
		{Text: "a", Style: ast.StyleItalic},
		{Text: "m", Style: ast.StyleItalic},
		{Text: "e", Style: ast.StyleItalic},
	}

	merged := CoalesceSpans(spans)
	if len(merged) != 3 {
		t.Fatalf("expected 3 merged spans, got %d", len(merged))
	}

	if merged[0].Text != "Project" || merged[0].Style != ast.StyleBold {
		t.Errorf("expected merged 'Project' bold, got: %q (%v)", merged[0].Text, merged[0].Style)
	}
	if merged[1].Text != " " || merged[1].Style != ast.StyleNone {
		t.Errorf("expected space span, got: %q", merged[1].Text)
	}
	if merged[2].Text != "Name" || merged[2].Style != ast.StyleItalic {
		t.Errorf("expected merged 'Name' italic, got: %q (%v)", merged[2].Text, merged[2].Style)
	}

	doc := ast.NewDocument()
	doc.Append(&ast.Node{
		Type:  ast.NodeParagraph,
		Spans: spans,
	})

	emitter := NewMarkdownEmitter()
	emitter.MergeFormatting = true
	rendered := emitter.Render(doc)

	expected := "**Project** *Name*\n"
	if rendered != expected {
		t.Fatalf("unexpected coalesced rendering:\nGot:\n%q\nExpected:\n%q", rendered, expected)
	}
}

func TestTableRendering(t *testing.T) {
	table := &ast.Node{
		Type: ast.NodeTable,
		Children: []*ast.Node{
			{
				Type: ast.NodeTableRow,
				Children: []*ast.Node{
					{Type: ast.NodeTableCell, Spans: []ast.TextSpan{{Text: "Name"}}},
					{Type: ast.NodeTableCell, Spans: []ast.TextSpan{{Text: "Role"}}},
				},
			},
			{
				Type: ast.NodeTableRow,
				Children: []*ast.Node{
					{Type: ast.NodeTableCell, Spans: []ast.TextSpan{{Text: "Alice"}}},
					{Type: ast.NodeTableCell, Spans: []ast.TextSpan{{Text: "Engineer"}}},
				},
			},
		},
	}

	doc := ast.NewDocument()
	doc.Append(table)

	emitter := NewMarkdownEmitter()
	out := emitter.Render(doc)

	expected := "| Name | Role |\n| --- | --- |\n| Alice | Engineer |\n"
	if out != expected {
		t.Fatalf("unexpected table rendering:\nGot:\n%q\nExpected:\n%q", out, expected)
	}
}
