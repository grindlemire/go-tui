package tui

import "testing"

// renderFocusedInput builds an app with the given input as an auto-focused root
// component, renders one frame, and returns the app so the test can inspect the
// MockTerminal cursor state set by placeCursor.
func renderFocusedInputApp(t *testing.T, inp *Input, termW, termH int) *App {
	t.Helper()
	app := newTestApp(termW, termH)
	inp.BindApp(app)
	app.SetRootComponent(inp)
	app.MarkDirty()
	app.Render()
	if app.Focused() == nil {
		t.Fatal("input was not focused after render")
	}
	return app
}

func TestInput_RealCursor_PlacedAtCell(t *testing.T) {
	type tc struct {
		text  string
		width int
		// cursor placed at end of text via CursorPos round-trip; expected absolute
		// column equals the text's display width.
		wantCol int
	}

	tests := map[string]tc{
		"ascii":      {text: "ab", width: 10, wantCol: 2},
		"cjk wide":   {text: "一二", width: 10, wantCol: 4},
		"flag":       {text: flagUS, width: 10, wantCol: 2},
		"zwj family": {text: zwjFamily, width: 12, wantCol: 2},
		"skin tone":  {text: skinWave, width: 10, wantCol: 2},
		"mixed wide": {text: "a" + flagUS + "b", width: 12, wantCol: 4},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := NewInput(WithInputWidth(tt.width), WithInputAutoFocus(true))
			inp.SetText(tt.text) // moves cursor to end

			app := renderFocusedInputApp(t, inp, 40, 5)
			mt := app.terminal.(*MockTerminal)

			if mt.IsCursorHidden() {
				t.Fatal("cursor should be visible for focused input")
			}
			x, y := mt.Cursor()
			if x != tt.wantCol || y != 0 {
				t.Fatalf("cursor at (%d,%d), want (%d,0)", x, y, tt.wantCol)
			}
		})
	}
}

// scrolledTextAreaRoot wraps a focused TextArea inside a short scrollable
// viewport scrolled so the cursor row falls outside the visible window.
type scrolledTextAreaRoot struct {
	ta      *TextArea
	scrollY int
}

func (r *scrolledTextAreaRoot) Render(app *App) *Element {
	scroller := New(
		WithDirection(Column),
		WithHeight(2),
		WithWidth(20),
		WithScrollable(ScrollVertical),
		WithScrollOffset(0, r.scrollY),
	)
	scroller.AddChild(r.ta.Render(app))
	return scroller
}

func (r *scrolledTextAreaRoot) BindApp(app *App) { r.ta.BindApp(app) }

func TestTextArea_RealCursor_HiddenWhenScrolledOutOfView(t *testing.T) {
	// A multi-line textarea inside a 2-row scrollable viewport. The cursor sits on
	// the last line; scrolling the viewport down by enough rows pushes the cursor
	// out of view, so the real cursor must be hidden.
	ta := NewTextArea(WithTextAreaWidth(20), WithTextAreaAutoFocus(true))
	ta.SetText("l0\nl1\nl2\nl3\nl4") // cursor lands on l4 (row 4)

	root := &scrolledTextAreaRoot{ta: ta, scrollY: 0}
	app := newTestApp(40, 10)
	root.BindApp(app)
	app.SetRootComponent(root)
	app.MarkDirty()
	app.Render()

	mt := app.terminal.(*MockTerminal)

	// With no scroll, the viewport shows rows 0..1; the cursor on row 4 is below
	// the viewport and must be hidden.
	if !mt.IsCursorHidden() {
		x, y := mt.Cursor()
		t.Fatalf("cursor should be hidden when below viewport, but shown at (%d,%d)", x, y)
	}

	// Scroll down so the cursor's row enters the viewport: it must become visible.
	root.scrollY = 3 // viewport now shows content rows 3..4
	app.MarkDirty()
	app.Render()
	if mt.IsCursorHidden() {
		t.Fatal("cursor should be visible once its row is scrolled into view")
	}
}

