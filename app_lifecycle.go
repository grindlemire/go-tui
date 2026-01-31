package tui

import (
	"fmt"
	"strings"
)

// Stop stops the currently running app. This is a package-level convenience function
// that allows stopping the app from event handlers without needing a direct reference.
// It is safe to call even if no app is running.
func Stop() {
	if currentApp != nil {
		currentApp.Stop()
	}
}

// Close restores the terminal to its original state.
// Must be called when the application exits.
func (a *App) Close() error {
	// Disable mouse event reporting (only if it was enabled)
	if a.mouseEnabled {
		a.terminal.DisableMouse()
	}

	// Show cursor (only if it was hidden)
	if !a.cursorVisible {
		a.terminal.ShowCursor()
	}

	// Handle screen cleanup based on mode
	if a.inlineHeight > 0 {
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
func (a *App) printAboveRaw(content string) {
	if a.inlineStartRow < 1 {
		return // No room above widget
	}

	text := strings.TrimSuffix(content, "\n")

	// To get content into terminal scrollback:
	// 1. Use reverse index (ESC M) at top of screen to scroll down, creating
	//    space at top - but this doesn't help us.
	// 2. Actually, we need to scroll UP to push content into scrollback.
	//
	// The approach: position at the line just above widget and use
	// scroll up (ESC[S) which scrolls the whole screen up, pushing
	// the top line into scrollback. Then print our text.
	//
	// inlineStartRow is 0-indexed. Widget starts at ANSI row inlineStartRow+1.

	var seq strings.Builder
	// Scroll entire screen up by 1 line (top line goes to scrollback)
	seq.WriteString("\033[1S")
	// Move to the row just above widget (which is now blank after scroll)
	seq.WriteString(fmt.Sprintf("\033[%d;1H", a.inlineStartRow))
	// Print text
	seq.WriteString(text)

	a.terminal.WriteDirect([]byte(seq.String()))

	// Force widget redraw (scroll affected it)
	a.needsFullRedraw = true
	MarkDirty()
}
