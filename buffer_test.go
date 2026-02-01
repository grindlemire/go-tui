package tui

import (
	"testing"
)

func TestNewBuffer(t *testing.T) {
	type tc struct {
		width  int
		height int
	}

	tests := map[string]tc{
		"standard size": {
			width:  80,
			height: 24,
		},
		"small size": {
			width:  10,
			height: 5,
		},
		"single cell": {
			width:  1,
			height: 1,
		},
		"zero width": {
			width:  0,
			height: 10,
		},
		"zero height": {
			width:  10,
			height: 0,
		},
		"negative dimensions": {
			width:  -5,
			height: -3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewBuffer(tt.width, tt.height)

			expectedWidth := tt.width
			if expectedWidth < 0 {
				expectedWidth = 0
			}
			expectedHeight := tt.height
			if expectedHeight < 0 {
				expectedHeight = 0
			}

			if b.Width() != expectedWidth {
				t.Errorf("Width() = %d, want %d", b.Width(), expectedWidth)
			}
			if b.Height() != expectedHeight {
				t.Errorf("Height() = %d, want %d", b.Height(), expectedHeight)
			}

			w, h := b.Size()
			if w != expectedWidth || h != expectedHeight {
				t.Errorf("Size() = (%d, %d), want (%d, %d)", w, h, expectedWidth, expectedHeight)
			}

			rect := b.Rect()
			if rect.X != 0 || rect.Y != 0 || rect.Width != expectedWidth || rect.Height != expectedHeight {
				t.Errorf("Rect() = %+v, want {0, 0, %d, %d}", rect, expectedWidth, expectedHeight)
			}
		})
	}
}

func TestBuffer_InitializedWithSpaces(t *testing.T) {
	b := NewBuffer(5, 3)
	defaultStyle := NewStyle()

	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			cell := b.Cell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("Cell(%d, %d).Rune = %q, want ' '", x, y, cell.Rune)
			}
			if !cell.Style.Equal(defaultStyle) {
				t.Errorf("Cell(%d, %d) has non-default style", x, y)
			}
			if cell.Width != 1 {
				t.Errorf("Cell(%d, %d).Width = %d, want 1", x, y, cell.Width)
			}
		}
	}
}

func TestBuffer_SetCell_GetCell(t *testing.T) {
	type tc struct {
		x, y     int
		cell     Cell
		expected Cell // what we expect to get back (empty if out of bounds)
	}

	b := NewBuffer(5, 3)
	style := NewStyle().Foreground(Red)

	tests := map[string]tc{
		"in bounds": {
			x:        2,
			y:        1,
			cell:     NewCell('A', style),
			expected: NewCell('A', style),
		},
		"top-left corner": {
			x:        0,
			y:        0,
			cell:     NewCell('B', style),
			expected: NewCell('B', style),
		},
		"bottom-right corner": {
			x:        4,
			y:        2,
			cell:     NewCell('C', style),
			expected: NewCell('C', style),
		},
		"negative x": {
			x:        -1,
			y:        1,
			cell:     NewCell('X', style),
			expected: Cell{}, // out of bounds returns empty
		},
		"negative y": {
			x:        1,
			y:        -1,
			cell:     NewCell('Y', style),
			expected: Cell{}, // out of bounds returns empty
		},
		"x out of bounds": {
			x:        5,
			y:        1,
			cell:     NewCell('Z', style),
			expected: Cell{}, // out of bounds returns empty
		},
		"y out of bounds": {
			x:        1,
			y:        3,
			cell:     NewCell('W', style),
			expected: Cell{}, // out of bounds returns empty
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Reset buffer for each test
			b = NewBuffer(5, 3)

			b.SetCell(tt.x, tt.y, tt.cell)
			got := b.Cell(tt.x, tt.y)

			if !got.Equal(tt.expected) {
				t.Errorf("Cell(%d, %d) = %+v, want %+v", tt.x, tt.y, got, tt.expected)
			}
		})
	}
}

func TestBuffer_Fill(t *testing.T) {
	b := NewBuffer(10, 5)
	style := NewStyle().Foreground(Green)

	rect := NewRect(2, 1, 4, 2)
	b.Fill(rect, '#', style)

	// Check filled area
	for y := 1; y <= 2; y++ {
		for x := 2; x <= 5; x++ {
			cell := b.Cell(x, y)
			if cell.Rune != '#' {
				t.Errorf("Cell(%d, %d).Rune = %q, want '#'", x, y, cell.Rune)
			}
			if !cell.Style.Equal(style) {
				t.Errorf("Cell(%d, %d) has wrong style", x, y)
			}
		}
	}

	// Check unfilled area (outside rect)
	if b.Cell(1, 1).Rune != ' ' {
		t.Error("Cell outside fill rect should be unchanged")
	}
	if b.Cell(6, 1).Rune != ' ' {
		t.Error("Cell outside fill rect should be unchanged")
	}
}

