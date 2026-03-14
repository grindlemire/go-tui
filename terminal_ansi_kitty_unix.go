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
	// CSI ? <digits> u, so we stop as soon as we see the 'u' terminator
	// after the CSI ? prefix.
	var resp [32]byte
	n := 0
	seenCSIQuestion := false
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
		// Track whether we've seen the CSI ? prefix
		if n >= 3 && resp[n-3] == 0x1b && resp[n-2] == '[' && resp[n-1] == '?' {
			seenCSIQuestion = true
		}
		// Stop at 'u' terminator only after seeing the response prefix
		if seenCSIQuestion && b[0] == 'u' {
			break
		}
	}

	if n > 0 && parseKittyQueryResponse(resp[:n]) {
		t.kittyKeyboard = true
		t.caps.KittyKeyboard = true
		return true
	}

	// Pop to undo the push. Skip if we got no response at all (timeout
	// with zero bytes), since the pop is a no-op per the Kitty spec but
	// avoids an unnecessary write.
	if n > 0 {
		t.popKittyKeyboard()
	}
	return false
}
