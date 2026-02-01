package tui

import (
	"testing"
)

func TestRenderTree_DrawsBackground(t *testing.T) {
	buf := NewBuffer(20, 10)
	bgStyle := NewStyle().Background(Blue)
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
			if cell.Style.Bg != Blue {
				t.Errorf("cell(%d,%d).Style.Bg = %v, want Blue", x, y, cell.Style.Bg)
			}
		}
	}
}

func TestRenderTree_DrawsBorder(t *testing.T) {
	buf := NewBuffer(20, 10)
	e := New(
		WithSize(10, 5),
		WithBorder(BorderSingle),
		WithBorderStyle(NewStyle().Foreground(Red)),
	)
	e.Calculate(20, 10)

	RenderTree(buf, e)

	// Check corners
	topLeft := buf.Cell(0, 0)
	if topLeft.Rune != '┌' {
		t.Errorf("top-left corner = %q, want '┌'", topLeft.Rune)
	}
	if topLeft.Style.Fg != Red {
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
	buf := NewBuffer(30, 20)

	parent := New(
		WithSize(20, 15),
		WithPadding(2),
		WithBackground(NewStyle().Background(Blue)),
	)

	child := New(
		WithSize(10, 5),
		WithBorder(BorderSingle),
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
	buf := NewBuffer(10, 10)

	// Element positioned outside buffer
	e := New(WithSize(5, 5))
	e.Calculate(100, 100)
	// Manually set position outside buffer
	e.layout.Rect = NewRect(50, 50, 5, 5)

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
	buf := NewBuffer(30, 10)

	elem := New(
		WithText("Hello"),
		WithTextStyle(NewStyle().Foreground(Green)),
		WithSize(20, 3),
	)
	elem.Calculate(30, 10)

	RenderTree(buf, elem)

	// Check that text was drawn at left-aligned position
	// Text should start at (0, 0) with content "Hello"
	checkString(t, buf, 0, 0, "Hello")

	// Check style
	cell := buf.Cell(0, 0)
	if cell.Style.Fg != Green {
		t.Errorf("text style.Fg = %v, want Green", cell.Style.Fg)
	}
}

func TestRenderTree_TextCenterAlignment(t *testing.T) {
	buf := NewBuffer(30, 10)

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
	buf := NewBuffer(30, 10)

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
	buf := NewBuffer(30, 10)

	// Border takes 1 cell on each side, padding adds additional space inside.
	// With border + padding=1, content starts at position 2 (1 for border + 1 for padding).
	elem := New(
		WithText("Test"),
		WithSize(20, 5),
		WithBorder(BorderSingle),
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
	buf := NewBuffer(30, 20)

	e := New(
		WithSize(20, 10),
		WithBorder(BorderSingle),
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
	buf := NewBuffer(30, 20)

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
	buf := NewBuffer(20, 10)

	// Element with empty text and background
	elem := New(
		WithText(""),
		WithSize(10, 5),
		WithBackground(NewStyle().Background(Blue)),
	)
	elem.Calculate(20, 10)

	RenderTree(buf, elem)

	// Background should be drawn, but no text rendering should occur
	// This just verifies no panic and background is correct
	cell := buf.Cell(5, 2)
	if cell.Style.Bg != Blue {
		t.Errorf("background not drawn correctly, got Bg=%v", cell.Style.Bg)
	}
}

func TestRenderTree_NestedTextElements(t *testing.T) {
	buf := NewBuffer(40, 20)

	parent := New(
		WithSize(30, 10),
		WithPadding(2),
		WithBackground(NewStyle().Background(Blue)),
	)

	child := New(
		WithText("Hello"),
		WithTextStyle(NewStyle().Foreground(Green)),
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
	if cell.Style.Fg != Green {
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
func checkString(t *testing.T, buf *Buffer, x, y int, expected string) {
	t.Helper()
	curX := x
	for _, r := range expected {
		cell := buf.Cell(curX, y)
		if cell.Rune != r {
			t.Errorf("buf.Cell(%d,%d).Rune = %q, want %q", curX, y, cell.Rune, r)
		}
		curX += RuneWidth(r)
	}
}

func TestRenderTree_TextWithBorder(t *testing.T) {
	type tc struct {
		name       string
		text       string
		wantWidth  int
		wantHeight int
	}

	tests := map[string]tc{
		"short text with border": {
			text:       "Hi",
			wantWidth:  4,  // 2 + border(2)
			wantHeight: 3,  // 1 + border(2)
		},
		"longer text with border": {
			text:       "Text Styles",
			wantWidth:  13, // 11 + border(2)
			wantHeight: 3,  // 1 + border(2)
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			span := New(
				WithText(tt.text),
				WithBorder(BorderSingle),
			)

			parent := New(
				WithDirection(Column),
				WithSize(40, 20),
			)
			parent.AddChild(span)

			buf := NewBuffer(40, 20)
			parent.Render(buf, 40, 20)

			rect := span.Rect()
			if rect.Width != tt.wantWidth {
				t.Errorf("Rect.Width = %d, want %d", rect.Width, tt.wantWidth)
			}
			if rect.Height != tt.wantHeight {
				t.Errorf("Rect.Height = %d, want %d", rect.Height, tt.wantHeight)
			}

			// Check border corners render correctly
			topLeft := buf.Cell(rect.X, rect.Y)
			if topLeft.Rune != '┌' {
				t.Errorf("top-left = %q, want '┌'", topLeft.Rune)
			}
			topRight := buf.Cell(rect.X+rect.Width-1, rect.Y)
			if topRight.Rune != '┐' {
				t.Errorf("top-right = %q, want '┐'", topRight.Rune)
			}
			bottomLeft := buf.Cell(rect.X, rect.Y+rect.Height-1)
			if bottomLeft.Rune != '└' {
				t.Errorf("bottom-left = %q, want '└'", bottomLeft.Rune)
			}
			bottomRight := buf.Cell(rect.X+rect.Width-1, rect.Y+rect.Height-1)
			if bottomRight.Rune != '┘' {
				t.Errorf("bottom-right = %q, want '┘'", bottomRight.Rune)
			}

			// Text inside border
			checkString(t, buf, rect.X+1, rect.Y+1, tt.text)
		})
	}
}

func TestRenderTree_TextBorderInScrollable(t *testing.T) {
	buf := NewBuffer(40, 20)

	// Scrollable parent (mirrors the styling example)
	parent := New(
		WithDirection(Column),
		WithGap(1),
		WithPadding(2),
		WithBorder(BorderRounded),
		WithHeightPercent(100),
		WithScrollable(ScrollVertical),
	)

	// Text span with border
	span := New(
		WithText("Text Styles"),
		WithBorder(BorderSingle),
		WithBorderStyle(NewStyle().Foreground(White)),
		WithTextStyle(NewStyle().Bold()),
	)
	parent.AddChild(span)

	parent.Render(buf, 40, 20)

	// Check span border corners in scrollable context
	parentCR := parent.ContentRect()
	spanR := span.Rect()
	sx := parentCR.X + spanR.X
	sy := parentCR.Y + spanR.Y

	topLeft := buf.Cell(sx, sy)
	if topLeft.Rune != '┌' {
		t.Errorf("span top-left at (%d,%d) = %q, want '┌'", sx, sy, topLeft.Rune)
	}

	topRight := buf.Cell(sx+spanR.Width-1, sy)
	if topRight.Rune != '┐' {
		t.Errorf("span top-right at (%d,%d) = %q, want '┐'", sx+spanR.Width-1, sy, topRight.Rune)
	}

	bottomLeft := buf.Cell(sx, sy+spanR.Height-1)
	if bottomLeft.Rune != '└' {
		t.Errorf("span bottom-left at (%d,%d) = %q, want '└'", sx, sy+spanR.Height-1, bottomLeft.Rune)
	}

	bottomRight := buf.Cell(sx+spanR.Width-1, sy+spanR.Height-1)
	if bottomRight.Rune != '┘' {
		t.Errorf("span bottom-right at (%d,%d) = %q, want '┘'", sx+spanR.Width-1, sy+spanR.Height-1, bottomRight.Rune)
	}

	// Text should be inside the border
	checkString(t, buf, sx+1, sy+1, "Text Styles")
}

func TestRenderTree_TextStyleInheritance(t *testing.T) {
	type tc struct {
		parentOpts []Option
		childOpts  []Option
		wantFg     Color
		wantAttrs  Attr
	}

	tests := map[string]tc{
		"child inherits parent fg color": {
			parentOpts: []Option{WithSize(20, 10), WithTextStyle(NewStyle().Foreground(Cyan))},
			childOpts:  []Option{WithText("hi")},
			wantFg:     Cyan,
			wantAttrs:  AttrNone,
		},
		"child inherits parent bold attr": {
			parentOpts: []Option{WithSize(20, 10), WithTextStyle(NewStyle().Bold())},
			childOpts:  []Option{WithText("hi")},
			wantFg:     DefaultColor(),
			wantAttrs:  AttrBold,
		},
		"child inherits parent fg and bold": {
			parentOpts: []Option{WithSize(20, 10), WithTextStyle(NewStyle().Foreground(Red).Bold())},
			childOpts:  []Option{WithText("hi")},
			wantFg:     Red,
			wantAttrs:  AttrBold,
		},
		"child with explicit style overrides parent": {
			parentOpts: []Option{WithSize(20, 10), WithTextStyle(NewStyle().Foreground(Red).Bold())},
			childOpts:  []Option{WithText("hi"), WithTextStyle(NewStyle().Foreground(Green))},
			wantFg:     Green,
			wantAttrs:  AttrNone, // child's style replaces entirely, no bold
		},
		"child with explicit default style overrides parent": {
			parentOpts: []Option{WithSize(20, 10), WithTextStyle(NewStyle().Foreground(Red))},
			childOpts:  []Option{WithText("hi"), WithTextStyle(NewStyle())},
			wantFg:     DefaultColor(),
			wantAttrs:  AttrNone,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := New(tt.parentOpts...)
			child := New(tt.childOpts...)
			parent.AddChild(child)

			buf := NewBuffer(20, 10)
			parent.Render(buf, 20, 10)

			childRect := child.Rect()
			cell := buf.Cell(childRect.X, childRect.Y)
			if cell.Style.Fg != tt.wantFg {
				t.Errorf("text Fg = %v, want %v", cell.Style.Fg, tt.wantFg)
			}
			if cell.Style.Attrs != tt.wantAttrs {
				t.Errorf("text Attrs = %v, want %v", cell.Style.Attrs, tt.wantAttrs)
			}
		})
	}
}

func TestRenderTree_BackgroundInheritance(t *testing.T) {
	type tc struct {
		parentOpts []Option
		childOpts  []Option
		wantBg     Color
	}

	tests := map[string]tc{
		"child inherits parent background": {
			parentOpts: []Option{WithSize(20, 10), WithBackground(NewStyle().Background(Blue))},
			childOpts:  []Option{WithText("hi")},
			wantBg:     Blue,
		},
		"child with explicit background overrides parent": {
			parentOpts: []Option{WithSize(20, 10), WithBackground(NewStyle().Background(Blue))},
			childOpts:  []Option{WithText("hi"), WithBackground(NewStyle().Background(Red))},
			wantBg:     Red,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := New(tt.parentOpts...)
			child := New(tt.childOpts...)
			parent.AddChild(child)

			buf := NewBuffer(20, 10)
			parent.Render(buf, 20, 10)

			// Check text cell has inherited/overridden background
			childRect := child.Rect()
			cell := buf.Cell(childRect.X, childRect.Y)
			if cell.Style.Bg != tt.wantBg {
				t.Errorf("text Bg = %v, want %v", cell.Style.Bg, tt.wantBg)
			}
		})
	}
}

func TestRenderTree_DeepInheritance(t *testing.T) {
	// grandparent (red, bold) -> parent (no style) -> child (text)
	// child should inherit red+bold through the parent
	grandparent := New(WithSize(30, 10), WithTextStyle(NewStyle().Foreground(Red).Bold()))
	parent := New() // no explicit style
	child := New(WithText("deep"))

	grandparent.AddChild(parent)
	parent.AddChild(child)

	buf := NewBuffer(30, 10)
	grandparent.Render(buf, 30, 10)

	childRect := child.Rect()
	cell := buf.Cell(childRect.X, childRect.Y)
	if cell.Style.Fg != Red {
		t.Errorf("deeply inherited Fg = %v, want Red", cell.Style.Fg)
	}
	if cell.Style.Attrs != AttrBold {
		t.Errorf("deeply inherited Attrs = %v, want Bold", cell.Style.Attrs)
	}
}

func TestRenderTree_BorderStyleDoesNotInherit(t *testing.T) {
	parent := New(
		WithSize(20, 10),
		WithBorderStyle(NewStyle().Foreground(Red)),
	)
	child := New(
		WithSize(10, 5),
		WithBorder(BorderSingle),
	)
	parent.AddChild(child)

	buf := NewBuffer(20, 10)
	parent.Render(buf, 20, 10)

	// Child's border should use default style, not parent's red border style
	childRect := child.Rect()
	corner := buf.Cell(childRect.X, childRect.Y)
	if corner.Rune != '┌' {
		t.Errorf("child border corner = %q, want '┌'", corner.Rune)
	}
	if corner.Style.Fg != DefaultColor() {
		t.Errorf("child border Fg = %v, want Default (border should not inherit)", corner.Style.Fg)
	}
}

func TestRenderTree_HRInheritsTextStyle(t *testing.T) {
	parent := New(
		WithSize(20, 10),
		WithDirection(Column),
		WithTextStyle(NewStyle().Foreground(Cyan)),
	)
	hr := New(WithHR())
	parent.AddChild(hr)

	buf := NewBuffer(20, 10)
	parent.Render(buf, 20, 10)

	// HR should render with inherited cyan color
	hrRect := hr.Rect()
	cell := buf.Cell(hrRect.X, hrRect.Y)
	if cell.Rune != '─' {
		t.Errorf("HR rune = %q, want '─'", cell.Rune)
	}
	if cell.Style.Fg != Cyan {
		t.Errorf("HR inherited Fg = %v, want Cyan", cell.Style.Fg)
	}
}

func TestRenderTree_AutoContrast(t *testing.T) {
	type tc struct {
		bgColor   Color
		textStyle Style
		wantFg    Color
		desc      string
	}

	tests := map[string]tc{
		"light background gets black text": {
			bgColor:   White,
			textStyle: Style{}, // default fg
			wantFg:    Black,
			desc:      "white bg should auto-set black text",
		},
		"bright white background gets black text": {
			bgColor:   BrightWhite,
			textStyle: Style{}, // default fg
			wantFg:    Black,
			desc:      "bright white bg should auto-set black text",
		},
		"bright yellow background gets black text": {
			bgColor:   BrightYellow,
			textStyle: Style{}, // default fg
			wantFg:    Black,
			desc:      "bright yellow bg should auto-set black text",
		},
		"dark background keeps default text": {
			bgColor:   Black,
			textStyle: Style{}, // default fg
			wantFg:    DefaultColor(),
			desc:      "black bg should keep default text",
		},
		"blue background keeps default text": {
			bgColor:   Blue,
			textStyle: Style{}, // default fg
			wantFg:    DefaultColor(),
			desc:      "blue bg should keep default text",
		},
		"explicit fg color not overridden on light bg": {
			bgColor:   White,
			textStyle: NewStyle().Foreground(Red),
			wantFg:    Red,
			desc:      "explicit red fg should not be changed by auto-contrast",
		},
		"light RGB background gets black text": {
			bgColor:   RGBColor(255, 255, 200), // light yellow
			textStyle: Style{},
			wantFg:    Black,
			desc:      "light RGB bg should auto-set black text",
		},
		"dark RGB background keeps default text": {
			bgColor:   RGBColor(20, 20, 30), // dark blue
			textStyle: Style{},
			wantFg:    DefaultColor(),
			desc:      "dark RGB bg should keep default text",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var parentOpts []Option
			parentOpts = append(parentOpts,
				WithSize(20, 10),
				WithBackground(NewStyle().Background(tt.bgColor)),
			)
			if !tt.textStyle.Fg.IsDefault() || tt.textStyle.Attrs != AttrNone {
				parentOpts = append(parentOpts, WithTextStyle(tt.textStyle))
			}

			parent := New(parentOpts...)
			child := New(WithText("test"))
			parent.AddChild(child)

			buf := NewBuffer(20, 10)
			parent.Render(buf, 20, 10)

			childRect := child.Rect()
			cell := buf.Cell(childRect.X, childRect.Y)
			if cell.Style.Fg != tt.wantFg {
				t.Errorf("%s: text Fg = %v, want %v", tt.desc, cell.Style.Fg, tt.wantFg)
			}
		})
	}
}

