package main

import (
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
