package tui

import (
	"fmt"

	"github.com/grindlemire/go-tui/internal/debug"
)

// Quit stops the currently running app. This is an alias for Stop().
func Quit() {
	Stop()
}

// Stop stops the currently running app. This is a package-level convenience function
// that allows stopping the app from event handlers without needing a direct reference.
// It is safe to call even if no app is running.
func Stop() {
	if currentApp != nil {
		currentApp.Stop()
	}
}

// PrintAbove prints content above the inline widget without a trailing newline.
// Only works in inline mode. Safe to call even if no app is running.
func PrintAbove(format string, args ...any) {
	if currentApp != nil {
		currentApp.PrintAbove(format, args...)
	}
}

// QueuePrintAbove queues content to print above the inline widget without a
// trailing newline.
// This variant is goroutine-safe and executes on the app event loop.
func QueuePrintAbove(format string, args ...any) {
	if currentApp != nil {
		currentApp.QueuePrintAbove(format, args...)
	}
}

// PrintAboveln prints content with a trailing newline above the inline widget.
// Only works in inline mode. Safe to call even if no app is running.
func PrintAboveln(format string, args ...any) {
	if currentApp != nil {
		currentApp.PrintAboveln(format, args...)
	}
}

// QueuePrintAboveln queues content with a trailing newline above the inline
// widget.
// This variant is goroutine-safe and executes on the app event loop.
func QueuePrintAboveln(format string, args ...any) {
	if currentApp != nil {
		currentApp.QueuePrintAboveln(format, args...)
	}
}

// PrintAboveAsync queues content above the inline widget without a trailing
// newline.
// Deprecated: use QueuePrintAbove.
func PrintAboveAsync(format string, args ...any) {
	QueuePrintAbove(format, args...)
}

// PrintAbovelnAsync queues content with a trailing newline above the inline
// widget.
// Deprecated: use QueuePrintAboveln.
func PrintAbovelnAsync(format string, args ...any) {
	QueuePrintAboveln(format, args...)
}

// SetInlineHeight changes the inline widget height at runtime.
// Only works in inline mode. Safe to call even if no app is running.
func SetInlineHeight(rows int) {
	if currentApp != nil {
		currentApp.SetInlineHeight(rows)
	}
}

// SnapshotFrame returns the current frame as a string for debugging.
// Returns an empty string if no app is running.
func SnapshotFrame() string {
	if currentApp != nil && currentApp.buffer != nil {
		return currentApp.buffer.StringTrimmed()
	}
	return ""
}

// Close restores the terminal to its original state.
// Must be called when the application exits.
func (a *App) Close() error {
	// Component watchers are stopped via stopCh (closed by Stop()).
	// No explicit cleanup needed here - they exit when stopCh closes.

	// Disable mouse event reporting (only if it was enabled)
	if a.mouseEnabled {
		a.terminal.DisableMouse()
	}

	// Show cursor (only if it was hidden)
	if !a.cursorVisible {
		a.terminal.ShowCursor()
	}

	// Handle screen cleanup based on mode
	if a.inAlternateScreen {
		// Currently in alternate screen overlay: exit alternate screen first
		a.terminal.ExitAltScreen()
		// Then handle based on the original mode (before entering alternate)
		if a.savedInlineHeight > 0 {
			// Was inline mode: clear the inline area
			a.terminal.SetCursor(0, a.savedInlineStartRow)
			a.terminal.ClearToEnd()
		}
		// If savedInlineHeight == 0, we were in full-screen mode which means
		// alternate screen was the normal state, so exiting it is sufficient
	} else if a.inlineHeight > 0 {
		// Inline mode: clear the widget area and position cursor for shell
		a.terminal.SetCursor(0, a.inlineStartRow)
		a.terminal.ClearToEnd()
	} else {
		// Full screen mode: exit alternate screen
		a.terminal.ExitAltScreen()
	}

	// Exit raw mode
	if err := a.terminal.ExitRawMode(); err != nil {
		a.reader.Close()
		return err
	}

	// Close EventReader
	return a.reader.Close()
}

// PrintAbove prints content that scrolls up above the inline widget.
// Does not add a trailing newline. Use PrintAboveln for auto-newline.
// Only works in inline mode (WithInlineHeight). In full-screen mode, this is a no-op.
// Must be called from the app's main loop.
func (a *App) PrintAbove(format string, args ...any) {
	a.printAboveFormatted(false, false, format, args...)
}

