# Streaming Data

## Overview

This guide builds a live data viewer that receives streaming data from a Go channel, using channel watchers, scrollable containers, and auto-scroll. It builds on the [Watchers](watchers) and [Scrolling](scrolling) guides.

## The Producer Pattern

The streaming pattern starts outside the component: a goroutine produces data and sends it through a channel. The component receives the channel in its constructor and watches it for new data.

Create a buffered channel and pass it to the component:

```go
func main() {
    dataCh := make(chan string, 100)

    app, err := tui.NewApp(
        tui.WithRootComponent(Streaming(dataCh)),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    go produce(dataCh, app.StopCh())

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "App error: %v\n", err)
        os.Exit(1)
    }
}
```

The producer goroutine uses `app.StopCh()` for clean shutdown. When the app stops, `StopCh()` closes, and the goroutine exits:

```go
func produce(ch chan<- string, stopCh <-chan struct{}) {
    defer close(ch)

    for {
        select {
        case <-stopCh:
            return
        default:
        }

        line := generateLine()

        select {
        case <-stopCh:
            return
        case ch <- line:
        }

        time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)
    }
}
```

The double `select` pattern (one before work, one before send) ensures the goroutine responds promptly to shutdown signals.

## Channel Watcher for Streaming

The component watches the channel using `tui.Watch` in its `Watchers()` method. Each time a value arrives, the callback fires and the UI re-renders:

```go
func (s *streamingApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.Watch(s.dataCh, s.addLine),
    }
}

func (s *streamingApp) addLine(line string) {
    current := s.lines.Get()
    s.lines.Set(append(current, line))
}
```

The `lines` state holds a `[]string` that grows over time. Each new line triggers a re-render, which updates the scrollable content area.

## Auto-Scroll (Sticky Bottom)

A live data viewer should stay scrolled to the bottom as new data arrives, but let the user scroll up to read history. The `stickToBottom` state boolean controls this:

```go
type streamingApp struct {
    lines         *tui.State[[]string]
    scrollY       *tui.State[int]
    stickToBottom *tui.State[bool]
    content       *tui.Ref
    // ...
}
```

When `stickToBottom` is `true`, new lines set `scrollY` to `math.MaxInt` (the framework clamps this to the actual maximum):

```go
func (s *streamingApp) addLine(line string) {
    current := s.lines.Get()
    s.lines.Set(append(current, line))
    if s.stickToBottom.Get() {
        s.scrollY.Set(math.MaxInt)
    }
}
```

Manual scrolling disables auto-scroll. The `scrollBy` helper checks whether the user has scrolled to the bottom and updates `stickToBottom` accordingly:

```go
func (s *streamingApp) scrollBy(delta int) {
    el := s.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := s.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    } else if newY > maxY {
        newY = maxY
    }
    s.scrollY.Set(newY)
    s.stickToBottom.Set(newY >= maxY)
}
```

The Space key toggles auto-scroll, and End re-enables it:

```go
tui.OnRune(' ', func(ke tui.KeyEvent) {
    if s.stickToBottom.Get() {
        s.stickToBottom.Set(false)
    } else {
        s.scrollY.Set(math.MaxInt)
        s.stickToBottom.Set(true)
    }
}),
tui.OnKey(tui.KeyEnd, func(ke tui.KeyEvent) {
    s.scrollY.Set(math.MaxInt)
    s.stickToBottom.Set(true)
}),
```

## Dynamic Content Coloring

To visually distinguish different types of streaming data, write a helper function that inspects the line content and returns a Tailwind class string:

```go
func lineColor(line string) string {
    if len(line) < 20 {
        return ""
    }
    for i := 0; i < len(line)-3; i++ {
        sub := line[i : i+3]
        if sub == "cpu" {
            return "text-cyan"
        }
        if sub == "mem" {
            return "text-magenta"
        }
        if sub == "net" {
            return "text-green"
        }
        if sub == "dis" {
            return "text-yellow"
        }
        if sub == "io:" {
            return "text-blue"
        }
    }
    return ""
}
```

Use it in the template with a dynamic `class` attribute:

```gsx
for _, line := range s.lines.Get() {
    <span class={lineColor(line)}>{line}</span>
}
```

The `class` attribute accepts any expression that returns a string, so you can compute it per-element.

## Timer and Channel Together

A component can combine multiple watcher types. The streaming example uses both `OnTimer` (for an elapsed time counter) and `Watch` (for incoming data) in the same `Watchers()` slice:

```go
func (s *streamingApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, s.tick),
        tui.Watch(s.dataCh, s.addLine),
    }
}

func (s *streamingApp) tick() {
    s.elapsed.Set(s.elapsed.Get() + 1)
}
```

The timer fires every second regardless of channel activity. Both watchers trigger re-renders independently.

## Complete Example

This live stream viewer receives timestamped metrics from a background goroutine, with auto-scroll, manual scrolling, and color-coded output:

