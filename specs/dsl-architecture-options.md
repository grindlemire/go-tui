# DSL Architecture Options: Detailed Comparison

This document explores three architecture approaches for the go-tui DSL, with focus on call flow and state presentation.

---

## Option A: Bubbletea-Style (TEA / Elm Architecture)

### Core Concept

Separate Model (state) from View (UI). The Update function handles all state changes and returns a new model. View is a pure function of model state.

### Core Types

```go
// pkg/tea/model.go

// Msg is a message that triggers an update.
type Msg interface{}

// Cmd is a command that produces messages (for async operations).
type Cmd func() Msg

// Model is the application state with TEA methods.
type Model interface {
    // Init returns the initial command to run.
    Init() Cmd

    // Update handles a message and returns the new model + command.
    // MUST return a new model, not mutate in place.
    Update(msg Msg) (Model, Cmd)

    // View builds the element tree from current state.
    View() *element.Element
}

// Program runs a TEA program.
type Program struct {
    model   Model
    app     *tui.App
}

func NewProgram(model Model) *Program
func (p *Program) Run() error
```

### DSL Syntax

```
// app.tui

@model Counter {
    count int
}

@init {
    return nil  // No initial command
}

@update(msg) {
    switch msg {
    case IncrementMsg:
        return Counter{count: model.count + 1}, nil
    case DecrementMsg:
        return Counter{count: model.count - 1}, nil
    }
    return model, nil
}

@view {
    <box direction="column" gap={1}>
        <text>Count: {model.count}</text>
        <box direction="row" gap={2}>
            <button onPress={IncrementMsg}>+</button>
            <button onPress={DecrementMsg}>-</button>
        </box>
    </box>
}

// Message types
@msg IncrementMsg
@msg DecrementMsg
```

### Generated Go Code

```go
// app_tui.go (generated)

package main

import (
    "github.com/grindlemire/go-tui/pkg/tea"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

// Message types
type IncrementMsg struct{}
type DecrementMsg struct{}

// Model
type Counter struct {
    count int
}

func (m Counter) Init() tea.Cmd {
    return nil
}

func (m Counter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case IncrementMsg:
        return Counter{count: m.count + 1}, nil
    case DecrementMsg:
        return Counter{count: m.count - 1}, nil
    }
    return m, nil
}

func (m Counter) View() *element.Element {
    root := element.New(
        element.WithDirection(layout.Column),
        element.WithGap(1),
    )

    countText := element.New(
        element.WithText(fmt.Sprintf("Count: %d", m.count)),
    )

    buttonRow := element.New(
        element.WithDirection(layout.Row),
        element.WithGap(2),
    )

    incButton := element.New(
        element.WithText("+"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                tea.Send(IncrementMsg{})  // Send message to Update
                return true
            }
            return false
        }),
    )

    decButton := element.New(
        element.WithText("-"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                tea.Send(DecrementMsg{})
                return true
            }
            return false
        }),
    )

    buttonRow.AddChild(incButton, decButton)
    root.AddChild(countText, buttonRow)
    return root
}
```

### Call Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  1. Program starts                                               │
│     model := Counter{count: 0}                                   │
│     cmd := model.Init()                                          │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. Initial render                                               │
│     tree := model.View()   // Build element tree from model      │
│     app.SetRoot(tree)                                            │
│     app.Render()                                                 │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. Event loop                                                   │
│     event, ok := app.PollEvent(timeout)                          │
│     if ok {                                                      │
│         // Convert event to Msg or dispatch to focused element   │
│         msg := eventToMsg(event)                                 │
│         if msg != nil {                                          │
│             goto step 4                                          │
│         }                                                        │
│     }                                                            │
│     // Run any pending commands                                  │
│     goto step 2 (re-render)                                      │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  4. Update                                                       │
│     newModel, cmd := model.Update(msg)                           │
│     model = newModel                                             │
│     // Queue cmd for execution                                   │
│     goto step 2 (rebuild view)                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Key Insight

