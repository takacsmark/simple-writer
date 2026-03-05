package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/term"
)

const writingWidth = 72
const maxHistory = 2000
const yankFlashDuration = 180 * time.Millisecond

const (
	ansiHome         = "\x1b[H"
	ansiClear        = "\x1b[2J"
	ansiAltScreenOn  = "\x1b[?1049h"
	ansiAltScreenOff = "\x1b[?1049l"
	ansiHideCursor   = "\x1b[?25l"
	ansiShowCursor   = "\x1b[?25h"
	ansiSteadyBlock  = "\x1b[2 q"
	ansiSteadyBar    = "\x1b[6 q"
	ansiReset        = "\x1b[0m"
)

type mode int

const (
	modeNormal mode = iota
	modeInsert
	modeVisual
	modeVisualLine
)

type keyType int

const (
	keyUnknown keyType = iota
	keyRune
	keyEnter
	keyBackspace
	keyEscape
	keyCtrlC
	keyCtrlR
	keyUp
	keyDown
	keyLeft
	keyRight
)

type key struct {
	t keyType
	r rune
}

type inputParser struct {
	buf []byte
}

type editorState struct {
	lines [][]rune
	row   int
	col   int
}

type editor struct {
	lines             [][]rune
	row               int
	col               int
	scroll            int
	mode              mode
	width             int
	height            int
	normalPending     rune
	undo              []editorState
	redo              []editorState
	visualRow         int
	visualCol         int
	clipText          string
	clipLinewise      bool
	flashLine         int
	flashUntil        time.Time
	style             styleCodes
	markdown          *markdownRenderer
	commandLineActive bool
	commandLine       []rune
	commandCol        int
}

func newEditor() *editor {
	return &editor{
		lines:     [][]rune{{}},
		mode:      modeNormal,
		flashLine: -1,
		style:     DefaultAppearance.toStyleCodes(),
		markdown:  newMarkdownRenderer(),
	}
}

func cloneLines(src [][]rune) [][]rune {
	out := make([][]rune, len(src))
	for i := range src {
		out[i] = append([]rune(nil), src[i]...)
	}
	return out
}

func (e *editor) snapshot() editorState {
	return editorState{
		lines: cloneLines(e.lines),
		row:   e.row,
		col:   e.col,
	}
}

func (e *editor) restore(s editorState) {
	e.lines = cloneLines(s.lines)
	e.row = min(max(0, s.row), len(e.lines)-1)
	if len(e.lines[e.row]) == 0 {
		e.col = 0
	} else {
		e.col = min(max(0, s.col), len(e.lines[e.row])-1)
	}
	e.setMode(modeNormal)
}

func (e *editor) saveUndo() {
	e.undo = append(e.undo, e.snapshot())
	if len(e.undo) > maxHistory {
		e.undo = e.undo[1:]
	}
	e.redo = nil
}

func (e *editor) undoAction() {
	if len(e.undo) == 0 {
		return
	}
	e.redo = append(e.redo, e.snapshot())
	state := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.restore(state)
}

func (e *editor) redoAction() {
	if len(e.redo) == 0 {
		return
	}
	e.undo = append(e.undo, e.snapshot())
	if len(e.undo) > maxHistory {
		e.undo = e.undo[1:]
	}
	state := e.redo[len(e.redo)-1]
	e.redo = e.redo[:len(e.redo)-1]
	e.restore(state)
}

func (p *inputParser) feed(data []byte) []key {
	p.buf = append(p.buf, data...)
	keys := make([]key, 0, len(data))

	for len(p.buf) > 0 {
		b := p.buf[0]
		switch {
		case b == 3:
			keys = append(keys, key{t: keyCtrlC})
			p.buf = p.buf[1:]
		case b == 18:
			keys = append(keys, key{t: keyCtrlR})
			p.buf = p.buf[1:]
		case b == 13 || b == 10:
			keys = append(keys, key{t: keyEnter})
			p.buf = p.buf[1:]
		case b == 127 || b == 8:
			keys = append(keys, key{t: keyBackspace})
			p.buf = p.buf[1:]
		case b == 27:
			if len(p.buf) >= 3 && p.buf[1] == '[' {
				switch p.buf[2] {
				case 'A':
					keys = append(keys, key{t: keyUp})
				case 'B':
					keys = append(keys, key{t: keyDown})
				case 'C':
					keys = append(keys, key{t: keyRight})
				case 'D':
					keys = append(keys, key{t: keyLeft})
				default:
					keys = append(keys, key{t: keyEscape})
				}
				p.buf = p.buf[3:]
				continue
			}
			keys = append(keys, key{t: keyEscape})
			p.buf = p.buf[1:]
		case b < 32:
			p.buf = p.buf[1:]
		case b < utf8.RuneSelf:
			keys = append(keys, key{t: keyRune, r: rune(b)})
			p.buf = p.buf[1:]
		default:
			if !utf8.FullRune(p.buf) {
				return keys
			}
			r, size := utf8.DecodeRune(p.buf)
			if r == utf8.RuneError && size == 1 {
				p.buf = p.buf[1:]
				continue
			}
			keys = append(keys, key{t: keyRune, r: r})
			p.buf = p.buf[size:]
		}
	}

	return keys
}

