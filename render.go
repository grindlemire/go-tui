package tui

// Render the buffer to the terminal.
// It computes the diff between front and back buffers, flushes only
// the changed cells, and then swaps the buffers.
//
// This is the primary rendering function for normal frame updates.
func Render(term Terminal, buf *Buffer) {
	changes := buf.Diff()
	if len(changes) > 0 {
		term.Flush(changes)
	}
	buf.Swap()
}

// RenderFull forces a complete redraw of the buffer to the terminal.
// Unlike Render(), this sends all cells regardless of whether they changed.
//
// Use this after:
//   - Initial application startup
//   - Terminal resize
//   - Recovering from external terminal corruption
//   - Switching back from alternate screen
func RenderFull(term Terminal, buf *Buffer) {
	// Build a list of all cells as changes, skipping trailing empty cells in
	// each row. term.Clear() blanks the screen first, so unpainted trailing
	// cells stay blank. Emitting them as spaces would mark them as written and
	// stop terminals from trimming trailing whitespace on copy.
	width := buf.Width()
	height := buf.Height()
	changes := make([]CellChange, 0, width*height)

	for y := range height {
		trimEnd := -1
		for x := width - 1; x >= 0; x-- {
			c := buf.Cell(x, y)
			if !c.IsEmpty() && !c.IsContinuation() {
				trimEnd = x
				break
			}
		}
		for x := range trimEnd + 1 {
			changes = append(changes, CellChange{X: x, Y: y, Cell: buf.Cell(x, y)})
		}
	}

	term.Clear()
	if len(changes) > 0 {
		term.Flush(changes)
	}
	buf.Swap()
}
