package tui

import "unicode/utf8"

// parseInput parses buffered bytes into events.
// Handles:
// - Single printable characters -> KeyEvent{Key: KeyRune, Rune: r}
// - Control characters (0x00-0x1F) -> appropriate KeyEvent
// - CSI sequences (\x1b[...) -> Arrow keys, function keys with modifiers
// - SS3 sequences (\x1bO...) -> Some function keys
// - Alt+key: \x1b + printable -> KeyRune with ModAlt
func parseInput(data []byte) []Event {
	var events []Event
	i := 0

	for i < len(data) {
		b := data[i]

		// Check for escape sequence
		if b == 0x1b {
			// Look ahead to determine sequence type
			if i+1 >= len(data) {
				// Lone escape at end - treat as escape key
				events = append(events, KeyEvent{Key: KeyEscape})
				i++
				continue
			}

			next := data[i+1]
			switch next {
			case '[':
				// Check for SGR mouse sequence (ESC [ <)
				if i+2 < len(data) && data[i+2] == '<' {
					mouseEvent, consumed := parseMouseSGR(data[i:])
					if consumed > 0 {
						events = append(events, mouseEvent)
						i += consumed
						continue
					}
				}
				// CSI sequence
				key, mod, consumed := parseCSISequence(data[i:])
				if consumed > 0 {
					if key != KeyNone {
						events = append(events, KeyEvent{Key: key, Mod: mod})
					}
					i += consumed
					continue
				}
				// Failed to parse, treat as escape
				events = append(events, KeyEvent{Key: KeyEscape})
				i++
				continue

			case 'O':
				// SS3 sequence (function keys)
				if i+2 < len(data) {
					key := parseSS3(data[i+2])
					if key != KeyNone {
						events = append(events, KeyEvent{Key: key})
						i += 3
						continue
					}
				}
				// Failed to parse, treat as escape
				events = append(events, KeyEvent{Key: KeyEscape})
				i++
				continue

			default:
				// Alt+key combination
				if next >= 0x20 && next < 0x7f {
					events = append(events, KeyEvent{Key: KeyRune, Rune: rune(next), Mod: ModAlt})
					i += 2
					continue
				}
				// Unknown sequence, treat as escape
				events = append(events, KeyEvent{Key: KeyEscape})
				i++
				continue
			}
		}

		// Control characters (0x00-0x1F, except 0x1b which is handled above)
		if b < 0x20 {
			key := controlToKey(b)
			events = append(events, KeyEvent{Key: key})
			i++
			continue
		}

		// DEL character (0x7F) is backspace on most terminals
		if b == 0x7f {
			events = append(events, KeyEvent{Key: KeyBackspace})
			i++
			continue
		}

		// Printable characters (including multi-byte UTF-8)
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8, skip byte
			i++
			continue
		}
		events = append(events, KeyEvent{Key: KeyRune, Rune: r})
		i += size
	}

	return events
}

// controlToKey converts a control character (0x00-0x1F) to a Key.
func controlToKey(b byte) Key {
	switch b {
	case 0x00: // Ctrl+Space or Ctrl+@
		return KeyCtrlSpace
	case 0x01: // Ctrl+A
		return KeyCtrlA
	case 0x02: // Ctrl+B
		return KeyCtrlB
	case 0x03: // Ctrl+C
		return KeyCtrlC
	case 0x04: // Ctrl+D
		return KeyCtrlD
	case 0x05: // Ctrl+E
		return KeyCtrlE
	case 0x06: // Ctrl+F
		return KeyCtrlF
	case 0x07: // Ctrl+G (bell)
		return KeyCtrlG
	case 0x08: // Ctrl+H (backspace on some terminals)
		return KeyBackspace
	case 0x09: // Ctrl+I (tab)
		return KeyTab
	case 0x0a: // Ctrl+J (newline/enter on some terminals)
		return KeyCtrlJ
	case 0x0b: // Ctrl+K
		return KeyCtrlK
	case 0x0c: // Ctrl+L
		return KeyCtrlL
	case 0x0d: // Ctrl+M (carriage return/enter)
		return KeyEnter
	case 0x0e: // Ctrl+N
		return KeyCtrlN
	case 0x0f: // Ctrl+O
		return KeyCtrlO
	case 0x10: // Ctrl+P
		return KeyCtrlP
	case 0x11: // Ctrl+Q
		return KeyCtrlQ
	case 0x12: // Ctrl+R
		return KeyCtrlR
	case 0x13: // Ctrl+S
		return KeyCtrlS
	case 0x14: // Ctrl+T
		return KeyCtrlT
	case 0x15: // Ctrl+U
		return KeyCtrlU
	case 0x16: // Ctrl+V
		return KeyCtrlV
	case 0x17: // Ctrl+W
		return KeyCtrlW
	case 0x18: // Ctrl+X
		return KeyCtrlX
	case 0x19: // Ctrl+Y
		return KeyCtrlY
	case 0x1a: // Ctrl+Z
		return KeyCtrlZ
	case 0x1b: // Escape
		return KeyEscape
	default:
		return KeyNone
	}
}

