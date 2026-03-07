//go:build !windows

package tui

import (
	"os"
	"os/signal"
	"syscall"
)

// suspendTerminal tears down terminal state before process suspension.
// Must be called from the main event loop.
func (a *App) suspendTerminal() {
	if a.onSuspend != nil {
		a.onSuspend()
	}

	if a.mouseEnabled {
		a.terminal.DisableMouse()
	}

	a.terminal.ShowCursor()

	// Exit alternate screen only in full-screen mode
	if a.inlineHeight == 0 && !a.inAlternateScreen {
		a.terminal.ExitAltScreen()
	}

	a.terminal.ExitRawMode()
}

// resumeTerminal restores terminal state after process resumption.
// Must be called from the main event loop.
func (a *App) resumeTerminal() {
	a.terminal.EnterRawMode()

	if a.inlineHeight == 0 && !a.inAlternateScreen {
		a.terminal.EnterAltScreen()
		a.terminal.Clear()
	}

	if !a.cursorVisible {
		a.terminal.HideCursor()
	}

	if a.mouseEnabled {
		a.terminal.EnableMouse()
	}

	a.needsFullRedraw = true
	a.MarkDirty()

	if a.onResume != nil {
		a.onResume()
	}
}

// suspend performs the full suspend sequence: tear down terminal, send SIGTSTP.
// Must be called from the main event loop (via eventQueue).
func (a *App) suspend() {
	a.suspendTerminal()

	// Restore default SIGTSTP handler so the kill actually stops the process
	signal.Reset(syscall.SIGTSTP)

	// Stop the process. Execution pauses here until SIGCONT.
	syscall.Kill(syscall.Getpid(), syscall.SIGTSTP)

	// Process has been resumed by SIGCONT.
	// Resume inline to avoid a race with the event queue.
	a.resumeTerminal()
}

// Suspend programmatically triggers a suspend (same as Ctrl+Z).
// Safe to call from any goroutine.
func (a *App) Suspend() {
	select {
	case a.eventQueue <- func() { a.suspend() }:
	case <-a.stopCh:
	}
}

// registerSuspendSignals sets up SIGTSTP signal handling.
// Returns a cleanup function to call when the app stops.
func (a *App) registerSuspendSignals() func() {
	suspendCh := make(chan os.Signal, 1)
	signal.Notify(suspendCh, syscall.SIGTSTP)

	go func() {
		for {
			select {
			case <-suspendCh:
				select {
				case a.eventQueue <- func() { a.suspend() }:
				case <-a.stopCh:
					return
				}
			case <-a.stopCh:
				return
			}
		}
	}()

	return func() {
		signal.Stop(suspendCh)
	}
}
