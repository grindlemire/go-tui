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

// Ctrl+letter helpers. Each is a RuneSpec that matches the corresponding
// Ctrl+letter combination via the modifier API.
var (
	KeyCtrlA     = Rune('a').Ctrl()
	KeyCtrlB     = Rune('b').Ctrl()
	KeyCtrlC     = Rune('c').Ctrl()
	KeyCtrlD     = Rune('d').Ctrl()
	KeyCtrlE     = Rune('e').Ctrl()
	KeyCtrlF     = Rune('f').Ctrl()
	KeyCtrlG     = Rune('g').Ctrl()
	KeyCtrlH     = Rune('h').Ctrl()
	KeyCtrlI     = Rune('i').Ctrl()
	KeyCtrlJ     = Rune('j').Ctrl()
	KeyCtrlK     = Rune('k').Ctrl()
	KeyCtrlL     = Rune('l').Ctrl()
	KeyCtrlM     = Rune('m').Ctrl()
	KeyCtrlN     = Rune('n').Ctrl()
	KeyCtrlO     = Rune('o').Ctrl()
	KeyCtrlP     = Rune('p').Ctrl()
	KeyCtrlQ     = Rune('q').Ctrl()
	KeyCtrlR     = Rune('r').Ctrl()
	KeyCtrlS     = Rune('s').Ctrl()
	KeyCtrlT     = Rune('t').Ctrl()
	KeyCtrlU     = Rune('u').Ctrl()
	KeyCtrlV     = Rune('v').Ctrl()
	KeyCtrlW     = Rune('w').Ctrl()
	KeyCtrlX     = Rune('x').Ctrl()
	KeyCtrlY     = Rune('y').Ctrl()
	KeyCtrlZ     = Rune('z').Ctrl()
	KeyCtrlSpace = Rune(' ').Ctrl()
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

// KeyMatcher describes what key events to match.
// Sealed to this package via unexported method.
type KeyMatcher interface {
	keyPattern() KeyPattern
}

// KeySpec matches a special key with specific modifiers.
type KeySpec struct {
	key Key
	mod Modifier
}

func (s KeySpec) keyPattern() KeyPattern {
	if s.mod != 0 {
		return KeyPattern{Key: s.key, Mod: s.mod}
	}
	return KeyPattern{Key: s.key, ExcludeMods: ModCtrl | ModAlt | ModShift}
}

// Ctrl returns a KeySpec requiring the Ctrl modifier.
func (s KeySpec) Ctrl() KeySpec { s.mod |= ModCtrl; return s }

// Alt returns a KeySpec requiring the Alt modifier.
func (s KeySpec) Alt() KeySpec { s.mod |= ModAlt; return s }

// Shift returns a KeySpec requiring the Shift modifier.
func (s KeySpec) Shift() KeySpec { s.mod |= ModShift; return s }

// keyPattern makes Key satisfy KeyMatcher directly.
// Matches the bare key with no modifiers (excludes Ctrl/Alt/Shift).
func (k Key) keyPattern() KeyPattern {
	return KeyPattern{Key: k, ExcludeMods: ModCtrl | ModAlt | ModShift}
}

// Ctrl returns a KeySpec for this key with the Ctrl modifier.
func (k Key) Ctrl() KeySpec { return KeySpec{key: k, mod: ModCtrl} }

// Alt returns a KeySpec for this key with the Alt modifier.
func (k Key) Alt() KeySpec { return KeySpec{key: k, mod: ModAlt} }

// Shift returns a KeySpec for this key with the Shift modifier.
func (k Key) Shift() KeySpec { return KeySpec{key: k, mod: ModShift} }

// RuneSpec matches a specific printable character with optional modifiers.
type RuneSpec struct {
	r   rune
	mod Modifier
}

// Rune returns a RuneSpec that matches a specific printable character.
// r must be a non-zero printable rune; Rune(0) will never match any event.
// Without modifiers, allows Shift (character-forming) but excludes Ctrl and Alt.
//
// Rune('a').Ctrl() works in both legacy and Kitty keyboard modes. In legacy
// mode, Ctrl+letter bytes (0x01-0x1A) are normalized to {KeyRune, letter, ModCtrl}
// by the parser, producing the same event as Kitty mode's CSI u encoding.
// Modifier combinations beyond what the terminal can distinguish (e.g.
// Ctrl+Shift+letter in legacy mode, where both produce the same control byte)
// will match whichever event the terminal actually sends.
func Rune(r rune) RuneSpec {
	return RuneSpec{r: r}
}

func (s RuneSpec) keyPattern() KeyPattern {
	if s.mod != 0 {
		return KeyPattern{Rune: s.r, Mod: s.mod}
	}
	return KeyPattern{Rune: s.r, ExcludeMods: ModCtrl | ModAlt}
}

// Ctrl returns a RuneSpec requiring the Ctrl modifier.
func (s RuneSpec) Ctrl() RuneSpec { s.mod |= ModCtrl; return s }

// Alt returns a RuneSpec requiring the Alt modifier.
func (s RuneSpec) Alt() RuneSpec { s.mod |= ModAlt; return s }

// Shift returns a RuneSpec requiring the Shift modifier.
func (s RuneSpec) Shift() RuneSpec { s.mod |= ModShift; return s }

// anyRuneSpec matches any printable character.
type anyRuneSpec struct{}

func (anyRuneSpec) keyPattern() KeyPattern {
	return KeyPattern{AnyRune: true, ExcludeMods: ModCtrl | ModAlt}
}

// AnyRune matches any printable character.
// Allows Shift (character-forming) but excludes Ctrl and Alt.
var AnyRune KeyMatcher = anyRuneSpec{}

// anyKeySpec matches any key event (printable or special).
type anyKeySpec struct{}

func (anyKeySpec) keyPattern() KeyPattern {
	return KeyPattern{AnyKey: true}
}

// AnyKey matches any key event. Used by modal overlays to block all parent handlers.
var AnyKey KeyMatcher = anyKeySpec{}

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