// parseCSISequence parses a CSI escape sequence starting at data[0].
// Returns the key, modifier, and number of bytes consumed.
// Returns (KeyNone, ModNone, 0) if parsing fails.
func parseCSISequence(data []byte) (Key, Modifier, int) {
	if len(data) < 3 || data[0] != 0x1b || data[1] != '[' {
		return KeyNone, ModNone, 0
	}

	// Parse parameters (numbers separated by ;)
	var params []int
	currentParam := 0
	hasParam := false
	i := 2

	for i < len(data) {
		b := data[i]

		if b >= '0' && b <= '9' {
			currentParam = currentParam*10 + int(b-'0')
			hasParam = true
			i++
			continue
		}

		if b == ';' {
			params = append(params, currentParam)
			currentParam = 0
			hasParam = false
			i++
			continue
		}

		// Final byte (determines the key)
		if b >= 0x40 && b <= 0x7e {
			if hasParam {
				params = append(params, currentParam)
			}
			key, mod := parseCSI(params, b)
			return key, mod, i + 1
		}

		// Unexpected character
		return KeyNone, ModNone, 0
	}

	// Incomplete sequence
	return KeyNone, ModNone, 0
}

// parseCSI parses a complete CSI sequence given parameters and final byte.
// Returns (Key, Modifier).
func parseCSI(params []int, final byte) (Key, Modifier) {
	mod := ModNone

	// Extract modifier from params (xterm-style: CSI 1;mod X)
	if len(params) >= 2 {
		mod = decodeModifier(params[1])
	}

	switch final {
	case 'A':
		return KeyUp, mod
	case 'B':
		return KeyDown, mod
	case 'C':
		return KeyRight, mod
	case 'D':
		return KeyLeft, mod
	case 'H':
		return KeyHome, mod
	case 'F':
		return KeyEnd, mod
	case '~':
		// Extended keys: CSI n ~
		if len(params) == 0 {
			return KeyNone, ModNone
		}
		switch params[0] {
		case 1:
			return KeyHome, mod
		case 2:
			return KeyInsert, mod
		case 3:
			return KeyDelete, mod
		case 4:
			return KeyEnd, mod
		case 5:
			return KeyPageUp, mod
		case 6:
			return KeyPageDown, mod
		case 11:
			return KeyF1, mod
		case 12:
			return KeyF2, mod
		case 13:
			return KeyF3, mod
		case 14:
			return KeyF4, mod
		case 15:
			return KeyF5, mod
		case 17:
			return KeyF6, mod
		case 18:
			return KeyF7, mod
		case 19:
			return KeyF8, mod
		case 20:
			return KeyF9, mod
		case 21:
			return KeyF10, mod
		case 23:
			return KeyF11, mod
		case 24:
			return KeyF12, mod
		}
	case 'P':
		return KeyF1, mod
	case 'Q':
		return KeyF2, mod
	case 'R':
		return KeyF3, mod
	case 'S':
		return KeyF4, mod
	case 'Z':
		// Backtab (Shift+Tab) - CSI Z
		return KeyTab, ModShift
	}

	return KeyNone, ModNone
}

