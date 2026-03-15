# State and Reactivity

## Overview

go-tui uses a reactive model: when state changes, the UI re-renders automatically. You don't manually update element text or toggle visibility. Instead, you store values in `State[T]` containers. Your `Render` method reads state with `.Get()`, and any call to `.Set()` or `.Update()` marks the UI as dirty and triggers a new render pass.

## Creating State

State lives as fields on your struct component. Declare each piece of state as a `*tui.State[T]` where `T` is whatever type you need, then initialize it in the constructor with `tui.NewState`:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

type myApp struct {
    count    *tui.State[int]
    name     *tui.State[string]
    visible  *tui.State[bool]
    items    *tui.State[[]string]
}

func MyApp() *myApp {
    return &myApp{
        count:   tui.NewState(0),
        name:    tui.NewState("world"),
        visible: tui.NewState(true),
        items:   tui.NewState([]string{"Go", "Rust", "Zig"}),
    }
}
```

`NewState` accepts any type through Go generics. The value you pass becomes the initial state. The framework binds each `State` to the running `App` automatically when your component mounts.

## Reading State

Call `.Get()` to read the current value. Use it directly in your render method, inside Go expressions, or in control flow conditions:

```gsx
templ (m *myApp) Render() {
    <div class="flex-col gap-1 p-1">
        <span class="text-cyan font-bold">{fmt.Sprintf("Count: %d", m.count.Get())}</span>
        <span>{"Hello, " + m.name.Get()}</span>

        if m.visible.Get() {
            <span class="text-green">This is visible</span>
        }

        for i, item := range m.items.Get() {
            <span>{fmt.Sprintf("%d. %s", i+1, item)}</span>
        }
    </div>
}
```

`.Get()` is thread-safe for reads from any goroutine. In render methods, just call it wherever you need the current value.

## Writing State

There are two ways to change state: `.Set()` for replacing the value outright, and `.Update()` for modifying it based on the current value.

### Set

`.Set()` replaces the current value and triggers a re-render:

```go
func (m *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('r'), func(ke tui.KeyEvent) { m.count.Set(0) }),
        tui.On(tui.Rune('h'), func(ke tui.KeyEvent) { m.visible.Set(!m.visible.Get()) }),
    }
}
```

Use `.Set()` when you have the exact value you want. Setting the count to zero, toggling a boolean, replacing a string with new content.

### Update

`.Update()` takes a function that receives the current value and returns the new value. It's shorthand for `s.Set(fn(s.Get()))`:

```go
tui.On(tui.Rune('+'), func(ke tui.KeyEvent) {
    m.count.Update(func(v int) int { return v + 1 })
}),
tui.On(tui.Rune('-'), func(ke tui.KeyEvent) {
    m.count.Update(func(v int) int { return v - 1 })
}),
```

Use `.Update()` when the new value depends on the old one. Incrementing a counter, appending to a slice, toggling a flag.

Both `.Set()` and `.Update()` must be called from the main event loop (key handlers, mouse handlers, watcher callbacks). For background goroutines, use `app.QueueUpdate()` to schedule the change on the event loop instead.

## Conditional Rendering

`if` and `else` let you show or hide elements based on state. The condition is a Go boolean expression, evaluated fresh on each render:

```gsx
templ (m *myApp) Render() {
    <div class="flex-col gap-1 p-1 border-rounded">
        <span class="font-bold">Status</span>
        <div class="flex gap-1">
            <span class="font-dim">Sign:</span>
            if m.count.Get() > 0 {
                <span class="text-green font-bold">Positive</span>
            } else if m.count.Get() < 0 {
                <span class="text-red font-bold">Negative</span>
            } else {
                <span class="text-blue font-bold">Zero</span>
            }
        </div>

        <div class="flex gap-1">
            <span class="font-dim">Parity:</span>
            if m.count.Get()%2 == 0 {
                <span class="text-cyan">Even</span>
            } else {
                <span class="text-magenta">Odd</span>
            }
        </div>
    </div>
}
```

Chain `else if` for multiple branches. These conditions are re-evaluated on every render, so changes to state are reflected immediately.

## List Rendering

`for` with `range` renders a collection. Combine it with state to build dynamic lists where the data or selection can change:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

type listApp struct {
    items    []string
    selected *tui.State[int]
}

func ListApp() *listApp {
    return &listApp{
        items:    []string{"Rust", "Go", "TypeScript", "Python", "Zig"},
        selected: tui.NewState(0),
    }
}

func (l *listApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) {
            l.selected.Update(func(v int) int {
                if v >= len(l.items)-1 {
                    return 0
                }
                return v + 1
            })
        }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) {
            l.selected.Update(func(v int) int {
                if v <= 0 {
                    return len(l.items) - 1
                }
                return v - 1
            })
        }),
    }
}

templ (l *listApp) Render() {
    <div class="flex-col p-1 border-rounded border-cyan">
        <span class="font-bold text-gradient-cyan-magenta">Pick a Language</span>
        <br />
        for i, item := range l.items {
            if i == l.selected.Get() {
                <span class="text-cyan font-bold">{fmt.Sprintf("  > %s", item)}</span>
            } else {
                <span class="font-dim">{fmt.Sprintf("    %s", item)}</span>
            }
        }
        <br />
        picked := l.items[l.selected.Get()]
        <span class="font-dim">{fmt.Sprintf("Selected: %s", picked)}</span>
    </div>
}
```

