package main

import (
	"os"
	"path"
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

func (app *App) loadPage(filePath string) (*Page, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	url := app.fmtPageURL(filePath)
	outPath := strings.TrimPrefix(url, "/")

	return &Page{
		SourcePath: filePath,
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
