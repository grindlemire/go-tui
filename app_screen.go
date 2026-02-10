package tui

// EnterAlternateScreen switches the currently running app to alternate screen mode.
// Safe to call even if no app is running.
func EnterAlternateScreen() error {
	if currentApp == nil {
		return nil
	}
	return currentApp.EnterAlternateScreen()
}

// ExitAlternateScreen returns the currently running app to its previous screen mode.
// Safe to call even if no app is running.
func ExitAlternateScreen() error {
	if currentApp == nil {
		return nil
	}
	return currentApp.ExitAlternateScreen()
}

// IsInAlternateScreen reports whether the currently running app is in alternate
// screen mode. Returns false when no app is running.
func IsInAlternateScreen() bool {
	if currentApp == nil {
		return false
	}
	return currentApp.IsInAlternateScreen()
}

// EnterAlternateScreen switches to alternate screen mode for full-screen UI.
// The current scrollback position is preserved and will be restored on exit.
// Use this for overlays like settings panels that should not affect terminal history.
// If already in alternate mode, this is a no-op.
func (a *App) EnterAlternateScreen() error {
	// Guard: already in alternate mode
	if a.inAlternateScreen {
		return nil
	}

	// Save current inline mode state for restoration
	a.savedInlineHeight = a.inlineHeight
	a.savedInlineStartRow = a.inlineStartRow
	a.savedInlineLayout = a.inlineLayout

	// Get terminal dimensions
	width, height := a.terminal.Size()

	// If currently in inline mode, clear the inline region first
	if a.inlineHeight > 0 {
		a.terminal.SetCursor(0, a.inlineStartRow)
		a.terminal.ClearToEnd()
	}

	// Enter alternate screen via terminal
	a.terminal.EnterAltScreen()

	// Update internal state to full-screen mode
	a.inAlternateScreen = true
	a.inlineHeight = 0
	a.inlineStartRow = 0

	// Resize buffer to full terminal size
	a.buffer.Resize(width, height)

	// Mark for full redraw
	if a.root != nil {
		a.root.MarkDirty()
	}
	a.needsFullRedraw = true
	MarkDirty()

	return nil
}

// ExitAlternateScreen returns to normal mode, restoring the previous scrollback.
// If not in alternate mode, this is a no-op.
func (a *App) ExitAlternateScreen() error {
	// Guard: not in alternate mode
	if !a.inAlternateScreen {
		return nil
	}

	// Exit alternate screen - this restores the original terminal content
	a.terminal.ExitAltScreen()

	// Restore inline mode state
	a.inlineHeight = a.savedInlineHeight
	a.inlineStartRow = a.savedInlineStartRow
	a.inlineLayout = a.savedInlineLayout
	a.inAlternateScreen = false

	// Get terminal dimensions
	width, height := a.terminal.Size()

	// Reconfigure buffer for restored mode
	if a.inlineHeight > 0 {
		// Inline mode: recalculate start row, resize buffer to inline height
		a.inlineStartRow = height - a.inlineHeight
		a.inlineLayout.clamp(a.inlineStartRow)
		a.buffer.Resize(width, a.inlineHeight)
	} else {
		// Was full-screen normal mode (started with alternate screen)
		// Re-enter alternate screen since that's the normal state for full-screen apps
		a.terminal.EnterAltScreen()
		a.buffer.Resize(width, height)
	}

	// Mark for full redraw
	if a.root != nil {
		a.root.MarkDirty()
	}
	a.needsFullRedraw = true
	MarkDirty()

	return nil
}

// IsInAlternateScreen returns true if currently in alternate screen mode.
func (a *App) IsInAlternateScreen() bool {
	return a.inAlternateScreen
}
