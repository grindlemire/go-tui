# Event Handling Specification

**Status:** Draft\
**Version:** 2.1\
**Last Updated:** 2026-01-27

---

## 1. Overview

### Purpose

Provide a unified event handling system with a hybrid push/poll architecture. The framework manages the main event loop, dispatches events to handlers, and only renders when changes occur (isDirty pattern).

### Goals

- **Framework-owned main loop**: `app.Run()` handles all event dispatch and rendering
- **Push-based external events**: Channels and timers push events immediately when they occur
- **Push-based input events**: Keyboard/mouse events are dispatched immediately
- **Automatic dirty tracking**: Mutating operations (state.Set, element.ScrollBy, etc.) mark dirty automatically
- **No bool returns**: Handlers don't need to return whether UI changed—the framework tracks this
- **Explicit watcher functions**: `tui.Watch(ch, handler)` and `tui.OnTimer(duration, handler)` - no special syntax
- **No Context parameter**: Components don't need to accept or pass Context—framework handles it internally
- **SetRoot takes view directly**: `app.SetRoot(view)` extracts Root and starts watchers automatically
- **Regular Go functions**: Event handlers are normal Go functions, no special DSL syntax
- **Batched rendering**: Multiple events are processed before a single render
- **Cleanup handles**: Watchers stopped automatically when app stops (via stopCh)
- **App-level key bindings**: `SetGlobalKeyHandler` for quit and other app-level keys—components don't manage app lifecycle
- **Graceful SIGINT handling**: Ctrl+C triggers automatic shutdown

### Non-Goals

- Poll-based `onUpdate` every frame (may add later for animations)
- Two-way data binding
- Event bubbling (may add later)

---

## 2. Architecture

### Hybrid Push Model

```
┌─────────────────────────────────────────────────────────────────┐
│                    Event Sources (Push)                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Terminal Input ──┐                                             │
│    (goroutine)    │                                             │
│                   │      ┌──────────────┐                       │
│  Channel Data  ───┼─────►│ Event Queue  │                       │
│    (goroutine)    │      │ chan func()  │                       │
│                   │      └──────┬───────┘                       │
│  Timer Ticks   ───┘             │                               │
│    (goroutine)                  │                               │
│                                 ▼                               │
│                    ┌────────────────────────┐                   │
│                    │     Main Loop          │                   │
│                    │  1. Drain queue        │                   │
│                    │  2. Check dirty flag   │  ◄── auto-set by  │
│                    │  3. Render if dirty    │      mutations    │
│                    └────────────────────────┘                   │
└─────────────────────────────────────────────────────────────────┘
```

### Component Overview

| Component | Change |
|-----------|--------|
| `pkg/tui/watcher.go` | NEW: Watcher interface, Watch(), OnTimer() functions |
| `pkg/tui/dirty.go` | NEW: Dirty flag management and automatic tracking |
| `pkg/tui/app.go` | Add Run(), Stop(), SetRoot(), SetGlobalKeyHandler(), SIGINT handling |
| `pkg/tui/element/element.go` | Add onKeyPress, onClick (no bool return); mutations mark dirty |
| `pkg/tui/element/options.go` | Add WithOnKeyPress, WithOnClick (no bool return) |
| `pkg/tuigen/generator.go` | Collect watchers from onChannel/onTimer attributes |

---

## 3. Core Entities

### 3.1 Dirty Tracking

Automatic dirty tracking eliminates the need for handlers to return bool. Mutating operations mark dirty internally.

```go
// pkg/tui/dirty.go

package tui

import "sync/atomic"

// Global dirty flag - set by any mutating operation
var dirty atomic.Bool

// MarkDirty signals that the UI needs re-rendering.
// Called automatically by State.Set(), Element.ScrollBy(), etc.
// Can also be called manually for custom mutations.
func MarkDirty() {
    dirty.Store(true)
}

// checkAndClearDirty returns true if dirty and clears the flag.
// Called by the main loop after processing events.
func checkAndClearDirty() bool {
    return dirty.Swap(false)
}
```

### 3.2 Watcher Types

Watchers are created during component construction and started when `SetRoot` is called. This allows components to declare watchers without needing access to the App.