func TestTextArea_RealCursor_PlacedAtCell(t *testing.T) {
	type tc struct {
		text    string
		width   int
		wantCol int
		wantRow int
	}

	tests := map[string]tc{
		"ascii end":    {text: "abc", width: 20, wantCol: 3, wantRow: 0},
		"cjk wide":     {text: "一二", width: 20, wantCol: 4, wantRow: 0},
		"flag":         {text: flagUS, width: 20, wantCol: 2, wantRow: 0},
		"second line":  {text: "ab\ncd", width: 20, wantCol: 2, wantRow: 1},
		"after family": {text: zwjFamily, width: 20, wantCol: 2, wantRow: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea(WithTextAreaWidth(tt.width), WithTextAreaAutoFocus(true))
			ta.SetText(tt.text) // cursor to end

			app := newTestApp(40, 10)
			ta.BindApp(app)
			app.SetRootComponent(ta)
			app.MarkDirty()
			app.Render()

			if app.Focused() == nil {
				t.Fatal("textarea not focused")
			}
			mt := app.terminal.(*MockTerminal)
			if mt.IsCursorHidden() {
				t.Fatal("cursor should be visible for focused textarea")
			}
			x, y := mt.Cursor()
			if x != tt.wantCol || y != tt.wantRow {
				t.Fatalf("cursor at (%d,%d), want (%d,%d)", x, y, tt.wantCol, tt.wantRow)
			}
		})
	}
}

func TestApp_PlaceCursor_HiddenWhenNoFocus(t *testing.T) {
	app := newTestApp(20, 5)
	root := New()
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	mt := app.terminal.(*MockTerminal)
	if !mt.IsCursorHidden() {
		t.Fatal("cursor should be hidden when nothing focused reports a cursor")
	}
}

func TestApp_ManualCursor_SkipsPlacement(t *testing.T) {
	app := newTestApp(40, 5)
	app.manualCursor = true

	inp := NewInput(WithInputWidth(10), WithInputAutoFocus(true))
	inp.SetText("ab")
	inp.BindApp(app)
	app.SetRootComponent(inp)

	mt := app.terminal.(*MockTerminal)
	mt.HideCursor() // baseline
	mt.SetCursor(7, 7)
	app.MarkDirty()
	app.Render()

	// With manual cursor, placeCursor is a no-op: it must not move or show the
	// cursor.
	x, y := mt.Cursor()
	if x != 7 || y != 7 {
		t.Fatalf("manual cursor moved to (%d,%d), want (7,7)", x, y)
	}
	if !mt.IsCursorHidden() {
		t.Fatal("manual cursor should leave cursor hidden as set")
	}
}

func TestElement_ReportCursor_NoSource(t *testing.T) {
	e := New()
	if _, _, vis := e.ReportCursor(); vis {
		t.Fatal("element with no cursor source should report invisible")
	}
}

func TestElement_ReportCursor_AddsContentOrigin(t *testing.T) {
	app := newTestApp(40, 10)
	// A bordered child so its content rect is offset from the root origin.
	child := New(WithWidth(10), WithHeight(3), WithBorder(BorderSingle))
	child.SetCursorSource(func() (int, int, bool) { return 2, 1, true })
	root := New(WithDirection(Column))
	root.AddChild(child)
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	cr := child.ContentRect()
	x, y, vis := child.ReportCursor()
	if !vis {
		t.Fatal("expected visible")
	}
	if x != cr.X+2 || y != cr.Y+1 {
		t.Fatalf("ReportCursor = (%d,%d), want (%d,%d)", x, y, cr.X+2, cr.Y+1)
	}
}

func TestElement_ReportCursor_ScrollableAncestorShiftsAndClips(t *testing.T) {
	app := newTestApp(40, 10)

	// A scrollable viewport of height 3, scrolled down by 2 rows.
	scroller := New(WithDirection(Column), WithHeight(3), WithWidth(10), WithScrollable(ScrollVertical), WithScrollOffset(0, 2))
	// Two stacked children; the cursor source lives on the second.
	first := New(WithHeight(2), WithWidth(10))
	cursorChild := New(WithHeight(4), WithWidth(10))
	cursorChild.SetCursorSource(func() (int, int, bool) { return 1, 0, true })
	scroller.AddChild(first)
	scroller.AddChild(cursorChild)

	root := New(WithDirection(Column))
	root.AddChild(scroller)
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	// cursorChild starts at content-row 2 (after first's 2 rows). With scrollY=2,
	// it shifts to screen row 0 of the viewport. The viewport content origin is
	// the scroller's content rect.
	clip := scroller.ContentRect()
	x, y, vis := cursorChild.ReportCursor()
	if !vis {
		t.Fatalf("expected cursor visible inside viewport, got hidden")
	}
	if x != clip.X+1 || y != clip.Y+0 {
		t.Fatalf("ReportCursor = (%d,%d), want (%d,%d)", x, y, clip.X+1, clip.Y+0)
	}
}

