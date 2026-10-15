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
