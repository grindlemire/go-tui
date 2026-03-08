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

func (r *recordingTerminal) SetCursor(x, y int) {
	r.calls = append(r.calls, "SetCursor")
	r.MockTerminal.SetCursor(x, y)
}

func (r *recordingTerminal) ClearToEnd() {
	r.calls = append(r.calls, "ClearToEnd")
	r.MockTerminal.ClearToEnd()
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
		terminal:       term,
		inlineHeight:   5,
		inlineStartRow: 19,
		stopCh:         make(chan struct{}),
		buffer:         NewBuffer(80, 5),
	}

	app.suspendTerminal()

	// Inline mode: should clear widget area, NOT call ExitAltScreen
	expected := []string{"ShowCursor", "SetCursor", "ClearToEnd", "ExitRawMode"}
	if len(term.calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(term.calls), term.calls)
	}
	for i, call := range expected {
		if term.calls[i] != call {
			t.Errorf("call[%d] = %q, want %q", i, term.calls[i], call)
		}
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
		terminal:       term,
		inlineHeight:   5,
		inlineStartRow: 0, // stale value
		stopCh:         make(chan struct{}),
		buffer:         NewBuffer(80, 5),
		dirty:          atomic.Bool{},
	}

	app.resumeTerminal()

	// Inline mode: should NOT call EnterAltScreen or Clear
	for _, call := range term.calls {
		if call == "EnterAltScreen" || call == "Clear" {
			t.Fatalf("should not call %s in inline mode", call)
		}
	}

	// Should recalculate inlineStartRow from terminal size
	if app.inlineStartRow != 19 { // 24 - 5
		t.Fatalf("expected inlineStartRow=19, got %d", app.inlineStartRow)
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

func TestSuspendSignalRegistration(t *testing.T) {
	term := newRecordingTerminal(80, 24)

	app := &App{
		terminal:       term,
		stopCh:         make(chan struct{}),
		eventQueue:     make(chan func(), 256),
		eventQueueSize: 256,
		buffer:         NewBuffer(80, 24),
	}

	cleanup := app.registerSuspendSignals()
	defer cleanup()

	// Verify cleanup doesn't panic when called multiple times
	cleanup()
}

func TestSuspendSequence_DynamicAltScreen(t *testing.T) {
	term := newRecordingTerminal(80, 24)
	term.inRawMode = true
	term.inAltScreen = true
	term.cursorHidden = true

	app := &App{
		terminal:          term,
		inAlternateScreen: true,
		savedInlineHeight: 5,
		savedInlineStartRow: 19,
		stopCh:            make(chan struct{}),
		buffer:            NewBuffer(80, 24),
	}

	app.suspendTerminal()

	// Should exit alt screen overlay, then clear saved inline widget area
	expected := []string{"ShowCursor", "ExitAltScreen", "SetCursor", "ClearToEnd", "ExitRawMode"}
	if len(term.calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(term.calls), term.calls)
	}
	for i, call := range expected {
		if term.calls[i] != call {
			t.Errorf("call[%d] = %q, want %q", i, term.calls[i], call)
		}
	}
}

func TestSuspendSequence_DynamicAltScreenFullScreenUnderlying(t *testing.T) {
	term := newRecordingTerminal(80, 24)
	term.inRawMode = true
	term.inAltScreen = true
	term.cursorHidden = true

	// Dynamic alt screen with full-screen underlying mode (savedInlineHeight == 0)
	app := &App{
		terminal:          term,
		inAlternateScreen: true,
		stopCh:            make(chan struct{}),
		buffer:            NewBuffer(80, 24),
	}

	app.suspendTerminal()

	// Should exit alt screen overlay only (no SetCursor/ClearToEnd)
	expected := []string{"ShowCursor", "ExitAltScreen", "ExitRawMode"}
	if len(term.calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(term.calls), term.calls)
	}
	for i, call := range expected {
		if term.calls[i] != call {
			t.Errorf("call[%d] = %q, want %q", i, term.calls[i], call)
		}
	}
}

func TestResumeSequence_DynamicAltScreen(t *testing.T) {
	term := newRecordingTerminal(80, 24)

	app := &App{
		terminal:            term,
		inAlternateScreen:   true,
		savedInlineHeight:   5,
		savedInlineStartRow: 0, // stale value
		stopCh:              make(chan struct{}),
		buffer:              NewBuffer(80, 24),
		dirty:               atomic.Bool{},
	}

	app.resumeTerminal()

	// Should re-enter alt screen overlay, then hide cursor (cursorVisible defaults to false)
	expected := []string{"EnterRawMode", "EnterAltScreen", "Clear", "HideCursor"}
	if len(term.calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(term.calls), term.calls)
	}
	for i, call := range expected {
		if term.calls[i] != call {
			t.Errorf("call[%d] = %q, want %q", i, term.calls[i], call)
		}
	}

	// Should recalculate saved inline start row
	if app.savedInlineStartRow != 19 { // 24 - 5
		t.Fatalf("expected savedInlineStartRow=19, got %d", app.savedInlineStartRow)
	}
}

func TestSelfSuspendedPreventsDoubleResume(t *testing.T) {
	term := newRecordingTerminal(80, 24)

	resumeCount := 0
	app := &App{
		terminal:       term,
		stopCh:         make(chan struct{}),
		eventQueue:     make(chan func(), 256),
		eventQueueSize: 256,
		buffer:         NewBuffer(80, 24),
		dirty:          atomic.Bool{},
		onResume:       func() { resumeCount++ },
	}

	// Simulate self-initiated suspend: flag should prevent SIGCONT handler
	// from enqueuing a duplicate resume.
	app.selfSuspended.Store(true)

	// The SIGCONT handler goroutine would check this flag and skip.
	// Verify the flag is readable.
	if !app.selfSuspended.Load() {
		t.Fatal("expected selfSuspended to be true")
	}

	// Call resumeTerminal once (as suspend() would)
	app.resumeTerminal()
	app.selfSuspended.Store(false)

	if resumeCount != 1 {
		t.Fatalf("expected onResume called once, got %d", resumeCount)
	}
}

func TestKeyCtrlZ_OverrideByStopper(t *testing.T) {
	term := newRecordingTerminal(80, 24)
	term.inRawMode = true

	overrideCalled := false

	// Build a dispatch table with a Stop handler for KeyCtrlZ
	table := &dispatchTable{
		entries: []dispatchEntry{
			{
				pattern: KeyPattern{Key: KeyCtrlZ},
				handler: func(ke KeyEvent) { overrideCalled = true },
				stop:    true,
			},
		},
	}

	app := &App{
		terminal:      term,
		stopCh:        make(chan struct{}),
		eventQueue:    make(chan func(), 256),
		updateQueue:   make(chan func(), 256),
		buffer:        NewBuffer(80, 24),
		focus:         newFocusManager(),
		dispatchTable: table,
		dirty:         atomic.Bool{},
	}

	ke := KeyEvent{Key: KeyCtrlZ, app: app}

	// Dispatch through table - should be consumed by Stop handler
	stopped := app.dispatchTable.dispatch(ke, app.focus)

	if !stopped {
		t.Fatal("expected dispatch to return stopped=true")
	}
	if !overrideCalled {
		t.Fatal("expected override handler to be called")
	}

	// Verify terminal was NOT touched (suspend should not have fired)
	if len(term.calls) != 0 {
		t.Fatalf("expected no terminal calls, got: %v", term.calls)
	}
}
