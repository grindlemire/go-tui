package tui

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// InputLatencyBlocking is a special value for WithInputLatency that makes
// the event reader block indefinitely until input is available.
// This is more efficient for CPU usage but requires proper interrupt handling.
const InputLatencyBlocking = -1 * time.Millisecond

// Viewable is implemented by types that provide a root element tree.
// Generated view structs, *Element, and struct components all implement this.
// Used by SetRootView, PrintAboveElement, and StreamWriter.WriteElement.
type Viewable interface {
	GetRoot() *Element
	GetWatchers() []Watcher
}

// App manages the application lifecycle: terminal setup, event loop, and rendering.
type App struct {
	terminal        Terminal
	buffer          *Buffer
	reader          EventReader
	focus           *focusManager
	root            *Element
	needsFullRedraw bool // Set after resize, cleared after RenderFull
	dirty           atomic.Bool
	batch           batchContext

	// Event loop fields
	inputEvents      chan Event  // Terminal input events (key, mouse, resize)
	updates          chan Event  // Background events (watchers, QueueUpdate, Suspend)
	merged           chan Event  // Fan-in of inputEvents + updates with input priority
	watcherQueue     chan func() // Bridge channel for Watcher interface compatibility
	stopCh           chan struct{}
	stopped          bool
	stopOnce         sync.Once
	closeOnce        sync.Once
	opened           atomic.Bool
	signalCleanup    func()              // Cleans up signal handlers (set by Open)
	selfSuspended    atomic.Bool         // True during self-initiated suspend; prevents double resume from SIGCONT handler
	globalKeyHandler func(KeyEvent) bool // Returns true if event consumed

	// Configuration (set via options)
	inputLatency     time.Duration // Polling timeout for event reader (default: blocking, use positive duration for polling)
	frameDuration    time.Duration // Duration per frame (default 16ms = 60fps)
	eventQueueSize   int           // Capacity of event queue (default 256, used during construction)
	mouseEnabled     bool          // Whether mouse events are enabled
	mouseExplicit    bool          // Whether mouse setting was explicitly configured
	cursorVisible    bool          // Whether cursor is visible (default false)
	legacyKeyboard   bool          // Force legacy keyboard mode (skip Kitty protocol negotiation)
	onSuspend        func()        // Called before suspending (Ctrl+Z / SIGTSTP)
	onResume         func()        // Called after resuming (SIGCONT)
	pendingRootApply func(*App)    // Root setter to run after initialization (used by WithRoot* options)

	// Inline mode (set via WithInlineHeight)
	inlineHeight       int               // Number of rows for inline widget (0 = full screen mode)
	inlineStartRow     int               // Terminal row where inline region starts (calculated at init)
	inlineStartupMode  InlineStartupMode // Startup behavior for inline viewport ownership
	inlineLayout       inlineLayoutState
	inlineSession      *inlineSession
	activeStreamWriter *inlineStreamWriter

	// Overlay rendering
	overlays []*overlayEntry

	// Dynamic alternate screen mode (for overlays like settings panels)
	inAlternateScreen   bool // Currently in alternate screen overlay
	savedInlineHeight   int  // Preserved inlineHeight when entering alternate
	savedInlineStartRow int  // Preserved inlineStartRow when entering alternate
	savedInlineLayout   inlineLayoutState

	// Component model (mount system for struct components)
	mounts        *mountState
	dispatchTable *dispatchTable // Key broadcast dispatch table, rebuilt on dirty frames
	rootComponent Component      // Root struct component (set via SetRoot with Component)

	// Component watchers (from WatcherProvider components)
	componentWatchers        []Watcher
	componentWatchersStarted bool

	// Root-scoped watcher lifecycle.
	rootStopCh    chan struct{}
	rootWatcherCh <-chan struct{}

	// Topic-based event registry (scoped per app instance).
	topicMu sync.RWMutex
	topics  map[string]*topicSubscription
}

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
	focus := newFocusManager()

	// Create app with defaults (options may override these)
	// Note: buffer is created after options are applied to handle inline mode
	app := &App{
		terminal:       terminal,
		reader:         reader,
		focus:          focus,
		stopCh:         make(chan struct{}),
		stopped:        false,
		inputLatency:   InputLatencyBlocking,  // Default: block until input arrives
		frameDuration:  16 * time.Millisecond, // Default ~60fps
		eventQueueSize: 256,                   // Default queue size
		cursorVisible:  false,                 // Cursor hidden by default
		mounts:         newMountState(),
		batch:          newBatchContext(),
	}
	app.resetRootSession()

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

	// Set up screen mode BEFORE Kitty negotiation. The Kitty keyboard
	// protocol stack is per-screen: entering alternate screen starts a
	// fresh stack, discarding any modes pushed on the primary screen.
	app.setupInitialScreen(width, termHeight)

	// Negotiate Kitty keyboard protocol unless legacy mode was requested
	if !app.legacyKeyboard {
		terminal.NegotiateKittyKeyboard()
	}

	// Create event channels and start background goroutines
	app.inputEvents = make(chan Event, app.eventQueueSize)
	app.updates = make(chan Event, app.eventQueueSize)
	app.merged = make(chan Event, app.eventQueueSize)
	app.watcherQueue = make(chan func(), app.eventQueueSize)
	app.startWatcherBridge()
	app.startEventMerge()

	// Apply terminal settings based on options
	if app.mouseEnabled {
		terminal.EnableMouse()
	}
	if !app.cursorVisible {
		terminal.HideCursor()
	}

	// Enable interrupt capability on the reader for SIGWINCH and shutdown wakeup
	if interruptible, ok := reader.(InterruptibleReader); ok {
		if err := interruptible.EnableInterrupt(); err != nil {
			app.Stop() // Stop background goroutines (startWatcherBridge, startEventMerge)
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
			terminal.DisableKittyKeyboard()
			terminal.ExitRawMode()
			return nil, fmt.Errorf("failed to enable interrupt: %w", err)
		}
	}

	// Set pending root if provided via WithRoot* option.
	if app.pendingRootApply != nil {
		app.pendingRootApply(app)
		app.pendingRootApply = nil
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
	focus := newFocusManager()

	// Create app with defaults (options may override these)
	// Note: buffer is created after options are applied to handle inline mode
	app := &App{
		terminal:       terminal,
		reader:         reader,
		focus:          focus,
		stopCh:         make(chan struct{}),
		stopped:        false,
		inputLatency:   InputLatencyBlocking,  // Default: block until input arrives
		frameDuration:  16 * time.Millisecond, // Default ~60fps
		eventQueueSize: 256,                   // Default queue size
		cursorVisible:  false,                 // Cursor hidden by default
		mounts:         newMountState(),
		batch:          newBatchContext(),
	}
	app.resetRootSession()

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

	// Set up screen mode BEFORE Kitty negotiation. The Kitty keyboard
	// protocol stack is per-screen: entering alternate screen starts a
	// fresh stack, discarding any modes pushed on the primary screen.
	app.setupInitialScreen(width, termHeight)

	// Negotiate Kitty keyboard protocol unless legacy mode was requested
	if !app.legacyKeyboard {
		terminal.NegotiateKittyKeyboard()
	}

	// Create event channels and start background goroutines
	app.inputEvents = make(chan Event, app.eventQueueSize)
	app.updates = make(chan Event, app.eventQueueSize)
	app.merged = make(chan Event, app.eventQueueSize)
	app.watcherQueue = make(chan func(), app.eventQueueSize)
	app.startWatcherBridge()
	app.startEventMerge()

	// Apply terminal settings based on options
	if app.mouseEnabled {
		terminal.EnableMouse()
	}
	if !app.cursorVisible {
		terminal.HideCursor()
	}

	// Enable interrupt capability on the reader for SIGWINCH and shutdown wakeup
	if interruptible, ok := reader.(InterruptibleReader); ok {
		if err := interruptible.EnableInterrupt(); err != nil {
			app.Stop() // Stop background goroutines (startWatcherBridge, startEventMerge)
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
			terminal.DisableKittyKeyboard()
			terminal.ExitRawMode()
			return nil, fmt.Errorf("failed to enable interrupt: %w", err)
		}
	}

	// Set pending root if provided via WithRoot* option.
	if app.pendingRootApply != nil {
		app.pendingRootApply(app)
		app.pendingRootApply = nil
	}

	return app, nil
}

