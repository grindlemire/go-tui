package tui

import (
	"testing"
)

func TestElement_ScrollTo(t *testing.T) {
	type tc struct {
		contentItems   int
		viewportHeight int
		scrollToY      int
		expectedY      int
	}

	tests := map[string]tc{
		"scroll within bounds": {
			contentItems:   20,
			viewportHeight: 10,
			scrollToY:      5,
			expectedY:      5,
		},
		"scroll clamped to max": {
			contentItems:   20,
			viewportHeight: 10,
			scrollToY:      100,
			expectedY:      10, // 20 items - 10 viewport = 10 max
		},
		"scroll clamped to zero": {
			contentItems:   20,
			viewportHeight: 10,
			scrollToY:      -5,
			expectedY:      0,
		},
		"no scroll needed when content fits": {
			contentItems:   5,
			viewportHeight: 10,
			scrollToY:      5,
			expectedY:      0, // Content fits, can't scroll
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(
				WithHeight(tt.viewportHeight),
				WithScrollable(ScrollVertical),
				WithDirection(Column),
			)

			for i := 0; i < tt.contentItems; i++ {
				e.AddChild(New(WithHeight(1)))
			}

			// Force layout calculation
			buf := NewBuffer(80, 25)
			e.Render(buf, 80, tt.viewportHeight)

			e.ScrollTo(0, tt.scrollToY)

			_, y := e.ScrollOffset()
			if y != tt.expectedY {
				t.Errorf("ScrollTo(%d) got scrollY=%d, want %d", tt.scrollToY, y, tt.expectedY)
			}
		})
	}
}

func TestElement_ScrollBy(t *testing.T) {
	type tc struct {
		initialY int
		deltaY   int
		expected int
	}

	tests := map[string]tc{
		"scroll down": {
			initialY: 0,
			deltaY:   3,
			expected: 3,
		},
		"scroll up": {
			initialY: 5,
			deltaY:   -2,
			expected: 3,
		},
		"scroll up clamped to zero": {
			initialY: 2,
			deltaY:   -10,
			expected: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(
				WithHeight(10),
				WithScrollable(ScrollVertical),
				WithDirection(Column),
			)
			for i := 0; i < 30; i++ {
				e.AddChild(New(WithHeight(1)))
			}

			buf := NewBuffer(80, 25)
			e.Render(buf, 80, 10)

			e.ScrollTo(0, tt.initialY)
			e.ScrollBy(0, tt.deltaY)

			_, y := e.ScrollOffset()
			if y != tt.expected {
				t.Errorf("ScrollBy(%d) from %d got scrollY=%d, want %d", tt.deltaY, tt.initialY, y, tt.expected)
			}
		})
	}
}

func TestElement_ContentSize(t *testing.T) {
	type tc struct {
		items    int
		expected int
	}

	tests := map[string]tc{
		"empty": {
			items:    0,
			expected: 0,
		},
		"single item": {
			items:    1,
			expected: 1,
		},
		"multiple items": {
			items:    25,
			expected: 25,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(
				WithHeight(10),
				WithScrollable(ScrollVertical),
				WithDirection(Column),
			)
			for i := 0; i < tt.items; i++ {
				e.AddChild(New(WithHeight(1)))
			}

			buf := NewBuffer(80, 25)
			e.Render(buf, 80, 10)

			_, h := e.ContentSize()
			if h != tt.expected {
				t.Errorf("ContentSize() got height=%d, want %d", h, tt.expected)
			}
		})
	}
}

func TestElement_ScrollEventHandling(t *testing.T) {
	type tc struct {
		key      Key
		initialY int
		expected int
	}

	tests := map[string]tc{
		"arrow down": {
			key:      KeyDown,
			initialY: 0,
			expected: 1,
		},
		"arrow up": {
			key:      KeyUp,
			initialY: 5,
			expected: 4,
		},
		"home": {
			key:      KeyHome,
			initialY: 10,
			expected: 0,
		},
		"end": {
			key:      KeyEnd,
			initialY: 0,
			expected: 20, // 30 items - 10 viewport = 20 max
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(
				WithHeight(10),
				WithScrollable(ScrollVertical),
				WithDirection(Column),
			)
			for i := 0; i < 30; i++ {
				e.AddChild(New(WithHeight(1)))
			}

			buf := NewBuffer(80, 25)
			e.Render(buf, 80, 10)

			e.ScrollTo(0, tt.initialY)

			event := KeyEvent{Key: tt.key}
			handled := e.HandleEvent(event)

			if !handled {
				t.Error("Event should have been handled")
			}
			_, y := e.ScrollOffset()
			if y != tt.expected {
				t.Errorf("Key %v from %d got scrollY=%d, want %d", tt.key, tt.initialY, y, tt.expected)
			}
		})
	}
}

