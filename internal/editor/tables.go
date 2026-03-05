package editor

import "strings"

func rangesOverlap(aStart, aEnd, bStart, bEnd int) bool {
	return aStart <= bEnd && bStart <= aEnd
}

func (e *editor) tableBlockStartAt(lineIdx int) (int, int, bool) {
	if lineIdx < 0 || lineIdx+1 >= len(e.lines) {
		return 0, 0, false
	}

	header := string(e.lines[lineIdx])
	delimiter := string(e.lines[lineIdx+1])
	if !isMarkdownTableHeaderLine(header) || !isMarkdownTableDelimiterLine(delimiter) {
		return 0, 0, false
	}

	end := lineIdx + 1
	for end+1 < len(e.lines) && isMarkdownTableRowLine(string(e.lines[end+1])) {
		end++
	}

	return lineIdx, end, true
}

func (e *editor) tableBlockForLine(lineIdx int) (int, int, bool) {
	if lineIdx < 0 || lineIdx >= len(e.lines) {
		return 0, 0, false
	}

	for i := 0; i < len(e.lines)-1; i++ {
		start, end, ok := e.tableBlockStartAt(i)
		if !ok {
			continue
		}
		if lineIdx >= start && lineIdx <= end {
			return start, end, true
		}
		i = end
	}

	return 0, 0, false
}

func isMarkdownTableHeaderLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "|")
}

func isMarkdownTableRowLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "|")
}

func isMarkdownTableDelimiterLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.Contains(trimmed, "|") {
		return false
	}

	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	if len(parts) == 0 {
		return false
	}

	hasColumn := false
	for _, part := range parts {
		cell := strings.TrimSpace(part)
		if cell == "" {
			return false
		}
		cell = strings.TrimPrefix(cell, ":")
		cell = strings.TrimSuffix(cell, ":")
		if len(cell) < 3 {
			return false
		}
		for _, r := range cell {
			if r != '-' {
				return false
			}
		}
		hasColumn = true
	}

	return hasColumn
}
