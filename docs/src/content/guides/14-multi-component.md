# Multi-Component Applications

## Overview

Real applications are rarely a single component. A file explorer needs a sidebar, a content panel, a search bar. A chat app needs a message list and an input widget, at minimum. Each piece has its own state and behavior, but they need to coordinate.

go-tui handles this through shared state and the mount system. Parent components create `State[T]` values and pass them to children through constructors. When any component changes a shared state value, every component that reads it sees the update on the next render. The mount system caches child component instances across renders, so each child keeps its own local state even when the parent re-renders.

This guide walks through splitting code across files, sharing state between components, conditional KeyMaps, and architecture patterns for larger apps.

## Multiple .gsx Files

Each component can live in its own `.gsx` file. All files in the same package share scope, so components defined in one file are callable from another without imports:

```
myapp/
  main.go
  app.gsx          -> app_gsx.go
  sidebar.gsx      -> sidebar_gsx.go
  search.gsx       -> search_gsx.go   (search bar + content panel)
```

Run `tui generate ./...` and it processes every `.gsx` file in the package, producing a corresponding `_gsx.go` file for each. The generated files are standard Go, so all the components end up in the same package and can reference each other directly.

There's no strict rule about how many components go in one file. A small pure component that's only used alongside its parent can share a file. A struct component with its own state and key handling usually deserves its own file. Split when it makes the code easier to navigate.

## Shared State

The main coordination mechanism between components is shared `*tui.State[T]`. The parent creates a state value and passes the pointer to each child that needs it. Because they all hold the same pointer, a `.Set()` or `.Update()` from any component marks the UI dirty and every component sees the new value on the next render.

