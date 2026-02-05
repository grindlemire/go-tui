# Component Model & Broadcast Key System Specification

**Status:** Approved\
**Version:** 1.1\
**Last Updated:** 2026-02-04

---

## 1. Overview

### Purpose

Replace the current focus-based event dispatch system with a broadcast key bus built on an interface-based component model. Components are Go structs that implement capability interfaces. The framework discovers capabilities via type assertion and orchestrates key dispatch, lifecycle, and rendering automatically.

### Motivation

The current system routes key events through a FocusManager to the single focused element, then bubbles up the tree. This is the web browser model. It's a poor fit for TUIs because:

- **Keyboard is the only input device** — there's no mouse pointer to disambiguate which component the user is "targeting"
- **Multiple components legitimately need the same keys** — e.g., `q` to quit AND `q` as search input. Focus makes this an either/or choice
- **Focus is invisible and confusing** — users don't know which component "has focus," and component authors have to understand the global focus system to get their handlers to fire
- **Components can't be self-contained** — a component's key handling depends on whether it has focus, which depends on the rest of the app

The existing frameworks (Ratatui, Bubble Tea, Ink) each solve this differently:

- **Ratatui**: No event system at all — developer routes everything manually
- **Bubble Tea**: Single root `Update()` receives all messages, manually forwards to children. Parent must understand children's input needs
- **Ink**: Broadcast model — every `useInput` hook fires for every keystroke. No consumption mechanism, no ordering

None of them combine broadcast-by-default with stop-propagation and compile-time safety. This design does.

### Goals

- **Broadcast by default**: When a key is pressed, all registered handlers fire unless one stops propagation
- **Interface-based components**: Components are Go structs implementing `Component`, `KeyListener`, `Initializer` — discovered via type assertion
- **Self-contained components**: A component declares its key bindings without knowing about other components in the app
- **Single-file components**: Type definition, behavior methods, and template all live in one `.gsx` file
- **Template-driven composition**: `@Component(args)` in templates handles instantiation, caching, and lifecycle
- **Compile-time safety**: Interface satisfaction checked by `go build`; key conflicts caught at mount time
- **Minimal grammar changes**: One addition — optional method receiver on `templ`. Everything else is plain Go
- **Automatic reactivity**: `State.Set()` triggers re-render, `KeyMap()` re-evaluated on dirty, dispatch table rebuilt. No manual subscription management

### Non-Goals

- Full virtual DOM / reconciliation (simple position-keyed instance cache is sufficient)
- Per-component dirty tracking (global dirty flag + full tree re-render is fine for TUI scale)
- Event bubbling (replaced by broadcast + stop propagation)
- Two-way data binding
- Custom event types beyond keys and mouse

### Dependencies

| Dependency | Location | Description |
|------------|----------|-------------|
| `State[T]` | `state.go` | Reactive state with `Get()`/`Set()`/`Bind()` |
| `MarkDirty()` | `dirty.go` | Global atomic dirty flag |
| `Element` | `element.go` | UI element tree nodes |
| App main loop | `app_loop.go` | Frame-based event processing and rendering |

---

## 2. Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────┐
│  .gsx Source                                                      │
│                                                                    │
│  type sidebar struct { ... }                                       │
│  func Sidebar(query *tui.State[string]) *sidebar { ... }          │
│  func (s *sidebar) KeyMap() tui.KeyMap { ... }                     │
│  templ (s *sidebar) Render() { <div>...</div> }                    │
└────────────────────────┬─────────────────────────────────────────┘
                         │ tui generate
                         ▼
┌──────────────────────────────────────────────────────────────────┐
│  Generated Go (_gsx.go)                                           │
│                                                                    │
│  func (s *sidebar) Render() *tui.Element {                         │
│      return tui.New(                                               │
│          tui.WithChildren(                                         │
│              tui.Mount(s, 0, func() tui.Component {                │
│                  return ChildComponent(s.query)                    │
│              }),                                                    │
│          ),                                                        │
│      )                                                             │
│  }                                                                 │
└────────────────────────┬─────────────────────────────────────────┘
                         │ app.SetRoot(root)
                         ▼
┌──────────────────────────────────────────────────────────────────┐
│  Framework Runtime                                                │
│                                                                    │
│  1. Call root.Render() → element tree                              │
│  2. Walk tree, find Mount-tagged elements                          │
│  3. Type-assert components for KeyListener, Initializer            │
│  4. Collect KeyMap() from all KeyListeners                         │
│  5. Validate: no conflicting Stop handlers for same key            │
│  6. Build dispatch table: KeyPattern → []handler (tree order)      │
│  7. On key press: iterate handlers, call each, stop if Stop=true   │
└──────────────────────────────────────────────────────────────────┘
```

### Event Dispatch Flow

```
Terminal Input
    │
    ▼
parseInput() → KeyEvent
    │
    ▼
readInputEvents goroutine
    │
    ▼
eventQueue <- func() { dispatch(KeyEvent) }
    │
    ▼
Main Loop drains queue
    │
    ▼
Dispatch table lookup: KeyPattern → []handler
    │
    ├── Handler 1 (broadcast): call handler, continue
    ├── Handler 2 (broadcast): call handler, continue
    └── Handler 3 (stop): call handler, STOP
    │
    ▼
Handlers call State.Set() → MarkDirty()
    │
    ▼
