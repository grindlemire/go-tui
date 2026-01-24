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

func TestBuffer_SetRune_ASCII(t *testing.T) {
	b := NewBuffer(10, 5)
	style := NewStyle().Bold()

	b.SetRune(3, 2, 'A', style)

	cell := b.Cell(3, 2)
	if cell.Rune != 'A' {
		t.Errorf("Cell(3, 2).Rune = %q, want 'A'", cell.Rune)
	}
	if !cell.Style.Equal(style) {
		t.Error("Cell(3, 2) has wrong style")
	}
	if cell.Width != 1 {
		t.Errorf("Cell(3, 2).Width = %d, want 1", cell.Width)
	}

	// Neighboring cells should be unchanged (spaces)
	neighbors := []struct{ x, y int }{{2, 2}, {4, 2}, {3, 1}, {3, 3}}
	for _, n := range neighbors {
		c := b.Cell(n.x, n.y)
		if c.Rune != ' ' {
			t.Errorf("Cell(%d, %d).Rune = %q, want ' ' (unchanged)", n.x, n.y, c.Rune)
		}
	}
}

func TestBuffer_SetRune_WideChar(t *testing.T) {
	b := NewBuffer(10, 5)
	style := NewStyle().Foreground(Blue)

	b.SetRune(3, 2, '你', style)

	// Primary cell
	cell := b.Cell(3, 2)
	if cell.Rune != '你' {
		t.Errorf("Cell(3, 2).Rune = %q, want '你'", cell.Rune)
	}
	if cell.Width != 2 {
		t.Errorf("Cell(3, 2).Width = %d, want 2", cell.Width)
	}

	// Continuation cell
	cont := b.Cell(4, 2)
	if !cont.IsContinuation() {
		t.Error("Cell(4, 2) should be a continuation cell")
	}
	if cont.Rune != 0 {
		t.Errorf("Cell(4, 2).Rune = %q, want 0", cont.Rune)
	}
	if cont.Width != 0 {
		t.Errorf("Cell(4, 2).Width = %d, want 0", cont.Width)
	}
}

func TestBuffer_SetRune_OverwriteContinuation(t *testing.T) {
	b := NewBuffer(10, 5)
	style := NewStyle()

	// Place a wide character at position 2
	b.SetRune(2, 0, '好', style)

	// Verify initial state
	if b.Cell(2, 0).Rune != '好' {
		t.Fatal("Failed to set initial wide char")
	}
	if !b.Cell(3, 0).IsContinuation() {
		t.Fatal("Failed to set continuation cell")
	}

	// Now write an ASCII char at the continuation position (3)
	b.SetRune(3, 0, 'X', style)

	// The wide char should be cleared (replaced with space)
	if b.Cell(2, 0).Rune != ' ' {
		t.Errorf("Cell(2, 0).Rune = %q, want ' ' (cleared)", b.Cell(2, 0).Rune)
	}

	// Position 3 should now have 'X'
	if b.Cell(3, 0).Rune != 'X' {
		t.Errorf("Cell(3, 0).Rune = %q, want 'X'", b.Cell(3, 0).Rune)
	}
}

func TestBuffer_SetRune_OverwriteWideCharStart(t *testing.T) {
	b := NewBuffer(10, 5)
	style := NewStyle()

	// Place a wide character at position 2
	b.SetRune(2, 0, '好', style)

	// Now write an ASCII char at the start position (2)
	b.SetRune(2, 0, 'Y', style)

	// Position 2 should now have 'Y'
	if b.Cell(2, 0).Rune != 'Y' {
		t.Errorf("Cell(2, 0).Rune = %q, want 'Y'", b.Cell(2, 0).Rune)
	}

	// Position 3 should be cleared (the continuation was replaced)
	if b.Cell(3, 0).Rune != ' ' {
		t.Errorf("Cell(3, 0).Rune = %q, want ' ' (cleared)", b.Cell(3, 0).Rune)
	}
}

func TestBuffer_SetRune_WideCharOverlapExisting(t *testing.T) {
	b := NewBuffer(10, 5)
	style := NewStyle()

	// Place a wide character at position 3
	b.SetRune(3, 0, '中', style)

	// Now place another wide character at position 2
	// This should clear the wide char at position 3
	b.SetRune(2, 0, '文', style)

	// Position 2 should have '文'
	if b.Cell(2, 0).Rune != '文' {
		t.Errorf("Cell(2, 0).Rune = %q, want '文'", b.Cell(2, 0).Rune)
	}
	// Position 3 should be continuation of '文'
	if !b.Cell(3, 0).IsContinuation() {
		t.Error("Cell(3, 0) should be continuation")
	}
	// Position 4 should be cleared (was continuation of '中')
	if b.Cell(4, 0).Rune != ' ' {
		t.Errorf("Cell(4, 0).Rune = %q, want ' ' (cleared)", b.Cell(4, 0).Rune)
	}
}