func (e *editor) updateSize(fd int) {
	w, h, err := term.GetSize(fd)
	if err != nil {
		return
	}
	e.width = max(1, w)
	e.height = max(1, h)
}

func (e *editor) contentWidth() int {
	return max(1, min(writingWidth, e.width))
}

func (e *editor) bodyHeight() int {
	if e.height <= 1 {
		return 1
	}
	return e.height - 1
}

func (e *editor) setMode(m mode) {
	e.mode = m
	e.normalPending = 0
	line := e.lines[e.row]
	if m == modeInsert {
		if e.col < 0 {
			e.col = 0
		}
		if e.col > len(line) {
			e.col = len(line)
		}
		return
	}

	if len(line) == 0 {
		e.col = 0
		return
	}
	if e.col >= len(line) {
		e.col = len(line) - 1
	}
	if e.col < 0 {
		e.col = 0
	}
}

func (e *editor) enterVisual(linewise bool) {
	e.visualRow = e.row
	e.visualCol = e.col
	if linewise {
		e.setMode(modeVisualLine)
		return
	}
	e.setMode(modeVisual)
}

func (e *editor) visualRangeRows() (int, int) {
	if e.visualRow <= e.row {
		return e.visualRow, e.row
	}
	return e.row, e.visualRow
}

func normalizeClipboardText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func (e *editor) setClipboard(text string, linewise bool) {
	text = normalizeClipboardText(text)
	e.clipText = text
	e.clipLinewise = linewise
	_ = writeSystemClipboard(text)
}

func (e *editor) clipboardPayload() (string, bool) {
	text := e.clipText
	linewise := e.clipLinewise
	if sysText, err := readSystemClipboard(); err == nil {
		sysText = normalizeClipboardText(sysText)
		if sysText != "" {
			text = sysText
			linewise = strings.HasSuffix(sysText, "\n")
		}
	}
	return text, linewise
}

func (e *editor) insertRune(r rune) {
	line := e.lines[e.row]
	if e.col < 0 {
		e.col = 0
	}
	if e.col > len(line) {
		e.col = len(line)
	}

	e.saveUndo()
	line = append(line, 0)
	copy(line[e.col+1:], line[e.col:])
	line[e.col] = r
	e.lines[e.row] = line
	e.col++
}

func (e *editor) insertNewline() {
	line := e.lines[e.row]
	if e.col < 0 {
		e.col = 0
	}
	if e.col > len(line) {
		e.col = len(line)
	}

	e.saveUndo()
	left := append([]rune(nil), line[:e.col]...)
	right := append([]rune(nil), line[e.col:]...)
	e.lines[e.row] = left

	e.lines = append(e.lines, nil)
	copy(e.lines[e.row+2:], e.lines[e.row+1:])
	e.lines[e.row+1] = right

	e.row++
	e.col = 0
}

func (e *editor) backspace() {
	if e.col > 0 {
		e.saveUndo()
		line := e.lines[e.row]
		line = append(line[:e.col-1], line[e.col:]...)
		e.lines[e.row] = line
		e.col--
		return
	}

	if e.row == 0 {
		return
	}

	e.saveUndo()
	prevLen := len(e.lines[e.row-1])
	e.lines[e.row-1] = append(e.lines[e.row-1], e.lines[e.row]...)
	e.lines = append(e.lines[:e.row], e.lines[e.row+1:]...)
	e.row--
	e.col = prevLen
}

