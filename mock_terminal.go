package tui

import (
	"strings"
)

// MockTerminal is a mock implementation of Terminal for testing.
// It captures all operations and maintains an internal buffer for verification.
type MockTerminal struct {
	width, height int
	cells         []Cell
	cursorX       int
	cursorY       int
	cursorHidden  bool
	inRawMode     bool
	inAltScreen   bool
	mouseEnabled  bool
	caps          Capabilities

	// Transition counters for testing screen mode switches
	altScreenEnterCount int
	altScreenExitCount  int
}

// Ensure MockTerminal implements Terminal.
var _ Terminal = (*MockTerminal)(nil)

// NewMockTerminal creates a new mock terminal with the given dimensions.
func NewMockTerminal(width, height int) *MockTerminal {
	size := width * height
	cells := make([]Cell, size)

	// Initialize with spaces
	defaultCell := NewCell(' ', NewStyle())
	for i := range cells {
		cells[i] = defaultCell
	}

	return &MockTerminal{
		width:  width,
		height: height,
		cells:  cells,
		caps: Capabilities{
			Colors:    Color256,
			Unicode:   true,
			TrueColor: true,
			AltScreen: true,
		},
	}
}

// Size returns the terminal dimensions.
func (m *MockTerminal) Size() (width, height int) {
	return m.width, m.height
}

// Flush applies the given cell changes to the mock terminal's buffer.
func (m *MockTerminal) Flush(changes []CellChange) {
	for _, ch := range changes {
		if ch.X >= 0 && ch.X < m.width && ch.Y >= 0 && ch.Y < m.height {
			idx := ch.Y*m.width + ch.X
			m.cells[idx] = ch.Cell
		}
	}
}

// Clear clears the entire terminal to spaces with default style.
func (m *MockTerminal) Clear() {
	defaultCell := NewCell(' ', NewStyle())
	for i := range m.cells {
		m.cells[i] = defaultCell
	}
	m.cursorX = 0
	m.cursorY = 0
}

// ClearToEnd clears from cursor position to end of screen.
func (m *MockTerminal) ClearToEnd() {
	defaultCell := NewCell(' ', NewStyle())
	// Start from current cursor position
	startIdx := m.cursorY*m.width + m.cursorX
	for i := startIdx; i < len(m.cells); i++ {
		m.cells[i] = defaultCell
	}
}

// SetCursor moves the cursor to the specified position.
func (m *MockTerminal) SetCursor(x, y int) {
	m.cursorX = x
	m.cursorY = y
}

// HideCursor makes the cursor invisible.
func (m *MockTerminal) HideCursor() {
	m.cursorHidden = true
}

// ShowCursor makes the cursor visible.
func (m *MockTerminal) ShowCursor() {
	m.cursorHidden = false
}

// EnterRawMode simulates entering raw mode.
func (m *MockTerminal) EnterRawMode() error {
	m.inRawMode = true
	return nil
}

// ExitRawMode simulates exiting raw mode.
func (m *MockTerminal) ExitRawMode() error {
	m.inRawMode = false
	return nil
}

// EnterAltScreen simulates entering the alternate screen buffer.
func (m *MockTerminal) EnterAltScreen() {
	m.inAltScreen = true
	m.altScreenEnterCount++
}

// ExitAltScreen simulates exiting the alternate screen buffer.
func (m *MockTerminal) ExitAltScreen() {
	m.inAltScreen = false
	m.altScreenExitCount++
}

// EnableMouse simulates enabling mouse event reporting.
func (m *MockTerminal) EnableMouse() {
	m.mouseEnabled = true
}

// DisableMouse simulates disabling mouse event reporting.
func (m *MockTerminal) DisableMouse() {
	m.mouseEnabled = false
}

// Caps returns the terminal's capabilities.
func (m *MockTerminal) Caps() Capabilities {
	return m.caps
}

