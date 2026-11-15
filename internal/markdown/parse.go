// Package markdown parses the Markdown subset used by docs-go.
//
// It is intentionally forgiving: unknown or malformed syntax degrades to
// text. Blocks, inline spans and rendering are separate passes, so a new
// construct stays local: one recognizer, one block kind, one render case.
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
// private; callers use Render, RewriteLinks and the exported metadata fields.
type Document struct {
	blocks []block

	Headings       []Heading
	FirstH1        string
	FirstParagraph string
	PlainText      string
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
// start, tight, children. List item: children. Table row: header, children.
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
	tight   bool

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
		PlainText:      collectPlainText(blocks),
	}
}

func (p *parser) parseBlocks(lines []string) []block {
	blocks := make([]block, 0, len(lines))

	for i := 0; i < len(lines); {
		line := lines[i]

		if isBlank(line) {
			i++
			continue
		}

		if char, length, info, ok := fenceStart(line); ok {
			code, language, next := parseFence(lines, i, char, length, info)
			blocks = append(blocks, block{
				kind:     blockCode,
				text:     code,
				language: language,
			})
			i = next
			continue
		}

		if indentWidth(line) >= 4 {
			code, next := parseIndentedCode(lines, i)
			blocks = append(blocks, block{kind: blockCode, text: code})
			i = next
			continue
		}

		if level, text, ok := headingStart(line); ok {
			inlines := parseInline(text)
			id := p.uniqueHeadingID(plainInlines(inlines))
			headingText := plainInlines(inlines)

			p.headings = append(p.headings, Heading{
				Level: level,
				Text:  headingText,
				ID:    id,
			})

			if level == 1 && p.firstH1 == "" {
				p.firstH1 = headingText
			}

			blocks = append(blocks, block{
				kind:    blockHeading,
				level:   level,
				id:      id,
				inlines: inlines,
			})
			i++
			continue
		}

		if thematicBreak(line) {
			blocks = append(blocks, block{kind: blockThematicBreak})
			i++
			continue
		}

		if blockquoteStart(line) {
			inner, next := p.collectBlockquote(lines, i)
			blocks = append(blocks, block{
				kind:     blockBlockquote,
				children: p.parseBlocks(inner),
			})
			i = next
			continue
		}

		if marker, ok := parseListMarker(line); ok {
			list, next := p.parseList(lines, i, marker)
			blocks = append(blocks, list)
			i = next
			continue
		}

		if kind, title, ok := containerStart(line); ok {
			inner, next := collectContainer(lines, i)
			blocks = append(blocks, block{
				kind:           blockContainer,
				containerKind:  kind,
				containerTitle: title,
				children:       p.parseBlocks(inner),
			})
			i = next
			continue
		}

		if isTableStart(lines, i) {
			table, next := p.parseTable(lines, i)
			blocks = append(blocks, table)
			i = next
			continue
		}

		start := i
		i++

		for i < len(lines) && !isBlank(lines[i]) && !p.interruptsParagraph(lines, i) {
			i++
		}

		blocks = append(blocks, block{
			kind:    blockParagraph,
			inlines: parseInline(strings.Join(lines[start:i], "\n")),
		})
	}

	return blocks
}

// startsBlock includes table starts, which need the next line to detect.
func (p *parser) startsBlock(lines []string, index int) bool {
	return startsBlockLine(lines[index]) || isTableStart(lines, index)
}

// interruptsParagraph is startsBlock with the CommonMark exception that an
// ordered list not starting at 1 does not split a paragraph.
func (p *parser) interruptsParagraph(lines []string, index int) bool {
	if marker, ok := parseListMarker(lines[index]); ok && marker.ordered && marker.start != 1 {
		return false
	}

	return p.startsBlock(lines, index)
}

