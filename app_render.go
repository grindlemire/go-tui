package tui

import "github.com/grindlemire/go-tui/internal/debug"

// Render performs layout and renders to the terminal if the dirty flag is set.
// No-op if nothing has changed since the last render. After rendering, the
// dispatch table is rebuilt from the current component tree.
// Use RenderFull() to force a re-render regardless of dirty state.
func (a *App) Render() {
	if !a.checkAndClearDirty() {
		return
	}
	a.renderFrame()
	a.rebuildDispatchTable()
}

// renderFrame performs the actual render cycle: clear buffer, re-render
// components, render element tree, flush to terminal.
func (a *App) renderFrame() {
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
			// Inline mode: keep buffer height fixed to inlineHeight.
			a.syncInlineGeometryOnResize(width, termHeight)
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

	// Clear overlay registrations from previous frame
	a.clearOverlays()

	// If a root component is set, re-render it to get a fresh element tree.
	// This is the core of the reactivity cycle: state changes → dirty → re-render
	// component → new element tree with updated state reads.
	a.rerenderComponent()

	// Re-read renderHeight in case SetInlineHeight was called during component render
	if !a.inAlternateScreen && a.inlineHeight > 0 {
		renderHeight = a.inlineHeight
	}

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, renderHeight)
	}

	a.renderOverlays(width, renderHeight)

	// Sweep mount cache: clean up components no longer in the tree.
	// Mount() marks active keys during Render(); sweep removes the rest.
	if a.mounts != nil {
		a.mounts.sweep()
	}

	// Collect and start component watchers (once after first render)
	if !a.componentWatchersStarted {
		if a.root != nil {
			a.componentWatchers = collectComponentWatchers(a.rootComponent, a.root)
			for _, w := range a.componentWatchers {
				w.Start(a.watcherQueue, a.rootWatcherCh)
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
// This performs a full render cycle: re-renders the component tree (triggering
// state reads, overlay registration, and focus refresh), then flushes every
// cell to the terminal. Use this after resize events or when the terminal may
// be corrupted.
func (a *App) RenderFull() {
	width, height := a.terminal.Size()

	// Clear buffer
	a.buffer.Clear()

	// Clear overlay registrations from previous frame
	a.clearOverlays()

	// Re-render the component tree so overlays and state are up to date.
	a.rerenderComponent()

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, height)
	}

	a.renderOverlays(width, height)

	// Full render to terminal
	RenderFull(a.terminal, a.buffer)

	a.rebuildDispatchTable()
}

// rerenderComponent re-renders the root component to produce a fresh element tree.
// Called by both Render() and RenderFull() to keep overlays and state in sync.
func (a *App) rerenderComponent() {
	if a.rootComponent == nil {
		return
	}
	el := a.rootComponent.Render(a)
	el.setAppRecursive(a)
	a.root = el
	// Refresh focusManager references: re-renders produce new Element
	// objects, so the focusManager's old references become stale.
	// Rebuild the focusable list from the current tree, preserving
	// the focus index so the focused element stays focused.
	a.focus.refreshFromTree(el)
}

// renderOverlays applies focus scoping and renders overlay elements (modals)
// on top of the main element tree. Called by both Render() and RenderFull().
func (a *App) renderOverlays(width, height int) {
	// Apply focus scoping so overlay elements get correct focus borders.
	focusScoped := false
	for i := len(a.overlays) - 1; i >= 0; i-- {
		if a.overlays[i].trapFocus {
			a.focus.ScopeTo(a.overlays[i].element)
			focusScoped = true
			// Move focus into the modal only on the first frame after open.
			// This avoids calling Next()/MarkDirty() on every render frame.
			if a.overlays[i].needsFocusInit {
				a.overlays[i].needsFocusInit = false
				current := a.focus.Focused()
				if current == nil || !a.focus.isInScope(current) {
					a.focus.Next()
				}
			}
			break
		}
	}
	if !focusScoped && a.focus.scope != nil {
		a.focus.ClearScope()
	}

	// Render overlay elements (modals) on top of the main tree
	for _, ov := range a.overlays {
		switch ov.backdrop {
		case "dim":
			a.buffer.ApplyDim()
		case "blank":
			a.buffer.FillBlank()
		}
		// Ensure overlay content children have an opaque background so the
		// backdrop effect doesn't bleed through the dialog body. The overlay
		// element itself stays transparent for backdrop click detection.
		for _, child := range ov.element.children {
			if child.background == nil {
				bg := NewStyle()
				child.background = &bg
			}
		}
		ov.element.Render(a.buffer, width, height)
	}
}