**View rebuilds the ENTIRE element tree on every state change.**

This is the fundamental trade-off of TEA:
- ✅ View is always consistent with model
- ✅ No stale state bugs
- ❌ Full tree rebuild on every update
- ❌ Need diffing or careful optimization

With your layout engine's dirty flags, the layout won't recompute unnecessarily, but:
- Element allocations happen on every render
- Child slices are reallocated
- Callbacks are recreated

### State Presentation

State is presented as **fields on the model struct**, accessed in View:

```
@view {
    <text>Count: {model.count}</text>   // model.count is the state
}
```

---

## Option B: Ratatui-Style (Stateful Widgets + App State)

### Core Concept

Application owns state directly. Widgets can be stateless (render-only) or stateful (have associated state). App controls the render loop and mutates state directly.

### Core Types

```go
// No framework types needed!
// App is just a regular struct
// Widgets are regular elements
// State is regular Go fields
```

### DSL Syntax

```
// counter.tui

// Stateless component - just a view function
@component CounterView(count int, onIncrement func(), onDecrement func()) {
    <box direction="column" gap={1}>
        <text>Count: {count}</text>
        <box direction="row" gap={2}>
            <button onPress={onIncrement}>+</button>
            <button onPress={onDecrement}>-</button>
        </box>
    </box>
}

// app.tui

// App struct with state - you write this in Go
@app CounterApp {
    count int
}

// Render method uses the DSL
@render {
    <box padding={2}>
        <CounterView
            count={app.count}
            onIncrement={() => app.count++}
            onDecrement={() => app.count--}
        />
    </box>
}
```

### Generated Go Code

```go
// counter_tui.go (generated)

package main

import "github.com/grindlemire/go-tui/pkg/tui/element"

// CounterView is a stateless component - pure function of props.
func CounterView(count int, onIncrement func(), onDecrement func()) *element.Element {
    root := element.New(
        element.WithDirection(layout.Column),
        element.WithGap(1),
    )

    countText := element.New(
        element.WithText(fmt.Sprintf("Count: %d", count)),
    )

    buttonRow := element.New(
        element.WithDirection(layout.Row),
        element.WithGap(2),
    )

    incButton := element.New(
        element.WithText("+"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                onIncrement()
                return true
            }
            return false
        }),
    )

    decButton := element.New(
        element.WithText("-"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                onDecrement()
                return true
            }
            return false
        }),
    )

    buttonRow.AddChild(incButton, decButton)
    root.AddChild(countText, buttonRow)
    return root
}
```

```go
// app_tui.go (generated)

package main

// CounterApp manages state and render loop.
type CounterApp struct {
    app   *tui.App
    count int
    root  *element.Element
}

func NewCounterApp() (*CounterApp, error) {
    tuiApp, err := tui.NewApp()
    if err != nil {
        return nil, err
    }
    ca := &CounterApp{app: tuiApp, count: 0}
    ca.rebuild()
    return ca, nil
}

func (ca *CounterApp) rebuild() {
    ca.root = element.New(element.WithPadding(2))
    ca.root.AddChild(CounterView(
        ca.count,
        func() { ca.count++; ca.rebuild() },
        func() { ca.count--; ca.rebuild() },
    ))
    ca.app.SetRoot(ca.root)
}

func (ca *CounterApp) Run() error {
    for {
        event, ok := ca.app.PollEvent(50 * time.Millisecond)
        if ok {
            if ke, ok := event.(tui.KeyEvent); ok && ke.Key == tui.KeyEscape {
                return nil
            }
            ca.app.Dispatch(event)
        }
        ca.app.Render()
    }
}
```

