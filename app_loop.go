package tui

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"
)

// ErrAlreadyOpen is returned by Open when the app has already been opened.
var ErrAlreadyOpen = errors.New("tui: app is already open")

// Open initializes the event loop: registers signal handlers, starts the
// input reader goroutine, and performs the initial render. Call this instead
// of Run() when driving your own event loop. Returns an error if already open.
//
// After Open(), use Events(), Dispatch(), and Render() to process events.
// Call Close() when done to restore terminal state.
func (a *App) Open() (retErr error) {
	if !a.opened.CompareAndSwap(false, true) {
		return ErrAlreadyOpen
	}

	// If Open fails after starting goroutines, clean them up.
	defer func() {
		if retErr != nil {
			a.Stop()
			if a.signalCleanup != nil {
				a.signalCleanup()
			}
		}
	}()

	// Handle Ctrl+C gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		select {
		case <-sigCh:
			a.Stop()
		case <-a.stopCh:
		}
		signal.Stop(sigCh)
	}()

	// Handle SIGWINCH (terminal resize)
	cleanupResize := a.registerResizeSignal()

	// Handle Ctrl+Z / SIGTSTP for job control
	cleanupSuspend := a.registerSuspendSignals()

	// Store cleanup functions for Close()
	a.signalCleanup = func() {
		cleanupResize()
		cleanupSuspend()
	}

	// Start input reader in background
	go a.readInputEvents()

	// Initial render
	a.MarkDirty()
	a.renderFrame()
	a.rebuildDispatchTable()

	return nil
}

// Run starts the main event loop. Blocks until Stop() is called or SIGINT received.
// This is equivalent to calling Open(), running a frame-timed loop with
// Dispatch/Render, and calling Close(). For custom event loops, use
// Open/Events/Dispatch/Render/Close directly.
func (a *App) Run() error {
	if err := a.Open(); err != nil && err != ErrAlreadyOpen {
		return err
	}
	defer a.Close()

	for {
		frameStart := time.Now()
		eventDeadline := frameStart.Add(a.frameDuration / 2)

		// Drain events for up to half the frame budget
	drain:
		for time.Now().Before(eventDeadline) {
			select {
			case ev := <-a.merged:
				a.Dispatch(ev)
			case <-a.stopCh:
				return nil
			default:
				break drain
			}
		}

		a.Render()

		// Sleep for remaining frame time. Events arriving during sleep are
		// processed on the next iteration. For lower latency, use Events()
		// in a custom select loop.
		elapsed := time.Since(frameStart)
		if remaining := a.frameDuration - elapsed; remaining > 0 {
			select {
			case <-time.After(remaining):
			case <-a.stopCh:
				return nil
			}
		}
	}
}

// Stop signals the Run loop to exit gracefully and stops all watchers.
// Watchers receive the stop signal via stopCh and exit their goroutines.
// Stop is idempotent - multiple calls are safe.
func (a *App) Stop() {
	a.stopOnce.Do(func() {
		a.stopped = true

		// Interrupt blocking reader before closing stopCh to wake it up
		if interruptible, ok := a.reader.(InterruptibleReader); ok {
			interruptible.Interrupt()
		}

		// Signal all watcher goroutines to stop
		close(a.stopCh)
	})
}

// Events returns a read-only channel carrying all events: key, mouse, resize,
// and queued updates. Input events are prioritized over background updates.
// Use this with select to multiplex go-tui events with your own event sources.
// The channel remains open until the App is garbage collected; use StopCh()
// to detect shutdown.
func (a *App) Events() <-chan Event {
	return a.merged
}

// DispatchEvents reads and dispatches all pending events from the Events channel.
// Returns false if the app has been stopped, true otherwise.
func (a *App) DispatchEvents() bool {
	for {
		select {
		case <-a.stopCh:
			return false
		case ev := <-a.merged:
			a.Dispatch(ev)
		default:
			return true
		}
	}
}

// Step is a convenience that calls DispatchEvents followed by Render.
// Returns false if the app has been stopped.
func (a *App) Step() bool {
	if !a.DispatchEvents() {
		return false
	}
	a.Render()
	return true
}

// QueueUpdate enqueues a function to run on the main loop.
// Safe to call from any goroutine. Use this for background thread safety.
// If the updates channel is full, the update is dropped to avoid blocking
// the caller.
func (a *App) QueueUpdate(fn func()) {
	if fn == nil {
		return
	}
	select {
	case a.updates <- UpdateEvent{fn: fn}:
	case <-a.stopCh:
	default:
		// Channel full; drop the update to avoid blocking the caller.
	}
}

// rebuildDispatchTable walks the rendered element tree and builds a new
// dispatch table from all mounted components' KeyMap() methods.
// If validation fails, the previous table is kept.
func (a *App) rebuildDispatchTable() {
	if a.root == nil {
		return
	}

	table, err := buildDispatchTable(a.rootComponent, a.root)
	if err != nil {
		// Validation error (e.g., conflicting Stop handlers).
		// Log and keep the previous valid table rather than crashing.
		fmt.Fprintf(os.Stderr, "tui: dispatch table error: %v\n", err)
		return
	}
	a.dispatchTable = table
}
