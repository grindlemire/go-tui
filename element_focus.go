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
// Only handles scroll events for scrollable elements.
// Returns true if the event was consumed.
func (e *Element) HandleEvent(event Event) bool {
	debug.Log("Element.HandleEvent: event=%T text=%q scrollMode=%v", event, e.text, e.scrollMode)

	// Handle scroll events for scrollable elements
	if e.scrollMode != ScrollNone {
		if e.handleScrollEvent(event) {
			return true
		}
	}

	debug.Log("Element.HandleEvent: event not consumed")
	return false
}

// ContainsPoint returns true if the point (x, y) is within the element's bounds.
// This is useful for hit testing in HandleMouse implementations.
func (e *Element) ContainsPoint(x, y int) bool {
	return e.layout.Rect.Contains(x, y)
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
