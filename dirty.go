package tui

// MarkDirty marks this app as needing a render.
func (a *App) MarkDirty() {
	if a == nil {
		panic("tui: nil app in MarkDirty")
	}
	a.dirty.Store(true)
}

func (a *App) checkAndClearDirty() bool {
	if a == nil {
		panic("tui: nil app in checkAndClearDirty")
	}
	return a.dirty.Swap(false)
}

func (a *App) resetDirty() {
	if a == nil {
		panic("tui: nil app in resetDirty")
	}
	a.dirty.Store(false)
}
