package tui

import (
	"strings"
	"unicode/utf8"
)

// sanitizeInlineText strips control/ANSI sequences from appended history text.
// Inline history content is always treated as plain text, not terminal control.
func sanitizeInlineText(s string) string {
	s = stripANSISequences(s)

	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch {
		case r == '\n':
			b.WriteRune('\n')
		case r == '\t':
			b.WriteRune(' ')
		case r < 0x20 || r == 0x7f:
			// Drop remaining C0/DEL control bytes.
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// stripANSISequences removes common escape-sequence forms from text.
func stripANSISequences(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if s[i] != 0x1b {
			r, size := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && size == 1 {
				i++
				continue
			}
			b.WriteRune(r)
			i += size
			continue
		}

		if i+1 >= len(s) {
			i++
			continue
		}

		switch s[i+1] {
		case '[':
			// CSI: ESC [ ... final-byte
			i += 2
			for i < len(s) {
				c := s[i]
				i++
				if c >= 0x40 && c <= 0x7e {
					break
				}
			}
		case ']':
			// OSC: ESC ] ... BEL or ST
			i += 2
			for i < len(s) {
				c := s[i]
				i++
				if c == 0x07 {
					break
				}
				if c == 0x1b && i < len(s) && s[i] == '\\' {
					i++
					break
				}
			}
		default:
			// Generic 2-byte escape.
			i += 2
		}
	}

	return b.String()
}

// wrapInlineVisualRows converts text into terminal visual rows using RuneWidth.
func wrapInlineVisualRows(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	if text == "" {
		return []string{""}
	}

	rows := make([]string, 0, 4)
	var row strings.Builder
	col := 0

	flush := func() {
		rows = append(rows, row.String())
		row.Reset()
		col = 0
	}

	for _, r := range text {
		if r == '\n' {
			flush()
			continue
		}

		w := RuneWidth(r)
		if w < 1 {
			w = 1
		}
		if w > width {
			r = '?'
			w = 1
		}

		if col+w > width {
			flush()
		}

		row.WriteRune(r)
		col += w
	}

	if row.Len() > 0 || len(rows) == 0 {
		rows = append(rows, row.String())
	}

	return rows
}
