//go:build unix

package tui

import (
	"golang.org/x/sys/unix"
)

// rawModeState stores the original terminal state for restoration.
type rawModeState struct {
	termios unix.Termios
}

// enableRawMode puts the terminal into raw mode and returns the previous state.
func enableRawMode(fd int) (*rawModeState, error) {
	// Get current terminal attributes
	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return nil, err
	}

	// Save original state
	state := &rawModeState{termios: *termios}

	// Modify for raw mode
	// Turn off:
	// - ECHO: don't echo input characters
	// - ICANON: disable canonical mode (read byte-by-byte instead of line-by-line)
	// - ISIG: disable signals (Ctrl+C, Ctrl+Z, etc.)
	// - IEXTEN: disable extended input processing
	termios.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN

	// Turn off:
	// - IXON: disable software flow control (Ctrl+S, Ctrl+Q)
	// - ICRNL: don't translate CR to NL
	// - BRKINT: don't send SIGINT on break
	// - INPCK: disable parity checking
	// - ISTRIP: don't strip 8th bit
	termios.Iflag &^= unix.IXON | unix.ICRNL | unix.BRKINT | unix.INPCK | unix.ISTRIP

	// Turn off:
	// - OPOST: disable output processing
	termios.Oflag &^= unix.OPOST

	// Set:
	// - CS8: 8-bit characters
	termios.Cflag |= unix.CS8

	// Set minimum bytes for read and timeout
	// VMIN = 1: read returns when at least 1 byte is available
	// VTIME = 0: no timeout
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	// Apply new settings
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, termios); err != nil {
		return nil, err
	}

	return state, nil
}

// disableRawMode restores the terminal to its previous state.
func disableRawMode(state *rawModeState) error {
	if state == nil {
		return nil
	}
	// Note: We use TIOCSETA for immediate change
	// TIOCSETAF would wait for output to drain and flush input
	return unix.IoctlSetTermios(unix.Stdin, unix.TIOCSETA, &state.termios)
}

// getTerminalSize returns the terminal dimensions.
func getTerminalSize(fd int) (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}
	return int(ws.Col), int(ws.Row), nil
}
