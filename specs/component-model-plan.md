# Component Model & Broadcast Key System Implementation Plan

Implementation phases for the component model and broadcast key dispatch system. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Core Interfaces and KeyMap Types

**Reference:** [component-model-design.md §3.1–§3.2](./component-model-design.md#31-component-interfaces)
**Review:** false

**Completed in commit:** (pending)

- [ ] Create `component.go`
  - Define `Component` interface: `Render() *Element`
  - Define `KeyListener` interface: `KeyMap() KeyMap`
  - Define `Initializer` interface: `Init() func()`
  - These are pure interface definitions with no implementation logic

- [ ] Create `keymap.go`
  - Define `KeyMap` type: `[]KeyBinding`
  - Define `KeyBinding` struct: `Pattern KeyPattern`, `Handler func(KeyEvent)`, `Stop bool`
  - Define `KeyPattern` struct: `Key Key`, `Rune rune`, `AnyRune bool`, `Mods Modifier`
  - Note: `Key`, `Modifier`, and `KeyEvent` types already exist in `key.go` and `event.go`
  - Implement helper constructors:
    - `OnKey(key Key, handler func(KeyEvent)) KeyBinding`
    - `OnKeyStop(key Key, handler func(KeyEvent)) KeyBinding`
    - `OnRune(r rune, handler func(KeyEvent)) KeyBinding`
    - `OnRuneStop(r rune, handler func(KeyEvent)) KeyBinding`
    - `OnRunes(handler func(KeyEvent)) KeyBinding`
    - `OnRunesStop(handler func(KeyEvent)) KeyBinding`
  - See [design §3.2](./component-model-design.md#32-keymap-types) for exact signatures

- [ ] Create `keymap_test.go`
  - Test helper constructors produce correct `KeyBinding` values
  - Test `KeyPattern` equality (used as map-adjacent lookups later)
  - Test `OnKey`/`OnKeyStop` set `Stop` flag correctly
  - Test `OnRunes`/`OnRunesStop` set `AnyRune` flag correctly

**Tests:** `go test -run TestKeyMap ./...` once at phase end

---

## Phase 2: Mount System and Instance Caching

**Reference:** [component-model-design.md §3.3](./component-model-design.md#33-mount-and-instance-caching)

**Completed in commit:** (pending)

- [ ] Create `mount.go`
  - Define `mountKey` struct: `parent Component`, `index int`
  - Define `mountState` struct: `cache map[mountKey]Component`, `cleanups map[mountKey]func()`, `activeKeys map[mountKey]bool`
  - Implement `Mount(parent Component, index int, factory func() Component) *Element`:
    - Access cache via `currentApp.mounts` (per-App, not global)
    - Mark key in `activeKeys` (for mark-and-sweep)
    - On cache miss: call factory, cache instance, call `Init()` if `Initializer`
    - Call `instance.Render()`, set `el.component = instance`, return element
  - Implement `mountState.sweep()`:
    - Iterate cache, remove entries not in `activeKeys`, call cleanup functions
    - Reset `activeKeys` map for next render pass
  - Implement `newMountState() *mountState` constructor

- [ ] Modify `element.go` — Add `component` field to `Element` struct
  - Add unexported field: `component Component` (after existing fields, near line ~100)
  - This field is set by `Mount()` and read during tree walks for component discovery

- [ ] Modify `app.go` — Add mount state to App struct
  - Add field: `mounts *mountState` (in App struct, near line ~86)
  - Initialize in `NewApp()`: `mounts: newMountState()`
  - Add `walkComponents(root *Element, fn func(Component))` helper:
    - DFS walk of element tree
    - For each element with non-nil `component` field, call `fn(el.component)`
    - Recurse into children

- [ ] Modify `app_render.go` — Call sweep after render
  - After `root.Render()` completes (near line ~39), call `a.mounts.sweep()`
  - This cleans up components removed by conditional rendering

- [ ] Create `mount_test.go`
  - Test first mount calls factory and caches instance
  - Test subsequent mount returns cached instance (factory not called again)
  - Test `Init()` called on first mount for `Initializer` components
  - Test `Init()` cleanup function called by `sweep()` for removed components
  - Test mark-and-sweep: active components survive, inactive are cleaned up
  - Test mount with different (parent, index) keys cache independently
  - Use mock components implementing `Component` and `Initializer`

**Tests:** `go test -run TestMount ./...` once at phase end

---

## Phase 3: Dispatch Table and Key Broadcast

**Reference:** [component-model-design.md §3.4](./component-model-design.md#34-dispatch-table), [§9](./component-model-design.md#9-ordering-and-stop-propagation)

**Completed in commit:** (pending)

- [ ] Create `dispatch.go`
  - Define `dispatchEntry` struct: `pattern KeyPattern`, `handler func(KeyEvent)`, `stop bool`, `position int`
  - Define `dispatchTable` struct: `entries []dispatchEntry` (single unified list, tree-ordered)
  - Implement `(e *dispatchEntry) matches(ke KeyEvent) bool`:
    - AnyRune: match if `ke.Key == KeyRune`
    - Exact rune: match if `ke.Rune == p.Rune && ke.Key == KeyRune`
    - Exact key: match if `ke.Key == p.Key`
    - Modifier matching when `p.Mods != 0`
  - Implement `buildDispatchTable(root *Element) (*dispatchTable, error)`:
    - Use `walkComponents()` from Phase 2
    - Type-assert each component for `KeyListener`
    - Call `KeyMap()`, append entries with position counter
    - All handlers (exact and AnyRune) go in single `entries` slice
    - Call `validate()` before returning
  - Implement `(dt *dispatchTable) validate() error`:
    - Check for two active Stop handlers matching the same pattern
    - Return descriptive error (not panic) naming the conflict
  - Implement `(dt *dispatchTable) dispatch(ke KeyEvent)`:
    - Iterate `entries` in order, call `matches()`, execute matching handlers
    - If a matching handler has `stop=true`, return immediately
  - See [design §3.4](./component-model-design.md#34-dispatch-table) for unified tree-order dispatch

- [ ] Modify `app.go` — Add dispatch table field
  - Add field: `dispatchTable *dispatchTable` (in App struct)

- [ ] Modify `app_loop.go` — Rebuild dispatch table on dirty frame
  - After render completes and sweep runs (when `checkAndClearDirty()` is true):
    - Call `buildDispatchTable(root)` with the rendered element tree
    - Store result in `a.dispatchTable`
    - Log error if validation fails (don't crash — use last valid table)
  - Initial dispatch table built on first render

- [ ] Modify `app_events.go` — Replace FocusManager dispatch for key events
  - In `readInputEvents()` goroutine (lines ~52-78):
    - Global key handler still runs first (existing behavior, lines 69-73)
    - For KeyEvents not consumed by global handler:
      - Use `a.dispatchTable.dispatch(ke)` instead of `a.Dispatch(event)` → `FocusManager.Dispatch()`
    - MouseEvent and ResizeEvent dispatch remain unchanged (still use `App.Dispatch()`)
  - FocusManager is no longer used for key dispatch but remains for mouse/focus visual state

- [ ] Create `dispatch_test.go`
  - Test broadcast: multiple handlers for same key all fire
  - Test stop propagation: handler with Stop=true prevents later handlers
  - Test tree order: handlers fire in DFS position order
  - Test unified ordering: exact and AnyRune handlers interleave by tree position
  - Test AnyRune matches printable characters only
  - Test exact key match (e.g., KeyEscape, KeyCtrlC)
  - Test exact rune match (e.g., 'q', '/')
  - Test modifier matching
  - Test conflict validation: two Stop handlers for same pattern returns error
  - Test no false positives: Stop handler + broadcast handler for same pattern is valid
  - Test dispatch with nil/empty KeyMap components

**Tests:** `go test -run TestDispatch ./...` once at phase end

---

## Phase 4: Parser — Method Receiver on templ

**Reference:** [component-model-design.md §5.1](./component-model-design.md#51-method-receiver-on-templ-grammar-addition)

**Completed in commit:** (pending)

- [ ] Modify `internal/tuigen/ast.go` — Add receiver to Component node
  - Add field to `Component` struct (near line ~79): `Receiver string` (e.g., `"s *sidebar"`)
  - Add field: `ReceiverName string` (e.g., `"s"`) — the variable name for generated code
  - Add field: `ReceiverType string` (e.g., `"*sidebar"`) — the type for generated code
  - When `Receiver` is empty, it's a function component (existing behavior)
  - When `Receiver` is set, it's a method component (new behavior)

- [ ] Modify `internal/tuigen/parser_component.go` — Parse optional receiver on templ
  - In `parseTempl()` (line ~210):
    - After consuming `templ` token, check if next token is `(`
    - If `(`: parse as receiver `(name *Type)`, then expect method name and `()`
    - If identifier: parse as function name with params (existing behavior)
    - Disambiguation: after `templ`, `(` means receiver; identifier means function name
  - Populate `Component.Receiver`, `ReceiverName`, `ReceiverType` fields
  - Method templs have no params: `templ (s *sidebar) Render()` — the `()` is required but empty
  - The method name should always be `Render` (validate this)

- [ ] Modify `internal/tuigen/parser_component.go` — Distinguish struct component mount
  - Currently `@Component(args)` creates a `ComponentCall` AST node
  - Add field to `ComponentCall` struct: `IsStructMount bool`
  - When parsing inside a method templ (has receiver), set `IsStructMount = true`
  - When parsing inside a function templ (no receiver), keep `IsStructMount = false`
  - This flag tells the generator whether to emit `tui.Mount()` or a plain function call

- [ ] Create/extend `internal/tuigen/parser_component_test.go`
  - Test parsing `templ (s *sidebar) Render() { ... }` — receiver fields populated
  - Test parsing `templ Header(title string) { ... }` — receiver fields empty (existing)
  - Test `@Component(args)` inside method templ sets `IsStructMount = true`
  - Test `@Component(args)` inside function templ sets `IsStructMount = false`
  - Test error: `templ (s *sidebar) Render(params) { ... }` — method templs don't accept params
  - Test error: `templ (s *sidebar) NotRender() { ... }` — method name must be `Render`
  - Test both forms coexist in the same file

**Tests:** `go test -run TestParse ./internal/tuigen/...` once at phase end

---

## Phase 5: Generator — Mount Code Generation

**Reference:** [component-model-design.md §6](./component-model-design.md#6-generated-output-examples)

**Completed in commit:** (pending)

- [ ] Modify `internal/tuigen/generator_component.go` — Method receiver Render()
  - In `generateComponent()` (line ~4):
    - Check `comp.Receiver != ""`
    - If method component:
      - Do NOT generate a view struct (`ComponentNameView`)
      - Generate `func (recv) Render() *tui.Element { ... }` method signature
      - Body generation reuses existing element generation logic
      - The receiver variable (e.g., `s`) is available for expressions in the template
    - If function component: existing generation path (unchanged)
  - Key difference: method components return `*tui.Element` directly, not a view struct

- [ ] Modify `internal/tuigen/generator.go` — Pass through non-templ Go code
  - Currently the generator handles `GoFunc` and `GoDecl` AST nodes
  - Ensure type definitions (`type sidebar struct { ... }`), constructors, methods, and
    interface checks (`var _ tui.KeyListener = ...`) pass through verbatim to output
  - This should already work via `GoDecl` passthrough — verify and fix if needed

- [ ] Modify `internal/tuigen/generator_component.go` — Generate Mount() for struct component calls
  - When generating a `ComponentCall` with `IsStructMount = true`:
    - Track a mount index counter per method component (reset per component)
    - Generate: `tui.Mount(receiverVar, mountIndex, func() tui.Component { return ComponentName(args) })`
    - Increment mount index for next `@Component` in same method
  - When `IsStructMount = false`: generate plain function call (existing behavior, unchanged)
  - See [design §6.2](./component-model-design.md#62-parent-component-with-mount-syntax) for output format

- [ ] Create/extend `internal/tuigen/generator_test.go` test cases
  - Test method templ generates `func (s *sidebar) Render() *tui.Element { ... }`
  - Test `@Component(args)` in method templ generates `tui.Mount(s, 0, factory)`
  - Test mount indices increment: first `@A()` is index 0, second `@B()` is index 1
  - Test `@Component(args)` in function templ generates plain function call (unchanged)
  - Test non-templ code (type defs, constructors, methods) passes through verbatim
  - Test mixed file: type + constructor + KeyMap method + templ all in one .gsx
  - Use existing generator test patterns (golden file comparison or inline assertions)

**Tests:** `go test -run TestGenerat ./internal/tuigen/...` once at phase end

---

## Phase 6: Integration, FocusGroup Helper, and Examples

**Reference:** [component-model-design.md §8.4](./component-model-design.md#84-focus-groups-helper-not-interface), [§14](./component-model-design.md#14-end-to-end-example)

**Completed in commit:** (pending)

- [ ] Create `focus_group.go` — FocusGroup helper
  - Define `FocusGroup` struct: `members []*State[bool]`, `current int`
  - Implement `NewFocusGroup(members ...*State[bool]) *FocusGroup`
  - Implement `Next()`: deactivate current member, activate next (wrapping)
  - Implement `Prev()`: deactivate current member, activate previous (wrapping)
  - Implement `KeyMap() KeyMap`: bind Tab → Next, Shift+Tab → Prev
  - `FocusGroup` implements `KeyListener` but not `Component` (it's a helper, not a renderable)
  - See [design §8.4](./component-model-design.md#84-focus-groups-helper-not-interface) for spec

- [ ] Create `focus_group_test.go`
  - Test Next() cycles through members
  - Test Prev() cycles backward
  - Test wrapping at boundaries
  - Test KeyMap() returns Tab/Shift+Tab bindings
  - Test mutual exclusion: only one member active at a time

- [ ] Create example: `examples/component-model/` — Multi-component app
  - Create `app.gsx`: root component with conditional KeyMap, mounts Sidebar and SearchInput
  - Create `sidebar.gsx`: struct component with ctrl+b toggle, renders conditionally
  - Create `search.gsx`: struct component with conditional OnRunesStop, text input
  - Create `main.go`: entry point with `tui.NewApp(tui.WithRoot(MyApp()))`
  - This validates the full pipeline: .gsx → parser → generator → compile → run
  - See [design §14](./component-model-design.md#14-end-to-end-example) for the complete example

- [ ] Integration test: end-to-end component lifecycle
  - Test: create mock app with root component mounting two children
  - Verify: Mount caches instances, Init called, KeyMap collected
  - Verify: dispatch table built with correct tree-ordered entries
  - Verify: key press dispatches to correct handlers
  - Verify: state change → dirty → re-render → dispatch table rebuilt
  - Verify: conditional KeyMap activation/deactivation works
  - Verify: sweep cleans up unmounted components

**Tests:** `go test ./...` full suite once at phase end, plus `go build ./examples/component-model/...`

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Core interfaces (`Component`, `KeyListener`, `Initializer`) and KeyMap types with helpers | Pending |
| 2 | Mount system with per-App cache, mark-and-sweep cleanup, Element.component field | Pending |
| 3 | Dispatch table with unified tree-order, broadcast + stop propagation, App integration | Pending |
| 4 | Parser support for `templ (recv) Name()` method receiver and struct mount detection | Pending |
| 5 | Generator: method Render(), `tui.Mount()` code gen, verbatim passthrough | Pending |
| 6 | FocusGroup helper, multi-component example, end-to-end integration tests | Pending |

## Files to Create

```
component.go          — Component, KeyListener, Initializer interfaces
keymap.go             — KeyMap, KeyBinding, KeyPattern types + helpers
keymap_test.go        — KeyMap unit tests
mount.go              — Mount(), mountState, sweep()
mount_test.go         — Mount lifecycle tests
dispatch.go           — dispatchTable, buildDispatchTable(), dispatch()
dispatch_test.go      — Dispatch/broadcast/stop tests
focus_group.go        — FocusGroup helper (Tab cycling)
focus_group_test.go   — FocusGroup tests
examples/component-model/
├── app.gsx           — Root component
├── sidebar.gsx       — Sidebar component
├── search.gsx        — Search input component
└── main.go           — Entry point
```

## Files to Modify

| File | Changes |
|------|---------|
| `element.go` | Add `component Component` field to Element struct |
| `app.go` | Add `mounts *mountState` and `dispatchTable *dispatchTable` fields; add `walkComponents()` |
| `app_events.go` | Use dispatch table for key events instead of FocusManager |
| `app_loop.go` | Rebuild dispatch table after render on dirty frames |
| `app_render.go` | Call `mounts.sweep()` after render |
| `internal/tuigen/ast.go` | Add `Receiver`, `ReceiverName`, `ReceiverType` to Component; add `IsStructMount` to ComponentCall |
| `internal/tuigen/parser_component.go` | Parse optional receiver on `templ`; set `IsStructMount` flag |
| `internal/tuigen/generator.go` | Ensure non-templ Go code passes through |
| `internal/tuigen/generator_component.go` | Method receiver Render() generation; `tui.Mount()` code gen |
