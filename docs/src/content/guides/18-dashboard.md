# Building a Dashboard

## What We're Building

We're going to build a live metrics dashboard: CPU, memory, and disk gauges, a network sparkline, and a scrollable streaming event feed, all updating in real time. It pulls in concepts from most of the previous guides.

Concepts used:

- **State** ([Guide 05](state)): reactive `State[T]` for metrics, sparkline data, and event lists
- **Components** ([Guide 06](components)): a struct component with constructor and render method
- **Events** ([Guide 07](events)): `KeyMap` for quit and scroll bindings, `HandleMouse` for scroll wheel
- **Watchers** ([Guide 09](watchers)): `OnTimer` for metric animation and `Watch` for channel-based event streaming
- **Scrolling** ([Guide 10](scrolling)): scrollable event feed with keyboard and mouse control
- **Styling** ([Guide 03](styling)): gradients, conditional color classes, borders
- **Layout** ([Guide 04](layout)): nested flex containers, gap, grow, padding

## Project Setup

Create a new directory and initialize the module:

```bash
mkdir dashboard && cd dashboard
go mod init dashboard
go get github.com/grindlemire/go-tui
```

You'll create two files:

- `dashboard.gsx` -- the component and all its logic
- `main.go` -- the entry point that wires everything up

## Layout Skeleton

Start with the outer structure. The dashboard is a vertical stack inside a bordered container, with sections for the title, metrics, a row containing the network chart and event feed side by side, and a key hint at the bottom.

Create `dashboard.gsx`:

```gsx
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type dashboardApp struct {
}

func Dashboard() *dashboardApp {
    return &dashboardApp{}
}

func (d *dashboardApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

templ (d *dashboardApp) Render() {
    <div class="flex-col p-1 gap-1 h-full border-rounded border-cyan">
        <div class="flex justify-center shrink-0">
            <span class="text-gradient-cyan-magenta font-bold">Dashboard</span>
        </div>

        // Metrics will go here

        // Network + Events will go here (side by side)

        <div class="flex justify-center shrink-0">
            <span class="font-dim">j/k scroll events | q to quit</span>
        </div>
    </div>
}
```

And `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(Dashboard()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "App error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run to check that the skeleton renders:

```bash
tui generate ./...
go run .
```

You should see a bordered box with "Dashboard" centered at the top and "q to quit" at the bottom. Press `q` or Escape to exit.

A few things to notice about the layout:

- The outer `<div>` uses `flex-col` to stack children vertically, with `h-full` to fill the terminal and `border-rounded border-cyan` for the outer frame.
- `shrink-0` on the title and hint bar prevents them from collapsing when the terminal is small. The content sections between them will flex to fill the remaining space.
- `justify-center` on the title and hint containers centers their text horizontally.

## Adding Metrics

Now add reactive state for three gauges: CPU, memory, and disk. Each is a `State[int]` representing a percentage.

Update the struct and constructor:

```gsx
type dashboardApp struct {
    cpu  *tui.State[int]
    mem  *tui.State[int]
    disk *tui.State[int]
}

func Dashboard() *dashboardApp {
    return &dashboardApp{
        cpu:  tui.NewState(45),
        mem:  tui.NewState(62),
        disk: tui.NewState(38),
    }
}
```

Add two helper functions that convert a percentage into a visual bar and a color class. These are regular Go functions, not components:

```gsx
func metricBar(value, max int) string {
    width := 20
    filled := value * width / max
    bar := ""
    for i := 0; i < width; i++ {
        if i < filled {
            bar += "█"
        } else {
            bar += "░"
        }
    }
    return bar
}

func metricColor(value int) string {
    if value >= 80 {
        return "text-red font-bold"
    }
    if value >= 60 {
        return "text-yellow"
    }
    return "text-green"
}
```

`metricBar` builds a 20-character bar using filled (`█`) and empty (`░`) block characters. `metricColor` returns a class string that shifts from green to yellow to red as the value rises. Because `class` accepts Go expressions, you can return different class strings based on state.

Now add the metric panels to the render method, replacing the `// Metrics will go here` comment:

