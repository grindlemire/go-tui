package tui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
)

// InputLatencyBlocking is a special value for WithInputLatency that makes
// the event reader block indefinitely until input is available.
// This is more efficient for CPU usage but requires proper interrupt handling.
const InputLatencyBlocking = -1 * time.Millisecond

// AppOption is a functional option for configuring an App.
type AppOption func(*App) error

// WithInputLatency sets the polling timeout for the event reader.
// Default is 50ms. Use InputLatencyBlocking (-1) for blocking mode.
// A value of 0 is not allowed and will return an error.
func WithInputLatency(d time.Duration) AppOption {
	return func(a *App) error {
		if d == 0 {
			return fmt.Errorf("input latency of 0 (busy polling) is not allowed; use a positive duration or InputLatencyBlocking")
		}
		a.inputLatency = d
		return nil
	}
}

// WithFrameRate sets the target frame rate for the render loop.
// Default is 60 fps (16ms frame duration). Valid range is 1-240 fps.
func WithFrameRate(fps int) AppOption {
	return func(a *App) error {
		if fps < 1 {
			return fmt.Errorf("frame rate must be at least 1 fps")
		}
		if fps > 240 {
			return fmt.Errorf("frame rate cannot exceed 240 fps")
		}
		a.frameDuration = time.Second / time.Duration(fps)
		return nil
	}
}

// WithEventQueueSize sets the capacity of the event queue buffer.
// Default is 256. Must be at least 1.
func WithEventQueueSize(size int) AppOption {
	return func(a *App) error {
		if size < 1 {
			return fmt.Errorf("event queue size must be at least 1")
		}
		a.eventQueueSize = size
		return nil
	}
}

// WithGlobalKeyHandler sets a handler that runs before dispatching to focused element.
// If the handler returns true, the event is consumed and not dispatched further.
// Use this for app-level key bindings like quit.
func WithGlobalKeyHandler(fn func(KeyEvent) bool) AppOption {
	return func(a *App) error {
		a.globalKeyHandler = fn
		return nil
	}
}

// WithRoot sets the root view for rendering. Accepts:
//   - A view struct implementing Viewable (extracts Root, starts watchers)
//   - A raw Renderable (element.Element)
//
// The root is set after the app is fully initialized.
func WithRoot(v any) AppOption {
	return func(a *App) error {
		a.pendingRoot = v
		return nil
	}
}

// WithoutMouse disables mouse event reporting.
// By default, mouse events are enabled.
func WithoutMouse() AppOption {
	return func(a *App) error {
		a.mouseEnabled = false
		return nil
	}
}

// WithCursor keeps the cursor visible during app execution.
// By default, the cursor is hidden.
func WithCursor() AppOption {
	return func(a *App) error {
		a.cursorVisible = true
		return nil
	}
}

// WithInlineHeight enables inline widget mode.
// The TUI manages only the bottom N rows of the terminal, allowing normal
// terminal output above. Use PrintAbove() or PrintAboveln() to print
// scrolling content above the widget.
// In inline mode, alternate screen is not used, so terminal history is preserved.
func WithInlineHeight(rows int) AppOption {
	return func(a *App) error {
		if rows < 1 {
			return fmt.Errorf("inline height must be at least 1 row")
		}
		a.inlineHeight = rows
		return nil
	}
}

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

// watcherTreeWalker is used internally by App to discover and start
// watchers attached to elements in a tree. element.Element implements this.
type watcherTreeWalker interface {
	// WalkWatchers calls fn for each watcher in the element tree.
	WalkWatchers(fn func(Watcher))
}

// mouseHitTester is used internally by App to find the element at a given point.
// element.Element implements this via its ElementAtPoint method.
type mouseHitTester interface {
	// ElementAtPoint returns the deepest focusable-compatible element containing the point (x, y).
	// Returns nil if no element contains the point.
	ElementAtPoint(x, y int) Focusable
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

	// Configuration (set via options)
	inputLatency   time.Duration // Polling timeout for event reader (default 50ms, -1 for blocking)
	frameDuration  time.Duration // Duration per frame (default 16ms = 60fps)
	eventQueueSize int           // Capacity of event queue (default 256, used during construction)
	mouseEnabled   bool          // Whether mouse events are enabled (default true)
	cursorVisible  bool          // Whether cursor is visible (default false)
	pendingRoot    any           // Root to set after initialization (used by WithRoot)

	// Inline mode (set via WithInlineHeight)
	inlineHeight   int // Number of rows for inline widget (0 = full screen mode)
	inlineStartRow int // Terminal row where inline region starts (calculated at init)
}

// currentApp holds a reference to the currently running app for package-level Stop().
var currentApp *App

