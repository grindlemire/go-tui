# Hybrid Event Handling Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Improve the component event handling API by adding a `HandleClicks` helper for automatic ref-based hit testing and a `WatcherProvider` interface for component-level timers/channels.

**Architecture:** Keep existing `KeyMap()` and `HandleMouse()` interfaces (supports dynamic bindings), add helper function for mouse hit testing, add new `WatcherProvider` interface for timers/channels collected via tree walk.

**Tech Stack:** Go interfaces, existing watcher infrastructure.

---

## Target API

```go
type counter struct {
    count *tui.State[int]
    btn   *tui.Ref
}

// Dynamic key bindings (called every render) - UNCHANGED
func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('+', func(ke tui.KeyEvent) { c.increment() }),
    }
}

// Mouse with automatic hit testing helper - NEW HELPER
func (c *counter) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.btn, c.increment),
    )
}

// Timers and channels - NEW INTERFACE
func (c *counter) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.NewTickerWatcher(time.Second, c.tick),
    }
}

// Clean template - no onTimer attribute needed
templ (c *counter) Render() {
    <div>
        <button ref={c.btn}>{" + "}</button>
    </div>
}
```

---

## Task 1: Add Click Helper Type

**Files:**
- Create: `click.go`
- Create: `click_test.go`

**Step 1: Write failing test for Click type**

```go
// click_test.go
package tui

import "testing"

func TestClick_Type(t *testing.T) {
    ref := NewRef()
    called := false
    c := Click(ref, func() { called = true })

    if c.Ref != ref {
        t.Fatal("ref not set")
    }

    c.Fn()
    if !called {
        t.Fatal("fn not called")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestClick_Type -v`
Expected: FAIL with "undefined: Click"

**Step 3: Write minimal implementation**

```go
// click.go
package tui

// ClickBinding represents a ref-to-function binding for mouse clicks.
type ClickBinding struct {
    Ref *Ref
    Fn  func()
}

// Click creates a click binding for use with HandleClicks.
func Click(ref *Ref, fn func()) ClickBinding {
    return ClickBinding{Ref: ref, Fn: fn}
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestClick_Type -v`
Expected: PASS

**Step 5: Commit**

```bash
git add click.go click_test.go
gcommit -m "feat: add ClickBinding type for ref-based mouse handling"
```

---

## Task 2: Add HandleClicks Helper Function

**Files:**
- Modify: `click.go`
- Modify: `click_test.go`

**Step 1: Write failing test for HandleClicks**

