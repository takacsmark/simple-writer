package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandLineOpenFromNormalWithColon(t *testing.T) {
	e := testEditorWithLines("hello")
	e.mode = modeNormal

	_ = e.handleKey(key{t: keyRune, r: ':'})
	if !e.commandLineActive {
		t.Fatalf("expected command line to be active")
	}
	if got := string(e.commandLine); got != "" {
		t.Fatalf("expected empty command line on open, got %q", got)
	}
}

func TestCommandLineEscCloses(t *testing.T) {
	e := testEditorWithLines("hello")
	e.mode = modeNormal
	e.openCommandLine()
	e.commandLine = []rune("q")
	e.commandCol = 1

	quit := e.handleKey(key{t: keyEscape})
	if quit {
		t.Fatalf("did not expect quit on escape")
	}
	if e.commandLineActive {
		t.Fatalf("expected command line to close on escape")
	}
}

func TestCommandLineEnterExecutesQuit(t *testing.T) {
	e := testEditorWithLines("hello")
	e.mode = modeNormal
	e.openCommandLine()
	e.commandLine = []rune("q")
	e.commandCol = 1

	quit := e.handleKey(key{t: keyEnter})
	if !quit {
		t.Fatalf("expected :q to request quit")
	}
	if e.commandLineActive {
		t.Fatalf("expected command line to close after enter")
	}
}

func TestCommandLineQuitBlockedWhenDirty(t *testing.T) {
	e := testEditorWithLines("hello")
	e.mode = modeNormal
	e.dirty = true
	e.openCommandLine()
	e.commandLine = []rune("q")
	e.commandCol = 1

	quit := e.handleKey(key{t: keyEnter})
	if quit {
		t.Fatalf("expected :q to be blocked when dirty")
	}
	if !e.commandLineActive {
		t.Fatalf("expected command line to remain open on dirty quit block")
	}
	if e.commandError != quitDirtyError {
		t.Fatalf("unexpected quit error: got %q, want %q", e.commandError, quitDirtyError)
	}
}

func TestCommandLineQuitBangBypassesDirty(t *testing.T) {
	e := testEditorWithLines("hello")
	e.mode = modeNormal
	e.dirty = true
	e.openCommandLine()
	e.commandLine = []rune("q!")
	e.commandCol = 2

	quit := e.handleKey(key{t: keyEnter})
	if !quit {
		t.Fatalf("expected :q! to quit even when dirty")
	}
}

func TestCommandLineDisplayHasPromptAndFixedWidth(t *testing.T) {
	e := testEditorWithLines("hello")
	e.width = 80
	e.height = 24
	e.openCommandLine()
	e.commandLine = []rune("write")
	e.commandCol = len(e.commandLine)

	w := e.commandLineWidth()
	display, cursor := e.commandLineDisplay(w)
	if len([]rune(display)) != w {
		t.Fatalf("display width mismatch: got %d, want %d", len([]rune(display)), w)
	}
	if !strings.HasPrefix(display, "> ") {
		t.Fatalf("display should start with \"> \", got %q", display)
	}
	if cursor < 1 || cursor >= w {
		t.Fatalf("cursor offset out of range: %d for width %d", cursor, w)
	}
}

func TestRenderFrameShowsCommandLineOnStatusRow(t *testing.T) {
	e := &editor{
		lines:     [][]rune{[]rune("hello")},
		mode:      modeNormal,
		row:       0,
		col:       0,
		width:     80,
		height:    2,
		flashLine: -1,
		style:     DefaultAppearance.toStyleCodes(),
		markdown:  newMarkdownRenderer(),
	}
	e.openCommandLine()
	e.commandLine = []rune("q")
	e.commandCol = 1

	frame := e.renderFrame()
	if !strings.Contains(frame, e.style.bgSelection+e.style.fgCommandPrompt+">") {
		t.Fatalf("expected status row to include styled command prompt; frame=%q", frame)
	}
	if !strings.Contains(frame, e.style.fgCommandText+" q") {
		t.Fatalf("expected status row to include styled command text; frame=%q", frame)
	}
}

func TestCommandLineWriteUnnamedBufferShowsErrorAndStaysOpen(t *testing.T) {
	e := testEditorWithLines("hello")
	e.mode = modeNormal
	e.openCommandLine()
	e.commandLine = []rune("w")
	e.commandCol = 1

	quit := e.handleKey(key{t: keyEnter})
	if quit {
		t.Fatalf("did not expect quit")
	}
	if !e.commandLineActive {
		t.Fatalf("expected command line to stay open on write error")
	}
	if e.commandError == "" {
		t.Fatalf("expected command error to be set")
	}
}

func TestCommandLineWriteErrorClearsOnFirstKeypress(t *testing.T) {
	e := testEditorWithLines("hello")
	e.mode = modeNormal
	e.openCommandLine()
	e.commandLine = []rune("w")
	e.commandCol = 1
	_ = e.handleKey(key{t: keyEnter})
	if e.commandError == "" {
		t.Fatalf("expected command error before clearing")
	}

	_ = e.handleKey(key{t: keyRune, r: ' '})
	if e.commandError != "" {
		t.Fatalf("expected command error to clear on first keypress")
	}
}

func TestCommandLineWriteToPathCreatesFileAndClosesPrompt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")

	e := testEditorWithLines("hello", "world")
	e.mode = modeNormal
	e.openCommandLine()
	e.commandLine = []rune("w " + path)
	e.commandCol = len(e.commandLine)

	quit := e.handleKey(key{t: keyEnter})
	if quit {
		t.Fatalf("did not expect quit")
	}
	if e.commandLineActive {
		t.Fatalf("expected command line to close after successful write")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if got := string(data); got != "hello\nworld" {
		t.Fatalf("written content mismatch: got %q", got)
	}
}

func TestRenderFrameShowsCommandLineErrorInRed(t *testing.T) {
	e := &editor{
		lines:     [][]rune{[]rune("hello")},
		mode:      modeNormal,
		row:       0,
		col:       0,
		width:     80,
		height:    2,
		flashLine: -1,
		style:     DefaultAppearance.toStyleCodes(),
		markdown:  newMarkdownRenderer(),
	}
	e.openCommandLine()
	e.commandError = "no file name"

	frame := e.renderFrame()
	if !strings.Contains(frame, e.style.fgCommandError+" no file name") {
		t.Fatalf("expected status row to include red command error text; frame=%q", frame)
	}
}