```go
// pkg/tui/watcher.go

package tui

import "time"

// Watcher represents a deferred event source that starts when the app runs.
// Watchers are collected during component construction and started by SetRoot.
type Watcher interface {
    // Start begins the watcher goroutine. Called by App.SetRoot().
    Start(app *App)
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
//   onChannel={tui.Watch(dataCh, handleData(lines))}
//
// Handlers don't return bool - mutations automatically mark dirty.
func Watch[T any](ch <-chan T, handler func(T)) Watcher {
    return &channelWatcher[T]{ch: ch, handler: handler}
}

func (w *channelWatcher[T]) Start(app *App) {
    go func() {
        for {
            select {
            case <-app.stopCh:
                return
            case val, ok := <-w.ch:
                if !ok {
                    return // Channel closed
                }
                app.eventQueue <- func() {
                    w.handler(val)
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
//   onTimer={tui.OnTimer(time.Second, tick(elapsed))}
//
// Handlers don't return bool - mutations automatically mark dirty.
func OnTimer(interval time.Duration, handler func()) Watcher {
    return &timerWatcher{interval: interval, handler: handler}
}

func (w *timerWatcher) Start(app *App) {
    go func() {
        ticker := time.NewTicker(w.interval)
        defer ticker.Stop()

        for {
            select {
            case <-app.stopCh:
                return
            case <-ticker.C:
                app.eventQueue <- w.handler
            }
        }
    }()
}
```

### 3.3 App Changes

```go
// pkg/tui/app.go

type App struct {
    // ... existing fields ...

    root             *element.Element
    eventQueue       chan func()
    stopCh           chan struct{}
    stopped          bool
    globalKeyHandler func(KeyEvent) bool  // Returns true if event consumed
}

func NewApp() (*App, error) {
    // ... existing setup ...

    app := &App{
        // ... existing fields ...
        eventQueue: make(chan func(), 256),
        stopCh:     make(chan struct{}),
        stopped:    false,
    }

    return app, nil
}

// SetGlobalKeyHandler sets a handler that runs before dispatching to focused element.
// If the handler returns true, the event is consumed and not dispatched further.
// Use this for app-level key bindings like quit.
func (a *App) SetGlobalKeyHandler(fn func(KeyEvent) bool) {
    a.globalKeyHandler = fn
}

// Viewable is implemented by generated view structs.
// Allows SetRoot to extract the root element and start watchers.
type Viewable interface {
    GetRoot() *element.Element
    GetWatchers() []Watcher
}

// SetRoot sets the root view. Accepts:
// - A view struct implementing Viewable (extracts Root, starts watchers)
// - A raw *element.Element
func (a *App) SetRoot(v any) {
    switch view := v.(type) {
    case Viewable:
        a.root = view.GetRoot()
        // Start all watchers collected during component construction
        for _, w := range view.GetWatchers() {
            w.Start(a)
        }
    case *element.Element:
        a.root = view
    }
}

// Run starts the main event loop. Blocks until Stop() is called or SIGINT received.
// Rendering occurs only when the dirty flag is set (by mutations).
func (a *App) Run() error {
    // Handle Ctrl+C gracefully
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt)
    go func() {
        <-sigCh
        a.Stop()
    }()

    // Start input reader in background
    go a.readInputEvents()

    for !a.stopped {
        // Block until at least one event arrives
        select {
        case handler := <-a.eventQueue:
            handler()
        case <-a.stopCh:
            return nil
        }

        // Drain any additional queued events (batch processing)
    drain:
        for {
            select {
            case handler := <-a.eventQueue:
                handler()
            default:
                break drain
            }
        }

        // Only render if something changed (dirty flag set by mutations)
        if checkAndClearDirty() {
            a.Render()
        }
    }

    return nil
}

// Stop signals the Run loop to exit gracefully and stops all watchers.
// Watchers receive the stop signal via stopCh and exit their goroutines.
func (a *App) Stop() {
    if a.stopped {
        return // Already stopped
    }
    a.stopped = true

    // Signal all watcher goroutines to stop
    close(a.stopCh)
}

// QueueUpdate enqueues a function to run on the main loop.
// Safe to call from any goroutine. Use this for background thread safety.
func (a *App) QueueUpdate(fn func()) {
    select {
    case a.eventQueue <- fn:
    default:
        // Queue full - this shouldn't happen with reasonable buffer size
        // Could log a warning here
    }
}

// readInputEvents reads terminal input in a goroutine and queues events.
func (a *App) readInputEvents() {
    for {
        select {
        case <-a.stopCh:
            return
        default:
        }

        event, ok := a.reader.PollEvent(50 * time.Millisecond)
        if !ok {
            continue
        }

        a.eventQueue <- func() {
            // Global key handler runs first (for app-level bindings like quit)
            if keyEvent, isKey := event.(KeyEvent); isKey {
                if a.globalKeyHandler != nil && a.globalKeyHandler(keyEvent) {
                    return // Event consumed by global handler
                }
            }
            a.Dispatch(event)
        }
    }
}
```

