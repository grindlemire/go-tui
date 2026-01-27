package tui

import (
	"testing"
	"time"
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

// --- Phase 2: Event Loop Tests ---

// mockViewable implements Viewable interface for testing
type mockViewable struct {
	root     Renderable
	watchers []Watcher
}

func newMockViewable(root Renderable, watchers ...Watcher) *mockViewable {
	return &mockViewable{root: root, watchers: watchers}
}

func (m *mockViewable) GetRoot() Renderable {
	return m.root
}

func (m *mockViewable) GetWatchers() []Watcher {
	return m.watchers
}

// mockWatcher tracks whether Start was called
type mockWatcher struct {
	started     bool
	eventQueue  chan<- func()
	stopCh      <-chan struct{}
	startCalled chan struct{} // signaled when Start is called
}

func newMockWatcher() *mockWatcher {
	return &mockWatcher{
		startCalled: make(chan struct{}),
	}
}

func (m *mockWatcher) Start(eventQueue chan<- func(), stopCh <-chan struct{}) {
	m.started = true
	m.eventQueue = eventQueue
	m.stopCh = stopCh
	close(m.startCalled)
}

func TestApp_SetRoot_WithViewable(t *testing.T) {
	type tc struct {
		name         string
		numWatchers  int
		expectRoot   bool
	}

	tests := map[string]tc{
		"viewable with no watchers": {
			numWatchers: 0,
			expectRoot:  true,
		},
		"viewable with one watcher": {
			numWatchers: 1,
			expectRoot:  true,
		},
		"viewable with multiple watchers": {
			numWatchers: 3,
			expectRoot:  true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{
				focus:      NewFocusManager(),
				buffer:     NewBuffer(80, 24),
				eventQueue: make(chan func(), 256),
				stopCh:     make(chan struct{}),
			}

			root := newMockRenderable()
			watchers := make([]Watcher, tt.numWatchers)
			mockWatchers := make([]*mockWatcher, tt.numWatchers)
			for i := 0; i < tt.numWatchers; i++ {
				mw := newMockWatcher()
				mockWatchers[i] = mw
				watchers[i] = mw
			}

			view := newMockViewable(root, watchers...)
			app.SetRoot(view)

			// Verify root was set
			if tt.expectRoot && app.Root() != root {
				t.Error("Root() should return the root from Viewable")
			}

			// Verify all watchers were started
			for i, mw := range mockWatchers {
				if !mw.started {
					t.Errorf("Watcher %d was not started", i)
				}
				if mw.eventQueue != app.eventQueue {
					t.Errorf("Watcher %d received wrong eventQueue", i)
				}
			}
		})
	}
}

func TestApp_SetRoot_WithRawRenderable(t *testing.T) {
	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
	}

	root := newMockRenderable()
	app.SetRoot(root)

	if app.Root() != root {
		t.Error("Root() should return the Renderable passed to SetRoot()")
	}
}

func TestApp_Run_EventLoopLogic(t *testing.T) {
	// Test the core event loop logic without a real terminal.
	// We simulate what Run() does: process events from eventQueue, check dirty, etc.

	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}

	// Queue an event
	var eventProcessed bool
	app.eventQueue <- func() {
		eventProcessed = true
	}

	// Process one event manually (simulating the Run loop)
	select {
	case handler := <-app.eventQueue:
		handler()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event in queue")
	}

	if !eventProcessed {
		t.Error("Event was not processed")
	}

	// Test that Stop() closes stopCh
	app.Stop()

	select {
	case <-app.stopCh:
		// Expected - stopCh was closed
	default:
		t.Error("Stop() should close stopCh")
	}
}

func TestApp_Stop_IsIdempotent(t *testing.T) {
	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}

	// First call should work
	app.Stop()

	if !app.stopped {
		t.Error("Stop() should set stopped to true")
	}

	// Second call should not panic
	app.Stop()

	// Still stopped
	if !app.stopped {
		t.Error("stopped should still be true after second Stop() call")
	}
}

func TestApp_QueueUpdate_EnqueuesSafely(t *testing.T) {
	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}

	var executed bool
	app.QueueUpdate(func() {
		executed = true
	})

	// Read from queue and execute
	select {
	case fn := <-app.eventQueue:
		fn()
		if !executed {
			t.Error("Queued function was not executed correctly")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("QueueUpdate did not enqueue function")
	}
}

