# Reactive Bindings Specification

**Status:** Draft\
**Version:** 1.0\
**Last Updated:** 2026-01-27

---

## 1. Overview

### Purpose

Provide a reactive state management system that automatically updates UI elements when state changes. This eliminates manual `SetText()` calls and reduces the need for element refs when displaying dynamic values.

### Goals

- **Reactive state**: `State[T]` wrapper type with `Get()`/`Set()` methods
- **Automatic bindings**: Generator detects state usage in element expressions and wires up update bindings
- **Explicit deps attribute**: `deps={[state1, state2]}` for complex cases where auto-detection fails
- **Type-safe**: Full Go generics support for any state type
- **Minimal boilerplate**: State declaration is a single line (`tui.NewState(initial)`)
- **No Context parameter**: Components don't need Context - framework handles internals
- **Automatic dirty tracking**: `State.Set()` calls `tui.MarkDirty()` - no bool returns needed
- **Batched updates**: `tui.Batch()` coalesces multiple `Set()` calls
- **Unbind support**: `Bind()` returns handle for cleanup
- **Thread safety**: Clear rules for main-loop-only mutations, leveraging existing `atomic.Bool` dirty flag

### Non-Goals

- Full virtual DOM / diffing (structural reactivity for loops is future work)
- Reactive primitives beyond `State[T]` (computed, effects, etc.)
- Global state store (this is component-local state)

### Dependencies

This specification depends on the following existing implementations:

| Dependency | Location | Description |
|------------|----------|-------------|
| `tui.MarkDirty()` | `pkg/tui/dirty.go` | Thread-safe atomic dirty flag |
| `element.Element` | `pkg/tui/element/element.go` | UI element with `SetText()`, etc. |
| `tui.Watcher` | `pkg/tui/watcher.go` | Channel/timer event integration |
| `App.eventQueue` | `pkg/tui/app.go` | Main loop event queue |

---

## 2. Architecture

### Component Overview

| Component | Change |
|-----------|--------|
| `pkg/tui/state.go` | NEW: `State[T]` type with Bind/Unbind, Batch support |
| `pkg/tuigen/analyzer.go` | Detect `State[T]` variables, track usage, handle `deps` attribute |
| `pkg/tuigen/generator.go` | Generate binding code, support explicit deps |

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  .tui Source                                                    │
│                                                                 │
│  count := tui.NewState(0)                                       │
│  <span>{fmt.Sprintf("Count: %d", count.Get())}</span>           │
└─────────────────────────┬───────────────────────────────────────┘
                          │ Analyzer detects State[T] usage
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│  Analysis Results                                               │
│                                                                 │
│  StateVars: [{Name: "count", Type: "int"}]                      │
│  Bindings: [{State: "count", Element: span, Expr: "fmt..."}]    │
└─────────────────────────┬───────────────────────────────────────┘
                          │ Generator creates binding code
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│  Generated Go                                                   │
│                                                                 │
│  count := tui.NewState(0)                                       │
│  span := element.New(element.WithText(                          │
│      fmt.Sprintf("Count: %d", count.Get()),                     │
│  ))                                                             │
│  count.Bind(func(v int) {                                       │
│      span.SetText(fmt.Sprintf("Count: %d", v))                  │
│  })                                                             │
│  // Set() automatically calls tui.MarkDirty()                   │
└─────────────────────────────────────────────────────────────────┘
```

### Integration with Existing Systems

```
┌─────────────────────────────────────────────────────────────────┐
│  User Code / Handler                                            │
│  count.Set(count.Get() + 1)                                     │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│  State[T].Set()                                                 │
│  1. Update value (mutex protected)                              │
│  2. Call tui.MarkDirty() (existing atomic flag)                 │
│  3. Execute bindings (or queue if batching)                     │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│  App Main Loop (existing in app.go)                             │
│  - Checks dirty.Swap(false)                                     │
│  - If dirty, calls Render()                                     │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Core Entities

