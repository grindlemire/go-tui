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

func TestRenderTree_DrawsTextContent(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	elem := New(
		WithText("Hello"),
		WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
		WithSize(20, 3),
	)
	elem.Calculate(30, 10)

	RenderTree(buf, elem)

	// Check that text was drawn at left-aligned position
	// Text should start at (0, 0) with content "Hello"
	checkString(t, buf, 0, 0, "Hello")

	// Check style
	cell := buf.Cell(0, 0)
	if cell.Style.Fg != tui.Green {
		t.Errorf("text style.Fg = %v, want Green", cell.Style.Fg)
	}
}

func TestRenderTree_TextCenterAlignment(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	elem := New(
		WithText("Hi"),
		WithTextAlign(TextAlignCenter),
		WithSize(10, 3),
	)
	elem.Calculate(30, 10)

	RenderTree(buf, elem)

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

func TestRenderTree_TextRightAlignment(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	elem := New(
		WithText("Hi"),
		WithTextAlign(TextAlignRight),
		WithSize(10, 3),
	)
	elem.Calculate(30, 10)

	RenderTree(buf, elem)

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

func TestRenderTree_TextWithBorderAndPadding(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	// Border takes 1 cell on each side, padding adds additional space inside.
	// With border + padding=1, content starts at position 2 (1 for border + 1 for padding).
	elem := New(
		WithText("Test"),
		WithSize(20, 5),
		WithBorder(tui.BorderSingle),
		WithPadding(1), // 1 cell padding inside the border
	)
	elem.Calculate(30, 10)

	RenderTree(buf, elem)

	// Border should be drawn at edge
	corner := buf.Cell(0, 0)
	if corner.Rune != '┌' {
		t.Errorf("border corner = %q, want '┌'", corner.Rune)
	}

	// Content rect accounts for border (1) + padding (1) = 2
	contentRect := elem.ContentRect()
	if contentRect.X != 2 || contentRect.Y != 2 {
		t.Errorf("content rect starts at (%d,%d), want (2,2)", contentRect.X, contentRect.Y)
	}

	// Text "Test" should be at content rect position (2, 2)
	checkString(t, buf, 2, 2, "Test")
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

func TestRenderTree_EmptyTextDoesNotRender(t *testing.T) {
	buf := tui.NewBuffer(20, 10)

	// Element with empty text and background
	elem := New(
		WithText(""),
		WithSize(10, 5),
		WithBackground(tui.NewStyle().Background(tui.Blue)),
	)
	elem.Calculate(20, 10)

	RenderTree(buf, elem)

	// Background should be drawn, but no text rendering should occur
	// This just verifies no panic and background is correct
	cell := buf.Cell(5, 2)
	if cell.Style.Bg != tui.Blue {
		t.Errorf("background not drawn correctly, got Bg=%v", cell.Style.Bg)
	}
}

func TestRenderTree_NestedTextElements(t *testing.T) {
	buf := tui.NewBuffer(40, 20)

	parent := New(
		WithSize(30, 10),
		WithPadding(2),
		WithBackground(tui.NewStyle().Background(tui.Blue)),
	)

	child := New(
		WithText("Hello"),
		WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
	)

	parent.AddChild(child)
	parent.Calculate(40, 20)

	RenderTree(buf, parent)

	// Child should be positioned at (2, 2) due to padding
	childRect := child.Rect()
	if childRect.X != 2 || childRect.Y != 2 {
		t.Errorf("child position = (%d,%d), want (2,2)", childRect.X, childRect.Y)
	}

	// Text should be rendered at child position
	checkString(t, buf, 2, 2, "Hello")

	// Check text style
	cell := buf.Cell(2, 2)
	if cell.Style.Fg != tui.Green {
		t.Errorf("text style.Fg = %v, want Green", cell.Style.Fg)
	}
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

// --- HR Rendering Tests ---

func TestRenderHRDefault(t *testing.T) {
	buf := tui.NewBuffer(20, 5)

	hr := New(WithHR(), WithWidth(10))
	hr.Calculate(20, 5)

	RenderTree(buf, hr)

	// HR should draw '─' characters across its width
	for x := 0; x < 10; x++ {
		cell := buf.Cell(x, 0)
		if cell.Rune != '─' {
			t.Errorf("HR at x=%d = %q, want '─'", x, cell.Rune)
		}
	}

	// Beyond the HR width should be untouched (spaces)
	for x := 10; x < 20; x++ {
		cell := buf.Cell(x, 0)
		if cell.Rune != ' ' {
			t.Errorf("beyond HR at x=%d = %q, want ' '", x, cell.Rune)
		}
	}
}

func TestRenderHRDouble(t *testing.T) {
	buf := tui.NewBuffer(20, 5)

	hr := New(WithHR(), WithWidth(10), WithBorder(tui.BorderDouble))
	hr.Calculate(20, 5)

	RenderTree(buf, hr)

	// HR with BorderDouble should draw '═' characters
	for x := 0; x < 10; x++ {
		cell := buf.Cell(x, 0)
		if cell.Rune != '═' {
			t.Errorf("HR double at x=%d = %q, want '═'", x, cell.Rune)
		}
	}
}

func TestRenderHRThick(t *testing.T) {
	buf := tui.NewBuffer(20, 5)

	hr := New(WithHR(), WithWidth(10), WithBorder(tui.BorderThick))
	hr.Calculate(20, 5)

	RenderTree(buf, hr)

	// HR with BorderThick should draw '━' characters
	for x := 0; x < 10; x++ {
		cell := buf.Cell(x, 0)
		if cell.Rune != '━' {
			t.Errorf("HR thick at x=%d = %q, want '━'", x, cell.Rune)
		}
	}
}

func TestRenderHRWithColor(t *testing.T) {
	buf := tui.NewBuffer(20, 5)

	hr := New(
		WithHR(),
		WithWidth(10),
		WithTextStyle(tui.NewStyle().Foreground(tui.Cyan)),
	)
	hr.Calculate(20, 5)

	RenderTree(buf, hr)

	// HR should respect textStyle for color
	for x := 0; x < 10; x++ {
		cell := buf.Cell(x, 0)
		if cell.Rune != '─' {
			t.Errorf("HR at x=%d = %q, want '─'", x, cell.Rune)
		}
		if cell.Style.Fg != tui.Cyan {
			t.Errorf("HR style at x=%d Fg = %v, want Cyan", x, cell.Style.Fg)
		}
	}
}

func TestRenderHRInContainer(t *testing.T) {
	buf := tui.NewBuffer(30, 10)

	// HR inside a column container should stretch to fill width
	container := New(
		WithSize(20, 5),
		WithDirection(layout.Column),
	)

	hr := New(WithHR())
	container.AddChild(hr)
	container.Calculate(30, 10)

	RenderTree(buf, container)

	// HR should stretch to fill container width (20)
	hrRect := hr.Rect()
	if hrRect.Width != 20 {
		t.Errorf("HR width = %d, want 20 (stretch to fill)", hrRect.Width)
	}

	// Check that HR drew '─' characters across the full width
	for x := 0; x < 20; x++ {
		cell := buf.Cell(x, 0)
		if cell.Rune != '─' {
			t.Errorf("HR at x=%d = %q, want '─'", x, cell.Rune)
		}
	}
}

func TestHRIntrinsicSize(t *testing.T) {
	hr := New(WithHR())

	w, h := hr.IntrinsicSize()

	// HR has 0 intrinsic width (relies on stretch) and height of 1
	if w != 0 {
		t.Errorf("HR intrinsic width = %d, want 0", w)
	}
	if h != 1 {
		t.Errorf("HR intrinsic height = %d, want 1", h)
	}
}

func TestHRIsHR(t *testing.T) {
	hr := New(WithHR())
	normal := New()

	if !hr.IsHR() {
		t.Error("WithHR() element.IsHR() = false, want true")
	}
	if normal.IsHR() {
		t.Error("normal element.IsHR() = true, want false")
	}
}
