package tui

// Option configures an Element.
type Option func(*Element)

// --- Dimension Options ---

// WithWidth sets a fixed width in terminal cells.
func WithWidth(cells int) Option {
	return func(e *Element) {
		e.style.Width = Fixed(cells)
	}
}

// WithWidthPercent sets width as a percentage of parent's available width.
func WithWidthPercent(percent float64) Option {
	return func(e *Element) {
		e.style.Width = Percent(percent)
	}
}

// WithHeight sets a fixed height in terminal cells.
func WithHeight(cells int) Option {
	return func(e *Element) {
		e.style.Height = Fixed(cells)
	}
}

// WithHeightPercent sets height as a percentage of parent's available height.
func WithHeightPercent(percent float64) Option {
	return func(e *Element) {
		e.style.Height = Percent(percent)
	}
}

// WithSize sets both width and height in terminal cells.
func WithSize(width, height int) Option {
	return func(e *Element) {
		e.style.Width = Fixed(width)
		e.style.Height = Fixed(height)
	}
}

// WithMinWidth sets the minimum width in terminal cells.
func WithMinWidth(cells int) Option {
	return func(e *Element) {
		e.style.MinWidth = Fixed(cells)
	}
}

// WithMinHeight sets the minimum height in terminal cells.
func WithMinHeight(cells int) Option {
	return func(e *Element) {
		e.style.MinHeight = Fixed(cells)
	}
}

// WithMaxWidth sets the maximum width in terminal cells.
func WithMaxWidth(cells int) Option {
	return func(e *Element) {
		e.style.MaxWidth = Fixed(cells)
	}
}

// WithMaxHeight sets the maximum height in terminal cells.
func WithMaxHeight(cells int) Option {
	return func(e *Element) {
		e.style.MaxHeight = Fixed(cells)
	}
}

// --- Flex Container Options ---

// WithDirection sets the main axis direction for laying out children.
func WithDirection(d Direction) Option {
	return func(e *Element) {
		e.style.Direction = d
	}
}

// WithJustify sets how children are distributed along the main axis.
func WithJustify(j Justify) Option {
	return func(e *Element) {
		e.style.JustifyContent = j
	}
}

// WithAlign sets how children are positioned on the cross axis.
func WithAlign(a Align) Option {
	return func(e *Element) {
		e.style.AlignItems = a
	}
}

// WithGap sets the space between children on the main axis.
func WithGap(cells int) Option {
	return func(e *Element) {
		e.style.Gap = cells
	}
}

// --- Flex Item Options ---

// WithFlexGrow sets how much this element should grow relative to siblings.
func WithFlexGrow(factor float64) Option {
	return func(e *Element) {
		e.style.FlexGrow = factor
	}
}

// WithFlexShrink sets how much this element should shrink relative to siblings.
func WithFlexShrink(factor float64) Option {
	return func(e *Element) {
		e.style.FlexShrink = factor
	}
}

// WithAlignSelf overrides the parent's AlignItems for this element.
func WithAlignSelf(a Align) Option {
	return func(e *Element) {
		e.style.AlignSelf = &a
	}
}

// --- Spacing Options ---

// WithPadding sets uniform padding on all sides.
func WithPadding(cells int) Option {
	return func(e *Element) {
		e.style.Padding = EdgeAll(cells)
	}
}

// WithPaddingTRBL sets padding using CSS order: Top, Right, Bottom, Left.
func WithPaddingTRBL(top, right, bottom, left int) Option {
	return func(e *Element) {
		e.style.Padding = EdgeTRBL(top, right, bottom, left)
	}
}

// WithMargin sets uniform margin on all sides.
func WithMargin(cells int) Option {
	return func(e *Element) {
		e.style.Margin = EdgeAll(cells)
	}
}

// WithMarginTRBL sets margin using CSS order: Top, Right, Bottom, Left.
func WithMarginTRBL(top, right, bottom, left int) Option {
	return func(e *Element) {
		e.style.Margin = EdgeTRBL(top, right, bottom, left)
	}
}

// --- Visual Options ---

// WithBorder sets the border style (e.g., BorderSingle, BorderRounded).
func WithBorder(style BorderStyle) Option {
	return func(e *Element) {
		e.border = style
	}
}

