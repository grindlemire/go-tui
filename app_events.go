package tui

// Dispatch sends an event to the focused element.
// Handles ResizeEvent internally by updating buffer size and scheduling a full redraw.
// Handles MouseEvent by hit-testing to find the element under the cursor.
// Returns true if the event was consumed.
func (a *App) Dispatch(event Event) bool {
	// Handle ResizeEvent specially
	if resize, ok := event.(ResizeEvent); ok {
		if a.inAlternateScreen {
			// Alternate screen mode: always use full-screen sizing
			a.buffer.Resize(resize.Width, resize.Height)
		} else if a.inlineHeight > 0 {
			// Inline mode: recalculate start row, keep buffer height fixed
			a.inlineStartRow = resize.Height - a.inlineHeight
			widthChanged := a.buffer.Width() != resize.Width
			// Only resize buffer width if it changed
			if widthChanged {
				a.buffer.Resize(resize.Width, a.inlineHeight)
				a.invalidateInlineLayoutForWidthChange(a.inlineStartRow)
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

// dispatchMouseToComponents walks the component tree and dispatches a mouse
// event to all MouseListener components. Returns true if any consumed it.
func (a *App) dispatchMouseToComponents(me MouseEvent) bool {
	root, ok := a.root.(*Element)
	if !ok {
		return false
	}
	consumed := false
	walkComponents(root, func(comp Component) {
		if consumed {
			return
		}
		if ml, ok := comp.(MouseListener); ok {
			if ml.HandleMouse(me) {
				consumed = true
			}
		}
	})
	return consumed
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
			if keyEvent, isKey := ev.(KeyEvent); isKey {
				// Component model path: use broadcast dispatch table exclusively.
				// globalKeyHandler is skipped — components use KeyMap() instead.
				if a.dispatchTable != nil {
					stopped := a.dispatchTable.dispatch(keyEvent)
					if stopped {
						return // Event consumed by a Stop handler
					}
					// Event was not stopped - continue to App.Dispatch for element handlers
					// This allows onEvent handlers to see key events for inspection/logging
				} else {
					// Legacy path: global key handler runs before FocusManager dispatch.
					if a.globalKeyHandler != nil && a.globalKeyHandler(keyEvent) {
						return // Event consumed by global handler
					}
				}
			}
			// Component model path for mouse events: dispatch to MouseListener
			// components before falling through to element hit-testing.
			if mouseEvent, isMouse := ev.(MouseEvent); isMouse {
				if a.dispatchMouseToComponents(mouseEvent) {
					return
				}
			}
			// Non-key events (mouse, resize) and fallback for key events
			// without a dispatch table still go through App.Dispatch.
			a.Dispatch(ev)
		}
	}
}
