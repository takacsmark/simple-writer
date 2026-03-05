package main

import "testing"

func TestCodeBlockStartAtClosedFence(t *testing.T) {
	e := testEditorWithLines(
		"intro",
		"```go",
		"fmt.Println(\"x\")",
		"```",
		"outro",
	)

	start, end, block, ok := e.codeBlockStartAt(1)
	if !ok {
		t.Fatal("expected fenced code block at line 1")
	}
	if start != 1 || end != 3 {
		t.Fatalf("got range [%d,%d], want [1,3]", start, end)
	}
	if block.lang != "go" {
		t.Fatalf("got lang %q, want %q", block.lang, "go")
	}
}

func TestCodeBlockStartAtUnclosedFenceRunsToEOF(t *testing.T) {
	e := testEditorWithLines(
		"```python",
		"print('x')",
		"print('y')",
	)

	start, end, block, ok := e.codeBlockStartAt(0)
	if !ok {
		t.Fatal("expected fenced code block at line 0")
	}
	if start != 0 || end != 2 {
		t.Fatalf("got range [%d,%d], want [0,2]", start, end)
	}
	if block.lang != "python" {
		t.Fatalf("got lang %q, want %q", block.lang, "python")
	}
}

func TestCodeBlockStartAtRequiresFencePrefix(t *testing.T) {
	e := testEditorWithLines("`` not-a-fence", "plain")
	if _, _, _, ok := e.codeBlockStartAt(0); ok {
		t.Fatal("did not expect code block for short fence")
	}
}
