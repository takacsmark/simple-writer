package main

type lineRenderKind int

const (
	lineRenderRaw lineRenderKind = iota
	lineRenderMarkdown
)

type visualRow struct {
	lineIndex int
	kind      lineRenderKind
	rawStart  int
	rawChunk  []rune
	rendered  string
}

func (e *editor) lineRenderKind(lineIdx int) lineRenderKind {
	if lineIdx < 0 || lineIdx >= len(e.lines) {
		return lineRenderMarkdown
	}

	// Keep copy flash behavior stable by keeping the flashed line renderable as raw.
	if e.flashLine == lineIdx && e.flashLine >= 0 {
		return lineRenderRaw
	}

	switch e.mode {
	case modeInsert, modeNormal:
		if lineIdx == e.row {
			return lineRenderRaw
		}
	case modeVisual, modeVisualLine:
		start, end := e.visualSelectionLineRange()
		if ts, te, ok := e.tableBlockForLine(lineIdx); ok && rangesOverlap(start, end, ts, te) {
			return lineRenderRaw
		}
		if lineIdx >= start && lineIdx <= end {
			return lineRenderRaw
		}
	}

	return lineRenderMarkdown
}

func (e *editor) visualSelectionLineRange() (int, int) {
	if e.mode == modeVisual {
		sr, _, er, _ := e.selectionEndpoints()
		return sr, er
	}
	return e.visualRangeRows()
}

func (e *editor) visualRows() []visualRow {
	if e.markdown == nil {
		e.markdown = newMarkdownRenderer()
	}

	cw := e.contentWidth()
	out := make([]visualRow, 0, len(e.lines))
	for i := 0; i < len(e.lines); {
		start, end, kind, isFrontMatter := e.frontMatterBlockAt(i)
		if isFrontMatter {
			out = append(out, e.visualRowsForFrontMatterBlock(start, end, kind, cw)...)
			i = end + 1
			continue
		}

		start, end, block, isCode := e.codeBlockStartAt(i)
		if isCode {
			out = append(out, e.visualRowsForCodeBlock(start, end, block, cw)...)
			i = end + 1
			continue
		}

		start, end, isTable := e.tableBlockStartAt(i)
		if isTable {
			out = append(out, e.visualRowsForTableBlock(start, end, cw)...)
			i = end + 1
			continue
		}

		line := e.lines[i]
		if e.lineRenderKind(i) == lineRenderRaw {
			out = append(out, rawWrappedRows(i, line, cw)...)
		} else {
			rendered, err := e.markdown.renderMarkdownLine(line, cw)
			if err != nil {
				out = append(out, rawWrappedRows(i, line, cw)...)
			} else {
				for _, row := range rendered {
					out = append(out, visualRow{
						lineIndex: i,
						kind:      lineRenderMarkdown,
						rendered:  row,
					})
				}
			}
		}
		i++
	}

	return out
}

func (e *editor) visualRowsForFrontMatterBlock(start, end int, kind frontMatterKind, width int) []visualRow {
	out := make([]visualRow, 0, end-start+1)
	for i := start; i <= end; i++ {
		if e.lineRenderKind(i) == lineRenderRaw {
			out = append(out, rawWrappedRows(i, e.lines[i], width)...)
			continue
		}

		rendered, err := e.markdown.renderMarkdownFrontMatterLine(kind, e.lines[i], width)
		if err != nil {
			out = append(out, rawWrappedRows(i, e.lines[i], width)...)
			continue
		}
		for _, row := range rendered {
			out = append(out, visualRow{
				lineIndex: i,
				kind:      lineRenderMarkdown,
				rendered:  row,
			})
		}
	}
	return out
}

func (e *editor) visualRowsForCodeBlock(start, end int, block fencedCodeBlock, width int) []visualRow {
	out := make([]visualRow, 0, end-start+1)
	for i := start; i <= end; i++ {
		if e.lineRenderKind(i) == lineRenderRaw {
			out = append(out, rawWrappedRows(i, e.lines[i], width)...)
			continue
		}

		var (
			rendered []string
			err      error
		)
		switch {
		case i == start:
			rendered, err = e.markdown.renderMarkdownLine(e.lines[i], width)
		case isCodeFenceEnd(e.lines[i], block.fenceRune, block.fenceLen):
			rendered, err = e.markdown.renderMarkdownLine(e.lines[i], width)
		default:
			rendered, err = e.markdown.renderMarkdownCodeLine(block.lang, e.lines[i], width)
		}
		if err != nil {
			out = append(out, rawWrappedRows(i, e.lines[i], width)...)
			continue
		}
		for _, row := range rendered {
			out = append(out, visualRow{
				lineIndex: i,
				kind:      lineRenderMarkdown,
				rendered:  row,
			})
		}
	}
	return out
}

