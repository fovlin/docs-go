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

// Options controls the small number of site-specific Markdown extensions.
type Options struct {
	// Containers maps a ::: block name to its default title, for example
	// {"tip": "TIP"}. Unknown names fall back to "tip".
	Containers map[string]string
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
	blockHTML
)

// block is a small union: each kind uses only some fields. Heading: level,
// id, inlines. Paragraph: inlines. Code: text, language. List: ordered,
// start, tight, tasks, children. List item: task, children. Table row:
// header, children. Table cell: align, inlines. Container: containerKind,
// containerTitle, children. HTML: text. Blockquote: children.
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
	tasks   bool
	task    *bool

	header bool
	align  string

	containerKind  string
	containerTitle string
}

type parser struct {
	labels         map[string]string
	usedHeadingIDs map[string]bool
	headings       []Heading
	firstH1        string
}

// Parse converts Markdown source into a document. It never returns an error:
// malformed constructs degrade to text, which is the useful behaviour for a
// documentation generator.
func Parse(source []byte, opts Options) *Document {
	text := strings.TrimPrefix(string(source), "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	p := &parser{
		labels:         opts.Containers,
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

		if level, text, explicitID, ok := headingStart(line); ok {
			inlines := parseInline(text)
			id := p.uniqueHeadingID(explicitID, plainInlines(inlines))
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
			inner, next := collectBlockquote(lines, i)
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

		if kind, title, ok := p.containerStart(line); ok {
			inner, next := p.collectContainer(lines, i)
			blocks = append(blocks, block{
				kind:           blockContainer,
				containerKind:  kind,
				containerTitle: title,
				children:       p.parseBlocks(inner),
			})
			i = next
			continue
		}

		if isHTMLBlockStart(line) {
			raw, next := collectHTMLBlock(lines, i)
			blocks = append(blocks, block{kind: blockHTML, text: raw})
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

		for i < len(lines) && !isBlank(lines[i]) && !p.startsBlock(lines[i]) && !isTableStart(lines, i) {
			i++
		}

		blocks = append(blocks, block{
			kind:    blockParagraph,
			inlines: parseInline(strings.Join(lines[start:i], "\n")),
		})
	}

	return blocks
}

func (p *parser) startsBlock(line string) bool {
	if startsBlockLine(line) {
		return true
	}

	if _, _, ok := p.containerStart(line); ok {
		return true
	}

	return isHTMLBlockStart(line)
}

func parseFence(lines []string, start int, char byte, length int, info string) (string, string, int) {
	language := ""
	if fields := strings.Fields(info); len(fields) > 0 {
		language = strings.ToLower(fields[0])
	}

	content := make([]string, 0, 8)
	i := start + 1

	for i < len(lines) {
		if fenceClose(lines[i], char, length) {
			i++
			break
		}

		content = append(content, lines[i])
		i++
	}

	code := strings.Join(content, "\n")
	if code != "" {
		code += "\n"
	}

	return code, language, i
}

func collectBlockquote(lines []string, start int) ([]string, int) {
	inner := make([]string, 0, 8)
	i := start

	for i < len(lines) {
		line := lines[i]

		if isBlank(line) {
			next := i + 1
			for next < len(lines) && isBlank(lines[next]) {
				next++
			}

			if next < len(lines) && blockquoteStart(lines[next]) {
				inner = append(inner, "")
				i++
				continue
			}

			break
		}

		if blockquoteStart(line) {
			inner = append(inner, stripBlockquote(line))
			i++
			continue
		}

		// Lazy continuation: a plain line directly after a quoted line still
		// belongs to the quote. A new block marker ends it.
		if !startsBlockLine(line) {
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
		if !ok || marker.indent != baseIndent || marker.ordered != first.ordered {
			break
		}

		content := marker.content
		var task *bool

		if done, rest, ok := taskMarker(content); ok {
			value := done
			task = &value
			content = rest
			list.tasks = true
		}

		itemLines := []string{content}
		hadBlank := false
		i++

		for i < len(lines) {
			if isBlank(lines[i]) {
				next := i + 1
				for next < len(lines) && isBlank(lines[next]) {
					next++
				}

				if next >= len(lines) {
					i = next
					break
				}

				nextMarker, nextIsItem := parseListMarker(lines[next])
				if nextIsItem && nextMarker.indent <= baseIndent {
					if nextMarker.ordered != first.ordered {
						i = next
						break
					}

					itemLines = append(itemLines, "")
					hadBlank = true
					i = next
					break
				}

				if indentWidth(lines[next]) <= baseIndent {
					i = next
					break
				}

				itemLines = append(itemLines, "")
				hadBlank = true
				i++
				continue
			}

			if next, ok := parseListMarker(lines[i]); ok && next.indent <= baseIndent {
				break
			}

			indent := indentWidth(lines[i])
			if indent <= baseIndent && startsBlockLine(lines[i]) {
				break
			}

			itemLines = append(itemLines, stripIndent(lines[i], marker.contentIndent))
			i++
		}

		itemLines = trimTrailingBlank(itemLines)
		if hadBlank {
			list.tight = false
		}

		list.children = append(list.children, block{
			kind:     blockListItem,
			children: p.parseBlocks(itemLines),
			task:     task,
		})
	}

	return list, i
}

func (p *parser) parseTable(lines []string, start int) (block, int) {
	aligns, _ := tableDelimiter(lines[start+1])
	if len(aligns) == 0 {
		aligns = []string{""}
	}

	headerCells := splitTableRow(lines[start])
	header := block{
		kind:   blockTableRow,
		header: true,
	}

	for index, align := range aligns {
		cellText := ""
		if index < len(headerCells) {
			cellText = headerCells[index]
		}

		header.children = append(header.children, block{
			kind:    blockTableCell,
			align:   align,
			inlines: parseInline(cellText),
		})
	}

	table := block{kind: blockTable}
	table.children = append(table.children, header)

	i := start + 2
	for i < len(lines) {
		if isBlank(lines[i]) || !strings.Contains(lines[i], "|") {
			break
		}

		cells := splitTableRow(lines[i])
		row := block{kind: blockTableRow}

		for index, align := range aligns {
			cellText := ""
			if index < len(cells) {
				cellText = cells[index]
			}

			row.children = append(row.children, block{
				kind:    blockTableCell,
				align:   align,
				inlines: parseInline(cellText),
			})
		}

		table.children = append(table.children, row)
		i++
	}

	return table, i
}

func (p *parser) containerStart(line string) (string, string, bool) {
	index := blockIndent(line)
	if index > 3 || !strings.HasPrefix(line[index:], ":::") {
		return "", "", false
	}

	rest := strings.TrimSpace(line[index+3:])
	if rest == "" {
		return "", "", false
	}

	kind := "tip"
	title := rest

	fields := strings.Fields(rest)
	if len(fields) > 0 {
		candidate := strings.ToLower(fields[0])
		if _, ok := p.labels[candidate]; ok {
			kind = candidate
			title = strings.TrimSpace(strings.TrimPrefix(rest, fields[0]))
		}
	}

	if title == "" {
		title = p.labels[kind]
		if title == "" {
			title = defaultContainerLabels[kind]
		}
		if title == "" {
			title = strings.ToUpper(kind)
		}
	}

	return kind, title, true
}

func (p *parser) collectContainer(lines []string, start int) ([]string, int) {
	inner := make([]string, 0, 16)
	depth := 1
	inFence := false
	var fenceChar byte
	var fenceLength int

	i := start + 1
	for i < len(lines) {
		line := lines[i]

		if inFence {
			inner = append(inner, line)
			if fenceClose(line, fenceChar, fenceLength) {
				inFence = false
			}
			i++
			continue
		}

		if char, length, _, ok := fenceStart(line); ok {
			inFence = true
			fenceChar = char
			fenceLength = length
			inner = append(inner, line)
			i++
			continue
		}

		if isContainerClose(line) {
			depth--
			if depth == 0 {
				i++
				break
			}

			inner = append(inner, line)
			i++
			continue
		}

		if _, _, ok := p.containerStart(line); ok {
			depth++
		}

		inner = append(inner, line)
		i++
	}

	return trimTrailingBlank(inner), i
}

func (p *parser) uniqueHeadingID(explicit string, text string) string {
	base := strings.TrimSpace(explicit)
	if base == "" {
		base = slugifyHeading(text)
	}
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
			case inlineText, inlineCode, inlineAutolink:
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

// voidHTMLElements never have a closing tag, so a raw HTML block made of one
// ends on the line where it starts.
var voidHTMLElements = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

func isHTMLBlockStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "<") {
		return false
	}

	if strings.HasPrefix(trimmed, "<!") || strings.HasPrefix(trimmed, "<?") {
		return true
	}

	name, closed, ok := htmlTagName(trimmed)
	if !ok {
		return false
	}

	// A void element has no closing tag, so it can only be a one-line block.
	// Treating an incomplete one as a block would swallow the rest of the file.
	if voidHTMLElements[name] {
		return closed
	}

	return true
}

// htmlTagName returns the lower-case element name at the start of a tag and
// whether that tag closes on this line. Autolinks such as
// <https://example.com> and text such as <3 do not produce a tag name followed
// by a delimiter, so they are rejected.
func htmlTagName(line string) (string, bool, bool) {
	rest := line[1:]
	if strings.HasPrefix(rest, "/") {
		rest = rest[1:]
	}
	if rest == "" || !isHTMLNameStart(rest[0]) {
		return "", false, false
	}

	end := 1
	for end < len(rest) && isHTMLNameChar(rest[end]) {
		end++
	}

	if end < len(rest) {
		switch rest[end] {
		case '>', '/', ' ', '\t', '\n', '\r':
		default:
			return "", false, false
		}
	}

	return strings.ToLower(rest[:end]), strings.Contains(rest[end:], ">"), true
}

func isHTMLNameChar(char byte) bool {
	return isHTMLNameStart(char) || (char >= '0' && char <= '9') || char == '-'
}

func collectHTMLBlock(lines []string, start int) (string, int) {
	trimmed := strings.TrimSpace(lines[start])
	if name, closed, ok := htmlTagName(trimmed); ok && closed && voidHTMLElements[name] {
		return lines[start], start + 1
	}

	end := start
	for end < len(lines) {
		line := strings.TrimSpace(lines[end])
		end++

		if endsHTMLBlock(line) {
			return strings.Join(lines[start:end], "\n"), end
		}
	}

	return strings.Join(lines[start:end], "\n"), end
}

// endsHTMLBlock reports whether a trimmed line closes a raw HTML block. The
// parser accepts the common endings instead of matching tags: a self-closing
// tag, a comment, a closing tag, a doctype or a processing instruction.
func endsHTMLBlock(line string) bool {
	if strings.HasSuffix(line, "/>") || strings.HasSuffix(line, "-->") {
		return true
	}

	if !strings.HasSuffix(line, ">") {
		return false
	}

	return strings.Contains(line, "</") ||
		strings.HasPrefix(line, "<!") ||
		strings.HasPrefix(line, "<?")
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
			case blockContainer:
				if item.containerTitle != "" {
					builder.WriteString(item.containerTitle)
					builder.WriteByte('\n')
				}
				walk(item.children)
			default:
				walk(item.children)
			}
		}
	}

	walk(blocks)

	return strings.Join(strings.Fields(builder.String()), " ")
}

