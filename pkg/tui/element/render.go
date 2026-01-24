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

	// 3. Recurse to children
	for _, child := range e.children {
		renderElement(buf, child)
	}
}

// RenderText renders a Text element to the buffer.
// It renders the embedded Element first, then draws the text content.
func RenderText(buf *tui.Buffer, t *Text) {
	// First render the element (background, border)
	renderElement(buf, t.Element)

	// Then render the text content
	renderTextContent(buf, t)
}

// renderTextContent draws the text content with proper alignment.
func renderTextContent(buf *tui.Buffer, t *Text) {
	contentRect := t.ContentRect()

	// Skip if content rect is empty or outside buffer
	if contentRect.IsEmpty() {
		return
	}

	textWidth := stringWidth(t.content)

	// Calculate x position based on alignment
	x := contentRect.X
	switch t.align {
	case TextAlignCenter:
		x += (contentRect.Width - textWidth) / 2
		if x < contentRect.X {
			x = contentRect.X
		}
	case TextAlignRight:
		x += contentRect.Width - textWidth
		if x < contentRect.X {
			x = contentRect.X
		}
	}

	// Draw the text
	buf.SetString(x, contentRect.Y, t.content, t.contentStyle)
}

// stringWidth returns the display width of a string in terminal cells.
func stringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += tui.RuneWidth(r)
	}
	return width
}

// Render calculates layout (if needed) and renders the entire tree to the buffer.
// This is the main entry point for rendering an Element tree.
func (e *Element) Render(buf *tui.Buffer, width, height int) {
	if e.dirty {
		layout.Calculate(e, width, height)
	}
	RenderTree(buf, e)
}

// RenderTextTree renders a tree that may contain Text elements.
// It checks if each element is a Text and renders accordingly.
// Note: Since Element doesn't track whether it's a Text, this function
// uses the standard element rendering. Use RenderText directly for Text elements.
func RenderTextTree(buf *tui.Buffer, root *Element, textElements map[*Element]*Text) {
	renderElementWithText(buf, root, textElements)
}

// renderElementWithText renders an element, checking if it's a Text element.
func renderElementWithText(buf *tui.Buffer, e *Element, textElements map[*Element]*Text) {
	// Check if this element is actually a Text
	if text, ok := textElements[e]; ok {
		RenderText(buf, text)
		return
	}

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

	// 3. Recurse to children
	for _, child := range e.children {
		renderElementWithText(buf, child, textElements)
	}
}
