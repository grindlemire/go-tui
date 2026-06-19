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

// sanitizeStyledText strips control characters but preserves ANSI escape sequences.
func sanitizeStyledText(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			// Preserve the entire ANSI sequence.
			start := i
			if i+1 < len(s) {
				switch s[i+1] {
				case '[':
					i += 2
					for i < len(s) {
						c := s[i]
						i++
						if c >= 0x40 && c <= 0x7e {
							break
						}
					}
				case ']':
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
					i += 2
				}
			} else {
				i++
			}
			b.WriteString(s[start:i])
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if r == utf8.RuneError && size == 1 {
			continue
		}
		switch {
		case r == '\n':
			b.WriteRune('\n')
		case r == '\t':
			b.WriteRune(' ')
		case r < 0x20 || r == 0x7f:
			// Drop control bytes.
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// wrapInlineStyledRows wraps text that may contain ANSI escape sequences.
// Escape sequences are preserved in the output but do not count toward column width.
// Text is segmented on grapheme cluster boundaries so multi-rune clusters (flags,
// ZWJ families, decomposed accents) are never split across lines or measured
// as the sum of their code points.
func wrapInlineStyledRows(text string, width int) []string {
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

	for i := 0; i < len(text); {
		// Pass ANSI sequences through without counting width.
		if text[i] == 0x1b {
			start := i
			if i+1 < len(text) && text[i+1] == '[' {
				i += 2
				for i < len(text) {
					c := text[i]
					i++
					if c >= 0x40 && c <= 0x7e {
						break
					}
				}
			} else {
				i += 2
				if i > len(text) {
					i = len(text)
				}
			}
			row.WriteString(text[start:i])
			continue
		}

		// Consume the next grapheme cluster: nextClusterWidth handles
		// decoding the base rune, extending past combining marks, ZWJ+base
		// sequences, and regional-indicator pairs. It advances i past the
		// whole cluster and returns the correct display width.
		clusterStart := i
		cw := nextClusterWidth(text, &i)

		if cw == 0 {
			// Reached end-of-string or newline.
			if i < len(text) && text[i] == '\n' {
				flush()
				i++
			}
			continue
		}

		if cw > width {
			cw = 1
			// Run the pre-flush check before writing "?" so an already-full row
			// is flushed first (mirrors wrapInlineVisualRows behavior).
			if col+cw > width {
				flush()
			}
			row.WriteString("?")
			col++
			continue
		}

		if col+cw > width {
			flush()
		}

		row.WriteString(text[clusterStart:i])
		col += cw
	}

	if row.Len() > 0 || len(rows) == 0 {
		rows = append(rows, row.String())
	}

	return rows
}

// nextClusterWidth consumes the next grapheme cluster from text starting at
// position *pos, advances *pos past the cluster (in bytes), and returns the
// cluster's display width. ANSI escape sequences are treated as cluster
// boundaries — they don't affect the width.
//
// This is the ANSI-aware analogue of clusterAdvance/nextCluster. Width
// updates for combining marks, VS16, and VS15 are delegated to the shared
// clusterExtendUpdateWidth helper so both state machines stay in sync.
func nextClusterWidth(text string, pos *int) int {
	start := *pos
	// Decode the base rune and its width.
	r, sz := utf8.DecodeRuneInString(text[start:])
	if sz == 0 {
		return 0
	}
	if r == '\n' {
		// Leave pos unchanged; caller can detect and flush.
		return 0
	}
	w := max(RuneWidth(r), 1)
	*pos += sz

	// Extend past any trailing combining marks, ZWJ sequences, and RI pairs
	// that form one grapheme cluster. ANSI sequences between runes are NOT
	// crossed — they act as cluster boundaries.
	lastWasZWJ := false
	for *pos < len(text) {
		// ANSI sequence between base and trailing runes: treat as boundary.
		if text[*pos] == 0x1b {
			break
		}

		r2, sz2 := utf8.DecodeRuneInString(text[*pos:])
		if sz2 == 0 || r2 == '\n' {
			break
		}

		// Regional-indicator pair: two consecutive RIs form one 2-col cluster.
		if regionalIndicator(r) && regionalIndicator(r2) && *pos == start+sz {
			*pos += sz2
			return 2
		}

		if graphemeExtend(r2) {
			*pos += sz2
			w = clusterExtendUpdateWidth(r2, r, w)
			lastWasZWJ = isZWJ(r2)
			continue
		}
		if lastWasZWJ {
			*pos += sz2
			w = 2
			lastWasZWJ = false
			continue
		}
		break
	}
	return w
}

// wrapInlineVisualRows converts text into terminal visual rows using
// grapheme cluster widths.
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

	rest := text
	for len(rest) > 0 {
		cluster, cw, size := nextCluster(rest)
		if size == 0 {
			break
		}
		rest = rest[size:]

		if cluster == "\n" {
			flush()
			continue
		}

		if cw > width {
			cluster = "?"
			cw = 1
		}

		if col+cw > width {
			flush()
		}

		row.WriteString(cluster)
		col += cw
	}

	if row.Len() > 0 || len(rows) == 0 {
		rows = append(rows, row.String())
	}

	return rows
}