// Stop stops the currently running app. This is a package-level convenience function
// that allows stopping the app from event handlers without needing a direct reference.
// It is safe to call even if no app is running.
func Stop() {
	if currentApp != nil {
		currentApp.Stop()
	}
}

// NewApp creates a new application with the terminal set up for TUI usage.
// The terminal is put into raw mode and alternate screen mode (unless inline mode).
// Options can be passed to configure the app (e.g., WithInputLatency, WithInlineHeight).
func NewApp(opts ...AppOption) (*App, error) {
	// Create ANSITerminal from stdout/stdin
	terminal, err := NewANSITerminal(os.Stdout, os.Stdin)
	if err != nil {
		return nil, err
	}

	// Enter raw mode
	if err := terminal.EnterRawMode(); err != nil {
		return nil, err
	}

	// Create EventReader from stdin
	reader, err := NewEventReader(os.Stdin)
	if err != nil {
		// Clean up terminal state before returning error
		terminal.ExitRawMode()
		return nil, err
	}

	// Create empty FocusManager
	focus := NewFocusManager()

	// Create app with defaults (options may override these)
	// Note: buffer is created after options are applied to handle inline mode
	app := &App{
		terminal:       terminal,
		reader:         reader,
		focus:          focus,
		stopCh:         make(chan struct{}),
		stopped:        false,
		inputLatency:   50 * time.Millisecond,  // Default polling timeout
		frameDuration:  16 * time.Millisecond,  // Default ~60fps
		eventQueueSize: 256,                    // Default queue size
		mouseEnabled:   true,                   // Mouse enabled by default
		cursorVisible:  false,                  // Cursor hidden by default
	}

	// Apply options (may modify defaults above, including inlineHeight)
	for _, opt := range opts {
		if err := opt(app); err != nil {
			// Clean up on option error
			reader.Close()
			terminal.ExitRawMode()
			return nil, err
		}
	}

	// Get terminal size
	width, termHeight := terminal.Size()

	// Set up screen mode based on inline configuration
	if app.inlineHeight > 0 {
		// Inline mode: don't use alternate screen, reserve space at bottom
		// Clamp inline height to terminal height
		if app.inlineHeight > termHeight {
			app.inlineHeight = termHeight
		}

		// Print newlines to reserve space for the widget at the bottom
		fmt.Print(strings.Repeat("\n", app.inlineHeight))
		// Move cursor back up to the start of our region
		fmt.Printf("\033[%dA", app.inlineHeight)

		// Calculate where our inline region starts
		app.inlineStartRow = termHeight - app.inlineHeight

		// Create buffer sized for inline region only
		app.buffer = NewBuffer(width, app.inlineHeight)
	} else {
		// Full screen mode: use alternate screen
		terminal.EnterAltScreen()

		// Create buffer for full terminal
		app.buffer = NewBuffer(width, termHeight)
	}

	// Create event queue with configured size
	app.eventQueue = make(chan func(), app.eventQueueSize)

	// Apply terminal settings based on options
	if app.mouseEnabled {
		terminal.EnableMouse()
	}
	if !app.cursorVisible {
		terminal.HideCursor()
	}

	// If blocking mode, enable interrupt capability on the reader
	if app.inputLatency < 0 {
		if interruptible, ok := reader.(InterruptibleReader); ok {
			if err := interruptible.EnableInterrupt(); err != nil {
				reader.Close()
				if app.mouseEnabled {
					terminal.DisableMouse()
				}
				if !app.cursorVisible {
					terminal.ShowCursor()
				}
				if app.inlineHeight == 0 {
					terminal.ExitAltScreen()
				}
				terminal.ExitRawMode()
				return nil, fmt.Errorf("failed to enable interrupt for blocking mode: %w", err)
			}
		}
	}

	// Set pending root if provided via WithRoot option
	if app.pendingRoot != nil {
		app.SetRoot(app.pendingRoot)
		app.pendingRoot = nil
	}

	return app, nil
}