The `selected` state tracks which item has the cursor. Pressing `j` and `k` moves it up and down with wrapping. On each render, the `for` loop checks every item against the selected index and applies highlighting to the match.

For lists stored in state (e.g. `*tui.State[[]string]`), call `.Get()` in the range expression:

```gsx
for i, msg := range m.messages.Get() {
    <span>{fmt.Sprintf("[%d] %s", i, msg)}</span>
}
```

## Computed Values

Inside a `templ` body you can write regular Go variable assignments. This is useful when you want to compute a value once and reference it in multiple places:

```gsx
templ (m *myApp) Render() {
    label := fmt.Sprintf("Count: %d", m.count.Get())
    status := statusText(m.count.Get())
    <div class="flex-col gap-1 p-1">
        <span class="text-cyan font-bold">{label}</span>
        <span class="font-dim">{status}</span>
    </div>
}
```

Where `statusText` is a plain Go helper:

```go
func statusText(n int) string {
    if n > 20 {
        return "high"
    }
    if n > 0 {
        return "medium"
    }
    if n == 0 {
        return "zero"
    }
    return "negative"
}
```

These are normal Go short variable declarations. They're re-evaluated on each render, so they always reflect the latest state.

### Element bindings (:=) for reusable elements

Element bindings are different from Go variable assignments. They bind an **element** to a name so you can reuse it in multiple places:

```gsx
templ (m *myApp) Render() {
    badge := <span class="bg-cyan text-black px-1 font-bold">{fmt.Sprintf("%d", m.count.Get())}</span>
    <div class="flex-col gap-1 p-1">
        <div class="flex gap-1">
            <span>Current count:</span>
            {badge}
        </div>
        <div class="flex gap-1">
            <span>Shown again:</span>
            {badge}
        </div>
    </div>
}
```

Use regular Go assignments for computed strings, numbers, and booleans. Use `:=` element bindings when you want to define a reusable element fragment.

## Batching Updates

When you need to change multiple state values at once, `app.Batch()` groups them into a single re-render. Without batching, each `.Set()` call marks the UI dirty independently. With batching, the framework waits until the batch function returns before running any binding callbacks:

```go
func (m *myApp) reset(ke tui.KeyEvent) {
    ke.App().Batch(func() {
        m.count.Set(0)
        m.name.Set("world")
        m.visible.Set(true)
    })
}
```

Inside the batch function, each `.Set()` records its bindings but doesn't fire them yet. When the function returns, all pending bindings run once with their final values. If the same binding is triggered multiple times within a batch (say, by setting the same state twice), it fires only once with the last value.

Batches can nest. Bindings are deferred until the outermost batch completes:

```go
ke.App().Batch(func() {
    m.count.Set(1)
    ke.App().Batch(func() {
        m.count.Set(2)
    })
    m.count.Set(3)
})
// Binding fires once with value 3
```

You typically don't need batching for single state changes. It's useful for resetting forms, swapping between modes, or any situation where multiple fields change together and you want a single coherent re-render.

## State Bindings

`.Bind()` registers a callback that fires whenever the state changes. It returns an `Unbind` function you can call to stop receiving updates:

```go
unbind := m.count.Bind(func(v int) {
    fmt.Println("count changed to", v)
})

// Later, to stop listening:
unbind()
```

Bindings run in registration order. They fire on every `.Set()` or `.Update()` call (or at the end of a batch). The callback receives the new value as its argument.

