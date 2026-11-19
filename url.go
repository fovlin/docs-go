package main

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (app *App) applyNavigation(state *RenderState) {
	links := make([]*SidebarNode, 0)

	var walk func([]*SidebarNode)
	walk = func(nodes []*SidebarNode) {
		for _, node := range nodes {
			if node.Href != "" && strings.HasPrefix(node.Href, "/") && len(node.Items) == 0 {
				links = append(links, node)
			}
			walk(node.Items)
		}
	}
	walk(state.Sidebar)

	for index, node := range links {
		if node.Href != state.URL {
			continue
		}
		if index > 0 {
			prev := links[index-1]
			state.Prev = &PageLink{Text: prev.Text, Href: prev.Href}
		}
		if index < len(links)-1 {
			next := links[index+1]
			state.Next = &PageLink{Text: next.Text, Href: next.Href}
		}
		return
	}
}

func (app *App) fmtURL(source string) string {
	parsed, err := url.Parse(source)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.Path == "" {
		return source
	}

	source = strings.TrimPrefix(path.Clean(source), "/")

	if app.CleanURLs || source == "" || strings.HasSuffix(parsed.Path, "/") {
		return "/" + source
	}

	return "/" + source + ".html"
}

func (app *App) fmtPagePathInfo(source string) (string, string) {
	if source == "" || source == "index" {
		return "/", "index.html"
	}

	if strings.HasSuffix(source, "/index") {
		source = strings.TrimSuffix(source, "/index")
		return app.fmtURL(source + "/"), source + "/index.html"
	}

	url := app.fmtURL(source)
	outPath := strings.TrimPrefix(url, "/")

	if app.CleanURLs {
		outPath += "/index.html"
	}

	return url, outPath
}

func fmtConfigPath(fsPath string) string {
	fsPath = strings.TrimSpace(fsPath)
	if fsPath == "" {
		return "."
	}

	return filepath.Clean(fsPath)
}

func (app *App) fileForURL(sitePath string) (string, []byte, error) {
	relative := strings.TrimPrefix(path.Clean("/"+sitePath), "/")

	switch {
	case strings.HasSuffix(relative, "/index.html") || relative == "index.html":
		relative = strings.TrimSuffix(strings.TrimSuffix(relative, ".html"), "/index")
	case !app.CleanURLs && strings.HasSuffix(strings.ToLower(relative), ".html"):
		relative = relative[:len(relative)-len(".html")]
	}

	for _, base := range []string{path.Join(relative, "index"), relative} {
		for _, extension := range []string{".md", ".markdown", ".MD", ".Markdown"} {
			file := filepath.Join(app.SrcDir, filepath.FromSlash(base+extension))
			if data, err := os.ReadFile(file); err == nil {
				return file, data, nil
			}
		}
	}

	names := []string{relative, path.Join(relative, "index.html")}
	if relative == "" {
		names = []string{"index.html"}
	}

	for _, root := range []string{app.SrcDir, app.PublicDir} {
		for _, name := range names {
			file := filepath.Join(root, filepath.FromSlash(name))
			if isMarkdownFile(file) && isWithin(app.SrcDir, file) {
				continue
			}
			if data, err := os.ReadFile(file); err == nil {
				return file, data, nil
			}
		}
	}

	return "", nil, os.ErrNotExist
}

func escapePath(sitePath string) string {
	if sitePath == "" {
		sitePath = "/"
	}

	if !strings.HasPrefix(sitePath, "/") {
		sitePath = "/" + sitePath
	}

	return (&url.URL{Path: sitePath}).EscapedPath()
}
