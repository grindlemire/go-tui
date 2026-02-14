package tui

import (
	"os"
	"testing"
)

// testApp is a lightweight App used by all unit tests.
// It is created in TestMain before any tests run.
var testApp *App

func TestMain(m *testing.M) {
	testApp = &App{
		stopCh:      make(chan struct{}),
		eventQueue:  make(chan func(), 1),
		updateQueue: make(chan func(), 1),
		focus:       NewFocusManager(),
		mounts:      newMountState(),
		batch:       newBatchContext(),
	}
	os.Exit(m.Run())
}
