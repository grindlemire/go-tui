package tui

import (
	"fmt"
	"slices"
	"testing"
)

// hoverTestRoot rebuilds its whole tree every render, like generated code, and
// sets state from its hover handlers so each enter/leave triggers a re-render.
type hoverTestRoot struct {
	hovered  *State[string]
	log      []string
	modal    *Modal
	rows     []string
	keyed    bool
	rowClass string
	onEnter  func(name string)
	mouse    []MouseEvent
}

func (r *hoverTestRoot) HandleMouse(me MouseEvent) bool {
	r.mouse = append(r.mouse, MouseEvent{Button: me.Button, Action: me.Action, X: me.X, Y: me.Y})
	return false
}

func (r *hoverTestRoot) track(name string) []Option {
	return []Option{
		WithOnMouseEnter(func() {
			r.log = append(r.log, "enter "+name)
			r.hovered.Set(name)
			if r.onEnter != nil {
				r.onEnter(name)
			}
		}),
		WithOnMouseLeave(func() {
			r.log = append(r.log, "leave "+name)
		}),
	}
}

func (r *hoverTestRoot) Render(app *App) *Element {
	root := New(append(r.track("root"), WithDirection(Column), WithWidth(20), WithHeight(4))...)
	rows := r.rows
	if rows == nil {
		rows = []string{"a", "b"}
	}
	for _, name := range rows {
		opts := append(r.track(name), WithWidth(20), WithHeight(1), WithText(name))
		if r.keyed {
			opts = append(opts, WithKey(name))
		}
		if r.rowClass != "" {
			opts = append(opts, WithClass(r.rowClass))
		}
		root.AddChild(New(opts...))
	}
	if r.modal != nil {
		overlay := r.modal.Render(app)
		overlay.AddChild(New(append(r.track("dialog"), WithWidth(6), WithHeight(1))...))
		root.AddChild(overlay)
	}
	return root
}

func newHoverTestApp(root *hoverTestRoot) *App {
	app := newTestApp(20, 4)
	app.mouseMotion = true
	root.hovered = NewState("")
	root.hovered.BindApp(app)
	if root.modal != nil {
		root.modal.BindApp(app)
	}
	app.rootComponent = root
	app.MarkDirty()
	app.Render()
	return app
}

// moveTo dispatches a no-button move, renders, and returns the hover callbacks
// that fired.
func moveTo(app *App, root *hoverTestRoot, x, y int) []string {
	root.log = nil
	app.Dispatch(MouseEvent{Button: MouseNone, Action: MouseMove, X: x, Y: y})
	app.Render()
	return root.log
}

func TestHover_EnterLeaveAcrossRerenders(t *testing.T) {
	root := &hoverTestRoot{}
	app := newHoverTestApp(root)

	steps := []struct {
		x, y int
		want []string
	}{
		// Entering fires outermost first.
		{1, 0, []string{"enter root", "enter a"}},
		// The enter handler set state, so the tree was rebuilt. Moving within
		// the rebuilt row must not re-fire enter/leave.
		{5, 0, nil},
		// Moving between siblings leaves one and enters the other; the shared
		// parent stays hovered.
		{5, 1, []string{"leave a", "enter b"}},
		// Off the rows but still inside root.
		{5, 3, []string{"leave b"}},
		// Off the app entirely (outside every element).
		{25, 3, []string{"leave root"}},
		// Re-enter, cross to a sibling, then leave everything at once: leaves
		// still fire deepest first.
		{1, 0, []string{"enter root", "enter a"}},
		{1, 1, []string{"leave a", "enter b"}},
		{25, 1, []string{"leave b", "leave root"}},
	}

	for i, s := range steps {
		got := moveTo(app, root, s.x, s.y)
		if !slices.Equal(got, s.want) {
			t.Fatalf("step %d move to (%d,%d): fired %v, want %v", i, s.x, s.y, got, s.want)
		}
	}
}

