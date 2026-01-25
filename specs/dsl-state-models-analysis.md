# DSL State Management Models: Analysis

This document analyzes four potential state management approaches for the go-tui DSL, with detailed API examples and generated Go code.

---

## 1. Signals/Reactive Model

### Concept

Reactive primitives where state changes automatically trigger re-renders. Inspired by SolidJS, Svelte, and Observable patterns.

### Core Types

```go
// pkg/reactive/signal.go

// Signal holds a value and notifies dependents when it changes.
type Signal[T any] struct {
    value    T
    watchers []func(T)
}

func NewSignal[T any](initial T) *Signal[T]
func (s *Signal[T]) Get() T
func (s *Signal[T]) Set(value T)          // Triggers watchers
func (s *Signal[T]) Update(fn func(T) T)  // Atomic update

// Derived creates a computed signal that depends on other signals.
func Derived[T, R any](source *Signal[T], transform func(T) R) *Signal[R]

// Effect runs a side effect whenever dependencies change.
func Effect(fn func())
```

### DSL Syntax

```
// components/counter.tui

@component Counter() {
    @state count = 0                    // Creates Signal[int]
    @derived doubled = count * 2        // Creates derived Signal

    <box direction="column" gap={1}>
        <text>Count: {count}</text>      // Auto-subscribes to signal
        <text>Doubled: {doubled}</text>
        <button onPress={() => count++}>Increment</button>
    </box>
}

@component TodoList() {
    @state items = []string{}
    @state filter = "all"

    @derived visible = items.filter(item => {
        if filter == "all" { return true }
        return item.done == (filter == "done")
    })

    <box direction="column">
        @for item in visible {
            <TodoItem item={item} />
        }
    </box>
}
```

### Generated Go Code

```go
// counter_tui.go (generated)

package components

import (
    "github.com/grindlemire/go-tui/pkg/reactive"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

func Counter() *element.Element {
    // Create signals
    count := reactive.NewSignal(0)
    doubled := reactive.Derived(count, func(c int) int { return c * 2 })

    // Build element tree
    root := element.New(
        element.WithDirection(layout.Column),
        element.WithGap(1),
    )

    countText := element.New(element.WithText("Count: 0"))
    doubledText := element.New(element.WithText("Doubled: 0"))

    // Subscribe to signals
    reactive.Effect(func() {
        countText.SetText(fmt.Sprintf("Count: %d", count.Get()))
    })
    reactive.Effect(func() {
        doubledText.SetText(fmt.Sprintf("Doubled: %d", doubled.Get()))
    })

    button := element.New(
        element.WithText("Increment"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                count.Update(func(c int) int { return c + 1 })
                return true
            }
            return false
        }),
    )

    root.AddChild(countText, doubledText, button)
    return root
}
```

### Characteristics

| Aspect | Assessment |
|--------|------------|
| **Automatic updates** | UI updates automatically when signals change |
| **Fine-grained** | Only affected elements re-render |
| **Memory overhead** | Signals + watcher lists per reactive value |
| **Complexity** | Moderate — subscription management needed |
| **Debugging** | Can be tricky — implicit dependencies |
| **Go idiom fit** | Fair — generics help, but not a native Go pattern |
| **Performance** | Good for complex UIs with many updates |

### Pros
- Declarative: describe what, not how
- Automatic dependency tracking
- Fine-grained updates (no full tree reconciliation)
- Composable with derived signals

### Cons
- Runtime overhead for signal infrastructure
- Memory for subscription lists
- Debugging can be harder (magic updates)
- Less idiomatic in Go

---

## 2. Props-Down/Events-Up Model (React-like)

### Concept

Components are pure functions of props. State lives at the top and flows down. Events bubble up to trigger state changes and re-renders.

### Core Types

```go
// No runtime types needed — components are pure functions
// Props are regular Go structs
// Events use callbacks

type CounterProps struct {
    Count    int
    OnChange func(int)
}
```

### DSL Syntax

```
// components/counter.tui

@component Counter(count int, onChange func(int)) {
    <box direction="column" gap={1}>
        <text>Count: {count}</text>
        <text>Doubled: {count * 2}</text>
        <button onPress={() => onChange(count + 1)}>Increment</button>
    </box>
}

@component TodoList(items []TodoItem, onUpdate func([]TodoItem)) {
    <box direction="column">
        @for i, item in items {
            <TodoItem
                item={item}
                onToggle={() => onUpdate(toggleItem(items, i))}
            />
        }
    </box>
}

// Parent manages state
@component App() {
    @state count = 0  // State at top level only

    <box>
        <Counter count={count} onChange={(c) => count = c} />
    </box>
}
```

### Generated Go Code