func (e *editor) deleteAtCursor() {
	line := e.lines[e.row]
	if len(line) == 0 {
		return
	}
	if e.col < 0 {
		e.col = 0
	}
	if e.col >= len(line) {
		e.col = len(line) - 1
	}
	e.saveUndo()
	line = append(line[:e.col], line[e.col+1:]...)
	e.lines[e.row] = line

	if len(line) == 0 {
		e.col = 0
	} else if e.col >= len(line) {
		e.col = len(line) - 1
	}
}

func (e *editor) replaceAtCursor(r rune) {
	line := e.lines[e.row]
	if len(line) == 0 {
		return
	}
	if e.col < 0 {
		e.col = 0
	}
	if e.col >= len(line) {
		e.col = len(line) - 1
	}
	e.saveUndo()
	line[e.col] = r
	e.lines[e.row] = line
}

func (e *editor) deleteToEndOfLine(andInsert bool) {
	line := e.lines[e.row]
	if len(line) > 0 {
		if e.col < 0 {
			e.col = 0
		}
		if e.col >= len(line) {
			e.col = len(line) - 1
		}
		e.saveUndo()
		e.lines[e.row] = append([]rune(nil), line[:e.col]...)
	}

	if andInsert {
		e.setMode(modeInsert)
		return
	}
	e.setMode(modeNormal)
}

func (e *editor) yankVisualSelection() {
	if e.mode == modeVisualLine {
		sr, er := e.visualRangeRows()
		var b strings.Builder
		for i := sr; i <= er; i++ {
			b.WriteString(string(e.lines[i]))
			b.WriteByte('\n')
		}
		e.setClipboard(b.String(), true)
		e.setMode(modeNormal)
		return
	}

	text := e.charwiseSelectionText()
	e.setClipboard(text, false)
	e.setMode(modeNormal)
}

func (e *editor) yankCurrentLine() {
	e.setClipboard(string(e.lines[e.row])+"\n", true)
	e.flashLine = e.row
	e.flashUntil = time.Now().Add(yankFlashDuration)
}

func (e *editor) activeFlashLine(now time.Time) int {
	if e.flashLine >= 0 && now.Before(e.flashUntil) {
		return e.flashLine
	}
	e.flashLine = -1
	return -1
}

func (e *editor) nextFlashTimer(now time.Time) <-chan time.Time {
	if e.flashLine < 0 {
		return nil
	}
	if !now.Before(e.flashUntil) {
		e.flashLine = -1
		return nil
	}
	return time.After(e.flashUntil.Sub(now))
}

func (e *editor) deleteVisualSelection() {
	if e.mode == modeVisualLine {
		sr, er := e.visualRangeRows()
		e.saveUndo()
		if sr == 0 && er == len(e.lines)-1 {
			e.lines = [][]rune{{}}
			e.row = 0
			e.col = 0
			e.setMode(modeNormal)
			return
		}
		e.lines = append(e.lines[:sr], e.lines[er+1:]...)
		if sr >= len(e.lines) {
			e.row = len(e.lines) - 1
		} else {
			e.row = sr
		}
		e.col = 0
		e.setMode(modeNormal)
		return
	}

	sr, sc, er, ec := e.selectionEndpoints()
	e.saveUndo()

	if sr == er {
		line := e.lines[sr]
		if len(line) > 0 {
			sc = min(max(0, sc), len(line)-1)
			ec = min(max(0, ec), len(line)-1)
			if sc > ec {
				sc, ec = ec, sc
			}
			newLine := append([]rune(nil), line[:sc]...)
			newLine = append(newLine, line[ec+1:]...)
			e.lines[sr] = newLine
			e.row = sr
			if len(newLine) == 0 {
				e.col = 0
			} else {
				e.col = min(sc, len(newLine)-1)
			}
		}
		e.setMode(modeNormal)
		return
	}

	first := e.lines[sr]
	last := e.lines[er]
	prefix := append([]rune(nil), first[:min(max(0, sc), len(first))]...)
	suffix := []rune{}
	if len(last) > 0 {
		cut := min(max(0, ec+1), len(last))
		suffix = append(suffix, last[cut:]...)
	}
	merged := append(prefix, suffix...)

	e.lines[sr] = merged
	e.lines = append(e.lines[:sr+1], e.lines[er+1:]...)
	e.row = sr
	if len(merged) == 0 {
		e.col = 0
	} else {
		e.col = min(sc, len(merged)-1)
	}
	e.setMode(modeNormal)
}

