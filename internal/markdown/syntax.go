package markdown

import (
	"strconv"
	"strings"
)

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

// fenceStart parses an opening backtick or tilde fence.
func fenceStart(line string) (byte, int, string, bool) {
	indent := blockIndent(line)
	if indent > 3 || indent >= len(line) {
		return 0, 0, "", false
	}

	char := line[indent]
	if char != '`' && char != '~' {
		return 0, 0, "", false
	}

	length := runLength(line, indent, char)
	if length < 3 {
		return 0, 0, "", false
	}

	info := strings.TrimSpace(line[indent+length:])
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

	run := runLength(line, indent, char)
	if run < length {
		return false
	}

	return strings.TrimSpace(line[indent+run:]) == ""
}

func headingStart(line string) (int, string, bool) {
	indent := blockIndent(line)
	if indent > 3 || indent >= len(line) {
		return 0, "", false
	}

	level := runLength(line, indent, '#')
	if level < 1 || level > 6 {
		return 0, "", false
	}

	end := indent + level
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

func tableDelimiter(line string) ([]string, bool) {
	cells := splitTableRow(line)
	aligns := make([]string, len(cells))

	for index, cell := range cells {
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

// tableStart reports whether a table starts here and returns its column
// alignments. The delimiter row is parsed once for both checks.
func tableStart(lines []string, index int) ([]string, bool) {
	if index+1 >= len(lines) || !isTableRaw(lines[index]) {
		return nil, false
	}

	aligns, ok := tableDelimiter(lines[index+1])

	return aligns, ok && len(aligns) == len(splitTableRow(lines[index]))
}

func isTableRaw(line string) bool {
	if !strings.ContainsRune(line, '|') {
		return false
	}

	trimmed := strings.TrimSpace(line)

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
