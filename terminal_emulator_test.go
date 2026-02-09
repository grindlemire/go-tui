package tui

import (
	"fmt"
	"strconv"
	"strings"
)

// EmulatorTerminal is a terminal emulator for testing that processes ANSI escape
// sequences and maintains both a visible screen and scrollback buffer.
// This allows tests to verify what SetInlineHeight, PrintAboveln, and other
// inline mode operations actually do to the terminal.
type EmulatorTerminal struct {
	width, height int
	screen        [][]rune // visible screen: screen[row][col]
	scrollback    []string // lines that scrolled off the top
	cursorRow     int      // 0-indexed
	cursorCol     int      // 0-indexed
	cursorHidden  bool
	inRawMode     bool
	inAltScreen   bool
	mouseEnabled  bool
	caps          Capabilities

	// Scroll region (0-indexed, inclusive). Defaults to full screen.
	scrollTop    int // top row of scroll region
	scrollBottom int // bottom row of scroll region
}

var _ Terminal = (*EmulatorTerminal)(nil)

// NewEmulatorTerminal creates a terminal emulator with the given dimensions.
// The screen is initialized with spaces.
func NewEmulatorTerminal(width, height int) *EmulatorTerminal {
	screen := make([][]rune, height)
	for i := range screen {
		screen[i] = make([]rune, width)
		for j := range screen[i] {
			screen[i][j] = ' '
		}
	}
	return &EmulatorTerminal{
		width:        width,
		height:       height,
		screen:       screen,
		scrollTop:    0,
		scrollBottom: height - 1,
		caps: Capabilities{
			Colors:    Color256,
			Unicode:   true,
			TrueColor: true,
			AltScreen: true,
		},
	}
}

func (e *EmulatorTerminal) Size() (int, int)    { return e.width, e.height }
func (e *EmulatorTerminal) Caps() Capabilities  { return e.caps }
func (e *EmulatorTerminal) EnterRawMode() error  { e.inRawMode = true; return nil }
func (e *EmulatorTerminal) ExitRawMode() error   { e.inRawMode = false; return nil }
func (e *EmulatorTerminal) EnterAltScreen()      { e.inAltScreen = true }
func (e *EmulatorTerminal) ExitAltScreen()       { e.inAltScreen = false }
func (e *EmulatorTerminal) EnableMouse()         { e.mouseEnabled = true }
func (e *EmulatorTerminal) DisableMouse()        { e.mouseEnabled = false }
func (e *EmulatorTerminal) HideCursor()          { e.cursorHidden = true }
func (e *EmulatorTerminal) ShowCursor()          { e.cursorHidden = false }
func (e *EmulatorTerminal) SetCursor(x, y int)   { e.cursorCol = x; e.cursorRow = y }

func (e *EmulatorTerminal) Clear() {
	for r := 0; r < e.height; r++ {
		for c := 0; c < e.width; c++ {
			e.screen[r][c] = ' '
		}
	}
	e.cursorRow = 0
	e.cursorCol = 0
}

func (e *EmulatorTerminal) ClearToEnd() {
	// Clear from cursor to end of screen
	for c := e.cursorCol; c < e.width; c++ {
		e.screen[e.cursorRow][c] = ' '
	}
	for r := e.cursorRow + 1; r < e.height; r++ {
		for c := 0; c < e.width; c++ {
			e.screen[r][c] = ' '
		}
	}
}

func (e *EmulatorTerminal) Flush(changes []CellChange) {
	for _, ch := range changes {
		if ch.X >= 0 && ch.X < e.width && ch.Y >= 0 && ch.Y < e.height {
			r := ch.Cell.Rune
			if r == 0 {
				r = ' '
			}
			e.screen[ch.Y][ch.X] = r
		}
	}
}

// WriteDirect processes raw bytes containing ANSI escape sequences.
// This is the key method that makes EmulatorTerminal useful for testing —
// it actually interprets the escape sequences instead of discarding them.
func (e *EmulatorTerminal) WriteDirect(b []byte) (int, error) {
	s := string(b)
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			// ESC sequence
			if i+1 < len(s) {
				switch s[i+1] {
				case '[':
					// CSI sequence
					consumed := e.parseCSI(s[i+2:])
					i += 2 + consumed
				case 'M':
					// Reverse Index
					e.reverseIndex()
					i += 2
				default:
					i += 2
				}
			} else {
				i++
			}
		} else if s[i] == '\n' {
			e.linefeed()
			i++
		} else if s[i] == '\r' {
			e.cursorCol = 0
			i++
		} else {
			// Printable character
			if e.cursorRow >= 0 && e.cursorRow < e.height &&
				e.cursorCol >= 0 && e.cursorCol < e.width {
				e.screen[e.cursorRow][e.cursorCol] = rune(s[i])
				e.cursorCol++
			} else if e.cursorCol >= e.width {
				// At end of line; real terminals wrap, but for our tests
				// we just stop advancing
			}
			i++
		}
	}
	return len(b), nil
}