// NewAppWithReader creates an App with a custom EventReader.
// This is useful for testing or custom input handling.
// Options can be passed to configure the app (e.g., WithInputLatency, WithInlineHeight).
func NewAppWithReader(reader EventReader, opts ...AppOption) (*App, error) {
	// Create ANSITerminal from stdout/stdin
	terminal, err := NewANSITerminal(os.Stdout, os.Stdin)
	if err != nil {
		return nil, err
	}

	// Enter raw mode
	if err := terminal.EnterRawMode(); err != nil {
		return nil, err
	}

	// Create empty FocusManager
	focus := NewFocusManager()

	// Create app with defaults (options may override these)
	// Note: buffer is created after options are applied to handle inline mode
	app := &App{
		terminal:       terminal,
		reader:         reader,
		focus:          focus,
		stopCh:         make(chan struct{}),
		stopped:        false,
		inputLatency:   50 * time.Millisecond,  // Default polling timeout
		frameDuration:  16 * time.Millisecond,  // Default ~60fps
		eventQueueSize: 256,                    // Default queue size
		mouseEnabled:   true,                   // Mouse enabled by default
		cursorVisible:  false,                  // Cursor hidden by default
	}

	// Apply options (may modify defaults above, including inlineHeight)
	for _, opt := range opts {
		if err := opt(app); err != nil {
			// Clean up on option error
			reader.Close()
			terminal.ExitRawMode()
			return nil, err
		}
	}

	// Get terminal size
	width, termHeight := terminal.Size()

	// Set up screen mode based on inline configuration
	if app.inlineHeight > 0 {
		// Inline mode: don't use alternate screen, reserve space at bottom
		// Clamp inline height to terminal height
		if app.inlineHeight > termHeight {
			app.inlineHeight = termHeight
		}

		// Print newlines to reserve space for the widget at the bottom
		fmt.Print(strings.Repeat("\n", app.inlineHeight))
		// Move cursor back up to the start of our region
		fmt.Printf("\033[%dA", app.inlineHeight)

		// Calculate where our inline region starts
		app.inlineStartRow = termHeight - app.inlineHeight

		// Create buffer sized for inline region only
		app.buffer = NewBuffer(width, app.inlineHeight)
	} else {
		// Full screen mode: use alternate screen
		terminal.EnterAltScreen()

		// Create buffer for full terminal
		app.buffer = NewBuffer(width, termHeight)
	}

	// Create event queue with configured size
	app.eventQueue = make(chan func(), app.eventQueueSize)

	// Apply terminal settings based on options
	if app.mouseEnabled {
		terminal.EnableMouse()
	}
	if !app.cursorVisible {
		terminal.HideCursor()
	}

	// If blocking mode, enable interrupt capability on the reader
	if app.inputLatency < 0 {
		if interruptible, ok := reader.(InterruptibleReader); ok {
			if err := interruptible.EnableInterrupt(); err != nil {
				reader.Close()
				if app.mouseEnabled {
					terminal.DisableMouse()
				}
				if !app.cursorVisible {
					terminal.ShowCursor()
				}
				if app.inlineHeight == 0 {
					terminal.ExitAltScreen()
				}
				terminal.ExitRawMode()
				return nil, fmt.Errorf("failed to enable interrupt for blocking mode: %w", err)
			}
		}
	}

	// Set pending root if provided via WithRoot option
	if app.pendingRoot != nil {
		app.SetRoot(app.pendingRoot)
		app.pendingRoot = nil
	}

	return app, nil
}

// Close restores the terminal to its original state.
// Must be called when the application exits.
func (a *App) Close() error {
	// Disable mouse event reporting (only if it was enabled)
	if a.mouseEnabled {
		a.terminal.DisableMouse()
	}

	// Show cursor (only if it was hidden)
	if !a.cursorVisible {
		a.terminal.ShowCursor()
	}

	// Handle screen cleanup based on mode
	if a.inlineHeight > 0 {
		// Inline mode: clear the widget area and position cursor for shell
		a.terminal.SetCursor(0, a.inlineStartRow)
		a.terminal.ClearToEnd()
	} else {
		// Full screen mode: exit alternate screen
		a.terminal.ExitAltScreen()
	}

	// Exit raw mode
	if err := a.terminal.ExitRawMode(); err != nil {
		a.reader.Close()
		return err
	}

	// Close EventReader
	return a.reader.Close()
}

// PrintAbove prints content that scrolls up above the inline widget.
// Does not add a trailing newline. Use PrintAboveln for auto-newline.
// Only works in inline mode (WithInlineHeight). In full-screen mode, this is a no-op.
func (a *App) PrintAbove(format string, args ...any) {
	if a.inlineHeight == 0 {
		return
	}
	a.printAboveRaw(fmt.Sprintf(format, args...))
}

// PrintAboveln prints content with a trailing newline that scrolls up above the inline widget.
// Only works in inline mode (WithInlineHeight). In full-screen mode, this is a no-op.
func (a *App) PrintAboveln(format string, args ...any) {
	if a.inlineHeight == 0 {
		return
	}
	a.printAboveRaw(fmt.Sprintf(format, args...) + "\n")
}

