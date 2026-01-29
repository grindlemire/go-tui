package tui

import (
	"strconv"
	"unicode/utf8"
)

// escBuilder efficiently builds ANSI escape sequences.
// It uses a pre-allocated buffer to minimize allocations.
type escBuilder struct {
	buf []byte
}

// newEscBuilder creates a new escape sequence builder with the given initial capacity.
func newEscBuilder(capacity int) *escBuilder {
	return &escBuilder{
		buf: make([]byte, 0, capacity),
	}
}

// Reset clears the buffer for reuse.
func (e *escBuilder) Reset() {
	e.buf = e.buf[:0]
}

// Bytes returns the built escape sequence.
func (e *escBuilder) Bytes() []byte {
	return e.buf
}

// Len returns the current length of the buffer.
func (e *escBuilder) Len() int {
	return len(e.buf)
}

// writeCSI writes the Control Sequence Introducer (ESC [).
func (e *escBuilder) writeCSI() {
	e.buf = append(e.buf, '\x1b', '[')
}

// writeInt writes an integer to the buffer.
func (e *escBuilder) writeInt(n int) {
	e.buf = strconv.AppendInt(e.buf, int64(n), 10)
}

// MoveTo moves the cursor to the specified position.
// x and y are 0-indexed; ANSI sequences use 1-indexed positions.
func (e *escBuilder) MoveTo(x, y int) {
	e.writeCSI()
	e.writeInt(y + 1) // Convert to 1-indexed
	e.buf = append(e.buf, ';')
	e.writeInt(x + 1) // Convert to 1-indexed
	e.buf = append(e.buf, 'H')
}

// MoveUp moves the cursor up by n rows.
func (e *escBuilder) MoveUp(n int) {
	if n <= 0 {
		return
	}
	e.writeCSI()
	if n > 1 {
		e.writeInt(n)
	}
	e.buf = append(e.buf, 'A')
}

// MoveDown moves the cursor down by n rows.
func (e *escBuilder) MoveDown(n int) {
	if n <= 0 {
		return
	}
	e.writeCSI()
	if n > 1 {
		e.writeInt(n)
	}
	e.buf = append(e.buf, 'B')
}

// MoveRight moves the cursor right by n columns.
func (e *escBuilder) MoveRight(n int) {
	if n <= 0 {
		return
	}
	e.writeCSI()
	if n > 1 {
		e.writeInt(n)
	}
	e.buf = append(e.buf, 'C')
}

// MoveLeft moves the cursor left by n columns.
func (e *escBuilder) MoveLeft(n int) {
	if n <= 0 {
		return
	}
	e.writeCSI()
	if n > 1 {
		e.writeInt(n)
	}
	e.buf = append(e.buf, 'D')
}

// ClearScreen clears the entire screen.
func (e *escBuilder) ClearScreen() {
	e.writeCSI()
	e.buf = append(e.buf, '2', 'J')
}

// ClearScrollback clears the scrollback buffer (ESC[3J).
// This helps ensure a clean screen after terminal resize.
func (e *escBuilder) ClearScrollback() {
	e.writeCSI()
	e.buf = append(e.buf, '3', 'J')
}

// ClearToEndOfScreen clears from cursor to end of screen (ESC[J or ESC[0J).
func (e *escBuilder) ClearToEndOfScreen() {
	e.writeCSI()
	e.buf = append(e.buf, 'J')
}

// ClearLine clears the entire current line.
func (e *escBuilder) ClearLine() {
	e.writeCSI()
	e.buf = append(e.buf, '2', 'K')
}

// HideCursor makes the cursor invisible.
func (e *escBuilder) HideCursor() {
	e.writeCSI()
	e.buf = append(e.buf, '?', '2', '5', 'l')
}

// ShowCursor makes the cursor visible.
func (e *escBuilder) ShowCursor() {
	e.writeCSI()
	e.buf = append(e.buf, '?', '2', '5', 'h')
}

// EnterAltScreen switches to the alternate screen buffer.
func (e *escBuilder) EnterAltScreen() {
	e.writeCSI()
	e.buf = append(e.buf, '?', '1', '0', '4', '9', 'h')
}

// ExitAltScreen switches back to the main screen buffer.
func (e *escBuilder) ExitAltScreen() {
	e.writeCSI()
	e.buf = append(e.buf, '?', '1', '0', '4', '9', 'l')
}

// BeginSyncUpdate starts a synchronized update block.
// The terminal will buffer all output until EndSyncUpdate is called,
// then display it atomically. This prevents tearing during updates.
// Supported by: iTerm2, Terminal.app, kitty, alacritty, foot, etc.
// Terminals that don't support it will simply ignore this sequence.
func (e *escBuilder) BeginSyncUpdate() {
	e.writeCSI()
	e.buf = append(e.buf, '?', '2', '0', '2', '6', 'h')
}

// EndSyncUpdate ends a synchronized update block.
// The terminal will now display all buffered output atomically.
func (e *escBuilder) EndSyncUpdate() {
	e.writeCSI()
	e.buf = append(e.buf, '?', '2', '0', '2', '6', 'l')
}