// findGlyph scans the buffer for the first cell whose reconstructed glyph equals
// want, returning its coordinates. Used to pin cursor reporting to the position
// the renderer actually drew at, rather than to a re-derivation of the formula.
func findGlyph(b *Buffer, want string) (x, y int, found bool) {
	w, h := b.Size()
	for yy := range h {
		for xx := range w {
			if cellGlyph(b.Cell(xx, yy)) == want {
				return xx, yy, true
			}
		}
	}
	return 0, 0, false
}

func TestElement_ReportCursor_NestedScrollablesMirrorRenderer(t *testing.T) {
	app := newTestApp(40, 10)

	// Outer scrollable viewport.
	outer := New(WithDirection(Column), WithWidth(20), WithHeight(6), WithScrollable(ScrollVertical))
	// Inner scrollable with a border, so its content origin is offset within the
	// outer scrollable's content space. The renderer applies only the outer
	// scrollable's transform to the whole clipped subtree; the previous
	// per-ancestor accumulation also added this inner offset and mis-placed the
	// cursor by exactly the inner scrollable's content origin.
	middle := New(WithDirection(Column), WithWidth(16), WithHeight(8), WithScrollable(ScrollVertical), WithBorder(BorderSingle))
	cursorChild := New(WithWidth(10), WithHeight(2))
	cursorChild.SetText("Z") // marker rendered at the cursor's content origin
	cursorChild.SetCursorSource(func() (int, int, bool) { return 0, 0, true })

	middle.AddChild(cursorChild)
	outer.AddChild(middle)
	root := New(WithDirection(Column))
	root.AddChild(outer)
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	zx, zy, found := findGlyph(app.buffer, "Z")
	if !found {
		t.Fatal("cursor marker 'Z' was not rendered inside the nested viewports")
	}

	x, y, vis := cursorChild.ReportCursor()
	if !vis {
		t.Fatal("expected cursor visible inside nested viewports")
	}
	if x != zx || y != zy {
		t.Fatalf("ReportCursor = (%d,%d), but glyph rendered at (%d,%d)", x, y, zx, zy)
	}
}

func TestElement_ReportCursor_HidesCursorUnderScrollbarGutter(t *testing.T) {
	app := newTestApp(40, 10)

	// A scrollable whose content overflows vertically, so the renderer reserves
	// the last content column for the vertical scrollbar and shrinks the clip by
	// one. A cursor in that last column sits under the gutter and must be hidden.
	scroller := New(WithDirection(Column), WithWidth(10), WithHeight(3), WithScrollable(ScrollVertical))
	tall := New(WithWidth(10), WithHeight(5)) // taller than the viewport => scrollbar shows
	tall.SetCursorSource(func() (int, int, bool) { return 9, 0, true })
	scroller.AddChild(tall)

	root := New(WithDirection(Column))
	root.AddChild(scroller)
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	if !scroller.needsVerticalScrollbar() {
		t.Fatal("test setup: expected scroller to need a vertical scrollbar")
	}
	if _, _, vis := tall.ReportCursor(); vis {
		t.Fatal("cursor in the scrollbar gutter column should be reported hidden")
	}
}

func TestElement_ReportCursor_HiddenScrollbarReclaimsGutter(t *testing.T) {
	app := newTestApp(40, 10)

	// Same overflowing viewport, but with the scrollbar hidden the gutter column
	// is reclaimed: needsVerticalScrollbar is false, so the last column stays
	// inside the clip and a cursor there remains visible.
	scroller := New(WithDirection(Column), WithWidth(10), WithHeight(3), WithScrollable(ScrollVertical), WithScrollbarHidden(true))
	tall := New(WithWidth(10), WithHeight(5))
	tall.SetCursorSource(func() (int, int, bool) { return 9, 0, true })
	scroller.AddChild(tall)

	root := New(WithDirection(Column))
	root.AddChild(scroller)
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	if scroller.needsVerticalScrollbar() {
		t.Fatal("test setup: hidden scrollbar should not reserve a gutter")
	}
	if _, _, vis := tall.ReportCursor(); !vis {
		t.Fatal("cursor in the last column should be visible when the gutter is reclaimed")
	}
}

