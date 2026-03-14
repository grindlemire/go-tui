# Timers, Watchers, and Channels

## Overview

Components can run background operations through the `WatcherProvider` interface. A watcher is a long-running goroutine: a timer that ticks every second, a channel that receives data from an API. Watchers start when your component mounts, stop when it unmounts, and their callbacks run on the main event loop so you can update state without synchronization.

## The WatcherProvider Interface

To give your component background operations, implement `Watchers() []tui.Watcher`:

```go
func (m *myApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, m.tick),
        tui.Watch(m.dataCh, m.onData),
    }
}
```

The framework calls `Watchers()` when your component mounts and starts each returned watcher. Each watcher spawns its own goroutine internally. When the app stops (via `app.Stop()` or Ctrl+C), all watchers receive a stop signal and shut down cleanly.

Watcher callbacks run on the main event loop, not in the watcher's goroutine. You can call `state.Set()`, `state.Update()`, and any other state mutation directly inside a callback without worrying about races or locks.

## Timers

`tui.OnTimer` fires a callback at a fixed interval:

```go
tui.OnTimer(interval time.Duration, handler func()) Watcher
```

The handler runs on the main event loop every time the interval elapses. It keeps firing until the component unmounts or the app stops.

Here's a stopwatch that counts seconds while running:

```gsx
package main

import (
    "fmt"
    "time"

    tui "github.com/grindlemire/go-tui"
)

type stopwatch struct {
    seconds *tui.State[int]
    running *tui.State[bool]
}

func Stopwatch() *stopwatch {
    return &stopwatch{
        seconds: tui.NewState(0),
        running: tui.NewState(false),
    }
}

func (s *stopwatch) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, s.tick),
    }
}

func (s *stopwatch) tick() {
    if s.running.Get() {
        s.seconds.Update(func(v int) int { return v + 1 })
    }
}

func (s *stopwatch) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune(' ', func(ke tui.KeyEvent) {
            s.running.Set(!s.running.Get())
        }),
        tui.OnRune('r', func(ke tui.KeyEvent) {
            s.seconds.Set(0)
            s.running.Set(false)
        }),
    }
}

templ (s *stopwatch) Render() {
    <div class="flex-col p-1 gap-1 border-rounded border-cyan items-center">
        <span class="font-bold text-gradient-cyan-magenta">Stopwatch</span>
        m := s.seconds.Get() / 60
        sec := s.seconds.Get() - m*60
        <span class="text-cyan font-bold">{fmt.Sprintf("%02d:%02d", m, sec)}</span>
        if s.running.Get() {
            <span class="text-green font-bold">Running</span>
        } else {
            <span class="text-yellow">Paused</span>
        }
        <span class="font-dim">space toggle | r reset | esc quit</span>
    </div>
}
```

The timer fires every second regardless of whether the stopwatch is running. The `tick` callback checks the `running` state and only increments when active. This is simpler than starting and stopping the timer itself.

## Channel Watchers

`tui.Watch` receives values from a Go channel and calls your handler for each one:

```go
tui.Watch[T any](ch <-chan T, handler func(T)) Watcher
```

The watcher reads from the channel in its own goroutine. When a value arrives, it queues the handler on the main event loop. If the channel closes, the watcher stops. If the app stops, the watcher exits even if the channel still has data.

Create a channel, start a producer goroutine, and let a channel watcher deliver the results:

```gsx
package main

import (
    "fmt"
    "time"

    tui "github.com/grindlemire/go-tui"
)

type feedApp struct {
    messages *tui.State[[]string]
    msgCh    <-chan string
}

func FeedApp(ch <-chan string) *feedApp {
    return &feedApp{
        messages: tui.NewState([]string{}),
        msgCh:    ch,
    }
}

func (f *feedApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.Watch(f.msgCh, f.addMessage),
    }
}

func (f *feedApp) addMessage(msg string) {
    current := f.messages.Get()
    // Keep last 10 messages
    if len(current) >= 10 {
        current = current[1:]
    }
    f.messages.Set(append(current, msg))
}

func (f *feedApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

templ (f *feedApp) Render() {
    <div class="flex-col p-1 gap-1 border-rounded border-cyan">
        <span class="font-bold text-gradient-cyan-magenta">Live Feed</span>
        for _, msg := range f.messages.Get() {
            <span class="text-green">{msg}</span>
        }
        if len(f.messages.Get()) == 0 {
            <span class="font-dim">Waiting for messages...</span>
        }
        <span class="font-dim">esc quit</span>
    </div>
}
```