### Call Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  1. App starts                                                   │
│     app := NewCounterApp()                                       │
│     app.rebuild()  // Build initial tree                         │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. Event loop                                                   │
│     event, ok := app.PollEvent(timeout)                          │
│     if ok {                                                      │
│         app.Dispatch(event)  // Routes to focused element        │
│     }                                                            │
│     app.Render()                                                 │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. Event handler (in button)                                    │
│     onIncrement()  // Closure captures app                       │
│       → app.count++                                              │
│       → app.rebuild()  // Rebuild tree with new state            │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  4. Rebuild                                                      │
│     // Create new element tree from current state                │
│     // Old tree is garbage collected                             │
│     app.SetRoot(newTree)                                         │
└─────────────────────────────────────────────────────────────────┘
```

### Key Insight

**Components are pure functions. App owns state and rebuilds when needed.**

This is the React props-down model but without React:
- Components receive data, return elements
- App decides when to rebuild
- State mutation is explicit and user-controlled

### State Presentation

State is presented as **fields on the app struct**, passed as props:

```
@render {
    <CounterView count={app.count} ... />   // app.count is the state
}
```

---

## Option C: Templ-Style ("Just Go" with Templates)

### Core Concept

The DSL is **syntactic sugar for constructing element trees**. No framework, no state management, no runtime. Components compile to Go functions. You manage state however you want.

This is the most flexible approach — the DSL only handles the **view layer**.

### Core Types

```go
// No framework types!
// The DSL compiles to element construction code
// State management is entirely your choice
```

### DSL Syntax

```
// components.tui

package components

import (
    "fmt"
    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
    "github.com/grindlemire/go-tui/pkg/layout"
)

// Components are Go functions with template syntax for the return
@component CounterView(count int, onInc, onDec func()) *element.Element {
    return <box direction={layout.Column} gap={1}>
        <text style={tui.NewStyle().Bold()}>Count: {fmt.Sprintf("%d", count)}</text>
        <box direction={layout.Row} gap={2}>
            <text onPress={onInc}>+</text>
            <text onPress={onDec}>-</text>
        </box>
    </box>
}

// Can embed Go code freely
@component ConditionalView(show bool) *element.Element {
    if !show {
        return <text>Hidden</text>
    }

    items := []string{"one", "two", "three"}

    return <box direction={layout.Column}>
        @for i, item := range items {
            <text key={i}>{item}</text>
        }
    </box>
}

// Can use closures and local state
@component StatefulCounter() *element.Element {
    count := 0  // Local variable

    // Elements created once, mutated via closures
    @let countText = <text>Count: 0</text>

    return <box direction={layout.Column}>
        {countText}
        <text onPress={func() {
            count++
            countText.SetText(fmt.Sprintf("Count: %d", count))
        }}>+</text>
    </box>
}
```

### Generated Go Code

```go
// components_tui.go (generated)

package components

import (
    "fmt"
    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
    "github.com/grindlemire/go-tui/pkg/layout"
)

// CounterView builds a counter display.
func CounterView(count int, onInc, onDec func()) *element.Element {
    __tui_0 := element.New(
        element.WithDirection(layout.Column),
        element.WithGap(1),
    )
    __tui_1 := element.New(
        element.WithText(fmt.Sprintf("Count: %d", count)),
        element.WithTextStyle(tui.NewStyle().Bold()),
    )
    __tui_2 := element.New(
        element.WithDirection(layout.Row),
        element.WithGap(2),
    )
    __tui_3 := element.New(
        element.WithText("+"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                onInc()
                return true
            }
            return false
        }),
    )
    __tui_4 := element.New(
        element.WithText("-"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                onDec()
                return true
            }
            return false
        }),
    )
    __tui_2.AddChild(__tui_3, __tui_4)
    __tui_0.AddChild(__tui_1, __tui_2)
    return __tui_0
}

// ConditionalView demonstrates conditional rendering.
func ConditionalView(show bool) *element.Element {
    if !show {
        return element.New(element.WithText("Hidden"))
    }

    items := []string{"one", "two", "three"}

    __tui_0 := element.New(element.WithDirection(layout.Column))
    for i, item := range items {
        _ = i  // key is informational
        __tui_child := element.New(element.WithText(item))
        __tui_0.AddChild(__tui_child)
    }
    return __tui_0
}