// parseCSI parses a CSI (Control Sequence Introducer) sequence starting after "\033[".
// Returns the number of bytes consumed from s.
func (e *EmulatorTerminal) parseCSI(s string) int {
	// Collect parameters and final character
	i := 0
	for i < len(s) {
		ch := s[i]
		if ch >= 0x40 && ch <= 0x7E {
			// Final character found
			params := s[:i]
			switch ch {
			case 'H': // CUP - Cursor Position
				e.cursorPosition(params)
			case 'r': // DECSTBM - Set Scrolling Region
				e.setScrollRegion(params)
			case 'K': // EL - Erase in Line
				e.eraseLine(params)
			case 'M': // DL - Delete Lines
				e.deleteLines(params)
			case 'J': // ED - Erase in Display
				e.eraseDisplay(params)
			}
			return i + 1
		}
		i++
	}
	return i
}

// cursorPosition handles ESC[row;colH (1-indexed)
func (e *EmulatorTerminal) cursorPosition(params string) {
	row, col := 1, 1
	if params != "" {
		parts := strings.Split(params, ";")
		if len(parts) >= 1 && parts[0] != "" {
			row, _ = strconv.Atoi(parts[0])
		}
		if len(parts) >= 2 && parts[1] != "" {
			col, _ = strconv.Atoi(parts[1])
		}
	}
	e.cursorRow = row - 1 // convert to 0-indexed
	e.cursorCol = col - 1
}

