package tui

import (
	"testing"
)

func TestBorderStyle_Chars_Single(t *testing.T) {
	chars := BorderSingle.Chars()

	expected := BorderChars{
		TopLeft:     '┌',
		Top:         '─',
		TopRight:    '┐',
		Left:        '│',
		Right:       '│',
		BottomLeft:  '└',
		Bottom:      '─',
		BottomRight: '┘',
	}

	if chars != expected {
		t.Errorf("BorderSingle.Chars() = %+v, want %+v", chars, expected)
	}
}

func TestBorderStyle_Chars_Double(t *testing.T) {
	chars := BorderDouble.Chars()

	expected := BorderChars{
		TopLeft:     '╔',
		Top:         '═',
		TopRight:    '╗',
		Left:        '║',
		Right:       '║',
		BottomLeft:  '╚',
		Bottom:      '═',
		BottomRight: '╝',
	}

	if chars != expected {
		t.Errorf("BorderDouble.Chars() = %+v, want %+v", chars, expected)
	}
}

func TestBorderStyle_Chars_Rounded(t *testing.T) {
	chars := BorderRounded.Chars()

	expected := BorderChars{
		TopLeft:     '╭',
		Top:         '─',
		TopRight:    '╮',
		Left:        '│',
		Right:       '│',
		BottomLeft:  '╰',
		Bottom:      '─',
		BottomRight: '╯',
	}

	if chars != expected {
		t.Errorf("BorderRounded.Chars() = %+v, want %+v", chars, expected)
	}
}

func TestBorderStyle_Chars_Thick(t *testing.T) {
	chars := BorderThick.Chars()

	expected := BorderChars{
		TopLeft:     '┏',
		Top:         '━',
		TopRight:    '┓',
		Left:        '┃',
		Right:       '┃',
		BottomLeft:  '┗',
		Bottom:      '━',
		BottomRight: '┛',
	}

	if chars != expected {
		t.Errorf("BorderThick.Chars() = %+v, want %+v", chars, expected)
	}
}

func TestBorderStyle_Chars_None(t *testing.T) {
	chars := BorderNone.Chars()

	// All characters should be spaces
	expected := BorderChars{
		TopLeft:     ' ',
		Top:         ' ',
		TopRight:    ' ',
		Left:        ' ',
		Right:       ' ',
		BottomLeft:  ' ',
		Bottom:      ' ',
		BottomRight: ' ',
	}

	if chars != expected {
		t.Errorf("BorderNone.Chars() = %+v, want %+v", chars, expected)
	}
}

func TestDrawBox_SingleBorder(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle()

	DrawBox(buf, NewRect(1, 1, 5, 3), BorderSingle, style)

	// Check corners
	if buf.Cell(1, 1).Rune != '┌' {
		t.Errorf("TopLeft = %q, want '┌'", buf.Cell(1, 1).Rune)
	}
	if buf.Cell(5, 1).Rune != '┐' {
		t.Errorf("TopRight = %q, want '┐'", buf.Cell(5, 1).Rune)
	}
	if buf.Cell(1, 3).Rune != '└' {
		t.Errorf("BottomLeft = %q, want '└'", buf.Cell(1, 3).Rune)
	}
	if buf.Cell(5, 3).Rune != '┘' {
		t.Errorf("BottomRight = %q, want '┘'", buf.Cell(5, 3).Rune)
	}

	// Check top edge
	for x := 2; x <= 4; x++ {
		if buf.Cell(x, 1).Rune != '─' {
			t.Errorf("Top edge at %d = %q, want '─'", x, buf.Cell(x, 1).Rune)
		}
	}

	// Check bottom edge
	for x := 2; x <= 4; x++ {
		if buf.Cell(x, 3).Rune != '─' {
			t.Errorf("Bottom edge at %d = %q, want '─'", x, buf.Cell(x, 3).Rune)
		}
	}

	// Check left edge
	if buf.Cell(1, 2).Rune != '│' {
		t.Errorf("Left edge = %q, want '│'", buf.Cell(1, 2).Rune)
	}

	// Check right edge
	if buf.Cell(5, 2).Rune != '│' {
		t.Errorf("Right edge = %q, want '│'", buf.Cell(5, 2).Rune)
	}

	// Check interior is untouched (still spaces)
	for x := 2; x <= 4; x++ {
		if buf.Cell(x, 2).Rune != ' ' {
			t.Errorf("Interior at (%d, 2) = %q, want ' '", x, buf.Cell(x, 2).Rune)
		}
	}
}

