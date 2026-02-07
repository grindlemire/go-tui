package tui

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

// Run starts the main event loop. Blocks until Stop() is called or SIGINT received.
// Rendering occurs only when the dirty flag is set (by mutations).
func (a *App) Run() error {
	// Set current app for package-level Stop(), saving previous for nested apps
	prevApp := currentApp
	currentApp = a
	defer func() { currentApp = prevApp }()

	// Handle Ctrl+C gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		select {
		case <-sigCh:
			a.Stop()
		case <-a.stopCh:
			// App already stopped, clean up signal handler
		}
		signal.Stop(sigCh)
	}()

	// Start input reader in background
	go a.readInputEvents()

	// Initial render
	a.Render()
	a.rebuildDispatchTable()

	// Frame-based loop with configurable frame timing
	for !a.stopped {
		frameStart := time.Now()

		// Process events for up to half the frame budget (non-blocking)
		eventDeadline := frameStart.Add(a.frameDuration / 2)
		for time.Now().Before(eventDeadline) {
			select {
			case handler := <-a.eventQueue:
				handler()
			case <-a.stopCh:
				return nil
			default:
				// No more events, move to render phase
				goto render
			}
		}

	render:
		// Always render if dirty
		if checkAndClearDirty() {
			a.Render()
			a.rebuildDispatchTable()
		}

		// Sleep for remaining frame time to maintain consistent framerate
		elapsed := time.Since(frameStart)
		if elapsed < a.frameDuration {
			select {
			case <-time.After(a.frameDuration - elapsed):
			case <-a.stopCh:
				return nil
			}
		}
	}

	return nil
}

// Stop signals the Run loop to exit gracefully and stops all watchers.
// Watchers receive the stop signal via stopCh and exit their goroutines.
// Stop is idempotent - multiple calls are safe.
func (a *App) Stop() {
	if a.stopped {
		return // Already stopped
	}
	a.stopped = true

	// Interrupt blocking reader before closing stopCh to wake it up
	if interruptible, ok := a.reader.(InterruptibleReader); ok {
		interruptible.Interrupt()
	}

	// Signal all watcher goroutines to stop
	close(a.stopCh)
}

// QueueUpdate enqueues a function to run on the main loop.
// Safe to call from any goroutine. Use this for background thread safety.
func (a *App) QueueUpdate(fn func()) {
	select {
	case a.eventQueue <- fn:
	case <-a.stopCh:
		// App is stopping, ignore update
	default:
		// Queue full - this shouldn't happen with reasonable buffer size
		// Could log a warning here
	}
}

// rebuildDispatchTable walks the rendered element tree and builds a new
// dispatch table from all mounted components' KeyMap() methods.
// If the root is not an *Element or validation fails, the previous table is kept.
func (a *App) rebuildDispatchTable() {
	root, ok := a.root.(*Element)
	if !ok {
		return
	}

	table, err := buildDispatchTable(root)
	if err != nil {
		// Validation error (e.g., conflicting Stop handlers).
		// Log and keep the previous valid table rather than crashing.
		fmt.Fprintf(os.Stderr, "tui: dispatch table error: %v\n", err)
		return
	}
	a.dispatchTable = table
}
