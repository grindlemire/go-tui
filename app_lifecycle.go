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
		// Strategy depends on history packing:
		// - bottom-packed (default): consume top blanks first, then scroll content
		// - top-packed (after shrink): consume bottom blanks by shrinking area; if
		//   content overflows, scroll oldest rows into scrollback
		a.clearWidgetArea(oldStartRow, oldHeight)
		linesToScroll := rows - oldHeight

		if a.historyTopAligned {
			blankRowsBottom := oldStartRow - a.historyRows
			if linesToScroll > blankRowsBottom {
				overflow := linesToScroll - blankRowsBottom
				a.scrollHistoryUpRegion(overflow, 0, oldStartRow)
				a.historyRows -= overflow
				if a.historyRows < 0 {
					a.historyRows = 0
				}
			}

			// Once history exactly fills the area, alignment is effectively neutral.
			// Switch back to default bottom-packed mode for future operations.
			if a.historyRows >= newStartRow {
				a.historyRows = newStartRow
				a.historyTopAligned = false
			}
		} else {
			blankRows := oldStartRow - a.historyRows

			if linesToScroll < blankRows {
				// Enough blank rows at the top: consume only blanks without touching row 0.
				// This keeps scrollback clean while preserving visible content.
				topRow := blankRows - linesToScroll
				a.scrollHistoryUpRegion(linesToScroll, topRow, oldStartRow)
			} else {
				// Need to remove all blank rows and some content rows.
				// First, remove as many blanks as possible without touching row 0.
				// Leave at most one blank at row 0, then scroll the remainder from row 0
				// so only real content (plus at most one unavoidable blank) enters scrollback.
				if blankRows > 1 {
					a.scrollHistoryUpRegion(blankRows-1, 1, oldStartRow)
				}

				remaining := linesToScroll
				if blankRows > 1 {
					remaining -= (blankRows - 1)
				}
				a.scrollHistoryUpRegion(remaining, 0, oldStartRow)

				// Content rows removed are growth minus true blank capacity.
				contentRowsRemoved := linesToScroll - blankRows
				if contentRowsRemoved > 0 {
					a.historyRows -= contentRowsRemoved
					if a.historyRows < 0 {
						a.historyRows = 0
					}
				}
			}
		}
	} else {
		// Shrinking: We need to handle the "released" rows (the rows that were part of
		// the old widget but won't be part of the new smaller widget).
		//
		// The challenge: These rows are now in the history area. If we leave them blank,
		// they'll scroll into the scrollback mixed with actual messages.
		//
		// Keep current history rows where they are and let released rows become
		// blanks at the bottom of the expanded history area. This preserves line
		// chronology (important for multiline submits) instead of inserting a blank
		// block between already-scrolled lines and still-visible lines.
		a.clearWidgetArea(oldStartRow, oldHeight)
		a.historyTopAligned = true
	}

	a.inlineHeight = rows
	a.inlineStartRow = newStartRow
	a.buffer.Resize(width, rows)
	a.needsFullRedraw = true // Terminal position shifted, need full redraw
	debug.Log("SetInlineHeight: buffer resized, new inlineStartRow=%d, needsFullRedraw=true", a.inlineStartRow)
}

// scrollHistoryUpRegion scrolls a history subregion up by n lines.
// topRow is 0-indexed and inclusive; oldStartRow is the widget start row (0-indexed),
// so the region bottom is oldStartRow-1.
//
// If topRow == 0, scrolled-off lines are pushed into terminal scrollback.
// If topRow > 0, scrolled-off lines are discarded instead.
func (a *App) scrollHistoryUpRegion(n int, topRow int, oldStartRow int) {
	if oldStartRow < 1 || n < 1 {
		return
	}
	if topRow < 0 {
		topRow = 0
	}
	if topRow >= oldStartRow {
		return
	}

	var seq strings.Builder

	// Set scroll region to the requested history subregion (ANSI is 1-indexed).
	seq.WriteString(fmt.Sprintf("\033[%d;%dr", topRow+1, oldStartRow))

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

// InlineHeight returns the current inline height (0 if not in inline mode).
func (a *App) InlineHeight() int {
	return a.inlineHeight
}

// printAboveRaw handles the actual printing and scrolling for inline mode.
// Prints content that scrolls into terminal scrollback buffer, allowing
// the user to scroll back through history with their terminal's scroll feature.
// Must be called from the main event loop (via QueueUpdate).
// Supports both history packing modes:
// - top-packed (blanks at bottom): fill blank rows first, then scroll when full
// - bottom-packed (blanks at top): scroll first, then print at bottom
func (a *App) printAboveRaw(content string) {
	if a.inlineStartRow < 1 {
		return // No room above widget
	}

	text := strings.TrimSuffix(content, "\n")
	lines := strings.Split(text, "\n")

	// Top-packed history mode (blanks at bottom): append directly into available
	// blank rows before scrolling. This avoids introducing blank blocks between
	// multiline submit lines when the widget just shrank.
	if a.historyTopAligned {
		var seq strings.Builder
		for _, line := range lines {
			if a.historyRows < a.inlineStartRow {
				row := a.historyRows + 1 // ANSI 1-indexed
				seq.WriteString(fmt.Sprintf("\033[%d;1H\033[2K", row))
				seq.WriteString(line)
				a.historyRows++
				continue
			}

			seq.WriteString(fmt.Sprintf("\033[1;%dr", a.inlineStartRow))
			seq.WriteString(fmt.Sprintf("\033[%d;1H", a.inlineStartRow))
			seq.WriteString("\n")
			seq.WriteString(fmt.Sprintf("\033[%d;1H\033[2K", a.inlineStartRow))
			seq.WriteString(line)
			seq.WriteString("\033[r")
		}

		a.terminal.WriteDirect([]byte(seq.String()))

		if a.historyRows > a.inlineStartRow {
			a.historyRows = a.inlineStartRow
		}
		if a.historyRows == a.inlineStartRow {
			a.historyTopAligned = false
		}

		MarkDirty()
		return
	}

	// Bottom-packed history mode (blanks at top): scroll then print at bottom.
	// In raw terminal mode, \n (LF) only moves cursor down without returning
	// to column 1. We need \r\n (CR+LF) to properly start each line at column 1.
	text = strings.ReplaceAll(text, "\n", "\r\n")

	var seq strings.Builder
	seq.WriteString(fmt.Sprintf("\033[1;%dr", a.inlineStartRow))
	seq.WriteString(fmt.Sprintf("\033[%d;1H", a.inlineStartRow))
	seq.WriteString("\n")
	seq.WriteString(text)
	seq.WriteString("\033[r")

	a.terminal.WriteDirect([]byte(seq.String()))

	a.historyRows += len(lines)
	if a.historyRows > a.inlineStartRow {
		a.historyRows = a.inlineStartRow
	}

	// Mark dirty to ensure consistent state
	MarkDirty()
}
