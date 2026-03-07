//go:build windows

package tui

// suspendTerminal is a no-op on Windows (no SIGTSTP support).
func (a *App) suspendTerminal() {}

// resumeTerminal is a no-op on Windows.
func (a *App) resumeTerminal() {}

// suspend is a no-op on Windows.
func (a *App) suspend() {}

// Suspend is a no-op on Windows.
func (a *App) Suspend() {}

// registerSuspendSignals is a no-op on Windows. Returns a no-op cleanup.
func (a *App) registerSuspendSignals() func() {
	return func() {}
}