With `main.go` that creates the channel and starts a producer:

```go
package main

import (
    "fmt"
    "os"
    "time"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    msgCh := make(chan string, 100)

    app, err := tui.NewApp(
        tui.WithRootComponent(FeedApp(msgCh)),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    // Start producer, using StopCh to know when to exit
    go func() {
        defer close(msgCh)
        i := 0
        for {
            select {
            case <-app.StopCh():
                return
            case msgCh <- fmt.Sprintf("[%s] Event #%d", time.Now().Format("15:04:05"), i):
                i++
            }
            time.Sleep(2 * time.Second)
        }
    }()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Note the use of `app.StopCh()`. This channel closes when the app stops, giving your producer goroutine a clean shutdown signal. Always check it in your producer's select to avoid goroutine leaks.

## Events Bus

`tui.Events[T]` is a typed broadcast bus routed by a topic key. Every call to `Emit` delivers the value to every subscriber on that same topic in the same app.

```go
bus := tui.NewEvents[string]("demo.events")
bus.Subscribe(func(msg string) { fmt.Println("got:", msg) })
bus.Emit("hello") // every subscriber receives "hello"
```

All subscribers run synchronously in registration order when `Emit` is called, and the UI is marked dirty afterward. If you need separate channels, use different topics.

The event bus must be bound to the app before `Emit` is called. The framework handles this automatically for `Events` fields on struct components through the generated `BindApp` method. If you create an event bus outside a component (say, in `main.go`), use `tui.NewEventsForApp(app, "topic")` or call `events.BindApp(app)` manually.

### Two components sharing a topic

The typical pattern: components independently create buses with the same topic string. One component emits, the other subscribes. Neither component knows about the other and no bus passing is needed.

```gsx
package main

import (
    "fmt"
    "time"

    tui "github.com/grindlemire/go-tui"
)

// --- Producer: emits actions onto the bus ---

type controls struct {
    bus     *tui.Events[string]
    counter *tui.State[int]
}

func Controls() *controls {
    return &controls{
        bus:     tui.NewEvents[string]("ci.events"),
        counter: tui.NewState(0),
    }
}

func (c *controls) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('1', func(ke tui.KeyEvent) { c.bus.Emit("build started") }),
        tui.OnRune('2', func(ke tui.KeyEvent) { c.bus.Emit("tests passed") }),
        tui.OnRune('3', func(ke tui.KeyEvent) { c.bus.Emit("deployed to staging") }),
    }
}

templ (c *controls) Render() {
    <div class="flex-col border-rounded p-1 gap-1">
        <span class="font-bold">Controls</span>
        <span class="font-dim">[1] build  [2] test  [3] deploy  [esc] quit</span>
    </div>
}

// --- Consumer: subscribes to the bus and displays events ---

type eventLog struct {
    bus      *tui.Events[string]
    messages *tui.State[[]string]
}

func EventLog() *eventLog {
    bus := tui.NewEvents[string]("ci.events")
    el := &eventLog{
        bus:      bus,
        messages: tui.NewState([]string{}),
    }
    el.bus.Subscribe(el.onEvent)
    return el
}

func (e *eventLog) onEvent(msg string) {
    ts := time.Now().Format("15:04:05")
    entry := fmt.Sprintf("[%s] %s", ts, msg)
    current := e.messages.Get()
    if len(current) >= 8 {
        current = current[1:]
    }
    e.messages.Set(append(current, entry))
}

