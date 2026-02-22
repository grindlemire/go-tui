# Components

## Overview

go-tui has two kinds of components: pure templ functions and struct components.

Pure templ functions are stateless. They take parameters, return elements, and that's it. Use them for reusable visual pieces like cards, badges, and headers.

Struct components carry state. They own `State[T]` fields, handle keyboard input, run timers, and manage their own lifecycle. Use them for anything interactive or anything that changes over time.

Both kinds compose freely. A struct component's render method can call pure components, and pure components can nest other pure components inside themselves. You build complex UIs by stacking simple pieces.

## Pure Components

A pure component is a `templ` function. It takes parameters and returns a chunk of UI:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

templ Badge(label string) {
    <span class="text-cyan font-bold px-1">{label}</span>
}

templ Header(title string) {
    <div class="border-rounded p-1 flex justify-center">
        <span class="text-gradient-cyan-magenta font-bold">{title}</span>
    </div>
}

templ StatusLine(label string, value string) {
    <div class="flex gap-1">
        <span class="font-dim">{label}</span>
        <span class="text-cyan font-bold">{value}</span>
    </div>
}
```

Call them from other templates just like HTML elements, using `@`:

```gsx
templ (a *myApp) Render() {
    <div class="flex-col gap-1 p-1">
        @Header("Dashboard")
        @Badge("v1.0")
        @StatusLine("Status:", "Online")
    </div>
}
```

Parameters can be any Go type: strings, ints, booleans, slices, custom structs. The only rule is that pure components don't hold state between renders. Every call constructs a fresh view from the parameters you give it.

### Children Slot

Pure components can accept nested content through the `{children...}` slot. This lets you build wrapper components that add layout, borders, or styling around arbitrary content:

```gsx
templ Card(title string) {
    <div class="border-rounded p-1 flex-col gap-1">
        <span class="text-gradient-cyan-magenta font-bold">{title}</span>
        <hr class="border-single" />
        {children...}
    </div>
}
```

Pass children by nesting elements inside the component call:

```gsx
templ (a *myApp) Render() {
    <div class="flex gap-2">
        @Card("System Info") {
            @StatusLine("Version:", "1.2.0")
            @StatusLine("Uptime:", "3d 14h")
        }
        @Card("Config") {
            @StatusLine("Theme:", "Dark")
            @StatusLine("Notify:", "On")
        }
    </div>
}
```

Each `Card` renders its title and divider, then places whatever children you passed where `{children...}` appears. The children can be elements, other pure components, or any mix.

`tui generate` compiles the children into a `[]*tui.Element` slice and passes it as a parameter to the generated function. The slot expands to a loop that adds each child element to the parent container.

Struct components also support `{children...}` — see [Children Slot](#children-slot-2) in the struct components section below.

### When to Use Pure Components

Use pure components for anything that doesn't need its own state or event handlers:

- Visual elements: cards, badges, headers, dividers
- Layout wrappers: bordered sections, padded containers, centered panels
- Data display: status lines, key-value pairs, formatted labels
- Repeated patterns: list items, table rows, form fields

If you find yourself wanting to add a `*tui.State` field, you need a struct component instead.

## Struct Components

Struct components are Go structs with a `templ` render method. They hold state, handle input, and manage their own lifecycle:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

type counter struct {
    count *tui.State[int]
}

func Counter() *counter {
    return &counter{
        count: tui.NewState(0),
    }
}

func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('+', func(ke tui.KeyEvent) {
            c.count.Update(func(v int) int { return v + 1 })
        }),
        tui.OnRune('-', func(ke tui.KeyEvent) {
            c.count.Update(func(v int) int { return v - 1 })
        }),
    }
}

templ (c *counter) Render() {
    <div class="flex-col p-1 border-rounded border-cyan items-center gap-1">
        <span class="font-bold">Counter</span>
        <span class="text-cyan font-bold">{fmt.Sprintf("%d", c.count.Get())}</span>
        <span class="font-dim">+/- to change, esc to quit</span>
    </div>
}
```

There are three parts: the struct with state fields, a constructor function that initializes it, and the render method using `templ (receiver) Render()` syntax.

The render method works exactly like a pure templ function, but it has access to the struct's fields through the receiver. The generated code turns it into a `Render(app *App) *Element` method, which satisfies the `Component` interface.

### Children Slot

Struct components can also accept children using `{children...}`. Add a `children []*tui.Element` field to the struct and accept it in the constructor:

