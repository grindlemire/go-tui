package tui

// Dispatch sends an event to the focused element.
// Handles ResizeEvent internally by updating buffer size and scheduling a full redraw.
// Handles MouseEvent by hit-testing to find the element under the cursor.
// Returns true if the event was consumed.
func (a *App) Dispatch(event Event) bool {
	// Handle ResizeEvent specially
	if resize, ok := event.(ResizeEvent); ok {
		if a.inlineHeight > 0 {
			// Inline mode: recalculate start row, keep buffer height fixed
			a.inlineStartRow = resize.Height - a.inlineHeight
			// Only resize buffer width if it changed
			if a.buffer.Width() != resize.Width {
				a.buffer.Resize(resize.Width, a.inlineHeight)
			}
		} else {
			// Full screen mode: resize buffer to match terminal
			a.buffer.Resize(resize.Width, resize.Height)
		}

		// Mark root dirty so layout is recalculated
		if a.root != nil {
			a.root.MarkDirty()
		}

		// Schedule full redraw to clear any visual artifacts
		a.needsFullRedraw = true

		return true
	}

	// Handle MouseEvent by hit-testing to find the element under the cursor
	if mouse, ok := event.(MouseEvent); ok {
		if a.root == nil {
			return false
		}
		// Check if root supports hit-testing
		if hitTester, ok := a.root.(mouseHitTester); ok {
			if target := hitTester.ElementAtPoint(mouse.X, mouse.Y); target != nil {
				return target.HandleEvent(event)
			}
		}
		return false
	}

	// Delegate to FocusManager for other events
	return a.focus.Dispatch(event)
}

// readInputEvents reads terminal input in a goroutine and queues events.
func (a *App) readInputEvents() {
	for {
		select {
		case <-a.stopCh:
			return
		default:
		}

		event, ok := a.reader.PollEvent(a.inputLatency)
		if !ok {
			continue
		}

		// Capture event for closure
		ev := event

		a.eventQueue <- func() {
			// Global key handler runs first (for app-level bindings like quit)
			if keyEvent, isKey := ev.(KeyEvent); isKey {
				if a.globalKeyHandler != nil && a.globalKeyHandler(keyEvent) {
					return // Event consumed by global handler
				}
			}
			a.Dispatch(ev)
		}
	}
}