Most of the time you won't need manual bindings. The render method already reads state with `.Get()`, and re-renders happen automatically when state changes. Where bindings become useful is when you need something to happen *outside* the render cycle. For example, writing to a log file whenever a value changes, updating a second piece of state that's derived from the first, or sending a message over a channel to notify another goroutine. If you find yourself reaching for `.Bind()` just to update the UI, you probably don't need it. The `.Get()` in your render method already covers that.

## Complete Example

Here's a full app with a counter, status display, selectable list, and a reset key that uses batching:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

type demoApp struct {
    count    *tui.State[int]
    selected *tui.State[int]
    items    []string
}

func Demo() *demoApp {
    return &demoApp{
        count:    tui.NewState(0),
        selected: tui.NewState(0),
        items:    []string{"Rust", "Go", "TypeScript", "Python", "Zig"},
    }
}

func (d *demoApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('+'), func(ke tui.KeyEvent) {
            d.count.Update(func(v int) int { return v + 1 })
        }),
        tui.On(tui.Rune('-'), func(ke tui.KeyEvent) {
            d.count.Update(func(v int) int { return v - 1 })
        }),
        tui.On(tui.Rune('r'), func(ke tui.KeyEvent) {
            ke.App().Batch(func() {
                d.count.Set(0)
                d.selected.Set(0)
            })
        }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { d.selectNext() }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { d.selectPrev() }),
        tui.On(tui.KeyDown, func(ke tui.KeyEvent) { d.selectNext() }),
        tui.On(tui.KeyUp, func(ke tui.KeyEvent) { d.selectPrev() }),
    }
}

func (d *demoApp) selectNext() {
    d.selected.Update(func(v int) int {
        if v >= len(d.items)-1 {
            return 0
        }
        return v + 1
    })
}

func (d *demoApp) selectPrev() {
    d.selected.Update(func(v int) int {
        if v <= 0 {
            return len(d.items) - 1
        }
        return v - 1
    })
}

func signLabel(n int) string {
    if n > 0 {
        return "Positive"
    }
    if n < 0 {
        return "Negative"
    }
    return "Zero"
}

func signClass(n int) string {
    if n > 0 {
        return "text-green font-bold"
    }
    if n < 0 {
        return "text-red font-bold"
    }
    return "text-blue font-bold"
}

templ (d *demoApp) Render() {
    <div class="flex-col p-1 border-rounded border-cyan">
        <span class="text-gradient-cyan-magenta font-bold">State Demo</span>

        <div class="flex">
            // Counter panel
            <div class="flex-col border-rounded p-1 gap-1 items-center justify-center" flexGrow={1.0}>
                <span class="font-bold">Counter</span>
                <span class="text-cyan font-bold">{fmt.Sprintf("%d", d.count.Get())}</span>
                <div class="flex gap-1 justify-center">
                    <span class="font-dim">+/-  r:reset</span>
                </div>
            </div>

            // Status panel
            <div class="flex-col border-rounded p-1 gap-1" flexGrow={2.0}>
                <span class="font-bold">Status</span>
                <div class="flex gap-1">
                    <span class="font-dim">Sign:</span>
                    <span class={signClass(d.count.Get())}>{signLabel(d.count.Get())}</span>
                </div>
                <div class="flex gap-1">
                    <span class="font-dim">Parity:</span>
                    if d.count.Get()%2 == 0 {
                        <span class="text-cyan">Even</span>
                    } else {
                        <span class="text-magenta">Odd</span>
                    }
                </div>
            </div>
        </div>

        // Selectable list
        <div class="flex-col border-rounded p-1 gap-1">
            <span class="font-bold">Languages</span>
            for i, item := range d.items {
                if i == d.selected.Get() {
                    <span class="text-cyan font-bold">{fmt.Sprintf("  > %s", item)}</span>
                } else {
                    <span class="font-dim">{fmt.Sprintf("    %s", item)}</span>
                }
            }
        </div>

        <div class="flex justify-center">
            <span class="font-dim">+/- count | j/k navigate | r reset | esc quit</span>
        </div>
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
        tui.WithRootComponent(Demo()),
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

You should see something like this:

![State and Reactivity screenshot](/guides/06.png)

## Next Steps

- [Components](components) -- Component patterns, composition, and lifecycle interfaces
- [Event Handling](events) -- Keyboard and mouse input in depth
- [Timers, Watchers, and Channels](watchers) -- Background operations that update state over time
