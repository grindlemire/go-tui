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
// captured for the element's cursor during the most recent render, or
// visible=false when no source is set, the source reported invisible, or the
// cursor was clipped out of view. The value is produced by captureCursor at the
// element's draw site, so it always matches where the renderer placed the
// content (nested scrollables, overflow-hidden clipping, and the scrollbar gutter
// included) rather than re-deriving the transform here.
func (e *Element) ReportCursor() (int, int, bool) {
	return e.cursorReport.x, e.cursorReport.y, e.cursorReport.visible
}

// captureCursor records where the element's content-local cursor lands on screen
// for ReportCursor to return. originX/originY is the element's content origin in
// screen space and clip is the visible region (the active ancestor clip already
// intersected with the element's own viewport when it is scrollable). It is
// called by the renderer at the element's draw site so the cursor uses the exact
// transform and clip the renderer applied to the element's content.
func (e *Element) captureCursor(originX, originY int, clip Rect) {
	e.cursorReport = cursorReport{}
	if e.cursorSource == nil {
		return
	}
	col, row, vis := e.cursorSource()
	if !vis {
		return
	}
	x, y := originX+col, originY+row
	if x < clip.X || x >= clip.Right() || y < clip.Y || y >= clip.Bottom() {
		return
	}
	e.cursorReport = cursorReport{x: x, y: y, visible: true}
}

// clearCursorReport marks the element's cursor hidden until the next capture. The
// renderer clears the focused element each frame so an element that no longer
// draws (hidden, or scrolled fully out of view) stops reporting a stale cursor.
func (e *Element) clearCursorReport() {
	e.cursorReport = cursorReport{}
}
