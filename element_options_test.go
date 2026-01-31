package tui

import (
	"testing"
)

func TestWithWidth(t *testing.T) {
	e := New(WithWidth(100))
	if e.style.Width != Fixed(100) {
		t.Errorf("WithWidth(100) = %+v, want Fixed(100)", e.style.Width)
	}
}

func TestWithWidthPercent(t *testing.T) {
	e := New(WithWidthPercent(50))
	if e.style.Width != Percent(50) {
		t.Errorf("WithWidthPercent(50) = %+v, want Percent(50)", e.style.Width)
	}
}

func TestWithHeight(t *testing.T) {
	e := New(WithHeight(80))
	if e.style.Height != Fixed(80) {
		t.Errorf("WithHeight(80) = %+v, want Fixed(80)", e.style.Height)
	}
}

func TestWithHeightPercent(t *testing.T) {
	e := New(WithHeightPercent(25))
	if e.style.Height != Percent(25) {
		t.Errorf("WithHeightPercent(25) = %+v, want Percent(25)", e.style.Height)
	}
}

func TestWithSize(t *testing.T) {
	e := New(WithSize(120, 60))
	if e.style.Width != Fixed(120) {
		t.Errorf("WithSize(120, 60) Width = %+v, want Fixed(120)", e.style.Width)
	}
	if e.style.Height != Fixed(60) {
		t.Errorf("WithSize(120, 60) Height = %+v, want Fixed(60)", e.style.Height)
	}
}

func TestWithMinWidth(t *testing.T) {
	e := New(WithMinWidth(20))
	if e.style.MinWidth != Fixed(20) {
		t.Errorf("WithMinWidth(20) = %+v, want Fixed(20)", e.style.MinWidth)
	}
}

func TestWithMinHeight(t *testing.T) {
	e := New(WithMinHeight(15))
	if e.style.MinHeight != Fixed(15) {
		t.Errorf("WithMinHeight(15) = %+v, want Fixed(15)", e.style.MinHeight)
	}
}

func TestWithMaxWidth(t *testing.T) {
	e := New(WithMaxWidth(200))
	if e.style.MaxWidth != Fixed(200) {
		t.Errorf("WithMaxWidth(200) = %+v, want Fixed(200)", e.style.MaxWidth)
	}
}

func TestWithMaxHeight(t *testing.T) {
	e := New(WithMaxHeight(150))
	if e.style.MaxHeight != Fixed(150) {
		t.Errorf("WithMaxHeight(150) = %+v, want Fixed(150)", e.style.MaxHeight)
	}
}

func TestWithDirection(t *testing.T) {
	type tc struct {
		dir    Direction
		expect Direction
	}

	tests := map[string]tc{
		"Row":    {dir: Row, expect: Row},
		"Column": {dir: Column, expect: Column},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithDirection(tt.dir))
			if e.style.Direction != tt.expect {
				t.Errorf("WithDirection(%v) = %v, want %v", tt.dir, e.style.Direction, tt.expect)
			}
		})
	}
}

func TestWithJustify(t *testing.T) {
	type tc struct {
		justify Justify
	}

	tests := map[string]tc{
		"Start":        {justify: JustifyStart},
		"End":          {justify: JustifyEnd},
		"Center":       {justify: JustifyCenter},
		"SpaceBetween": {justify: JustifySpaceBetween},
		"SpaceAround":  {justify: JustifySpaceAround},
		"SpaceEvenly":  {justify: JustifySpaceEvenly},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithJustify(tt.justify))
			if e.style.JustifyContent != tt.justify {
				t.Errorf("WithJustify(%v) = %v", tt.justify, e.style.JustifyContent)
			}
		})
	}
}

func TestWithAlign(t *testing.T) {
	type tc struct {
		align Align
	}

	tests := map[string]tc{
		"Start":   {align: AlignStart},
		"End":     {align: AlignEnd},
		"Center":  {align: AlignCenter},
		"Stretch": {align: AlignStretch},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithAlign(tt.align))
			if e.style.AlignItems != tt.align {
				t.Errorf("WithAlign(%v) = %v", tt.align, e.style.AlignItems)
			}
		})
	}
}

func TestWithGap(t *testing.T) {
	e := New(WithGap(10))
	if e.style.Gap != 10 {
		t.Errorf("WithGap(10) = %d, want 10", e.style.Gap)
	}
}

func TestWithFlexGrow(t *testing.T) {
	e := New(WithFlexGrow(2.5))
	if e.style.FlexGrow != 2.5 {
		t.Errorf("WithFlexGrow(2.5) = %f, want 2.5", e.style.FlexGrow)
	}
}

func TestWithFlexShrink(t *testing.T) {
	e := New(WithFlexShrink(0.5))
	if e.style.FlexShrink != 0.5 {
		t.Errorf("WithFlexShrink(0.5) = %f, want 0.5", e.style.FlexShrink)
	}
}

func TestWithAlignSelf(t *testing.T) {
	e := New(WithAlignSelf(AlignCenter))
	if e.style.AlignSelf == nil || *e.style.AlignSelf != AlignCenter {
		t.Error("WithAlignSelf(AlignCenter) should set AlignSelf to AlignCenter")
	}
}

func TestWithPadding(t *testing.T) {
	e := New(WithPadding(5))
	expected := EdgeAll(5)
	if e.style.Padding != expected {
		t.Errorf("WithPadding(5) = %+v, want %+v", e.style.Padding, expected)
	}
}

func TestWithPaddingTRBL(t *testing.T) {
	e := New(WithPaddingTRBL(1, 2, 3, 4))
	expected := EdgeTRBL(1, 2, 3, 4)
	if e.style.Padding != expected {
		t.Errorf("WithPaddingTRBL(1,2,3,4) = %+v, want %+v", e.style.Padding, expected)
	}
}

