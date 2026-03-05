package main

import (
	"fmt"
	"strings"
	"time"
)

type rgb struct {
	R int
	G int
	B int
}

type appearanceConfig struct {
	Background    rgb
	Highlight     rgb
	Text          rgb
	ModeIndicator rgb
	CommandText   rgb
	CommandPrompt rgb
	CommandError  rgb
}

type styleCodes struct {
	bgDark          string
	bgSelection     string
	fgText          string
	fgModeIndicator string
	fgCommandText   string
	fgCommandPrompt string
	fgCommandError  string
}

var DefaultAppearance = appearanceConfig{
	Background:    rgb{R: 17, G: 18, B: 20},
	Highlight:     rgb{R: 43, G: 45, B: 51},
	Text:          rgb{R: 231, G: 231, B: 232},
	ModeIndicator: rgb{R: 109, G: 109, B: 114},
	CommandText:   rgb{R: 223, G: 225, B: 228},
	CommandPrompt: rgb{R: 173, G: 176, B: 182},
	CommandError:  rgb{R: 231, G: 98, B: 98},
}

func (c appearanceConfig) toStyleCodes() styleCodes {
	return styleCodes{
		bgDark:          ansiBgRGB(c.Background),
		bgSelection:     ansiBgRGB(c.Highlight),
		fgText:          ansiFgRGB(c.Text),
		fgModeIndicator: ansiFgRGB(c.ModeIndicator),
		fgCommandText:   ansiFgRGB(c.CommandText),
		fgCommandPrompt: ansiFgRGB(c.CommandPrompt),
		fgCommandError:  ansiFgRGB(c.CommandError),
	}
}

func ansiBgRGB(v rgb) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", v.R, v.G, v.B)
}

func ansiFgRGB(v rgb) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", v.R, v.G, v.B)
}

func (e *editor) isSelectedCell(lineIdx, col int, hasChar bool) bool {
	switch e.mode {
	case modeVisualLine:
		sr, er := e.visualRangeRows()
		return lineIdx >= sr && lineIdx <= er
	case modeVisual:
		sr, sc, er, ec := e.selectionEndpoints()
		if lineIdx < sr || lineIdx > er {
			return false
		}
		if sr == er {
			if hasChar {
				return col >= sc && col <= ec
			}
			return len(e.lines[lineIdx]) == 0 && col == 0
		}
		if lineIdx == sr {
			if hasChar {
				return col >= sc
			}
			return len(e.lines[lineIdx]) == 0 && col == 0
		}
		if lineIdx == er {
			if hasChar {
				return col <= ec
			}
			return len(e.lines[lineIdx]) == 0 && col == 0
		}
		if hasChar {
			return true
		}
		return len(e.lines[lineIdx]) == 0 && col == 0
	default:
		return false
	}
}

func (e *editor) renderWrappedSegment(lineIdx, start int, chunk []rune, width int, flashLine int) string {
	var b strings.Builder
	baseFg := e.style.fgText
	line := []rune(nil)
	if lineIdx >= 0 && lineIdx < len(e.lines) {
		line = e.lines[lineIdx]
		if fg, ok := headingAnsiFg(rawHeadingLevel(e.lines[lineIdx])); ok {
			baseFg = fg
		}
	}
	spans := make([]linkColorSpan, 0, 8)
	if kind, ok := e.frontMatterKindForLine(lineIdx); ok {
		if e.markdown == nil {
			e.markdown = newMarkdownRenderer()
		}
		fmSpans, err := e.markdown.rawFrontMatterColorSpans(kind, line, max(width, len(line)))
		if err == nil {
			spans = append(spans, fmSpans...)
		}
	}
	spans = append(spans, rawLinkColorSpans(line)...)

	fgForCol := func(col int, hasChar bool) string {
		if !hasChar {
			return baseFg
		}
		for i := len(spans) - 1; i >= 0; i-- {
			span := spans[i]
			if col >= span.start && col < span.end {
				return span.fg
			}
		}
		return baseFg
	}

	setStyle := func(selected bool, fg string) {
		if selected {
			b.WriteString(e.style.bgSelection)
		} else {
			b.WriteString(e.style.bgDark)
		}
		b.WriteString(fg)
	}

	currentSelected := false
	currentFg := ""

	for i := 0; i < width; i++ {
		hasChar := i < len(chunk)
		col := start + i
		shouldSelect := e.isSelectedCell(lineIdx, col, hasChar) || lineIdx == flashLine
		fg := fgForCol(col, hasChar)
		if shouldSelect != currentSelected || fg != currentFg {
			setStyle(shouldSelect, fg)
			currentSelected = shouldSelect
			currentFg = fg
		}
		if hasChar {
			b.WriteRune(chunk[i])
		} else {
			b.WriteByte(' ')
		}
	}

	return b.String()
}