// EnableMouse enables mouse reporting using SGR-1006 extended mode.
// This enables button events (press/release) with SGR encoding for better
// coordinate support (works beyond column 223).
// Supported by most modern terminals: iTerm2, Terminal.app, kitty, alacritty, etc.
func (e *escBuilder) EnableMouse() {
	// Enable X10 mouse button tracking (basic click events)
	e.writeCSI()
	e.buf = append(e.buf, '?', '1', '0', '0', '0', 'h')
	// Enable SGR extended mouse mode (better coordinate encoding)
	e.writeCSI()
	e.buf = append(e.buf, '?', '1', '0', '0', '6', 'h')
}

// DisableMouse disables mouse reporting.
func (e *escBuilder) DisableMouse() {
	// Disable SGR extended mouse mode
	e.writeCSI()
	e.buf = append(e.buf, '?', '1', '0', '0', '6', 'l')
	// Disable X10 mouse button tracking
	e.writeCSI()
	e.buf = append(e.buf, '?', '1', '0', '0', '0', 'l')
}

// ResetStyle resets all text attributes to default.
func (e *escBuilder) ResetStyle() {
	e.writeCSI()
	e.buf = append(e.buf, '0', 'm')
}

// SetStyle sets the text style based on the given Style and terminal capabilities.
// It generates minimal escape sequences by only setting non-default attributes.
func (e *escBuilder) SetStyle(s Style, caps Capabilities) {
	// Always start with a reset to ensure clean state
	e.writeCSI()
	e.buf = append(e.buf, '0')

	// Add attributes
	if s.HasAttr(AttrBold) {
		e.buf = append(e.buf, ';', '1')
	}
	if s.HasAttr(AttrDim) {
		e.buf = append(e.buf, ';', '2')
	}
	if s.HasAttr(AttrItalic) {
		e.buf = append(e.buf, ';', '3')
	}
	if s.HasAttr(AttrUnderline) {
		e.buf = append(e.buf, ';', '4')
	}
	if s.HasAttr(AttrBlink) {
		e.buf = append(e.buf, ';', '5')
	}
	if s.HasAttr(AttrReverse) {
		e.buf = append(e.buf, ';', '7')
	}
	if s.HasAttr(AttrStrikethrough) {
		e.buf = append(e.buf, ';', '9')
	}

	// Add foreground color
	e.appendColor(s.Fg, true, caps)

	// Add background color
	e.appendColor(s.Bg, false, caps)

	e.buf = append(e.buf, 'm')
}

// appendColor appends the appropriate escape sequence for a color.
// fg indicates whether this is a foreground (true) or background (false) color.
func (e *escBuilder) appendColor(c Color, fg bool, caps Capabilities) {
	if c.IsDefault() {
		return
	}

	// Determine color code base: 38 for foreground, 48 for background
	var base int
	if fg {
		base = 38
	} else {
		base = 48
	}

	switch c.Type() {
	case ColorANSI:
		idx := c.ANSI()
		if idx < 16 && caps.Colors >= Color16 {
			// Use basic color codes for colors 0-15
			// Foreground: 30-37 (normal), 90-97 (bright)
			// Background: 40-47 (normal), 100-107 (bright)
			if idx < 8 {
				if fg {
					e.buf = append(e.buf, ';')
					e.writeInt(30 + int(idx))
				} else {
					e.buf = append(e.buf, ';')
					e.writeInt(40 + int(idx))
				}
			} else {
				if fg {
					e.buf = append(e.buf, ';')
					e.writeInt(90 + int(idx) - 8)
				} else {
					e.buf = append(e.buf, ';')
					e.writeInt(100 + int(idx) - 8)
				}
			}
		} else if caps.Colors >= Color256 {
			// Use 256-color mode: ESC[38;5;{n}m or ESC[48;5;{n}m
			e.buf = append(e.buf, ';')
			e.writeInt(base)
			e.buf = append(e.buf, ';', '5', ';')
			e.writeInt(int(idx))
		}

	case ColorRGB:
		if caps.TrueColor && caps.Colors >= ColorTrue {
			// Use true color: ESC[38;2;{r};{g};{b}m or ESC[48;2;{r};{g};{b}m
			r, g, b := c.RGB()
			e.buf = append(e.buf, ';')
			e.writeInt(base)
			e.buf = append(e.buf, ';', '2', ';')
			e.writeInt(int(r))
			e.buf = append(e.buf, ';')
			e.writeInt(int(g))
			e.buf = append(e.buf, ';')
			e.writeInt(int(b))
		} else if caps.Colors >= Color256 {
			// Fall back to ANSI approximation
			ansi := c.ToANSI()
			e.buf = append(e.buf, ';')
			e.writeInt(base)
			e.buf = append(e.buf, ';', '5', ';')
			e.writeInt(int(ansi.ANSI()))
		}
	}
}

// WriteRune appends a UTF-8 encoded rune to the buffer.
func (e *escBuilder) WriteRune(r rune) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	e.buf = append(e.buf, buf[:n]...)
}

// WriteString appends a string to the buffer.
func (e *escBuilder) WriteString(s string) {
	e.buf = append(e.buf, s...)
}

// WriteBytes appends bytes to the buffer.
func (e *escBuilder) WriteBytes(b []byte) {
	e.buf = append(e.buf, b...)
}
