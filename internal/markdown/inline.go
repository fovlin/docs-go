package markdown

import (
	"html"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type inlineKind int

const (
	inlineText inlineKind = iota
	inlineCode
	inlineEmphasis
	inlineStrong
	inlineLink
	inlineImage
	inlineBreak
	inlineHTML
)

type inline struct {
	kind     inlineKind
	text     string
	url      string
	title    string
	children []inline
}

// inlineTagPattern matches one inline HTML tag; quoted attribute values may
// contain '>'.
var inlineTagPattern = regexp.MustCompile(`^</?[A-Za-z][A-Za-z0-9-]*(?:\s(?:[^<>"']|"[^"]*"|'[^']*')*)?/?>`)

func parseInline(value string) []inline {
	items := make([]inline, 0, 8)

	for index := 0; index < len(value); {
		if value[index] == '\n' {
			// Trailing spaces before a break would render inside the paragraph.
			if last := len(items) - 1; last >= 0 && items[last].kind == inlineText {
				items[last].text = strings.TrimRight(items[last].text, " ")
				if items[last].text == "" {
					items = items[:last]
				}
			}

			items = append(items, inline{kind: inlineBreak})
			index++
			continue
		}

		if item, next, ok := parseAtom(value, index); ok {
			items = append(items, item)
			index = next
			continue
		}

		// A malformed construct is literal text.
		start := index
		index++
		for index < len(value) && !isInlineSpecial(value[index]) {
			index++
		}
		items = append(items, inline{kind: inlineText, text: value[start:index]})
	}

	return items
}

// parseAtom parses one inline construct at index. A false result means the
// character carries no syntax here and is literal text.
func parseAtom(value string, index int) (inline, int, bool) {
	switch value[index] {
	case '\\':
		if index+1 >= len(value) {
			return inline{}, 0, false
		}
		switch next := value[index+1]; {
		case next == '\n':
			return inline{kind: inlineBreak}, index + 2, true
		case isEscapable(next):
			return inline{kind: inlineText, text: value[index+1 : index+2]}, index + 2, true
		}

	case '`':
		count := runLength(value, index, '`')
		end := skipCodeSpan(value, index)
		if end == index+count {
			// No closing run: the whole opening run is literal, so its
			// second backtick cannot open a span.
			return inline{kind: inlineText, text: value[index:end]}, end, true
		}

		content := value[index+count : end-count]
		content = strings.ReplaceAll(content, "\n", " ")

		if len(content) >= 2 && content[0] == ' ' && content[len(content)-1] == ' ' && strings.TrimSpace(content) != "" {
			content = content[1 : len(content)-1]
		}

		return inline{kind: inlineCode, text: content}, end, true

	case '*', '_':
		return parseDelimiter(value, index)

	case '!', '[':
		return parseLink(value, index)

	case '<':
		return parseHTMLInline(value, index)

	case '&':
		if text, next, ok := parseEntity(value, index); ok {
			return inline{kind: inlineText, text: text}, next, true
		}
	}

	return inline{}, 0, false
}

// skipCodeSpan returns the position just past the code span at start. Without
// a closing run it returns the position just past the opening backtick run.
func skipCodeSpan(value string, start int) int {
	count := runLength(value, start, '`')

	for index := start + count; index < len(value); {
		if value[index] != '`' {
			index++
			continue
		}

		run := runLength(value, index, '`')
		if run == count {
			return index + count
		}

		index += run
	}

	return start + count
}

func parseDelimiter(value string, start int) (inline, int, bool) {
	char := value[start]
	count := runLength(value, start, char)

	previous := previousRune(value, start)
	next := nextRune(value, start+count)

	if !delimiterCanOpen(char, previous, next) {
		return inline{}, 0, false
	}

	if count >= 2 {
		if end, ok := findClosingDelimiter(value, start+2, char, 2); ok {
			return inline{kind: inlineStrong, children: parseInline(value[start+2 : end])}, end + 2, true
		}
	}

	if end, ok := findClosingDelimiter(value, start+count, char, 1); ok {
		return inline{kind: inlineEmphasis, children: parseInline(value[start+1 : end])}, end + 1, true
	}

	return inline{}, 0, false
}

// findClosingDelimiter finds a closing run of char. A single-character
// delimiter must close on a run of exactly one, leaving a double run to strong.
func findClosingDelimiter(value string, start int, char byte, need int) (int, bool) {
	for index := start; index < len(value); {
		switch value[index] {
		case '\\':
			index += 2

		case '`':
			index = skipCodeSpan(value, index)

		case char:
			run := runLength(value, index, char)

			if need == 1 && run != 1 {
				index += run
				continue
			}

			if run >= need {
				closeAt := index + run - need
				if delimiterCanClose(char, previousRune(value, closeAt), nextRune(value, closeAt+need)) {
					return closeAt, true
				}
			}

			index += run

		default:
			index++
		}
	}

	return 0, false
}

func parseLink(value string, start int) (inline, int, bool) {
	image := value[start] == '!'
	bracket := start
	if image {
		bracket++
	}

	if bracket >= len(value) || value[bracket] != '[' {
		return inline{}, 0, false
	}

	close, ok := findClosingBracket(value, bracket)
	if !ok || close+1 >= len(value) || value[close+1] != '(' {
		return inline{}, 0, false
	}

	destination, title, next, ok := parseLinkTarget(value, close+2)
	if !ok {
		return inline{}, 0, false
	}

	kind := inlineLink
	if image {
		kind = inlineImage
	}

	return inline{
		kind:     kind,
		url:      destination,
		title:    title,
		children: parseInline(value[bracket+1 : close]),
	}, next, true
}

func findClosingBracket(value string, open int) (int, bool) {
	depth := 1

	for index := open + 1; index < len(value); index++ {
		switch value[index] {
		case '\\':
			index++
		case '`':
			index = skipCodeSpan(value, index) - 1
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return index, true
			}
		}
	}

	return 0, false
}

