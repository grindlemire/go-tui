package tui

import (
	"fmt"
	"strings"

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

// PrintAboveln prints content with a trailing newline above the inline widget.
// Only works in inline mode. Safe to call even if no app is running.
func PrintAboveln(format string, args ...any) {
	if currentApp != nil {
		currentApp.PrintAboveln(format, args...)
	}
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

	if rows > oldHeight {
		// Growing: need to make room by shifting history up.
		//
		// Two-phase approach:
		// Phase 1: Delete blank rows at top of history area using ANSI Delete Line.
		//   DL removes lines without pushing them to scrollback — they just vanish.
		//   Content below shifts up naturally to fill the gap.
		// Phase 2: If growth exceeds blank space, scroll remaining content rows
		//   into scrollback using a scroll region (these ARE preserved in scrollback).
		a.clearWidgetArea(oldStartRow, oldHeight)
		linesToScroll := rows - oldHeight
		blankRows := oldStartRow - a.historyRows

		if linesToScroll <= blankRows {
			// Enough blank space — just delete blank rows, content shifts up
			a.deleteBlankHistoryRows(linesToScroll, oldStartRow)
		} else {
			// Delete all blanks first, then scroll content for the rest
			if blankRows > 0 {
				a.deleteBlankHistoryRows(blankRows, oldStartRow)
			}
			remaining := linesToScroll - blankRows
			// After deleting blanks, content starts at row 0
			a.scrollHistoryUp(remaining, oldStartRow)
			a.historyRows -= remaining
		}
	} else {
		// Shrinking: We need to handle the "released" rows (the rows that were part of
		// the old widget but won't be part of the new smaller widget).
		//
		// The challenge: These rows are now in the history area. If we leave them blank,
		// they'll scroll into the scrollback mixed with actual messages.
		//
		// Solution: Use Reverse Index to scroll content DOWN, which:
		// 1. Inserts blank lines at the TOP of the screen
		// 2. Pushes existing history DOWN to fill the released rows
		// 3. The old widget blanks at the bottom fall off the screen
		//
		// This way, blanks are at the TOP and scroll into scrollback FIRST (before
		// messages), appearing at the "oldest" end of scroll history.
		a.clearWidgetArea(oldStartRow, oldHeight)
		releasedRows := oldHeight - rows
		a.scrollContentDown(releasedRows)
	}

	a.inlineHeight = rows
	a.inlineStartRow = newStartRow
	a.buffer.Resize(width, rows)
	a.needsFullRedraw = true // Terminal position shifted, need full redraw
	debug.Log("SetInlineHeight: buffer resized, new inlineStartRow=%d, needsFullRedraw=true", a.inlineStartRow)
}

// scrollHistoryUp scrolls the history area up by n lines to make room for widget growth.
// This uses a scroll region to push content into scrollback. The caller should ensure
// blank rows have been removed first (via deleteBlankHistoryRows) so that only content
// rows are pushed to scrollback.
func (a *App) scrollHistoryUp(n int, oldStartRow int) {
	if oldStartRow < 1 {
		return
	}

	var seq strings.Builder

	// Set scroll region to the history area (rows 1 to oldStartRow, 1-indexed)
	seq.WriteString(fmt.Sprintf("\033[1;%dr", oldStartRow))

	// Move to bottom of scroll region and emit newlines to scroll up
	seq.WriteString(fmt.Sprintf("\033[%d;1H", oldStartRow))
	for i := 0; i < n; i++ {
		seq.WriteString("\n")
	}

	// Reset scroll region to full screen
	seq.WriteString("\033[r")

	a.terminal.WriteDirect([]byte(seq.String()))
}

// clearWidgetArea clears the entire widget area before resizing.
// This prevents widget content (borders, text) from being scrolled into history.
func (a *App) clearWidgetArea(startRow, height int) {
	var seq strings.Builder

	for i := 0; i < height; i++ {
		row := startRow + i
		// Move to row (1-indexed) and clear the line
		seq.WriteString(fmt.Sprintf("\033[%d;1H\033[2K", row+1))
	}

	a.terminal.WriteDirect([]byte(seq.String()))
}

// deleteBlankHistoryRows removes n blank rows from the top of the history area
// using ANSI Delete Line (CSI n M). Unlike scrolling, DL does NOT push deleted
// lines into the terminal's scrollback buffer — they simply vanish. Content below
// shifts up to fill the gap, and blank lines appear at the bottom of the scroll
// region (just above the widget). This preserves visible content while making room
// for widget growth.
func (a *App) deleteBlankHistoryRows(n int, oldStartRow int) {
	if oldStartRow < 1 || n < 1 {
		return
	}

	var seq strings.Builder

	// Set scroll region to the history area so DL doesn't affect widget rows
	seq.WriteString(fmt.Sprintf("\033[1;%dr", oldStartRow))

	// Move to top of history area and delete n lines
	seq.WriteString("\033[1;1H")
	seq.WriteString(fmt.Sprintf("\033[%dM", n))

	// Reset scroll region to full screen
	seq.WriteString("\033[r")

	a.terminal.WriteDirect([]byte(seq.String()))
}

// scrollContentDown uses Reverse Index to scroll visible content down by n lines.
// This inserts blank lines at the TOP of the screen and pushes content down.
// Content at the bottom of the screen falls off (is lost).
// Used when shrinking the widget to put blanks at top (where they'll scroll
// into scrollback first) rather than at the bottom (where they'd be mixed with messages).
func (a *App) scrollContentDown(n int) {
	var seq strings.Builder

	// Move cursor to the top row (row 1 in ANSI 1-indexed)
	seq.WriteString("\033[1;1H")

	// Use Reverse Index (ESC M) n times to scroll content down
	// Each ESC M at the top of the screen inserts a blank line at row 1
	// and pushes everything down by 1 (bottom row falls off)
	for i := 0; i < n; i++ {
		seq.WriteString("\033M")
	}

	a.terminal.WriteDirect([]byte(seq.String()))
}

// InlineHeight returns the current inline height (0 if not in inline mode).
func (a *App) InlineHeight() int {
	return a.inlineHeight
}

// printAboveRaw handles the actual printing and scrolling for inline mode.
// Prints content that scrolls into terminal scrollback buffer, allowing
// the user to scroll back through history with their terminal's scroll feature.
// Must be called from the main event loop (via QueueUpdate).
//
// The scroll-then-print order is important: when the widget shrinks, we use
// Reverse Index to put blanks at the TOP of the visible area. The last history
// line is now at the bottom of the scroll region. If we printed first, we'd
// overwrite that line. By scrolling first, we push the top line (a blank from
// shrinking, or an old message) to scrollback, then print our new message
// over the blank line that appears at the bottom.
func (a *App) printAboveRaw(content string) {
	if a.inlineStartRow < 1 {
		return // No room above widget
	}

	text := strings.TrimSuffix(content, "\n")

	// In raw terminal mode, \n (LF) only moves cursor down without returning
	// to column 1. We need \r\n (CR+LF) to properly start each line at column 1.
	text = strings.ReplaceAll(text, "\n", "\r\n")

	// Use a scroll region to protect the widget at the bottom.
	// ANSI escape sequences use 1-indexed rows.
	// inlineStartRow is 0-indexed, so the widget occupies rows
	// (inlineStartRow+1) through termHeight in ANSI terms.

	var seq strings.Builder

	// Set scroll region to exclude widget area (rows 1 to inlineStartRow)
	seq.WriteString(fmt.Sprintf("\033[1;%dr", a.inlineStartRow))

	// Move to bottom of scroll region (last row before widget)
	seq.WriteString(fmt.Sprintf("\033[%d;1H", a.inlineStartRow))

	// Scroll first, then print (scroll-then-print order)
	// The newline scrolls content up within the region, pushing top line to scrollback
	// This creates a blank line at the bottom of the scroll region
	seq.WriteString("\n")

	// Now print the text over the blank line that just appeared
	// Move back to the bottom row (cursor is still there after newline in scroll region)
	seq.WriteString(text)

	// Reset scroll region to full screen
	seq.WriteString("\033[r")

	a.terminal.WriteDirect([]byte(seq.String()))

	// Track content rows. Each \n in the original content is one displayed line.
	lines := strings.Count(content, "\n")
	if lines == 0 {
		lines = 1
	}
	a.historyRows += lines
	if a.historyRows > a.inlineStartRow {
		a.historyRows = a.inlineStartRow
	}

	// Mark dirty to ensure consistent state
	MarkDirty()
}
