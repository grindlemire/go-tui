package tui

// --- Element's own API ---

// AddChild appends children to this Element.
// Notifies root's onChildAdded callback for each child.
func (e *Element) AddChild(children ...*Element) {
	for _, child := range children {
		child.parent = e
		child.setAppRecursive(e.app)
		e.children = append(e.children, child)
		e.notifyChildAdded(child)
	}
	e.MarkDirty()
}

// notifyChildAdded walks up to root and calls appropriate callbacks.
func (e *Element) notifyChildAdded(child *Element) {
	root := e
	for root.parent != nil {
		root = root.parent
	}
	if root.onChildAdded != nil {
		root.onChildAdded(child)
	}
	// Notify App about focusable elements for auto-registration
	if root.onFocusableAdded != nil && child.IsFocusable() {
		root.onFocusableAdded(child)
	}
}

// SetOnChildAdded sets the callback for when any descendant is added.
func (e *Element) SetOnChildAdded(fn func(*Element)) {
	e.onChildAdded = fn
}

// RemoveChild removes a child from this Element.
// Returns true if the child was found and removed.
func (e *Element) RemoveChild(child *Element) bool {
	for i, c := range e.children {
		if c == child {
			// Remove by swapping with last element and truncating
			e.children[i] = e.children[len(e.children)-1]
			e.children = e.children[:len(e.children)-1]
			child.parent = nil
			child.setAppRecursive(nil)
			e.MarkDirty()
			return true
		}
	}
	return false
}

// RemoveAllChildren removes all children from this Element.
// Automatically marks dirty.
func (e *Element) RemoveAllChildren() {
	for _, child := range e.children {
		child.parent = nil
		child.setAppRecursive(nil)
	}
	e.children = nil
	e.MarkDirty()
}

// Children returns the child elements.
func (e *Element) Children() []*Element {
	return e.children
}

// Parent returns the parent element, or nil if this is the root.
func (e *Element) Parent() *Element {
	return e.parent
}

func (e *Element) setAppRecursive(app *App) {
	if e == nil {
		return
	}
	e.app = app
	for _, child := range e.children {
		child.setAppRecursive(app)
	}
}
