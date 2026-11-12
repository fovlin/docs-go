package main

import (
	"os"
	"strings"
	"sync"
)

type titleEntry struct {
	title string
	mod   int64
	size  int64
}

// TitleCache maps absolute source paths to their H1 title. Entries are
// validated with modtime+size, so an edited file is re-parsed on next lookup.
type TitleCache struct {
	mu      sync.RWMutex
	entries map[string]titleEntry
}

func newTitleCache() *TitleCache {
	return &TitleCache{entries: make(map[string]titleEntry)}
}

func (cache *TitleCache) lookup(absPath string, mod int64, size int64) (string, bool) {
	if cache == nil {
		return "", false
	}

	cache.mu.RLock()
	entry, ok := cache.entries[absPath]
	cache.mu.RUnlock()

	if !ok || entry.mod != mod || entry.size != size {
		return "", false
	}

	return entry.title, true
}

func (cache *TitleCache) store(absPath string, title string, mod int64, size int64) {
	if cache == nil {
		return
	}

	cache.mu.Lock()
	cache.entries[absPath] = titleEntry{title: title, mod: mod, size: size}
	cache.mu.Unlock()
}

// pageTitle resolves the display title for a configured site path. It is used
// at load time to pre-resolve navbar and sidebar labels.
func (app *App) pageTitle(sitePath string) string {
	if page := app.findPageByURL(sitePath); page != nil {
		return app.titleFor(page)
	}

	if page := app.findPageBySource(strings.TrimPrefix(sitePath, "/")); page != nil {
		return app.titleFor(page)
	}

	return humanizeTitle(strings.TrimSuffix(sitePath, "/"))
}

func (app *App) titleFor(page *Page) string {
	info, statErr := os.Stat(page.AbsPath)
	var mod, size int64
	if statErr == nil {
		mod = info.ModTime().UnixNano()
		size = info.Size()
	}

	if title, ok := app.Titles.lookup(page.AbsPath, mod, size); ok {
		return title
	}

	title := ""
	if data, err := os.ReadFile(page.AbsPath); err == nil {
		if parsed, err := parseMarkdown(data, app.Config); err == nil {
			title = parsed.FirstH1
		}
	}
	if title == "" {
		title = humanizeTitle(page.SourcePath)
	}

	app.Titles.store(page.AbsPath, title, mod, size)
	return title
}

func (app *App) titleForHref(href string) string {
	target := href
	if app.Base != "/" {
		target = strings.TrimPrefix(href, strings.TrimSuffix(app.Base, "/"))
	}
	if page := app.findPageByURL(target); page != nil {
		return app.titleFor(page)
	}
	return href
}