func TestBuffer_SetRune_WideCharAtLastColumn(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	// Try to place a wide char at the last column
	b.SetRune(4, 0, '你', style)

	// Should place a space instead (wide char doesn't fit)
	cell := b.Cell(4, 0)
	if cell.Rune != ' ' {
		t.Errorf("Cell(4, 0).Rune = %q, want ' ' (wide char doesn't fit)", cell.Rune)
	}
}

func TestBuffer_SetString_ASCII(t *testing.T) {
	b := NewBuffer(20, 5)
	style := NewStyle().Bold()

	width := b.SetString(2, 1, "Hello", style)

	if width != 5 {
		t.Errorf("SetString returned width %d, want 5", width)
	}

	expected := "Hello"
	for i, r := range expected {
		cell := b.Cell(2+i, 1)
		if cell.Rune != r {
			t.Errorf("Cell(%d, 1).Rune = %q, want %q", 2+i, cell.Rune, r)
		}
		if !cell.Style.Equal(style) {
			t.Errorf("Cell(%d, 1) has wrong style", 2+i)
		}
	}
}

func TestBuffer_SetString_MixedWidths(t *testing.T) {
	b := NewBuffer(20, 5)
	style := NewStyle()

	// "Hi你好" = H(1) + i(1) + 你(2) + 好(2) = 6 columns
	width := b.SetString(0, 0, "Hi你好", style)

	if width != 6 {
		t.Errorf("SetString returned width %d, want 6", width)
	}

	// Check each position
	if b.Cell(0, 0).Rune != 'H' {
		t.Error("Position 0 should be 'H'")
	}
	if b.Cell(1, 0).Rune != 'i' {
		t.Error("Position 1 should be 'i'")
	}
	if b.Cell(2, 0).Rune != '你' {
		t.Error("Position 2 should be '你'")
	}
	if !b.Cell(3, 0).IsContinuation() {
		t.Error("Position 3 should be continuation")
	}
	if b.Cell(4, 0).Rune != '好' {
		t.Error("Position 4 should be '好'")
	}
	if !b.Cell(5, 0).IsContinuation() {
		t.Error("Position 5 should be continuation")
	}
}

func TestBuffer_SetString_Truncation(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	// Try to write "Hello World" in a 5-column buffer
	width := b.SetString(0, 0, "Hello World", style)

	if width != 5 {
		t.Errorf("SetString returned width %d, want 5 (truncated)", width)
	}

	// Only "Hello" should fit
	expected := "Hello"
	for i, r := range expected {
		if b.Cell(i, 0).Rune != r {
			t.Errorf("Cell(%d, 0).Rune = %q, want %q", i, b.Cell(i, 0).Rune, r)
		}
	}
}

func TestBuffer_SetString_WideCharTruncation(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	// "abc你" would need 5 columns (3 + 2), fits exactly
	width := b.SetString(0, 0, "abc你", style)
	if width != 5 {
		t.Errorf("SetString(\"abc你\") returned width %d, want 5", width)
	}

	// "abcd你" would need 6 columns - wide char shouldn't fit
	b = NewBuffer(5, 3)
	width = b.SetString(0, 0, "abcd你", style)
	if width != 4 {
		t.Errorf("SetString(\"abcd你\") returned width %d, want 4 (truncated)", width)
	}
}

func TestBuffer_SetString_NegativeStart(t *testing.T) {
	b := NewBuffer(10, 3)
	style := NewStyle()

	// Start before visible area - should skip leading chars
	width := b.SetString(-2, 0, "Hello", style)

	// Only "llo" should be visible (starting at x=0)
	if width != 3 {
		t.Errorf("SetString returned width %d, want 3", width)
	}
	if b.Cell(0, 0).Rune != 'l' {
		t.Errorf("Cell(0, 0).Rune = %q, want 'l'", b.Cell(0, 0).Rune)
	}
}

