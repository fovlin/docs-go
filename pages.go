package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

func findPageByURL(target string) *Page {
	target = canonicalURL(target)
	trailing := strings.HasSuffix(target, "/")
	relative := strings.TrimSuffix(strings.TrimPrefix(target, "/"), ".html")
	relative = strings.TrimSuffix(relative, "/")

	if relative == "" || relative == "index" {
		return findSource([]string{"index.md", "index.markdown"})
	}

	candidates := make([]string, 0, 4)
	if trailing {
		candidates = append(candidates, relative+"/index.md", relative+"/index.markdown")
	}

	candidates = append(candidates, relative+".md", relative+".markdown")

	if !trailing {
		candidates = append(candidates, relative+"/index.md", relative+"/index.markdown")
	}

	return findSource(candidates)
}

func findPageBySource(relative string) *Page {
	relative = strings.TrimPrefix(path.Clean("/"+relative), "/")

	if isMarkdownPath(relative) {
		if page, err := loadPage(relative); err == nil {
			return page
		}
	}

	if relative == "" || strings.HasSuffix(relative, "/") {
		relative = strings.TrimSuffix(relative, "/") + "/index"
	}

	return findSource([]string{
		relative + ".md",
		relative + ".markdown",
		relative + "/index.md",
		relative + "/index.markdown",
	})
}

func findSource(candidates []string) *Page {
	for _, relative := range candidates {
		absolute := filepath.Join(srcDir, filepath.FromSlash(relative))
		if info, err := os.Stat(absolute); err == nil && !info.IsDir() {
			if page, err := loadPage(relative); err == nil {
				return page
			}
		}
	}

	return nil
}
