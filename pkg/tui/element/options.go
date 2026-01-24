package element

import (
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

// Option configures an Element.
type Option func(*Element)

// --- Dimension Options ---

// WithWidth sets a fixed width in terminal cells.
func WithWidth(cells int) Option {
	return func(e *Element) {
		e.style.Width = layout.Fixed(cells)
	}
}

// WithWidthPercent sets width as a percentage of parent's available width.
func WithWidthPercent(percent float64) Option {
	return func(e *Element) {
		e.style.Width = layout.Percent(percent)
	}
}

// WithHeight sets a fixed height in terminal cells.
func WithHeight(cells int) Option {
	return func(e *Element) {
		e.style.Height = layout.Fixed(cells)
	}
}

// WithHeightPercent sets height as a percentage of parent's available height.
func WithHeightPercent(percent float64) Option {
	return func(e *Element) {
		e.style.Height = layout.Percent(percent)
	}
}

// WithSize sets both width and height in terminal cells.
func WithSize(width, height int) Option {
	return func(e *Element) {
		e.style.Width = layout.Fixed(width)
		e.style.Height = layout.Fixed(height)
	}
}

// WithMinWidth sets the minimum width in terminal cells.
func WithMinWidth(cells int) Option {
	return func(e *Element) {
		e.style.MinWidth = layout.Fixed(cells)
	}
}

// WithMinHeight sets the minimum height in terminal cells.
func WithMinHeight(cells int) Option {
	return func(e *Element) {
		e.style.MinHeight = layout.Fixed(cells)
	}
}

// WithMaxWidth sets the maximum width in terminal cells.
func WithMaxWidth(cells int) Option {
	return func(e *Element) {
		e.style.MaxWidth = layout.Fixed(cells)
	}
}

// WithMaxHeight sets the maximum height in terminal cells.
func WithMaxHeight(cells int) Option {
	return func(e *Element) {
		e.style.MaxHeight = layout.Fixed(cells)
	}
}

// --- Flex Container Options ---

// WithDirection sets the main axis direction for laying out children.
func WithDirection(d layout.Direction) Option {
	return func(e *Element) {
		e.style.Direction = d
	}
}

// WithJustify sets how children are distributed along the main axis.
func WithJustify(j layout.Justify) Option {
	return func(e *Element) {
		e.style.JustifyContent = j
	}
}

// WithAlign sets how children are positioned on the cross axis.
func WithAlign(a layout.Align) Option {
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
func WithAlignSelf(a layout.Align) Option {
	return func(e *Element) {
		e.style.AlignSelf = &a
	}
}

// --- Spacing Options ---

// WithPadding sets uniform padding on all sides.
func WithPadding(cells int) Option {
	return func(e *Element) {
		e.style.Padding = layout.EdgeAll(cells)
	}
}

// WithPaddingTRBL sets padding using CSS order: Top, Right, Bottom, Left.
func WithPaddingTRBL(top, right, bottom, left int) Option {
	return func(e *Element) {
		e.style.Padding = layout.EdgeTRBL(top, right, bottom, left)
	}
}

// WithMargin sets uniform margin on all sides.
func WithMargin(cells int) Option {
	return func(e *Element) {
		e.style.Margin = layout.EdgeAll(cells)
	}
}

// WithMarginTRBL sets margin using CSS order: Top, Right, Bottom, Left.
func WithMarginTRBL(top, right, bottom, left int) Option {
	return func(e *Element) {
		e.style.Margin = layout.EdgeTRBL(top, right, bottom, left)
	}
}

// --- Visual Options ---

// WithBorder sets the border style (e.g., BorderSingle, BorderRounded).
func WithBorder(style tui.BorderStyle) Option {
	return func(e *Element) {
		e.border = style
	}
}

// WithBorderStyle sets the color/attributes for the border.
func WithBorderStyle(style tui.Style) Option {
	return func(e *Element) {
		e.borderStyle = style
	}
}

// WithBackground sets the background style.
func WithBackground(style tui.Style) Option {
	return func(e *Element) {
		e.background = &style
	}
}