Here's the pattern. A root component creates a `category` state and hands it to both a sidebar and a content panel:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

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
        tui.On(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

templ (a *myApp) Render() {
    <div class="flex h-full">
        @Sidebar(a.category)
        @Content(a.category)
    </div>
}
```

The sidebar receives the shared state and also creates its own local `selected` state for tracking which list item has the cursor:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

var categories = []string{"Documents", "Images", "Music", "Projects", "Downloads"}

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
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) {
            s.selected.Update(func(v int) int {
                if v >= len(categories)-1 {
                    return 0
                }
                return v + 1
            })
            s.category.Set(categories[s.selected.Get()])
        }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) {
            s.selected.Update(func(v int) int {
                if v <= 0 {
                    return len(categories) - 1
                }
                return v - 1
            })
            s.category.Set(categories[s.selected.Get()])
        }),
    }
}

templ (s *sidebar) Render() {
    <div class="flex-col border-single shrink-0 px-1" width={20}>
        <span class="text-gradient-cyan-magenta font-bold">Folders</span>
        <hr />
        for i, cat := range categories {
            if i == s.selected.Get() {
                <span class="text-cyan font-bold">{"> " + cat}</span>
            } else {
                <span class="font-dim">{"  " + cat}</span>
            }
        }
    </div>
}
```

The content panel reads the same `category` state to display the right files:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

var filesByCategory = map[string][]string{
    "Documents": {"report.pdf", "notes.md", "budget.xlsx", "readme.txt"},
    "Images":    {"photo.jpg", "screenshot.png", "logo.svg", "banner.gif"},
    "Music":     {"song.mp3", "album.flac", "podcast.ogg"},
    "Projects":  {"go-tui/", "website/", "api-server/"},
    "Downloads": {"setup.exe", "archive.zip", "data.csv"},
}

type content struct {
    category *tui.State[string]
}

func Content(category *tui.State[string]) *content {
    return &content{category: category}
}

templ (c *content) Render() {
    <div class="flex-col grow px-2">
        <span class="font-bold text-cyan">{c.category.Get() + "/"}</span>
        <hr />
        files := filesByCategory[c.category.Get()]
        for i, file := range files {
            if i == len(files)-1 {
                <span>{fmt.Sprintf("└── %s", file)}</span>
            } else {
                <span>{fmt.Sprintf("├── %s", file)}</span>
            }
        }
    </div>
}
```

When the sidebar calls `s.category.Set(categories[s.selected.Get()])`, the content panel picks it up through `c.category.Get()` on the next render. Neither component knows about the other. They just read and write the same state.

### What Gets Shared vs. What Stays Local

A good rule of thumb: share state that represents data two or more components need to agree on. Keep state local when it's purely about that component's internal behavior.

In the example above, `category` is shared because both the sidebar and the content panel need it. But `selected` is local to the sidebar because only the sidebar cares about which list item has the cursor highlight.

## Conditional KeyMaps

`KeyMap()` is called on every render pass, so you can return different bindings depending on component state. This is how you implement modes: a search bar that captures all keystrokes when active, a sidebar that ignores navigation keys when collapsed, or a vim-style editor with distinct modes.

Here's a search bar that returns `nil` (no bindings) when inactive, and captures all runes when active:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

type searchBar struct {
    active *tui.State[bool]
    query  *tui.State[string]
}

func SearchBar(active *tui.State[bool], query *tui.State[string]) *searchBar {
    return &searchBar{active: active, query: query}
}

func (s *searchBar) KeyMap() tui.KeyMap {
    if !s.active.Get() {
        return nil
    }
    return tui.KeyMap{
        tui.OnStop(tui.AnyRune,s.appendChar),
        tui.OnStop(tui.KeyBackspace, s.deleteChar),
        tui.OnStop(tui.KeyEnter, s.submit),
        tui.OnStop(tui.KeyEscape, s.deactivate),
    }
}

func (s *searchBar) appendChar(ke tui.KeyEvent) {
    s.query.Set(s.query.Get() + string(ke.Rune))
}

func (s *searchBar) deleteChar(ke tui.KeyEvent) {
    q := s.query.Get()
    if len(q) > 0 {
        s.query.Set(q[:len(q)-1])
    }
}

func (s *searchBar) submit(ke tui.KeyEvent) {
    s.active.Set(false)
}

func (s *searchBar) deactivate(ke tui.KeyEvent) {
    s.active.Set(false)
    s.query.Set("")
}

templ (s *searchBar) Render() {
    <div class="shrink-0">
        if s.active.Get() {
            <hr />
            <div class="px-1 flex gap-1">
                <span class="text-cyan font-bold">Search:</span>
                <span>{s.query.Get()}</span>
                <span class="text-cyan blink">_</span>
            </div>
        }
    </div>
}
```

The `OnStop` variants prevent the event from reaching other components. When the search bar is active, pressing `j` types a `j` into the search field instead of moving the sidebar cursor. When it's inactive, it returns `nil` and stays out of the way entirely.

The root component can also have conditional bindings. For example, only registering `/` to open search and `q` to quit when search is not active:

```go
func (a *myApp) KeyMap() tui.KeyMap {
    km := tui.KeyMap{
        tui.On(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
    if !a.searchActive.Get() {
        km = append(km, tui.On(tui.Rune('/'), func(ke tui.KeyEvent) {
            a.searchActive.Set(true)
        }))
        km = append(km, tui.On(tui.Rune('q'), func(ke tui.KeyEvent) {
            ke.App().Stop()
        }))
    }
    return km
}
```

This keeps the `q` key from quitting the app while you're typing a search query.

## Event Dispatch Order

The framework collects key bindings from all mounted components by walking the element tree in breadth-first order. Components closer to the root of the tree appear earlier in the dispatch list. For a given key event, every matching handler fires in tree order unless a `Stop` handler consumes the event first.

For the file explorer layout:

```
myApp (root)
├── Sidebar
├── Content
└── SearchBar
```

The dispatch order is: `myApp` first, then `Sidebar`, then `Content`, then `SearchBar`. If the search bar registers `OnStop(tui.AnyRune, ...)` for all runes, those handlers run after the root and sidebar handlers for the same key. But because the search bar's handler has `Stop=true`, it prevents any subsequent handlers from firing. Since the search bar is last in tree order for this layout, the stop flag mainly serves as documentation of intent and future-proofing.

If you need a child component to intercept a key before its parent sees it, restructure the tree so the intercepting component appears first in BFS order, or have the parent check state before registering its bindings (the conditional KeyMap pattern above).

Two components cannot both register `Stop` handlers for the same key pattern at the same time. The framework validates this and returns an error. This catches bugs early: if two components both claim exclusive ownership of `Escape`, something needs to change.

## Sub-Component Mounting

When a struct component renders another struct component using `@Component(args)` syntax, the generated code calls `app.Mount()`:

```go
// Generated from: @Sidebar(a.category)
app.Mount(a, 0, func() tui.Component {
    return Sidebar(a.category)
})
```

`Mount` takes the parent component, a position index, and a factory function. It uses the `(parent, index)` pair as a cache key:

- **First render**: The factory runs, creating a new component instance. The framework calls `BindApp()` to wire up state fields, then calls `Init()` if the component implements `Initializer`. The instance is cached.
- **Subsequent renders**: The cached instance is reused. If the component implements `PropsUpdater`, a fresh instance is created from the factory and passed to `UpdateProps()` so the cached instance can pick up changed props without losing its internal state.

This is why a child's local state survives parent re-renders. The sidebar's `selected` state persists because the framework returns the cached sidebar instance instead of creating a new one.

When a component is no longer rendered (removed from the tree by an `if` condition, for example), the framework's mark-and-sweep cleanup calls the cleanup function returned by `Init()` and removes the instance from cache.

You don't call `Mount` directly. The code generator handles it. Write `@Sidebar(a.category)` in your `.gsx` file and the generated code does the rest.

### Generated BindApp and UpdateProps

The code generator automatically produces `BindApp` and `UpdateProps` methods for struct components. `BindApp` wires every `*tui.State` and `*tui.Events` field to the running app so that `.Set()` calls mark the UI dirty. `UpdateProps` copies non-state fields from a fresh instance to the cached one, keeping state intact while updating configuration.

You don't write these methods by hand. For a component with one shared state field and one local state field:

```go
// Auto-generated
func (s *sidebar) BindApp(app *tui.App) {
    if s.category != nil {
        s.category.BindApp(app)
    }
    if s.selected != nil {
        s.selected.BindApp(app)
    }
    if s.expanded != nil {
        s.expanded.BindApp(app)
    }
}
```

## Architecture Patterns

### Orchestrator Pattern

One root component owns all shared state. Children receive state through their constructors and read/write it, but the root is the single source of truth.

```gsx
type myApp struct {
    searchActive *tui.State[bool]
    query        *tui.State[string]
    category     *tui.State[string]
}

func MyApp() *myApp {
    return &myApp{
        searchActive: tui.NewState(false),
        query:        tui.NewState(""),
        category:     tui.NewState("Documents"),
    }
}

templ (a *myApp) Render() {
    <div class="flex-col h-full border-rounded border-cyan">
        <div class="flex justify-center px-1 shrink-0">
            <span class="text-gradient-cyan-magenta font-bold">File Explorer</span>
        </div>
        <hr />
        <div class="flex grow min-h-0 overflow-hidden">
            @Sidebar(a.category)
            @Content(a.category, a.query)
        </div>
        @SearchBar(a.searchActive, a.query)
        <hr />
        <div class="flex justify-center px-1 shrink-0">
            <span class="font-dim">/search | Ctrl+B sidebar | j/k navigate | q quit</span>
        </div>
    </div>
}
```

This works well for apps where a few pieces of state drive the whole UI. The root component's KeyMap can check `searchActive` to decide which keys to register, and every child component reads the same state values.

### Distributed State

Each component owns the state it's responsible for. Coordination happens through a small number of shared state values, while most state stays local.

In the file explorer example, the sidebar owns `expanded` and `selected` locally. Only `category` is shared with the content panel. The search bar owns its text-input mechanics locally; only `active` and `query` are shared.

This scales better as apps grow. Each component manages its own concerns, and you only share what's necessary for cross-component coordination.

### Choosing Between Them

For small apps (2-3 components, a handful of state values), the orchestrator pattern is simpler and keeps everything in one place.

For larger apps, distributed state keeps individual components self-contained. New components can be added without modifying the root, as long as they receive the shared state they need through their constructor.

Most apps end up somewhere in between: a root that owns the shared coordination state, with children that manage their own internal state.

## Complete Example

Here's a full file explorer with three components: sidebar, content panel, and search bar. The root component wires them together with shared state.

**app.gsx**:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

type myApp struct {
    searchActive *tui.State[bool]
    query        *tui.State[string]
    category     *tui.State[string]
}

func MyApp() *myApp {
    return &myApp{
        searchActive: tui.NewState(false),
        query:        tui.NewState(""),
        category:     tui.NewState("Documents"),
    }
}

func (a *myApp) KeyMap() tui.KeyMap {
    km := tui.KeyMap{
        tui.On(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
    if !a.searchActive.Get() {
        km = append(km, tui.On(tui.Rune('/'), func(ke tui.KeyEvent) {
            a.searchActive.Set(true)
        }))
        km = append(km, tui.On(tui.Rune('q'), func(ke tui.KeyEvent) {
            ke.App().Stop()
        }))
    }
    return km
}

templ (a *myApp) Render() {
    <div class="flex-col h-full border-rounded border-cyan">
        <div class="flex justify-center px-1 shrink-0">
            <span class="text-gradient-cyan-magenta font-bold">File Explorer</span>
        </div>
        <hr />
        <div class="flex grow min-h-0 overflow-hidden">
            @Sidebar(a.category)
            @Content(a.category, a.query)
        </div>
        @SearchBar(a.searchActive, a.query)
        <hr />
        <div class="flex justify-center px-1 shrink-0">
            <span class="font-dim">/search | Ctrl+B sidebar | j/k navigate | q quit</span>
        </div>
    </div>
}
```

**sidebar.gsx**:

```gsx
package main

import tui "github.com/grindlemire/go-tui"

type sidebar struct {
    category *tui.State[string]
    expanded *tui.State[bool]
    selected *tui.State[int]
}

var categories = []string{"Documents", "Images", "Music", "Projects", "Downloads"}

func Sidebar(category *tui.State[string]) *sidebar {
    return &sidebar{
        category: category,
        expanded: tui.NewState(true),
        selected: tui.NewState(0),
    }
}

func (s *sidebar) KeyMap() tui.KeyMap {
    km := tui.KeyMap{
        tui.On(tui.KeyCtrlB, func(ke tui.KeyEvent) {
            s.expanded.Set(!s.expanded.Get())
        }),
    }
    if s.expanded.Get() {
        km = append(km, tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { s.moveDown() }))
        km = append(km, tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { s.moveUp() }))
        km = append(km, tui.On(tui.KeyDown, func(ke tui.KeyEvent) { s.moveDown() }))
        km = append(km, tui.On(tui.KeyUp, func(ke tui.KeyEvent) { s.moveUp() }))
        km = append(km, tui.On(tui.KeyEnter, func(ke tui.KeyEvent) {
            idx := s.selected.Get()
            if idx >= 0 && idx < len(categories) {
                s.category.Set(categories[idx])
            }
        }))
    }
    return km
}

func (s *sidebar) moveDown() {
    s.selected.Update(func(v int) int {
        if v >= len(categories)-1 {
            return 0
        }
        return v + 1
    })
    s.category.Set(categories[s.selected.Get()])
}

func (s *sidebar) moveUp() {
    s.selected.Update(func(v int) int {
        if v <= 0 {
            return len(categories) - 1
        }
        return v - 1
    })
    s.category.Set(categories[s.selected.Get()])
}

func (s *sidebar) sidebarWidth() int {
    if s.expanded.Get() {
        return 22
    }
    return 5
}

templ (s *sidebar) Render() {
    <div class="flex-col border-single shrink-0" width={s.sidebarWidth()}>
        if s.expanded.Get() {
            <div class="flex-col px-1">
                <span class="text-gradient-cyan-magenta font-bold">Folders</span>
            </div>
            <hr />
            <div class="flex-col px-1 grow">
                for i, cat := range categories {
                    if i == s.selected.Get() {
                        <span class="text-cyan font-bold">{"> " + cat}</span>
                    } else {
                        <span class="font-dim">{"  " + cat}</span>
                    }
                }
            </div>
            <hr />
            <div class="flex-col px-1">
                <span class="font-dim text-cyan">Ctrl+B hide</span>
            </div>
        } else {
            <span class="text-cyan font-bold px-1">F</span>
        }
    </div>
}
```

**search.gsx** (search bar and content panel):

```gsx
package main

import (
    "fmt"
    "strings"

    tui "github.com/grindlemire/go-tui"
)

type searchBar struct {
    active *tui.State[bool]
    query  *tui.State[string]
}

func SearchBar(active *tui.State[bool], query *tui.State[string]) *searchBar {
    return &searchBar{active: active, query: query}
}

func (s *searchBar) KeyMap() tui.KeyMap {
    if !s.active.Get() {
        return nil
    }
    return tui.KeyMap{
        tui.OnStop(tui.AnyRune,s.appendChar),
        tui.OnStop(tui.KeyBackspace, s.deleteChar),
        tui.OnStop(tui.KeyEnter, s.submit),
        tui.OnStop(tui.KeyEscape, s.deactivate),
    }
}

func (s *searchBar) appendChar(ke tui.KeyEvent) {
    s.query.Set(s.query.Get() + string(ke.Rune))
}

func (s *searchBar) deleteChar(ke tui.KeyEvent) {
    q := s.query.Get()
    if len(q) > 0 {
        s.query.Set(q[:len(q)-1])
    }
}

func (s *searchBar) submit(ke tui.KeyEvent) {
    s.active.Set(false)
}

func (s *searchBar) deactivate(ke tui.KeyEvent) {
    s.active.Set(false)
    s.query.Set("")
}

templ (s *searchBar) Render() {
    <div class="shrink-0">
        if s.active.Get() {
            <hr />
            <div class="px-1 flex gap-1">
                <span class="text-cyan font-bold">Search:</span>
                <span>{s.query.Get()}</span>
                <span class="text-cyan blink">_</span>
            </div>
        }
    </div>
}

// Content displays files for the selected category
type content struct {
    category *tui.State[string]
    query    *tui.State[string]
}

func Content(category *tui.State[string], query *tui.State[string]) *content {
    return &content{category: category, query: query}
}

var filesByCategory = map[string][]string{
    "Documents": {"report.pdf", "notes.md", "budget.xlsx", "readme.txt", "design.doc"},
    "Images":    {"photo.jpg", "screenshot.png", "logo.svg", "banner.gif", "icon.ico"},
    "Music":     {"song.mp3", "album.flac", "podcast.ogg", "ringtone.wav"},
    "Projects":  {"go-tui/", "website/", "api-server/", "mobile-app/", "scripts/"},
    "Downloads": {"setup.exe", "archive.zip", "data.csv", "patch-v2.tar.gz"},
}

func (c *content) filteredFiles() []string {
    cat := c.category.Get()
    files, ok := filesByCategory[cat]
    if !ok {
        return nil
    }
    q := strings.ToLower(c.query.Get())
    if q == "" {
        return files
    }
    var result []string
    for _, f := range files {
        if strings.Contains(strings.ToLower(f), q) {
            result = append(result, f)
        }
    }
    return result
}

templ (c *content) Render() {
    <div class="flex-col grow px-2 overflow-hidden">
        <span class="font-bold text-cyan">{c.category.Get() + "/"}</span>
        <hr />
        for i, file := range c.filteredFiles() {
            if i == len(c.filteredFiles())-1 {
                <span>{fmt.Sprintf("└── %s", file)}</span>
            } else {
                <span>{fmt.Sprintf("├── %s", file)}</span>
            }
        }
        if len(c.filteredFiles()) == 0 {
            <span class="font-dim">No matching files</span>
        }
        if c.query.Get() != "" {
            <br />
            <span class="font-dim">{fmt.Sprintf("Filtering: \"%s\"", c.query.Get())}</span>
        }
    </div>
}
```

**main.go**:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(MyApp()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "App error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

Here's what it looks like with the sidebar, file list, and preview all wired together:

![Multi-Component Applications screenshot](/guides/14.png)

## Next Steps

- [Inline Mode and Alternate Screen](inline-mode) -- Rendering inline at the bottom of the terminal and switching to full-screen overlays
- [Focus Management](focus) -- Tab navigation and focus groups for form-like interfaces
- [Components](components) -- Component basics, lifecycle interfaces, and composition patterns
- [State and Reactivity](state) -- Reactive state management with `State[T]`