func TestBuffer_Fill_WideChar(t *testing.T) {
	b := NewBuffer(10, 3)
	style := NewStyle()

	// Fill with a wide character
	rect := NewRect(0, 0, 6, 1)
	b.Fill(rect, '好', style)

	// Should have 3 wide chars (each taking 2 columns)
	for i := 0; i < 3; i++ {
		x := i * 2
		if b.Cell(x, 0).Rune != '好' {
			t.Errorf("Cell(%d, 0).Rune = %q, want '好'", x, b.Cell(x, 0).Rune)
		}
		if !b.Cell(x+1, 0).IsContinuation() {
			t.Errorf("Cell(%d, 0) should be continuation", x+1)
		}
	}
}

func TestBuffer_Fill_ClipsToBuffer(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	// Fill rect that extends beyond buffer
	rect := NewRect(-1, -1, 10, 10)
	b.Fill(rect, 'X', style)

	// All cells should be filled
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			if b.Cell(x, y).Rune != 'X' {
				t.Errorf("Cell(%d, %d).Rune = %q, want 'X'", x, y, b.Cell(x, y).Rune)
			}
		}
	}
}

func TestBuffer_Clear(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle().Bold().Foreground(Red)

	// Fill with styled content
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			b.SetRune(x, y, 'X', style)
		}
	}

	// Clear
	b.Clear()

	// All cells should be space with default style
	defaultStyle := NewStyle()
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			cell := b.Cell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("Cell(%d, %d).Rune = %q, want ' '", x, y, cell.Rune)
			}
			if !cell.Style.Equal(defaultStyle) {
				t.Errorf("Cell(%d, %d) should have default style", x, y)
			}
		}
	}
}

func TestBuffer_ClearRect(t *testing.T) {
	b := NewBuffer(10, 5)
	style := NewStyle().Bold()

	// Fill entire buffer
	b.Fill(b.Rect(), 'X', style)

	// Clear a portion
	rect := NewRect(2, 1, 3, 2)
	b.ClearRect(rect)

	// Check cleared area
	defaultStyle := NewStyle()
	for y := 1; y <= 2; y++ {
		for x := 2; x <= 4; x++ {
			cell := b.Cell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("Cell(%d, %d).Rune = %q, want ' ' (cleared)", x, y, cell.Rune)
			}
			if !cell.Style.Equal(defaultStyle) {
				t.Errorf("Cell(%d, %d) should have default style", x, y)
			}
		}
	}

	// Check non-cleared area
	if b.Cell(1, 1).Rune != 'X' {
		t.Error("Cell outside clear rect should be unchanged")
	}
	if b.Cell(5, 1).Rune != 'X' {
		t.Error("Cell outside clear rect should be unchanged")
	}
}

func TestBuffer_ClearRect_ClearsWideCharEdges(t *testing.T) {
	b := NewBuffer(10, 3)
	style := NewStyle()

	// Place wide chars at positions 1-2 and 4-5
	b.SetRune(1, 0, '好', style)
	b.SetRune(4, 0, '你', style)

	// Clear rect starting at continuation (2) and ending at wide char start (4)
	rect := NewRect(2, 0, 3, 1) // clears columns 2, 3, 4
	b.ClearRect(rect)

	// Position 1 should be cleared (was start of wide char, continuation was in clear zone)
	if b.Cell(1, 0).Rune != ' ' {
		t.Errorf("Cell(1, 0).Rune = %q, want ' ' (wide char cleared)", b.Cell(1, 0).Rune)
	}

	// Position 5 should be cleared (was continuation, start was in clear zone)
	if b.Cell(5, 0).Rune != ' ' {
		t.Errorf("Cell(5, 0).Rune = %q, want ' ' (continuation cleared)", b.Cell(5, 0).Rune)
	}
}

func TestBuffer_Resize_Grow(t *testing.T) {
	b := NewBuffer(3, 2)
	style := NewStyle()

	// Set some content
	b.SetRune(0, 0, 'A', style)
	b.SetRune(2, 1, 'B', style)

	// Grow
	b.Resize(5, 4)

	if b.Width() != 5 || b.Height() != 4 {
		t.Errorf("Size = (%d, %d), want (5, 4)", b.Width(), b.Height())
	}

	// Original content should be preserved
	if b.Cell(0, 0).Rune != 'A' {
		t.Errorf("Cell(0, 0).Rune = %q, want 'A'", b.Cell(0, 0).Rune)
	}
	if b.Cell(2, 1).Rune != 'B' {
		t.Errorf("Cell(2, 1).Rune = %q, want 'B'", b.Cell(2, 1).Rune)
	}

	// New area should be spaces
	if b.Cell(4, 3).Rune != ' ' {
		t.Errorf("Cell(4, 3).Rune = %q, want ' '", b.Cell(4, 3).Rune)
	}
}