### 3.4 Element Handler Changes

Handlers no longer return bool. Mutating element methods mark dirty automatically.

```go
// pkg/tui/element/element.go

type Element struct {
    // ... existing fields ...

    onKeyPress func(tui.KeyEvent)
    onClick    func()
}

// SetOnKeyPress sets a handler for key press events.
// No return value needed - mutations mark dirty automatically.
func (e *Element) SetOnKeyPress(fn func(tui.KeyEvent)) {
    e.onKeyPress = fn
}

// SetOnClick sets a handler for click events.
// No return value needed - mutations mark dirty automatically.
func (e *Element) SetOnClick(fn func()) {
    e.onClick = fn
}

// ScrollBy scrolls the element by the given amount.
// Automatically marks dirty.
func (e *Element) ScrollBy(dx, dy int) {
    e.scrollX += dx
    e.scrollY += dy
    tui.MarkDirty()
}

// SetText updates the element's text content.
// Automatically marks dirty.
func (e *Element) SetText(text string) {
    e.text = text
    tui.MarkDirty()
}

// AddChild adds a child element.
// Automatically marks dirty.
func (e *Element) AddChild(children ...*Element) {
    e.children = append(e.children, children...)
    tui.MarkDirty()
}

// RemoveAllChildren removes all children.
// Automatically marks dirty.
func (e *Element) RemoveAllChildren() {
    e.children = nil
    tui.MarkDirty()
}
```

### 3.5 Element Options

```go
// pkg/tui/element/options.go

// WithOnKeyPress sets the key press handler.
// No return value needed - mutations mark dirty automatically.
func WithOnKeyPress(fn func(tui.KeyEvent)) Option {
    return func(e *Element) {
        e.onKeyPress = fn
    }
}

// WithOnClick sets the click handler.
// No return value needed - mutations mark dirty automatically.
func WithOnClick(fn func()) Option {
    return func(e *Element) {
        e.onClick = fn
    }
}
```

---

## 4. DSL Syntax

### 4.1 Channel Watcher

Use `tui.Watch(ch, handler)` to create a channel watcher. Watchers are collected during component construction and started when `SetRoot` is called.

```tui
@component StreamBox(dataCh chan string) {
    lines := tui.NewState([]string{})

    <div onChannel={tui.Watch(dataCh, handleData(lines))}>
        @for _, line := range lines.Get() {
            <span>{line}</span>
        }
    </div>
}

// Regular Go function - no bool return needed
func handleData(lines *tui.State[[]string]) func(string) {
    return func(line string) {
        lines.Set(append(lines.Get(), line))
        // No return - State.Set() marks dirty automatically
    }
}
```

### 4.2 Timer Watcher

Use `tui.OnTimer(duration, handler)` to create a timer watcher.

```tui
@component Clock() {
    elapsed := tui.NewState(0)

    <div onTimer={tui.OnTimer(time.Second, tick(elapsed))}>
        <span>{fmt.Sprintf("Elapsed: %ds", elapsed.Get())}</span>
    </div>
}

// No bool return needed
func tick(elapsed *tui.State[int]) func() {
    return func() {
        elapsed.Set(elapsed.Get() + 1)
        // No return - State.Set() marks dirty automatically
    }
}
```

### 4.3 Key Press Handler

```tui
@component Counter() {
    count := tui.NewState(0)

    <div onKeyPress={handleKeys(count)} focusable={true}>
        <span>{fmt.Sprintf("Count: %d", count.Get())}</span>
    </div>
}

// No bool return needed - mutations mark dirty automatically
// Note: App-level concerns like quit are handled via SetGlobalKeyHandler
func handleKeys(count *tui.State[int]) func(tui.KeyEvent) {
    return func(e tui.KeyEvent) {
        switch e.Rune {
        case '+':
            count.Set(count.Get() + 1)
        case '-':
            count.Set(count.Get() - 1)
        }
        // No return needed
    }
}
```

