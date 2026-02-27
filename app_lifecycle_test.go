package tui

import (
	"testing"
)

func TestApp_PollEventWithMockReader(t *testing.T) {
	type tc struct {
		events      []Event
		expectedOk  bool
		expectedKey Key
	}

	tests := map[string]tc{
		"returns queued event": {
			events:      []Event{KeyEvent{Key: KeyEnter}},
			expectedOk:  true,
			expectedKey: KeyEnter,
		},
		"returns false when exhausted": {
			events:     []Event{},
			expectedOk: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockReader := NewMockEventReader(tt.events...)

			app := &App{
				reader: mockReader,
				focus:  newFocusManager(),
				buffer: NewBuffer(80, 24),
			}

			event, ok := app.PollEvent(0)

			if ok != tt.expectedOk {
				t.Errorf("PollEvent() ok = %v, want %v", ok, tt.expectedOk)
			}

			if tt.expectedOk {
				ke, isKey := event.(KeyEvent)
				if !isKey {
					t.Fatalf("PollEvent() returned %T, want KeyEvent", event)
				}
				if ke.Key != tt.expectedKey {
					t.Errorf("PollEvent() key = %v, want %v", ke.Key, tt.expectedKey)
				}
			}
		})
	}
}

func TestApp_MultipleEventsFromMockReader(t *testing.T) {
	events := []Event{
		KeyEvent{Key: KeyEnter},
		KeyEvent{Key: KeyTab},
		KeyEvent{Key: KeyEscape},
	}

	mockReader := NewMockEventReader(events...)

	app := &App{
		reader: mockReader,
		focus:  newFocusManager(),
		buffer: NewBuffer(80, 24),
	}

	// Should return events in order
	for i, expected := range events {
		event, ok := app.PollEvent(0)
		if !ok {
			t.Fatalf("PollEvent() %d returned ok=false, want true", i)
		}

		ke, isKey := event.(KeyEvent)
		if !isKey {
			t.Fatalf("PollEvent() %d returned %T, want KeyEvent", i, event)
		}

		expectedKey := expected.(KeyEvent).Key
		if ke.Key != expectedKey {
			t.Errorf("PollEvent() %d key = %v, want %v", i, ke.Key, expectedKey)
		}
	}

	// Should now be exhausted
	_, ok := app.PollEvent(0)
	if ok {
		t.Error("PollEvent() should return false when exhausted")
	}
}

func TestApp_BufferReturnsBuffer(t *testing.T) {
	buffer := NewBuffer(80, 24)
	app := &App{
		buffer: buffer,
		focus:  newFocusManager(),
	}

	if app.Buffer() != buffer {
		t.Error("Buffer() should return the app's buffer")
	}
}

func TestApp_FocusNext(t *testing.T) {
	app := &App{
		focus:  newFocusManager(),
		buffer: NewBuffer(80, 24),
	}

	elem1 := newMockFocusable("elem1", true)
	elem2 := newMockFocusable("elem2", true)
	app.focus.Register(elem1)
	app.focus.Register(elem2)

	// Initially focused on elem1
	if app.Focused().(*mockFocusable).id != "elem1" {
		t.Error("Initial focus should be elem1")
	}

	// FocusNext should move to elem2
	app.FocusNext()

	if app.Focused().(*mockFocusable).id != "elem2" {
		t.Error("After FocusNext(), focus should be elem2")
	}
}

func TestApp_FocusPrev(t *testing.T) {
	app := &App{
		focus:  newFocusManager(),
		buffer: NewBuffer(80, 24),
	}

	elem1 := newMockFocusable("elem1", true)
	elem2 := newMockFocusable("elem2", true)
	app.focus.Register(elem1)
	app.focus.Register(elem2)

	// Initially focused on elem1
	if app.Focused().(*mockFocusable).id != "elem1" {
		t.Error("Initial focus should be elem1")
	}

	// FocusPrev should wrap to elem2
	app.FocusPrev()

	if app.Focused().(*mockFocusable).id != "elem2" {
		t.Error("After FocusPrev(), focus should be elem2")
	}
}

func TestApp_Focused(t *testing.T) {
	app := &App{
		focus:  newFocusManager(),
		buffer: NewBuffer(80, 24),
	}

	// No focused element initially
	if app.Focused() != nil {
		t.Error("Focused() should return nil when no elements registered")
	}

	// Register an element
	elem := newMockFocusable("elem", true)
	app.focus.Register(elem)

	// Now should return the focused element
	focused := app.Focused()
	if focused == nil {
		t.Error("Focused() should return non-nil after registering element")
	}
	if focused.(*mockFocusable).id != "elem" {
		t.Error("Focused() should return the registered element")
	}
}

func TestApp_SetRoot_AutoRegistration(t *testing.T) {
	type tc struct {
		numFocusable int
	}

	tests := map[string]tc{
		"single focusable": {
			numFocusable: 1,
		},
		"multiple focusables": {
			numFocusable: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{
				focus:  newFocusManager(),
				buffer: NewBuffer(80, 24),
			}

			root := New()
			for i := 0; i < tt.numFocusable; i++ {
				root.AddChild(New(WithFocusable(true)))
			}
			app.SetRoot(root)

			// Verify focusables were auto-registered
			focused := app.Focused()
			if focused == nil {
				t.Fatal("Focused() returned nil")
			}
		})
	}
}

func TestApp_SetRoot_DynamicFocusableRegistration(t *testing.T) {
	app := &App{
		focus:  newFocusManager(),
		buffer: NewBuffer(80, 24),
	}

	root := New()
	app.SetRoot(root)

	// No focusables initially
	if app.Focused() != nil {
		t.Fatal("Focused() should be nil with no focusable children")
	}

	// Dynamically add a focusable child — the callback set by applyRoot should register it
	child := New(WithFocusable(true))
	root.AddChild(child)

	// Verify it was registered
	focused := app.Focused()
	if focused == nil {
		t.Fatal("Focused() returned nil after adding focusable child")
	}
}
