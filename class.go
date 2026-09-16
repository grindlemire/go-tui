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
// then the new string is applied, so omitting a class removes it. A property
// that an option or setter changed after the class string was applied is left
// alone, so options set outside the class string keep their effect in either
// order.
func (e *Element) SetClass(classes string) {
	if b, a := e.classBase, e.classApplied; b != nil && a != nil {
		for _, op := range e.classOps {
			classRestore(e, b, a, op)
		}
		if e.classText && e.textStyle == a.textStyle && e.textStyleSet == a.textStyleSet {
			e.textStyle, e.textStyleSet = b.textStyle, b.textStyleSet
		}
	}
	e.classOps, e.classText, e.classBase, e.classApplied = nil, false, nil, nil
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
	e.classApplied = e.captureClassBase()
}

// classBase is the state class ops can write. It is captured before the first
// class string touches the element and again after each string is applied.
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
		style: e.style, textAlign: e.textAlign, border: e.border, borderStyle: *e.borderStyleSlot(),
		background: e.background, scrollbarStyle: e.scrollbarStyle, scrollbarThumbStyle: e.scrollbarThumbStyle,
		scrollMode: e.scrollMode, focusable: e.focusable, tabStop: e.tabStop, overflow: e.overflow,
		hidden: e.hidden, truncate: e.truncate, scrollbarHidden: e.scrollbarHidden, noWrap: e.noWrap,
		textGradient: e.textGradient, bgGradient: e.bgGradient, borderGradient: e.borderGradient,
		textStyle: e.textStyle, textStyleSet: e.textStyleSet,
	}
}

// boxFromOptions reports the border and fixed width a set of element options
// would apply, so components that size their content from their own fields
// can pick up class-derived values before rendering. Only a fixed width
// reaches the component; percent, fraction, and auto widths size the root
// element but leave the input viewport and textarea wrap width at their defaults.
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
// before the class string was applied, but only while the property still
// holds what the class wrote; a later option or setter wins. Kept next to
// classOption so the two switches stay in step;
// TestSetClass_EveryClassResetsToDefault pins it.
func classRestore(e *Element, b, a *classBase, op tailwind.Op) {
	switch op.(type) {
	case tailwind.Display:
		restore(&e.style.Display, b.style.Display, a.style.Display)
	case tailwind.Direction:
		restore(&e.style.Direction, b.style.Direction, a.style.Direction)
	case tailwind.FlexWrap:
		restore(&e.style.FlexWrap, b.style.FlexWrap, a.style.FlexWrap)
	case tailwind.AlignContent:
		restore(&e.style.AlignContent, b.style.AlignContent, a.style.AlignContent)
	case tailwind.FlexGrow:
		restore(&e.style.FlexGrow, b.style.FlexGrow, a.style.FlexGrow)
	case tailwind.FlexShrink:
		restore(&e.style.FlexShrink, b.style.FlexShrink, a.style.FlexShrink)
	case tailwind.Justify:
		restore(&e.style.JustifyContent, b.style.JustifyContent, a.style.JustifyContent)
	case tailwind.AlignItems:
		restore(&e.style.AlignItems, b.style.AlignItems, a.style.AlignItems)
	case tailwind.AlignSelf:
		restore(&e.style.AlignSelf, b.style.AlignSelf, a.style.AlignSelf)
	case tailwind.TextAlign:
		restore(&e.textAlign, b.textAlign, a.textAlign)
	case tailwind.Border:
		restore(&e.border, b.border, a.border)
	case tailwind.BorderColor:
		restore(e.borderStyleSlot(), b.borderStyle, a.borderStyle)
	case tailwind.Background:
		restore(&e.background, b.background, a.background)
	case tailwind.ScrollbarColor:
		restore(&e.scrollbarStyle, b.scrollbarStyle, a.scrollbarStyle)
	case tailwind.ScrollbarThumbColor:
		restore(&e.scrollbarThumbStyle, b.scrollbarThumbStyle, a.scrollbarThumbStyle)
	case tailwind.Scroll:
		// Mirrors what WithScrollable writes.
		restore(&e.scrollMode, b.scrollMode, a.scrollMode)
		restore(&e.focusable, b.focusable, a.focusable)
		restore(&e.scrollbarStyle, b.scrollbarStyle, a.scrollbarStyle)
		restore(&e.scrollbarThumbStyle, b.scrollbarThumbStyle, a.scrollbarThumbStyle)
	case tailwind.OverflowHidden:
		restore(&e.overflow, b.overflow, a.overflow)
	case tailwind.Focusable:
		restore(&e.focusable, b.focusable, a.focusable)
		restore(&e.tabStop, b.tabStop, a.tabStop)
	case tailwind.Hidden:
		restore(&e.hidden, b.hidden, a.hidden)
	case tailwind.Truncate:
		restore(&e.truncate, b.truncate, a.truncate)
	case tailwind.ScrollbarHidden:
		restore(&e.scrollbarHidden, b.scrollbarHidden, a.scrollbarHidden)
	case tailwind.Wrap:
		restore(&e.noWrap, b.noWrap, a.noWrap)
	case tailwind.Gap:
		restore(&e.style.Gap, b.style.Gap, a.style.Gap)
	case tailwind.Padding, tailwind.PaddingEdges:
		restore(&e.style.Padding, b.style.Padding, a.style.Padding)
	case tailwind.Margin, tailwind.MarginEdges:
		restore(&e.style.Margin, b.style.Margin, a.style.Margin)
	case tailwind.Width, tailwind.WidthPercent, tailwind.WidthFraction, tailwind.WidthAuto:
		restore(&e.style.Width, b.style.Width, a.style.Width)
	case tailwind.Height, tailwind.HeightPercent, tailwind.HeightFraction, tailwind.HeightAuto:
		restore(&e.style.Height, b.style.Height, a.style.Height)
	case tailwind.MinWidth:
		restore(&e.style.MinWidth, b.style.MinWidth, a.style.MinWidth)
	case tailwind.MaxWidth:
		restore(&e.style.MaxWidth, b.style.MaxWidth, a.style.MaxWidth)
	case tailwind.MinHeight:
		restore(&e.style.MinHeight, b.style.MinHeight, a.style.MinHeight)
	case tailwind.MaxHeight:
		restore(&e.style.MaxHeight, b.style.MaxHeight, a.style.MaxHeight)
	case tailwind.TextGradient:
		restore(&e.textGradient, b.textGradient, a.textGradient)
	case tailwind.BackgroundGradient:
		restore(&e.bgGradient, b.bgGradient, a.bgGradient)
	case tailwind.BorderGradient:
		restore(&e.borderGradient, b.borderGradient, a.borderGradient)
	default:
		panic(fmt.Sprintf("tui: unhandled tailwind op %T", op))
	}
}

// restore reverts dst to base only if it still holds the class-written value.
func restore[T comparable](dst *T, base, applied T) {
	if *dst == applied {
		*dst = base
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
