package tui

import "strings"

// wrapText wraps text to fit within maxWidth terminal cells using word boundaries.
// It breaks at spaces, falling back to mid-character breaks when a single word
// exceeds maxWidth. Existing newlines in the text are preserved.
func wrapText(text string, maxWidth int) []string {
	if maxWidth < 1 {
		return []string{""}
	}
	if text == "" {
		return []string{""}
	}

	var result []string
	for _, paragraph := range strings.Split(text, "\n") {
		result = append(result, wrapParagraph(paragraph, maxWidth)...)
	}
	return result
}

// wrapParagraph wraps a single paragraph (no newlines) to maxWidth.
func wrapParagraph(text string, maxWidth int) []string {
	if text == "" {
		return []string{""}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var buf strings.Builder
	lineWidth := 0

	for _, word := range words {
		ww := stringWidth(word)

		if ww > maxWidth {
			// Word is longer than line — flush then break character by character
			if lineWidth > 0 {
				lines = append(lines, buf.String())
				buf.Reset()
				lineWidth = 0
			}
			for _, r := range word {
				rw := RuneWidth(r)
				if lineWidth+rw > maxWidth && lineWidth > 0 {
					lines = append(lines, buf.String())
					buf.Reset()
					lineWidth = 0
				}
				buf.WriteRune(r)
				lineWidth += rw
			}
			continue
		}

		if lineWidth == 0 {
			// First word on line
			buf.WriteString(word)
			lineWidth = ww
		} else if lineWidth+1+ww <= maxWidth {
			// Fits with space
			buf.WriteByte(' ')
			buf.WriteString(word)
			lineWidth += 1 + ww
		} else {
			// Doesn't fit — start new line
			lines = append(lines, buf.String())
			buf.Reset()
			buf.WriteString(word)
			lineWidth = ww
		}
	}

	lines = append(lines, buf.String())
	return lines
}

// styledRune is one rune carrying the style of its source span.
type styledRune struct {
	r  rune
	st Style
}

// wrapSpans wraps styled spans to maxWidth using word boundaries, mirroring
// wrapParagraph. Words are delimited by actual whitespace in the concatenated
// text (a span boundary is NOT a word boundary), so a single word may mix styles
// (e.g. "a**b**c" is one word "abc"). Each rune keeps its source span's style, so
// a styled run stays styled across a line break. Adjacent same-style runes on a
// line are merged into one segment. Newlines start new lines.
func wrapSpans(spans []TextSpan, maxWidth int) [][]TextSpan {
	if maxWidth < 1 {
		return [][]TextSpan{{}}
	}

	var lines [][]TextSpan
	var cur []TextSpan // current line segments
	lineWidth := 0

	flush := func() {
		lines = append(lines, cur)
		cur = nil
		lineWidth = 0
	}
	// emit appends styled runes to cur, merging same-style into the last segment.
	emit := func(rs []styledRune) {
		for _, sr := range rs {
			if n := len(cur); n > 0 && cur[n-1].Style == sr.st {
				cur[n-1].Text += string(sr.r)
			} else {
				cur = append(cur, TextSpan{Text: string(sr.r), Style: sr.st})
			}
		}
	}
	// emitSpace appends a single separator space (default style; it merges to the
	// element's base style at render time).
	emitSpace := func() {
		if n := len(cur); n > 0 && cur[n-1].Style == (Style{}) {
			cur[n-1].Text += " "
		} else {
			cur = append(cur, TextSpan{Text: " "})
		}
	}

	var word []styledRune
	wordWidth := 0
	placeWord := func() {
		if len(word) == 0 {
			return
		}
		if wordWidth > maxWidth {
			// Word longer than the line: flush current, then hard-break by rune.
			if lineWidth > 0 {
				flush()
			}
			for _, sr := range word {
				rw := RuneWidth(sr.r)
				if lineWidth+rw > maxWidth && lineWidth > 0 {
					flush()
				}
				emit([]styledRune{sr})
				lineWidth += rw
			}
			word = word[:0]
			wordWidth = 0
			return
		}
		switch {
		case lineWidth == 0:
			emit(word)
			lineWidth = wordWidth
		case lineWidth+1+wordWidth <= maxWidth:
			emitSpace()
			emit(word)
			lineWidth += 1 + wordWidth
		default:
			flush()
			emit(word)
			lineWidth = wordWidth
		}
		word = word[:0]
		wordWidth = 0
	}

	for _, sp := range spans {
		for _, r := range sp.Text {
			switch r {
			case '\n':
				placeWord()
				flush()
			case ' ', '\t':
				placeWord()
			default:
				word = append(word, styledRune{r: r, st: sp.Style})
				wordWidth += RuneWidth(r)
			}
		}
	}
	placeWord()
	flush()
	return lines
}