func TestElement_ReportCursor_OverflowHiddenAncestorHidesCursor(t *testing.T) {
	app := newTestApp(40, 10)

	// An overflow-hidden parent clips its children to its content box just like a
	// scrollable does, but it is not scrollable. A cursor on a child positioned
	// below the clip must be reported hidden, matching the renderer (which draws
	// nothing for that child).
	parent := New(WithDirection(Column), WithWidth(10), WithHeight(2), WithOverflow(OverflowHidden))
	inClip := New(WithWidth(10), WithHeight(2))
	inClip.SetText("A")
	below := New(WithWidth(10), WithHeight(2))
	below.SetText("Z")
	below.SetCursorSource(func() (int, int, bool) { return 0, 0, true })
	parent.AddChild(inClip)
	parent.AddChild(below)

	root := New(WithDirection(Column))
	root.AddChild(parent)
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	if _, _, found := findGlyph(app.buffer, "A"); !found {
		t.Fatal("test setup: in-clip child should render")
	}
	if _, _, found := findGlyph(app.buffer, "Z"); found {
		t.Fatal("test setup: below-clip child should be clipped out by overflow-hidden")
	}
	if _, _, vis := below.ReportCursor(); vis {
		t.Fatal("cursor clipped out by an overflow-hidden ancestor should be reported hidden")
	}
}

func TestElement_ReportCursor_OverflowHiddenInClipVisible(t *testing.T) {
	app := newTestApp(40, 10)

	// Counterpart to the hidden case: a cursor within the overflow-hidden clip is
	// reported at the cell where its glyph actually rendered.
	parent := New(WithDirection(Column), WithWidth(10), WithHeight(3), WithOverflow(OverflowHidden))
	child := New(WithWidth(10), WithHeight(2))
	child.SetText("Z")
	child.SetCursorSource(func() (int, int, bool) { return 0, 0, true })
	parent.AddChild(child)

	root := New(WithDirection(Column))
	root.AddChild(parent)
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	zx, zy, found := findGlyph(app.buffer, "Z")
	if !found {
		t.Fatal("cursor marker 'Z' should render inside the overflow-hidden clip")
	}
	x, y, vis := child.ReportCursor()
	if !vis {
		t.Fatal("cursor inside the overflow-hidden clip should be visible")
	}
	if x != zx || y != zy {
		t.Fatalf("ReportCursor = (%d,%d), but glyph rendered at (%d,%d)", x, y, zx, zy)
	}
}

func TestElement_ReportCursor_SelfScrollableClipsOwnCursor(t *testing.T) {
	app := newTestApp(40, 10)

	// A cursor source installed directly on a scrollable element must be clipped to
	// that element's own viewport. A row past the viewport height is out of view.
	scroller := New(WithDirection(Column), WithWidth(10), WithHeight(2), WithScrollable(ScrollVertical))
	scroller.SetCursorSource(func() (int, int, bool) { return 0, 5, true }) // row 5, viewport is 2 tall

	root := New(WithDirection(Column))
	root.AddChild(scroller)
	app.SetRoot(root)
	app.MarkDirty()
	app.Render()

	if _, _, vis := scroller.ReportCursor(); vis {
		t.Fatal("cursor past the scrollable's own viewport should be reported hidden")
	}
}

func TestApp_RenderFull_ResetsCursorForElementThatStopsDrawing(t *testing.T) {
	app := newTestApp(40, 10)

	// A persistent (SetRoot) focused element with a cursor source. After it stops
	// drawing, a full redraw must not leave a stale cursor on screen.
	cursorEl := New(WithWidth(10), WithHeight(2), WithFocusable(true))
	cursorEl.SetText("X")
	cursorEl.SetCursorSource(func() (int, int, bool) { return 0, 0, true })
	root := New(WithDirection(Column))
	root.AddChild(cursorEl)
	app.SetRoot(root)
	app.focus.Register(cursorEl)
	app.focus.SetFocus(cursorEl)

	app.MarkDirty()
	app.Render()

	mt := app.terminal.(*MockTerminal)
	if mt.IsCursorHidden() {
		t.Fatal("setup: cursor should be visible while the focused element draws")
	}

	cursorEl.SetHidden(true)
	app.RenderFull()
	if !mt.IsCursorHidden() {
		x, y := mt.Cursor()
		t.Fatalf("cursor should be hidden after the focused element stops drawing, shown at (%d,%d)", x, y)
	}
}
