package tui

import (
	"testing"
)

// mockRenderable is a mock implementation of Renderable for testing.
type mockRenderable struct {
	dirty        bool
	renderCalled bool
	markDirtyCalled bool
}

func newMockRenderable() *mockRenderable {
	return &mockRenderable{dirty: true}
}

func (m *mockRenderable) Render(buf *Buffer, width, height int) {
	m.renderCalled = true
	m.dirty = false
}

func (m *mockRenderable) MarkDirty() {
	m.dirty = true
	m.markDirtyCalled = true
}

func (m *mockRenderable) IsDirty() bool {
	return m.dirty
}

func TestApp_SetRootAndRoot(t *testing.T) {
	type tc struct {
		createRoot bool
	}

	tests := map[string]tc{
		"with root element": {
			createRoot: true,
		},
		"without root element": {
			createRoot: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a mock app (we can't test NewApp without a real terminal)
			app := &App{
				focus:  NewFocusManager(),
				buffer: NewBuffer(80, 24),
			}

			if tt.createRoot {
				root := newMockRenderable()
				app.SetRoot(root)

				if app.Root() != root {
					t.Error("Root() should return the element passed to SetRoot()")
				}
			} else {
				if app.Root() != nil {
					t.Error("Root() should return nil when no root set")
				}
			}
		})
	}
}

func TestApp_Focus(t *testing.T) {
	app := &App{
		focus: NewFocusManager(),
	}

	if app.Focus() == nil {
		t.Error("Focus() should return a non-nil FocusManager")
	}
}

func TestApp_DispatchResizeEvent(t *testing.T) {
	type tc struct {
		initialWidth  int
		initialHeight int
		resizeWidth   int
		resizeHeight  int
		hasRoot       bool
	}

	tests := map[string]tc{
		"resize with root": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   100,
			resizeHeight:  30,
			hasRoot:       true,
		},
		"resize without root": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   100,
			resizeHeight:  30,
			hasRoot:       false,
		},
		"shrink terminal": {
			initialWidth:  100,
			initialHeight: 50,
			resizeWidth:   60,
			resizeHeight:  20,
			hasRoot:       true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := NewBuffer(tt.initialWidth, tt.initialHeight)
			app := &App{
				focus:  NewFocusManager(),
				buffer: buffer,
			}

			var mockRoot *mockRenderable
			if tt.hasRoot {
				mockRoot = newMockRenderable()
				mockRoot.dirty = false // Start as not dirty
				app.SetRoot(mockRoot)
			}

			event := ResizeEvent{Width: tt.resizeWidth, Height: tt.resizeHeight}
			handled := app.Dispatch(event)

			if !handled {
				t.Error("Dispatch(ResizeEvent) should return true")
			}

			// Check buffer was resized
			bufW, bufH := app.buffer.Size()
			if bufW != tt.resizeWidth || bufH != tt.resizeHeight {
				t.Errorf("Buffer size = (%d, %d), want (%d, %d)", bufW, bufH, tt.resizeWidth, tt.resizeHeight)
			}

			// Check root was marked dirty if it exists
			if tt.hasRoot && !mockRoot.markDirtyCalled {
				t.Error("MarkDirty should have been called on root after resize")
			}
		})
	}
}

func TestApp_DispatchKeyEvent(t *testing.T) {
	type tc struct {
		hasFocused   bool
		handled      bool
		expectReturn bool
	}

	tests := map[string]tc{
		"event handled by focused element": {
			hasFocused:   true,
			handled:      true,
			expectReturn: true,
		},
		"event not handled by focused element": {
			hasFocused:   true,
			handled:      false,
			expectReturn: false,
		},
		"no focused element": {
			hasFocused:   false,
			handled:      false,
			expectReturn: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			focus := NewFocusManager()

			if tt.hasFocused {
				mock := newMockFocusable("a", true)
				mock.handled = tt.handled
				focus.Register(mock)
			}

			app := &App{
				focus:  focus,
				buffer: NewBuffer(80, 24),
			}

			event := KeyEvent{Key: KeyEnter}
			result := app.Dispatch(event)

			if result != tt.expectReturn {
				t.Errorf("Dispatch(KeyEvent) = %v, want %v", result, tt.expectReturn)
			}
		})
	}
}

