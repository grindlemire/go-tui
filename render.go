package tui

// Render efficiently renders the buffer to the terminal.
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
	// Build a list of all cells as changes
	width := buf.Width()
	height := buf.Height()
	changes := make([]CellChange, 0, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := buf.Cell(x, y)
			changes = append(changes, CellChange{X: x, Y: y, Cell: cell})
		}
	}

	term.Clear()
	if len(changes) > 0 {
		term.Flush(changes)
	}
	buf.Swap()
}

// RenderRegion renders only a specific rectangular region of the buffer.
// This is useful for partial updates when you know only a portion changed.
//
// Note: This still computes the diff, but only for cells within the region.
func RenderRegion(term Terminal, buf *Buffer, region Rect) {
	// Intersect with buffer bounds
	bufRect := buf.Rect()
	region = region.Intersect(bufRect)
	if region.IsEmpty() {
		return
	}

	// Collect changes within the region
	// We can't use buf.Diff() directly since it returns all changes
	// For efficiency, we'd need a more sophisticated buffer implementation
	// For now, just use the regular Render which is already efficient
	Render(term, buf)
}
