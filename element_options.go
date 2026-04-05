package tui

import "github.com/grindlemire/go-tui/internal/debug"

// Option configures an Element.
type Option func(*Element)

// clampNonNeg returns 0 if v is negative, logging a debug warning.
func clampNonNeg(v int, name string) int {
	if v < 0 {
		debug.Log("tui: %s received negative value %d, clamping to 0", name, v)
		return 0
	}
	return v
}

// --- Dimension Options ---

// WithWidth sets a fixed width in terminal cells.
// Negative values are clamped to 0.
func WithWidth(cells int) Option {
	return func(e *Element) {
		e.style.Width = Fixed(clampNonNeg(cells, "WithWidth"))
	}
}

// WithWidthPercent sets width as a percentage of parent's available width.
// Negative values are clamped to 0.
func WithWidthPercent(percent float64) Option {
	return func(e *Element) {
		if percent < 0 {
			debug.Log("tui: WithWidthPercent received negative value %f, clamping to 0", percent)
			percent = 0
		}
		e.style.Width = Percent(percent)
	}
}

// WithHeight sets a fixed height in terminal cells.
// Negative values are clamped to 0.
func WithHeight(cells int) Option {
	return func(e *Element) {
		e.style.Height = Fixed(clampNonNeg(cells, "WithHeight"))
	}
}

// WithHeightPercent sets height as a percentage of parent's available height.
// Negative values are clamped to 0.
func WithHeightPercent(percent float64) Option {
	return func(e *Element) {
		if percent < 0 {
			debug.Log("tui: WithHeightPercent received negative value %f, clamping to 0", percent)
			percent = 0
		}
		e.style.Height = Percent(percent)
	}
}

// WithSize sets both width and height in terminal cells.
// Negative values are clamped to 0.
func WithSize(width, height int) Option {
	return func(e *Element) {
		e.style.Width = Fixed(clampNonNeg(width, "WithSize(width)"))
		e.style.Height = Fixed(clampNonNeg(height, "WithSize(height)"))
	}
}

// WithMinWidth sets the minimum width in terminal cells.
// Negative values are clamped to 0.
func WithMinWidth(cells int) Option {
	return func(e *Element) {
		e.style.MinWidth = Fixed(clampNonNeg(cells, "WithMinWidth"))
	}
}

// WithMinHeight sets the minimum height in terminal cells.
// Negative values are clamped to 0.
func WithMinHeight(cells int) Option {
	return func(e *Element) {
		e.style.MinHeight = Fixed(clampNonNeg(cells, "WithMinHeight"))
	}
}

// WithMaxWidth sets the maximum width in terminal cells.
// Negative values are clamped to 0.
func WithMaxWidth(cells int) Option {
	return func(e *Element) {
		e.style.MaxWidth = Fixed(clampNonNeg(cells, "WithMaxWidth"))
	}
}

// WithMaxHeight sets the maximum height in terminal cells.
// Negative values are clamped to 0.
func WithMaxHeight(cells int) Option {
	return func(e *Element) {
		e.style.MaxHeight = Fixed(clampNonNeg(cells, "WithMaxHeight"))
	}
}

// --- Display Options ---

