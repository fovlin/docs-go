package markdown

import (
	"regexp"
	"strconv"
	"strings"
)

// blockKind is the kind of block that starts on a source line.
type blockKind int

const (
	blockParagraph blockKind = iota
	blockFence
	blockIndentedCode
	blockHeading
	blockThematicBreak
	blockBlockquote
	blockList
	blockContainer
	blockTable
)

// lineKind classifies the block that starts on one line. A table is only
// identifiable together with its delimiter line and is resolved by blockKindAt.
func lineKind(line string) blockKind {
	if _, _, _, ok := fenceStart(line); ok {
		return blockFence
	}
	if width, _ := leadingSpace(line); width >= 4 {
		return blockIndentedCode
	}
	if _, _, ok := headingStart(line); ok {
		return blockHeading
	}
	if thematicBreak(line) {
		return blockThematicBreak
	}
	if blockquoteStart(line) {
		return blockBlockquote
	}
	if _, ok := parseListMarker(line); ok {
		return blockList
	}
	if _, _, ok := containerStart(line); ok {
		return blockContainer
	}
	return blockParagraph
}

// blockKindAt classifies the block at index, resolving the one construct that
// needs more than a single line: a table header and its delimiter.
func blockKindAt(lines []string, index int) blockKind {
	kind := lineKind(lines[index])
	if kind == blockParagraph {
		if _, _, ok := tableStart(lines, index); ok {
			return blockTable
		}
	}
	return kind
}

// listMarker describes one list item marker as recognised on a line.
type listMarker struct {
	ordered bool
	marker  byte
	start   int
	indent  int
	content string

	// contentIndent is the visual column where the item's content begins. It
	// is the indent used to strip continuation lines.
	contentIndent int
}

// continues reports whether m keeps the same list as first.
func (m listMarker) continues(first listMarker) bool {
	return m.indent >= first.indent && m.ordered == first.ordered && m.marker == first.marker
}

// listMarkerPattern captures the marker, the single space or tab separating it
// from the content, and the content itself.
var listMarkerPattern = regexp.MustCompile(`^([-+*]|[0-9]{1,9}[.)])([ \t])(.*\S.*)$`)

// parseListMarker recognises an unordered bullet (-, * or +) or an ordered
// marker (one to nine digits followed by . or )). Exactly one space or tab is
// consumed; the rest is content and keeps its own leading whitespace.
func parseListMarker(line string) (listMarker, bool) {
	indent, index := leadingSpace(line)
	if indent > 3 {
		return listMarker{}, false
	}

	match := listMarkerPattern.FindStringSubmatch(line[index:])
	if match == nil || strings.TrimSpace(match[3]) == "" {
		return listMarker{}, false
	}

	separator := 1
	if match[2][0] == '\t' {
		separator = 4 - (indent+len(match[1]))%4
	}

	marker := listMarker{
		indent:        indent,
		content:       match[3],
		contentIndent: indent + len(match[1]) + separator,
	}
	if match[1][0] >= '0' && match[1][0] <= '9' {
		marker.ordered = true
		marker.start, _ = strconv.Atoi(match[1][:len(match[1])-1])
		marker.marker = match[1][len(match[1])-1]
	} else {
		marker.marker = match[1][0]
	}

	return marker, true
}

// fenceStart parses an opening backtick or tilde fence.
func fenceStart(line string) (char byte, length int, info string, ok bool) {
	indent, index := leadingSpace(line)
	if indent > 3 || index >= len(line) {
		return 0, 0, "", false
	}

	char = line[index]
	if char != '`' && char != '~' {
		return 0, 0, "", false
	}
	length = runLength(line, index, char)
	if length < 3 {
		return 0, 0, "", false
	}
	info = strings.TrimSpace(line[index+length:])
	if char == '`' && strings.ContainsRune(info, '`') {
		return 0, 0, "", false
	}
	return char, length, info, true
}

// fenceEnd finds the first line at or after start that closes the fence: at
// most three spaces of indent, at least length fence characters, and nothing
// but whitespace after them. Unclosed fences return len(lines) and false.
func fenceEnd(lines []string, start int, char byte, length int) (int, bool) {
	for index := start; index < len(lines); index++ {
		line := lines[index]
		indent, at := leadingSpace(line)
		if indent > 3 || at >= len(line) || line[at] != char {
			continue
		}
		if run := runLength(line, at, char); run >= length && strings.TrimSpace(line[at+run:]) == "" {
			return index, true
		}
	}
	return len(lines), false
}

