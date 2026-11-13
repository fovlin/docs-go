package markdown

import (
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
	inlineStrike
	inlineLink
	inlineImage
	inlineAutolink
	inlineSoftBreak
	inlineHardBreak
	inlineHTML
)

type inline struct {
	kind     inlineKind
	text     string
	url      string
	title    string
	children []inline
}

func parseInline(value string) []inline {
	return parseInlineSegment(value)
}

func parseInlineSegment(value string) []inline {
	items := make([]inline, 0, 8)

	for i := 0; i < len(value); {
		switch value[i] {
		case '\\':
			if i+1 < len(value) && value[i+1] == '\n' {
				items = append(items, inline{kind: inlineHardBreak})
				i += 2
				continue
			}

			if i+1 < len(value) && isEscapable(value[i+1]) {
				items = append(items, inline{kind: inlineText, text: value[i+1 : i+2]})
				i += 2
				continue
			}

			items = append(items, inline{kind: inlineText, text: "\\"})
			i++

		case '`':
			if item, next, ok := parseCodeSpan(value, i); ok {
				items = append(items, item)
				i = next
				continue
			}

			count := runLength(value, i, '`')
			items = append(items, inline{kind: inlineText, text: value[i : i+count]})
			i += count

		case '*', '_', '~':
			if item, next, ok := parseDelimiter(value, i); ok {
				items = append(items, item)
				i = next
				continue
			}

			items = append(items, inline{kind: inlineText, text: value[i : i+1]})
			i++

		case '!':
			if i+1 < len(value) && value[i+1] == '[' {
				if item, next, ok := parseLink(value, i); ok {
					items = append(items, item)
					i = next
					continue
				}
			}

			items = append(items, inline{kind: inlineText, text: "!"})
			i++

		case '[':
			if item, next, ok := parseLink(value, i); ok {
				items = append(items, item)
				i = next
				continue
			}

			items = append(items, inline{kind: inlineText, text: "["})
			i++

		case '<':
			if item, next, ok := parseAngleAutolink(value, i); ok {
				items = append(items, item)
				i = next
				continue
			}
			if item, next, ok := parseHTMLInline(value, i); ok {
				items = append(items, item)
				i = next
				continue
			}

			items = append(items, inline{kind: inlineText, text: "<"})
			i++

		case '&':
			if text, next, ok := parseEntity(value, i); ok {
				items = append(items, inline{kind: inlineText, text: text})
				i = next
				continue
			}

			items = append(items, inline{kind: inlineText, text: "&"})
			i++

		case '\n':
			var spaces int
			items, spaces = trimTrailingSpaces(items)

			if spaces >= 2 {
				items = append(items, inline{kind: inlineHardBreak})
			} else {
				items = append(items, inline{kind: inlineSoftBreak})
			}

			i++

		default:
			if item, next, ok := parseBareAutolink(value, i); ok {
				items = append(items, item)
				i = next
				continue
			}

			start := i
			i++

			for i < len(value) && !isInlineSpecial(value[i]) {
				if i > start && urlBoundary(value[i-1]) && looksLikeURL(value[i:]) {
					break
				}

				i++
			}

			items = append(items, inline{kind: inlineText, text: value[start:i]})
		}
	}

	return items
}

func parseCodeSpan(value string, start int) (inline, int, bool) {
	count := runLength(value, start, '`')
	if count == 0 {
		return inline{}, start, false
	}

	for index := start + count; index < len(value); {
		if value[index] != '`' {
			index++
			continue
		}

		run := runLength(value, index, '`')
		if run != count {
			index += run
			continue
		}

		content := value[start+count : index]
		content = strings.ReplaceAll(content, "\n", " ")

		if len(content) >= 2 && content[0] == ' ' && content[len(content)-1] == ' ' && strings.TrimSpace(content) != "" {
			content = content[1 : len(content)-1]
		}

		return inline{kind: inlineCode, text: content}, index + count, true
	}

	return inline{}, start, false
}

