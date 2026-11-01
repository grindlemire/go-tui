package tui

import "testing"

// wheelTestRoot renders a bordered, scrollable list of six rows under a
// title. With controlled set, the app pins the offset with WithScrollOffset
// every render, as apps that keep scroll position in state do.
type wheelTestRoot struct {
	controlled bool
	modal      *Modal
	list       *Element
}

func (r *wheelTestRoot) Render(app *App) *Element {
	root := New(WithDirection(Column), WithWidth(12), WithHeight(6))
	root.AddChild(New(WithText("title"), WithHeight(1)))
	opts := []Option{WithDirection(Column), WithHeight(4), WithWidth(12), WithBorder(BorderSingle), WithScrollable(ScrollVertical)}
	if r.controlled {
		opts = append(opts, WithScrollOffset(0, 0))
	}
	// Declared before the list, so plain last-child-first hit testing would
	// reach the list before the overlay.
	if r.modal != nil {
		overlay := r.modal.Render(app)
		overlay.AddChild(New(WithText("dialog"), WithWidth(6), WithHeight(1)))
		root.AddChild(overlay)
	}
	r.list = New(opts...)
	for range 6 {
		r.list.AddChild(New(WithText("row"), WithHeight(1)))
	}
	root.AddChild(r.list)
	return root
}

func newWheelTestApp(root *wheelTestRoot) *App {
	app := newTestApp(12, 6)
	app.inAlternateScreen = true
	if root.modal != nil {
		root.modal.BindApp(app)
	}
	app.rootComponent = root
	app.MarkDirty()
	app.Render()
	return app
}

// wheelDown scrolls over the first visible row (screen row 2) and renders
// twice, so the offset has to survive a rebuild of the tree.
func wheelDown(app *App) {
	app.Dispatch(MouseEvent{Button: MouseWheelDown, Action: MousePress, X: 2, Y: 2})
	app.Render()
	app.MarkDirty()
	app.Render()
}

func TestWheel_ScrollsContainerUnderPointer(t *testing.T) {
	type tc struct {
		controlled bool
		want       int
	}

	tests := map[string]tc{
		// The wheel lands on a row, reaches the list, and the offset is
		// carried to the rebuilt list on each render.
		"uncontrolled list scrolls and keeps its offset": {want: 1},
		// The app owns the offset, so the carried value never overrides it.
		"controlled offset is left to the app": {controlled: true, want: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := &wheelTestRoot{controlled: tt.controlled}
			app := newWheelTestApp(root)
			wheelDown(app)
			if _, y := root.list.ScrollOffset(); y != tt.want {
				t.Errorf("scroll offset after wheel and re-render = %d, want %d", y, tt.want)
			}
		})
	}
}

func TestWheel_OpenModalBlocksBackgroundScroll(t *testing.T) {
	root := &wheelTestRoot{modal: NewModal(WithModalOpen(NewState(true)), WithModalBackdrop("none"))}
	app := newWheelTestApp(root)
	wheelDown(app)
	if _, y := root.list.ScrollOffset(); y != 0 {
		t.Errorf("background list scrolled to %d behind an open modal", y)
	}
}