func (e *editor) charwiseSelectionText() string {
	sr, sc, er, ec := e.selectionEndpoints()
	if sr == er {
		line := e.lines[sr]
		if len(line) == 0 {
			return ""
		}
		sc = min(max(0, sc), len(line)-1)
		ec = min(max(0, ec), len(line)-1)
		if sc > ec {
			sc, ec = ec, sc
		}
		return string(line[sc : ec+1])
	}

	var b strings.Builder
	first := e.lines[sr]
	if len(first) > 0 {
		sc = min(max(0, sc), len(first)-1)
		b.WriteString(string(first[sc:]))
	}
	b.WriteByte('\n')

	for i := sr + 1; i < er; i++ {
		b.WriteString(string(e.lines[i]))
		b.WriteByte('\n')
	}

	last := e.lines[er]
	if len(last) > 0 {
		ec = min(max(0, ec), len(last)-1)
		b.WriteString(string(last[:ec+1]))
	}

	return b.String()
}

func (e *editor) selectionEndpoints() (int, int, int, int) {
	ar, ac := e.visualRow, e.visualCol
	br, bc := e.row, e.col
	if ar > br || (ar == br && ac > bc) {
		return br, bc, ar, ac
	}
	return ar, ac, br, bc
}

func (e *editor) pasteAfterCursor() {
	text, linewise := e.clipboardPayload()
	if text == "" {
		return
	}

	if linewise {
		e.pasteLinewise(text)
		return
	}
	e.pasteCharwise(text)
}

func (e *editor) pasteLinewise(text string) {
	text = normalizeClipboardText(text)
	if strings.HasSuffix(text, "\n") {
		text = strings.TrimSuffix(text, "\n")
	}
	parts := strings.Split(text, "\n")
	if len(parts) == 0 {
		return
	}

	insertAt := e.row + 1
	newLines := make([][]rune, len(parts))
	for i := range parts {
		newLines[i] = []rune(parts[i])
	}

	e.saveUndo()
	merged := make([][]rune, 0, len(e.lines)+len(newLines))
	merged = append(merged, e.lines[:insertAt]...)
	merged = append(merged, newLines...)
	merged = append(merged, e.lines[insertAt:]...)
	e.lines = merged
	e.row = insertAt
	e.col = 0
}

func (e *editor) pasteCharwise(text string) {
	text = normalizeClipboardText(text)
	line := e.lines[e.row]
	insertCol := 0
	if len(line) > 0 {
		insertCol = min(e.col+1, len(line))
	}

	e.saveUndo()
	endRow, endCol := e.insertTextAt(e.row, insertCol, text)
	e.row = endRow
	e.col = endCol
}

func (e *editor) insertTextAt(row, col int, text string) (int, int) {
	line := e.lines[row]
	col = min(max(0, col), len(line))
	parts := strings.Split(text, "\n")
	if len(parts) == 1 {
		insert := []rune(parts[0])
		newLine := make([]rune, 0, len(line)+len(insert))
		newLine = append(newLine, line[:col]...)
		newLine = append(newLine, insert...)
		newLine = append(newLine, line[col:]...)
		e.lines[row] = newLine
		if len(insert) == 0 {
			if col > 0 {
				return row, col - 1
			}
			return row, 0
		}
		return row, col + len(insert) - 1
	}

	before := append([]rune(nil), line[:col]...)
	after := append([]rune(nil), line[col:]...)

	newLines := make([][]rune, 0, len(parts))
	first := append(before, []rune(parts[0])...)
	newLines = append(newLines, first)
	for i := 1; i < len(parts)-1; i++ {
		newLines = append(newLines, []rune(parts[i]))
	}
	last := append([]rune(parts[len(parts)-1]), after...)
	newLines = append(newLines, last)

	merged := make([][]rune, 0, len(e.lines)-1+len(newLines))
	merged = append(merged, e.lines[:row]...)
	merged = append(merged, newLines...)
	merged = append(merged, e.lines[row+1:]...)
	e.lines = merged

	endRow := row + len(parts) - 1
	lastInserted := []rune(parts[len(parts)-1])
	if len(lastInserted) == 0 {
		return endRow, 0
	}
	return endRow, len(lastInserted) - 1
}

func (e *editor) deleteCurrentLine() {
	e.saveUndo()
	if len(e.lines) == 1 {
		e.lines[0] = []rune{}
		e.row = 0
		e.col = 0
		return
	}

	e.lines = append(e.lines[:e.row], e.lines[e.row+1:]...)
	if e.row >= len(e.lines) {
		e.row = len(e.lines) - 1
	}
	e.col = 0
}