```go
// click_test.go
func TestHandleClicks_Hit(t *testing.T) {
    ref := NewRef()
    called := false

    // Create element and attach to ref
    el := New(WithWidth(Fixed(10)), WithHeight(Fixed(5)))
    el.layout.X = 0
    el.layout.Y = 0
    el.layout.Width = 10
    el.layout.Height = 5
    ref.SetEl(el)

    // Simulate click inside element
    handled := HandleClicks(
        MouseEvent{Button: MouseLeft, Action: MousePress, X: 5, Y: 2},
        Click(ref, func() { called = true }),
    )

    if !handled {
        t.Fatal("expected click to be handled")
    }
    if !called {
        t.Fatal("handler not called")
    }
}

func TestHandleClicks_Miss(t *testing.T) {
    ref := NewRef()
    called := false

    // Create element and attach to ref
    el := New(WithWidth(Fixed(10)), WithHeight(Fixed(5)))
    el.layout.X = 0
    el.layout.Y = 0
    el.layout.Width = 10
    el.layout.Height = 5
    ref.SetEl(el)

    // Simulate click outside element
    handled := HandleClicks(
        MouseEvent{Button: MouseLeft, Action: MousePress, X: 15, Y: 10},
        Click(ref, func() { called = true }),
    )

    if handled {
        t.Fatal("expected click to not be handled")
    }
    if called {
        t.Fatal("handler should not be called")
    }
}

func TestHandleClicks_NilRef(t *testing.T) {
    ref := NewRef()
    // Don't attach element - ref.El() will be nil

    handled := HandleClicks(
        MouseEvent{Button: MouseLeft, Action: MousePress, X: 5, Y: 2},
        Click(ref, func() { t.Fatal("should not be called") }),
    )

    if handled {
        t.Fatal("expected nil ref to not handle click")
    }
}

func TestHandleClicks_MultipleBindings(t *testing.T) {
    ref1 := NewRef()
    ref2 := NewRef()
    var clicked string

    // Create elements
    el1 := New(WithWidth(Fixed(10)), WithHeight(Fixed(5)))
    el1.layout.X = 0
    el1.layout.Y = 0
    el1.layout.Width = 10
    el1.layout.Height = 5
    ref1.SetEl(el1)

    el2 := New(WithWidth(Fixed(10)), WithHeight(Fixed(5)))
    el2.layout.X = 20
    el2.layout.Y = 0
    el2.layout.Width = 10
    el2.layout.Height = 5
    ref2.SetEl(el2)

    // Click on second element
    handled := HandleClicks(
        MouseEvent{Button: MouseLeft, Action: MousePress, X: 25, Y: 2},
        Click(ref1, func() { clicked = "first" }),
        Click(ref2, func() { clicked = "second" }),
    )

    if !handled {
        t.Fatal("expected click to be handled")
    }
    if clicked != "second" {
        t.Fatalf("expected 'second', got '%s'", clicked)
    }
}

func TestHandleClicks_IgnoresNonLeftClick(t *testing.T) {
    ref := NewRef()
    el := New(WithWidth(Fixed(10)), WithHeight(Fixed(5)))
    el.layout.X = 0
    el.layout.Y = 0
    el.layout.Width = 10
    el.layout.Height = 5
    ref.SetEl(el)

    // Right click
    handled := HandleClicks(
        MouseEvent{Button: MouseRight, Action: MousePress, X: 5, Y: 2},
        Click(ref, func() { t.Fatal("should not be called") }),
    )

    if handled {
        t.Fatal("expected right click to not be handled")
    }
}

func TestHandleClicks_IgnoresRelease(t *testing.T) {
    ref := NewRef()
    el := New(WithWidth(Fixed(10)), WithHeight(Fixed(5)))
    el.layout.X = 0
    el.layout.Y = 0
    el.layout.Width = 10
    el.layout.Height = 5
    ref.SetEl(el)

    // Mouse release (not press)
    handled := HandleClicks(
        MouseEvent{Button: MouseLeft, Action: MouseRelease, X: 5, Y: 2},
        Click(ref, func() { t.Fatal("should not be called") }),
    )

    if handled {
        t.Fatal("expected release to not be handled")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestHandleClicks -v`
Expected: FAIL with "undefined: HandleClicks"

**Step 3: Write implementation**

```go
// click.go - add function

// HandleClicks checks a mouse event against click bindings and calls
// the first matching handler. Returns true if a click was handled.
// Use this in HandleMouse to simplify ref-based hit testing.
//
// Example:
//
//     func (c *counter) HandleMouse(me tui.MouseEvent) bool {
//         return tui.HandleClicks(me,
//             tui.Click(c.incrementBtn, c.increment),
//             tui.Click(c.decrementBtn, c.decrement),
//         )
//     }
func HandleClicks(me MouseEvent, bindings ...ClickBinding) bool {
    if me.Button != MouseLeft || me.Action != MousePress {
        return false
    }

    for _, b := range bindings {
        if b.Ref.El() != nil && b.Ref.El().ContainsPoint(me.X, me.Y) {
            b.Fn()
            return true
        }
    }

    return false
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestHandleClicks -v`
Expected: PASS

**Step 5: Commit**

```bash
git add click.go click_test.go
gcommit -m "feat: add HandleClicks helper for automatic ref hit testing"
```

---

## Task 3: Add WatcherProvider Interface

**Files:**
- Modify: `component.go`

**Step 1: Write failing test for WatcherProvider interface**

```go
// component_test.go (create if needed)
package tui

import "testing"

type mockWatcherProvider struct{}

func (m *mockWatcherProvider) Render() *Element { return New() }
func (m *mockWatcherProvider) Watchers() []Watcher {
    return []Watcher{
        NewTickerWatcher(time.Second, func() {}),
    }
}

func TestWatcherProvider_Interface(t *testing.T) {
    var _ WatcherProvider = &mockWatcherProvider{}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestWatcherProvider_Interface -v`
