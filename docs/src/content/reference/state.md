# State Reference

## Overview

`State[T]` and `Events[T]` are the two generic types that handle reactivity in go-tui. `State[T]` holds a value and triggers re-renders when it changes. `Events[T]` broadcasts messages between components without shared mutable state.

Both types follow the same lifecycle pattern: create them unbound, and the framework binds them to an `App` during component mounting. Once bound, any mutation marks the UI dirty so the next frame picks up the change.

```go
// In a struct component
type counter struct {
    count *tui.State[int]
}

func Counter() *counter {
    return &counter{
        count: tui.NewState(0),
    }
}
```

## State[T]

A reactive container for a single value of type `T`. When the value changes via `Set` or `Update`, the framework marks the app dirty (triggering a re-render) and notifies any registered bindings.

### NewState

```go
func NewState[T any](initial T) *State[T]
```

Creates a new state with the given initial value. The state starts unbound and the framework calls `BindApp` automatically when the owning component mounts.

Works with any type: primitives, slices, structs, maps.

```go
count := tui.NewState(0)               // State[int]
name := tui.NewState("hello")          // State[string]
items := tui.NewState([]string{})      // State[[]string]
visible := tui.NewState(true)          // State[bool]
```

### NewStateForApp

```go
func NewStateForApp[T any](app *App, initial T) *State[T]
```

Creates a state already bound to the given app. Use this when creating state outside of a component (for example, in `main.go` before passing it to a component constructor). Panics if `app` is nil.

```go
app, _ := tui.NewApp(...)
data := tui.NewStateForApp(app, []string{"a", "b", "c"})
```

### BindApp

```go
func (s *State[T]) BindApp(app *App)
```

Binds the state to an app for dirty-marking and batching. The framework calls this during component mounting; you rarely need to call it yourself. Panics if `app` is nil. Idempotent if called with the same app; overwrites if called with a different one.

### Get

```go
func (s *State[T]) Get() T
```

Returns the current value. Safe to call from any goroutine.

```go
current := count.Get()
```

In `.gsx` render methods, call `Get()` inside expressions:

```gsx
<span>{fmt.Sprintf("Count: %d", s.count.Get())}</span>
```

### Set

```go
func (s *State[T]) Set(v T)
```

Replaces the value, marks the app dirty, and notifies all bindings. If called within a `Batch`, binding execution is deferred until the batch completes.

Must be called from the main event loop. For updates from background goroutines, use `app.QueueUpdate` or channel watchers.

```go
count.Set(42)
```

### Update

```go
func (s *State[T]) Update(fn func(T) T)
```

Reads the current value, applies `fn`, and sets the result. A convenience for read-modify-write operations.

```go
count.Update(func(v int) int { return v + 1 })

items.Update(func(list []string) []string {
    return append(list, "new item")
})
```

### Bind

```go
func (s *State[T]) Bind(fn func(T)) Unbind
```

Registers a callback that fires whenever the value changes via `Set` or `Update`. The callback receives the new value. Bindings execute in registration order.

Returns an `Unbind` function. Call it to remove the binding and prevent future invocations.

```go
unbind := count.Bind(func(v int) {
    fmt.Println("count changed to", v)
})

// Later, to stop receiving updates:
unbind()
```

### Unbind

```go
type Unbind func()
```

A handle returned by `Bind`. Calling it deactivates the associated binding. Safe to call multiple times.

## Events[T]

An event bus for broadcasting messages between components. Unlike `State[T]`, `Events[T]` does not store a value. It delivers each emitted event to all current subscribers and marks the app dirty.

### NewEvents

```go
func NewEvents[T any](topic string) *Events[T]
```

Creates a new topic-based event bus. `topic` is required and is used to route events across components.

```go
notifications := tui.NewEvents[string]("app.notifications")
```

### NewEventsForApp

```go
func NewEventsForApp[T any](app *App, topic string) *Events[T]
```

Creates an event bus already bound to the given app. Use this when creating a bus outside of a component. Panics if `app` is nil.

