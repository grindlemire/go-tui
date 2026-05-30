package tui

// enableInputReporting turns on the input-reporting mode appropriate for the
// run. With mouse enabled we report mouse events (clicks, wheel). With mouse
// disabled in full-screen mode we instead enable alternate-scroll, so the mouse
// wheel still scrolls (the terminal translates wheel notches into cursor keys)
// while native text selection and OSC 8 link clicking keep working. Inline mode
// with mouse off enables neither: there is no alternate screen for the terminal
// to gate alternate-scroll on, and we leave the wheel to the host terminal so
// scrollback stays usable. Dynamic alternate-screen overlays are left alone too;
// alternate-scroll tracks the base full-screen mode only.
func (a *App) enableInputReporting() {
	if a.mouseEnabled {
		a.terminal.EnableMouse()
		return
	}
	if a.baseFullScreenNoMouse() {
		a.terminal.EnableAltScroll()
	}
}

// disableInputReporting undoes enableInputReporting, matching the same mode
// selection so we only disable what we turned on.
func (a *App) disableInputReporting() {
	if a.mouseEnabled {
		a.terminal.DisableMouse()
		return
	}
	if a.baseFullScreenNoMouse() {
		a.terminal.DisableAltScroll()
	}
}

// baseFullScreenNoMouse reports whether the app is in its base full-screen mode
// (not inline, not a dynamic alternate-screen overlay) with mouse reporting off,
// which is exactly when alternate-scroll should be active.
func (a *App) baseFullScreenNoMouse() bool {
	return !a.mouseEnabled && a.inlineHeight == 0 && !a.inAlternateScreen
}