Expected: FAIL with "undefined: WatcherProvider"

**Step 3: Write implementation**

```go
// component.go - add interface

// WatcherProvider is an optional interface for components that provide
// timers, tickers, or channel watchers. Watchers() is called after the
// component is mounted and the returned watchers are started.
type WatcherProvider interface {
    Watchers() []Watcher
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestWatcherProvider_Interface -v`
Expected: PASS

**Step 5: Commit**

```bash
git add component.go component_test.go
gcommit -m "feat: add WatcherProvider interface for component-level watchers"
```

---

## Task 4: Collect Watchers from Components

**Files:**
- Modify: `app.go` or create `watcher_collect.go`
- Add tests

**Step 1: Write failing test for collecting watchers**

```go
// watcher_collect_test.go
package tui

import (
    "testing"
    "time"
)

type testWatcherComponent struct {
    watchers []Watcher
}

func (t *testWatcherComponent) Render() *Element { return New() }
func (t *testWatcherComponent) Watchers() []Watcher { return t.watchers }

func TestCollectComponentWatchers(t *testing.T) {
    // Create root with component
    root := New()

    comp := &testWatcherComponent{
        watchers: []Watcher{
            NewTickerWatcher(time.Second, func() {}),
        },
    }

    // Simulate component being mounted on element
    child := New()
    child.component = comp
    root.AddChild(child)

    watchers := collectComponentWatchers(root)

    if len(watchers) != 1 {
        t.Fatalf("expected 1 watcher, got %d", len(watchers))
    }
}

func TestCollectComponentWatchers_Nested(t *testing.T) {
    root := New()

    comp1 := &testWatcherComponent{
        watchers: []Watcher{NewTickerWatcher(time.Second, func() {})},
    }
    comp2 := &testWatcherComponent{
        watchers: []Watcher{
            NewTickerWatcher(time.Second, func() {}),
            NewTickerWatcher(time.Millisecond*500, func() {}),
        },
    }

    child1 := New()
    child1.component = comp1

    child2 := New()
    child2.component = comp2
    child1.AddChild(child2)

    root.AddChild(child1)

    watchers := collectComponentWatchers(root)

    if len(watchers) != 3 {
        t.Fatalf("expected 3 watchers, got %d", len(watchers))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestCollectComponentWatchers -v`
Expected: FAIL with "undefined: collectComponentWatchers"

**Step 3: Write implementation**

```go
// watcher_collect.go
package tui

// collectComponentWatchers walks the element tree and collects watchers
// from all components that implement WatcherProvider.
func collectComponentWatchers(root *Element) []Watcher {
    var watchers []Watcher

    walkComponents(root, func(comp Component) {
        if wp, ok := comp.(WatcherProvider); ok {
            watchers = append(watchers, wp.Watchers()...)
        }
    })

    return watchers
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestCollectComponentWatchers -v`
Expected: PASS

**Step 5: Commit**

```bash
git add watcher_collect.go watcher_collect_test.go
gcommit -m "feat: add tree walk to collect watchers from WatcherProvider components"
```

---

## Task 5: Integrate Component Watchers into App

**Files:**
- Modify: `app.go`
- Modify: `app_render.go`

**Step 1: Add component watchers tracking to App**

```go
// app.go - add field to App struct
type App struct {
    // ... existing fields ...

    // componentWatchers holds watchers from WatcherProvider components
    componentWatchers []Watcher
    componentWatchersStarted bool
}
```

**Step 2: Collect and start watchers after first render**

```go
// app_render.go - after render completes, before sweep
// Around line 53, add:

// Collect and start component watchers (once after first render)
if !a.componentWatchersStarted {
    a.componentWatchers = collectComponentWatchers(root)
    for _, w := range a.componentWatchers {
        w.Start()
    }
    a.componentWatchersStarted = true
}
```

**Step 3: Stop watchers on app close**

```go
// app_lifecycle.go - in Close() method
// Add before existing cleanup:

// Stop component watchers
for _, w := range a.componentWatchers {
    w.Stop()
}
```

**Step 4: Run all tests**

Run: `go test ./... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add app.go app_render.go app_lifecycle.go
gcommit -m "feat: integrate WatcherProvider watchers into app lifecycle"
```

