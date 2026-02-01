package tui

import (
	"testing"
)

// --- HR Rendering Tests ---

func TestRenderHRDefault(t *testing.T) {
	buf := NewBuffer(20, 5)

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
	buf := NewBuffer(20, 5)

	hr := New(WithHR(), WithWidth(10), WithBorder(BorderDouble))
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
	buf := NewBuffer(20, 5)

	hr := New(WithHR(), WithWidth(10), WithBorder(BorderThick))
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
	buf := NewBuffer(20, 5)

	hr := New(
		WithHR(),
		WithWidth(10),
		WithTextStyle(NewStyle().Foreground(Cyan)),
	)
	hr.Calculate(20, 5)

	RenderTree(buf, hr)

	// HR should respect textStyle for color
	for x := 0; x < 10; x++ {
		cell := buf.Cell(x, 0)
		if cell.Rune != '─' {
			t.Errorf("HR at x=%d = %q, want '─'", x, cell.Rune)
		}
		if cell.Style.Fg != Cyan {
			t.Errorf("HR style at x=%d Fg = %v, want Cyan", x, cell.Style.Fg)
		}
	}
}

func TestRenderHRInContainer(t *testing.T) {
	buf := NewBuffer(30, 10)

	// HR inside a column container should stretch to fill width
	container := New(
		WithSize(20, 5),
		WithDirection(Column),
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

func TestRenderHRInScrollableContainer(t *testing.T) {
	buf := NewBuffer(30, 10)

	// Scrollable container with HR child
	container := New(
		WithSize(20, 8),
		WithDirection(Column),
		WithScrollable(ScrollVertical),
		WithPadding(1),
	)

	hr := New(WithHR())
	container.AddChild(hr)
	container.Calculate(30, 10)

	RenderTree(buf, container)

	// HR should be rendered inside the scrollable container
	// Container padding is 1, so HR should start at x=1, y=1
	hrRect := hr.Rect()

	// The HR should stretch to fill the content width (20 - 2 padding = 18)
	expectedWidth := 18
	if hrRect.Width != expectedWidth {
		t.Errorf("HR width = %d, want %d (stretch to fill content area)", hrRect.Width, expectedWidth)
	}

	// Check that HR drew '─' characters at the correct position
	// The container starts at (0,0), padding is 1, so HR line is at y=1, x=1 to x=18
	for x := 1; x < 19; x++ {
		cell := buf.Cell(x, 1)
		if cell.Rune != '─' {
			t.Errorf("HR at x=%d, y=1 = %q, want '─'", x, cell.Rune)
		}
	}
}
