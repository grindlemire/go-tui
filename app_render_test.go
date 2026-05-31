package tui

import (
	"testing"
)

func TestPostRenderHook_FiresInRenderFrame(t *testing.T) {
	called := 0
	a := &App{
		stopCh:       make(chan struct{}),
		merged:       make(chan Event, 1),
		watcherQueue: make(chan func(), 1),
		focus:        newFocusManager(),
		mounts:       newMountState(),
		batch:        newBatchContext(),
		PostRenderHook: func() {
			called++
		},
	}
	a.buffer = NewBuffer(80, 24)

	a.renderFrame()

	if called != 1 {
		t.Fatalf("PostRenderHook called %d times, want 1", called)
	}
}

func TestPostRenderHook_FiresInRenderFull(t *testing.T) {
	called := 0
	a := &App{
		stopCh:       make(chan struct{}),
		merged:       make(chan Event, 1),
		watcherQueue: make(chan func(), 1),
		focus:        newFocusManager(),
		mounts:       newMountState(),
		batch:        newBatchContext(),
		PostRenderHook: func() {
			called++
		},
	}
	a.buffer = NewBuffer(80, 24)

	a.RenderFull()

	if called != 1 {
		t.Fatalf("PostRenderHook called %d times, want 1", called)
	}
}