Next frame: dirty=true
    │
    ├── root.Render() → new element tree
    ├── Walk tree → collect KeyMap() from all components
    ├── Validate exclusive conflicts
    └── Rebuild dispatch table
```

### Component Overview

| Component | Change |
|-----------|--------|
| `component.go` | NEW: `Component`, `KeyListener`, `Initializer` interfaces |
| `keymap.go` | NEW: `KeyMap`, `KeyBinding`, `KeyPattern` types; `OnKey`, `OnKeyStop`, `OnRunes`, `OnRunesStop` helpers |
| `mount.go` | NEW: `Mount()` function with instance caching and lifecycle management |
| `app.go` | MODIFY: `SetRoot` walks tree for component discovery; dispatch table replaces FocusManager for keys |
| `app_events.go` | MODIFY: Key dispatch uses broadcast dispatch table instead of FocusManager.Dispatch() |
| `app_loop.go` | MODIFY: Rebuild dispatch table when dirty |
| `focus.go` | DEPRECATE: FocusManager no longer used for key dispatch (may retain for visual focus via FocusGroup helper) |
| `internal/tuigen/generator.go` | MODIFY: Support `templ` with method receiver; generate `tui.Mount()` for `@Component()` calls |
| `internal/tuigen/parser.go` | MODIFY: Parse optional method receiver on `templ` declarations |

---

## 3. Core Entities

### 3.1 Component Interfaces

Three interfaces. Components implement the ones they need. The framework discovers capabilities via type assertion.

```go
// component.go

package tui

// Component is the base interface. Any struct with a Render() method
// can be used as a component in the element tree.
type Component interface {
    Render() *Element
}

// KeyListener is implemented by components that handle keyboard input.
// KeyMap() returns the current set of key bindings. It is called on every
// tree walk (when dirty), so it can return different bindings based on state.
type KeyListener interface {
    KeyMap() KeyMap
}

// Initializer is implemented by components that need setup when mounted.
// Init() is called once when the component first enters the tree.
// The returned function (if non-nil) is called when the component leaves
// the tree. This pairs setup and cleanup at the same call site.
type Initializer interface {
    Init() func()
}
```

Compile-time interface checks use standard Go pattern:

```go
var _ tui.KeyListener = (*sidebar)(nil)
var _ tui.Initializer = (*streamPanel)(nil)
```

### 3.2 KeyMap Types

KeyMap is data, not registration. The framework inspects and validates it.

```go
// keymap.go

package tui

// KeyMap is a list of key bindings returned by KeyListener.KeyMap().
// It is a value, not a registration — the framework collects and manages it.
type KeyMap []KeyBinding

// KeyBinding associates a key pattern with a handler.
type KeyBinding struct {
    Pattern KeyPattern
    Handler func(KeyEvent)
    Stop    bool // If true, prevent later handlers from firing for this key
}

// KeyPattern identifies which key events match a binding.
type KeyPattern struct {
    Key    Key      // Specific key (KeyCtrlB, KeyEscape, etc.), or 0
    Rune   rune     // Specific rune, or 0
    AnyRune bool    // Match any printable character
    Mods   Modifier // Required modifiers
}
```

Helper constructors for ergonomic KeyMap building:

```go
// keymap.go (continued)

// OnKey creates a broadcast binding for a specific key.
// Other handlers for the same key will also fire.
func OnKey(key Key, handler func(KeyEvent)) KeyBinding {
    return KeyBinding{
        Pattern: KeyPattern{Key: key},
        Handler: handler,
        Stop:    false,
    }
}

// OnKeyStop creates a stop-propagation binding for a specific key.
// No handlers registered after this one (in tree order) will fire.
func OnKeyStop(key Key, handler func(KeyEvent)) KeyBinding {
    return KeyBinding{
        Pattern: KeyPattern{Key: key},
        Handler: handler,
        Stop:    true,
    }
}

// OnRune creates a broadcast binding for a specific printable character.
func OnRune(r rune, handler func(KeyEvent)) KeyBinding {
    return KeyBinding{
        Pattern: KeyPattern{Rune: r},
        Handler: handler,
        Stop:    false,
    }
}

// OnRuneStop creates a stop-propagation binding for a specific printable character.
func OnRuneStop(r rune, handler func(KeyEvent)) KeyBinding {
    return KeyBinding{
        Pattern: KeyPattern{Rune: r},
        Handler: handler,
        Stop:    true,
    }
}

// OnRunes creates a broadcast binding for all printable characters.
func OnRunes(handler func(KeyEvent)) KeyBinding {
    return KeyBinding{
        Pattern: KeyPattern{AnyRune: true},
        Handler: handler,
        Stop:    false,
    }
}

// OnRunesStop creates a stop-propagation binding for all printable characters.
// Use this for text inputs that need exclusive access to character keys.
func OnRunesStop(handler func(KeyEvent)) KeyBinding {
    return KeyBinding{
        Pattern: KeyPattern{AnyRune: true},
        Handler: handler,
        Stop:    true,
    }
}
```

### 3.3 Mount and Instance Caching

Components are instantiated once (by their constructor) and cached across renders. The cache lives on the `App` struct (accessed via `currentApp` during render) for test isolation and proper scoping.

**Cache lifecycle uses mark-and-sweep:** Each render pass marks active mount keys. After render completes, sweep unmarked entries and call their cleanup functions. This handles conditional mounts — when a component is removed from the tree by an `@if`, its cache entry is swept and cleanup fires.

```go
// mount.go

