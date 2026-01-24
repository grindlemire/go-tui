package tui

// Attr represents text attributes as a bitfield for efficient comparison and storage.
type Attr uint8

const (
	// AttrNone represents no text attributes.
	AttrNone Attr = 0
	// AttrBold makes text bold/bright.
	AttrBold Attr = 1 << iota
	// AttrDim makes text dimmed/faint.
	AttrDim
	// AttrItalic makes text italic.
	AttrItalic
	// AttrUnderline underlines the text.
	AttrUnderline
	// AttrBlink makes text blink (rarely supported).
	AttrBlink
	// AttrReverse swaps foreground and background colors.
	AttrReverse
	// AttrStrikethrough draws a line through the text.
	AttrStrikethrough
)

// Style combines text attributes with foreground and background colors.
// Zero value represents default styling (no attributes, default colors).
type Style struct {
	Fg    Color
	Bg    Color
	Attrs Attr
}

// NewStyle returns a new Style with default colors and no attributes.
func NewStyle() Style {
	return Style{}
}

// Foreground returns a new Style with the given foreground color.
func (s Style) Foreground(c Color) Style {
	s.Fg = c
	return s
}

// Background returns a new Style with the given background color.
func (s Style) Background(c Color) Style {
	s.Bg = c
	return s
}

// Bold returns a new Style with the bold attribute set.
func (s Style) Bold() Style {
	s.Attrs |= AttrBold
	return s
}

// Dim returns a new Style with the dim attribute set.
func (s Style) Dim() Style {
	s.Attrs |= AttrDim
	return s
}

// Italic returns a new Style with the italic attribute set.
func (s Style) Italic() Style {
	s.Attrs |= AttrItalic
	return s
}

// Underline returns a new Style with the underline attribute set.
func (s Style) Underline() Style {
	s.Attrs |= AttrUnderline
	return s
}

// Blink returns a new Style with the blink attribute set.
func (s Style) Blink() Style {
	s.Attrs |= AttrBlink
	return s
}

// Reverse returns a new Style with the reverse attribute set.
func (s Style) Reverse() Style {
	s.Attrs |= AttrReverse
	return s
}

// Strikethrough returns a new Style with the strikethrough attribute set.
func (s Style) Strikethrough() Style {
	s.Attrs |= AttrStrikethrough
	return s
}

// Equal returns true if both styles are identical.
func (s Style) Equal(other Style) bool {
	return s.Fg.Equal(other.Fg) && s.Bg.Equal(other.Bg) && s.Attrs == other.Attrs
}

// HasAttr returns true if the style has the given attribute(s) set.
func (s Style) HasAttr(a Attr) bool {
	return s.Attrs&a == a
}
