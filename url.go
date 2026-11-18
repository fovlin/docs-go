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
			if node.Href != "" && strings.HasPrefix(node.Href, "/") {
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
	source = strings.TrimPrefix(path.Clean("/"+source), "/")

	switch {
	case source == "" || source == "index":
		return "/"
	case strings.HasSuffix(source, "/index"):
		return "/" + strings.TrimSuffix(source, "/index")
	case app.CleanURLs:
		return "/" + source
	default:
		return "/" + source + ".html"
	}
}

func (app *App) fmtPageURL(source string) (string, string) {
	url := app.fmtURL(source)
	outPath := strings.TrimPrefix(url, "/")

	switch {
	case outPath == "":
		outPath = "index.html"
	case app.CleanURLs || strings.HasSuffix(source, "/index"):
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

	bases := make([]string, 0, 2)
	switch {
	case relative == "" || relative == "index.html":
		bases = append(bases, "index")
	case strings.HasSuffix(relative, "/index.html"):
		dir := strings.TrimSuffix(relative, "/index.html")
		bases = append(bases, dir+"/index", dir)
	case app.CleanURLs:
		bases = append(bases, relative+"/index", relative)
	case strings.HasSuffix(strings.ToLower(relative), ".html"):
		bases = append(bases, relative[:len(relative)-len(".html")])
	default:
		bases = append(bases, relative+"/index")
	}

	for _, base := range bases {
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
