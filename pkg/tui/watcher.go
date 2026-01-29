package tui

import (
	"time"

	"github.com/grindlemire/go-tui/pkg/debug"
)

// Watcher represents a deferred event source that starts when the app runs.
// Watchers are collected during component construction and started by SetRoot.
type Watcher interface {
	// Start begins the watcher goroutine. Called by App.SetRoot().
	// The eventQueue channel and stopCh are provided by the App.
	Start(eventQueue chan<- func(), stopCh <-chan struct{})
}

// channelWatcher watches a channel and calls handler for each value.
type channelWatcher[T any] struct {
	ch      <-chan T
	handler func(T)
}

// Watch creates a channel watcher. The handler is called on the main loop
// whenever data arrives on the channel.
//
// Usage in .tui file:
//
//	onChannel={tui.Watch(dataCh, handleData(lines))}
//
// Handlers don't return bool - mutations automatically mark dirty.
func Watch[T any](ch <-chan T, handler func(T)) Watcher {
	return &channelWatcher[T]{ch: ch, handler: handler}
}

func (w *channelWatcher[T]) Start(eventQueue chan<- func(), stopCh <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case val, ok := <-w.ch:
				if !ok {
					return // Channel closed
				}
				// Capture val for closure
				v := val
				select {
				case eventQueue <- func() { w.handler(v) }:
				case <-stopCh:
					return
				}
			}
		}
	}()
}

// timerWatcher fires at a regular interval.
type timerWatcher struct {
	interval time.Duration
	handler  func()
}

// OnTimer creates a timer watcher that fires at the given interval.
// The handler is called on the main loop.
//
// Usage in .tui file:
//
//	onTimer={tui.OnTimer(time.Second, tick(elapsed))}
//
// Handlers don't return bool - mutations automatically mark dirty.
func OnTimer(interval time.Duration, handler func()) Watcher {
	return &timerWatcher{interval: interval, handler: handler}
}

func (w *timerWatcher) Start(eventQueue chan<- func(), stopCh <-chan struct{}) {
	go func() {
		debug.Log("timerWatcher started")
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				debug.Log("timerWatcher ticked")
				select {
				case eventQueue <- w.handler:
				case <-stopCh:
					return
				}
			}
		}
	}()
}
