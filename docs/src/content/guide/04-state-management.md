---
title: "State Management"
order: 4
---

## State[T]

The generic State[T] type provides reactive values that trigger re-renders:

```go
// Create state
count := tui.NewState(0)
name := tui.NewState("World")

// Read and write
current := count.Get()
count.Set(current + 1)

// Bind callbacks (called on change)
count.Bind(func(v int) {
    fmt.Println("Count changed to:", v)
})

// Batch updates (single re-render)
tui.Batch(func() {
    count.Set(10)
    name.Set("Go")
})
```

## Watchers

Components can provide watchers for timers, tickers, and channels:

```go
func (d *dashboard) Watchers() []tui.Watcher {
    return []tui.Watcher{
        // Fire every second
        tui.OnTimer(time.Second, d.tick),

        // Watch a channel
        tui.NewChannelWatcher(d.dataCh, d.onData),
    }
}
```