type linkColorSpan struct {
	start int
	end   int
	fg    string
}

func rawLinkColorSpans(line []rune) []linkColorSpan {
	if len(line) == 0 {
		return nil
	}

	out := make([]linkColorSpan, 0, 2)
	for i := 0; i < len(line); {
		if line[i] == '!' && i+1 < len(line) && line[i+1] == '[' && !isEscapedRune(line, i) {
			i++
			continue
		}
		if line[i] != '[' || isEscapedRune(line, i) {
			i++
			continue
		}

		textEnd, ok := scanBracketRunesEnd(line, i)
		if !ok {
			i++
			continue
		}

		j := textEnd + 1
		for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
			j++
		}

		if j < len(line) && line[j] == '(' {
			destEnd, ok := scanParenRunesEnd(line, j)
			if ok {
				out = append(out,
					linkColorSpan{start: i, end: textEnd + 1, fg: linkLabelAnsiFg()},
					linkColorSpan{start: j, end: destEnd + 1, fg: linkDestAnsiFg()},
				)
				i = destEnd + 1
				continue
			}
		}

		if j < len(line) && line[j] == '[' {
			refEnd, ok := scanBracketRunesEnd(line, j)
			if ok {
				out = append(out,
					linkColorSpan{start: i, end: textEnd + 1, fg: linkLabelAnsiFg()},
					linkColorSpan{start: j, end: refEnd + 1, fg: linkDestAnsiFg()},
				)
				i = refEnd + 1
				continue
			}
		}

		i++
	}

	return out
}