```gsx
// Metric gauges
<div class="flex gap-1 shrink-0">
    <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
        <span class="text-gradient-cyan-magenta font-bold">CPU</span>
        <span class={metricColor(d.cpu.Get())}>{metricBar(d.cpu.Get(), 100)}</span>
        <span class={metricColor(d.cpu.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.cpu.Get())}</span>
    </div>
    <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
        <span class="text-gradient-cyan-magenta font-bold">Memory</span>
        <span class={metricColor(d.mem.Get())}>{metricBar(d.mem.Get(), 100)}</span>
        <span class={metricColor(d.mem.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.mem.Get())}</span>
    </div>
    <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
        <span class="text-gradient-cyan-magenta font-bold">Disk</span>
        <span class={metricColor(d.disk.Get())}>{metricBar(d.disk.Get(), 100)}</span>
        <span class={metricColor(d.disk.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.disk.Get())}</span>
    </div>
</div>
```

Each gauge is a `flex-col` container with a title, bar, and percentage label. The three panels sit side-by-side in a `flex` (horizontal) row, each with `flexGrow={1.0}` so they share the available width equally. The `border-rounded` on each panel gives them individual frames.

Run `tui generate ./...` and `go run .` to see the three gauges. They display static values for now.

### Animating the Metrics

To make the gauges move, add a timer watcher. Import `"math/rand"` and `"time"`, then implement `WatcherProvider`:

```gsx
import (
    "fmt"
    "math/rand"
    "time"
    tui "github.com/grindlemire/go-tui"
)
```

```go
func (d *dashboardApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(500*time.Millisecond, d.updateMetrics),
    }
}

func (d *dashboardApp) updateMetrics() {
    d.cpu.Set(clampVal(d.cpu.Get()+rand.Intn(11)-5, 5, 95))
    d.mem.Set(clampVal(d.mem.Get()+rand.Intn(7)-3, 20, 90))
    d.disk.Set(clampVal(d.disk.Get()+rand.Intn(3)-1, 20, 80))
}

func clampVal(v, min, max int) int {
    if v < min {
        return min
    }
    if v > max {
        return max
    }
    return v
}
```

`OnTimer` fires `updateMetrics` every 500 milliseconds. Each call nudges the metric values by a random amount, clamped to stay within bounds. Because `Set` marks the state as dirty, the framework re-renders automatically after each update. The `clampVal` helper keeps values from wandering out of range.

Regenerate and run. The bars should shift every half second, with colors changing as values cross the 60% and 80% thresholds.

## Network Sparkline

The network section shows a scrolling sparkline for inbound and outbound traffic. Add state for the current rates and the sparkline data arrays:

```gsx
type dashboardApp struct {
    cpu      *tui.State[int]
    mem      *tui.State[int]
    disk     *tui.State[int]
    netIn    *tui.State[int]
    netOut   *tui.State[int]
    sparkIn  *tui.State[[]int]
    sparkOut *tui.State[[]int]
}

func Dashboard() *dashboardApp {
    return &dashboardApp{
        cpu:      tui.NewState(45),
        mem:      tui.NewState(62),
        disk:     tui.NewState(38),
        netIn:    tui.NewState(142),
        netOut:   tui.NewState(89),
        sparkIn:  tui.NewState([]int{3, 5, 4, 6, 7, 5, 4, 3, 5, 6, 7, 8, 6, 5, 4, 3, 5, 6, 7, 5}),
        sparkOut: tui.NewState([]int{2, 3, 4, 3, 5, 4, 3, 2, 3, 4, 5, 6, 4, 3, 2, 3, 4, 5, 4, 3}),
    }
}
```

The sparkline data is a fixed-length slice of integers. Each value maps to a Unicode block character that represents the height of that data point. Add the `sparkline` helper:

```gsx
func sparkline(data []int) string {
    blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
    maxVal := 1
    for _, v := range data {
        if v > maxVal {
            maxVal = v
        }
    }
    s := ""
    for _, v := range data {
        idx := v * 7 / maxVal
        if idx > 7 {
            idx = 7
        }
        s += string(blocks[idx])
    }
    return s
}
```

This normalizes values against the current maximum, so the sparkline always uses the full vertical range of the block characters (`▁` through `█`).

Add the network panel to the render method. We'll place it inside a horizontal row that will also hold the events panel (added in the next section), replacing `// Network + Events will go here`:

```gsx
// Network Traffic + Recent Events
<div class="flex gap-1 flex-grow">
    <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
        <span class="text-gradient-cyan-magenta font-bold">Network Traffic</span>
        <div class="flex gap-1">
            <span class="font-dim">In: </span>
            <span class="text-cyan">{sparkline(d.sparkIn.Get())}</span>
        </div>
        <div class="flex gap-1">
            <span class="font-dim">Out:</span>
            <span class="text-magenta">{sparkline(d.sparkOut.Get())}</span>
        </div>
        <div class="flex gap-2">
            <span class="text-cyan font-bold">{fmt.Sprintf("In: %d MB/s", d.netIn.Get())}</span>
            <span class="text-magenta font-bold">{fmt.Sprintf("Out: %d MB/s", d.netOut.Get())}</span>
        </div>
    </div>

    // Events panel will go here
</div>
```