// StatefulCounter demonstrates local state via closures.
func StatefulCounter() *element.Element {
    count := 0

    countText := element.New(element.WithText("Count: 0"))

    __tui_0 := element.New(element.WithDirection(layout.Column))
    __tui_1 := element.New(
        element.WithText("+"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                count++
                countText.SetText(fmt.Sprintf("Count: %d", count))
                return true
            }
            return false
        }),
    )
    __tui_0.AddChild(countText, __tui_1)
    return __tui_0
}
```

### Call Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  1. Code Generation (build time)                                 │
│     tui generate ./...                                           │
│     components.tui → components_tui.go                           │
│     (Pure Go code, no runtime)                                   │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. Your application (you write this in Go)                      │
│                                                                  │
│     func main() {                                                │
│         app, _ := tui.NewApp()                                   │
│         defer app.Close()                                        │
│                                                                  │
│         // Option A: Stateless, rebuild on change                │
│         count := 0                                               │
│         var root *element.Element                                │
│         rebuild := func() {                                      │
│             root = CounterView(count, ...)                       │
│             app.SetRoot(root)                                    │
│         }                                                        │
│         rebuild()                                                │
│                                                                  │
│         // Option B: Stateful component                          │
│         root := StatefulCounter()                                │
│         app.SetRoot(root)                                        │
│                                                                  │
│         // Event loop (you control this)                         │
│         for {                                                    │
│             event, ok := app.PollEvent(50*time.Millisecond)      │
│             if ok { app.Dispatch(event) }                        │
│             app.Render()                                         │
│         }                                                        │
│     }                                                            │
└─────────────────────────────────────────────────────────────────┘
```

### Key Insight

**The DSL is ONLY for element construction. Everything else is your Go code.**

Like templ:
- `.tui` files compile to `.go` files
- At runtime, it's just Go function calls
- Zero framework overhead
- Use any state management you want (or none)

### State Presentation

State is presented however you want:

```
// As function parameters (stateless)
@component View(count int) { ... }

// As local variables (closure-based)
@component View() {
    count := 0
    @let text = <text />
    return <box><text onPress={() => { count++; text.SetText(...) }}>+</text></box>
}

// As struct fields (you define the struct in Go)
// Then call the component from your struct's methods
```

---

## Comparison: What Gets Generated

| Feature | Bubbletea-style | Ratatui-style | Templ-style |
|---------|-----------------|---------------|-------------|
| State location | Model struct | App struct | Your choice |
| State access | `model.field` | `app.field` | Parameters or closures |
| Update trigger | Return new model | Call rebuild() | Call SetText() etc. |
| Tree rebuild | Every Update | On rebuild() | Never (mutate in place) or your choice |
| Framework code | tea.Program | None | None |
| Generated code | Model methods | Component functions | Component functions |

---

## Recommendation for go-tui

Based on your priorities (flexibility, maintainability, performance, idiomatic Go), I recommend **Templ-style**:

### Why Templ-style?

1. **Maximum flexibility**: Users choose their own state model
2. **Zero runtime**: Just compiled Go functions
3. **Matches your API**: Direct mapping to Element construction
4. **Familiar to Go devs**: It's literally just Go with nicer syntax
5. **Incremental adoption**: Can use DSL for some components, pure Go for others
6. **No lock-in**: The generated code is readable, editable, no magic

### What the DSL provides:

1. **Element construction syntax**: `<box>` instead of `element.New(...)`
2. **Attribute mapping**: `direction={layout.Column}` instead of `element.WithDirection(...)`
3. **Child nesting**: Natural tree structure instead of `AddChild` calls
4. **Control flow**: `@for`, `@if` for dynamic content
5. **Element references**: `@let` for capturing elements to mutate later
6. **Go interop**: Full Go expressions, imports, control flow

### What the DSL does NOT provide (intentionally):

1. **State management**: You choose
2. **Reactivity**: You choose
3. **Event loop**: You write it
4. **Application structure**: You design it

This is the most Go-idiomatic approach: simple tools that compose, no magic, user controls the architecture.