// parseSS3 parses an SS3 function key sequence.
// Returns the key constant for the given final byte.
func parseSS3(b byte) Key {
	switch b {
	case 'P':
		return KeyF1
	case 'Q':
		return KeyF2
	case 'R':
		return KeyF3
	case 'S':
		return KeyF4
	case 'A':
		return KeyUp
	case 'B':
		return KeyDown
	case 'C':
		return KeyRight
	case 'D':
		return KeyLeft
	case 'H':
		return KeyHome
	case 'F':
		return KeyEnd
	}
	return KeyNone
}

// decodeModifier decodes the xterm modifier parameter.
// The parameter is encoded as: 1 + (shift ? 1 : 0) + (alt ? 2 : 0) + (ctrl ? 4 : 0)
// So: 1=none, 2=shift, 3=alt, 4=shift+alt, 5=ctrl, 6=ctrl+shift, 7=ctrl+alt, 8=all
func decodeModifier(param int) Modifier {
	if param <= 1 {
		return ModNone
	}

	// Subtract 1 to get the raw flags
	flags := param - 1
	var mod Modifier
	if flags&1 != 0 {
		mod |= ModShift
	}
	if flags&2 != 0 {
		mod |= ModAlt
	}
	if flags&4 != 0 {
		mod |= ModCtrl
	}
	return mod
}

// parseMouseSGR parses an SGR-1006 mouse sequence.
// Format: ESC [ < button ; x ; y M (press) or ESC [ < button ; x ; y m (release)
// The button field encodes: button number + modifier bits
//
//	bits 0-1: button (0=left, 1=middle, 2=right, 3=release/none)
//	bit 2: shift
//	bit 3: meta/alt
//	bit 4: ctrl
//	bit 5: motion (drag)
//	bit 6: wheel (64=up, 65=down)
//
// Returns (MouseEvent, bytes consumed). Returns (MouseEvent{}, 0) on failure.
func parseMouseSGR(data []byte) (MouseEvent, int) {
	// Minimum: ESC [ < b ; x ; y M = 10 bytes for single digits
	if len(data) < 9 || data[0] != 0x1b || data[1] != '[' || data[2] != '<' {
		return MouseEvent{}, 0
	}

	// Parse: button ; x ; y
	i := 3
	button := 0
	x := 0
	y := 0
	stage := 0 // 0=button, 1=x, 2=y

	for i < len(data) {
		b := data[i]

		if b >= '0' && b <= '9' {
			switch stage {
			case 0:
				button = button*10 + int(b-'0')
			case 1:
				x = x*10 + int(b-'0')
			case 2:
				y = y*10 + int(b-'0')
			}
			i++
			continue
		}

		if b == ';' {
			stage++
			if stage > 2 {
				// Too many semicolons
				return MouseEvent{}, 0
			}
			i++
			continue
		}

		// Final byte: 'M' for press, 'm' for release
		if b == 'M' || b == 'm' {
			if stage != 2 {
				// Didn't get all three parameters
				return MouseEvent{}, 0
			}

			event := MouseEvent{
				X: x - 1, // Convert from 1-indexed to 0-indexed
				Y: y - 1,
			}

			// Decode button and modifiers
			if button&4 != 0 {
				event.Mod |= ModShift
			}
			if button&8 != 0 {
				event.Mod |= ModAlt
			}
			if button&16 != 0 {
				event.Mod |= ModCtrl
			}

			// Check for wheel events (bit 6 set)
			if button&64 != 0 {
				if button&1 != 0 {
					event.Button = MouseWheelDown
				} else {
					event.Button = MouseWheelUp
				}
				event.Action = MousePress // Wheel events are instantaneous
			} else {
				// Regular button event
				buttonNum := button & 3
				switch buttonNum {
				case 0:
					event.Button = MouseLeft
				case 1:
					event.Button = MouseMiddle
				case 2:
					event.Button = MouseRight
				case 3:
					event.Button = MouseNone // Release (legacy encoding)
				}

				// Determine action from final byte and motion bit
				if button&32 != 0 {
					event.Action = MouseDrag
				} else if b == 'M' {
					event.Action = MousePress
				} else {
					event.Action = MouseRelease
				}
			}

			return event, i + 1
		}

		// Unexpected character
		return MouseEvent{}, 0
	}

	// Incomplete sequence
	return MouseEvent{}, 0
}
