# Component Interfaces Reference

## Overview

go-tui uses interfaces to define component behavior. A struct component only needs to implement `Component` (the `Render` method). All other interfaces are optional -- implement them to add keyboard handling, mouse support, timers, or lifecycle hooks.

The framework discovers these interfaces through type assertions during the mount process. The order of operations when a component first mounts:

1. `AppBinder.BindApp` -- wire up `State` and `Events` fields
2. `Initializer.Init` -- run setup logic, capture cleanup function
3. `Component.Render` -- produce the element tree
4. `WatcherProvider.Watchers` -- discover and start background watchers
5. `KeyListener.KeyMap` and `MouseListener.HandleMouse` -- discovered during event dispatch via tree walks

On subsequent renders (cached mount), the framework calls `PropsUpdater.UpdateProps` (if implemented), re-calls `AppBinder.BindApp`, then `Render`.

## Component

```go
type Component interface {
    Render(app *App) *Element
}
```

The only required interface. Every struct component must have a `Render` method that returns an element tree. In `.gsx` files, the `templ (s *myStruct) Render()` syntax generates this method.

`Render` is called on every frame where the UI is dirty. It should be a pure function of the component's state -- read state, build elements, return. Side effects belong in `Init`, `Watchers`, or event handlers.

**When to implement:** Always. This is what makes a struct a component.

```gsx
type counter struct {
    count *tui.State[int]
}

func Counter() *counter {
    return &counter{count: tui.NewState(0)}
}

templ (c *counter) Render() {
    <div class="p-1">
        <span class="text-cyan font-bold">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
    </div>
}
```

## KeyListener

```go
type KeyListener interface {
    KeyMap() KeyMap
}
```

Provides keyboard bindings for the component. `KeyMap()` is called during event dispatch when the framework walks the component tree, so it can return different bindings depending on the component's current state.

`KeyMap` is a `[]KeyBinding`. Bindings are checked in order; the first match wins. Use `OnKey`, `OnRune`, `OnKeyStop`, `OnRuneStop`, and `OnRunes` helpers to build bindings.

**When to implement:** When your component needs to respond to keyboard input.

```go
func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('+', func(ke tui.KeyEvent) {
            c.count.Update(func(v int) int { return v + 1 })
        }),
        tui.OnRune('-', func(ke tui.KeyEvent) {
            c.count.Update(func(v int) int { return v - 1 })
        }),
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) {
            ke.App().Stop()
        }),
    }
}
```

## MouseListener

```go
type MouseListener interface {
    HandleMouse(MouseEvent) bool
}
```

Handles mouse events (clicks, wheel, drag). Return `true` if the event was consumed, `false` to let it propagate. Like `KeyListener`, discovered by walking the component tree.

Mouse events are only delivered when mouse support is enabled via `tui.WithMouse()` (enabled by default in full-screen mode, disabled by default in inline mode).

**When to implement:** When your component has clickable elements or needs mouse wheel handling.

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

## WatcherProvider

```go
type WatcherProvider interface {
    Watchers() []Watcher
}
```

Returns background watchers (timers, channel readers) that should run while the component is mounted. The framework calls `Watchers()` after mounting and starts each returned watcher. When the component unmounts or the app stops, the framework closes the stop channel, which signals all watchers to exit.

Watcher handlers run on the main event loop, so they can safely mutate state without synchronization.

**When to implement:** When your component needs periodic updates (timers) or receives data from Go channels.

```go
func (c *dashboard) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, func() {
            c.elapsed.Update(func(v int) int { return v + 1 })
        }),
        tui.Watch(c.dataCh, func(msg string) {
            c.messages.Update(func(list []string) []string {
                return append(list, msg)
            })
        }),
    }
}
```

## Initializer

```go
type Initializer interface {
    Init() func()
}
```

Called once when the component first enters the tree (first `Mount` call). The returned function, if non-nil, is called when the component leaves the tree (unmount cleanup). This pairs setup and teardown at the same call site.

`Init` runs after `BindApp` but before `Render`, so state and events are already wired up.

**When to implement:** When your component needs one-time setup (open a file, start a connection) or cleanup (close resources, cancel goroutines).

