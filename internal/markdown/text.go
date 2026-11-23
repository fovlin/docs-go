package markdown

import (
	"strconv"
	"strings"
	"unicode"
)

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
	for _, item := range blocks {
		if item.kind == blockParagraph {
			return plainInlines(item.inlines)
		}
		if text := firstParagraphText(item.children); text != "" {
			return text
		}
	}

	return ""
}
