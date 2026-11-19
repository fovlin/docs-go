package markdown

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

// ResolveFunc rewrites a link or image destination before it is written.
// Returning the original value leaves it unchanged.
type ResolveFunc func(destination string, image bool) string

// Render writes the document as an HTML fragment.
func (document *Document) Render(resolve ResolveFunc) []byte {
	renderer := &renderer{resolve: resolve}
	renderer.blocks(document.blocks)

	return []byte(renderer.builder.String())
}

type renderer struct {
	builder strings.Builder
	resolve ResolveFunc
}

func (r *renderer) blocks(items []block) {
	for _, item := range items {
		r.block(item)
	}
}

func (r *renderer) block(item block) {
	switch item.kind {
	case blockHeading:
		r.heading(item)

	case blockParagraph:
		r.builder.WriteString("<p>")
		r.inlines(item.inlines)
		r.builder.WriteString("</p>\n")

	case blockCode:
		r.code(item)

	case blockBlockquote:
		r.builder.WriteString("<blockquote>\n")
		r.blocks(item.children)
		r.builder.WriteString("</blockquote>\n")

	case blockList:
		r.list(item)

	case blockTable:
		r.table(item)

	case blockContainer:
		r.container(item)

	case blockThematicBreak:
		r.builder.WriteString("<hr>\n")

	default:
		r.blocks(item.children)
	}
}

func (r *renderer) heading(item block) {
	fmt.Fprintf(&r.builder, `<h%d id="%s">`, item.level, html.EscapeString(item.id))
	r.inlines(item.inlines)

	if item.id != "" {
		fmt.Fprintf(&r.builder, `<a class="header-anchor" href="#%s" aria-label="Link to this heading"></a>`, html.EscapeString(item.id))
	}

	fmt.Fprintf(&r.builder, "</h%d>\n", item.level)
}

func (r *renderer) code(item block) {
	language := strings.ToLower(strings.TrimSpace(item.language))
	if fields := strings.Fields(language); len(fields) > 0 {
		language = fields[0]
	}

	if language == "" {
		r.builder.WriteString(`<div class="code-block"><pre><code>`)
	} else {
		fmt.Fprintf(&r.builder, `<div class="code-block" data-lang="%s"><pre><code class="language-%s">`, html.EscapeString(language), html.EscapeString(language))
	}

	r.builder.WriteString(html.EscapeString(item.text))

	r.builder.WriteString("</code></pre></div>\n")
}

func (r *renderer) list(item block) {
	tag := "ul"
	if item.ordered {
		tag = "ol"
	}

	fmt.Fprintf(&r.builder, "<%s", tag)

	if item.ordered && item.start != 1 {
		fmt.Fprintf(&r.builder, ` start="%d"`, item.start)
	}

	r.builder.WriteString(">\n")

	for _, child := range item.children {
		r.builder.WriteString("<li>")

		if len(child.children) > 0 && child.children[0].kind == blockParagraph {
			r.inlines(child.children[0].inlines)
			if len(child.children) > 1 {
				r.builder.WriteByte('\n')
			}
			r.blocks(child.children[1:])
		} else {
			r.blocks(child.children)
		}

		r.builder.WriteString("</li>\n")
	}

	fmt.Fprintf(&r.builder, "</%s>\n", tag)
}

func (r *renderer) table(item block) {
	r.builder.WriteString("<table>\n")

	rows := item.children
	if len(rows) == 0 {
		r.builder.WriteString("</table>\n")
		return
	}

	r.builder.WriteString("<thead>\n")
	r.tableRow(rows[0])
	r.builder.WriteString("</thead>\n")

	if len(rows) > 1 {
		r.builder.WriteString("<tbody>\n")
		for _, row := range rows[1:] {
			r.tableRow(row)
		}
		r.builder.WriteString("</tbody>\n")
	}

	r.builder.WriteString("</table>\n")
}

func (r *renderer) tableRow(row block) {
	r.builder.WriteString("<tr>\n")

	for _, cell := range row.children {
		tag := "td"
		if row.header {
			tag = "th"
		}

		if cell.align == "" {
			fmt.Fprintf(&r.builder, "<%s>", tag)
		} else {
			fmt.Fprintf(&r.builder, `<%s style="text-align:%s">`, tag, html.EscapeString(cell.align))
		}

		r.inlines(cell.inlines)
		fmt.Fprintf(&r.builder, "</%s>\n", tag)
	}

	r.builder.WriteString("</tr>\n")
}

func (r *renderer) container(item block) {
	kind := item.containerKind
	if kind == "" {
		kind = "tip"
	}

	title := item.containerTitle
	if title == "" {
		title = strings.ToUpper(kind)
	}

	if kind == "details" {
		r.builder.WriteString(`<details class="custom-block details">` + "\n")
		fmt.Fprintf(&r.builder, `<summary class="custom-block-title">%s</summary>`+"\n", html.EscapeString(title))
		r.blocks(item.children)
		r.builder.WriteString("</details>\n")
		return
	}

	fmt.Fprintf(&r.builder, `<div class="custom-block %s">`+"\n", html.EscapeString(kind))
	fmt.Fprintf(&r.builder, `<p class="custom-block-title">%s</p>`+"\n", html.EscapeString(title))
	r.blocks(item.children)
	r.builder.WriteString("</div>\n")
}

func (r *renderer) inlines(items []inline) {
	for _, item := range items {
		switch item.kind {
		case inlineText:
			r.builder.WriteString(html.EscapeString(item.text))

		case inlineCode:
			r.builder.WriteString("<code>")
			r.builder.WriteString(html.EscapeString(item.text))
			r.builder.WriteString("</code>")

		case inlineEmphasis:
			r.builder.WriteString("<em>")
			r.inlines(item.children)
			r.builder.WriteString("</em>")

		case inlineStrong:
			r.builder.WriteString("<strong>")
			r.inlines(item.children)
			r.builder.WriteString("</strong>")

		case inlineLink:
			r.link(item)

		case inlineImage:
			r.image(item)

		case inlineHTML:
			r.builder.WriteString(item.text)

		case inlineSoftBreak:
			r.builder.WriteString("<br>\n")

		case inlineHardBreak:
			r.builder.WriteString("<br>\n")

		default:
			r.inlines(item.children)
			if item.text != "" {
				r.builder.WriteString(html.EscapeString(item.text))
			}
		}
	}
}

func (r *renderer) link(item inline) {
	destination := r.safeDestination(item.url, false)

	fmt.Fprintf(&r.builder, `<a href="%s"`, html.EscapeString(destination))

	if item.title != "" {
		fmt.Fprintf(&r.builder, ` title="%s"`, html.EscapeString(item.title))
	}

	r.builder.WriteString(">")
	r.inlines(item.children)
	r.builder.WriteString("</a>")
}

func (r *renderer) image(item inline) {
	destination := r.safeDestination(item.url, true)

	fmt.Fprintf(&r.builder, `<img src="%s" alt="%s"`, html.EscapeString(destination), html.EscapeString(plainInlines(item.children)))

	if item.title != "" {
		fmt.Fprintf(&r.builder, ` title="%s"`, html.EscapeString(item.title))
	}

	r.builder.WriteString(">")
}

func (r *renderer) safeDestination(value string, image bool) string {
	if r.resolve != nil {
		value = r.resolve(value, image)
	}

	if safeURL(value, image) {
		return value
	}

	if image {
		return ""
	}

	return "#"
}

func safeURL(value string, image bool) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	switch strings.ToLower(parsed.Scheme) {
	case "", "http", "https":
		return true
	case "mailto":
		return !image
	default:
		return false
	}
}
