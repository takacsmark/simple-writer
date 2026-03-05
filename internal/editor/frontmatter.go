package editor

import "strings"

type frontMatterKind int

const (
	frontMatterYAML frontMatterKind = iota
	frontMatterTOML
)

func (e *editor) frontMatterBlockAt(lineIdx int) (int, int, frontMatterKind, bool) {
	if lineIdx != 0 || len(e.lines) < 2 {
		return 0, 0, frontMatterYAML, false
	}

	kind, delim, ok := parseFrontMatterDelimiter(e.lines[0])
	if !ok {
		return 0, 0, frontMatterYAML, false
	}

	for i := 1; i < len(e.lines); i++ {
		if strings.TrimSpace(string(e.lines[i])) == delim {
			return 0, i, kind, true
		}
	}

	// Require a closing delimiter to avoid treating a top-level hr as frontmatter.
	return 0, 0, frontMatterYAML, false
}

func (e *editor) frontMatterKindForLine(lineIdx int) (frontMatterKind, bool) {
	start, end, kind, ok := e.frontMatterBlockAt(0)
	if !ok {
		return frontMatterYAML, false
	}
	if lineIdx < start || lineIdx > end {
		return frontMatterYAML, false
	}
	return kind, true
}

func parseFrontMatterDelimiter(line []rune) (frontMatterKind, string, bool) {
	trimmed := strings.TrimSpace(string(line))
	switch trimmed {
	case "---":
		return frontMatterYAML, "---", true
	case "+++":
		return frontMatterTOML, "+++", true
	default:
		return frontMatterYAML, "", false
	}
}
