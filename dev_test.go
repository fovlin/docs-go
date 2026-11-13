package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func devGet(t *testing.T, dev *Dev, target string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	dev.handle(recorder, httptest.NewRequest(http.MethodGet, target, nil))

	return recorder
}

func TestDevReadsPageFromDiskEachTime(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)
	dev := newDev(app, "docs.json", "", "")

	first := devGet(t, dev, "/guide/start")
	if first.Code != http.StatusOK {
		t.Fatalf("initial status = %d", first.Code)
	}
	if !strings.Contains(first.Body.String(), "<h1 id=\"start\">Start") {
		t.Fatal("initial page did not render")
	}

	writeFile(t, filepath.Join(root, "docs", "guide", "start.md"), "# Changed\n\nNew body.\n")

	second := devGet(t, dev, "/guide/start")
	if second.Code != http.StatusOK {
		t.Fatalf("second status = %d", second.Code)
	}
	if !strings.Contains(second.Body.String(), "<h1 id=\"changed\">Changed") {
		t.Fatal("dev did not read the edited source")
	}

	if dev.snapshot() != app {
		t.Fatal("editing a page must not replace the project snapshot")
	}
}

func TestDevSnapshotSwap(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)
	dev := newDev(app, "docs.json", "", "")

	if dev.snapshot() != app {
		t.Fatal("snapshot changed without a config edit")
	}

	writeFile(t, filepath.Join(root, "docs.json"), `{
		"title": "Changed docs",
		"clean_urls": true
	}`)

	reloaded := dev.snapshot()
	if reloaded == app {
		t.Fatal("snapshot was not reloaded after a config edit")
	}
	if reloaded.Title != "Changed docs" {
		t.Fatalf("reloaded title = %q", reloaded.Title)
	}

	response := devGet(t, dev, "/guide/start")
	if !strings.Contains(response.Body.String(), "<title>Start | Changed docs</title>") {
		t.Fatal("request did not use the swapped snapshot")
	}
}

func TestDevFailedReloadKeepsPreviousSnapshot(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)
	dev := newDev(app, "docs.json", "", "")

	writeFile(t, filepath.Join(root, "docs.json"), `{"title": "Broken docs"`)

	if kept := dev.snapshot(); kept != app {
		t.Fatal("failed reload replaced the previous snapshot")
	}

	response := devGet(t, dev, "/guide/start")
	if !strings.Contains(response.Body.String(), "<title>Start | Test docs</title>") {
		t.Fatal("request did not use the previous snapshot after a failed reload")
	}

	writeFile(t, filepath.Join(root, "docs.json"), `{
		"title": "Fixed docs",
		"clean_urls": true
	}`)

	fixed := dev.snapshot()
	if fixed == app {
		t.Fatal("snapshot did not reload after the config was fixed")
	}
	if fixed.Title != "Fixed docs" {
		t.Fatalf("reloaded title = %q", fixed.Title)
	}
}

func TestDevCacheResetSeesNewSource(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app, err := LoadApp(root, "docs.json", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if page := app.findPageByURL("/guide/later"); page != nil {
		t.Fatal("unexpected page before creation")
	}

	writeFile(t, filepath.Join(root, "docs", "guide", "later.md"), "# Later\n")
	app.clearLookups()

	if page := app.findPageByURL("/guide/later"); page == nil {
		t.Fatal("new source stayed cached as missing")
	}
}

func TestSidebarTitleFollowsEditedSource(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	writeFile(t, filepath.Join(root, "sidebar.json"), `{
		"/guide/": [{ "items": [{ "link": "/guide/start" }] }]
	}`)
	app, err := LoadApp(root, "docs.json", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if text := app.sidebarFor("/guide/start")[0].Items[0].Text; text != "Start" {
		t.Fatalf("initial sidebar text = %q", text)
	}

	writeFile(t, filepath.Join(root, "docs", "guide", "start.md"), "# Changed\n\nBody.\n")
	if text := app.sidebarFor("/guide/start")[0].Items[0].Text; text != "Changed" {
		t.Fatalf("sidebar text did not follow source: %q", text)
	}
}
