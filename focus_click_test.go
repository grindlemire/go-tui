package tui

import "testing"

// clickFocusRoot lays out one target per row: an input, a focusable box
// holding a child, plain text, and a scrollable container (focusable, but
// not a tab stop).
type clickFocusRoot struct {
	input *Input
	refs  map[string]*Ref
}

func (r *clickFocusRoot) Render(app *App) *Element {
	root := New(WithDirection(Column), WithWidth(20), WithHeight(4))
	inputEl := app.MountInstance(r, 0, r.input)
	r.refs["input"].Set(inputEl)
	box := New(WithFocusable(true), WithHeight(1))
	box.AddChild(New(WithText("inner")))
	r.refs["box"].Set(box)
	scroll := New(WithScrollable(ScrollVertical), WithHeight(1))
	scroll.AddChild(New(WithText("scroll")))
	root.AddChild(inputEl, box, New(WithText("plain")), scroll)
	return root
}

func newClickFocusApp() (*App, *clickFocusRoot) {
	r := &clickFocusRoot{
		input: NewInput(WithInputWidth(10)),
		refs:  map[string]*Ref{"input": NewRef(), "box": NewRef()},
	}
	app := newTestApp(20, 4)
	app.SetRootComponent(r)
	app.Render()
	return app, r
}

func press(button MouseButton, x, y int) MouseEvent {
	return MouseEvent{Button: button, Action: MousePress, X: x, Y: y}
}

func TestClickFocusesFocusableUnderPointer(t *testing.T) {
	type tc struct {
		events []MouseEvent
		want   string // ref name of the focused element, "" for none
	}

	tests := map[string]tc{
		"click an input": {
			events: []MouseEvent{press(MouseLeft, 1, 0)},
			want:   "input",
		},
		"click a child of a focusable element": {
			events: []MouseEvent{press(MouseLeft, 1, 1)},
			want:   "box",
		},
		"click another focusable element": {
			events: []MouseEvent{press(MouseLeft, 1, 0), press(MouseLeft, 1, 1)},
			want:   "box",
		},
		"click with nothing focusable keeps focus": {
			events: []MouseEvent{press(MouseLeft, 1, 0), press(MouseLeft, 1, 2)},
			want:   "input",
		},
		"scrollable container is not a tab stop": {
			events: []MouseEvent{press(MouseLeft, 1, 3)},
			want:   "",
		},
		"other buttons and actions do not focus": {
			events: []MouseEvent{
				press(MouseRight, 1, 0),
				press(MouseWheelDown, 1, 0),
				{Button: MouseLeft, Action: MouseRelease, X: 1, Y: 1},
				{Button: MouseLeft, Action: MouseDrag, X: 1, Y: 1},
			},
			want: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app, r := newClickFocusApp()
			for _, ev := range tt.events {
				app.Dispatch(ev)
				app.Render()
			}
			got := ""
			for name, ref := range r.refs {
				if el := ref.El(); el != nil && app.Focused() == Focusable(el) {
					got = name
				}
			}
			if got != tt.want {
				t.Errorf("focused %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClickedInputTakesTyping(t *testing.T) {
	app, r := newClickFocusApp()
	app.Dispatch(KeyEvent{Key: KeyRune, Rune: 'a'})
	app.Dispatch(press(MouseLeft, 1, 0))
	app.Render()
	app.Dispatch(KeyEvent{Key: KeyRune, Rune: 'b'})
	if got := r.input.Text(); got != "b" {
		t.Errorf("input text = %q, want %q (typed only after the click)", got, "b")
	}
}

// clickFocusModalRoot puts a focusable box behind an open modal whose dialog
// holds a focusable button.
type clickFocusModalRoot struct {
	modal *Modal
}

func (r *clickFocusModalRoot) Render(app *App) *Element {
	root := New(WithDirection(Column), WithWidth(20), WithHeight(4))
	root.AddChild(New(WithHeight(2)))
	overlay := app.MountInstance(r, 0, r.modal)
	overlay.AddChild(New(WithFocusable(true), WithText("dialog"), WithWidth(6), WithHeight(1)))
	// Declared before the box, so only the hit scope keeps the click off it.
	root.AddChild(overlay, New(WithFocusable(true), WithText("behind"), WithHeight(1)))
	return root
}

func TestClickBehindModalDoesNotTakeFocus(t *testing.T) {
	open := NewState(true)
	r := &clickFocusModalRoot{
		modal: NewModal(WithModalOpen(open), WithModalTrapFocus(false), WithModalCloseOnBackdropClick(false)),
	}
	app := newTestApp(20, 4)
	app.SetRootComponent(r)
	app.Render()

	app.Dispatch(press(MouseLeft, 1, 2))
	app.Render()
	if app.Focused() != nil {
		t.Errorf("click behind the backdrop focused %q", focusedText(app))
	}
	app.Dispatch(press(MouseLeft, 1, 0))
	app.Render()
	if got := focusedText(app); got != "dialog" {
		t.Errorf("click on the dialog focused %q, want %q", got, "dialog")
	}
}

// stackedModalsRoot opens a modal without a focus trap over one that traps
// focus.
type stackedModalsRoot struct {
	lower, upper *Modal
}

func (r *stackedModalsRoot) Render(app *App) *Element {
	root := New(WithWidth(20), WithHeight(4))
	lower := app.MountInstance(r, 0, r.lower)
	lower.AddChild(New(WithFocusable(true), WithText("lower"), WithWidth(6), WithHeight(1)))
	upper := app.MountInstance(r, 1, r.upper)
	upper.AddChild(New(WithFocusable(true), WithText("upper"), WithWidth(6), WithHeight(1)))
	root.AddChild(lower, upper)
	return root
}

func TestClickOutsideTrappingModalDoesNotTakeFocus(t *testing.T) {
	r := &stackedModalsRoot{
		lower: NewModal(WithModalOpen(NewState(true))),
		upper: NewModal(WithModalOpen(NewState(true)), WithModalTrapFocus(false), WithModalCloseOnBackdropClick(false)),
	}
	app := newTestApp(20, 4)
	app.SetRootComponent(r)
	app.Render()
	before := focusedText(app)

	upper := app.root.ElementAt(0, 0)
	for upper != nil && upper.Text() != "upper" {
		upper = upper.parent
	}
	if upper == nil {
		t.Fatal("setup: upper dialog not at (0, 0)")
	}
	app.Dispatch(press(MouseLeft, 0, 0))
	app.Render()
	if got := focusedText(app); got != before {
		t.Errorf("click outside the focus trap focused %q, want %q", got, before)
	}
}
