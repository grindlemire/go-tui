package tui

import (
	"testing"
	"time"
)

type testWatcherComponent struct {
	watchers []Watcher
}

func (t *testWatcherComponent) Render(app *App) *Element { return New() }
func (t *testWatcherComponent) Watchers() []Watcher      { return t.watchers }

func TestCollectComponentWatchers(t *testing.T) {
	type tc struct {
		setup    func() *Element
		expected int
	}

	tests := map[string]tc{
		"single component with one watcher": {
			setup: func() *Element {
				root := New()
				comp := &testWatcherComponent{
					watchers: []Watcher{
						OnTimer(time.Second, func() {}),
					},
				}
				child := New()
				child.component = comp
				root.AddChild(child)
				return root
			},
			expected: 1,
		},
		"nested components with multiple watchers": {
			setup: func() *Element {
				root := New()
				comp1 := &testWatcherComponent{
					watchers: []Watcher{OnTimer(time.Second, func() {})},
				}
				comp2 := &testWatcherComponent{
					watchers: []Watcher{
						OnTimer(time.Second, func() {}),
						OnTimer(time.Millisecond*500, func() {}),
					},
				}

				child1 := New()
				child1.component = comp1

				child2 := New()
				child2.component = comp2
				child1.AddChild(child2)

				root.AddChild(child1)
				return root
			},
			expected: 3,
		},
		"no components": {
			setup: func() *Element {
				root := New()
				root.AddChild(New())
				root.AddChild(New())
				return root
			},
			expected: 0,
		},
		"component without WatcherProvider": {
			setup: func() *Element {
				root := New()
				// Use a component that doesn't implement WatcherProvider
				comp := &simpleComponent{}
				child := New()
				child.component = comp
				root.AddChild(child)
				return root
			},
			expected: 0,
		},
		"nil root": {
			setup: func() *Element {
				return nil
			},
			expected: 0,
		},
		"component on root element": {
			setup: func() *Element {
				root := New()
				comp := &testWatcherComponent{
					watchers: []Watcher{
						OnTimer(time.Second, func() {}),
					},
				}
				root.component = comp
				return root
			},
			expected: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := tt.setup()
			watchers := collectComponentWatchers(nil, root)

			if len(watchers) != tt.expected {
				t.Fatalf("expected %d watchers, got %d", tt.expected, len(watchers))
			}
		})
	}
}

// simpleComponent implements Component but not WatcherProvider
type simpleComponent struct{}

func (s *simpleComponent) Render(app *App) *Element { return New() }

// testViewComponent mimics a generated view struct: it implements Viewable
// (GetRoot/GetWatchers) but not WatcherProvider.
type testViewComponent struct {
	root     *Element
	watchers []Watcher
}

func (v *testViewComponent) Render(app *App) *Element { return v.root }
func (v *testViewComponent) GetRoot() *Element        { return v.root }
func (v *testViewComponent) GetWatchers() []Watcher   { return v.watchers }

// testDualWatcherComponent implements both interfaces; only Watchers() counts.
type testDualWatcherComponent struct {
	testViewComponent
}

func (d *testDualWatcherComponent) Watchers() []Watcher { return d.watchers }

func TestCollectComponentWatchers_MountedView(t *testing.T) {
	type tc struct {
		comp     Component
		expected int
	}

	two := []Watcher{OnTimer(time.Second, func() {}), OnTimer(time.Minute, func() {})}
	tests := map[string]tc{
		"view exposing only GetWatchers is collected": {
			comp:     &testViewComponent{root: New(), watchers: two},
			expected: 2,
		},
		"component with both interfaces is counted once": {
			comp:     &testDualWatcherComponent{testViewComponent{root: New(), watchers: two}},
			expected: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := New()
			child := New()
			child.component = tt.comp
			root.AddChild(child)

			got := collectComponentWatchers(nil, root)
			if len(got) != tt.expected {
				t.Errorf("collected %d watchers, want %d", len(got), tt.expected)
			}
		})
	}
}

type countStartWatcher struct{ starts int }

func (c *countStartWatcher) Start(chan<- func(), <-chan struct{}) { c.starts++ }

type elementMountHost struct{ el *Element }

func (h *elementMountHost) Render(app *App) *Element {
	root := New()
	root.AddChild(app.Mount(h, 0, func() Component { return h.el }))
	return root
}

// An Element used directly as a component must not have its watchers
// started twice (once by the tree walk, once by the component walk).
func TestElementAsComponent_WatchersStartOnce(t *testing.T) {
	type tc struct {
		setRoot func(app *App, el *Element)
	}

	tests := map[string]tc{
		"SetRootComponent(el)": {
			setRoot: func(app *App, el *Element) { app.SetRootComponent(el) },
		},
		"Mount(el)": {
			setRoot: func(app *App, el *Element) { app.SetRootComponent(&elementMountHost{el: el}) },
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			w := &countStartWatcher{}
			el := New()
			el.AddWatcher(w)
			app := newTestApp(20, 5)
			tt.setRoot(app, el)
			app.Render()
			if w.starts != 1 {
				t.Fatalf("watcher Start calls = %d, want 1", w.starts)
			}
		})
	}
}
