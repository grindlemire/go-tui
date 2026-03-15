package tui

// KeyMap is a list of key bindings returned by KeyListener.KeyMap().
// It is a value, not a registration — the framework collects and manages it.
type KeyMap []KeyBinding

// KeyBinding associates a key pattern with a handler.
type KeyBinding struct {
	Pattern KeyPattern
	Handler func(KeyEvent)
	Stop    bool // If true, prevent later handlers from firing for this key
}

// KeyPattern identifies which key events match a binding.
type KeyPattern struct {
	Key           Key      // Specific key (KeyEscape, KeyBackspace, etc.), or 0
	Rune          rune     // Specific rune, or 0
	AnyRune       bool     // Match any printable character
	Mod           Modifier // Required modifiers (when non-zero, event must have exactly these mods)
	ExcludeMods   Modifier // Reject event if any of these modifiers are present
	FocusRequired bool     // When true, only dispatch when owning component is focused
}

// On creates a broadcast binding. Other handlers for the same key will also fire.
func On(m KeyMatcher, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{Pattern: m.keyPattern(), Handler: handler}
}

// OnStop creates a stop-propagation binding.
// No handlers registered after this one (in tree order) will fire for this event.
func OnStop(m KeyMatcher, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{Pattern: m.keyPattern(), Handler: handler, Stop: true}
}

// OnFocused creates a focus-gated stop-propagation binding.
// Only fires when the owning component's element is focused.
func OnFocused(m KeyMatcher, handler func(KeyEvent)) KeyBinding {
	p := m.keyPattern()
	p.FocusRequired = true
	return KeyBinding{Pattern: p, Handler: handler, Stop: true}
}

