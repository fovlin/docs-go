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

func parsePage(page *Page) (*parsedMarkdown, error) {
	data, err := os.ReadFile(page.AbsPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", page.SourcePath, err)
	}

	return parseMarkdown(data, config)
}

func renderPageHTML(page *Page) ([]byte, error) {
	parsed, err := parsePage(page)
	if err != nil {
		return nil, err
	}

	page.Title = parsed.FirstH1
	if page.Title == "" {
		page.Title = humanizeTitle(page.SourcePath)
	}
	page.Description = truncateRunes(strings.TrimSpace(parsed.FirstParagraph), 160)
	if page.Description == "" {
		page.Description = config.Description
	}

	state := &RenderState{Page: page}
	applyNavigation(state, page)
	state.Body = string(renderMarkdown(parsed, func(destination string, image bool) string {
		return resolveDestination(page, destination, image)
	}))
	state.Outline = filterOutline(parsed.Headings, config.Outline.Level)
	state.Nav = buildNav(page)
	state.Sidebar = buildSidebar(page)
	state.EditURL = editURL(page)
	state.LastUpdated = lastUpdated(page)
	fillPageChrome(state)

	return renderLayout(layout, state), nil
}

func renderPage(page *Page) error {
	document, err := renderPageHTML(page)
	if err != nil {
		return err
	}

	return writeOutputFile(outDir, page.OutPath, document)
}

func fillPageChrome(state *RenderState) {
	title := state.Page.Title
	if state.URL == "/" {
		title = config.Title
	} else if !strings.EqualFold(title, config.Title) {
		title += " | " + config.Title
	}

	state.Config = config
	state.UI = config.UI
	state.Title = title
	state.Lang = config.Lang
	state.Appearance = config.Appearance
	state.HomeURL = href("/")
	state.StyleURL = href("/assets/style.css") + "?v=" + assetVersion
	state.InitURL = href("/assets/init.js") + "?v=" + assetVersion
	state.AppURL = href("/assets/app.js") + "?v=" + assetVersion
	state.FaviconURL = href("/favicon.ico") + "?v=" + assetVersion
	state.LogoURL = publicURL(config.Logo)
	state.PrevText = config.PrevText
	state.NextText = config.NextText
	state.FooterMessage = config.Footer.Message
	state.FooterCopyright = strings.ReplaceAll(config.Footer.Copyright, "{year}", fmt.Sprint(time.Now().Year()))
	state.Bootstrap = bootstrap
	state.Version = version

	if config.SiteURL != "" {
		state.Canonical = config.SiteURL + state.Href
	}
	if config.Favicon != "" {
		state.FaviconURL = publicURL(config.Favicon)
	}
}

func publicURL(value string) string {
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

	return href(value)
}

func editURL(page *Page) string {
	if strings.TrimSpace(config.EditLink) == "" {
		return ""
	}

	result := strings.ReplaceAll(config.EditLink, ":path", page.SourcePath)
	result = strings.ReplaceAll(result, ":url", strings.TrimPrefix(href(page.URL), config.Base))

	return result
}

func lastUpdated(page *Page) string {
	if !config.LastUpdated || page.ModTime.IsZero() {
		return ""
	}

	return page.ModTime.Format("2006-01-02")
}

func resolveDestination(page *Page, raw string, image bool) string {
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
		if linked := findPageBySource(strings.TrimPrefix(target, "/")); linked != nil {
			target = linked.URL
		} else {
			logWarn("%s links to a page that does not exist: %s", page.SourcePath, parsed.Path)
			target = markdownFallbackURL(target, config.CleanURLs)
		}
	} else if !image {
		if linked := findPageByURL(target); linked != nil {
			target = linked.URL
		}
	} else if !assetExists(target) {
		logWarn("%s references an image that does not exist: %s", page.SourcePath, parsed.Path)
	}

	return href(target) + urlSuffix(parsed)
}

func assetExists(sitePath string) bool {
	relative := strings.TrimPrefix(sitePath, "/")
	if relative == "" {
		return false
	}

	for _, directory := range []string{srcDir, publicDir} {
		info, err := os.Stat(filepath.Join(directory, filepath.FromSlash(relative)))
		if err == nil && !info.IsDir() {
			return true
		}
	}

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

func notFoundPage() *RenderState {
	state := &RenderState{Page: &Page{
		Title: config.UI.NotFoundTitle,
		URL:   "/404",
		Href:  href("/404"),
	}, Body: `<div class="not-found">` +
		`<p class="not-found-code">404</p>` +
		`<h1>` + html.EscapeString(config.UI.NotFoundTitle) + `</h1>` +
		`<p>` + html.EscapeString(config.UI.NotFoundText) + `</p>` +
		`<p><a class="btn btn-primary" href="` + html.EscapeString(href("/")) + `">` + html.EscapeString(config.UI.BackHome) + `</a></p>` +
		`</div>`,
	}

	fillPageChrome(state)
	state.Title = config.UI.NotFoundTitle + " | " + config.Title
	state.Nav = nil
	state.Sidebar = nil

	return state
}
