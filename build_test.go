package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, name string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func loadTestProject(t *testing.T, root string) *App {
	t.Helper()

	logQuiet = true
	app, err := LoadApp(root, "docs.json", "", "")
	if err != nil {
		t.Fatal(err)
	}

	return app
}

func writeTestProject(t *testing.T, root string) {
	t.Helper()

	writeFile(t, filepath.Join(root, "docs.json"), `{
		"title": "Test docs",
		"clean_urls": true
	}`)
	writeFile(t, filepath.Join(root, "sidebar.json"), `{
		"/guide/": [{
			"text": "Guide",
			"items": [
				{ "text": "Overview", "link": "/guide/" },
				{ "text": "Start", "link": "/guide/start" },
				{ "text": "Next", "link": "/guide/next" }
			]
		}]
	}`)
	writeFile(t, filepath.Join(root, "docs", "index.md"), "# Home\n\nWelcome.\n")
	writeFile(t, filepath.Join(root, "docs", "guide", "index.md"), "# Guide\n\nOverview.\n")
	writeFile(t, filepath.Join(root, "docs", "guide", "start.md"), "# Start\n\nStart page.\n\n## Details\n\nBody.\n")
	writeFile(t, filepath.Join(root, "docs", "guide", "next.md"), "# Next\n\nNext page.\n")
}

func TestBuild(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)

	if err := runBuild(app); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		"index.html",
		"guide/index.html",
		"guide/start.html",
		"guide/next.html",
		"404.html",
		"favicon.ico",
		"assets/style.css",
	} {
		if _, err := os.Stat(filepath.Join(root, "dist", name)); err != nil {
			t.Errorf("missing %s: %v", name, err)
		}
	}

	page := readFile(t, filepath.Join(root, "dist", "guide", "start.html"))
	for _, want := range []string{
		`<title>Start | Test docs</title>`,
		`<h2 id="details">`,
		`prev-next-title">Guide`,
		`prev-next-title">Next`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in start page", want)
		}
	}
}

func TestBuildRejectsRouteMismatch(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	writeFile(t, filepath.Join(root, "sidebar.json"), `{
		"/guide/": [{ "text": "Guide", "link": "/outside" }]
	}`)
	writeFile(t, filepath.Join(root, "docs", "outside.md"), "# Outside\n")

	logQuiet = true
	_, err := LoadApp(root, "docs.json", "", "")
	if err == nil || !strings.Contains(err.Error(), "outside its route") {
		t.Fatalf("expected route mismatch, got %v", err)
	}
}

func TestBuildRejectsDuplicatePageAcrossRoutes(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	writeFile(t, filepath.Join(root, "sidebar.json"), `{
		"/": [{ "text": "Home route", "link": "/guide/start" }],
		"/guide/": [{ "text": "Guide", "link": "/guide/start" }]
	}`)

	logQuiet = true
	_, err := LoadApp(root, "docs.json", "", "")
	if err == nil || !strings.Contains(err.Error(), "more than one sidebar route") {
		t.Fatalf("expected duplicate route error, got %v", err)
	}
}

func TestAppPrevNext(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)

	info, ok := app.URLTab["/guide/start"]
	if !ok {
		t.Fatal("page missing from URLTab")
	}
	if info.PrevHref != "/guide/" || info.NextHref != "/guide/next" {
		t.Errorf("navigation = %+v", info)
	}

	page := mustPage(t, app, "/guide/start")

	state := &RenderState{Page: page}
	app.applyNavigation(state, page)
	if state.Prev == nil || state.Prev.Text != "Guide" || state.Prev.Href != "/guide/" {
		t.Errorf("prev = %+v", state.Prev)
	}
	if state.Next == nil || state.Next.Text != "Next" || state.Next.Href != "/guide/next" {
		t.Errorf("next = %+v", state.Next)
	}
}

