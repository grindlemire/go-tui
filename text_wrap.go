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

// styledWord is one whitespace-delimited token with the style of its source span.
type styledWord struct {
	text  string
	style Style
}

// wrapSpans wraps styled spans to maxWidth using word boundaries, mirroring
// wrapParagraph. Each emitted word keeps its source span's style, so a multi-word
// styled run stays styled across a line break. Adjacent same-style segments on a
// line are merged. Newlines inside span text start new lines.
func wrapSpans(spans []TextSpan, maxWidth int) [][]TextSpan {
	if maxWidth < 1 {
		return [][]TextSpan{{}}
	}

	// Flatten spans into styled words, treating '\n' as a hard line break.
	var words []styledWord
	for _, sp := range spans {
		for i, para := range strings.Split(sp.Text, "\n") {
			if i > 0 {
				words = append(words, styledWord{text: "\n"}) // marker
			}
			for _, w := range strings.Fields(para) {
				words = append(words, styledWord{text: w, style: sp.Style})
			}
		}
	}

	var lines [][]TextSpan
	var cur []TextSpan
	lineWidth := 0

	flush := func() {
		lines = append(lines, cur)
		cur = nil
		lineWidth = 0
	}
	// appendWord adds a word to cur, merging into the last segment if same style.
	appendWord := func(w styledWord, leadingSpace bool) {
		text := w.text
		if leadingSpace {
			text = " " + text
		}
		if n := len(cur); n > 0 && cur[n-1].Style == w.style {
			cur[n-1].Text += text
		} else {
			cur = append(cur, TextSpan{Text: text, Style: w.style})
		}
	}

	for _, w := range words {
		if w.text == "\n" {
			flush()
			continue
		}
		ww := stringWidth(w.text)
		if ww > maxWidth {
			// Word longer than the line: flush current, then hard-break by rune.
			if lineWidth > 0 {
				flush()
			}
			for _, r := range w.text {
				rw := RuneWidth(r)
				if lineWidth+rw > maxWidth && lineWidth > 0 {
					flush()
				}
				appendWord(styledWord{text: string(r), style: w.style}, false)
				lineWidth += rw
			}
			continue
		}
		switch {
		case lineWidth == 0:
			appendWord(w, false)
			lineWidth = ww
		case lineWidth+1+ww <= maxWidth:
			appendWord(w, true)
			lineWidth += 1 + ww
		default:
			flush()
			appendWord(w, false)
			lineWidth = ww
		}
	}
	flush()
	return lines
}
