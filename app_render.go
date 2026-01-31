package tui

// Render clears the buffer, renders the element tree, and flushes to terminal.
// If a resize occurred since the last render, this automatically performs a full
// redraw to eliminate visual artifacts.
func (a *App) Render() {
	width, termHeight := a.terminal.Size()

	// Determine the render height based on mode
	renderHeight := termHeight
	if a.inlineHeight > 0 {
		renderHeight = a.inlineHeight
	}

	// Ensure buffer matches expected size (handles rapid resize)
	if a.buffer.Width() != width || a.buffer.Height() != renderHeight {
		if a.inlineHeight > 0 {
			// Inline mode: update start row, resize buffer width only if needed
			a.inlineStartRow = termHeight - a.inlineHeight
			if a.buffer.Width() != width {
				a.buffer.Resize(width, a.inlineHeight)
			}
		} else {
			// Full screen mode: clear terminal and resize buffer
			a.terminal.Clear()
			a.buffer.Resize(width, termHeight)
		}
		if a.root != nil {
			a.root.MarkDirty()
		}
		a.needsFullRedraw = true
	}

	// Clear buffer
	a.buffer.Clear()

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, renderHeight)
	}

	// Flush to terminal (inline mode offsets Y coordinates)
	if a.inlineHeight > 0 {
		a.renderInline()
	} else if a.needsFullRedraw {
		RenderFull(a.terminal, a.buffer)
		a.needsFullRedraw = false
	} else {
		Render(a.terminal, a.buffer)
	}
}

// renderInline handles rendering for inline mode by offsetting Y coordinates.
func (a *App) renderInline() {
	var changes []CellChange

	if a.needsFullRedraw {
		// Build all cells as changes
		width := a.buffer.Width()
		height := a.buffer.Height()
		changes = make([]CellChange, 0, width*height)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				cell := a.buffer.Cell(x, y)
				changes = append(changes, CellChange{X: x, Y: y + a.inlineStartRow, Cell: cell})
			}
		}
		// Clear only the inline region, not the whole screen
		a.terminal.SetCursor(0, a.inlineStartRow)
		a.terminal.ClearToEnd()
		a.needsFullRedraw = false
	} else {
		// Get diff and offset Y coordinates
		diff := a.buffer.Diff()
		changes = make([]CellChange, len(diff))
		for i, ch := range diff {
			changes[i] = CellChange{X: ch.X, Y: ch.Y + a.inlineStartRow, Cell: ch.Cell}
		}
	}

	if len(changes) > 0 {
		a.terminal.Flush(changes)
	}
	a.buffer.Swap()
}

// RenderFull forces a complete redraw of the buffer to the terminal.
// Use this after resize events or when the terminal may be corrupted.
func (a *App) RenderFull() {
	width, height := a.terminal.Size()

	// Clear buffer
	a.buffer.Clear()

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, height)
	}

	// Full render to terminal
	RenderFull(a.terminal, a.buffer)
}