func TestNavigationIsPreResolved(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)

	if len(app.Nav) != 0 {
		t.Fatalf("expected no navbar links, got %d", len(app.Nav))
	}

	tree := app.SidebarTrees["/guide"]
	if len(tree) != 1 || len(tree[0].Items) != 3 {
		t.Fatalf("unexpected resolved sidebar tree: %+v", tree)
	}
	if tree[0].Items[1].Href != "/guide/start" {
		t.Errorf("resolved href = %q", tree[0].Items[1].Href)
	}
	if tree[0].HasActive {
		t.Fatal("stored sidebar tree was mutated before rendering")
	}

	active := app.sidebarFor("/guide/start")
	if len(active) != 1 || !active[0].HasActive || !active[0].Open {
		t.Fatalf("active sidebar state = %+v", active)
	}
	if !active[0].Items[1].Active {
		t.Fatal("start link is not marked active")
	}
	if tree[0].HasActive || tree[0].Items[1].Active {
		t.Fatal("per-page marking mutated the shared sidebar tree")
	}
}

func TestPageHeadingAndDisplayTitle(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)

	page := mustPage(t, app, "/guide/start")
	if page.Heading != "" {
		t.Fatalf("heading should be empty before render, got %q", page.Heading)
	}

	document, err := app.renderPageHTML(page)
	if err != nil {
		t.Fatal(err)
	}
	if page.Heading != "Start" {
		t.Errorf("page heading = %q", page.Heading)
	}
	if !strings.Contains(string(document), "<title>Start | Test docs</title>") {
		t.Errorf("display title = %q, want %q", titleOf(t, string(document)), "Start | Test docs")
	}

	home := mustPage(t, app, "/")
	homeDocument, err := app.renderPageHTML(home)
	if err != nil {
		t.Fatal(err)
	}
	if home.Heading != "Home" {
		t.Errorf("home heading = %q", home.Heading)
	}
	if !strings.Contains(string(homeDocument), "<title>Test docs</title>") {
		t.Errorf("home display title = %q, want site title", titleOf(t, string(homeDocument)))
	}

	state := &RenderState{Page: &Page{Heading: "Start", URL: "/guide/start"}}
	app.fillPageChrome(state)
	if state.Title != "Start | Test docs" {
		t.Errorf("RenderState.Title = %q", state.Title)
	}
	if state.Page.Heading != "Start" {
		t.Errorf("Page.Heading changed to %q", state.Page.Heading)
	}
}

func TestTitleCacheInvalidatesOnChange(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)

	page := mustPage(t, app, "/guide/start")
	if title := app.titleFor(page); title != "Start" {
		t.Fatalf("initial title = %q", title)
	}

	writeFile(t, filepath.Join(root, "docs", "guide", "start.md"), "# Changed\n\nNew body.\n")

	changed := mustPage(t, app, "/guide/start")
	if title := app.titleFor(changed); title != "Changed" {
		t.Fatalf("cached title after edit = %q", title)
	}
}

func mustPage(t *testing.T, app *App, target string) *Page {
	t.Helper()

	page := app.findPageByURL(target)
	if page == nil {
		t.Fatalf("missing page %s", target)
	}

	return page
}

func titleOf(t *testing.T, document string) string {
	t.Helper()

	start := strings.Index(document, "<title>")
	if start < 0 {
		return ""
	}
	end := strings.Index(document[start:], "</title>")
	if end < 0 {
		return ""
	}

	return document[start+len("<title>") : start+end]
}

func readFile(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func TestBuildRejectsDuplicateOutput(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	writeFile(t, filepath.Join(root, "docs", "guide", "start.markdown"), "# Duplicate\n")
	app := loadTestProject(t, root)

	if err := runBuild(app); err == nil || !strings.Contains(err.Error(), "both map to") {
		t.Fatalf("expected duplicate output error, got %v", err)
	}
}

func TestShellClassDoesNotReserveEmptySidebar(t *testing.T) {
	state := &RenderState{Page: &Page{}, Nav: []*NavLink{{Text: "Guide"}}}
	if got := shellClass(state); got != "shell no-sidebar" {
		t.Fatalf("shell class = %q, want shell no-sidebar", got)
	}

	state.Sidebar = []*SidebarNode{{Text: "Guide"}}
	if got := shellClass(state); got != "shell" {
		t.Fatalf("shell class = %q, want shell", got)
	}
}
