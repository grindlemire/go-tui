package tui

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

// EventReader reads events from the terminal.
// It is designed for polling-based event loops.
type EventReader interface {
	// PollEvent reads the next event with a timeout.
	// Returns (event, true) if an event was read, or (nil, false) on timeout.
	// A timeout of 0 performs a non-blocking check.
	// A negative timeout blocks indefinitely.
	PollEvent(timeout time.Duration) (Event, bool)

	// Close releases resources. Must be called when done.
	Close() error
}

// stdinReader implements EventReader for a real terminal.
type stdinReader struct {
	fd         int           // stdin file descriptor
	buf        []byte        // Read buffer for escape sequences
	partialBuf []byte        // Buffer for incomplete UTF-8 sequences
	pending    []Event       // Parsed events waiting to be returned
	sigCh      chan os.Signal // For SIGWINCH (resize) handling
}

// NewEventReader creates an EventReader for the given terminal input.
// The terminal should already be in raw mode.
func NewEventReader(in *os.File) (EventReader, error) {
	r := &stdinReader{
		fd:    int(in.Fd()),
		buf:   make([]byte, 256),
		sigCh: make(chan os.Signal, 1),
	}

	// Set up SIGWINCH signal for resize events
	signal.Notify(r.sigCh, syscall.SIGWINCH)

	return r, nil
}

// PollEvent reads the next event with a timeout.
// Returns (event, true) if an event was read, or (nil, false) on timeout.
func (r *stdinReader) PollEvent(timeout time.Duration) (Event, bool) {
	// Return pending events first
	if len(r.pending) > 0 {
		ev := r.pending[0]
		r.pending = r.pending[1:]
		return ev, true
	}

	// Check for resize signal (non-blocking)
	select {
	case <-r.sigCh:
		w, h := getTerminalSizeForReader(r.fd)
		return ResizeEvent{Width: w, Height: h}, true
	default:
	}

	// Use select() with timeout for non-blocking stdin check
	ready, err := selectWithTimeout(r.fd, timeout)
	if err != nil || !ready {
		return nil, false
	}

	// Read available bytes
	n, err := syscall.Read(r.fd, r.buf)
	if err != nil || n == 0 {
		return nil, false
	}

	// Combine with any partial UTF-8 buffer from previous read
	data := r.buf[:n]
	if len(r.partialBuf) > 0 {
		data = append(r.partialBuf, data...)
		r.partialBuf = nil
	}

	// Parse into events
	events, remaining := parseInputWithRemainder(data)
	if len(remaining) > 0 {
		r.partialBuf = make([]byte, len(remaining))
		copy(r.partialBuf, remaining)
	}

	r.pending = events
	if len(r.pending) > 0 {
		ev := r.pending[0]
		r.pending = r.pending[1:]
		return ev, true
	}

	return nil, false
}

// Close releases resources.
func (r *stdinReader) Close() error {
	signal.Stop(r.sigCh)
	close(r.sigCh)
	return nil
}

// parseInputWithRemainder parses input and returns any incomplete trailing bytes.
// This handles partial UTF-8 sequences at the end of the buffer.
func parseInputWithRemainder(data []byte) ([]Event, []byte) {
	// Check for trailing incomplete UTF-8 sequence
	// A UTF-8 leading byte (0xC0-0xFF) without enough continuation bytes
	remaining := findIncompleteUTF8Suffix(data)
	if len(remaining) > 0 {
		data = data[:len(data)-len(remaining)]
	}

	events := parseInput(data)
	return events, remaining
}

// findIncompleteUTF8Suffix finds any incomplete UTF-8 sequence at the end of data.
// Returns the incomplete bytes (if any).
func findIncompleteUTF8Suffix(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	// Check last 1-3 bytes for incomplete UTF-8 sequences
	for i := 1; i <= 3 && i <= len(data); i++ {
		b := data[len(data)-i]

		// If this is a UTF-8 leading byte, check if sequence is complete
		if b >= 0xC0 {
			// Determine expected sequence length
			var expectedLen int
			switch {
			case b < 0xE0:
				expectedLen = 2
			case b < 0xF0:
				expectedLen = 3
			default:
				expectedLen = 4
			}

			// If we don't have enough bytes for the full sequence, it's incomplete
			if i < expectedLen {
				return data[len(data)-i:]
			}
			// Sequence is complete
			return nil
		}

		// If this is a continuation byte (0x80-0xBF), keep looking for the lead byte
		if b >= 0x80 && b < 0xC0 {
			continue
		}

		// ASCII byte - no incomplete sequence
		return nil
	}

	return nil
}
