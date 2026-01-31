package tui

import (
	"testing"
)

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