func parseDelimiter(value string, start int) (inline, int, bool) {
	char := value[start]
	count := runLength(value, start, char)

	previous := previousRune(value, start)
	next := nextRune(value, start+count)

	if !delimiterCanOpen(char, previous, next) {
		return inline{}, start, false
	}

	if char == '~' {
		if count < 2 {
			return inline{}, start, false
		}

		if end, ok := findClosingDelimiter(value, start+2, char, 2); ok {
			inner := parseInlineSegment(value[start+2 : end])

			return inline{kind: inlineStrike, children: inner}, end + 2, true
		}

		return inline{}, start, false
	}

	if count >= 2 {
		if end, ok := findClosingDelimiter(value, start+2, char, 2); ok {
			inner := parseInlineSegment(value[start+2 : end])

			return inline{kind: inlineStrong, children: inner}, end + 2, true
		}
	}

	if end, ok := findClosingDelimiter(value, start+1, char, 1); ok {
		inner := parseInlineSegment(value[start+1 : end])

		return inline{kind: inlineEmphasis, children: inner}, end + 1, true
	}

	return inline{}, start, false
}

func findClosingDelimiter(value string, start int, char byte, need int) (int, bool) {
	for index := start; index < len(value); {
		switch value[index] {
		case '\\':
			index += 2
			continue

		case '`':
			if _, next, ok := parseCodeSpan(value, index); ok {
				index = next
				continue
			}

			index += runLength(value, index, '`')
			continue

		case char:
			run := runLength(value, index, char)

			if need == 1 {
				if run == 1 && delimiterCanClose(char, previousRune(value, index), nextRune(value, index+1)) {
					return index, true
				}

				index += run
				continue
			}

			if run >= 2 {
				closeAt := index + run - 2
				if delimiterCanClose(char, previousRune(value, closeAt), nextRune(value, closeAt+2)) {
					return closeAt, true
				}
			}

			index += run
			continue
		}

		index++
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
		return inline{}, start, false
	}

	close, ok := findClosingBracket(value, bracket)
	if !ok || close+1 >= len(value) || value[close+1] != '(' {
		return inline{}, start, false
	}

	destination, title, next, ok := parseLinkTarget(value, close+2)
	if !ok {
		return inline{}, start, false
	}

	children := parseInlineSegment(value[bracket+1 : close])
	kind := inlineLink
	if image {
		kind = inlineImage
	}

	return inline{
		kind:     kind,
		url:      destination,
		title:    title,
		children: children,
	}, next, true
}

