// Package parser converts raw PDF contents into structured AST documents.
package parser

import (
	"context"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/aserron/pdf2md/internal/ast"
	"github.com/ledongthuc/pdf"
)

// Options holds configuration needed by the parser.
type Options struct {
	PreserveStyles bool
}

// Line represents a single horizontal line of text fragments.
type Line struct {
	Y        float64
	PageNum  int
	Spans    []ast.TextSpan
	Elements []pdf.Text
}

// RawPage captures the extracted text elements and page dimensions.
type RawPage struct {
	Number int
	Texts  []pdf.Text
	Lines  []Line
}

// PDFParser orchestrates reading and AST generation.
type PDFParser struct{}

// NewPDFParser creates a new parser instance.
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

// Parse extracts text and layout from an io.ReaderAt into an AST Document.
func (p *PDFParser) Parse(ctx context.Context, r io.ReaderAt, size int64, opts Options) (*ast.Document, int, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open pdf reader: %w", err)
	}

	numPages := reader.NumPage()
	if numPages == 0 {
		return ast.NewDocument(), 0, nil
	}

	rawPages := make([]RawPage, numPages)
	var allFontSizes []float64

	for i := 1; i <= numPages; i++ {
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		default:
		}

		page := reader.Page(i)
		content := page.Content()
		rawPages[i-1] = RawPage{
			Number: i,
			Texts:  content.Text,
		}

		for _, t := range content.Text {
			if strings.TrimSpace(t.S) != "" && t.FontSize > 0 {
				allFontSizes = append(allFontSizes, math.Round(t.FontSize*10)/10)
			}
		}
	}

	// Calculate font size clustering to determine heading thresholds
	fontHierarchy := computeFontHierarchy(allFontSizes)

	doc := ast.NewDocument()

	for _, page := range rawPages {
		lines := groupIntoLines(page.Texts, page.Number)
		blockNodes := processLinesToBlocks(lines, fontHierarchy, opts)
		for _, node := range blockNodes {
			doc.Append(node)
		}
	}

	return doc, numPages, nil
}

// computeFontHierarchy clusters font sizes and identifies body font vs headings.
type fontHierarchy struct {
	bodyFontSize float64
	headingSizes []float64 // Sorted descending: [0] = H1, [1] = H2, etc.
}

func computeFontHierarchy(sizes []float64) fontHierarchy {
	if len(sizes) == 0 {
		return fontHierarchy{bodyFontSize: 12.0}
	}

	freq := make(map[float64]int)
	for _, s := range sizes {
		freq[s]++
	}

	// Find the most frequent size as body font
	var bodySize float64
	var maxCount int
	for s, count := range freq {
		if count > maxCount {
			maxCount = count
			bodySize = s
		}
	}

	// Collect unique sizes larger than body font
	var largerSizes []float64
	for s := range freq {
		if s > bodySize+1.0 { // at least 1pt larger than body
			largerSizes = append(largerSizes, s)
		}
	}

	sort.Slice(largerSizes, func(i, j int) bool {
		return largerSizes[i] > largerSizes[j]
	})

	return fontHierarchy{
		bodyFontSize: bodySize,
		headingSizes: largerSizes,
	}
}

// groupIntoLines clusters text elements with similar Y coordinates into ordered lines.
func groupIntoLines(texts []pdf.Text, pageNum int) []Line {
	if len(texts) == 0 {
		return nil
	}

	// Filter empty text
	var valid []pdf.Text
	for _, t := range texts {
		if strings.TrimSpace(t.S) != "" {
			valid = append(valid, t)
		}
	}

	if len(valid) == 0 {
		return nil
	}

	// Sort primary by Y descending (PDF top to bottom), secondary by X ascending (left to right)
	const yTolerance = 2.5
	sort.Slice(valid, func(i, j int) bool {
		diff := valid[i].Y - valid[j].Y
		if math.Abs(diff) <= yTolerance {
			return valid[i].X < valid[j].X
		}
		return valid[i].Y > valid[j].Y
	})

	var lines []Line
	var currentElements []pdf.Text
	currentY := valid[0].Y

	for _, elem := range valid {
		if math.Abs(elem.Y-currentY) <= yTolerance {
			currentElements = append(currentElements, elem)
		} else {
			if len(currentElements) > 0 {
				lines = append(lines, assembleLine(currentElements, currentY, pageNum))
			}
			currentElements = []pdf.Text{elem}
			currentY = elem.Y
		}
	}

	if len(currentElements) > 0 {
		lines = append(lines, assembleLine(currentElements, currentY, pageNum))
	}

	return lines
}

// assembleLine orders text fragments by X position and combines them into styled TextSpans.
func assembleLine(elements []pdf.Text, y float64, pageNum int) Line {
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].X < elements[j].X
	})

	var spans []ast.TextSpan

	for i, elem := range elements {
		style := detectStyle(elem.Font)

		// Insert space if there's a horizontal gap between adjacent fragments
		if i > 0 {
			prev := elements[i-1]
			gap := elem.X - (prev.X + prev.W)
			// If gap is greater than ~1.5 points, insert spacing
			if gap > 1.5 && !strings.HasSuffix(spans[len(spans)-1].Text, " ") && !strings.HasPrefix(elem.S, " ") {
				spans[len(spans)-1].Text += " "
			}
		}

		spans = append(spans, ast.TextSpan{
			Text:     elem.S,
			Style:    style,
			FontSize: elem.FontSize,
			FontName: elem.Font,
		})
	}

	return Line{
		Y:        y,
		PageNum:  pageNum,
		Spans:    spans,
		Elements: elements,
	}
}

