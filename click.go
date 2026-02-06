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
