package tui

// ModalOption configures a Modal component.
type ModalOption func(*Modal)

// WithModalOpen binds the modal's visibility to a boolean state.
func WithModalOpen(state *State[bool]) ModalOption {
	return func(m *Modal) {
		m.open = state
	}
}

// WithModalBackdrop sets the backdrop style: "dim" (default), "blank", or "none".
func WithModalBackdrop(b string) ModalOption {
	return func(m *Modal) {
		m.backdrop = b
	}
}

// WithModalCloseOnEscape configures whether Escape closes the modal (default true).
func WithModalCloseOnEscape(v bool) ModalOption {
	return func(m *Modal) {
		m.closeOnEscape = v
	}
}

// WithModalCloseOnBackdropClick configures whether clicking the backdrop closes the modal (default true).
func WithModalCloseOnBackdropClick(v bool) ModalOption {
	return func(m *Modal) {
		m.closeOnBackdrop = v
	}
}

// WithModalTrapFocus configures whether Tab/Shift+Tab is restricted to modal children (default true).
func WithModalTrapFocus(v bool) ModalOption {
	return func(m *Modal) {
		m.trapFocus = v
	}
}

// WithModalElementOptions passes through standard Element options to the modal's
// container element. Used by the code generator to apply class-derived layout options.
func WithModalElementOptions(opts ...Option) ModalOption {
	return func(m *Modal) {
		m.elementOpts = append(m.elementOpts, opts...)
	}
}
