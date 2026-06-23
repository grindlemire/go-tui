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
