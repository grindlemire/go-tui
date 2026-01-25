package element

import (
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

// RenderTree traverses the Element tree and renders to the buffer.
// This renders the element and all its descendants.
func RenderTree(buf *tui.Buffer, root *Element) {
	renderElement(buf, root)
}

// renderElement renders a single element and recurses to its children.
func renderElement(buf *tui.Buffer, e *Element) {
	rect := e.Rect()

	// Skip if outside buffer bounds
	bufRect := buf.Rect()
	if !rect.Intersects(bufRect) {
		return
	}

	// 1. Fill background
	if e.background != nil {
		buf.Fill(rect, ' ', *e.background)
	}

	// 2. Draw border
	if e.border != tui.BorderNone {
		tui.DrawBox(buf, rect, e.border, e.borderStyle)
	}

	// 3. Draw text content if present
	if e.text != "" {
		renderTextContent(buf, e)
	}

	// 4. Recurse to children
	for _, child := range e.children {
		renderElement(buf, child)
	}
}

// renderTextContent draws the text content within the element's content rect.
//
// When the element width equals text width (intrinsic sizing), the text is drawn
// at the content rect origin - the parent's AlignItems handles centering.
//
// When the element width is larger than text width (explicit sizing), text-level
// alignment is applied. This supports use cases like centered text in a fixed-width
// button, while avoiding jitter for intrinsic-width text in a centered layout.
func renderTextContent(buf *tui.Buffer, e *Element) {
	contentRect := e.ContentRect()

	// Skip if content rect is empty or outside buffer
	if contentRect.IsEmpty() {
		return
	}

	textWidth := stringWidth(e.text)
	x := contentRect.X

	// Only apply text-level alignment if element is wider than text content
	// (i.e., user set explicit size larger than intrinsic)
	if contentRect.Width > textWidth {
		switch e.textAlign {
		case TextAlignCenter:
			x += (contentRect.Width - textWidth) / 2
		case TextAlignRight:
			x += contentRect.Width - textWidth
		}
	}

	buf.SetString(x, contentRect.Y, e.text, e.textStyle)
}

// Render calculates layout (if needed) and renders the entire tree to the buffer.
// This is the main entry point for rendering an Element tree.
func (e *Element) Render(buf *tui.Buffer, width, height int) {
	if e.dirty {
		layout.Calculate(e, width, height)
	}
	RenderTree(buf, e)
}

