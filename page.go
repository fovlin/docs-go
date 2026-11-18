package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"acovia.net/docs-go/internal/markdown"
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
	OutPath     string
	Heading     string
	Description string
	ModTime     time.Time
}

// RenderState is the transient render state for one Page.
type RenderState struct {
	*Page

	Body        string
	Outline     []markdown.Heading
	Prev        *PageLink
	Next        *PageLink
	EditURL     string
	LastUpdated string

	Nav     []*NavLink
	Sidebar []*SidebarNode

	Config Config
	UI     UIConfig

	Title      string
	Canonical  string
	Lang       string
	Appearance string
	HomeURL    string
	StyleURL   string
	InitURL    string
	AppURL     string
	FaviconURL string
	LogoURL    string
	PrevText   string
	NextText   string
	Bootstrap  string
	Version    string
}

// urlForFile maps a source file under SrcDir to its site URL path. The result
// is unescaped; escapePath is applied only when writing HTML.
func (app *App) urlForFile(filePath string) string {
	relative, err := filepath.Rel(app.SrcDir, filePath)
	if err != nil {
		return "/"
	}

	relative = filepath.ToSlash(relative)
	relative = strings.TrimSuffix(relative, path.Ext(relative))
	relative = strings.TrimPrefix(path.Clean("/"+relative), "/")

	switch {
	case relative == "" || relative == "index":
		return "/"
	case strings.HasSuffix(relative, "/index"):
		return "/" + strings.TrimSuffix(relative, "index")
	case app.CleanURLs:
		return "/" + relative + "/"
	default:
		return "/" + relative + ".html"
	}
}

func (app *App) loadPage(filePath string) (*Page, error) {
	absolute, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absolute)
	if err != nil {
		return nil, err
	}

	relative, err := filepath.Rel(app.SrcDir, absolute)
	if err != nil {
		return nil, err
	}
	relative = filepath.ToSlash(relative)

	url := app.urlForFile(absolute)
	outPath := strings.TrimPrefix(url, "/")
	switch {
	case outPath == "" || strings.HasSuffix(url, "/"):
		outPath += "index.html"
	case !strings.HasSuffix(outPath, ".html"):
		outPath += ".html"
	}

	return &Page{
		SourcePath: relative,
		AbsPath:    absolute,
		URL:        url,
		OutPath:    outPath,
		ModTime:    info.ModTime(),
	}, nil
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
