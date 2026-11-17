package markdown

import "strings"

// blockSpec describes one top-level block: its source range and the metadata
// needed by the parser. It is produced before any AST is built.
type listItemSpec struct {
	start  int
	end    int
	marker listMarker
}

type tableRowSpec struct {
	line   int
	header bool
	cells  []string
}

type blockSpec struct {
	kind  blockKind
	start int
	end   int

	fenced      bool
	fenceChar   byte
	fenceLength int
	fenceInfo   string

	headingLevel int
	headingText  string

	listItems []listItemSpec

	containerKind   string
	containerTitle  string
	containerClosed bool

	tableAligns []string
	tableRows   []tableRowSpec
}

// scanBlocks finds top-level block ranges and their metadata in one pass.
func scanBlocks(lines []string) []blockSpec {
	specs := make([]blockSpec, 0, len(lines))

	for i := 0; i < len(lines); {
		if strings.TrimSpace(lines[i]) == "" {
			i++
			continue
		}

		spec := scanBlock(lines, i)
		if spec.end < i {
			// Defensive: a scanner must always advance, even if a future
			// recognizer forgets to consume its first line.
			spec.end = i
		}

		specs = append(specs, spec)
		i = spec.end + 1
	}

	return specs
}

func scanBlock(lines []string, start int) blockSpec {
	line := lines[start]
	spec := blockSpec{start: start}

	if char, length, info, ok := fenceStart(line); ok {
		spec.kind = blockCode
		spec.fenced = true
		spec.fenceChar = char
		spec.fenceLength = length
		spec.fenceInfo = info
		spec.end = scanFenceEnd(lines, start)
		return spec
	}

	if indentWidth(line) >= 4 {
		spec.kind = blockCode
		spec.end = scanIndentedCodeEnd(lines, start)
		return spec
	}

	if level, text, ok := headingStart(line); ok {
		spec.kind = blockHeading
		spec.headingLevel = level
		spec.headingText = text
		spec.end = start
		return spec
	}

	if thematicBreak(line) {
		spec.kind = blockThematicBreak
		spec.end = start
		return spec
	}

	if blockquoteStart(line) {
		spec.kind = blockBlockquote
		spec.end = scanBlockquoteEnd(lines, start)
		return spec
	}

	if _, ok := parseListMarker(line); ok {
		spec.kind = blockList
		end, items := scanListEnd(lines, start)
		spec.end, spec.listItems = end, items
		return spec
	}

	if kind, title, ok := containerStart(line); ok {
		spec.kind = blockContainer
		spec.containerKind = kind
		spec.containerTitle = title
		end, closed := scanContainerEnd(lines, start)
		spec.end = end
		spec.containerClosed = closed
		return spec
	}

	if isTableStart(lines, start) {
		spec.kind = blockTable
		spec.end = scanTableEnd(lines, start)
		spec.tableAligns, _ = tableDelimiter(lines[start+1])
		if len(spec.tableAligns) == 0 {
			spec.tableAligns = []string{""}
		}
		spec.tableRows = append(spec.tableRows, tableRowSpec{
			line:   start,
			header: true,
			cells:  splitTableRow(lines[start]),
		})
		for i := start + 2; i <= spec.end; i++ {
			spec.tableRows = append(spec.tableRows, tableRowSpec{
				line:  i,
				cells: splitTableRow(lines[i]),
			})
		}
		return spec
	}

	spec.kind = blockParagraph
	spec.end = scanParagraphEnd(lines, start)
	return spec
}

func scanFenceEnd(lines []string, start int) int {
	char, length, _, _ := fenceStart(lines[start])
	end := start
	for end+1 < len(lines) {
		if fenceClose(lines[end+1], char, length) {
			return end + 1
		}
		end++
	}
	return end
}

func scanIndentedCodeEnd(lines []string, start int) int {
	end := start
	for {
		next := end + 1
		for next < len(lines) && strings.TrimSpace(lines[next]) == "" {
			next++
		}
		if next >= len(lines) || indentWidth(lines[next]) < 4 {
			return end
		}
		end = next
	}
}

func scanBlockquoteEnd(lines []string, start int) int {
	end := start
	for end+1 < len(lines) && blockquoteStart(lines[end+1]) {
		end++
	}
	return end
}

func scanListEnd(lines []string, start int) (int, []listItemSpec) {
	first, _ := parseListMarker(lines[start])
	items := make([]listItemSpec, 0, 4)
	baseIndent := first.indent
	end := start
	next := start

	for next < len(lines) {
		marker, ok := parseListMarker(lines[next])
		if !ok || marker.indent < baseIndent || !sameListMarker(marker, first) {
			break
		}

		itemStart := next
		itemEnd := next
		next++

		for next < len(lines) {
			if strings.TrimSpace(lines[next]) == "" {
				blankEnd := next
				for blankEnd+1 < len(lines) && strings.TrimSpace(lines[blankEnd+1]) == "" {
					blankEnd++
				}
				itemEnd = blankEnd
				if blankEnd+1 >= len(lines) {
					next = len(lines)
					break
				}
				if indentWidth(lines[blankEnd+1]) < marker.indent+2 {
					next = blankEnd + 1
					break
				}
				itemEnd = blankEnd + 1
				next = blankEnd + 2
				continue
			}

			if indentWidth(lines[next]) < marker.indent+2 {
				break
			}
			itemEnd = next
			next++
		}

		items = append(items, listItemSpec{start: itemStart, end: itemEnd, marker: marker})
		end = itemEnd
	}

	return end, items
}

func scanTableEnd(lines []string, start int) int {
	end := start + 1
	for end+1 < len(lines) && isTableRaw(lines[end+1]) {
		end++
	}
	return end
}

func scanContainerEnd(lines []string, start int) (int, bool) {
	_, end, closed := collectContainer(lines, start)
	return end, closed
}

func scanParagraphEnd(lines []string, start int) int {
	end := start
	for end+1 < len(lines) {
		if strings.TrimSpace(lines[end+1]) == "" || startsBlock(lines, end+1) {
			break
		}
		end++
	}
	return end
}

func startsBlock(lines []string, i int) bool {
	line := lines[i]

	if _, _, _, ok := fenceStart(line); ok {
		return true
	}
	if indentWidth(line) >= 4 {
		return true
	}
	if _, _, ok := headingStart(line); ok {
		return true
	}
	if thematicBreak(line) || blockquoteStart(line) {
		return true
	}
	if _, ok := parseListMarker(line); ok {
		return true
	}
	if _, _, ok := containerStart(line); ok {
		return true
	}
	return isTableStart(lines, i)
}
