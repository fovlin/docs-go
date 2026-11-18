package main

import (
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
			if node.Href != "" && node.Path != "" {
				links = append(links, node)
			}
			walk(node.Items)
		}
	}
	walk(state.Sidebar)

	for index, node := range links {
		if node.Path != state.URL {
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

func (app *App) fileForURL(sitePath string) (string, []byte, error) {
	relative := strings.TrimPrefix(path.Clean("/"+sitePath), "/")

	if app.CleanURLs && path.Base(relative) == "index.html" {
		relative = strings.TrimSuffix(strings.TrimSuffix(relative, "index.html"), "/")
	}

	bases := make([]string, 0, 2)
	switch {
	case relative == "":
		bases = append(bases, "index")

	case app.CleanURLs && strings.HasSuffix(relative, "/"):
		bases = append(bases, relative+"/index", relative)
	case app.CleanURLs && strings.HasSuffix(relative, "/index.html"):
		bases = append(bases, relative[:len("index.html")])

	case strings.HasSuffix(strings.ToLower(relative), ".html"):
		bases = append(bases, relative[:len(relative)-len(".html")])
	default:
		bases = append(bases, relative+"/index")
	}

	namelist := make([]string, 0, 2)

	for _, base := range bases {
		for _, extension := range []string{".md", ".markdown", ".MD", ".Markdown"} {
			namelist = append(namelist, base + extension)
		}
	}

	for _, root := range []string{app.SrcDir, app.PublicDir} {
		for _, name := range namelist {

			filename := filepath.Join(root, filepath.FromSlash(name))
			info, err := os.Stat(filename)
			if err != nil || info.IsDir() || !isWithin(app.SrcDir, filename) {
				continue
			}

			fileData, err := os.ReadFile(filename)
			return filename, fileData, err
		}
	}

	return "", nil, os.ErrNotExist
}

func fmtURL(path string) (string) {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	if path == "" {
		return "/"
	}
	return "/" + strings.Trim(path, "/") + "/"
}

func (app *App) fmtPageURL(filePath string) (string) {
	switch {
	case app.CleanURLs:
		return fmtURL(filePath) + "index.html"
	default:
		return fmtURL(filePath)
	}
}