func parseFence(lines []string, start int, char byte, length int, info string) (string, string, int) {
	language := ""
	if fields := strings.Fields(info); len(fields) > 0 {
		language = strings.ToLower(fields[0])
	}

	content := make([]string, 0, 8)
	indent := blockIndent(lines[start])
	i := start + 1

	for i < len(lines) {
		if fenceClose(lines[i], char, length) {
			i++
			break
		}

		content = append(content, stripIndent(lines[i], indent))
		i++
	}

	code := strings.Join(content, "\n")
	if code != "" {
		code += "\n"
	}

	return code, language, i
}

func parseIndentedCode(lines []string, start int) (string, int) {
	content := make([]string, 0, 8)
	i := start

	for i < len(lines) {
		if isBlank(lines[i]) {
			content = append(content, "")
			i++
			continue
		}

		if indentWidth(lines[i]) < 4 {
			break
		}

		content = append(content, stripIndent(lines[i], 4))
		i++
	}

	code := strings.Join(trimTrailingBlank(content), "\n")
	if code != "" {
		code += "\n"
	}

	return code, i
}

func (p *parser) collectBlockquote(lines []string, start int) ([]string, int) {
	inner := make([]string, 0, 8)
	i := start

	for i < len(lines) {
		line := lines[i]

		if isBlank(line) {
			break
		}

		if blockquoteStart(line) {
			inner = append(inner, stripBlockquote(line))
			i++
			continue
		}

		// Lazy continuation: a plain line directly after a quoted line still
		// belongs to the quote. A new block marker ends it.
		if !p.interruptsParagraph(lines, i) {
			inner = append(inner, line)
			i++
			continue
		}

		break
	}

	return trimTrailingBlank(inner), i
}

func (p *parser) parseList(lines []string, start int, first listMarker) (block, int) {
	list := block{
		kind:    blockList,
		ordered: first.ordered,
		start:   first.start,
		tight:   true,
	}

	i := start
	baseIndent := first.indent

	for i < len(lines) {
		marker, ok := parseListMarker(lines[i])
		if !ok || marker.indent < baseIndent || !sameListMarker(marker, first) {
			break
		}

		empty := marker.content == ""
		itemLines := []string{marker.content}
		i++

		atItemLevel := func(m listMarker) bool {
			return m.indent >= baseIndent && m.indent < marker.contentIndent
		}

		for i < len(lines) {
			line := lines[i]

			if isBlank(line) {
				next := i + 1
				for next < len(lines) && isBlank(lines[next]) {
					next++
				}

				if next >= len(lines) {
					i = next
					break
				}

				if nextMarker, ok := parseListMarker(lines[next]); ok &&
					atItemLevel(nextMarker) && sameListMarker(nextMarker, first) {
					list.tight = false
					i = next
					break
				}

				if empty {
					list.tight = false
					i = next
					break
				}

				if indentWidth(lines[next]) < marker.contentIndent {
					i = next
					break
				}

				itemLines = append(itemLines, "")
				list.tight = false
				i = next
				continue
			}

			if nextMarker, ok := parseListMarker(line); ok && atItemLevel(nextMarker) {
				break
			}

			if indentWidth(line) < marker.contentIndent && p.startsBlock(lines, i) {
				break
			}

			if empty && indentWidth(line) < marker.contentIndent {
				break
			}

			itemLines = append(itemLines, stripIndent(line, marker.contentIndent))
			i++
		}

		list.children = append(list.children, block{
			kind:     blockListItem,
			children: p.parseBlocks(trimTrailingBlank(itemLines)),
		})
	}

	return list, i
}

func (p *parser) parseTable(lines []string, start int) (block, int) {
	aligns, _ := tableDelimiter(lines[start+1])
	if len(aligns) == 0 {
		aligns = []string{""}
	}

	table := block{
		kind:     blockTable,
		children: []block{tableRow(splitTableRow(lines[start]), aligns, true)},
	}

	i := start + 2
	for i < len(lines) {
		if !strings.Contains(lines[i], "|") || startsBlockLine(lines[i]) {
			break
		}

		table.children = append(table.children, tableRow(splitTableRow(lines[i]), aligns, false))
		i++
	}

	return table, i
}

