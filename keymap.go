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

// OnKey creates a broadcast binding for a specific key.
// Without modifiers, excludes Ctrl/Alt/Shift so only unmodified presses match.
// With modifiers (e.g. OnKey(KeyTab, handler, ModShift)), matches that exact combo.
func OnKey(key Key, handler func(KeyEvent), mods ...Modifier) KeyBinding {
	var mod Modifier
	for _, m := range mods {
		mod |= m
	}
	pattern := KeyPattern{Key: key}
	if mod != 0 {
		pattern.Mod = mod
	} else {
		pattern.ExcludeMods = ModCtrl | ModAlt | ModShift
	}
	return KeyBinding{Pattern: pattern, Handler: handler}
}

// OnKeyStop creates a stop-propagation binding for a specific key.
// No handlers registered after this one (in tree order) will fire.
func OnKeyStop(key Key, handler func(KeyEvent), mods ...Modifier) KeyBinding {
	b := OnKey(key, handler, mods...)
	b.Stop = true
	return b
}

// OnRune creates a broadcast binding for a specific printable character.
// Without modifiers, allows Shift (character-forming) but excludes Ctrl and Alt.
// With modifiers (e.g. OnRune('s', handler, ModCtrl)), matches that exact combo.
func OnRune(r rune, handler func(KeyEvent), mods ...Modifier) KeyBinding {
	var mod Modifier
	for _, m := range mods {
		mod |= m
	}
	pattern := KeyPattern{Rune: r}
	if mod != 0 {
		pattern.Mod = mod
	} else {
		pattern.ExcludeMods = ModCtrl | ModAlt
	}
	return KeyBinding{Pattern: pattern, Handler: handler}
}

// OnRuneStop creates a stop-propagation binding for a specific printable character.
func OnRuneStop(r rune, handler func(KeyEvent), mods ...Modifier) KeyBinding {
	b := OnRune(r, handler, mods...)
	b.Stop = true
	return b
}

// OnRunes creates a broadcast binding for all printable characters.
// Allows Shift (character-forming) but excludes Ctrl and Alt.
func OnRunes(handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{AnyRune: true, ExcludeMods: ModCtrl | ModAlt},
		Handler: handler,
		Stop:    false,
	}
}

// OnRunesStop creates a stop-propagation binding for all printable characters.
// Use this for text inputs that need exclusive access to character keys.
// Allows Shift (character-forming) but excludes Ctrl and Alt.
func OnRunesStop(handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{AnyRune: true, ExcludeMods: ModCtrl | ModAlt},
		Handler: handler,
		Stop:    true,
	}
}

// OnKeyFocused creates a focus-gated stop-propagation binding for a specific key.
// Only fires when the owning component's element is focused.
func OnKeyFocused(key Key, handler func(KeyEvent), mods ...Modifier) KeyBinding {
	var mod Modifier
	for _, m := range mods {
		mod |= m
	}
	pattern := KeyPattern{Key: key, FocusRequired: true}
	if mod != 0 {
		pattern.Mod = mod
	} else {
		pattern.ExcludeMods = ModCtrl | ModAlt | ModShift
	}
	return KeyBinding{Pattern: pattern, Handler: handler, Stop: true}
}

// OnRuneFocused creates a focus-gated stop-propagation binding for a specific rune.
// Only fires when the owning component's element is focused.
func OnRuneFocused(r rune, handler func(KeyEvent), mods ...Modifier) KeyBinding {
	var mod Modifier
	for _, m := range mods {
		mod |= m
	}
	pattern := KeyPattern{Rune: r, FocusRequired: true}
	if mod != 0 {
		pattern.Mod = mod
	} else {
		pattern.ExcludeMods = ModCtrl | ModAlt
	}
	return KeyBinding{Pattern: pattern, Handler: handler, Stop: true}
}

// OnRunesFocused creates a focus-gated stop-propagation binding for all printable characters.
// Only fires when the owning component's element is focused.
// Allows Shift (character-forming) but excludes Ctrl and Alt.
func OnRunesFocused(handler func(KeyEvent)) KeyBinding {
	return KeyBinding{
		Pattern: KeyPattern{AnyRune: true, ExcludeMods: ModCtrl | ModAlt, FocusRequired: true},
		Handler: handler,
		Stop:    true,
	}
}
