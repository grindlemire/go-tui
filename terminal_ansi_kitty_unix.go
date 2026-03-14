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

	// Read the response byte-by-byte to avoid consuming extra stdin data
	// (e.g. keystrokes typed during startup). The expected response is
	// CSI ? <digits> u, so we stop as soon as we see the 'u' terminator.
	var resp [32]byte
	n := 0
	deadline := time.Now().Add(50 * time.Millisecond)

	for n < len(resp) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		ready, err := selectWithTimeout(stdinFd, remaining)
		if err != nil || !ready {
			break
		}
		var b [1]byte
		nr, err := syscall.Read(stdinFd, b[:])
		if err != nil || nr == 0 {
			break
		}
		resp[n] = b[0]
		n++
		// Minimum valid response is \x1b[?1u (5 bytes). Stop at 'u' terminator.
		if b[0] == 'u' && n >= 5 {
			break
		}
	}

	if n > 0 && parseKittyQueryResponse(resp[:n]) {
		t.kittyKeyboard = true
		t.caps.KittyKeyboard = true
		return true
	}

	t.popKittyKeyboard()
	return false
}
