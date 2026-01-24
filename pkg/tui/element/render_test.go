package element

import (
	"testing"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

func TestRenderTree_DrawsBackground(t *testing.T) {
	buf := tui.NewBuffer(20, 10)
	bgStyle := tui.NewStyle().Background(tui.Blue)
	e := New(
		WithSize(10, 5),
		WithBackground(bgStyle),
	)
	e.Calculate(20, 10)

	RenderTree(buf, e)

	// Check that background was filled
	// The entire 10x5 area should have the background style
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			cell := buf.Cell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("cell(%d,%d).Rune = %q, want ' '", x, y, cell.Rune)
			}
			if cell.Style.Bg != tui.Blue {
				t.Errorf("cell(%d,%d).Style.Bg = %v, want Blue", x, y, cell.Style.Bg)
			}
		}
	}
}

func TestRenderTree_DrawsBorder(t *testing.T) {
	buf := tui.NewBuffer(20, 10)
	e := New(
		WithSize(10, 5),
		WithBorder(tui.BorderSingle),
		WithBorderStyle(tui.NewStyle().Foreground(tui.Red)),
	)
	e.Calculate(20, 10)

	RenderTree(buf, e)

	// Check corners
	topLeft := buf.Cell(0, 0)
	if topLeft.Rune != '┌' {
		t.Errorf("top-left corner = %q, want '┌'", topLeft.Rune)
	}
	if topLeft.Style.Fg != tui.Red {
		t.Errorf("top-left color = %v, want Red", topLeft.Style.Fg)
	}

	topRight := buf.Cell(9, 0)
	if topRight.Rune != '┐' {
		t.Errorf("top-right corner = %q, want '┐'", topRight.Rune)
	}

	bottomLeft := buf.Cell(0, 4)
	if bottomLeft.Rune != '└' {
		t.Errorf("bottom-left corner = %q, want '└'", bottomLeft.Rune)
	}

	bottomRight := buf.Cell(9, 4)
	if bottomRight.Rune != '┘' {
		t.Errorf("bottom-right corner = %q, want '┘'", bottomRight.Rune)
	}

	// Check horizontal edges
	for x := 1; x < 9; x++ {
		top := buf.Cell(x, 0)
		if top.Rune != '─' {
			t.Errorf("top edge at %d = %q, want '─'", x, top.Rune)
		}
		bottom := buf.Cell(x, 4)
		if bottom.Rune != '─' {
			t.Errorf("bottom edge at %d = %q, want '─'", x, bottom.Rune)
		}
	}

	// Check vertical edges
	for y := 1; y < 4; y++ {
		left := buf.Cell(0, y)
		if left.Rune != '│' {
			t.Errorf("left edge at %d = %q, want '│'", y, left.Rune)
		}
		right := buf.Cell(9, y)
		if right.Rune != '│' {
			t.Errorf("right edge at %d = %q, want '│'", y, right.Rune)
		}
	}
}

func TestRenderTree_NestedElements(t *testing.T) {
	buf := tui.NewBuffer(30, 20)

	parent := New(
		WithSize(20, 15),
		WithPadding(2),
		WithBackground(tui.NewStyle().Background(tui.Blue)),
	)

	child := New(
		WithSize(10, 5),
		WithBorder(tui.BorderSingle),
	)

	parent.AddChild(child)
	parent.Calculate(30, 20)

	RenderTree(buf, parent)

	// Parent background should cover 20x15
	// Child should be at (2, 2) with 10x5 size
	childRect := child.Rect()
	if childRect.X != 2 || childRect.Y != 2 {
		t.Errorf("child position = (%d,%d), want (2,2)", childRect.X, childRect.Y)
	}

	// Check child border corner exists
	topLeft := buf.Cell(2, 2)
	if topLeft.Rune != '┌' {
		t.Errorf("child top-left = %q, want '┌'", topLeft.Rune)
	}
}

func TestRenderTree_CullsElementsOutsideBuffer(t *testing.T) {
	buf := tui.NewBuffer(10, 10)

	// Element positioned outside buffer
	e := New(WithSize(5, 5))
	e.Calculate(100, 100)
	// Manually set position outside buffer
	e.layout.Rect = layout.NewRect(50, 50, 5, 5)

	// Should not panic and should not draw anything
	RenderTree(buf, e)

	// All cells should be spaces (untouched)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			cell := buf.Cell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("cell(%d,%d) should be space, got %q", x, y, cell.Rune)
			}
		}
	}
}

func TestRenderText_DrawsContent(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	text := NewText("Hello",
		WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
		WithElementOption(WithSize(20, 3)),
	)
	text.Calculate(30, 10)

	RenderText(buf, text)

	// Check that text was drawn at left-aligned position
	// Text should start at (0, 0) with content "Hello"
	checkString(t, buf, 0, 0, "Hello")

	// Check style
	cell := buf.Cell(0, 0)
	if cell.Style.Fg != tui.Green {
		t.Errorf("text style.Fg = %v, want Green", cell.Style.Fg)
	}
}

