# State Management Specification

**Status:** Draft\
**Version:** 2.1\
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
- **Automatic dirty tracking**: `State.Set()` calls `MarkDirty()` - no bool returns needed
- **Batched updates**: `tui.Batch()` coalesces multiple `Set()` calls
- **Unbind support**: `Bind()` returns handle for cleanup
- **Thread safety**: Clear rules for main-loop-only mutations

### Non-Goals

- Full virtual DOM / diffing (structural reactivity for loops is future work)
- Reactive primitives beyond `State[T]` (computed, effects, etc.)
- Global state store (this is component-local state)

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
│  // Set() automatically marks dirty - no bool return needed     │
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
}

type binding[T any] struct {
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
// updates, use app.QueueUpdate().
func (s *State[T]) Set(v T) {
    s.mu.Lock()
    s.value = v
    // Copy active bindings while holding lock
    bindings := make([]func(T), 0, len(s.bindings))
    for _, b := range s.bindings {
        if b.active {
            bindings = append(bindings, b.fn)
        }
    }
    s.mu.Unlock()

    // Mark dirty (automatic - no bool return needed from handlers)
    MarkDirty()

    // Execute bindings outside lock (they may call Get())
    if batchDepth == 0 {
        // Immediate execution
        for _, fn := range bindings {
            fn(v)
        }
    } else {
        // Deferred execution during batch
        for _, fn := range bindings {
            fn := fn // capture
            pendingBindings = append(pendingBindings, func() { fn(v) })
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
    b := &binding[T]{fn: fn, active: true}
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

Batch defers binding execution until after all state updates complete, avoiding redundant updates.

```go
// pkg/tui/state.go (continued)

var (
    batchDepth      int
    pendingBindings []func()
)

// Batch executes fn and defers all binding callbacks until fn returns.
// Use this when updating multiple states to avoid redundant element updates.
//
// Example:
//   tui.Batch(func() {
//       firstName.Set("Bob")
//       lastName.Set("Smith")
//       age.Set(30)
//   })
//   // Bindings fire once here, not three times
func Batch(fn func()) {
    batchDepth++
    fn()
    batchDepth--

    if batchDepth == 0 && len(pendingBindings) > 0 {
        // Deduplicate bindings (same function may be queued multiple times)
        seen := make(map[*func()]bool)
        for _, b := range pendingBindings {
            ptr := &b
            if !seen[ptr] {
                seen[ptr] = true
                b()
            }
        }
        pendingBindings = nil
    }
}
```

### 3.3 Thread Safety

State operations have specific thread safety requirements:

```go
// SAFE: Get() from any goroutine
go func() {
    value := count.Get()  // OK - uses RLock
}()

// UNSAFE: Set() from background goroutine
go func() {
    count.Set(5)  // WRONG - races with binding execution
}()

// SAFE: Use QueueUpdate for background updates
go func() {
    app.QueueUpdate(func() {
        count.Set(5)  // OK - runs on main loop
    })
}()

// PREFERRED: Use channel watchers for streaming data (in .tui file)
// onChannel={tui.Watch(dataCh, func(value int) {
//     count.Set(value)  // OK - handler runs on main loop
// })}
```

> **Rule:** State must only be mutated from the main event loop. Use channel watchers (`tui.Watch`) or `app.QueueUpdate()` for background updates.

### 3.4 Analyzer Changes

Add state tracking and explicit deps detection to the analyzer:

```go
// pkg/tuigen/analyzer.go

type StateVar struct {
    Name     string   // Variable name (e.g., "count")
    Type     string   // Go type (e.g., "int", "string", "[]Item")
    InitExpr string   // Initialization expression
    Pos      Position
}

type StateBinding struct {
    StateVars   []string  // State variables referenced in expression
    Element     *Element  // Element that uses this expression
    Attribute   string    // Which attribute ("text", "class", etc.)
    Expr        string    // The expression (e.g., "fmt.Sprintf(...)")
    ExplicitDeps bool     // True if deps={...} was used
}

type ComponentAnalysis struct {
    // ... existing fields
    StateVars []StateVar
    Bindings  []StateBinding
}

// Detect tui.NewState calls
func (a *Analyzer) detectStateVars(comp *Component) []StateVar {
    // Look for: varName := tui.NewState(initialValue)
    // Extract variable name, infer type from initialValue
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

Generate binding code when state is used:

```go
// pkg/tuigen/generator.go

func (g *Generator) generateComponent(comp *Component) {
    analysis := g.analyzer.Analyze(comp)

    // Generate state variable declarations (no Context needed)
    for _, sv := range analysis.StateVars {
        g.writef("%s := tui.NewState(%s)\n", sv.Name, sv.InitExpr)
    }

    // Generate elements (existing logic)
    // ...

    // Generate bindings
    for _, binding := range analysis.Bindings {
        g.generateBinding(binding)
    }
}

func (g *Generator) generateBinding(b StateBinding) {
    // For single state variable:
    // count.Bind(func(v int) {
    //     span.SetText(fmt.Sprintf("Count: %d", v))
    // })

    // For multiple state variables (or explicit deps):
    // updateSpan := func() { span.SetText(expr) }
    // count.Bind(func(_ int) { updateSpan() })
    // name.Bind(func(_ string) { updateSpan() })
}
```

---

## 4. DSL Syntax

### 4.1 State Declaration

State is created with just the initial value - no Context parameter needed:

```tui
@component Counter() {
    // Simple types
    count := tui.NewState(0)
    name := tui.NewState("default")
    enabled := tui.NewState(true)

    // Complex types
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

Handlers don't return bool - `Set()` marks dirty automatically:

```tui
// No bool return needed
func increment(count *tui.State[int]) func() {
    return func() {
        count.Set(count.Get() + 1)
        // Set() automatically calls MarkDirty()
    }
}

// Using Update helper
func increment(count *tui.State[int]) func() {
    return func() {
        count.Update(func(v int) int { return v + 1 })
    }
}

// Batched updates for multiple states
func updateProfile(name, age *tui.State[string], *tui.State[int]) func() {
    return func() {
        tui.Batch(func() {
            name.Set("Bob")
            age.Set(30)
        })
        // Bindings fire once, not twice
    }
}
```

### 4.5 Manual Binding (Escape Hatch)

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

// No bool return needed
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
    var view CounterView

    count := tui.NewState(0)

    // Create elements
    span := element.New(
        element.WithText(fmt.Sprintf("Count: %d", count.Get())),
    )

    button := element.New(
        element.WithText("+"),
        element.WithOnClick(increment(count)),  // no bool return
    )

    Root := element.New(
        element.WithDirection(layout.Column),
        element.WithGap(1),
    )
    Root.AddChild(span, button)

    // Bind state to elements
    count.Bind(func(v int) {
        span.SetText(fmt.Sprintf("Count: %d", v))
    })

    view = CounterView{Root: Root, watchers: nil}
    return view
}

// No bool return - Set() marks dirty automatically
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
    var view ProfileView

    name := tui.NewState("Alice")
    age := tui.NewState(30)

    span := element.New(
        element.WithText(fmt.Sprintf("%s is %d years old", name.Get(), age.Get())),
    )

    Root := element.New()
    Root.AddChild(span)

    // Shared update function for expression with multiple state deps
    updateSpan := func() {
        span.SetText(fmt.Sprintf("%s is %d years old", name.Get(), age.Get()))
    }
    name.Bind(func(_ string) { updateSpan() })
    age.Bind(func(_ int) { updateSpan() })

    view = ProfileView{Root: Root, watchers: nil}
    return view
}
```

### 5.3 State with Refs (Hybrid)

**Input:**

```tui
@component StreamBox() {
    lineCount := tui.NewState(0)

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

func StreamBox() StreamBoxView {
    var view StreamBoxView

    lineCount := tui.NewState(0)

    span := element.New(
        element.WithText(fmt.Sprintf("Lines: %d", lineCount.Get())),
    )

    Content := element.New(
        element.WithScrollable(element.ScrollVertical),
    )

    Root := element.New(element.WithDirection(layout.Column))
    Root.AddChild(span, Content)

    lineCount.Bind(func(v int) {
        span.SetText(fmt.Sprintf("Lines: %d", v))
    })

    view = StreamBoxView{Root: Root, Content: Content, watchers: nil}
    return view
}
```

**Usage:**

```go
view := StreamBox()
app.SetRoot(view)  // Takes view directly

// Add line using ref (imperative)
view.Content.AddChild(element.New(element.WithText(newLine)))
view.Content.ScrollToBottom()

// Update count using state (reactive)
lineCount.Set(lineCount.Get() + 1)  // span updates automatically, dirty marked
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
    var view UserCardView

    user := tui.NewState(&User{Name: "Alice", Age: 30})

    span := element.New(
        element.WithText(formatUser(user.Get())),
    )

    Root := element.New()
    Root.AddChild(span)

    // Explicit deps - binds to user even though Get() isn't in span expression
    user.Bind(func(_ *User) {
        span.SetText(formatUser(user.Get()))
    })

    view = UserCardView{Root: Root, watchers: nil}
    return view
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

// No bool return - Set() marks dirty automatically
func updateInput(input *tui.State[string]) func(string) {
    return func(value string) {
        input.Set(value)
    }
}

// Using Batch for multiple state updates
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
4. **State is component-local** - not shared between components unless passed as parameter
5. **Handlers don't return bool** - `Set()` automatically marks dirty
6. **Thread safety: main loop only** - call `Set()` only from main loop; use `QueueUpdate()` for background
7. **`tui.Batch()` for multiple updates** - defers bindings until batch completes
8. **`Bind()` returns `Unbind` handle** - for cleanup when needed
9. **Use `deps={...}` for complex cases** - when auto-detection fails
10. **Refs still needed for imperative operations** - scroll, focus, dynamic children
11. **Loops are not reactive** - `@for` over `state.Get()` doesn't auto-update children

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
   - Add `Unbind` return from `Bind()` with active flag pattern
   - Add mutex for thread-safe `Get()` and binding management
   - Integrate with `MarkDirty()` from Event Handling spec
   - Unit tests for basic operations

2. **Phase 2: Batching** (Small)
   - Add `batchDepth` and `pendingBindings` variables
   - Implement `Batch()` function
   - Defer binding execution during batch
   - Unit tests for batching behavior

3. **Phase 3: Analyzer Detection** (Medium)
   - Detect `tui.NewState(...)` declarations
   - Track state variable names and types
   - Detect `.Get()` calls in element expressions
   - Parse `deps={[state1, state2]}` attribute
   - Build binding list with explicit/detected deps

4. **Phase 4: Generator Binding Code** (Medium)
   - Generate `Bind()` calls for each state-element binding
   - Handle multiple state variables in single expression
   - Handle explicit `deps` attribute
   - Generate Viewable interface for view structs
   - Update examples to use state pattern

---

## 9. Success Criteria

1. `tui.NewState(initialValue)` creates a `State[T]` with correct type inference (no Context)
2. `state.Get()` returns current value (thread-safe)
3. `state.Set(newValue)` updates value, marks dirty, and calls all bindings
4. `state.Bind(fn)` returns `Unbind` handle
5. Calling `Unbind()` prevents future binding calls
6. `tui.Batch()` defers binding execution until batch completes
7. Multiple `Set()` calls in a batch trigger bindings once
8. Generator detects state usage in element text expressions
9. Generator handles `deps={[state1, state2]}` attribute
10. Generator produces correct `Bind()` calls
11. Bound elements update automatically when state changes
12. Multiple state variables in one expression all trigger update
13. State works alongside refs without conflict
14. Handlers don't need bool return - dirty tracking is automatic
15. Example apps (counter, todo) work with state pattern (no Context)

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
// Use state for count display
lineCount := tui.NewState(0)

// Use ref for dynamic children
view.Content.AddChild(element.New(element.WithText(newLine)))
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
        // 2. Calls MarkDirty()
        // 3. Executes bindings
    }
}
```

### Channel Watchers

For streaming data, use `tui.Watch()` in the .tui file. The handler runs on the main loop, making state mutations safe:

```tui
<div onChannel={tui.Watch(dataCh, func(value int) {
    count.Set(value)  // Safe - runs on main loop
})}>
```

### Thread Safety

State.Get() is safe from any goroutine. State.Set() must only be called from the main loop. For background updates:

```go
// Option 1: Channel watcher (preferred) - in .tui file
// onChannel={tui.Watch(dataCh, handler)}

// Option 2: QueueUpdate - in Go code
go func() {
    app.QueueUpdate(func() {
        count.Set(computedValue)  // Safe - runs on main loop
    })
}()
```
