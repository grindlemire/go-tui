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
	app              *App
	element          *Element
	previousFocusIdx int // focus index before modal opened (-1 = none)
	wasOpen          bool
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
		backdrop:         "dim",
		closeOnEscape:    true,
		closeOnBackdrop:  true,
		trapFocus:        true,
		previousFocusIdx: -1,
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
			if m.previousFocusIdx >= 0 && m.app != nil {
				m.app.focus.setFocusIndex(m.previousFocusIdx)
				m.previousFocusIdx = -1
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

	// Handle open transition: save previous focus index
	needsFocusInit := false
	if !m.wasOpen {
		if m.app != nil {
			m.previousFocusIdx = m.app.focus.focusedIndex()
		}
		m.wasOpen = true
		needsFocusInit = true
	}

	// Register overlay for the App's render pass.
	// needsFocusInit tells the render pass to move focus into the modal
	// on the first frame (after children are attached).
	app.registerOverlay(m.element, m.backdrop, m.trapFocus, needsFocusInit)

	return m.element
}

// KeyMap returns key bindings for the modal.
// All bindings are preemptive: they fire before parent component handlers,
// preventing parent keys from leaking through when the modal is open.
// A catch-all binding consumes any unhandled keys.
func (m *Modal) KeyMap() KeyMap {
	if m.open == nil || !m.open.Get() {
		return nil
	}
	// In inline mode without alternate screen, overlays are not rendered
	// (registerOverlay silently skips them). Returning preemptive bindings
	// here would block all keyboard input with no visible modal.
	if m.app != nil && !m.app.inAlternateScreen && m.app.inlineHeight > 0 {
		return nil
	}
	var km KeyMap
	if m.closeOnEscape {
		km = append(km, OnPreemptStop(KeyEscape, func(ke KeyEvent) {
			m.open.Set(false)
		}))
	}
	if m.trapFocus && m.app != nil {
		km = append(km,
			OnPreemptStop(KeyTab, func(ke KeyEvent) {
				m.app.FocusNext()
			}),
			OnPreemptStop(KeyTab.Shift(), func(ke KeyEvent) {
				m.app.FocusPrev()
			}),
		)
	}
	// Enter activates the focused element's onActivate callback
	km = append(km, OnPreemptStop(KeyEnter, func(ke KeyEvent) {
		if m.app == nil {
			return
		}
		if focused, ok := m.app.Focused().(*Element); ok && focused != nil {
			focused.Activate()
		}
	}))
	// Catch-all: block all other keys from reaching parent handlers
	km = append(km, OnPreemptStop(AnyKey, func(ke KeyEvent) {}))
	return km
}

// HandleMouse handles click events within the modal.
// Clicking a child with onActivate triggers it. Clicking the backdrop closes the modal.
func (m *Modal) HandleMouse(me MouseEvent) bool {
	if m.open == nil || !m.open.Get() {
		return false
	}
	if me.Action != MousePress || me.Button != MouseLeft {
		return false
	}
	if m.element == nil {
		return false
	}
	hit := m.element.ElementAt(me.X, me.Y)
	if hit == nil {
		return false
	}
	// Backdrop click (hit the overlay container itself, not a child)
	if hit == m.element {
		if m.closeOnBackdrop {
			m.open.Set(false)
		}
		return true // always consume backdrop clicks
	}
	// Check if the clicked element (or an ancestor up to the overlay) has onActivate
	for el := hit; el != nil && el != m.element; el = el.parent {
		if el.onActivate != nil {
			el.Activate()
			return true
		}
	}
	// Click landed inside the modal on non-activatable content.
	// Consume it to prevent leaking to parent handlers.
	return true
}
