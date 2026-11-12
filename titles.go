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

var (
	titlesMu sync.RWMutex
	titles   = map[string]titleEntry{}
)

func pageTitle(sitePath string) string {
	if page := findPageByURL(sitePath); page != nil {
		return titleFor(page)
	}

	if page := findPageBySource(strings.TrimPrefix(sitePath, "/")); page != nil {
		return titleFor(page)
	}

	return humanizeTitle(strings.TrimSuffix(sitePath, "/"))
}

func titleFor(page *Page) string {
	info, statErr := os.Stat(page.AbsPath)
	var mod, size int64
	if statErr == nil {
		mod = info.ModTime().UnixNano()
		size = info.Size()
	}

	titlesMu.RLock()
	entry, ok := titles[page.AbsPath]
	titlesMu.RUnlock()
	if ok && entry.mod == mod && entry.size == size {
		return entry.title
	}

	title := ""
	if data, err := os.ReadFile(page.AbsPath); err == nil {
		if parsed, err := parseMarkdown(data, config); err == nil {
			title = parsed.FirstH1
		}
	}
	if title == "" {
		title = humanizeTitle(page.SourcePath)
	}

	titlesMu.Lock()
	titles[page.AbsPath] = titleEntry{title: title, mod: mod, size: size}
	titlesMu.Unlock()
	return title
}

func titleForHref(href string) string {
	target := href
	if config.Base != "/" {
		target = strings.TrimPrefix(href, strings.TrimSuffix(config.Base, "/"))
	}
	if page := findPageByURL(target); page != nil {
		return titleFor(page)
	}
	return href
}
