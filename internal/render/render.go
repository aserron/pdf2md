// Package render formats an AST Document into standard Markdown.
package render

import (
	"fmt"
	"strings"

	"github.com/aserron/pdf2md/internal/ast"
)

// MarkdownEmitter converts an AST Document into standard Markdown text.
type MarkdownEmitter struct {
	builder strings.Builder
}

// NewMarkdownEmitter instantiates an emitter.
func NewMarkdownEmitter() *MarkdownEmitter {
	return &MarkdownEmitter{}
}

// Render traverses the AST and produces a markdown string.
func (e *MarkdownEmitter) Render(doc *ast.Document) string {
	if doc == nil || len(doc.Nodes) == 0 {
		return ""
	}

	e.builder.Reset()
	for i, node := range doc.Nodes {
		e.renderNode(node)
		if i < len(doc.Nodes)-1 {
			e.builder.WriteString("\n\n")
		}
	}
	e.builder.WriteString("\n")
	return e.builder.String()
}

func (e *MarkdownEmitter) renderNode(node *ast.Node) {
	if node == nil {
		return
	}

	switch node.Type {
	case ast.NodeHeading:
		prefix := strings.Repeat("#", clamp(node.Level, 1, 6))
		e.builder.WriteString(prefix)
		e.builder.WriteString(" ")
		e.renderSpans(node.Spans)

	case ast.NodeParagraph:
		e.renderSpans(node.Spans)

	case ast.NodeCodeBlock:
		e.builder.WriteString("```\n")
		e.renderSpans(node.Spans)
		e.builder.WriteString("\n```")

	case ast.NodeBlockquote:
		e.builder.WriteString("> ")
		e.renderSpans(node.Spans)

	case ast.NodeImage:
		e.builder.WriteString(fmt.Sprintf("![%s](%s)", node.AltText, node.URL))

	case ast.NodeList:
		for idx, item := range node.Children {
			if node.Ordered {
				e.builder.WriteString(fmt.Sprintf("%d. ", idx+1))
			} else {
				e.builder.WriteString("- ")
			}
			e.renderSpans(item.Spans)
			if idx < len(node.Children)-1 {
				e.builder.WriteString("\n")
			}
		}

	case ast.NodeTable:
		if len(node.Children) == 0 {
			return
		}
		header := node.Children[0]
		e.builder.WriteString("|")
		for _, cell := range header.Children {
			e.builder.WriteString(" ")
			e.renderSpans(cell.Spans)
			e.builder.WriteString(" |")
		}
		e.builder.WriteString("\n|")
		for range header.Children {
			e.builder.WriteString(" --- |")
		}
		for _, row := range node.Children[1:] {
			e.builder.WriteString("\n|")
			for _, cell := range row.Children {
				e.builder.WriteString(" ")
				e.renderSpans(cell.Spans)
				e.builder.WriteString(" |")
			}
		}

	case ast.NodeHorizontalRule:
		e.builder.WriteString("---")
	}
}

func (e *MarkdownEmitter) renderSpans(spans []ast.TextSpan) {
	for _, span := range spans {
		text := span.Text
		if span.Style&ast.StyleBold != 0 {
			text = "**" + text + "**"
		}
		if span.Style&ast.StyleItalic != 0 {
			text = "*" + text + "*"
		}
		if span.Style&ast.StyleCode != 0 {
			text = "`" + text + "`"
		}
		e.builder.WriteString(text)
	}
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
