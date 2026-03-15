package tui

// overlayEntry tracks a registered overlay for modal rendering.
type overlayEntry struct {
	element   *Element
	backdrop  string // "dim", "blank", "none"
	trapFocus bool
}

// registerOverlay registers an element to be rendered in the overlay pass.
// Called by Modal.Render() when the modal is open.
func (a *App) registerOverlay(el *Element, backdrop string, trapFocus bool) {
	a.overlays = append(a.overlays, &overlayEntry{
		element:   el,
		backdrop:  backdrop,
		trapFocus: trapFocus,
	})
}

// clearOverlays removes all registered overlays.
// Called at the start of each render frame.
func (a *App) clearOverlays() {
	a.overlays = a.overlays[:0]
}
