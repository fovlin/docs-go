package markdown

import (
	"strings"
	"testing"
)

// TestWhitespaceOnlyListMarkerIsNotList pins the fix for a marker followed
// only by whitespace, which used to produce an empty list item.
func TestWhitespaceOnlyListMarkerIsNotList(t *testing.T) {
	html := render("-  \n")

	if strings.Contains(html, "<ul>") || strings.Contains(html, "<li>") {
		t.Fatalf("whitespace-only marker became a list: %s", html)
	}
	if !strings.Contains(html, "<p>-  </p>") {
		t.Fatalf("whitespace-only marker did not stay a paragraph: %s", html)
	}
}

// TestListMarkerContentIndent pins the unified marker handling: exactly one
// separator follows the marker, tabs advance to the next tab stop, and the rest
// of the line keeps its own indentation.
func TestListMarkerContentIndent(t *testing.T) {
	cases := []struct {
		line   string
		indent int
	}{
		{"- a", 2},
		{"  - a", 4},
		{"1. a", 3},
		{"10. a", 4},
		{"-\ta", 4},
		{"   -\ta", 8},
	}

	for _, tc := range cases {
		marker, ok := parseListMarker(tc.line)
		if !ok {
			t.Fatalf("parseListMarker(%q) failed", tc.line)
		}
		if marker.contentIndent != tc.indent {
			t.Errorf("parseListMarker(%q).contentIndent = %d, want %d", tc.line, marker.contentIndent, tc.indent)
		}
	}

	for _, source := range []string{"-     code\n", "1.     code\n", "10.     code\n"} {
		html := render(source)

		if !strings.Contains(html, `<li><div class="code-block"><pre><code>code`) {
			t.Errorf("%q did not keep the code indentation: %s", source, html)
		}
	}
}

// TestListContinuationIndent pins the item boundary at the marker's content
// column, including multi-digit ordered markers and tab-separated content.
func TestListContinuationIndent(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{"1. a\n  b\n", "</ol>\n<p>  b</p>"},
		{"1. a\n   b\n", "<li>a<br>\nb</li>"},
		{"10. a\n   b\n", "</ol>\n<p>   b</p>"},
		{"10. a\n    b\n", "<li>a<br>\nb</li>"},
		{"-\tcode\n  cont\n", "</ul>\n<p>  cont</p>"},
		{"-\tcode\n    cont\n", "<li>code<br>\ncont</li>"},
	}

	for _, tc := range cases {
		if html := render(tc.source); !strings.Contains(html, tc.want) {
			t.Errorf("%q: missing %q in %s", tc.source, tc.want, html)
		}
	}
}

// TestLooseListShape pins CommonMark's loose-list HTML: when any item is
// separated by a blank line, every direct paragraph is wrapped in <p>.
func TestLooseListShape(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{"- one\n\n- two\n", "<li><p>one</p>\n</li>\n<li><p>two</p>"},
		{"- one\n\n  two\n", "<li><p>one</p>\n<p>two</p>\n</li>"},
		{"- one\n- two\n", "<li>one</li>"},
	}

	for _, tc := range cases {
		if html := render(tc.source); !strings.Contains(html, tc.want) {
			t.Errorf("%q: missing %q in %s", tc.source, tc.want, html)
		}
	}
}

// TestTableRowsAreOrdinaryText pins the replacement of the old first-byte
// heuristic: these lines contain inline syntax but are not block starts.
func TestTableRowsAreOrdinaryText(t *testing.T) {
	html := render("| a | b |\n| --- | --- |\n*em* | x\n#tag | y\n+1 | z\n-1 | w\n")

	for _, want := range []string{
		"<td><em>em</em></td>",
		"<td>#tag</td>",
		"<td>+1</td>",
		"<td>-1</td>",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in %s", want, html)
		}
	}
}

// TestTableStopsAtBlockStarts pins that a table body ends when the next line
// starts any other block, not just when its first byte is on a list.
func TestTableStopsAtBlockStarts(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{"| h |\n| --- |\n> q | x\n", "</table>\n<blockquote>"},
		{"| h |\n| --- |\n## Next | section\n", "</table>\n<h2 id=\"next-section\">"},
		{"| h |\n| --- |\n1. b | c\n", "</table>\n<ol>"},
		{"| h |\n| --- |\n    | c\n", "</table>\n<div class=\"code-block\">"},
		{"| h |\n| --- |\n```\nx\n```\n", "</table>\n<div class=\"code-block\">"},
		{"| h |\n| --- |\n::: tip\nx\n:::\n", "</table>\n<div class=\"custom-block tip\">"},
	}

	for _, tc := range cases {
		html := render(tc.source)
		if !strings.Contains(html, tc.want) {
			t.Errorf("%q: missing %q in %q", tc.source, tc.want, html)
		}
		if strings.Contains(html, "<tbody>") {
			t.Errorf("%q: table body swallowed the following block: %q", tc.source, html)
		}
	}
}

// TestTableHeaderBlockOpenerWins pins that a header line starting another
// block is that block, not a table header.
func TestTableHeaderBlockOpenerWins(t *testing.T) {
	for _, source := range []string{
		"1. h | x\n--- | ---\n",
		"# h | x\n--- | ---\n",
		"> h | x\n--- | ---\n",
	} {
		if html := render(source); strings.Contains(html, "<table>") {
			t.Errorf("%q: block opener became a table header: %s", source, html)
		}
	}
}