// SetRoot sets the root element for rendering.
func (a *App) SetRoot(root *Element) {
	a.rootComponent = nil
	a.applyRoot(root)
}

// SetRootView sets the root from a Viewable and starts its watchers.
func (a *App) SetRootView(view Viewable) {
	a.rootComponent = nil
	if binder, ok := view.(AppBinder); ok {
		binder.BindApp(a)
	}
	a.applyRoot(view.GetRoot())
	if binder, ok := view.(AppBinder); ok {
		binder.BindApp(a)
	}
	for _, w := range view.GetWatchers() {
		w.Start(a.watcherQueue, a.rootWatcherCh)
	}
}

// SetRootComponent sets the root from a struct component.
func (a *App) SetRootComponent(component Component) {
	if binder, ok := component.(AppBinder); ok {
		binder.BindApp(a)
	}
	a.rootComponent = component
	el := component.Render(a)
	a.applyRoot(el)
	if binder, ok := component.(AppBinder); ok {
		binder.BindApp(a)
	}
}

func (a *App) applyRoot(root *Element) {
	a.resetRootSession()
	a.root = root
	if root == nil {
		return
	}
	root.setAppRecursive(a)
	a.MarkDirty()

	// Set up callback for future focusable elements
	root.SetOnFocusableAdded(func(f Focusable) {
		a.focus.Register(f)
	})

	// Discover existing focusable elements in tree
	root.WalkFocusables(func(f Focusable) {
		a.focus.Register(f)
	})

	// Start all watchers in the tree
	root.WalkWatchers(func(w Watcher) {
		w.Start(a.watcherQueue, a.rootWatcherCh)
	})
}

