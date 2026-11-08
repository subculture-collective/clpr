package countdownrender

import (
	"strings"
	"unicode"
)

// cleanText keeps text the bundled fonts can draw: it drops control and
// formatting characters, emoji and pictographs (Noto Sans CJK has none), and
// collapses whitespace.
func cleanText(s string) string {
	var b strings.Builder
	space := false
	for _, r := range s {
		switch {
		case r == '�' || r == '‍' || (r >= '︀' && r <= '️'):
			continue
		case r >= 0x1F000 || (r >= 0x2600 && r <= 0x27BF) || (r >= 0x2B00 && r <= 0x2BFF) || (r >= 0xE000 && r <= 0xF8FF):
			continue
		case unicode.IsSpace(r):
			space = b.Len() > 0
			continue
		case unicode.IsControl(r) || unicode.Is(unicode.Cf, r):
			continue
		}
		if space {
			b.WriteByte(' ')
			space = false
		}
		b.WriteRune(r)
	}
	return b.String()
}

// runeWidth counts East Asian wide characters as two columns.
func runeWidth(r rune) int {
	if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) || (r >= 0xFF00 && r <= 0xFF60) {
		return 2
	}
	return 1
}

func textWidth(s string) int {
	n := 0
	for _, r := range s {
		n += runeWidth(r)
	}
	return n
}

// truncate cuts s to at most width columns, ending with an ellipsis when cut.
func truncate(s string, width int) string {
	if textWidth(s) <= width {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		w := runeWidth(r)
		if n+w > width-1 {
			break
		}
		b.WriteRune(r)
		n += w
	}
	return strings.TrimRight(b.String(), " ") + "…"
}

// wrap breaks s into at most maxLines lines of width columns, at spaces where
// possible and anywhere in text without spaces (CJK). The last line is
// truncated with an ellipsis.
func wrap(s string, width, maxLines int) []string {
	s = cleanText(s)
	if s == "" {
		return nil
	}
	var lines []string
	var line []rune
	lineWidth := 0
	flush := func() {
		lines = append(lines, strings.TrimSpace(string(line)))
		line, lineWidth = nil, 0
	}
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if len(lines) == maxLines-1 {
			rest := strings.TrimSpace(string(line) + string(runes[i:]))
			lines = append(lines, truncate(rest, width))
			return lines
		}
		r := runes[i]
		w := runeWidth(r)
		if lineWidth+w > width {
			// Move a partly written word to the next line when it fits there.
			if cut := strings.LastIndexByte(string(line), ' '); cut > 0 && r != ' ' {
				head, tail := string(line)[:cut], strings.TrimLeft(string(line)[cut:], " ")
				line = []rune(head)
				flush()
				line, lineWidth = []rune(tail), textWidth(tail)
			} else {
				flush()
			}
			if r == ' ' {
				continue
			}
		}
		line = append(line, r)
		lineWidth += w
	}
	if len(line) > 0 {
		flush()
	}
	return lines
}
