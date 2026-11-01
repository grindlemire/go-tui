package tui

import (
	"reflect"
	"slices"

	"github.com/grindlemire/go-tui/internal/debug"
)

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

// setFocusIndex focuses the element at the given index. Used by modal to
// restore focus after close. The Blur/Focus calls here target stale elements
// from the previous render tree, but are harmless; refreshFromTree runs
// immediately after and finds the same element in the fresh tree.
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
	startIdx := max(f.current, 0)

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
// element tree. Re-renders produce new Element objects, so the focused one is
// found again by its focusTarget; if it is gone, it gets its Blur and nothing
// is focused.
func (f *focusManager) refreshFromTree(root *Element) {
	if root == nil {
		return
	}
	prev := f.Focused()
	target, known := f.currentTarget()
	savedIdx := f.current
	f.elements = f.elements[:0]
	root.WalkFocusables(func(elem Focusable) {
		f.elements = append(f.elements, elem)
	})
	f.current = -1
	switch {
	case known:
		f.current = f.find(target)
	case prev != nil && savedIdx < len(f.elements):
		// Not an *Element, so it has no target; keep its index.
		f.current = savedIdx
	}
	if f.current >= 0 {
		// The new element was just created by Render() and doesn't know
		// it's focused. Call Focus() to sync so that Blur() on the next
		// Tab press targets the correct element state.
		f.elements[f.current].Focus()
		return
	}
	if prev != nil {
		// Its owner (an unmounted Input, say) stops claiming focus.
		prev.Blur()
	}
	// On first refresh with no prior focus, check for autoFocus elements
	if !f.focusApplied {
		f.applyAutoFocus(root)
	}
}

// focusTarget names a focused element across re-renders: the component that
// rendered it, plus its path below that component when the path crosses a
// key (so it follows a reordered row), or else its position among that
// component's other unkeyed focusables. Scoping to the component keeps focus
// from landing on another page's element at the same position.
type focusTarget struct {
	owner   Component
	path    []pathSeg
	ordinal int
}

// focusOwner returns the nearest component that rendered el (nil for the
// root's own elements) and whether an element below that component has a key.
func focusOwner(el *Element) (owner Component, keyed bool) {
	for e := el; e != nil; e = e.parent {
		if e.component != nil {
			return e.component, keyed
		}
		keyed = keyed || (e.hasKey && e.parent != nil)
	}
	return nil, keyed
}

// pathBelowOwner is el's path from the element of the component that
// rendered it.
func pathBelowOwner(el *Element) []pathSeg {
	var path []pathSeg
	for e := el; e != nil && e.component == nil && e.parent != nil; e = e.parent {
		path = append(path, segmentOf(e))
	}
	slices.Reverse(path)
	return path
}

func sameOwner(a, b Component) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	// A struct value holding a slice or map never equals itself, so match
	// it by type; focus then keeps its position among such siblings.
	if v := reflect.ValueOf(a); v.Kind() == reflect.Struct && !v.Comparable() {
		return v.Type() == reflect.TypeOf(b)
	}
	return sameInstance(a, b)
}

// currentTarget describes the focused element; ok is false when nothing is
// focused or the focused Focusable is not an *Element.
func (f *focusManager) currentTarget() (focusTarget, bool) {
	el, ok := f.Focused().(*Element)
	if !ok {
		return focusTarget{}, false
	}
	owner, keyed := focusOwner(el)
	t := focusTarget{owner: owner}
	if keyed {
		t.path = pathBelowOwner(el)
		return t, true
	}
	for _, e := range f.elements[:f.current] {
		if o, ok := e.(*Element); ok {
			if oo, k := focusOwner(o); !k && sameOwner(oo, owner) {
				t.ordinal++
			}
		}
	}
	return t, true
}

// find returns the index of the element t names, or -1. A keyed path that
// no longer matches exactly (a line was added above the list) falls back to
// its keys alone when exactly one element shares them.
func (f *focusManager) find(t focusTarget) int {
	n, loose, looseHits := 0, -1, 0
	tail := keyedTail(t.path)
	for i, e := range f.elements {
		el, ok := e.(*Element)
		if !ok {
			continue
		}
		owner, keyed := focusOwner(el)
		switch {
		case !sameOwner(owner, t.owner):
		case t.path != nil:
			if !keyed {
				break
			}
			p := pathBelowOwner(el)
			if slices.Equal(p, t.path) {
				return i
			}
			if slices.Equal(keyedTail(p), tail) {
				loose, looseHits = i, looseHits+1
			}
		case keyed:
		case n == t.ordinal:
			return i
		default:
			n++
		}
	}
	if looseHits == 1 {
		return loose
	}
	return -1
}

// keyedTail keeps a path's keyed segments and everything below the deepest
// key, dropping the sibling indexes above it.
func keyedTail(path []pathSeg) []pathSeg {
	last := -1
	for i, s := range path {
		if s.keyed {
			last = i
		}
	}
	var tail []pathSeg
	for i, s := range path {
		if s.keyed || i > last {
			tail = append(tail, s)
		}
	}
	return tail
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
		return slices.ContainsFunc(e.children, walk)
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
