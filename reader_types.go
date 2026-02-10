package tui

import "time"

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

// InterruptibleReader extends EventReader with the ability to be interrupted.
// This is used for blocking mode where PollEvent(-1) would otherwise block forever.
type InterruptibleReader interface {
	EventReader

	// EnableInterrupt sets up the interrupt mechanism (e.g., a self-pipe).
	// Must be called before using blocking mode.
	EnableInterrupt() error

	// Interrupt wakes up a blocking PollEvent call.
	// Safe to call even if not currently blocking.
	Interrupt() error
}

// resizeDebounceWindow is the duration to wait for additional resize events before emitting.
// This coalesces rapid resize signals during window dragging into a single event.
const resizeDebounceWindow = 16 * time.Millisecond