func TestDrawBox_MinimalSize(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle()

	// Minimal 2x2 box
	DrawBox(buf, NewRect(1, 1, 2, 2), BorderSingle, style)

	// Should draw just corners
	if buf.Cell(1, 1).Rune != '┌' {
		t.Errorf("TopLeft = %q, want '┌'", buf.Cell(1, 1).Rune)
	}
	if buf.Cell(2, 1).Rune != '┐' {
		t.Errorf("TopRight = %q, want '┐'", buf.Cell(2, 1).Rune)
	}
	if buf.Cell(1, 2).Rune != '└' {
		t.Errorf("BottomLeft = %q, want '└'", buf.Cell(1, 2).Rune)
	}
	if buf.Cell(2, 2).Rune != '┘' {
		t.Errorf("BottomRight = %q, want '┘'", buf.Cell(2, 2).Rune)
	}
}

func TestDrawBox_TooSmall(t *testing.T) {
	type tc struct {
		width  int
		height int
	}

	tests := map[string]tc{
		"1x1": {width: 1, height: 1},
		"1x5": {width: 1, height: 5},
		"5x1": {width: 5, height: 1},
		"0x0": {width: 0, height: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf := NewBuffer(10, 10)
			style := NewStyle()

			// Mark a cell to verify nothing was drawn
			buf.SetRune(0, 0, 'X', style)

			DrawBox(buf, NewRect(0, 0, tt.width, tt.height), BorderSingle, style)

			// The 'X' should still be there (nothing drawn)
			if buf.Cell(0, 0).Rune != 'X' {
				t.Error("DrawBox should do nothing for rect smaller than 2x2")
			}
		})
	}
}

func TestDrawBox_BorderNone(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle()

	// Mark a cell to verify nothing was drawn
	buf.SetRune(1, 1, 'X', style)

	DrawBox(buf, NewRect(1, 1, 5, 3), BorderNone, style)

	// The 'X' should still be there
	if buf.Cell(1, 1).Rune != 'X' {
		t.Error("DrawBox with BorderNone should do nothing")
	}
}

func TestDrawBox_WithStyle(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle().Foreground(Red).Bold()

	DrawBox(buf, NewRect(1, 1, 5, 3), BorderSingle, style)

	// Check that corners have the style
	cell := buf.Cell(1, 1)
	if !cell.Style.HasAttr(AttrBold) {
		t.Error("Border should have bold style")
	}
	if !cell.Style.Fg.Equal(Red) {
		t.Error("Border should have red foreground")
	}
}

func TestDrawBox_ClipsToBuffer(t *testing.T) {
	buf := NewBuffer(10, 10)
	style := NewStyle()

	// Draw a box that extends beyond buffer on the right and bottom
	// Box from (5,5) with size 10x10 will clip to buffer bounds (10x10)
	// So visible portion is from (5,5) to (9,9) = 5x5 visible
	DrawBox(buf, NewRect(5, 5, 10, 10), BorderSingle, style)

	// Only the visible portion should be drawn
	// Top-left corner of the box at (5,5)
	if buf.Cell(5, 5).Rune != '┌' {
		t.Errorf("Visible top-left corner should be drawn, got %q", buf.Cell(5, 5).Rune)
	}
	// Top edge
	if buf.Cell(6, 5).Rune != '─' {
		t.Errorf("Visible top edge should be drawn, got %q", buf.Cell(6, 5).Rune)
	}
	// Left edge
	if buf.Cell(5, 6).Rune != '│' {
		t.Errorf("Visible left edge should be drawn, got %q", buf.Cell(5, 6).Rune)
	}
	// The right and bottom edges are clipped but should still draw what's visible
	// Right edge at x=9 (buffer boundary) - clipped to visible part
	if buf.Cell(9, 5).Rune != '─' && buf.Cell(9, 5).Rune != '┐' {
		t.Errorf("Right boundary should have border char, got %q", buf.Cell(9, 5).Rune)
	}
}

func TestDrawBoxWithTitle_Centered(t *testing.T) {
	buf := NewBuffer(20, 5)
	style := NewStyle()

	DrawBoxWithTitle(buf, NewRect(0, 0, 15, 3), BorderSingle, "Test", style)

	// Title "Test" should be centered in the top border
	// Available width = 15 - 2 = 13
	// Title width = 4
	// Start position = 1 + (13-4)/2 = 1 + 4 = 5

	// Check corners are still correct
	if buf.Cell(0, 0).Rune != '┌' {
		t.Errorf("TopLeft = %q, want '┌'", buf.Cell(0, 0).Rune)
	}
	if buf.Cell(14, 0).Rune != '┐' {
		t.Errorf("TopRight = %q, want '┐'", buf.Cell(14, 0).Rune)
	}

	// Check title is present
	title := "Test"
	startX := 1 + (13-4)/2 // = 5
	for i, r := range title {
		cell := buf.Cell(startX+i, 0)
		if cell.Rune != r {
			t.Errorf("Title at %d = %q, want %q", startX+i, cell.Rune, r)
		}
	}
}

