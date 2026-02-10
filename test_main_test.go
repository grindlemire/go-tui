package tui

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	prev := DefaultApp()
	testApp := &App{
		stopCh:      make(chan struct{}),
		eventQueue:  make(chan func(), 1),
		updateQueue: make(chan func(), 1),
		focus:       NewFocusManager(),
		mounts:      newMountState(),
		batch:       newBatchContext(),
	}
	SetDefaultApp(testApp)
	code := m.Run()
	SetDefaultApp(prev)
	os.Exit(code)
}
