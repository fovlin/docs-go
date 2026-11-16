package markdown

import (
	"strings"
	"testing"
)

func render(source string) string {
	document := Parse([]byte(source))

	return string(document.Render(nil))
}

func TestHeadingsAndIDs(t *testing.T) {
	document := Parse([]byte("# 中文文档\n\n## 快速开始\n\n## 快速开始\n"))

	if len(document.Headings) != 3 {
		t.Fatalf("expected 3 headings, got %d", len(document.Headings))
	}

	expected := []string{"中文文档", "快速开始", "快速开始-1"}
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
	html := render("A **bold** and *em* and `code`.\n")

	for _, expected := range []string{
		"<strong>bold</strong>",
		"<em>em</em>",
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

func TestLinksImages(t *testing.T) {
	html := render("[next](./next.md) ![pic](./pic.png)\n")

	for _, expected := range []string{
		`<a href="./next.md">next</a>`,
		`<img src="./pic.png" alt="pic">`,
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

func TestUnsafeImageIsNeutralised(t *testing.T) {
	html := render("![bad](javascript:alert(1))\n")

	if strings.Contains(html, "javascript:") {
		t.Errorf("javascript image survived: %s", html)
	}

	if !strings.Contains(html, `src=""`) {
		t.Errorf("unsafe image was not neutralised: %s", html)
	}
}

func TestLists(t *testing.T) {
	html := render("- first\n- second\n\n1. first\n2. second\n")

	for _, expected := range []string{
		"<ul>",
		"<li>first</li>",
		"<li>second</li>",
		"<ol>",
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

func TestBlankLineContinuesList(t *testing.T) {
	html := render("- one\n\n- two\n")

	if strings.Count(html, "<ul>") != 1 {
		t.Errorf("blank line should not end the list: %s", html)
	}

	if !strings.Contains(html, "<li>one</li>") || !strings.Contains(html, "<li>two</li>") {
		t.Errorf("list items missing: %s", html)
	}
}

func TestNestedList(t *testing.T) {
	html := render("- parent\n  - child\n")

	if !strings.Contains(html, "<li>parent\n<ul>") {
		t.Errorf("nested list was not parsed: %s", html)
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

func TestEntity(t *testing.T) {
	html := render("A &amp; B\n")

	if !strings.Contains(html, "A &amp; B") {
		t.Errorf("entity was not preserved: %s", html)
	}
}

func TestLineBreakSplitsParagraph(t *testing.T) {
	html := render("one  \ntwo\n")

	if !strings.Contains(html, "<p>one  </p>\n<p>two</p>") {
		t.Errorf("line break should split the paragraph: %s", html)
	}
}

func TestRewriteLinks(t *testing.T) {
	document := Parse([]byte("[next](./next.md) ![pic](./pic.png)\n"))

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
	document := Parse([]byte("# Title\n\nFirst **paragraph** with `code`.\n\n## Section\n"))

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
		"- item\n- item\n",
		"**bold** and [link](./x.md)\n",
		"```go\nfmt.Println(\"x\")\n```\n",
		"    indented code\n",
		"::: tip\ncontent\n:::\n",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, source string) {
		document := Parse([]byte(source))
		_ = document.Render(nil)
	})
}

func TestTextStartingWithAngleIsParagraph(t *testing.T) {
	html := render("<3 is love\n\nAfter\n")

	if !strings.Contains(html, "<p>&lt;3 is love</p>") {
		t.Errorf("angle text was treated as HTML: %s", html)
	}

	if !strings.Contains(html, "<p>After</p>") {
		t.Errorf("angle text swallowed following markdown: %s", html)
	}
}

func TestEmphasisDelimiters(t *testing.T) {
	cases := []struct{ in, want string }{
		{"__A__ 与 __B\n", "<strong>A</strong> 与 __B"},
		{"__init__ 和 __main__\n", "<strong>init</strong> 和 <strong>main</strong>"},
		{"___c___\n", "<strong><em>c</em></strong>"},
		{"a_b_c\n", "a_b_c"},
		{"a ** b\n", "a ** b"},
		{"*i* 和 __s__\n", "<em>i</em> 和 <strong>s</strong>"},
	}

	for _, tc := range cases {
		if got := render(tc.in); !strings.Contains(got, tc.want) {
			t.Errorf("%q = %q, want it to contain %q", tc.in, got, tc.want)
		}
	}
}

func TestInlineHTMLStaysInline(t *testing.T) {
	html := render("<kbd>Ctrl</kbd> + C\n\n# Heading\n")

	if !strings.Contains(html, "<p><kbd>Ctrl</kbd> + C</p>") {
		t.Errorf("inline HTML swallowed the paragraph: %s", html)
	}

	if !strings.Contains(html, `<h1 id="heading">`) {
		t.Errorf("following heading was swallowed: %s", html)
	}
}

func TestListIndentation(t *testing.T) {
	cases := []struct {
		source string
		lists  int
		want   string
	}{
		{"- a\n - b\n", 1, "<li>a</li>\n<li>b</li>"},
		{"- a\n  - b\n - c\n", 2, "</ul>\n</li>\n<li>c</li>"},
		{"- a\n\n text\n", 1, "</ul>\n<p> text</p>"},
	}

	for _, tc := range cases {
		html := render(tc.source)

		if got := strings.Count(html, "<ul>") + strings.Count(html, "<ol>"); got != tc.lists {
			t.Errorf("%q: got %d lists, want %d", tc.source, got, tc.lists)
		}

		if !strings.Contains(html, tc.want) {
			t.Errorf("%q: missing %q", tc.source, tc.want)
		}
	}
}

func TestIndentedCode(t *testing.T) {
	html := render("    code\n    more\n\nAfter\n")

	if !strings.Contains(html, "<code>code\nmore\n</code>") {
		t.Errorf("indented code block missing: %s", html)
	}

	if !strings.Contains(html, "<p>After</p>") {
		t.Errorf("paragraph after code block missing: %s", html)
	}

	if tabbed := render("\tcode\n"); !strings.Contains(tabbed, "<code>code\n</code>") {
		t.Errorf("tab-indented code block missing: %s", tabbed)
	}

	if paragraph := render("text\n    still text\n"); !strings.Contains(paragraph, "<code>still text") {
		t.Errorf("indented line should become a code block: %s", paragraph)
	}
}

func TestTable(t *testing.T) {
	html := render("| Left | Right |\n| :--- | ---: |\n| a | b |\n")

	for _, expected := range []string{
		"<table>",
		`<th style="text-align:left">Left</th>`,
		`<td style="text-align:right">b</td>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("missing %q in %s", expected, html)
		}
	}
}

func TestListMarkers(t *testing.T) {
	if html := render("- a\n* b\n"); strings.Count(html, "<ul>") != 2 {
		t.Errorf("different bullet markers should start a new list: %s", html)
	}

	if html := render("1. a\n1) b\n"); strings.Count(html, "<ol>") != 2 {
		t.Errorf("different ordered markers should start a new list: %s", html)
	}

	if html := render("text\n2. item\n"); !strings.Contains(html, `<ol start="2">`) {
		t.Errorf("ordered list did not start after the paragraph: %s", html)
	}
}

func TestBlockquoteBoundaries(t *testing.T) {
	html := render("> one\n\n> two\n")

	if strings.Count(html, "<blockquote>") != 2 {
		t.Errorf("blank line should start a new blockquote: %s", html)
	}

	multi := render("> one\n>\n> two\n")
	if !strings.Contains(multi, "<p>one</p>\n<p>two</p>") {
		t.Errorf("quoted blank line should keep one blockquote: %s", multi)
	}
}

func TestTableCells(t *testing.T) {
	html := render("| a | b |\n| --- | --- |\n| x | y \\| z |\n")

	for _, expected := range []string{"<td>x</td>", "<td>y | z</td>"} {
		if !strings.Contains(html, expected) {
			t.Errorf("missing %q in %s", expected, html)
		}
	}

	if bad := render("| a | b | c |\n| --- | --- |\n| 1 | 2 | 3 |\n"); strings.Contains(bad, "<table>") {
		t.Errorf("mismatched header should not be a table: %s", bad)
	}

	if heading := render("| a | b |\n| --- | --- |\n## Next | section\n"); !strings.Contains(heading, `<h2 id="next-section">`) {
		t.Errorf("table did not end before a heading: %s", heading)
	}
}

func TestInlineHTML(t *testing.T) {
	html := render("<span>**bold**</span>\n")

	if !strings.Contains(html, "<p><span><strong>bold</strong></span></p>") {
		t.Errorf("inline HTML should stay in a paragraph: %s", html)
	}
}

func TestDocumentedVariants(t *testing.T) {
	if html := render("~~~go\nfmt.Println(1)\n~~~\n"); !strings.Contains(html, `class="language-go"`) {
		t.Errorf("tilde fence missing: %s", html)
	}

	if html := render("___\n"); !strings.Contains(html, "<hr>") {
		t.Errorf("underscore thematic break missing: %s", html)
	}

	if html := render("[x](/x \"title\")\n"); !strings.Contains(html, `title="title"`) {
		t.Errorf("link title missing: %s", html)
	}
}

func TestFencedCode(t *testing.T) {
	cases := []string{
		"```go\nfmt.Println(1)\n```\n",
		"~~~go\nfmt.Println(1)\n~~~\n",
	}

	for _, source := range cases {
		if html := render(source); !strings.Contains(html, `class="language-go"`) {
			t.Errorf("fenced code missing for %q: %s", source, html)
		}
	}
}

func TestContainers(t *testing.T) {
	html := render("::: tip\nhello\n:::\n")
	for _, expected := range []string{`<div class="custom-block tip">`, `<p class="custom-block-title">TIP</p>`, "<p>hello</p>"} {
		if !strings.Contains(html, expected) {
			t.Errorf("missing %q in:\n%s", expected, html)
		}
	}

	custom := render("::: warning Watch out\nbody\n:::\n")
	if !strings.Contains(custom, `<p class="custom-block-title">Watch out</p>`) {
		t.Errorf("custom title missing: %s", custom)
	}

	nested := render("::: tip\n::: warning\ninner\n:::\n:::\n")
	if strings.Count(nested, `<div class="custom-block tip">`) != 1 || strings.Count(nested, `<div class="custom-block warning">`) != 1 {
		t.Errorf("nested containers missing: %s", nested)
	}

	code := render("::: tip\n```text\n:::\n```\n:::\n")
	if !strings.Contains(code, ":::\n</code>") {
		t.Errorf("container marker inside fence was parsed: %s", code)
	}

	details := render("::: details\nhidden\n:::\n")
	if !strings.Contains(details, `<details class="custom-block details">`) || !strings.Contains(details, "<summary") {
		t.Errorf("details container missing: %s", details)
	}

	unknown := render("::: custom\nbody\n:::\n")
	if strings.Contains(unknown, "custom-block") {
		t.Errorf("unknown container name should be text: %s", unknown)
	}
}

func TestOrderedStartZero(t *testing.T) {
	html := render("0. zero\n1. one\n")

	if !strings.Contains(html, `<ol start="0">`) {
		t.Errorf("zero start missing: %s", html)
	}
}

func TestFenceIndent(t *testing.T) {
	html := render("  ```\n  code\n    extra\n  ```\n")

	if !strings.Contains(html, "<code>code\n  extra\n</code>") {
		t.Errorf("fence indentation was not stripped: %s", html)
	}
}

func TestTableAfterListAndQuote(t *testing.T) {
	list := render("- item\n| a | b |\n| --- | --- |\n")
	if !strings.Contains(list, "</ul>\n<table>") {
		t.Errorf("table did not end the list: %s", list)
	}

	quote := render("> quote\n| a | b |\n| --- | --- |\n")
	if !strings.Contains(quote, "</blockquote>\n<table>") {
		t.Errorf("table did not end the blockquote: %s", quote)
	}
}
