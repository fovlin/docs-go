package main

import (
	"fmt"
	"html"
	"strings"

	"acovia.net/docs-go/internal/record"
)

type layoutPart struct {
	text string
	name string
}

func compileLayout(source string) []layoutPart {
	parts := make([]layoutPart, 0, 32)

	for {
		start := strings.Index(source, "{{ ")
		if start < 0 {
			if source != "" {
				return append(parts, layoutPart{text: source})
			}
			return parts
		}

		end := strings.Index(source[start+3:], " }}")
		if end < 0 {
			if source != "" {
				return append(parts, layoutPart{text: source})
			}
			return parts
		}
		end += start + 3

		parts = append(parts, layoutPart{text: source[:start]})
		raw := source[start+3 : end]
		parts = append(parts, layoutPart{name: strings.TrimSpace(strings.TrimPrefix(raw, "@"))})
		source = source[end+3:]
	}
}

func (app *App) renderLayout(page *RenderState) []byte {
	values := layoutValues(page)

	var builder strings.Builder
	for _, part := range app.Layout {

		if part.name == "" {
			builder.WriteString(part.text)
			continue
		}
			
		value, ok := values[part.name]
		if !ok {
			record.Error("unknown field:", part.name)
		}
		
		builder.WriteString(value)
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
	}
}

func headExtra(page *RenderState) string {
	var b strings.Builder
	if page.Canonical != "" {
		fmt.Fprintf(&b, "<link rel=\"canonical\" href=\"%s\">", escapeHTML(page.Canonical))
		fmt.Fprintf(&b, "<meta property=\"og:url\" content=\"%s\">", escapeHTML(page.Canonical))
	}
	if page.FaviconURL != "" {
		fmt.Fprintf(&b, "<link rel=\"icon\" href=\"%s\">", escapeHTML(page.FaviconURL))
	}
	return b.String()
}

func menuButton(page *RenderState) string {
	if len(page.Sidebar) == 0 && len(page.Nav) == 0 {
		return ""
	}

	return fmt.Sprintf("<button class=\"icon-btn menu-btn\" id=\"menu-btn\" type=\"button\" aria-label=\"%s\" aria-expanded=\"false\" aria-controls=\"sidebar\"><svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\"><path d=\"M4 7h16M4 12h16M4 17h16\"/></svg></button>", escapeHTML(page.UI.Menu))
}

func brand(page *RenderState) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<a class=\"brand\" href=\"%s\" aria-label=\"%s\">", escapeHTML(page.HomeURL), escapeHTML(page.Config.Title))

	if page.LogoURL != "" {
		fmt.Fprintf(&b, "<img class=\"brand-logo\" src=\"%s\" alt=\"\">", escapeHTML(page.LogoURL))
	}

	fmt.Fprintf(&b, "<span class=\"brand-name\">%s</span></a>", escapeHTML(page.Config.Title))
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

		b.WriteString(">")
		b.WriteString(escapeHTML(link.Text))
		b.WriteString("</a>")
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
		fmt.Fprintf(&b, "<a class=\"social-link\" href=\"%s\" target=\"_blank\" rel=\"noopener\" aria-label=\"%s\" title=\"%s\">%s</a>", escapeHTML(link.Link), escapeHTML(label), escapeHTML(label), socialIcon(link.Icon))
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
	fmt.Fprintf(&b, "<aside class=\"%s\" id=\"sidebar\"><div class=\"sidebar-inner\">", class)

	if len(page.Sidebar) > 0 {
		b.WriteString("<nav class=\"sidebar-nav\" aria-label=\"Sidebar\">")
		b.WriteString(sidebarItems(page.Sidebar))
		b.WriteString("</nav>")
	}

	if len(page.Nav) > 0 {
		fmt.Fprintf(&b, "<nav class=\"mobile-nav\" aria-label=\"Main\"><p class=\"mobile-nav-title\">%s</p><ul class=\"mobile-nav-list\">", escapeHTML(page.UI.Menu))

		for _, link := range page.Nav {
			class := "mobile-nav-link"
			if link.Active {
				class += " active"
			}

			fmt.Fprintf(&b, "<li class=\"mobile-nav-entry\"><a class=\"%s\" href=\"%s\"", class, escapeHTML(link.Href))
			fmt.Fprintf(&b, ">%s</a></li>", escapeHTML(link.Text))
		}

		b.WriteString("</ul></nav>")
	}

	b.WriteString("</div></aside><div class=\"sidebar-backdrop\" id=\"sidebar-backdrop\" hidden></div>")
	return b.String()
}