### 3.1 State[T] Type

```go
// pkg/tui/state.go

package tui

import "sync"

// State wraps a value and notifies bindings when it changes.
type State[T any] struct {
    mu       sync.RWMutex
    value    T
    bindings []*binding[T]
    nextID   uint64
}

type binding[T any] struct {
    id     uint64
    fn     func(T)
    active bool
}

// Unbind is a handle to remove a binding.
type Unbind func()

// NewState creates a new state with the given initial value.
// No Context needed - just pass the initial value.
func NewState[T any](initial T) *State[T] {
    return &State[T]{value: initial}
}

// Get returns the current value. Thread-safe for reading.
func (s *State[T]) Get() T {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.value
}

// Set updates the value, marks dirty, and notifies all bindings.
// IMPORTANT: Must be called from main loop only. For background
// updates, use app.QueueUpdate() or channel watchers.
func (s *State[T]) Set(v T) {
    s.mu.Lock()
    s.value = v
    // Copy active bindings while holding lock
    bindings := make([]*binding[T], 0, len(s.bindings))
    for _, b := range s.bindings {
        if b.active {
            bindings = append(bindings, b)
        }
    }
    s.mu.Unlock()

    // Mark dirty using existing atomic flag
    MarkDirty()

    // Execute bindings outside lock (they may call Get())
    if batchCtx.depth == 0 {
        // Immediate execution
        for _, b := range bindings {
            b.fn(v)
        }
    } else {
        // Deferred execution during batch - track by binding ID
        for _, b := range bindings {
            batchCtx.pending[b.id] = func() { b.fn(v) }
        }
    }
}

// Update applies a function to the current value and sets the result.
func (s *State[T]) Update(fn func(T) T) {
    s.Set(fn(s.Get()))
}

// Bind registers a function to be called when the value changes.
// Returns an Unbind handle to remove the binding.
func (s *State[T]) Bind(fn func(T)) Unbind {
    s.mu.Lock()
    id := s.nextID
    s.nextID++
    b := &binding[T]{id: id, fn: fn, active: true}
    s.bindings = append(s.bindings, b)
    s.mu.Unlock()

    return func() {
        s.mu.Lock()
        b.active = false
        s.mu.Unlock()
    }
}
```

### 3.2 Batching

Batch defers binding execution until after all state updates complete, avoiding redundant updates. The implementation uses binding IDs to properly deduplicate.

```go
// pkg/tui/state.go (continued)

// batchContext tracks batch state
type batchContext struct {
    depth   int
    pending map[uint64]func() // keyed by binding ID for deduplication
}

var batchCtx = batchContext{
    pending: make(map[uint64]func()),
}

// Batch executes fn and defers all binding callbacks until fn returns.
// Use this when updating multiple states to avoid redundant element updates.
//
// When the same binding is triggered multiple times during a batch,
// it only executes once with the final value.
//
// Example:
//
//	tui.Batch(func() {
//	    firstName.Set("Bob")
//	    lastName.Set("Smith")
//	    age.Set(30)
//	})
//	// Bindings fire once here, not three times
func Batch(fn func()) {
    batchCtx.depth++
    fn()
    batchCtx.depth--

    if batchCtx.depth == 0 && len(batchCtx.pending) > 0 {
        // Execute each unique binding once
        // Map ensures deduplication - same binding ID overwrites previous
        for _, execute := range batchCtx.pending {
            execute()
        }
        // Reset for next batch
        batchCtx.pending = make(map[uint64]func())
    }
}
```

**Why this works**: Each binding has a unique ID. When the same state is set multiple times during a batch, the pending map entry for that binding ID is overwritten with a closure capturing the latest value. Only the final value is used when bindings execute.

### 3.3 Thread Safety

State operations have specific thread safety requirements that integrate with the existing `tui.MarkDirty()` atomic flag:

