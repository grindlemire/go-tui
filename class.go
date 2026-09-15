package tui

import (
	"fmt"

	"github.com/grindlemire/go-tui/internal/layout"
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
// string set is restored to the value it had before that string was applied,
// then the new string is applied, so omitting a class removes it while options
// set outside the class string keep their effect.
func (e *Element) SetClass(classes string) {
	if b := e.classBase; b != nil {
		for _, op := range e.classOps {
			classRestore(e, b, op)
		}
		if e.classText {
			e.textStyle, e.textStyleSet = b.textStyle, b.textStyleSet
		}
	}
	e.classOps, e.classText, e.classBase = nil, false, nil
	e.applyClasses(classes)
	e.MarkDirty()
}

func (e *Element) applyClasses(classes string) {
	if e.classBase == nil {
		e.classBase = e.captureClassBase()
	}
	result := tailwind.Parse(classes)
	ops := result.Ops()
	for _, op := range ops {
		e.Apply(classOption(op))
	}
	e.classOps = append(e.classOps, ops...)
	if text := result.Text(); len(text) > 0 {
		e.Apply(WithTextStyle(classTextStyle(text)))
		e.classText = true
	}
}

// classBase is the state class ops can write, captured before the first class
// string touches the element.
type classBase struct {
	style                                     LayoutStyle
	textAlign                                 TextAlign
	border                                    BorderStyle
	borderStyle                               Style
	background                                *Style
	scrollbarStyle, scrollbarThumbStyle       Style
	scrollMode                                ScrollMode
	focusable, tabStop                        bool
	overflow                                  OverflowMode
	hidden, truncate, scrollbarHidden, noWrap bool
	textGradient, bgGradient, borderGradient  *Gradient
	textStyle                                 Style
	textStyleSet                              bool
}

func (e *Element) captureClassBase() *classBase {
	return &classBase{
		style: e.style, textAlign: e.textAlign, border: e.border, borderStyle: e.borderStyle,
		background: e.background, scrollbarStyle: e.scrollbarStyle, scrollbarThumbStyle: e.scrollbarThumbStyle,
		scrollMode: e.scrollMode, focusable: e.focusable, tabStop: e.tabStop, overflow: e.overflow,
		hidden: e.hidden, truncate: e.truncate, scrollbarHidden: e.scrollbarHidden, noWrap: e.noWrap,
		textGradient: e.textGradient, bgGradient: e.bgGradient, borderGradient: e.borderGradient,
		textStyle: e.textStyle, textStyleSet: e.textStyleSet,
	}
}

// boxFromOptions reports the border and fixed width a set of element options
// would apply, so components that size their content from their own fields
// can pick up class-derived values before rendering.
func boxFromOptions(opts []Option) (border BorderStyle, width int) {
	if len(opts) == 0 {
		return BorderNone, 0
	}
	probe := New(opts...)
	if w := probe.LayoutStyle().Width; w.Unit == layout.UnitFixed {
		width = int(w.Amount)
	}
	return probe.Border(), width
}

// classRestore puts back the property op overwrote, from the snapshot taken
// before the class string was applied. Kept next to classOption so the two
// switches stay in step; TestSetClass_EveryClassResetsToDefault pins it.
func classRestore(e *Element, b *classBase, op tailwind.Op) {
	switch op.(type) {
	case tailwind.Display:
		e.style.Display = b.style.Display
	case tailwind.Direction:
		e.style.Direction = b.style.Direction
	case tailwind.FlexWrap:
		e.style.FlexWrap = b.style.FlexWrap
	case tailwind.AlignContent:
		e.style.AlignContent = b.style.AlignContent
	case tailwind.FlexGrow:
		e.style.FlexGrow = b.style.FlexGrow
	case tailwind.FlexShrink:
		e.style.FlexShrink = b.style.FlexShrink
	case tailwind.Justify:
		e.style.JustifyContent = b.style.JustifyContent
	case tailwind.AlignItems:
		e.style.AlignItems = b.style.AlignItems
	case tailwind.AlignSelf:
		e.style.AlignSelf = b.style.AlignSelf
	case tailwind.TextAlign:
		e.textAlign = b.textAlign
	case tailwind.Border:
		e.border = b.border
	case tailwind.BorderColor:
		e.borderStyle = b.borderStyle
	case tailwind.Background:
		e.background = b.background
	case tailwind.ScrollbarColor:
		e.scrollbarStyle = b.scrollbarStyle
	case tailwind.ScrollbarThumbColor:
		e.scrollbarThumbStyle = b.scrollbarThumbStyle
	case tailwind.Scroll:
		// Mirrors what WithScrollable writes.
		e.scrollMode, e.focusable = b.scrollMode, b.focusable
		e.scrollbarStyle, e.scrollbarThumbStyle = b.scrollbarStyle, b.scrollbarThumbStyle
	case tailwind.OverflowHidden:
		e.overflow = b.overflow
	case tailwind.Focusable:
		e.focusable, e.tabStop = b.focusable, b.tabStop
	case tailwind.Hidden:
		e.hidden = b.hidden
	case tailwind.Truncate:
		e.truncate = b.truncate
	case tailwind.ScrollbarHidden:
		e.scrollbarHidden = b.scrollbarHidden
	case tailwind.Wrap:
		e.noWrap = b.noWrap
	case tailwind.Gap:
		e.style.Gap = b.style.Gap
	case tailwind.Padding, tailwind.PaddingEdges:
		e.style.Padding = b.style.Padding
	case tailwind.Margin, tailwind.MarginEdges:
		e.style.Margin = b.style.Margin
	case tailwind.Width, tailwind.WidthPercent, tailwind.WidthFraction, tailwind.WidthAuto:
		e.style.Width = b.style.Width
	case tailwind.Height, tailwind.HeightPercent, tailwind.HeightFraction, tailwind.HeightAuto:
		e.style.Height = b.style.Height
	case tailwind.MinWidth:
		e.style.MinWidth = b.style.MinWidth
	case tailwind.MaxWidth:
		e.style.MaxWidth = b.style.MaxWidth
	case tailwind.MinHeight:
		e.style.MinHeight = b.style.MinHeight
	case tailwind.MaxHeight:
		e.style.MaxHeight = b.style.MaxHeight
	case tailwind.TextGradient:
		e.textGradient = b.textGradient
	case tailwind.BackgroundGradient:
		e.bgGradient = b.bgGradient
	case tailwind.BorderGradient:
		e.borderGradient = b.borderGradient
	default:
		panic(fmt.Sprintf("tui: unhandled tailwind op %T", op))
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
