package tui

import (
	"testing"
	"time"
)

type mockWatcherProvider struct{}

func (m *mockWatcherProvider) Render() *Element { return New() }
func (m *mockWatcherProvider) Watchers() []Watcher {
	return []Watcher{
		OnTimer(time.Second, func() {}),
	}
}

func TestWatcherProvider_Interface(t *testing.T) {
	var _ WatcherProvider = &mockWatcherProvider{}
}
