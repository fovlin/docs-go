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
	dev := newDev(app, "config/docs.json", "", "")

	first := devGet(t, dev, "/guide/start/")
	if first.Code != http.StatusOK {
		t.Fatalf("initial status = %d", first.Code)
	}
	if !strings.Contains(first.Body.String(), "<h1 id=\"start\">Start") {
		t.Fatal("initial page did not render")
	}

	writeFile(t, filepath.Join(root, "docs", "guide", "start.md"), "# Changed\n\nNew body.\n")

	second := devGet(t, dev, "/guide/start/")
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
	dev := newDev(app, "config/docs.json", "", "")

	if dev.snapshot() != app {
		t.Fatal("snapshot changed without a config edit")
	}

	writeFile(t, filepath.Join(root, "config/docs.json"), `{
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

	response := devGet(t, dev, "/guide/start/")
	if !strings.Contains(response.Body.String(), "<title>Start | Changed docs</title>") {
		t.Fatal("request did not use the swapped snapshot")
	}
}

func TestDevFailedReloadKeepsPreviousSnapshot(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app := loadTestProject(t, root)
	dev := newDev(app, "config/docs.json", "", "")

	writeFile(t, filepath.Join(root, "config/docs.json"), `{"title": "Broken docs"`)

	if kept := dev.snapshot(); kept != app {
		t.Fatal("failed reload replaced the previous snapshot")
	}

	response := devGet(t, dev, "/guide/start/")
	if !strings.Contains(response.Body.String(), "<title>Start | Test docs</title>") {
		t.Fatal("request did not use the previous snapshot after a failed reload")
	}

	writeFile(t, filepath.Join(root, "config/docs.json"), `{
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

func TestDevSeesNewSourceWithoutCache(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	app, err := LoadApp(root, "config/docs.json", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := app.fileForURL("/guide/later/"); ok {
		t.Fatal("unexpected page before creation")
	}

	writeFile(t, filepath.Join(root, "docs", "guide", "later.md"), "# Later\n")

	file, ok := app.fileForURL("/guide/later/")
	if !ok {
		t.Fatal("new source not found")
	}
	if page, err := app.loadPage(file); err != nil || page.SourcePath != "guide/later.md" {
		t.Fatalf("new source = %+v, %v", page, err)
	}
}

func plainProject(t *testing.T) *App {
	t.Helper()

	root := t.TempDir()
	writeTestProject(t, root)
	writeFile(t, filepath.Join(root, "config", "docs.json"), `{"title": "Test docs", "clean_urls": false}`)

	return loadTestProject(t, root)
}

func TestFileForURLFollowsBuildRoutes(t *testing.T) {
	cases := []struct {
		name    string
		app     *App
		found   []string
		missing []string
	}{
		{
			name:    "clean",
			app:     loadTestProject(t, tempProject(t)),
			found:   []string{"/", "/guide", "/guide/", "/guide/index.html", "/guide/start", "/guide/start/", "/guide/start/index.html"},
			missing: []string{"/guide/start.html", "/guide/start.md"},
		},
		{
			name:    "plain",
			app:     plainProject(t),
			found:   []string{"/", "/guide", "/guide/", "/guide/index.html", "/guide/start.html"},
			missing: []string{"/guide/start", "/guide/start/", "/guide/start.md"},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			for _, target := range test.found {
				file, ok := test.app.fileForURL(target)
				if !ok {
					t.Errorf("%s not found", target)
					continue
				}
				if !isWithin(test.app.SrcDir, file) {
					t.Errorf("%s resolved to %s", target, file)
				}
			}

			for _, target := range test.missing {
				if file, ok := test.app.fileForURL(target); ok {
					t.Errorf("%s resolved to %s", target, file)
				}
			}
		})
	}
}

func TestFileForURLBindsIndexHTMLToSource(t *testing.T) {
	root := tempProject(t)
	writeFile(t, filepath.Join(root, "docs", "guide", "deep", "index.md"), "# Deep\n")
	app := loadTestProject(t, root)

	cases := map[string]string{
		"/guide/index.html":       "guide/index.md",
		"/guide/start/index.html": "guide/start.md",
		"/guide/deep/index.html":  "guide/deep/index.md",
	}

	for target, source := range cases {
		file, ok := app.fileForURL(target)
		if !ok {
			t.Errorf("%s not found", target)
			continue
		}

		relative, err := filepath.Rel(app.SrcDir, file)
		if err != nil {
			t.Fatal(err)
		}
		if got := filepath.ToSlash(relative); got != source {
			t.Errorf("%s -> %s, want %s", target, got, source)
		}
	}
}

func TestDevRoutesMatchBuild(t *testing.T) {
	clean := newDev(loadTestProject(t, tempProject(t)), "config/docs.json", "", "")
	for target, want := range map[string]int{
		"/":                       http.StatusOK,
		"/guide":                  http.StatusOK,
		"/guide/":                 http.StatusOK,
		"/guide/index.html":       http.StatusOK,
		"/guide/start":            http.StatusOK,
		"/guide/start/":           http.StatusOK,
		"/guide/start/index.html": http.StatusOK,
		"/guide/start.html":       http.StatusNotFound,
		"/guide/start.md":         http.StatusNotFound,
	} {
		if got := devGet(t, clean, target).Code; got != want {
			t.Errorf("clean %s = %d, want %d", target, got, want)
		}
	}

	plain := newDev(plainProject(t), "config/docs.json", "", "")
	for target, want := range map[string]int{
		"/":                 http.StatusOK,
		"/guide":            http.StatusOK,
		"/guide/":           http.StatusOK,
		"/guide/start.html": http.StatusOK,
		"/guide/index.html": http.StatusOK,
		"/guide/start/":     http.StatusNotFound,
		"/guide/start":      http.StatusNotFound,
		"/guide/start.md":   http.StatusNotFound,
	} {
		if got := devGet(t, plain, target).Code; got != want {
			t.Errorf("plain %s = %d, want %d", target, got, want)
		}
	}
}

func tempProject(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	writeTestProject(t, root)

	return root
}
