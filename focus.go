package tui

import "github.com/grindlemire/go-tui/internal/debug"

// Focusable is implemented by elements that can receive keyboard focus.
// Element implements this interface directly. For custom focus handling,
// use WithOnFocus, WithOnBlur, or WithOnEvent options on Element.
type Focusable interface {
	// IsFocusable returns whether this element can currently receive focus.
	// May return false for disabled elements.
	IsFocusable() bool

	// IsTabStop returns whether this element participates in Tab/Shift+Tab
	// navigation. Elements like scrollable containers are focusable (can
	// receive keyboard events) but are not tab stops by default.
	IsTabStop() bool

	// HandleEvent processes a keyboard event.
	// Returns true if the event was consumed, false to allow propagation.
	HandleEvent(event Event) bool

	// Focus is called when this element gains focus.
	// Implementations typically update visual state (e.g., highlight border).
	Focus()

	// Blur is called when this element loses focus.
	// Implementations typically revert visual state.
	Blur()
}

// focusManager tracks focus state for a set of focusable elements.
// It does NOT automatically handle Tab navigation; the user controls
// when focus moves by calling Next(), Prev(), or SetFocus().
type focusManager struct {
	elements     []Focusable // Registered focusable elements in order
	current      int         // Index of currently focused element (-1 = none)
	focusApplied bool        // true after focus has been set (prevents re-applying autoFocus)
	scope        *Element    // if set, only elements within this subtree are navigable
}

// newFocusManager creates an empty focusManager.
// Use Register to add focusable elements.
func newFocusManager() *focusManager {
	return &focusManager{
		current: -1,
	}
}

// Register adds a focusable element to the manager.
// Does not auto-focus; use Tab or SetFocus to focus an element.
func (f *focusManager) Register(elem Focusable) {
	debug.Log("FocusManager.Register: adding element %T (focusable=%v)", elem, elem.IsFocusable())
	f.elements = append(f.elements, elem)
	debug.Log("FocusManager.Register: total elements=%d, current=%d", len(f.elements), f.current)
}

// Unregister removes a focusable element from the manager.
func (f *focusManager) Unregister(elem Focusable) {
	// Find the element
	idx := -1
	for i, e := range f.elements {
		if e == elem {
			idx = i
			break
		}
	}
	if idx == -1 {
		return // Not found
	}

	// Check if we're removing the focused element
	wasCurrentlyFocused := idx == f.current

	// Call Blur if this element was focused
	if wasCurrentlyFocused {
		elem.Blur()
	}

	// Remove from slice
	f.elements = append(f.elements[:idx], f.elements[idx+1:]...)

	// Adjust current index
	if len(f.elements) == 0 {
		f.current = -1
	} else if wasCurrentlyFocused {
		// Focus the next element (or wrap to beginning)
		f.current = idx
		if f.current >= len(f.elements) {
			f.current = 0
		}
		// Find next focusable element
		f.focusNextFrom(f.current)
	} else if idx < f.current {
		// Shift current down since an element before it was removed
		f.current--
	}
}

// Focused returns the currently focused element, or nil if none.
func (f *focusManager) Focused() Focusable {
	if f.current < 0 || f.current >= len(f.elements) {
		return nil
	}
	return f.elements[f.current]
}

// IsFocused returns true if the given Focusable is the currently focused element.
func (f *focusManager) IsFocused(elem Focusable) bool {
	if f.current < 0 || f.current >= len(f.elements) {
		return false
	}
	return f.elements[f.current] == elem
}

// focusedIndex returns the current focus index, or -1 if nothing is focused.
func (f *focusManager) focusedIndex() int {
	return f.current
}

// setFocusIndex restores focus to the element at the given index.
// Used by modal to restore focus after close. The Blur/Focus calls here
// target stale elements from the previous render tree, but are harmless;
// refreshFromTree runs immediately after and calls Focus on the fresh
// element at the preserved index.
func (f *focusManager) setFocusIndex(idx int) {
	if idx < 0 || idx >= len(f.elements) {
		return
	}
	if f.current >= 0 && f.current < len(f.elements) && f.current != idx {
		f.elements[f.current].Blur()
	}
	f.current = idx
	f.focusApplied = true
	f.elements[idx].Focus()
}

// SetFocus moves focus to the specified element.
// Does nothing if the element is not registered or not focusable.
func (f *focusManager) SetFocus(elem Focusable) {
	// Find the element
	idx := -1
	for i, e := range f.elements {
		if e == elem {
			idx = i
			break
		}
	}
	if idx == -1 || !elem.IsFocusable() {
		return // Not found or not focusable
	}

	// Blur current if different
	if f.current >= 0 && f.current < len(f.elements) && f.current != idx {
		f.elements[f.current].Blur()
	}

	// Focus new element
	f.current = idx
	f.focusApplied = true
	elem.Focus()
}

