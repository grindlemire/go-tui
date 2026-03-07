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

	if a.inlineHeight > 0 {
		// Inline mode: clear the widget area and position cursor there,
		// so the shell's "[1]+ Stopped" message appears cleanly below
		// the scrollback content instead of overwriting the widget.
		a.terminal.SetCursor(0, a.inlineStartRow)
		a.terminal.ClearToEnd()
	} else if !a.inAlternateScreen {
		// Full-screen mode: exit alternate screen
		a.terminal.ExitAltScreen()
	}

	a.terminal.ExitRawMode()
}

// resumeTerminal restores terminal state after process resumption.
// Must be called from the main event loop.
func (a *App) resumeTerminal() {
	a.terminal.EnterRawMode()

	if a.inlineHeight > 0 {
		// Inline mode: the shell printed job control messages while stopped.
		// Recalculate where the widget should be drawn.
		_, termHeight := a.terminal.Size()
		a.inlineStartRow = termHeight - a.inlineHeight
		if a.inlineStartRow < 0 {
			a.inlineStartRow = 0
		}
	} else if !a.inAlternateScreen {
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
//
// We never register signal.Notify for SIGTSTP, so its disposition remains at
// the OS default (stop the process). signal.Reset after Notify doesn't reliably
// restore SIG_DFL in Go's runtime, so avoiding Notify entirely is the fix.
func (a *App) suspend() {
	a.suspendTerminal()

	// Stop the process. Execution pauses here until SIGCONT.
	// SIGTSTP disposition is SIG_DFL since we never called signal.Notify for it.
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

// registerSuspendSignals sets up a SIGCONT handler to restore terminal state
// when the process is resumed after an external kill -TSTP (where we didn't
// get to run suspendTerminal/resumeTerminal ourselves).
// Returns a cleanup function to call when the app stops.
func (a *App) registerSuspendSignals() func() {
	contCh := make(chan os.Signal, 1)
	signal.Notify(contCh, syscall.SIGCONT)

	go func() {
		for {
			select {
			case <-contCh:
				// SIGCONT after an external SIGTSTP. The terminal may be
				// in a bad state since we didn't get to tear down cleanly.
				// Force a full redraw on the event loop.
				select {
				case a.eventQueue <- func() {
					a.needsFullRedraw = true
					a.MarkDirty()
					if a.onResume != nil {
						a.onResume()
					}
				}:
				case <-a.stopCh:
					return
				}
			case <-a.stopCh:
				return
			}
		}
	}()

	return func() {
		signal.Stop(contCh)
	}
}