func (e *editor) visualRowsForTableBlock(start, end, width int) []visualRow {
	if e.shouldRenderWholeTableRaw(start, end) {
		out := make([]visualRow, 0, end-start+1)
		for i := start; i <= end; i++ {
			out = append(out, rawWrappedRows(i, e.lines[i], width)...)
		}
		return out
	}

	rendered, err := e.markdown.renderMarkdownTableBlock(e.lines[start:end+1], width)
	if err != nil || len(rendered) != end-start+1 {
		out := make([]visualRow, 0, end-start+1)
		for i := start; i <= end; i++ {
			if e.lineRenderKind(i) == lineRenderRaw {
				out = append(out, rawWrappedRows(i, e.lines[i], width)...)
				continue
			}
			lineRows, lineErr := e.markdown.renderMarkdownLine(e.lines[i], width)
			if lineErr != nil {
				out = append(out, rawWrappedRows(i, e.lines[i], width)...)
				continue
			}
			for _, row := range lineRows {
				out = append(out, visualRow{
					lineIndex: i,
					kind:      lineRenderMarkdown,
					rendered:  row,
				})
			}
		}
		return out
	}

	out := make([]visualRow, 0, end-start+1)
	for i := start; i <= end; i++ {
		if e.lineRenderKind(i) == lineRenderRaw {
			out = append(out, rawWrappedRows(i, e.lines[i], width)...)
			continue
		}
		out = append(out, visualRow{
			lineIndex: i,
			kind:      lineRenderMarkdown,
			rendered:  rendered[i-start],
		})
	}
	return out
}

func (e *editor) shouldRenderWholeTableRaw(start, end int) bool {
	if e.mode != modeVisual && e.mode != modeVisualLine {
		return false
	}
	sr, er := e.visualSelectionLineRange()
	return rangesOverlap(sr, er, start, end)
}

func rawWrappedRows(lineIndex int, line []rune, width int) []visualRow {
	if len(line) == 0 {
		return []visualRow{{
			lineIndex: lineIndex,
			kind:      lineRenderRaw,
			rawStart:  0,
			rawChunk:  nil,
		}}
	}

	out := make([]visualRow, 0, (len(line)-1)/width+1)
	for s := 0; s < len(line); s += width {
		end := min(len(line), s+width)
		out = append(out, visualRow{
			lineIndex: lineIndex,
			kind:      lineRenderRaw,
			rawStart:  s,
			rawChunk:  line[s:end],
		})
	}
	return out
}

func (e *editor) cursorLineColumn() int {
	line := e.lines[e.row]
	c := e.col
	if e.mode == modeInsert {
		if c < 0 {
			return 0
		}
		if c > len(line) {
			return len(line)
		}
		return c
	}

	maxCol := 0
	if len(line) > 0 {
		maxCol = len(line) - 1
	}
	if c < 0 {
		return 0
	}
	if c > maxCol {
		return maxCol
	}
	return c
}

func (e *editor) cursorVisualWithRows(rows []visualRow) (row, col int) {
	cw := e.contentWidth()
	line := e.lines[e.row]
	c := e.cursorLineColumn()

	lineStartRow := -1
	for i := range rows {
		if rows[i].lineIndex == e.row {
			lineStartRow = i
			break
		}
	}
	if lineStartRow < 0 {
		return 0, 0
	}

	row = lineStartRow + (c / cw)
	col = c % cw

	// Preserve the insert-mode cursor behavior at visual wrap boundaries.
	if e.mode == modeInsert && len(line) > 0 && c == len(line) && c%cw == 0 {
		row++
		col = 0
	}

	return row, col
}

func (e *editor) cursorVisual() (row, col int) {
	return e.cursorVisualWithRows(e.visualRows())
}

func (e *editor) ensureCursorRowVisible(cursorRow int) {
	body := e.bodyHeight()
	if cursorRow < e.scroll {
		e.scroll = cursorRow
	}
	if cursorRow >= e.scroll+body {
		e.scroll = cursorRow - body + 1
	}
	if e.scroll < 0 {
		e.scroll = 0
	}
}

func (e *editor) ensureCursorVisible() {
	cursorRow, _ := e.cursorVisualWithRows(e.visualRows())
	e.ensureCursorRowVisible(cursorRow)
}
