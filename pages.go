package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (app *App) findPageByURL(target string) *Page {
	target = canonicalURL(target)
	trailing := strings.HasSuffix(target, "/")
	relative := strings.TrimSuffix(strings.TrimPrefix(target, "/"), ".html")
	relative = strings.TrimSuffix(relative, "/")

	if relative == "" || relative == "index" {
		return app.findSource([]string{"index.md", "index.markdown"})
	}

	candidates := make([]string, 0, 4)
	if trailing {
		candidates = append(candidates, relative+"/index.md", relative+"/index.markdown")
	}

	candidates = append(candidates, relative+".md", relative+".markdown")

	if !trailing {
		candidates = append(candidates, relative+"/index.md", relative+"/index.markdown")
	}

	return app.findSource(candidates)
}

func (app *App) findPageBySource(relative string) *Page {
	relative = strings.TrimPrefix(path.Clean("/"+relative), "/")

	if isMarkdownPath(relative) {
		if page, err := app.loadPage(relative); err == nil {
			return page
		}
	}

	if relative == "" || strings.HasSuffix(relative, "/") {
		relative = strings.TrimSuffix(relative, "/") + "/index"
	}

	return app.findSource([]string{
		relative + ".md",
		relative + ".markdown",
		relative + "/index.md",
		relative + "/index.markdown",
	})
}

func (app *App) findSource(candidates []string) *Page {
	for _, relative := range candidates {
		absolute := filepath.Join(app.SrcDir, filepath.FromSlash(relative))
		if info, err := os.Stat(absolute); err == nil && !info.IsDir() {
			if page, err := app.loadPage(relative); err == nil {
				return page
			}
		}
	}

	return nil
}