---

## Task 6: Add NewChannelWatcher Generic Helper

**Files:**
- Modify: `watcher.go`
- Modify: `watcher_test.go` (if exists)

**Step 1: Write failing test for NewChannelWatcher**

```go
// watcher_test.go
func TestNewChannelWatcher(t *testing.T) {
    ch := make(chan string, 1)
    var received string

    w := NewChannelWatcher(ch, func(s string) {
        received = s
    })

    w.Start()
    ch <- "hello"

    // Give goroutine time to process
    time.Sleep(10 * time.Millisecond)

    w.Stop()

    if received != "hello" {
        t.Fatalf("expected 'hello', got '%s'", received)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestNewChannelWatcher -v`
Expected: FAIL with "undefined: NewChannelWatcher"

**Step 3: Write implementation**

```go
// watcher.go - add generic channel watcher

// ChannelWatcher watches a channel and calls a handler for each received value.
type ChannelWatcher[T any] struct {
    ch     <-chan T
    fn     func(T)
    stop   chan struct{}
    done   chan struct{}
}

// NewChannelWatcher creates a watcher that calls fn for each value received on ch.
func NewChannelWatcher[T any](ch <-chan T, fn func(T)) *ChannelWatcher[T] {
    return &ChannelWatcher[T]{
        ch:   ch,
        fn:   fn,
        stop: make(chan struct{}),
        done: make(chan struct{}),
    }
}

func (w *ChannelWatcher[T]) Start() {
    go func() {
        defer close(w.done)
        for {
            select {
            case <-w.stop:
                return
            case val, ok := <-w.ch:
                if !ok {
                    return
                }
                w.fn(val)
                MarkDirty()
            }
        }
    }()
}

func (w *ChannelWatcher[T]) Stop() {
    close(w.stop)
    <-w.done
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestNewChannelWatcher -v`
Expected: PASS

**Step 5: Commit**

```bash
git add watcher.go watcher_test.go
gcommit -m "feat: add generic NewChannelWatcher helper"
```

---

## Task 7: Update Interactive Example

**Files:**
- Modify: `examples/06-interactive/counter.gsx`
- Modify: `examples/06-interactive/timer.gsx`
- Modify: `examples/06-interactive/toggles.gsx`
- Modify: `examples/06-interactive/interactive.gsx`
- Delete: `examples/06-interactive/events.go` (prototype)
- Delete: `examples/06-interactive/registrar.go` (prototype)

**Step 1: Update counter.gsx**

```go
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type counter struct {
    count        *tui.State[int]
    decrementBtn *tui.Ref
    incrementBtn *tui.Ref
    resetBtn     *tui.Ref
}

func Counter() *counter {
    return &counter{
        count:        tui.NewState(0),
        decrementBtn: tui.NewRef(),
        incrementBtn: tui.NewRef(),
        resetBtn:     tui.NewRef(),
    }
}

func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('+', func(ke tui.KeyEvent) { c.increment() }),
        tui.OnRune('=', func(ke tui.KeyEvent) { c.increment() }),
        tui.OnRune('-', func(ke tui.KeyEvent) { c.decrement() }),
        tui.OnRune('0', func(ke tui.KeyEvent) { c.reset() }),
    }
}

func (c *counter) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.decrementBtn, c.decrement),
        tui.Click(c.incrementBtn, c.increment),
        tui.Click(c.resetBtn, c.reset),
    )
}

func (c *counter) increment() { c.count.Set(c.count.Get() + 1) }
func (c *counter) decrement() { c.count.Set(c.count.Get() - 1) }
func (c *counter) reset()     { c.count.Set(0) }

templ (c *counter) Render() {
    <div class="border-single p-1 flex-col gap-1" flexGrow={1.0}>
        <span class="text-gradient-cyan-blue font-bold">{"Counter"}</span>
        <div class="flex gap-1 items-center">
            <span class="font-dim">Count:</span>
            <span class="text-cyan font-bold">{fmt.Sprintf("%d", c.count.Get())}</span>
        </div>
        <div class="flex gap-1">
            <button ref={c.decrementBtn}>{" - "}</button>
            <button ref={c.incrementBtn}>{" + "}</button>
            <button ref={c.resetBtn}>{" 0 "}</button>
        </div>
        @if c.count.Get() > 0 {
            <span class="text-green font-bold">{"Positive"}</span>
        } @else @if c.count.Get() < 0 {
            <span class="text-red font-bold">{"Negative"}</span>
        } @else {
            <span class="text-blue font-bold">{"Zero"}</span>
        }
        <span class="font-dim">{"click btns or +/-/0"}</span>
    </div>
}
```