templ (e *eventLog) Render() {
    <div class="flex-col border-rounded p-1 gap-1">
        <span class="font-bold">Event Log</span>
        for _, msg := range e.messages.Get() {
            <span class="text-green">{msg}</span>
        }
        if len(e.messages.Get()) == 0 {
            <span class="font-dim">No events yet</span>
        }
    </div>
}

// --- Root: no bus wiring needed ---

type app struct{}

func App() *app { return &app{} }

templ (a *app) Render() {
    <div class="flex-col p-1 gap-1 border-rounded border-cyan">
        <span class="text-gradient-cyan-magenta font-bold">Event Bus Demo</span>
        @Controls()
        @EventLog()
    </div>
}
```

The `controls` component emits strings onto the topic when you press a key. The `eventLog` component subscribes in its constructor and appends each message to a rolling list. Neither references the other. The root just mounts both components.

The event bus works well for fire-and-forget notifications and cross-component communication where the sender doesn't need to know who's listening.

## State Change Watchers

`tui.OnChange` watches a `State[T]` value and calls your handler when it changes:

```go
tui.OnChange[T any](state *State[T], handler func(T)) Watcher
```

The handler fires once on startup with the current value, then again each time the state changes. Like other watchers, it starts and stops with the component lifecycle.

Good for side effects that don't belong inside `Render()`. Scrolling a container to the bottom when new messages arrive, for instance:

```go
func (w *myApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnChange(w.messages, func(_ []string) {
            el := w.feed.El()
            if el != nil {
                _, maxY := el.MaxScroll()
                w.scrollY.Set(maxY + 1)
            }
        }),
    }
}
```

The handler runs synchronously during `State.Set()`, so it doesn't go through the event queue like `OnTimer` and `Watch` callbacks do. It runs on the main loop because `Set()` itself must be called from the main loop.

## Combining Watchers

A single component can return multiple watchers. Return them all from `Watchers()`:

```gsx
package main

import (
    "fmt"
    "time"

    tui "github.com/grindlemire/go-tui"
)

type dashboard struct {
    elapsed  *tui.State[int]
    messages *tui.State[[]string]
    msgCh    chan string
    msgCount *tui.State[int]
    feed     *tui.Ref
    scrollY  *tui.State[int]
}

func Dashboard() *dashboard {
    msgCh := make(chan string, 100)
    d := &dashboard{
        elapsed:  tui.NewState(0),
        messages: tui.NewState([]string{}),
        msgCh:    msgCh,
        msgCount: tui.NewState(0),
        feed:     tui.NewRef(),
        scrollY:  tui.NewState(0),
    }

    go func() {
        labels := []string{
            "Heartbeat OK",
            "Data synced",
            "Cache refreshed",
            "Checkpoint saved",
        }
        i := 0
        for {
            time.Sleep(3 * time.Second)
            msgCh <- labels[i%len(labels)]
            i++
        }
    }()

    return d
}

func (d *dashboard) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, d.tick),
        tui.Watch(d.msgCh, d.addMessage),
        tui.OnChange(d.messages, d.autoScrollToBottom),
    }
}

func (d *dashboard) tick() {
    d.elapsed.Update(func(v int) int { return v + 1 })
}

func (d *dashboard) addMessage(msg string) {
    d.msgCount.Update(func(v int) int { return v + 1 })
    ts := time.Now().Format("15:04:05")
    entry := fmt.Sprintf("[%s] #%d: %s", ts, d.msgCount.Get(), msg)
    current := d.messages.Get()
    if len(current) >= 10 {
        current = current[1:]
    }
    d.messages.Set(append(current, entry))
}

func (d *dashboard) autoScrollToBottom(_ []string) {
    el := d.feed.El()
    if el != nil {
        _, maxY := el.MaxScroll()
        d.scrollY.Set(maxY + 1)
    }
}

func (d *dashboard) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

