//go:build unix

package tui

import (
	"syscall"
	"time"
)

// NegotiateKittyKeyboard attempts to enable Kitty keyboard protocol (flag 1 =
// disambiguate). Uses push/pop stack semantics so nested TUI apps coexist.
//
// Sequence:
//  1. Push flag 1: CSI > 1 u
//  2. Query current mode: CSI ? u
//  3. Poll stdin (50ms timeout) for response: CSI ? flags u
//  4. If response includes flag 1, success. Otherwise pop: CSI < u
func (t *ANSITerminal) NegotiateKittyKeyboard(stdinFd int) bool {
	// Push disambiguate mode onto the keyboard stack and query
	t.esc.Reset()
	t.esc.KittyKeyboardPush(1)
	t.esc.KittyKeyboardQuery()
	t.out.Write(t.esc.Bytes())

	// Poll for the terminal's response with a short timeout.
	// We read directly from the fd since the EventReader isn't created yet.
	ready, err := selectWithTimeout(stdinFd, 50*time.Millisecond)
	if err != nil || !ready {
		// No response: terminal doesn't support Kitty protocol. Pop to undo.
		t.esc.Reset()
		t.esc.KittyKeyboardPop()
		t.out.Write(t.esc.Bytes())
		return false
	}

	// Read the response
	var buf [64]byte
	n, err := syscall.Read(stdinFd, buf[:])
	if err != nil || n == 0 {
		t.esc.Reset()
		t.esc.KittyKeyboardPop()
		t.out.Write(t.esc.Bytes())
		return false
	}

	// Parse response: expect CSI ? flags u
	response := buf[:n]
	if parseKittyQueryResponse(response) {
		t.kittyKeyboard = true
		t.caps.KittyKeyboard = true
		return true
	}

	// Unrecognized response: pop to undo
	t.esc.Reset()
	t.esc.KittyKeyboardPop()
	t.out.Write(t.esc.Bytes())
	return false
}