// WithBorderStyle sets the color/attributes for the border.
func WithBorderStyle(style Style) Option {
	return func(e *Element) {
		e.borderStyle = style
	}
}

// WithBackground sets the background style.
func WithBackground(style Style) Option {
	return func(e *Element) {
		e.background = &style
	}
}

// --- Text Options ---

// WithText sets the text content and calculates intrinsic size.
// Width is set to the text width, height is set to 1 (single line).
func WithText(content string) Option {
	return func(e *Element) {
		e.text = content
		e.style.Width = Fixed(stringWidth(content))
		e.style.Height = Fixed(1)
	}
}

// WithTextStyle sets the style for text content.
func WithTextStyle(style Style) Option {
	return func(e *Element) {
		e.textStyle = style
	}
}

// WithTextAlign sets text alignment within the content area.
func WithTextAlign(align TextAlign) Option {
	return func(e *Element) {
		e.textAlign = align
	}
}

// --- Focus Options ---

// WithOnFocus sets the callback for when this element gains focus.
// Implicitly sets focusable = true.
func WithOnFocus(fn func()) Option {
	return func(e *Element) {
		e.focusable = true
		e.onFocus = fn
	}
}

// WithOnBlur sets the callback for when this element loses focus.
// Implicitly sets focusable = true.
func WithOnBlur(fn func()) Option {
	return func(e *Element) {
		e.focusable = true
		e.onBlur = fn
	}
}

// WithOnEvent sets the event handler for this element.
// Implicitly sets focusable = true.
func WithOnEvent(fn func(Event) bool) Option {
	return func(e *Element) {
		e.focusable = true
		e.onEvent = fn
	}
}

// WithFocusable sets whether this element can receive focus.
func WithFocusable(focusable bool) Option {
	return func(e *Element) {
		e.focusable = focusable
	}
}

// --- Event Handler Options ---

// WithOnKeyPress sets the key press handler.
// No return value needed - mutations mark dirty automatically via MarkDirty().
// Implicitly sets focusable = true.
func WithOnKeyPress(fn func(KeyEvent)) Option {
	return func(e *Element) {
		e.focusable = true
		e.onKeyPress = fn
	}
}

// WithOnClick sets the click handler.
// No return value needed - mutations mark dirty automatically via MarkDirty().
// Implicitly sets focusable = true.
func WithOnClick(fn func()) Option {
	return func(e *Element) {
		e.focusable = true
		e.onClick = fn
	}
}

// --- Scroll Options ---

// WithScrollable enables scrolling in the specified mode.
// Implicitly sets focusable = true so the element can receive scroll events.
func WithScrollable(mode ScrollMode) Option {
	return func(e *Element) {
		e.scrollMode = mode
		e.focusable = true
		e.scrollbarStyle = NewStyle().Foreground(BrightBlack)
		e.scrollbarThumbStyle = NewStyle().Foreground(White)
	}
}

// WithScrollbarStyle sets the style for the scrollbar track.
func WithScrollbarStyle(style Style) Option {
	return func(e *Element) {
		e.scrollbarStyle = style
	}
}

// WithScrollbarThumbStyle sets the style for the scrollbar thumb.
func WithScrollbarThumbStyle(style Style) Option {
	return func(e *Element) {
		e.scrollbarThumbStyle = style
	}
}

// --- HR Options ---

// WithHR configures an element as a horizontal rule.
// The element renders a horizontal line character across its width.
// Uses ─ (U+2500) by default, or other characters based on border style:
//   - BorderDouble → ═ (U+2550)
//   - BorderThick  → ━ (U+2501)
//
// Sets AlignSelf to Stretch so HR fills container width regardless
// of parent's AlignItems setting.
func WithHR() Option {
	return func(e *Element) {
		e.hr = true
		e.style.Height = Fixed(1)
		stretch := AlignStretch
		e.style.AlignSelf = &stretch // Always stretch to fill width
	}
}

// --- OnUpdate Hook Options ---

// WithOnUpdate sets a function called before each render.
// Useful for polling channels, updating animations, etc.
func WithOnUpdate(fn func()) Option {
	return func(e *Element) {
		e.onUpdate = fn
	}
}
