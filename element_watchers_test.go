package tui

import (
	"testing"
)

func TestElement_Watchers(t *testing.T) {
	type tc struct {
		setup func(e *Element)
		want  int
	}

	tests := map[string]tc{
		"no watchers by default": {
			setup: func(e *Element) {},
			want:  0,
		},
		"add one watcher": {
			setup: func(e *Element) {
				ch := make(chan string)
				e.AddWatcher(Watch(ch, func(s string) {}))
			},
			want: 1,
		},
		"add multiple watchers": {
			setup: func(e *Element) {
				ch1 := make(chan string)
				ch2 := make(chan int)
				e.AddWatcher(Watch(ch1, func(s string) {}))
				e.AddWatcher(Watch(ch2, func(i int) {}))
			},
			want: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New()
			tt.setup(e)
			if got := len(e.Watchers()); got != tt.want {
				t.Errorf("len(Watchers()) = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestElement_WalkWatchers(t *testing.T) {
	parent := New()
	child := New()
	parent.AddChild(child)

	ch1 := make(chan string)
	ch2 := make(chan int)
	parent.AddWatcher(Watch(ch1, func(s string) {}))
	child.AddWatcher(Watch(ch2, func(i int) {}))

	var count int
	parent.WalkWatchers(func(w Watcher) {
		count++
	})

	if count != 2 {
		t.Errorf("WalkWatchers visited %d watchers, want 2", count)
	}
}

func TestElement_SetOnUpdate(t *testing.T) {
	e := New()
	called := false

	e.SetOnUpdate(func() {
		called = true
	})

	// onUpdate is called during render
	buf := NewBuffer(10, 10)
	e.Render(buf, 10, 10)

	if !called {
		t.Error("SetOnUpdate callback should be called during Render")
	}
}
