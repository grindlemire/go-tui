// Package tailwind resolves Tailwind-style utility classes into a typed
// intermediate form shared by the .gsx compiler (which renders it to Go
// source) and the runtime (which applies it to an Element). It has no
// dependency on the root tui package so both can import it.
package tailwind

// Op is one element option produced by a class. Each concrete type below
// corresponds to a With* option in the root package. Adding a type means a
// case in tuigen.renderOp and tui.classOption; TestRenderCoversEveryClass and
// TestWithClass_CoversEveryClass catch a missing case.
type Op interface{ op() }

// TextOp is one modifier chained onto the element's text style.
type TextOp interface{ textOp() }

// Display sets the display mode: block (default) or flex.
type Display struct{ Flex bool }

// Direction sets the flex direction: row (default) or column.
type Direction struct{ Column bool }

// WrapMode is a flex-wrap value.
type WrapMode int

// Flex wrap modes.
const (
	WrapNone WrapMode = iota
	WrapOn
	WrapReverse
)

// FlexWrap sets the flex-wrap mode.
type FlexWrap struct{ Mode WrapMode }

// ContentValue is an align-content value.
type ContentValue int

// Align-content values.
const (
	ContentStart ContentValue = iota
	ContentEnd
	ContentCenter
	ContentStretch
	ContentSpaceBetween
	ContentSpaceAround
)

// AlignContent sets cross-axis line distribution for wrapped flex.
type AlignContent struct{ Value ContentValue }

// FlexGrow sets the flex-grow factor.
type FlexGrow struct{ N int }

// FlexShrink sets the flex-shrink factor.
type FlexShrink struct{ N int }

// JustifyValue is a justify-content value.
type JustifyValue int

// Justify-content values.
const (
	JustifyStart JustifyValue = iota
	JustifyCenter
	JustifyEnd
	SpaceBetween
	SpaceAround
	SpaceEvenly
)

// Justify sets main-axis alignment.
type Justify struct{ Value JustifyValue }

// AlignValue is an align-items or align-self value.
type AlignValue int

// Align values.
const (
	AlignStart AlignValue = iota
	AlignCenter
	AlignEnd
	AlignStretch
)

// AlignItems sets cross-axis alignment of children.
type AlignItems struct{ Value AlignValue }

// AlignSelf overrides the parent's align-items for this element.
type AlignSelf struct{ Value AlignValue }

// TextAlignValue is a text alignment.
type TextAlignValue int

// Text alignments.
const (
	TextLeft TextAlignValue = iota
	TextCenter
	TextRight
)

// TextAlign sets text alignment.
type TextAlign struct{ Value TextAlignValue }

// BorderStyle is a border line style.
type BorderStyle int

// Border styles.
const (
	BorderSingle BorderStyle = iota
	BorderRounded
	BorderDouble
	BorderThick
)

// Border sets the border style.
type Border struct{ Style BorderStyle }

// BorderColor sets the border foreground color.
type BorderColor struct{ Color Color }

// Background sets the background color.
type Background struct{ Color Color }

// ScrollbarColor sets the scrollbar track color.
type ScrollbarColor struct{ Color Color }

// ScrollbarThumbColor sets the scrollbar thumb color.
type ScrollbarThumbColor struct{ Color Color }

// ScrollMode is a scroll direction set.
type ScrollMode int

// Scroll modes.
const (
	ScrollBoth ScrollMode = iota
	ScrollVertical
	ScrollHorizontal
)

// Scroll enables scrolling.
type Scroll struct{ Mode ScrollMode }

// OverflowHidden clips children without scrollbars.
type OverflowHidden struct{}

// Focusable makes the element focusable.
type Focusable struct{}

// Hidden removes the element from layout and rendering.
type Hidden struct{}

// Truncate truncates overflowing text with an ellipsis.
type Truncate struct{}

// ScrollbarHidden hides the scrollbar and reclaims its gutter.
type ScrollbarHidden struct{}

// Wrap enables or disables text wrapping.
type Wrap struct{ Enabled bool }

// Gap sets the gap between children.
type Gap struct{ N int }

// Padding sets padding on all sides.
type Padding struct{ N int }

// Margin sets margin on all sides.
type Margin struct{ N int }

// Width sets a fixed width.
type Width struct{ N int }

// Height sets a fixed height.
type Height struct{ N int }

// MinWidth sets the minimum width.
type MinWidth struct{ N int }

// MaxWidth sets the maximum width.
type MaxWidth struct{ N int }

// MinHeight sets the minimum height.
type MinHeight struct{ N int }

// MaxHeight sets the maximum height.
type MaxHeight struct{ N int }

