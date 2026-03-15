package tui

import "strings"

// Key represents a keyboard key.
type Key uint16

const (
	// KeyNone represents no key (zero value).
	KeyNone Key = iota

	// KeyRune represents a printable character. Check KeyEvent.Rune for the character.
	KeyRune

	// Special keys
	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyDelete
	KeyInsert

	// Arrow keys
	KeyUp
	KeyDown
	KeyLeft
	KeyRight

	// Navigation keys
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown

	// Function keys
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// String returns a human-readable representation of the key.
func (k Key) String() string {
	switch k {
	case KeyNone:
		return "None"
	case KeyRune:
		return "Rune"
	case KeyEscape:
		return "Escape"
	case KeyEnter:
		return "Enter"
	case KeyTab:
		return "Tab"
	case KeyBackspace:
		return "Backspace"
	case KeyDelete:
		return "Delete"
	case KeyInsert:
		return "Insert"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	default:
		return "Unknown"
	}
}

// Modifier represents keyboard modifier flags.
type Modifier uint8

const (
	// ModNone represents no modifiers.
	ModNone Modifier = 0
	// ModCtrl represents the Ctrl modifier.
	ModCtrl Modifier = 1 << iota
	// ModAlt represents the Alt modifier.
	ModAlt
	// ModShift represents the Shift modifier.
	ModShift
)

// Has checks if the modifier set includes the given modifier.
func (m Modifier) Has(mod Modifier) bool {
	return m&mod != 0
}

// String returns a human-readable representation of the modifiers.
func (m Modifier) String() string {
	if m == ModNone {
		return "None"
	}

	var parts []string
	if m.Has(ModCtrl) {
		parts = append(parts, "Ctrl")
	}
	if m.Has(ModAlt) {
		parts = append(parts, "Alt")
	}
	if m.Has(ModShift) {
		parts = append(parts, "Shift")
	}
	return strings.Join(parts, "+")
}