```gsx
type panel struct {
    title    string
    children []*tui.Element
}

func NewPanel(title string, children []*tui.Element) *panel {
    return &panel{title: title, children: children}
}

templ (p *panel) Render() {
    <div class="border-rounded p-1 flex-col gap-1">
        <span class="font-bold text-gradient-cyan-magenta">{p.title}</span>
        {children...}
    </div>
}
```

Callers use the same block syntax as pure components:

```gsx
templ (a *myApp) Render() {
    @NewPanel("Items") {
        <span>First item</span>
        <span>Second item</span>
    }
}
```

The generated code builds a `[]*tui.Element` slice from the children block and passes it as the last constructor argument. On re-renders, `UpdateProps` copies the fresh children to the cached instance automatically, so the panel always reflects the latest content.

This is useful when you need a wrapper component that carries its own state (timers, scroll position, internal selections) while still accepting arbitrary content from the parent.

### The Component Interface

Every struct component implements this interface:

```go
type Component interface {
    Render(app *App) *Element
}
```

You don't write `Render(app *App) *Element` by hand. The `templ` keyword handles the signature. You write `templ (c *counter) Render()` and `tui generate` produces the correct Go method.

## Component Interfaces

Struct components can implement additional interfaces for input handling, background tasks, and lifecycle hooks. All are optional -- implement only what you need.

### KeyListener

Handle keyboard input by implementing `KeyMap() tui.KeyMap`:

```go
type KeyListener interface {
    KeyMap() KeyMap
}
```

`KeyMap()` returns a slice of key bindings. The framework calls it on every render pass, so you can return different bindings based on current state:

```go
func (s *search) KeyMap() tui.KeyMap {
    if s.active.Get() {
        return tui.KeyMap{
            tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { s.active.Set(false) }),
            tui.OnRunesStop(func(ke tui.KeyEvent) {
                s.query.Update(func(q string) string { return q + string(ke.Rune) })
            }),
        }
    }
    return tui.KeyMap{
        tui.OnRune('/', func(ke tui.KeyEvent) { s.active.Set(true) }),
    }
}
```

### MouseListener

Handle mouse input by implementing `HandleMouse(MouseEvent) bool`:

```go
type MouseListener interface {
    HandleMouse(MouseEvent) bool
}
```

Return `true` if you handled the event, `false` to let it propagate. The typical pattern uses refs and `HandleClicks`:

```go
func (c *counter) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.incBtn, func() {
            c.count.Update(func(v int) int { return v + 1 })
        }),
        tui.Click(c.decBtn, func() {
            c.count.Update(func(v int) int { return v - 1 })
        }),
    )
}
```

Mouse events require `tui.WithMouse()` when creating the app. See the [Event Handling](events) guide for details.

### WatcherProvider

Run background timers or receive from Go channels by implementing `Watchers() []Watcher`:

```go
type WatcherProvider interface {
    Watchers() []Watcher
}
```

Watchers start when the component mounts and stop when it unmounts. Callbacks run on the UI thread, so you can safely update state:

```go
func (t *timer) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, func() {
            t.elapsed.Update(func(v int) int { return v + 1 })
        }),
    }
}
```

See [Timers, Watchers, and Channels](watchers) for the full API.

### Initializer

Run setup code when a component first mounts by implementing `Init() func()`:

```go
type Initializer interface {
    Init() func()
}
```

`Init()` is called once when the component enters the tree. The returned function, if non-nil, runs when the component leaves the tree. This pairs setup and cleanup at the same call site:

```go
func (c *conn) Init() func() {
    c.ws = connectWebSocket(c.url)
    go c.readLoop()
    return func() {
        c.ws.Close()
    }
}
```

### AppBinder

Wire up `State` and `Events` fields to the app by implementing `BindApp(app *App)`:

```go
type AppBinder interface {
    BindApp(app *App)
}
```

You rarely write this by hand. The code generator produces `BindApp` methods automatically for any struct component that has `*tui.State` or `*tui.Events` fields. The mount system calls it when the component is first created and again after props updates.

### PropsUpdater

Receive updated props when a cached component is re-rendered:

```go
type PropsUpdater interface {
    UpdateProps(fresh Component)
}
```