func headingStart(line string) (int, string, bool) {
	indent, index := leadingSpace(line)
	if indent > 3 || index >= len(line) {
		return 0, "", false
	}

	level := runLength(line, index, '#')
	if level < 1 || level > 6 {
		return 0, "", false
	}
	end := index + level
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

// thematicBreakPattern matches three or more of the same break character,
// optionally separated by spaces or tabs.
var thematicBreakPattern = regexp.MustCompile(`^ {0,3}(?:(?:\*[ \t]*){3,}|(?:-[ \t]*){3,}|(?:_[ \t]*){3,})$`)

func thematicBreak(line string) bool {
	return thematicBreakPattern.MatchString(line)
}

func blockquoteStart(line string) bool {
	indent, index := leadingSpace(line)

	return indent <= 3 && index < len(line) && line[index] == '>'
}

func stripBlockquote(line string) string {
	indent, index := leadingSpace(line)
	if indent > 3 || index >= len(line) || line[index] != '>' {
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

func containerStart(line string) (kind, title string, ok bool) {
	indent, index := leadingSpace(line)
	if indent > 3 || !strings.HasPrefix(line[index:], ":::") {
		return "", "", false
	}

	rest := strings.TrimSpace(line[index+3:])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return "", "", false
	}

	kind = strings.ToLower(fields[0])
	title, ok = containerTitles[kind]
	if !ok {
		return "", "", false
	}
	if custom := strings.TrimSpace(rest[len(fields[0]):]); custom != "" {
		title = custom
	}
	return kind, title, true
}

func isContainerClose(line string) bool {
	indent, _ := leadingSpace(line)

	return indent <= 3 && strings.TrimSpace(line) == ":::"
}

// isTableRow reports whether a line can be a table row: it contains a pipe and
// does not start any other block, which is what ends a table body.
func isTableRow(line string) bool {
	return strings.ContainsRune(line, '|') && lineKind(line) == blockParagraph
}

// tableAlignPattern matches one delimiter cell: optional colons around at
// least three dashes.
var tableAlignPattern = regexp.MustCompile(`^:?-{3,}:?$`)

func tableDelimiter(line string) ([]string, bool) {
	cells := splitTableRow(line)
	aligns := make([]string, len(cells))

	for index, cell := range cells {
		if !tableAlignPattern.MatchString(cell) {
			return nil, false
		}
		switch {
		case strings.HasPrefix(cell, ":"):
			if strings.HasSuffix(cell, ":") {
				aligns[index] = "center"
			} else {
				aligns[index] = "left"
			}
		case strings.HasSuffix(cell, ":"):
			aligns[index] = "right"
		}
	}

	return aligns, true
}

// tableStart reports whether a table starts here, returning its header cells
// and column alignments.
func tableStart(lines []string, index int) (header, aligns []string, ok bool) {
	if index+1 >= len(lines) || !isTableRow(lines[index]) {
		return nil, nil, false
	}

	header = splitTableRow(lines[index])
	if aligns, ok = tableDelimiter(lines[index+1]); !ok || len(aligns) != len(header) {
		return nil, nil, false
	}
	return header, aligns, true
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	// A trailing pipe is a border only when no backslash escapes it.
	if strings.HasSuffix(line, "|") {
		backslashes := 0
		for index := len(line) - 2; index >= 0 && line[index] == '\\'; index-- {
			backslashes++
		}
		if backslashes%2 == 0 {
			line = line[:len(line)-1]
		}
	}

	cells := make([]string, 0, 4)
	var cell strings.Builder

	for index := 0; index < len(line); index++ {
		switch {
		case line[index] == '\\' && index+1 < len(line):
			if line[index+1] == '|' {
				cell.WriteByte('|')
			} else {
				cell.WriteString(line[index : index+2])
			}
			index++

		case line[index] == '|':
			cells = append(cells, strings.TrimSpace(cell.String()))
			cell.Reset()

		default:
			cell.WriteByte(line[index])
		}
	}

	return append(cells, strings.TrimSpace(cell.String()))
}

// leadingSpace returns the visual width of the line's leading whitespace and
// the byte index where its content starts. Tab stops are four columns wide.
func leadingSpace(line string) (width, index int) {
	for index < len(line) {
		switch line[index] {
		case ' ':
			width++
		case '\t':
			width += 4 - width%4
		default:
			return width, index
		}
		index++
	}
	return width, index
}

func stripIndent(line string, width int) string {
	consumed := 0

	for index := 0; index < len(line); index++ {
		if consumed >= width {
			return line[index:]
		}
		switch line[index] {
		case ' ':
			consumed++
		case '\t':
			next := consumed + 4 - consumed%4
			if next > width {
				return strings.Repeat(" ", next-width) + line[index+1:]
			}
			consumed = next
		default:
			return line[index:]
		}
	}
	return ""
}

func skipSpaces(value string, index int) int {
	for index < len(value) && (value[index] == ' ' || value[index] == '\t' || value[index] == '\n') {
		index++
	}
	return index
}

func runLength(value string, start int, char byte) int {
	end := start
	for end < len(value) && value[end] == char {
		end++
	}
	return end - start
}
