package tui

import (
	"github.com/grindlemire/go-tui/internal/debug"
)

// --- Focus API ---

// IsFocusable returns whether this element can receive focus.
func (e *Element) IsFocusable() bool {
	return e.focusable
}

// IsTabStop returns whether this element participates in Tab/Shift+Tab navigation.
func (e *Element) IsTabStop() bool {
	return e.tabStop
}

// IsFocused returns whether this element currently has focus.
func (e *Element) IsFocused() bool {
	return e.focused
}

// Focus marks this element as focused and calls onFocus callback if set.
// Idempotent: no-op if already focused.
// Does not cascade to children — only the FocusManager target receives focus.
func (e *Element) Focus() {
	if e.focused {
		debug.Log("Element.Focus: already focused, text=%q, skipping", e.text)
		return
	}
	debug.Log("Element.Focus: text=%q, hasOnFocus=%v", e.text, e.onFocus != nil)
	e.focused = true
	if e.onFocus != nil {
		e.onFocus(e)
	}
}

// Blur marks this element as not focused and calls onBlur callback if set.
// Idempotent: no-op if already blurred.
// Does not cascade to children — only the FocusManager target loses focus.
func (e *Element) Blur() {
	if !e.focused {
		debug.Log("Element.Blur: already blurred, text=%q, skipping", e.text)
		return
	}
	debug.Log("Element.Blur: text=%q, hasOnBlur=%v", e.text, e.onBlur != nil)
	e.focused = false
	if e.onBlur != nil {
		e.onBlur(e)
	}
}

// SetFocusable sets whether this element can receive focus.
// Also sets tabStop to the same value.
func (e *Element) SetFocusable(focusable bool) {
	e.focusable = focusable
	e.tabStop = focusable
}

// SetOnFocus sets a handler that's called when this element gains focus.
// The handler receives the element as its first parameter (self-inject).
// Implicitly sets focusable = true and tabStop = true.
func (e *Element) SetOnFocus(fn func(*Element)) {
	e.focusable = true
	e.tabStop = true
	e.onFocus = fn
}

// SetOnBlur sets a handler that's called when this element loses focus.
// The handler receives the element as its first parameter (self-inject).
// Implicitly sets focusable = true and tabStop = true.
func (e *Element) SetOnBlur(fn func(*Element)) {
	e.focusable = true
	e.tabStop = true
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

// hasWrapOverflow returns true if this element has wrapped text that overflows
// its content area, enabling auto-scroll with no visible scrollbar.
func (e *Element) hasWrapOverflow() bool {
	if e.scrollMode != ScrollNone || e.text == "" || e.noWrap {
		return false
	}
	return e.contentHeight > 0 && e.contentHeight > e.ContentRect().Height
}

// scrollWrapOverflow adjusts scrollY for wrap-overflow elements.
// This bypasses the normal ScrollTo/ScrollBy which require scrollMode.
func (e *Element) scrollWrapOverflow(dy int) {
	cr := e.ContentRect()
	maxY := e.contentHeight - cr.Height
	if maxY < 0 {
		maxY = 0
	}
	newY := e.scrollY + dy
	if newY < 0 {
		newY = 0
	}
	if newY > maxY {
		newY = maxY
	}
	if newY != e.scrollY {
		e.scrollY = newY
		e.MarkDirty()
	}
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
			// Auto-scroll for wrap-overflow text
			if e.hasWrapOverflow() {
				e.scrollWrapOverflow(-1)
				return true
			}
		case MouseWheelDown:
			if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
				e.ScrollBy(0, 1)
				return true
			}
			// Auto-scroll for wrap-overflow text
			if e.hasWrapOverflow() {
				e.scrollWrapOverflow(1)
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
