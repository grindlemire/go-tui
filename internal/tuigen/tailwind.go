package tuigen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/grindlemire/go-tui/internal/tailwind"
)

// TailwindMapping is the Go source rendering of a single Tailwind class.
type TailwindMapping struct {
	Option      string // Element option code (e.g., "tui.WithDirection(tui.Column)")
	NeedsImport string // Import needed by Option/TextMethod ("tui" or "")
	IsTextStyle bool   // Whether this class modifies the text style instead
	TextMethod  string // Method to chain on tui.NewStyle() (e.g., "Bold()")
}

// TailwindParseResult contains the parsed results from a class string.
type TailwindParseResult struct {
	Options      []string        // Direct element options
	TextMethods  []string        // Text style methods to chain
	NeedsImports map[string]bool // Imports needed
}

// ParseTailwindClass renders a single class. ok is false for unknown classes.
func ParseTailwindClass(class string) (TailwindMapping, bool) {
	c, ok := tailwind.Resolve(class)
	if !ok {
		return TailwindMapping{}, false
	}
	return renderClass(c), true
}

// ParseTailwindClasses renders a full class attribute string.
func ParseTailwindClasses(classes string) TailwindParseResult {
	result := TailwindParseResult{NeedsImports: make(map[string]bool)}
	for _, c := range tailwind.Parse(classes).Classes {
		m := renderClass(c)
		if m.IsTextStyle {
			result.TextMethods = append(result.TextMethods, m.TextMethod)
		} else if m.Option != "" {
			result.Options = append(result.Options, m.Option)
		}
		if m.NeedsImport != "" {
			result.NeedsImports[m.NeedsImport] = true
		}
	}
	return result
}

// BuildTextStyleOption builds the combined text style option from accumulated methods.
func BuildTextStyleOption(methods []string) string {
	if len(methods) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("tui.WithTextStyle(tui.NewStyle()")
	for _, m := range methods {
		b.WriteString(".")
		b.WriteString(m)
	}
	b.WriteString(")")
	return b.String()
}

// renderClass renders one resolved class. A class yields either element
// options (joined on one line) or a single text style method.
func renderClass(c tailwind.Class) TailwindMapping {
	if len(c.Text) > 0 {
		method, needs := renderTextOp(c.Text[0])
		return TailwindMapping{IsTextStyle: true, TextMethod: method, NeedsImport: needs}
	}
	var opts []string
	var needs string
	for _, op := range c.Ops {
		code, n := renderOp(op)
		opts = append(opts, code)
		if n != "" {
			needs = n
		}
	}
	return TailwindMapping{Option: strings.Join(opts, ", "), NeedsImport: needs}
}

// renderTextOp renders a text style modifier as a Style method call.
func renderTextOp(op tailwind.TextOp) (code, needsImport string) {
	switch op := op.(type) {
	case tailwind.TextAttr:
		return textAttrMethods[op.Attr], ""
	case tailwind.Foreground:
		return "Foreground(" + colorExpr(op.Color) + ")", "tui"
	}
	panic(fmt.Sprintf("tuigen: unhandled text op %T", op))
}

var textAttrMethods = map[tailwind.Attr]string{
	tailwind.Bold:          "Bold()",
	tailwind.Dim:           "Dim()",
	tailwind.Italic:        "Italic()",
	tailwind.Underline:     "Underline()",
	tailwind.Blink:         "Blink()",
	tailwind.Reverse:       "Reverse()",
	tailwind.Strikethrough: "Strikethrough()",
}