// TestTableHeaderUsesOrdinaryText pins that the header follows the same
// ordinary-text rule as body rows.
func TestTableHeaderUsesOrdinaryText(t *testing.T) {
	html := render("*em* | h2\n--- | ---\n")

	if !strings.Contains(html, "<th><em>em</em></th>") || !strings.Contains(html, "<th>h2</th>") {
		t.Fatalf("header with inline syntax was not a table: %s", html)
	}
}

// TestFirstParagraphSkipsEmptyText pins the parser-side metadata recording:
// an invisible first paragraph must not block the description fallback.
func TestFirstParagraphSkipsEmptyText(t *testing.T) {
	document := Parse([]byte("<!-- hidden -->\n\nVisible paragraph.\n"))

	if document.FirstParagraph != "Visible paragraph." {
		t.Fatalf("FirstParagraph = %q", document.FirstParagraph)
	}
}

// TestStripIndentTabStop guards the tab crossing the requested column: the
// part beyond the column must remain as spaces, not be swallowed.
func TestStripIndentTabStop(t *testing.T) {
	cases := []struct {
		line  string
		width int
		want  string
	}{
		{"\t   code", 3, "    code"},
		{"  \tcode", 3, " code"},
	}

	for _, tc := range cases {
		if got := stripIndent(tc.line, tc.width); got != tc.want {
			t.Errorf("stripIndent(%q, %d) = %q, want %q", tc.line, tc.width, got, tc.want)
		}
	}

	for _, tc := range []struct{ source, want string }{
		{"1. a\n\t   code\n", "<li>a\n<div class=\"code-block\"><pre><code>code"},
		{"  ```\n\tcode\n  ```\n", "<code>  code\n</code>"},
	} {
		if html := render(tc.source); !strings.Contains(html, tc.want) {
			t.Errorf("%q: missing %q in %q", tc.source, tc.want, html)
		}
	}
}

// TestContainerEndsAtBlockBoundary pins the block-scoped container terminator:
// nested list fences and indented ::: lines belong to their block, not to the
// outer container.
func TestContainerEndsAtBlockBoundary(t *testing.T) {
	html := render("::: tip\n- ```\n  :::\n  ```\n:::\n")

	if !strings.Contains(html, "<li><div class=\"code-block\"><pre><code>:::\n</code></pre></div>") {
		t.Fatalf("list fenced block was closed early: %s", html)
	}
	if strings.Count(html, "custom-block tip") != 1 {
		t.Fatalf("container boundary wrong: %s", html)
	}

	indented := render("::: outer\n- ::: inner\n  x\n  :::\n:::\n")
	if !strings.Contains(indented, "::: inner<br>\nx<br>\n:::") {
		t.Fatalf("indented container marker leaked out of list content: %s", indented)
	}

	nested := render("::: warning\n- ::: tip\n  x\n  :::\n:::\n")
	if strings.Count(nested, "custom-block warning") != 1 ||
		strings.Count(nested, "custom-block tip") != 1 ||
		!strings.Contains(nested, `<li><div class="custom-block tip">`) ||
		strings.Contains(nested, "<p>:::</p>") {
		t.Fatalf("container did not nest inside a list item: %s", nested)
	}
}

// TestUnclosedFenceTailInsideContainer pins the block-scoped behavior for an
// unclosed fence: trailing blank lines stay with the fence, as at top level.
func TestUnclosedFenceTailInsideContainer(t *testing.T) {
	html := render("::: tip\n```\ncode\n\n\n")

	if !strings.Contains(html, "<code>code\n\n\n</code>") {
		t.Fatalf("unclosed fence tail changed: %q", html)
	}
}

// TestUnclosedContainerAtEOF guards the container body slice; an unclosed
// container with no following line used to panic.
func TestUnclosedContainerAtEOF(t *testing.T) {
	for _, source := range []string{"::: tip\n", ":::tip\n", "::: tip", ":::tip"} {
		html := render(source)
		if !strings.Contains(html, `<div class="custom-block tip">`) || strings.Count(html, "</div>") != 1 {
			t.Errorf("%q did not render one closed container: %s", source, html)
		}
	}
}

// TestMalformedInlineHTMLIsText pins the stricter tag matcher: malformed tags
// are escaped text, not trusted raw HTML.
func TestMalformedInlineHTMLIsText(t *testing.T) {
	for _, source := range []string{"<x_y>\n", "<x/ >\n", "<!DOCTYPE html>\n"} {
		html := render(source)
		if strings.Contains(html, "<x_y>") || strings.Contains(html, "<x/ >") || strings.Contains(html, "<!DOCTYPE") {
			t.Errorf("%q was passed through as HTML: %s", source, html)
		}
		if !strings.Contains(html, "&lt;") {
			t.Errorf("%q was not escaped: %s", source, html)
		}
	}
}

// TestTableDelimiterRejectsMultipleColons pins the stricter delimiter regex:
// a malformed cell is not a table and the lines stay a paragraph.
func TestTableDelimiterRejectsMultipleColons(t *testing.T) {
	for _, source := range []string{"| a |\n| ::--- |\n", "| a |\n| ---:: |\n"} {
		if html := render(source); strings.Contains(html, "<table>") {
			t.Fatalf("malformed delimiter became a table: %s", html)
		}
	}
}

// TestInlineHTMLQuotedAngle keeps quoted '>' inside inline HTML attributes.
func TestInlineHTMLQuotedAngle(t *testing.T) {
	html := render(`<span title="a > b">x</span>` + "\n")

	if !strings.Contains(html, `<p><span title="a > b">x</span></p>`) {
		t.Fatalf("quoted angle broke inline HTML: %s", html)
	}
}
