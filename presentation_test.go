package main

import "testing"

func testEditorWithLines(lines ...string) *editor {
	out := make([][]rune, 0, len(lines))
	for _, line := range lines {
		out = append(out, []rune(line))
	}
	if len(out) == 0 {
		out = [][]rune{{}}
	}
	return &editor{
		lines:     out,
		mode:      modeNormal,
		row:       0,
		col:       0,
		flashLine: -1,
	}
}

func TestLineRenderKindInsertAndNormal(t *testing.T) {
	e := testEditorWithLines("# title", "**focus**", "- item")

	e.mode = modeInsert
	e.row = 1
	if got := e.lineRenderKind(0); got != lineRenderMarkdown {
		t.Fatalf("insert mode line 0: got %v, want markdown", got)
	}
	if got := e.lineRenderKind(1); got != lineRenderRaw {
		t.Fatalf("insert mode line 1: got %v, want raw", got)
	}
	if got := e.lineRenderKind(2); got != lineRenderMarkdown {
		t.Fatalf("insert mode line 2: got %v, want markdown", got)
	}

	e.mode = modeNormal
	e.row = 2
	if got := e.lineRenderKind(0); got != lineRenderMarkdown {
		t.Fatalf("normal mode line 0: got %v, want markdown", got)
	}
	if got := e.lineRenderKind(2); got != lineRenderRaw {
		t.Fatalf("normal mode line 2: got %v, want raw", got)
	}
}

func TestLineRenderKindVisualAndVisualLine(t *testing.T) {
	e := testEditorWithLines("a", "b", "c", "d")

	e.mode = modeVisual
	e.visualRow = 3
	e.visualCol = 0
	e.row = 1
	e.col = 0

	for i := 1; i <= 3; i++ {
		if got := e.lineRenderKind(i); got != lineRenderRaw {
			t.Fatalf("visual mode line %d: got %v, want raw", i, got)
		}
	}
	if got := e.lineRenderKind(0); got != lineRenderMarkdown {
		t.Fatalf("visual mode line 0: got %v, want markdown", got)
	}

	e.mode = modeVisualLine
	e.visualRow = 0
	e.row = 2

	for i := 0; i <= 2; i++ {
		if got := e.lineRenderKind(i); got != lineRenderRaw {
			t.Fatalf("visual line mode line %d: got %v, want raw", i, got)
		}
	}
	if got := e.lineRenderKind(3); got != lineRenderMarkdown {
		t.Fatalf("visual line mode line 3: got %v, want markdown", got)
	}
}

func TestLineRenderKindFlashLineStaysRaw(t *testing.T) {
	e := testEditorWithLines("a", "b", "c")
	e.mode = modeNormal
	e.row = 1
	e.flashLine = 0
	if got := e.lineRenderKind(0); got != lineRenderRaw {
		t.Fatalf("flash line render kind: got %v, want raw", got)
	}
}

func TestLineRenderKindNormalTableOnlyCursorLineRaw(t *testing.T) {
	e := testEditorWithLines(
		"| A | B |",
		"| --- | --- |",
		"| 1 | 2 |",
	)
	e.mode = modeNormal
	e.row = 1

	if got := e.lineRenderKind(0); got != lineRenderMarkdown {
		t.Fatalf("table header in normal mode: got %v, want markdown", got)
	}
	if got := e.lineRenderKind(1); got != lineRenderRaw {
		t.Fatalf("cursor table row in normal mode: got %v, want raw", got)
	}
	if got := e.lineRenderKind(2); got != lineRenderMarkdown {
		t.Fatalf("other table row in normal mode: got %v, want markdown", got)
	}
}

func TestLineRenderKindVisualTableWholeBlockRaw(t *testing.T) {
	e := testEditorWithLines(
		"before",
		"| A | B |",
		"| --- | --- |",
		"| 1 | 2 |",
		"after",
	)
	e.mode = modeVisual
	e.visualRow = 3
	e.visualCol = 0
	e.row = 3
	e.col = 0

	for i := 1; i <= 3; i++ {
		if got := e.lineRenderKind(i); got != lineRenderRaw {
			t.Fatalf("table line %d in visual mode: got %v, want raw", i, got)
		}
	}
	if got := e.lineRenderKind(0); got != lineRenderMarkdown {
		t.Fatalf("line before table in visual mode: got %v, want markdown", got)
	}
	if got := e.lineRenderKind(4); got != lineRenderMarkdown {
		t.Fatalf("line after table in visual mode: got %v, want markdown", got)
	}
}

func TestVisualRowsTableCursorLineRawOnly(t *testing.T) {
	e := testEditorWithLines(
		"| A | B |",
		"| --- | --- |",
		"| 1 | 2 |",
	)
	e.mode = modeNormal
	e.row = 1
	e.width = 80
	e.height = 24
	e.markdown = newMarkdownRenderer()

	rows := e.visualRows()
	if len(rows) < 3 {
		t.Fatalf("expected at least 3 visual rows, got %d", len(rows))
	}

	kindByLine := map[int]lineRenderKind{}
	for _, r := range rows {
		if _, seen := kindByLine[r.lineIndex]; !seen {
			kindByLine[r.lineIndex] = r.kind
		}
	}

	if got := kindByLine[0]; got != lineRenderMarkdown {
		t.Fatalf("table header kind: got %v, want markdown", got)
	}
	if got := kindByLine[1]; got != lineRenderRaw {
		t.Fatalf("cursor table line kind: got %v, want raw", got)
	}
	if got := kindByLine[2]; got != lineRenderMarkdown {
		t.Fatalf("table data line kind: got %v, want markdown", got)
	}
}

func TestVisualRowsCodeBlockCursorLineRawOnly(t *testing.T) {
	e := testEditorWithLines(
		"```go",
		`fmt.Println("x")`,
		"```",
	)
	e.mode = modeNormal
	e.row = 1
	e.width = 80
	e.height = 24
	e.markdown = newMarkdownRenderer()

	rows := e.visualRows()
	if len(rows) < 3 {
		t.Fatalf("expected at least 3 visual rows, got %d", len(rows))
	}

	kindByLine := map[int]lineRenderKind{}
	for _, r := range rows {
		if _, seen := kindByLine[r.lineIndex]; !seen {
			kindByLine[r.lineIndex] = r.kind
		}
	}

	if got := kindByLine[0]; got != lineRenderMarkdown {
		t.Fatalf("opening fence kind: got %v, want markdown", got)
	}
	if got := kindByLine[1]; got != lineRenderRaw {
		t.Fatalf("active code line kind: got %v, want raw", got)
	}
	if got := kindByLine[2]; got != lineRenderMarkdown {
		t.Fatalf("closing fence kind: got %v, want markdown", got)
	}
}