```gsx
package main

import (
    "fmt"
    "math"
    "time"
    tui "github.com/grindlemire/go-tui"
)

type streamingApp struct {
    dataCh        <-chan string
    lines         *tui.State[[]string]
    scrollY       *tui.State[int]
    stickToBottom *tui.State[bool]
    elapsed       *tui.State[int]
    content       *tui.Ref
}

func Streaming(dataCh <-chan string) *streamingApp {
    return &streamingApp{
        dataCh:        dataCh,
        lines:         tui.NewState([]string{}),
        scrollY:       tui.NewState(0),
        stickToBottom: tui.NewState(true),
        elapsed:       tui.NewState(0),
        content:       tui.NewRef(),
    }
}

func (s *streamingApp) scrollBy(delta int) {
    el := s.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := s.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    } else if newY > maxY {
        newY = maxY
    }
    s.scrollY.Set(newY)
    s.stickToBottom.Set(newY >= maxY)
}

func (s *streamingApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('j', func(ke tui.KeyEvent) { s.scrollBy(1) }),
        tui.OnRune('k', func(ke tui.KeyEvent) { s.scrollBy(-1) }),
        tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { s.scrollBy(-1) }),
        tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { s.scrollBy(1) }),
        tui.OnKey(tui.KeyPageUp, func(ke tui.KeyEvent) { s.scrollBy(-10) }),
        tui.OnKey(tui.KeyPageDown, func(ke tui.KeyEvent) { s.scrollBy(10) }),
        tui.OnKey(tui.KeyHome, func(ke tui.KeyEvent) {
            s.scrollY.Set(0)
            s.stickToBottom.Set(false)
        }),
        tui.OnKey(tui.KeyEnd, func(ke tui.KeyEvent) {
            s.scrollY.Set(math.MaxInt)
            s.stickToBottom.Set(true)
        }),
        tui.OnRune(' ', func(ke tui.KeyEvent) {
            if s.stickToBottom.Get() {
                s.stickToBottom.Set(false)
            } else {
                s.scrollY.Set(math.MaxInt)
                s.stickToBottom.Set(true)
            }
        }),
    }
}

func (s *streamingApp) HandleMouse(me tui.MouseEvent) bool {
    switch me.Button {
    case tui.MouseWheelUp:
        s.scrollBy(-1)
        return true
    case tui.MouseWheelDown:
        s.scrollBy(1)
        return true
    }
    return false
}

func (s *streamingApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, s.tick),
        tui.Watch(s.dataCh, s.addLine),
    }
}

func (s *streamingApp) tick() {
    s.elapsed.Set(s.elapsed.Get() + 1)
}

func (s *streamingApp) addLine(line string) {
    current := s.lines.Get()
    s.lines.Set(append(current, line))
    if s.stickToBottom.Get() {
        s.scrollY.Set(math.MaxInt)
    }
}

func lineColor(line string) string {
    if len(line) < 20 {
        return ""
    }
    for i := 0; i < len(line)-3; i++ {
        sub := line[i : i+3]
        if sub == "cpu" {
            return "text-cyan"
        }
        if sub == "mem" {
            return "text-magenta"
        }
        if sub == "net" {
            return "text-green"
        }
        if sub == "dis" {
            return "text-yellow"
        }
        if sub == "io:" {
            return "text-blue"
        }
    }
    return ""
}

templ (s *streamingApp) Render() {
    <div class="flex-col gap-1 p-1 h-full border-rounded border-cyan">
        <div class="flex justify-between shrink-0">
            <span class="text-gradient-cyan-magenta font-bold shrink-0">Live Stream</span>
            <span class="text-cyan font-bold" minWidth={0}>{fmt.Sprintf("%d lines", len(s.lines.Get()))}</span>
        </div>
        <div
            ref={s.content}
            class="flex-col flex-grow border-single p-1"
            scrollable={tui.ScrollVertical}
            scrollOffset={0, s.scrollY.Get()}
        >
            for _, line := range s.lines.Get() {
                <span class={lineColor(line)}>{line}</span>
            }
        </div>

        <div class="flex gap-2 shrink-0 justify-center">
            <span class="font-dim">Elapsed:</span>
            <span class="text-cyan font-bold">{fmt.Sprintf("%ds", s.elapsed.Get())}</span>
            <span class="font-dim">Auto-scroll:</span>
            if s.stickToBottom.Get() {
                <span class="text-green font-bold">ON</span>
            } else {
                <span class="text-yellow">OFF</span>
            }
        </div>

        <span class="font-dim shrink-0">j/k scroll | Space toggle auto-scroll | q quit</span>
    </div>
}
```

With `main.go`:

```go
package main

import (
    "fmt"
    "math/rand"
    "os"
    "time"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    dataCh := make(chan string, 100)

    app, err := tui.NewApp(
        tui.WithRootComponent(Streaming(dataCh)),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    go produce(dataCh, app.StopCh())

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "App error: %v\n", err)
        os.Exit(1)
    }
}

func produce(ch chan<- string, stopCh <-chan struct{}) {
    defer close(ch)

    metrics := []string{"cpu", "mem", "net", "disk", "io"}

    for {
        select {
        case <-stopCh:
            return
        default:
        }

        metric := metrics[rand.Intn(len(metrics))]
        value := rand.Intn(100)
        ts := time.Now().Format("15:04:05.000")

        var line string
        switch metric {
        case "cpu":
            line = fmt.Sprintf("[%s] cpu:  %d%%", ts, value)
        case "mem":
            line = fmt.Sprintf("[%s] mem:  %.1fG", ts, float64(value)/10.0)
        case "net":
            line = fmt.Sprintf("[%s] net:  %d req/s", ts, value*5)
        case "disk":
            line = fmt.Sprintf("[%s] disk: %d%% used", ts, 40+value/2)
        case "io":
            line = fmt.Sprintf("[%s] io:   %d MB/s", ts, value*2)
        }

        select {
        case <-stopCh:
            return
        case ch <- line:
        }

        time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

Watch the live metrics stream in, scroll up to pause auto-scroll:

![Streaming Data screenshot](/guides/16.png)

## Next Steps

- [Watchers Guide](watchers) - Timers, channels, and the WatcherProvider interface
- [Inline Streaming Guide](inline-streaming) - Stream character-by-character output above an inline widget
- [Building a Dashboard](dashboard) - Combine streaming, layout, and state into a full application