func TestWithMargin(t *testing.T) {
	e := New(WithMargin(8))
	expected := EdgeAll(8)
	if e.style.Margin != expected {
		t.Errorf("WithMargin(8) = %+v, want %+v", e.style.Margin, expected)
	}
}

func TestWithMarginTRBL(t *testing.T) {
	e := New(WithMarginTRBL(2, 4, 6, 8))
	expected := EdgeTRBL(2, 4, 6, 8)
	if e.style.Margin != expected {
		t.Errorf("WithMarginTRBL(2,4,6,8) = %+v, want %+v", e.style.Margin, expected)
	}
}

func TestWithBorder(t *testing.T) {
	type tc struct {
		border BorderStyle
	}

	tests := map[string]tc{
		"None":    {border: BorderNone},
		"Single":  {border: BorderSingle},
		"Double":  {border: BorderDouble},
		"Rounded": {border: BorderRounded},
		"Thick":   {border: BorderThick},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithBorder(tt.border))
			if e.border != tt.border {
				t.Errorf("WithBorder(%v) = %v", tt.border, e.border)
			}
		})
	}
}

func TestWithBorderStyle(t *testing.T) {
	style := NewStyle().Foreground(Red).Bold()
	e := New(WithBorderStyle(style))
	if e.borderStyle != style {
		t.Errorf("WithBorderStyle() = %+v, want %+v", e.borderStyle, style)
	}
}

func TestWithBackground(t *testing.T) {
	style := NewStyle().Background(Blue)
	e := New(WithBackground(style))
	if e.background == nil {
		t.Error("WithBackground() should set background")
	}
	if *e.background != style {
		t.Errorf("WithBackground() = %+v, want %+v", *e.background, style)
	}
}

func TestOptions_Compose(t *testing.T) {
	// Test that multiple options can be composed
	e := New(
		WithSize(100, 50),
		WithDirection(Column),
		WithJustify(JustifyCenter),
		WithAlign(AlignCenter),
		WithPadding(5),
		WithBorder(BorderRounded),
		WithBorderStyle(NewStyle().Foreground(Cyan)),
	)

	if e.style.Width != Fixed(100) {
		t.Error("Width not set correctly")
	}
	if e.style.Height != Fixed(50) {
		t.Error("Height not set correctly")
	}
	if e.style.Direction != Column {
		t.Error("Direction not set correctly")
	}
	if e.style.JustifyContent != JustifyCenter {
		t.Error("JustifyContent not set correctly")
	}
	if e.style.AlignItems != AlignCenter {
		t.Error("AlignItems not set correctly")
	}
	if e.style.Padding != EdgeAll(5) {
		t.Error("Padding not set correctly")
	}
	if e.border != BorderRounded {
		t.Error("Border not set correctly")
	}
	if e.borderStyle.Fg != Cyan {
		t.Error("BorderStyle not set correctly")
	}
}

func TestOptions_Override(t *testing.T) {
	// Test that later options override earlier ones
	e := New(
		WithWidth(100),
		WithWidth(200),
	)

	if e.style.Width != Fixed(200) {
		t.Errorf("Later WithWidth should override earlier, got %+v", e.style.Width)
	}
}

func TestWithWidthAuto(t *testing.T) {
	e := New(WithWidthAuto())
	if e.style.Width != Auto() {
		t.Errorf("WithWidthAuto() = %+v, want Auto()", e.style.Width)
	}
}

func TestWithHeightAuto(t *testing.T) {
	e := New(WithHeightAuto())
	if e.style.Height != Auto() {
		t.Errorf("WithHeightAuto() = %+v, want Auto()", e.style.Height)
	}
}

// --- Event Handler Option Tests ---

func TestWithOnKeyPress_SetsHandler(t *testing.T) {
	handlerCalled := false
	e := New(WithOnKeyPress(func(_ *Element, event KeyEvent) {
		handlerCalled = true
	}))

	if e.onKeyPress == nil {
		t.Error("WithOnKeyPress should set the onKeyPress handler")
	}

	// Verify handler works
	e.onKeyPress(e, KeyEvent{Key: KeyEnter})
	if !handlerCalled {
		t.Error("onKeyPress handler should be callable")
	}
}

func TestWithOnKeyPress_SetsFocusable(t *testing.T) {
	e := New(WithOnKeyPress(func(*Element, KeyEvent) {}))

	if !e.focusable {
		t.Error("WithOnKeyPress should set focusable = true")
	}
}

func TestWithOnClick_SetsHandler(t *testing.T) {
	clickCalled := false
	e := New(WithOnClick(func(_ *Element) {
		clickCalled = true
	}))

	if e.onClick == nil {
		t.Error("WithOnClick should set the onClick handler")
	}

	// Verify handler works
	e.onClick(e)
	if !clickCalled {
		t.Error("onClick handler should be callable")
	}
}

func TestWithOnClick_SetsFocusable(t *testing.T) {
	e := New(WithOnClick(func(*Element) {}))

	if !e.focusable {
		t.Error("WithOnClick should set focusable = true")
	}
}

func TestWithFocusable_True(t *testing.T) {
	e := New(WithFocusable(true))

	if !e.focusable {
		t.Error("WithFocusable(true) should set focusable = true")
	}
}

func TestWithFocusable_False(t *testing.T) {
	// First make it focusable via another option
	e := New(
		WithOnFocus(func(*Element) {}), // This sets focusable = true
		WithFocusable(false),           // This should override to false
	)

	if e.focusable {
		t.Error("WithFocusable(false) should override focusable to false")
	}
}
