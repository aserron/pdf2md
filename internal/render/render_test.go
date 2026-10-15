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
