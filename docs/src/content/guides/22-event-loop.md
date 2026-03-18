# Custom Event Loops

## Overview

go-tui has three ways to drive its event loop. `Run()` works for most apps, but when you need to integrate external event sources like LLM streaming, network I/O, or background workers, you can take control of the loop using `Open`, `Step`, `Events`, `Dispatch`, `Render`, and `Close`.

## The Standard Loop (Run)

`Run()` manages signal setup, input reading, event dispatch, dirty checking, rendering, and frame timing. Background goroutines push data into the UI through `QueueUpdate`.

```go
func runMode() {
    comp := NewFeedApp("Run()")
    app, err := tui.NewApp(tui.WithRootComponent(comp))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    // The producer must use QueueUpdate to safely mutate state from its
    // goroutine. This is the only way to get external data into the UI
    // when Run() owns the event loop.
    go func() {
        for i := 1; ; i++ {
            time.Sleep(200 * time.Millisecond)
            if comp.IsPaused() {
                continue
            }
            msg := fmt.Sprintf("[%s] Message #%d", time.Now().Format("15:04:05.000"), i)
            app.QueueUpdate(func() {
                comp.AddMessage(msg)
            })
        }
    }()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Because `Run()` owns the event loop, all external data must funnel through `QueueUpdate` so that state mutations happen on the main goroutine.

## Owning the Frame Loop (Step)

`Open()` performs the same startup work as `Run()` (registering signals, starting the input reader, performing the initial render) but returns immediately instead of blocking.

After that, your code controls when frames happen. `Step()` combines `DispatchEvents()` and `Render()` into a single call. Between steps you can read from your own channels and mutate state directly, since you are the main goroutine.

```go
func stepMode() {
    comp := NewFeedApp("Step()")
    app, err := tui.NewApp(tui.WithRootComponent(comp))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if err := app.Open(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    msgCh := startProducer(comp.IsPaused)
    ticker := time.NewTicker(16 * time.Millisecond)
    defer ticker.Stop()

    for {
        // Wait for next frame (acts as a frame rate limiter)
        select {
        case <-ticker.C:
        case <-app.StopCh():
            return
        }

        // Drain pending messages from the producer. This runs on the
        // main goroutine so we can mutate state directly.
    drain:
        for {
            select {
            case msg := <-msgCh:
                comp.AddMessage(msg)
            default:
                break drain
            }
        }

        if !app.Step() {
            return
        }
    }
}
```

The ticker controls frame rate. Between ticks, the drain loop pulls all pending messages from the producer channel and updates state without `QueueUpdate`. `Step()` returns `false` when the app should exit.

`Close()` restores terminal state and is safe to call multiple times.

## Full Control with Select (Events)

`Events()` returns a read-only channel you can use in a standard Go `select`, placing your external channels alongside go-tui events as peers.

```go
func selectMode() {
    comp := NewFeedApp("Select()")
    app, err := tui.NewApp(tui.WithRootComponent(comp))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if err := app.Open(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    msgCh := startProducer(comp.IsPaused)

    for {
        select {
        case ev := <-app.Events():
            app.Dispatch(ev)
        case msg := <-msgCh:
            comp.AddMessage(msg)
        case <-app.StopCh():
            return
        }
        app.Render()
    }
}
```

`Dispatch(ev)` routes the event through the key/mouse/resize dispatch system. `Render()` checks for dirty state and redraws if needed, so calling it after every select case is fine since it short-circuits when nothing changed.

This is the cleanest option when you have external event sources, because each source gets its own select case instead of funneling through `QueueUpdate`.

## When to Use Which

Start with **`Run()`**. It handles frame timing, signal setup, and event dispatch internally. If your app receives external data, `QueueUpdate` and channel watchers cover most cases without leaving `Run()`.

Use **`Step()`** when you need to control frame timing yourself, for example to implement variable frame rates, skip rendering during heavy computation, or pause the render loop entirely while waiting for a resource. It also lets you drain your own channels between frames without `QueueUpdate`, which avoids the overhead of serializing closures through the event queue.

Use **`Events()` + select** when you are building something that is driven by multiple event sources at once, like a chat client that handles keyboard input, incoming messages, and connection status changes in a single loop. Each source gets its own select case, which is the standard Go pattern for multiplexing channels.

## Complete Example

The UI component lives in `feed.gsx` and displays a scrollable message feed with pause/resume and sticky-bottom scrolling.

```gsx
package main

import (
    "fmt"
    "math"
    tui "github.com/grindlemire/go-tui"
)

type feedApp struct {
    messages      *tui.State[[]string]
    paused        *tui.State[bool]
    scrollY       *tui.State[int]
    stickToBottom *tui.State[bool]
    content       *tui.Ref
    mode          string
}

func NewFeedApp(mode string) *feedApp {
    return &feedApp{
        messages:      tui.NewState([]string{}),
        paused:        tui.NewState(false),
        scrollY:       tui.NewState(0),
        stickToBottom: tui.NewState(false),
        content:       tui.NewRef(),
        mode:          mode,
    }
}

func (f *feedApp) scrollBy(delta int) {
    el := f.content.El()
    if el == nil {
        return
    }
    _, curY := el.ScrollOffset()
    _, maxY := el.MaxScroll()
    newY := curY + delta
    if newY < 0 {
        newY = 0
    } else if newY > maxY {
        newY = maxY
    }
    f.scrollY.Set(newY)
    f.stickToBottom.Set(false)
}

func (f *feedApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnStop(tui.Rune('p'), func(ke tui.KeyEvent) {
            f.paused.Set(!f.paused.Get())
        }),
        tui.OnStop(tui.Rune('s'), func(ke tui.KeyEvent) {
            if f.stickToBottom.Get() {
                if el := f.content.El(); el != nil {
                    _, y := el.ScrollOffset()
                    f.scrollY.Set(y)
                }
                f.stickToBottom.Set(false)
            } else {
                f.stickToBottom.Set(true)
            }
        }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { f.scrollBy(1) }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { f.scrollBy(-1) }),
        tui.On(tui.KeyUp, func(ke tui.KeyEvent) { f.scrollBy(-1) }),
        tui.On(tui.KeyDown, func(ke tui.KeyEvent) { f.scrollBy(1) }),
        tui.On(tui.KeyPageUp, func(ke tui.KeyEvent) { f.scrollBy(-10) }),
        tui.On(tui.KeyPageDown, func(ke tui.KeyEvent) { f.scrollBy(10) }),
        tui.On(tui.KeyHome, func(ke tui.KeyEvent) {
            f.scrollY.Set(0)
            f.stickToBottom.Set(false)
        }),
    }
}

func (f *feedApp) HandleMouse(me tui.MouseEvent) bool {
    switch me.Button {
    case tui.MouseWheelUp:
        f.scrollBy(-1)
        return true
    case tui.MouseWheelDown:
        f.scrollBy(1)
        return true
    }
    return false
}

func (f *feedApp) AddMessage(msg string) {
    f.messages.Update(func(msgs []string) []string {
        return append(msgs, msg)
    })
    if f.stickToBottom.Get() {
        f.scrollY.Set(math.MaxInt)
    }
}

func (f *feedApp) IsPaused() bool {
    return f.paused.Get()
}

templ (f *feedApp) Render() {
    <div class="flex-col h-full border-rounded border-cyan">
        <div class="flex justify-between px-1 shrink-0">
            <span class="text-gradient-cyan-magenta font-bold">Event Loop Demo</span>
            <div class="flex gap-1">
                <span class="font-dim">mode:</span>
                <span class="text-cyan font-bold">{f.mode}</span>
            </div>
        </div>
        <hr />
        <div
            ref={f.content}
            class="flex-col flex-grow border-single p-1"
            scrollable={tui.ScrollVertical}
            scrollOffset={0, f.scrollY.Get()}
        >
            for _, msg := range f.messages.Get() {
                <span class="font-dim">{msg}</span>
            }
        </div>
        <hr />
        <div class="flex justify-between px-1 shrink-0">
            <div class="flex gap-2">
                <span class="font-dim">p: pause</span>
                <span class="font-dim">s: sticky</span>
                <span class="font-dim">j/k: scroll</span>
                <span class="font-dim">q: quit</span>
            </div>
        </div>
    </div>
}
```

The `main.go` file contains the `startProducer` helper and all three modes:

```go
package main

import (
    "fmt"
    "os"
    "time"

    tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate feed.gsx

func main() {
    mode := "run"
    if len(os.Args) > 1 {
        mode = os.Args[1]
    }

    switch mode {
    case "run":
        runMode()
    case "step":
        stepMode()
    case "select":
        selectMode()
    default:
        fmt.Fprintf(os.Stderr, "Unknown mode %q. Use: run, step, or select\n", mode)
        os.Exit(1)
    }
}

func startProducer(paused func() bool) <-chan string {
    ch := make(chan string, 10)
    go func() {
        for i := 1; ; i++ {
            time.Sleep(200 * time.Millisecond)
            if paused() {
                continue
            }
            select {
            case ch <- fmt.Sprintf("[%s] Message #%d", time.Now().Format("15:04:05.000"), i):
            default:
            }
        }
    }()
    return ch
}

// ... runMode, stepMode, selectMode as shown above ...
```

Generate and run:

```bash
tui generate ./...
go run . run       # Standard Run() mode
go run . step      # Step-based loop
go run . select    # Select-based loop
```

All three modes produce the same UI, and the difference is how the event loop is wired:

![Event Loop Demo screenshot](/guides/22.png)

## Next Steps

- [Streaming Data](streaming) for channel watchers, auto-scroll, and the producer pattern
- [Watchers](watchers) for timers, channels, and the WatcherProvider interface
- [App Reference](../reference/app) for full documentation of Open, Step, Events, Dispatch, Render, and Close