```go
// counter_tui.go (generated)

package components

// Counter is a pure function that builds an element tree from props.
func Counter(count int, onChange func(int)) *element.Element {
    root := element.New(
        element.WithDirection(layout.Column),
        element.WithGap(1),
    )

    countText := element.New(element.WithText(fmt.Sprintf("Count: %d", count)))
    doubledText := element.New(element.WithText(fmt.Sprintf("Doubled: %d", count*2)))

    button := element.New(
        element.WithText("Increment"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                onChange(count + 1)
                return true
            }
            return false
        }),
    )

    root.AddChild(countText, doubledText, button)
    return root
}

// App manages state and rebuilds on change
type App struct {
    root  *element.Element
    count int
}

func NewApp() *App {
    app := &App{count: 0}
    app.rebuild()
    return app
}

func (a *App) rebuild() {
    a.root = element.New()
    a.root.AddChild(Counter(a.count, func(c int) {
        a.count = c
        a.rebuild()  // Re-render entire tree
    }))
}

func (a *App) Root() *element.Element { return a.root }
```

### Characteristics

| Aspect | Assessment |
|--------|------------|
| **Predictability** | Excellent — unidirectional data flow |
| **Testability** | Excellent — components are pure functions |
| **Memory overhead** | Low — no runtime infrastructure |
| **Complexity** | Low — simple mental model |
| **Debugging** | Easy — data flows in one direction |
| **Go idiom fit** | Good — functions, structs, callbacks |
| **Performance** | Rebuild entire tree on state change (but layout is dirty-checked) |

### Pros
- Simple mental model
- Easy to test (pure functions)
- Predictable data flow
- No runtime infrastructure
- Good fit for Go's function-oriented style

### Cons
- Rebuilds entire component tree on state change
- Prop drilling (passing callbacks through layers)
- State lives at top, can get bloated
- No fine-grained updates

---

## 3. Direct Mutation Model (Current Element API)

### Concept

Components hold references to elements. State changes are applied directly with explicit `SetText()`, `SetStyle()`, etc. Manual `MarkDirty()` if needed. No magic.

### Core Types

```go
// Uses existing Element API directly
// No additional runtime types needed
// Components are factories that return elements
```

### DSL Syntax

```
// components/counter.tui

@component Counter() {
    // Elements are created and referenced
    @let countText = <text>Count: 0</text>
    @let count = 0

    <box direction="column" gap={1}>
        {countText}
        <text>Doubled: {count * 2}</text>
        <button onPress={() => {
            count++
            countText.SetText("Count: " + count)
        }}>Increment</button>
    </box>
}

@component TodoList() {
    @let container = <box direction="column" />
    @let items = []string{}

    // Manual list management
    func addItem(text string) {
        items = append(items, text)
        item := <TodoItem text={text} />
        container.AddChild(item)
    }

    <box>
        {container}
        <button onPress={() => addItem("New Item")}>Add</button>
    </box>
}
```

### Generated Go Code

```go
// counter_tui.go (generated)

package components

// Counter returns an interactive counter element.
// The returned struct provides access to the element and control methods.
type Counter struct {
    Root      *element.Element
    countText *element.Element
    count     int
}

func NewCounter() *Counter {
    c := &Counter{count: 0}

    c.countText = element.New(element.WithText("Count: 0"))

    doubledText := element.New(element.WithText("Doubled: 0"))

    button := element.New(
        element.WithText("Increment"),
        element.WithOnEvent(func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                c.count++
                c.countText.SetText(fmt.Sprintf("Count: %d", c.count))
                doubledText.SetText(fmt.Sprintf("Doubled: %d", c.count*2))
                return true
            }
            return false
        }),
    )

    c.Root = element.New(
        element.WithDirection(layout.Column),
        element.WithGap(1),
    )
    c.Root.AddChild(c.countText, doubledText, button)

    return c
}

// Element returns the root element for adding to trees.
func (c *Counter) Element() *element.Element { return c.Root }
```

### Characteristics

| Aspect | Assessment |
|--------|------------|
| **Explicitness** | Excellent — no magic, every update is visible |
| **Performance** | Excellent — only changed elements are updated |
| **Memory overhead** | Minimal — just element pointers |
| **Complexity** | Low for simple, high for complex (manual updates) |
| **Debugging** | Easy — explicit updates, no hidden behavior |
| **Go idiom fit** | Excellent — matches current API perfectly |
| **Scalability** | Challenging — manual updates for every state change |

### Pros
- Zero runtime overhead
- Matches existing Element API exactly
- No magic — every update is explicit
- Familiar to Go developers
- Fine-grained control

### Cons
- Verbose for complex state
- Easy to forget to update an element
- Manual synchronization between state and UI
- Doesn't scale well for complex components

---

## 4. Hybrid Approach

### Concept

Support multiple patterns. Components can choose their own state strategy:
- Simple components: Direct mutation
- Complex state: Signals/reactive
- Reusable components: Props-down

Provide primitives, let users compose.

### Core Types

```go
// pkg/reactive/state.go

// State wraps a value with optional automatic binding.
type State[T any] struct {
    value    T
    bindings []*element.Element  // Elements to update
    updater  func(*element.Element, T)
}

func NewState[T any](initial T) *State[T]
func (s *State[T]) Get() T
func (s *State[T]) Set(value T)
func (s *State[T]) Bind(elem *element.Element, updater func(*element.Element, T))

// Convenience for common bindings
func BindText[T any](elem *element.Element, format string) func(*element.Element, T)
```