You never write this by hand. The code generator produces `UpdateProps` methods automatically for all components. For struct components, it copies non-state fields (skipping `*tui.State`, `*tui.Ref`, channels, and functions) from a fresh instance to the cached one. The mount system calls it on subsequent renders so the cached instance picks up any changed props without losing its internal state.

## Component Composition

Components compose by nesting. Call a pure component with `@`, and it inlines its elements into the parent. Call a struct component the same way, and the framework handles mounting and caching behind the scenes.

### Pure Components Inside Struct Renders

The most common pattern. A struct component's render method calls pure components for layout and display:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

templ Badge(label string, style string) {
    <span class={style + " font-bold px-1"}>{label}</span>
}

templ InfoRow(label string, value string) {
    <div class="flex gap-1">
        <span class="font-dim">{label}</span>
        <span class="text-cyan font-bold">{value}</span>
    </div>
}

templ Card(title string) {
    <div class="border-rounded p-1 flex-col gap-1">
        <span class="text-gradient-cyan-magenta font-bold">{title}</span>
        <hr class="border-single" />
        {children...}
    </div>
}

type userProfile struct {
    name  string
    role  string
    email string
}

func UserProfile(name, role, email string) *userProfile {
    return &userProfile{name: name, role: role, email: email}
}

func (u *userProfile) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

templ (u *userProfile) Render() {
    <div class="flex-col gap-1 p-1">
        @Card("User Profile") {
            <div class="flex-col gap-1">
                @InfoRow("Name:", u.name)
                @InfoRow("Role:", u.role)
                @InfoRow("Email:", u.email)
                <div class="flex gap-1">
                    <span class="font-dim">Status:</span>
                    @Badge("Active", "text-green")
                </div>
            </div>
        }
    </div>
}
```

### Struct Components Inside Struct Renders

When one struct component renders another struct component, the framework uses `app.Mount` to cache the child instance. This means the child keeps its state across parent re-renders:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

type sidebar struct {
    category *tui.State[string]
    selected *tui.State[int]
}

func Sidebar(category *tui.State[string]) *sidebar {
    return &sidebar{
        category: category,
        selected: tui.NewState(0),
    }
}

func (s *sidebar) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('j', func(ke tui.KeyEvent) {
            s.selected.Update(func(v int) int { return v + 1 })
        }),
        tui.OnRune('k', func(ke tui.KeyEvent) {
            s.selected.Update(func(v int) int { return v - 1 })
        }),
    }
}

templ (s *sidebar) Render() {
    <div class="flex-col border-single p-1" width={20}>
        <span class="font-bold text-cyan">Sidebar</span>
        // ... sidebar content ...
    </div>
}

type myApp struct {
    category *tui.State[string]
}

func MyApp() *myApp {
    return &myApp{
        category: tui.NewState("Documents"),
    }
}

func (a *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

templ (a *myApp) Render() {
    <div class="flex h-full">
        @Sidebar(a.category)
        <div class="flex-col flex-1 p-1">
            <span>Main content</span>
        </div>
    </div>
}
```

When `tui generate` processes this, it turns `@Sidebar(a.category)` inside a struct render method into an `app.Mount` call:

```go
app.Mount(a, 0, func() tui.Component {
    return Sidebar(a.category)
})
```

The mount system caches the `sidebar` instance by its parent and position index. On the first render, it creates the sidebar, binds its state to the app, and calls `Init()` if implemented. On subsequent renders, it reuses the cached instance and calls `UpdateProps` if the child implements it.

The sidebar's `selected` state persists even when the parent re-renders, because the framework reuses the cached instance instead of creating a new one.

### Shared State

Parent and child components can share state by passing `*tui.State` values through constructors. Changes from either side trigger a re-render:

```gsx
type myApp struct {
    category *tui.State[string]
}

func MyApp() *myApp {
    return &myApp{
        category: tui.NewState("Documents"),
    }
}

templ (a *myApp) Render() {
    <div class="flex h-full">
        @Sidebar(a.category)
        @Content(a.category)
    </div>
}
```

Both `Sidebar` and `Content` receive the same `category` state. When the sidebar changes it, the content panel sees the new value on the next render.

## Composition Patterns

### Layout Wrapper

A pure component that wraps content in a consistent layout:

```gsx
templ PageLayout(title string) {
    <div class="flex-col h-full">
        <div class="flex justify-center p-1 border-single">
            <span class="text-gradient-cyan-magenta font-bold">{title}</span>
        </div>
        <div class="flex-col flex-1 p-1">
            {children...}
        </div>
        <div class="flex justify-center p-1 border-single">
            <span class="font-dim">Press q to quit</span>
        </div>
    </div>
}
```

