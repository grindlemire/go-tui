package tui

import (
	"testing"
)

// mockFocusable is a mock implementation of Focusable for testing.
type mockFocusable struct {
	id         string
	focusable  bool
	focused    bool
	focusCalls int
	blurCalls  int
	lastEvent  Event
	handled    bool
}

func newMockFocusable(id string, focusable bool) *mockFocusable {
	return &mockFocusable{
		id:        id,
		focusable: focusable,
	}
}

func (m *mockFocusable) IsFocusable() bool {
	return m.focusable
}

func (m *mockFocusable) IsTabStop() bool {
	return m.focusable
}

func (m *mockFocusable) HandleEvent(event Event) bool {
	m.lastEvent = event
	return m.handled
}

func (m *mockFocusable) Focus() {
	m.focused = true
	m.focusCalls++
}

func (m *mockFocusable) Blur() {
	m.focused = false
	m.blurCalls++
}

func (m *mockFocusable) IsFocused() bool {
	return m.focused
}

// registerAll registers all elements to the focusManager.
func registerAll(fm *focusManager, elements ...*mockFocusable) {
	for _, elem := range elements {
		fm.Register(elem)
	}
}

func TestFocusManager_IsFocused(t *testing.T) {
	type tc struct {
		elements      []*mockFocusable
		focusIndex    int // which element to SetFocus on (-1 for none)
		checkIndex    int // which element to check IsFocused (-1 for unregistered)
		expectFocused bool
	}

	tests := map[string]tc{
		"focused element returns true": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
			},
			focusIndex:    0,
			checkIndex:    0,
			expectFocused: true,
		},
		"unfocused element returns false": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
			},
			focusIndex:    0,
			checkIndex:    1,
			expectFocused: false,
		},
		"no focus returns false": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
			},
			focusIndex:    -1,
			checkIndex:    0,
			expectFocused: false,
		},
		"unregistered element returns false": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
			},
			focusIndex:    0,
			checkIndex:    -1,
			expectFocused: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := newFocusManager()
			registerAll(fm, tt.elements...)

			if tt.focusIndex >= 0 {
				fm.SetFocus(tt.elements[tt.focusIndex])
			}

			var check Focusable
			if tt.checkIndex >= 0 {
				check = tt.elements[tt.checkIndex]
			} else {
				check = newMockFocusable("unregistered", true)
			}

			got := fm.IsFocused(check)
			if got != tt.expectFocused {
				t.Errorf("IsFocused() = %v, want %v", got, tt.expectFocused)
			}
		})
	}
}

func TestFocusManager_NoAutoFocus(t *testing.T) {
	type tc struct {
		elements []*mockFocusable
	}

	tests := map[string]tc{
		"single focusable element": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
			},
		},
		"multiple focusable elements": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := newFocusManager()
			registerAll(fm, tt.elements...)

			if fm.Focused() != nil {
				t.Error("Focused() should return nil after registration (no auto-focus)")
			}

			for _, elem := range tt.elements {
				if elem.focusCalls != 0 {
					t.Errorf("Element %q should not have Focus() called, got %d calls", elem.id, elem.focusCalls)
				}
			}
		})
	}
}

func TestNewFocusManager_NoFocusableElements(t *testing.T) {
	type tc struct {
		elements []*mockFocusable
	}

	tests := map[string]tc{
		"empty": {
			elements: []*mockFocusable{},
		},
		"all non-focusable": {
			elements: []*mockFocusable{
				newMockFocusable("a", false),
				newMockFocusable("b", false),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := newFocusManager()
			registerAll(fm, tt.elements...)

			if fm.Focused() != nil {
				t.Error("Focused() should return nil when no focusable elements")
			}
		})
	}
}

