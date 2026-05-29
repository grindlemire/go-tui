package tui

// TextSpan is a run of text sharing one style. A zero-value Style means the
// span inherits the element's resolved textStyle; set fields override it
// (attributes OR in, non-default colors replace). When an Element has rich
// text it renders as a sequence of spans that wrap together at word
// boundaries, allowing mixed styling within one wrapped paragraph.
type TextSpan struct {
	Text  string
	Style Style
	// Link is an optional OSC 8 hyperlink target. It is stored here so the type
	// is stable, but it is inert until the OSC 8 layer wires it into the cell
	// pipeline. Plain styled rendering ignores it.
	Link string
}

// WithRichText sets styled, multi-segment text on an element. When set it takes
// precedence over WithText and clears any plain text. Wrapping and alignment
// behave as for plain text.
func WithRichText(spans ...TextSpan) Option {
	return func(e *Element) {
		e.richText = spans
		e.text = ""
	}
}

// RichText returns the element's rich-text spans (nil if none).
func (e *Element) RichText() []TextSpan {
	return e.richText
}

// SetRichText replaces the element's rich-text spans and clears any plain text.
func (e *Element) SetRichText(spans ...TextSpan) {
	e.richText = spans
	e.text = ""
	e.MarkDirty()
}

// mergeSpanStyle layers a span's style over the element's resolved base style:
// attributes OR together, and a non-default span color replaces the base color.
func mergeSpanStyle(base, span Style) Style {
	out := base
	out.Attrs |= span.Attrs
	if !span.Fg.IsDefault() {
		out.Fg = span.Fg
	}
	if !span.Bg.IsDefault() {
		out.Bg = span.Bg
	}
	return out
}

// richTextWidth returns the total display width of all spans concatenated,
// used for intrinsic (unwrapped) sizing.
func richTextWidth(spans []TextSpan) int {
	w := 0
	for _, s := range spans {
		w += stringWidth(s.Text)
	}
	return w
}

// spanLineWidth returns the display width of one wrapped line of spans.
func spanLineWidth(line []TextSpan) int {
	w := 0
	for _, s := range line {
		w += stringWidth(s.Text)
	}
	return w
}
