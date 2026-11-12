package markdown

import (
	"strings"
	"testing"
)

func render(source string) string {
	document := Parse([]byte(source), Options{Containers: defaultContainerLabels})

	return string(document.Render(nil))
}

func TestHeadingsAndIDs(t *testing.T) {
	document := Parse([]byte("# 中文文档\n\n## 快速开始\n\n## 快速开始\n\n### Custom {#mine}\n"), Options{})

	if len(document.Headings) != 4 {
		t.Fatalf("expected 4 headings, got %d", len(document.Headings))
	}

	expected := []string{"中文文档", "快速开始", "快速开始-1", "mine"}
	for index, id := range expected {
		if document.Headings[index].ID != id {
			t.Errorf("heading %d id = %q, want %q", index, document.Headings[index].ID, id)
		}
	}

	if document.FirstH1 != "中文文档" {
		t.Errorf("first H1 = %q", document.FirstH1)
	}
}

func TestInlineFormatting(t *testing.T) {
	html := render("A **bold** and *em* and ~~gone~~ and `code`.\n")

	for _, expected := range []string{
		"<strong>bold</strong>",
		"<em>em</em>",
		"<del>gone</del>",
		"<code>code</code>",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("missing %q in %s", expected, html)
		}
	}
}

func TestNestedEmphasis(t *testing.T) {
	html := render("**bold *em* bold** and *em **strong** em*.\n")

	if !strings.Contains(html, "<strong>bold <em>em</em> bold</strong>") {
		t.Errorf("nested strong/em failed: %s", html)
	}

	if !strings.Contains(html, "<em>em <strong>strong</strong> em</em>") {
		t.Errorf("nested em/strong failed: %s", html)
	}
}

func TestLinksImagesAutolinks(t *testing.T) {
	html := render("[next](./next.md) ![pic](./pic.png) <https://example.com> https://example.org.\n")

	for _, expected := range []string{
		`<a href="./next.md">next</a>`,
		`<img src="./pic.png" alt="pic">`,
		`<a href="https://example.com">https://example.com</a>`,
		`<a href="https://example.org">https://example.org</a>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("missing %q in %s", expected, html)
		}
	}
}

func TestUnsafeLinkIsNeutralised(t *testing.T) {
	html := render("[bad](javascript:alert(1))\n")

	if strings.Contains(html, "javascript:") {
		t.Errorf("javascript link survived: %s", html)
	}

	if !strings.Contains(html, `href="#"`) {
		t.Errorf("unsafe link was not neutralised: %s", html)
	}
}

func TestCodeFenceKeepsContainerMarkers(t *testing.T) {
	source := strings.Join([]string{
		"::: warning",
		"Before",
		"",
		"~~~text",
		":::",
		"~~~",
		"",
		":::",
	}, "\n")

	html := render(source)

	if !strings.Contains(html, `<div class="custom-block warning">`) {
		t.Errorf("missing container: %s", html)
	}

	if !strings.Contains(html, "<code class=\"language-text\">:::\n</code>") {
		t.Errorf("container marker inside fence was parsed: %s", html)
	}
}

func TestNestedContainers(t *testing.T) {
	html := render("::: tip\nouter\n\n::: warning\ninner\n:::\n:::\n")

	if strings.Count(html, `<div class="custom-block tip">`) != 1 {
		t.Errorf("outer container missing: %s", html)
	}

	if strings.Count(html, `<div class="custom-block warning">`) != 1 {
		t.Errorf("inner container missing: %s", html)
	}
}

func TestListsAndTasks(t *testing.T) {
	source := "- [x] done\n- [ ] todo\n\n1. first\n2. second\n"
	html := render(source)

	for _, expected := range []string{
		`<ul class="task-list">`,
		`<li class="task-list-item"><input type="checkbox" disabled checked> done</li>`,
		`<li class="task-list-item"><input type="checkbox" disabled> todo</li>`,
		`<ol>`,
		`<li>first</li>`,
		`<li>second</li>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("missing %q in %s", expected, html)
		}
	}
}