func (e *editor) goToFirstLine() {
	e.row = 0
	e.col = 0
}

func (e *editor) goToLastLine() {
	e.row = len(e.lines) - 1
	e.col = 0
}

func (e *editor) openLineBelow() {
	e.saveUndo()
	insertAt := e.row + 1
	e.lines = append(e.lines, nil)
	copy(e.lines[insertAt+1:], e.lines[insertAt:])
	e.lines[insertAt] = []rune{}
	e.row = insertAt
	e.col = 0
	e.setMode(modeInsert)
}

func (e *editor) moveLeftInsert() {
	if e.col > 0 {
		e.col--
		return
	}
	if e.row > 0 {
		e.row--
		e.col = len(e.lines[e.row])
	}
}

func (e *editor) moveRightInsert() {
	lineLen := len(e.lines[e.row])
	if e.col < lineLen {
		e.col++
		return
	}
	if e.row+1 < len(e.lines) {
		e.row++
		e.col = 0
	}
}

func (e *editor) moveUpInsert() {
	if e.row == 0 {
		return
	}
	e.row--
	e.col = min(e.col, len(e.lines[e.row]))
}

func (e *editor) moveDownInsert() {
	if e.row+1 >= len(e.lines) {
		return
	}
	e.row++
	e.col = min(e.col, len(e.lines[e.row]))
}

func (e *editor) moveLeftNormal() {
	if e.col > 0 {
		e.col--
	}
}

func (e *editor) moveRightNormal() {
	lineLen := len(e.lines[e.row])
	maxCol := 0
	if lineLen > 0 {
		maxCol = lineLen - 1
	}
	if e.col < maxCol {
		e.col++
	}
}

func (e *editor) moveWordForward() {
	// Vim-like: move to next word start, but also stop on empty lines at BOL.
	// If there is no next stop, land on the last character of the last
	// reachable word.
	// If the current line ends in a trailing dot after its last word, stop on
	// that dot before advancing to a later line.
	if e.row >= 0 && e.row < len(e.lines) {
		line := e.lines[e.row]
		dotCol := len(line) - 1
		if dotCol > 0 && line[dotCol] == '.' && isWordRune(line[dotCol-1]) && e.col < dotCol {
			hasNextWordStartOnLine := false
			start := max(0, e.col+1)
			for c := start; c < len(line); c++ {
				if !isWordRune(line[c]) {
					continue
				}
				if c == 0 || !isWordRune(line[c-1]) {
					hasNextWordStartOnLine = true
					break
				}
			}
			if !hasNextWordStartOnLine {
				e.col = dotCol
				return
			}
		}
	}

	for r := e.row; r < len(e.lines); r++ {
		line := e.lines[r]
		start := 0
		if r == e.row {
			start = e.col + 1
		}
		if start < 0 {
			start = 0
		}

		if len(line) == 0 {
			// Empty lines are valid motion stops at column 0.
			if r > e.row {
				e.row = r
				e.col = 0
				return
			}
			continue
		}

		for c := start; c < len(line); c++ {
			if !isWordRune(line[c]) {
				continue
			}
			if c == 0 || !isWordRune(line[c-1]) {
				e.row = r
				e.col = c
				return
			}
		}
	}

	lastRow, lastCol := -1, -1
	lastDotRow, lastDotCol := -1, -1
	for r := e.row; r < len(e.lines); r++ {
		line := e.lines[r]
		start := 0
		if r == e.row {
			start = e.col
		}
		for c := start; c < len(line); c++ {
			if isWordRune(line[c]) {
				lastRow, lastCol = r, c
				if c+1 < len(line) && line[c+1] == '.' {
					lastDotRow, lastDotCol = r, c+1
				}
			}
		}
	}
	if lastDotRow >= 0 {
		e.row = lastDotRow
		e.col = lastDotCol
		return
	}
	if lastRow >= 0 {
		e.row = lastRow
		e.col = lastCol
	}
}

