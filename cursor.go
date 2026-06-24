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

	// Cursor cell in the element's own layout base. Under a scrollable ancestor the
	// whole subtree is laid out in that ancestor's content space (rebased toward
	// 0,0), so this is already content-local there.
	cr := e.ContentRect()
	x := cr.X + col
	y := cr.Y + row

	// Mirror the renderer: only the OUTERMOST scrollable ancestor's transform is
	// applied to the entire clipped subtree (renderClippedElement recurses into
	// descendants without re-basing at nested scrollables). Apply that one
	// transform and clip against its viewport. No scrollable ancestor means the
	// content-local position is already screen-absolute.
	var outer *Element
	for anc := e.parent; anc != nil; anc = anc.parent {
		if anc.IsScrollable() {
			outer = anc
		}
	}
	if outer == nil {
		return x, y, true
	}

	clip := outer.ContentRect()
	sx, sy := outer.ScrollOffset()
	x += clip.X - sx
	y += clip.Y - sy

	// The renderer reserves the last content column for a vertical scrollbar, so
	// shrink the clip to match before the visibility test.
	if outer.needsVerticalScrollbar() {
		clip.Width = max(0, clip.Width-1)
	}

	// A cursor scrolled outside the viewport is hidden.
	if x < clip.X || x >= clip.Right() || y < clip.Y || y >= clip.Bottom() {
		return 0, 0, false
	}
	return x, y, true
}
