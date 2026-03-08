package tui

import "github.com/grindlemire/go-tui/internal/debug"

// Focusable is implemented by elements that can receive keyboard focus.
// Element implements this interface directly. For custom focus handling,
// use WithOnFocus, WithOnBlur, or WithOnEvent options on Element.
type Focusable interface {
	// IsFocusable returns whether this element can currently receive focus.
	// May return false for disabled elements.
	IsFocusable() bool

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
	elements []Focusable // Registered focusable elements in order
	current  int         // Index of currently focused element (-1 = none)
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

	// Find next focusable element
	startIdx := f.current
	if startIdx < 0 {
		startIdx = -1
	}

	for i := 0; i < len(f.elements); i++ {
		nextIdx := (startIdx + 1 + i) % len(f.elements)
		if f.elements[nextIdx].IsFocusable() {
			f.current = nextIdx
			f.elements[nextIdx].Focus()
			return
		}
	}

	// No focusable elements found
	f.current = -1
}

// Prev moves focus to the previous focusable element.
// Wraps around to the last element if at the beginning.
func (f *focusManager) Prev() {
	if len(f.elements) == 0 {
		return
	}

	// Blur current
	if f.current >= 0 && f.current < len(f.elements) {
		f.elements[f.current].Blur()
	}

	// Find previous focusable element
	startIdx := f.current
	if startIdx < 0 {
		startIdx = 0
	}

	for i := 0; i < len(f.elements); i++ {
		prevIdx := startIdx - 1 - i
		if prevIdx < 0 {
			prevIdx += len(f.elements)
		}
		if f.elements[prevIdx].IsFocusable() {
			f.current = prevIdx
			f.elements[prevIdx].Focus()
			return
		}
	}

	// No focusable elements found
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

// focusNextFrom finds the next focusable element starting from idx.
// Used internally after unregister.
func (f *focusManager) focusNextFrom(startIdx int) {
	for i := 0; i < len(f.elements); i++ {
		idx := (startIdx + i) % len(f.elements)
		if f.elements[idx].IsFocusable() {
			f.current = idx
			f.elements[idx].Focus()
			return
		}
	}
	// No focusable element found
	f.current = -1
}