// setScrollRegion handles ESC[top;bottomr (1-indexed)
// Real terminals move cursor to home position (0,0) on any DECSTBM.
func (e *EmulatorTerminal) setScrollRegion(params string) {
	if params == "" {
		// Reset to full screen
		e.scrollTop = 0
		e.scrollBottom = e.height - 1
		e.cursorRow = 0
		e.cursorCol = 0
		return
	}
	parts := strings.Split(params, ";")
	top := 1
	bottom := e.height
	if len(parts) >= 1 && parts[0] != "" {
		top, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 && parts[1] != "" {
		bottom, _ = strconv.Atoi(parts[1])
	}
	e.scrollTop = top - 1       // convert to 0-indexed
	e.scrollBottom = bottom - 1 // convert to 0-indexed
	e.cursorRow = 0
	e.cursorCol = 0
}

// eraseLine handles ESC[nK
func (e *EmulatorTerminal) eraseLine(params string) {
	n := 0
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	if e.cursorRow < 0 || e.cursorRow >= e.height {
		return
	}
	switch n {
	case 0: // clear from cursor to end of line
		for c := e.cursorCol; c < e.width; c++ {
			e.screen[e.cursorRow][c] = ' '
		}
	case 1: // clear from start of line to cursor
		for c := 0; c <= e.cursorCol && c < e.width; c++ {
			e.screen[e.cursorRow][c] = ' '
		}
	case 2: // clear entire line
		for c := 0; c < e.width; c++ {
			e.screen[e.cursorRow][c] = ' '
		}
	}
}

// deleteLines handles ESC[nM (Delete n lines at cursor, shift below up)
// Matches real terminal behavior: when the cursor is at the top of a scroll
// region that starts at row 0 (screen top), deleted lines are pushed to
// scrollback — the same as normal scrolling. When the scroll region starts
// below row 0, or the cursor is not at the top of the region, lines are discarded.
func (e *EmulatorTerminal) deleteLines(params string) {
	n := 1
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	for count := 0; count < n; count++ {
		if e.cursorRow <= e.scrollBottom {
			// Real terminal behavior: if scroll region starts at screen top
			// and cursor is at top of region, push deleted line to scrollback
			if e.scrollTop == 0 && e.cursorRow == 0 {
				line := strings.TrimRight(string(e.screen[0]), " ")
				e.scrollback = append(e.scrollback, line)
			}
			// Shift lines up
			for r := e.cursorRow; r < e.scrollBottom; r++ {
				copy(e.screen[r], e.screen[r+1])
			}
			// Blank the bottom line of scroll region
			for c := 0; c < e.width; c++ {
				e.screen[e.scrollBottom][c] = ' '
			}
		}
	}
}

// eraseDisplay handles ESC[nJ
func (e *EmulatorTerminal) eraseDisplay(params string) {
	n := 0
	if params != "" {
		n, _ = strconv.Atoi(params)
	}
	switch n {
	case 0: // clear from cursor to end of screen
		e.ClearToEnd()
	case 2: // clear entire screen
		e.Clear()
	}
}

// linefeed handles \n — if cursor is at the bottom of the scroll region,
// scrolls the region up (top line goes to scrollback). Otherwise moves cursor down.
func (e *EmulatorTerminal) linefeed() {
	if e.cursorRow == e.scrollBottom {
		// At bottom of scroll region: scroll up
		e.scrollRegionUp()
	} else if e.cursorRow < e.height-1 {
		e.cursorRow++
	}
}

// scrollRegionUp scrolls the scroll region up by one line.
// Matches real terminal behavior: the top line of the region goes to scrollback
// ONLY when the scroll region starts at the screen top (row 0). When the region
// starts below row 0, the top line is discarded.
func (e *EmulatorTerminal) scrollRegionUp() {
	// Real terminals only push to scrollback when scroll region starts at screen top
	line := strings.TrimRight(string(e.screen[e.scrollTop]), " ")
	if e.scrollTop == 0 {
		e.scrollback = append(e.scrollback, line)
	}

	// Shift lines up within the scroll region
	for r := e.scrollTop; r < e.scrollBottom; r++ {
		copy(e.screen[r], e.screen[r+1])
	}

	// Blank the bottom line of the region
	for c := 0; c < e.width; c++ {
		e.screen[e.scrollBottom][c] = ' '
	}
}

// reverseIndex handles ESC M — if cursor is at the top of the scroll region,
// scrolls the region down (bottom line falls off). Otherwise moves cursor up.
func (e *EmulatorTerminal) reverseIndex() {
	if e.cursorRow == e.scrollTop {
		// At top of scroll region: scroll down (insert blank at top, bottom falls off)
		for r := e.scrollBottom; r > e.scrollTop; r-- {
			copy(e.screen[r], e.screen[r-1])
		}
		// Blank the top line
		for c := 0; c < e.width; c++ {
			e.screen[e.scrollTop][c] = ' '
		}
	} else if e.cursorRow > 0 {
		e.cursorRow--
	}
}

// --- Test helper methods ---

// Scrollback returns all lines that have been scrolled into scrollback.
func (e *EmulatorTerminal) Scrollback() []string {
	return e.scrollback
}

// ScrollbackString returns the scrollback buffer as a single string.
func (e *EmulatorTerminal) ScrollbackString() string {
	return strings.Join(e.scrollback, "\n")
}

// NonBlankScrollback returns only non-empty lines from scrollback.
func (e *EmulatorTerminal) NonBlankScrollback() []string {
	var result []string
	for _, line := range e.scrollback {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// BlankScrollbackCount returns the number of blank lines in scrollback.
func (e *EmulatorTerminal) BlankScrollbackCount() int {
	count := 0
	for _, line := range e.scrollback {
		if line == "" {
			count++
		}
	}
	return count
}

// ScreenRow returns the content of a screen row as a trimmed string.
func (e *EmulatorTerminal) ScreenRow(row int) string {
	if row < 0 || row >= e.height {
		return ""
	}
	return strings.TrimRight(string(e.screen[row]), " ")
}

// ScreenString returns the entire visible screen as a string (rows joined by \n).
func (e *EmulatorTerminal) ScreenString() string {
	var lines []string
	for r := 0; r < e.height; r++ {
		lines = append(lines, strings.TrimRight(string(e.screen[r]), " "))
	}
	return strings.Join(lines, "\n")
}

// SetScreenRow sets the content of a screen row (for test setup).
func (e *EmulatorTerminal) SetScreenRow(row int, text string) {
	if row < 0 || row >= e.height {
		return
	}
	for c := 0; c < e.width; c++ {
		if c < len(text) {
			e.screen[row][c] = rune(text[c])
		} else {
			e.screen[row][c] = ' '
		}
	}
}

// DumpState returns a human-readable dump of the terminal state for debugging.
func (e *EmulatorTerminal) DumpState() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Terminal %dx%d, cursor=(%d,%d), scrollRegion=[%d,%d]\n",
		e.width, e.height, e.cursorRow, e.cursorCol, e.scrollTop, e.scrollBottom))
	sb.WriteString("--- Screen ---\n")
	for r := 0; r < e.height; r++ {
		sb.WriteString(fmt.Sprintf("  %2d: |%s|\n", r, string(e.screen[r])))
	}
	sb.WriteString(fmt.Sprintf("--- Scrollback (%d lines) ---\n", len(e.scrollback)))
	for i, line := range e.scrollback {
		if line == "" {
			sb.WriteString(fmt.Sprintf("  %2d: (blank)\n", i))
		} else {
			sb.WriteString(fmt.Sprintf("  %2d: |%s|\n", i, line))
		}
	}
	return sb.String()
}
