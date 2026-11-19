package main

import "testing"

func TestFmtURL(t *testing.T) {
	cases := []struct {
		source string
		clean  string
		flat   string
	}{
		{"", "", ""},
		{"/", "/", "/"},
		{"/guide", "/guide", "/guide.html"},
		{"/guide/", "/guide", "/guide"},
		{"/guide/start", "/guide/start", "/guide/start.html"},
		{"guide/start", "/guide/start", "/guide/start.html"},
		{"https://example.com/a", "https://example.com/a", "https://example.com/a"},
		{"mailto:a@b", "mailto:a@b", "mailto:a@b"},
		{"#a", "#a", "#a"},
		{"?x=1", "?x=1", "?x=1"},
		{"//example.com/a", "//example.com/a", "//example.com/a"},
	}

	clean := &App{Config: Config{CleanURLs: true}}
	flat := &App{Config: Config{CleanURLs: false}}

	for _, test := range cases {
		if got := clean.fmtURL(test.source); got != test.clean {
			t.Errorf("clean fmtURL(%q) = %q, want %q", test.source, got, test.clean)
		}
		if got := flat.fmtURL(test.source); got != test.flat {
			t.Errorf("flat fmtURL(%q) = %q, want %q", test.source, got, test.flat)
		}
	}
}

func TestFmtPagePathInfo(t *testing.T) {
	cases := []struct {
		source   string
		cleanURL string
		cleanOut string
		flatURL  string
		flatOut  string
	}{
		{"index", "/", "index.html", "/", "index.html"},
		{"guide/index", "/guide", "guide/index.html", "/guide", "guide/index.html"},
		{"guide/start", "/guide/start", "guide/start/index.html", "/guide/start.html", "guide/start.html"},
	}

	clean := &App{Config: Config{CleanURLs: true}}
	flat := &App{Config: Config{CleanURLs: false}}

	for _, test := range cases {
		if url, out := clean.fmtPagePathInfo(test.source); url != test.cleanURL || out != test.cleanOut {
			t.Errorf("clean fmtPagePathInfo(%q) = %q, %q; want %q, %q", test.source, url, out, test.cleanURL, test.cleanOut)
		}
		if url, out := flat.fmtPagePathInfo(test.source); url != test.flatURL || out != test.flatOut {
			t.Errorf("flat fmtPagePathInfo(%q) = %q, %q; want %q, %q", test.source, url, out, test.flatURL, test.flatOut)
		}
	}
}

func TestResolveDestination(t *testing.T) {
	cases := []struct {
		raw   string
		page  string
		image bool
		clean string
		flat  string
	}{
		{"https://example.com/a", "guide/index.md", false, "https://example.com/a", "https://example.com/a"},
		{"mailto:a@b", "guide/index.md", false, "mailto:a@b", "mailto:a@b"},
		{"#a", "guide/index.md", false, "#a", "#a"},
		{"?x=1", "guide/index.md", false, "?x=1", "?x=1"},
		{"other.md", "guide/index.md", false, "/guide/other", "/guide/other.html"},
		{"other.html", "guide/index.md", false, "/guide/other", "/guide/other.html"},
		{"../guide/other.md", "guide/index.md", false, "/guide/other", "/guide/other.html"},
		{"/guide/other.md", "guide/index.md", false, "/guide/other", "/guide/other.html"},
		{"other.md#x", "guide/index.md", false, "/guide/other#x", "/guide/other.html#x"},
		{"other.md?x=1", "guide/index.md", false, "/guide/other?x=1", "/guide/other.html?x=1"},
		{"files/x.pdf", "guide/index.md", false, "files/x.pdf", "files/x.pdf"},
		{"/files/x.pdf", "guide/index.md", false, "/files/x.pdf", "/files/x.pdf"},
		{"img.png", "guide/index.md", true, "img.png", "img.png"},
	}

	clean := &App{Config: Config{CleanURLs: true}}
	flat := &App{Config: Config{CleanURLs: false}}

	for _, test := range cases {
		page := &Page{SourcePath: test.page}
		if got := clean.resolveDestination(page, test.raw, test.image); got != test.clean {
			t.Errorf("clean resolveDestination(%q) = %q, want %q", test.raw, got, test.clean)
		}
		if got := flat.resolveDestination(page, test.raw, test.image); got != test.flat {
			t.Errorf("flat resolveDestination(%q) = %q, want %q", test.raw, got, test.flat)
		}
	}
}

func TestNavActiveMatch(t *testing.T) {
	app := &App{Nav: []*NavLink{
		{Text: "Guide", Href: "/guide", ActiveMatch: "/guide"},
		{Text: "Other", Href: "/other"},
	}}

	links := app.navFor("/guide/start")
	if !links[0].Active || links[1].Active {
		t.Fatalf("guide page active state = %v, %v", links[0].Active, links[1].Active)
	}

	links = app.navFor("/other")
	if links[0].Active || !links[1].Active {
		t.Fatalf("other page active state = %v, %v", links[0].Active, links[1].Active)
	}

	if links = app.navFor("/other/x"); links[1].Active {
		t.Fatal("link without active_match matched a prefix")
	}
}