package tui

// mountKey identifies a component instance by its parent and position.
type mountKey struct {
    parent Component
    index  int
}

// mountState is per-App state for component instance caching.
// Stored on the App struct, accessed via currentApp during render.
type mountState struct {
    cache      map[mountKey]Component
    cleanups   map[mountKey]func()
    activeKeys map[mountKey]bool // Marked during render, swept after
}

// Mount creates or retrieves a cached component instance and returns
// its rendered element tree. Called by generated code from @Component() syntax.
//
// On first call: executes factory, caches instance, calls Init() if Initializer.
// On subsequent calls: returns cached instance's Render() result.
// Mark-and-sweep: marks the key as active. Sweep after render cleans stale entries.
func Mount(parent Component, index int, factory func() Component) *Element {
    ms := currentApp.mounts
    key := mountKey{parent: parent, index: index}
    ms.activeKeys[key] = true // Mark as active this render

    instance, cached := ms.cache[key]
    if !cached {
        instance = factory()
        ms.cache[key] = instance

        // Call Init() if component implements Initializer
        if init, ok := instance.(Initializer); ok {
            cleanup := init.Init()
            if cleanup != nil {
                ms.cleanups[key] = cleanup
            }
        }
    }

    // Render the component and tag the element for framework discovery
    el := instance.Render()
    el.component = instance
    return el
}

// sweepMounts removes cached instances that were not marked active
// during the last render pass. Calls cleanup functions for removed components.
func (ms *mountState) sweep() {
    for key := range ms.cache {
        if !ms.activeKeys[key] {
            if cleanup, ok := ms.cleanups[key]; ok {
                cleanup()
                delete(ms.cleanups, key)
            }
            delete(ms.cache, key)
        }
    }
    // Reset active keys for next render
    ms.activeKeys = make(map[mountKey]bool)
}
```

### 3.4 Dispatch Table

Built from collected KeyMap data during tree walk. Rebuilt when dirty. **All handlers (exact key, exact rune, and AnyRune) are stored in a single list ordered by tree position.** This ensures one simple rule: tree order determines dispatch order.

```go
// dispatch.go

package tui

// dispatchEntry is a handler with its tree position for ordering.
type dispatchEntry struct {
    pattern  KeyPattern
    handler  func(KeyEvent)
    stop     bool
    position int // DFS order index from tree walk
}

// dispatchTable holds all handlers in a single tree-ordered list.
// Handlers are matched against incoming KeyEvents by pattern.
type dispatchTable struct {
    entries []dispatchEntry // All handlers, ordered by tree position
}

// buildDispatchTable walks the element tree, collects KeyMap() from
// all mounted components, validates exclusive conflicts, and builds
// the dispatch table ordered by tree position.
func buildDispatchTable(root *Element) (*dispatchTable, error) {
    table := &dispatchTable{}
    position := 0

    walkComponents(root, func(comp Component) {
        kl, ok := comp.(KeyListener)
        if !ok {
            return
        }
        km := kl.KeyMap()
        if km == nil {
            return
        }
        for _, binding := range km {
            table.entries = append(table.entries, dispatchEntry{
                pattern:  binding.Pattern,
                handler:  binding.Handler,
                stop:     binding.Stop,
                position: position,
            })
        }
        position++
    })

    // Validate: no two active Stop handlers for the same pattern
    if err := table.validate(); err != nil {
        return nil, err
    }

    return table, nil
}

// matches checks if a dispatch entry matches a key event.
func (e *dispatchEntry) matches(ke KeyEvent) bool {
    p := e.pattern
    if p.AnyRune && ke.Key == KeyRune {
        return true
    }
    if p.Rune != 0 && ke.Rune == p.Rune && ke.Key == KeyRune {
        return true
    }
    if p.Key != 0 && ke.Key == p.Key {
        return true
    }
    return false
}

// dispatch sends a key event to all matching handlers in tree order.
// Stops early if a matching handler has Stop=true.
func (dt *dispatchTable) dispatch(ke KeyEvent) {
    for i := range dt.entries {
        if dt.entries[i].matches(ke) {
            dt.entries[i].handler(ke)
            if dt.entries[i].stop {
                return
            }
        }
    }
}
```

---

## 4. Component Pattern

### 4.1 Single-File Component

Everything for a component lives in one `.gsx` file: type definition, constructor, behavior methods, and template.

```gsx
// sidebar.gsx
package myapp

import tui "github.com/grindlemire/go-tui"

// Unexported type — implementation detail
type sidebar struct {
    Query    *tui.State[string]  // Passed in by parent
    expanded *tui.State[bool]    // Internal state
}

// Exported constructor — this IS the component's public API.
// @Sidebar(args) in a parent template calls this function.
func Sidebar(query *tui.State[string]) *sidebar {
    return &sidebar{
        Query:    query,
        expanded: tui.NewState(true),
    }
}

// Compile-time interface check
var _ tui.KeyListener = (*sidebar)(nil)

// KeyMap returns current key bindings. Called by framework on each tree walk.
func (s *sidebar) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyCtrlB, s.toggle),
    }
}

func (s *sidebar) toggle(ke tui.KeyEvent) {
    s.expanded.Set(!s.expanded.Get())
}

