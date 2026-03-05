package main

import (
	"strings"

	"github.com/charmbracelet/glamour"
	glamansi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	xansi "github.com/charmbracelet/x/ansi"
)

type markdownRenderer struct {
	width    int
	renderer *glamour.TermRenderer
}

func newMarkdownRenderer() *markdownRenderer {
	return &markdownRenderer{width: -1}
}

func (m *markdownRenderer) ensureRenderer(width int) error {
	if m.renderer != nil && m.width == width {
		return nil
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithTableWrap(false),
		glamour.WithPreservedNewLines(),
		glamour.WithStyles(minimalMarkdownStyle()),
	)
	if err != nil {
		return err
	}

	m.renderer = r
	m.width = width
	return nil
}

func (m *markdownRenderer) renderMarkdownLine(line []rune, width int) ([]string, error) {
	if width <= 0 {
		return []string{""}, nil
	}
	if len(line) == 0 {
		return []string{fitANSIWidth("", width)}, nil
	}
	preprocessed := stripInlineLinkDestinations(string(line))
	return m.renderMarkdownText(preprocessed, width)
}

func (m *markdownRenderer) renderMarkdownTableBlock(lines [][]rune, width int) ([]string, error) {
	if width <= 0 {
		return []string{""}, nil
	}
	if len(lines) == 0 {
		return []string{}, nil
	}

	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		parts = append(parts, stripInlineLinkDestinations(string(line)))
	}
	return m.renderMarkdownText(strings.Join(parts, "\n"), width)
}

func (m *markdownRenderer) renderMarkdownCodeLine(lang string, line []rune, width int) ([]string, error) {
	if width <= 0 {
		return []string{""}, nil
	}
	input := fencedCodeSnippet(lang, string(line))
	return m.renderMarkdownText(input, width)
}

func fencedCodeSnippet(lang, code string) string {
	fenceLen := max(3, maxByteRun(code, '`')+1)
	fence := strings.Repeat("`", fenceLen)
	if strings.ContainsRune(lang, '`') {
		lang = ""
	}
	if lang != "" {
		return fence + lang + "\n" + code + "\n" + fence
	}
	return fence + "\n" + code + "\n" + fence
}

func maxByteRun(s string, b byte) int {
	best := 0
	run := 0
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			run++
			if run > best {
				best = run
			}
			continue
		}
		run = 0
	}
	return best
}

func (m *markdownRenderer) renderMarkdownText(input string, width int) ([]string, error) {
	if err := m.ensureRenderer(width); err != nil {
		return nil, err
	}

	rendered, err := m.renderer.Render(input)
	if err != nil {
		return nil, err
	}
	rendered = strings.ReplaceAll(rendered, "\r\n", "\n")

	rows := strings.Split(rendered, "\n")
	for len(rows) > 1 && isVisuallyBlankANSI(rows[0]) {
		rows = rows[1:]
	}
	for len(rows) > 1 && isVisuallyBlankANSI(rows[len(rows)-1]) {
		rows = rows[:len(rows)-1]
	}
	if len(rows) == 0 {
		rows = []string{""}
	}

	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, fitANSIWidth(row, width))
	}
	return out, nil
}

func isVisuallyBlankANSI(s string) bool {
	plain := xansi.Strip(s)
	return strings.TrimSpace(plain) == ""
}

func stripInlineLinkDestinations(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		// Leave image syntax untouched: ![alt](url)
		if s[i] == '!' && i+1 < len(s) && s[i+1] == '[' && !isEscapedByte(s, i) {
			textEnd, ok := scanBracketContentEnd(s, i+1)
			if ok {
				j := textEnd + 1
				for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
					j++
				}
				if j < len(s) && s[j] == '(' {
					destEnd, ok := scanParenContentEnd(s, j)
					if ok {
						b.WriteString(s[i : destEnd+1])
						i = destEnd + 1
						continue
					}
				}
			}
		}

		if s[i] != '[' || isEscapedByte(s, i) {
			b.WriteByte(s[i])
			i++
			continue
		}

		textEnd, ok := scanBracketContentEnd(s, i)
		if !ok {
			b.WriteByte(s[i])
			i++
			continue
		}

		j := textEnd + 1
		for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
			j++
		}
		if j < len(s) && s[j] == '[' {
			refEnd, ok := scanBracketContentEnd(s, j)
			if ok {
				b.WriteString(s[i : textEnd+1])
				b.WriteString("()")
				i = refEnd + 1
				continue
			}
		}

		if j >= len(s) || s[j] != '(' {
			b.WriteByte(s[i])
			i++
			continue
		}

		destEnd, ok := scanParenContentEnd(s, j)
		if !ok {
			b.WriteByte(s[i])
			i++
			continue
		}

		b.WriteString(s[i : textEnd+1])
		b.WriteString("()")
		i = destEnd + 1
	}

	return b.String()
}

func scanBracketContentEnd(s string, start int) (int, bool) {
	depth := 0
	for i := start; i < len(s); i++ {
		if isEscapedByte(s, i) {
			continue
		}
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func scanParenContentEnd(s string, start int) (int, bool) {
	depth := 0
	quote := byte(0)
	inAngle := false
	for i := start; i < len(s); i++ {
		if isEscapedByte(s, i) {
			continue
		}

		ch := s[i]
		if quote != 0 {
			if ch == quote {
				quote = 0
			}
			continue
		}
		if inAngle {
			if ch == '>' {
				inAngle = false
			}
			continue
		}

		switch ch {
		case '"', '\'':
			quote = ch
		case '<':
			inAngle = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func isEscapedByte(s string, i int) bool {
	backslashes := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func fitANSIWidth(s string, width int) string {
	clipped := xansi.Truncate(s, width, "")
	w := xansi.StringWidth(clipped)
	if w < width {
		clipped += strings.Repeat(" ", width-w)
	}
	return clipped
}

func minimalMarkdownStyle() glamansi.StyleConfig {
	style := styles.DarkStyleConfig
	zero := uint(0)

	// Keep default Glamour look, but remove extra structural margins so rows
	// don't jump horizontally/vertically when a line flips raw/rendered by mode.
	style.Document.Margin = &zero
	style.Document.BlockPrefix = ""
	style.Document.BlockSuffix = ""
	style.Paragraph.Margin = &zero
	style.Paragraph.BlockPrefix = ""
	style.Paragraph.BlockSuffix = ""
	style.CodeBlock.Margin = &zero
	style.DefinitionDescription.BlockPrefix = ""

	// We hide link destinations in render mode; keep label styling obvious.
	style.LinkText.Underline = boolPtr(true)
	style.LinkText.Color = stringPtr(linkLabelColorCode())
	style.Link.Color = stringPtr(linkDestColorCode())
	style.Link.Underline = boolPtr(true)

	return style
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
