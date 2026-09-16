package tui

import "testing"

// lifecycleWidget records the app binding calls it receives.
type lifecycleWidget struct {
	binds   int
	unbinds int
}

func (w *lifecycleWidget) Render(app *App) *Element { return New() }
func (w *lifecycleWidget) BindApp(app *App)         { w.binds++ }
func (w *lifecycleWidget) UnbindApp()               { w.unbinds++ }

// widgetHost mirrors the shape the gsx generator emits for a struct component
// that renders a slice field through a for loop.
type widgetHost struct {
	widgets []Component
}

func (h *widgetHost) Render(app *App) *Element {
	root := New()
	for _, w := range h.widgets {
		root.AddChild(w.Render(app))
	}
	return root
}

func (h *widgetHost) UpdateProps(fresh Component) {
	f, ok := fresh.(*widgetHost)
	if !ok {
		return
	}
	for _, item := range h.widgets {
		if unbinder, ok := any(item).(AppUnbinder); ok {
			unbinder.UnbindApp()
		}
	}
	h.widgets = f.widgets
}

func (h *widgetHost) BindApp(app *App) {
	for _, item := range h.widgets {
		if binder, ok := any(item).(AppBinder); ok {
			binder.BindApp(app)
		}
	}
}

// A cached mount calls UpdateProps then BindApp, so a widget dropped by the
// prop swap only sees UnbindApp if UpdateProps unbinds the old slice first.
func TestMount_UpdatePropsUnbindsReplacedWidgets(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	a := &lifecycleWidget{}
	b := &lifecycleWidget{}

	testApp.Mount(parent, 0, func() Component { return &widgetHost{widgets: []Component{a}} })
	if a.binds != 1 {
		t.Fatalf("widget a bound %d times after first mount, want 1", a.binds)
	}

	testApp.Mount(parent, 0, func() Component { return &widgetHost{widgets: []Component{b}} })
	if a.unbinds != 1 {
		t.Errorf("widget a unbound %d times after prop swap, want 1", a.unbinds)
	}
	if b.binds != 1 {
		t.Errorf("widget b bound %d times after prop swap, want 1", b.binds)
	}
	if b.unbinds != 0 {
		t.Errorf("widget b unbound %d times after prop swap, want 0", b.unbinds)
	}
}
