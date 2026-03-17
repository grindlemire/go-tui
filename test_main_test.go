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
		stopCh:       make(chan struct{}),
		merged:       make(chan Event, 1),
		watcherQueue: make(chan func(), 1),
		focus:        newFocusManager(),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}
	os.Exit(m.Run())
}
