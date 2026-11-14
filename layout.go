package main

import (
	"fmt"
	"html"
	"strings"
)

type layoutPart struct {
	text string
	name string
}

// compileLayout splits the static template once. Rendering only concatenates
// text and values; it does not build a replacer or scan the template per page.
func compileLayout(source string) []layoutPart {
	parts := make([]layoutPart, 0, 32)

	for {
		start := strings.Index(source, "{{")
		if start < 0 {
			return append(parts, layoutPart{text: source})
		}

		end := strings.Index(source[start+2:], "}}")
		if end < 0 {
			return append(parts, layoutPart{text: source})
		}
		end += start + 2

		if start > 0 {
			parts = append(parts, layoutPart{text: source[:start]})
		}

		raw := strings.TrimSpace(source[start+2 : end])
		parts = append(parts, layoutPart{name: strings.TrimSpace(strings.TrimPrefix(raw, "@"))})
		source = source[end+2:]
	}
}

func (app *App) renderLayout(page *RenderState) []byte {
	values := layoutValues(page)

	var builder strings.Builder
	for _, part := range app.Layout {
		if part.name == "" {
			builder.WriteString(part.text)
		} else if value, ok := values[part.name]; ok {
			builder.WriteString(value)
		}
	}

	return []byte(builder.String())
}

// Dy generate text with page value
func layoutValues(page *RenderState) map[string]string {
	ogType := "article"
	if page.URL == "/" {
		ogType = "website"
	}

	return map[string]string{
		"Lang":             escapeHTML(page.Lang),
		"Appearance":       escapeHTML(page.Appearance),
		"Version":          escapeHTML(page.Version),
		"Title":            escapeHTML(page.Title),
		"Description":      escapeHTML(page.Description),
		"HeadExtra":        headExtra(page),
		"OGType":           ogType,
		"SiteTitle":        escapeHTML(page.Config.Title),
		"InitURL":          escapeHTML(page.InitURL),
		"StyleURL":         escapeHTML(page.StyleURL),
		"AppURL":           escapeHTML(page.AppURL),
		"Bootstrap":        page.Bootstrap,
		"SkipToContent":    escapeHTML(page.UI.SkipToContent),
		"MenuButton":       menuButton(page),
		"Brand":            brand(page),
		"Nav":              navigation(page),
		"SocialLinks":      socialLinks(page),
		"PageMenuButton":   pageMenuButton(page),
		"PageMenu":         pageMenu(page),
		"ThemeLabel":       escapeHTML(page.UI.Theme),
		"ShellClass":       shellClass(page),
		"Sidebar":          sidebarHTML(page),
		"ContentGridClass": contentGridClass(page),
		"PageBody":         page.Body,
		"DocMeta":          docMeta(page),
		"PrevNext":         prevNext(page),
		"Outline":          outlineHTML(page),
		"FooterText":       footerText(page),
	}
}

func headExtra(page *RenderState) string {
	var b strings.Builder
	if page.Canonical != "" {
		fmt.Fprintf(&b, "<link rel=\"canonical\" href=\"%s\">\n", escapeHTML(page.Canonical))
		fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"%s\">\n", escapeHTML(page.Canonical))
	}
	if page.FaviconURL != "" {
		fmt.Fprintf(&b, "<link rel=\"icon\" href=\"%s\">\n", escapeHTML(page.FaviconURL))
	}
	return b.String()
}

func menuButton(page *RenderState) string {
	if len(page.Sidebar) == 0 && len(page.Nav) == 0 {
		return ""
	}

	return fmt.Sprintf("<button class=\"icon-btn menu-btn\" id=\"menu-btn\" type=\"button\" aria-label=\"%s\" aria-expanded=\"false\" aria-controls=\"sidebar\">\n<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\"><path d=\"M4 7h16M4 12h16M4 17h16\"/></svg>\n</button>\n", escapeHTML(page.UI.Menu))
}

func brand(page *RenderState) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<a class=\"brand\" href=\"%s\" aria-label=\"%s\">", escapeHTML(page.HomeURL), escapeHTML(page.Config.Title))

	if page.LogoURL != "" {
		fmt.Fprintf(&b, "<img class=\"brand-logo\" src=\"%s\" alt=\"\">", escapeHTML(page.LogoURL))
	}

	fmt.Fprintf(&b, "<span class=\"brand-name\">%s</span></a>\n", escapeHTML(page.Config.Title))
	return b.String()
}

func navigation(page *RenderState) string {
	var b strings.Builder
	for _, link := range page.Nav {
		class := "nav-link"
		if link.Active {
			class += " active"
		}

		fmt.Fprintf(&b, "<a class=\"%s\" href=\"%s\"", class, escapeHTML(link.Href))
		if link.External {
			b.WriteString(" target=\"_blank\" rel=\"noopener\"")
		}
		b.WriteString(">")
		b.WriteString(escapeHTML(link.Text))
		if link.External {
			b.WriteString("<svg class=\"external-icon\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M7 17 17 7M9 7h8v8\"/></svg>")
		}
		b.WriteString("</a>\n")
	}
	return b.String()
}