func TestListEndsBeforeOutdentedParagraph(t *testing.T) {
	html := render("- one\n- two\n\nAfter\n")

	if strings.Contains(html, "<li>two<p>After") {
		t.Errorf("paragraph was swallowed by the list: %s", html)
	}

	if !strings.Contains(html, "</ul>\n<p>After</p>") {
		t.Errorf("list did not end before the paragraph: %s", html)
	}
}

func TestLooseList(t *testing.T) {
	html := render("- one\n\n- two\n")

	if !strings.Contains(html, "<li><p>one</p>") {
		t.Errorf("blank line did not make the list loose: %s", html)
	}
}

func TestNestedList(t *testing.T) {
	html := render("- parent\n  - child\n")

	if !strings.Contains(html, "<li>parent\n<ul>") {
		t.Errorf("nested list was not parsed: %s", html)
	}
}

func TestTable(t *testing.T) {
	source := "| Left | Right |\n| :--- | ---: |\n| a | b |\n"
	html := render(source)

	for _, expected := range []string{
		"<table>",
		`<th style="text-align:left">Left</th>`,
		`<th style="text-align:right">Right</th>`,
		`<td style="text-align:left">a</td>`,
		`<td style="text-align:right">b</td>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("missing %q in %s", expected, html)
		}
	}
}

func TestBlockquote(t *testing.T) {
	html := render("> quoted\n>\n> > nested\n")

	if !strings.Contains(html, "<blockquote>\n<p>quoted</p>") {
		t.Errorf("blockquote failed: %s", html)
	}

	if !strings.Contains(html, "<blockquote>\n<p>nested</p>") {
		t.Errorf("nested blockquote failed: %s", html)
	}
}

func TestRawHTMLIsEscaped(t *testing.T) {
	html := render("<figure><img src=\"/x.png\" alt=\"x\"></figure>\n")

	if strings.Contains(html, "<figure>") {
		t.Errorf("raw HTML was not escaped: %s", html)
	}

	if !strings.Contains(html, "&lt;figure&gt;") {
		t.Errorf("escaped HTML missing: %s", html)
	}
}

func TestEntity(t *testing.T) {
	html := render("A &amp; B\n")

	if !strings.Contains(html, "A &amp; B") {
		t.Errorf("entity was not preserved: %s", html)
	}
}

func TestHardBreak(t *testing.T) {
	html := render("one  \ntwo\n")

	if !strings.Contains(html, "one<br>\ntwo") {
		t.Errorf("hard break failed: %s", html)
	}
}

func TestRewriteLinks(t *testing.T) {
	document := Parse([]byte("[next](./next.md) ![pic](./pic.png)\n"), Options{})

	document.RewriteLinks(func(destination string, image bool) (string, bool) {
		if image {
			return "/images/pic.png", true
		}

		return "/guide/next", true
	})

	html := string(document.Render(nil))

	if !strings.Contains(html, `href="/guide/next"`) {
		t.Errorf("link rewrite failed: %s", html)
	}

	if !strings.Contains(html, `src="/images/pic.png"`) {
		t.Errorf("image rewrite failed: %s", html)
	}
}

func TestPlainTextAndFirstParagraph(t *testing.T) {
	document := Parse([]byte("# Title\n\nFirst **paragraph** with `code`.\n\n## Section\n"), Options{})

	if document.FirstParagraph != "First paragraph with code." {
		t.Errorf("first paragraph = %q", document.FirstParagraph)
	}

	if !strings.Contains(document.PlainText, "Title") || !strings.Contains(document.PlainText, "Section") {
		t.Errorf("plain text = %q", document.PlainText)
	}
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"# heading\n\nparagraph\n",
		"- [x] task\n- item\n",
		"::: tip\ncontent\n:::\n",
		"| a | b |\n| --- | --- |\n| 1 | 2 |\n",
		"**bold** and [link](./x.md)\n",
		"```go\nfmt.Println(\"x\")\n```\n",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, source string) {
		document := Parse([]byte(source), Options{Containers: defaultContainerLabels})
		_ = document.Render(nil)
	})
}
