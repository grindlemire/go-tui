package tui

import (
	"os"
	"os/signal"
	"time"
)

// Renderable is implemented by types that can be rendered to a buffer.
// This is typically implemented by element.Element.
type Renderable interface {
	// Render calculates layout (if dirty) and renders to the buffer.
	Render(buf *Buffer, width, height int)

	// MarkDirty marks the element as needing layout recalculation.
	MarkDirty()

	// IsDirty returns whether the element needs recalculation.
	IsDirty() bool
}

// focusableTreeWalker is used internally by App to discover and register
// focusable elements in an element tree. element.Element implements this.
type focusableTreeWalker interface {
	// SetOnFocusableAdded sets a callback called when a focusable descendant is added.
	SetOnFocusableAdded(fn func(Focusable))

	// WalkFocusables calls fn for each focusable element in the tree.
	WalkFocusables(fn func(Focusable))
}

// Viewable is implemented by generated view structs.
// Allows SetRoot to extract the root element and start watchers.
type Viewable interface {
	GetRoot() Renderable
	GetWatchers() []Watcher
}

// App manages the application lifecycle: terminal setup, event loop, and rendering.
type App struct {
	terminal        *ANSITerminal
	buffer          *Buffer
	reader          EventReader
	focus           *FocusManager
	root            Renderable
	needsFullRedraw bool // Set after resize, cleared after RenderFull

	// Event loop fields
	eventQueue       chan func()
	stopCh           chan struct{}
	stopped          bool
	globalKeyHandler func(KeyEvent) bool // Returns true if event consumed
}

// NewApp creates a new application with the terminal set up for TUI usage.
// The terminal is put into raw mode and alternate screen mode.
func NewApp() (*App, error) {
	// Create ANSITerminal from stdout/stdin
	terminal, err := NewANSITerminal(os.Stdout, os.Stdin)
	if err != nil {
		return nil, err
	}

	// Enter raw mode
	if err := terminal.EnterRawMode(); err != nil {
		return nil, err
	}

	// Enter alternate screen
	terminal.EnterAltScreen()

	// Hide cursor
	terminal.HideCursor()

	// Get terminal size and create buffer
	width, height := terminal.Size()
	buffer := NewBuffer(width, height)

	// Create EventReader from stdin
	reader, err := NewEventReader(os.Stdin)
	if err != nil {
		// Clean up terminal state before returning error
		terminal.ShowCursor()
		terminal.ExitAltScreen()
		terminal.ExitRawMode()
		return nil, err
	}

	// Create empty FocusManager
	focus := NewFocusManager()

	return &App{
		terminal:   terminal,
		buffer:     buffer,
		reader:     reader,
		focus:      focus,
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}, nil
}

// NewAppWithReader creates an App with a custom EventReader.
// This is useful for testing or custom input handling.
func NewAppWithReader(reader EventReader) (*App, error) {
	// Create ANSITerminal from stdout/stdin
	terminal, err := NewANSITerminal(os.Stdout, os.Stdin)
	if err != nil {
		return nil, err
	}

	// Enter raw mode
	if err := terminal.EnterRawMode(); err != nil {
		return nil, err
	}

	// Enter alternate screen
	terminal.EnterAltScreen()

	// Hide cursor
	terminal.HideCursor()

	// Get terminal size and create buffer
	width, height := terminal.Size()
	buffer := NewBuffer(width, height)

	// Create empty FocusManager
	focus := NewFocusManager()

	return &App{
		terminal:   terminal,
		buffer:     buffer,
		reader:     reader,
		focus:      focus,
		eventQueue: make(chan func(), 256),
		stopCh:     make(chan struct{}),
		stopped:    false,
	}, nil
}

// Close restores the terminal to its original state.
// Must be called when the application exits.
func (a *App) Close() error {
	// Show cursor
	a.terminal.ShowCursor()

	// Exit alternate screen
	a.terminal.ExitAltScreen()

	// Exit raw mode
	if err := a.terminal.ExitRawMode(); err != nil {
		a.reader.Close()
		return err
	}

	// Close EventReader
	return a.reader.Close()
}

// SetRoot sets the root view for rendering. Accepts:
//   - A view struct implementing Viewable (extracts Root, starts watchers)
//   - A raw Renderable (element.Element)
//
// If the root supports focus discovery, focusable elements are auto-registered.
func (a *App) SetRoot(v any) {
	var root Renderable

	switch view := v.(type) {
	case Viewable:
		root = view.GetRoot()
		// Start all watchers collected during component construction
		for _, w := range view.GetWatchers() {
			w.Start(a.eventQueue, a.stopCh)
		}
	case Renderable:
		root = view
	default:
		// Invalid type - ignore
		return
	}

	a.root = root

	// If root supports focus discovery, set up auto-registration
	if walker, ok := root.(focusableTreeWalker); ok {
		// Set up callback for future focusable elements
		walker.SetOnFocusableAdded(func(f Focusable) {
			a.focus.Register(f)
		})

		// Discover existing focusable elements in tree
		walker.WalkFocusables(func(f Focusable) {
			a.focus.Register(f)
		})
	}
}