// renderOp renders one element op as Go source. needsImport is "" for
// options whose arguments need no tui identifiers.
func renderOp(op tailwind.Op) (code, needsImport string) {
	tui := func(s string) (string, string) { return s, "tui" }
	plain := func(s string) (string, string) { return s, "" }
	switch op := op.(type) {
	case tailwind.Display:
		if op.Flex {
			return tui("tui.WithDisplay(tui.DisplayFlex)")
		}
		return tui("tui.WithDisplay(tui.DisplayBlock)")
	case tailwind.Direction:
		if op.Column {
			return tui("tui.WithDirection(tui.Column)")
		}
		return tui("tui.WithDirection(tui.Row)")
	case tailwind.FlexWrap:
		return tui("tui.WithFlexWrap(tui." + [...]string{"WrapNone", "Wrap", "WrapReverse"}[op.Mode] + ")")
	case tailwind.AlignContent:
		return tui("tui.WithAlignContent(tui." + [...]string{"ContentStart", "ContentEnd", "ContentCenter", "ContentStretch", "ContentSpaceBetween", "ContentSpaceAround"}[op.Value] + ")")
	case tailwind.FlexGrow:
		return plain("tui.WithFlexGrow(" + strconv.Itoa(op.N) + ")")
	case tailwind.FlexShrink:
		return plain("tui.WithFlexShrink(" + strconv.Itoa(op.N) + ")")
	case tailwind.Justify:
		return tui("tui.WithJustify(tui." + [...]string{"JustifyStart", "JustifyCenter", "JustifyEnd", "JustifySpaceBetween", "JustifySpaceAround", "JustifySpaceEvenly"}[op.Value] + ")")
	case tailwind.AlignItems:
		return tui("tui.WithAlign(tui." + alignNames[op.Value] + ")")
	case tailwind.AlignSelf:
		return tui("tui.WithAlignSelf(tui." + alignNames[op.Value] + ")")
	case tailwind.TextAlign:
		return plain("tui.WithTextAlign(tui." + [...]string{"TextAlignLeft", "TextAlignCenter", "TextAlignRight"}[op.Value] + ")")
	case tailwind.Border:
		return tui("tui.WithBorder(tui." + [...]string{"BorderSingle", "BorderRounded", "BorderDouble", "BorderThick"}[op.Style] + ")")
	case tailwind.BorderColor:
		return tui("tui.WithBorderStyle(tui.NewStyle().Foreground(" + colorExpr(op.Color) + "))")
	case tailwind.Background:
		return tui("tui.WithBackground(tui.NewStyle().Background(" + colorExpr(op.Color) + "))")
	case tailwind.ScrollbarColor:
		return tui("tui.WithScrollbarStyle(tui.NewStyle().Foreground(" + colorExpr(op.Color) + "))")
	case tailwind.ScrollbarThumbColor:
		return tui("tui.WithScrollbarThumbStyle(tui.NewStyle().Foreground(" + colorExpr(op.Color) + "))")
	case tailwind.Scroll:
		return plain("tui.WithScrollable(tui." + [...]string{"ScrollBoth", "ScrollVertical", "ScrollHorizontal"}[op.Mode] + ")")
	case tailwind.OverflowHidden:
		return tui("tui.WithOverflow(tui.OverflowHidden)")
	case tailwind.Focusable:
		return plain("tui.WithFocusable(true)")
	case tailwind.Hidden:
		return plain("tui.WithHidden(true)")
	case tailwind.Truncate:
		return plain("tui.WithTruncate(true)")
	case tailwind.ScrollbarHidden:
		return tui("tui.WithScrollbarHidden(true)")
	case tailwind.Wrap:
		return plain("tui.WithWrap(" + strconv.FormatBool(op.Enabled) + ")")
	case tailwind.Gap:
		return plain("tui.WithGap(" + strconv.Itoa(op.N) + ")")
	case tailwind.Padding:
		return plain("tui.WithPadding(" + strconv.Itoa(op.N) + ")")
	case tailwind.Margin:
		return plain("tui.WithMargin(" + strconv.Itoa(op.N) + ")")
	case tailwind.Width:
		return plain("tui.WithWidth(" + strconv.Itoa(op.N) + ")")
	case tailwind.Height:
		return plain("tui.WithHeight(" + strconv.Itoa(op.N) + ")")
	case tailwind.MinWidth:
		return plain("tui.WithMinWidth(" + strconv.Itoa(op.N) + ")")
	case tailwind.MaxWidth:
		return plain("tui.WithMaxWidth(" + strconv.Itoa(op.N) + ")")
	case tailwind.MinHeight:
		return plain("tui.WithMinHeight(" + strconv.Itoa(op.N) + ")")
	case tailwind.MaxHeight:
		return plain("tui.WithMaxHeight(" + strconv.Itoa(op.N) + ")")
	case tailwind.PaddingEdges:
		return plain(fmt.Sprintf("tui.WithPaddingTRBL(%d, %d, %d, %d)", op.Top, op.Right, op.Bottom, op.Left))
	case tailwind.MarginEdges:
		return plain(fmt.Sprintf("tui.WithMarginTRBL(%d, %d, %d, %d)", op.Top, op.Right, op.Bottom, op.Left))
	case tailwind.WidthPercent:
		return plain("tui.WithWidthPercent(" + strconv.FormatFloat(op.Percent, 'f', 2, 64) + ")")
	case tailwind.HeightPercent:
		return plain("tui.WithHeightPercent(" + strconv.FormatFloat(op.Percent, 'f', 2, 64) + ")")
	case tailwind.WidthFraction:
		// A constant expression rounds once, matching the runtime's single division.
		return plain(fmt.Sprintf("tui.WithWidthPercent(100.0 * %d / %d)", op.Num, op.Den))
	case tailwind.HeightFraction:
		return plain(fmt.Sprintf("tui.WithHeightPercent(100.0 * %d / %d)", op.Num, op.Den))
	case tailwind.WidthAuto:
		return plain("tui.WithWidthAuto()")
	case tailwind.HeightAuto:
		return plain("tui.WithHeightAuto()")
	case tailwind.TextGradient:
		return tui("tui.WithTextGradient(" + gradientExpr(op.Start, op.End, op.Direction) + ")")
	case tailwind.BackgroundGradient:
		return tui("tui.WithBackgroundGradient(" + gradientExpr(op.Start, op.End, op.Direction) + ")")
	case tailwind.BorderGradient:
		return tui("tui.WithBorderGradient(" + gradientExpr(op.Start, op.End, op.Direction) + ")")
	}
	panic(fmt.Sprintf("tuigen: unhandled tailwind op %T", op))
}

var alignNames = [...]string{"AlignStart", "AlignCenter", "AlignEnd", "AlignStretch"}

var colorIdents = [16]string{
	"Black", "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White",
	"BrightBlack", "BrightRed", "BrightGreen", "BrightYellow", "BrightBlue", "BrightMagenta", "BrightCyan", "BrightWhite",
}

// colorExpr renders a color as a tui expression.
func colorExpr(c tailwind.Color) string {
	if c.IsRGB {
		return fmt.Sprintf("tui.RGBColor(%d, %d, %d)", c.R, c.G, c.B)
	}
	return "tui." + colorIdents[c.Index]
}

func gradientExpr(start, end tailwind.Color, dir tailwind.GradientDirection) string {
	d := [...]string{"GradientHorizontal", "GradientVertical", "GradientDiagonalDown", "GradientDiagonalUp"}[dir]
	return "tui.NewGradient(" + colorExpr(start) + ", " + colorExpr(end) + ").WithDirection(tui." + d + ")"
}
