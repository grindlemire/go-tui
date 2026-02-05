package tui

import (
	"github.com/grindlemire/go-tui/internal/debug"
)

// --- Focus API ---

// IsFocusable returns whether this element can receive focus.
func (e *Element) IsFocusable() bool {
	return e.focusable
}

// IsFocused returns whether this element currently has focus.
func (e *Element) IsFocused() bool {
	return e.focused
}

// Focus marks this element as focused and calls onFocus callback if set.
// Does not cascade to children — only the FocusManager target receives focus.
func (e *Element) Focus() {
	debug.Log("Element.Focus: text=%q", e.text)
	e.focused = true
	if e.onFocus != nil {
		e.onFocus(e)
	}
}

// Blur marks this element as not focused and calls onBlur callback if set.
// Does not cascade to children — only the FocusManager target loses focus.
func (e *Element) Blur() {
	e.focused = false
	if e.onBlur != nil {
		e.onBlur(e)
	}
}

// SetFocusable sets whether this element can receive focus.
func (e *Element) SetFocusable(focusable bool) {
	e.focusable = focusable
}

// --- Event Handler API ---

// SetOnKeyPress sets a handler for key press events.
// The handler receives the element as its first parameter (self-inject).
// Return true to consume the event (stops bubbling to parent elements).
// Return false to let the event bubble up to the parent.
// Does NOT set focusable — this handler participates in bubbling without
// making the element a Tab navigation target. Use SetFocusable(true) explicitly
// if this element should receive focus directly.
func (e *Element) SetOnKeyPress(fn func(*Element, KeyEvent) bool) {
	e.onKeyPress = fn
}

// SetOnClick sets a handler for click events.
// The handler receives the element as its first parameter (self-inject).
// No return value needed - mutations mark dirty automatically via MarkDirty().
// Implicitly sets focusable = true so the element can receive mouse and keyboard events.
func (e *Element) SetOnClick(fn func(*Element)) {
	e.focusable = true
	e.onClick = fn
}

// SetOnEvent sets the event handler for this element.
// The handler receives the element as its first parameter (self-inject).
// Does NOT set focusable — this handler participates in bubbling without
// making the element a Tab navigation target. Use SetFocusable(true) explicitly
// if this element should receive focus directly.
func (e *Element) SetOnEvent(fn func(*Element, Event) bool) {
	e.onEvent = fn
}

// SetOnFocus sets a handler that's called when this element gains focus.
// The handler receives the element as its first parameter (self-inject).
// Implicitly sets focusable = true.
func (e *Element) SetOnFocus(fn func(*Element)) {
	e.focusable = true
	e.onFocus = fn
}

// SetOnBlur sets a handler that's called when this element loses focus.
// The handler receives the element as its first parameter (self-inject).
// Implicitly sets focusable = true.
func (e *Element) SetOnBlur(fn func(*Element)) {
	e.focusable = true
	e.onBlur = fn
}

// HandleEvent dispatches an event to this element's handler.
// Returns true if the event was consumed.
func (e *Element) HandleEvent(event Event) bool {
	debug.Log("Element.HandleEvent: event=%T text=%q focusable=%v onClick=%v", event, e.text, e.focusable, e.onClick != nil)

	// First, let user handler try to consume the event (legacy bool-returning handler)
	if e.onEvent != nil {
		debug.Log("Element.HandleEvent: calling onEvent handler")
		if e.onEvent(e, event) {
			debug.Log("Element.HandleEvent: onEvent consumed event")
			return true
		}
	}

	// Handle key events with bubbling support
	if keyEvent, ok := event.(KeyEvent); ok {
		debug.Log("Element.HandleEvent: KeyEvent key=%d rune=%c", keyEvent.Key, keyEvent.Rune)
		if e.onKeyPress != nil {
			debug.Log("Element.HandleEvent: calling onKeyPress handler")
			if e.onKeyPress(e, keyEvent) {
				debug.Log("Element.HandleEvent: onKeyPress consumed event")
				return true
			}
		}

		// Trigger onClick on Enter or Space when focused
		if e.onClick != nil && (keyEvent.Key == KeyEnter || keyEvent.Rune == ' ') {
			debug.Log("Element.HandleEvent: triggering onClick via Enter/Space")
			e.onClick(e)
			return true
		}

		// Bubble to parent (same pattern as mouse events)
		if e.parent != nil {
			debug.Log("Element.HandleEvent: bubbling key event to parent")
			return e.parent.HandleEvent(event)
		}
	}

	// Handle MouseEvent - trigger onClick for left click press
	// Bubbles up to parent elements if this element doesn't handle it
	if mouseEvent, ok := event.(MouseEvent); ok {
		debug.Log("Element.HandleEvent: MouseEvent button=%d action=%d x=%d y=%d", mouseEvent.Button, mouseEvent.Action, mouseEvent.X, mouseEvent.Y)
		if mouseEvent.Button == MouseLeft && mouseEvent.Action == MousePress {
			if e.onClick != nil {
				debug.Log("Element.HandleEvent: triggering onClick via mouse click")
				e.onClick(e)
				return true
			}
			// Bubble up to parent if we didn't handle it
			if e.parent != nil {
				debug.Log("Element.HandleEvent: bubbling mouse event to parent")
				return e.parent.HandleEvent(event)
			}
		}
		// Bubble wheel events up to parent (for scrollable containers)
		if mouseEvent.Button == MouseWheelUp || mouseEvent.Button == MouseWheelDown {
			if e.parent != nil {
				debug.Log("Element.HandleEvent: bubbling wheel event to parent")
				return e.parent.HandleEvent(event)
			}
		}
		return false
	}

	// Handle scroll events for scrollable elements
	if e.scrollMode != ScrollNone {
		if e.handleScrollEvent(event) {
			return true
		}
	}

	debug.Log("Element.HandleEvent: event not consumed")
	return false
}

// handleScrollEvent handles keyboard and mouse wheel events for scrolling.
func (e *Element) handleScrollEvent(event Event) bool {
	// Handle mouse wheel events
	if mouse, ok := event.(MouseEvent); ok {
		switch mouse.Button {
		case MouseWheelUp:
			if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
				e.ScrollBy(0, -1)
				return true
			}
		case MouseWheelDown:
			if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
				e.ScrollBy(0, 1)
				return true
			}
		}
		return false
	}

	key, ok := event.(KeyEvent)
	if !ok {
		return false
	}

	_, viewportHeight := e.ViewportSize()

	switch key.Key {
	case KeyUp:
		if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
			e.ScrollBy(0, -1)
			return true
		}
	case KeyDown:
		if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
			e.ScrollBy(0, 1)
			return true
		}
	case KeyLeft:
		if e.scrollMode == ScrollHorizontal || e.scrollMode == ScrollBoth {
			e.ScrollBy(-1, 0)
			return true
		}
	case KeyRight:
		if e.scrollMode == ScrollHorizontal || e.scrollMode == ScrollBoth {
			e.ScrollBy(1, 0)
			return true
		}
	case KeyPageUp:
		if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
			e.ScrollBy(0, -viewportHeight)
			return true
		}
	case KeyPageDown:
		if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
			e.ScrollBy(0, viewportHeight)
			return true
		}
	case KeyHome:
		e.ScrollTo(0, 0)
		return true
	case KeyEnd:
		e.ScrollToBottom()
		return true
	}

	return false
}
