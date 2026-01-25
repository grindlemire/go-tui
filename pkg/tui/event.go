package tui

// Event is the base interface for all terminal events.
// Use type switch to handle specific event types.
type Event interface {
	// isEvent is a marker method to prevent external implementations.
	isEvent()
}

// KeyEvent represents a keyboard input event.
type KeyEvent struct {
	// Key is the key pressed. For printable characters, this is KeyRune.
	// For special keys (arrows, function keys), this is the specific constant.
	Key Key

	// Rune is the character for KeyRune events. Zero for special keys.
	Rune rune

	// Mod contains modifier flags (Ctrl, Alt, Shift).
	Mod Modifier
}

func (KeyEvent) isEvent() {}

// IsRune returns true if this is a printable character event.
func (e KeyEvent) IsRune() bool {
	return e.Key == KeyRune
}

// Is checks if the event matches a specific key with optional modifiers.
// Example: event.Is(KeyEnter) or event.Is(KeyRune, ModCtrl)
func (e KeyEvent) Is(key Key, mods ...Modifier) bool {
	if e.Key != key {
		return false
	}
	if len(mods) == 0 {
		return true
	}
	// Combine all provided modifiers and check if they all match
	var combined Modifier
	for _, m := range mods {
		combined |= m
	}
	return e.Mod == combined
}

// Char returns the rune if this is a KeyRune event, or 0 otherwise.
func (e KeyEvent) Char() rune {
	if e.Key == KeyRune {
		return e.Rune
	}
	return 0
}

// ResizeEvent is emitted when the terminal is resized.
type ResizeEvent struct {
	Width  int
	Height int
}

func (ResizeEvent) isEvent() {}
