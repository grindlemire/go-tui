# Reactive Bindings Implementation Plan

Implementation phases for the reactive state management system with `State[T]` type, automatic bindings, and batch support. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: State[T] Core Type

**Reference:** [reactive-bindings-design.md §3.1](./reactive-bindings-design.md#31-statet-type)

**Status:** Complete

- [x] Create `pkg/tui/state.go`
  - Add package documentation explaining State purpose and thread safety rules
  - See [design §3.1](./reactive-bindings-design.md#31-statet-type)

- [x] Implement `State[T]` struct in `pkg/tui/state.go`
  - Add `mu sync.RWMutex` field for thread safety
  - Add `value T` field for current value
  - Add `bindings []*binding[T]` field for registered bindings
  - Add `nextID uint64` field for unique binding IDs

- [x] Implement `binding[T]` struct in `pkg/tui/state.go`
  - Add `id uint64` field for unique identification
  - Add `fn func(T)` field for callback
  - Add `active bool` field for unbind support

- [x] Implement `Unbind` type in `pkg/tui/state.go`
  - Define as `type Unbind func()`

- [x] Implement `NewState[T any](initial T) *State[T]` in `pkg/tui/state.go`
  - Return new State with value set to initial
  - No Context parameter needed

- [x] Implement `Get() T` method in `pkg/tui/state.go`
  - Acquire read lock (`mu.RLock()`)
  - Return value
  - Release lock with defer

- [x] Implement `Set(v T)` method in `pkg/tui/state.go`
  - Acquire write lock (`mu.Lock()`)
  - Update value
  - Copy active bindings to local slice while holding lock
  - Release lock before calling bindings
  - Call `tui.MarkDirty()` (existing function in `dirty.go`)
  - Execute bindings (or queue if batching - Phase 2)
  - See [design §3.1](./reactive-bindings-design.md#31-statet-type)

- [x] Implement `Update(fn func(T) T)` method in `pkg/tui/state.go`
  - Call `s.Set(fn(s.Get()))`

- [x] Implement `Bind(fn func(T)) Unbind` method in `pkg/tui/state.go`
  - Acquire write lock
  - Generate unique ID from `nextID` counter
  - Create binding with ID, fn, and active=true
  - Append to bindings slice
  - Release lock
  - Return Unbind function that sets active=false

- [x] Add tests to `pkg/tui/state_test.go`
  - Test `NewState` creates state with initial value
  - Test `Get` returns current value
  - Test `Set` updates value
  - Test `Set` calls `MarkDirty()` (verify dirty flag set)
  - Test `Set` calls all registered bindings
  - Test `Update` applies function and sets result
  - Test `Bind` registers callback
  - Test `Bind` callback receives new value on Set
  - Test `Unbind` prevents future callback invocations
  - Test multiple bindings all called
  - Test binding removed by Unbind not called
  - Test concurrent `Get` calls are safe (parallel test)
  - Test binding execution happens outside lock (no deadlock on Get inside binding)

**Acceptance:** `go test ./pkg/tui/... -run State` passes

---

## Phase 2: Batching

**Reference:** [reactive-bindings-design.md §3.2](./reactive-bindings-design.md#32-batching)

**Status:** Complete

- [x] Add batch context to `pkg/tui/state.go`
  - Define `batchContext` struct with `depth int` and `pending map[uint64]func()` fields
  - Create package-level `batchCtx` variable initialized with empty map
  - See [design §3.2](./reactive-bindings-design.md#32-batching)

- [x] Modify `Set()` method in `pkg/tui/state.go`
  - Check `batchCtx.depth == 0` for immediate execution
  - If batching (depth > 0), store binding closures in `batchCtx.pending` keyed by binding ID
  - Keying by ID ensures deduplication (later Set overwrites earlier)

- [x] Implement `Batch(fn func())` function in `pkg/tui/state.go`
  - Increment `batchCtx.depth`
  - Call `fn()`
  - Decrement `batchCtx.depth`
  - If depth returns to 0 and pending map not empty:
    - Execute all pending binding callbacks
    - Reset pending map to empty

- [x] Add tests to `pkg/tui/state_test.go`
  - Test `Batch` defers binding execution until fn returns
  - Test multiple Sets to same state in batch calls binding once
  - Test binding receives final value (not intermediate values)
  - Test multiple different states in batch each call their bindings once
  - Test nested Batch calls work correctly (bindings fire at outermost)
  - Test deduplication uses binding ID (not function pointer)
  - Test Batch with no Sets doesn't error

**Acceptance:** `go test ./pkg/tui/... -run "State|Batch"` passes

---

## Phase 3: Analyzer Detection

**Reference:** [reactive-bindings-design.md §3.4](./reactive-bindings-design.md#34-analyzer-changes)

**Status:** Complete

- [x] Add state tracking types to `pkg/tuigen/analyzer.go`
  - Add `StateVar` struct with fields: Name, Type, InitExpr, Pos
  - Add `StateBinding` struct with fields: StateVars, Element, Attribute, Expr, ExplicitDeps
  - Add `StateVars []StateVar` field to `ComponentAnalysis` struct
  - Add `Bindings []StateBinding` field to `ComponentAnalysis` struct
  - See [design §3.4](./reactive-bindings-design.md#34-analyzer-changes)

- [x] Implement `detectStateVars()` method in `pkg/tuigen/analyzer.go`
  - Walk component code section looking for assignment statements
  - Match pattern: `varName := tui.NewState(expr)`
  - Extract variable name
  - Infer type from initializer expression:
    - Integer literal → `int`
    - Float literal → `float64`
    - String literal → `string`
    - Bool literal → `bool`
    - Composite literal → extract type from literal
    - Function call or complex expr → require type annotation or report error
  - Store initialization expression for code generation
  - Return list of StateVar

- [x] Implement `detectStateBindings()` method in `pkg/tuigen/analyzer.go`
  - Walk element tree
  - For each element with text expression or dynamic attribute:
    - Check for `deps={[...]}` attribute first
    - If present, parse explicit dependencies
    - If not present, scan expression for `.Get()` calls
    - Match state variable names against detected Get calls
    - Record binding with state vars, element, attribute, and expression
  - Return list of StateBinding

- [x] Implement `parseExplicitDeps()` method in `pkg/tuigen/analyzer.go`
  - Parse `deps={[state1, state2]}` attribute syntax
  - Extract list of state variable names
  - Validate each name exists in detected StateVars
  - Return error if unknown state variable referenced

- [x] Handle state as component parameters in `pkg/tuigen/analyzer.go`
  - Detect `*tui.State[T]` types in component parameter list
  - Add these to StateVars with appropriate type info
  - Parameter states don't have InitExpr

- [x] Add tests to `pkg/tuigen/analyzer_test.go`
  - Test detection of `tui.NewState(0)` with int type inference
  - Test detection of `tui.NewState("hello")` with string type inference
  - Test detection of `tui.NewState(true)` with bool type inference
  - Test detection of `tui.NewState([]string{})` with slice type
  - Test detection of `tui.NewState(&User{})` with pointer type
  - Test state usage detection via `.Get()` calls
  - Test binding detection in text expressions
  - Test binding detection in attribute expressions
  - Test explicit `deps={[state1]}` parsing
  - Test explicit deps with multiple states
  - Test error on unknown state in deps
  - Test state parameter detection in component signature
  - Test multiple state variables in single expression detected

**Acceptance:** `go test ./pkg/tuigen/... -run Analyzer` passes

---

## Phase 4: Generator Binding Code

**Reference:** [reactive-bindings-design.md §3.5](./reactive-bindings-design.md#35-generator-changes)

**Status:** Complete

- [x] Modify `pkg/tuigen/generator.go` - state variable generation
  - For each StateVar in analysis, generate declaration
  - Format: `varName := tui.NewState(initExpr)`
  - Skip for parameter states (already passed in)

- [x] Modify `pkg/tuigen/generator.go` - binding generation for single state
  - For bindings with single state variable:
  - Generate direct Bind call:
    ```go
    stateName.Bind(func(v Type) {
        element.SetText(expr)  // or other setter
    })
    ```
  - The `v` parameter can be used in expression if it's simple `state.Get()` replacement
  - See [design §3.5](./reactive-bindings-design.md#35-generator-changes)

- [x] Modify `pkg/tuigen/generator.go` - binding generation for multiple states
  - For bindings with multiple state variables:
  - Generate shared update function:
    ```go
    updateElement := func() { element.SetText(expr) }
    state1.Bind(func(_ Type1) { updateElement() })
    state2.Bind(func(_ Type2) { updateElement() })
    ```
  - Update function calls Get() on all states to get current values

- [x] Modify `pkg/tuigen/generator.go` - explicit deps handling
  - When `ExplicitDeps` is true, use deps list instead of auto-detected
  - Generate same binding code as auto-detected

- [x] Modify `pkg/tuigen/generator.go` - attribute bindings
  - Handle bindings on `class` attribute (SetClass or similar)
  - Handle bindings on other dynamic attributes
  - Generate appropriate setter call for each attribute type

- [x] Ensure proper import handling in `pkg/tuigen/generator.go`
  - Add `"github.com/grindlemire/go-tui/pkg/tui"` import when State is used
  - Existing import logic should handle this

- [x] Add tests to `pkg/tuigen/generator_test.go`
  - Test state declaration is generated correctly
  - Test single state binding generates direct Bind call
  - Test multiple state binding generates shared update function
  - Test explicit deps generates correct bindings
  - Test parameter state doesn't generate declaration
  - Test binding uses correct setter method
  - Test generated code compiles (go build check)
  - Test generated code runs correctly (functional test with mock)

- [x] Create example `examples/counter-state/`
  - Create `counter.tui` with state-based counter
  - Create `main.go` to run the example
  - Verify example compiles and runs
  - Demonstrate state binding in action

- [x] Update existing examples if beneficial
  - Consider updating `examples/streaming-dsl/` to use state where appropriate
  - Document state usage patterns

**Acceptance:** `go test ./pkg/tuigen/... -run Generator` passes, examples build and run correctly

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | State[T] Core Type | Complete |
| 2 | Batching | Complete |
| 3 | Analyzer Detection | Complete |
| 4 | Generator Binding Code | Complete |

## Files to Create

```
pkg/tui/
├── state.go           # NEW: State[T] type with Bind/Unbind, Batch
└── state_test.go      # NEW: State and Batch tests

examples/counter-state/
├── counter.tui        # NEW: State-based counter example
├── counter_tui.go     # NEW: Generated code
└── main.go            # NEW: Example runner
```

## Files to Modify

```
pkg/tuigen/
├── analyzer.go        # Add state detection, binding detection, deps parsing
├── analyzer_test.go   # Add state detection tests
├── generator.go       # Add state declaration and binding generation
└── generator_test.go  # Add binding generation tests
```

## Files Unchanged

| File | Reason |
|------|--------|
| `pkg/tui/dirty.go` | Already exists, provides MarkDirty() |
| `pkg/tui/app.go` | No changes needed - eventQueue already supports this pattern |
| `pkg/tui/watcher.go` | No changes needed - integrates via main loop |
| `pkg/tuigen/lexer.go` | No new tokens needed |
| `pkg/tuigen/parser.go` | Attributes already parsed, deps={} is standard attribute |
| `pkg/tuigen/tailwind.go` | No changes for state |
| `pkg/formatter/` | No state-specific formatting needed |
| `pkg/lsp/` | LSP support can be added later |
| `pkg/tui/element/` | Existing SetText, etc. already call MarkDirty() |

## Dependencies

```
Phase 1 ─────► Phase 2
    │              │
    │              ▼
    └─────────► Phase 3 ─────► Phase 4
```

- Phase 2 depends on Phase 1 (batching modifies Set() behavior)
- Phase 3 depends on Phase 1 (needs State type to exist)
- Phase 4 depends on Phases 2 and 3 (generates code using both)

## Integration Points

This feature integrates with existing systems:

| System | Integration |
|--------|-------------|
| `tui.MarkDirty()` | State.Set() calls MarkDirty() to trigger render |
| `App.eventQueue` | Watchers queue updates, handlers run on main loop |
| `element.SetText()` | Bindings call existing mutator methods |
| `tui.Watcher` | Channel watchers can safely call State.Set() |

## Success Criteria

From [design §9](./reactive-bindings-design.md#9-success-criteria):

1. `tui.NewState(initialValue)` creates a `State[T]` with correct type inference
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
15. Handlers don't need bool return - dirty tracking is automatic
16. State passed as component parameters works correctly
17. Example apps work with state pattern
18. Channel watchers can safely call `Set()` since handlers run on main loop