```go
app, _ := tui.NewApp(...)
bus := tui.NewEventsForApp[string](app, "app.notifications")
```

### BindApp

```go
func (e *Events[T]) BindApp(app *App)
```

Binds the event bus to an app for dirty-marking. The framework calls this during component mounting. Panics if `app` is nil. Idempotent for the same app; overwrites if called with a different one.

### UnbindApp

```go
func (e *Events[T]) UnbindApp()
```

Detaches the event bus from app topic routing. Called automatically when components unmount.

### Emit

```go
func (e *Events[T]) Emit(event T)
```

Sends an event to all subscribers and marks the app dirty. Subscribers are called synchronously in registration order.

Panics if the event bus has no bound app.

```go
notifications.Emit("task completed")
```

### Subscribe

```go
func (e *Events[T]) Subscribe(fn func(T)) func()
```

Registers a handler that will be called for every emitted event. Handlers run on the UI thread, so it is safe to update state from within a handler.

```go
unsub := notifications.Subscribe(func(msg string) {
    log.Set(append(log.Get(), msg))
})
defer unsub()
```

## Batching

### App.Batch

```go
func (a *App) Batch(fn func())
```

Executes `fn` with all state binding callbacks deferred. Within a batch, each `Set` or `Update` call still marks the app dirty immediately, but binding callbacks accumulate and run once when the outermost `Batch` returns. If the same binding is triggered multiple times during a batch, only the last value is delivered.

Batches can nest. Deferred callbacks only fire when the outermost batch completes.

```go
app.Batch(func() {
    firstName.Set("Alice")
    lastName.Set("Smith")
    age.Set(30)
}) // All binding callbacks fire here, once
```

Use `Batch` when you need to update several state values at once and want to avoid intermediate binding work.

## Thread Safety

`State[T]` and `Events[T]` follow the same threading rules as the rest of go-tui:

- **`Get()` is safe from any goroutine.** Read the value whenever you need it.
- **`Set()`, `Update()`, and `Emit()` must run on the main event loop.** The event loop is the single thread that processes input and renders frames.
- **Watcher callbacks run on the event loop.** Handlers passed to `tui.OnTimer` and `tui.Watch` are safe to call `Set` from.
- **For background goroutines, use `app.QueueUpdate`.** This enqueues a function on the event loop where `Set` is safe.

```go
// Background goroutine updating state safely
go func() {
    result := fetchData()
    app.QueueUpdate(func() {
        data.Set(result) // runs on the event loop
    })
}()
```

## Practical Patterns

### Derived display values

Use Go expressions directly in `.gsx` to compute display strings from state:

```gsx
templ (c *counter) Render() {
    <span class="font-bold">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
}
```

### Reusable element fragments with @let

`@let` binds an element to a name so you can reuse it in multiple places. It requires an element (starting with `<`), not a Go expression.

```gsx
templ (s *stateApp) Render() {
    @let countBadge = <span class="text-cyan font-bold">{fmt.Sprintf("%d", s.count.Get())}</span>
    <div class="flex-col gap-1">
        {countBadge}
    </div>
}
```

### Conditional rendering from state

```gsx
templ (a *myApp) Render() {
    @if a.loading.Get() {
        <span class="text-yellow">Loading...</span>
    } @else {
        <span class="text-green">Ready</span>
    }
}
```

### Shared state between components

Create state in a parent and pass it to child constructors. Both components react to changes from either side.

```go
type app struct {
    selected *tui.State[int]
    sidebar  *sidebar
    content  *contentPanel
}

func App() *app {
    sel := tui.NewState(0)
    return &app{
        selected: sel,
        sidebar:  Sidebar(sel),
        content:  ContentPanel(sel),
    }
}
```

### Cross-component events

Use `Events[T]` when components need to communicate without sharing mutable state:

```go
type app struct {
    header *header
    body   *body
}

func App() *app {
    return &app{
        header: Header(),
        body:   Body(),
    }
}
```

Both components can construct `tui.NewEvents[string]("app.alerts")` internally and communicate through that shared topic without passing bus pointers.