### DSL Syntax

```
// components/counter.tui

// Option 1: Direct mutation (simple)
@component SimpleCounter() {
    @let count = 0
    @let countText = <text />

    <box>
        {countText}
        <button onPress={() => {
            count++
            countText.SetText("Count: " + count)
        }}>+</button>
    </box>
}

// Option 2: State binding (medium complexity)
@component BoundCounter() {
    @state count = 0  // Creates State[int]

    <box>
        <text bind:text={count, "Count: %d"} />  // Auto-update on change
        <button onPress={() => count++}>+</button>
    </box>
}

// Option 3: Props-down (reusable)
@component ReusableCounter(count int, onIncrement func()) {
    <box>
        <text>Count: {count}</text>
        <button onPress={onIncrement}>+</button>
    </box>
}

// Option 4: Full reactive (complex)
@component ReactiveCounter() {
    @signal count = 0
    @derived doubled = count * 2

    <box>
        <text>{count}</text>
        <text>{doubled}</text>
    </box>
}
```

### Generated Go Code

```go
// Option 1: Direct mutation
type SimpleCounter struct {
    Root      *element.Element
    countText *element.Element
    count     int
}

func NewSimpleCounter() *SimpleCounter { /* ... */ }

// Option 2: State binding
func BoundCounter() *element.Element {
    count := reactive.NewState(0)

    countText := element.New()
    count.Bind(countText, reactive.BindText[int]("Count: %d"))

    button := element.New(
        element.WithText("+"),
        element.WithOnEvent(func(e tui.Event) bool {
            count.Set(count.Get() + 1)  // Automatically updates countText
            return true
        }),
    )

    root := element.New()
    root.AddChild(countText, button)
    return root
}

// Option 3: Props-down (pure function)
func ReusableCounter(count int, onIncrement func()) *element.Element {
    root := element.New()

    countText := element.New(element.WithText(fmt.Sprintf("Count: %d", count)))
    button := element.New(
        element.WithText("+"),
        element.WithOnEvent(func(e tui.Event) bool {
            onIncrement()
            return true
        }),
    )

    root.AddChild(countText, button)
    return root
}

// Option 4: Full reactive
func ReactiveCounter() *element.Element {
    count := reactive.NewSignal(0)
    doubled := reactive.Derived(count, func(c int) int { return c * 2 })

    // ... with Effect for auto-subscription
}
```

### Characteristics

| Aspect | Assessment |
|--------|------------|
| **Flexibility** | Excellent — choose the right tool |
| **Learning curve** | Higher — multiple patterns to learn |
| **Consistency** | Lower — codebases may mix patterns |
| **Implementation** | Higher — must implement all patterns |
| **Performance** | Varies by pattern chosen |
| **Go idiom fit** | Good — offers the simple path and advanced path |

### Pros
- Use simple patterns for simple components
- Scale up to reactive for complex state
- Reusable components via props
- Users choose based on needs

### Cons
- More to learn
- Inconsistent codebases
- More implementation work
- Decision fatigue

---

## Comparison Matrix

| Criteria | Signals/Reactive | Props-Down | Direct Mutation | Hybrid |
|----------|-----------------|------------|-----------------|--------|
| **Simplicity** | Medium | High | High (simple cases) | Medium |
| **Scalability** | High | Medium | Low | High |
| **Performance** | Fine-grained | Tree rebuild | Fine-grained | Varies |
| **Memory** | Higher | Low | Low | Varies |
| **Testability** | Medium | Excellent | Good | Varies |
| **Go idiom** | Fair | Good | Excellent | Good |
| **Implementation effort** | High | Medium | Low | High |
| **Debugging** | Harder | Easy | Easy | Varies |

---

## Recommendation

Given your priorities (flexibility, maintainability, performance, idiomatic Go), I recommend:

### Primary: **Direct Mutation with Optional Bindings**

1. **Default to direct mutation** — matches existing Element API, zero overhead, explicit
2. **Add simple State[T] bindings** — for common patterns (text updates, style changes)
3. **Props-style components** — for reusable/composable components
4. **No full reactive system initially** — can be added later if needed

This gives you:
- Simple mental model for most cases
- Escape hatch for common patterns
- Reusability through function parameters
- Path to add reactivity later without breaking changes

### Example of Recommended Approach

```
@component Counter() {
    @let count = 0

    // Direct references for updates
    @let countText = <text>Count: 0</text>

    <box direction="column" gap={1}>
        {countText}
        <button onPress={() => {
            count++
            countText.SetText("Count: " + count)
        }}>Increment</button>
    </box>
}

// Reusable via parameters
@component Display(label string, value int) {
    <text>{label}: {value}</text>
}
```

This is the most Go-idiomatic approach while still providing the ergonomics of a DSL.