// Template — only this block is transformed by the compiler.
// Everything else passes through to generated Go verbatim.
templ (s *sidebar) Render() {
    @if s.expanded.Get() {
        <div class="border-single p-1">
            <span class="font-bold">Results for: {s.Query.Get()}</span>
        </div>
    }
}
```

### 4.2 Constructor Function Pattern

The exported function name IS the component name. The unexported struct is the implementation. This follows Go conventions (`errors.New()` returns `*errorString`).

```go
// Exported constructor: the public API
func Sidebar(query *tui.State[string]) *sidebar { ... }

// Unexported struct: implementation detail
type sidebar struct { ... }
```

- Parent templates use `@Sidebar(args)` — calls the constructor directly
- The framework uses interface type assertions (`KeyListener`, `Component`) — never the concrete type
- Internal state (unexported fields) is created by the constructor, invisible to parents
- Shared state (`*tui.State[T]` pointers) is passed in as constructor arguments

### 4.3 Conditional Key Activation

`KeyMap()` is a method that returns data. It uses normal Go control flow for conditional bindings:

```go
func (s *searchInput) KeyMap() tui.KeyMap {
    if !s.Active.Get() {
        return nil // No bindings when inactive
    }
    return tui.KeyMap{
        tui.OnRunesStop(s.appendChar),
        tui.OnKeyStop(tui.KeyBackspace, s.deleteChar),
        tui.OnKeyStop(tui.KeyEnter, s.submit),
        tui.OnKeyStop(tui.KeyEscape, s.deactivate),
    }
}
```

When `Active` is false, `KeyMap()` returns nil — no bindings registered. When `Active` becomes true, the state change marks dirty, the tree is re-walked, `KeyMap()` is called again, and now it returns Stop bindings. The dispatch table is rebuilt automatically.

No special "conditional exclusive" concept. No activation flags. Just an `if` statement.

### 4.4 Lifecycle with Init

Components that need setup (goroutines, timers, connections) implement `Initializer`. The returned cleanup function is called when the component leaves the tree.

```go
func (s *streamPanel) Init() func() {
    ctx, cancel := context.WithCancel(context.Background())
    go s.pollData(ctx)
    return cancel
}

func (s *clock) Init() func() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            s.Time.Set(time.Now())
        }
    }()
    return ticker.Stop
}
```

Setup and cleanup are paired at the same call site — they can't get out of sync.

---

## 5. GSX Syntax

### 5.1 Method Receiver on templ (Grammar Addition)

One grammar addition: `templ` declarations accept an optional Go method receiver.

**Before:**
```gsx
templ Header(title string) {
    <div><span>{title}</span></div>
}
```

**After (also supported):**
```gsx
templ (s *sidebar) Render() {
    <div><span>{s.Query.Get()}</span></div>
}
```

This matches Go's method declaration syntax exactly. The parser handles an optional parenthesized receiver before the function name.

Both forms coexist:
- `templ Name(params)` — function component (stateless, current behavior)
- `templ (recv) Name()` — method component (stateful struct)

### 5.2 Component Mounting with @

`@Component(args)` mounts a struct component. The compiler generates `tui.Mount()` calls.

```gsx
templ (a *myApp) Render() {
    <div class="flex">
        @Sidebar(a.query)
        <div class="flex-col flex-grow-1">
            @SearchInput(a.searchActive, a.query)
            @ContentPanel(a.query)
        </div>
    </div>
}
```

The `@` prefix is already used for control flow (`@if`, `@for`, `@let`). The rule:
- `@keyword` (if, for, let) → control flow
- `@Expression(args)` (starts with uppercase identifier, has parens) → component mount
- `<Name ... />` → function component call (existing behavior, unchanged)

This gives three visually distinct forms in templates:
```gsx
@Sidebar(a.query)               // struct component (stateful, has KeyMap/Init)
<Header title="Welcome" />      // function component (stateless presentation)
@if a.showFooter.Get() { ... }  // control flow
```

---

## 6. Generated Output Examples

### 6.1 Single-File Component

**Input:**

```gsx
// search.gsx
package myapp

import tui "github.com/grindlemire/go-tui"

type searchInput struct {
    Active *tui.State[bool]
    Query  *tui.State[string]
}

func SearchInput(active *tui.State[bool], query *tui.State[string]) *searchInput {
    return &searchInput{Active: active, Query: query}
}

var _ tui.KeyListener = (*searchInput)(nil)

func (s *searchInput) KeyMap() tui.KeyMap {
    if !s.Active.Get() {
        return nil
    }
    return tui.KeyMap{
        tui.OnRunesStop(s.appendChar),
        tui.OnKeyStop(tui.KeyBackspace, s.deleteChar),
        tui.OnKeyStop(tui.KeyEscape, s.deactivate),
    }
}

func (s *searchInput) appendChar(ke tui.KeyEvent) {
    s.Query.Set(s.Query.Get() + string(ke.Rune))
}

func (s *searchInput) deleteChar(ke tui.KeyEvent) {
    q := s.Query.Get()
    if len(q) > 0 {
        s.Query.Set(q[:len(q)-1])
    }
}

func (s *searchInput) deactivate(ke tui.KeyEvent) {
    s.Active.Set(false)
    s.Query.Set("")
}

