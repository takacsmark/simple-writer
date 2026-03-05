package editor

import (
	"path/filepath"
	"testing"
)

func TestDirtyStartsClean(t *testing.T) {
	e := newEditor()
	if e.dirty {
		t.Fatalf("new editor should start clean")
	}
}

func TestDirtyMutationsAndSaveLifecycle(t *testing.T) {
	e := newEditor()
	e.mode = modeInsert
	e.insertRune('a')
	if !e.dirty {
		t.Fatalf("insert should mark dirty")
	}

	path := filepath.Join(t.TempDir(), "note.md")
	if err := e.saveBuffer(path); err != nil {
		t.Fatalf("saveBuffer returned error: %v", err)
	}
	if e.dirty {
		t.Fatalf("successful save should clear dirty")
	}

	e.mode = modeInsert
	e.insertRune('b')
	if !e.dirty {
		t.Fatalf("post-save edit should mark dirty")
	}

	e.undoAction()
	if e.dirty {
		t.Fatalf("undo should restore clean snapshot")
	}

	e.redoAction()
	if !e.dirty {
		t.Fatalf("redo should restore dirty snapshot")
	}
}

func TestDirtyMutationCategories(t *testing.T) {
	tests := []struct {
		name string
		run  func(e *editor)
	}{
		{name: "insert newline", run: func(e *editor) { e.mode = modeInsert; e.insertNewline() }},
		{name: "backspace", run: func(e *editor) { e.lines = [][]rune{[]rune("ab")}; e.row = 0; e.col = 1; e.backspace() }},
		{name: "delete at cursor", run: func(e *editor) { e.lines = [][]rune{[]rune("ab")}; e.row = 0; e.col = 0; e.deleteAtCursor() }},
		{name: "replace", run: func(e *editor) { e.lines = [][]rune{[]rune("ab")}; e.row = 0; e.col = 0; e.replaceAtCursor('z') }},
		{name: "delete to eol", run: func(e *editor) { e.lines = [][]rune{[]rune("abc")}; e.row = 0; e.col = 1; e.deleteToEndOfLine(false) }},
		{name: "paste linewise", run: func(e *editor) { e.lines = [][]rune{[]rune("a")}; e.row = 0; e.col = 0; e.pasteLinewise("x\ny") }},
		{name: "paste charwise", run: func(e *editor) { e.lines = [][]rune{[]rune("ab")}; e.row = 0; e.col = 0; e.pasteCharwise("x") }},
		{name: "open line below", run: func(e *editor) { e.lines = [][]rune{[]rune("ab")}; e.row = 0; e.col = 0; e.openLineBelow() }},
		{name: "delete current line", run: func(e *editor) {
			e.lines = [][]rune{[]rune("ab"), []rune("cd")}
			e.row = 0
			e.col = 0
			e.deleteCurrentLine()
		}},
		{name: "delete visual selection", run: func(e *editor) {
			e.lines = [][]rune{[]rune("ab"), []rune("cd")}
			e.mode = modeVisualLine
			e.visualRow = 0
			e.row = 1
			e.col = 0
			e.deleteVisualSelection()
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEditor()
			e.dirty = false
			tc.run(e)
			if !e.dirty {
				t.Fatalf("%s should mark dirty", tc.name)
			}
		})
	}
}