**Step 2: Update timer.gsx**

```go
package main

import (
    "fmt"
    "time"
    tui "github.com/grindlemire/go-tui"
)

type timer struct {
    elapsed *tui.State[int]
    running *tui.State[bool]
}

func Timer() *timer {
    return &timer{
        elapsed: tui.NewState(0),
        running: tui.NewState(true),
    }
}

func (t *timer) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune(' ', func(ke tui.KeyEvent) { t.toggleRunning() }),
        tui.OnRune('r', func(ke tui.KeyEvent) { t.resetTimer() }),
    }
}

func (t *timer) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.NewTickerWatcher(time.Second, t.tick),
    }
}

func (t *timer) toggleRunning() { t.running.Set(!t.running.Get()) }
func (t *timer) resetTimer()    { t.elapsed.Set(0) }

func (t *timer) tick() {
    if t.running.Get() {
        t.elapsed.Set(t.elapsed.Get() + 1)
    }
}

func formatTime(seconds int) string {
    m := seconds / 60
    s := seconds - (m * 60)
    return fmt.Sprintf("%02d:%02d", m, s)
}

templ (t *timer) Render() {
    <div class="border-single p-1 flex-col gap-1" flexGrow={1.0}>
        <span class="text-gradient-blue-cyan font-bold">{"Timer"}</span>
        <div class="flex gap-1 items-center">
            <span class="font-dim">Elapsed:</span>
            <span class="text-blue font-bold">{formatTime(t.elapsed.Get())}</span>
        </div>
        @if t.running.Get() {
            <span class="text-green font-bold">{"Running"}</span>
        } @else {
            <span class="text-red font-bold">{"Stopped"}</span>
        }
        <span class="font-dim">{"[space] toggle [r] reset"}</span>
    </div>
}
```

**Step 3: Update toggles.gsx**

```go
package main

import tui "github.com/grindlemire/go-tui"

type toggles struct {
    sound     *tui.State[bool]
    notify    *tui.State[bool]
    dark      *tui.State[bool]
    soundBtn  *tui.Ref
    notifyBtn *tui.Ref
    themeBtn  *tui.Ref
}

func Toggles() *toggles {
    return &toggles{
        sound:     tui.NewState(true),
        notify:    tui.NewState(false),
        dark:      tui.NewState(false),
        soundBtn:  tui.NewRef(),
        notifyBtn: tui.NewRef(),
        themeBtn:  tui.NewRef(),
    }
}

func (t *toggles) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('1', func(ke tui.KeyEvent) { t.toggleSound() }),
        tui.OnRune('2', func(ke tui.KeyEvent) { t.toggleNotify() }),
        tui.OnRune('3', func(ke tui.KeyEvent) { t.toggleTheme() }),
    }
}

func (t *toggles) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(t.soundBtn, t.toggleSound),
        tui.Click(t.notifyBtn, t.toggleNotify),
        tui.Click(t.themeBtn, t.toggleTheme),
    )
}

func (t *toggles) toggleSound()  { t.sound.Set(!t.sound.Get()) }
func (t *toggles) toggleNotify() { t.notify.Set(!t.notify.Get()) }
func (t *toggles) toggleTheme()  { t.dark.Set(!t.dark.Get()) }

templ (t *toggles) Render() {
    <div class="border-single p-1 flex-col gap-1" flexGrow={1.0}>
        <span class="text-gradient-green-cyan font-bold">{"Toggles"}</span>
        <div class="flex gap-1 items-center">
            <button ref={t.soundBtn}>{"Sound  "}</button>
            @if t.sound.Get() {
                <span class="text-green font-bold">ON</span>
            } @else {
                <span class="text-red font-bold">OFF</span>
            }
        </div>
        <div class="flex gap-1 items-center">
            <button ref={t.notifyBtn}>{"Notify "}</button>
            @if t.notify.Get() {
                <span class="text-green font-bold">ON</span>
            } @else {
                <span class="text-red font-bold">OFF</span>
            }
        </div>
        <div class="flex gap-1 items-center">
            <button ref={t.themeBtn}>{"Theme  "}</button>
            @if t.dark.Get() {
                <span class="text-cyan font-bold">Dark</span>
            } @else {
                <span class="text-yellow font-bold">Light</span>
            }
        </div>
        <span class="font-dim">{"click or press 1/2/3"}</span>
    </div>
}
```

