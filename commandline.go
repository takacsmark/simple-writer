package main

import "strings"

func (e *editor) openCommandLine() {
	e.commandLineActive = true
	e.commandLine = e.commandLine[:0]
	e.commandCol = 0
}

func (e *editor) closeCommandLine() {
	e.commandLineActive = false
	e.commandLine = e.commandLine[:0]
	e.commandCol = 0
}

func (e *editor) commandLineWidth() int {
	cw := e.contentWidth()
	if cw <= 1 {
		return 1
	}
	w := cw / 2
	if w < 12 {
		w = min(cw, 12)
	}
	if w > cw {
		w = cw
	}
	return max(1, w)
}

func (e *editor) commandLineStartCol(fieldWidth int) int {
	cw := e.contentWidth()
	leftPad := max(0, (e.width-cw)/2)
	return leftPad + max(0, (cw-fieldWidth)/2) + 1
}

func (e *editor) commandLineDisplay(fieldWidth int) (string, int) {
	if fieldWidth <= 0 {
		return "", 0
	}
	if fieldWidth == 1 {
		return ">", 0
	}
	prefix := "> "
	if fieldWidth <= len(prefix) {
		return prefix[:fieldWidth], fieldWidth - 1
	}

	editable := fieldWidth - len(prefix)
	if e.commandCol < 0 {
		e.commandCol = 0
	}
	if e.commandCol > len(e.commandLine) {
		e.commandCol = len(e.commandLine)
	}

	start := 0
	if e.commandCol >= editable {
		start = e.commandCol - editable + 1
	}
	maxStart := max(0, len(e.commandLine)-editable)
	if start > maxStart {
		start = maxStart
	}

	end := min(len(e.commandLine), start+editable)
	visible := append([]rune(nil), e.commandLine[start:end]...)
	for len(visible) < editable {
		visible = append(visible, ' ')
	}

	cursorPos := e.commandCol - start
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos >= editable {
		cursorPos = editable - 1
	}

	return prefix + string(visible), len(prefix) + cursorPos
}

func (e *editor) executeCommandLine() bool {
	cmd := strings.TrimSpace(string(e.commandLine))
	switch cmd {
	case "":
		return false
	case "q", "q!", "quit", "qa", "qa!":
		return true
	default:
		return false
	}
}

func (e *editor) handleCommandLine(k key) bool {
	switch k.t {
	case keyCtrlC:
		return true
	case keyEscape:
		e.closeCommandLine()
	case keyEnter:
		quit := e.executeCommandLine()
		e.closeCommandLine()
		return quit
	case keyBackspace:
		if e.commandCol > 0 && len(e.commandLine) > 0 {
			e.commandLine = append(e.commandLine[:e.commandCol-1], e.commandLine[e.commandCol:]...)
			e.commandCol--
		}
	case keyLeft:
		if e.commandCol > 0 {
			e.commandCol--
		}
	case keyRight:
		if e.commandCol < len(e.commandLine) {
			e.commandCol++
		}
	case keyRune:
		r := k.r
		e.commandLine = append(e.commandLine, 0)
		copy(e.commandLine[e.commandCol+1:], e.commandLine[e.commandCol:])
		e.commandLine[e.commandCol] = r
		e.commandCol++
	}
	return false
}
