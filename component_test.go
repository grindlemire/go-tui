package tui

import (
	"testing"
	"time"
)

type mockWatcherProvider struct{}

func (m *mockWatcherProvider) Render(app *App) *Element { return New() }
func (m *mockWatcherProvider) Watchers() []Watcher {
	return []Watcher{
		OnTimer(time.Second, func() {}),
	}
}

func TestWatcherProvider_Interface(t *testing.T) {
	var _ WatcherProvider = &mockWatcherProvider{}
}

// Element satisfies Component so a prebuilt element can be used anywhere a
// component is accepted, including @expr in gsx.
var _ Component = (*Element)(nil)

func TestElement_Render_ReturnsSelf(t *testing.T) {
	app := newTestApp(20, 5)
	el := New(WithText("hi"))

	var comp Component = el
	got := comp.Render(app)
	if got != el {
		t.Fatalf("Element.Render(app) = %p, want the same element %p", got, el)
	}
}
