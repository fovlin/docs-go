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
		PlainText:      collectPlainText(blocks),
	}
}

func (p *parser) parseBlocks(lines []string) []block {
	specs := scanBlocks(lines)
	blocks := make([]block, 0, len(specs))

	for _, spec := range specs {
		blocks = append(blocks, p.parseBlock(lines, spec))
	}

	return blocks
}

func (p *parser) parseBlock(lines []string, spec blockSpec) block {
	switch spec.kind {
	case blockCode:
		if spec.fenced {
			code, language := parseFenceRange(lines, spec.start, spec.end, spec.fenceChar, spec.fenceLength, spec.fenceInfo)
			return block{kind: blockCode, text: code, language: language}
		}
		return block{kind: blockCode, text: parseIndentedCodeRange(lines, spec.start, spec.end)}

	case blockHeading:
		inlines := parseInline(spec.headingText)
		id := p.uniqueHeadingID(plainInlines(inlines))
		headingText := plainInlines(inlines)

		p.headings = append(p.headings, Heading{Level: spec.headingLevel, Text: headingText, ID: id})
		if spec.headingLevel == 1 && p.firstH1 == "" {
			p.firstH1 = headingText
		}

		return block{kind: blockHeading, level: spec.headingLevel, id: id, inlines: inlines}

	case blockThematicBreak:
		return block{kind: blockThematicBreak}

	case blockBlockquote:
		inner := make([]string, 0, spec.end-spec.start)
		for j := spec.start; j < spec.end; j++ {
			inner = append(inner, stripBlockquote(lines[j]))
		}
		return block{kind: blockBlockquote, children: p.parseBlocks(trimTrailingBlank(inner))}

	case blockList:
		return p.parseListRange(lines, spec.end, spec.listItems)

	case blockContainer:
		return block{
			kind:           blockContainer,
			containerKind:  spec.containerKind,
			containerTitle: spec.containerTitle,
			children:       p.parseBlocks(trimTrailingBlank(lines[spec.contentStart:spec.contentEnd])),
		}

	case blockTable:
		return parseTableSpec(spec)
	}

	return block{
		kind:    blockParagraph,
		inlines: parseInline(strings.Join(lines[spec.start:spec.end], "\n")),
	}
}

func parseFenceRange(lines []string, start, end int, char byte, length int, info string) (string, string) {
	language := ""
	if fields := strings.Fields(info); len(fields) > 0 {
		language = strings.ToLower(fields[0])
	}

	contentEnd := end
	if contentEnd > start+1 && fenceClose(lines[contentEnd-1], char, length) {
		contentEnd--
	} else if contentEnd == len(lines) && contentEnd > start+1 && lines[contentEnd-1] == "" {
		contentEnd--
	}

	indent := blockIndent(lines[start])
	content := make([]string, 0, contentEnd-start)
	for i := start + 1; i < contentEnd; i++ {
		content = append(content, stripIndent(lines[i], indent))
	}

	code := strings.Join(content, "\n")
	if code != "" {
		code += "\n"
	}

	return code, language
}

func parseIndentedCodeRange(lines []string, start, end int) string {
	content := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			content = append(content, "")
			continue
		}
		content = append(content, stripIndent(lines[i], 4))
	}

	code := strings.Join(trimTrailingBlank(content), "\n")
	if code != "" {
		code += "\n"
	}

	return code
}

func (p *parser) parseListRange(lines []string, end int, items []listItemSpec) block {
	first := items[0].marker
	list := block{
		kind:    blockList,
		ordered: first.ordered,
		start:   first.start,
	}

	for _, item := range items {
		itemLines := []string{item.marker.content}
		for i := item.start + 1; i < item.end; i++ {
			if strings.TrimSpace(lines[i]) == "" {
				continue
			}
			itemLines = append(itemLines, stripIndent(lines[i], item.marker.indent+2))
		}

		list.children = append(list.children, block{
			kind:     blockListItem,
			children: p.parseBlocks(trimTrailingBlank(itemLines)),
		})
	}

	return list
}

func parseTableSpec(spec blockSpec) block {
	table := block{kind: blockTable}
	for _, row := range spec.tableRows {
		table.children = append(table.children, tableRow(row.cells, spec.tableAligns, row.header))
	}
	return table
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
	ordered bool
	marker  byte
	start   int
	indent  int
	content string
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
		if next < len(line) && line[next] != ' ' && line[next] != '\t' {
			return listMarker{}, false
		}
		next++
		if next >= len(line) {
			return listMarker{}, false
		}

		return listMarker{
			marker:  line[index],
			indent:  indent,
			content: line[next:],
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
	if end >= len(line) {
		return listMarker{}, false
	}

	return listMarker{
		ordered: true,
		marker:  marker,
		start:   start,
		indent:  indent,
		content: line[end:],
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
	if index < len(line) && (line[index] == ' ' || line[index] == '\t') {
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

func isContainerClose(line string) bool {
	return indentWidth(line) <= 3 && strings.TrimSpace(line) == ":::"
}

func collectContainer(lines []string, start int) ([]string, int, bool) {
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

		// is container close?
		if indentWidth(line) <= 3 && strings.TrimSpace(line) == ":::" {
			depth--
			if depth == 0 {
				return trimTrailingBlank(inner), i + 1, true
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

	return trimTrailingBlank(inner), i, false
}

func tableDelimiter(line string) ([]string, bool) {
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
	if index+1 >= len(lines) || !isTableRaw(lines[index]) {
		return false
	}

	aligns, ok := tableDelimiter(lines[index+1])

	return ok && len(aligns) == len(splitTableRow(lines[index]))
}

func isTableRaw(line string) bool {
	if !strings.ContainsRune(line, '|') {
		return false
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	switch trimmed[0] {
	case '#', '>', '+', '*':
		return false
	case '-':
		return len(trimmed) == 1 || (trimmed[1] != ' ' && trimmed[1] != '\t')
	}

	return true
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)

	line = strings.Trim(line, "|")

	cells := make([]string, 0, 4)
	var builder strings.Builder
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

		if value == '|' {
			cells = append(cells, strings.TrimSpace(builder.String()))
			builder.Reset()
			continue
		}

		builder.WriteByte(value)
	}

	cells = append(cells, strings.TrimSpace(builder.String()))

	return cells
}

func skipBlank(lines []string, i int) (int, bool) {
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "" {
			i++
			continue
		}
		break
	}

	return i, i >= len(lines)
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
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}