```go
// SAFE: Get() from any goroutine
go func() {
    value := count.Get()  // OK - uses RLock
}()

// UNSAFE: Set() from background goroutine
go func() {
    count.Set(5)  // WRONG - races with binding execution
}()

// SAFE: Use channel watchers (handlers run on main loop)
// In .tui file:
// onChannel={tui.Watch(dataCh, func(value int) {
//     count.Set(value)  // OK - handler runs on main loop
// })}

// SAFE: Use QueueUpdate for ad-hoc background updates
go func() {
    result := expensiveComputation()
    app.QueueUpdate(func() {
        count.Set(result)  // OK - runs on main loop
    })
}()
```

**Rule:** State must only be mutated from the main event loop. The existing `tui.Watcher` pattern and `app.QueueUpdate()` provide thread-safe mechanisms for background updates.

### 3.4 Analyzer Changes

Add state tracking and explicit deps detection to the existing analyzer:

```go
// pkg/tuigen/analyzer.go

type StateVar struct {
    Name     string   // Variable name (e.g., "count")
    Type     string   // Go type (e.g., "int", "string", "[]Item")
    InitExpr string   // Initialization expression
    Pos      Position
}

type StateBinding struct {
    StateVars    []string  // State variables referenced in expression
    Element      *Element  // Element that uses this expression
    Attribute    string    // Which attribute ("text", "class", etc.)
    Expr         string    // The expression (e.g., "fmt.Sprintf(...)")
    ExplicitDeps bool      // True if deps={...} was used
}

type ComponentAnalysis struct {
    // ... existing fields (refs, imports, etc.)
    StateVars []StateVar
    Bindings  []StateBinding
}

// Detect tui.NewState calls
func (a *Analyzer) detectStateVars(comp *Component) []StateVar {
    // Look for: varName := tui.NewState(initialValue)
    // Extract variable name, infer type from initialValue
    // Type inference rules:
    //   - Integer literal → int
    //   - Float literal → float64
    //   - String literal → string
    //   - Bool literal → bool
    //   - Composite literal → explicit type from literal
    //   - Function call → requires explicit type annotation
}

// Detect state usage in expressions
func (a *Analyzer) detectStateBindings(comp *Component, stateVars []StateVar) []StateBinding {
    // For each element:
    //   1. Check for explicit deps={[state1, state2]} attribute
    //   2. If no explicit deps, scan expression for stateVar.Get() calls
    //   3. Record the binding with detected/explicit state dependencies
}

// Handle explicit deps attribute
func (a *Analyzer) parseExplicitDeps(attr *Attribute) []string {
    // Parse: deps={[count, name]}
    // Returns: ["count", "name"]
}
```

### 3.5 Generator Changes

Generate binding code when state is used, following existing patterns from ref generation:

```go
// pkg/tuigen/generator.go

func (g *Generator) generateComponent(comp *Component) {
    analysis := g.analyzer.Analyze(comp)

    // Generate state variable declarations (no Context needed)
    for _, sv := range analysis.StateVars {
        g.writef("%s := tui.NewState(%s)\n", sv.Name, sv.InitExpr)
    }

    // Generate elements (existing logic)
    // Uses __tui_N naming for unnamed elements
    // Uses ref names for named elements (#RefName)
    // ...

    // Generate bindings after all elements are created
    for _, binding := range analysis.Bindings {
        g.generateBinding(binding)
    }
}

func (g *Generator) generateBinding(b StateBinding) {
    if len(b.StateVars) == 1 {
        // Single state variable - direct binding
        // count.Bind(func(v int) {
        //     span.SetText(fmt.Sprintf("Count: %d", v))
        // })
        g.writef("%s.Bind(func(v %s) {\n", b.StateVars[0], b.stateType)
        g.writef("    %s.%s(%s)\n", b.Element.VarName, b.setter, b.Expr)
        g.writef("})\n")
    } else {
        // Multiple state variables - shared update function
        // updateSpan := func() { span.SetText(expr) }
        // count.Bind(func(_ int) { updateSpan() })
        // name.Bind(func(_ string) { updateSpan() })
        updateFn := fmt.Sprintf("update%s", b.Element.VarName)
        g.writef("%s := func() { %s.%s(%s) }\n", updateFn, b.Element.VarName, b.setter, b.Expr)
        for _, stateVar := range b.StateVars {
            g.writef("%s.Bind(func(_ %s) { %s() })\n", stateVar, b.stateType, updateFn)
        }
    }
}
```