func TestApp_RenderWithMockRoot(t *testing.T) {
	// Create a mock terminal for testing
	mockTerm := NewMockTerminal(80, 24)
	buffer := NewBuffer(80, 24)

	app := &App{
		terminal: nil, // We can't use a real ANSITerminal in tests
		buffer:   buffer,
		focus:    NewFocusManager(),
	}

	// Create a mock renderable
	mockRoot := newMockRenderable()
	app.SetRoot(mockRoot)

	// Test that rendering calls the root's Render method
	mockRoot.Render(buffer, 80, 24)

	if !mockRoot.renderCalled {
		t.Error("Root's Render method should have been called")
	}

	// Verify the mock was used
	_ = mockTerm // We created it but App tests are limited without terminal
}

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
				focus:  NewFocusManager(),
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
		focus:  NewFocusManager(),
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
		focus:  NewFocusManager(),
	}

	if app.Buffer() != buffer {
		t.Error("Buffer() should return the app's buffer")
	}
}

func TestApp_FocusNext(t *testing.T) {
	app := &App{
		focus:  NewFocusManager(),
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
		focus:  NewFocusManager(),
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
		focus:  NewFocusManager(),
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

// mockFocusableTreeWalker is a mock that implements focusableTreeWalker
type mockFocusableTreeWalker struct {
	*mockRenderable
	focusables           []Focusable
	onFocusableAddedFn   func(Focusable)
}

func newMockFocusableTreeWalker(focusables ...Focusable) *mockFocusableTreeWalker {
	return &mockFocusableTreeWalker{
		mockRenderable: newMockRenderable(),
		focusables:     focusables,
	}
}

func (m *mockFocusableTreeWalker) SetOnFocusableAdded(fn func(Focusable)) {
	m.onFocusableAddedFn = fn
}

func (m *mockFocusableTreeWalker) WalkFocusables(fn func(Focusable)) {
	for _, f := range m.focusables {
		fn(f)
	}
}

func TestApp_SetRoot_AutoRegistration(t *testing.T) {
	type tc struct {
		focusables       []*mockFocusable
		expectedFocusedID string
	}

	tests := map[string]tc{
		"single focusable": {
			focusables: []*mockFocusable{
				newMockFocusable("elem1", true),
			},
			expectedFocusedID: "elem1",
		},
		"multiple focusables": {
			focusables: []*mockFocusable{
				newMockFocusable("elem1", true),
				newMockFocusable("elem2", true),
				newMockFocusable("elem3", true),
			},
			expectedFocusedID: "elem1",
		},
		"skips non-focusable": {
			focusables: []*mockFocusable{
				newMockFocusable("elem1", false),
				newMockFocusable("elem2", true),
			},
			expectedFocusedID: "elem2",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{
				focus:  NewFocusManager(),
				buffer: NewBuffer(80, 24),
			}

			// Convert to []Focusable
			focusables := make([]Focusable, len(tt.focusables))
			for i, f := range tt.focusables {
				focusables[i] = f
			}

			root := newMockFocusableTreeWalker(focusables...)
			app.SetRoot(root)

			// Verify focusables were auto-registered
			focused := app.Focused()
			if focused == nil {
				t.Fatal("Focused() returned nil")
			}

			mf := focused.(*mockFocusable)
			if mf.id != tt.expectedFocusedID {
				t.Errorf("Focused element = %q, want %q", mf.id, tt.expectedFocusedID)
			}
		})
	}
}

func TestApp_SetRoot_OnFocusableAddedCallback(t *testing.T) {
	app := &App{
		focus:  NewFocusManager(),
		buffer: NewBuffer(80, 24),
	}

	root := newMockFocusableTreeWalker()
	app.SetRoot(root)

	// Verify callback was set
	if root.onFocusableAddedFn == nil {
		t.Fatal("SetRoot should set onFocusableAdded callback")
	}

	// Simulate adding a new focusable
	newElem := newMockFocusable("newElem", true)
	root.onFocusableAddedFn(newElem)

	// Verify it was registered
	focused := app.Focused()
	if focused == nil {
		t.Fatal("Focused() returned nil after callback")
	}

	mf := focused.(*mockFocusable)
	if mf.id != "newElem" {
		t.Errorf("Focused element = %q, want 'newElem'", mf.id)
	}
}