templ (s *searchInput) Render() {
    @if s.Active.Get() {
        <div class="border-rounded p-1">
            <span class="text-cyan">Search: </span>
            <span>{s.Query.Get()}</span>
            <span class="font-dim">|</span>
        </div>
    }
}
```

**Output (`search_gsx.go`):**

```go
package myapp

import tui "github.com/grindlemire/go-tui"

type searchInput struct {
    Active *tui.State[bool]
    Query  *tui.State[string]
}

func SearchInput(active *tui.State[bool], query *tui.State[string]) *searchInput {
    return &searchInput{Active: active, Query: query}
}

var _ tui.KeyListener = (*searchInput)(nil)

func (s *searchInput) KeyMap() tui.KeyMap {
    if !s.Active.Get() {
        return nil
    }
    return tui.KeyMap{
        tui.OnRunesStop(s.appendChar),
        tui.OnKeyStop(tui.KeyBackspace, s.deleteChar),
        tui.OnKeyStop(tui.KeyEscape, s.deactivate),
    }
}

func (s *searchInput) appendChar(ke tui.KeyEvent) {
    s.Query.Set(s.Query.Get() + string(ke.Rune))
}

func (s *searchInput) deleteChar(ke tui.KeyEvent) {
    q := s.Query.Get()
    if len(q) > 0 {
        s.Query.Set(q[:len(q)-1])
    }
}

func (s *searchInput) deactivate(ke tui.KeyEvent) {
    s.Active.Set(false)
    s.Query.Set("")
}

// Only this method is generated from the templ block:
func (s *searchInput) Render() *tui.Element {
    var children []*tui.Element
    if s.Active.Get() {
        children = append(children, tui.New(
            tui.WithBorder(tui.BorderRounded),
            tui.WithPadding(tui.EdgeAll(1)),
            tui.WithChildren(
                tui.New(tui.WithText("Search: "), tui.WithTextStyle(tui.Style{}.Fg(tui.Cyan))),
                tui.New(tui.WithText(s.Query.Get())),
                tui.New(tui.WithText("|"), tui.WithTextStyle(tui.Style{}.Dim())),
            ),
        ))
    }
    return tui.New(tui.WithChildren(children...))
}
```

Everything except the `Render()` method passes through verbatim. The `templ` block is the only thing the compiler transforms.

### 6.2 Parent Component with @Mount Syntax

**Input:**

```gsx
// app.gsx
package myapp

import tui "github.com/grindlemire/go-tui"

type myApp struct {
    searchActive *tui.State[bool]
    query        *tui.State[string]
}

func MyApp() *myApp {
    return &myApp{
        searchActive: tui.NewState(false),
        query:        tui.NewState(""),
    }
}

var _ tui.KeyListener = (*myApp)(nil)

// Conditional KeyMap: '/' only binds when search is not active.
// When search IS active, searchInput's OnRunesStop captures all runes
// (including '/') via unified tree-order dispatch.
func (a *myApp) KeyMap() tui.KeyMap {
    km := tui.KeyMap{
        tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { tui.Quit() }),
    }
    if !a.searchActive.Get() {
        km = append(km, tui.OnRune('/', func(ke tui.KeyEvent) {
            a.searchActive.Set(true)
        }))
    }
    return km
}

templ (a *myApp) Render() {
    <div class="flex">
        @Sidebar(a.query)
        <div class="flex-col flex-grow-1">
            @SearchInput(a.searchActive, a.query)
            @ContentPanel(a.query)
        </div>
    </div>
}
```

**Output (`app_gsx.go`):**

```go
package myapp

import tui "github.com/grindlemire/go-tui"

type myApp struct {
    searchActive *tui.State[bool]
    query        *tui.State[string]
}

func MyApp() *myApp {
    return &myApp{
        searchActive: tui.NewState(false),
        query:        tui.NewState(""),
    }
}

var _ tui.KeyListener = (*myApp)(nil)

func (a *myApp) KeyMap() tui.KeyMap {
    km := tui.KeyMap{
        tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { tui.Quit() }),
    }
    if !a.searchActive.Get() {
        km = append(km, tui.OnRune('/', func(ke tui.KeyEvent) {
            a.searchActive.Set(true)
        }))
    }
    return km
}

// Generated from templ block. @Component(args) becomes tui.Mount().
func (a *myApp) Render() *tui.Element {
    return tui.New(
        tui.WithDirection(tui.Row),
        tui.WithChildren(
            tui.Mount(a, 0, func() tui.Component {
                return Sidebar(a.query)
            }),
            tui.New(
                tui.WithDirection(tui.Column),
                tui.WithFlexGrow(1),
                tui.WithChildren(
                    tui.Mount(a, 1, func() tui.Component {
                        return SearchInput(a.searchActive, a.query)
                    }),
                    tui.Mount(a, 2, func() tui.Component {
                        return ContentPanel(a.query)
                    }),
                ),
            ),
        ),
    )
}
```

### 6.3 Entry Point

```go
// main.go
package main