func socialLinks(page *RenderState) string {
	var b strings.Builder
	for _, link := range page.Config.SocialLinks {
		label := strings.TrimSpace(link.Label)
		if label == "" {
			label = link.Icon
		}
		fmt.Fprintf(&b, "<a class=\"social-link\" href=\"%s\" target=\"_blank\" rel=\"noopener\" aria-label=\"%s\" title=\"%s\">%s</a>\n", escapeHTML(link.Link), escapeHTML(label), escapeHTML(label), socialIcon(link.Icon))
	}
	return b.String()
}

func sidebarHTML(page *RenderState) string {
	if len(page.Sidebar) == 0 && len(page.Nav) == 0 {
		return ""
	}

	class := "sidebar"
	if len(page.Sidebar) == 0 {
		class += " mobile-only"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<aside class=\"%s\" id=\"sidebar\">\n<div class=\"sidebar-inner\">\n", class)

	if len(page.Sidebar) > 0 {
		b.WriteString("<nav class=\"sidebar-nav\" aria-label=\"Sidebar\">\n")
		b.WriteString(sidebarItems(page.Sidebar))
		b.WriteString("</nav>\n")
	}

	if len(page.Nav) > 0 {
		fmt.Fprintf(&b, "<nav class=\"mobile-nav\" aria-label=\"Main\">\n<p class=\"mobile-nav-title\">%s</p>\n", escapeHTML(page.UI.Menu))
		for _, link := range page.Nav {
			class := "mobile-nav-link"
			if link.Active {
				class += " active"
			}
			fmt.Fprintf(&b, "<a class=\"%s\" href=\"%s\"", class, escapeHTML(link.Href))
			if link.External {
				b.WriteString(" target=\"_blank\" rel=\"noopener\"")
			}
			fmt.Fprintf(&b, ">%s</a>\n", escapeHTML(link.Text))
		}
		b.WriteString("</nav>\n")
	}

	b.WriteString("</div>\n</aside>\n<div class=\"sidebar-backdrop\" id=\"sidebar-backdrop\" hidden></div>\n")
	return b.String()
}

func sidebarItems(nodes []*SidebarNode) string {
	var b strings.Builder
	b.WriteString("<ul class=\"sidebar-list\">\n")

	for _, node := range nodes {
		b.WriteString("<li class=\"sidebar-entry")
		if node.Active || node.HasActive {
			b.WriteString(" is-active")
		}
		b.WriteString("\">\n")

		if len(node.Items) > 0 {
			open := "false"
			if node.Open {
				open = "true"
			}
			rowClass := "sidebar-group-row"
			rowMeta := ""
			if node.Href == "" {
				rowClass += " is-toggle"
				rowMeta = fmt.Sprintf(" role=\"button\" tabindex=\"0\" aria-expanded=\"%s\"", open)
			}
			fmt.Fprintf(&b, "<div class=\"sidebar-group%s\" data-open=\"%s\">\n<div class=\"%s\"%s>\n", openClass(node), open, rowClass, rowMeta)
			if node.Href != "" {
				active := ""
				if node.Active {
					active = " active"
				}
				fmt.Fprintf(&b, "<a class=\"sidebar-group-link%s\" href=\"%s\"", active, escapeHTML(node.Href))
				if node.External {
					b.WriteString(" target=\"_blank\" rel=\"noopener\"")
				}
				fmt.Fprintf(&b, ">%s</a>\n", escapeHTML(node.Text))
			} else {
				fmt.Fprintf(&b, "<span class=\"sidebar-group-label\">%s</span>\n", escapeHTML(node.Text))
			}
			b.WriteString("<span class=\"sidebar-arrow\" aria-hidden=\"true\"><svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"m9 6 6 6-6 6\"/></svg></span>\n")
			b.WriteString("</div>\n<div class=\"sidebar-group-items\">\n")
			b.WriteString(sidebarItems(node.Items))
			b.WriteString("</div>\n</div>\n")
		} else {
			active := ""
			if node.Active {
				active = " active"
			}
			fmt.Fprintf(&b, "<a class=\"sidebar-link%s\" href=\"%s\"", active, escapeHTML(node.Href))
			if node.External {
				b.WriteString(" target=\"_blank\" rel=\"noopener\"")
			}
			fmt.Fprintf(&b, "><span>%s</span>", escapeHTML(node.Text))
			if node.External {
				b.WriteString("<svg class=\"external-icon\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M7 17 17 7M9 7h8v8\"/></svg>")
			}
			b.WriteString("</a>\n")
		}

		b.WriteString("</li>\n")
	}

	b.WriteString("</ul>\n")
	return b.String()
}

func openClass(node *SidebarNode) string {
	if node.Open {
		return " is-open"
	}
	return ""
}

