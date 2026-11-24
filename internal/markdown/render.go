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
func (document *Document) Render(resolve ResolveFunc) string {
	r := renderer{resolve: resolve}
	r.blocks(document.blocks)

	return r.builder.String()
}

type renderer struct {
	resolve ResolveFunc
	builder strings.Builder
}

func (r *renderer) blocks(items []block) {
	for _, item := range items {
		item.render(r)
	}
}

func (b *headingBlock) render(r *renderer) {
	fmt.Fprintf(&r.builder, `<h%d id="%s">`, b.level, html.EscapeString(b.id))
	r.inlines(b.inlines)
	fmt.Fprintf(&r.builder, `<a class="header-anchor" href="#%s" aria-label="Link to this heading"></a>`, html.EscapeString(b.id))
	fmt.Fprintf(&r.builder, "</h%d>\n", b.level)
}

func (b *paragraphBlock) render(r *renderer) {
	r.builder.WriteString("<p>")
	r.inlines(b.inlines)
	r.builder.WriteString("</p>\n")
}

func (b *codeBlock) render(r *renderer) {
	r.builder.WriteString(`<div class="code-block"><pre><code`)
	if b.language != "" {
		fmt.Fprintf(&r.builder, ` class="language-%s"`, html.EscapeString(b.language))
	}
	r.builder.WriteString(">")
	r.builder.WriteString(html.EscapeString(b.text))
	r.builder.WriteString("</code></pre></div>\n")
}

func (thematicBreakBlock) render(r *renderer) {
	r.builder.WriteString("<hr>\n")
}

func (b *blockquoteBlock) render(r *renderer) {
	r.builder.WriteString("<blockquote>\n")
	r.blocks(b.children)
	r.builder.WriteString("</blockquote>\n")
}

func (b *listBlock) render(r *renderer) {
	tag := "ul"
	if b.ordered {
		tag = "ol"
	}

	fmt.Fprintf(&r.builder, "<%s", tag)
	if b.ordered && b.start != 1 {
		fmt.Fprintf(&r.builder, ` start="%d"`, b.start)
	}
	r.builder.WriteString(">\n")

	for _, item := range b.items {
		r.builder.WriteString("<li>")
		// A loose list renders every paragraph as a block; a tight list
		// inlines the first paragraph and renders the rest as nested blocks.
		rest := item.children
		if len(rest) > 0 {
			if first, ok := rest[0].(*paragraphBlock); ok && !b.loose {
				r.inlines(first.inlines)
				rest = rest[1:]
				if len(rest) > 0 {
					r.builder.WriteByte('\n')
				}
			}
		}
		r.blocks(rest)
		r.builder.WriteString("</li>\n")
	}

	fmt.Fprintf(&r.builder, "</%s>\n", tag)
}

func (b *tableBlock) render(r *renderer) {
	r.builder.WriteString("<table>\n")

	if len(b.rows) == 0 {
		r.builder.WriteString("</table>\n")
		return
	}

	r.builder.WriteString("<thead>\n")
	b.rows[0].render(r, "th")
	r.builder.WriteString("</thead>\n")

	if len(b.rows) > 1 {
		r.builder.WriteString("<tbody>\n")
		for _, row := range b.rows[1:] {
			row.render(r, "td")
		}
		r.builder.WriteString("</tbody>\n")
	}

	r.builder.WriteString("</table>\n")
}

func (row tableRow) render(r *renderer, tag string) {
	r.builder.WriteString("<tr>\n")
	for _, cell := range row.cells {
		fmt.Fprintf(&r.builder, "<%s", tag)
		if cell.align != "" {
			fmt.Fprintf(&r.builder, ` style="text-align:%s"`, html.EscapeString(cell.align))
		}
		r.builder.WriteString(">")
		r.inlines(cell.inlines)
		fmt.Fprintf(&r.builder, "</%s>\n", tag)
	}
	r.builder.WriteString("</tr>\n")
}

func (b *containerBlock) render(r *renderer) {
	tag, titleTag := "div", "p"
	if b.kind == "details" {
		tag, titleTag = "details", "summary"
	}

	fmt.Fprintf(&r.builder, `<%s class="custom-block %s">`+"\n", tag, html.EscapeString(b.kind))
	fmt.Fprintf(&r.builder, `<%s class="custom-block-title">%s</%s>`+"\n", titleTag, html.EscapeString(b.title), titleTag)
	r.blocks(b.children)
	fmt.Fprintf(&r.builder, "</%s>\n", tag)
}

func (r *renderer) inlines(items []inline) {
	for _, item := range items {
		r.inline(item)
	}
}

func (r *renderer) inline(item inline) {
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
		fmt.Fprintf(&r.builder, `<a href="%s"`, html.EscapeString(r.safeDestination(item.url, false)))
		if item.title != "" {
			fmt.Fprintf(&r.builder, ` title="%s"`, html.EscapeString(item.title))
		}
		r.builder.WriteString(">")
		r.inlines(item.children)
		r.builder.WriteString("</a>")

	case inlineImage:
		fmt.Fprintf(&r.builder, `<img src="%s" alt="%s"`, html.EscapeString(r.safeDestination(item.url, true)), html.EscapeString(plainInlines(item.children)))
		if item.title != "" {
			fmt.Fprintf(&r.builder, ` title="%s"`, html.EscapeString(item.title))
		}
		r.builder.WriteString(">")

	case inlineHTML:
		r.builder.WriteString(item.text)

	case inlineBreak:
		r.builder.WriteString("<br>\n")

	default:
		// Unknown inline kinds keep their children and text rather than
		// disappearing when a new kind is added without a renderer case.
		r.inlines(item.children)
		if item.text != "" {
			r.builder.WriteString(html.EscapeString(item.text))
		}
	}
}

func (r *renderer) safeDestination(value string, image bool) string {
	if r.resolve != nil {
		value = r.resolve(value, image)
	}

	switch {
	case safeURL(value, image):
		return value
	case image:
		return ""
	default:
		return "#"
	}
}

func safeURL(value string, image bool) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return false
	}

	switch strings.ToLower(parsed.Scheme) {
	case "", "http", "https":
		return true
	case "mailto":
		return !image
	}

	return false
}