func TestHover_RenderNeverFiresCallbacks(t *testing.T) {
	root := &hoverTestRoot{}
	app := newHoverTestApp(root)
	moveTo(app, root, 1, 0)

	root.log = nil
	for i := range 3 {
		root.hovered.Set(fmt.Sprint(i))
		app.Render()
	}
	if len(root.log) != 0 {
		t.Errorf("renders fired hover callbacks: %v", root.log)
	}
}

func TestHover_ClicksAndWheelUpdateHover(t *testing.T) {
	type tc struct {
		event MouseEvent
	}

	tests := map[string]tc{
		"left press": {event: MouseEvent{Button: MouseLeft, Action: MousePress, X: 1, Y: 1}},
		"left drag":  {event: MouseEvent{Button: MouseLeft, Action: MouseDrag, X: 1, Y: 1}},
		"wheel down": {event: MouseEvent{Button: MouseWheelDown, Action: MousePress, X: 1, Y: 1}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := &hoverTestRoot{}
			app := newHoverTestApp(root)
			app.Dispatch(tt.event)
			want := []string{"enter root", "enter b"}
			if !slices.Equal(root.log, want) {
				t.Errorf("fired %v, want %v", root.log, want)
			}
		})
	}
}

func TestHover_InlinePointerLeavingRegion(t *testing.T) {
	root := &hoverTestRoot{}
	app := newHoverTestApp(root)
	app.inlineHeight = 4
	app.inlineStartRow = 10

	// Terminal row 10 is buffer row 0.
	if got := moveTo(app, root, 1, 10); !slices.Equal(got, []string{"enter root", "enter a"}) {
		t.Fatalf("entering inline region fired %v", got)
	}
	// Terminal row 2 is above the inline region.
	if got := moveTo(app, root, 1, 2); !slices.Equal(got, []string{"leave a", "leave root"}) {
		t.Fatalf("leaving inline region fired %v", got)
	}
	// Moving around outside stays quiet.
	if got := moveTo(app, root, 3, 1); len(got) != 0 {
		t.Fatalf("moving outside inline region fired %v", got)
	}
}

func TestHover_OpenModalBlocksBackground(t *testing.T) {
	open := NewState(true)
	root := &hoverTestRoot{modal: NewModal(WithModalOpen(open), WithModalBackdrop("none"))}
	app := newHoverTestApp(root)
	app.inAlternateScreen = true
	app.MarkDirty()
	app.Render()

	// Find a point over row "a" that the dialog does not cover.
	dialog := app.overlays[0].element.children[0].Rect()
	x, y := 0, 0
	if dialog.Contains(x, y) {
		x = dialog.Right() + 1
	}
	if got := moveTo(app, root, x, y); len(got) != 0 {
		t.Fatalf("background hovered behind open modal: %v", got)
	}
	if got := moveTo(app, root, dialog.X, dialog.Y); !slices.Equal(got, []string{"enter dialog"}) {
		t.Fatalf("hovering dialog fired %v, want [enter dialog]", got)
	}
}

func TestHover_ModalOpeningUnderStillPointerLeavesBackground(t *testing.T) {
	open := NewState(false)
	root := &hoverTestRoot{modal: NewModal(WithModalOpen(open), WithModalBackdrop("none"))}
	app := newHoverTestApp(root)
	app.inAlternateScreen = true
	moveTo(app, root, 0, 0)

	open.Set(true)
	app.Render()
	dialog := app.overlays[0].element.children[0].Rect()
	x, y := 0, 0
	if dialog.Contains(x, y) {
		x = dialog.Right() + 1
	}

	// The next move leaves everything entered before the modal opened, even
	// though the pointer is still over the same background element.
	if got := moveTo(app, root, x, y); !slices.Equal(got, []string{"leave a", "leave root"}) {
		t.Fatalf("first move after modal opened fired %v, want [leave a leave root]", got)
	}
	if got := moveTo(app, root, dialog.X, dialog.Y); !slices.Equal(got, []string{"enter dialog"}) {
		t.Fatalf("hovering dialog fired %v, want [enter dialog]", got)
	}

	open.Set(false)
	app.Render()
	if got := moveTo(app, root, 1, 1); !slices.Equal(got, []string{"leave dialog", "enter root", "enter b"}) {
		t.Fatalf("first move after modal closed fired %v", got)
	}
}

