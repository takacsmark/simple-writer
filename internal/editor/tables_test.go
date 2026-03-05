package editor

import "testing"

func TestTableBlockForLine(t *testing.T) {
	e := testEditorWithLines(
		"intro",
		"| A | B |",
		"| --- | --- |",
		"| 1 | 2 |",
		"| 3 | 4 |",
		"outro",
	)

	start, end, ok := e.tableBlockForLine(2)
	if !ok {
		t.Fatal("expected table block for delimiter line")
	}
	if start != 1 || end != 4 {
		t.Fatalf("got table range [%d,%d], want [1,4]", start, end)
	}

	if _, _, ok := e.tableBlockForLine(0); ok {
		t.Fatal("did not expect table block for non-table line")
	}
}

func TestMarkdownTableDelimiterDetection(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{line: "| --- | --- |", want: true},
		{line: "| :--- | ---: |", want: true},
		{line: " --- | --- ", want: true},
		{line: "| - | --- |", want: false},
		{line: "| --- | text |", want: false},
		{line: "plain text", want: false},
	}

	for _, tc := range cases {
		if got := isMarkdownTableDelimiterLine(tc.line); got != tc.want {
			t.Fatalf("line %q: got %v, want %v", tc.line, got, tc.want)
		}
	}
}
