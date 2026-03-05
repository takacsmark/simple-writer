package main

import (
	"strings"
	"testing"
)

func TestStripInlineLinkDestinations(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain inline link",
			in:   "Read [docs](https://example.com/docs) now",
			want: "Read [docs]() now",
		},
		{
			name: "link with title and parens",
			in:   `[x](https://example.com/path(a) "Title (v2)")`,
			want: `[x]()`,
		},
		{
			name: "multiple links",
			in:   "[a](https://a) and [b](https://b)",
			want: "[a]() and [b]()",
		},
		{
			name: "reference style converted",
			in:   "[a][ref]",
			want: "[a]()",
		},
		{
			name: "collapsed reference converted",
			in:   "[a][]",
			want: "[a]()",
		},
		{
			name: "image unchanged",
			in:   "![alt](https://img.example/x.png)",
			want: "![alt](https://img.example/x.png)",
		},
		{
			name: "escaped opener unchanged",
			in:   `\[a](https://example.com)`,
			want: `\[a](https://example.com)`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := stripInlineLinkDestinations(tc.in)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRenderMarkdownTableBlockLineCount(t *testing.T) {
	m := newMarkdownRenderer()
	lines := [][]rune{
		[]rune("| A | B |"),
		[]rune("| --- | --- |"),
		[]rune("| 1 | 2 |"),
	}

	rows, err := m.renderMarkdownTableBlock(lines, 72)
	if err != nil {
		t.Fatalf("renderMarkdownTableBlock returned error: %v", err)
	}
	if len(rows) != len(lines) {
		t.Fatalf("got %d rows, want %d", len(rows), len(lines))
	}
}

func TestRenderMarkdownCodeLineDiffersFromPlainLine(t *testing.T) {
	m := newMarkdownRenderer()
	line := []rune(`fmt.Println("x")`)

	plainRows, err := m.renderMarkdownLine(line, 72)
	if err != nil {
		t.Fatalf("renderMarkdownLine returned error: %v", err)
	}
	codeRows, err := m.renderMarkdownCodeLine("go", line, 72)
	if err != nil {
		t.Fatalf("renderMarkdownCodeLine returned error: %v", err)
	}
	if len(plainRows) != 1 || len(codeRows) != 1 {
		t.Fatalf("unexpected row counts plain=%d code=%d", len(plainRows), len(codeRows))
	}
	if plainRows[0] == codeRows[0] {
		t.Fatalf("expected code rendering to differ from plain rendering")
	}
	if strings.TrimSpace(codeRows[0]) == "" {
		t.Fatalf("expected non-empty rendered code line")
	}
}
