//go:build windows

package tui

// registerResizeSignal is a no-op on Windows: resize events are delivered
// in-band by the console input handle as WINDOW_BUFFER_SIZE_EVENT records,
// which the Windows EventReader decodes directly into ResizeEvent. There is
// no SIGWINCH-style signal to install. Returns a no-op cleanup.
func (a *App) registerResizeSignal() func() {
	return func() {}
}