---

## 4. DSL Syntax

### 4.1 State Declaration

State is created with just the initial value - no Context parameter needed:

```tui
@component Counter() {
    // Simple types - type inferred from literal
    count := tui.NewState(0)
    name := tui.NewState("default")
    enabled := tui.NewState(true)

    // Complex types - type from composite literal
    items := tui.NewState([]string{})
    user := tui.NewState(&User{Name: "Alice"})

    // ...
}
```

### 4.2 State Usage in Elements

```tui
// Text content - auto-detected binding
<span>{count.Get()}</span>
<span>{fmt.Sprintf("Count: %d", count.Get())}</span>

// With formatting
<span class="font-bold">{name.Get()}</span>

// Conditional styling
<span class={enabled.Get() ? "text-green" : "text-red"}>{status.Get()}</span>
```

### 4.3 Explicit Dependencies (deps attribute)

For complex cases where auto-detection fails (helper functions, computed values), use explicit deps:

```tui
// Auto-detection works for direct .Get() calls
<span>{fmt.Sprintf("%d", count.Get())}</span>

// Explicit deps for helper functions
<span deps={[user, settings]}>{formatUserDisplay(user, settings)}</span>

// Explicit deps for complex expressions
<span deps={[items]}>{computeTotal(items.Get())}</span>
```

The analyzer will:
1. First check for `deps={...}` attribute
2. If not present, scan expression for `.Get()` calls
3. Generate bindings for all detected/explicit dependencies

### 4.4 State Updates in Handlers

Handlers don't return bool - `Set()` marks dirty automatically via `tui.MarkDirty()`:

```tui
// No bool return needed
func increment(count *tui.State[int]) func() {
    return func() {
        count.Set(count.Get() + 1)
        // Set() automatically calls tui.MarkDirty()
    }
}

// Using Update helper
func increment(count *tui.State[int]) func() {
    return func() {
        count.Update(func(v int) int { return v + 1 })
    }
}

// Batched updates for multiple states
func updateProfile(name *tui.State[string], age *tui.State[int]) func() {
    return func() {
        tui.Batch(func() {
            name.Set("Bob")
            age.Set(30)
        })
        // Bindings fire once, not twice
    }
}
```

### 4.5 Accessing State Outside Components

State is component-local by default. To access state from outside the component, pass it as a parameter:

```tui
@component Counter(count *tui.State[int]) {
    <div>
        <span>{fmt.Sprintf("Count: %d", count.Get())}</span>
        <button onClick={increment(count)}>+</button>
    </div>
}
```

Usage:

```go
// Create state outside component
count := tui.NewState(0)
view := Counter(count)
app.SetRoot(view)

// Access state from outside
count.Set(10)  // Updates UI automatically
```

This pattern keeps the component pure and testable while allowing external control when needed.

### 4.6 Manual Binding (Escape Hatch)

For cases the DSL can't handle, write manual bindings in Go:

```go
// In helper function or after generated code
count.Bind(func(v int) {
    // Complex update logic
    span.SetText(computeExpensiveValue(v))
})
```

---

## 5. Generated Output Examples

### 5.1 Simple Counter

**Input:**

