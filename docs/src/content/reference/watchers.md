# Watchers Reference

## Overview

Watchers connect Go's concurrency primitives to the go-tui event loop. A watcher is a goroutine that listens for something (a timer tick, a channel message) and queues a handler function on the main event loop when data arrives.

Because handlers run on the event loop, they can safely call `state.Set()`, `state.Update()`, and any other state mutation without synchronization. The watcher goroutine itself never touches UI state directly.

Watchers start when their owning component mounts (or when the element tree is set as the app root) and stop when the component unmounts or the app shuts down.

## Watcher Interface

```go
type Watcher interface {
    Start(eventQueue chan<- func(), stopCh <-chan struct{})
}
```

All watcher types implement this single-method interface.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `eventQueue` | `chan<- func()` | The app's event loop queue. Send a `func()` to have it run on the main thread. |
| `stopCh` | `<-chan struct{}` | Closed when the watcher should shut down. Select on this to exit the goroutine. |

You rarely call `Start` yourself. The framework calls it during component mounting after walking the element tree with `WalkWatchers`. If you're building a custom watcher, implement this interface and launch a goroutine that selects on both your data source and `stopCh`.

```go
type myWatcher struct {
    handler func()
}

func (w *myWatcher) Start(eventQueue chan<- func(), stopCh <-chan struct{}) {
    go func() {
        for {
            select {
            case <-stopCh:
                return
            case <-someCondition:
                select {
                case eventQueue <- w.handler:
                case <-stopCh:
                    return
                }
            }
        }
    }()
}
```

The double-select pattern — first waiting for data, then trying to enqueue — prevents the goroutine from blocking on a full event queue when the app is shutting down.

## OnTimer

```go
func OnTimer(interval time.Duration, handler func()) Watcher
```

Creates a watcher that fires `handler` at a fixed interval. The handler runs on the main event loop.

Under the hood, `OnTimer` uses `time.NewTicker`. The ticker is cleaned up automatically when `stopCh` closes.

```go
func (s *stopwatch) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, func() {
            s.elapsed.Update(func(v int) int { return v + 1 })
        }),
    }
}
```

The interval is approximate — the handler won't fire faster than the app's frame rate, and a backed-up event queue can introduce extra delay. For UI animation, intervals below 16ms (60fps) rarely produce visible improvement.

## ChannelWatcher

### NewChannelWatcher

```go
func NewChannelWatcher[T any](ch <-chan T, fn func(T)) *ChannelWatcher[T]
```

Creates a watcher that reads from a Go channel and calls `fn` for each value received. The handler runs on the main event loop, not in the goroutine that reads the channel.

When the channel closes, the watcher exits its goroutine. When `stopCh` closes (app shutdown), the watcher also exits.

```go
dataCh := make(chan string)
w := tui.NewChannelWatcher(dataCh, func(s string) {
    messages.Update(func(list []string) []string {
        return append(list, s)
    })
})
```

### Watch

```go
func Watch[T any](ch <-chan T, handler func(T)) Watcher
```

A convenience wrapper around `NewChannelWatcher`. Returns the `Watcher` interface directly, which is what `Watchers()` expects.

```go
func (f *feed) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.Watch(f.dataCh, func(item DataItem) {
            f.items.Update(func(list []DataItem) []DataItem {
                return append(list, item)
            })
        }),
    }
}
```

Use `NewChannelWatcher` when you need the concrete `*ChannelWatcher[T]` type; use `Watch` when you just need a `Watcher` to return from `Watchers()`.

### Start

```go
func (w *ChannelWatcher[T]) Start(eventQueue chan<- func(), stopCh <-chan struct{})
```

Launches the goroutine that reads from the channel. Called automatically by the framework. The goroutine exits when either the channel closes or `stopCh` is closed.

## WatcherProvider Interface

```go
type WatcherProvider interface {
    Watchers() []Watcher
}
```

Implement this on a struct component to attach watchers to its lifecycle. The framework calls `Watchers()` after the component mounts and starts each returned watcher.

When the component unmounts, the framework closes the stop channel, which signals all watchers to exit.

```go
type dashboard struct {
    elapsed  *tui.State[int]
    messages *tui.State[[]string]
    dataCh   <-chan string
}

func Dashboard(dataCh <-chan string) *dashboard {
    return &dashboard{
        elapsed:  tui.NewState(0),
        messages: tui.NewState([]string{}),
        dataCh:   dataCh,
    }
}

func (d *dashboard) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, func() {
            d.elapsed.Update(func(v int) int { return v + 1 })
        }),
        tui.Watch(d.dataCh, func(msg string) {
            d.messages.Update(func(list []string) []string {
                return append(list, msg)
            })
        }),
    }
}
```

## Element-Level Watchers

Watchers can also be attached directly to elements, outside the component model.

### AddWatcher

```go
func (e *Element) AddWatcher(w Watcher)
```

Attaches a watcher to a specific element. The watcher starts when the element tree is set as the app root.

### Watchers

```go
func (e *Element) Watchers() []Watcher
```

Returns all watchers attached to the element.

### WalkWatchers

```go
func (e *Element) WalkWatchers(fn func(Watcher))
```

Recursively walks the element tree and calls `fn` for each attached watcher. Skips hidden elements and their subtrees. The framework uses this during `applyRoot` to discover and start all watchers.

## Lifecycle

Watchers follow this lifecycle:

1. Watchers are created in `Watchers()` or attached via `AddWatcher`.
2. The framework calls `Start(eventQueue, stopCh)` after the component mounts or the root is set. A goroutine begins running.
3. The goroutine loops, sending handler functions to the event queue.
4. When the component unmounts, the root changes, or the app stops, the framework closes `stopCh`. The goroutine detects the closed channel and exits.

Each time the root changes (via `SetRoot`, `SetRootView`, or `SetRootComponent`), the previous stop channel is closed and a new one is created. This ensures old watchers don't leak.

## Thread Safety

Watcher handlers run on the main event loop. This makes them safe for state mutation:

- Call `state.Set()` and `state.Update()` freely inside handlers.
- Access element properties without locks.
- Emit events via `Events[T].Emit()`.

If you have a background goroutine that is *not* a watcher, use `app.QueueUpdate` to safely run code on the event loop:

```go
go func() {
    result := expensiveComputation()
    app.QueueUpdate(func() {
        data.Set(result)
    })
}()
```

## See Also

- [State Reference](state.md) — reactive state that watcher handlers typically modify
- [App Reference](app.md) — app lifecycle, `QueueUpdate`, component mounting
- [Events Reference](events.md) — keyboard and mouse event handling