func main() {
    app := tui.NewApp(tui.WithRoot(MyApp()))
    app.Run()
}
```

---

## 7. Reactivity Cycle

### 7.1 How State Changes Propagate

The reactivity model is simple: `State.Set()` → global dirty flag → full tree re-render → re-read via `Get()`.

1. User presses `/`
2. KeyEvent enters dispatch table
3. `myApp.KeyMap()` handler fires: `a.searchActive.Set(true)`
4. `Set()` internally calls `MarkDirty()`
5. Next frame: `checkAndClearDirty()` returns true
6. Framework calls `root.Render()`
7. `root.Render()` calls `tui.Mount()` for each child — returns cached instances' `Render()` results
8. `searchInput.Render()` calls `s.Active.Get()` — now returns `true` — renders the input UI
9. Framework walks new element tree, finds components, calls `KeyMap()` on each
10. `searchInput.KeyMap()` now returns Stop bindings for runes, backspace, escape
11. Dispatch table is rebuilt with new bindings
12. Future rune keypresses go exclusively to searchInput (Stop handlers prevent broadcast)

### 7.2 How Dispatch Table Rebuilds

The dispatch table is rebuilt from scratch on every dirty frame. This is intentional:

- `KeyMap()` is called on every component during tree walk
- `KeyMap()` can return different bindings based on current state (e.g., nil when inactive)
- No registration/unregistration bookkeeping
- No stale handlers from unmounted components
- Cost is a tree walk per state change — acceptable for TUI-scale trees

### 7.3 What Parents Don't Do

Parents never explicitly update child state. The parent passes a `*tui.State[T]` pointer to the child. When anyone calls `.Set()` on that state value (the parent, the child, or another component entirely), the global dirty flag is set and the entire tree re-renders. The child's `Render()` calls `.Get()` and sees the new value.

No subscriptions. No event propagation. No "notify children." The shared pointer and the dirty flag handle everything.

---

## 8. Coordination Between Components

### 8.1 Shared State

Components share `*tui.State[T]` values passed through constructors:

```go
func MyApp() *myApp {
    searchActive := tui.NewState(false)
    query := tui.NewState("")
    return &myApp{
        searchActive: searchActive,
        query:        query,
    }
}
```

In the template, these are passed to children:

```gsx
@Sidebar(a.query)
@SearchInput(a.searchActive, a.query)
```

When `SearchInput` calls `s.Query.Set("new text")`, `Sidebar` sees the new value on next render via `s.Query.Get()`. The parent did nothing.

### 8.2 Callbacks

For "do something the parent controls" patterns, use function fields:

```go
type deleteButton struct {
    OnConfirm func()
}

func DeleteButton(onConfirm func()) *deleteButton {
    return &deleteButton{OnConfirm: onConfirm}
}

func (d *deleteButton) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKeyStop(tui.KeyEnter, func(ke tui.KeyEvent) {
            d.OnConfirm()
        }),
    }
}
```

The parent decides what "confirm" means. The button doesn't know or care.

### 8.3 App-Level Operations

`tui.Quit()` exists as a package-level function for truly global operations. Everything else is wired through constructor arguments (shared state or callbacks).

### 8.4 Focus Groups (Helper, Not Interface)

Focus is visual state + Tab cycling, not an event dispatch concept. A `FocusGroup` helper manages the common case:

```go
type FocusGroup struct {
    members []*tui.State[bool]
    current int
}

func NewFocusGroup(members ...*tui.State[bool]) *FocusGroup { ... }
func (fg *FocusGroup) Next() { /* deactivate current, activate next */ }
func (fg *FocusGroup) Prev() { /* deactivate current, activate prev */ }