The outer `<div class="flex gap-1 flex-grow">` is a horizontal row that fills the remaining vertical space. The network panel uses `flexGrow={1.0}` so it shares the width equally with the events panel we'll add next.

Update `updateMetrics` to include the network values and shift the sparkline data:

```go
func (d *dashboardApp) updateMetrics() {
    d.cpu.Set(clampVal(d.cpu.Get()+rand.Intn(11)-5, 5, 95))
    d.mem.Set(clampVal(d.mem.Get()+rand.Intn(7)-3, 20, 90))
    d.disk.Set(clampVal(d.disk.Get()+rand.Intn(3)-1, 20, 80))
    d.netIn.Set(clampVal(d.netIn.Get()+rand.Intn(41)-20, 50, 300))
    d.netOut.Set(clampVal(d.netOut.Get()+rand.Intn(31)-15, 30, 200))

    // Shift sparkline data left and append new point
    inData := d.sparkIn.Get()
    inData = append(inData[1:], d.netIn.Get()/30)
    d.sparkIn.Set(inData)

    outData := d.sparkOut.Get()
    outData = append(outData[1:], d.netOut.Get()/30)
    d.sparkOut.Set(outData)
}
```

Each tick drops the oldest data point (`inData[1:]`) and appends the latest value scaled down to the sparkline range. This creates a scrolling chart effect.

## Event Feed

The event feed receives messages from outside the component through a Go channel, using the `Watch` watcher to pipe goroutine-produced data into the UI. The feed is scrollable so you can keep a longer history and scroll back through it.

Add the channel field, event state, scroll state, and a ref to the struct:

```gsx
type dashboardApp struct {
    cpu       *tui.State[int]
    mem       *tui.State[int]
    disk      *tui.State[int]
    netIn     *tui.State[int]
    netOut    *tui.State[int]
    sparkIn   *tui.State[[]int]
    sparkOut  *tui.State[[]int]
    events    *tui.State[[]string]
    eventCh   <-chan string
    scrollY   *tui.State[int]
    eventsRef *tui.Ref
}

func Dashboard(eventCh <-chan string) *dashboardApp {
    return &dashboardApp{
        cpu:       tui.NewState(45),
        mem:       tui.NewState(62),
        disk:      tui.NewState(38),
        netIn:     tui.NewState(142),
        netOut:    tui.NewState(89),
        sparkIn:   tui.NewState([]int{3, 5, 4, 6, 7, 5, 4, 3, 5, 6, 7, 8, 6, 5, 4, 3, 5, 6, 7, 5}),
        sparkOut:  tui.NewState([]int{2, 3, 4, 3, 5, 4, 3, 2, 3, 4, 5, 6, 4, 3, 2, 3, 4, 5, 4, 3}),
        events:    tui.NewState([]string{}),
        eventCh:   eventCh,
        scrollY:   tui.NewState(0),
        eventsRef: tui.NewRef(),
    }
}
```

The constructor now takes an `eventCh` parameter. The channel is receive-only (`<-chan string`) inside the component. The producer runs elsewhere. The `scrollY` state tracks the current scroll position, and `eventsRef` is a ref to the scrollable container so we can query its maximum scroll offset.

Add a scroll helper and update `KeyMap` with scroll bindings:

```go
func (d *dashboardApp) scrollBy(delta int) {
    el := d.eventsRef.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := d.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    } else if newY > maxY {
        newY = maxY
    }
    d.scrollY.Set(newY)
}

func (d *dashboardApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('j', func(ke tui.KeyEvent) { d.scrollBy(1) }),
        tui.OnRune('k', func(ke tui.KeyEvent) { d.scrollBy(-1) }),
        tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { d.scrollBy(1) }),
        tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { d.scrollBy(-1) }),
    }
}
```

`scrollBy` clamps the new offset between 0 and the container's maximum scroll value. `MaxScroll()` returns the furthest the content can scroll given its total height versus the visible area.

Add mouse wheel support with `HandleMouse`:

