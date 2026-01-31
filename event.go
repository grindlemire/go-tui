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

// MouseButton represents which mouse button was involved in an event.
type MouseButton int

const (
	// MouseLeft is the left (primary) mouse button.
	MouseLeft MouseButton = iota
	// MouseMiddle is the middle mouse button (scroll wheel click).
	MouseMiddle
	// MouseRight is the right (secondary) mouse button.
	MouseRight
	// MouseWheelUp is a scroll wheel up event.
	MouseWheelUp
	// MouseWheelDown is a scroll wheel down event.
	MouseWheelDown
	// MouseNone indicates no button (used for motion events).
	MouseNone
)

// MouseAction represents the type of mouse action.
type MouseAction int

const (
	// MousePress indicates a button was pressed.
	MousePress MouseAction = iota
	// MouseRelease indicates a button was released.
	MouseRelease
	// MouseDrag indicates motion while a button is held.
	MouseDrag
)

// MouseEvent represents a mouse input event.
type MouseEvent struct {
	// Button is which mouse button was involved.
	Button MouseButton
	// Action is the type of mouse action (press, release, drag).
	Action MouseAction
	// X is the column position (0-indexed).
	X int
	// Y is the row position (0-indexed).
	Y int
	// Mod contains modifier flags (Ctrl, Alt, Shift).
	Mod Modifier
}

func (MouseEvent) isEvent() {}