### 4.4 Click Handler

```tui
@component Button(label string, onClick func()) {
    <button onClick={onClick} focusable={true}>
        {label}
    </button>
}
```

### 4.5 Component Composition

Components don't need Context - watchers are collected automatically and started by SetRoot.

```tui
// No context threading needed
@component App(dataCh chan string) {
    <div class="flex-col">
        @Header()           // Simple component
        @StreamBox(dataCh)  // Has channel watcher
        @Clock()            // Has timer watcher
    </div>
}

// Simple component
@component Header() {
    <div class="border-single">
        <span>{"My App"}</span>
    </div>
}
```

**No Context needed** - the framework handles watcher collection and startup internally.

---

## 5. Generated Code

### 5.1 Channel Watcher Registration

**Input:**

```tui
@component StreamBox(dataCh chan string) {
    lines := tui.NewState([]string{})

    <div #Content onChannel={tui.Watch(dataCh, handleData(lines))}>
    </div>
}
```

**Output:**

```go
type StreamBoxView struct {
    Root     *element.Element
    Content  *element.Element
    watchers []tui.Watcher  // Collected watchers for deferred startup
}

// GetRoot implements tui.Viewable
func (v StreamBoxView) GetRoot() *element.Element { return v.Root }

// GetWatchers implements tui.Viewable
func (v StreamBoxView) GetWatchers() []tui.Watcher { return v.watchers }

func StreamBox(dataCh chan string) StreamBoxView {
    var view StreamBoxView
    var watchers []tui.Watcher

    lines := tui.NewState([]string{})

    Content := element.New()

    Root := element.New()
    Root.AddChild(Content)

    // Generated: Collect channel watcher (started by SetRoot)
    watchers = append(watchers, tui.Watch(dataCh, handleData(lines)))

    view = StreamBoxView{Root: Root, Content: Content, watchers: watchers}
    return view
}
```

### 5.2 Key Handler Registration

**Input:**

```tui
@component Counter() {
    count := tui.NewState(0)

    <div onKeyPress={handleKeys(count)} focusable={true}>
        <span>{count.Get()}</span>
    </div>
}
```

**Output:**

```go
type CounterView struct {
    Root     *element.Element
    watchers []tui.Watcher
}

func (v CounterView) GetRoot() *element.Element   { return v.Root }
func (v CounterView) GetWatchers() []tui.Watcher { return v.watchers }

func Counter() CounterView {
    var view CounterView

    count := tui.NewState(0)

    span := element.New(element.WithText(fmt.Sprintf("%d", count.Get())))
    count.Bind(func(v int) {
        span.SetText(fmt.Sprintf("%d", v))
    })

    Root := element.New(
        element.WithOnKeyPress(handleKeys(count)),  // no bool return
        element.WithFocusable(true),
    )
    Root.AddChild(span)

    view = CounterView{Root: Root, watchers: nil}
    return view
}
```

### 5.3 Timer Registration

**Input:**

```tui
@component Clock() {
    elapsed := tui.NewState(0)

    <div onTimer={tui.OnTimer(time.Second, tick(elapsed))}>
        <span>{fmt.Sprintf("%ds", elapsed.Get())}</span>
    </div>
}
```

**Output:**

```go
type ClockView struct {
    Root     *element.Element
    watchers []tui.Watcher
}

func (v ClockView) GetRoot() *element.Element   { return v.Root }
func (v ClockView) GetWatchers() []tui.Watcher { return v.watchers }

func Clock() ClockView {
    var view ClockView
    var watchers []tui.Watcher

    elapsed := tui.NewState(0)

    span := element.New(element.WithText(fmt.Sprintf("%ds", elapsed.Get())))
    elapsed.Bind(func(v int) {
        span.SetText(fmt.Sprintf("%ds", v))
    })

    Root := element.New()
    Root.AddChild(span)

    // Generated: Collect timer watcher (started by SetRoot)
    watchers = append(watchers, tui.OnTimer(time.Second, tick(elapsed)))

    view = ClockView{Root: Root, watchers: watchers}
    return view
}
```

