package editor

import "testing"

func TestWordCount(t *testing.T) {
	e := &editor{
		lines: [][]rune{
			[]rune("hello world"),
			[]rune("two_words and 123"),
			[]rune(""),
			[]rune("punctuation,split!ok"),
		},
	}

	if got, want := e.wordCount(), 8; got != want {
		t.Fatalf("wordCount() = %d, want %d", got, want)
	}
}
