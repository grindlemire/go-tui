package tui

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/grindlemire/go-tui/internal/debug"
)

// Run starts the main event loop. Blocks until Stop() is called or SIGINT received.
// Rendering occurs only when the dirty flag is set (by mutations).
func (a *App) Run() error {
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

	// Handle Ctrl+Z / SIGTSTP for job control
	cleanupSuspend := a.registerSuspendSignals()
	defer cleanupSuspend()

	// Start input reader in background
	go a.readInputEvents()

	// Initial render
	debug.Log("App.Run: initial render")
	a.Render()
	a.rebuildDispatchTable()

	// Frame-based loop with configurable frame timing
	for {
		frameStart := time.Now()

		// Process events for up to half the frame budget (non-blocking)
		eventDeadline := frameStart.Add(a.frameDuration / 2)
	drain:
		for time.Now().Before(eventDeadline) {
			select {
			case handler := <-a.eventQueue:
				handler()
			case handler := <-a.updateQueue:
				handler()
			case <-a.stopCh:
				return nil
			default:
				// No more events, move to render phase
				break drain
			}
		}

		// Always render if dirty
		if a.checkAndClearDirty() {
			debug.Log("App.Run: dirty frame, re-rendering")
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

// QueueUpdate enqueues a function to run on the main loop.
// Safe to call from any goroutine. Use this for background thread safety.
func (a *App) QueueUpdate(fn func()) {
	if fn == nil {
		return
	}
	if a.updateQueue == nil {
		// Back-compat path for tests/mocks that construct App manually.
		select {
		case a.eventQueue <- fn:
		case <-a.stopCh:
		default:
		}
		return
	}

	// Bounded queue: drop oldest background update when full.
	// Input/watcher events use eventQueue and are lossless.
	for {
		select {
		case a.updateQueue <- fn:
			return
		case <-a.stopCh:
			return
		default:
			select {
			case <-a.updateQueue:
			case <-a.stopCh:
				return
			default:
			}
		}
	}
}

// rebuildDispatchTable walks the rendered element tree and builds a new
// dispatch table from all mounted components' KeyMap() methods.
// If validation fails, the previous table is kept.
func (a *App) rebuildDispatchTable() {
	if a.root == nil {
		return
	}

	table, err := buildDispatchTable(a.rootComponent, a.root, a.focus)
	if err != nil {
		// Validation error (e.g., conflicting Stop handlers).
		// Log and keep the previous valid table rather than crashing.
		fmt.Fprintf(os.Stderr, "tui: dispatch table error: %v\n", err)
		return
	}
	debug.Log("rebuildDispatchTable: %d entries", len(table.entries))
	for i, e := range table.entries {
		debug.Log("rebuildDispatchTable:   [%d] pos=%d key=%v rune=%c anyRune=%v focusReq=%v stop=%v focusCheck=%v",
			i, e.position, e.pattern.Key, e.pattern.Rune, e.pattern.AnyRune, e.pattern.FocusRequired, e.stop, e.focusCheck != nil)
	}
	a.dispatchTable = table
}
