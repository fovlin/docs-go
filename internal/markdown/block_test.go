package markdown

import (
	"strings"
	"testing"
)

func testBlocks(source string) []block {
	return Parse([]byte(source)).blocks
}

func paragraphOf(t *testing.T, item block) *paragraphBlock {
	t.Helper()

	paragraph, ok := item.(*paragraphBlock)
	if !ok {
		t.Fatalf("block is %T, want paragraph", item)
	}

	return paragraph
}

func TestFencedCodeBlock(t *testing.T) {
	blocks := testBlocks("before\n```go\n# heading\n\ncode\n```\nafter\n")
	if len(blocks) != 3 {
		t.Fatalf("got %d blocks, want 3", len(blocks))
	}

	code, ok := blocks[1].(*codeBlock)
	if !ok || code.language != "go" || code.text != "# heading\n\ncode\n" {
		t.Fatalf("code block = %#v", blocks[1])
	}
}

func TestUnclosedFence(t *testing.T) {
	blocks := testBlocks("```go\ncode\n")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}

	code, ok := blocks[0].(*codeBlock)
	if !ok || code.language != "go" || code.text != "code\n" {
		t.Fatalf("code block = %#v", blocks[0])
	}
}

func TestParagraphBlocks(t *testing.T) {
	blocks := testBlocks("a\nb\n\nc\n")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}

	if got := plainInlines(paragraphOf(t, blocks[0]).inlines); got != "a b" {
		t.Fatalf("first paragraph = %q", got)
	}
}

func TestTableBlock(t *testing.T) {
	blocks := testBlocks("| a | b |\n| --- | --- |\n| 1 | 2 |\n\nafter\n")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}

	table, ok := blocks[0].(*tableBlock)
	if !ok || len(table.rows) != 2 {
		t.Fatalf("table = %#v", blocks[0])
	}

	header := table.rows[0]
	if len(header.cells) != 2 || plainInlines(header.cells[0].inlines) != "a" || header.cells[1].align != "" {
		t.Fatalf("header = %#v", header)
	}

	body := table.rows[1]
	if len(body.cells) != 2 || plainInlines(body.cells[0].inlines) != "1" || plainInlines(body.cells[1].inlines) != "2" {
		t.Fatalf("body = %#v", body)
	}
}

func TestListBlock(t *testing.T) {
	blocks := testBlocks("- a\n\n- b\n\nafter\n")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}

	list, ok := blocks[0].(*listBlock)
	if !ok || len(list.items) != 2 {
		t.Fatalf("list = %#v", blocks[0])
	}

	for index, want := range []string{"a", "b"} {
		if got := plainInlines(paragraphOf(t, list.items[index].children[0]).inlines); got != want {
			t.Fatalf("item %d = %q, want %q", index, got, want)
		}
	}

	paragraphOf(t, blocks[1])
}

func TestBlockquoteBlock(t *testing.T) {
	blocks := testBlocks("> a\n> b\n\nafter\n")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}

	quote, ok := blocks[0].(*blockquoteBlock)
	if !ok || len(quote.children) != 1 {
		t.Fatalf("blockquote = %#v", blocks[0])
	}

	if got := plainInlines(paragraphOf(t, quote.children[0]).inlines); got != "a b" {
		t.Fatalf("quote text = %q", got)
	}
}

func TestContainerBlock(t *testing.T) {
	blocks := testBlocks("::: tip\nbody\n:::\nafter\n")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}

	container, ok := blocks[0].(*containerBlock)
	if !ok || container.kind != "tip" || container.title != "TIP" || len(container.children) != 1 {
		t.Fatalf("container = %#v", blocks[0])
	}

	if got := plainInlines(paragraphOf(t, container.children[0]).inlines); got != "body" {
		t.Fatalf("container text = %q", got)
	}
}

func TestAdjacentDifferentLists(t *testing.T) {
	blocks := testBlocks("- a\n* b\n")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}

	for _, item := range blocks {
		if _, ok := item.(*listBlock); !ok {
			t.Fatalf("block is %T, want list", item)
		}
	}
}

func TestListItemBlankLineSeparatesParagraphs(t *testing.T) {
	blocks := testBlocks("- one\n\n  two\n")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}

	list := blocks[0].(*listBlock)
	if len(list.items) != 1 || len(list.items[0].children) != 2 {
		t.Fatalf("list item = %#v", list.items)
	}
}

func TestOpensBlock(t *testing.T) {
	openers := []string{
		"# heading\n",
		"```\ncode\n```\n",
		"    code\n",
		"---\n",
		"> quote\n",
		"- item\n",
		"::: tip\nx\n:::\n",
		"| a | b |\n| --- | --- |\n",
	}

	for _, source := range openers {
		if blockKindAt(strings.Split(source, "\n"), 0) == blockParagraph {
			t.Errorf("%q did not open a block", source)
		}
	}

	if blockKindAt([]string{"plain text"}, 0) != blockParagraph {
		t.Fatal("plain paragraph opened a block")
	}
}