```go
func (c *logger) Init() func() {
    f, err := os.Open("app.log")
    if err != nil {
        return nil
    }
    c.file = f

    // Returned function runs on unmount
    return func() {
        f.Close()
    }
}
```

## AppBinder

```go
type AppBinder interface {
    BindApp(app *App)
}
```

Called by the framework to wire up `State` and `Events` fields to the app. Generated code from `.gsx` files emits `BindApp` methods automatically -- each `State` and `Events` field gets its `BindApp` called in turn.

Users never call `BindApp` directly. The mount system calls it before `Init` on first mount, and again after `UpdateProps` on subsequent renders (to bind any fresh `Events` fields from the props update).

**When to implement:** Almost never manually. The `.gsx` code generator handles this. Only implement it if you're building a component without `.gsx` and it contains `State` or `Events` fields.

```go
func (c *counter) BindApp(app *tui.App) {
    c.count.BindApp(app)
}
```

## AppUnbinder

```go
type AppUnbinder interface {
    UnbindApp()
}
```

Called by the framework when a mounted component leaves the tree. This detaches app-bound resources such as topic-based `Events[T]` subscriptions.

You usually don't implement this manually. Generated code handles it for `.gsx` components that contain `Events` fields.

## PropsUpdater

```go
type PropsUpdater interface {
    UpdateProps(fresh Component)
}
```

Called on cached component instances when the parent re-renders and the component is mounted again at the same position. The `fresh` parameter is a new instance created by the factory function, containing the updated props. The cached instance should copy any relevant fields from `fresh`.

Without `PropsUpdater`, a mounted component's props are fixed at the values passed during the first render. Implement this interface when the component's constructor takes parameters that may change across renders.

**When to implement:** When your component receives props from a parent and those props can change.

```go
type statusBar struct {
    label   string
    count   *tui.State[int]
}

func StatusBar(label string) *statusBar {
    return &statusBar{
        label: label,
        count: tui.NewState(0),
    }
}

func (s *statusBar) UpdateProps(fresh tui.Component) {
    if f, ok := fresh.(*statusBar); ok {
        s.label = f.label
        // Don't copy count -- that's internal state, not a prop
    }
}
```

## Renderable

```go
type Renderable interface {
    Render(buf *Buffer, width, height int)
    MarkDirty()
    IsDirty() bool
}
```

The low-level rendering interface. `Element` implements this. `Render` calculates layout (if dirty) and draws to the buffer. `MarkDirty` marks the element and its ancestors as needing layout recalculation. `IsDirty` reports whether recalculation is needed.

`Renderable` is what `App.SetRoot` accepts. Most users pass components (via `WithRootComponent`), which the framework wraps internally. You'd only interact with `Renderable` directly when building custom rendering pipelines or testing.

**When to implement:** Rarely. Use `Component` with `WithRootComponent` instead.

## Viewable

```go
type Viewable interface {
    GetRoot() Renderable
    GetWatchers() []Watcher
}
```

A bundle of a root `Renderable` and its associated watchers. Used by `App.SetRootView` to set the root element and start watchers in one call.

Generated view structs implement this interface. Like `Renderable`, most users don't need to interact with it directly -- `WithRootComponent` handles everything.

**When to implement:** Rarely. Use `Component` with `WithRootComponent` instead.

## Focusable

```go
type Focusable interface {
    IsFocusable() bool
    HandleEvent(event Event) bool
    Focus()
    Blur()
}
```

Implemented by elements that can receive keyboard focus. `Element` satisfies this interface directly via `WithFocusable(true)`, so custom implementations are uncommon.

`IsFocusable` returns whether the element can receive focus. `HandleEvent` processes keyboard/mouse events (returns `true` if consumed). `Focus` and `Blur` are called by the focus manager when focus changes.

**When to implement:** Rarely. Use `WithFocusable`, `WithOnFocus`, and `WithOnBlur` options on elements instead. See [Focus Reference](focus.md) for details.

## Watcher

```go
type Watcher interface {
    Start(eventQueue chan<- func(), stopCh <-chan struct{})
}
```