### 5.4 Watcher Aggregation from Nested Components

When a component contains children with watchers, all watchers are aggregated upward.

**Input:**

```tui
@component App(dataCh chan string) {
    <div>
        @StreamBox(dataCh)
        @Clock()
    </div>
}
```

**Output:**

```go
type AppView struct {
    Root     *element.Element
    watchers []tui.Watcher
}

func (v AppView) GetRoot() *element.Element   { return v.Root }
func (v AppView) GetWatchers() []tui.Watcher { return v.watchers }

func App(dataCh chan string) AppView {
    var view AppView
    var watchers []tui.Watcher

    // Child components may have watchers
    streamBox := StreamBox(dataCh)
    watchers = append(watchers, streamBox.GetWatchers()...)

    clock := Clock()
    watchers = append(watchers, clock.GetWatchers()...)

    Root := element.New()
    Root.AddChild(streamBox.Root)
    Root.AddChild(clock.Root)

    view = AppView{Root: Root, watchers: watchers}
    return view
}
```

**Usage in main.go:**

```go
func main() {
    app, _ := tui.NewApp()
    defer app.Close()

    dataCh := make(chan string, 100)
    go produce(dataCh)

    // SetRoot takes the view directly - extracts Root and starts watchers
    app.SetRoot(App(dataCh))

    app.Run()
}
```

---

## 6. Handler Types Summary

All handlers have unified signatures without bool returns. Mutations automatically mark dirty.

| Attribute | Handler Signature | When Called | DSL Syntax |
|-----------|-------------------|-------------|------------|
| `onClick` | `func()` | On mouse click | `onClick={handler}` |
| `onKeyPress` | `func(KeyEvent)` | On key press (when focused) | `onKeyPress={handler}` |
| `onInput` | `func(string)` | On text input change | `onInput={handler}` |
| `onChannel` | `func(T)` | When channel receives data | `onChannel={tui.Watch(ch, handler)}` |
| `onTimer` | `func()` | At interval | `onTimer={tui.OnTimer(duration, handler)}` |

**Key design decisions:**
- No `bool` return values - mutations mark dirty automatically
- No Context parameter - watchers collected during construction, started by SetRoot
- Explicit `tui.Watch()` and `tui.OnTimer()` function calls - no special parsing syntax
- Unified function signatures across all handler types

---

## 7. Event Flow

### 7.1 Input Event Flow

```
Terminal → EventReader.PollEvent()
    → readInputEvents goroutine
    → eventQueue <- func() { ... }
    → Main loop drains queue
    → Global key handler runs first (if set)
        → If returns true: event consumed, stop here
        → If returns false: continue to dispatch
    → FocusManager.Dispatch()
    → Element.HandleEvent()
    → element.onKeyPress(event)
    → Handler calls state.Set() or element.ScrollBy(), etc.
    → Mutation automatically calls MarkDirty()
    → After draining queue, check dirty flag
    → Render() if dirty, clear flag
```

### 7.1.1 SIGINT Flow

```
User presses Ctrl+C
    → OS sends SIGINT
    → signal.Notify receives on sigCh
    → Goroutine calls app.Stop()
    → stopCh is closed
    → All watcher goroutines exit
    → readInputEvents goroutine exits
    → Run() returns
```

### 7.2 Channel Event Flow

```
External goroutine sends to channel
    → OnChannel goroutine receives
    → eventQueue <- func() { handler(value) }
    → Main loop drains queue
    → handler(value) executes
    → Handler mutations call MarkDirty()
    → After draining queue, check dirty flag
    → Render() if dirty, clear flag
```

### 7.3 Timer Event Flow

```
time.Ticker fires
    → OnTimer goroutine receives
    → eventQueue <- handler
    → Main loop drains queue
    → handler() executes
    → Handler mutations call MarkDirty()
    → After draining queue, check dirty flag
    → Render() if dirty, clear flag
```

### 7.4 Dirty Flag Behavior

The dirty flag is:
- Set atomically by any mutating operation (`State.Set()`, `Element.SetText()`, `Element.ScrollBy()`, etc.)
- Checked and cleared atomically after event batch is drained
- If multiple mutations occur during one event batch, only one render occurs

---

## 8. User Experience