// printAboveRaw handles the actual printing and scrolling for inline mode.
func (a *App) printAboveRaw(content string) {
	// Save cursor position
	fmt.Print("\033[s")
	// Move to the line just above our widget region
	fmt.Printf("\033[%d;1H", a.inlineStartRow)
	// Scroll the region above widget up by 1 line (makes room at bottom of scrollable area)
	fmt.Print("\033[S")
	// Move to the new blank line (one above the widget)
	fmt.Printf("\033[%d;1H", a.inlineStartRow)
	// Print the content
	fmt.Print(content)
	// Restore cursor position
	fmt.Print("\033[u")

	// Mark dirty to ensure widget redraws after we've scrolled
	MarkDirty()
}

// SetRoot sets the root view for rendering. Accepts:
//   - A view struct implementing Viewable (extracts Root, starts watchers)
//   - A raw Renderable (element.Element)
//
// If the root supports focus discovery, focusable elements are auto-registered.
// If the root supports watcher discovery, watchers are auto-started.
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

	// If root supports watcher discovery, start all watchers in the tree
	if walker, ok := root.(watcherTreeWalker); ok {
		walker.WalkWatchers(func(w Watcher) {
			w.Start(a.eventQueue, a.stopCh)
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

// EventQueue returns the event queue channel for manual watcher setup.
// Use with caution - prefer using SetRoot with Viewable for automatic watcher management.
func (a *App) EventQueue() chan<- func() {
	return a.eventQueue
}

// StopCh returns the stop channel for manual watcher setup.
// Use with caution - prefer using SetRoot with Viewable for automatic watcher management.
func (a *App) StopCh() <-chan struct{} {
	return a.stopCh
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
// Handles MouseEvent by hit-testing to find the element under the cursor.
// Returns true if the event was consumed.
func (a *App) Dispatch(event Event) bool {
	// Handle ResizeEvent specially
	if resize, ok := event.(ResizeEvent); ok {
		if a.inlineHeight > 0 {
			// Inline mode: recalculate start row, keep buffer height fixed
			a.inlineStartRow = resize.Height - a.inlineHeight
			// Only resize buffer width if it changed
			if a.buffer.Width() != resize.Width {
				a.buffer.Resize(resize.Width, a.inlineHeight)
			}
		} else {
			// Full screen mode: resize buffer to match terminal
			a.buffer.Resize(resize.Width, resize.Height)
		}

		// Mark root dirty so layout is recalculated
		if a.root != nil {
			a.root.MarkDirty()
		}

		// Schedule full redraw to clear any visual artifacts
		a.needsFullRedraw = true

		return true
	}

	// Handle MouseEvent by hit-testing to find the element under the cursor
	if mouse, ok := event.(MouseEvent); ok {
		if a.root == nil {
			return false
		}
		// Check if root supports hit-testing
		if hitTester, ok := a.root.(mouseHitTester); ok {
			if target := hitTester.ElementAtPoint(mouse.X, mouse.Y); target != nil {
				return target.HandleEvent(event)
			}
		}
		return false
	}

	// Delegate to FocusManager for other events
	return a.focus.Dispatch(event)
}

// Render clears the buffer, renders the element tree, and flushes to terminal.
// If a resize occurred since the last render, this automatically performs a full
// redraw to eliminate visual artifacts.
func (a *App) Render() {
	width, termHeight := a.terminal.Size()

	// Determine the render height based on mode
	renderHeight := termHeight
	if a.inlineHeight > 0 {
		renderHeight = a.inlineHeight
	}

	// Ensure buffer matches expected size (handles rapid resize)
	if a.buffer.Width() != width || a.buffer.Height() != renderHeight {
		if a.inlineHeight > 0 {
			// Inline mode: update start row, resize buffer width only if needed
			a.inlineStartRow = termHeight - a.inlineHeight
			if a.buffer.Width() != width {
				a.buffer.Resize(width, a.inlineHeight)
			}
		} else {
			// Full screen mode: clear terminal and resize buffer
			a.terminal.Clear()
			a.buffer.Resize(width, termHeight)
		}
		if a.root != nil {
			a.root.MarkDirty()
		}
		a.needsFullRedraw = true
	}

	// In inline mode, position cursor at start of managed region
	if a.inlineHeight > 0 {
		a.terminal.SetCursor(0, a.inlineStartRow)
	}

	// Clear buffer
	a.buffer.Clear()

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, renderHeight)
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
	// Set current app for package-level Stop()
	currentApp = a
	defer func() { currentApp = nil }()

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

// readInputEvents reads terminal input in a goroutine and queues events.
func (a *App) readInputEvents() {
	for {
		select {
		case <-a.stopCh:
			return
		default:
		}

		event, ok := a.reader.PollEvent(a.inputLatency)
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