func tableRow(cells, aligns []string, header bool) block {
	row := block{kind: blockTableRow, header: header}

	for index, align := range aligns {
		cell := ""
		if index < len(cells) {
			cell = cells[index]
		}

		row.children = append(row.children, block{
			kind:    blockTableCell,
			align:   align,
			inlines: parseInline(cell),
		})
	}

	return row
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

func plainInlines(inlines []inline) string {
	var builder strings.Builder

	var walk func([]inline)
	walk = func(items []inline) {
		for _, item := range items {
			switch item.kind {
			case inlineText, inlineCode:
				builder.WriteString(item.text)
			case inlineSoftBreak, inlineHardBreak:
				builder.WriteByte(' ')
			default:
				walk(item.children)
			}
		}
	}

	walk(inlines)

	return strings.Join(strings.Fields(builder.String()), " ")
}

func firstParagraphText(blocks []block) string {
	var walk func([]block) string

	walk = func(items []block) string {
		for _, item := range items {
			switch item.kind {
			case blockParagraph:
				return plainInlines(item.inlines)
			case blockBlockquote, blockList, blockListItem, blockContainer:
				if text := walk(item.children); text != "" {
					return text
				}
			}
		}

		return ""
	}

	return walk(blocks)
}

func collectPlainText(blocks []block) string {
	var builder strings.Builder

	var walk func([]block)
	walk = func(items []block) {
		for _, item := range items {
			switch item.kind {
			case blockHeading, blockParagraph, blockTableCell:
				builder.WriteString(plainInlines(item.inlines))
				builder.WriteByte('\n')
			case blockCode:
				builder.WriteString(item.text)
				builder.WriteByte('\n')
			default:
				walk(item.children)
			}
		}
	}

	walk(blocks)

	return strings.Join(strings.Fields(builder.String()), " ")
}

type listMarker struct {
	ordered       bool
	marker        byte
	start         int
	indent        int
	contentIndent int
	content       string
}

func sameListMarker(a, b listMarker) bool {
	return a.ordered == b.ordered && a.marker == b.marker
}

func parseListMarker(line string) (listMarker, bool) {
	indent := blockIndent(line)
	if indent > 3 || indent >= len(line) {
		return listMarker{}, false
	}

	index := indent

	switch line[index] {
	case '-', '*', '+':
		next := index + 1
		if next + 1 < len(line) && line[next] != ' ' && line[next] != '\t' {
			return listMarker{}, false
		}

		next = skipSpaces(line, next)

		return listMarker{
			marker:        line[index],
			indent:        indent,
			contentIndent: next,
			content:       line[next:],
		}, true
	}

	if line[index] < '0' || line[index] > '9' {
		return listMarker{}, false
	}

	end := index
	for end < len(line) && end-index < 9 && line[end] >= '0' && line[end] <= '9' {
		end++
	}

	if end >= len(line) || (line[end] != '.' && line[end] != ')') {
		return listMarker{}, false
	}

	start, err := strconv.Atoi(line[index:end])
	if err != nil {
		return listMarker{}, false
	}

	marker := line[end]
	end++
	if end < len(line) && line[end] != ' ' && line[end] != '\t' {
		return listMarker{}, false
	}

	end = skipSpaces(line, end)

	return listMarker{
		ordered:       true,
		marker:        marker,
		start:         start,
		indent:        indent,
		contentIndent: end,
		content:       line[end:],
	}, true
}

// parse code block
func fenceStart(line string) (byte, int, string, bool) {
	indent := blockIndent(line)
	if indent > 3 || indent >= len(line) {
		return 0, 0, "", false
	}

	char := line[indent]
	if char != '`' && char != '~' {
		return 0, 0, "", false
	}

	end := indent
	for end < len(line) && line[end] == char {
		end++
	}

	length := end - indent
	if length < 3 {
		return 0, 0, "", false
	}

	info := strings.TrimSpace(line[end:])
	if char == '`' && strings.Contains(info, "`") {
		return 0, 0, "", false
	}

	return char, length, info, true
}

func fenceClose(line string, char byte, length int) bool {
	indent := blockIndent(line)
	if indent > 3 || indent >= len(line) || line[indent] != char {
		return false
	}

	end := indent
	for end < len(line) && line[end] == char {
		end++
	}

	if end-indent < length {
		return false
	}

	return strings.TrimSpace(line[end:]) == ""
}

func headingStart(line string) (int, string, bool) {
	indent := blockIndent(line)
	if indent > 3 || indent >= len(line) {
		return 0, "", false
	}

	end := indent
	for end < len(line) && line[end] == '#' {
		end++
	}

	level := end - indent
	if level < 1 || level > 6 {
		return 0, "", false
	}

	if end < len(line) && line[end] != ' ' && line[end] != '\t' {
		return 0, "", false
	}

	text := strings.TrimSpace(line[end:])

	if trimmed := strings.TrimRight(text, "#"); len(trimmed) < len(text) {
		if trimmed == "" || strings.HasSuffix(trimmed, " ") || strings.HasSuffix(trimmed, "\t") {
			text = strings.TrimSpace(trimmed)
		}
	}

	return level, text, true
}

func thematicBreak(line string) bool {
	index := blockIndent(line)
	if index > 3 {
		return false
	}

	var char byte
	count := 0

	for ; index < len(line); index++ {
		value := line[index]
		if value == ' ' || value == '\t' {
			continue
		}

		if char == 0 {
			if value != '*' && value != '-' && value != '_' {
				return false
			}
			char = value
		}

		if value != char {
			return false
		}

		count++
	}

	return count >= 3
}

func blockquoteStart(line string) bool {
	index := blockIndent(line)

	return index <= 3 && index < len(line) && line[index] == '>'
}

func stripBlockquote(line string) string {
	index := blockIndent(line)
	if index > 3 || index >= len(line) || line[index] != '>' {
		return line
	}

	index++
	if index < len(line) && line[index] == ' ' {
		index++
	}

	return line[index:]
}

var containerTitles = map[string]string{
	"tip": "TIP", "info": "INFO", "note": "NOTE",
	"warning": "WARNING", "danger": "DANGER", "details": "Details",
}

func containerStart(line string) (string, string, bool) {
	indent := blockIndent(line)
	if indent > 3 || !strings.HasPrefix(line[indent:], ":::") {
		return "", "", false
	}

	rest := strings.TrimSpace(line[indent+3:])
	if rest == "" {
		return "", "", false
	}

	fields := strings.Fields(rest)
	name := strings.ToLower(fields[0])
	title, ok := containerTitles[name]
	if !ok {
		return "", "", false
	}

	if custom := strings.TrimSpace(strings.TrimPrefix(rest, fields[0])); custom != "" {
		title = custom
	}

	return name, title, true
}

func collectContainer(lines []string, start int) ([]string, int) {
	inner := make([]string, 0, 16)
	depth := 1
	var fenceChar byte
	var fenceLength int

	i := start + 1
	for i < len(lines) {
		line := lines[i]

		if fenceChar != 0 {
			inner = append(inner, line)
			if fenceClose(line, fenceChar, fenceLength) {
				fenceChar = 0
			}
			i++
			continue
		}

		if char, length, _, ok := fenceStart(line); ok {
			fenceChar, fenceLength = char, length
			inner = append(inner, line)
			i++
			continue
		}

		if isContainerClose(line) {
			depth--
			if depth == 0 {
				return trimTrailingBlank(inner), i + 1
			}

			inner = append(inner, line)
			i++
			continue
		}

		if _, _, ok := containerStart(line); ok {
			depth++
		}

		inner = append(inner, line)
		i++
	}

	return trimTrailingBlank(inner), i
}

func isContainerClose(line string) bool {
	return indentWidth(line) <= 3 && strings.TrimSpace(line) == ":::"
}

func startsBlockLine(line string) bool {
	if isBlank(line) || thematicBreak(line) {
		return true
	}

	if _, _, ok := headingStart(line); ok {
		return true
	}

	if _, _, _, ok := fenceStart(line); ok {
		return true
	}

	if blockquoteStart(line) {
		return true
	}

	if _, ok := parseListMarker(line); ok {
		return true
	}

	if _, _, ok := containerStart(line); ok {
		return true
	}

	return false
}

func tableDelimiter(line string) ([]string, bool) {
	if !strings.Contains(line, "|") {
		return nil, false
	}

	cells := splitTableRow(line)
	if len(cells) == 0 {
		return nil, false
	}

	aligns := make([]string, len(cells))

	for index, cell := range cells {
		cell = strings.TrimSpace(cell)
		if len(cell) < 3 {
			return nil, false
		}

		left := strings.HasPrefix(cell, ":")
		right := strings.HasSuffix(cell, ":")
		core := strings.Trim(cell, ":")

		if len(core) < 3 || strings.Trim(core, "-") != "" {
			return nil, false
		}

		switch {
		case left && right:
			aligns[index] = "center"
		case left:
			aligns[index] = "left"
		case right:
			aligns[index] = "right"
		}
	}

	return aligns, true
}

func isTableStart(lines []string, index int) bool {
	if index+1 >= len(lines) || !strings.Contains(lines[index], "|") {
		return false
	}

	aligns, ok := tableDelimiter(lines[index+1])

	return ok && len(aligns) == len(splitTableRow(lines[index]))
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)

	line = strings.TrimPrefix(line, "|")

	if strings.HasSuffix(line, "|") && !strings.HasSuffix(line, "\\|") {
		line = line[:len(line)-1]
	}

	cells := make([]string, 0, 4)
	var builder strings.Builder
	inCode := 0

	for index := 0; index < len(line); index++ {
		value := line[index]

		if value == '\\' && index+1 < len(line) {
			if line[index+1] == '|' {
				builder.WriteByte('|')
			} else {
				builder.WriteByte('\\')
				builder.WriteByte(line[index+1])
			}
			index++
			continue
		}

		if value == '`' {
			run := 1
			for index+run < len(line) && line[index+run] == '`' {
				run++
			}

			if inCode == 0 {
				inCode = run
			} else if inCode == run {
				inCode = 0
			}

			builder.WriteString(line[index : index+run])
			index += run - 1
			continue
		}

		if value == '|' && inCode == 0 {
			cells = append(cells, strings.TrimSpace(builder.String()))
			builder.Reset()
			continue
		}

		builder.WriteByte(value)
	}

	cells = append(cells, strings.TrimSpace(builder.String()))

	return cells
}

func isBlank(line string) bool {
	return strings.TrimSpace(line) == ""
}

// blockIndent returns the leading-space width of a block marker, capped at 4.
// Callers use it to reject indentation beyond three spaces; it is not a
// general-purpose indentation counter.
func blockIndent(line string) int {
	index := 0
	for index < len(line) && index < 4 && line[index] == ' ' {
		index++
	}

	return index
}

func indentWidth(line string) int {
	width := 0

	for _, char := range line {
		switch char {
		case ' ':
			width++
		case '\t':
			width += 4 - width%4
		default:
			return width
		}
	}

	return width
}

func stripIndent(line string, width int) string {
	consumed := 0

	for index, char := range line {
		if consumed >= width {
			return line[index:]
		}

		switch char {
		case ' ':
			consumed++
		case '\t':
			consumed += 4 - consumed%4
		default:
			return line[index:]
		}
	}

	return ""
}

func skipSpaces(line string, index int) int {
	for index < len(line) && (line[index] == ' ' || line[index] == '\t') {
		index++
	}

	return index
}

func trimTrailingBlank(lines []string) []string {
	for len(lines) > 0 && isBlank(lines[len(lines)-1]) {
		lines = lines[:len(lines)-1]
	}

	return lines
}