// PaddingEdges sets per-side padding, accumulated from pt-/pr-/pb-/pl-/px-/py- classes.
type PaddingEdges struct{ Top, Right, Bottom, Left int }

// MarginEdges sets per-side margin, accumulated from mt-/mr-/mb-/ml-/mx-/my- classes.
type MarginEdges struct{ Top, Right, Bottom, Left int }

// WidthPercent sets width as a percentage of the parent.
type WidthPercent struct{ Percent float64 }

// HeightPercent sets height as a percentage of the parent.
type HeightPercent struct{ Percent float64 }

// WidthAuto sizes width to content.
type WidthAuto struct{}

// HeightAuto sizes height to content.
type HeightAuto struct{}

// GradientDirection is the axis a gradient runs along.
type GradientDirection int

// Gradient directions.
const (
	Horizontal GradientDirection = iota
	Vertical
	DiagonalDown
	DiagonalUp
)

// TextGradient applies a gradient to text.
type TextGradient struct {
	Start, End Color
	Direction  GradientDirection
}

// BackgroundGradient applies a gradient to the background.
type BackgroundGradient struct {
	Start, End Color
	Direction  GradientDirection
}

// BorderGradient applies a gradient to the border.
type BorderGradient struct {
	Start, End Color
	Direction  GradientDirection
}

// Attr is a text attribute flag.
type Attr int

// Text attributes.
const (
	Bold Attr = iota
	Dim
	Italic
	Underline
	Blink
	Reverse
	Strikethrough
)

// TextAttr enables one text attribute.
type TextAttr struct{ Attr Attr }

// Foreground sets the text color.
type Foreground struct{ Color Color }

func (Display) op()             {}
func (Direction) op()           {}
func (FlexWrap) op()            {}
func (AlignContent) op()        {}
func (FlexGrow) op()            {}
func (FlexShrink) op()          {}
func (Justify) op()             {}
func (AlignItems) op()          {}
func (AlignSelf) op()           {}
func (TextAlign) op()           {}
func (Border) op()              {}
func (BorderColor) op()         {}
func (Background) op()          {}
func (ScrollbarColor) op()      {}
func (ScrollbarThumbColor) op() {}
func (Scroll) op()              {}
func (OverflowHidden) op()      {}
func (Focusable) op()           {}
func (Hidden) op()              {}
func (Truncate) op()            {}
func (ScrollbarHidden) op()     {}
func (Wrap) op()                {}
func (Gap) op()                 {}
func (Padding) op()             {}
func (Margin) op()              {}
func (Width) op()               {}
func (Height) op()              {}
func (MinWidth) op()            {}
func (MaxWidth) op()            {}
func (MinHeight) op()           {}
func (MaxHeight) op()           {}
func (PaddingEdges) op()        {}
func (MarginEdges) op()         {}
func (WidthPercent) op()        {}
func (HeightPercent) op()       {}
func (WidthAuto) op()           {}
func (HeightAuto) op()          {}
func (TextGradient) op()        {}
func (BackgroundGradient) op()  {}
func (BorderGradient) op()      {}

func (TextAttr) textOp()   {}
func (Foreground) textOp() {}

// Color is a terminal color: one of the 16 named ANSI colors or 24-bit RGB.
type Color struct {
	IsRGB   bool
	Index   uint8 // ANSI-16 index when !IsRGB
	R, G, B uint8 // components when IsRGB
}

// RGB returns a 24-bit color.
func RGB(r, g, b uint8) Color { return Color{IsRGB: true, R: r, G: g, B: b} }

// ColorNames lists the 16 named colors in ANSI index order.
var ColorNames = [16]string{
	"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
	"bright-black", "bright-red", "bright-green", "bright-yellow", "bright-blue", "bright-magenta", "bright-cyan", "bright-white",
}

// Named returns the ANSI color for a class color name. Unknown names fall
// back to black, matching the generator's historical behavior.
func Named(name string) Color {
	for i, n := range ColorNames {
		if n == name {
			return Color{Index: uint8(i)}
		}
	}
	return Color{}
}

// ParseHex parses a 3 or 6 digit hex color (without the leading '#').
func ParseHex(hex string) (Color, bool) {
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return Color{}, false
	}
	var out [3]uint8
	for i := range out {
		hi, ok1 := hexNibble(hex[2*i])
		lo, ok2 := hexNibble(hex[2*i+1])
		if !ok1 || !ok2 {
			return Color{}, false
		}
		out[i] = hi<<4 | lo
	}
	return RGB(out[0], out[1], out[2]), true
}

func hexNibble(c byte) (uint8, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}