```tui
@component Counter() {
    count := tui.NewState(0)

    <div class="flex-col gap-1">
        <span>{fmt.Sprintf("Count: %d", count.Get())}</span>
        <button onClick={increment(count)}>+</button>
    </div>
}

func increment(count *tui.State[int]) func() {
    return func() {
        count.Set(count.Get() + 1)
    }
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
    count := tui.NewState(0)

    // Create elements (using existing __tui_N pattern for unnamed elements)
    __tui_0 := element.New(
        element.WithText(fmt.Sprintf("Count: %d", count.Get())),
    )

    __tui_1 := element.New(
        element.WithText("+"),
        element.WithOnClick(increment(count)),
    )

    Root := element.New(
        element.WithDirection(layout.Column),
        element.WithGap(1),
    )
    Root.AddChild(__tui_0, __tui_1)

    // Bind state to elements
    count.Bind(func(v int) {
        __tui_0.SetText(fmt.Sprintf("Count: %d", v))
    })

    return CounterView{Root: Root, watchers: nil}
}

func increment(count *tui.State[int]) func() {
    return func() {
        count.Set(count.Get() + 1)
    }
}
```

### 5.2 Multiple State Variables

**Input:**

```tui
@component Profile() {
    name := tui.NewState("Alice")
    age := tui.NewState(30)

    <div>
        <span>{fmt.Sprintf("%s is %d years old", name.Get(), age.Get())}</span>
    </div>
}
```

**Output:**

```go
func Profile() ProfileView {
    name := tui.NewState("Alice")
    age := tui.NewState(30)

    __tui_0 := element.New(
        element.WithText(fmt.Sprintf("%s is %d years old", name.Get(), age.Get())),
    )

    Root := element.New()
    Root.AddChild(__tui_0)

    // Shared update function for expression with multiple state deps
    update__tui_0 := func() {
        __tui_0.SetText(fmt.Sprintf("%s is %d years old", name.Get(), age.Get()))
    }
    name.Bind(func(_ string) { update__tui_0() })
    age.Bind(func(_ int) { update__tui_0() })

    return ProfileView{Root: Root, watchers: nil}
}
```

### 5.3 State with Refs (Hybrid)

**Input:**

```tui
@component StreamBox(lineCount *tui.State[int]) {
    <div class="flex-col">
        <span>{fmt.Sprintf("Lines: %d", lineCount.Get())}</span>
        <div #Content scrollable={element.ScrollVertical}></div>
    </div>
}
```

**Output:**

```go
type StreamBoxView struct {
    Root     *element.Element
    Content  *element.Element
    watchers []tui.Watcher
}

func (v StreamBoxView) GetRoot() *element.Element   { return v.Root }
func (v StreamBoxView) GetWatchers() []tui.Watcher { return v.watchers }

func StreamBox(lineCount *tui.State[int]) StreamBoxView {
    __tui_0 := element.New(
        element.WithText(fmt.Sprintf("Lines: %d", lineCount.Get())),
    )

    Content := element.New(
        element.WithScrollable(element.ScrollVertical),
    )

    Root := element.New(element.WithDirection(layout.Column))
    Root.AddChild(__tui_0, Content)

    lineCount.Bind(func(v int) {
        __tui_0.SetText(fmt.Sprintf("Lines: %d", v))
    })

    return StreamBoxView{
        Root:     Root,
        Content:  Content,
        watchers: nil,
    }
}
```

**Usage:**

```go
// Create state outside component
lineCount := tui.NewState(0)
view := StreamBox(lineCount)
app.SetRoot(view)

// Add line using ref (imperative)
view.Content.AddChild(element.New(element.WithText(newLine)))
view.Content.ScrollToBottom()

// Update count using state parameter (reactive)
lineCount.Set(lineCount.Get() + 1)  // span updates automatically
```

### 5.4 Explicit Dependencies

**Input:**

```tui
@component UserCard() {
    user := tui.NewState(&User{Name: "Alice", Age: 30})

    // Auto-detection can't trace into formatUser()
    <span deps={[user]}>{formatUser(user.Get())}</span>
}

func formatUser(u *User) string {
    return fmt.Sprintf("%s (%d)", u.Name, u.Age)
}
```

