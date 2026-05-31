package tui

import (
	"testing"
)

func TestPostRenderHook_FiresInRenderFrame(t *testing.T) {
	called := 0
	a := &App{
		terminal:   NewMockTerminal(80, 24),
		focus:      newFocusManager(),
		buffer:     NewBuffer(80, 24),
		merged:     make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		mounts:     newMountState(),
		batch:      newBatchContext(),
		postRenderHook: func() {
			called++
		},
	}

	a.renderFrame()

	if called != 1 {
		t.Fatalf("postRenderHook called %d times, want 1", called)
	}
}

func TestPostRenderHook_FiresInRenderFull(t *testing.T) {
	called := 0
	a := &App{
		terminal:   NewMockTerminal(80, 24),
		focus:      newFocusManager(),
		buffer:     NewBuffer(80, 24),
		merged:     make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		mounts:     newMountState(),
		batch:      newBatchContext(),
		postRenderHook: func() {
			called++
		},
	}

	a.RenderFull()

	if called != 1 {
		t.Fatalf("postRenderHook called %d times, want 1", called)
	}
}
