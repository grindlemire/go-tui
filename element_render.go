package tui

// inheritedStyle carries cascading visual properties down the element tree.
// Text style (Fg, Attrs) and background color cascade from parent to child.
// Each field is only used when the child does not explicitly set its own value.
type inheritedStyle struct {
	textStyle Style
	bg        *Style // nil = no inherited background
}

// effectiveStyles returns the resolved text style and background for an element,
// taking inheritance into account. If the element explicitly set its textStyle,
// that is used; otherwise the inherited textStyle is used. Similarly for background.
//
// Automatic contrast: If the background is light and no explicit foreground color
// is set, the foreground is automatically set to black for readability.
func effectiveStyles(e *Element, inherited inheritedStyle) (textStyle Style, bg *Style) {
	if e.textStyleSet {
		textStyle = e.textStyle
	} else {
		textStyle = inherited.textStyle
	}

	if e.background != nil {
		bg = e.background
	} else {
		bg = inherited.bg
	}

	// Auto-contrast: if background is light and foreground is default, use black text
	if bg != nil && !bg.Bg.IsDefault() && textStyle.Fg.IsDefault() && bg.Bg.IsLight() {
		textStyle.Fg = Black
	}

	return textStyle, bg
}

// RenderTree traverses the Element tree and renders to the buffer.
// This renders the element and all its descendants.
func RenderTree(buf *Buffer, root *Element) {
	renderElement(buf, root, inheritedStyle{})
}

// renderElement renders a single element and recurses to its children.
func renderElement(buf *Buffer, e *Element, inherited inheritedStyle) {
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

	// Resolve effective styles (inheritance applied)
	textStyle, bg := effectiveStyles(e, inherited)

	// Handle HR specially - draws a horizontal line and returns (no children)
	if e.hr {
		renderHR(buf, e, textStyle)
		return
	}

	// 1. Fill background
	if e.bgGradient != nil {
		// Use gradient background (create default style if bg is nil)
		bgStyle := NewStyle()
		if bg != nil {
			bgStyle = *bg
		}
		buf.FillGradient(rect, ' ', *e.bgGradient, bgStyle)
	} else if bg != nil {
		buf.Fill(rect, ' ', *bg)
	}

	// 2. Draw border (border style does NOT inherit)
	if e.border != BorderNone {
		if e.borderGradient != nil {
			DrawBoxGradient(buf, rect, e.border, *e.borderGradient, e.borderStyle)
		} else {
			DrawBox(buf, rect, e.border, e.borderStyle)
		}
	}

	// 3. Draw text content if present
	if e.text != "" {
		renderTextContent(buf, e, textStyle, bg)
	}

	// 4. Build inherited style for children
	childInherited := inheritedStyle{
		textStyle: textStyle,
		bg:        bg,
	}

	// 5. Render children (with scroll handling if scrollable)
	if e.IsScrollable() {
		renderScrollableChildren(buf, e, childInherited)
	} else {
		for _, child := range e.children {
			renderElement(buf, child, childInherited)
		}
	}
}

// renderScrollableChildren renders children with scroll offset and clipping.
func renderScrollableChildren(buf *Buffer, e *Element, childInherited inheritedStyle) {
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
		renderClippedElement(buf, child, clipRect, e.scrollX, e.scrollY, clipRect.X, clipRect.Y, childInherited)
	}

	// Draw scrollbar (scrollbar styles are independent, not inherited)
	if e.needsVerticalScrollbar() {
		renderVerticalScrollbar(buf, e)
	}
}

