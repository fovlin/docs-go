// Package markdown parses the Markdown subset used by docs-go.
//
// It is intentionally forgiving: unknown or malformed syntax degrades to
// text. Blocks, inline spans and rendering are separate passes.
package markdown

import "strings"

// Heading is one heading in a document, in source order.
type Heading struct {
	Level int
	Text  string
	ID    string
}

// Document is the parsed form of one Markdown file. The block tree is kept
// private; callers use Render and the exported metadata fields.
type Document struct {
	blocks []block

	Headings       []Heading
	FirstH1        string
	FirstParagraph string
}

type blockKind int

const (
	blockHeading blockKind = iota
	blockParagraph
	blockCode
	blockBlockquote
	blockList
	blockListItem
	blockTable
	blockTableRow
	blockTableCell
	blockThematicBreak
	blockContainer
)

// block is a small union: each kind uses only some fields. Heading: level,
// id, inlines. Paragraph: inlines. Code: text, language. List: ordered,
// start. List item: children. Table row: header, children.
// Table cell: align, inlines. Blockquote: children. Container:
// containerKind, containerTitle, children.
type block struct {
	kind blockKind

	level    int
	id       string
	inlines  []inline
	children []block

	text     string
	language string

	ordered bool
	start   int

	header bool
	align  string

	containerKind  string
	containerTitle string
}

type parser struct {
	usedHeadingIDs map[string]bool
	headings       []Heading
	firstH1        string
}

// Parse converts Markdown source into a document. It never returns an error:
// malformed constructs degrade to text, which is the useful behaviour for a
// documentation generator.
func Parse(source []byte) *Document {
	text := strings.TrimPrefix(string(source), "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	p := &parser{
		usedHeadingIDs: make(map[string]bool),
	}

	blocks := p.parseBlocks(strings.Split(text, "\n"))

	return &Document{
		blocks:         blocks,
		Headings:       p.headings,
		FirstH1:        p.firstH1,
		FirstParagraph: firstParagraphText(blocks),
	}
}
