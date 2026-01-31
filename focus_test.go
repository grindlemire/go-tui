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

// registerAll registers all elements to the FocusManager.
func registerAll(fm *FocusManager, elements ...*mockFocusable) {
	for _, elem := range elements {
		fm.Register(elem)
	}
}

func TestNewFocusManager_FocusesFirstElement(t *testing.T) {
	type tc struct {
		elements          []*mockFocusable
		expectedFocusedID string
	}

	tests := map[string]tc{
		"single focusable element": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
			},
			expectedFocusedID: "a",
		},
		"first of multiple focusable": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			expectedFocusedID: "a",
		},
		"skips non-focusable first element": {
			elements: []*mockFocusable{
				newMockFocusable("a", false),
				newMockFocusable("b", true),
			},
			expectedFocusedID: "b",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := NewFocusManager()
			registerAll(fm, tt.elements...)

			focused := fm.Focused()
			if focused == nil {
				t.Fatal("Focused() returned nil")
			}

			mf, ok := focused.(*mockFocusable)
			if !ok {
				t.Fatalf("Focused() returned wrong type: %T", focused)
			}

			if mf.id != tt.expectedFocusedID {
				t.Errorf("Focused element id = %q, want %q", mf.id, tt.expectedFocusedID)
			}

			if !mf.focused {
				t.Error("Focused element should have focused=true")
			}

			if mf.focusCalls != 1 {
				t.Errorf("Focus() calls = %d, want 1", mf.focusCalls)
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
			fm := NewFocusManager()
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
		"next from first to second": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			nextCalls:         1,
			expectedFocusedID: "b",
		},
		"wraps to beginning": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
			},
			nextCalls:         2,
			expectedFocusedID: "a",
		},
		"skips non-focusable": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", false),
				newMockFocusable("c", true),
			},
			nextCalls:         1,
			expectedFocusedID: "c",
		},
		"full cycle through all": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			nextCalls:         3,
			expectedFocusedID: "a",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := NewFocusManager()
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
		"prev from first wraps to last": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			prevCalls:         1,
			expectedFocusedID: "c",
		},
		"prev twice from first": {
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
			prevCalls:         1,
			expectedFocusedID: "c",
		},
		"full cycle backwards": {
			elements: []*mockFocusable{
				newMockFocusable("a", true),
				newMockFocusable("b", true),
				newMockFocusable("c", true),
			},
			prevCalls:         3,
			expectedFocusedID: "a",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := NewFocusManager()
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
			fm := NewFocusManager()
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

			// Verify previous element was blurred
			if tt.elements[0].id != tt.expectedFocusedID && !tt.elements[0].focused {
				// First element should have been blurred
				if tt.elements[0].blurCalls == 0 {
					t.Error("Previous focused element should have Blur() called")
				}
			}
		})
	}
}

func TestFocusManager_SetFocusNonFocusable(t *testing.T) {
	a := newMockFocusable("a", true)
	b := newMockFocusable("b", false) // Not focusable

	fm := NewFocusManager()
	fm.Register(a)
	fm.Register(b)

	// Try to focus non-focusable element
	fm.SetFocus(b)

	// Focus should remain on 'a'
	focused := fm.Focused()
	if focused == nil {
		t.Fatal("Focused() returned nil")
	}
	mf := focused.(*mockFocusable)
	if mf.id != "a" {
		t.Errorf("SetFocus() on non-focusable should not change focus, got %q", mf.id)
	}
}

func TestFocusManager_Register(t *testing.T) {
	type tc struct {
		initialElements   []*mockFocusable
		registerElement   *mockFocusable
		expectedFocusedID string
	}

	tests := map[string]tc{
		"register to empty manager focuses first": {
			initialElements:   []*mockFocusable{},
			registerElement:   newMockFocusable("new", true),
			expectedFocusedID: "new",
		},
		"register to existing does not change focus": {
			initialElements: []*mockFocusable{
				newMockFocusable("a", true),
			},
			registerElement:   newMockFocusable("new", true),
			expectedFocusedID: "a",
		},
		"register non-focusable to empty does not focus": {
			initialElements:   []*mockFocusable{},
			registerElement:   newMockFocusable("new", false),
			expectedFocusedID: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fm := NewFocusManager()
			registerAll(fm, tt.initialElements...)
			fm.Register(tt.registerElement)

			focused := fm.Focused()
			if tt.expectedFocusedID == "" {
				if focused != nil {
					t.Error("Expected no focused element")
				}
				return
			}

			if focused == nil {
				t.Fatal("Focused() returned nil")
			}
			mf := focused.(*mockFocusable)
			if mf.id != tt.expectedFocusedID {
				t.Errorf("After Register(), focused = %q, want %q", mf.id, tt.expectedFocusedID)
			}
		})
	}
}

