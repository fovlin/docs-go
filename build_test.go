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

func loadTestProject(t *testing.T, root string) {
	t.Helper()

	configFile = "docs.json"
	srcFlag = ""
	outFlag = ""
	logQuiet = true
	if err := loadProject(root); err != nil {
		t.Fatal(err)
	}
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
	loadTestProject(t, root)

	if err := runBuild(); err != nil {
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

	configFile = "docs.json"
	srcFlag = ""
	outFlag = ""
	logQuiet = true
	err := loadProject(root)
	if err == nil || !strings.Contains(err.Error(), "outside its route") {
		t.Fatalf("expected route mismatch, got %v", err)
	}
}

func TestURLTable(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	loadTestProject(t, root)

	info, ok := lookupURLInfo("/guide/start")
	if !ok {
		t.Fatal("page missing from urlTab")
	}
	if info.PrevHref != "/guide/" || info.NextHref != "/guide/next" {
		t.Errorf("navigation = %+v", info)
	}
}

func readFile(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}
