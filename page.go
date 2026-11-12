package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type PageLink struct {
	Text string
	Href string
}

// Page is the source-backed metadata for one Markdown file. Heading is the
// source H1; the display title lives on RenderState.Title.
type Page struct {
	SourcePath  string
	AbsPath     string
	URL         string
	Href        string
	OutPath     string
	Heading     string
	Description string
	ModTime     time.Time
}

// RenderState is the transient render state for one Page.
type RenderState struct {
	*Page

	Body        string
	Outline     []OutlineEntry
	Prev        *PageLink
	Next        *PageLink
	EditURL     string
	LastUpdated string

	Nav     []*NavLink
	Sidebar []*SidebarNode

	Config Config
	UI     UIConfig

	Title           string
	Canonical       string
	Lang            string
	Appearance      string
	HomeURL         string
	StyleURL        string
	InitURL         string
	AppURL          string
	FaviconURL      string
	LogoURL         string
	PrevText        string
	NextText        string
	FooterMessage   string
	FooterCopyright string
	Bootstrap       string
	Version         string
}

func pageURL(sourcePath string, cleanURLs bool) string {
	relative := strings.TrimSuffix(sourcePath, path.Ext(sourcePath))
	relative = strings.TrimPrefix(path.Clean("/"+relative), "/")

	switch {
	case relative == "" || relative == "index":
		return "/"
	case strings.HasSuffix(relative, "/index"):
		return "/" + strings.TrimSuffix(relative, "index")
	case cleanURLs:
		return "/" + relative
	default:
		return "/" + relative + ".html"
	}
}

func pageOutPath(sourcePath string) string {
	relative := strings.TrimSuffix(sourcePath, path.Ext(sourcePath))
	relative = strings.TrimPrefix(path.Clean("/"+relative), "/")

	if relative == "" || relative == "index" {
		return "index.html"
	}

	return relative + ".html"
}

func humanizeTitle(sourcePath string) string {
	name := strings.TrimSuffix(path.Base(sourcePath), path.Ext(sourcePath))
	if name == "index" {
		name = path.Base(path.Dir(sourcePath))
	}

	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.TrimSpace(name)
	if name == "" || name == "." {
		return "Home"
	}

	runes := []rune(name)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]

	return string(runes)
}

func (app *App) loadPage(relative string) (*Page, error) {
	absolute := filepath.Join(app.SrcDir, filepath.FromSlash(relative))
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, err
	}

	page := &Page{
		SourcePath: relative,
		AbsPath:    absolute,
		URL:        pageURL(relative, app.CleanURLs),
		OutPath:    pageOutPath(relative),
		ModTime:    info.ModTime(),
	}
	page.Href = app.href(page.URL)

	return page, nil
}