func TestHover_ContentChangingUnderPointerKeepsEnterLeaveBalanced(t *testing.T) {
	root := &hoverTestRoot{}
	app := newHoverTestApp(root)
	moveTo(app, root, 1, 0)

	// Row "a" is replaced by "z" at the same position while the pointer is still.
	root.rows = []string{"z", "b"}
	root.hovered.Set("changed")
	app.Render()

	if got := moveTo(app, root, 3, 0); len(got) != 0 {
		t.Fatalf("moving within the same position fired %v", got)
	}
	// Leaving fires the handler of the row that was actually entered.
	if got := moveTo(app, root, 3, 1); !slices.Equal(got, []string{"leave a", "enter b"}) {
		t.Fatalf("moving to row b fired %v, want [leave a enter b]", got)
	}
}

func TestHover_RequiresMouseMotion(t *testing.T) {
	root := &hoverTestRoot{}
	app := newHoverTestApp(root)
	app.mouseMotion = false

	app.Dispatch(MouseEvent{Button: MouseLeft, Action: MousePress, X: 1, Y: 0})
	app.Render()
	if len(root.log) != 0 {
		t.Errorf("hover fired without WithMouseMotion: %v", root.log)
	}
}

func TestHover_RecheckAfterContentChangesUnderStillPointer(t *testing.T) {
	type tc struct {
		keyed bool
		want  []string
	}

	tests := map[string]tc{
		// Keys follow the item: "a" moved away, "b" is now under the pointer.
		"keyed rows follow the item": {keyed: true, want: []string{"leave a", "enter b"}},
		// Without keys identity is positional: the first row is still hovered.
		"unkeyed rows are positional": {keyed: false, want: nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := &hoverTestRoot{keyed: tt.keyed}
			app := newHoverTestApp(root)
			moveTo(app, root, 1, 0)

			root.log = nil
			app.Dispatch(UpdateEvent{fn: func() {
				root.rows = []string{"b", "a"}
				root.hovered.Set("reordered")
			}})
			app.Render()
			if !slices.Equal(root.log, tt.want) {
				t.Errorf("after reorder under still pointer fired %v, want %v", root.log, tt.want)
			}
		})
	}
}

func TestHover_HandlerMovingItsElementDoesNotLoop(t *testing.T) {
	root := &hoverTestRoot{keyed: true}
	// Entering a row moves it to the end, so a different row slides under
	// the still pointer every time.
	root.onEnter = func(name string) {
		rows := slices.DeleteFunc(slices.Clone(root.rows), func(r string) bool { return r == name })
		root.rows = append(rows, name)
	}
	root.rows = []string{"a", "b", "c"}
	app := newHoverTestApp(root)

	moveTo(app, root, 1, 0)
	for range 5 {
		app.MarkDirty()
		app.Render()
	}
	// One move: the enter, plus at most one re-check round after the render.
	want := []string{"enter root", "enter a", "leave a", "enter b"}
	if !slices.Equal(root.log, want) {
		t.Errorf("fired %v, want %v (bounded to one re-check per input)", root.log, want)
	}
}

func TestDispatchEvents_CoalescesMoves(t *testing.T) {
	root := &hoverTestRoot{}
	app := newHoverTestApp(root)
	for _, ev := range []MouseEvent{
		{Button: MouseNone, Action: MouseMove, X: 1, Y: 0},
		{Button: MouseNone, Action: MouseMove, X: 2, Y: 0},
		{Button: MouseLeft, Action: MousePress, X: 2, Y: 1},
		{Button: MouseNone, Action: MouseMove, X: 3, Y: 1},
		{Button: MouseNone, Action: MouseMove, X: 4, Y: 1},
	} {
		app.merged <- ev
	}

	app.DispatchEvents()
	want := []MouseEvent{
		{Button: MouseNone, Action: MouseMove, X: 2, Y: 0},
		{Button: MouseLeft, Action: MousePress, X: 2, Y: 1},
		{Button: MouseNone, Action: MouseMove, X: 4, Y: 1},
	}
	if !slices.Equal(root.mouse, want) {
		t.Errorf("components saw %v, want %v", root.mouse, want)
	}
}

