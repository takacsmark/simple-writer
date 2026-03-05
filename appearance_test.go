package main

import (
	"strings"
	"testing"
)

func TestRawHeadingLevel(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{line: "# h1", want: 1},
		{line: "## h2", want: 2},
		{line: "   ### h3", want: 3},
		{line: "####### too many", want: 0},
		{line: "##no-space", want: 0},
		{line: "text # not heading", want: 0},
		{line: "######", want: 6},
	}

	for _, tc := range tests {
		got := rawHeadingLevel([]rune(tc.line))
		if got != tc.want {
			t.Fatalf("line %q: got %d, want %d", tc.line, got, tc.want)
		}
	}
}

func TestRenderWrappedSegmentUsesHeadingColorWithoutSelection(t *testing.T) {
	e := &editor{
		lines:     [][]rune{[]rune("# Heading")},
		mode:      modeNormal,
		row:       0,
		col:       0,
		flashLine: -1,
		style:     DefaultAppearance.toStyleCodes(),
	}

	out := e.renderWrappedSegment(0, 0, []rune("# Heading"), 10, -1)
	fg, ok := headingAnsiFg(1)
	if !ok {
		t.Fatal("missing heading color for h1")
	}
	if !strings.Contains(out, fg) {
		t.Fatalf("expected output to contain heading color %q, got %q", fg, out)
	}
}

func TestRawLinkColorSpans(t *testing.T) {
	line := []rune(`See [label](https://example.com)`)
	spans := rawLinkColorSpans(line)
	if len(spans) != 2 {
		t.Fatalf("got %d spans, want 2", len(spans))
	}
	if spans[0].fg != linkLabelAnsiFg() {
		t.Fatalf("label span fg: got %q, want %q", spans[0].fg, linkLabelAnsiFg())
	}
	if spans[1].fg != linkDestAnsiFg() {
		t.Fatalf("dest span fg: got %q, want %q", spans[1].fg, linkDestAnsiFg())
	}
	if string(line[spans[0].start:spans[0].end]) != "[label]" {
		t.Fatalf("label span text mismatch: got %q", string(line[spans[0].start:spans[0].end]))
	}
	if string(line[spans[1].start:spans[1].end]) != "(https://example.com)" {
		t.Fatalf("dest span text mismatch: got %q", string(line[spans[1].start:spans[1].end]))
	}
}