func findClosingBracket(value string, open int) (int, bool) {
	depth := 1

	for index := open + 1; index < len(value); index++ {
		switch value[index] {
		case '\\':
			index++
		case '`':
			if _, next, ok := parseCodeSpan(value, index); ok {
				index = next - 1
				continue
			}
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

func parseLinkTarget(value string, start int) (string, string, int, bool) {
	index := skipInlineSpaces(value, start)

	var destination string

	if index < len(value) && value[index] == '<' {
		end := index + 1
		for end < len(value) && value[end] != '>' {
			if value[end] == '\\' && end+1 < len(value) {
				end++
			}
			end++
		}

		if end >= len(value) {
			return "", "", start, false
		}

		destination = unescapeLinkText(value[index+1 : end])
		index = end + 1
	} else {
		var builder strings.Builder
		depth := 0

		for index < len(value) {
			char := value[index]

			if char == '\\' && index+1 < len(value) {
				builder.WriteByte(value[index+1])
				index += 2
				continue
			}

			if char == '(' {
				depth++
				builder.WriteByte(char)
				index++
				continue
			}

			if char == ')' {
				if depth == 0 {
					break
				}

				depth--
				builder.WriteByte(char)
				index++
				continue
			}

			if char == ' ' || char == '\t' || char == '\n' {
				break
			}

			builder.WriteByte(char)
			index++
		}

		destination = builder.String()
	}

	index = skipInlineSpaces(value, index)
	title := ""

	if index < len(value) && (value[index] == '"' || value[index] == '\'' || value[index] == '(') {
		opening := value[index]
		closing := opening
		if opening == '(' {
			closing = ')'
		}

		index++
		begin := index

		for index < len(value) && value[index] != closing {
			if value[index] == '\\' && index+1 < len(value) {
				index++
			}
			index++
		}

		if index >= len(value) {
			return "", "", start, false
		}

		title = unescapeLinkText(value[begin:index])
		index++
		index = skipInlineSpaces(value, index)
	}

	if index >= len(value) || value[index] != ')' {
		return "", "", start, false
	}

	return destination, title, index + 1, true
}

func parseAngleAutolink(value string, start int) (inline, int, bool) {
	end := strings.IndexByte(value[start+1:], '>')
	if end < 0 {
		return inline{}, start, false
	}

	end += start + 1
	content := value[start+1 : end]

	if content == "" || strings.ContainsAny(content, " \t\n") {
		return inline{}, start, false
	}

	lower := strings.ToLower(content)

	switch {
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "https://"), strings.HasPrefix(lower, "mailto:"):
		return inline{kind: inlineAutolink, text: content, url: content}, end + 1, true

	case looksLikeEmail(content):
		return inline{kind: inlineAutolink, text: content, url: "mailto:" + content}, end + 1, true
	}

	return inline{}, start, false
}

func parseBareAutolink(value string, start int) (inline, int, bool) {
	if !looksLikeURL(value[start:]) {
		return inline{}, start, false
	}

	if start > 0 && !urlBoundary(value[start-1]) {
		return inline{}, start, false
	}

	end := start
	for end < len(value) && value[end] != ' ' && value[end] != '\t' && value[end] != '\n' && value[end] != '<' && value[end] != '>' {
		end++
	}

	for end > start && strings.ContainsRune(".,;:!?", rune(value[end-1])) {
		end--
	}

	for end > start && value[end-1] == ')' && !strings.Contains(value[start:end], "(") {
		end--
	}

	if end == start {
		return inline{}, start, false
	}

	url := value[start:end]

	return inline{kind: inlineAutolink, text: url, url: url}, end, true
}

func parseEntity(value string, start int) (string, int, bool) {
	end := strings.IndexByte(value[start:], ';')
	if end < 0 || end > 32 {
		return "", start, false
	}

	candidate := value[start : start+end+1]
	decoded := htmlUnescape(candidate)

	if decoded == candidate {
		return "", start, false
	}

	return decoded, start + end + 1, true
}

func trimTrailingSpaces(items []inline) ([]inline, int) {
	if len(items) == 0 || items[len(items)-1].kind != inlineText {
		return items, 0
	}

	last := &items[len(items)-1]
	trimmed := strings.TrimRight(last.text, " ")
	spaces := len(last.text) - len(trimmed)
	last.text = trimmed

	if last.text == "" {
		items = items[:len(items)-1]
	}

	return items, spaces
}

func delimiterCanOpen(char byte, previous rune, next rune) bool {
	if next == 0 || unicode.IsSpace(next) {
		return false
	}

	if char == '_' && isWordRune(previous) && isWordRune(next) {
		return false
	}

	return true
}

func delimiterCanClose(char byte, previous rune, next rune) bool {
	if previous == 0 || unicode.IsSpace(previous) {
		return false
	}

	if char == '_' && isWordRune(previous) && isWordRune(next) {
		return false
	}

	return true
}

func isWordRune(value rune) bool {
	return unicode.IsLetter(value) || unicode.IsDigit(value)
}

func isEscapable(value byte) bool {
	return strings.ContainsRune(`!"#$%&'()*+,-./:;<=>?@[\]^_`+"`"+`{|}~`, rune(value))
}

func isInlineSpecial(value byte) bool {
	switch value {
	case '\\', '`', '*', '_', '~', '[', '!', '<', '&', '\n':
		return true
	}

	return false
}

func looksLikeURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func looksLikeEmail(value string) bool {
	at := strings.IndexByte(value, '@')
	if at <= 0 || at == len(value)-1 || strings.ContainsAny(value, " \t\n") {
		return false
	}

	domain := value[at+1:]

	return strings.Contains(domain, ".") && !strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".")
}

func urlBoundary(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '(', '[', '<', '"', '\'', '*', '_', '~':
		return true
	}

	return false
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

func skipInlineSpaces(value string, index int) int {
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
	if start >= len(value) || value[start] != '<' {
		return inline{}, start, false
	}

	if strings.HasPrefix(value[start:], "<!--") {
		end := strings.Index(value[start+4:], "-->")
		if end < 0 {
			return inline{}, start, false
		}
		end += start + 7
		return inline{kind: inlineHTML, text: value[start:end]}, end, true
	}

	end := start + 1
	var quote byte
	for end < len(value) {
		char := value[end]
		if quote != 0 {
			if char == quote {
				quote = 0
			}
		} else if char == '\'' || char == '"' {
			quote = char
		} else if char == '>' {
			break
		}
		end++
	}
	if end >= len(value) {
		return inline{}, start, false
	}

	raw := value[start : end+1]
	inner := raw[1 : len(raw)-1]
	if inner == "" {
		return inline{}, start, false
	}
	if inner[0] == '/' {
		inner = inner[1:]
	}
	if inner == "" || (inner[0] != '!' && !isHTMLNameStart(inner[0])) {
		return inline{}, start, false
	}

	return inline{kind: inlineHTML, text: raw}, end + 1, true
}

func isHTMLNameStart(char byte) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
}
