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

func TestRawFrontMatterColorSpansYAML(t *testing.T) {
	m := newMarkdownRenderer()
	line := []rune(`title: "Hello" # note`)
	spans, err := m.rawFrontMatterColorSpans(frontMatterYAML, line, 72)
	if err != nil {
		t.Fatalf("rawFrontMatterColorSpans returned error: %v", err)
	}
	if len(spans) == 0 {
		t.Fatalf("expected non-empty frontmatter color spans")
	}
	if !spansContainFG(spans, ansiFg256(140)) {
		t.Fatalf("expected spans to contain key purple fg %q, got %#v", ansiFg256(140), spans)
	}
	if !spansContainFG(spans, ansiFg256(173)) {
		t.Fatalf("expected spans to contain value orange fg %q, got %#v", ansiFg256(173), spans)
	}
}

func TestRenderWrappedSegmentUsesFrontMatterColors(t *testing.T) {
	e := &editor{
		lines: [][]rune{
			[]rune("---"),
			[]rune(`title: "Hello"`),
			[]rune("---"),
		},
		mode:      modeNormal,
		row:       1,
		col:       0,
		flashLine: -1,
		style:     DefaultAppearance.toStyleCodes(),
		markdown:  newMarkdownRenderer(),
	}

	out := e.renderWrappedSegment(1, 0, e.lines[1], 20, -1)
	if !strings.Contains(out, ansiFg256(140)) {
		t.Fatalf("expected output to contain frontmatter key color %q, got %q", ansiFg256(140), out)
	}
	if !strings.Contains(out, ansiFg256(173)) {
		t.Fatalf("expected output to contain frontmatter value color %q, got %q", ansiFg256(173), out)
	}
}

func TestFrontMatterRawPaletteMatchesRenderedTheme(t *testing.T) {
	e := &editor{
		lines: [][]rune{
			[]rune("---"),
			[]rune(`title: "Hello"`),
			[]rune("---"),
		},
		mode:     modeNormal,
		row:      1,
		col:      0,
		style:    DefaultAppearance.toStyleCodes(),
		markdown: newMarkdownRenderer(),
	}

	renderedRows, err := e.markdown.renderMarkdownFrontMatterLine(frontMatterYAML, e.lines[1], 72)
	if err != nil || len(renderedRows) == 0 {
		t.Fatalf("renderMarkdownFrontMatterLine failed: err=%v rows=%d", err, len(renderedRows))
	}
	renderedSpans := colorSpansForSource(renderedRows[0], string(e.lines[1]))
	rawOut := e.renderWrappedSegment(1, 0, e.lines[1], 20, -1)
	for _, span := range renderedSpans {
		if !strings.Contains(rawOut, span.fg) {
			t.Fatalf("raw output missing rendered fg %q; raw=%q rendered=%q", span.fg, rawOut, renderedRows[0])
		}
	}
}

func spansContainFG(spans []linkColorSpan, fg string) bool {
	for _, span := range spans {
		if span.fg == fg {
			return true
		}
	}
	return false
}

func TestRenderFrameKeepsSelectionInsideTextArea(t *testing.T) {
	e := &editor{
		lines:     [][]rune{[]rune("selected")},
		mode:      modeVisualLine,
		row:       0,
		col:       0,
		visualRow: 0,
		visualCol: 0,
		width:     80,
		height:    2,
		flashLine: -1,
		style:     DefaultAppearance.toStyleCodes(),
		markdown:  newMarkdownRenderer(),
	}

	frame := e.renderFrame()
	rowStart := strings.Index(frame, "\x1b[1;1H")
	rowEnd := strings.Index(frame, "\x1b[2;1H")
	if rowStart < 0 || rowEnd <= rowStart {
		t.Fatalf("failed to isolate first row from frame: %q", frame)
	}
	row := frame[rowStart:rowEnd]
	if !strings.Contains(row, e.style.bgSelection) {
		t.Fatalf("expected selected row to contain selection background")
	}

	// width=80 and content width=72 gives a 4-cell right outer padding.
	wantRightPad := e.style.bgDark + e.style.fgText + strings.Repeat(" ", 4)
	if !strings.HasSuffix(row, wantRightPad) {
		t.Fatalf("expected first row to end with dark right padding; row=%q", row)
	}
}

func TestRenderFrameStatusShowsWordCountOnRight(t *testing.T) {
	e := &editor{
		lines:     [][]rune{[]rune("one two three")},
		mode:      modeInsert,
		row:       0,
		col:       0,
		width:     20,
		height:    2,
		flashLine: -1,
		style:     DefaultAppearance.toStyleCodes(),
		markdown:  newMarkdownRenderer(),
	}

	frame := e.renderFrame()
	rowStart := strings.Index(frame, "\x1b[2;1H")
	if rowStart < 0 {
		t.Fatalf("failed to isolate status row from frame: %q", frame)
	}
	status := frame[rowStart:]
	want := e.style.fgModeIndicator + "3w"
	if !strings.Contains(status, want) {
		t.Fatalf("expected status row to contain right-side word counter %q, got %q", want, status)
	}
}