```go
func (d *dashboardApp) HandleMouse(me tui.MouseEvent) bool {
    switch me.Button {
    case tui.MouseWheelUp:
        d.scrollBy(-1)
        return true
    case tui.MouseWheelDown:
        d.scrollBy(1)
        return true
    }
    return false
}
```

Add a `Watch` call alongside the timer in `Watchers`:

```go
func (d *dashboardApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(500*time.Millisecond, d.updateMetrics),
        tui.Watch(d.eventCh, d.addEvent),
    }
}
```

`Watch` receives from the channel and calls `addEvent` on the UI thread for each value. No manual synchronization needed.

Add the handler that timestamps and stores events, with auto-scroll to keep the latest event visible:

```go
func (d *dashboardApp) addEvent(event string) {
    current := d.events.Get()
    ts := time.Now().Format("15:04:05")
    entry := fmt.Sprintf("%s  %s", ts, event)
    current = append(current, entry)
    if len(current) > 50 {
        current = current[len(current)-50:]
    }
    d.events.Set(current)

    // Auto-scroll to bottom
    el := d.eventsRef.El()
    if el != nil {
        _, maxY := el.MaxScroll()
        d.scrollY.Set(maxY + 1)
    }
}
```

This keeps the last 50 events. After appending, it auto-scrolls to the bottom so the newest event is always visible. If the user has scrolled up manually, the next event will snap back to the bottom.

Add the events panel inside the network row (replacing `// Events panel will go here`), right after the network traffic `</div>`:

```gsx
<div
    ref={d.eventsRef}
    class="flex-col border-rounded p-1 gap-1"
    flexGrow={1.0}
    scrollable={tui.ScrollVertical}
    scrollOffset={0, d.scrollY.Get()}
>
    <span class="text-gradient-cyan-magenta font-bold">Recent Events</span>
    for _, event := range d.events.Get() {
        <span class="text-green">{event}</span>
    }
    if len(d.events.Get()) == 0 {
        <span class="font-dim">Waiting for events...</span>
    }
</div>
```

The `scrollable={tui.ScrollVertical}` attribute enables vertical scrolling on the container, and `scrollOffset={0, d.scrollY.Get()}` binds the scroll position to our state. The `ref` connects the element to `eventsRef` so the scroll helper can query `MaxScroll()`. Both panels use `flexGrow={1.0}` to share the row width equally.

### The Event Producer

Now update `main.go` to create the channel, start a producer goroutine, pass the channel to the component, and enable mouse support for scroll wheel:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    eventCh := make(chan string, 100)

    app, err := tui.NewApp(
        tui.WithRootComponent(Dashboard(eventCh)),
        tui.WithMouse(),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    go produceEvents(eventCh, app.StopCh())

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "App error: %v\n", err)
        os.Exit(1)
    }
}
```

`tui.WithMouse()` enables mouse event reporting so the scroll wheel works in the event feed.

The `produceEvents` function lives in `dashboard.gsx` alongside the component. It sends random events at 2-5 second intervals and respects the stop channel so it shuts down cleanly:

```go
func produceEvents(ch chan<- string, stopCh <-chan struct{}) {
    defer close(ch)
    events := []string{
        "Deploy completed",
        "Health check passed",
        "New connection from 10.0.0.5",
        "Cache invalidated",
        "Backup complete",
        "Certificate renewed",
        "Config reloaded",
        "Scale up: 3 replicas",
        "Alert cleared: cpu",
        "Metrics exported",
    }
    for {
        delay := time.Duration(2000+rand.Intn(3000)) * time.Millisecond
        select {
        case <-stopCh:
            return
        case <-time.After(delay):
        }
        event := events[rand.Intn(len(events))]
        select {
        case <-stopCh:
            return
        case ch <- event:
        }
    }
}
```

Two things to note about the goroutine pattern:

- `app.StopCh()` returns a channel that closes when the app exits. Using it in `select` ensures the goroutine doesn't leak.
- The buffered channel (`make(chan string, 100)`) prevents the producer from blocking if the UI falls behind.

## Polish

The dashboard already has gradient titles (`text-gradient-cyan-magenta`) and colored borders. A few things worth calling out:

The `metricColor` function returns different class strings depending on the value: green below 60%, yellow below 80%, red and bold above. You can use this pattern anywhere: write a function that returns a class string, and pass it to `class={}`.

Each panel gets its own `border-rounded` frame for visual separation, and `gap-1` between sections plus `p-1` padding inside panels keeps everything readable.

## Full Code

Here's the complete `dashboard.gsx`:

```gsx
package main

