package tui

import (
	"slices"
	"testing"
)

// exprChild records which framework hooks reach it when rendered through an
// @expr (MountInstance) site.
type exprChild struct {
	name   string
	events *[]string
}

func (c *exprChild) Render(app *App) *Element {
	return New(WithText(c.name), WithWidth(10), WithHeight(1))
}

func (c *exprChild) KeyMap() KeyMap {
	return KeyMap{On(Rune('k'), func(KeyEvent) { *c.events = append(*c.events, c.name+" key") })}
}

func (c *exprChild) HandleMouse(me MouseEvent) bool {
	*c.events = append(*c.events, c.name+" mouse")
	return false
}

func (c *exprChild) Init() func() {
	*c.events = append(*c.events, c.name+" init")
	return func() { *c.events = append(*c.events, c.name+" cleanup") }
}

// exprParent mirrors generated code for `@p.child` behind an if.
type exprParent struct {
	child *exprChild
	show  bool
}

func (p *exprParent) Render(app *App) *Element {
	root := New(WithWidth(10), WithHeight(2))
	if p.show {
		root.AddChild(app.MountInstance(p, 0, p.child))
	}
	return root
}

func TestMountInstance_ComponentGetsFrameworkHooks(t *testing.T) {
	var events []string
	a := &exprChild{name: "a", events: &events}
	parent := &exprParent{child: a, show: true}
	app := newTestApp(10, 2)
	app.rootComponent = parent
	render := func() {
		app.MarkDirty()
		app.Render()
	}

	render()
	render()
	app.Dispatch(KeyEvent{Key: KeyRune, Rune: 'k'})
	app.Dispatch(MouseEvent{Button: MouseLeft, Action: MousePress, X: 1, Y: 0})
	if want := []string{"a init", "a key", "a mouse"}; !slices.Equal(events, want) {
		t.Fatalf("hooks = %v, want %v (init once, then key and mouse dispatch)", events, want)
	}

	// Pointing the field at another instance mounts it during the render and
	// sweeps the old one after, as with a changed @Component(...) key.
	events = nil
	b := &exprChild{name: "b", events: &events}
	parent.child = b
	render()
	if want := []string{"b init", "a cleanup"}; !slices.Equal(events, want) {
		t.Fatalf("after swapping the field: %v, want %v", events, want)
	}

	// No longer rendered: swept like any other mount.
	events = nil
	parent.show = false
	render()
	if want := []string{"b cleanup"}; !slices.Equal(events, want) {
		t.Errorf("after hiding: %v, want %v", events, want)
	}
}

func TestMountInstance_ElementRendersAsItself(t *testing.T) {
	app := newTestApp(10, 2)
	el := New(WithText("plain"))
	if got := app.MountInstance(nil, 0, el); got != el {
		t.Errorf("MountInstance(element) = %p, want the element itself %p", got, el)
	}
}

// exprListParent mirrors generated code for `for _, c := range p.items { @c }`,
// which passes the loop position as the site key.
type exprListParent struct {
	items []*exprChild
}

func (p *exprListParent) Render(app *App) *Element {
	root := New(WithWidth(10), WithHeight(3))
	for i, c := range p.items {
		root.AddChild(app.MountInstance(p, MountKey(0, i), c))
	}
	return root
}

func TestMountInstance_ReorderKeepsInstancesMounted(t *testing.T) {
	var events []string
	a := &exprChild{name: "a", events: &events}
	b := &exprChild{name: "b", events: &events}
	parent := &exprListParent{items: []*exprChild{a, b}}
	app := newTestApp(10, 3)
	app.rootComponent = parent
	app.MarkDirty()
	app.Render()

	events = nil
	parent.items = []*exprChild{b, a}
	app.MarkDirty()
	app.Render()
	if len(events) != 0 {
		t.Errorf("reordering mounted components re-ran hooks: %v", events)
	}
}

// exprPane renders the shared child when it holds it.
type exprPane struct {
	child *exprChild
}

func (p *exprPane) Render(app *App) *Element {
	root := New(WithWidth(5), WithHeight(1))
	if p.child != nil {
		root.AddChild(app.MountInstance(p, 0, p.child))
	}
	return root
}

// exprWrapperRoot builds a fresh wrapper around child every render, the way a
// hand-written layout helper might.
type exprWrapperRoot struct {
	child *exprChild
	panes [2]*exprPane
	twice bool
}

func (r *exprWrapperRoot) Render(app *App) *Element {
	root := New(WithWidth(10), WithHeight(2))
	switch {
	case r.panes[0] != nil:
		root.AddChild(r.panes[0].Render(app), r.panes[1].Render(app))
	case r.twice:
		root.AddChild(app.MountInstance(r, 0, r.child), app.MountInstance(r, 1, r.child))
	default:
		root.AddChild((&exprPane{child: r.child}).Render(app))
	}
	return root
}

func TestMountInstance_SameInstanceIsOneMountWhereverRendered(t *testing.T) {
	type tc struct {
		setup func(r *exprWrapperRoot)
		move  func(r *exprWrapperRoot)
	}

	tests := map[string]tc{
		"moved between parents": {
			setup: func(r *exprWrapperRoot) { r.panes = [2]*exprPane{{child: r.child}, {}} },
			move:  func(r *exprWrapperRoot) { r.panes[0].child, r.panes[1].child = nil, r.child },
		},
		"parent rebuilt every render": {},
		"rendered at two sites": {
			setup: func(r *exprWrapperRoot) { r.twice = true },
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var events []string
			r := &exprWrapperRoot{child: &exprChild{name: "c", events: &events}}
			if tt.setup != nil {
				tt.setup(r)
			}
			app := newTestApp(10, 2)
			app.rootComponent = r
			for i := range 3 {
				if i == 1 && tt.move != nil {
					tt.move(r)
				}
				app.MarkDirty()
				app.Render()
			}
			if want := []string{"c init"}; !slices.Equal(events, want) {
				t.Errorf("hooks = %v, want %v (one mount for one instance)", events, want)
			}
		})
	}
}
