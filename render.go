package main

import (
	"fmt"
	"html"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func (app *App) parsePage(page *Page) (*parsedMarkdown, error) {
	data, err := os.ReadFile(page.AbsPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", page.SourcePath, err)
	}

	return parseMarkdown(data, app.Config)
}

func (app *App) renderPageHTML(page *Page) ([]byte, error) {
	parsed, err := app.parsePage(page)
	if err != nil {
		return nil, err
	}

	page.Heading = parsed.FirstH1
	if page.Heading == "" {
		page.Heading = humanizeTitle(page.SourcePath)
	}
	page.Description = truncateRunes(strings.TrimSpace(parsed.FirstParagraph), 160)
	if page.Description == "" {
		page.Description = app.Description
	}

	state := &RenderState{Page: page}
	app.applyNavigation(state, page)
	state.Body = string(renderMarkdown(parsed, func(destination string, image bool) string {
		return app.resolveDestination(page, destination, image)
	}))
	state.Outline = filterOutline(parsed.Headings, app.Outline.Level)
	state.Nav = app.navFor(page.URL)
	state.Sidebar = app.sidebarFor(page.URL)
	state.EditURL = app.editURL(page)
	state.LastUpdated = app.lastUpdated(page)
	app.fillPageChrome(state)

	return app.renderLayout(state), nil
}

func (app *App) renderPage(page *Page) error {
	document, err := app.renderPageHTML(page)
	if err != nil {
		return err
	}

	return app.writeOutputFile(page.OutPath, document)
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
	state.HomeURL = app.href("/")
	state.StyleURL = app.href("/assets/style.css") + "?v=" + app.AssetVersion
	state.InitURL = app.href("/assets/init.js") + "?v=" + app.AssetVersion
	state.AppURL = app.href("/assets/app.js") + "?v=" + app.AssetVersion
	state.FaviconURL = app.href("/favicon.ico") + "?v=" + app.AssetVersion
	state.LogoURL = app.publicURL(app.Logo)
	state.PrevText = app.PrevText
	state.NextText = app.NextText
	state.FooterMessage = app.Footer.Message
	state.FooterCopyright = strings.ReplaceAll(app.Footer.Copyright, "{year}", fmt.Sprint(time.Now().Year()))
	state.Bootstrap = app.Bootstrap
	state.Version = version

	if app.SiteURL != "" {
		state.Canonical = app.SiteURL + state.Href
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

	return app.href(value)
}

func (app *App) editURL(page *Page) string {
	if strings.TrimSpace(app.EditLink) == "" {
		return ""
	}

	result := strings.ReplaceAll(app.EditLink, ":path", page.SourcePath)
	result = strings.ReplaceAll(result, ":url", strings.TrimPrefix(app.href(page.URL), app.Base))

	return result
}

func (app *App) lastUpdated(page *Page) string {
	if !app.LastUpdated || page.ModTime.IsZero() {
		return ""
	}

	return page.ModTime.Format("2006-01-02")
}

func (app *App) resolveDestination(page *Page, raw string, image bool) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if parsed.Scheme != "" || parsed.Host != "" || strings.HasPrefix(raw, "//") {
		return raw
	}

	if parsed.Path == "" {
		return raw
	}

	target := parsed.Path
	if !strings.HasPrefix(target, "/") {
		target = path.Join(path.Dir(page.URL), target)
	}
	target = canonicalURL(target)

	if isMarkdownPath(target) {
		if linked := app.findPageBySource(strings.TrimPrefix(target, "/")); linked != nil {
			target = linked.URL
		} else {
			logWarn("%s links to a page that does not exist: %s", page.SourcePath, parsed.Path)
			target = markdownFallbackURL(target, app.CleanURLs)
		}
	} else if !image {
		if linked := app.findPageByURL(target); linked != nil {
			target = linked.URL
		}
	} else if !app.assetExists(target) {
		logWarn("%s references an image that does not exist: %s", page.SourcePath, parsed.Path)
	}

	return app.href(target) + urlSuffix(parsed)
}

func (app *App) assetExists(sitePath string) bool {
	relative := strings.TrimPrefix(sitePath, "/")
	if relative == "" {
		return false
	}

	app.lookupMu.Lock()
	exists, cached := app.assetCache[relative]
	app.lookupMu.Unlock()
	if cached {
		return exists
	}

	for _, directory := range []string{app.SrcDir, app.PublicDir} {
		info, err := os.Stat(filepath.Join(directory, filepath.FromSlash(relative)))
		if err == nil && !info.IsDir() {
			app.lookupMu.Lock()
			app.assetCache[relative] = true
			app.lookupMu.Unlock()
			return true
		}
	}

	app.lookupMu.Lock()
	app.assetCache[relative] = false
	app.lookupMu.Unlock()
	return false
}

func isMarkdownPath(value string) bool {
	extension := strings.ToLower(path.Ext(value))
	return extension == ".md" || extension == ".markdown"
}

func markdownFallbackURL(value string, cleanURLs bool) string {
	extension := strings.ToLower(path.Ext(value))
	if extension != ".md" && extension != ".markdown" {
		return value
	}

	if cleanURLs {
		return strings.TrimSuffix(value, path.Ext(value))
	}

	return strings.TrimSuffix(value, path.Ext(value)) + ".html"
}

func filterOutline(headings []OutlineEntry, levels []int) []OutlineEntry {
	if len(levels) == 0 {
		return nil
	}

	allowed := make(map[int]bool, len(levels))
	for _, level := range levels {
		allowed[level] = true
	}

	outline := make([]OutlineEntry, 0, len(headings))
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
		Href:    app.href("/404"),
	}, Body: `<div class="not-found">` +
		`<p class="not-found-code">404</p>` +
		`<h1>` + html.EscapeString(app.UI.NotFoundTitle) + `</h1>` +
		`<p>` + html.EscapeString(app.UI.NotFoundText) + `</p>` +
		`<p><a class="btn btn-primary" href="` + html.EscapeString(app.href("/")) + `">` + html.EscapeString(app.UI.BackHome) + `</a></p>` +
		`</div>`,
	}

	app.fillPageChrome(state)
	state.Title = app.UI.NotFoundTitle + " | " + app.Title
	state.Nav = nil
	state.Sidebar = nil

	return state
}
