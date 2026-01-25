package tui

import (
	"testing"
)

// TestIntegration_MockReaderToFocusManager tests the full flow from
// MockEventReader through FocusManager to mock Focusable elements.
func TestIntegration_MockReaderToFocusManager(t *testing.T) {
	type tc struct {
		events           []Event
		expectedHandled  []bool
		focusedAfter     string
		focusCycles      int // number of Next() calls to make
	}

	tests := map[string]tc{
		"single event to focused element": {
			events:          []Event{KeyEvent{Key: KeyEnter}},
			expectedHandled: []bool{true},
			focusedAfter:    "a",
		},
		"multiple events same element": {
			events: []Event{
				KeyEvent{Key: KeyEnter},
				KeyEvent{Key: KeyTab},
				KeyEvent{Key: KeyEscape},
			},
			expectedHandled: []bool{true, true, true},
			focusedAfter:    "a",
		},
		"events after focus change": {
			events: []Event{
				KeyEvent{Key: KeyEnter},
			},
			expectedHandled: []bool{true},
			focusedAfter:    "b",
			focusCycles:     1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock focusable elements
			elemA := newMockFocusable("a", true)
			elemA.handled = true
			elemB := newMockFocusable("b", true)
			elemB.handled = true

			// Create FocusManager with elements
			fm := NewFocusManager()
			fm.Register(elemA)
			fm.Register(elemB)

			// Cycle focus if requested
			for i := 0; i < tt.focusCycles; i++ {
				fm.Next()
			}

			// Create MockEventReader with events
			reader := NewMockEventReader(tt.events...)

			// Process events through the system
			for i, expectedHandled := range tt.expectedHandled {
				event, ok := reader.PollEvent(0)
				if !ok {
					t.Fatalf("Event %d: PollEvent returned false unexpectedly", i)
				}

				handled := fm.Dispatch(event)
				if handled != expectedHandled {
					t.Errorf("Event %d: Dispatch() = %v, want %v", i, handled, expectedHandled)
				}
			}

			// Verify final focus state
			focused := fm.Focused()
			if focused == nil {
				t.Fatal("Focused() returned nil")
			}
			mf := focused.(*mockFocusable)
			if mf.id != tt.focusedAfter {
				t.Errorf("Final focused element = %q, want %q", mf.id, tt.focusedAfter)
			}
		})
	}
}

// TestIntegration_ResizeEventUpdatesBuffer tests that ResizeEvent
// properly updates buffer dimensions through App.Dispatch.
func TestIntegration_ResizeEventUpdatesBuffer(t *testing.T) {
	type tc struct {
		initialWidth  int
		initialHeight int
		resizeWidth   int
		resizeHeight  int
	}

	tests := map[string]tc{
		"grow terminal": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   120,
			resizeHeight:  40,
		},
		"shrink terminal": {
			initialWidth:  120,
			initialHeight: 40,
			resizeWidth:   80,
			resizeHeight:  24,
		},
		"change width only": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   100,
			resizeHeight:  24,
		},
		"change height only": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   80,
			resizeHeight:  30,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := NewBuffer(tt.initialWidth, tt.initialHeight)
			app := &App{
				buffer: buffer,
				focus:  NewFocusManager(),
			}

			// Create resize event
			event := ResizeEvent{Width: tt.resizeWidth, Height: tt.resizeHeight}

			// Dispatch the event
			handled := app.Dispatch(event)
			if !handled {
				t.Error("Dispatch(ResizeEvent) should return true")
			}

			// Verify buffer dimensions
			w, h := app.buffer.Size()
			if w != tt.resizeWidth || h != tt.resizeHeight {
				t.Errorf("Buffer size = (%d, %d), want (%d, %d)", w, h, tt.resizeWidth, tt.resizeHeight)
			}
		})
	}
}

