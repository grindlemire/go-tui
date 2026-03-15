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

	if a.inAlternateScreen {
		// Dynamic alternate screen overlay: exit overlay first, then
		// handle the underlying mode (inline or full-screen).
		a.terminal.ExitAltScreen()
		if a.savedInlineHeight > 0 {
			a.terminal.SetCursor(0, a.savedInlineStartRow)
			a.terminal.ClearToEnd()
		}
	} else if a.inlineHeight > 0 {
		// Inline mode: clear the widget area and position the cursor there.
		// The scrollback history above the widget is untouched. Shell job
		// control messages ("Stopped", "fg") appear where the widget was.
		// On resume, the widget redraws at the recalculated bottom position.
		a.terminal.SetCursor(0, a.inlineStartRow)
		a.terminal.ClearToEnd()
	} else {
		// Full-screen mode: exit alternate screen
		a.terminal.ExitAltScreen()
	}

	// Disable Kitty keyboard protocol (pop from stack)
	a.terminal.DisableKittyKeyboard()

	a.terminal.ExitRawMode()
}

// resumeTerminal restores terminal state after process resumption.
// Must be called from the main event loop.
func (a *App) resumeTerminal() {
	a.terminal.EnterRawMode()

	// Re-enable Kitty keyboard protocol. We use EnableKittyKeyboard (push
	// without query) instead of NegotiateKittyKeyboard to avoid a stdin
	// query/response race: after SIGCONT the terminal may be slow to respond,
	// and a late response leaks onto stdin where the reader parses it as
	// keypresses (e.g., "[?1u" typed into a textarea). Since we already
	// negotiated successfully at startup, we know the terminal supports it.
	if !a.legacyKeyboard {
		a.terminal.EnableKittyKeyboard()
	}

	if a.inAlternateScreen {
		// Dynamic alternate screen overlay: recalculate saved inline
		// geometry (if the underlying mode was inline), then re-enter
		// the overlay alt screen.
		if a.savedInlineHeight > 0 {
			_, termHeight := a.terminal.Size()
			a.savedInlineStartRow = termHeight - a.savedInlineHeight
			if a.savedInlineStartRow < 0 {
				a.savedInlineStartRow = 0
			}
		}
		a.terminal.EnterAltScreen()
		a.terminal.Clear()
	} else if a.inlineHeight > 0 {
		// Inline mode: the shell printed job control messages while stopped.
		// Recalculate where the widget should be drawn.
		_, termHeight := a.terminal.Size()
		a.inlineStartRow = termHeight - a.inlineHeight
		if a.inlineStartRow < 0 {
			a.inlineStartRow = 0
		}
		// Reset style tracking: the terminal's SGR state is unknown after
		// going through cooked mode and shell interaction. Without this,
		// Flush may skip emitting style codes for cells whose style matches
		// the stale lastStyle, producing wrong colors on the first frame.
		a.terminal.ResetStyle()
	} else {
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
	a.selfSuspended.Store(true)
	a.suspendTerminal()

	// Stop the process. Execution pauses here until SIGCONT.
	// SIGTSTP disposition is SIG_DFL since we never called signal.Notify for it.
	syscall.Kill(syscall.Getpid(), syscall.SIGTSTP)

	// Process has been resumed by SIGCONT.
	// Resume inline to avoid a race with the event queue.
	a.resumeTerminal()

	// Note: selfSuspended is cleared by the SIGCONT handler goroutine via
	// CompareAndSwap. If the handler hasn't run yet, clear it as a fallback.
	a.selfSuspended.Store(false)
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
				// Use CompareAndSwap to avoid a race with suspend().
				// After SIGCONT, both this goroutine and the main goroutine
				// (in suspend()) resume simultaneously. If we used Load(),
				// suspend() might clear the flag before we check it, causing
				// a spurious double-resume. CAS atomically checks and clears,
				// so exactly one side wins.
				if a.selfSuspended.CompareAndSwap(true, false) {
					// Self-initiated suspend: suspend() calls
					// resumeTerminal() inline. Nothing to do here.
					continue
				}
				// SIGCONT after an external SIGTSTP (kill -TSTP).
				// The terminal is likely in a bad state (cooked mode,
				// no alt screen, cursor visible, mouse disabled).
				// Run the full resume sequence on the event loop.
				select {
				case a.eventQueue <- func() {
					a.resumeTerminal()
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