func docMeta(page *RenderState) string {
	if page.EditURL == "" && page.LastUpdated == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("<div class=\"doc-meta\">\n")
	if page.EditURL != "" {
		fmt.Fprintf(&b, "<a class=\"edit-link\" href=\"%s\" target=\"_blank\" rel=\"noopener\">\n<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.7\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M4 20h4L19 9a2.8 2.8 0 0 0-4-4L4 16v4Z\"/><path d=\"m13.5 6.5 4 4\"/></svg>\n%s\n</a>\n", escapeHTML(page.EditURL), escapeHTML(page.UI.EditPage))
	}
	if page.LastUpdated != "" {
		fmt.Fprintf(&b, "<span class=\"last-updated\">%s %s</span>\n", escapeHTML(page.UI.LastUpdated), escapeHTML(page.LastUpdated))
	}
	b.WriteString("</div>\n")
	return b.String()
}

func prevNext(page *RenderState) string {
	if !page.Config.PrevNext || (page.Prev == nil && page.Next == nil) {
		return ""
	}

	var b strings.Builder
	b.WriteString("<nav class=\"prev-next\" aria-label=\"Pages\">\n")
	if page.Prev != nil {
		fmt.Fprintf(&b, "<a class=\"prev-next-link prev\" href=\"%s\">\n<span class=\"prev-next-label\"><svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M19 12H5M11 6l-6 6 6 6\"/></svg>%s</span>\n<span class=\"prev-next-title\">%s</span>\n</a>\n", escapeHTML(page.Prev.Href), escapeHTML(page.PrevText), escapeHTML(page.Prev.Text))
	}
	if page.Next != nil && page.Prev == nil {
		b.WriteString("<span class=\"prev-next-placeholder\" aria-hidden=\"true\"></span>\n")
	}
	if page.Next != nil {
		fmt.Fprintf(&b, "<a class=\"prev-next-link next\" href=\"%s\">\n<span class=\"prev-next-label\">%s<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M5 12h14M13 6l6 6-6 6\"/></svg></span>\n<span class=\"prev-next-title\">%s</span>\n</a>\n", escapeHTML(page.Next.Href), escapeHTML(page.NextText), escapeHTML(page.Next.Text))
	}
	b.WriteString("</nav>\n")
	return b.String()
}

func outlineList(page *RenderState) string {
	if len(page.Outline) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("<ul class=\"outline-list\">\n")
	fmt.Fprintf(&b, "<li class=\"outline-item\"><a href=\"#\">%s</a></li>\n", escapeHTML(page.Heading))
	for _, heading := range page.Outline {
		fmt.Fprintf(&b, "<li class=\"outline-item outline-level-%d\"><a href=\"#%s\">%s</a></li>\n", heading.Level, escapeHTML(heading.ID), escapeHTML(heading.Text))
	}
	b.WriteString("</ul>\n")
	return b.String()
}

func outlineHTML(page *RenderState) string {
	if len(page.Outline) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("<aside class=\"outline\">\n")
	fmt.Fprintf(&b, "<p class=\"outline-title\">%s</p>\n", escapeHTML(page.Config.Outline.Label))
	b.WriteString(outlineList(page))
	b.WriteString("</aside>\n")
	return b.String()
}

func pageMenuButton(page *RenderState) string {
	if len(page.Outline) == 0 && len(page.Config.SocialLinks) == 0 {
		return ""
	}

	return "<button class=\"icon-btn page-menu-btn\" id=\"page-menu-btn\" type=\"button\" aria-label=\"Page menu\" aria-expanded=\"false\" aria-controls=\"page-menu\">\n<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\"><path d=\"M4 7h16M4 12h16M4 17h16\"/></svg>\n</button>\n"
}

func pageMenu(page *RenderState) string {
	if len(page.Outline) == 0 && len(page.Config.SocialLinks) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("<div class=\"page-menu\" id=\"page-menu\" hidden>\n")
	if len(page.Outline) > 0 {
		fmt.Fprintf(&b, "<div class=\"page-menu-section\"><p class=\"page-menu-title\">%s</p>\n", escapeHTML(page.Config.Outline.Label))
		b.WriteString(outlineList(page))
		b.WriteString("</div>\n")
	}
	if len(page.Config.SocialLinks) > 0 {
		b.WriteString("<div class=\"page-menu-social\">\n")
		b.WriteString(socialLinks(page))
		b.WriteString("</div>\n")
	}
	b.WriteString("</div>\n")
	return b.String()
}

func footerText(page *RenderState) string {
	var b strings.Builder
	if page.FooterMessage != "" {
		fmt.Fprintf(&b, "<p>%s</p>\n", escapeHTML(page.FooterMessage))
	}
	if page.FooterCopyright != "" {
		fmt.Fprintf(&b, "<p>%s</p>\n", escapeHTML(page.FooterCopyright))
	}
	return b.String()
}

func shellClass(page *RenderState) string {
	if len(page.Sidebar) == 0 {
		return "shell no-sidebar"
	}
	return "shell"
}

func contentGridClass(page *RenderState) string {
	if len(page.Outline) == 0 {
		return "content-grid no-outline"
	}
	return "content-grid"
}

func escapeHTML(value string) string {
	return html.EscapeString(value)
}
