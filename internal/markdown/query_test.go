package markdown

import (
	"strings"
	"testing"
)

func assertKinds(t *testing.T, source string, want ...blockKind) {
	t.Helper()

	lines := strings.Split(source, "\n")
	specs := scanBlocks(lines)
	if len(specs) != len(want) {
		t.Fatalf("got %d blocks, want %d: %+v", len(specs), len(want), specs)
	}
	for i := range want {
		if specs[i].kind != want[i] {
			t.Fatalf("block %d: got %v, want %v (all: %+v)", i, specs[i].kind, want[i], specs)
		}
	}
}

func TestScanFenceRun(t *testing.T) {
	lines := strings.Split("before\n```go\n# heading\n\ncode\n```\nafter\n", "\n")
	specs := scanBlocks(lines)

	if len(specs) != 3 {
		t.Fatalf("got %d blocks: %+v", len(specs), specs)
	}
	if specs[1].kind != blockCode || !specs[1].fenced {
		t.Fatalf("middle block = %+v", specs[1])
	}
	if specs[1].start != 1 || specs[1].end != 6 {
		t.Fatalf("fence range = [%d,%d)", specs[1].start, specs[1].end)
	}
	if specs[1].fenceChar != '`' || specs[1].fenceLength != 3 || specs[1].fenceInfo != "go" {
		t.Fatalf("fence metadata = %+v", specs[1])
	}
}

func TestScanUnclosedFence(t *testing.T) {
	lines := strings.Split("```go\ncode\n", "\n")
	specs := scanBlocks(lines)

	if len(specs) != 1 || specs[0].kind != blockCode || specs[0].end != len(lines) {
		t.Fatalf("specs = %+v", specs)
	}
}

func TestScanParagraphsAndBlank(t *testing.T) {
	assertKinds(t, "a\nb\n\nc\n",
		blockParagraph,
		blockParagraph,
	)
}

func TestScanTableRun(t *testing.T) {
	lines := strings.Split("| a | b |\n| --- | --- |\n| 1 | 2 |\n\nafter\n", "\n")
	specs := scanBlocks(lines)

	if len(specs) != 2 || specs[0].kind != blockTable || specs[1].kind != blockParagraph {
		t.Fatalf("specs = %+v", specs)
	}
	if len(specs[0].tableAligns) != 2 {
		t.Fatalf("table aligns = %+v", specs[0].tableAligns)
	}
	if len(specs[0].tableRows) != 2 || !specs[0].tableRows[0].header {
		t.Fatalf("table rows = %+v", specs[0].tableRows)
	}
	if got := specs[0].tableRows[1].cells; len(got) != 2 || got[0] != "1" || got[1] != "2" {
		t.Fatalf("table body cells = %+v", got)
	}
}

func TestScanListRun(t *testing.T) {
	lines := strings.Split("- a\n\n- b\n\nafter\n", "\n")
	specs := scanBlocks(lines)

	if len(specs) != 2 || specs[0].kind != blockList || specs[1].kind != blockParagraph {
		t.Fatalf("specs = %+v", specs)
	}
	if len(specs[0].listItems) != 2 {
		t.Fatalf("list items = %+v", specs[0].listItems)
	}
	for index, item := range specs[0].listItems {
		if item.marker.indent != 0 || item.start != index*2 {
			t.Fatalf("item %d = %+v", index, item)
		}
	}
}

func TestScanQuoteRun(t *testing.T) {
	assertKinds(t, "> a\n> b\n\nafter\n",
		blockBlockquote,
		blockParagraph,
	)
}

func TestScanContainerRun(t *testing.T) {
	lines := strings.Split("::: tip\nbody\n:::\nafter\n", "\n")
	specs := scanBlocks(lines)

	if len(specs) != 2 || specs[0].kind != blockContainer {
		t.Fatalf("specs = %+v", specs)
	}
	if specs[0].containerKind != "tip" || specs[0].containerTitle != "TIP" {
		t.Fatalf("container metadata = %+v", specs[0])
	}
	if specs[0].contentStart != 1 || specs[0].contentEnd != 2 {
		t.Fatalf("container content range = [%d,%d)", specs[0].contentStart, specs[0].contentEnd)
	}
}
