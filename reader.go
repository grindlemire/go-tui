//go:build !windows

package tui

import (
	"os"
	"sync/atomic"
	"syscall"
	"time"
)

// stdinReader implements EventReader for a real terminal.
type stdinReader struct {
	fd         int    // stdin file descriptor
	buf        []byte // Read buffer for escape sequences
	partialBuf []byte // Buffer for incomplete UTF-8 sequences
	pending    []Event // Parsed events waiting to be returned

	// Interrupt mechanism for blocking mode
	interruptPipe [2]int // [0]=read, [1]=write
	hasInterrupt  bool   // Whether interrupt is enabled

	paused atomic.Bool // When true, PollEvent returns immediately
}

// Ensure stdinReader implements InterruptibleReader and PausableReader.
var _ InterruptibleReader = (*stdinReader)(nil)
var _ PausableReader = (*stdinReader)(nil)

// NewEventReader creates an EventReader for the given terminal input.
// The terminal should already be in raw mode.
func NewEventReader(in *os.File) (EventReader, error) {
	r := &stdinReader{
		fd:  int(in.Fd()),
		buf: make([]byte, 256),
	}
	return r, nil
}

// Pause causes PollEvent to return immediately without reading stdin.
// Interrupts any in-progress blocking read.
func (r *stdinReader) Pause() {
	r.paused.Store(true)
	if r.hasInterrupt {
		r.Interrupt()
	}
}

// Resume allows PollEvent to read stdin again.
func (r *stdinReader) Resume() {
	r.paused.Store(false)
}

// PollEvent reads the next event with a timeout.
// Returns (event, true) if an event was read, or (nil, false) on timeout.
func (r *stdinReader) PollEvent(timeout time.Duration) (Event, bool) {
	if r.paused.Load() {
		return nil, false
	}

	// Return pending events first
	if len(r.pending) > 0 {
		ev := r.pending[0]
		r.pending = r.pending[1:]
		return ev, true
	}

	// Use select() with timeout for non-blocking stdin check
	// If interrupt is enabled, use the interruptible version
	var ready bool
	var err error
	if r.hasInterrupt {
		var interrupted bool
		ready, interrupted, err = selectWithTimeoutAndInterrupt(r.fd, r.interruptPipe[0], timeout)
		if interrupted {
			return nil, false // Interrupted, return immediately
		}
	} else {
		ready, err = selectWithTimeout(r.fd, timeout)
	}

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
	if r.hasInterrupt {
		syscall.Close(r.interruptPipe[0])
		syscall.Close(r.interruptPipe[1])
	}
	return nil
}

// EnableInterrupt sets up the interrupt mechanism using a self-pipe.
// This allows Interrupt() to wake up a blocking PollEvent call.
func (r *stdinReader) EnableInterrupt() error {
	if r.hasInterrupt {
		return nil // Already enabled
	}
	var fds [2]int
	if err := syscall.Pipe(fds[:]); err != nil {
		return err
	}
	r.interruptPipe = fds
	r.hasInterrupt = true
	return nil
}

// Interrupt wakes up a blocking PollEvent call by writing to the interrupt pipe.
func (r *stdinReader) Interrupt() error {
	if !r.hasInterrupt {
		return nil
	}
	_, err := syscall.Write(r.interruptPipe[1], []byte{0})
	return err
}

// parseInputWithRemainder parses input and returns any incomplete trailing bytes.
// This handles partial UTF-8 sequences and incomplete escape sequences at the end of the buffer.
func parseInputWithRemainder(data []byte) ([]Event, []byte) {
	// Check for trailing incomplete escape sequence first
	escRemaining := findIncompleteEscapeSequence(data)
	if len(escRemaining) > 0 {
		// Only buffer the escape sequence if there's other data before it.
		// If the entire buffer is just the incomplete sequence, don't buffer -
		// this handles the case where user actually pressed ESC key.
		if len(escRemaining) < len(data) {
			data = data[:len(data)-len(escRemaining)]
		} else {
			// Entire buffer is the incomplete sequence - don't buffer it
			escRemaining = nil
		}
	}

	// Check for trailing incomplete UTF-8 sequence
	// A UTF-8 leading byte (0xC0-0xFF) without enough continuation bytes
	remaining := findIncompleteUTF8Suffix(data)
	if len(remaining) > 0 {
		data = data[:len(data)-len(remaining)]
	}

	events := parseInput(data)

	// Combine remainders (escape sequence remainder takes precedence if both exist)
	if len(escRemaining) > 0 {
		return events, escRemaining
	}
	return events, remaining
}

// findIncompleteEscapeSequence finds any incomplete escape sequence at the end of data.
// Returns the incomplete bytes (if any).
// Handles:
// - Lone ESC at end (could be start of ESC[... or ESC O...)
// - ESC[ without terminating byte (CSI sequence in progress)
// - ESC[< without M/m terminator (SGR mouse sequence in progress)
func findIncompleteEscapeSequence(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	// Scan backwards to find a potential incomplete escape sequence
	// Look for ESC (0x1b) in the last ~64 bytes (max reasonable escape sequence length)
	searchStart := len(data) - 64
	if searchStart < 0 {
		searchStart = 0
	}

	for i := len(data) - 1; i >= searchStart; i-- {
		if data[i] != 0x1b {
			continue
		}

		// Found ESC at position i
		suffix := data[i:]

		// Lone ESC at end - definitely incomplete
		if len(suffix) == 1 {
			return suffix
		}

		// Check what follows ESC
		switch suffix[1] {
		case '[':
			// CSI sequence - check if it's complete
			// Complete CSI ends with byte in range 0x40-0x7e (@ through ~)
			if len(suffix) == 2 {
				// Just "ESC[" - incomplete
				return suffix
			}

			// Check for SGR mouse sequence: ESC[<...M or ESC[<...m
			if suffix[2] == '<' {
				// SGR mouse - look for terminating M or m
				for j := 3; j < len(suffix); j++ {
					if suffix[j] == 'M' || suffix[j] == 'm' {
						// Found terminator - sequence is complete
						// Continue scanning for another ESC
						break
					}
					// Check for invalid characters in mouse sequence
					// Valid: 0-9, ;, and final M/m
					if suffix[j] != ';' && (suffix[j] < '0' || suffix[j] > '9') {
						// Invalid character - not an SGR mouse sequence
						break
					}
					if j == len(suffix)-1 {
						// Reached end without finding M/m - incomplete
						return suffix
					}
				}
			} else {
				// Regular CSI - look for terminating byte (0x40-0x7e)
				for j := 2; j < len(suffix); j++ {
					if suffix[j] >= 0x40 && suffix[j] <= 0x7e {
						// Found terminator - sequence is complete
						break
					}
					if j == len(suffix)-1 {
						// Reached end without finding terminator - incomplete
						return suffix
					}
				}
			}

		case 'O':
			// SS3 sequence - needs one more byte
			if len(suffix) == 2 {
				return suffix
			}
			// SS3 sequences are always 3 bytes total, so complete

		default:
			// Could be Alt+key (ESC + printable) - these are complete at 2 bytes
			// Or unknown sequence - treat as complete to avoid infinite buffering
		}
	}

	return nil
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
