package tui

// ClickBinding represents a ref-to-function binding for mouse clicks.
type ClickBinding struct {
	Ref *Ref
	Fn  func()
}

// Click creates a click binding for use with HandleClicks.
func Click(ref *Ref, fn func()) ClickBinding {
	return ClickBinding{Ref: ref, Fn: fn}
}

// HandleClicks checks a mouse event against click bindings and calls
// the first matching handler. Returns true if a click was handled.
// Use this in HandleMouse to simplify ref-based hit testing.
//
// Example:
//
//	func (c *counter) HandleMouse(me tui.MouseEvent) bool {
//	    return tui.HandleClicks(me,
//	        tui.Click(c.incrementBtn, c.increment),
//	        tui.Click(c.decrementBtn, c.decrement),
//	    )
//	}
func HandleClicks(me MouseEvent, bindings ...ClickBinding) bool {
	if me.Button != MouseLeft || me.Action != MousePress {
		return false
	}

	for _, b := range bindings {
		if b.Ref.El() != nil && b.Ref.El().ContainsPoint(me.X, me.Y) {
			b.Fn()
			return true
		}
	}

	return false
}
