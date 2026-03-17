package tui

import (
	"testing"
	"time"
)

func TestApp_QueueUpdate_EnqueuesSafely(t *testing.T) {
	app := &App{
		focus:        newFocusManager(),
		buffer:       NewBuffer(80, 24),
		updates:      make(chan Event, 256),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		stopped:      false,
	}

	var executed bool
	app.QueueUpdate(func() {
		executed = true
	})

	// QueueUpdate sends to updates channel; read directly (no fan-in in test)
	select {
	case ev := <-app.updates:
		app.Dispatch(ev)
		if !executed {
			t.Error("Queued function was not executed correctly")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("QueueUpdate did not enqueue function")
	}
}

func TestApp_QueueUpdate_FromGoroutine(t *testing.T) {
	app := &App{
		focus:        newFocusManager(),
		buffer:       NewBuffer(80, 24),
		updates:      make(chan Event, 256),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		stopped:      false,
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

	// Read all queued functions from updates channel
	go func() {
		for i := 0; i < 10; i++ {
			select {
			case ev := <-app.updates:
				app.Dispatch(ev)
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

func TestApp_QueueUpdate_DropsWhenFull(t *testing.T) {
	app := &App{
		focus:        newFocusManager(),
		buffer:       NewBuffer(80, 24),
		updates:      make(chan Event, 1),
		merged:       make(chan Event, 1),
		watcherQueue: make(chan func(), 1),
		stopCh:       make(chan struct{}),
	}

	seen := make([]int, 0, 2)
	app.QueueUpdate(func() { seen = append(seen, 1) }) // fits in buffer
	app.QueueUpdate(func() { seen = append(seen, 2) }) // channel full, dropped

	// Drain: only the first update should be present
	select {
	case ev := <-app.updates:
		app.Dispatch(ev)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected queued update")
	}

	// Channel should be empty now
	select {
	case <-app.updates:
		t.Fatal("expected channel to be empty after draining one event")
	default:
	}

	if len(seen) != 1 || seen[0] != 1 {
		t.Fatalf("expected only first update to run, got %v", seen)
	}
}

func TestApp_SetGlobalKeyHandler(t *testing.T) {
	app := &App{
		focus:        newFocusManager(),
		buffer:       NewBuffer(80, 24),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		stopped:      false,
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
		focus:        newFocusManager(),
		buffer:       NewBuffer(80, 24),
		reader:       mockReader,
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		stopped:      false,
	}
	app.focus.Register(focusable)
	app.focus.SetFocus(focusable)

	var globalHandlerCalled bool
	app.SetGlobalKeyHandler(func(e KeyEvent) bool {
		globalHandlerCalled = true
		if e.Rune == 'q' {
			return true // Consume event
		}
		return false
	})

	// Dispatch goes through Dispatch() which handles globalKeyHandler in legacy path
	event := KeyEvent{Key: KeyRune, Rune: 'q'}
	app.Dispatch(event)

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
		focus:        newFocusManager(),
		buffer:       NewBuffer(80, 24),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		stopped:      false,
	}
	app.focus.Register(focusable)
	app.focus.SetFocus(focusable)

	var globalHandlerCalled bool
	app.SetGlobalKeyHandler(func(e KeyEvent) bool {
		globalHandlerCalled = true
		// Don't consume - let it pass through
		return false
	})

	// Dispatch goes through Dispatch() which handles globalKeyHandler in legacy path
	event := KeyEvent{Key: KeyRune, Rune: 'j'}
	app.Dispatch(event)

	if !globalHandlerCalled {
		t.Error("Global handler was not called")
	}

	if focusable.lastEvent == nil {
		t.Error("Event should have been passed to focused element")
	}
}

func TestApp_EventBatching(t *testing.T) {
	// Reset dirty flag for clean test
	testApp.resetDirty()

	mockReader := NewMockEventReader()

	app := &App{
		focus:        newFocusManager(),
		buffer:       NewBuffer(80, 24),
		reader:       mockReader,
		root:         New(),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		stopped:      false,
	}

	// Queue multiple events directly to merged (simulating fan-in output)
	for i := 0; i < 5; i++ {
		app.merged <- UpdateEvent{fn: func() {
			testApp.MarkDirty()
		}}
	}

	// Process one batch manually (simulating the Run() loop logic)
	// Block until at least one event arrives
	select {
	case ev := <-app.merged:
		app.Dispatch(ev)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event in queue")
	}

	// Drain additional queued events
drain:
	for {
		select {
		case ev := <-app.merged:
			app.Dispatch(ev)
		default:
			break drain
		}
	}

	// Only check dirty once, clear it
	var renderCount int
	if testApp.checkAndClearDirty() {
		// Would call Render() here in the real loop
		renderCount++
	}

	// Should only have rendered once despite multiple events
	if renderCount != 1 {
		t.Errorf("Expected 1 render after batched events, got %d", renderCount)
	}
}
