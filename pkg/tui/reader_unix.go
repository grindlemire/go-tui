//go:build unix

package tui

import (
	"time"

	"golang.org/x/sys/unix"
)

// getTerminalSizeForReader returns the terminal dimensions for the EventReader.
// This is separate from getTerminalSize in terminal_unix.go to avoid circular deps.
func getTerminalSizeForReader(fd int) (width, height int) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		// Default to standard terminal size on error
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}

// selectWithTimeout performs a select() call on the given fd with timeout.
// Returns (true, nil) if the fd is ready for reading.
// Returns (false, nil) on timeout.
// Returns (false, err) on error.
func selectWithTimeout(fd int, timeout time.Duration) (ready bool, err error) {
	// Prepare the fd set
	var readFds unix.FdSet
	readFds.Zero()
	readFds.Set(fd)

	// Convert timeout to timeval
	var tv *unix.Timeval
	if timeout >= 0 {
		tvVal := unix.NsecToTimeval(timeout.Nanoseconds())
		tv = &tvVal
	}
	// If timeout < 0, tv is nil which means block indefinitely

	// Call select
	n, err := unix.Select(fd+1, &readFds, nil, nil, tv)
	if err != nil {
		// EINTR is expected when signals arrive
		if err == unix.EINTR {
			return false, nil
		}
		return false, err
	}

	return n > 0, nil
}
