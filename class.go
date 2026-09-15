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

// SetClass replaces the element's classes. Every property the previous class
// string set returns to its default first, then the new string is applied, so
// omitting a class removes it. Properties set by other options are left alone
// unless the new string sets them too.
func (e *Element) SetClass(classes string) {
	for _, op := range e.classOps {
		classReset(op)(e)
	}
	if e.classText {
		e.textStyle = Style{}
		e.textStyleSet = false
	}
	e.classOps, e.classText = nil, false
	e.applyClasses(classes)
	e.MarkDirty()
}

func (e *Element) applyClasses(classes string) {
	result := tailwind.Parse(classes)
	ops := result.Ops()
	for _, op := range ops {
		classOption(op)(e)
	}
	e.classOps = append(e.classOps, ops...)
	if text := result.Text(); len(text) > 0 {
		WithTextStyle(classTextStyle(text))(e)
		e.classText = true
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
	case tailwind.WidthFraction:
		return WithWidthPercent(op.Percent())
	case tailwind.HeightFraction:
		return WithHeightPercent(op.Percent())
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

// classReset returns the option that restores the New() default of the
// property op set. Kept next to classOption so the two switches stay in step;
// TestSetClass_EveryClassResetsToDefault pins that every op resets cleanly.
func classReset(op tailwind.Op) Option {
	def := DefaultLayoutStyle()
	return func(e *Element) {
		switch op.(type) {
		case tailwind.Display:
			e.style.Display = def.Display
		case tailwind.Direction:
			e.style.Direction = def.Direction
		case tailwind.FlexWrap:
			e.style.FlexWrap = def.FlexWrap
		case tailwind.AlignContent:
			e.style.AlignContent = def.AlignContent
		case tailwind.FlexGrow:
			e.style.FlexGrow = def.FlexGrow
		case tailwind.FlexShrink:
			e.style.FlexShrink = def.FlexShrink
		case tailwind.Justify:
			e.style.JustifyContent = def.JustifyContent
		case tailwind.AlignItems:
			e.style.AlignItems = def.AlignItems
		case tailwind.AlignSelf:
			e.style.AlignSelf = nil
		case tailwind.TextAlign:
			e.textAlign = TextAlignLeft
		case tailwind.Border:
			e.border = BorderNone
		case tailwind.BorderColor:
			e.borderStyle = Style{}
		case tailwind.Background:
			e.background = nil
		case tailwind.ScrollbarColor:
			e.scrollbarStyle = Style{}
		case tailwind.ScrollbarThumbColor:
			e.scrollbarThumbStyle = Style{}
		case tailwind.Scroll:
			e.scrollMode = ScrollNone
			e.focusable, e.tabStop = false, false
			e.scrollbarStyle, e.scrollbarThumbStyle = Style{}, Style{}
		case tailwind.OverflowHidden:
			e.overflow = OverflowVisible
		case tailwind.Focusable:
			e.focusable, e.tabStop = false, false
		case tailwind.Hidden:
			e.hidden = false
		case tailwind.Truncate:
			e.truncate = false
		case tailwind.ScrollbarHidden:
			e.scrollbarHidden = false
		case tailwind.Wrap:
			e.noWrap = false
		case tailwind.Gap:
			e.style.Gap = 0
		case tailwind.Padding, tailwind.PaddingEdges:
			e.style.Padding = Edges{}
		case tailwind.Margin, tailwind.MarginEdges:
			e.style.Margin = Edges{}
		case tailwind.Width, tailwind.WidthPercent, tailwind.WidthFraction, tailwind.WidthAuto:
			e.style.Width = def.Width
		case tailwind.Height, tailwind.HeightPercent, tailwind.HeightFraction, tailwind.HeightAuto:
			e.style.Height = def.Height
		case tailwind.MinWidth:
			e.style.MinWidth = def.MinWidth
		case tailwind.MaxWidth:
			e.style.MaxWidth = def.MaxWidth
		case tailwind.MinHeight:
			e.style.MinHeight = def.MinHeight
		case tailwind.MaxHeight:
			e.style.MaxHeight = def.MaxHeight
		case tailwind.TextGradient:
			e.textGradient = nil
		case tailwind.BackgroundGradient:
			e.bgGradient = nil
		case tailwind.BorderGradient:
			e.borderGradient = nil
		default:
			panic(fmt.Sprintf("tui: unhandled tailwind op %T", op))
		}
	}
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