### Data Display

Pure components for showing key-value data:

```gsx
templ KeyValue(key string, value string) {
    <div class="flex gap-1">
        <span class="font-dim">{key + ":"}</span>
        <span class="text-cyan">{value}</span>
    </div>
}

templ Section(title string) {
    <div class="flex-col border-rounded p-1 gap-1">
        <span class="font-bold">{title}</span>
        <hr class="border-single" />
        {children...}
    </div>
}
```

### Container Component

A reusable scrollable panel:

```gsx
templ ScrollPanel(title string, height int) {
    <div class="flex-col border-rounded" height={height}>
        <div class="flex p-1">
            <span class="font-bold text-cyan">{title}</span>
        </div>
        <div class="flex-col overflow-y-scroll flex-1 p-1">
            {children...}
        </div>
    </div>
}
```

## Complete Example

A full app with pure components, a struct component, and composition:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

// Pure components

templ Badge(label string, color string) {
    <span class={color + " font-bold px-1"}>{label}</span>
}

templ StatusLine(label string, value string) {
    <div class="flex gap-1">
        <span class="font-dim">{label}</span>
        <span class="text-cyan font-bold">{value}</span>
    </div>
}

templ Card(title string) {
    <div class="border-rounded p-1 flex-col gap-1 w-full" flexGrow={1.0}>
        <span class="text-gradient-cyan-magenta font-bold">{title}</span>
        <hr class="border-single" />
        {children...}
    </div>
}

// Tab content components

templ OverviewTab() {
    <div class="flex gap-1">
        @Card("System") {
            @StatusLine("CPU:", "42%")
            @StatusLine("Memory:", "1.2 GB")
            @StatusLine("Disk:", "68%")
        }
        @Card("Network") {
            @StatusLine("In:", "12 MB/s")
            @StatusLine("Out:", "3 MB/s")
            @StatusLine("Latency:", "24ms")
        }
    </div>
}

templ MetricsTab() {
    <div class="flex gap-1">
        @Card("Performance") {
            @StatusLine("Requests:", "1.2k/s")
            @StatusLine("P99:", "145ms")
            @StatusLine("Errors:", "0.02%")
        }
        @Card("Storage") {
            @StatusLine("Used:", "42 GB")
            @StatusLine("Free:", "118 GB")
            @StatusLine("IOPS:", "3.4k")
        }
    </div>
}

templ LogsTab() {
    <div class="flex gap-1">
        @Card("Application") {
            @StatusLine("Level:", "INFO")
            @StatusLine("Rate:", "84/min")
            @StatusLine("Errors:", "2")
        }
        @Card("Security") {
            @StatusLine("Auth:", "OK")
            @StatusLine("Blocked:", "17")
            @StatusLine("Alerts:", "0")
        }
    </div>
}

// Struct component

type dashboard struct {
    selected *tui.State[int]
    tabs     []string
}

func Dashboard() *dashboard {
    return &dashboard{
        selected: tui.NewState(0),
        tabs:     []string{"Overview", "Metrics", "Logs"},
    }
}

func (d *dashboard) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) {
            d.selected.Update(func(v int) int {
                return (v + 1) % len(d.tabs)
            })
        }),
    }
}

templ (d *dashboard) Render() {
    <div class="flex-col p-1 gap-1 border-rounded border-cyan">
        <div class="flex gap-2">
            @for i, tab := range d.tabs {
                @if i == d.selected.Get() {
                    @Badge(tab, "text-cyan")
                } @else {
                    <span class="font-dim">{tab}</span>
                }
            }
        </div>

        @if d.selected.Get() == 0 {
            @OverviewTab()
        } @else @if d.selected.Get() == 1 {
            @MetricsTab()
        } @else {
            @LogsTab()
        }

        <div class="flex justify-center">
            <span class="font-dim">{fmt.Sprintf("Tab: switch tabs | esc: quit | Viewing: %s", d.tabs[d.selected.Get()])}</span>
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
        tui.WithRootComponent(Dashboard()),
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

## Next Steps

- [Event Handling](events) -- Keyboard and mouse input in depth
- [Timers, Watchers, and Channels](watchers) -- Background operations and the event bus
- [State and Reactivity](state) -- Reactive state management with `State[T]`