// FocusGroup itself implements KeyListener
func (fg *FocusGroup) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) { fg.Next() }),
        tui.OnKey(tui.KeyShiftTab, func(ke tui.KeyEvent) { fg.Prev() }),
    }
}
```

Components that participate in a focus group use a `*tui.State[bool]` field for their selected state and check it in `Render()` for visual highlighting. The focus group toggles these states. No framework-level focus concept needed.

---

## 9. Ordering and Stop Propagation

### 9.1 Tree Order (Unified)

**All handlers — exact key, exact rune, and AnyRune — are stored in a single list ordered by DFS tree position.** When a key fires, the framework iterates the list, checks each entry for a pattern match, and calls matching handlers in order. There is no separate priority for exact vs AnyRune matches; tree position is the only ordering axis.

A component author doesn't choose their ordering — their position in the component tree determines it.

### 9.2 Why Order Usually Doesn't Matter

For broadcast handlers (no Stop), all fire regardless of order. If the three handlers for `ctrl+b` toggle a sidebar, log an event, and update a status indicator — these are independent side effects. Order between them is irrelevant.

**If order between broadcast handlers matters, that's a signal the handlers aren't truly independent.** The fix is shared state coordination or Stop propagation, not ordering control.

### 9.3 When Order Matters: Stop Propagation

Stop handlers capture a key. Tree position determines priority:
- Components higher in the tree (parents, earlier siblings) run first
- A parent can intercept before children
- A "stop" handler prevents later handlers from firing

For modal/overlay patterns where you want something to intercept keys regardless of tree position, mount the modal high in the tree (at the root level). This is natural — modals are typically rendered at the top of the component hierarchy.

### 9.4 Conflict Validation

At mount time (`SetRoot`), the framework validates the dispatch table:

- Two Stop handlers for the same key pattern where both are active = error returned from `buildDispatchTable` with a clear message naming both components
- This catches conflicts before any events are processed
- For conditional Stop handlers (where `KeyMap()` returns nil when inactive), the validation only checks currently active bindings
- If two components could theoretically both be active with conflicting Stop handlers, the developer must ensure mutual exclusion via shared state

---

## 10. Rules and Constraints

1. **Components are Go structs** implementing `Component` (Render). Optional: `KeyListener` (KeyMap), `Initializer` (Init)
2. **Constructor is the component name** — `func Sidebar(args) *sidebar`. `@Sidebar(args)` calls this function
3. **Struct is unexported, constructor is exported** — follows Go convention (`errors.New` → `*errorString`)
4. **KeyMap returns data, not registration** — framework collects and validates it
5. **KeyMap is re-evaluated every dirty frame** — can return different bindings based on state
6. **Broadcast by default** — handlers without Stop all fire for matching keys
7. **Stop propagation prevents later handlers** — in tree order, not just children
8. **No bool returns from handlers** — mutations call `MarkDirty()` automatically
9. **Init returns cleanup** — paired at the call site, called by framework on unmount
10. **Mount caches instances** — constructor called once, `Render()` called every frame
11. **Shared state via `*tui.State[T]` pointers** — not copied, shared by reference
12. **One grammar addition** — optional method receiver on `templ`
13. **Everything else is plain Go** — type definitions, methods, constructors all pass through verbatim

---

## 11. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Large\
**Recommended Phases:** 6

### Phase Breakdown

1. **Phase 1: Core Interfaces and KeyMap Types**
   - Create `Component`, `KeyListener`, `Initializer` interfaces
   - Create `KeyMap`, `KeyBinding`, `KeyPattern` types
   - Create helper constructors (`OnKey`, `OnKeyStop`, `OnRunes`, `OnRunesStop`)
   - Unit tests for KeyMap construction and pattern matching

2. **Phase 2: Mount and Instance Caching**
   - Implement `Mount()` with position-keyed cache
   - Implement Init/cleanup lifecycle
   - Implement cache invalidation on tree changes
   - Unit tests for mount/unmount lifecycle

3. **Phase 3: Dispatch Table and Key Broadcast**
   - Implement `buildDispatchTable()` with tree walk
   - Implement `dispatch()` with broadcast + stop propagation
   - Implement conflict validation
   - Replace FocusManager key dispatch in `app_events.go`
   - Integrate dispatch table rebuild into dirty frame cycle
   - Unit tests for broadcast, stop, conflict detection

4. **Phase 4: Parser — Method Receiver on templ**
   - Extend `templ` parsing to accept optional `(receiver)` before name
   - Parse `@Component(args)` as component mount (vs control flow)
   - Unit tests for new parse paths

5. **Phase 5: Generator — Mount Code Generation**
   - Generate method-receiver `Render()` from `templ (recv) Name()`
   - Generate `tui.Mount(parent, index, factory)` from `@Component(args)`
   - Assign position indices to mount calls
   - Pass through type definitions, methods, constructors verbatim
   - Integration tests with example components

6. **Phase 6: Migration and Examples**
   - Migrate existing examples to component model
   - Add FocusGroup helper
   - Update documentation
   - End-to-end tests with multi-component apps

---

## 12. Success Criteria

1. Components are Go structs implementing `Component` interface with `Render() *Element`
2. `KeyListener.KeyMap()` returns `KeyMap` data that the framework collects
3. When a key is pressed, all matching broadcast handlers fire in tree order
4. A Stop handler prevents subsequent handlers from firing for that key
5. Two active Stop handlers for the same key pattern cause a panic at mount time
6. `KeyMap()` is called on every dirty frame and can return different bindings based on state
7. `Init()` is called once on first mount; returned cleanup function is called on unmount
8. `Mount()` caches instances by (parent, index) and reuses across renders
9. `State.Set()` marks dirty, triggering full re-render and dispatch table rebuild
10. Child components automatically see updated state via shared `*tui.State[T]` pointers
11. Parents never explicitly update child state — shared pointers and dirty flag handle it
12. `templ (recv) Name()` generates a Go method with the given receiver
13. `@Component(args)` generates `tui.Mount(parent, index, func() Component { return Component(args) })`
14. `.gsx` files support type definitions, methods, and constructors alongside `templ` blocks
15. All non-templ Go code passes through to generated output verbatim
16. FocusGroup helper handles Tab/Shift+Tab cycling without framework-level focus
17. Example multi-component app works: key broadcast, stop propagation, shared state, conditional activation

---

## 13. Relationship to Other Features

### Existing State[T] System

The component model builds directly on `State[T]`. Components receive shared state via constructor arguments. State changes trigger the same dirty/re-render cycle. No changes to `State[T]` needed.

### Existing Element System

Components produce `*Element` trees from their `Render()` methods. The element system is unchanged — it remains the rendering primitive. Components add structure and behavior on top.

### Existing Watcher System

Components that need channel/timer watching implement `Initializer`. The `Init()` method starts watchers and returns cleanup. This replaces the current watcher aggregation pattern with a simpler per-component lifecycle.

### Current Focus System (Deprecated for Keys)

The current `FocusManager` is no longer used for key dispatch. It may be retained for internal use or removed entirely. The `FocusGroup` helper replaces its Tab-cycling functionality with a simpler, state-based approach.

### Mouse Events

Mouse handling stays on elements (`onClick`, hit-testing). Mouse events are spatial — they target the element under the cursor. Only keyboard events use the broadcast dispatch system.

---

## 14. End-to-End Example

### Complete App with Multiple Components

```gsx
// sidebar.gsx
package myapp

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type sidebar struct {
    Query    *tui.State[string]
    expanded *tui.State[bool]
}

