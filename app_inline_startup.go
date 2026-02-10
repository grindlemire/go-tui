package tui

import (
	"fmt"
	"strings"
)

func (a *App) setupInitialScreen(width, termHeight int) {
	if a.inlineHeight > 0 {
		a.setupInlineScreen(width, termHeight)
		return
	}

	// Full screen mode: use alternate screen.
	a.terminal.EnterAltScreen()
	a.buffer = NewBuffer(width, termHeight)
}

func (a *App) setupInlineScreen(width, termHeight int) {
	// Clamp inline height to terminal height.
	if a.inlineHeight > termHeight {
		a.inlineHeight = termHeight
	}
	if a.inlineHeight < 0 {
		a.inlineHeight = 0
	}

	// Reserve space for the inline widget at the bottom.
	a.reserveInlineRegion()

	// Calculate where our inline region starts.
	a.inlineStartRow = termHeight - a.inlineHeight
	if a.inlineStartRow < 0 {
		a.inlineStartRow = 0
	}
	a.inlineSession = newInlineSession(a.terminal)

	// Create buffer sized for inline region only.
	a.buffer = NewBuffer(width, a.inlineHeight)

	a.applyInlineStartupPolicy(termHeight)

	// First inline frame must repaint owned rows fully; startup state in the
	// terminal is not guaranteed to match the buffer's initial front state.
	a.needsFullRedraw = true
}

func (a *App) reserveInlineRegion() {
	if a.inlineHeight < 1 {
		return
	}

	// Reserve rows via linefeeds and return to the start of the reserved block.
	_, _ = a.terminal.WriteDirect([]byte(strings.Repeat("\r\n", a.inlineHeight)))
	_, _ = a.terminal.WriteDirect([]byte(fmt.Sprintf("\033[%dA", a.inlineHeight)))
}

func (a *App) applyInlineStartupPolicy(termHeight int) {
	historyCapacity := a.inlineStartRow

	switch a.inlineStartupMode {
	case InlineStartupFreshViewport:
		a.clearVisibleViewport()
		a.inlineLayout = newInlineLayoutState(historyCapacity)
	case InlineStartupSoftReset:
		a.softResetVisibleViewport(termHeight)
		a.inlineLayout = newInlineLayoutState(historyCapacity)
	case InlineStartupPreserveVisible:
		fallthrough
	default:
		// Unknown content may already be present in the history region.
		// Mark layout invalid so first append conservatively treats history as full.
		a.inlineLayout = newInlineLayoutState(historyCapacity)
		a.inlineLayout.invalidate(historyCapacity)
	}
}

func (a *App) clearVisibleViewport() {
	a.terminal.SetCursor(0, 0)
	a.terminal.ClearToEnd()
}

func (a *App) softResetVisibleViewport(termHeight int) {
	if termHeight < 1 {
		return
	}

	// Scroll existing visible rows into scrollback without issuing a global
	// scrollback erase command.
	a.terminal.SetCursor(0, termHeight-1)
	_, _ = a.terminal.WriteDirect([]byte(strings.Repeat("\r\n", termHeight)))
}