// WithDisplay sets the layout mode (block or flex).
func WithDisplay(d Display) Option {
	return func(e *Element) {
		e.style.Display = d
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
// Negative values are clamped to 0.
func WithGap(cells int) Option {
	return func(e *Element) {
		e.style.Gap = clampNonNeg(cells, "WithGap")
	}
}

// WithFlexWrap sets whether flex items wrap to new lines when they overflow.
func WithFlexWrap(w FlexWrap) Option {
	return func(e *Element) {
		e.style.FlexWrap = w
	}
}

// WithAlignContent sets how flex lines are distributed along the cross axis.
// Only applies when FlexWrap is Wrap or WrapReverse.
func WithAlignContent(a AlignContent) Option {
	return func(e *Element) {
		e.style.AlignContent = a
	}
}

// --- Flex Item Options ---

// WithFlexGrow sets how much this element should grow relative to siblings.
// Negative values are clamped to 0.
func WithFlexGrow(factor float64) Option {
	return func(e *Element) {
		if factor < 0 {
			debug.Log("tui: WithFlexGrow received negative value %f, clamping to 0", factor)
			factor = 0
		}
		e.style.FlexGrow = factor
	}
}

// WithFlexShrink sets how much this element should shrink relative to siblings.
// Negative values are clamped to 0.
func WithFlexShrink(factor float64) Option {
	return func(e *Element) {
		if factor < 0 {
			debug.Log("tui: WithFlexShrink received negative value %f, clamping to 0", factor)
			factor = 0
		}
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
// Negative values are clamped to 0.
func WithPadding(cells int) Option {
	return func(e *Element) {
		e.style.Padding = EdgeAll(clampNonNeg(cells, "WithPadding"))
	}
}

// WithPaddingTRBL sets padding using CSS order: Top, Right, Bottom, Left.
// Negative values are clamped to 0.
func WithPaddingTRBL(top, right, bottom, left int) Option {
	return func(e *Element) {
		e.style.Padding = EdgeTRBL(
			clampNonNeg(top, "WithPaddingTRBL(top)"),
			clampNonNeg(right, "WithPaddingTRBL(right)"),
			clampNonNeg(bottom, "WithPaddingTRBL(bottom)"),
			clampNonNeg(left, "WithPaddingTRBL(left)"),
		)
	}
}

// WithMargin sets uniform margin on all sides.
// Negative values are clamped to 0.
func WithMargin(cells int) Option {
	return func(e *Element) {
		e.style.Margin = EdgeAll(clampNonNeg(cells, "WithMargin"))
	}
}

// WithMarginTRBL sets margin using CSS order: Top, Right, Bottom, Left.
// Negative values are clamped to 0.
func WithMarginTRBL(top, right, bottom, left int) Option {
	return func(e *Element) {
		e.style.Margin = EdgeTRBL(
			clampNonNeg(top, "WithMarginTRBL(top)"),
			clampNonNeg(right, "WithMarginTRBL(right)"),
			clampNonNeg(bottom, "WithMarginTRBL(bottom)"),
			clampNonNeg(left, "WithMarginTRBL(left)"),
		)
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

// WithText sets the text content.
// Width and Height remain Auto so the flex algorithm uses IntrinsicSize(),
// which correctly accounts for text dimensions, padding, and border.
func WithText(content string) Option {
	return func(e *Element) {
		e.text = content
	}
}

// WithTextStyle sets the style for text content.
// Setting this explicitly prevents inheritance from the parent element.
func WithTextStyle(style Style) Option {
	return func(e *Element) {
		e.textStyle = style
		e.textStyleSet = true
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
// The handler receives the element as its first parameter (self-inject).
// Implicitly sets focusable = true and tabStop = true.
func WithOnFocus(fn func(*Element)) Option {
	return func(e *Element) {
		e.focusable = true
		e.tabStop = true
		e.onFocus = fn
	}
}

// WithOnBlur sets the callback for when this element loses focus.
// The handler receives the element as its first parameter (self-inject).
// Implicitly sets focusable = true and tabStop = true.
func WithOnBlur(fn func(*Element)) Option {
	return func(e *Element) {
		e.focusable = true
		e.tabStop = true
		e.onBlur = fn
	}
}

// WithFocusable sets whether this element can receive focus and appear in Tab navigation.
func WithFocusable(focusable bool) Option {
	return func(e *Element) {
		e.focusable = focusable
		e.tabStop = focusable
	}
}

// WithAutoFocus sets whether this element should automatically receive focus
// when the element tree is first applied. Only the first autoFocus element
// in tree order takes effect. Implies focusable and tabStop.
func WithAutoFocus(auto bool) Option {
	return func(e *Element) {
		e.autoFocus = auto
		if auto {
			e.focusable = true
			e.tabStop = true
		}
	}
}

// WithTabStop sets whether this element appears in Tab/Shift+Tab navigation.
// Use this to override the default tabStop behavior set by WithFocusable or WithScrollable.
func WithTabStop(tabStop bool) Option {
	return func(e *Element) {
		e.tabStop = tabStop
	}
}

// WithOnActivate sets a callback for when this element is activated (Enter while focused).
// Implicitly sets focusable = true and tabStop = true.
func WithOnActivate(fn func()) Option {
	return func(e *Element) {
		e.focusable = true
		e.tabStop = true
		e.onActivate = fn
	}
}

// --- Scroll Options ---

// WithScrollable enables scrolling in the specified mode.
// Implicitly sets focusable = true so the element can receive scroll events,
// but does NOT set tabStop (scrollable elements are not in the Tab cycle by default).
func WithScrollable(mode ScrollMode) Option {
	return func(e *Element) {
		e.scrollMode = mode
		e.focusable = true
		e.scrollbarStyle = NewStyle().Foreground(BrightBlack)
		e.scrollbarThumbStyle = NewStyle().Foreground(White)
	}
}

// WithScrollOffset sets the initial scroll offset for a scrollable element.
// This is useful in the component model where elements are recreated each render
// and scroll state needs to be preserved via State[int].
// The offset is clamped to valid range during layout.
func WithScrollOffset(x, y int) Option {
	return func(e *Element) {
		e.scrollX = x
		e.scrollY = y
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

// --- Truncate Options ---

// WithTruncate enables text truncation with ellipsis when text overflows.
func WithTruncate(truncate bool) Option {
	return func(e *Element) {
		e.truncate = truncate
	}
}

// --- Wrap Options ---

// WithWrap sets whether text content should wrap within the element's width.
// Wrapping is enabled by default. Use WithWrap(false) or the "nowrap" class to disable.
func WithWrap(wrap bool) Option {
	return func(e *Element) {
		e.noWrap = !wrap
	}
}

// --- Hidden Options ---

// WithHidden sets whether this element is excluded from layout and rendering.
func WithHidden(hidden bool) Option {
	return func(e *Element) {
		e.hidden = hidden
	}
}

// --- Overflow Options ---

// WithOverflow sets how the element handles content that exceeds its bounds.
func WithOverflow(mode OverflowMode) Option {
	return func(e *Element) {
		e.overflow = mode
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

// --- Gradient Options ---

// WithTextGradient sets a gradient for text (overrides textStyle.Fg).
// The gradient is applied per-character horizontally by default.
func WithTextGradient(g Gradient) Option {
	return func(e *Element) {
		e.textGradient = &g
	}
}

// WithBackgroundGradient sets a gradient for the background.
// The gradient direction determines how it's applied across the element.
func WithBackgroundGradient(g Gradient) Option {
	return func(e *Element) {
		e.bgGradient = &g
	}
}

// WithBorderGradient sets a gradient for the border color.
// The gradient is applied around the perimeter of the border.
func WithBorderGradient(g Gradient) Option {
	return func(e *Element) {
		e.borderGradient = &g
	}
}

// WithOverlay marks the element as an overlay that renders in the overlay pass.
func WithOverlay(overlay bool) Option {
	return func(e *Element) {
		e.overlay = overlay
	}
}

// WithTag sets the element tag for layout dispatch.
// Used by generated code to identify table elements.
func WithTag(tag string) Option {
	return func(e *Element) {
		e.tag = tag
	}
}