**Output:**

```go
func UserCard() UserCardView {
    user := tui.NewState(&User{Name: "Alice", Age: 30})

    __tui_0 := element.New(
        element.WithText(formatUser(user.Get())),
    )

    Root := element.New()
    Root.AddChild(__tui_0)

    // Explicit deps - binds to user even though Get() isn't visible in expression
    user.Bind(func(_ *User) {
        __tui_0.SetText(formatUser(user.Get()))
    })

    return UserCardView{Root: Root, watchers: nil}
}
```

### 5.5 State with Channel Watchers

**Input:**

```tui
@component LiveData(dataCh <-chan int) {
    value := tui.NewState(0)

    <div onChannel={tui.Watch(dataCh, func(v int) {
        value.Set(v)  // Safe - handler runs on main loop
    })}>
        <span>{fmt.Sprintf("Value: %d", value.Get())}</span>
    </div>
}
```

**Output:**

```go
func LiveData(dataCh <-chan int) LiveDataView {
    value := tui.NewState(0)

    __tui_0 := element.New(
        element.WithText(fmt.Sprintf("Value: %d", value.Get())),
    )

    Root := element.New()
    Root.AddChild(__tui_0)

    value.Bind(func(v int) {
        __tui_0.SetText(fmt.Sprintf("Value: %d", v))
    })

    watchers := []tui.Watcher{
        tui.Watch(dataCh, func(v int) {
            value.Set(v)
        }),
    }

    return LiveDataView{Root: Root, watchers: watchers}
}
```

---

## 6. User Experience

### 6.1 Complete Example

```tui
// todo.tui
package main

import "fmt"

@component TodoApp() {
    todos := tui.NewState([]string{})
    input := tui.NewState("")

    <div class="flex-col gap-1 p-1 border-single">
        <span class="font-bold">{"Todo List"}</span>

        <div class="flex gap-1">
            <input
                value={input.Get()}
                onInput={updateInput(input)}
                width={30}
            />
            <button onClick={addTodo(todos, input)}>Add</button>
        </div>

        <div #List class="flex-col">
            @for i, todo := range todos.Get() {
                <span>{fmt.Sprintf("%d. %s", i+1, todo)}</span>
            }
        </div>

        <span class="text-dim">{fmt.Sprintf("%d items", len(todos.Get()))}</span>
    </div>
}

func updateInput(input *tui.State[string]) func(string) {
    return func(value string) {
        input.Set(value)
    }
}

func addTodo(todos *tui.State[[]string], input *tui.State[string]) func() {
    return func() {
        if input.Get() != "" {
            tui.Batch(func() {
                todos.Set(append(todos.Get(), input.Get()))
                input.Set("")
            })
        }
    }
}
```

```go
// main.go
package main

import "github.com/grindlemire/go-tui/pkg/tui"

func main() {
    app, _ := tui.NewApp()
    defer app.Close()

    // SetRoot takes view directly - no Context needed
    app.SetRoot(TodoApp())

    app.Run()
}
```

> **Note:** The `@for` loop over `todos.Get()` generates static children at construction time. When `todos` changes, the bindings update the count display but NOT the list children. For dynamic lists, use refs with `AddChild()`/`RemoveAllChildren()`, or see Future Considerations for reactive loops.

---

## 7. Rules and Constraints

1. **State declared with `tui.NewState(initial)`** - no Context parameter needed
2. **Access via `.Get()`, update via `.Set()`** - required for binding detection
3. **Bindings are one-way (state → UI)** - no two-way binding magic
4. **State is component-local** - pass as parameter if external access is needed
5. **Handlers don't return bool** - `Set()` automatically calls `tui.MarkDirty()`
6. **Thread safety: main loop only** - call `Set()` only from main loop; use watchers or `QueueUpdate()` for background
7. **`tui.Batch()` for multiple updates** - defers and deduplicates bindings until batch completes
8. **`Bind()` returns `Unbind` handle** - for cleanup when needed
9. **Use `deps={...}` for complex cases** - when auto-detection fails
10. **Refs still needed for imperative operations** - scroll, focus, dynamic children
11. **Loops are not reactive** - `@for` over `state.Get()` doesn't auto-update children
12. **`#Name` is for element refs only** - not for state variables

