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
			// Inline mode: keep buffer height fixed to inlineHeight.
			a.syncInlineGeometryOnResize(resize.Width, resize.Height)
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

	// Handle MouseEvent by hit-testing to find the element under the cursor.
	// Coordinates are already translated to buffer-space by readInputEvents
	// when in inline mode.
	if mouse, ok := event.(MouseEvent); ok {
		mouse.app = a
		if a.root == nil {
			return false
		}
		if target := a.root.ElementAtPoint(mouse.X, mouse.Y); target != nil {
			return target.HandleEvent(mouse)
		}
		return false
	}

	// Inject app on key events reaching Dispatch directly
	if ke, ok := event.(KeyEvent); ok {
		ke.app = a
		return a.focus.Dispatch(ke)
	}

	// Delegate to focusManager for other events
	return a.focus.Dispatch(event)
}

// dispatchMouseToComponents walks the component tree and dispatches a mouse
// event to all MouseListener components. Returns true if any consumed it.
func (a *App) dispatchMouseToComponents(me MouseEvent) bool {
	if a.root == nil {
		return false
	}
	consumed := false
	walkComponents(a.rootComponent, a.root, func(comp Component) {
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
				keyEvent.app = a
				// Component model path: use broadcast dispatch table exclusively.
				// globalKeyHandler is skipped — components use KeyMap() instead.
				if a.dispatchTable != nil {
					stopped := a.dispatchTable.dispatch(keyEvent)
					if stopped {
						return // Event consumed by a Stop handler
					}
					// Fallback: Ctrl+Z triggers suspend if not consumed
					if keyEvent.Key == KeyRune && keyEvent.Rune == 'z' && keyEvent.Mod == ModCtrl {
						a.suspend()
						return
					}
					// Event was not stopped - continue to App.Dispatch for element handlers
					// This allows onEvent handlers to see key events for inspection/logging
				} else {
					// Legacy path: global key handler runs before focusManager dispatch.
					if a.globalKeyHandler != nil && a.globalKeyHandler(keyEvent) {
						return // Event consumed by global handler
					}
					// Fallback: Ctrl+Z triggers suspend if not consumed by global handler
					if keyEvent.Key == KeyRune && keyEvent.Rune == 'z' && keyEvent.Mod == ModCtrl {
						a.suspend()
						return
					}
				}
				a.Dispatch(keyEvent)
				return
			}
			// Component model path for mouse events: dispatch to MouseListener
			// components before falling through to element hit-testing.
			if mouseEvent, isMouse := ev.(MouseEvent); isMouse {
				mouseEvent.app = a
				// In inline mode, translate terminal-space Y to buffer-space Y
				// before any dispatch path sees the event. Both HandleClicks
				// (via ContainsPoint) and ElementAtPoint use buffer-relative
				// coordinates.
				if !a.inAlternateScreen && a.inlineHeight > 0 {
					mouseEvent.Y -= a.inlineStartRow
					if mouseEvent.Y < 0 || mouseEvent.Y >= a.inlineHeight {
						return // click outside the inline region
					}
				}
				if a.dispatchMouseToComponents(mouseEvent) {
					return
				}
				a.Dispatch(mouseEvent)
				return
			}
			// Non-key events (resize) go through App.Dispatch.
			a.Dispatch(ev)
		}
	}
}

func (a *App) syncInlineGeometryOnResize(width, termHeight int) {
	a.inlineStartRow = termHeight - a.inlineHeight
	if a.buffer.Width() == width {
		return
	}

	a.buffer.Resize(width, a.inlineHeight)
	a.invalidateInlineLayoutForWidthChange(a.inlineStartRow)
}
