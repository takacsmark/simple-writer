package editor

import "testing"

func TestInputParserParsesScrollControlKeys(t *testing.T) {
	p := &inputParser{}
	keys := p.feed([]byte{2, 5, 6, 25})
	if len(keys) != 4 {
		t.Fatalf("expected 4 keys, got %d", len(keys))
	}
	if keys[0].t != keyCtrlB {
		t.Fatalf("expected Ctrl-B, got %v", keys[0].t)
	}
	if keys[1].t != keyCtrlE {
		t.Fatalf("expected Ctrl-E, got %v", keys[1].t)
	}
	if keys[2].t != keyCtrlF {
		t.Fatalf("expected Ctrl-F, got %v", keys[2].t)
	}
	if keys[3].t != keyCtrlY {
		t.Fatalf("expected Ctrl-Y, got %v", keys[3].t)
	}
}

func TestInputParserParsesPageKeys(t *testing.T) {
	p := &inputParser{}
	keys := p.feed([]byte{27, '[', '5', '~', 27, '[', '6', '~'})
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].t != keyPageUp {
		t.Fatalf("expected PageUp, got %v", keys[0].t)
	}
	if keys[1].t != keyPageDown {
		t.Fatalf("expected PageDown, got %v", keys[1].t)
	}
}

func TestNormalCtrlEScrollsDownOneLine(t *testing.T) {
	e := testEditorWithLines(
		"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10",
		"l11", "l12", "l13", "l14", "l15", "l16", "l17", "l18", "l19", "l20",
	)
	e.width = 80
	e.height = 10
	e.row = 4
	e.col = 0
	e.scroll = 0

	_ = e.handleKey(key{t: keyCtrlE})
	if e.scroll != 1 {
		t.Fatalf("expected scroll=1 after Ctrl-E, got %d", e.scroll)
	}
}

func TestNormalCtrlYScrollsUpOneLine(t *testing.T) {
	e := testEditorWithLines(
		"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10",
		"l11", "l12", "l13", "l14", "l15", "l16", "l17", "l18", "l19", "l20",
	)
	e.width = 80
	e.height = 10
	e.row = 6
	e.col = 0
	e.scroll = 3

	_ = e.handleKey(key{t: keyCtrlY})
	if e.scroll != 2 {
		t.Fatalf("expected scroll=2 after Ctrl-Y, got %d", e.scroll)
	}
}

func TestNormalCtrlFPageDown(t *testing.T) {
	e := testEditorWithLines(
		"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10",
		"l11", "l12", "l13", "l14", "l15", "l16", "l17", "l18", "l19", "l20",
		"l21", "l22", "l23", "l24", "l25", "l26", "l27", "l28", "l29", "l30",
	)
	e.width = 80
	e.height = 10
	e.row = 4
	e.col = 0
	e.scroll = 0

	_ = e.handleKey(key{t: keyCtrlF})
	if e.scroll != 8 {
		t.Fatalf("expected scroll=8 after Ctrl-F, got %d", e.scroll)
	}
}

func TestNormalCtrlBPageUp(t *testing.T) {
	e := testEditorWithLines(
		"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10",
		"l11", "l12", "l13", "l14", "l15", "l16", "l17", "l18", "l19", "l20",
		"l21", "l22", "l23", "l24", "l25", "l26", "l27", "l28", "l29", "l30",
	)
	e.width = 80
	e.height = 10
	e.row = 15
	e.col = 0
	e.scroll = 10

	_ = e.handleKey(key{t: keyCtrlB})
	if e.scroll != 2 {
		t.Fatalf("expected scroll=2 after Ctrl-B, got %d", e.scroll)
	}
}

func TestNormalCtrlFCanReachEndOnSingleWrappedLine(t *testing.T) {
	long := make([]rune, writingWidth*35)
	for i := range long {
		long[i] = 'a'
	}
	e := testEditorWithLines(string(long))
	e.width = 80
	e.height = 10
	e.row = 0
	e.col = 0
	e.scroll = 0

	maxScroll := e.maxScrollForRows(e.visualRows())
	if maxScroll == 0 {
		t.Fatalf("expected wrapped content to be scrollable")
	}

	for i := 0; i < 20; i++ {
		_ = e.handleKey(key{t: keyCtrlF})
		e.ensureCursorVisible()
	}

	if e.scroll != maxScroll {
		t.Fatalf("expected to reach end scroll=%d, got %d", maxScroll, e.scroll)
	}
}

func TestPageDownKeyCanReachEndOnSingleWrappedLine(t *testing.T) {
	long := make([]rune, writingWidth*35)
	for i := range long {
		long[i] = 'a'
	}
	e := testEditorWithLines(string(long))
	e.width = 80
	e.height = 10
	e.row = 0
	e.col = 0
	e.scroll = 0

	maxScroll := e.maxScrollForRows(e.visualRows())
	if maxScroll == 0 {
		t.Fatalf("expected wrapped content to be scrollable")
	}

	for i := 0; i < 20; i++ {
		_ = e.handleKey(key{t: keyPageDown})
		e.ensureCursorVisible()
	}

	if e.scroll != maxScroll {
		t.Fatalf("expected to reach end scroll=%d, got %d", maxScroll, e.scroll)
	}
}