func parseLinkTarget(value string, start int) (destination, title string, next int, ok bool) {
	index := skipSpaces(value, start)

	if index < len(value) && value[index] == '<' {
		end, ok := scanDelimited(value, index+1, '>')
		if !ok {
			return "", "", 0, false
		}

		destination = unescapeLinkText(value[index+1 : end])
		index = end + 1
	} else {
		end := scanLinkDestination(value, index)
		destination = unescapeLinkText(value[index:end])
		index = end
	}

	index = skipSpaces(value, index)
	if index < len(value) && (value[index] == '"' || value[index] == '\'' || value[index] == '(') {
		closing := value[index]
		if closing == '(' {
			closing = ')'
		}

		end, ok := scanDelimited(value, index+1, closing)
		if !ok {
			return "", "", 0, false
		}

		title = unescapeLinkText(value[index+1 : end])
		index = skipSpaces(value, end+1)
	}

	if index >= len(value) || value[index] != ')' {
		return "", "", 0, false
	}

	return html.UnescapeString(destination), html.UnescapeString(title), index + 1, true
}

// scanDelimited returns the index of the next unescaped closing byte after
// start.
func scanDelimited(value string, start int, closing byte) (int, bool) {
	for index := start; index < len(value); index++ {
		switch value[index] {
		case '\\':
			index++
		case closing:
			return index, true
		}
	}

	return 0, false
}

// scanLinkDestination returns the end of a bare (non-angle) destination.
// Nested parentheses are part of the destination; whitespace ends it.
func scanLinkDestination(value string, index int) int {
	depth := 0

	for index < len(value) {
		switch value[index] {
		case '\\':
			if index+1 < len(value) {
				index++
			}
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return index
			}
			depth--
		case ' ', '\t', '\n':
			return index
		}

		index++
	}

	return index
}

func parseEntity(value string, start int) (string, int, bool) {
	end := strings.IndexByte(value[start:], ';')
	if end < 0 || end > 32 {
		return "", 0, false
	}

	candidate := value[start : start+end+1]
	if decoded := html.UnescapeString(candidate); decoded != candidate {
		return decoded, start + end + 1, true
	}

	return "", 0, false
}

func delimiterCanOpen(char byte, previous, next rune) bool {
	return next != 0 && !unicode.IsSpace(next) && !(char == '_' && isWordRune(previous) && isWordRune(next))
}

func delimiterCanClose(char byte, previous, next rune) bool {
	return previous != 0 && !unicode.IsSpace(previous) && !(char == '_' && isWordRune(previous) && isWordRune(next))
}

func isWordRune(value rune) bool {
	return unicode.IsLetter(value) || unicode.IsDigit(value)
}

func isEscapable(value byte) bool {
	return strings.ContainsRune(`!"#$%&'()*+,-./:;<=>?@[\]^_`+"`"+`{|}~`, rune(value))
}

func isInlineSpecial(value byte) bool {
	return strings.IndexByte("\\`*_[!<&\n", value) >= 0
}

func unescapeLinkText(value string) string {
	var builder strings.Builder

	for index := 0; index < len(value); index++ {
		if value[index] == '\\' && index+1 < len(value) && isEscapable(value[index+1]) {
			builder.WriteByte(value[index+1])
			index++
			continue
		}

		builder.WriteByte(value[index])
	}

	return builder.String()
}

func previousRune(value string, index int) rune {
	if index <= 0 {
		return 0
	}

	run, _ := utf8.DecodeLastRuneInString(value[:index])

	return run
}

func nextRune(value string, index int) rune {
	if index >= len(value) {
		return 0
	}

	run, _ := utf8.DecodeRuneInString(value[index:])

	return run
}

func parseHTMLInline(value string, start int) (inline, int, bool) {
	if strings.HasPrefix(value[start:], "<!--") {
		end := strings.Index(value[start+4:], "-->")
		if end < 0 {
			return inline{}, 0, false
		}
		end += start + 7
		return inline{kind: inlineHTML, text: value[start:end]}, end, true
	}

	match := inlineTagPattern.FindStringIndex(value[start:])
	if match == nil {
		return inline{}, 0, false
	}
	return inline{kind: inlineHTML, text: value[start : start+match[1]]}, start + match[1], true
}

// plainInlines strips inline syntax down to readable text. Breaks become
// spaces; images and links keep the text they wrap.
func plainInlines(inlines []inline) string {
	var builder strings.Builder

	var walk func([]inline)
	walk = func(items []inline) {
		for _, item := range items {
			switch item.kind {
			case inlineText, inlineCode:
				builder.WriteString(item.text)
			case inlineBreak:
				builder.WriteByte(' ')
			default:
				walk(item.children)
			}
		}
	}

	walk(inlines)

	return strings.Join(strings.Fields(builder.String()), " ")
}
