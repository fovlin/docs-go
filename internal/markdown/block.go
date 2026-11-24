package markdown

import "strings"

func (p *parser) parseBlocks(lines []string) []block {
	specs := scanBlocks(lines)
	blocks := make([]block, 0, len(specs))

	for _, spec := range specs {
		blocks = append(blocks, p.parseBlock(lines, spec))
	}

	return blocks
}

func (p *parser) parseBlock(lines []string, spec blockSpec) block {
	switch spec.kind {
	case blockCode:
		if spec.fenced {
			code, language := parseFenceRange(lines, spec)
			return block{kind: blockCode, text: code, language: language}
		}
		return block{kind: blockCode, text: parseIndentedCodeRange(lines, spec)}

	case blockHeading:
		inlines := parseInline(spec.headingText)
		headingText := plainInlines(inlines)
		id := p.uniqueHeadingID(headingText)

		p.headings = append(p.headings, Heading{Level: spec.headingLevel, Text: headingText, ID: id})
		if spec.headingLevel == 1 && p.firstH1 == "" {
			p.firstH1 = headingText
		}

		return block{kind: blockHeading, level: spec.headingLevel, id: id, inlines: inlines}

	case blockThematicBreak:
		return block{kind: blockThematicBreak}

	case blockBlockquote:
		inner := make([]string, 0, spec.end-spec.start+1)
		for j := spec.start; j <= spec.end; j++ {
			inner = append(inner, stripBlockquote(lines[j]))
		}
		return block{kind: blockBlockquote, children: p.parseBlocks(trimTrailingBlank(inner))}

	case blockList:
		return p.parseListRange(lines, spec.listItems)

	case blockContainer:
		innerEnd := spec.end
		if spec.containerClosed {
			innerEnd--
		}
		children := []string(nil)
		if innerEnd >= spec.start+1 {
			children = lines[spec.start+1 : innerEnd+1]
		}
		return block{
			kind:           blockContainer,
			containerKind:  spec.containerKind,
			containerTitle: spec.containerTitle,
			children:       p.parseBlocks(trimTrailingBlank(children)),
		}

	case blockTable:
		return parseTableSpec(spec)
	}

	return block{
		kind:    blockParagraph,
		inlines: parseInline(strings.Join(lines[spec.start:spec.end+1], "\n")),
	}
}

func parseFenceRange(lines []string, spec blockSpec) (string, string) {
	language := ""
	if fields := strings.Fields(spec.fenceInfo); len(fields) > 0 {
		language = strings.ToLower(fields[0])
	}

	contentEnd := spec.end
	if contentEnd > spec.start && fenceClose(lines[contentEnd], spec.fenceChar, spec.fenceLength) {
		contentEnd--
	} else if contentEnd == len(lines)-1 && lines[contentEnd] == "" {
		contentEnd--
	}

	indent := blockIndent(lines[spec.start])
	content := make([]string, 0, contentEnd-spec.start)
	for i := spec.start + 1; i <= contentEnd; i++ {
		content = append(content, stripIndent(lines[i], indent))
	}

	code := strings.Join(content, "\n")
	if code != "" {
		code += "\n"
	}

	return code, language
}

func parseIndentedCodeRange(lines []string, spec blockSpec) string {
	content := make([]string, 0, spec.end-spec.start+1)
	for i := spec.start; i <= spec.end; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			content = append(content, "")
			continue
		}
		content = append(content, stripIndent(lines[i], 4))
	}

	code := strings.Join(trimTrailingBlank(content), "\n")
	if code != "" {
		code += "\n"
	}

	return code
}

func (p *parser) parseListRange(lines []string, items []listItemSpec) block {
	first := items[0].marker
	list := block{
		kind:    blockList,
		ordered: first.ordered,
		start:   first.start,
	}

	for _, item := range items {
		itemLines := []string{item.marker.content}
		for i := item.start + 1; i <= item.end; i++ {
			if strings.TrimSpace(lines[i]) == "" {
				continue
			}
			itemLines = append(itemLines, stripIndent(lines[i], item.marker.indent+2))
		}

		list.children = append(list.children, block{
			kind:     blockListItem,
			children: p.parseBlocks(trimTrailingBlank(itemLines)),
		})
	}

	return list
}

func parseTableSpec(spec blockSpec) block {
	table := block{kind: blockTable}
	for _, row := range spec.tableRows {
		table.children = append(table.children, tableRow(row.cells, spec.tableAligns, row.header))
	}
	return table
}

func tableRow(cells, aligns []string, header bool) block {
	row := block{kind: blockTableRow, header: header}

	for index, align := range aligns {
		cell := ""
		if index < len(cells) {
			cell = cells[index]
		}

		row.children = append(row.children, block{
			kind:    blockTableCell,
			align:   align,
			inlines: parseInline(cell),
		})
	}

	return row
}