templ (d *dashboard) Render() {
    <div class="flex-col p-1 gap-1 border-rounded border-cyan">
        <span class="font-bold text-gradient-cyan-magenta">Dashboard</span>
        <div class="flex gap-2">
            <span class="font-dim">Uptime:</span>
            <span class="text-cyan font-bold">{fmt.Sprintf("%ds", d.elapsed.Get())}</span>
            <span class="font-dim">Events:</span>
            <span class="text-cyan font-bold">{fmt.Sprintf("%d", d.msgCount.Get())}</span>
        </div>

        <div class="flex-col border-single p-1">
            <span class="font-bold">Event Log</span>
            for _, msg := range d.messages.Get() {
                <span class="text-green">{msg}</span>
            }
            if len(d.messages.Get()) == 0 {
                <span class="font-dim">Waiting for events...</span>
            }
        </div>

        <span class="font-dim">q quit</span>
    </div>
}
```

The timer handles the uptime counter, the channel watcher receives messages from the background producer, and `OnChange` keeps the feed scrolled to the bottom whenever new messages arrive. All three callbacks run on the main event loop, so they can update state without coordination.

## Thread Safety

Watcher callbacks are queued on the main event loop and run one at a time. This gives you a simple guarantee: inside any callback, you are the only code running, so you can update state directly without locks.

This applies to:
- `OnTimer` callbacks
- `Watch` channel handlers
- `OnChange` state change handlers
- Key handlers in `KeyMap()`
- Mouse handlers in `HandleMouse()`

`Events[T]` subscriber callbacks work slightly differently. They aren't queued through the event loop. Instead, `Emit` calls each subscriber directly at the call site. Since `Emit` should only be called from the main loop (see the table below), subscribers end up running on the main loop too, but through direct invocation rather than queuing.

For code that runs in your own goroutines outside the watcher system, you need `app.QueueUpdate()`:

```go
go func() {
    result := expensiveComputation()
    app.QueueUpdate(func() {
        myState.Set(result)
    })
}()
```

`QueueUpdate` is safe to call from any goroutine. It queues a function on the event loop, where it will run alongside watcher callbacks and input handlers. Inside that function, you can call `state.Set()` and `state.Update()` normally.

Remember: `state.Get()` is safe from any goroutine, but `state.Set()` and `state.Update()` must run on the main event loop. Watcher callbacks are already on the main loop. For anything else, wrap mutations in `QueueUpdate`.

| Operation | Main Loop | Any Goroutine |
|---|---|---|
| `state.Get()` | Yes | Yes |
| `state.Set()` | Yes | Use `QueueUpdate` |
| `state.Update()` | Yes | Use `QueueUpdate` |
| `events.Emit()` | Yes | Use `QueueUpdate` |
| `events.Subscribe()` | Yes | Yes |
| `app.QueueUpdate()` | Yes | Yes |

## Complete Example

This app combines a stopwatch timer with a channel-fed message stream. The timer ticks every second and conditionally increments the stopwatch. The channel watcher appends messages from a background producer. An `OnChange` watcher on the messages state keeps the feed scrolled to the bottom automatically.

```gsx
package main

import (
    "fmt"
    "time"

    tui "github.com/grindlemire/go-tui"
)

type watcherApp struct {
    stopwatchSec *tui.State[int]
    stopwatchOn  *tui.State[bool]
    messages     *tui.State[[]string]
    msgCh        chan string
    msgCount     *tui.State[int]
    feed         *tui.Ref
    scrollY      *tui.State[int]
}

func WatcherApp() *watcherApp {
    msgCh := make(chan string, 100)
    app := &watcherApp{
        stopwatchSec: tui.NewState(0),
        stopwatchOn:  tui.NewState(false),
        messages:     tui.NewState([]string{}),
        msgCh:        msgCh,
        msgCount:     tui.NewState(0),
        feed:         tui.NewRef(),
        scrollY:      tui.NewState(0),
    }

    go func() {
        msgs := []string{
            "Hello from producer",
            "Data packet received",
            "Processing complete",
            "Checkpoint saved",
            "Heartbeat OK",
        }
        i := 0
        for {
            time.Sleep(3 * time.Second)
            msgCh <- msgs[i%len(msgs)]
            i++
        }
    }()

    return app
}

