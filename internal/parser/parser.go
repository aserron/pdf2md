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
	PreserveStyles  bool
	FormatTables    bool
	CleanWatermarks bool
	CleanMargins    bool
	TextFilter      func([]pdf.Text) ([]pdf.Text, int)
	LineFilter      func([]Line) ([]Line, int)
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
// Returns document AST, page count, cleaned watermarks, cleaned margins, detected encoding, and error.
func (p *PDFParser) Parse(ctx context.Context, r io.ReaderAt, size int64, opts Options) (*ast.Document, int, int, int, string, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return nil, 0, 0, 0, "Unknown", fmt.Errorf("failed to open pdf reader: %w", err)
	}

	numPages := reader.NumPage()
	if numPages == 0 {
		return ast.NewDocument(), 0, 0, 0, "UTF-8", nil
	}

	detectedEncoding := detectEncoding(reader)
	rawPages := make([]RawPage, numPages)
	var allFontSizes []float64
	totalWatermarksCleaned := 0

	for i := 1; i <= numPages; i++ {
		select {
		case <-ctx.Done():
			return nil, 0, 0, 0, detectedEncoding, ctx.Err()
		default:
		}

		page := reader.Page(i)
		content := page.Content()
		pageTexts := content.Text

		// Apply watermark pre-filter before font hierarchy analysis
		if opts.TextFilter != nil {
			var wCleaned int
			pageTexts, wCleaned = opts.TextFilter(pageTexts)
			totalWatermarksCleaned += wCleaned
		}

		rawPages[i-1] = RawPage{
			Number: i,
			Texts:  pageTexts,
		}

		// Only sample human-readable font sizes (>= 6pt) to avoid microprint watermarks skewing the modal body size
		for _, t := range pageTexts {
			if strings.TrimSpace(t.S) != "" && t.FontSize >= 6.0 {
				allFontSizes = append(allFontSizes, math.Round(t.FontSize*10)/10)
			}
		}
	}

	// Calculate font size clustering to determine heading thresholds
	fontHierarchy := computeFontHierarchy(allFontSizes)

	doc := ast.NewDocument()
	totalMarginsCleaned := 0

	for _, page := range rawPages {
		lines := groupIntoLines(page.Texts, page.Number)
		if opts.LineFilter != nil {
			var mCleaned int
			lines, mCleaned = opts.LineFilter(lines)
			totalMarginsCleaned += mCleaned
		}
		blockNodes := processLinesToBlocks(lines, fontHierarchy, opts)
		for _, node := range blockNodes {
			doc.Append(node)
		}
	}

	return doc, numPages, totalWatermarksCleaned, totalMarginsCleaned, detectedEncoding, nil
}

