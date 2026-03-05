package editor

import "strings"

type fencedCodeBlock struct {
	fenceRune rune
	fenceLen  int
	lang      string
}

func (e *editor) codeBlockStartAt(lineIdx int) (int, int, fencedCodeBlock, bool) {
	if lineIdx < 0 || lineIdx >= len(e.lines) {
		return 0, 0, fencedCodeBlock{}, false
	}

	block, ok := parseCodeFenceStart(e.lines[lineIdx])
	if !ok {
		return 0, 0, fencedCodeBlock{}, false
	}

	for end := lineIdx + 1; end < len(e.lines); end++ {
		if isCodeFenceEnd(e.lines[end], block.fenceRune, block.fenceLen) {
			return lineIdx, end, block, true
		}
	}

	// Markdown fenced blocks can run until EOF when there is no closing fence.
	return lineIdx, len(e.lines) - 1, block, true
}

func parseCodeFenceStart(line []rune) (fencedCodeBlock, bool) {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	if i >= len(line) {
		return fencedCodeBlock{}, false
	}

	fenceRune := line[i]
	if fenceRune != '`' && fenceRune != '~' {
		return fencedCodeBlock{}, false
	}

	j := i
	for j < len(line) && line[j] == fenceRune {
		j++
	}
	fenceLen := j - i
	if fenceLen < 3 {
		return fencedCodeBlock{}, false
	}

	rest := strings.TrimSpace(string(line[j:]))
	if fenceRune == '`' && strings.ContainsRune(rest, '`') {
		return fencedCodeBlock{}, false
	}

	lang := ""
	if rest != "" {
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			lang = fields[0]
		}
	}

	return fencedCodeBlock{
		fenceRune: fenceRune,
		fenceLen:  fenceLen,
		lang:      lang,
	}, true
}

func isCodeFenceEnd(line []rune, fenceRune rune, minLen int) bool {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	if i >= len(line) || line[i] != fenceRune {
		return false
	}

	j := i
	for j < len(line) && line[j] == fenceRune {
		j++
	}
	if j-i < minLen {
		return false
	}

	for ; j < len(line); j++ {
		if line[j] != ' ' && line[j] != '\t' {
			return false
		}
	}
	return true
}
