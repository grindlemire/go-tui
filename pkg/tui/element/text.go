package element

import (
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

// TextAlign specifies how text is aligned within its content area.
type TextAlign int

const (
	// TextAlignLeft aligns text to the left edge (default).
	TextAlignLeft TextAlign = iota
	// TextAlignCenter centers text horizontally.
	TextAlignCenter
	// TextAlignRight aligns text to the right edge.
	TextAlignRight
)

// Text is an Element variant that displays text content.
// It embeds *Element and adds text-specific properties.
type Text struct {
	*Element
	content      string
	contentStyle tui.Style
	align        TextAlign
}

// NewText creates a new Text element with the given content.
// Text elements are sized to their intrinsic content size by default.
// This enables the parent's AlignItems to center the element, eliminating
// the need for separate text-level centering (and preventing jitter).
func NewText(content string, opts ...TextOption) *Text {
	t := &Text{
		Element: New(),
		content: content,
		align:   TextAlignLeft,
	}

	// Set intrinsic size based on content
	// Width = display width of text, Height = 1 (single line)
	t.Element.style.Width = layout.Fixed(stringWidth(content))
	t.Element.style.Height = layout.Fixed(1)

	// Apply options (may override the intrinsic size)
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// SetContent updates the text content and recalculates intrinsic width.
func (t *Text) SetContent(content string) {
	t.content = content
	// Update intrinsic width to match new content
	t.Element.style.Width = layout.Fixed(stringWidth(content))
	t.Element.MarkDirty()
}

// Content returns the text content.
func (t *Text) Content() string {
	return t.content
}

// ContentStyle returns the style used to render the text.
func (t *Text) ContentStyle() tui.Style {
	return t.contentStyle
}

// SetContentStyle sets the style used to render the text.
func (t *Text) SetContentStyle(style tui.Style) {
	t.contentStyle = style
}

// Align returns the text alignment.
func (t *Text) Align() TextAlign {
	return t.align
}

// SetAlign sets the text alignment.
func (t *Text) SetAlign(align TextAlign) {
	t.align = align
}

// TextOption configures a Text element.
type TextOption func(*Text)

// WithTextStyle sets the style used to render the text.
func WithTextStyle(style tui.Style) TextOption {
	return func(t *Text) {
		t.contentStyle = style
	}
}

// WithTextAlign sets the text alignment.
func WithTextAlign(align TextAlign) TextOption {
	return func(t *Text) {
		t.align = align
	}
}

// WithElementOption applies an Element option to the Text's embedded Element.
// This allows using all standard Element options on a Text.
func WithElementOption(opt Option) TextOption {
	return func(t *Text) {
		opt(t.Element)
	}
}