func TestRenderText_CenterAlignment(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	text := NewText("Hi",
		WithTextAlign(TextAlignCenter),
		WithElementOption(WithSize(10, 3)),
	)
	text.Calculate(30, 10)

	RenderText(buf, text)

	// "Hi" is 2 chars wide, in a 10-wide container
	// Center position: (10 - 2) / 2 = 4
	// Text should be at x=4
	cell := buf.Cell(4, 0)
	if cell.Rune != 'H' {
		t.Errorf("centered text at x=4 = %q, want 'H'", cell.Rune)
	}
	cell = buf.Cell(5, 0)
	if cell.Rune != 'i' {
		t.Errorf("centered text at x=5 = %q, want 'i'", cell.Rune)
	}
}

func TestRenderText_RightAlignment(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	text := NewText("Hi",
		WithTextAlign(TextAlignRight),
		WithElementOption(WithSize(10, 3)),
	)
	text.Calculate(30, 10)

	RenderText(buf, text)

	// "Hi" is 2 chars wide, in a 10-wide container
	// Right-aligned position: 10 - 2 = 8
	cell := buf.Cell(8, 0)
	if cell.Rune != 'H' {
		t.Errorf("right-aligned text at x=8 = %q, want 'H'", cell.Rune)
	}
	cell = buf.Cell(9, 0)
	if cell.Rune != 'i' {
		t.Errorf("right-aligned text at x=9 = %q, want 'i'", cell.Rune)
	}
}

func TestRenderText_WithBorderAndPadding(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	// Note: Border is a visual property that draws on the element's rect.
	// Padding creates space inside the border box for content.
	// If you want space between border and content, use padding >= 1.
	text := NewText("Test",
		WithElementOption(WithSize(20, 5)),
		WithElementOption(WithBorder(tui.BorderSingle)),
		WithElementOption(WithPadding(1)), // 1 cell padding on all sides
	)
	text.Calculate(30, 10)

	RenderText(buf, text)

	// Border should be drawn
	corner := buf.Cell(0, 0)
	if corner.Rune != '┌' {
		t.Errorf("border corner = %q, want '┌'", corner.Rune)
	}

	// Content rect is border box inset by padding
	// With padding of 1, content starts at (1, 1)
	contentRect := text.ContentRect()
	if contentRect.X != 1 || contentRect.Y != 1 {
		t.Errorf("content rect starts at (%d,%d), want (1,1)", contentRect.X, contentRect.Y)
	}

	// Text "Test" should be at content rect position (1, 1)
	// Note: The text will overlap with the border because padding=1
	// means content starts where the border is drawn.
	// For proper spacing, use padding >= 2 when using a border.
	checkString(t, buf, 1, 1, "Test")
}

func TestElement_Render_CalculatesIfDirty(t *testing.T) {
	buf := tui.NewBuffer(30, 20)

	e := New(
		WithSize(20, 10),
		WithBorder(tui.BorderSingle),
	)

	// Element starts dirty, Render should calculate
	e.Render(buf, 30, 20)

	if e.IsDirty() {
		t.Error("Render should clear dirty flag after calculating")
	}

	// Border should be drawn
	corner := buf.Cell(0, 0)
	if corner.Rune != '┌' {
		t.Errorf("border corner = %q, want '┌'", corner.Rune)
	}
}

func TestElement_Render_SkipsCalculateIfClean(t *testing.T) {
	buf := tui.NewBuffer(30, 20)

	e := New(WithSize(20, 10))
	// Manually calculate and clear dirty
	e.Calculate(30, 20)

	// Manually set a different position to verify Calculate isn't called
	originalX := e.layout.Rect.X
	e.layout.Rect.X = 100

	// Render should skip calculate since not dirty
	e.Render(buf, 30, 20)

	// Position should be unchanged (Calculate wasn't called)
	if e.layout.Rect.X != 100 {
		t.Errorf("Render called Calculate when element was clean")
	}

	_ = originalX // Unused intentionally
}

func TestStringWidth(t *testing.T) {
	type tc struct {
		input    string
		expected int
	}

	tests := map[string]tc{
		"ASCII":       {input: "Hello", expected: 5},
		"empty":       {input: "", expected: 0},
		"with spaces": {input: "Hi there", expected: 8},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := stringWidth(tt.input)
			if got != tt.expected {
				t.Errorf("stringWidth(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// checkString verifies that the given string appears at the specified position in the buffer.
func checkString(t *testing.T, buf *tui.Buffer, x, y int, expected string) {
	t.Helper()
	curX := x
	for _, r := range expected {
		cell := buf.Cell(curX, y)
		if cell.Rune != r {
			t.Errorf("buf.Cell(%d,%d).Rune = %q, want %q", curX, y, cell.Rune, r)
		}
		curX += tui.RuneWidth(r)
	}
}