func (a *App) resetRootSession() {
	if a.rootStopCh != nil {
		close(a.rootStopCh)
	}
	a.rootStopCh = make(chan struct{})
	a.rootWatcherCh = mergeStopChannels(a.stopCh, a.rootStopCh)
	a.focus = newFocusManager()
	a.dispatchTable = nil
	a.mounts = newMountState()
	a.componentWatchers = nil
	a.componentWatchersStarted = false
	a.topicMu.Lock()
	a.topics = make(map[string]*topicSubscription)
	a.topicMu.Unlock()
}

func mergeStopChannels(ch1, ch2 <-chan struct{}) <-chan struct{} {
	merged := make(chan struct{})
	go func() {
		select {
		case <-ch1:
		case <-ch2:
		}
		close(merged)
	}()
	return merged
}

// SetGlobalKeyHandler sets a handler that runs before dispatching to focused element.
// If the handler returns true, the event is consumed and not dispatched further.
// Use this for app-level key bindings like quit.
func (a *App) SetGlobalKeyHandler(fn func(KeyEvent) bool) {
	a.globalKeyHandler = fn
}

// Root returns the current root element.
func (a *App) Root() *Element {
	return a.root
}

// Size returns the current terminal size.
func (a *App) Size() (width, height int) {
	return a.terminal.Size()
}

// FocusNext moves focus to the next focusable element.
func (a *App) FocusNext() {
	a.focus.Next()
}

// FocusPrev moves focus to the previous focusable element.
func (a *App) FocusPrev() {
	a.focus.Prev()
}

// BlurFocused blurs the currently focused element, leaving nothing focused.
func (a *App) BlurFocused() {
	a.focus.ClearFocus()
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

// EventQueue returns the watcher queue channel for manual watcher setup.
// Use with caution - prefer using SetRoot with Viewable for automatic watcher management.
func (a *App) EventQueue() chan<- func() {
	return a.watcherQueue
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

// startWatcherBridge starts a goroutine that forwards closures from the
// watcherQueue to the updates channel as UpdateEvents.
func (a *App) startWatcherBridge() {
	go func() {
		for {
			select {
			case fn, ok := <-a.watcherQueue:
				if !ok {
					return
				}
				select {
				case a.updates <- UpdateEvent{fn: fn}:
				case <-a.stopCh:
					return
				}
			case <-a.stopCh:
				return
			}
		}
	}()
}

// startEventMerge starts a goroutine that merges inputEvents and updates
// into the merged channel with input priority. Input events are always
// forwarded before background updates when both are pending.
func (a *App) startEventMerge() {
	go func() {
		for {
			// Priority: always drain inputEvents first
			select {
			case ev := <-a.inputEvents:
				select {
				case a.merged <- ev:
				case <-a.stopCh:
					return
				}
				continue
			default:
			}
			// No input pending; take from either channel
			select {
			case ev := <-a.inputEvents:
				select {
				case a.merged <- ev:
				case <-a.stopCh:
					return
				}
			case ev := <-a.updates:
				select {
				case a.merged <- ev:
				case <-a.stopCh:
					return
				}
			case <-a.stopCh:
				return
			}
		}
	}()
}

// walkComponents performs a BFS walk of the element tree, calling fn for
// each element that has an associated component (set by Mount).
// If rootComp is non-nil it is visited first, before the tree walk.
// BFS order means shallower components are visited before deeper ones,
// so a parent's handlers always fire before any descendant's handlers
// regardless of tree branching structure.
func walkComponents(rootComp Component, root *Element, fn func(Component)) {
	if rootComp != nil {
		fn(rootComp)
	}
	if root == nil {
		return
	}
	queue := []*Element{root}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node.component != nil {
			fn(node.component)
		}
		queue = append(queue, node.children...)
	}
}
