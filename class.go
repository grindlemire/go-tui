package tui

import (
	"fmt"

	"github.com/grindlemire/go-tui/internal/tailwind"
)

// WithClass applies Tailwind-style utility classes at runtime, resolving the
// same class set the .gsx compiler accepts for literal class attributes. The
// generator emits it for expression-valued class attributes. Unknown classes
// are ignored, so validation happens only for literals at compile time.
func WithClass(classes string) Option {
	return func(e *Element) {
		e.applyClasses(classes)
	}
}

// SetClass re-applies a class string to an existing element. Properties the
// classes mention are overwritten; everything else is left as is.
func (e *Element) SetClass(classes string) {
	e.applyClasses(classes)
	e.MarkDirty()
}

func (e *Element) applyClasses(classes string) {
	result := tailwind.Parse(classes)
	for _, op := range result.Ops() {
		classOption(op)(e)
	}
	if text := result.Text(); len(text) > 0 {
		WithTextStyle(classTextStyle(text))(e)
	}
}

// classTextStyle chains text modifiers onto a fresh style in class order, so
// a later color class overrides an earlier one just like the compiled form.
func classTextStyle(ops []tailwind.TextOp) Style {
	s := NewStyle()
	for _, op := range ops {
		switch op := op.(type) {
		case tailwind.TextAttr:
			switch op.Attr {
			case tailwind.Bold:
				s = s.Bold()
			case tailwind.Dim:
				s = s.Dim()
			case tailwind.Italic:
				s = s.Italic()
			case tailwind.Underline:
				s = s.Underline()
			case tailwind.Blink:
				s = s.Blink()
			case tailwind.Reverse:
				s = s.Reverse()
			case tailwind.Strikethrough:
				s = s.Strikethrough()
			}
		case tailwind.Foreground:
			s = s.Foreground(classColor(op.Color))
		}
	}
	return s
}

// classOption maps one resolved op to the element option it stands for.
func classOption(op tailwind.Op) Option {
	switch op := op.(type) {
	case tailwind.Display:
		if op.Flex {
			return WithDisplay(DisplayFlex)
		}
		return WithDisplay(DisplayBlock)
	case tailwind.Direction:
		if op.Column {
			return WithDirection(Column)
		}
		return WithDirection(Row)
	case tailwind.FlexWrap:
		return WithFlexWrap([...]FlexWrap{WrapNone, Wrap, WrapReverse}[op.Mode])
	case tailwind.AlignContent:
		return WithAlignContent([...]AlignContent{ContentStart, ContentEnd, ContentCenter, ContentStretch, ContentSpaceBetween, ContentSpaceAround}[op.Value])
	case tailwind.FlexGrow:
		return WithFlexGrow(float64(op.N))
	case tailwind.FlexShrink:
		return WithFlexShrink(float64(op.N))
	case tailwind.Justify:
		return WithJustify([...]Justify{JustifyStart, JustifyCenter, JustifyEnd, JustifySpaceBetween, JustifySpaceAround, JustifySpaceEvenly}[op.Value])
	case tailwind.AlignItems:
		return WithAlign(classAlign(op.Value))
	case tailwind.AlignSelf:
		return WithAlignSelf(classAlign(op.Value))
	case tailwind.TextAlign:
		return WithTextAlign([...]TextAlign{TextAlignLeft, TextAlignCenter, TextAlignRight}[op.Value])
	case tailwind.Border:
		return WithBorder([...]BorderStyle{BorderSingle, BorderRounded, BorderDouble, BorderThick}[op.Style])
	case tailwind.BorderColor:
		return WithBorderStyle(NewStyle().Foreground(classColor(op.Color)))
	case tailwind.Background:
		return WithBackground(NewStyle().Background(classColor(op.Color)))
	case tailwind.ScrollbarColor:
		return WithScrollbarStyle(NewStyle().Foreground(classColor(op.Color)))
	case tailwind.ScrollbarThumbColor:
		return WithScrollbarThumbStyle(NewStyle().Foreground(classColor(op.Color)))
	case tailwind.Scroll:
		return WithScrollable([...]ScrollMode{ScrollBoth, ScrollVertical, ScrollHorizontal}[op.Mode])
	case tailwind.OverflowHidden:
		return WithOverflow(OverflowHidden)
	case tailwind.Focusable:
		return WithFocusable(true)
	case tailwind.Hidden:
		return WithHidden(true)
	case tailwind.Truncate:
		return WithTruncate(true)
	case tailwind.ScrollbarHidden:
		return WithScrollbarHidden(true)
	case tailwind.Wrap:
		return WithWrap(op.Enabled)
	case tailwind.Gap:
		return WithGap(op.N)
	case tailwind.Padding:
		return WithPadding(op.N)
	case tailwind.Margin:
		return WithMargin(op.N)
	case tailwind.Width:
		return WithWidth(op.N)
	case tailwind.Height:
		return WithHeight(op.N)
	case tailwind.MinWidth:
		return WithMinWidth(op.N)
	case tailwind.MaxWidth:
		return WithMaxWidth(op.N)
	case tailwind.MinHeight:
		return WithMinHeight(op.N)
	case tailwind.MaxHeight:
		return WithMaxHeight(op.N)
	case tailwind.PaddingEdges:
		return WithPaddingTRBL(op.Top, op.Right, op.Bottom, op.Left)
	case tailwind.MarginEdges:
		return WithMarginTRBL(op.Top, op.Right, op.Bottom, op.Left)
	case tailwind.WidthPercent:
		return WithWidthPercent(op.Percent)
	case tailwind.HeightPercent:
		return WithHeightPercent(op.Percent)
	case tailwind.WidthAuto:
		return WithWidthAuto()
	case tailwind.HeightAuto:
		return WithHeightAuto()
	case tailwind.TextGradient:
		return WithTextGradient(classGradient(op.Start, op.End, op.Direction))
	case tailwind.BackgroundGradient:
		return WithBackgroundGradient(classGradient(op.Start, op.End, op.Direction))
	case tailwind.BorderGradient:
		return WithBorderGradient(classGradient(op.Start, op.End, op.Direction))
	}
	panic(fmt.Sprintf("tui: unhandled tailwind op %T", op))
}

func classAlign(v tailwind.AlignValue) Align {
	return [...]Align{AlignStart, AlignCenter, AlignEnd, AlignStretch}[v]
}

func classColor(c tailwind.Color) Color {
	if c.IsRGB {
		return RGBColor(c.R, c.G, c.B)
	}
	return ANSIColor(c.Index)
}

func classGradient(start, end tailwind.Color, dir tailwind.GradientDirection) Gradient {
	d := [...]GradientDirection{GradientHorizontal, GradientVertical, GradientDiagonalDown, GradientDiagonalUp}[dir]
	return NewGradient(classColor(start), classColor(end)).WithDirection(d)
}
