package editor

func (e *editor) handleInsert(k key) bool {
	switch k.t {
	case keyCtrlC:
		return true
	case keyEscape:
		if e.col > 0 {
			e.col--
		}
		e.setMode(modeNormal)
	case keyLeft:
		e.moveLeftInsert()
	case keyRight:
		e.moveRightInsert()
	case keyUp:
		e.moveUpInsert()
	case keyDown:
		e.moveDownInsert()
	case keyEnter:
		e.insertNewline()
	case keyBackspace:
		e.backspace()
	case keyRune:
		e.insertRune(k.r)
	}
	return false
}

func (e *editor) handleNormal(k key) bool {
	switch k.t {
	case keyCtrlC:
		return true
	case keyCtrlR:
		e.normalPending = 0
		e.redoAction()
	case keyEscape:
		e.normalPending = 0
	case keyLeft:
		e.normalPending = 0
		e.moveLeftNormal()
	case keyRight:
		e.normalPending = 0
		e.moveRightNormal()
	case keyUp:
		e.normalPending = 0
		e.moveUpNormal()
	case keyDown:
		e.normalPending = 0
		e.moveDownNormal()
	case keyRune:
		r := k.r

		if e.normalPending != 0 {
			pending := e.normalPending
			e.normalPending = 0
			switch pending {
			case 'g':
				if r == 'g' {
					e.goToFirstLine()
					return false
				}
			case 'd':
				if r == 'd' {
					e.deleteCurrentLine()
					return false
				}
			case 'y':
				if r == 'y' {
					e.yankCurrentLine()
					return false
				}
			case 'r':
				e.replaceAtCursor(r)
				return false
			}
		}

		switch k.r {
		case 'h':
			e.moveLeftNormal()
		case 'j':
			e.moveDownNormal()
		case 'k':
			e.moveUpNormal()
		case 'l':
			e.moveRightNormal()
		case 'b':
			e.moveWordBackward()
		case 'w':
			e.moveWordForward()
		case '0':
			e.col = 0
		case '$':
			lineLen := len(e.lines[e.row])
			if lineLen == 0 {
				e.col = 0
			} else {
				e.col = lineLen - 1
			}
		case 'x':
			e.deleteAtCursor()
		case 'D':
			e.deleteToEndOfLine(false)
		case 'C':
			e.deleteToEndOfLine(true)
		case 'g':
			e.normalPending = 'g'
		case 'G':
			e.goToLastLine()
		case 'd':
			e.normalPending = 'd'
		case 'y':
			e.normalPending = 'y'
		case 'r':
			e.normalPending = 'r'
		case 'u':
			e.undoAction()
		case 'v':
			e.enterVisual(false)
		case 'V':
			e.enterVisual(true)
		case 'p':
			e.pasteAfterCursor()
		case 'i':
			e.setMode(modeInsert)
		case 'a':
			lineLen := len(e.lines[e.row])
			if lineLen > 0 && e.col < lineLen {
				e.col++
			}
			e.setMode(modeInsert)
		case 'I':
			e.col = 0
			e.setMode(modeInsert)
		case 'A':
			e.col = len(e.lines[e.row])
			e.setMode(modeInsert)
		case 'o':
			e.openLineBelow()
		case ':':
			e.openCommandLine()
		}
	}
	return false
}

func (e *editor) handleVisual(k key) bool {
	switch k.t {
	case keyCtrlC:
		return true
	case keyEscape:
		e.setMode(modeNormal)
	case keyLeft:
		e.normalPending = 0
		e.moveLeftNormal()
	case keyRight:
		e.normalPending = 0
		e.moveRightNormal()
	case keyUp:
		e.normalPending = 0
		e.moveUpNormal()
	case keyDown:
		e.normalPending = 0
		e.moveDownNormal()
	case keyRune:
		r := k.r

		if e.normalPending != 0 {
			pending := e.normalPending
			e.normalPending = 0
			if pending == 'g' && r == 'g' {
				e.goToFirstLine()
				return false
			}
		}

		switch r {
		case 'h':
			e.moveLeftNormal()
		case 'j':
			e.moveDownNormal()
		case 'k':
			e.moveUpNormal()
		case 'l':
			e.moveRightNormal()
		case 'b':
			e.moveWordBackward()
		case 'w':
			e.moveWordForward()
		case '0':
			e.col = 0
		case '$':
			lineLen := len(e.lines[e.row])
			if lineLen == 0 {
				e.col = 0
			} else {
				e.col = lineLen - 1
			}
		case 'g':
			e.normalPending = 'g'
		case 'G':
			e.goToLastLine()
		case 'v':
			if e.mode == modeVisual {
				e.setMode(modeNormal)
			} else {
				e.setMode(modeVisual)
			}
		case 'V':
			if e.mode == modeVisualLine {
				e.setMode(modeNormal)
			} else {
				e.setMode(modeVisualLine)
			}
		case 'y':
			e.yankVisualSelection()
		case 'x':
			e.deleteVisualSelection()
		}
	}
	return false
}

func (e *editor) handleKey(k key) bool {
	if e.commandLineActive {
		return e.handleCommandLine(k)
	}
	if e.mode == modeInsert {
		return e.handleInsert(k)
	}
	if e.mode == modeVisual || e.mode == modeVisualLine {
		return e.handleVisual(k)
	}
	return e.handleNormal(k)
}