---

## 8. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Medium\
**Recommended Phases:** 4

### Phase Breakdown

1. **Phase 1: State[T] Core Type** (Medium)
   - Create `pkg/tui/state.go`
   - Implement `State[T]` with `Get`, `Set`, `Update`, `Bind`
   - Add `Unbind` return from `Bind()` with ID-based deactivation
   - Add mutex for thread-safe `Get()` and binding management
   - Integrate with existing `tui.MarkDirty()` from `dirty.go`
   - Unit tests for basic operations

2. **Phase 2: Batching** (Small)
   - Add `batchContext` with depth and ID-keyed pending map
   - Implement `Batch()` function with proper deduplication
   - Defer binding execution during batch
   - Execute unique bindings with final values
   - Unit tests for batching behavior

3. **Phase 3: Analyzer Detection** (Medium)
   - Detect `tui.NewState(...)` declarations
   - Track state variable names and infer types
   - Detect `.Get()` calls in element expressions
   - Parse `deps={[state1, state2]}` attribute
   - Build binding list with explicit/detected deps
   - Recognize state parameters in component signatures

4. **Phase 4: Generator Binding Code** (Medium)
   - Generate `Bind()` calls for each state-element binding
   - Handle multiple state variables in single expression
   - Handle explicit `deps` attribute
   - Handle state passed as component parameters
   - Update examples to use state pattern

---

## 9. Success Criteria

1. `tui.NewState(initialValue)` creates a `State[T]` with correct type inference (no Context)
2. `state.Get()` returns current value (thread-safe via RLock)
3. `state.Set(newValue)` updates value, calls `tui.MarkDirty()`, and calls all bindings
4. `state.Bind(fn)` returns `Unbind` handle with unique ID
5. Calling `Unbind()` prevents future binding calls for that binding
6. `tui.Batch()` defers binding execution until batch completes
7. Multiple `Set()` calls to same state in batch trigger binding once with final value
8. Batch deduplication uses binding IDs (not function pointer comparison)
9. Generator detects state usage in element text expressions
10. Generator handles `deps={[state1, state2]}` attribute
11. Generator produces correct `Bind()` calls
12. Bound elements update automatically when state changes
13. Multiple state variables in one expression all trigger update
14. State works alongside refs without conflict
15. Handlers don't need bool return - dirty tracking is automatic via `tui.MarkDirty()`
16. State passed as component parameters works correctly
17. Example apps (counter, todo) work with state pattern (no Context)
18. Channel watchers can safely call `Set()` since handlers run on main loop

---

## 10. Future Considerations

### 10.1 Reactive Loops (Structural Reactivity)

Currently, `@for` loops execute once at construction time:

```tui
@for _, item := range items.Get() {
    <span>{item}</span>
}
```

When `items` changes, the children are NOT updated. Users must use refs for dynamic lists.

**Future:** Add `@reactive for` directive for auto-updating loops:

```tui
@reactive for _, item := range items.Get() {
    <span>{item}</span>
}
```

This would generate reconciliation code that:
1. Detects when `items` state changes
2. Compares new items to existing children
3. Adds/removes/reorders children as needed
4. Optionally uses `key` for stable identity

**Implementation sketch:**

```go
// Generated code for @reactive for
items.Bind(func(newItems []string) {
    // Simple replace-all strategy
    listContainer.RemoveAllChildren()
    for _, item := range newItems {
        listContainer.AddChild(element.New(element.WithText(item)))
    }
})
```

More sophisticated implementations could diff and patch for better performance.

### 10.2 Computed State

Derived values that update when dependencies change:

