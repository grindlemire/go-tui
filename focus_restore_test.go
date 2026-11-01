package tui

import "testing"

// focusRestoreRoot renders focusable rows whose positions change between
// renders, optionally after a focusable lead row or a plain note, or a
// mounted input in place of the rows. nested puts the rows in a container.
type focusRestoreRoot struct {
	lead      bool
	note      bool
	rows      []string
	keyed     bool
	nested    bool
	input     *Input
	showInput bool
}

func (r *focusRestoreRoot) Render(app *App) *Element {
	root := New(WithDirection(Column), WithWidth(20), WithHeight(8))
	if r.lead {
		root.AddChild(New(WithFocusable(true), WithText("lead"), WithHeight(1)))
	}
	if r.note {
		root.AddChild(New(WithText("note"), WithHeight(1)))
	}
	if r.showInput {
		root.AddChild(app.MountInstance(r, 0, r.input))
		return root
	}
	list := root
	if r.nested {
		list = New(WithDirection(Column))
		root.AddChild(list)
	}
	for _, name := range r.rows {
		opts := []Option{WithFocusable(true), WithText(name), WithHeight(1)}
		if r.keyed {
			opts = append(opts, WithKey(name))
		}
		list.AddChild(New(opts...))
	}
	return root
}

func newFocusRestoreApp(r *focusRestoreRoot) *App {
	app := newTestApp(20, 8)
	app.SetRootComponent(r)
	app.Render()
	return app
}

// focusText focuses the registered element showing text.
func focusText(t *testing.T, app *App, text string) {
	t.Helper()
	for _, f := range app.focus.elements {
		if el, ok := f.(*Element); ok && el.Text() == text {
			app.focus.SetFocus(el)
			return
		}
	}
	t.Fatalf("no focusable element shows %q", text)
}

func focusedText(app *App) string {
	if el, ok := app.Focused().(*Element); ok {
		return el.Text()
	}
	return ""
}

func TestFocusFollowsItsElementAcrossRenders(t *testing.T) {
	type tc struct {
		root   focusRestoreRoot
		focus  string
		change func(r *focusRestoreRoot)
		want   string
	}

	tests := map[string]tc{
		"focusable added before a keyed row": {
			root:   focusRestoreRoot{rows: []string{"x", "y"}, keyed: true},
			focus:  "y",
			change: func(r *focusRestoreRoot) { r.lead = true },
			want:   "y",
		},
		// Without a key, focus keeps its position among the component's
		// focusables, like an unkeyed list in React.
		"focusable added before an unkeyed row": {
			root:   focusRestoreRoot{rows: []string{"x", "y"}},
			focus:  "y",
			change: func(r *focusRestoreRoot) { r.lead = true },
			want:   "x",
		},
		"plain line added before it": {
			root:   focusRestoreRoot{rows: []string{"x", "y"}},
			focus:  "y",
			change: func(r *focusRestoreRoot) { r.note = true },
			want:   "y",
		},
		"plain line added above a keyed list's container": {
			root:   focusRestoreRoot{rows: []string{"x", "y"}, keyed: true, nested: true},
			focus:  "y",
			change: func(r *focusRestoreRoot) { r.note = true },
			want:   "y",
		},
		"keyed rows reordered": {
			root:   focusRestoreRoot{rows: []string{"a", "b", "c"}, keyed: true},
			focus:  "b",
			change: func(r *focusRestoreRoot) { r.rows = []string{"b", "a", "c"} },
			want:   "b",
		},
		"focused row removed": {
			root:   focusRestoreRoot{rows: []string{"a", "b"}, keyed: true},
			focus:  "a",
			change: func(r *focusRestoreRoot) { r.rows = []string{"b"} },
			want:   "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := tt.root
			app := newFocusRestoreApp(&r)
			focusText(t, app, tt.focus)
			tt.change(&r)
			app.MarkDirty()
			app.Render()
			if got := focusedText(app); got != tt.want {
				t.Errorf("focused %q after the change, want %q", got, tt.want)
			}
		})
	}
}

func TestFocusedComponentUnmountedIsBlurred(t *testing.T) {
	r := &focusRestoreRoot{rows: []string{"other"}, input: NewInput(), showInput: true}
	app := newFocusRestoreApp(r)
	app.FocusNext()
	if !r.input.IsFocused() {
		t.Fatal("setup: input not focused")
	}

	// Another page replaces the input: nothing there takes its focus, and
	// the input hears the blur so it stops claiming keys.
	r.showInput = false
	app.MarkDirty()
	app.Render()
	if got := focusedText(app); got != "" || app.Focused() != nil {
		t.Errorf("focused %q after the input unmounted, want nothing", got)
	}
	if r.input.IsFocused() {
		t.Error("unmounted input still reports focus")
	}

	r.showInput = true
	app.MarkDirty()
	app.Render()
	if r.input.IsFocused() {
		t.Error("input came back focused without being focused again")
	}
}

// valuePage is a value-type component holding a slice, so it never compares
// equal to itself.
type valuePage struct{ items []string }

func (p valuePage) Render(app *App) *Element {
	root := New(WithDirection(Column))
	for _, name := range p.items {
		root.AddChild(New(WithFocusable(true), WithText(name), WithHeight(1)))
	}
	return root
}

type valuePageRoot struct{}

func (r *valuePageRoot) Render(app *App) *Element {
	root := New(WithWidth(20), WithHeight(4))
	root.AddChild(app.Mount(r, 0, func() Component { return valuePage{items: []string{"x", "y"}} }))
	return root
}

func TestFocusInValueComponentSurvivesRender(t *testing.T) {
	app := newTestApp(20, 4)
	app.SetRootComponent(&valuePageRoot{})
	app.Render()
	focusText(t, app, "y")
	app.MarkDirty()
	app.Render()
	if got := focusedText(app); got != "y" {
		t.Errorf("focused %q after a re-render, want %q", got, "y")
	}
}

// modalBeforeRoot declares a focus-trapping modal ahead of a focusable row,
// so the dialog's focusables precede the row while the modal is open.
type modalBeforeRoot struct {
	modal *Modal
}

func (r *modalBeforeRoot) Render(app *App) *Element {
	root := New(WithDirection(Column), WithWidth(20), WithHeight(4))
	overlay := app.MountInstance(r, 0, r.modal)
	overlay.AddChild(New(WithFocusable(true), WithText("dialog"), WithWidth(6), WithHeight(1)))
	root.AddChild(overlay, New(WithFocusable(true), WithText("row"), WithHeight(1)))
	return root
}

func TestModalCloseRestoresFocusedElement(t *testing.T) {
	open := NewState(false)
	r := &modalBeforeRoot{modal: NewModal(WithModalOpen(open))}
	app := newTestApp(20, 4)
	app.SetRootComponent(r)
	app.Render()
	focusText(t, app, "row")

	open.Set(true)
	app.Render()
	if got := focusedText(app); got != "dialog" {
		t.Fatalf("setup: opening the modal focused %q, want %q", got, "dialog")
	}
	open.Set(false)
	app.Render()
	if got := focusedText(app); got != "row" {
		t.Errorf("closing the modal focused %q, want %q", got, "row")
	}
}
