package markdown

import "strings"

// parseBlocks parses source lines into blocks; blank lines only separate them.
func (p *parser) parseBlocks(lines []string) []block {
	blocks, _ := p.parseBlocksUntil(lines, nil)
	return blocks
}

// parseBlocksUntil parses blocks until stop matches an unconsumed line and
// returns that line's index. A nil stop parses to the end.
func (p *parser) parseBlocksUntil(lines []string, stop func(string) bool) ([]block, int) {
	blocks := make([]block, 0, 8)

	index := 0
	for index < len(lines) {
		if strings.TrimSpace(lines[index]) == "" {
			index++
			continue
		}
		if stop != nil && stop(lines[index]) {
			return blocks, index
		}
		item, next := p.parseBlock(lines, index, stop)
		blocks = append(blocks, item)
		index = next
	}

	return blocks, index
}

// parseBlock parses the block starting at start and returns the first line
// after it. Each parser re-recognises its own opener, which keeps the
// construct's metadata and its parser in one place. stop only constrains
// paragraph continuation; other parsers consume their own content.
func (p *parser) parseBlock(lines []string, start int, stop func(string) bool) (block, int) {
	switch blockKindAt(lines, start) {
	case blockFence:
		return fencedCode(lines, start)
	case blockIndentedCode:
		return indentedCode(lines, start)
	case blockHeading:
		return p.heading(lines[start]), start + 1
	case blockThematicBreak:
		return thematicBreakBlock{}, start + 1
	case blockBlockquote:
		return p.blockquote(lines, start)
	case blockList:
		return p.list(lines, start)
	case blockContainer:
		return p.container(lines, start)
	case blockTable:
		return table(lines, start)
	default:
		return p.paragraph(lines, start, stop)
	}
}

type headingBlock struct {
	level   int
	id      string
	inlines []inline
}

func (p *parser) heading(line string) block {
	level, text, _ := headingStart(line)
	inlines := parseInline(text)
	plain := plainInlines(inlines)
	id := p.uniqueHeadingID(plain)

	p.headings = append(p.headings, Heading{Level: level, Text: plain, ID: id})
	if level == 1 && p.firstH1 == "" {
		p.firstH1 = plain
	}
	return &headingBlock{level: level, id: id, inlines: inlines}
}

type paragraphBlock struct {
	inlines []inline
}

func (p *parser) paragraph(lines []string, start int, stop func(string) bool) (block, int) {
	end := start
	for end+1 < len(lines) && strings.TrimSpace(lines[end+1]) != "" {
		if stop != nil && stop(lines[end+1]) {
			break
		}
		if blockKindAt(lines, end+1) != blockParagraph {
			break
		}
		end++
	}

	inlines := parseInline(strings.Join(lines[start:end+1], "\n"))
	if p.firstParagraph == "" {
		if text := plainInlines(inlines); text != "" {
			p.firstParagraph = text
		}
	}
	return &paragraphBlock{inlines: inlines}, end + 1
}

type codeBlock struct {
	text     string
	language string
}

func fencedCode(lines []string, start int) (block, int) {
	char, length, info, _ := fenceStart(lines[start])
	end, closed := fenceEnd(lines, start+1, char, length)

	indent, _ := leadingSpace(lines[start])
	content := make([]string, 0, end-start-1)
	for index := start + 1; index < end; index++ {
		content = append(content, stripIndent(lines[index], indent))
	}

	// Splitting the source on newlines leaves an empty trailing line; it is
	// not part of an unclosed fence's content.
	if !closed && len(content) > 0 && content[len(content)-1] == "" {
		content = content[:len(content)-1]
	}

	var language string
	if fields := strings.Fields(info); len(fields) > 0 {
		language = strings.ToLower(fields[0])
	}

	next := end
	if closed {
		next++
	}
	return &codeBlock{text: joinCode(content), language: language}, next
}