### 8.1 Complete Example

```tui
// streaming.tui
package main

import (
    "fmt"
    "time"
)

@component StreamApp(dataCh chan string) {
    lines := tui.NewState([]string{})
    count := tui.NewState(0)
    elapsed := tui.NewState(0)

    <div class="flex-col">
        <div class="border-single" height={3}>
            <span class="font-bold">
                {fmt.Sprintf("Stream (%d lines, %ds)", count.Get(), elapsed.Get())}
            </span>
        </div>

        <div #Content
             class="flex-col border-cyan"
             onChannel={tui.Watch(dataCh, handleData(lines, count))}
             onKeyPress={handleKeys(view)}
             onTimer={tui.OnTimer(time.Second, tick(elapsed))}
             scrollable={element.ScrollVertical}
             focusable={true}
             flexGrow={1}>
            @for _, line := range lines.Get() {
                <span>{line}</span>
            }
        </div>

        <div class="border-single" height={1}>
            <span class="text-dim">{"j/k scroll | q quit (app-level)"}</span>
        </div>
    </div>
}

// No bool return - State.Set() marks dirty automatically
func handleData(lines *tui.State[[]string], count *tui.State[int]) func(string) {
    return func(line string) {
        lines.Set(append(lines.Get(), line))
        count.Set(count.Get() + 1)
    }
}

// No bool return - ScrollBy() marks dirty automatically
// Note: Quit is handled at app level via SetGlobalKeyHandler, not here
func handleKeys(v StreamAppView) func(tui.KeyEvent) {
    return func(e tui.KeyEvent) {
        switch e.Rune {
        case 'j':
            v.Content.ScrollBy(0, 1)
        case 'k':
            v.Content.ScrollBy(0, -1)
        }
    }
}

// No bool return - State.Set() marks dirty automatically
func tick(elapsed *tui.State[int]) func() {
    return func() {
        elapsed.Set(elapsed.Get() + 1)
    }
}
```

```go
// main.go
package main

import (
    "fmt"
    "time"

    "github.com/grindlemire/go-tui/pkg/tui"
)

func main() {
    app, _ := tui.NewApp()
    defer app.Close()

    // App-level key bindings (quit on 'q' or Ctrl+C)
    // Ctrl+C is also handled automatically via SIGINT
    app.SetGlobalKeyHandler(func(e tui.KeyEvent) bool {
        if e.Rune == 'q' {
            app.Stop()
            return true // Event consumed
        }
        return false // Pass to focused element
    })

    dataCh := make(chan string, 100)
    go produce(dataCh)

    // SetRoot takes view directly - extracts Root and starts watchers
    app.SetRoot(StreamApp(dataCh))

    app.Run()
}

func produce(ch chan<- string) {
    for i := 0; i < 100; i++ {
        ch <- fmt.Sprintf("[%s] Line %d", time.Now().Format("15:04:05"), i)
        time.Sleep(200 * time.Millisecond)
    }
    close(ch)
}
```

---

## 9. Rules and Constraints

1. **Handlers don't return bool** - mutations mark dirty automatically via `tui.MarkDirty()`
2. **No Context parameter** - components don't need Context; watchers are collected automatically
3. **Channel watchers are push-based** - handler called immediately when data arrives
4. **Timer watchers are push-based** - handler called at interval
5. **Input events are push-based** - dispatched immediately when received
6. **Rendering is batched** - multiple events processed before single render
7. **Explicit watcher functions** - `tui.Watch(ch, handler)` and `tui.OnTimer(duration, handler)` return Watcher interface
8. **Deferred watcher startup** - watchers are collected during construction, started by SetRoot
9. **SetRoot takes view directly** - implements Viewable interface to extract Root and watchers
10. **Handler functions are regular Go** - no special DSL syntax for channel/timer watchers
11. **App.Stop() stops all watchers** - closes stopCh, automatic cleanup when app exits
12. **Global key handler for app-level bindings** - `SetGlobalKeyHandler` runs before dispatch to components
13. **SIGINT handled automatically** - Ctrl+C triggers graceful shutdown via `app.Stop()`
14. **Component handlers don't manage app lifecycle** - quit/stop is handled at app level, not in component handlers

---

## 10. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Medium\
**Recommended Phases:** 4

### Phase Breakdown