// renderClippedElement renders an element with scroll offset and clipping.
func renderClippedElement(buf *Buffer, e *Element, clipRect Rect, scrollX, scrollY, viewportX, viewportY int, inherited inheritedStyle) {
	childRect := e.Rect()

	// Translate from content space to screen space
	// Children are laid out starting from (0,0) in content space
	// We add viewport origin and subtract scroll offset
	screenX := viewportX + childRect.X - scrollX
	screenY := viewportY + childRect.Y - scrollY

	screenRect := Rect{
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

	// Resolve effective styles (inheritance applied)
	textStyle, bg := effectiveStyles(e, inherited)

	// Handle HR specially - draws a horizontal line and returns (no children)
	if e.hr {
		char := hrCharacter(e.border)
		// Draw only within visible bounds
		for x := visibleRect.X; x < visibleRect.Right(); x++ {
			buf.SetRune(x, screenY, char, textStyle)
		}
		return
	}

	// Render background (only visible portion)
	if e.bgGradient != nil {
		// Use gradient background (create default style if bg is nil)
		bgStyle := NewStyle()
		if bg != nil {
			bgStyle = *bg
		}
		buf.FillGradient(visibleRect, ' ', *e.bgGradient, bgStyle)
	} else if bg != nil {
		buf.Fill(visibleRect, ' ', *bg)
	}

	// Render border clipped to viewport (border style does NOT inherit)
	if e.border != BorderNone {
		if e.borderGradient != nil {
			DrawBoxGradientClipped(buf, screenRect, e.border, *e.borderGradient, e.borderStyle, clipRect)
		} else {
			DrawBoxClipped(buf, screenRect, e.border, e.borderStyle, clipRect)
		}
	}

	// Render text with clipping
	if e.text != "" {
		textX := screenX + e.style.Padding.Left
		textY := screenY + e.style.Padding.Top
		if e.border != BorderNone {
			textX += 1
			textY += 1
		}

		if textY >= clipRect.Y && textY < clipRect.Bottom() {
			ts := textStyle
			// Merge background color into text style so text preserves the background
			if bg != nil && !bg.Bg.IsDefault() {
				ts.Bg = bg.Bg
			}
			// When the text background is unset or a text gradient is active,
			// render char-by-char to preserve existing buffer backgrounds
			// (e.g. gradient backgrounds painted by a parent) and apply clipping.
			needPerCell := e.textGradient != nil || ts.Bg.IsDefault()
			if needPerCell {
				runes := []rune(e.text)
				if len(runes) > 0 {
					curX := textX
					for i, r := range runes {
						if curX >= clipRect.Right() {
							break
						}
						if curX < clipRect.X {
							curX += RuneWidth(r)
							continue
						}
						width := RuneWidth(r)
						if width == 2 && curX+1 >= clipRect.Right() {
							break
						}
						if curX >= clipRect.X && curX < clipRect.Right() {
							style := ts
							if ts.Bg.IsDefault() {
								cellBg := buf.Cell(curX, textY).Style.Bg
								if !cellBg.IsDefault() {
									style.Bg = cellBg
								}
							}
							if e.textGradient != nil {
								t := float64(i) / float64(len(runes)-1)
								if len(runes) == 1 {
									t = 0
								}
								style.Fg = e.textGradient.At(t)
							}
							buf.SetRune(curX, textY, r, style)
						}
						curX += width
					}
				}
			} else {
				buf.SetStringClipped(textX, textY, e.text, ts, clipRect)
			}
		}
	}

	// Build inherited style for children
	childInherited := inheritedStyle{
		textStyle: textStyle,
		bg:        bg,
	}

	// Recurse to children
	// Propagate the original viewport and scroll offsets rather than re-basing
	// to the parent's screen position. Child Rect() values are absolute in the
	// temp container's coordinate space (from layoutScrollContent), so the same
	// viewport+scroll translation applies at every depth.
	for _, child := range e.children {
		renderClippedElement(buf, child, clipRect, scrollX, scrollY, viewportX, viewportY, childInherited)
	}
}

// renderVerticalScrollbar draws the vertical scrollbar for a scrollable element.
func renderVerticalScrollbar(buf *Buffer, e *Element) {
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
func renderTextContent(buf *Buffer, e *Element, textStyle Style, bg *Style) {
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

	ts := textStyle
	// Merge background color into text style so text preserves the background
	if bg != nil && !bg.Bg.IsDefault() {
		ts.Bg = bg.Bg
	}

	// When the text background is unset (default) or a text gradient is active,
	// render char-by-char so we can read each cell's existing background from
	// the buffer. This preserves gradient backgrounds painted by a parent element.
	needPerCell := e.textGradient != nil || ts.Bg.IsDefault()
	if needPerCell {
		runes := []rune(e.text)
		curX := x
		for i, r := range runes {
			if curX >= buf.Width() {
				break
			}
			width := RuneWidth(r)
			if width == 2 && curX+1 >= buf.Width() {
				break
			}
			style := ts
			if ts.Bg.IsDefault() {
				cellBg := buf.Cell(curX, contentRect.Y).Style.Bg
				if !cellBg.IsDefault() {
					style.Bg = cellBg
				}
			}
			if e.textGradient != nil {
				t := float64(i) / float64(len(runes)-1)
				if len(runes) == 1 {
					t = 0
				}
				style.Fg = e.textGradient.At(t)
			}
			buf.SetRune(curX, contentRect.Y, r, style)
			curX += width
		}
	} else {
		buf.SetString(x, contentRect.Y, e.text, ts)
	}
}

// Render calculates layout (if needed) and renders the entire tree to the buffer.
// This is the main entry point for rendering an Element tree.
// Note: onUpdate hooks are called in renderElement for each element in the tree.
func (e *Element) Render(buf *Buffer, width, height int) {
	if e.dirty {
		Calculate(e, width, height)
	}
	RenderTree(buf, e)
}

// hrCharacter returns the horizontal rule character based on border style.
func hrCharacter(border BorderStyle) rune {
	switch border {
	case BorderDouble:
		return '═' // U+2550
	case BorderThick:
		return '━' // U+2501
	default:
		return '─' // U+2500
	}
}

// renderHR draws a horizontal rule across the element's width.
func renderHR(buf *Buffer, e *Element, textStyle Style) {
	rect := e.ContentRect()
	char := hrCharacter(e.border)

	for x := rect.X; x < rect.Right(); x++ {
		buf.SetRune(x, rect.Y, char, textStyle)
	}
}
