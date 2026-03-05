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
}

type styleCodes struct {
	bgDark          string
	bgSelection     string
	fgText          string
	fgModeIndicator string
}

var DefaultAppearance = appearanceConfig{
	Background:    rgb{R: 17, G: 18, B: 20},
	Highlight:     rgb{R: 43, G: 45, B: 51},
	Text:          rgb{R: 231, G: 231, B: 232},
	ModeIndicator: rgb{R: 109, G: 109, B: 114},
}

func (c appearanceConfig) toStyleCodes() styleCodes {
	return styleCodes{
		bgDark:          ansiBgRGB(c.Background),
		bgSelection:     ansiBgRGB(c.Highlight),
		fgText:          ansiFgRGB(c.Text),
		fgModeIndicator: ansiFgRGB(c.ModeIndicator),
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
	selected := false
	setSelected := func(v bool) {
		if v {
			b.WriteString(e.style.bgSelection)
			b.WriteString(e.style.fgText)
		} else {
			b.WriteString(e.style.bgDark)
			b.WriteString(e.style.fgText)
		}
	}

	for i := 0; i < width; i++ {
		hasChar := i < len(chunk)
		col := start + i
		shouldSelect := e.isSelectedCell(lineIdx, col, hasChar) || lineIdx == flashLine
		if shouldSelect != selected {
			setSelected(shouldSelect)
			selected = shouldSelect
		}
		if hasChar {
			b.WriteRune(chunk[i])
		} else {
			b.WriteByte(' ')
		}
	}

	if selected {
		setSelected(false)
	}

	return b.String()
}

func (e *editor) renderFrame() string {
	cw := e.contentWidth()
	body := e.bodyHeight()
	leftPad := max(0, (e.width-cw)/2)
	rightPad := max(0, e.width-leftPad-cw)
	flashLine := e.activeFlashLine(time.Now())

	wrapped := e.wrappedLines()
	cursorVisualRow, cursorVisualCol := e.cursorVisual()
	e.ensureCursorVisible()

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
		b.WriteString(fmt.Sprintf("\x1b[%d;1H", y+1))
		b.WriteString(e.style.bgDark)
		b.WriteString(e.style.fgText)
		if leftPad > 0 {
			b.WriteString(strings.Repeat(" ", leftPad))
		}

		if visualIdx >= 0 && visualIdx < len(wrapped) {
			seg := wrapped[visualIdx]
			chunk := e.lines[seg.lineIndex][seg.start:seg.end]
			if len(chunk) > cw {
				chunk = chunk[:cw]
			}
			b.WriteString(e.renderWrappedSegment(seg.lineIndex, seg.start, chunk, cw, flashLine))
		} else {
			b.WriteString(strings.Repeat(" ", cw))
		}

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
		b.WriteString(fmt.Sprintf("\x1b[%d;1H", e.height))
		b.WriteString(e.style.bgDark)
		b.WriteString(e.style.fgModeIndicator)
		b.WriteString(indicator)
		if e.width > 1 {
			b.WriteString(e.style.fgText)
			b.WriteString(strings.Repeat(" ", e.width-1))
		}
	}

	b.WriteString(fmt.Sprintf("\x1b[%d;%dH", cursorScreenRow, cursorScreenCol))
	b.WriteString(ansiShowCursor)

	return b.String()
}