func indentedCode(lines []string, start int) (block, int) {
	end := start
	for end < len(lines) {
		width, _ := leadingSpace(lines[end])
		if strings.TrimSpace(lines[end]) != "" && width < 4 {
			break
		}
		end++
	}

	// Blank lines that trail the code belong to whatever follows it.
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	content := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		if strings.TrimSpace(line) == "" {
			content = append(content, "")
			continue
		}
		content = append(content, stripIndent(line, 4))
	}
	return &codeBlock{text: joinCode(content)}, end
}

func joinCode(lines []string) string {
	code := strings.Join(lines, "\n")
	if code != "" {
		code += "\n"
	}
	return code
}

type thematicBreakBlock struct{}

type blockquoteBlock struct {
	children []block
}

func (p *parser) blockquote(lines []string, start int) (block, int) {
	inner := make([]string, 0, 4)

	end := start
	for ; end < len(lines) && blockquoteStart(lines[end]); end++ {
		inner = append(inner, stripBlockquote(lines[end]))
	}
	return &blockquoteBlock{children: p.parseBlocks(inner)}, end
}

type listBlock struct {
	ordered bool
	start   int
	loose   bool

	items []*listItem
}

type listItem struct {
	children []block
	loose    bool
}

func (p *parser) list(lines []string, start int) (block, int) {
	first, _ := parseListMarker(lines[start])
	list := &listBlock{ordered: first.ordered, start: first.start}

	index := start
	for index < len(lines) {
		marker, ok := parseListMarker(lines[index])
		if !ok || !marker.continues(first) {
			break
		}
		item, next := p.listItem(lines, index, marker)
		list.items = append(list.items, item)
		list.loose = list.loose || item.loose
		index = next

		blankStart := index
		for index < len(lines) && strings.TrimSpace(lines[index]) == "" {
			index++
		}
		if index > blankStart && index < len(lines) {
			if nextMarker, ok := parseListMarker(lines[index]); ok && nextMarker.continues(first) {
				list.loose = true
			}
		}
	}
	return list, index
}

func (p *parser) listItem(lines []string, start int, marker listMarker) (*listItem, int) {
	end := start
	loose := false
	for next := start + 1; next < len(lines); next++ {
		if strings.TrimSpace(lines[next]) == "" {
			continue
		}
		if width, _ := leadingSpace(lines[next]); width < marker.contentIndent {
			break
		}
		if strings.TrimSpace(lines[next-1]) == "" {
			loose = true
		}
		end = next
	}

	inner := make([]string, 0, end-start+2)
	inner = append(inner, marker.content)
	for _, line := range lines[start+1 : end+1] {
		if strings.TrimSpace(line) == "" {
			inner = append(inner, "")
			continue
		}
		inner = append(inner, stripIndent(line, marker.contentIndent))
	}
	return &listItem{children: p.parseBlocks(inner), loose: loose}, end + 1
}

type tableBlock struct {
	rows []tableRow
}

type tableRow struct {
	cells []tableCell
}

type tableCell struct {
	align   string
	inlines []inline
}

func table(lines []string, start int) (block, int) {
	header, aligns, _ := tableStart(lines, start)
	end := start + 1
	for end+1 < len(lines) && isTableRow(lines[end+1]) {
		end++
	}

	rows := make([]tableRow, 0, end-start)
	rows = append(rows, newTableRow(header, aligns))
	for index := start + 2; index <= end; index++ {
		rows = append(rows, newTableRow(splitTableRow(lines[index]), aligns))
	}
	return &tableBlock{rows: rows}, end + 1
}

func newTableRow(cells, aligns []string) tableRow {
	row := tableRow{cells: make([]tableCell, len(aligns))}
	for index, align := range aligns {
		text := ""
		if index < len(cells) {
			text = cells[index]
		}
		row.cells[index] = tableCell{align: align, inlines: parseInline(text)}
	}
	return row
}

type containerBlock struct {
	kind     string
	title    string
	children []block
}

func (p *parser) container(lines []string, start int) (block, int) {
	kind, title, _ := containerStart(lines[start])

	children, consumed := p.parseBlocksUntil(lines[start+1:], isContainerClose)
	next := start + 1 + consumed
	if next < len(lines) && isContainerClose(lines[next]) {
		next++
	}

	return &containerBlock{kind: kind, title: title, children: children}, next
}