func TestBuffer_Resize_Shrink(t *testing.T) {
	b := NewBuffer(5, 4)
	style := NewStyle()

	// Set content including outside new bounds
	b.SetRune(0, 0, 'A', style)
	b.SetRune(4, 3, 'Z', style)
	b.SetRune(2, 1, 'M', style)

	// Shrink
	b.Resize(3, 2)

	if b.Width() != 3 || b.Height() != 2 {
		t.Errorf("Size = (%d, %d), want (3, 2)", b.Width(), b.Height())
	}

	// Content within new bounds preserved
	if b.Cell(0, 0).Rune != 'A' {
		t.Errorf("Cell(0, 0).Rune = %q, want 'A'", b.Cell(0, 0).Rune)
	}
	if b.Cell(2, 1).Rune != 'M' {
		t.Errorf("Cell(2, 1).Rune = %q, want 'M'", b.Cell(2, 1).Rune)
	}

	// Old position (4, 3) is now out of bounds
	if b.Cell(4, 3).Rune != 0 {
		t.Error("Cell outside new bounds should return empty")
	}
}

func TestBuffer_Resize_SameSize(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	b.SetRune(2, 1, 'X', style)

	// Resize to same size - should be no-op
	b.Resize(5, 3)

	if b.Width() != 5 || b.Height() != 3 {
		t.Errorf("Size changed unexpectedly")
	}
	if b.Cell(2, 1).Rune != 'X' {
		t.Errorf("Content changed unexpectedly")
	}
}

func TestBuffer_Resize_PreservesFrontBuffer(t *testing.T) {
	b := NewBuffer(3, 2)
	style := NewStyle()

	// Make changes and swap
	b.SetRune(0, 0, 'A', style)
	b.Swap()

	// Make more changes (not swapped)
	b.SetRune(1, 0, 'B', style)

	// Resize
	b.Resize(4, 3)

	// Front buffer content should be preserved
	// After resize, Diff should still show 'B' as the only change
	b.SetRune(0, 0, 'A', style) // Reset to match front buffer

	changes := b.Diff()
	// Should have change for 'B' at (1,0) since it wasn't swapped
	found := false
	for _, c := range changes {
		if c.X == 1 && c.Y == 0 && c.Cell.Rune == 'B' {
			found = true
		}
	}
	if !found {
		t.Error("Resize didn't preserve pending changes")
	}
}

func TestBuffer_SetStringGradient(t *testing.T) {
	type tc struct {
		text     string
		gradient Gradient
		wantLen  int
	}

	tests := map[string]tc{
		"simple gradient": {
			text:     "Hello",
			gradient: NewGradient(Red, Blue),
			wantLen:  5,
		},
		"single char": {
			text:     "A",
			gradient: NewGradient(Red, Blue),
			wantLen:  1,
		},
		"empty string": {
			text:     "",
			gradient: NewGradient(Red, Blue),
			wantLen:  0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf := NewBuffer(20, 5)
			baseStyle := NewStyle()
			got := buf.SetStringGradient(0, 0, tt.text, tt.gradient, baseStyle)
			if got != tt.wantLen {
				t.Errorf("SetStringGradient() = %d, want %d", got, tt.wantLen)
			}
			// Verify first character has start color
			if len(tt.text) > 0 {
				cell := buf.Cell(0, 0)
				if cell.Rune != rune(tt.text[0]) {
					t.Errorf("First cell rune = %c, want %c", cell.Rune, rune(tt.text[0]))
				}
			}
		})
	}
}

func TestBuffer_FillGradient(t *testing.T) {
	type tc struct {
		rect     Rect
		gradient Gradient
	}

	tests := map[string]tc{
		"horizontal gradient": {
			rect:     NewRect(0, 0, 10, 5),
			gradient: NewGradient(Red, Blue).WithDirection(GradientHorizontal),
		},
		"vertical gradient": {
			rect:     NewRect(0, 0, 10, 5),
			gradient: NewGradient(Red, Blue).WithDirection(GradientVertical),
		},
		"diagonal down": {
			rect:     NewRect(0, 0, 10, 5),
			gradient: NewGradient(Red, Blue).WithDirection(GradientDiagonalDown),
		},
		"diagonal up": {
			rect:     NewRect(0, 0, 10, 5),
			gradient: NewGradient(Red, Blue).WithDirection(GradientDiagonalUp),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf := NewBuffer(20, 10)
			baseStyle := NewStyle()
			buf.FillGradient(tt.rect, ' ', tt.gradient, baseStyle)
			// Verify that cells have gradient colors applied
			cell := buf.Cell(tt.rect.X, tt.rect.Y)
			if cell.Style.Bg.IsDefault() {
				t.Error("FillGradient should set background color")
			}
		})
	}
}
