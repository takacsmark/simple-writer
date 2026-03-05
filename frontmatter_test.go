package main

import "testing"

func TestFrontMatterBlockAtYAML(t *testing.T) {
	e := testEditorWithLines(
		"---",
		"title: Hello",
		"draft: true",
		"---",
		"# Body",
	)

	start, end, kind, ok := e.frontMatterBlockAt(0)
	if !ok {
		t.Fatal("expected frontmatter block at top")
	}
	if start != 0 || end != 3 {
		t.Fatalf("got range [%d,%d], want [0,3]", start, end)
	}
	if kind != frontMatterYAML {
		t.Fatalf("got kind %v, want YAML", kind)
	}
}

func TestFrontMatterBlockAtTOML(t *testing.T) {
	e := testEditorWithLines(
		"+++",
		"title = \"Hello\"",
		"+++",
	)

	_, _, kind, ok := e.frontMatterBlockAt(0)
	if !ok {
		t.Fatal("expected frontmatter block at top")
	}
	if kind != frontMatterTOML {
		t.Fatalf("got kind %v, want TOML", kind)
	}
}

func TestFrontMatterBlockAtRequiresTopLineAndClosing(t *testing.T) {
	e := testEditorWithLines(
		"# Heading",
		"---",
		"title: x",
		"---",
	)
	if _, _, _, ok := e.frontMatterBlockAt(1); ok {
		t.Fatal("did not expect frontmatter away from top of file")
	}

	e = testEditorWithLines(
		"---",
		"title: x",
	)
	if _, _, _, ok := e.frontMatterBlockAt(0); ok {
		t.Fatal("did not expect unclosed frontmatter block")
	}
}

func TestFrontMatterKindForLine(t *testing.T) {
	e := testEditorWithLines(
		"+++",
		`title = "Hello"`,
		"+++",
		"# Body",
	)

	if kind, ok := e.frontMatterKindForLine(1); !ok || kind != frontMatterTOML {
		t.Fatalf("line 1: got (kind=%v, ok=%v), want TOML true", kind, ok)
	}
	if _, ok := e.frontMatterKindForLine(3); ok {
		t.Fatalf("line 3 should not be treated as frontmatter")
	}
}