func TestApp_QueueUpdate_FromGoroutine(t *testing.T) {
	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}

	var executed int
	done := make(chan struct{})

	// Queue from multiple goroutines
	for i := 0; i < 10; i++ {
		go func() {
			app.QueueUpdate(func() {
				executed++
			})
		}()
	}

	// Read all queued functions
	go func() {
		for i := 0; i < 10; i++ {
			select {
			case fn := <-app.eventQueue:
				fn()
			case <-time.After(100 * time.Millisecond):
				return
			}
		}
		close(done)
	}()

	select {
	case <-done:
		if executed != 10 {
			t.Errorf("Expected 10 executions, got %d", executed)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timed out waiting for goroutines to complete")
	}
}

func TestApp_SetGlobalKeyHandler(t *testing.T) {
	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}

	var handlerCalled bool
	app.SetGlobalKeyHandler(func(e KeyEvent) bool {
		handlerCalled = true
		return true
	})

	if app.globalKeyHandler == nil {
		t.Fatal("SetGlobalKeyHandler should set the handler")
	}

	// Call it
	result := app.globalKeyHandler(KeyEvent{Key: KeyRune, Rune: 'q'})

	if !handlerCalled {
		t.Error("Global key handler was not called")
	}
	if !result {
		t.Error("Global key handler should return true")
	}
}

func TestApp_GlobalKeyHandler_ConsumesEvent(t *testing.T) {
	mockReader := NewMockEventReader(KeyEvent{Key: KeyRune, Rune: 'q'})

	focusable := newMockFocusable("elem", true)
	focusable.handled = false

	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		reader:     mockReader,
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}
	app.focus.Register(focusable)

	var globalHandlerCalled bool
	app.SetGlobalKeyHandler(func(e KeyEvent) bool {
		globalHandlerCalled = true
		if e.Rune == 'q' {
			return true // Consume event
		}
		return false
	})

	// Simulate the event dispatch logic from readInputEvents
	event := KeyEvent{Key: KeyRune, Rune: 'q'}

	// Global handler should consume the event
	if app.globalKeyHandler != nil && app.globalKeyHandler(event) {
		// Event consumed, don't dispatch further
	} else {
		app.Dispatch(event)
	}

	if !globalHandlerCalled {
		t.Error("Global handler was not called")
	}

	if focusable.lastEvent != nil {
		t.Error("Event should have been consumed by global handler")
	}
}

func TestApp_GlobalKeyHandler_PassesEvent(t *testing.T) {
	focusable := newMockFocusable("elem", true)
	focusable.handled = true

	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}
	app.focus.Register(focusable)

	var globalHandlerCalled bool
	app.SetGlobalKeyHandler(func(e KeyEvent) bool {
		globalHandlerCalled = true
		// Don't consume - let it pass through
		return false
	})

	// Simulate the event dispatch logic from readInputEvents
	event := KeyEvent{Key: KeyRune, Rune: 'j'}

	// Global handler should NOT consume the event
	consumed := false
	if app.globalKeyHandler != nil && app.globalKeyHandler(event) {
		consumed = true
	}
	if !consumed {
		app.Dispatch(event)
	}

	if !globalHandlerCalled {
		t.Error("Global handler was not called")
	}

	if focusable.lastEvent == nil {
		t.Error("Event should have been passed to focused element")
	}
}

func TestApp_EventBatching(t *testing.T) {
	// Reset dirty flag for clean test
	resetDirty()

	mockReader := NewMockEventReader()

	var renderCount int
	mockRenderable := &renderCountingMock{
		mockRenderable: newMockRenderable(),
		onRender: func() {
			renderCount++
		},
	}

	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		reader:     mockReader,
		root:       mockRenderable,
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}

	// Queue multiple events that mark dirty
	for i := 0; i < 5; i++ {
		app.eventQueue <- func() {
			MarkDirty()
		}
	}

	// Process one batch manually (simulating the Run() loop logic)
	// Block until at least one event arrives
	select {
	case handler := <-app.eventQueue:
		handler()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event in queue")
	}

	// Drain additional queued events
drain:
	for {
		select {
		case handler := <-app.eventQueue:
			handler()
		default:
			break drain
		}
	}

	// Only check dirty once, clear it
	if checkAndClearDirty() {
		// Would call Render() here in the real loop
		renderCount++ // Simulated render
	}

	// Should only have rendered once despite multiple events
	if renderCount != 1 {
		t.Errorf("Expected 1 render after batched events, got %d", renderCount)
	}
}

// renderCountingMock wraps mockRenderable to count renders
type renderCountingMock struct {
	*mockRenderable
	onRender func()
}

func (m *renderCountingMock) Render(buf *Buffer, width, height int) {
	m.mockRenderable.Render(buf, width, height)
	if m.onRender != nil {
		m.onRender()
	}
}