func Sidebar(query *tui.State[string]) *sidebar {
    return &sidebar{
        Query:    query,
        expanded: tui.NewState(true),
    }
}

var _ tui.KeyListener = (*sidebar)(nil)

func (s *sidebar) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyCtrlB, s.toggle),
    }
}

func (s *sidebar) toggle(ke tui.KeyEvent) {
    s.expanded.Set(!s.expanded.Get())
}

templ (s *sidebar) Render() {
    @if s.expanded.Get() {
        <div class="flex-col border-single p-1" width={30}>
            <span class="font-bold">Sidebar</span>
            @if s.Query.Get() != "" {
                <span>Filtering: {s.Query.Get()}</span>
            }
        </div>
    }
}
```

```gsx
// search.gsx
package myapp

import tui "github.com/grindlemire/go-tui"

type searchInput struct {
    Active *tui.State[bool]
    Query  *tui.State[string]
}

func SearchInput(active *tui.State[bool], query *tui.State[string]) *searchInput {
    return &searchInput{Active: active, Query: query}
}

var _ tui.KeyListener = (*searchInput)(nil)

func (s *searchInput) KeyMap() tui.KeyMap {
    if !s.Active.Get() {
        return nil
    }
    return tui.KeyMap{
        tui.OnRunesStop(s.appendChar),
        tui.OnKeyStop(tui.KeyBackspace, s.deleteChar),
        tui.OnKeyStop(tui.KeyEnter, s.submit),
        tui.OnKeyStop(tui.KeyEscape, s.deactivate),
    }
}

func (s *searchInput) appendChar(ke tui.KeyEvent) {
    s.Query.Set(s.Query.Get() + string(ke.Rune))
}

func (s *searchInput) deleteChar(ke tui.KeyEvent) {
    q := s.Query.Get()
    if len(q) > 0 {
        s.Query.Set(q[:len(q)-1])
    }
}

func (s *searchInput) submit(ke tui.KeyEvent) {
    s.Active.Set(false)
}

func (s *searchInput) deactivate(ke tui.KeyEvent) {
    s.Active.Set(false)
    s.Query.Set("")
}

templ (s *searchInput) Render() {
    @if s.Active.Get() {
        <div class="border-rounded p-1">
            <span class="text-cyan">Search: </span>
            <span>{s.Query.Get()}</span>
            <span class="font-dim">|</span>
        </div>
    }
}
```

```gsx
// app.gsx
package myapp

import tui "github.com/grindlemire/go-tui"

type myApp struct {
    searchActive *tui.State[bool]
    query        *tui.State[string]
}

func MyApp() *myApp {
    return &myApp{
        searchActive: tui.NewState(false),
        query:        tui.NewState(""),
    }
}

var _ tui.KeyListener = (*myApp)(nil)

// Conditional KeyMap: '/' only binds when search is inactive.
func (a *myApp) KeyMap() tui.KeyMap {
    km := tui.KeyMap{
        tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { tui.Quit() }),
    }
    if !a.searchActive.Get() {
        km = append(km, tui.OnRune('/', func(ke tui.KeyEvent) {
            a.searchActive.Set(true)
        }))
    }
    return km
}

templ (a *myApp) Render() {
    <div class="flex">
        @Sidebar(a.query)
        <div class="flex-col flex-grow-1">
            @SearchInput(a.searchActive, a.query)
            <div class="flex-grow-1 p-1">
                <span>Main content area</span>
            </div>
        </div>
    </div>
}
```

```go
// main.go
package main

func main() {
    app := tui.NewApp(tui.WithRoot(myapp.MyApp()))
    app.Run()
}
```

### What Happens at Runtime

1. `MyApp()` constructor creates shared state: `searchActive=false`, `query=""`
2. `app.SetRoot(myApp)` triggers first render
3. `myApp.Render()` → mounts Sidebar and SearchInput (first mount: creates instances, calls Init if any)
4. Tree walk collects KeyMap: myApp returns `[ctrl+c, /]`, Sidebar returns `[ctrl+b]`, SearchInput returns `nil` (inactive)
5. Dispatch table (unified tree order): `ctrl+c → myApp, / → myApp, ctrl+b → sidebar`
6. User presses `/` → myApp handler fires → `searchActive.Set(true)` → dirty
7. Re-render: SearchInput now renders the input UI
8. Tree walk: myApp.KeyMap() returns `[ctrl+c]` only (no `/` — searchActive is true). SearchInput.KeyMap() returns `[runes(stop), backspace(stop), enter(stop), escape(stop)]`
9. Dispatch table rebuilt (unified tree order): myApp has ctrl+c, sidebar has ctrl+b, searchInput has runes(stop)+keys(stop). No `/` handler exists — all rune keys go exclusively to searchInput via AnyRune stop
10. User types "hello" → each rune goes to searchInput exclusively
11. User presses Escape → searchInput.deactivate → `active=false, query=""` → dirty
12. Re-render: SearchInput returns nil. KeyMap returns nil. myApp.KeyMap() returns `[ctrl+c, /]` again. Dispatch table reverts
13. `/` key goes to myApp again — the cycle restarts
