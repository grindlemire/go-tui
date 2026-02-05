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
	Key           Key      // Specific key (KeyCtrlB, KeyEscape, etc.), or 0
	Rune          rune     // Specific rune, or 0
	AnyRune       bool     // Match any printable character
	Mod           Modifier // Required modifiers (when non-zero, event must have exactly these mods)
	RequireNoMods bool     // When true, event must have no modifiers (Mod field is ignored)
}

// OnKey creates a broadcast binding for a specific key.
// Other handlers for the same key will also fire.
func OnKey(key Key, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Key: key},
		Handler: handler,
		Stop:    false,
	}
}

// OnKeyStop creates a stop-propagation binding for a specific key.
// No handlers registered after this one (in tree order) will fire.
func OnKeyStop(key Key, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Key: key},
		Handler: handler,
		Stop:    true,
	}
}

// OnRune creates a broadcast binding for a specific printable character.
func OnRune(r rune, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Rune: r},
		Handler: handler,
		Stop:    false,
	}
}

// OnRuneStop creates a stop-propagation binding for a specific printable character.
func OnRuneStop(r rune, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Rune: r},
		Handler: handler,
		Stop:    true,
	}
}

// OnRunes creates a broadcast binding for all printable characters.
func OnRunes(handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{AnyRune: true},
		Handler: handler,
		Stop:    false,
	}
}

// OnRunesStop creates a stop-propagation binding for all printable characters.
// Use this for text inputs that need exclusive access to character keys.
func OnRunesStop(handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{AnyRune: true},
		Handler: handler,
		Stop:    true,
	}
}