**Step 4: Update interactive.gsx (main app)**

```go
package main

import tui "github.com/grindlemire/go-tui"

type interactiveApp struct{}

func Interactive() *interactiveApp {
    return &interactiveApp{}
}

func (a *interactiveApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('q', func(ke tui.KeyEvent) { tui.Stop() }),
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
    }
}

templ (a *interactiveApp) Render() {
    <div class="flex-col p-1 border-rounded gap-1">
        <div class="flex justify-between">
            <span class="text-gradient-cyan-magenta font-bold">{"Interactive Elements"}</span>
            <span class="font-dim">{"[q] quit"}</span>
        </div>
        <div class="flex gap-1">
            @Counter()
            @Timer()
        </div>
        <div class="flex gap-1">
            @Toggles()
        </div>
    </div>
}
```

**Step 5: Delete prototype files**

```bash
rm examples/06-interactive/events.go
rm examples/06-interactive/registrar.go
rm examples/06-interactive/event_inspector.gsx
rm examples/06-interactive/event_inspector_gsx.go
```

**Step 6: Generate and test**

```bash
go run ./cmd/tui generate ./examples/06-interactive/
cd examples/06-interactive && go run .
```

**Step 7: Commit**

```bash
git add examples/06-interactive/
gcommit -m "refactor: update interactive example to use HandleClicks and Watchers"
```

---

## Task 8: Documentation

**Files:**
- Update: `CLAUDE.md`

**Step 1: Add HandleClicks documentation**

Add to "Key Types" or new "Event Handling" section:

```markdown
### Mouse Click Handling

Use `HandleClicks` for automatic ref-based hit testing:

    func (c *counter) HandleMouse(me tui.MouseEvent) bool {
        return tui.HandleClicks(me,
            tui.Click(c.incrementBtn, c.increment),
            tui.Click(c.decrementBtn, c.decrement),
        )
    }

### Component Watchers

Implement `WatcherProvider` for component-level timers/channels:

    func (t *timer) Watchers() []tui.Watcher {
        return []tui.Watcher{
            tui.NewTickerWatcher(time.Second, t.tick),
            tui.NewChannelWatcher(t.dataChan, t.onData),
        }
    }
```

**Step 2: Commit**

```bash
git add CLAUDE.md
gcommit -m "docs: document HandleClicks and WatcherProvider APIs"
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Add Click type | click.go |
| 2 | Add HandleClicks helper | click.go |
| 3 | Add WatcherProvider interface | component.go |
| 4 | Collect watchers from components | watcher_collect.go |
| 5 | Integrate into app lifecycle | app.go, app_render.go |
| 6 | Add generic NewChannelWatcher | watcher.go |
| 7 | Update interactive example | examples/06-interactive/ |
| 8 | Documentation | CLAUDE.md |

**Total: ~8 tasks, ~40 steps**

## API Summary

| Interface | Method | Purpose |
|-----------|--------|---------|
| `Component` | `Render()` | Return element tree |
| `KeyListener` | `KeyMap()` | Dynamic key bindings (every render) |
| `MouseListener` | `HandleMouse()` | Handle mouse with `HandleClicks` helper |
| `WatcherProvider` | `Watchers()` | Timers, channels (collected once) |
| `Initializer` | `Init()` | One-time setup, return cleanup |

## Helper Functions

| Function | Purpose |
|----------|---------|
| `Click(ref, fn)` | Create click binding |
| `HandleClicks(me, ...bindings)` | Automatic hit testing |
| `NewTickerWatcher(interval, fn)` | Create timer watcher |
| `NewChannelWatcher(ch, fn)` | Create channel watcher |