// QueuePrintAbove queues content to print above the inline widget without a
// trailing newline.
// Safe to call from any goroutine.
func (a *App) QueuePrintAbove(format string, args ...any) {
	a.printAboveFormatted(true, false, format, args...)
}

// PrintAboveln prints content with a trailing newline that scrolls up above the inline widget.
// Only works in inline mode (WithInlineHeight). In full-screen mode, this is a no-op.
// Must be called from the app's main loop.
func (a *App) PrintAboveln(format string, args ...any) {
	a.printAboveFormatted(false, true, format, args...)
}

// QueuePrintAboveln queues content with a trailing newline that scrolls up
// above the inline widget.
// Safe to call from any goroutine.
func (a *App) QueuePrintAboveln(format string, args ...any) {
	a.printAboveFormatted(true, true, format, args...)
}

// PrintAboveAsync queues content above the inline widget without a trailing
// newline.
// Deprecated: use QueuePrintAbove.
func (a *App) PrintAboveAsync(format string, args ...any) {
	a.QueuePrintAbove(format, args...)
}

// PrintAbovelnAsync queues content with a trailing newline that scrolls up
// above the inline widget.
// Deprecated: use QueuePrintAboveln.
func (a *App) PrintAbovelnAsync(format string, args ...any) {
	a.QueuePrintAboveln(format, args...)
}

func (a *App) printAboveFormatted(async, trailingNewline bool, format string, args ...any) {
	if a.inlineHeight == 0 {
		return
	}

	content := fmt.Sprintf(format, args...)
	if trailingNewline {
		content += "\n"
	}

	if async {
		a.QueueUpdate(func() {
			a.printAboveRaw(content)
		})
		return
	}

	a.printAboveRaw(content)
}

// SetInlineHeight changes the inline widget height at runtime.
// Only works in inline mode (WithInlineHeight was used at creation).
// The height change takes effect immediately.
// Should be called from render functions or the main event loop.
func (a *App) SetInlineHeight(rows int) {
	if a.inlineHeight == 0 {
		return // Not in inline mode
	}
	if rows < 1 {
		rows = 1
	}

	// Get current terminal size
	width, termHeight := a.terminal.Size()

	// Cap to terminal height
	if rows > termHeight {
		rows = termHeight
	}

	// Only update if height actually changed
	if rows == a.inlineHeight {
		debug.Log("SetInlineHeight: no change needed (already %d)", rows)
		return
	}

	oldHeight := a.inlineHeight
	oldStartRow := a.inlineStartRow
	newStartRow := termHeight - rows

	debug.Log("SetInlineHeight: changing from %d to %d (termHeight=%d, width=%d)", oldHeight, rows, termHeight, width)
	a.ensureInlineSession()
	a.inlineSession.ensureInitialized(&a.inlineLayout, oldStartRow)
	a.inlineSession.resize(&a.inlineLayout, oldStartRow, oldHeight, newStartRow)

	a.inlineHeight = rows
	a.inlineStartRow = newStartRow
	a.buffer.Resize(width, rows)
	a.needsFullRedraw = true // Terminal position shifted, need full redraw
	debug.Log("SetInlineHeight: buffer resized, new inlineStartRow=%d, needsFullRedraw=true", a.inlineStartRow)
}

// InlineHeight returns the current inline height (0 if not in inline mode).
func (a *App) InlineHeight() int {
	return a.inlineHeight
}

// printAboveRaw handles the actual printing and scrolling for inline mode.
// Prints content that scrolls into terminal scrollback buffer, allowing
// the user to scroll back through history with their terminal's scroll feature.
// Must be called from the main event loop.
func (a *App) printAboveRaw(content string) {
	if a.inlineStartRow < 1 {
		return // No room above widget
	}
	a.ensureInlineSession()
	a.inlineSession.ensureInitialized(&a.inlineLayout, a.inlineStartRow)
	width, _ := a.terminal.Size()
	a.inlineSession.appendText(&a.inlineLayout, a.inlineStartRow, width, content)

	// Mark dirty to ensure consistent state
	MarkDirty()
}

func (a *App) ensureInlineSession() {
	if a.inlineSession == nil {
		a.inlineSession = newInlineSession(a.terminal)
	}
}
