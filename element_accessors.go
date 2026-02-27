package tui

// GetRoot implements Viewable. Returns the element itself as a Renderable.
func (e *Element) GetRoot() Renderable { return e }

// GetWatchers implements Viewable. Elements have no standalone watchers.
func (e *Element) GetWatchers() []Watcher { return nil }

// SetStyle updates the layout style and marks the element dirty.
func (e *Element) SetStyle(style LayoutStyle) {
	e.style = style
	e.MarkDirty()
}

// Style returns the current layout style.
func (e *Element) Style() LayoutStyle {
	return e.style
}

// Border returns the border style.
func (e *Element) Border() BorderStyle {
	return e.border
}

// SetBorder sets the border style.
func (e *Element) SetBorder(border BorderStyle) {
	e.border = border
}

// BorderStyle returns the style used to render the border.
func (e *Element) BorderStyle() Style {
	return e.borderStyle
}

// SetBorderStyle sets the style used to render the border.
func (e *Element) SetBorderStyle(style Style) {
	e.borderStyle = style
}

// Background returns the background style, or nil if transparent.
func (e *Element) Background() *Style {
	return e.background
}

// SetBackground sets the background style. Pass nil for transparent.
func (e *Element) SetBackground(style *Style) {
	e.background = style
}

// --- Text API ---

// Text returns the text content.
func (e *Element) Text() string {
	return e.text
}

// SetText updates the text content.
// Width remains Auto so the flex algorithm uses IntrinsicSize(),
// which correctly accounts for text dimensions, padding, and border.
func (e *Element) SetText(content string) {
	e.text = content
	e.MarkDirty()
}

// TextStyle returns the style used to render the text.
func (e *Element) TextStyle() Style {
	return e.textStyle
}

// SetTextStyle sets the style used to render the text.
// Setting this explicitly prevents inheritance from the parent element.
func (e *Element) SetTextStyle(style Style) {
	e.textStyle = style
	e.textStyleSet = true
}

// TextAlign returns the text alignment.
func (e *Element) TextAlign() TextAlign {
	return e.textAlign
}

// SetTextAlign sets the text alignment.
func (e *Element) SetTextAlign(align TextAlign) {
	e.textAlign = align
}

// --- Truncate API ---

// Truncate returns whether text truncation is enabled.
func (e *Element) Truncate() bool {
	return e.truncate
}

// SetTruncate sets whether text should be truncated with ellipsis on overflow.
func (e *Element) SetTruncate(truncate bool) {
	e.truncate = truncate
	e.MarkDirty()
}

// --- Wrap API ---

// wrapsText returns true if this element should wrap text content.
func (e *Element) wrapsText() bool {
	return !e.noWrap
}

// --- Hidden API ---

// Hidden returns whether this element is hidden.
func (e *Element) Hidden() bool {
	return e.hidden
}

// SetHidden sets whether this element is excluded from layout and rendering.
func (e *Element) SetHidden(hidden bool) {
	e.hidden = hidden
	e.MarkDirty()
}

// --- Overflow API ---

// Overflow returns the overflow mode.
func (e *Element) Overflow() OverflowMode {
	return e.overflow
}

// SetOverflow sets the overflow mode.
func (e *Element) SetOverflow(mode OverflowMode) {
	e.overflow = mode
	e.MarkDirty()
}

// stringWidth returns the display width of a string in terminal cells.
func stringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += RuneWidth(r)
	}
	return width
}
