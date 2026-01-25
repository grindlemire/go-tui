package tui

import (
	"os"
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

// App manages the application lifecycle: terminal setup, event loop, and rendering.
type App struct {
	terminal *ANSITerminal
	buffer   *Buffer
	reader   EventReader
	focus    *FocusManager
	root     Renderable
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
		terminal: terminal,
		buffer:   buffer,
		reader:   reader,
		focus:    focus,
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
		terminal: terminal,
		buffer:   buffer,
		reader:   reader,
		focus:    focus,
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

// SetRoot sets the root element tree for rendering.
// The root must implement Renderable (element.Element satisfies this).
// If root is an element.Element, focusable elements are auto-registered.
func (a *App) SetRoot(root Renderable) {
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
// Handles ResizeEvent internally by updating buffer size.
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

		return true
	}

	// Delegate to FocusManager for other events
	return a.focus.Dispatch(event)
}

// Render clears the buffer, renders the element tree, and flushes to terminal.
func (a *App) Render() {
	width, height := a.terminal.Size()

	// Clear buffer
	a.buffer.Clear()

	// If root exists, render the element tree
	if a.root != nil {
		a.root.Render(a.buffer, width, height)
	}

	// Render to terminal
	Render(a.terminal, a.buffer)
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