func TestDrawBoxWithTitle_LongTitle(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle()

	// Title longer than available space
	DrawBoxWithTitle(buf, NewRect(0, 0, 6, 3), BorderSingle, "VeryLongTitle", style)

	// Available width = 6 - 2 = 4
	// Title should be truncated to fit
	// Check that corners are intact
	if buf.Cell(0, 0).Rune != '┌' {
		t.Errorf("TopLeft = %q, want '┌'", buf.Cell(0, 0).Rune)
	}
	if buf.Cell(5, 0).Rune != '┐' {
		t.Errorf("TopRight = %q, want '┐'", buf.Cell(5, 0).Rune)
	}

	// Check that some of the title is visible
	if buf.Cell(1, 0).Rune != 'V' {
		t.Errorf("First title char = %q, want 'V'", buf.Cell(1, 0).Rune)
	}
}

func TestDrawBoxWithTitle_EmptyTitle(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle()

	DrawBoxWithTitle(buf, NewRect(0, 0, 6, 3), BorderSingle, "", style)

	// Should just draw a normal box
	if buf.Cell(0, 0).Rune != '┌' {
		t.Errorf("TopLeft = %q, want '┌'", buf.Cell(0, 0).Rune)
	}

	// Top edge should be all horizontal lines (no title)
	for x := 1; x < 5; x++ {
		if buf.Cell(x, 0).Rune != '─' {
			t.Errorf("Top edge at %d = %q, want '─'", x, buf.Cell(x, 0).Rune)
		}
	}
}

func TestDrawBoxWithTitle_TooSmallForTitle(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle()

	// Box too small to fit any title
	DrawBoxWithTitle(buf, NewRect(0, 0, 2, 2), BorderSingle, "X", style)

	// Should still draw the box
	if buf.Cell(0, 0).Rune != '┌' {
		t.Errorf("TopLeft = %q, want '┌'", buf.Cell(0, 0).Rune)
	}
	if buf.Cell(1, 0).Rune != '┐' {
		t.Errorf("TopRight = %q, want '┐'", buf.Cell(1, 0).Rune)
	}
}

func TestDrawBoxWithTitle_WideCharTitle(t *testing.T) {
	buf := NewBuffer(20, 5)
	style := NewStyle()

	// Title with wide characters
	DrawBoxWithTitle(buf, NewRect(0, 0, 10, 3), BorderSingle, "你好", style)

	// "你好" takes 4 columns
	// Available width = 10 - 2 = 8
	// Start position = 1 + (8-4)/2 = 1 + 2 = 3

	// Check corners
	if buf.Cell(0, 0).Rune != '┌' {
		t.Errorf("TopLeft = %q, want '┌'", buf.Cell(0, 0).Rune)
	}

	// Check that the wide characters are present
	// They should be centered
	found := false
	for x := 1; x < 9; x++ {
		if buf.Cell(x, 0).Rune == '你' {
			found = true
			// Next cell should be continuation
			if !buf.Cell(x+1, 0).IsContinuation() {
				t.Error("Wide char should have continuation")
			}
			break
		}
	}
	if !found {
		t.Error("Wide character title not found")
	}
}

func TestFillBox(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle().Foreground(Blue)

	// Draw a box first
	DrawBox(buf, NewRect(1, 1, 6, 4), BorderSingle, style)

	// Fill the interior
	fillStyle := NewStyle().Background(Red)
	FillBox(buf, NewRect(1, 1, 6, 4), '.', fillStyle)

	// Check interior is filled
	for y := 2; y <= 3; y++ {
		for x := 2; x <= 5; x++ {
			cell := buf.Cell(x, y)
			if cell.Rune != '.' {
				t.Errorf("Interior at (%d, %d) = %q, want '.'", x, y, cell.Rune)
			}
		}
	}

	// Check border is unchanged
	if buf.Cell(1, 1).Rune != '┌' {
		t.Error("Border should be unchanged")
	}
}

func TestFillBox_TooSmall(t *testing.T) {
	buf := NewBuffer(10, 5)
	style := NewStyle()

	// Fill a box that's too small to have an interior
	buf.SetRune(0, 0, 'X', style)
	FillBox(buf, NewRect(0, 0, 2, 2), '.', style)

	// Should do nothing
	if buf.Cell(0, 0).Rune != 'X' {
		t.Error("FillBox should do nothing for box without interior")
	}
}
