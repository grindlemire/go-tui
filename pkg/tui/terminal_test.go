package tui

import (
	"testing"
)

func TestMockTerminal_ImplementsInterface(t *testing.T) {
	// Compile-time check that MockTerminal implements Terminal
	var _ Terminal = (*MockTerminal)(nil)
}

func TestMockTerminal_Size(t *testing.T) {
	type tc struct {
		width, height int
	}

	tests := map[string]tc{
		"standard 80x24": {
			width:  80,
			height: 24,
		},
		"large terminal": {
			width:  200,
			height: 60,
		},
		"small terminal": {
			width:  40,
			height: 10,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewMockTerminal(tt.width, tt.height)
			w, h := m.Size()
			if w != tt.width || h != tt.height {
				t.Errorf("Size() = (%d, %d), want (%d, %d)", w, h, tt.width, tt.height)
			}
		})
	}
}

func TestMockTerminal_FlushUpdatesCorrectCells(t *testing.T) {
	type tc struct {
		changes  []CellChange
		checkX   int
		checkY   int
		expected rune
	}

	tests := map[string]tc{
		"single cell at origin": {
			changes: []CellChange{
				{X: 0, Y: 0, Cell: NewCell('A', NewStyle())},
			},
			checkX:   0,
			checkY:   0,
			expected: 'A',
		},
		"cell at arbitrary position": {
			changes: []CellChange{
				{X: 5, Y: 3, Cell: NewCell('X', NewStyle())},
			},
			checkX:   5,
			checkY:   3,
			expected: 'X',
		},
		"multiple cells": {
			changes: []CellChange{
				{X: 0, Y: 0, Cell: NewCell('H', NewStyle())},
				{X: 1, Y: 0, Cell: NewCell('i', NewStyle())},
			},
			checkX:   1,
			checkY:   0,
			expected: 'i',
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewMockTerminal(80, 24)
			m.Flush(tt.changes)

			cell := m.CellAt(tt.checkX, tt.checkY)
			if cell.Rune != tt.expected {
				t.Errorf("CellAt(%d, %d).Rune = %q, want %q", tt.checkX, tt.checkY, cell.Rune, tt.expected)
			}
		})
	}
}

func TestMockTerminal_FlushIgnoresOutOfBounds(t *testing.T) {
	m := NewMockTerminal(10, 10)

	changes := []CellChange{
		{X: -1, Y: 0, Cell: NewCell('X', NewStyle())},
		{X: 10, Y: 0, Cell: NewCell('X', NewStyle())},
		{X: 0, Y: -1, Cell: NewCell('X', NewStyle())},
		{X: 0, Y: 10, Cell: NewCell('X', NewStyle())},
	}

	// Should not panic
	m.Flush(changes)

	// Border cells should still be spaces
	if m.CellAt(0, 0).Rune != ' ' {
		t.Error("Out-of-bounds flush affected (0,0)")
	}
	if m.CellAt(9, 0).Rune != ' ' {
		t.Error("Out-of-bounds flush affected (9,0)")
	}
	if m.CellAt(0, 9).Rune != ' ' {
		t.Error("Out-of-bounds flush affected (0,9)")
	}
}

func TestMockTerminal_CursorTracking(t *testing.T) {
	type tc struct {
		x, y int
	}

	tests := map[string]tc{
		"move to origin": {
			x: 0,
			y: 0,
		},
		"move to center": {
			x: 40,
			y: 12,
		},
		"move to corner": {
			x: 79,
			y: 23,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewMockTerminal(80, 24)
			m.SetCursor(tt.x, tt.y)

			x, y := m.Cursor()
			if x != tt.x || y != tt.y {
				t.Errorf("Cursor() = (%d, %d), want (%d, %d)", x, y, tt.x, tt.y)
			}
		})
	}
}

func TestMockTerminal_Clear(t *testing.T) {
	m := NewMockTerminal(10, 10)

	// Set some content
	m.Flush([]CellChange{
		{X: 0, Y: 0, Cell: NewCell('A', NewStyle())},
		{X: 5, Y: 5, Cell: NewCell('B', NewStyle())},
	})

	// Clear
	m.Clear()

	// Check that cells are cleared
	if m.CellAt(0, 0).Rune != ' ' {
		t.Errorf("Clear() did not reset cell (0,0)")
	}
	if m.CellAt(5, 5).Rune != ' ' {
		t.Errorf("Clear() did not reset cell (5,5)")
	}

	// Check that cursor was reset
	x, y := m.Cursor()
	if x != 0 || y != 0 {
		t.Errorf("Clear() did not reset cursor, got (%d, %d)", x, y)
	}
}

