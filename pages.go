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
	key := strings.Join(candidates, "\x00")

	app.lookupMu.Lock()
	relative, cached := app.sourceCache[key]
	app.lookupMu.Unlock()
	if cached {
		if relative == "" {
			return nil
		}
		page, _ := app.loadPage(relative)
		return page
	}

	for _, candidate := range candidates {
		absolute := filepath.Join(app.SrcDir, filepath.FromSlash(candidate))
		if info, err := os.Stat(absolute); err == nil && !info.IsDir() {
			app.lookupMu.Lock()
			app.sourceCache[key] = candidate
			app.lookupMu.Unlock()
			page, _ := app.loadPage(candidate)
			return page
		}
	}

	app.lookupMu.Lock()
	app.sourceCache[key] = ""
	app.lookupMu.Unlock()
	return nil
}
