package tailwind

import "strings"

// HoverPrefix marks a class that applies only while the pointer is over the
// element (or one of its descendants), e.g. hover:bg-blue.
const HoverPrefix = "hover:"

// SplitHover separates hover: variants from the rest of a class list. hover
// holds the variant classes with the prefix removed, space-separated.
func SplitHover(classes string) (base, hover string) {
	var b, h []string
	for class := range strings.FieldsSeq(classes) {
		if inner, ok := strings.CutPrefix(class, HoverPrefix); ok {
			h = append(h, inner)
		} else {
			b = append(b, class)
		}
	}
	return strings.Join(b, " "), strings.Join(h, " ")
}

// HoverAllowed reports whether c may be used as a hover: variant. Only
// classes that restyle without changing layout qualify (colors, gradients,
// text styles), so hovering can never move an element out from under the
// pointer and flicker.
func HoverAllowed(c Class) bool {
	for _, op := range c.Ops {
		switch op.(type) {
		case BorderColor, Background, ScrollbarColor, ScrollbarThumbColor,
			TextGradient, BackgroundGradient, BorderGradient:
		default:
			return false
		}
	}
	return true
}

// ParseHover resolves the classes SplitHover returned, dropping unknown
// classes and classes HoverAllowed rejects.
func ParseHover(hover string) Result {
	var result Result
	for class := range strings.FieldsSeq(hover) {
		if c, ok := Resolve(class); ok && HoverAllowed(c) {
			result.Classes = append(result.Classes, c)
		}
	}
	return result
}
