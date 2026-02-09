package tui

import "github.com/grindlemire/go-tui/internal/debug"

// Render clears the buffer, renders the element tree, and flushes to terminal.
// If a resize occurred since the last render, this automatically performs a full
// redraw to eliminate visual artifacts.
func (a *App) Render() {
	width, termHeight := a.terminal.Size()

	// Determine the render height based on mode
	renderHeight := termHeight
	if !a.inAlternateScreen && a.inlineHeight > 0 {
		renderHeight = a.inlineHeight
	}

	// Ensure buffer matches expected size (handles rapid resize)
	if a.buffer.Width() != width || a.buffer.Height() != renderHeight {
		if a.inAlternateScreen {
			// Alternate screen mode: always use full-screen sizing
			a.terminal.Clear()
			a.buffer.Resize(width, termHeight)
		} else if a.inlineHeight > 0 {
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

	// If a root component is set, re-render it to get a fresh element tree.
	// This is the core of the reactivity cycle: state changes → dirty → re-render
	// component → new element tree with updated state reads.
	if a.rootComponent != nil {
		el := a.rootComponent.Render()
		el.component = a.rootComponent
		a.root = el
	}

	// Re-read renderHeight in case SetInlineHeight was called during component render
	if !a.inAlternateScreen && a.inlineHeight > 0 {
		renderHeight = a.inlineHeight
	}

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, renderHeight)
	}

	// Sweep mount cache: clean up components no longer in the tree.
	// Mount() marks active keys during Render(); sweep removes the rest.
	if a.mounts != nil {
		a.mounts.sweep()
	}

	// Collect and start component watchers (once after first render)
	if !a.componentWatchersStarted {
		if root, ok := a.root.(*Element); ok {
			a.componentWatchers = collectComponentWatchers(root)
			for _, w := range a.componentWatchers {
				w.Start(a.eventQueue, a.stopCh)
			}
		}
		a.componentWatchersStarted = true
	}

	// Flush to terminal (inline mode offsets Y coordinates)
	if !a.inAlternateScreen && a.inlineHeight > 0 {
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
		debug.Log("renderInline: fullRedraw — SetCursor(0, %d), ClearToEnd, flushing %dx%d cells at Y offset %d",
			a.inlineStartRow, width, height, a.inlineStartRow)
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
