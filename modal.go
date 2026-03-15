package tui

// Modal is a built-in component that renders a full-screen overlay.
// It supports backdrop dimming, focus trapping, and close-on-escape behavior.
type Modal struct {
	// Configuration (set via options)
	open            *State[bool]
	backdrop        string // "dim", "blank", "none"
	closeOnEscape   bool
	closeOnBackdrop bool
	trapFocus       bool
	elementOpts     []Option

	// Internal state
	app           *App
	element       *Element
	previousFocus Focusable
	wasOpen       bool
}

var (
	_ Component     = (*Modal)(nil)
	_ KeyListener   = (*Modal)(nil)
	_ MouseListener = (*Modal)(nil)
	_ AppBinder     = (*Modal)(nil)
)

// NewModal creates a new Modal with the given options.
func NewModal(opts ...ModalOption) *Modal {
	m := &Modal{
		backdrop:        "dim",
		closeOnEscape:   true,
		closeOnBackdrop: true,
		trapFocus:       true,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// BindApp wires the modal's state to the App.
func (m *Modal) BindApp(app *App) {
	m.app = app
	if m.open != nil {
		m.open.BindApp(app)
	}
}

// Render returns the modal's element tree.
// When open, registers the element as an overlay for post-render compositing.
// When closed, returns a hidden placeholder.
func (m *Modal) Render(app *App) *Element {
	isOpen := m.open != nil && m.open.Get()

	if !isOpen {
		// Closed: return hidden overlay placeholder
		if m.wasOpen {
			// Transition from open to closed
			if m.trapFocus && m.app != nil {
				m.app.focus.ClearScope()
			}
			if m.previousFocus != nil && m.app != nil {
				m.app.focus.SetFocus(m.previousFocus)
				m.previousFocus = nil
			}
			m.wasOpen = false
		}
		m.element = New(WithOverlay(true), WithHidden(true))
		return m.element
	}

	// Open: build the overlay container
	m.element = New(WithOverlay(true))
	for _, opt := range m.elementOpts {
		opt(m.element)
	}

	// Register overlay for the App's render pass
	app.registerOverlay(m.element, m.backdrop, m.trapFocus)

	// Handle open transition (save focus, scope)
	if !m.wasOpen {
		if m.app != nil {
			m.previousFocus = m.app.focus.Focused()
		}
		m.wasOpen = true
	}

	return m.element
}

// KeyMap returns key bindings for the modal.
// Escape closes the modal when closeOnEscape is true.
func (m *Modal) KeyMap() KeyMap {
	if m.open == nil || !m.open.Get() || !m.closeOnEscape {
		return nil
	}
	return KeyMap{
		OnStop(KeyEscape, func(ke KeyEvent) {
			m.open.Set(false)
		}),
	}
}

// HandleMouse handles backdrop click events.
// Clicking on the modal's backdrop (outside children) closes the modal.
func (m *Modal) HandleMouse(me MouseEvent) bool {
	if m.open == nil || !m.open.Get() || !m.closeOnBackdrop {
		return false
	}
	if me.Action != MousePress || me.Button != MouseLeft {
		return false
	}
	if m.element == nil {
		return false
	}
	// Check if click landed on the modal container itself (not on a child)
	hit := m.element.ElementAt(me.X, me.Y)
	if hit == m.element {
		m.open.Set(false)
		return true
	}
	return false
}