func (e *editor) moveWordBackward() {
	line := e.lines[e.row]
	// If we are inside a word (not already at its start), first jump to the
	// beginning of the current word.
	if e.col > 0 && e.col < len(line) && isWordRune(line[e.col]) && isWordRune(line[e.col-1]) {
		c := e.col
		for c > 0 && isWordRune(line[c-1]) {
			c--
		}
		e.col = c
		return
	}

	// Otherwise jump to the previous word start, or empty line at BOL.
	for r := e.row; r >= 0; r-- {
		line = e.lines[r]
		if len(line) == 0 {
			// Empty lines are valid motion stops at column 0.
			if r < e.row {
				e.row = r
				e.col = 0
				return
			}
			continue
		}

		end := len(line) - 1
		if r == e.row {
			end = min(end, e.col-1)
		}
		for c := end; c >= 0; c-- {
			if !isWordRune(line[c]) {
				continue
			}
			if c == 0 || !isWordRune(line[c-1]) {
				e.row = r
				e.col = c
				return
			}
		}
	}
}

func (e *editor) moveUpNormal() {
	if e.row == 0 {
		return
	}
	e.row--
	lineLen := len(e.lines[e.row])
	maxCol := 0
	if lineLen > 0 {
		maxCol = lineLen - 1
	}
	e.col = min(e.col, maxCol)
}

func (e *editor) moveDownNormal() {
	if e.row+1 >= len(e.lines) {
		return
	}
	e.row++
	lineLen := len(e.lines[e.row])
	maxCol := 0
	if lineLen > 0 {
		maxCol = lineLen - 1
	}
	e.col = min(e.col, maxCol)
}

func readSystemClipboard() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return runClipboardRead("pbpaste")
	case "windows":
		return runClipboardRead("powershell", "-NoProfile", "-Command", "Get-Clipboard")
	default:
		if out, err := runClipboardRead("wl-paste", "-n"); err == nil {
			return out, nil
		}
		if out, err := runClipboardRead("xclip", "-selection", "clipboard", "-o"); err == nil {
			return out, nil
		}
		return runClipboardRead("xsel", "--clipboard", "--output")
	}
}

func writeSystemClipboard(text string) error {
	switch runtime.GOOS {
	case "darwin":
		return runClipboardWrite(text, "pbcopy")
	case "windows":
		return runClipboardWrite(text, "powershell", "-NoProfile", "-Command", "Set-Clipboard")
	default:
		if err := runClipboardWrite(text, "wl-copy"); err == nil {
			return nil
		}
		if err := runClipboardWrite(text, "xclip", "-selection", "clipboard"); err == nil {
			return nil
		}
		return runClipboardWrite(text, "xsel", "--clipboard", "--input")
	}
}

func runClipboardRead(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func runClipboardWrite(text string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	ed := newEditor()
	if err := ed.loadFromArgs(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
		os.Exit(1)
	}

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to enter raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(fd, state)

	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()

	cleanup := func() {
		_, _ = stdout.WriteString(ansiReset + ansiSteadyBlock + ansiShowCursor + ansiAltScreenOff)
		_ = stdout.Flush()
	}
	defer cleanup()
	_, _ = stdout.WriteString(ansiAltScreenOn + ansiHideCursor + ansiClear + ansiHome + ed.style.bgDark + ed.style.fgText)
	_ = stdout.Flush()

	ed.updateSize(fd)
	_, _ = stdout.WriteString(ed.renderFrame())
	_ = stdout.Flush()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)
	defer signal.Stop(sig)

	input := make(chan []byte, 8)
	errs := make(chan error, 1)

	go func() {
		buf := make([]byte, 256)
		for {
			n, readErr := os.Stdin.Read(buf)
			if readErr != nil {
				errs <- readErr
				return
			}
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				input <- chunk
			}
		}
	}()

	parser := &inputParser{}
	var flashTimer <-chan time.Time
	for {
		select {
		case <-sig:
			ed.updateSize(fd)
			ed.ensureCursorVisible()
			_, _ = stdout.WriteString(ed.renderFrame())
			_ = stdout.Flush()
			flashTimer = ed.nextFlashTimer(time.Now())
		case chunk := <-input:
			keys := parser.feed(chunk)
			quit := false
			for _, k := range keys {
				if ed.handleKey(k) {
					quit = true
					break
				}
			}
			if quit {
				return
			}
			ed.updateSize(fd)
			ed.ensureCursorVisible()
			_, _ = stdout.WriteString(ed.renderFrame())
			_ = stdout.Flush()
			flashTimer = ed.nextFlashTimer(time.Now())
		case <-flashTimer:
			ed.updateSize(fd)
			ed.ensureCursorVisible()
			_, _ = stdout.WriteString(ed.renderFrame())
			_ = stdout.Flush()
			flashTimer = ed.nextFlashTimer(time.Now())
		case <-errs:
			return
		}
	}
}
