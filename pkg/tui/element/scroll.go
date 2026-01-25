package element

import "github.com/grindlemire/go-tui/pkg/layout"

// --- Scroll Query Methods ---

// IsScrollable returns whether this element has scrolling enabled.
func (e *Element) IsScrollable() bool {
	return e.scrollMode != ScrollNone
}

// ScrollMode returns the current scroll mode.
func (e *Element) ScrollModeValue() ScrollMode {
	return e.scrollMode
}

// ScrollOffset returns the current scroll position.
func (e *Element) ScrollOffset() (x, y int) {
	return e.scrollX, e.scrollY
}

// ContentSize returns the total scrollable content dimensions.
// This is computed during layout and may exceed the viewport size.
func (e *Element) ContentSize() (width, height int) {
	return e.contentWidth, e.contentHeight
}

// ViewportSize returns the visible area dimensions (content rect size).
func (e *Element) ViewportSize() (width, height int) {
	cr := e.ContentRect()
	return cr.Width, cr.Height
}

// MaxScroll returns the maximum scroll offset in each direction.
func (e *Element) MaxScroll() (maxX, maxY int) {
	vw, vh := e.ViewportSize()
	maxX = max(0, e.contentWidth-vw)
	maxY = max(0, e.contentHeight-vh)
	return
}

// --- Scroll Control Methods ---

// ScrollTo sets the scroll offset directly, clamped to valid range.
func (e *Element) ScrollTo(x, y int) {
	if e.scrollMode == ScrollNone {
		return
	}

	maxX, maxY := e.MaxScroll()

	newX := clamp(x, 0, maxX)
	newY := clamp(y, 0, maxY)

	// Only update if changed
	if newX != e.scrollX || newY != e.scrollY {
		e.scrollX = newX
		e.scrollY = newY
		e.MarkDirty()
	}
}

// ScrollBy adjusts scroll offset by delta.
func (e *Element) ScrollBy(dx, dy int) {
	e.ScrollTo(e.scrollX+dx, e.scrollY+dy)
}

// ScrollToTop scrolls to the top of the content.
func (e *Element) ScrollToTop() {
	e.ScrollTo(e.scrollX, 0)
}

// ScrollToBottom scrolls to the bottom of the content.
func (e *Element) ScrollToBottom() {
	_, maxY := e.MaxScroll()
	e.ScrollTo(e.scrollX, maxY)
}

// ScrollIntoView scrolls minimally to make the child element fully visible.
// Does nothing if the child is not a descendant of this element.
func (e *Element) ScrollIntoView(child *Element) {
	if e.scrollMode == ScrollNone {
		return
	}

	// Find child's position relative to this element's content
	childRect := child.Rect()
	contentRect := e.ContentRect()

	// Child rect is in absolute coordinates, convert to relative
	relativeX := childRect.X - contentRect.X + e.scrollX
	relativeY := childRect.Y - contentRect.Y + e.scrollY

	vw, vh := e.ViewportSize()

	// Vertical scrolling
	if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
		if relativeY < e.scrollY {
			// Child is above viewport
			e.scrollY = relativeY
		} else if relativeY+childRect.Height > e.scrollY+vh {
			// Child is below viewport
			e.scrollY = relativeY + childRect.Height - vh
		}
	}

	// Horizontal scrolling
	if e.scrollMode == ScrollHorizontal || e.scrollMode == ScrollBoth {
		if relativeX < e.scrollX {
			// Child is left of viewport
			e.scrollX = relativeX
		} else if relativeX+childRect.Width > e.scrollX+vw {
			// Child is right of viewport
			e.scrollX = relativeX + childRect.Width - vw
		}
	}

	e.MarkDirty()
}

// --- Internal Layout ---

// layoutScrollContent re-layouts children with "infinite" space in the scroll direction.
// This is called after normal layout to compute actual content size.
func (e *Element) layoutScrollContent() {
	if e.scrollMode == ScrollNone || len(e.children) == 0 {
		e.contentWidth = 0
		e.contentHeight = 0
		return
	}

	contentRect := e.ContentRect()

	// Determine available space - infinite in scroll direction
	availableWidth := contentRect.Width
	availableHeight := contentRect.Height

	switch e.scrollMode {
	case ScrollVertical:
		availableHeight = 100000 // "infinite"
	case ScrollHorizontal:
		availableWidth = 100000
	case ScrollBoth:
		availableWidth = 100000
		availableHeight = 100000
	}

	// Check if we'll need a scrollbar and reserve space
	// We do a preliminary measure to decide
	needsVScrollbar := false
	needsHScrollbar := false

	// First pass: layout with full space to measure content
	e.layoutChildrenWithSpace(availableWidth, availableHeight)
	e.measureContentBounds()

	// Check if scrollbars are needed
	if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
		needsVScrollbar = e.contentHeight > contentRect.Height
	}
	// TODO: horizontal scrollbar support
	_ = needsHScrollbar

	// If vertical scrollbar needed, reduce available width and re-layout
	if needsVScrollbar {
		availableWidth = contentRect.Width - 1
		if e.scrollMode == ScrollHorizontal || e.scrollMode == ScrollBoth {
			availableWidth = 100000
		}
		e.layoutChildrenWithSpace(availableWidth, availableHeight)
		e.measureContentBounds()
	}

	// Clamp scroll offset to valid range
	e.clampScrollOffset()
}

// layoutChildrenWithSpace lays out children within the given available space.
// Child positions are relative to (0, 0) in content space.
func (e *Element) layoutChildrenWithSpace(availableWidth, availableHeight int) {
	if len(e.children) == 0 {
		return
	}

	// Create a temporary "content" element to layout children
	// This gives us positions starting from (0, 0) in content space
	tempContainer := &Element{
		style:    e.style,
		children: e.children,
		dirty:    true,
	}
	tempContainer.style.Width = layout.Fixed(availableWidth)
	tempContainer.style.Height = layout.Fixed(availableHeight)
	// Clear padding since we're laying out in content space (already inside padding)
	tempContainer.style.Padding = layout.Edges{}

	layout.Calculate(tempContainer, availableWidth, availableHeight)
}

// measureContentBounds computes the bounding box of all children.
func (e *Element) measureContentBounds() {
	e.contentWidth = 0
	e.contentHeight = 0

	for _, child := range e.children {
		r := child.Rect()
		e.contentWidth = max(e.contentWidth, r.Right())
		e.contentHeight = max(e.contentHeight, r.Bottom())
	}
}

// clampScrollOffset ensures scroll offset is within valid range.
func (e *Element) clampScrollOffset() {
	maxX, maxY := e.MaxScroll()
	e.scrollX = clamp(e.scrollX, 0, maxX)
	e.scrollY = clamp(e.scrollY, 0, maxY)
}

// --- Internal Helpers ---

// needsVerticalScrollbar returns whether a vertical scrollbar should be shown.
func (e *Element) needsVerticalScrollbar() bool {
	if e.scrollMode != ScrollVertical && e.scrollMode != ScrollBoth {
		return false
	}
	_, vh := e.ViewportSize()
	return e.contentHeight > vh
}

// needsHorizontalScrollbar returns whether a horizontal scrollbar should be shown.
func (e *Element) needsHorizontalScrollbar() bool {
	if e.scrollMode != ScrollHorizontal && e.scrollMode != ScrollBoth {
		return false
	}
	vw, _ := e.ViewportSize()
	return e.contentWidth > vw
}

// clamp restricts v to the range [minVal, maxVal].
func clamp(v, minVal, maxVal int) int {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}
