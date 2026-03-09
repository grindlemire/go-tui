//go:build unix

package tui

import (
	"time"

	"golang.org/x/sys/unix"
)

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

// selectWithTimeoutAndInterrupt performs a select() call on fd and optionally an interrupt fd.
// Returns (ready, interrupted, err) where:
// - ready=true if the main fd is ready for reading
// - interrupted=true if the interrupt fd was triggered
// - err is non-nil on error
func selectWithTimeoutAndInterrupt(fd, interruptFd int, timeout time.Duration) (ready, interrupted bool, err error) {
	var readFds unix.FdSet
	readFds.Zero()
	readFds.Set(fd)

	maxFd := fd
	if interruptFd >= 0 {
		readFds.Set(interruptFd)
		if interruptFd > maxFd {
			maxFd = interruptFd
		}
	}

	var tv *unix.Timeval
	if timeout >= 0 {
		tvVal := unix.NsecToTimeval(timeout.Nanoseconds())
		tv = &tvVal
	}
	// If timeout < 0, tv is nil which means block indefinitely

	n, err := unix.Select(maxFd+1, &readFds, nil, nil, tv)
	if err != nil {
		if err == unix.EINTR {
			return false, false, nil
		}
		return false, false, err
	}

	if n == 0 {
		return false, false, nil // Timeout
	}

	// Check if interrupt fd was triggered
	if interruptFd >= 0 && readFds.IsSet(interruptFd) {
		// Drain the interrupt byte so the pipe is ready for the next interrupt
		var buf [1]byte
		unix.Read(interruptFd, buf[:])
		return false, true, nil
	}

	return readFds.IsSet(fd), false, nil
}