func TestBuffer_SetString_OutOfBoundsY(t *testing.T) {
	b := NewBuffer(10, 3)
	style := NewStyle()

	width := b.SetString(0, -1, "Test", style)
	if width != 0 {
		t.Errorf("SetString with y=-1 returned width %d, want 0", width)
	}

	width = b.SetString(0, 3, "Test", style)
	if width != 0 {
		t.Errorf("SetString with y=3 (out of bounds) returned width %d, want 0", width)
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

func TestBuffer_Diff_Empty(t *testing.T) {
	b := NewBuffer(5, 3)

	// No changes - diff should be empty
	changes := b.Diff()
	if len(changes) != 0 {
		t.Errorf("Diff() returned %d changes, want 0", len(changes))
	}
}

func TestBuffer_Diff_SingleChange(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	b.SetRune(2, 1, 'A', style)

	changes := b.Diff()
	if len(changes) != 1 {
		t.Fatalf("Diff() returned %d changes, want 1", len(changes))
	}

	if changes[0].X != 2 || changes[0].Y != 1 {
		t.Errorf("Change at (%d, %d), want (2, 1)", changes[0].X, changes[0].Y)
	}
	if changes[0].Cell.Rune != 'A' {
		t.Errorf("Change cell rune = %q, want 'A'", changes[0].Cell.Rune)
	}
}

func TestBuffer_Diff_MultipleChanges(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	b.SetRune(0, 0, 'A', style)
	b.SetRune(4, 0, 'B', style)
	b.SetRune(2, 2, 'C', style)

	changes := b.Diff()
	if len(changes) != 3 {
		t.Fatalf("Diff() returned %d changes, want 3", len(changes))
	}

	// Changes should be in row-major order
	expected := []struct{ x, y int; r rune }{
		{0, 0, 'A'},
		{4, 0, 'B'},
		{2, 2, 'C'},
	}

	for i, e := range expected {
		if changes[i].X != e.x || changes[i].Y != e.y {
			t.Errorf("Change %d at (%d, %d), want (%d, %d)", i, changes[i].X, changes[i].Y, e.x, e.y)
		}
		if changes[i].Cell.Rune != e.r {
			t.Errorf("Change %d rune = %q, want %q", i, changes[i].Cell.Rune, e.r)
		}
	}
}

func TestBuffer_Diff_RowMajorOrder(t *testing.T) {
	b := NewBuffer(3, 3)
	style := NewStyle()

	// Fill in non-row-major order
	b.SetRune(2, 2, 'I', style)
	b.SetRune(0, 0, 'A', style)
	b.SetRune(1, 1, 'E', style)

	changes := b.Diff()
	if len(changes) != 3 {
		t.Fatalf("Diff() returned %d changes, want 3", len(changes))
	}

	// Should come out in row-major order regardless of insertion order
	if changes[0].X != 0 || changes[0].Y != 0 {
		t.Errorf("First change at (%d, %d), want (0, 0)", changes[0].X, changes[0].Y)
	}
	if changes[1].X != 1 || changes[1].Y != 1 {
		t.Errorf("Second change at (%d, %d), want (1, 1)", changes[1].X, changes[1].Y)
	}
	if changes[2].X != 2 || changes[2].Y != 2 {
		t.Errorf("Third change at (%d, %d), want (2, 2)", changes[2].X, changes[2].Y)
	}
}

func TestBuffer_Swap(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	// Make a change
	b.SetRune(2, 1, 'X', style)

	// Diff should show the change
	changes1 := b.Diff()
	if len(changes1) != 1 {
		t.Fatal("Expected 1 change before swap")
	}

	// Swap
	b.Swap()

	// Diff should now be empty
	changes2 := b.Diff()
	if len(changes2) != 0 {
		t.Errorf("Diff() after Swap() returned %d changes, want 0", len(changes2))
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

func TestBuffer_WideChar_ChainedOverwrite(t *testing.T) {
	b := NewBuffer(10, 1)
	style := NewStyle()

	// Place wide chars in sequence
	b.SetRune(0, 0, '你', style) // occupies 0-1
	b.SetRune(2, 0, '好', style) // occupies 2-3
	b.SetRune(4, 0, '吗', style) // occupies 4-5

	// Verify initial state
	if b.Cell(0, 0).Rune != '你' {
		t.Error("Initial: position 0 should be '你'")
	}
	if b.Cell(2, 0).Rune != '好' {
		t.Error("Initial: position 2 should be '好'")
	}
	if b.Cell(4, 0).Rune != '吗' {
		t.Error("Initial: position 4 should be '吗'")
	}

	// Now overwrite middle with ASCII
	b.SetRune(2, 0, 'X', style)
	b.SetRune(3, 0, 'Y', style)

	// Position 2 should be X, position 3 should be Y
	if b.Cell(2, 0).Rune != 'X' {
		t.Errorf("Cell(2, 0).Rune = %q, want 'X'", b.Cell(2, 0).Rune)
	}
	if b.Cell(3, 0).Rune != 'Y' {
		t.Errorf("Cell(3, 0).Rune = %q, want 'Y'", b.Cell(3, 0).Rune)
	}

	// Surrounding wide chars should still be intact
	if b.Cell(0, 0).Rune != '你' {
		t.Errorf("Cell(0, 0).Rune = %q, want '你'", b.Cell(0, 0).Rune)
	}
	if b.Cell(4, 0).Rune != '吗' {
		t.Errorf("Cell(4, 0).Rune = %q, want '吗'", b.Cell(4, 0).Rune)
	}
}
