package main

import (
	"net/url"
	"path"
	"strings"
)

func (app *App) href(sitePath string) string {
	if sitePath == "" {
		sitePath = "/"
	}

	if !strings.HasPrefix(sitePath, "/") {
		sitePath = "/" + sitePath
	}

	escaped := (&url.URL{Path: sitePath}).EscapedPath()
	if app.Base == "/" {
		return escaped
	}

	return strings.TrimSuffix(app.Base, "/") + escaped
}

func canonicalURL(target string) string {
	trailing := strings.HasSuffix(target, "/")
	cleaned := path.Clean("/" + target)
	if cleaned == "." {
		cleaned = "/"
	}

	if trailing && cleaned != "/" {
		cleaned += "/"
	}

	return cleaned
}

// resolveConfigLink resolves one configured navbar/sidebar link. It is used
// only while loading an App, never while rendering a page.
func (app *App) resolveConfigLink(raw string) (string, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || strings.HasPrefix(raw, "//") {
		return raw, "", true
	}

	if parsed.Path == "" {
		return raw, "", false
	}

	target := parsed.Path
	if !strings.HasPrefix(target, "/") {
		target = "/" + target
	}
	target = canonicalURL(target)

	if page := app.findPageByURL(target); page != nil {
		target = page.URL
	} else if page := app.findPageBySource(strings.TrimPrefix(target, "/")); page != nil {
		target = page.URL
	} else if isMarkdownPath(target) {
		target = markdownFallbackURL(target, app.CleanURLs)
	}

	return app.href(target) + urlSuffix(parsed), target, false
}

func urlSuffix(parsed *url.URL) string {
	var builder strings.Builder

	if parsed.RawQuery != "" || parsed.ForceQuery {
		builder.WriteByte('?')
		builder.WriteString(parsed.RawQuery)
	}

	if parsed.Fragment != "" {
		builder.WriteByte('#')
		builder.WriteString(parsed.Fragment)
	}

	return builder.String()
}