func sidebarItems(nodes []*SidebarNode) string {
	var b strings.Builder
	b.WriteString("<ul class=\"sidebar-list\">")

	for _, node := range nodes {
		b.WriteString("<li class=\"sidebar-entry")
		if node.Active {
			b.WriteString(" is-active")
		}
		b.WriteString("\">")

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
			fmt.Fprintf(&b, "<div class=\"sidebar-group%s\" data-open=\"%s\"><div class=\"%s\"%s>", openClass(node), open, rowClass, rowMeta)
			arrow := "<span class=\"sidebar-arrow\" aria-hidden=\"true\"><svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"m9 6 6 6-6 6\"/></svg></span>"
			if node.Href != "" {
				active := ""
				if node.Active {
					active = " active"
				}
				fmt.Fprintf(&b, "<a class=\"sidebar-group-link%s\" href=\"%s\">%s", active, escapeHTML(node.Href), escapeHTML(node.Text))
				b.WriteString(arrow)
				b.WriteString("</a>")
			} else {
				fmt.Fprintf(&b, "<span class=\"sidebar-group-label\">%s</span>", escapeHTML(node.Text))
				b.WriteString(arrow)
			}
			b.WriteString("</div><div class=\"sidebar-group-items\">")
			b.WriteString(sidebarItems(node.Items))
			b.WriteString("</div></div>")
		} else {
			active := ""
			if node.Active {
				active = " active"
			}
			fmt.Fprintf(&b, "<a class=\"sidebar-link%s\" href=\"%s\"", active, escapeHTML(node.Href))

			fmt.Fprintf(&b, "><span>%s</span>", escapeHTML(node.Text))
			b.WriteString("</a>")
		}

		b.WriteString("</li>")
	}

	b.WriteString("</ul>")
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
	b.WriteString("<div class=\"doc-meta\">")
	if page.EditURL != "" {
		fmt.Fprintf(&b, "<a class=\"edit-link\" href=\"%s\" target=\"_blank\" rel=\"noopener\"><svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.7\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M4 20h4L19 9a2.8 2.8 0 0 0-4-4L4 16v4Z\"/><path d=\"m13.5 6.5 4 4\"/></svg>%s</a>", escapeHTML(page.EditURL), escapeHTML(page.UI.EditPage))
	}
	if page.LastUpdated != "" {
		fmt.Fprintf(&b, "<span class=\"last-updated\">%s %s</span>", escapeHTML(page.UI.LastUpdated), escapeHTML(page.LastUpdated))
	}
	b.WriteString("</div>")
	return b.String()
}

func prevNext(page *RenderState) string {
	if !page.Config.PrevNext || (page.Prev == nil && page.Next == nil) {
		return ""
	}

	var b strings.Builder
	b.WriteString("<nav class=\"prev-next\" aria-label=\"Pages\">")
	if page.Prev != nil {
		fmt.Fprintf(&b, "<a class=\"prev-next-link prev\" href=\"%s\"><span class=\"prev-next-label\"><svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M19 12H5M11 6l-6 6 6 6\"/></svg>%s</span><span class=\"prev-next-title\">%s</span></a>", escapeHTML(page.Prev.Href), escapeHTML(page.PrevText), escapeHTML(page.Prev.Text))
	}
	if page.Next != nil && page.Prev == nil {
		b.WriteString("<span class=\"prev-next-placeholder\" aria-hidden=\"true\"></span>")
	}
	if page.Next != nil {
		fmt.Fprintf(&b, "<a class=\"prev-next-link next\" href=\"%s\"><span class=\"prev-next-label\">%s<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M5 12h14M13 6l6 6-6 6\"/></svg></span><span class=\"prev-next-title\">%s</span></a>", escapeHTML(page.Next.Href), escapeHTML(page.NextText), escapeHTML(page.Next.Text))
	}
	b.WriteString("</nav>")
	return b.String()
}

func outlineList(page *RenderState) string {
	if len(page.Outline) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("<ul class=\"outline-list\">")
	fmt.Fprintf(&b, "<li class=\"outline-item\"><a href=\"#\">%s</a></li>", escapeHTML(page.Heading))
	for _, heading := range page.Outline {
		fmt.Fprintf(&b, "<li class=\"outline-item outline-level-%d\"><a href=\"#%s\">%s</a></li>", heading.Level, escapeHTML(heading.ID), escapeHTML(heading.Text))
	}
	b.WriteString("</ul>")
	return b.String()
}

func outlineHTML(page *RenderState) string {
	if len(page.Outline) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("<aside class=\"outline\">")
	fmt.Fprintf(&b, "<p class=\"outline-title\">%s</p>", escapeHTML(page.Config.Outline.Label))
	b.WriteString(outlineList(page))
	b.WriteString("</aside>")
	return b.String()
}

func pageMenuButton(page *RenderState) string {
	if len(page.Outline) == 0 && len(page.Config.SocialLinks) == 0 {
		return ""
	}

	return "<button class=\"icon-btn page-menu-btn\" id=\"page-menu-btn\" type=\"button\" aria-label=\"Page menu\" aria-expanded=\"false\" aria-controls=\"page-menu\"><svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\"><path d=\"M4 7h16M4 12h16M4 17h16\"/></svg></button>"
}

func pageMenu(page *RenderState) string {
	if len(page.Outline) == 0 && len(page.Config.SocialLinks) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("<div class=\"page-menu\" id=\"page-menu\" hidden>")
	if len(page.Outline) > 0 {
		fmt.Fprintf(&b, "<div class=\"page-menu-section\"><p class=\"page-menu-title\">%s</p>", escapeHTML(page.Config.Outline.Label))
		b.WriteString(outlineList(page))
		b.WriteString("</div>")
	}
	if len(page.Config.SocialLinks) > 0 {
		b.WriteString("<div class=\"page-menu-social\">")
		b.WriteString(socialLinks(page))
		b.WriteString("</div>")
	}
	b.WriteString("</div>")
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