func detectEncoding(reader *pdf.Reader) string {
	if reader == nil || reader.NumPage() == 0 {
		return "UTF-8"
	}
	hasUnicode := false
	for i := 1; i <= reader.NumPage() && i <= 3; i++ {
		page := reader.Page(i)
		for _, fontName := range page.Fonts() {
			lower := strings.ToLower(fontName)
			if strings.Contains(lower, "identity") || strings.Contains(lower, "unicode") {
				hasUnicode = true
				break
			}
		}
		if hasUnicode {
			break
		}
	}
	if hasUnicode {
		return "UTF-8 (Identity-H / Unicode CMap)"
	}
	return "UTF-8 (WinAnsi / TrueType)"
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
		if t.S != "" {
			valid = append(valid, t)
		}
	}

	if len(valid) == 0 {
		return nil
	}

	// Sort primary by Y descending (top to bottom).
	// Within vertical line tolerance, preserve stable stream reading order!
	const yTolerance = 3.0
	sort.SliceStable(valid, func(i, j int) bool {
		diff := valid[i].Y - valid[j].Y
		if math.Abs(diff) > yTolerance {
			return valid[i].Y > valid[j].Y
		}
		return false
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

// assembleLine preserves natural stream reading order and combines elements into styled TextSpans.
func assembleLine(elements []pdf.Text, y float64, pageNum int) Line {
	// Preserve natural reading order from PDF stream.
	// Only sort if elements represent out-of-order columns separated by large distance (> 30pt).
	needsSort := false
	for i := 1; i < len(elements); i++ {
		if elements[i].X < elements[i-1].X-30.0 {
			needsSort = true
			break
		}
	}

	if needsSort {
		sort.SliceStable(elements, func(i, j int) bool {
			diff := elements[i].X - elements[j].X
			if math.Abs(diff) > 20.0 {
				return elements[i].X < elements[j].X
			}
			return false
		})
	}

	var spans []ast.TextSpan

	for i, elem := range elements {
		style := detectStyle(elem.Font)

		// Insert space if there is an explicit horizontal gap between separate words
		if i > 0 {
			prev := elements[i-1]
			if prev.W > 0 && elem.X > prev.X+prev.W+2.0 {
				if len(spans) > 0 && !strings.HasSuffix(spans[len(spans)-1].Text, " ") && !strings.HasPrefix(elem.S, " ") {
					spans[len(spans)-1].Text += " "
				}
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

	var tableRows [][]*ast.Node

	flushTable := func() {
		if len(tableRows) >= 2 {
			tableNode := &ast.Node{Type: ast.NodeTable}
			for _, rowCells := range tableRows {
				rowNode := &ast.Node{
					Type:     ast.NodeTableRow,
					Children: rowCells,
				}
				tableNode.Children = append(tableNode.Children, rowNode)
			}
			nodes = append(nodes, tableNode)
		} else if len(tableRows) == 1 {
			var spans []ast.TextSpan
			for idx, cell := range tableRows[0] {
				spans = append(spans, cell.Spans...)
				if idx < len(tableRows[0])-1 {
					spans = append(spans, ast.TextSpan{Text: " "})
				}
			}
			nodes = append(nodes, &ast.Node{
				Type:  ast.NodeParagraph,
				Spans: spans,
			})
		}
		tableRows = nil
	}

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
			flushTable()
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
			flushTable()
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
			flushTable()
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
			flushTable()
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
			flushTable()
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
			flushTable()
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

		// Table row check
		if opts.FormatTables {
			if cells, isRow := extractRowCells(line); isRow {
				flushList()
				flushCodeBlock()
				var cellNodes []*ast.Node
				for _, c := range cells {
					cellNodes = append(cellNodes, &ast.Node{
						Type:  ast.NodeTableCell,
						Spans: []ast.TextSpan{c},
					})
				}
				tableRows = append(tableRows, cellNodes)
				continue
			}
		}

		// Default: Paragraph
		flushTable()
		flushList()
		flushCodeBlock()
		nodes = append(nodes, &ast.Node{
			Type:  ast.NodeParagraph,
			Spans: line.Spans,
		})
	}

	flushTable()
	flushList()
	flushCodeBlock()

	return nodes
}

func getHeadingLevel(line Line, hierarchy fontHierarchy) int {
	if len(line.Spans) == 0 {
		return 0
	}

	var maxFontSize float64
	for _, span := range line.Spans {
		if span.FontSize > maxFontSize {
			maxFontSize = span.FontSize
		}
	}

	for idx, size := range hierarchy.headingSizes {
		if math.Abs(maxFontSize-size) <= 0.5 {
			return idx + 1
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

func extractRowCells(line Line) ([]ast.TextSpan, bool) {
	if len(line.Elements) < 2 {
		return nil, false
	}
	var cells []ast.TextSpan
	var currentCell strings.Builder
	var currentStyle ast.InlineStyle
	hasGap := false

	for i, elem := range line.Elements {
		if elem.S == "" {
			continue
		}
		if i > 0 {
			prev := line.Elements[i-1]
			gap := elem.X - (prev.X + prev.W)
			if gap >= 15.0 {
				hasGap = true
				if currentCell.Len() > 0 {
					cells = append(cells, ast.TextSpan{
						Text:  strings.TrimSpace(currentCell.String()),
						Style: currentStyle,
					})
					currentCell.Reset()
				}
			} else if prev.W > 0 && gap >= 2.5 {
				if currentCell.Len() > 0 && !strings.HasSuffix(currentCell.String(), " ") && !strings.HasPrefix(elem.S, " ") {
					currentCell.WriteString(" ")
				}
			}
		}
		currentCell.WriteString(elem.S)
		currentStyle = detectStyle(elem.Font)
	}

	if currentCell.Len() > 0 {
		cells = append(cells, ast.TextSpan{
			Text:  strings.TrimSpace(currentCell.String()),
			Style: currentStyle,
		})
	}

	return cells, hasGap && len(cells) >= 2
}
