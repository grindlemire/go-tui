package tui

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

// SetText updates the text content and recalculates intrinsic width.
func (e *Element) SetText(content string) {
	e.text = content
	e.style.Width = Fixed(stringWidth(content))
	e.MarkDirty()
}

// TextStyle returns the style used to render the text.
func (e *Element) TextStyle() Style {
	return e.textStyle
}

// SetTextStyle sets the style used to render the text.
func (e *Element) SetTextStyle(style Style) {
	e.textStyle = style
}

// TextAlign returns the text alignment.
func (e *Element) TextAlign() TextAlign {
	return e.textAlign
}

// SetTextAlign sets the text alignment.
func (e *Element) SetTextAlign(align TextAlign) {
	e.textAlign = align
}

// stringWidth returns the display width of a string in terminal cells.
func stringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += RuneWidth(r)
	}
	return width
}
