package tui

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// InputLatencyBlocking is a special value for WithInputLatency that makes
// the event reader block indefinitely until input is available.
// This is more efficient for CPU usage but requires proper interrupt handling.
const InputLatencyBlocking = -1 * time.Millisecond

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
	terminal        Terminal
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
	mouseEnabled   bool          // Whether mouse events are enabled
	mouseExplicit  bool          // Whether mouse setting was explicitly configured
	cursorVisible  bool          // Whether cursor is visible (default false)
	pendingRoot    any           // Root to set after initialization (used by WithRoot)

	// Inline mode (set via WithInlineHeight)
	inlineHeight   int // Number of rows for inline widget (0 = full screen mode)
	inlineStartRow int // Terminal row where inline region starts (calculated at init)
	historyRows    int // Number of history area rows that contain actual content

	// Dynamic alternate screen mode (for overlays like settings panels)
	inAlternateScreen   bool // Currently in alternate screen overlay
	savedInlineHeight   int  // Preserved inlineHeight when entering alternate
	savedInlineStartRow int  // Preserved inlineStartRow when entering alternate

	// Component model (mount system for struct components)
	mounts        *mountState
	dispatchTable *dispatchTable // Key broadcast dispatch table, rebuilt on dirty frames
	rootComponent Component      // Root struct component (set via SetRoot with Component)

	// Component watchers (from WatcherProvider components)
	componentWatchers        []Watcher
	componentWatchersStarted bool
}

// currentApp holds a reference to the currently running app for package-level Stop().
var currentApp *App

// NewApp creates a new application with the terminal set up for TUI usage.
// The terminal is put into raw mode and alternate screen mode (unless inline mode).
// Options can be passed to configure the app (e.g., WithInputLatency, WithInlineHeight).
//
// Mouse behavior:
//   - Full screen mode: mouse events enabled by default
//   - Inline mode: mouse events disabled by default (preserves terminal scrollback)
//   - Use WithMouse() or WithoutMouse() to explicitly override
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
		inputLatency:   50 * time.Millisecond, // Default polling timeout
		frameDuration:  16 * time.Millisecond, // Default ~60fps
		eventQueueSize: 256,                   // Default queue size
		cursorVisible:  false,                 // Cursor hidden by default
		mounts:         newMountState(),
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

	// Default mouse behavior based on mode (if not explicitly configured)
	if !app.mouseExplicit {
		// Inline mode: disable mouse to preserve terminal scrollback
		// Full screen mode: enable mouse for click/scroll handling
		app.mouseEnabled = app.inlineHeight == 0
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

	// Set currentApp so that Mount() works during SetRoot (Component.Render may call Mount).
	// Run() will set it again; this ensures it's available during initialization.
	currentApp = app

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
		inputLatency:   50 * time.Millisecond, // Default polling timeout
		frameDuration:  16 * time.Millisecond, // Default ~60fps
		eventQueueSize: 256,                   // Default queue size
		cursorVisible:  false,                 // Cursor hidden by default
		mounts:         newMountState(),
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

	// Default mouse behavior based on mode (if not explicitly configured)
	if !app.mouseExplicit {
		// Inline mode: disable mouse to preserve terminal scrollback
		// Full screen mode: enable mouse for click/scroll handling
		app.mouseEnabled = app.inlineHeight == 0
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

	// Set currentApp so that Mount() works during SetRoot (Component.Render may call Mount).
	// Run() will set it again; this ensures it's available during initialization.
	currentApp = app

	// Set pending root if provided via WithRoot option
	if app.pendingRoot != nil {
		app.SetRoot(app.pendingRoot)
		app.pendingRoot = nil
	}

	return app, nil
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
	case Component:
		// Struct component: render to get element tree, tag root for discovery.
		// Store as rootComponent so Run() re-renders the component each frame.
		a.rootComponent = view
		el := view.Render()
		el.component = view
		root = el
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

// walkComponents performs a DFS walk of the element tree, calling fn for
// each element that has an associated component (set by Mount).
// This is used to discover KeyListener and other component capabilities.
func walkComponents(root *Element, fn func(Component)) {
	if root == nil {
		return
	}
	if root.component != nil {
		fn(root.component)
	}
	for _, child := range root.children {
		walkComponents(child, fn)
	}
}