// WriteDirect is a no-op for the mock terminal.
// In tests, raw escape sequences are not processed.
func (m *MockTerminal) WriteDirect(b []byte) (int, error) {
	return len(b), nil
}

// SetCaps sets the terminal's capabilities for testing.
func (m *MockTerminal) SetCaps(caps Capabilities) {
	m.caps = caps
}

// --- Test helper methods ---

// CellAt returns the cell at the given position.
// Returns an empty Cell if out of bounds.
func (m *MockTerminal) CellAt(x, y int) Cell {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return Cell{}
	}
	return m.cells[y*m.width+x]
}

// String renders the terminal buffer to a string for snapshot testing.
// Each row is separated by a newline.
func (m *MockTerminal) String() string {
	var sb strings.Builder
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			cell := m.cells[y*m.width+x]
			if cell.IsContinuation() {
				continue // Skip continuation cells
			}
			if cell.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune)
			}
		}
		if y < m.height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// StringTrimmed returns the terminal content with trailing spaces removed from each line.
func (m *MockTerminal) StringTrimmed() string {
	var sb strings.Builder
	for y := 0; y < m.height; y++ {
		var line strings.Builder
		for x := 0; x < m.width; x++ {
			cell := m.cells[y*m.width+x]
			if cell.IsContinuation() {
				continue
			}
			if cell.Rune == 0 {
				line.WriteRune(' ')
			} else {
				line.WriteRune(cell.Rune)
			}
		}
		sb.WriteString(strings.TrimRight(line.String(), " "))
		if y < m.height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// Cursor returns the current cursor position.
func (m *MockTerminal) Cursor() (x, y int) {
	return m.cursorX, m.cursorY
}

// IsCursorHidden returns whether the cursor is hidden.
func (m *MockTerminal) IsCursorHidden() bool {
	return m.cursorHidden
}

// IsInRawMode returns whether the terminal is in raw mode.
func (m *MockTerminal) IsInRawMode() bool {
	return m.inRawMode
}

// IsInAltScreen returns whether the terminal is using the alternate screen buffer.
func (m *MockTerminal) IsInAltScreen() bool {
	return m.inAltScreen
}

// AltScreenEnterCount returns the number of times EnterAltScreen was called.
func (m *MockTerminal) AltScreenEnterCount() int {
	return m.altScreenEnterCount
}

// AltScreenExitCount returns the number of times ExitAltScreen was called.
func (m *MockTerminal) AltScreenExitCount() int {
	return m.altScreenExitCount
}

// IsMouseEnabled returns whether mouse event reporting is enabled.
func (m *MockTerminal) IsMouseEnabled() bool {
	return m.mouseEnabled
}

// Reset resets the mock terminal to its initial state.
func (m *MockTerminal) Reset() {
	m.Clear()
	m.cursorHidden = false
	m.inRawMode = false
	m.inAltScreen = false
	m.mouseEnabled = false
	m.altScreenEnterCount = 0
	m.altScreenExitCount = 0
}

// Resize changes the terminal dimensions, preserving content where possible.
func (m *MockTerminal) Resize(width, height int) {
	newSize := width * height
	newCells := make([]Cell, newSize)

	defaultCell := NewCell(' ', NewStyle())
	for i := range newCells {
		newCells[i] = defaultCell
	}

	// Copy existing content
	copyWidth := min(width, m.width)
	copyHeight := min(height, m.height)

	for y := 0; y < copyHeight; y++ {
		for x := 0; x < copyWidth; x++ {
			newCells[y*width+x] = m.cells[y*m.width+x]
		}
	}

	m.width = width
	m.height = height
	m.cells = newCells
}

// SetCell directly sets a cell in the buffer (for test setup).
func (m *MockTerminal) SetCell(x, y int, c Cell) {
	if x >= 0 && x < m.width && y >= 0 && y < m.height {
		m.cells[y*m.width+x] = c
	}
}