func (w *watcherApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('s', func(ke tui.KeyEvent) {
            w.stopwatchOn.Set(!w.stopwatchOn.Get())
        }),
        tui.OnRune('r', func(ke tui.KeyEvent) {
            w.stopwatchSec.Set(0)
            w.stopwatchOn.Set(false)
        }),
    }
}

func (w *watcherApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, w.tick),
        tui.Watch(w.msgCh, w.addMessage),
        tui.OnChange(w.messages, w.autoScrollToBottom),
    }
}

func (w *watcherApp) tick() {
    if w.stopwatchOn.Get() {
        w.stopwatchSec.Update(func(v int) int { return v + 1 })
    }
}

func (w *watcherApp) addMessage(msg string) {
    w.msgCount.Update(func(v int) int { return v + 1 })
    ts := time.Now().Format("15:04:05")
    entry := fmt.Sprintf("[%s] #%d: %s", ts, w.msgCount.Get(), msg)
    w.messages.Set(append(w.messages.Get(), entry))
}

func (w *watcherApp) autoScrollToBottom(_ []string) {
    el := w.feed.El()
    if el != nil {
        _, maxY := el.MaxScroll()
        w.scrollY.Set(maxY + 1)
    }
}

func (w *watcherApp) scrollBy(delta int) {
    el := w.feed.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := w.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    }
    if newY > maxY {
        newY = maxY
    }
    w.scrollY.Set(newY)
}

func (w *watcherApp) HandleMouse(me tui.MouseEvent) bool {
    switch me.Button {
    case tui.MouseWheelUp:
        w.scrollBy(-3)
        return true
    case tui.MouseWheelDown:
        w.scrollBy(3)
        return true
    }
    return false
}

func formatDuration(seconds int) string {
    m := seconds / 60
    s := seconds - m*60
    return fmt.Sprintf("%02d:%02d", m, s)
}

templ (w *watcherApp) Render() {
    <div class="flex-col p-1 gap-1 border-rounded border-cyan">
        <span class="text-gradient-cyan-magenta font-bold">Timers & Watchers</span>

        <div class="flex gap-2 shrink-0">
            <span class="font-bold">Stopwatch</span>
            <span class="text-cyan font-bold">{formatDuration(w.stopwatchSec.Get())}</span>
            if w.stopwatchOn.Get() {
                <span class="text-green font-bold">Running</span>
            } else {
                <span class="text-yellow">Paused</span>
            }
            <span class="font-dim">[s] toggle  [r] reset</span>
        </div>

        <div class="flex-col border-rounded p-1 grow overflow-y-scroll scrollbar-cyan scrollbar-thumb-bright-cyan"
            ref={w.feed} scrollOffset={0, w.scrollY.Get()}>
            <div class="flex gap-2">
                <span class="font-bold">Live Feed</span>
                <span class="font-dim">{fmt.Sprintf("(%d received)", w.msgCount.Get())}</span>
            </div>
            for _, msg := range w.messages.Get() {
                <span class="text-green">{msg}</span>
            }
            if len(w.messages.Get()) == 0 {
                <span class="font-dim">Waiting for messages...</span>
            }
        </div>

        <span class="font-dim">s stopwatch | r reset | q quit</span>
    </div>
}
```

With `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(WatcherApp()),
        tui.WithMouse(),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

You'll see the clock ticking, the stopwatch ready to go, and the channel watcher waiting for messages. When messages arrive, the `OnChange` watcher scrolls the feed to the bottom automatically:

![Timers, Watchers, and Channels screenshot](/guides/12.png)

## Next Steps

- [Scrolling](scrolling) -- Scrollable containers for content that exceeds available space
- [Testing](testing) -- Testing components with MockTerminal and MockEventReader
