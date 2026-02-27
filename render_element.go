package tui

// bufferRowToANSI converts a single row of buffer cells to an ANSI-escaped
// string suitable for direct terminal output. Trailing empty cells are trimmed.
// The caller's escBuilder is reused to minimize allocations across rows.
func bufferRowToANSI(buf *Buffer, row int, esc *escBuilder, caps Capabilities) string {
	width := buf.Width()
	if width == 0 {
		return ""
	}

	// Find rightmost non-empty cell (trim point).
	trimEnd := -1
	for x := width - 1; x >= 0; x-- {
		c := buf.Cell(x, row)
		if !c.IsEmpty() && !c.IsContinuation() {
			trimEnd = x
			break
		}
	}
	if trimEnd < 0 {
		return ""
	}

	esc.Reset()

	var prevStyle Style
	styleSet := false

	for x := 0; x <= trimEnd; x++ {
		c := buf.Cell(x, row)

		// Skip continuation cells of wide characters.
		if c.IsContinuation() {
			continue
		}

		// Emit style change if needed.
		if !styleSet || !c.Style.Equal(prevStyle) {
			if c.Style.Equal(NewStyle()) {
				esc.ResetStyle()
			} else {
				esc.SetStyle(c.Style, caps)
			}
			prevStyle = c.Style
			styleSet = true
		}

		// Emit the rune (zero rune = empty cell, render as space).
		r := c.Rune
		if r == 0 {
			r = ' '
		}
		esc.WriteRune(r)
	}

	// Reset at end so styling doesn't bleed.
	esc.ResetStyle()

	return string(esc.Bytes())
}
