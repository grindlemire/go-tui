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
	FocusRequired bool     // When true, only dispatch when owning component is focused
}

// OnKey creates a broadcast binding for a specific key with no modifiers.
// Other handlers for the same key will also fire.
// Use OnKeyMod to match a key with specific modifiers (e.g., Shift+Tab).
func OnKey(key Key, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Key: key, RequireNoMods: true},
		Handler: handler,
		Stop:    false,
	}
}

// OnKeyStop creates a stop-propagation binding for a specific key with no modifiers.
// No handlers registered after this one (in tree order) will fire.
// Use OnKeyModStop to match a key with specific modifiers.
func OnKeyStop(key Key, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Key: key, RequireNoMods: true},
		Handler: handler,
		Stop:    true,
	}
}

// OnKeyMod creates a broadcast binding for a key with specific modifiers.
// Example: OnKeyMod(KeyTab, ModShift, handler) matches Shift+Tab.
func OnKeyMod(key Key, mod Modifier, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Key: key, Mod: mod},
		Handler: handler,
		Stop:    false,
	}
}

// OnKeyModStop creates a stop-propagation binding for a key with specific modifiers.
// Example: OnKeyModStop(KeyTab, ModShift, handler) matches Shift+Tab exclusively.
func OnKeyModStop(key Key, mod Modifier, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Key: key, Mod: mod},
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

// OnKeyFocused creates a focus-gated stop-propagation binding for a specific key.
// Only fires when the owning component's element is focused.
func OnKeyFocused(key Key, handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{Key: key, RequireNoMods: true, FocusRequired: true},
		Handler: handler,
		Stop:    true,
	}
}

// OnRunesFocused creates a focus-gated stop-propagation binding for all printable characters.
// Only fires when the owning component's element is focused.
func OnRunesFocused(handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{AnyRune: true, FocusRequired: true},
		Handler: handler,
		Stop:    true,
	}
}