import (
    "fmt"
    "math/rand"
    "time"
    tui "github.com/grindlemire/go-tui"
)

type dashboardApp struct {
    cpu       *tui.State[int]
    mem       *tui.State[int]
    disk      *tui.State[int]
    netIn     *tui.State[int]
    netOut    *tui.State[int]
    sparkIn   *tui.State[[]int]
    sparkOut  *tui.State[[]int]
    events    *tui.State[[]string]
    eventCh   <-chan string
    scrollY   *tui.State[int]
    eventsRef *tui.Ref
}

func Dashboard(eventCh <-chan string) *dashboardApp {
    return &dashboardApp{
        cpu:       tui.NewState(45),
        mem:       tui.NewState(62),
        disk:      tui.NewState(38),
        netIn:     tui.NewState(142),
        netOut:    tui.NewState(89),
        sparkIn:   tui.NewState([]int{3, 5, 4, 6, 7, 5, 4, 3, 5, 6, 7, 8, 6, 5, 4, 3, 5, 6, 7, 5}),
        sparkOut:  tui.NewState([]int{2, 3, 4, 3, 5, 4, 3, 2, 3, 4, 5, 6, 4, 3, 2, 3, 4, 5, 4, 3}),
        events:    tui.NewState([]string{}),
        eventCh:   eventCh,
        scrollY:   tui.NewState(0),
        eventsRef: tui.NewRef(),
    }
}

func (d *dashboardApp) scrollBy(delta int) {
    el := d.eventsRef.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := d.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    } else if newY > maxY {
        newY = maxY
    }
    d.scrollY.Set(newY)
}

func (d *dashboardApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('j', func(ke tui.KeyEvent) { d.scrollBy(1) }),
        tui.OnRune('k', func(ke tui.KeyEvent) { d.scrollBy(-1) }),
        tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { d.scrollBy(1) }),
        tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { d.scrollBy(-1) }),
    }
}

func (d *dashboardApp) HandleMouse(me tui.MouseEvent) bool {
    switch me.Button {
    case tui.MouseWheelUp:
        d.scrollBy(-1)
        return true
    case tui.MouseWheelDown:
        d.scrollBy(1)
        return true
    }
    return false
}

func (d *dashboardApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(500*time.Millisecond, d.updateMetrics),
        tui.Watch(d.eventCh, d.addEvent),
    }
}

func (d *dashboardApp) updateMetrics() {
    d.cpu.Set(clampVal(d.cpu.Get()+rand.Intn(11)-5, 5, 95))
    d.mem.Set(clampVal(d.mem.Get()+rand.Intn(7)-3, 20, 90))
    d.disk.Set(clampVal(d.disk.Get()+rand.Intn(3)-1, 20, 80))
    d.netIn.Set(clampVal(d.netIn.Get()+rand.Intn(41)-20, 50, 300))
    d.netOut.Set(clampVal(d.netOut.Get()+rand.Intn(31)-15, 30, 200))

    inData := d.sparkIn.Get()
    inData = append(inData[1:], d.netIn.Get()/30)
    d.sparkIn.Set(inData)

    outData := d.sparkOut.Get()
    outData = append(outData[1:], d.netOut.Get()/30)
    d.sparkOut.Set(outData)
}

func (d *dashboardApp) addEvent(event string) {
    current := d.events.Get()
    ts := time.Now().Format("15:04:05")
    entry := fmt.Sprintf("%s  %s", ts, event)
    current = append(current, entry)
    if len(current) > 50 {
        current = current[len(current)-50:]
    }
    d.events.Set(current)

    // Auto-scroll to bottom
    el := d.eventsRef.El()
    if el != nil {
        _, maxY := el.MaxScroll()
        d.scrollY.Set(maxY + 1)
    }
}

func clampVal(v, min, max int) int {
    if v < min {
        return min
    }
    if v > max {
        return max
    }
    return v
}

func metricBar(value, max int) string {
    width := 20
    filled := value * width / max
    bar := ""
    for i := 0; i < width; i++ {
        if i < filled {
            bar += "█"
        } else {
            bar += "░"
        }
    }
    return bar
}

func metricColor(value int) string {
    if value >= 80 {
        return "text-red font-bold"
    }
    if value >= 60 {
        return "text-yellow"
    }
    return "text-green"
}

func sparkline(data []int) string {
    blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
    maxVal := 1
    for _, v := range data {
        if v > maxVal {
            maxVal = v
        }
    }
    s := ""
    for _, v := range data {
        idx := v * 7 / maxVal
        if idx > 7 {
            idx = 7
        }
        s += string(blocks[idx])
    }
    return s
}

