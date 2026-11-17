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

	marker    listMarker
	listItems []listItemSpec

	containerKind  string
	containerTitle string
	contentStart   int
	contentEnd     int

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
		if spec.end <= i {
			spec.end = i + 1
		}
		specs = append(specs, spec)
		i = spec.end
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
		spec.end = start + 1
		return spec
	}

	if thematicBreak(line) {
		spec.kind = blockThematicBreak
		spec.end = start + 1
		return spec
	}

	if blockquoteStart(line) {
		spec.kind = blockBlockquote
		spec.end = scanBlockquoteEnd(lines, start)
		spec.contentStart = start
		spec.contentEnd = spec.end
		return spec
	}

	if marker, ok := parseListMarker(line); ok {
		spec.kind = blockList
		spec.marker = marker
		spec.end, spec.listItems = scanListEnd(lines, start)
		return spec
	}

	if kind, title, ok := containerStart(line); ok {
		spec.kind = blockContainer
		spec.containerKind = kind
		spec.containerTitle = title
		spec.end = scanContainerEnd(lines, start)
		spec.contentStart = start + 1
		spec.contentEnd = spec.end
		if spec.contentEnd > spec.contentStart && isContainerClose(lines[spec.contentEnd-1]) {
			spec.contentEnd--
		}
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
		for i := start + 2; i < spec.end; i++ {
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
	i := start + 1
	for i < len(lines) {
		if fenceClose(lines[i], char, length) {
			return i + 1
		}
		i++
	}
	return len(lines)
}

func scanIndentedCodeEnd(lines []string, start int) int {
	i := start
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "" {
			next := i + 1
			for next < len(lines) && strings.TrimSpace(lines[next]) == "" {
				next++
			}
			if next >= len(lines) || indentWidth(lines[next]) < 4 {
				return i
			}
			i = next
			continue
		}
		if indentWidth(lines[i]) < 4 {
			break
		}
		i++
	}
	return i
}

func scanBlockquoteEnd(lines []string, start int) int {
	i := start
	for i < len(lines) && blockquoteStart(lines[i]) {
		i++
	}
	return i
}

func scanListEnd(lines []string, start int) (int, []listItemSpec) {
	first, _ := parseListMarker(lines[start])
	items := make([]listItemSpec, 0, 4)
	i := start
	baseIndent := first.indent

	for i < len(lines) {
		marker, ok := parseListMarker(lines[i])
		if !ok || marker.indent < baseIndent || !sameListMarker(marker, first) {
			break
		}

		itemStart := i
		i++
		for i < len(lines) {
			next := i
			for next < len(lines) && strings.TrimSpace(lines[next]) == "" {
				next++
			}
			if next >= len(lines) {
				i = next
				break
			}
			i = next
			if indentWidth(lines[i]) < marker.indent+2 {
				break
			}
			i++
		}

		items = append(items, listItemSpec{start: itemStart, end: i, marker: marker})
	}

	return i, items
}

func scanTableEnd(lines []string, start int) int {
	i := start + 2
	for i < len(lines) && isTableRaw(lines[i]) {
		i++
	}
	return i
}

func scanContainerEnd(lines []string, start int) int {
	_, next := collectContainer(lines, start)
	return next
}

func scanParagraphEnd(lines []string, start int) int {
	i := start + 1
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "" || startsBlock(lines, i) {
			break
		}
		i++
	}
	return i
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
