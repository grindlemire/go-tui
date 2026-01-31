package element

import (
	"github.com/grindlemire/go-tui/internal/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

// RenderTree traverses the Element tree and renders to the buffer.
// This renders the element and all its descendants.
func RenderTree(buf *tui.Buffer, root *Element) {
	renderElement(buf, root)
}

// renderElement renders a single element and recurses to its children.
func renderElement(buf *tui.Buffer, e *Element) {
	// Call pre-render hook for custom update logic (polling, animations, etc.)
	if e.onUpdate != nil {
		e.onUpdate()
	}

	rect := e.Rect()

	// Skip if outside buffer bounds
	bufRect := buf.Rect()
	if !rect.Intersects(bufRect) {
		return
	}

	// Use custom render hook if set (used by wrappers)
	if e.onRender != nil {
		e.onRender(e, buf)
		return
	}

	// Handle HR specially - draws a horizontal line and returns (no children)
	if e.hr {
		renderHR(buf, e)
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

	// 4. Render children (with scroll handling if scrollable)
	if e.IsScrollable() {
		renderScrollableChildren(buf, e)
	} else {
		for _, child := range e.children {
			renderElement(buf, child)
		}
	}
}

// renderScrollableChildren renders children with scroll offset and clipping.
func renderScrollableChildren(buf *tui.Buffer, e *Element) {
	// First, do scroll-aware layout
	e.layoutScrollContent()

	// Get viewport (clip region)
	clipRect := e.ContentRect()

	// Reserve space for vertical scrollbar if needed
	if e.needsVerticalScrollbar() {
		clipRect.Width = max(0, clipRect.Width-1)
	}

	// Render each child with scroll offset and clipping
	for _, child := range e.children {
		renderClippedElement(buf, child, clipRect, e.scrollX, e.scrollY, clipRect.X, clipRect.Y)
	}

	// Draw scrollbar
	if e.needsVerticalScrollbar() {
		renderVerticalScrollbar(buf, e)
	}
}

// renderClippedElement renders an element with scroll offset and clipping.
func renderClippedElement(buf *tui.Buffer, e *Element, clipRect layout.Rect, scrollX, scrollY, viewportX, viewportY int) {
	childRect := e.Rect()

	// Translate from content space to screen space
	// Children are laid out starting from (0,0) in content space
	// We add viewport origin and subtract scroll offset
	screenX := viewportX + childRect.X - scrollX
	screenY := viewportY + childRect.Y - scrollY

	screenRect := layout.Rect{
		X:      screenX,
		Y:      screenY,
		Width:  childRect.Width,
		Height: childRect.Height,
	}

	// Check if visible within clip region
	visibleRect := screenRect.Intersect(clipRect)
	if visibleRect.IsEmpty() {
		return
	}

	// Check if fully visible (for border rendering decision)
	fullyVisible := clipRect.ContainsRect(screenRect)

	// Render background (only visible portion)
	if e.background != nil {
		buf.Fill(visibleRect, ' ', *e.background)
	}

	// Render border only if fully visible
	if e.border != tui.BorderNone && fullyVisible {
		tui.DrawBox(buf, screenRect, e.border, e.borderStyle)
	}

	// Render text with clipping
	if e.text != "" {
		textX := screenX + e.style.Padding.Left
		textY := screenY + e.style.Padding.Top

		if textY >= clipRect.Y && textY < clipRect.Bottom() {
			buf.SetStringClipped(textX, textY, e.text, e.textStyle, clipRect)
		}
	}

	// Recurse to children
	for _, child := range e.children {
		renderClippedElement(buf, child, clipRect, 0, 0, screenX, screenY)
	}
}

// renderVerticalScrollbar draws the vertical scrollbar for a scrollable element.
func renderVerticalScrollbar(buf *tui.Buffer, e *Element) {
	viewportRect := e.ContentRect()

	// Scrollbar position: right edge of content area
	trackX := viewportRect.Right() - 1
	trackTop := viewportRect.Y
	trackHeight := viewportRect.Height

	if trackHeight <= 0 {
		return
	}

	viewportHeight := viewportRect.Height
	contentHeight := e.contentHeight

	if contentHeight <= viewportHeight {
		return
	}

	// Thumb size proportional to viewport/content ratio
	thumbHeight := max(1, trackHeight*viewportHeight/contentHeight)

	// Thumb position based on scroll offset
	maxScroll := contentHeight - viewportHeight
	if maxScroll <= 0 {
		return
	}
	thumbTop := e.scrollY * (trackHeight - thumbHeight) / maxScroll

	// Draw track and thumb
	for y := 0; y < trackHeight; y++ {
		screenY := trackTop + y
		if y >= thumbTop && y < thumbTop+thumbHeight {
			buf.SetRune(trackX, screenY, '█', e.scrollbarThumbStyle)
		} else {
			buf.SetRune(trackX, screenY, '│', e.scrollbarStyle)
		}
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
// Note: onUpdate hooks are called in renderElement for each element in the tree.
func (e *Element) Render(buf *tui.Buffer, width, height int) {
	if e.dirty {
		layout.Calculate(e, width, height)
	}
	RenderTree(buf, e)
}

// hrCharacter returns the horizontal rule character based on border style.
func hrCharacter(border tui.BorderStyle) rune {
	switch border {
	case tui.BorderDouble:
		return '═' // U+2550
	case tui.BorderThick:
		return '━' // U+2501
	default:
		return '─' // U+2500
	}
}

// renderHR draws a horizontal rule across the element's width.
func renderHR(buf *tui.Buffer, e *Element) {
	rect := e.ContentRect()
	char := hrCharacter(e.border)

	for x := rect.X; x < rect.Right(); x++ {
		buf.SetRune(x, rect.Y, char, e.textStyle)
	}
}