func scanBracketRunesEnd(line []rune, start int) (int, bool) {
	depth := 0
	for i := start; i < len(line); i++ {
		if isEscapedRune(line, i) {
			continue
		}
		switch line[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func scanParenRunesEnd(line []rune, start int) (int, bool) {
	depth := 0
	quote := rune(0)
	inAngle := false
	for i := start; i < len(line); i++ {
		if isEscapedRune(line, i) {
			continue
		}

		ch := line[i]
		if quote != 0 {
			if ch == quote {
				quote = 0
			}
			continue
		}
		if inAngle {
			if ch == '>' {
				inAngle = false
			}
			continue
		}

		switch ch {
		case '"', '\'':
			quote = ch
		case '<':
			inAngle = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func isEscapedRune(line []rune, i int) bool {
	backslashes := 0
	for j := i - 1; j >= 0 && line[j] == '\\'; j-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func rawHeadingLevel(line []rune) int {
	spaceCount := 0
	i := 0
	for i < len(line) && line[i] == ' ' && spaceCount < 3 {
		i++
		spaceCount++
	}
	if i >= len(line) || line[i] != '#' {
		return 0
	}

	level := 0
	for i < len(line) && line[i] == '#' && level < 6 {
		level++
		i++
	}
	if level == 0 || level > 6 {
		return 0
	}
	// Require a separator space after the hash run for ATX headings.
	if i < len(line) && line[i] != ' ' && line[i] != '\t' {
		return 0
	}
	return level
}

func applyBaseStyleAfterReset(s, base string) string {
	if s == "" || !strings.Contains(s, ansiReset) {
		return s
	}
	return strings.ReplaceAll(s, ansiReset, ansiReset+base)
}

func (e *editor) renderFrame() string {
	cw := e.contentWidth()
	body := e.bodyHeight()
	leftPad := max(0, (e.width-cw)/2)
	rightPad := max(0, e.width-leftPad-cw)
	flashLine := e.activeFlashLine(time.Now())
	rows := e.visualRows()
	cursorVisualRow, cursorVisualCol := e.cursorVisualWithRows(rows)
	e.ensureCursorRowVisible(cursorVisualRow)

	cursorScreenRow := cursorVisualRow - e.scroll + 1
	if cursorScreenRow < 1 {
		cursorScreenRow = 1
	}
	if cursorScreenRow > body {
		cursorScreenRow = body
	}
	cursorScreenCol := min(e.width, max(1, leftPad+cursorVisualCol+1))

	var b strings.Builder
	b.Grow((e.width + 32) * (e.height + 2))

	b.WriteString(ansiHideCursor)
	if e.mode == modeInsert {
		b.WriteString(ansiSteadyBar)
	} else {
		b.WriteString(ansiSteadyBlock)
	}

	for y := 0; y < body; y++ {
		visualIdx := e.scroll + y
		base := e.style.bgDark + e.style.fgText
		b.WriteString(fmt.Sprintf("\x1b[%d;1H", y+1))
		b.WriteString(e.style.bgDark)
		b.WriteString(e.style.fgText)
		if leftPad > 0 {
			b.WriteString(strings.Repeat(" ", leftPad))
		}

		if visualIdx >= 0 && visualIdx < len(rows) {
			row := rows[visualIdx]
			if row.kind == lineRenderRaw {
				b.WriteString(e.renderWrappedSegment(row.lineIndex, row.rawStart, row.rawChunk, cw, flashLine))
			} else {
				b.WriteString(applyBaseStyleAfterReset(row.rendered, base))
				b.WriteString(base)
			}
		} else {
			b.WriteString(strings.Repeat(" ", cw))
		}

		// Always restore base style before outer padding so highlights stay
		// inside the centered writing area.
		b.WriteString(base)
		if rightPad > 0 {
			b.WriteString(strings.Repeat(" ", rightPad))
		}
	}

	if e.height > 1 {
		indicator := "I"
		switch e.mode {
		case modeNormal:
			indicator = "N"
		case modeVisual:
			indicator = "V"
		case modeVisualLine:
			indicator = "L"
		}
		counter := fmt.Sprintf("%dw", e.wordCount())
		counterRunes := []rune(counter)
		maxCounter := max(0, e.width-1)
		if len(counterRunes) > maxCounter {
			counterRunes = counterRunes[len(counterRunes)-maxCounter:]
		}
		counterStart := e.width - len(counterRunes)
		if counterStart < 1 {
			counterStart = 1
		}

		b.WriteString(fmt.Sprintf("\x1b[%d;1H", e.height))
		b.WriteString(e.style.bgDark)
		b.WriteString(e.style.fgModeIndicator)
		b.WriteString(indicator)
		if e.width > 1 {
			if counterStart > 1 {
				b.WriteString(e.style.fgText)
				b.WriteString(strings.Repeat(" ", counterStart-1))
			}
			if len(counterRunes) > 0 {
				b.WriteString(e.style.fgModeIndicator)
				b.WriteString(string(counterRunes))
			}
		}

		if e.commandLineActive {
			cmdWidth := e.commandLineWidth()
			cmdStart := e.commandLineStartCol(cmdWidth)
			display, cursorOffset := e.commandLineDisplay(cmdWidth)
			if e.commandError != "" {
				display = e.commandLineErrorDisplay(cmdWidth)
				cursorOffset = 2
			}
			b.WriteString(fmt.Sprintf("\x1b[%d;%dH", e.height, cmdStart))
			b.WriteString(e.style.bgSelection)
			displayRunes := []rune(display)
			if len(displayRunes) > 0 {
				b.WriteString(e.style.fgCommandPrompt)
				b.WriteRune(displayRunes[0])
			}
			if len(displayRunes) > 1 {
				if e.commandError != "" {
					b.WriteString(e.style.fgCommandError)
				} else {
					b.WriteString(e.style.fgCommandText)
				}
				b.WriteString(string(displayRunes[1:]))
			}
			b.WriteString(e.style.bgDark)
			b.WriteString(e.style.fgText)
			cursorScreenRow = e.height
			cursorScreenCol = min(e.width, max(1, cmdStart+cursorOffset))
		}
	}

	b.WriteString(fmt.Sprintf("\x1b[%d;%dH", cursorScreenRow, cursorScreenCol))
	b.WriteString(ansiShowCursor)

	return b.String()
}
