package tui

import (
	"testing"
	"time"
)

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

type mockComponentRoot struct{}

func (m *mockComponentRoot) Render() *Element { return New() }

type stopAwareWatcher struct {
	stopped chan struct{}
}

func newStopAwareWatcher() *stopAwareWatcher {
	return &stopAwareWatcher{stopped: make(chan struct{}, 1)}
}

func (w *stopAwareWatcher) Start(eventQueue chan<- func(), stopCh <-chan struct{}) {
	go func() {
		<-stopCh
		w.stopped <- struct{}{}
	}()
}

func TestApp_SetRoot_WithViewable(t *testing.T) {
	type tc struct {
		name        string
		numWatchers int
		expectRoot  bool
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
			app.SetRootView(view)

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

func TestApp_SetRoot_ClearsRootComponentForRenderable(t *testing.T) {
	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		mounts:     newMountState(),
	}

	app.SetRootComponent(&mockComponentRoot{})
	if app.rootComponent == nil {
		t.Fatal("expected rootComponent to be set after SetRootComponent")
	}

	app.SetRoot(newMockRenderable())
	if app.rootComponent != nil {
		t.Fatal("expected rootComponent to be cleared after SetRoot")
	}
}

func TestApp_SetRoot_StopsPreviousRootWatchers(t *testing.T) {
	app := &App{
		focus:      NewFocusManager(),
		buffer:     NewBuffer(80, 24),
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		mounts:     newMountState(),
	}

	w1 := newStopAwareWatcher()
	view1 := newMockViewable(newMockRenderable(), w1)
	app.SetRootView(view1)

	app.SetRoot(newMockRenderable())

	select {
	case <-w1.stopped:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected previous root watcher to be stopped")
	}
}
