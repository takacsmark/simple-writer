package main

import "strings"

const quitDirtyError = "no write since last change (add ! to override)"

func (e *editor) openCommandLine() {
	e.commandLineActive = true
	e.commandLine = e.commandLine[:0]
	e.commandCol = 0
	e.commandError = ""
}

func (e *editor) closeCommandLine() {
	e.commandLineActive = false
	e.commandLine = e.commandLine[:0]
	e.commandCol = 0
	e.commandError = ""
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

func (e *editor) commandLineErrorDisplay(fieldWidth int) string {
	if fieldWidth <= 0 {
		return ""
	}
	prefix := "> "
	if fieldWidth <= len(prefix) {
		return prefix[:fieldWidth]
	}
	available := fieldWidth - len(prefix)
	msg := []rune(e.commandError)
	if len(msg) > available {
		msg = msg[:available]
	}
	for len(msg) < available {
		msg = append(msg, ' ')
	}
	return prefix + string(msg)
}

func (e *editor) executeCommandLine() (quit bool, closePrompt bool) {
	cmd := strings.TrimSpace(string(e.commandLine))
	switch cmd {
	case "":
		return false, true
	case "q", "quit", "qa":
		if e.dirty {
			e.commandError = quitDirtyError
			return false, false
		}
		return true, true
	case "q!", "qa!":
		return true, true
	case "w":
		if err := e.saveBuffer(""); err != nil {
			e.commandError = err.Error()
			return false, false
		}
		return false, true
	default:
		if strings.HasPrefix(cmd, "w ") {
			path := strings.TrimSpace(cmd[1:])
			if err := e.saveBuffer(path); err != nil {
				e.commandError = err.Error()
				return false, false
			}
			return false, true
		}
		return false, true
	}
}

func (e *editor) handleCommandLine(k key) bool {
	if e.commandError != "" {
		e.commandError = ""
	}

	switch k.t {
	case keyCtrlC:
		return true
	case keyEscape:
		e.closeCommandLine()
	case keyEnter:
		quit, closePrompt := e.executeCommandLine()
		if closePrompt {
			e.closeCommandLine()
		}
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
