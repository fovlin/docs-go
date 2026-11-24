// Package markdown parses the Markdown subset used by docs-go.
//
// Parsing and rendering are separate: Parse builds a tree of blocks that hold
// their parsed content, and each block renders itself to HTML when
// Document.Render is called. Parsing is intentionally forgiving: unknown or
// malformed syntax degrades to text.
package markdown

import (
	"strconv"
	"strings"
	"unicode"
)

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

// block is one parsed block. Rendering writes into the renderer's builder.
type block interface {
	render(*renderer)
}

type parser struct {
	usedHeadingIDs map[string]bool
	headings       []Heading
	firstH1        string
	firstParagraph string
}

func (p *parser) uniqueHeadingID(text string) string {
	base := slugifyHeading(text)
	if base == "" {
		base = "section"
	}

	id := base
	for count := 1; p.usedHeadingIDs[id]; count++ {
		id = base + "-" + strconv.Itoa(count)
	}
	p.usedHeadingIDs[id] = true

	return id
}

func slugifyHeading(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))

	var builder strings.Builder
	lastHyphen := false

	for _, char := range value {
		switch {
		case unicode.IsLetter(char) || unicode.IsDigit(char):
			builder.WriteRune(char)
			lastHyphen = false
		case char == ' ' || char == '-' || char == '_':
			if builder.Len() > 0 && !lastHyphen {
				builder.WriteByte('-')
				lastHyphen = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

// Parse converts Markdown source into a document. It never returns an error:
// malformed constructs degrade to text, which is the useful behaviour for a
// documentation generator.
func Parse(source []byte) *Document {
	text := strings.TrimPrefix(string(source), "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	p := &parser{usedHeadingIDs: make(map[string]bool)}
	blocks := p.parseBlocks(strings.Split(text, "\n"))

	return &Document{
		blocks:         blocks,
		Headings:       p.headings,
		FirstH1:        p.firstH1,
		FirstParagraph: p.firstParagraph,
	}
}
