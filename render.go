package main

import (
	"fmt"
	"html"
	"net/url"
	"os"
	"path"
	"strings"

	"acovia.net/docs-go/internal/markdown"
)

func (app *App) parsePage(page *Page) (*markdown.Document, error) {
	data, err := os.ReadFile(page.AbsPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", page.SourcePath, err)
	}

	return markdown.Parse(data), nil
}

func (app *App) renderPageHTML(page *Page) ([]byte, error) {
	document, err := app.parsePage(page)
	if err != nil {
		return nil, err
	}

	page.Heading = document.FirstH1
	if page.Heading == "" {
		page.Heading = humanizeTitle(page.SourcePath)
	}
	page.Description = truncateRunes(strings.TrimSpace(document.FirstParagraph), 160)
	if page.Description == "" {
		page.Description = app.Description
	}

	state := &RenderState{Page: page}
	state.Nav = app.navFor(page.URL)
	state.Sidebar = app.sidebarFor(page.URL)
	app.applyNavigation(state)
	state.Body = string(document.Render(func(destination string, image bool) string {
		return app.resolveDestination(page, destination, image)
	}))
	state.Outline = filterOutline(document.Headings, app.Outline.Level)
	state.EditURL = app.editURL(page)
	state.LastUpdated = app.lastUpdated(page)
	app.fillPageChrome(state)

	return app.renderLayout(state), nil
}

func (app *App) fillPageChrome(state *RenderState) {
	title := state.Page.Heading
	if state.URL == "/" {
		title = app.Title
	} else if !strings.EqualFold(title, app.Title) {
		title += " | " + app.Title
	}

	state.Config = app.Config
	state.UI = app.UI
	state.Title = title
	state.Lang = app.Lang
	state.Appearance = app.Appearance
	state.HomeURL = escapePath("/")
	state.StyleURL = escapePath("/assets/style.css") + "?v=" + app.AssetVersion
	state.InitURL = escapePath("/assets/init.js") + "?v=" + app.AssetVersion
	state.AppURL = escapePath("/assets/app.js") + "?v=" + app.AssetVersion
	state.FaviconURL = escapePath("/favicon.ico") + "?v=" + app.AssetVersion
	state.LogoURL = app.publicURL(app.Logo)
	state.PrevText = app.PrevText
	state.NextText = app.NextText
	state.Bootstrap = app.Bootstrap
	state.Version = version

	if app.SiteURL != "" {
		state.Canonical = strings.TrimSuffix(app.SiteURL, "/") + escapePath(state.URL)
	}
	if app.Config.Favicon != "" {
		state.FaviconURL = app.publicURL(app.Config.Favicon)
	}
}

func (app *App) publicURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "//") {
		return value
	}

	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}

	return escapePath(value)
}

func (app *App) editURL(page *Page) string {
	if strings.TrimSpace(app.EditLink) == "" {
		return ""
	}

	result := strings.ReplaceAll(app.EditLink, ":path", strings.TrimPrefix(escapePath(page.SourcePath), "/"))
	result = strings.ReplaceAll(result, ":url", escapePath(page.URL))

	return result
}

func (app *App) lastUpdated(page *Page) string {
	if !app.LastUpdated || page.ModTime.IsZero() {
		return ""
	}

	return page.ModTime.Format("2006-01-02")
}

func (app *App) resolveDestination(page *Page, raw string, _ bool) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || strings.HasPrefix(raw, "//") || parsed.Path == "" {
		return raw
	}

	target := parsed.Path
	if !strings.HasPrefix(target, "/") {
		target = path.Join("/", path.Dir(page.SourcePath), target)
	}
	target = canonicalURL(target)

	return escapePath(target)
}

func filterOutline(headings []markdown.Heading, levels []int) []markdown.Heading {
	if len(levels) == 0 {
		return nil
	}

	allowed := make(map[int]bool, len(levels))
	for _, level := range levels {
		allowed[level] = true
	}

	outline := make([]markdown.Heading, 0, len(headings))
	for _, heading := range headings {
		if heading.ID != "" && heading.Text != "" && allowed[heading.Level] {
			outline = append(outline, heading)
		}
	}

	return outline
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit])
}

func (app *App) notFoundPage() *RenderState {
	state := &RenderState{Page: &Page{
		Heading: app.UI.NotFoundTitle,
		URL:     "/404",
	}, Body: `<div class="not-found">` +
		`<p class="not-found-code">404</p>` +
		`<h1>` + html.EscapeString(app.UI.NotFoundTitle) + `</h1>` +
		`<p>` + html.EscapeString(app.UI.NotFoundText) + `</p>` +
		`<p><a class="btn btn-primary" href="` + html.EscapeString(escapePath("/")) + `">` + html.EscapeString(app.UI.BackHome) + `</a></p>` +
		`</div>`,
	}

	app.fillPageChrome(state)
	state.Title = app.UI.NotFoundTitle + " | " + app.Title
	state.Nav = nil
	state.Sidebar = nil

	return state
}
