//go:build !windows

package tui

import (
	"sync/atomic"
	"testing"
)

// recordingTerminal wraps MockTerminal and records method calls in order.
type recordingTerminal struct {
	*MockTerminal
	calls []string
}

func newRecordingTerminal(width, height int) *recordingTerminal {
	return &recordingTerminal{
		MockTerminal: NewMockTerminal(width, height),
	}
}

func (r *recordingTerminal) DisableMouse() {
	r.calls = append(r.calls, "DisableMouse")
	r.MockTerminal.DisableMouse()
}

func (r *recordingTerminal) ShowCursor() {
	r.calls = append(r.calls, "ShowCursor")
	r.MockTerminal.ShowCursor()
}

func (r *recordingTerminal) HideCursor() {
	r.calls = append(r.calls, "HideCursor")
	r.MockTerminal.HideCursor()
}

func (r *recordingTerminal) ExitAltScreen() {
	r.calls = append(r.calls, "ExitAltScreen")
	r.MockTerminal.ExitAltScreen()
}

func (r *recordingTerminal) EnterAltScreen() {
	r.calls = append(r.calls, "EnterAltScreen")
	r.MockTerminal.EnterAltScreen()
}

func (r *recordingTerminal) ExitRawMode() error {
	r.calls = append(r.calls, "ExitRawMode")
	return r.MockTerminal.ExitRawMode()
}

func (r *recordingTerminal) EnterRawMode() error {
	r.calls = append(r.calls, "EnterRawMode")
	return r.MockTerminal.EnterRawMode()
}

func (r *recordingTerminal) EnableMouse() {
	r.calls = append(r.calls, "EnableMouse")
	r.MockTerminal.EnableMouse()
}

func (r *recordingTerminal) Clear() {
	r.calls = append(r.calls, "Clear")
	r.MockTerminal.Clear()
}

func TestSuspendSequence_FullScreen(t *testing.T) {
	term := newRecordingTerminal(80, 24)
	term.inRawMode = true
	term.inAltScreen = true
	term.cursorHidden = true
	term.mouseEnabled = true

	suspendCalled := false
	app := &App{
		terminal:     term,
		mouseEnabled: true,
		onSuspend:    func() { suspendCalled = true },
		stopCh:       make(chan struct{}),
		buffer:       NewBuffer(80, 24),
	}

	// Call the testable part of suspend (without the actual SIGTSTP)
	app.suspendTerminal()

	if !suspendCalled {
		t.Fatal("expected onSuspend to be called")
	}

	// Verify call order
	expected := []string{"DisableMouse", "ShowCursor", "ExitAltScreen", "ExitRawMode"}
	if len(term.calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(term.calls), term.calls)
	}
	for i, call := range expected {
		if term.calls[i] != call {
			t.Errorf("call[%d] = %q, want %q", i, term.calls[i], call)
		}
	}
}

func TestSuspendSequence_InlineMode(t *testing.T) {
	term := newRecordingTerminal(80, 24)
	term.inRawMode = true
	term.cursorHidden = true

	app := &App{
		terminal:     term,
		inlineHeight: 5,
		stopCh:       make(chan struct{}),
		buffer:       NewBuffer(80, 5),
	}

	app.suspendTerminal()

	// Inline mode: should NOT call ExitAltScreen
	for _, call := range term.calls {
		if call == "ExitAltScreen" {
			t.Fatal("should not exit alt screen in inline mode")
		}
	}

	// Should still exit raw mode
	found := false
	for _, call := range term.calls {
		if call == "ExitRawMode" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected ExitRawMode in inline mode")
	}
}

func TestResumeSequence_FullScreen(t *testing.T) {
	term := newRecordingTerminal(80, 24)

	app := &App{
		terminal:     term,
		mouseEnabled: true,
		stopCh:       make(chan struct{}),
		buffer:       NewBuffer(80, 24),
		dirty:        atomic.Bool{},
	}

	resumeCalled := false
	app.onResume = func() { resumeCalled = true }

	app.resumeTerminal()

	if !resumeCalled {
		t.Fatal("expected onResume to be called")
	}

	expected := []string{"EnterRawMode", "EnterAltScreen", "Clear", "HideCursor", "EnableMouse"}
	if len(term.calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(term.calls), term.calls)
	}
	for i, call := range expected {
		if term.calls[i] != call {
			t.Errorf("call[%d] = %q, want %q", i, term.calls[i], call)
		}
	}

	if !app.needsFullRedraw {
		t.Fatal("expected needsFullRedraw to be set")
	}
}

func TestResumeSequence_InlineMode(t *testing.T) {
	term := newRecordingTerminal(80, 24)

	app := &App{
		terminal:     term,
		inlineHeight: 5,
		stopCh:       make(chan struct{}),
		buffer:       NewBuffer(80, 5),
		dirty:        atomic.Bool{},
	}

	app.resumeTerminal()

	// Inline mode: should NOT call EnterAltScreen or Clear
	for _, call := range term.calls {
		if call == "EnterAltScreen" || call == "Clear" {
			t.Fatalf("should not call %s in inline mode", call)
		}
	}
}

func TestResumeSequence_CursorVisible(t *testing.T) {
	term := newRecordingTerminal(80, 24)

	app := &App{
		terminal:      term,
		cursorVisible: true,
		stopCh:        make(chan struct{}),
		buffer:        NewBuffer(80, 24),
		dirty:         atomic.Bool{},
	}

	app.resumeTerminal()

	// Should NOT hide cursor when cursorVisible is true
	for _, call := range term.calls {
		if call == "HideCursor" {
			t.Fatal("should not hide cursor when cursorVisible is true")
		}
	}
}

func TestSuspendSequence_MouseDisabled(t *testing.T) {
	term := newRecordingTerminal(80, 24)
	term.inRawMode = true
	term.inAltScreen = true

	app := &App{
		terminal:     term,
		mouseEnabled: false,
		stopCh:       make(chan struct{}),
		buffer:       NewBuffer(80, 24),
	}

	app.suspendTerminal()

	// Should NOT call DisableMouse when mouse is not enabled
	for _, call := range term.calls {
		if call == "DisableMouse" {
			t.Fatal("should not call DisableMouse when mouse is not enabled")
		}
	}
}

func TestSuspendResume_HooksNilSafe(t *testing.T) {
	term := newRecordingTerminal(80, 24)
	term.inRawMode = true

	app := &App{
		terminal: term,
		stopCh:   make(chan struct{}),
		buffer:   NewBuffer(80, 24),
		dirty:    atomic.Bool{},
	}

	// Should not panic with nil hooks
	app.suspendTerminal()
	app.resumeTerminal()
}
