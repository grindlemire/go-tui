package tui

import (
	"fmt"
	"strings"
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
// Safe to call from any goroutine.
func (a *App) PrintAbove(format string, args ...any) {
	if a.inlineHeight == 0 {
		return
	}
	content := fmt.Sprintf(format, args...)
	a.QueueUpdate(func() {
		a.printAboveRaw(content)
	})
}

// PrintAboveln prints content with a trailing newline that scrolls up above the inline widget.
// Only works in inline mode (WithInlineHeight). In full-screen mode, this is a no-op.
// Safe to call from any goroutine.
func (a *App) PrintAboveln(format string, args ...any) {
	if a.inlineHeight == 0 {
		return
	}
	content := fmt.Sprintf(format, args...) + "\n"
	a.QueueUpdate(func() {
		a.printAboveRaw(content)
	})
}

// printAboveRaw handles the actual printing and scrolling for inline mode.
// Prints content that scrolls into terminal scrollback buffer, allowing
// the user to scroll back through history with their terminal's scroll feature.
// Must be called from the main event loop (via QueueUpdate).
//
// Note: The first N lines printed (where N = rows above widget) will push
// blank lines into scrollback. This is a known limitation of the current
// approach. After N lines, actual content starts appearing in scrollback.
func (a *App) printAboveRaw(content string) {
	if a.inlineStartRow < 1 {
		return // No room above widget
	}

	text := strings.TrimSuffix(content, "\n")

	// Use a scroll region to protect the widget at the bottom.
	// ANSI escape sequences use 1-indexed rows.
	// inlineStartRow is 0-indexed, so the widget occupies rows
	// (inlineStartRow+1) through termHeight in ANSI terms.

	var seq strings.Builder

	// Set scroll region to exclude widget area (rows 1 to inlineStartRow)
	seq.WriteString(fmt.Sprintf("\033[1;%dr", a.inlineStartRow))

	// Move to bottom of scroll region (last row before widget)
	seq.WriteString(fmt.Sprintf("\033[%d;1H", a.inlineStartRow))

	// Print text followed by newline
	// The newline scrolls content up within the region, pushing top line to scrollback
	seq.WriteString(text)
	seq.WriteString("\n")

	// Reset scroll region to full screen
	seq.WriteString("\033[r")

	a.terminal.WriteDirect([]byte(seq.String()))

	// Mark dirty to ensure consistent state
	MarkDirty()
}