var defaultContainerLabels = map[string]string{
	"tip":     "TIP",
	"info":    "INFO",
	"note":    "NOTE",
	"warning": "WARNING",
	"danger":  "DANGER",
	"details": "Details",
}

type listMarker struct {
	ordered       bool
	start         int
	indent        int
	contentIndent int
	content       string
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
		if next >= len(line) || (line[next] != ' ' && line[next] != '\t') {
			return listMarker{}, false
		}

		next = skipSpaces(line, next)

		return listMarker{
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

	end++
	if end >= len(line) || (line[end] != ' ' && line[end] != '\t') {
		return listMarker{}, false
	}

	end = skipSpaces(line, end)

	return listMarker{
		ordered:       true,
		start:         start,
		indent:        indent,
		contentIndent: end,
		content:       line[end:],
	}, true
}

func taskMarker(content string) (bool, string, bool) {
	if len(content) < 4 || content[0] != '[' || content[2] != ']' {
		return false, content, false
	}

	if content[3] != ' ' && content[3] != '\t' {
		return false, content, false
	}

	rest := strings.TrimLeft(content[3:], " \t")

	switch content[1] {
	case ' ':
		return false, rest, true
	case 'x', 'X':
		return true, rest, true
	}

	return false, content, false
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

func headingStart(line string) (int, string, string, bool) {
	indent := blockIndent(line)
	if indent > 3 || indent >= len(line) {
		return 0, "", "", false
	}

	end := indent
	for end < len(line) && line[end] == '#' {
		end++
	}

	level := end - indent
	if level < 1 || level > 6 {
		return 0, "", "", false
	}

	if end < len(line) && line[end] != ' ' && line[end] != '\t' {
		return 0, "", "", false
	}

	text := strings.TrimSpace(line[end:])

	if trimmed := strings.TrimRight(text, "#"); len(trimmed) < len(text) {
		if trimmed == "" || strings.HasSuffix(trimmed, " ") || strings.HasSuffix(trimmed, "\t") {
			text = strings.TrimSpace(trimmed)
		}
	}

	id := ""
	if strings.HasSuffix(text, "}") {
		if open := strings.LastIndex(text, "{#"); open >= 0 {
			candidate := strings.TrimSpace(text[open+2 : len(text)-1])
			if candidate != "" && !strings.ContainsAny(candidate, " \t") {
				id = candidate
				text = strings.TrimSpace(text[:open])
			}
		}
	}

	return level, text, id, true
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

func startsBlockLine(line string) bool {
	if isBlank(line) || thematicBreak(line) {
		return true
	}

	if _, _, _, ok := headingStart(line); ok {
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

	return ok && len(aligns) > 0
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)

	line = strings.TrimPrefix(line, "|")

	if strings.HasSuffix(line, "|") && !strings.HasSuffix(line, "\\|") {
		line = line[:len(line)-1]
	}

	cells := make([]string, 0, 4)
	var builder strings.Builder
	inCode := false

	for index := 0; index < len(line); index++ {
		value := line[index]

		if value == '\\' && index+1 < len(line) {
			builder.WriteByte(line[index+1])
			index++
			continue
		}

		if value == '`' {
			inCode = !inCode
			builder.WriteByte(value)
			continue
		}

		if value == '|' && !inCode {
			cells = append(cells, strings.TrimSpace(builder.String()))
			builder.Reset()
			continue
		}

		builder.WriteByte(value)
	}

	cells = append(cells, strings.TrimSpace(builder.String()))

	return cells
}

func isContainerClose(line string) bool {
	return strings.TrimSpace(line) == ":::"
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
