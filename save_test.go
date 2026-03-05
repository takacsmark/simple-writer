package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveBufferUsesExistingFilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	e := newEditor()
	e.filePath = path
	e.lines = [][]rune{[]rune("hello"), []rune("world")}

	if err := e.saveBuffer(""); err != nil {
		t.Fatalf("saveBuffer returned error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if got := string(data); got != "hello\nworld" {
		t.Fatalf("saved content mismatch: got %q", got)
	}
}

func TestSaveBufferNeedsNameWhenUnnamed(t *testing.T) {
	e := newEditor()
	e.lines = [][]rune{[]rune("hello")}
	if err := e.saveBuffer(""); err == nil {
		t.Fatalf("expected error when saving unnamed buffer with :w")
	}
}

func TestSaveBufferWithProvidedNameSetsBufferPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	e := newEditor()
	e.lines = [][]rune{[]rune("alpha")}

	if err := e.saveBuffer(path); err != nil {
		t.Fatalf("saveBuffer returned error: %v", err)
	}
	if e.filePath != path {
		t.Fatalf("filePath not updated: got %q, want %q", e.filePath, path)
	}
}

func TestSaveBufferRejectsUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.go")
	e := newEditor()
	e.lines = [][]rune{[]rune("alpha")}

	if err := e.saveBuffer(path); err == nil {
		t.Fatalf("expected unsupported extension error")
	}
}
