// Package ast defines the abstract syntax tree for document elements extracted from PDFs.
package ast

// NodeType represents the structural classification of a document node.
type NodeType int

const (
	NodeDocument NodeType = iota
	NodeHeading
	NodeParagraph
	NodeCodeBlock
	NodeBlockquote
	NodeList
	NodeListItem
	NodeTable
	NodeTableRow
	NodeTableCell
	NodeImage
	NodeHorizontalRule
)

// InlineStyle flags typographic formatting for text spans.
type InlineStyle uint8

const (
	StyleNone InlineStyle = 0
	StyleBold InlineStyle = 1 << iota
	StyleItalic
	StyleCode
	StyleStrikethrough
)

// TextSpan represents an inline styled sequence of text.
type TextSpan struct {
	Text     string
	Style    InlineStyle
	FontSize float64
	FontName string
}

// Node represents a generic node in the document hierarchy.
type Node struct {
	Type     NodeType
	Level    int        // Heading level (1-6) or list nesting level
	Ordered  bool       // Ordered vs unordered list
	Spans    []TextSpan // Text contents for paragraphs/headings
	Children []*Node    // Sub-elements (list items, table rows, etc.)
	AltText  string     // For NodeImage
	URL      string     // For NodeImage link or path
}

// Document is the root node containing all page elements.
type Document struct {
	Title string
	Nodes []*Node
}

// NewDocument returns an empty Document AST.
func NewDocument() *Document {
	return &Document{
		Nodes: make([]*Node, 0),
	}
}

// Append adds a node to the document root.
func (d *Document) Append(n *Node) {
	d.Nodes = append(d.Nodes, n)
}
