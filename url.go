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
			if node.Href != "" && node.Path != "" && !node.External {
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

func (app *App) fileForURL(sitePath string) (string, bool) {
	relative := strings.TrimPrefix(path.Clean("/"+sitePath), "/")

	if app.CleanURLs && path.Base(relative) == "index.html" {
		relative = strings.TrimSuffix(strings.TrimSuffix(relative, "index.html"), "/")
	}

	bases := make([]string, 0, 2)
	switch {
	case relative == "":
		bases = append(bases, "index")
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
			if info, err := os.Stat(file); err == nil && !info.IsDir() {
				return file, true
			}
		}
	}

	names := []string{relative, path.Join(relative, "index.html")}
	for _, root := range []string{app.SrcDir, app.PublicDir} {
		for _, name := range names {
			file := filepath.Join(root, filepath.FromSlash(name))
			info, err := os.Stat(file)
			if err != nil || info.IsDir() || isMarkdownFile(file) && isWithin(app.SrcDir, file) {
				continue
			}
			return file, true
		}
	}

	return "", false
}
