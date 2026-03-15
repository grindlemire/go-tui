package tui

import "github.com/grindlemire/go-tui/internal/debug"

// overlayEntry tracks a registered overlay for modal rendering.
type overlayEntry struct {
	element        *Element
	backdrop       string // "dim", "blank", "none"
	trapFocus      bool
	needsFocusInit bool // true on first frame after open; triggers initial focus move
}

// registerOverlay registers an element to be rendered in the overlay pass.
// Called by Modal.Render() when the modal is open.
//
// Overlays are not supported in inline mode because the buffer covers only
// a partial terminal region: backdrop effects, centering, and mouse hit
// testing all assume a full-screen buffer. Use EnterAlternateScreen()
// before opening a modal from inline mode.
func (a *App) registerOverlay(el *Element, backdrop string, trapFocus bool, needsFocusInit bool) {
	if !a.inAlternateScreen && a.inlineHeight > 0 {
		debug.Log("registerOverlay: ignored in inline mode (inlineHeight=%d); use EnterAlternateScreen() before opening a modal", a.inlineHeight)
		return
	}
	a.overlays = append(a.overlays, &overlayEntry{
		element:        el,
		backdrop:       backdrop,
		trapFocus:      trapFocus,
		needsFocusInit: needsFocusInit,
	})
}

// clearOverlays removes all registered overlays.
// Called at the start of each render frame.
func (a *App) clearOverlays() {
	a.overlays = a.overlays[:0]
}
