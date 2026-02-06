package tui

import (
	"testing"
	"time"
)

type testWatcherComponent struct {
	watchers []Watcher
}

func (t *testWatcherComponent) Render() *Element { return New() }
func (t *testWatcherComponent) Watchers() []Watcher { return t.watchers }

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
			watchers := collectComponentWatchers(root)

			if len(watchers) != tt.expected {
				t.Fatalf("expected %d watchers, got %d", tt.expected, len(watchers))
			}
		})
	}
}

// simpleComponent implements Component but not WatcherProvider
type simpleComponent struct{}

func (s *simpleComponent) Render() *Element { return New() }
