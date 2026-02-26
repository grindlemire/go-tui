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