func TestMockTerminal_CursorVisibility(t *testing.T) {
	m := NewMockTerminal(80, 24)

	// Initially visible
	if m.IsCursorHidden() {
		t.Error("Cursor should be visible initially")
	}

	m.HideCursor()
	if !m.IsCursorHidden() {
		t.Error("HideCursor() should hide cursor")
	}

	m.ShowCursor()
	if m.IsCursorHidden() {
		t.Error("ShowCursor() should show cursor")
	}
}

func TestMockTerminal_RawModeTracking(t *testing.T) {
	m := NewMockTerminal(80, 24)

	// Initially not in raw mode
	if m.IsInRawMode() {
		t.Error("Should not be in raw mode initially")
	}

	err := m.EnterRawMode()
	if err != nil {
		t.Errorf("EnterRawMode() error = %v", err)
	}
	if !m.IsInRawMode() {
		t.Error("EnterRawMode() should enable raw mode")
	}

	err = m.ExitRawMode()
	if err != nil {
		t.Errorf("ExitRawMode() error = %v", err)
	}
	if m.IsInRawMode() {
		t.Error("ExitRawMode() should disable raw mode")
	}
}

func TestMockTerminal_AltScreenTracking(t *testing.T) {
	m := NewMockTerminal(80, 24)

	// Initially not in alt screen
	if m.IsInAltScreen() {
		t.Error("Should not be in alt screen initially")
	}

	m.EnterAltScreen()
	if !m.IsInAltScreen() {
		t.Error("EnterAltScreen() should enable alt screen")
	}

	m.ExitAltScreen()
	if m.IsInAltScreen() {
		t.Error("ExitAltScreen() should disable alt screen")
	}
}

func TestMockTerminal_Caps(t *testing.T) {
	m := NewMockTerminal(80, 24)

	caps := m.Caps()
	if caps.Colors != Color256 {
		t.Errorf("Caps().Colors = %v, want Color256", caps.Colors)
	}
	if !caps.Unicode {
		t.Error("Caps().Unicode should be true")
	}
	if !caps.TrueColor {
		t.Error("Caps().TrueColor should be true")
	}

	// Test SetCaps
	newCaps := Capabilities{
		Colors:    Color16,
		Unicode:   false,
		TrueColor: false,
		AltScreen: false,
	}
	m.SetCaps(newCaps)

	caps = m.Caps()
	if caps.Colors != Color16 {
		t.Errorf("After SetCaps, Colors = %v, want Color16", caps.Colors)
	}
	if caps.Unicode {
		t.Error("After SetCaps, Unicode should be false")
	}
}

func TestMockTerminal_String(t *testing.T) {
	m := NewMockTerminal(5, 3)

	// Set "Hi" at top-left
	m.Flush([]CellChange{
		{X: 0, Y: 0, Cell: NewCell('H', NewStyle())},
		{X: 1, Y: 0, Cell: NewCell('i', NewStyle())},
	})

	result := m.String()
	expected := "Hi   \n     \n     "
	if result != expected {
		t.Errorf("String() = %q, want %q", result, expected)
	}
}

func TestMockTerminal_StringTrimmed(t *testing.T) {
	m := NewMockTerminal(10, 3)

	m.Flush([]CellChange{
		{X: 0, Y: 0, Cell: NewCell('H', NewStyle())},
		{X: 1, Y: 0, Cell: NewCell('i', NewStyle())},
		{X: 2, Y: 1, Cell: NewCell('X', NewStyle())},
	})

	result := m.StringTrimmed()
	expected := "Hi\n  X\n"
	if result != expected {
		t.Errorf("StringTrimmed() = %q, want %q", result, expected)
	}
}

func TestMockTerminal_CellAt(t *testing.T) {
	m := NewMockTerminal(10, 10)

	// Set a styled cell
	style := NewStyle().Bold().Foreground(Red)
	m.Flush([]CellChange{
		{X: 3, Y: 4, Cell: NewCell('Z', style)},
	})

	cell := m.CellAt(3, 4)
	if cell.Rune != 'Z' {
		t.Errorf("CellAt(3,4).Rune = %q, want 'Z'", cell.Rune)
	}
	if !cell.Style.HasAttr(AttrBold) {
		t.Error("Cell should be bold")
	}
	if !cell.Style.Fg.Equal(Red) {
		t.Error("Cell foreground should be red")
	}
}

func TestMockTerminal_CellAtOutOfBounds(t *testing.T) {
	m := NewMockTerminal(10, 10)

	// Out of bounds should return empty cell
	cell := m.CellAt(-1, 0)
	if cell.Rune != 0 {
		t.Errorf("CellAt(-1,0).Rune = %q, want 0", cell.Rune)
	}

	cell = m.CellAt(10, 0)
	if cell.Rune != 0 {
		t.Errorf("CellAt(10,0).Rune = %q, want 0", cell.Rune)
	}
}