func produceEvents(ch chan<- string, stopCh <-chan struct{}) {
    defer close(ch)
    events := []string{
        "Deploy completed",
        "Health check passed",
        "New connection from 10.0.0.5",
        "Cache invalidated",
        "Backup complete",
        "Certificate renewed",
        "Config reloaded",
        "Scale up: 3 replicas",
        "Alert cleared: cpu",
        "Metrics exported",
    }
    for {
        delay := time.Duration(2000+rand.Intn(3000)) * time.Millisecond
        select {
        case <-stopCh:
            return
        case <-time.After(delay):
        }
        event := events[rand.Intn(len(events))]
        select {
        case <-stopCh:
            return
        case ch <- event:
        }
    }
}

templ (d *dashboardApp) Render() {
    <div class="flex-col p-1 gap-1 h-full border-rounded border-cyan">
        <div class="flex justify-center shrink-0">
            <span class="text-gradient-cyan-magenta font-bold">Dashboard</span>
        </div>

        // Metric gauges
        <div class="flex gap-1 shrink-0">
            <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
                <span class="text-gradient-cyan-magenta font-bold">CPU</span>
                <span class={metricColor(d.cpu.Get())}>{metricBar(d.cpu.Get(), 100)}</span>
                <span class={metricColor(d.cpu.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.cpu.Get())}</span>
            </div>
            <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
                <span class="text-gradient-cyan-magenta font-bold">Memory</span>
                <span class={metricColor(d.mem.Get())}>{metricBar(d.mem.Get(), 100)}</span>
                <span class={metricColor(d.mem.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.mem.Get())}</span>
            </div>
            <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
                <span class="text-gradient-cyan-magenta font-bold">Disk</span>
                <span class={metricColor(d.disk.Get())}>{metricBar(d.disk.Get(), 100)}</span>
                <span class={metricColor(d.disk.Get()) + " font-bold"}>{fmt.Sprintf("%d%%", d.disk.Get())}</span>
            </div>
        </div>

        // Network Traffic + Recent Events
        <div class="flex gap-1 flex-grow">
            <div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
                <span class="text-gradient-cyan-magenta font-bold">Network Traffic</span>
                <div class="flex gap-1">
                    <span class="font-dim">In: </span>
                    <span class="text-cyan">{sparkline(d.sparkIn.Get())}</span>
                </div>
                <div class="flex gap-1">
                    <span class="font-dim">Out:</span>
                    <span class="text-magenta">{sparkline(d.sparkOut.Get())}</span>
                </div>
                <div class="flex gap-2">
                    <span class="text-cyan font-bold">{fmt.Sprintf("In: %d MB/s", d.netIn.Get())}</span>
                    <span class="text-magenta font-bold">{fmt.Sprintf("Out: %d MB/s", d.netOut.Get())}</span>
                </div>
            </div>

            <div
                ref={d.eventsRef}
                class="flex-col border-rounded p-1 gap-1"
                flexGrow={1.0}
                scrollable={tui.ScrollVertical}
                scrollOffset={0, d.scrollY.Get()}
            >
                <span class="text-gradient-cyan-magenta font-bold">Recent Events</span>
                for _, event := range d.events.Get() {
                    <span class="text-green">{event}</span>
                }
                if len(d.events.Get()) == 0 {
                    <span class="font-dim">Waiting for events...</span>
                }
            </div>
        </div>

        <div class="flex justify-center shrink-0">
            <span class="font-dim">j/k scroll events | q to quit</span>
        </div>
    </div>
}
```

And the complete `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    eventCh := make(chan string, 100)

    app, err := tui.NewApp(
        tui.WithRootComponent(Dashboard(eventCh)),
        tui.WithMouse(),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    go produceEvents(eventCh, app.StopCh())

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "App error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

The finished dashboard with live-updating metrics, sparklines, and an event log:

![Building a Dashboard screenshot](/guides/18.png)

## Next Steps

That covers it. Here are some ways you could extend the dashboard:

- Add clickable panels that expand when clicked ([Events Guide](events))
- Add focus navigation between panels with Tab/Shift-Tab ([Focus Guide](focus))
- Split the dashboard into multiple `.gsx` files with separate components for each panel ([Multi-Component Guide](multi-component))
- Add an alternate screen settings overlay ([Inline Mode Guide](inline-mode))
