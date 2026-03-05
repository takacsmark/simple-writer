package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromArgsNoFileKeepsDefaultBuffer(t *testing.T) {
	e := newEditor()
	if err := e.loadFromArgs(nil); err != nil {
		t.Fatalf("loadFromArgs(nil) returned error: %v", err)
	}
	if len(e.lines) != 1 || string(e.lines[0]) != "" {
		t.Fatalf("expected default empty buffer, got %#v", e.lines)
	}
}

func TestLoadFromArgsLoadsTxtFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("hello\nworld"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	e := newEditor()
	if err := e.loadFromArgs([]string{path}); err != nil {
		t.Fatalf("loadFromArgs returned error: %v", err)
	}
	if len(e.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(e.lines))
	}
	if got := string(e.lines[0]); got != "hello" {
		t.Fatalf("line 0: got %q, want %q", got, "hello")
	}
	if got := string(e.lines[1]); got != "world" {
		t.Fatalf("line 1: got %q, want %q", got, "world")
	}
	if e.filePath != path {
		t.Fatalf("filePath: got %q, want %q", e.filePath, path)
	}
}

func TestLoadFromArgsLoadsMdFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte("# title\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	e := newEditor()
	if err := e.loadFromArgs([]string{path}); err != nil {
		t.Fatalf("loadFromArgs returned error: %v", err)
	}
	if len(e.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(e.lines))
	}
	if got := string(e.lines[0]); got != "# title" {
		t.Fatalf("line 0: got %q, want %q", got, "# title")
	}
	if got := string(e.lines[1]); got != "" {
		t.Fatalf("line 1: got %q, want empty", got)
	}
}

func TestLoadFromArgsRejectsMultipleFiles(t *testing.T) {
	e := newEditor()
	err := e.loadFromArgs([]string{"a.md", "b.md"})
	if err == nil {
		t.Fatalf("expected error for multiple args")
	}
}

func TestLoadFromArgsRejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	e := newEditor()
	err := e.loadFromArgs([]string{dir})
	if err == nil {
		t.Fatalf("expected error for directory arg")
	}
}

func TestLoadFromArgsRejectsUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.go")
	if err := os.WriteFile(path, []byte("package main"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	e := newEditor()
	err := e.loadFromArgs([]string{path})
	if err == nil {
		t.Fatalf("expected error for unsupported extension")
	}
}
