package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aserron/pdf2md/internal/ast"
	"github.com/aserron/pdf2md/internal/cleaner"
	"github.com/aserron/pdf2md/internal/parser"
	"github.com/aserron/pdf2md/internal/render"
	"github.com/aserron/pdf2md/pkg/pdf2md"
)

func main() {
	fmt.Println("Running pdf2md verification suite...")

	// 1. Cleaner tests
	fmt.Print("- Testing cleaner...")
	cl, err := cleaner.NewCleaner(`\[ID:\d+\]`)
	if err != nil {
		fatal("cleaner init failed: %v", err)
	}

	testLines := []parser.Line{
		{Y: 780, Spans: []ast.TextSpan{{Text: "Page 1 of 5"}}},
		{Y: 500, Spans: []ast.TextSpan{{Text: "CONFIDENTIAL DRAFT"}}},
		{Y: 300, Spans: []ast.TextSpan{{Text: "Active item [ID:999] in table."}}},
		{Y: 20, Spans: []ast.TextSpan{{Text: "1"}}},
	}

	filtered, removed := cl.FilterMarginLines(testLines, true)
	if removed != 2 || len(filtered) != 2 {
		fatal("cleaner expected 2 removed, 2 remaining; got %d removed, %d remaining", removed, len(filtered))
	}
	if !strings.Contains(filtered[1].Spans[0].Text, "Active item  in table.") {
		fatal("regex stripping failed: %q", filtered[1].Spans[0].Text)
	}
	fmt.Println(" OK")

	// 2. Render & Table tests
	fmt.Print("- Testing markdown renderer & table emitter...")
	doc := ast.NewDocument()
	doc.Append(&ast.Node{
		Type:  ast.NodeHeading,
		Level: 2,
		Spans: []ast.TextSpan{{Text: "Summary Table"}},
	})
	doc.Append(&ast.Node{
		Type: ast.NodeTable,
		Children: []*ast.Node{
			{
				Type: ast.NodeTableRow,
				Children: []*ast.Node{
					{Type: ast.NodeTableCell, Spans: []ast.TextSpan{{Text: "Metric"}}},
					{Type: ast.NodeTableCell, Spans: []ast.TextSpan{{Text: "Value"}}},
				},
			},
			{
				Type: ast.NodeTableRow,
				Children: []*ast.Node{
					{Type: ast.NodeTableCell, Spans: []ast.TextSpan{{Text: "Throughput"}}},
					{Type: ast.NodeTableCell, Spans: []ast.TextSpan{{Text: "150 MB/s"}}},
				},
			},
		},
	})

	emitter := render.NewMarkdownEmitter()
	rendered := emitter.Render(doc)
	if !strings.Contains(rendered, "## Summary Table") || !strings.Contains(rendered, "| Metric | Value |") {
		fatal("render failed, got:\n%s", rendered)
	}
	fmt.Println(" OK")

	// 3. Engine End-to-End Test
	fmt.Print("- Testing engine conversion with sample PDF...")
	engine := pdf2md.NewEngine()
	opts := pdf2md.DefaultOptions()
	res, err := engine.ConvertFile(context.Background(), "testdata/sample.pdf", opts)
	if err != nil {
		fatal("engine conversion failed: %v", err)
	}
	if res.PageCount < 1 || !strings.Contains(res.Markdown, "Hello") {
		fatal("unexpected engine output:\n%s", res.Markdown)
	}
	fmt.Println(" OK")

	fmt.Println("\nAll verification checks PASSED successfully!")
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\nFAIL: "+format+"\n", args...)
	os.Exit(1)
}