A background task that sends handler functions to the app's event loop. The framework calls `Start` during component mounting. The goroutine should select on both its data source and `stopCh`, exiting when `stopCh` closes.

Built-in implementations: `OnTimer` (periodic callbacks) and `Watch`/`NewChannelWatcher` (channel readers). See [Watchers Reference](watchers.md) for details.

**When to implement:** When you need a custom background data source beyond timers and channels.

```go
type fileWatcher struct {
    path    string
    handler func([]byte)
}

func (w *fileWatcher) Start(eventQueue chan<- func(), stopCh <-chan struct{}) {
    go func() {
        ticker := time.NewTicker(2 * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-stopCh:
                return
            case <-ticker.C:
                data, err := os.ReadFile(w.path)
                if err != nil {
                    continue
                }
                d := data
                select {
                case eventQueue <- func() { w.handler(d) }:
                case <-stopCh:
                    return
                }
            }
        }
    }()
}
```

## Terminal

```go
type Terminal interface {
    Size() (width, height int)
    Flush(changes []CellChange)
    Clear()
    ClearToEnd()
    SetCursor(x, y int)
    HideCursor()
    ShowCursor()
    EnterRawMode() error
    ExitRawMode() error
    EnterAltScreen()
    ExitAltScreen()
    EnableMouse()
    DisableMouse()
    Caps() Capabilities
    WriteDirect([]byte) (int, error)
}
```

Abstracts terminal operations. The framework ships two implementations: `ANSITerminal` for real terminals and `MockTerminal` for testing.

**When to implement:** You won't need to. Use `ANSITerminal` for production and `MockTerminal` for tests. See [Terminal Reference](terminal.md) for the full API.

## EventReader

```go
type EventReader interface {
    PollEvent(timeout time.Duration) (Event, bool)
    Close() error
}
```

Reads input events from the terminal. `PollEvent` blocks up to `timeout` waiting for an event. Returns `(event, true)` on success, `(nil, false)` on timeout. A timeout of 0 performs a non-blocking check. A negative timeout blocks indefinitely.

The framework ships `NewEventReader` for real stdin input and `MockEventReader` for testing.

**When to implement:** You won't need to. Use `NewEventReader` for production and `MockEventReader` for tests.

## InterruptibleReader

```go
type InterruptibleReader interface {
    EventReader

    EnableInterrupt() error
    Interrupt() error
}
```

Extends `EventReader` with the ability to wake up a blocking `PollEvent` call. Used internally for blocking input mode (`InputLatencyBlocking`) where `PollEvent(-1)` would otherwise block forever. The interrupt mechanism uses a self-pipe: `EnableInterrupt` creates the pipe, and `Interrupt` writes a byte to it, causing `select()` to return.

**When to implement:** This is an internal detail of the blocking input mode. The built-in `stdinReader` and `MockEventReader` already implement it.

## Interface Discovery Order

When the framework mounts a component via `App.Mount`, it checks interfaces in this order:

1. **AppBinder** -- bind state and events to the app
2. **Initializer** -- run setup, capture cleanup
3. **Component** -- call `Render` to get the element tree

On cached re-renders (same mount position):

1. **PropsUpdater** -- update props from a fresh factory instance
2. **AppBinder** -- rebind (fresh Events fields may be unbound after props update)
3. **Component** -- call `Render`

During event dispatch (tree walk):

- **KeyListener** -- checked on each component in the tree for key bindings
- **MouseListener** -- checked on each component in the tree for mouse handling

During root setup:

- **WatcherProvider** -- watchers are discovered and started

## See Also

- [App Reference](app.md) -- app lifecycle, `Mount`, root management
- [Events Reference](events.md) -- `KeyEvent`, `MouseEvent`, `KeyMap` types
- [Focus Reference](focus.md) -- `Focusable` interface and focus management
- [Watchers Reference](watchers.md) -- `Watcher` implementations in detail
- [Terminal Reference](terminal.md) -- `Terminal` interface and `ANSITerminal`
- [Testing Reference](testing.md) -- `MockTerminal` and `MockEventReader`