1. **Phase 1: Dirty Tracking & Watcher Types** (Medium)
   - Create `pkg/tui/dirty.go` with atomic dirty flag and `MarkDirty()`
   - Create `pkg/tui/watcher.go` with Watcher interface
   - Implement `Watch[T]()` for channel watchers
   - Implement `OnTimer()` for timer watchers
   - Unit tests for dirty tracking

2. **Phase 2: App.Run() & SetRoot** (Medium)
   - Add eventQueue, stopped, stopCh, globalKeyHandler to App struct
   - Implement Viewable interface
   - Implement SetRoot() that extracts Root and starts watchers
   - Implement SetGlobalKeyHandler() for app-level key bindings
   - Implement Run() with dirty flag checking, SIGINT handling
   - Implement Stop() that closes stopCh for automatic watcher cleanup
   - Implement QueueUpdate() for background thread safety
   - Add readInputEvents() goroutine with global key handler support
   - Unit tests for event queue, global handler, and cleanup

3. **Phase 3: Element Handler Changes** (Small)
   - Add onKeyPress, onClick fields (no bool return)
   - Add WithOnKeyPress, WithOnClick options
   - Update mutating methods (ScrollBy, SetText, AddChild) to call MarkDirty()
   - Update HandleEvent to use new handler signatures
   - Unit tests

4. **Phase 4: Generator Updates** (Medium)
   - Generate Viewable interface (GetRoot, GetWatchers) for view structs
   - Collect watchers from `onChannel={tui.Watch(...)}` attributes into view struct
   - Collect watchers from `onTimer={tui.OnTimer(...)}` attributes into view struct
   - Aggregate watchers from nested component calls
   - Update examples to use new pattern

---

## 11. Success Criteria

1. `app.Run()` blocks and processes events until `app.Stop()` is called or SIGINT received
2. Channel data triggers handler immediately (no polling delay)
3. Timer fires at specified interval
4. Input events are dispatched to focused element
5. Mutations (State.Set, Element.ScrollBy, etc.) automatically mark dirty
6. Handlers don't need to return bool - dirty tracking is automatic
7. Multiple events in queue are batched before single render
8. `tui.Watch(ch, handler)` returns Watcher that is collected during construction
9. `tui.OnTimer(duration, handler)` returns Watcher that is collected during construction
10. `app.SetRoot(view)` extracts Root and starts all collected watchers
11. View structs implement `Viewable` interface (GetRoot, GetWatchers)
12. `app.Stop()` closes stopCh and automatically stops all running watchers
13. `app.SetGlobalKeyHandler()` allows app-level key bindings (e.g., quit on 'q')
14. Global key handler runs before dispatch; returns true to consume event
15. Ctrl+C (SIGINT) triggers graceful shutdown automatically
16. Example streaming app works with channel + timer + key handlers (no bool returns, no Context)
17. Zero CPU usage when idle (no events, no rendering)

---

## 12. Relationship to Other Features

### State Management

Event handlers update state, which triggers bindings and marks dirty automatically:

```go
func handleData(lines *tui.State[[]string]) func(string) {
    return func(line string) {
        lines.Set(append(lines.Get(), line))  // Triggers bindings AND marks dirty
        // No return needed - framework knows to render
    }
}
```

The state bindings update elements immediately when `Set()` is called. The `Set()` call also invokes `MarkDirty()`, so the framework knows to render after the event batch is processed.

### Named Element Refs

Refs are used in handlers to perform imperative operations:

```go
func handleKeys(v StreamAppView) func(tui.KeyEvent) {
    return func(e tui.KeyEvent) {
        switch e.Rune {
        case 'j':
            v.Content.ScrollBy(0, 1)  // Marks dirty automatically
        }
        // No return needed
    }
}
```

State handles reactive updates (text, counts). Refs handle imperative operations (scroll, focus). Both mark dirty automatically.

### Watcher Aggregation

Watchers from child components are automatically aggregated up to the root view struct:

```go
// App.tui contains StreamBox and Clock children
app.SetRoot(App(dataCh))

// Internally, App's watchers field contains all watchers from:
// - StreamBox's channel watcher (tui.Watch)
// - Clock's timer watcher (tui.OnTimer)
// SetRoot starts them all
```

This happens automatically in the generated code - no manual watcher management needed.
