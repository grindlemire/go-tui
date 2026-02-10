package tui

// MarkDirty signals that the UI needs re-rendering.
// Called automatically by State.Set(), Element.ScrollBy(), etc.
// Can also be called manually for custom mutations.
func MarkDirty() {
	if app := DefaultApp(); app != nil {
		app.MarkDirty()
		return
	}
	panic("tui.MarkDirty requires a default app; call SetDefaultApp or use app.MarkDirty")
}

// MarkDirty marks this app as needing a render.
func (a *App) MarkDirty() {
	if a == nil {
		panic("tui: nil app in MarkDirty")
	}
	a.dirty.Store(true)
}

// checkAndClearDirty returns true if dirty and clears the flag.
// Called by the main loop after processing events.
func checkAndClearDirty() bool {
	if app := DefaultApp(); app != nil {
		return app.checkAndClearDirty()
	}
	panic("tui.checkAndClearDirty requires a default app")
}

func (a *App) checkAndClearDirty() bool {
	if a == nil {
		panic("tui: nil app in checkAndClearDirty")
	}
	return a.dirty.Swap(false)
}

// resetDirty clears the dirty flag without returning its value.
// Used for testing to reset state between tests.
func resetDirty() {
	if app := DefaultApp(); app != nil {
		app.resetDirty()
		return
	}
	panic("tui.resetDirty requires a default app")
}

func (a *App) resetDirty() {
	if a == nil {
		panic("tui: nil app in resetDirty")
	}
	a.dirty.Store(false)
}