func detectStyle(fontName string) ast.InlineStyle {
	lower := strings.ToLower(fontName)
	var style ast.InlineStyle

	if strings.Contains(lower, "bold") || strings.Contains(lower, "black") || strings.Contains(lower, "heavy") {
		style |= ast.StyleBold
	}
	if strings.Contains(lower, "italic") || strings.Contains(lower, "oblique") {
		style |= ast.StyleItalic
	}
	if strings.Contains(lower, "courier") || strings.Contains(lower, "mono") || strings.Contains(lower, "code") || strings.Contains(lower, "consolas") {
		style |= ast.StyleCode
	}

	return style
}

var (
	bulletRegex = regexp.MustCompile(`^[-*•◦▪]\s+(.*)$`)
	numberRegex = regexp.MustCompile(`^(\d+)[\.\)]\s+(.*)$`)
)

// processLinesToBlocks classifies lines into Markdown AST nodes (headings, code blocks, lists, paragraphs, tables).
func processLinesToBlocks(lines []Line, hierarchy fontHierarchy, opts Options) []*ast.Node {
	var nodes []*ast.Node
	var currentList *ast.Node
	var codeBlockLines []string
	inCodeBlock := false

	flushList := func() {
		if currentList != nil {
			nodes = append(nodes, currentList)
			currentList = nil
		}
	}

	flushCodeBlock := func() {
		if inCodeBlock {
			nodes = append(nodes, &ast.Node{
				Type:  ast.NodeCodeBlock,
				Spans: []ast.TextSpan{{Text: strings.Join(codeBlockLines, "\n")}},
			})
			codeBlockLines = nil
			inCodeBlock = false
		}
	}

	for _, line := range lines {
		rawText := lineText(line)
		trimmed := strings.TrimSpace(rawText)
		if trimmed == "" {
			continue
		}

		// Horizontal rule check
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			flushList()
			flushCodeBlock()
			nodes = append(nodes, &ast.Node{Type: ast.NodeHorizontalRule})
			continue
		}

		// Check if line is purely monospace font (potential code block)
		isCodeLine := true
		for _, s := range line.Spans {
			if s.Style&ast.StyleCode == 0 {
				isCodeLine = false
				break
			}
		}

		if isCodeLine && opts.PreserveStyles {
			flushList()
			inCodeBlock = true
			codeBlockLines = append(codeBlockLines, rawText)
			continue
		} else if inCodeBlock {
			flushCodeBlock()
		}

		// Check for heading level
		headingLevel := getHeadingLevel(line, hierarchy)
		if headingLevel > 0 {
			flushList()
			flushCodeBlock()
			nodes = append(nodes, &ast.Node{
				Type:  ast.NodeHeading,
				Level: headingLevel,
				Spans: line.Spans,
			})
			continue
		}

		// Blockquote check
		if strings.HasPrefix(trimmed, ">") {
			flushList()
			flushCodeBlock()
			content := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			nodes = append(nodes, &ast.Node{
				Type:  ast.NodeBlockquote,
				Spans: []ast.TextSpan{{Text: content}},
			})
			continue
		}

		// Unordered list check
		if match := bulletRegex.FindStringSubmatch(trimmed); match != nil {
			flushCodeBlock()
			if currentList == nil || currentList.Ordered {
				flushList()
				currentList = &ast.Node{Type: ast.NodeList, Ordered: false}
			}
			currentList.Children = append(currentList.Children, &ast.Node{
				Type:  ast.NodeListItem,
				Spans: []ast.TextSpan{{Text: match[1]}},
			})
			continue
		}

		// Ordered list check
		if match := numberRegex.FindStringSubmatch(trimmed); match != nil {
			flushCodeBlock()
			if currentList == nil || !currentList.Ordered {
				flushList()
				currentList = &ast.Node{Type: ast.NodeList, Ordered: true}
			}
			currentList.Children = append(currentList.Children, &ast.Node{
				Type:  ast.NodeListItem,
				Spans: []ast.TextSpan{{Text: match[2]}},
			})
			continue
		}

		// Default: Paragraph
		flushList()
		flushCodeBlock()
		nodes = append(nodes, &ast.Node{
			Type:  ast.NodeParagraph,
			Spans: line.Spans,
		})
	}

	flushList()
	flushCodeBlock()

	return nodes
}

func getHeadingLevel(line Line, hierarchy fontHierarchy) int {
	if len(line.Spans) == 0 {
		return 0
	}

	// Calculate maximum font size in line
	var maxFontSize float64
	for _, span := range line.Spans {
		if span.FontSize > maxFontSize {
			maxFontSize = span.FontSize
		}
	}

	for idx, size := range hierarchy.headingSizes {
		if math.Abs(maxFontSize-size) <= 0.5 {
			return idx + 1 // 1-indexed heading level
		}
	}

	return 0
}

func lineText(line Line) string {
	var b strings.Builder
	for _, s := range line.Spans {
		b.WriteString(s.Text)
	}
	return b.String()
}
