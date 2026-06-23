package tui

// CursorReporter is implemented by things that can report where the real
// terminal cursor should be drawn. The App queries the focused element through
// this interface at the end of each frame and drives the hardware cursor.
//
// The returned coordinates are absolute terminal cells (0-indexed, full-screen
// space; the App offsets them by the inline start row in inline mode). When
// visible is false the cursor is hidden.
type CursorReporter interface {
	// ReportCursor returns the absolute terminal cell and whether to show it.
	ReportCursor() (x, y int, visible bool)
}

// cursorSource computes the cursor position local to an element's content area:
// (col, row) measured in display cells from the content origin, and whether the
// cursor is currently visible (e.g. false when scrolled out of view). Widgets
// like TextArea and Input install one via SetCursorSource so the framework can
// place the real terminal cursor.
type cursorSource func() (col, row int, visible bool)

// Compile-time check that Element implements CursorReporter.
var _ CursorReporter = (*Element)(nil)

// SetCursorSource installs a content-local cursor source on the element. Pass
// nil to clear it. The source reports (col, row) within the element's content
// area; ReportCursor converts that to an absolute terminal cell, accounting for
// the scroll offset and clip of any scrollable ancestor.
func (e *Element) SetCursorSource(fn cursorSource) {
	e.cursorSource = fn
}

// ReportCursor implements CursorReporter. It returns the absolute terminal cell
// for the element's content-local cursor, or visible=false when no source is
// set, the source reports invisible, or the cursor is scrolled out of view in a
// scrollable ancestor.
func (e *Element) ReportCursor() (int, int, bool) {
	if e.cursorSource == nil {
		return 0, 0, false
	}
	col, row, vis := e.cursorSource()
	if !vis {
		return 0, 0, false
	}

	// Start at the cursor cell in the element's own layout base. An element under
	// a scrollable ancestor is laid out in that ancestor's content space (the
	// whole subtree is rebased toward 0,0), so the position must be translated up
	// through each scrollable ancestor exactly as the renderer does:
	//   screen = ancestorContentOrigin - ancestorScroll + posInAncestorContentSpace
	// and clipped to each ancestor's viewport (out-of-view hides the cursor).
	cr := e.ContentRect()
	x := cr.X + col
	y := cr.Y + row

	for anc := e.parent; anc != nil; anc = anc.parent {
		if !anc.IsScrollable() {
			continue
		}
		// Map (x,y) from this scrollable's content space into the scrollable's own
		// base (its parent's coordinate space), mirroring renderClippedElement.
		clip := anc.ContentRect() // viewport, expressed in the scrollable's base
		sx, sy := anc.ScrollOffset()
		x += clip.X - sx
		y += clip.Y - sy

		// Clip against the viewport in the scrollable's base. A cursor scrolled
		// out of view is hidden.
		if x < clip.X || x >= clip.Right() || y < clip.Y || y >= clip.Bottom() {
			return 0, 0, false
		}
	}

	return x, y, true
}