// TestIntegration_EventDispatchToMultipleFocusables tests event dispatch
// across multiple focusable elements with focus navigation.
func TestIntegration_EventDispatchToMultipleFocusables(t *testing.T) {
	// Create elements
	elem1 := newMockFocusable("elem1", true)
	elem1.handled = true
	elem2 := newMockFocusable("elem2", true)
	elem2.handled = true
	elem3 := newMockFocusable("elem3", true)
	elem3.handled = true

	fm := NewFocusManager()
	fm.Register(elem1)
	fm.Register(elem2)
	fm.Register(elem3)

	// Verify initial focus on elem1
	if fm.Focused().(*mockFocusable).id != "elem1" {
		t.Errorf("Initial focus should be elem1, got %s", fm.Focused().(*mockFocusable).id)
	}

	// Send event to elem1
	event1 := KeyEvent{Key: KeyRune, Rune: 'a'}
	fm.Dispatch(event1)
	if elem1.lastEvent == nil {
		t.Error("elem1 should have received the event")
	}

	// Move focus to elem2
	fm.Next()
	if fm.Focused().(*mockFocusable).id != "elem2" {
		t.Error("Focus should have moved to elem2")
	}

	// Send event to elem2
	event2 := KeyEvent{Key: KeyRune, Rune: 'b'}
	fm.Dispatch(event2)
	if elem2.lastEvent == nil {
		t.Error("elem2 should have received the event")
	}

	// elem1's last event should still be event1
	if elem1.lastEvent.(KeyEvent).Rune != 'a' {
		t.Error("elem1's last event should still be 'a'")
	}

	// Move focus to elem3
	fm.Next()
	if fm.Focused().(*mockFocusable).id != "elem3" {
		t.Error("Focus should have moved to elem3")
	}

	// Move back to elem1 (wrap around)
	fm.Next()
	if fm.Focused().(*mockFocusable).id != "elem1" {
		t.Error("Focus should have wrapped to elem1")
	}
}

// TestIntegration_FocusCycleWithNonFocusable tests focus cycling
// correctly skips non-focusable elements.
func TestIntegration_FocusCycleWithNonFocusable(t *testing.T) {
	// Create elements - some not focusable
	elem1 := newMockFocusable("elem1", true)
	elem2 := newMockFocusable("elem2", false) // Not focusable
	elem3 := newMockFocusable("elem3", true)
	elem4 := newMockFocusable("elem4", false) // Not focusable
	elem5 := newMockFocusable("elem5", true)

	fm := NewFocusManager()
	fm.Register(elem1)
	fm.Register(elem2)
	fm.Register(elem3)
	fm.Register(elem4)
	fm.Register(elem5)

	// Track focus order
	focusOrder := []string{}
	focusOrder = append(focusOrder, fm.Focused().(*mockFocusable).id)

	// Navigate forward through all focusable elements
	for i := 0; i < 5; i++ { // More iterations than needed to test wraparound
		fm.Next()
		focusOrder = append(focusOrder, fm.Focused().(*mockFocusable).id)
	}

	// Expected order: elem1 -> elem3 -> elem5 -> elem1 -> elem3 -> elem5
	expected := []string{"elem1", "elem3", "elem5", "elem1", "elem3", "elem5"}
	if len(focusOrder) != len(expected) {
		t.Fatalf("Focus order length = %d, want %d", len(focusOrder), len(expected))
	}
	for i, id := range expected {
		if focusOrder[i] != id {
			t.Errorf("Focus order[%d] = %q, want %q", i, focusOrder[i], id)
		}
	}
}

// TestIntegration_EventReaderWithFocusManager tests the integration
// of MockEventReader with FocusManager for a sequence of events.
func TestIntegration_EventReaderWithFocusManager(t *testing.T) {
	events := []Event{
		KeyEvent{Key: KeyRune, Rune: 'h'},
		KeyEvent{Key: KeyRune, Rune: 'e'},
		KeyEvent{Key: KeyRune, Rune: 'l'},
		KeyEvent{Key: KeyRune, Rune: 'l'},
		KeyEvent{Key: KeyRune, Rune: 'o'},
		KeyEvent{Key: KeyEnter},
	}

	reader := NewMockEventReader(events...)
	elem := newMockFocusable("input", true)
	elem.handled = true
	fm := NewFocusManager()
	fm.Register(elem)

	// Process all events
	eventCount := 0
	for {
		event, ok := reader.PollEvent(0)
		if !ok {
			break
		}
		fm.Dispatch(event)
		eventCount++
	}

	// Verify all events were processed
	if eventCount != len(events) {
		t.Errorf("Processed %d events, want %d", eventCount, len(events))
	}

	// Last event should be KeyEnter
	lastEvent := elem.lastEvent.(KeyEvent)
	if lastEvent.Key != KeyEnter {
		t.Errorf("Last event key = %v, want KeyEnter", lastEvent.Key)
	}
}

// TestIntegration_UnhandledEventPropagation tests that unhandled events
// return false, allowing applications to implement bubbling/fallback.
func TestIntegration_UnhandledEventPropagation(t *testing.T) {
	elem := newMockFocusable("elem", true)
	elem.handled = false // Element doesn't handle events

	fm := NewFocusManager()
	fm.Register(elem)

	event := KeyEvent{Key: KeyEnter}
	handled := fm.Dispatch(event)

	if handled {
		t.Error("Dispatch should return false when element doesn't handle event")
	}

	// Event should still have been passed to element
	if elem.lastEvent == nil {
		t.Error("Event should have been passed to element even if not handled")
	}
}