```tui
count := tui.NewState(0)
doubled := tui.Computed(func() int {
    return count.Get() * 2
})

<span>{doubled.Get()}</span>  // auto-updates when count changes
```

### 10.3 Other Considerations

- **State persistence**: Save/restore state across sessions
- **DevTools**: State inspection and time-travel debugging
- **Effects**: Side effects that run when state changes (beyond UI updates)

---

## 11. Relationship to Other Features

### Named Element Refs

State and refs are complementary:

- **State**: For reactive value display (text, counts, labels)
- **Refs**: For imperative operations (scroll, focus, dynamic children)

Most components will use state for display and refs only when imperative access is needed.

**Pattern for dynamic lists (until reactive loops are implemented):**

```go
// Create state outside component for external access
lineCount := tui.NewState(0)
view := StreamBox(lineCount)
app.SetRoot(view)

// Use ref for dynamic children
view.Content.AddChild(element.New(element.WithText(newLine)))

// Use state parameter for count display
lineCount.Set(lineCount.Get() + 1)
```

### Event Handlers

Handlers receive state as parameter and call `Set()` to trigger updates. No bool return needed:

```tui
func onClick(count *tui.State[int]) func() {
    return func() {
        count.Set(count.Get() + 1)
        // Set() automatically:
        // 1. Updates value
        // 2. Calls tui.MarkDirty() (existing atomic flag)
        // 3. Executes bindings
    }
}
```

### Channel Watchers

For streaming data, use `tui.Watch()` in the .tui file. The handler runs on the main loop via `app.eventQueue`, making state mutations safe:

```tui
<div onChannel={tui.Watch(dataCh, func(value int) {
    count.Set(value)  // Safe - runs on main loop
})}>
```

This integrates with the existing `tui.Watcher` interface and `App.SetRoot()` watcher startup.

### Thread Safety

`State.Get()` is safe from any goroutine. `State.Set()` must only be called from the main loop. For background updates:

```go
// Option 1: Channel watcher (preferred) - in .tui file
// onChannel={tui.Watch(dataCh, handler)}

// Option 2: QueueUpdate - in Go code
go func() {
    result := expensiveComputation()
    app.QueueUpdate(func() {
        count.Set(result)  // Safe - runs on main loop
    })
}()
```

Both options leverage the existing `eventQueue` channel in `App` to safely marshal updates to the main loop.

---

## 12. Testing Strategy

### Unit Tests (pkg/tui/state_test.go)

```go
func TestState_GetSet(t *testing.T) {
    // Basic get/set operations
}

func TestState_Bind(t *testing.T) {
    // Binding called on Set
}

func TestState_Unbind(t *testing.T) {
    // Unbind prevents future calls
}

func TestState_Update(t *testing.T) {
    // Update helper function
}

func TestBatch_SingleState(t *testing.T) {
    // Multiple sets, binding called once
}

func TestBatch_MultipleStates(t *testing.T) {
    // Different states, each binding called once
}

func TestBatch_Deduplication(t *testing.T) {
    // Same binding ID not called multiple times
}

func TestBatch_Nested(t *testing.T) {
    // Nested Batch calls work correctly
}

func TestState_ThreadSafeGet(t *testing.T) {
    // Concurrent Get calls are safe
}
```

### Integration Tests (pkg/tuigen/*_test.go)

```go
func TestAnalyzer_DetectsStateVars(t *testing.T) {
    // Detects tui.NewState declarations
}

func TestAnalyzer_DetectsStateBindings(t *testing.T) {
    // Detects .Get() calls in expressions
}

func TestAnalyzer_ParsesExplicitDeps(t *testing.T) {
    // Parses deps={[state1, state2]}
}

func TestGenerator_GeneratesBindings(t *testing.T) {
    // Generates correct Bind() calls
}

func TestGenerator_StateAsParameter(t *testing.T) {
    // State passed as component parameter works correctly
}
```
