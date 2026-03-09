//go:build windows

package tui

// registerResizeSignal is a no-op on Windows (no SIGWINCH support).
// Returns a no-op cleanup.
func (a *App) registerResizeSignal() func() {
	return func() {}
}
