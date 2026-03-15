//go:build windows

package tui

// NegotiateKittyKeyboard is a no-op on Windows.
// The Kitty keyboard protocol is not supported on Windows terminals.
func (t *ANSITerminal) NegotiateKittyKeyboard() bool {
	return false
}