// SetGlobalKeyHandler sets a handler that runs before dispatching to focused element.
// If the handler returns true, the event is consumed and not dispatched further.
// Use this for app-level key bindings like quit.
func (a *App) SetGlobalKeyHandler(fn func(KeyEvent) bool) {
	a.globalKeyHandler = fn
}

// Root returns the current root element.
func (a *App) Root() Renderable {
	return a.root
}

// Size returns the current terminal size.
func (a *App) Size() (width, height int) {
	return a.terminal.Size()
}

// Focus returns the FocusManager for this app.
// Deprecated: Use FocusNext, FocusPrev, and Focused instead.
func (a *App) Focus() *FocusManager {
	return a.focus
}

// FocusNext moves focus to the next focusable element.
func (a *App) FocusNext() {
	a.focus.Next()
}

// FocusPrev moves focus to the previous focusable element.
func (a *App) FocusPrev() {
	a.focus.Prev()
}

// Focused returns the currently focused element, or nil if none.
func (a *App) Focused() Focusable {
	return a.focus.Focused()
}

// Terminal returns the underlying terminal.
// Use with caution for advanced use cases.
func (a *App) Terminal() Terminal {
	return a.terminal
}

// Buffer returns the underlying buffer.
// Use with caution for advanced use cases.
func (a *App) Buffer() *Buffer {
	return a.buffer
}

// PollEvent reads the next event with a timeout.
// Convenience wrapper around the EventReader.
func (a *App) PollEvent(timeout time.Duration) (Event, bool) {
	return a.reader.PollEvent(timeout)
}

// Dispatch sends an event to the focused element.
// Handles ResizeEvent internally by updating buffer size and scheduling a full redraw.
// Returns true if the event was consumed.
func (a *App) Dispatch(event Event) bool {
	// Handle ResizeEvent specially
	if resize, ok := event.(ResizeEvent); ok {
		// Resize buffer
		a.buffer.Resize(resize.Width, resize.Height)

		// Mark root dirty so layout is recalculated
		if a.root != nil {
			a.root.MarkDirty()
		}

		// Schedule full redraw to clear any visual artifacts
		a.needsFullRedraw = true

		return true
	}

	// Delegate to FocusManager for other events
	return a.focus.Dispatch(event)
}

// Render clears the buffer, renders the element tree, and flushes to terminal.
// If a resize occurred since the last render, this automatically performs a full
// redraw to eliminate visual artifacts.
func (a *App) Render() {
	width, height := a.terminal.Size()

	// Ensure buffer matches current terminal size (handles rapid resize)
	if a.buffer.Width() != width || a.buffer.Height() != height {
		// Clear terminal to remove any corrupt content from resize
		a.terminal.Clear()

		// Resize buffer to match terminal
		a.buffer.Resize(width, height)
		if a.root != nil {
			a.root.MarkDirty()
		}
		a.needsFullRedraw = true
	}

	// Clear buffer
	a.buffer.Clear()

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, height)
	}

	// Use full redraw after resize to clear artifacts, otherwise use diff-based render
	if a.needsFullRedraw {
		RenderFull(a.terminal, a.buffer)
		a.needsFullRedraw = false
	} else {
		Render(a.terminal, a.buffer)
	}
}

// RenderFull forces a complete redraw of the buffer to the terminal.
// Use this after resize events or when the terminal may be corrupted.
func (a *App) RenderFull() {
	width, height := a.terminal.Size()

	// Clear buffer
	a.buffer.Clear()

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, height)
	}

	// Full render to terminal
	RenderFull(a.terminal, a.buffer)
}

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

	// Start input reader in background
	go a.readInputEvents()

	// Initial render
	a.Render()

	for !a.stopped {
		// Block until at least one event arrives
		select {
		case handler := <-a.eventQueue:
			handler()
		case <-a.stopCh:
			return nil
		}

		// Drain any additional queued events (batch processing)
	drain:
		for {
			select {
			case handler := <-a.eventQueue:
				handler()
			default:
				break drain
			}
		}

		// Only render if something changed (dirty flag set by mutations)
		if checkAndClearDirty() {
			a.Render()
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

// readInputEvents reads terminal input in a goroutine and queues events.
func (a *App) readInputEvents() {
	for {
		select {
		case <-a.stopCh:
			return
		default:
		}

		event, ok := a.reader.PollEvent(50 * time.Millisecond)
		if !ok {
			continue
		}

		// Capture event for closure
		ev := event

		a.eventQueue <- func() {
			// Global key handler runs first (for app-level bindings like quit)
			if keyEvent, isKey := ev.(KeyEvent); isKey {
				if a.globalKeyHandler != nil && a.globalKeyHandler(keyEvent) {
					return // Event consumed by global handler
				}
			}
			a.Dispatch(ev)
		}
	}
}
