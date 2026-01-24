package element

import (
	"testing"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

func TestWithWidth(t *testing.T) {
	e := New(WithWidth(100))
	if e.style.Width != layout.Fixed(100) {
		t.Errorf("WithWidth(100) = %+v, want Fixed(100)", e.style.Width)
	}
}

func TestWithWidthPercent(t *testing.T) {
	e := New(WithWidthPercent(50))
	if e.style.Width != layout.Percent(50) {
		t.Errorf("WithWidthPercent(50) = %+v, want Percent(50)", e.style.Width)
	}
}

func TestWithHeight(t *testing.T) {
	e := New(WithHeight(80))
	if e.style.Height != layout.Fixed(80) {
		t.Errorf("WithHeight(80) = %+v, want Fixed(80)", e.style.Height)
	}
}

func TestWithHeightPercent(t *testing.T) {
	e := New(WithHeightPercent(25))
	if e.style.Height != layout.Percent(25) {
		t.Errorf("WithHeightPercent(25) = %+v, want Percent(25)", e.style.Height)
	}
}

func TestWithSize(t *testing.T) {
	e := New(WithSize(120, 60))
	if e.style.Width != layout.Fixed(120) {
		t.Errorf("WithSize(120, 60) Width = %+v, want Fixed(120)", e.style.Width)
	}
	if e.style.Height != layout.Fixed(60) {
		t.Errorf("WithSize(120, 60) Height = %+v, want Fixed(60)", e.style.Height)
	}
}

func TestWithMinWidth(t *testing.T) {
	e := New(WithMinWidth(20))
	if e.style.MinWidth != layout.Fixed(20) {
		t.Errorf("WithMinWidth(20) = %+v, want Fixed(20)", e.style.MinWidth)
	}
}

func TestWithMinHeight(t *testing.T) {
	e := New(WithMinHeight(15))
	if e.style.MinHeight != layout.Fixed(15) {
		t.Errorf("WithMinHeight(15) = %+v, want Fixed(15)", e.style.MinHeight)
	}
}

func TestWithMaxWidth(t *testing.T) {
	e := New(WithMaxWidth(200))
	if e.style.MaxWidth != layout.Fixed(200) {
		t.Errorf("WithMaxWidth(200) = %+v, want Fixed(200)", e.style.MaxWidth)
	}
}

func TestWithMaxHeight(t *testing.T) {
	e := New(WithMaxHeight(150))
	if e.style.MaxHeight != layout.Fixed(150) {
		t.Errorf("WithMaxHeight(150) = %+v, want Fixed(150)", e.style.MaxHeight)
	}
}

func TestWithDirection(t *testing.T) {
	type tc struct {
		dir    layout.Direction
		expect layout.Direction
	}

	tests := map[string]tc{
		"Row":    {dir: layout.Row, expect: layout.Row},
		"Column": {dir: layout.Column, expect: layout.Column},
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
		justify layout.Justify
	}

	tests := map[string]tc{
		"Start":        {justify: layout.JustifyStart},
		"End":          {justify: layout.JustifyEnd},
		"Center":       {justify: layout.JustifyCenter},
		"SpaceBetween": {justify: layout.JustifySpaceBetween},
		"SpaceAround":  {justify: layout.JustifySpaceAround},
		"SpaceEvenly":  {justify: layout.JustifySpaceEvenly},
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
		align layout.Align
	}

	tests := map[string]tc{
		"Start":   {align: layout.AlignStart},
		"End":     {align: layout.AlignEnd},
		"Center":  {align: layout.AlignCenter},
		"Stretch": {align: layout.AlignStretch},
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
	e := New(WithAlignSelf(layout.AlignCenter))
	if e.style.AlignSelf == nil || *e.style.AlignSelf != layout.AlignCenter {
		t.Error("WithAlignSelf(AlignCenter) should set AlignSelf to AlignCenter")
	}
}

func TestWithPadding(t *testing.T) {
	e := New(WithPadding(5))
	expected := layout.EdgeAll(5)
	if e.style.Padding != expected {
		t.Errorf("WithPadding(5) = %+v, want %+v", e.style.Padding, expected)
	}
}

func TestWithPaddingTRBL(t *testing.T) {
	e := New(WithPaddingTRBL(1, 2, 3, 4))
	expected := layout.EdgeTRBL(1, 2, 3, 4)
	if e.style.Padding != expected {
		t.Errorf("WithPaddingTRBL(1,2,3,4) = %+v, want %+v", e.style.Padding, expected)
	}
}

func TestWithMargin(t *testing.T) {
	e := New(WithMargin(8))
	expected := layout.EdgeAll(8)
	if e.style.Margin != expected {
		t.Errorf("WithMargin(8) = %+v, want %+v", e.style.Margin, expected)
	}
}

func TestWithMarginTRBL(t *testing.T) {
	e := New(WithMarginTRBL(2, 4, 6, 8))
	expected := layout.EdgeTRBL(2, 4, 6, 8)
	if e.style.Margin != expected {
		t.Errorf("WithMarginTRBL(2,4,6,8) = %+v, want %+v", e.style.Margin, expected)
	}
}

func TestWithBorder(t *testing.T) {
	type tc struct {
		border tui.BorderStyle
	}

	tests := map[string]tc{
		"None":    {border: tui.BorderNone},
		"Single":  {border: tui.BorderSingle},
		"Double":  {border: tui.BorderDouble},
		"Rounded": {border: tui.BorderRounded},
		"Thick":   {border: tui.BorderThick},
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
	style := tui.NewStyle().Foreground(tui.Red).Bold()
	e := New(WithBorderStyle(style))
	if e.borderStyle != style {
		t.Errorf("WithBorderStyle() = %+v, want %+v", e.borderStyle, style)
	}
}

func TestWithBackground(t *testing.T) {
	style := tui.NewStyle().Background(tui.Blue)
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
		WithDirection(layout.Column),
		WithJustify(layout.JustifyCenter),
		WithAlign(layout.AlignCenter),
		WithPadding(5),
		WithBorder(tui.BorderRounded),
		WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
	)

	if e.style.Width != layout.Fixed(100) {
		t.Error("Width not set correctly")
	}
	if e.style.Height != layout.Fixed(50) {
		t.Error("Height not set correctly")
	}
	if e.style.Direction != layout.Column {
		t.Error("Direction not set correctly")
	}
	if e.style.JustifyContent != layout.JustifyCenter {
		t.Error("JustifyContent not set correctly")
	}
	if e.style.AlignItems != layout.AlignCenter {
		t.Error("AlignItems not set correctly")
	}
	if e.style.Padding != layout.EdgeAll(5) {
		t.Error("Padding not set correctly")
	}
	if e.border != tui.BorderRounded {
		t.Error("Border not set correctly")
	}
	if e.borderStyle.Fg != tui.Cyan {
		t.Error("BorderStyle not set correctly")
	}
}

func TestOptions_Override(t *testing.T) {
	// Test that later options override earlier ones
	e := New(
		WithWidth(100),
		WithWidth(200),
	)

	if e.style.Width != layout.Fixed(200) {
		t.Errorf("Later WithWidth should override earlier, got %+v", e.style.Width)
	}
}