func TestFocusManager_Next(t *testing.T) {
	type tc struct {
		elements          []*mockFocusable
		nextCalls         int
		expectedFocusedID string
	}

	tests := map[string]tc{
		"first Next focuses first element": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			nextCalls:         1,
			expectedFocusedID: "a",
		},
		"next from first to second": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			nextCalls:         2,
			expectedFocusedID: "b",
		},
		"wraps to beginning": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
			},
			nextCalls:         3,
			expectedFocusedID: "a",
		},
		"skips non-focusable": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", false),
				newMockFocusable("c", true),
			},
			nextCalls:         2,
			expectedFocusedID: "c",
		},
		"full cycle through all": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			nextCalls:         4,
			expectedFocusedID: "a",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := newFocusManager()
			registerAll(fm, tt.elements...)

			for i := 0; i < tt.nextCalls; i++ {
				fm.Next()
			}

			focused := fm.Focused()
			if focused == nil {
				t.Fatal("Focused() returned nil")
			}

			mf := focused.(*mockFocusable)
			if mf.id != tt.expectedFocusedID {
				t.Errorf("After %d Next() calls, focused = %q, want %q", tt.nextCalls, mf.id, tt.expectedFocusedID)
			}
		})
	}
}

func TestFocusManager_Prev(t *testing.T) {
	type tc struct {
		elements          []*mockFocusable
		prevCalls         int
		expectedFocusedID string
	}

	tests := map[string]tc{
		"first Prev wraps to last": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			prevCalls:         1,
			expectedFocusedID: "c",
		},
		"prev twice from none": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			prevCalls:         2,
			expectedFocusedID: "b",
		},
		"skips non-focusable backward": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", false),
				newMockFocusable("c", true),
			},
			prevCalls:         2,
			expectedFocusedID: "a",
		},
		"full cycle backwards": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			prevCalls:         4,
			expectedFocusedID: "c",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := newFocusManager()
			registerAll(fm, tt.elements...)

			for i := 0; i < tt.prevCalls; i++ {
				fm.Prev()
			}

			focused := fm.Focused()
			if focused == nil {
				t.Fatal("Focused() returned nil")
			}

			mf := focused.(*mockFocusable)
			if mf.id != tt.expectedFocusedID {
				t.Errorf("After %d Prev() calls, focused = %q, want %q", tt.prevCalls, mf.id, tt.expectedFocusedID)
			}
		})
	}
}

func TestFocusManager_SetFocus(t *testing.T) {
	type tc struct {
		elements          []*mockFocusable
		focusIndex        int
		expectedFocusedID string
	}

	tests := map[string]tc{
		"set focus to second": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			focusIndex:        1,
			expectedFocusedID: "b",
		},
		"set focus to last": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			focusIndex:        2,
			expectedFocusedID: "c",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := newFocusManager()
			registerAll(fm, tt.elements...)

			fm.SetFocus(tt.elements[tt.focusIndex])

			focused := fm.Focused()
			if focused == nil {
				t.Fatal("Focused() returned nil")
			}

			mf := focused.(*mockFocusable)
			if mf.id != tt.expectedFocusedID {
				t.Errorf("SetFocus() focused = %q, want %q", mf.id, tt.expectedFocusedID)
			}

			// Verify focus state
			if !mf.focused {
				t.Error("SetFocus() target should have focused=true")
			}
		})
	}
}

func TestFocusManager_SetFocusNonFocusable(t *testing.T) {
	a := newMockFocusable("a", true)
	b := newMockFocusable("b", false) // Not focusable

	fm := newFocusManager()
	fm.Register(a)
	fm.Register(b)

	// Try to focus non-focusable element
	fm.SetFocus(b)

	// Nothing should be focused (no auto-focus, SetFocus on non-focusable is no-op)
	if fm.Focused() != nil {
		t.Error("SetFocus() on non-focusable should not set focus")
	}
}

func TestFocusManager_Register(t *testing.T) {
	type tc struct {
		initialElements []*mockFocusable
		registerElement *mockFocusable
	}

	tests := map[string]tc{
		"register to empty manager does not auto-focus": {
			initialElements: []*mockFocusable{},
			registerElement: newMockFocusable("new", true),
		},
		"register to existing does not change focus": {
			initialElements: []*mockFocusable{
				newMockFocusable("a", true),
			},
			registerElement: newMockFocusable("new", true),
		},
		"register non-focusable to empty does not focus": {
			initialElements: []*mockFocusable{},
			registerElement: newMockFocusable("new", false),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := newFocusManager()
			registerAll(fm, tt.initialElements...)
			fm.Register(tt.registerElement)

			// Register never auto-focuses
			if fm.Focused() != nil {
				t.Error("Register() should not auto-focus any element")
			}
		})
	}
}

