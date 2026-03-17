package tui

import (
	"testing"
	"time"
)

func TestUpdateEvent_ImplementsEvent(t *testing.T) {
	var ev Event = UpdateEvent{fn: func() {}}
	if ev == nil {
		t.Fatal("UpdateEvent should implement Event")
	}
}

func TestUpdateEvent_RunsClosure(t *testing.T) {
	called := false
	ev := UpdateEvent{fn: func() { called = true }}
	ev.fn()
	if !called {
		t.Fatal("UpdateEvent closure should have been called")
	}
}

func TestRender_SkipsWhenNotDirty(t *testing.T) {
	mock := NewMockTerminal(80, 24)
	app := &App{
		terminal:     mock,
		buffer:       NewBuffer(80, 24),
		focus:        newFocusManager(),
		merged:       make(chan Event, 1),
		watcherQueue: make(chan func(), 1),
		stopCh:       make(chan struct{}),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}
	root := New(WithText("hello"))
	app.SetRoot(root)

	// First render: dirty from SetRoot
	app.Render()

	// Capture buffer state after first render
	snap1 := app.buffer.StringTrimmed()

	// Second render without marking dirty: should be no-op
	root.SetText("changed")
	app.resetDirty()
	app.Render()

	snap2 := app.buffer.StringTrimmed()
	if snap1 != snap2 {
		t.Errorf("Render() should be no-op when not dirty\nbefore: %q\nafter: %q", snap1, snap2)
	}
}

func TestClose_Idempotent(t *testing.T) {
	app := &App{
		terminal:     NewMockTerminal(80, 24),
		reader:       &MockEventReader{},
		buffer:       NewBuffer(80, 24),
		focus:        newFocusManager(),
		inputEvents:  make(chan Event, 256),
		updates:      make(chan Event, 256),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}

	if err := app.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	// First close
	app.Close()

	// Second close should not panic
	app.Close()
}

func TestEvents_ReturnsChannel(t *testing.T) {
	app := &App{
		merged: make(chan Event, 256),
		stopCh: make(chan struct{}),
	}
	ch := app.Events()
	if ch == nil {
		t.Fatal("Events() should return non-nil channel")
	}

	// Send an event and verify it's received
	app.merged <- KeyEvent{Key: KeyEnter}
	select {
	case ev := <-ch:
		if _, ok := ev.(KeyEvent); !ok {
			t.Fatalf("expected KeyEvent, got %T", ev)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected event on channel")
	}
}

func TestDispatchEvents_ReturnsFalseOnStop(t *testing.T) {
	app := &App{
		terminal:     NewMockTerminal(80, 24),
		buffer:       NewBuffer(80, 24),
		focus:        newFocusManager(),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}
	app.Stop()

	if app.DispatchEvents() {
		t.Fatal("DispatchEvents should return false when stopped")
	}
}

func TestStep_ReturnsFalseOnStop(t *testing.T) {
	app := &App{
		terminal:     NewMockTerminal(80, 24),
		buffer:       NewBuffer(80, 24),
		focus:        newFocusManager(),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}
	app.Stop()

	if app.Step() {
		t.Fatal("Step should return false when stopped")
	}
}

// Integration tests for manual event loop patterns

func TestManualLoop_StepExitsOnStop(t *testing.T) {
	app := &App{
		terminal:      NewMockTerminal(80, 24),
		reader:        &MockEventReader{},
		buffer:        NewBuffer(80, 24),
		focus:         newFocusManager(),
		merged:        make(chan Event, 256),
		watcherQueue:  make(chan func(), 256),
		stopCh:        make(chan struct{}),
		mounts:        newMountState(),
		batch:         newBatchContext(),
		frameDuration: 16 * time.Millisecond,
	}

	root := New(WithText("hello"))
	app.SetRoot(root)

	// Run a step
	if !app.Step() {
		t.Fatal("Step should return true while running")
	}

	// Stop and verify Step returns false
	app.Stop()
	if app.Step() {
		t.Fatal("Step should return false after Stop")
	}
}

func TestManualLoop_EventsChannelReceivesUpdates(t *testing.T) {
	app := &App{
		terminal:     NewMockTerminal(80, 24),
		reader:       &MockEventReader{},
		buffer:       NewBuffer(80, 24),
		focus:        newFocusManager(),
		inputEvents:  make(chan Event, 256),
		updates:      make(chan Event, 256),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}
	app.startEventMerge()
	defer app.Stop()

	called := false
	app.QueueUpdate(func() { called = true })

	select {
	case ev, ok := <-app.Events():
		if !ok {
			t.Fatal("channel should be open")
		}
		app.Dispatch(ev)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected event on channel")
	}

	if !called {
		t.Fatal("QueueUpdate closure should have been dispatched")
	}
}

func TestManualLoop_DispatchEventsProcessesAll(t *testing.T) {
	app := &App{
		terminal:     NewMockTerminal(80, 24),
		buffer:       NewBuffer(80, 24),
		focus:        newFocusManager(),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}

	count := 0
	for i := 0; i < 5; i++ {
		app.merged <- UpdateEvent{fn: func() { count++ }}
	}

	ok := app.DispatchEvents()
	if !ok {
		t.Fatal("DispatchEvents should return true")
	}
	if count != 5 {
		t.Fatalf("expected 5 dispatches, got %d", count)
	}
}

func TestManualLoop_OpenDoubleCallErrors(t *testing.T) {
	app := &App{
		terminal:      NewMockTerminal(80, 24),
		reader:        &MockEventReader{},
		buffer:        NewBuffer(80, 24),
		focus:         newFocusManager(),
		inputEvents:   make(chan Event, 256),
		updates:       make(chan Event, 256),
		merged:        make(chan Event, 256),
		watcherQueue:  make(chan func(), 256),
		stopCh:        make(chan struct{}),
		mounts:        newMountState(),
		batch:         newBatchContext(),
		frameDuration: 16 * time.Millisecond,
	}

	if err := app.Open(); err != nil {
		t.Fatalf("first Open should succeed: %v", err)
	}
	defer app.Close()

	if err := app.Open(); err == nil {
		t.Fatal("second Open should return error")
	}
}

func TestManualLoop_SelectMultiplexing(t *testing.T) {
	app := &App{
		terminal:     NewMockTerminal(80, 24),
		reader:       &MockEventReader{},
		buffer:       NewBuffer(80, 24),
		focus:        newFocusManager(),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}

	customCh := make(chan string, 1)
	customCh <- "external"

	app.merged <- UpdateEvent{fn: func() {}}

	tuiHandled := false
	customHandled := false

	// Process both sources
	for i := 0; i < 2; i++ {
		select {
		case ev, ok := <-app.Events():
			if ok {
				app.Dispatch(ev)
				tuiHandled = true
			}
		case msg := <-customCh:
			if msg == "external" {
				customHandled = true
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timed out waiting for events")
		}
	}

	if !tuiHandled {
		t.Error("tui event should have been handled")
	}
	if !customHandled {
		t.Error("custom event should have been handled")
	}
}
