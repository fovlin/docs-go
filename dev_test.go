package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDevReadsPageFromDiskEachTime(t *testing.T) {
	root := t.TempDir()
	writeTestProject(t, root)
	loadTestProject(t, root)

	first, err := renderPageHTML(findPageByURL("/guide/start"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), "<h1 id=\"start\">Start") {
		t.Fatal("initial page did not render")
	}

	writeFile(t, filepath.Join(root, "docs", "guide", "start.md"), "# Changed\n\nNew body.\n")

	second, err := renderPageHTML(findPageByURL("/guide/start"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(second), "<h1 id=\"changed\">Changed") {
		t.Fatal("dev did not read the edited source")
	}
}
