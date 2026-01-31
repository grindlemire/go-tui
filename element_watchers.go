package tui

// --- Focus Tree Discovery API ---

// SetOnFocusableAdded sets a callback called when a focusable descendant is added.
// This is used by App to auto-register focusable elements.
func (e *Element) SetOnFocusableAdded(fn func(Focusable)) {
	e.onFocusableAdded = fn
}

// WalkFocusables calls fn for each focusable element in the tree.
// This is used by App to discover existing focusable elements.
func (e *Element) WalkFocusables(fn func(Focusable)) {
	if e.IsFocusable() {
		fn(e)
	}
	for _, child := range e.children {
		child.WalkFocusables(fn)
	}
}

// --- OnUpdate Hook API ---

// SetOnUpdate sets a function called before each render.
// Useful for polling channels, updating animations, etc.
func (e *Element) SetOnUpdate(fn func()) {
	e.onUpdate = fn
}

// --- Watcher API ---

// AddWatcher attaches a watcher (timer, channel watcher) to this element.
// Watchers are started automatically when the element tree is set as app root.
func (e *Element) AddWatcher(w Watcher) {
	e.watchers = append(e.watchers, w)
}

// Watchers returns the watchers attached to this element.
func (e *Element) Watchers() []Watcher {
	return e.watchers
}

// WalkWatchers calls fn for each watcher in the element tree.
// This is used by App.SetRoot to discover and start all watchers.
func (e *Element) WalkWatchers(fn func(Watcher)) {
	for _, w := range e.watchers {
		fn(w)
	}
	for _, child := range e.children {
		child.WalkWatchers(fn)
	}
}

// --- Hit Testing API ---

// ElementAt finds the deepest element containing the point (x, y).
// Returns nil if no element contains the point.
// Children are checked in reverse order since last child renders on top.
func (e *Element) ElementAt(x, y int) *Element {
	bounds := e.Rect()
	if !bounds.Contains(x, y) {
		return nil
	}

	// Check children in reverse order (last child renders on top)
	for i := len(e.children) - 1; i >= 0; i-- {
		if hit := e.children[i].ElementAt(x, y); hit != nil {
			return hit
		}
	}

	// No child hit, this element is the target
	return e
}

// ElementAtPoint finds the deepest element containing the point (x, y).
// Returns nil if no element contains the point.
// This method returns Focusable to satisfy the mouseHitTester interface.
func (e *Element) ElementAtPoint(x, y int) Focusable {
	elem := e.ElementAt(x, y)
	if elem == nil {
		return nil
	}
	return elem
}