func TestElement_IsScrollable(t *testing.T) {
	t.Run("not scrollable by default", func(t *testing.T) {
		e := New(WithHeight(10))
		if e.IsScrollable() {
			t.Error("Element should not be scrollable by default")
		}
	})

	t.Run("scrollable when enabled", func(t *testing.T) {
		e := New(WithHeight(10), WithScrollable(ScrollVertical))
		if !e.IsScrollable() {
			t.Error("Element should be scrollable when WithScrollable is used")
		}
	})
}

func TestElement_ScrollModes(t *testing.T) {
	type tc struct {
		mode         ScrollMode
		expectVert   bool
		expectHoriz  bool
	}

	tests := map[string]tc{
		"vertical only": {
			mode:        ScrollVertical,
			expectVert:  true,
			expectHoriz: false,
		},
		"horizontal only": {
			mode:        ScrollHorizontal,
			expectVert:  false,
			expectHoriz: true,
		},
		"both": {
			mode:        ScrollBoth,
			expectVert:  true,
			expectHoriz: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(
				WithSize(10, 10),
				WithScrollable(tt.mode),
			)

			// Add content that exceeds viewport in both directions
			for i := 0; i < 30; i++ {
				e.AddChild(New(WithSize(30, 1)))
			}

			buf := NewBuffer(80, 25)
			e.Render(buf, 10, 10)

			// Test vertical scroll
			e.ScrollTo(0, 0)
			downEvent := KeyEvent{Key: KeyDown}
			handledDown := e.HandleEvent(downEvent)
			if handledDown != tt.expectVert {
				t.Errorf("KeyDown handled=%v, want %v", handledDown, tt.expectVert)
			}

			// Test horizontal scroll
			e.ScrollTo(0, 0)
			rightEvent := KeyEvent{Key: KeyRight}
			handledRight := e.HandleEvent(rightEvent)
			if handledRight != tt.expectHoriz {
				t.Errorf("KeyRight handled=%v, want %v", handledRight, tt.expectHoriz)
			}
		})
	}
}

func TestElement_ScrollRendersThroughTree(t *testing.T) {
	// Create a parent that contains a scrollable element
	parent := New(
		WithSize(40, 20),
		WithDirection(Column),
	)

	scrollable := New(
		WithHeight(10),
		WithScrollable(ScrollVertical),
		WithDirection(Column),
	)

	// Add items with text
	for i := 0; i < 20; i++ {
		scrollable.AddChild(New(
			WithText("Item"),
			WithHeight(1),
		))
	}

	// Add scrollable directly to parent (no unwrapping needed!)
	parent.AddChild(scrollable)

	// Render through parent
	buf := NewBuffer(40, 20)
	parent.Render(buf, 40, 20)

	// Check that content size was computed
	_, contentH := scrollable.ContentSize()
	if contentH != 20 {
		t.Errorf("ContentSize height=%d, want 20", contentH)
	}

	// Check that scrolling works
	scrollable.ScrollTo(0, 5)
	_, y := scrollable.ScrollOffset()
	if y != 5 {
		t.Errorf("ScrollOffset y=%d, want 5", y)
	}

	// Render again to apply scroll
	parent.MarkDirty()
	parent.Render(buf, 40, 20)

	// Verify scroll is clamped correctly (20 items - 10 viewport = 10 max)
	scrollable.ScrollTo(0, 100)
	_, y = scrollable.ScrollOffset()
	if y != 10 {
		t.Errorf("ScrollOffset y=%d after clamping, want 10", y)
	}
}

func TestElement_ScrollUnhandledEvent(t *testing.T) {
	e := New(
		WithHeight(10),
		WithScrollable(ScrollVertical),
	)

	// Test that unhandled events return false
	event := KeyEvent{Key: KeyRune, Rune: 'a'}
	handled := e.HandleEvent(event)

	if handled {
		t.Error("Rune events should not be handled by scroll")
	}
}

