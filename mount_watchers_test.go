package tui

import (
	"fmt"
	"slices"
	"testing"
	"time"
)

// probeWatcher records Start calls and exposes the stop channel it was given.
type probeWatcher struct {
	starts int
	stop   <-chan struct{}
}

func (w *probeWatcher) Start(_ chan<- func(), stop <-chan struct{}) {
	w.starts++
	w.stop = stop
}

// stopped waits briefly for the stop channel.
func (w *probeWatcher) stopped(wait time.Duration) bool {
	select {
	case <-w.stop:
		return true
	case <-time.After(wait):
		return false
	}
}

type watchedChild struct{ w *probeWatcher }

func (c *watchedChild) Render(app *App) *Element { return New(WithText("child")) }
func (c *watchedChild) Watchers() []Watcher      { return []Watcher{c.w} }

type watchedParent struct {
	child *watchedChild
	show  bool
}

func (p *watchedParent) Render(app *App) *Element {
	root := New()
	if p.show {
		root.AddChild(app.MountInstance(p, 0, p.child))
	}
	return root
}

func TestMountedComponentWatchers(t *testing.T) {
	type tc struct {
		shownAtStart bool
	}

	tests := map[string]tc{
		"shown at the first render": {shownAtStart: true},
		"shown after startup":       {shownAtStart: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			w := &probeWatcher{}
			parent := &watchedParent{child: &watchedChild{w: w}, show: tt.shownAtStart}
			app := newTestApp(10, 2)
			app.SetRootComponent(parent)
			render := func() {
				app.MarkDirty()
				app.Render()
			}
			render()

			parent.show = true
			render()
			render()
			if w.starts != 1 {
				t.Fatalf("watcher started %d times, want once", w.starts)
			}
			if w.stopped(20 * time.Millisecond) {
				t.Fatal("watcher stopped while its component is mounted")
			}

			parent.show = false
			render()
			if !w.stopped(time.Second) {
				t.Error("watcher still running after its component was unmounted")
			}
		})
	}
}

// layoutWatchChild records what its OnChange handler sees on its first call.
type layoutWatchChild struct {
	ref   *Ref
	state *State[int]
	seen  []string
}

func (c *layoutWatchChild) Render(app *App) *Element {
	el := New(WithText("child"), WithWidth(6), WithHeight(1))
	c.ref.Set(el)
	return el
}

func (c *layoutWatchChild) Watchers() []Watcher {
	return []Watcher{OnChange(c.state, func(int) {
		width := -1
		if el := c.ref.El(); el != nil {
			width = el.Rect().Width
		}
		c.seen = append(c.seen, fmt.Sprintf("width=%d", width))
	})}
}

func TestMountedWatchersStartAfterLayout(t *testing.T) {
	child := &layoutWatchChild{ref: NewRef(), state: NewState(0)}
	parent := &mountInstanceParent{child: child}
	app := newTestApp(10, 2)
	app.SetRootComponent(parent)
	app.Render()
	if want := []string{"width=6"}; !slices.Equal(child.seen, want) {
		t.Errorf("first OnChange call saw %v, want %v (after layout)", child.seen, want)
	}
}

// mountInstanceParent renders child through MountInstance unless hidden.
type mountInstanceParent struct {
	child Component
	hide  bool
}

func (p *mountInstanceParent) Render(app *App) *Element {
	root := New()
	if !p.hide {
		root.AddChild(app.MountInstance(p, 0, p.child))
	}
	return root
}

type watcherListChild struct{ ws []Watcher }

func (c *watcherListChild) Render(app *App) *Element { return New(WithText("child")) }
func (c *watcherListChild) Watchers() []Watcher      { return c.ws }

func TestUnmountedWatcherHandlersDoNotRun(t *testing.T) {
	type tc struct {
		watcher func(fired *int) (Watcher, func())
	}

	tests := map[string]tc{
		"state change after unmount": {
			watcher: func(fired *int) (Watcher, func()) {
				s := NewState(0)
				w := OnChange(s, func(int) { *fired++ })
				return w, func() { s.Set(1) }
			},
		},
		"timer tick queued before unmount": {
			watcher: func(fired *int) (Watcher, func()) {
				return OnTimer(time.Millisecond, func() { *fired++ }), nil
			},
		},
		"channel value queued before unmount": {
			watcher: func(fired *int) (Watcher, func()) {
				ch := make(chan int, 1)
				ch <- 1
				return Watch(ch, func(int) { *fired++ }), nil
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fired := 0
			w, afterUnmount := tt.watcher(&fired)
			parent := &mountInstanceParent{child: &watcherListChild{ws: []Watcher{w}}}
			app := newTestApp(10, 2)
			app.SetRootComponent(parent)
			app.Render()
			deadline := time.Now().Add(time.Second)
			for afterUnmount == nil && len(app.watcherQueue) == 0 && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			if afterUnmount == nil && len(app.watcherQueue) == 0 {
				t.Fatal("no handler queued before unmount")
			}
			fired = 0

			parent.hide = true
			app.MarkDirty()
			app.Render()
			if afterUnmount != nil {
				afterUnmount()
			}
			for len(app.watcherQueue) > 0 {
				(<-app.watcherQueue)()
			}
			if fired != 0 {
				t.Errorf("handler ran %d times after its component unmounted", fired)
			}
		})
	}
}