func TestMockTerminal_Reset(t *testing.T) {
	m := NewMockTerminal(10, 10)

	// Set various states
	m.Flush([]CellChange{{X: 0, Y: 0, Cell: NewCell('X', NewStyle())}})
	m.SetCursor(5, 5)
	m.HideCursor()
	m.EnterRawMode()
	m.EnterAltScreen()

	// Reset
	m.Reset()

	// Verify all states are reset
	if m.CellAt(0, 0).Rune != ' ' {
		t.Error("Reset() should clear cells")
	}
	x, y := m.Cursor()
	if x != 0 || y != 0 {
		t.Error("Reset() should reset cursor position")
	}
	if m.IsCursorHidden() {
		t.Error("Reset() should show cursor")
	}
	if m.IsInRawMode() {
		t.Error("Reset() should exit raw mode")
	}
	if m.IsInAltScreen() {
		t.Error("Reset() should exit alt screen")
	}
}

func TestMockTerminal_Resize(t *testing.T) {
	type tc struct {
		initialW, initialH int
		newW, newH         int
		setCell            struct{ x, y int }
		checkAfter         struct {
			x, y          int
			expectedRune  rune
			shouldPreserve bool
		}
	}

	tests := map[string]tc{
		"grow preserves content": {
			initialW: 5, initialH: 5,
			newW: 10, newH: 10,
			setCell: struct{ x, y int }{2, 2},
			checkAfter: struct {
				x, y          int
				expectedRune  rune
				shouldPreserve bool
			}{2, 2, 'X', true},
		},
		"shrink preserves overlapping content": {
			initialW: 10, initialH: 10,
			newW: 5, newH: 5,
			setCell: struct{ x, y int }{2, 2},
			checkAfter: struct {
				x, y          int
				expectedRune  rune
				shouldPreserve bool
			}{2, 2, 'X', true},
		},
		"shrink loses outside content": {
			initialW: 10, initialH: 10,
			newW: 5, newH: 5,
			setCell: struct{ x, y int }{8, 8},
			checkAfter: struct {
				x, y          int
				expectedRune  rune
				shouldPreserve bool
			}{8, 8, 0, false}, // Out of bounds after resize
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewMockTerminal(tt.initialW, tt.initialH)

			// Set a cell
			m.Flush([]CellChange{
				{X: tt.setCell.x, Y: tt.setCell.y, Cell: NewCell('X', NewStyle())},
			})

			// Resize
			m.Resize(tt.newW, tt.newH)

			// Check dimensions
			w, h := m.Size()
			if w != tt.newW || h != tt.newH {
				t.Errorf("After Resize, Size() = (%d, %d), want (%d, %d)", w, h, tt.newW, tt.newH)
			}

			// Check content preservation
			cell := m.CellAt(tt.checkAfter.x, tt.checkAfter.y)
			if tt.checkAfter.shouldPreserve {
				if cell.Rune != tt.checkAfter.expectedRune {
					t.Errorf("CellAt(%d,%d).Rune = %q, want %q", tt.checkAfter.x, tt.checkAfter.y, cell.Rune, tt.checkAfter.expectedRune)
				}
			} else {
				// Out of bounds should return empty cell
				if cell.Rune != 0 && cell.Rune != ' ' {
					t.Errorf("CellAt(%d,%d) should be empty after shrink, got %q", tt.checkAfter.x, tt.checkAfter.y, cell.Rune)
				}
			}
		})
	}
}

func TestMockTerminal_FlushWithWideCharacters(t *testing.T) {
	m := NewMockTerminal(10, 3)

	// Flush a CJK character (width 2)
	changes := []CellChange{
		{X: 0, Y: 0, Cell: NewCellWithWidth('中', NewStyle(), 2)},
		{X: 1, Y: 0, Cell: NewCellWithWidth(0, NewStyle(), 0)}, // Continuation
	}
	m.Flush(changes)

	// Check primary cell
	cell := m.CellAt(0, 0)
	if cell.Rune != '中' {
		t.Errorf("CellAt(0,0).Rune = %q, want '中'", cell.Rune)
	}
	if cell.Width != 2 {
		t.Errorf("CellAt(0,0).Width = %d, want 2", cell.Width)
	}

	// Check continuation cell
	cont := m.CellAt(1, 0)
	if !cont.IsContinuation() {
		t.Error("CellAt(1,0) should be a continuation cell")
	}
}