func TestHover_ClassStylesHoveredElementAcrossRerenders(t *testing.T) {
	root := &hoverTestRoot{rowClass: "hover:bg-blue"}
	app := newHoverTestApp(root)
	bg := func(y int) Color { return app.buffer.Cell(5, y).Style.Bg }

	moveTo(app, root, 1, 0)
	if bg(0) != Blue || bg(1) == Blue {
		t.Fatalf("hovering row 0: bg row0 %v row1 %v, want row0 blue only", bg(0), bg(1))
	}

	// A state change rebuilds every element; the fresh row keeps the style.
	root.hovered.Set("rebuilt")
	app.Render()
	if bg(0) != Blue {
		t.Fatalf("hover style lost after re-render: bg %v", bg(0))
	}

	moveTo(app, root, 1, 1)
	if bg(0) == Blue || bg(1) != Blue {
		t.Fatalf("after moving to row 1: bg row0 %v row1 %v, want row1 blue only", bg(0), bg(1))
	}
	moveTo(app, root, 1, 3)
	if bg(0) == Blue || bg(1) == Blue {
		t.Errorf("after leaving rows: bg row0 %v row1 %v, want neither blue", bg(0), bg(1))
	}
}

// scrollTestRoot is a bordered two-row viewport over five rows, below a title
// so its content does not start at the screen origin.
type scrollTestRoot struct {
	list *Element
	log  []string
}

func (r *scrollTestRoot) Render(app *App) *Element {
	root := New(WithDirection(Column), WithWidth(12), WithHeight(6))
	root.AddChild(New(WithText("title"), WithHeight(1)))
	r.list = New(WithDirection(Column), WithHeight(4), WithWidth(12), WithBorder(BorderSingle), WithScrollable(ScrollVertical))
	for _, name := range []string{"r0", "r1", "r2", "r3", "r4"} {
		r.list.AddChild(New(WithHeight(1), WithWidth(10), WithText(name),
			WithOnMouseEnter(func() { r.log = append(r.log, "enter "+name) }),
			WithOnMouseLeave(func() { r.log = append(r.log, "leave "+name) }),
		))
	}
	root.AddChild(r.list)
	return root
}

func TestHover_WheelScrollUnderStillPointer(t *testing.T) {
	root := &scrollTestRoot{}
	app := newTestApp(12, 6)
	app.mouseMotion = true
	app.rootComponent = root
	app.MarkDirty()
	app.Render()

	// Row r0 is drawn at screen row 2: title at 0, list border at 1.
	app.Dispatch(MouseEvent{Button: MouseNone, Action: MouseMove, X: 2, Y: 2})
	app.Render()
	if want := []string{"enter r0"}; !slices.Equal(root.log, want) {
		t.Fatalf("hovering the first visible row fired %v, want %v", root.log, want)
	}
	root.log = nil

	// Two wheel notches with the pointer still over the first visible row.
	for range 2 {
		app.Dispatch(MouseEvent{Button: MouseWheelDown, Action: MousePress, X: 2, Y: 2})
		app.Render()
	}
	if want := []string{"leave r0", "enter r1", "leave r1", "enter r2"}; !slices.Equal(root.log, want) {
		t.Errorf("after scrolling under the pointer fired %v, want %v", root.log, want)
	}

	// The border is part of the list, not of any row.
	root.log = nil
	app.Dispatch(MouseEvent{Button: MouseNone, Action: MouseMove, X: 2, Y: 1})
	app.Render()
	if want := []string{"leave r2"}; !slices.Equal(root.log, want) {
		t.Errorf("moving onto the list border fired %v, want %v", root.log, want)
	}
}