// Next moves focus to the next focusable element.
// Wraps around to the first element if at the end.
// Does nothing if there are no focusable elements.
func (f *focusManager) Next() {
	if len(f.elements) == 0 {
		return
	}

	// Blur current
	if f.current >= 0 && f.current < len(f.elements) {
		f.elements[f.current].Blur()
	}

	// Find next tab-stop element, skipping the current one.
	// If no other tab-stop exists, clear focus instead of re-focusing the same element.
	startIdx := f.current
	if startIdx < 0 {
		startIdx = -1
	}

	for i := 0; i < len(f.elements); i++ {
		nextIdx := (startIdx + 1 + i) % len(f.elements)
		if nextIdx == f.current {
			// Wrapped back to the same element; clear focus instead
			f.current = -1
			return
		}
		if f.elements[nextIdx].IsTabStop() && f.isInScope(f.elements[nextIdx]) {
			f.current = nextIdx
			f.elements[nextIdx].Focus()
			return
		}
	}

	// No focusable elements found
	f.current = -1
}

// Prev moves focus to the previous focusable element.
// If no other tab-stop exists, clears focus.
func (f *focusManager) Prev() {
	if len(f.elements) == 0 {
		return
	}

	// Blur current
	if f.current >= 0 && f.current < len(f.elements) {
		f.elements[f.current].Blur()
	}

	// Find previous tab-stop element, skipping the current one.
	startIdx := f.current
	if startIdx < 0 {
		startIdx = 0
	}

	for i := 0; i < len(f.elements); i++ {
		prevIdx := startIdx - 1 - i
		if prevIdx < 0 {
			prevIdx += len(f.elements)
		}
		if prevIdx == f.current {
			// Wrapped back to the same element; clear focus instead
			f.current = -1
			return
		}
		if f.elements[prevIdx].IsTabStop() && f.isInScope(f.elements[prevIdx]) {
			f.current = prevIdx
			f.elements[prevIdx].Focus()
			return
		}
	}

	// No focusable elements found
	f.current = -1
}

// ScopeTo restricts Tab/Shift+Tab navigation to focusable elements
// within the given element's subtree. Used by modal focus trapping.
func (f *focusManager) ScopeTo(el *Element) {
	f.scope = el
}

// ClearScope removes focus navigation restriction.
func (f *focusManager) ClearScope() {
	f.scope = nil
}

// isInScope returns true if the focusable element is within the current scope.
// If no scope is set, all elements are in scope.
func (f *focusManager) isInScope(elem Focusable) bool {
	if f.scope == nil {
		return true
	}
	el, ok := elem.(*Element)
	if !ok {
		return false
	}
	for e := el; e != nil; e = e.parent {
		if e == f.scope {
			return true
		}
	}
	return false
}

// ClearFocus blurs the currently focused element and sets focus to none.
func (f *focusManager) ClearFocus() {
	if f.current >= 0 && f.current < len(f.elements) {
		f.elements[f.current].Blur()
	}
	f.current = -1
}

// Dispatch sends an event to the currently focused element.
// Returns true if the event was handled.
func (f *focusManager) Dispatch(event Event) bool {
	focused := f.Focused()
	debug.Log("FocusManager.Dispatch: event=%T focused=%v (current=%d, total=%d)", event, focused != nil, f.current, len(f.elements))
	if focused == nil {
		debug.Log("FocusManager.Dispatch: no focused element, returning false")
		return false
	}
	result := focused.HandleEvent(event)
	debug.Log("FocusManager.Dispatch: HandleEvent returned %v", result)
	return result
}

// refreshFromTree rebuilds the focusable element list from the current
// element tree while preserving the focus index. This is needed because
// re-renders produce new Element objects, making old references stale.
func (f *focusManager) refreshFromTree(root *Element) {
	if root == nil {
		return
	}
	savedIdx := f.current
	f.elements = f.elements[:0]
	root.WalkFocusables(func(elem Focusable) {
		f.elements = append(f.elements, elem)
	})
	if savedIdx >= 0 && savedIdx < len(f.elements) {
		f.current = savedIdx
		// The new element was just created by Render() and doesn't know
		// it's focused. Call Focus() to sync so that Blur() on the next
		// Tab press targets the correct element state.
		f.elements[savedIdx].Focus()
	} else {
		f.current = -1
		// On first refresh with no prior focus, check for autoFocus elements
		if !f.focusApplied {
			f.applyAutoFocus(root)
		}
	}
}

// applyAutoFocus finds the first autoFocus element in the tree and sets
// focus on it. Only called when no focus has been applied yet.
func (f *focusManager) applyAutoFocus(root *Element) {
	if root == nil {
		return
	}
	var target Focusable
	var walk func(e *Element) bool
	walk = func(e *Element) bool {
		if e.hidden {
			return false
		}
		if e.autoFocus && e.IsTabStop() {
			target = e
			return true
		}
		for _, child := range e.children {
			if walk(child) {
				return true
			}
		}
		return false
	}
	walk(root)
	if target != nil {
		f.SetFocus(target)
	}
}

// focusNextFrom finds the next tab-stop element starting from idx.
// Used internally after unregister.
func (f *focusManager) focusNextFrom(startIdx int) {
	for i := 0; i < len(f.elements); i++ {
		idx := (startIdx + i) % len(f.elements)
		if f.elements[idx].IsTabStop() && f.isInScope(f.elements[idx]) {
			f.current = idx
			f.elements[idx].Focus()
			return
		}
	}
	// No focusable element found
	f.current = -1
}
