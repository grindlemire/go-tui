# Documentation Design Spec

## Overview

A complete documentation set for go-tui consisting of **Guides** (progressive learning path) and **Reference** (exhaustive API documentation). Each document is independent and can be written by a separate agent.

All documents live in `docs/content/` with subdirectories `guides/` and `reference/`.

## Writing Conventions (apply to ALL documents)

- Use GitHub-flavored Markdown
- Every code example must be complete and runnable (no "..." elision unless showing output)
- Use `.gsx` syntax for UI examples, plain Go for non-UI code
- Always show the `import` block when introducing new packages
- Use the import alias `tui "github.com/grindlemire/go-tui"` in all examples
- Cross-reference related docs with relative links: `[State](../reference/state.md)`
- Each document should be self-contained — an agent should be able to write it without reading other docs
- Do not duplicate full API signatures from reference docs in guides; instead link to them
- Include practical "when to use this" guidance, not just "what it does"

## Framework Context (for all documents)

### Project Import Path

```go
import tui "github.com/grindlemire/go-tui"
```

### .gsx File Structure

Every `.gsx` file follows this pattern:

```gsx
package mypackage

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

// Struct component (stateful)
type myApp struct {
    someState *tui.State[int]
    someRef   *tui.Ref
}

// Constructor
func MyApp() *myApp {
    return &myApp{
        someState: tui.NewState(0),
        someRef:   tui.NewRef(),
    }
}

// Render method — the `templ` keyword generates a method returning *tui.Element
templ (a *myApp) Render() {
    <div class="flex-col gap-1 p-1">
        <span class="font-bold text-cyan">{fmt.Sprintf("Count: %d", a.someState.Get())}</span>
    </div>
}

// KeyMap — returns key bindings
func (a *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('+', func(ke tui.KeyEvent) { a.someState.Update(func(v int) int { return v + 1 }) }),
    }
}

// HandleMouse — handles mouse events using refs
func (a *myApp) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(a.someRef, a.doSomething),
    )
}

// Watchers — returns background watchers
func (a *myApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, a.tick),
    }
}

// Pure templ component (no state, just params)
templ Card(title string) {
    <div class="border-rounded p-1">
        <span class="font-bold">{title}</span>
        {children...}
    </div>
}
```

### main.go Structure

Every app's `main.go` follows this pattern:

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
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Available Tailwind Classes

**Layout**: `flex`, `flex-col`, `flex-row`, `gap-N`, `grow`, `shrink`, `flex-1`, `flex-none`, `flex-grow-N`, `flex-shrink-N`
**Justify**: `justify-start`, `justify-center`, `justify-end`, `justify-between`, `justify-around`, `justify-evenly`
**Align**: `items-start`, `items-center`, `items-end`, `items-stretch`, `self-start`, `self-center`, `self-end`, `self-stretch`
**Text align**: `text-left`, `text-center`, `text-right`
**Spacing**: `p-N`, `px-N`, `py-N`, `pt-N`, `pr-N`, `pb-N`, `pl-N`, `m-N`, `mx-N`, `my-N`, `mt-N`, `mr-N`, `mb-N`, `ml-N`
**Sizing**: `w-N`, `h-N`, `w-full`, `h-full`, `w-auto`, `h-auto`, `w-1/2`, `w-1/3`, `w-2/3`, `min-w-N`, `max-w-N`, `min-h-N`, `max-h-N`
**Borders**: `border`, `border-single`, `border-double`, `border-rounded`, `border-thick`, `border-{color}` (e.g. `border-cyan`)
**Text styles**: `font-bold`, `font-dim`, `text-dim`, `italic`, `underline`, `strikethrough`, `reverse`
**Text colors**: `text-red`, `text-green`, `text-blue`, `text-cyan`, `text-magenta`, `text-yellow`, `text-white`, `text-black`, `text-bright-red`, `text-bright-green`, `text-bright-blue`, `text-bright-cyan`, `text-bright-magenta`, `text-bright-yellow`, `text-bright-white`, `text-bright-black`
**Background colors**: `bg-red`, `bg-green`, `bg-blue`, `bg-cyan`, `bg-magenta`, `bg-yellow`, `bg-white`, `bg-black`, `bg-bright-red`, `bg-bright-green`, `bg-bright-blue`, `bg-bright-cyan`, `bg-bright-magenta`, `bg-bright-yellow`, `bg-bright-white`, `bg-bright-black`
**Gradients**: `text-gradient-{c1}-{c2}[-direction]`, `bg-gradient-{c1}-{c2}[-direction]`, `border-gradient-{c1}-{c2}[-direction]` where direction is `-h` (horizontal), `-v` (vertical), `-dd` (diagonal down), `-du` (diagonal up)
**Scroll**: `overflow-scroll`, `overflow-y-scroll`, `overflow-x-scroll`

### Available Elements

`<div>` (block container), `<span>` (inline text), `<p>` (paragraph), `<ul>` (unordered list), `<li>` (list item), `<button>` (clickable), `<input>` (text input, self-closing), `<table>` (table), `<progress>` (progress bar, self-closing), `<hr>` (horizontal rule, self-closing), `<br>` (line break, self-closing)

### Element Attributes

**Common**: `id`, `class`, `disabled`, `ref`, `deps`
**Layout**: `width`, `widthPercent`, `height`, `heightPercent`, `minWidth`, `minHeight`, `maxWidth`, `maxHeight`, `direction`, `justify`, `align`, `gap`, `flexGrow`, `flexShrink`, `alignSelf`, `padding`, `margin`
**Visual**: `border`, `borderStyle`, `background`, `text`, `textStyle`, `textAlign`
**Scroll**: `scrollable`, `scrollOffset`, `scrollbarStyle`, `scrollbarThumbStyle`
**Input-specific**: `value`, `placeholder`
**Progress-specific**: `value` (current), `max` (maximum)
**Focus**: `focusable`, `onFocus`, `onBlur`

### Control Flow in .gsx

```gsx
// Conditionals
@if condition {
    <span>Shown when true</span>
} @else {
    <span>Shown when false</span>
}

// Loops
@for i, item := range items {
    <span>{item}</span>
}

// Let bindings
@let label = fmt.Sprintf("Count: %d", count)
<span>{label}</span>
```

### Key Types Quick Reference

```go
// Dimensions
tui.Fixed(10), tui.Percent(50), tui.Auto()

// Borders
tui.BorderNone, tui.BorderSingle, tui.BorderDouble, tui.BorderRounded, tui.BorderThick

// Direction
tui.Row, tui.Column

// Justify
tui.JustifyStart, tui.JustifyCenter, tui.JustifyEnd, tui.JustifySpaceBetween, tui.JustifySpaceAround, tui.JustifySpaceEvenly

// Align
tui.AlignStart, tui.AlignCenter, tui.AlignEnd, tui.AlignStretch

// Scroll
tui.ScrollNone, tui.ScrollVertical, tui.ScrollHorizontal, tui.ScrollBoth

// Style construction
tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan)).Bold()

// Colors
tui.Black, tui.Red, tui.Green, tui.Yellow, tui.Blue, tui.Magenta, tui.Cyan, tui.White
tui.BrightBlack, tui.BrightRed, tui.BrightGreen, tui.BrightYellow, tui.BrightBlue, tui.BrightMagenta, tui.BrightCyan, tui.BrightWhite
tui.ANSIColor(index), tui.RGBColor(r,g,b), tui.HexColor("#RRGGBB")

// State
state := tui.NewState(initialValue)
state.Get(), state.Set(v), state.Update(func(v T) T)

// Watchers
tui.OnTimer(duration, handler)
tui.Watch(channel, handler)

// Events bus
events := tui.NewEvents[string]()

// KeyMap entries
tui.OnKey(tui.KeyEscape, handler)
tui.OnRune('q', handler)
tui.OnRunes(handler)           // catch-all for any rune
tui.OnRuneStop('x', handler)   // stops propagation
tui.OnKeyStop(tui.KeyEnter, handler)

// Mouse
tui.HandleClicks(mouseEvent, tui.Click(ref, handler), ...)

// Refs
ref := tui.NewRef()      // single element ref
ref.El()                 // get the *Element (may be nil)
refs := tui.NewRefList() // slice of elements from @for loops
refMap := tui.NewRefMap[string]() // keyed map of elements

// App options
tui.WithRootComponent(component)
tui.WithRootView(viewable)
tui.WithInlineHeight(rows)
tui.WithFrameRate(fps)
tui.WithMouse()
tui.WithoutMouse()
```

### Code Generation

After writing `.gsx` files, run `go run ./cmd/tui generate ./path/to/...` from the project root to produce `_gsx.go` files. The generated files should NOT be hand-edited.

### Source Code Locations

All public API types live in the root `tui` package. Internal packages (`internal/layout`, `internal/tuigen`, `internal/formatter`, `internal/lsp`, `internal/debug`) are not importable by external consumers. When referencing source code for accuracy, consult:

- `element.go`, `element_options.go`, `element_accessors.go` — Element type and configuration
- `app.go`, `app_options.go`, `app_loop.go`, `app_lifecycle.go`, `app_render.go`, `app_events.go`, `app_screen.go` — App type and lifecycle
- `state.go` — State[T] generic reactive state
- `events.go` — Events[T] generic event bus
- `event.go`, `key.go`, `keymap.go` — Event types and key handling
- `style.go`, `color.go`, `border.go` — Visual styling
- `layout.go` — Re-exported layout types (Direction, Justify, Align, Value, etc.)
- `focus.go`, `focus_group.go` — Focus management
- `watcher.go` — Watcher interface and implementations
- `ref.go`, `click.go` — Element references and click handling
- `buffer.go`, `cell.go`, `render.go` — Rendering primitives
- `terminal.go`, `terminal_ansi.go`, `caps.go` — Terminal abstraction
- `textarea.go`, `textarea_options.go` — Built-in TextArea component
- `mock_terminal.go`, `mock_reader.go` — Testing utilities
- `component.go`, `mount.go` — Component interfaces and mounting
- `element_scroll.go` — Scroll API
- `element_tree.go` — Element tree manipulation
- `element_focus.go` — Element focus API
- `element_render.go` — Element rendering
- `element_watchers.go` — Element watcher/focus discovery

---

## Document Inventory

### Guides

Guides follow a progressive learning path. Earlier guides introduce foundational concepts; later guides build on them. Each guide should include practical, runnable examples.

---

#### Guide 01: Getting Started ✅
**File**: `docs/content/guides/01-getting-started.md`

**Purpose**: Get a new user from zero to a running go-tui app. First document anyone reads.

**Sections**:
1. **What is go-tui** — One paragraph: declarative terminal UI framework for Go with templ-like syntax and flexbox layout. Mention key selling points: pure Go (no CGO), minimal dependencies, type-safe generated code.
2. **Installation** — `go get github.com/grindlemire/go-tui`, install the `tui` CLI tool: `go install github.com/grindlemire/go-tui/cmd/tui@latest`
3. **Your First App** — Step-by-step walkthrough building a "Hello, Terminal!" app:
   - Create project directory, `go mod init`
   - Write `hello.gsx` with a simple struct component that renders a centered message
   - Write `main.go` with `tui.NewApp(tui.WithRootComponent(...))` and `app.Run()`
   - Run `tui generate ./...` to generate `_gsx.go`
   - Run `go run .` to see the result
   - Add a quit key (`q` via KeyMap) and explain the pattern
4. **How It Works** — Brief architecture overview:
   - `.gsx` files compile to Go code via `tui generate`
   - Components define a `Render` method using `templ` syntax
   - The framework handles layout (flexbox), rendering (double-buffered), and input (raw terminal)
   - Diagram: `.gsx` → `tui generate` → `_gsx.go` → `go build` → binary
5. **Core Concepts** — Brief intro to each concept (1-2 sentences each, with links to deeper guides):
   - Components (struct vs pure templ)
   - Elements (`<div>`, `<span>`, etc.)
   - Styling (Tailwind-like classes)
   - Layout (flexbox)
   - State (reactive with `State[T]`)
   - Events (keyboard, mouse)
6. **Next Steps** — Links to Guide 02 (GSX Syntax) and Guide 03 (Styling)

**Key Examples**:
- Complete hello world app (hello.gsx + main.go)
- The hello world but with a quit key added

**Cross-references**: Link to [GSX Syntax](02-gsx-syntax.md), [CLI Reference](../reference/cli.md)

---

#### Guide 02: GSX Syntax ✅
**File**: `docs/content/guides/02-gsx-syntax.md`

**Purpose**: Teach the .gsx file format — the templ-like syntax that is the primary way users write go-tui UIs.

**Sections**:
1. **Overview** — `.gsx` files are Go files extended with templ-like syntax for declaring UI. They compile to standard Go via `tui generate`.
2. **Package and Imports** — Standard Go package declaration. Imports work exactly like Go. The tui import alias convention: `tui "github.com/grindlemire/go-tui"`.
3. **Pure Components (templ functions)** — Stateless, parameterized UI fragments:
   - Basic: `templ Greeting(name string) { <span>{"Hello, " + name}</span> }`
   - With children: `templ Card(title string) { <div>{children...}</div> }`
   - Calling components: `<Card title="My Card"><span>content</span></Card>`
   - When to use: reusable visual elements that don't need their own state
4. **Struct Components** — Stateful components with lifecycle:
   - Define a struct with fields (state, refs, etc.)
   - Write a constructor function
   - Use `templ (s *myStruct) Render()` for the render method
   - Explain the Component interface: `Render(app *App) *Element`
   - Optional interfaces: `KeyListener`, `MouseListener`, `WatcherProvider`, `Initializer`, `AppBinder`
5. **Elements** — HTML-like tags:
   - Block containers: `<div>`, `<p>`
   - Inline: `<span>`
   - Lists: `<ul>`, `<li>`
   - Interactive: `<button>`, `<input />`
   - Data display: `<table>`, `<progress />`
   - Formatting: `<hr />`, `<br />`
   - Self-closing vs container elements
6. **Attributes** — Setting element properties:
   - String attributes: `class="flex-col"`, `id="header"`
   - Go expression attributes: `width={42}`, `textStyle={myStyle}`
   - Boolean attributes: `disabled={true}`, `focusable={true}`
   - Ref attributes: `ref={myRef}`
7. **Go Expressions** — Embedding Go code:
   - Text content: `<span>{fmt.Sprintf("Count: %d", n)}</span>`
   - Attribute values: `width={computeWidth()}`
   - Method calls: `textStyle={s.getStyle()}`
8. **Control Flow** — Template directives:
   - `@if` / `@else`: conditional rendering with examples
   - `@for`: loop rendering with `range`, index variables
   - `@let`: local bindings for computed values
   - Nesting control flow inside elements
9. **Helper Functions** — Regular Go functions in .gsx files:
   - `func helper(s string) string { ... }` — not a component, just a Go function
   - When to use: formatting, computation, shared logic
10. **Code Generation** — How `tui generate` works:
    - Reads `.gsx`, produces `_gsx.go` in same directory
    - Never hand-edit generated files
    - Re-run after any `.gsx` change
    - `tui check` validates without generating
    - `tui fmt` formats `.gsx` files

**Key Examples**:
- Pure component with children slot
- Struct component with state and KeyMap
- Control flow: conditional rendering, list rendering, let bindings
- Calling a pure component from a struct component's render

**Cross-references**: Link to [Elements Reference](../reference/gsx-syntax.md), [Components Guide](06-components.md), [CLI Reference](../reference/cli.md)

---

#### Guide 03: Styling and Colors ✅
**File**: `docs/content/guides/03-styling.md`

**Purpose**: Teach the visual styling system — Tailwind-like classes, colors, text styles, borders, and gradients.

**Sections**:
1. **Overview** — go-tui uses a Tailwind-inspired class system in the `class` attribute. Classes map to element options at compile time.
2. **Text Styles** — Available text decorations:
   - `font-bold`, `font-dim`/`text-dim`, `italic`, `underline`, `strikethrough`, `reverse`
   - Example showing each style applied to text
3. **Text Colors** — Standard and bright ANSI colors:
   - Standard: `text-red`, `text-green`, `text-blue`, `text-cyan`, `text-magenta`, `text-yellow`, `text-white`, `text-black`
   - Bright: `text-bright-red`, `text-bright-green`, etc.
   - Example showing all 16 colors
4. **Background Colors** — Same color set with `bg-` prefix:
   - `bg-red`, `bg-cyan`, etc.
   - Example showing colored blocks
5. **Combining Styles** — Multiple classes compose:
   - `class="font-bold text-cyan bg-black underline"`
   - Example showing combined styles
6. **Borders** — Border styles and colors:
   - Styles: `border-single`, `border-double`, `border-rounded`, `border-thick`
   - Show the actual characters for each: `┌─┐│└─┘`, `╔═╗║╚═╝`, `╭─╮│╰─╯`, `┏━┓┃┗━┛`
   - Colors: `border-cyan`, `border-red`, etc.
   - Example showing all four border styles
7. **Gradients** — Text, background, and border gradients:
   - Syntax: `text-gradient-{start}-{end}[-direction]`
   - Directions: `-h` (horizontal, default), `-v` (vertical), `-dd` (diagonal down), `-du` (diagonal up)
   - Works for text, background, and borders
   - Example showing gradients in each direction
8. **Programmatic Styling** — Using the Go `Style` type directly:
   - `tui.NewStyle().Bold().Foreground(tui.ANSIColor(tui.Cyan))`
   - Color constructors: `tui.ANSIColor()`, `tui.RGBColor()`, `tui.HexColor()`
   - Setting via attributes: `textStyle={myStyle}`, `borderStyle={myStyle}`, `background={myStyle}`
   - When to use programmatic vs class-based styling
9. **Color Capabilities** — Terminal color support:
   - 16 colors (ANSI), 256 colors, true color (24-bit RGB)
   - `tui.DetectCapabilities()` for runtime detection
   - Graceful degradation: RGB colors auto-downgrade on limited terminals

**Key Examples**:
- Text style sampler (all decorations)
- Color palette display (16 standard + 16 bright)
- Border style comparison
- Gradient showcase (text, background, border in all 4 directions)
- Programmatic style construction

**Cross-references**: Link to [Style Reference](../reference/styling.md), [Layout Guide](04-layout.md)

---

#### Guide 04: Layout ✅
**File**: `docs/content/guides/04-layout.md`

**Purpose**: Teach the flexbox layout system — how to arrange elements on screen.

**Sections**:
1. **Overview** — go-tui uses a CSS flexbox-compatible layout engine. Every `<div>` is a flex container. Layout is controlled via Tailwind classes or element attributes.
2. **Direction** — Row vs Column:
   - `flex` / `flex-row` — children arranged horizontally (default)
   - `flex-col` — children arranged vertically
   - Example showing the same 3 children in row vs column
3. **Justify Content** — Main axis alignment:
   - `justify-start` (default), `justify-center`, `justify-end`
   - `justify-between`, `justify-around`, `justify-evenly`
   - Visual example for each showing 3 blocks with spacing
4. **Align Items** — Cross axis alignment:
   - `items-start`, `items-center`, `items-end`, `items-stretch` (default)
   - Example with children of different heights to show alignment
5. **Gap** — Spacing between children:
   - `gap-N` — N characters between each child
   - Example comparing gap-0, gap-1, gap-2
6. **Flex Grow and Shrink** — Flexible sizing:
   - `grow` / `flex-grow-N` — element expands to fill space
   - `shrink` / `flex-shrink-N` — element shrinks when space is tight
   - `flex-1` — shorthand for grow-1
   - `flex-none` — no grow, no shrink
   - Example: sidebar (fixed) + main content (grow)
7. **Sizing** — Explicit dimensions:
   - Fixed: `w-N`, `h-N` (in characters/rows)
   - Percentage: `w-1/2`, `w-1/3`, `w-2/3`, or `widthPercent={50}`
   - Full: `w-full`, `h-full`
   - Auto: `w-auto`, `h-auto` (size to content)
   - Min/max: `min-w-N`, `max-w-N`, `min-h-N`, `max-h-N`
8. **Padding and Margin** — Spacing inside and outside:
   - Padding: `p-N`, `px-N`, `py-N`, `pt-N`, `pr-N`, `pb-N`, `pl-N`
   - Margin: `m-N`, `mx-N`, `my-N`, `mt-N`, `mr-N`, `mb-N`, `ml-N`
   - Example showing padding (inside border) vs margin (outside border)
9. **Common Layout Patterns** — Practical examples:
   - Full-screen app with header, content, footer
   - Sidebar + main content
   - Centered card (both axes)
   - Dashboard grid (2x2 or 3-column)
   - Stacked form fields

**Key Examples**:
- Direction comparison (row vs column)
- Justify content gallery (all 6 options)
- Sidebar + main content layout
- Full-screen app skeleton (header/content/footer)
- Centered card

**Cross-references**: Link to [Layout Reference](../reference/layout.md), [Styling Guide](03-styling.md)

---

#### Guide 05: State and Reactivity ✅
**File**: `docs/content/guides/05-state.md`

**Purpose**: Teach reactive state management — how to make UIs that respond to data changes.

**Sections**:
1. **Overview** — go-tui uses a reactive model: when state changes, the UI automatically re-renders. State is managed via the generic `State[T]` type.
2. **Creating State** — `tui.NewState[T](initialValue)`:
   - Declare as struct field: `count *tui.State[int]`
   - Initialize in constructor: `tui.NewState(0)`
   - Works with any type: `int`, `string`, `bool`, `[]string`, custom structs
3. **Reading State** — `state.Get()`:
   - Use in render: `{s.count.Get()}`
   - Use in expressions: `@if s.count.Get() > 0 { ... }`
4. **Writing State** — `state.Set(value)` and `state.Update(fn)`:
   - Direct set: `s.count.Set(42)`
   - Functional update: `s.count.Update(func(v int) int { return v + 1 })`
   - When to use each: Set for absolute values, Update for relative changes
   - State changes automatically trigger re-render
5. **Conditional Rendering** — `@if` / `@else` with state:
   - Show/hide elements based on state
   - Chained conditions: `@if ... { } @else { @if ... { } @else { } }`
   - Example: status panel that changes based on a value
6. **List Rendering** — `@for` with state:
   - Rendering a `State[[]string]` with `@for`
   - Highlighting selected items using state index
   - Example: selectable list with keyboard navigation
7. **Computed Values** — `@let` bindings:
   - Derive display values: `@let label = fmt.Sprintf("Count: %d", s.count.Get())`
   - Avoid expensive recomputation
8. **Batching Updates** — `app.Batch(fn)`:
   - Multiple state changes in a single re-render
   - Example: resetting multiple state values at once
9. **State Bindings** — `state.Bind(fn)`:
   - Register a callback for state changes
   - Returns an `Unbind` function for cleanup
   - Use case: side effects when state changes (logging, derived state)

**Key Examples**:
- Counter with increment/decrement/reset
- Status panel reacting to counter value (@if/@else chains)
- Selectable list with highlighted current item
- Batched state reset

**Cross-references**: Link to [State Reference](../reference/state.md), [Events Guide](07-events.md)

---

#### Guide 06: Components ✅
**File**: `docs/content/guides/06-components.md`

**Purpose**: Teach component patterns — pure vs struct components, children slots, composition, and component interfaces.

**Sections**:
1. **Overview** — Two kinds of components: pure templ functions (stateless) and struct components (stateful). Composition is the primary pattern for building complex UIs.
2. **Pure Components** — `templ` functions:
   - Syntax: `templ Name(params...) { ... }`
   - No state, no lifecycle — just parameters in, elements out
   - Children slot: `{children...}` to accept nested content
   - When to use: reusable visual elements (cards, badges, headers, layouts)
   - Example: `Card`, `Badge`, `Header` components
3. **Struct Components** — Stateful components:
   - Define struct with state fields
   - Constructor function (convention: `func NewMyComp() *myComp`)
   - `templ (s *myComp) Render()` for the render method
   - The `Component` interface: `Render(app *App) *Element`
   - Example: counter component with its own state
4. **Component Interfaces** — Optional behaviors:
   - `KeyListener` — `KeyMap() tui.KeyMap` for keyboard handling
   - `MouseListener` — `HandleMouse(MouseEvent) bool` for mouse events
   - `WatcherProvider` — `Watchers() []Watcher` for timers/channels
   - `Initializer` — `Init() func()` for setup (returns cleanup function)
   - `AppBinder` — `BindApp(app *App)` for app-level access
   - `PropsUpdater` — `UpdateProps(fresh Component)` for prop change handling
   - Each with a brief explanation and example
5. **Children Slot** — `{children...}`:
   - How it works: parent passes nested elements
   - Example: wrapper component that adds border + padding
   - Only available in pure templ components (not struct Render methods)
6. **Component Composition** — Building complex UIs:
   - Nesting pure components
   - Using struct components inside templ renders
   - Mounting struct components: `tui.Mount` for sub-components with lifecycle
   - Example: user card composed of Card + Badge + text
7. **Composition Patterns** — Practical approaches:
   - Layout components (sidebar layout, header-content-footer)
   - Data display components (table row, list item)
   - Container components (scrollable panel, bordered section)

**Key Examples**:
- Pure component family: Card, Badge, Header
- Struct component: interactive counter
- Composition: UserCard using Card and Badge
- Layout component with children slot
- App using Mount for sub-components

**Cross-references**: Link to [State Guide](05-state.md), [Events Guide](07-events.md), [App Reference](../reference/app.md)

---

#### Guide 07: Event Handling ✅
**File**: `docs/content/guides/07-events.md`

**Purpose**: Teach keyboard and mouse event handling — KeyMap, special keys, modifiers, mouse clicks.

**Sections**:
1. **Overview** — go-tui routes keyboard and mouse events to components via the `KeyListener` and `MouseListener` interfaces. The `KeyMap` system provides declarative key binding.
2. **KeyMap Basics** — Returning key bindings:
   - Implement `KeyMap() tui.KeyMap` on your struct
   - `tui.KeyMap` is a slice of `KeyBinding`
   - Bindings are checked in order; first match wins
3. **Key Bindings** — The binding helpers:
   - `tui.OnKey(key, handler)` — match a special key
   - `tui.OnRune(r, handler)` — match a specific character
   - `tui.OnRunes(handler)` — catch-all for any printable character
   - `tui.OnKeyStop(key, handler)` — match and stop propagation
   - `tui.OnRuneStop(r, handler)` — match rune and stop propagation
   - `tui.OnRunesStop(handler)` — catch-all and stop propagation
4. **Special Keys** — Available key constants:
   - Navigation: `KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`, `KeyHome`, `KeyEnd`, `KeyPageUp`, `KeyPageDown`
   - Editing: `KeyEnter`, `KeyTab`, `KeyBackspace`, `KeyDelete`, `KeyInsert`
   - Control: `KeyEscape`, `KeyCtrlA` through `KeyCtrlZ`, `KeyCtrlSpace`
   - Function: `KeyF1` through `KeyF12`
5. **KeyEvent Properties** — What you get in the handler:
   - `ke.Key` — the Key constant
   - `ke.Rune` — the character (for `KeyRune`)
   - `ke.Mod` — modifier flags (`ModCtrl`, `ModAlt`, `ModShift`)
   - `ke.IsRune()` — true if it's a printable character
   - `ke.Is(key, mods...)` — check key and modifiers
   - `ke.App()` — access the App instance (e.g., `ke.App().Stop()`)
6. **Stop Propagation** — Preventing parent handling:
   - Default: events bubble up through component tree
   - `OnKeyStop`/`OnRuneStop`/`OnRunesStop`: handler runs and event stops
   - Use case: search input captures all runes when active
7. **Mouse Events** — `HandleMouse(MouseEvent) bool`:
   - `MouseEvent` fields: `Button`, `Action`, `X`, `Y`, `Mod`
   - Buttons: `MouseLeft`, `MouseMiddle`, `MouseRight`, `MouseWheelUp`, `MouseWheelDown`
   - Actions: `MousePress`, `MouseRelease`, `MouseDrag`
   - Enable mouse: `tui.WithMouse()` app option
8. **Click Handling with Refs** — `tui.HandleClicks`:
   - Create refs: `btn := tui.NewRef()`
   - Attach to elements: `ref={btn}`
   - Handle clicks: `tui.HandleClicks(me, tui.Click(btn, handler))`
   - The framework does hit-testing automatically
   - Example: clickable buttons with visual feedback
9. **App-Level Key Handling** — Global handlers:
   - `tui.WithGlobalKeyHandler(fn)` — catch keys before component tree
   - `app.SetGlobalKeyHandler(fn)` — set/change at runtime

**Key Examples**:
- Basic KeyMap with quit, increment, decrement
- Keyboard explorer showing key name and modifiers
- Clickable color mixer with ref-based buttons
- Search input that captures all runes when focused

**Cross-references**: Link to [KeyMap Reference](../reference/events.md), [Refs Reference](../reference/refs.md), [State Guide](05-state.md)

---

#### Guide 08: Scrolling ✅
**File**: `docs/content/guides/08-scrolling.md`

**Purpose**: Teach scrollable containers — when content exceeds available space.

**Sections**:
1. **Overview** — When content is taller or wider than its container, go-tui supports scrollable regions with keyboard, mouse, and programmatic scroll control.
2. **Making an Element Scrollable** — The `scrollable` attribute:
   - Via class: `overflow-y-scroll`, `overflow-x-scroll`, `overflow-scroll`
   - Via attribute: `scrollable={tui.ScrollVertical}`, `scrollable={tui.ScrollHorizontal}`, `scrollable={tui.ScrollBoth}`
   - Content inside is laid out at its natural size, then clipped to the container
3. **Controlling Scroll Position** — State-driven scrolling:
   - Track position: `scrollY *tui.State[int]`
   - Bind to element: `scrollOffset={0, s.scrollY.Get()}`
   - Get a ref for bounds: `ref={s.content}`
4. **Keyboard Scrolling** — Key bindings:
   - Line-by-line: j/k or arrow keys
   - Page: PageUp/PageDown
   - Jump: Home/End
   - Helper method pattern: `scrollBy(delta int)` with clamping
5. **Mouse Scrolling** — Wheel events:
   - `HandleMouse`: check for `MouseWheelUp`/`MouseWheelDown`
   - Typical delta: 3 lines per wheel tick
6. **Bounds Checking** — Preventing over-scroll:
   - `ref.El().MaxScroll()` returns `(maxX, maxY int)`
   - Clamp scroll position: `max(0, min(newY, maxY))`
   - `ref.El().IsAtBottom()` — check if scrolled to bottom
7. **Auto-Scroll (Sticky Bottom)** — Following new content:
   - Track sticky state: `stickToBottom *tui.State[bool]`
   - On new content: if sticky, set scrollY to maxY
   - Manual scroll disables sticky; End key re-enables
   - Example: streaming log viewer
8. **Scrollbar Styling** — Customizing the scrollbar:
   - `scrollbarStyle={style}` — scrollbar track style
   - `scrollbarThumbStyle={style}` — scrollbar thumb style
9. **Scroll Methods** — Programmatic scroll control:
   - `el.ScrollTo(x, y)`, `el.ScrollBy(dx, dy)`
   - `el.ScrollToTop()`, `el.ScrollToBottom()`
   - `el.ScrollIntoView(child)` — ensure a child element is visible
   - `el.ScrollOffset()`, `el.ContentSize()`, `el.ViewportSize()`

**Key Examples**:
- Basic scrollable list with keyboard control
- Log viewer with auto-scroll and sticky toggle
- Styled scrollbar

**Cross-references**: Link to [Element Reference](../reference/element.md), [Events Guide](07-events.md)

---

#### Guide 09: Timers, Watchers, and Channels ✅
**File**: `docs/content/guides/09-watchers.md`

**Purpose**: Teach background operations — timers, Go channels, and the event bus.

**Sections**:
1. **Overview** — Components can run background operations via the `WatcherProvider` interface. Watchers integrate Go's concurrency primitives (timers, channels) with the UI event loop.
2. **The WatcherProvider Interface** — Returning watchers:
   - Implement `Watchers() []tui.Watcher` on your struct
   - Watchers start when the component mounts, stop when it unmounts
   - Watcher callbacks run on the UI thread (safe to update state)
3. **Timers** — `tui.OnTimer(interval, handler)`:
   - Periodic callback at the given interval
   - Handler signature: `func()`
   - Example: stopwatch, countdown timer
   - The timer keeps firing until the component unmounts
4. **Channel Watchers** — `tui.Watch(ch, handler)`:
   - Receives values from a Go channel
   - Handler signature: `func(T)` where T is the channel type
   - Example: receiving data from a goroutine producer
   - Channel can be created in main.go and passed to component
5. **Events Bus** — `tui.Events[T]`:
   - Create: `tui.NewEvents[string]()`
   - Emit: `events.Emit("something happened")`
   - Subscribe: `events.Subscribe(func(msg string) { ... })`
   - Use case: cross-component communication without shared state
   - Bind to app: `events.BindApp(app)` or `tui.NewEventsForApp(app)`
6. **Combining Watchers** — Multiple watchers per component:
   - Return multiple watchers from `Watchers()`
   - Timer + channel together
   - Example: dashboard with timer-based animation and channel-based data feed
7. **Thread Safety** — How watchers interact with the UI:
   - Watcher callbacks are queued on the event loop
   - Safe to call `state.Set()`, `state.Update()` from callbacks
   - No manual synchronization needed
   - For external goroutines, use `app.QueueUpdate(fn)` instead

**Key Examples**:
- Stopwatch using OnTimer
- Live data feed using Watch with a Go channel
- Events bus for component communication
- Dashboard combining timer animation and channel data

**Cross-references**: Link to [Watcher Reference](../reference/watchers.md), [State Guide](05-state.md), [App Reference](../reference/app.md)

---

#### Guide 10: Testing ✅
**File**: `docs/content/guides/10-testing.md`

**Purpose**: Teach how to test go-tui applications and components.

**Sections**:
1. **Overview** — go-tui provides `MockTerminal` and `MockEventReader` for testing components without a real terminal.
2. **MockTerminal** — Simulated terminal:
   - Create: `tui.NewMockTerminal(80, 24)` — 80 columns, 24 rows
   - Read output: `mock.String()` returns the full screen, `mock.StringTrimmed()` removes trailing whitespace
   - Read cells: `mock.CellAt(x, y)` returns the Cell at a position
   - Check state: `mock.IsInRawMode()`, `mock.IsInAltScreen()`, `mock.IsMouseEnabled()`, `mock.IsCursorHidden()`
   - Resize: `mock.Resize(width, height)`
   - Reset: `mock.Reset()` clears all state
   - Set capabilities: `mock.SetCaps(caps)` for testing color support
3. **MockEventReader** — Simulated input:
   - Create: `tui.NewMockEventReader(events...)` with pre-queued events
   - Add events: `reader.AddEvents(events...)`
   - Check remaining: `reader.Remaining()`
   - Events: `tui.KeyEvent{Key: tui.KeyEnter}`, `tui.KeyEvent{Rune: 'q'}`, etc.
4. **Testing Components** — Rendering and asserting:
   - Create app with mock terminal and reader: `tui.NewAppWithReader(reader, tui.WithRoot(...))`
   - Render once and check buffer
   - Pattern: create component, render, assert on string output
5. **Testing Key Handling** — Simulating input:
   - Queue key events in MockEventReader
   - Dispatch via `app.Dispatch(event)`
   - Assert state changes after key handling
6. **Table-Driven Tests** — go-tui's testing convention:
   - Define `tc` struct separately
   - Use `map[string]tc` for test cases
   - Run with `t.Run(name, func(t *testing.T) { ... })`
   - Example: testing a component with different initial states

**Key Examples**:
- Basic render test: create component, render, assert output contains expected text
- Key event test: dispatch keys, assert state changes
- Table-driven test for a counter component

**Cross-references**: Link to [Testing Reference](../reference/testing.md), [App Reference](../reference/app.md)

---

#### Guide 11: Multi-Component Applications ✅
**File**: `docs/content/guides/11-multi-component.md`

**Purpose**: Teach building larger apps with multiple struct components, shared state, and conditional behavior.

**Sections**:
1. **Overview** — Real applications have multiple components with their own state and behavior. go-tui supports composing struct components with shared state and conditional KeyMaps.
2. **Multiple .gsx Files** — Organizing code:
   - Each component can live in its own `.gsx` file
   - All files in the same package share scope
   - `tui generate` processes all `.gsx` files in a directory
3. **Shared State** — Passing state between components:
   - Parent creates `State[T]` and passes to child constructors
   - Multiple components can read/write the same state
   - Changes from any component trigger re-render
   - Example: sidebar selection state shared with content panel
4. **Conditional KeyMaps** — Context-dependent key handling:
   - KeyMap can return different bindings based on state
   - Example: search mode captures all runes, normal mode uses navigation keys
   - Use `OnRuneStop` to prevent propagation when capturing input
5. **Sub-Component Mounting** — `app.Mount()`:
   - Mount struct components within parent renders
   - Component lifecycle (init, cleanup) managed automatically
   - Props updates via `PropsUpdater` interface
6. **Architecture Patterns**:
   - Orchestrator pattern: one root component owns all state, children are display-only
   - Distributed state: each component owns relevant state, shared state for coordination
   - Example: file explorer with sidebar, content, and search components

**Key Examples**:
- File explorer: sidebar component + content component + search component
- Shared state between sidebar selection and content display
- Conditional KeyMap: search mode vs navigation mode

**Cross-references**: Link to [Components Guide](06-components.md), [State Guide](05-state.md)

---

#### Guide 12: Inline Mode and Alternate Screen ✅
**File**: `docs/content/guides/12-inline-mode.md`

**Purpose**: Teach inline mode (rendering at the bottom of existing terminal output) and alternate screen switching.

**Sections**:
1. **Overview** — By default, go-tui uses the full terminal (alternate screen). Inline mode renders a widget at the bottom of the terminal, preserving scroll history. Apps can switch between modes.
2. **Inline Mode** — Rendering inline:
   - Enable: `tui.WithInlineHeight(rows)` — widget occupies N rows at the bottom
   - Content above the widget is preserved (normal terminal scrollback)
   - Example: chat input that sits at the bottom
3. **PrintAbove** — Adding content above the widget:
   - `app.PrintAboveln(format, args...)` — print a line above the widget
   - `app.QueuePrintAboveln(format, args...)` — thread-safe version for callbacks
   - Lines scroll up naturally in the terminal
   - Example: chat messages appearing above an input widget
4. **Dynamic Height** — Growing the inline widget:
   - `app.SetInlineHeight(rows)` — change height at runtime
   - `app.InlineHeight()` — current height
   - Use case: text area that grows as content is typed
5. **Alternate Screen** — Full-screen overlay:
   - `app.EnterAlternateScreen()` — switch to full-screen mode
   - `app.ExitAlternateScreen()` — return to inline mode
   - `app.IsInAlternateScreen()` — check current mode
   - Use case: settings screen, help overlay
6. **Inline Startup Modes** — How the widget appears initially:
   - `tui.InlineStartupPreserveVisible` (default) — preserves visible terminal content
   - `tui.InlineStartupFreshViewport` — clears viewport first
   - `tui.InlineStartupSoftReset` — soft terminal reset
   - Set via: `tui.WithInlineStartupMode(mode)`
7. **Combining Inline + Alt Screen** — Multi-screen pattern:
   - Main UI in inline mode (e.g., chat input)
   - Settings/help in alternate screen
   - Toggle between them with a key binding

**Key Examples**:
- Inline chat input with messages above
- Settings overlay using alternate screen
- Full chat app pattern: inline + alt screen for settings

**Cross-references**: Link to [App Reference](../reference/app.md), [TextArea Reference](../reference/built-in-components.md)

---

#### Guide 13: Focus Management ✅
**File**: `docs/content/guides/13-focus.md`

**Purpose**: Teach focus management — making elements focusable, Tab/Shift-Tab navigation, and focus groups.

**Sections**:
1. **Overview** — Focus determines which element receives keyboard input. go-tui provides automatic focus management with Tab/Shift-Tab cycling and programmatic control.
2. **Making Elements Focusable** — The `focusable` attribute:
   - Set via class or attribute: `focusable={true}`
   - Focus callbacks: `onFocus={fn}`, `onBlur={fn}`
   - `WithOnFocus`, `WithOnBlur` option functions
   - Visual feedback: check `IsFocused()` to apply different styles
3. **Focus Navigation** — Tab cycling:
   - `app.FocusNext()` — move to next focusable (Tab)
   - `app.FocusPrev()` — move to previous focusable (Shift-Tab)
   - `app.Focused()` — get currently focused element
   - Focus order follows document order (DFS traversal of element tree)
4. **FocusGroup** — State-based focus management:
   - Create: `tui.MustNewFocusGroup(state1, state2, state3)` with `*State[bool]` members
   - `fg.Next()`, `fg.Prev()`, `fg.Current()`
   - `fg.KeyMap()` — returns Tab/Shift-Tab bindings
   - Each member state is true when focused, false otherwise
   - Use case: form with multiple focusable sections
5. **FocusManager** — Lower-level API:
   - `tui.NewFocusManager()` — create a focus manager
   - `fm.Register(elem)`, `fm.Unregister(elem)`
   - `fm.SetFocus(elem)`, `fm.Focused()`
   - `fm.Dispatch(event)` — route events to focused element
6. **Programmatic Focus** — Setting focus directly:
   - `element.Focus()`, `element.Blur()`
   - `element.IsFocused()`, `element.IsFocusable()`
   - `element.SetFocusable(bool)` — change focusability at runtime

**Key Examples**:
- Form with multiple focusable fields, Tab navigation
- FocusGroup controlling section highlight
- Custom focus indicators (border color change on focus)

**Cross-references**: Link to [Focus Reference](../reference/focus.md), [Events Guide](07-events.md)

---

#### Guide 14: Building a Dashboard (Capstone) ✅
**File**: `docs/content/guides/14-dashboard.md`

**Purpose**: Capstone tutorial that brings everything together — build a live metrics dashboard from scratch, exercising all major concepts.

**Sections**:
1. **What We're Building** — A live dashboard with CPU/memory/disk gauges, network sparkline, event feed. ASCII art mockup of the final result.
2. **Project Setup** — Create directory, init module, create main.go and dashboard.gsx
3. **Layout Skeleton** — Start with the overall layout:
   - Outer container with border
   - Top row: three metric panels
   - Middle: network chart
   - Bottom: event feed
   - Build this step by step, running after each addition
4. **Adding Metrics** — State and timers:
   - Create `State[int]` for CPU, memory, disk
   - Add `OnTimer` watcher for animation
   - Build bar visualization helper
   - Dynamic color based on value
5. **Network Sparkline** — Channel-based data:
   - `State[[]int]` for sparkline data
   - Map values to block characters (▁▂▃▄▅▆▇█)
   - Gradient text for visual appeal
6. **Event Feed** — Channel watcher:
   - Channel producer in main.go
   - `Watch(ch, handler)` to receive events
   - Scrollable event list, newest first
7. **Polish** — Final touches:
   - Gradient borders and titles
   - Key hint bar
   - Clean quit handling
8. **Full Code** — Complete listing of the finished dashboard

**Key Examples**:
- Complete dashboard.gsx and main.go
- Each section shown incrementally as it's built

**Cross-references**: Links to all relevant guides and references

---

### Reference Documents

Reference documents are exhaustive API documentation. Every exported type, function, constant, and method gets documented with its signature, description, and at least one example.

---

#### Reference 01: App ✅
**File**: `docs/content/reference/app.md`

**Purpose**: Complete reference for the App type — creation, configuration, lifecycle, and methods.

**Sections**:
1. **Overview** — The App is the top-level container that manages the event loop, rendering, and component tree.
2. **Creating an App**:
   - `tui.NewApp(opts ...AppOption) (*App, error)` — create with options
   - `tui.NewAppWithReader(reader EventReader, opts ...AppOption) (*App, error)` — create with custom input reader (for testing)
3. **AppOption Functions** — Each option documented with signature, description, default value, example:
   - `WithRoot(root Renderable)` — set root element
   - `WithRootView(view Viewable)` — set root with watchers
   - `WithRootComponent(component Component)` — set root as a component (most common)
   - `WithFrameRate(fps int)` — set render frame rate (default: 60)
   - `WithInputLatency(d time.Duration)` — input batching window
   - `WithEventQueueSize(size int)` — event queue buffer size
   - `WithGlobalKeyHandler(fn func(KeyEvent) bool)` — global key handler
   - `WithMouse()` — enable mouse events
   - `WithoutMouse()` — disable mouse events (default)
   - `WithCursor()` — show cursor
   - `WithInlineHeight(rows int)` — inline mode with N rows
   - `WithInlineStartupMode(mode InlineStartupMode)` — inline startup behavior
4. **Lifecycle Methods**:
   - `Run() error` — start the event loop (blocks until Stop)
   - `Stop()` — stop the event loop
   - `Close() error` — clean up resources (raw mode, alt screen)
5. **Root Management**:
   - `SetRoot(root Renderable)` — change root element
   - `SetRootView(view Viewable)` — change root view
   - `SetRootComponent(component Component)` — change root component
   - `Root() Renderable` — get current root
6. **Rendering**:
   - `Render()` — render if dirty
   - `RenderFull()` — force full redraw
   - `MarkDirty()` — mark for re-render
   - `Buffer() *Buffer` — access the render buffer
   - `Size() (width, height int)` — terminal dimensions
7. **Event Handling**:
   - `Dispatch(event Event) bool` — dispatch an event
   - `SetGlobalKeyHandler(fn func(KeyEvent) bool)` — set global key handler
   - `QueueUpdate(fn func())` — queue a function on the event loop (thread-safe)
   - `PollEvent(timeout time.Duration) (Event, bool)` — poll for input events
8. **Focus**:
   - `FocusNext()` — focus next focusable element
   - `FocusPrev()` — focus previous focusable element
   - `Focused() Focusable` — get focused element
   - `Focus() *FocusManager` — get focus manager (deprecated)
9. **Inline Mode**:
   - `PrintAbove(format string, args ...any)` — print above widget (must be called from event loop)
   - `QueuePrintAbove(format string, args ...any)` — thread-safe PrintAbove
   - `PrintAboveln(format string, args ...any)` — PrintAbove with newline
   - `QueuePrintAboveln(format string, args ...any)` — thread-safe PrintAboveln
   - `SetInlineHeight(rows int)` — change inline height
   - `InlineHeight() int` — current inline height
10. **Alternate Screen**:
    - `EnterAlternateScreen() error` — switch to alt screen
    - `ExitAlternateScreen() error` — return from alt screen
    - `IsInAlternateScreen() bool` — check current mode
11. **State**:
    - `Batch(fn func())` — batch multiple state changes into one render
12. **Component Mounting**:
    - `Mount(parent Component, index int, factory func() Component) *Element` — mount a sub-component
13. **Other**:
    - `Terminal() Terminal` — access the terminal
    - `EventQueue() chan<- func()` — access the event queue channel
    - `StopCh() <-chan struct{}` — channel closed when app stops
    - `SnapshotFrame() string` — capture current frame as string
    - `InputLatencyBlocking` constant — blocking input mode

**InlineStartupMode Constants**:
- `InlineStartupPreserveVisible` — preserve visible terminal content
- `InlineStartupFreshViewport` — clear viewport
- `InlineStartupSoftReset` — soft terminal reset

---

#### Reference 02: Element ✅
**File**: `docs/content/reference/element.md`

**Purpose**: Complete reference for the Element type — the building block of all UIs.

**Sections**:
1. **Overview** — Element is the fundamental UI node. Created via `tui.New(opts...)` or generated from `.gsx` templates.
2. **Creating Elements**:
   - `tui.New(opts ...Option) *Element` — create with options
   - In `.gsx`: `<div class="...">content</div>`
3. **Option Functions** — Full list, grouped by category. Each with signature, description, example:
   - **Dimensions**: `WithWidth(int)`, `WithWidthPercent(float64)`, `WithWidthAuto()`, `WithHeight(int)`, `WithHeightPercent(float64)`, `WithHeightAuto()`, `WithSize(w, h int)`, `WithMinWidth(int)`, `WithMinHeight(int)`, `WithMaxWidth(int)`, `WithMaxHeight(int)`
   - **Flex Container**: `WithDirection(Direction)`, `WithJustify(Justify)`, `WithAlign(Align)`, `WithGap(int)`
   - **Flex Item**: `WithFlexGrow(float64)`, `WithFlexShrink(float64)`, `WithAlignSelf(Align)`
   - **Spacing**: `WithPadding(int)`, `WithPaddingTRBL(t,r,b,l int)`, `WithMargin(int)`, `WithMarginTRBL(t,r,b,l int)`
   - **Visual**: `WithBorder(BorderStyle)`, `WithBorderStyle(Style)`, `WithBackground(Style)`, `WithText(string)`, `WithTextStyle(Style)`, `WithTextAlign(TextAlign)`, `WithTextGradient(Gradient)`, `WithBackgroundGradient(Gradient)`, `WithBorderGradient(Gradient)`
   - **Focus**: `WithFocusable(bool)`, `WithOnFocus(func(*Element))`, `WithOnBlur(func(*Element))`
   - **Scroll**: `WithScrollable(ScrollMode)`, `WithScrollOffset(x, y int)`, `WithScrollbarStyle(Style)`, `WithScrollbarThumbStyle(Style)`
   - **Behavior**: `WithHR()`, `WithTruncate(bool)`, `WithHidden(bool)`, `WithOverflow(OverflowMode)`, `WithOnUpdate(func())`
4. **Accessors** — Get/set methods:
   - Text: `Text() string`, `SetText(string)`
   - Style: `Style() LayoutStyle`, `SetStyle(LayoutStyle)`
   - Border: `Border() BorderStyle`, `SetBorder(BorderStyle)`, `BorderStyle() Style`, `SetBorderStyle(Style)`
   - Background: `Background() *Style`, `SetBackground(*Style)`
   - Text style: `TextStyle() Style`, `SetTextStyle(Style)`, `TextAlign() TextAlign`, `SetTextAlign(TextAlign)`
   - Truncate: `Truncate() bool`, `SetTruncate(bool)`
   - Hidden: `Hidden() bool`, `SetHidden(bool)`
   - Overflow: `Overflow() OverflowMode`, `SetOverflow(OverflowMode)`
5. **Tree Methods**:
   - `AddChild(children ...*Element)`, `RemoveChild(child) bool`, `RemoveAllChildren()`
   - `Children() []*Element`, `Parent() *Element`
   - `SetOnChildAdded(func(*Element))`
6. **Layout Methods**:
   - `LayoutStyle() LayoutStyle`, `LayoutChildren() []Layoutable`
   - `SetLayout(LayoutResult)`, `GetLayout() LayoutResult`
   - `Calculate(availableWidth, availableHeight int)`
   - `Rect() Rect`, `ContentRect() Rect`
   - `IntrinsicSize() (width, height int)`
   - `IsHR() bool`
7. **Focus Methods**:
   - `IsFocusable() bool`, `IsFocused() bool`, `SetFocusable(bool)`
   - `Focus()`, `Blur()`
   - `HandleEvent(Event) bool`
   - `SetOnFocus(func(*Element))`, `SetOnBlur(func(*Element))`
   - `ContainsPoint(x, y int) bool`
8. **Scroll Methods**:
   - `IsScrollable() bool`, `ScrollModeValue() ScrollMode`
   - `ScrollOffset() (x, y int)`, `ContentSize() (w, h int)`, `ViewportSize() (w, h int)`, `MaxScroll() (maxX, maxY int)`
   - `ScrollTo(x, y int)`, `ScrollBy(dx, dy int)`, `ScrollToTop()`, `ScrollToBottom()`
   - `IsAtBottom() bool`, `ScrollIntoView(child *Element)`
9. **Rendering**:
   - `Render(buf *Buffer, width, height int)`
   - `RenderTree(buf *Buffer, root *Element)` (package-level)
   - `IsDirty() bool`, `SetDirty(bool)`, `MarkDirty()`
10. **Watcher/Discovery**:
    - `AddWatcher(Watcher)`, `Watchers() []Watcher`, `WalkWatchers(func(Watcher))`
    - `WalkFocusables(func(Focusable))`, `SetOnFocusableAdded(func(Focusable))`
    - `SetOnUpdate(func())`
    - `ElementAt(x, y int) *Element`, `ElementAtPoint(x, y int) Focusable`
11. **Enums**:
    - `TextAlign`: `TextAlignLeft`, `TextAlignCenter`, `TextAlignRight`
    - `ScrollMode`: `ScrollNone`, `ScrollVertical`, `ScrollHorizontal`, `ScrollBoth`
    - `OverflowMode`: `OverflowVisible`, `OverflowHidden`

---

#### Reference 03: Layout ✅
**File**: `docs/content/reference/layout.md`

**Purpose**: Complete reference for the layout system types and algorithm.

**Sections**:
1. **Overview** — go-tui implements CSS flexbox layout. All types are re-exported from `internal/layout`.
2. **Direction**:
   - `tui.Row` — horizontal (left to right)
   - `tui.Column` — vertical (top to bottom)
3. **Justify** — Main axis alignment:
   - `JustifyStart`, `JustifyCenter`, `JustifyEnd`, `JustifySpaceBetween`, `JustifySpaceAround`, `JustifySpaceEvenly`
   - Visual diagram for each
4. **Align** — Cross axis alignment:
   - `AlignStart`, `AlignCenter`, `AlignEnd`, `AlignStretch`
   - Visual diagram for each
5. **Value** — Dimension specification:
   - `tui.Fixed(n int)` — exact character count
   - `tui.Percent(p float64)` — percentage of available space
   - `tui.Auto()` — size to content
6. **LayoutStyle** — Full layout configuration struct:
   - All fields documented: Direction, Justify, Align, Width, Height, MinWidth, MinHeight, MaxWidth, MaxHeight, FlexGrow, FlexShrink, AlignSelf, Padding, Margin, Gap
   - `tui.DefaultLayoutStyle()` — returns defaults
7. **Rect** — Rectangle type:
   - `tui.NewRect(x, y, width, height int)`
   - Fields: X, Y, Width, Height
   - `InsetRect(r, top, right, bottom, left int)`, `InsetUniform(r, n int)`
8. **Edges** — TRBL spacing:
   - `EdgeAll(n int)` — same on all sides
   - `EdgeSymmetric(v, h int)` — vertical and horizontal
   - `EdgeTRBL(t, r, b, l int)` — individual sides
   - Fields: Top, Right, Bottom, Left
9. **Other Types**:
   - `Size` — Width, Height pair
   - `Point` — X, Y pair
   - `LayoutResult` — computed layout result
   - `Layoutable` — interface for layout participation
10. **Calculate** — Layout computation:
    - `tui.Calculate(root Layoutable, availableWidth, availableHeight int)`
    - Runs the flexbox algorithm on the tree

---

#### Reference 04: Styling ✅
**File**: `docs/content/reference/styling.md`

**Purpose**: Complete reference for Style, Color, Gradient, and BorderStyle types.

**Sections**:
1. **Style Type**:
   - `tui.NewStyle() Style` — zero-value style
   - Fields: `Fg Color`, `Bg Color`, `Attrs Attr`
   - Chainable methods: `Foreground(Color)`, `Background(Color)`, `Bold()`, `Dim()`, `Italic()`, `Underline()`, `Blink()`, `Reverse()`, `Strikethrough()`
   - Query: `HasAttr(Attr) bool`, `Equal(Style) bool`
2. **Attr Flags**:
   - `AttrNone`, `AttrBold`, `AttrDim`, `AttrItalic`, `AttrUnderline`, `AttrBlink`, `AttrReverse`, `AttrStrikethrough`
   - Can be combined with `|`
3. **Color Type**:
   - Constructors: `DefaultColor()`, `ANSIColor(uint8)`, `RGBColor(r, g, b uint8)`, `HexColor(string) (Color, error)`
   - Query: `Type() ColorType`, `IsDefault() bool`, `ANSI() uint8`, `RGB() (r, g, b uint8)`, `Equal(Color) bool`
   - Utility: `ToANSI() Color`, `ToRGBValues() (r, g, b uint8)`, `Luminance() float64`, `IsLight() bool`
4. **Color Constants**:
   - Standard (0-7): `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White`
   - Bright (8-15): `BrightBlack`, `BrightRed`, `BrightGreen`, `BrightYellow`, `BrightBlue`, `BrightMagenta`, `BrightCyan`, `BrightWhite`
5. **ColorType Enum**:
   - `ColorDefault`, `ColorANSI`, `ColorRGB`
6. **Gradient Type**:
   - `tui.NewGradient(start, end Color) Gradient`
   - `gradient.WithDirection(GradientDirection) Gradient`
   - `gradient.At(t float64) Color` — interpolate at position t (0.0-1.0)
   - Directions: `GradientHorizontal`, `GradientVertical`, `GradientDiagonalDown`, `GradientDiagonalUp`
7. **BorderStyle Enum**:
   - `BorderNone`, `BorderSingle`, `BorderDouble`, `BorderRounded`, `BorderThick`
   - `border.Chars() BorderChars` — get the actual rune characters
   - `BorderChars` struct: TopLeft, Top, TopRight, Left, Right, BottomLeft, Bottom, BottomRight
8. **Drawing Functions**:
   - `DrawBox(buf, rect, border, style)`
   - `DrawBoxGradient(buf, rect, border, gradient, baseStyle)`
   - `DrawBoxClipped(buf, rect, border, style, clipRect)`
   - `DrawBoxGradientClipped(buf, rect, border, gradient, baseStyle, clipRect)`
   - `DrawBoxWithTitle(buf, rect, border, title, style)`
   - `FillBox(buf, rect, rune, style)`
9. **Capabilities**:
   - `tui.DetectCapabilities() Capabilities`
   - `Capabilities` struct: Colors, Unicode, TrueColor, AltScreen
   - `ColorCapability` enum: `ColorNone`, `Color16`, `Color256`, `ColorTrue`
   - `caps.SupportsColor(Color) bool`, `caps.EffectiveColor(Color) Color`

---

#### Reference 05: State ✅
**File**: `docs/content/reference/state.md`

**Purpose**: Complete reference for State[T] and Events[T] generic types.

**Sections**:
1. **State[T]** — Reactive state container:
   - `tui.NewState[T](initial T) *State[T]` — create unbound state
   - `tui.NewStateForApp[T](app *App, initial T) *State[T]` — create app-bound state
   - `state.BindApp(app *App)` — bind to app for dirty marking
   - `state.Get() T` — read current value
   - `state.Set(v T)` — set new value (marks dirty)
   - `state.Update(fn func(T) T)` — functional update (marks dirty)
   - `state.Bind(fn func(T)) Unbind` — register change callback, returns unbind function
   - `Unbind` type: `type Unbind func()` — call to unregister
2. **Events[T]** — Generic event bus:
   - `tui.NewEvents[T]() *Events[T]` — create unbound event bus
   - `tui.NewEventsForApp[T](app *App) *Events[T]` — create app-bound event bus
   - `events.BindApp(app *App)` — bind to app
   - `events.Emit(event T)` — emit an event to all subscribers
   - `events.Subscribe(fn func(T))` — register event handler
3. **Batching**:
   - `app.Batch(fn func())` — batch multiple state changes
   - All changes within fn coalesce into a single re-render

---

#### Reference 06: Events ✅
**File**: `docs/content/reference/events.md`

**Purpose**: Complete reference for Event types, Key constants, and Modifier flags.

**Sections**:
1. **Event Interface** — Marker interface for all events
2. **KeyEvent**:
   - Fields: `Key Key`, `Rune rune`, `Mod Modifier`
   - Methods: `App() *App`, `IsRune() bool`, `Is(Key, ...Modifier) bool`, `Char() rune`
3. **MouseEvent**:
   - Fields: `Button MouseButton`, `Action MouseAction`, `X int`, `Y int`, `Mod Modifier`
   - Method: `App() *App`
4. **ResizeEvent**:
   - Fields: `Width int`, `Height int`
5. **Key Constants** — Full list:
   - `KeyNone`, `KeyRune`, `KeyEscape`, `KeyEnter`, `KeyTab`, `KeyBackspace`, `KeyDelete`, `KeyInsert`
   - `KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`, `KeyHome`, `KeyEnd`, `KeyPageUp`, `KeyPageDown`
   - `KeyF1` through `KeyF12`
   - `KeyCtrlA` through `KeyCtrlZ`, `KeyCtrlSpace`
   - `key.String() string`
6. **Modifier Flags**:
   - `ModNone`, `ModCtrl`, `ModAlt`, `ModShift`
   - `mod.Has(Modifier) bool`, `mod.String() string`
7. **MouseButton Constants**:
   - `MouseLeft`, `MouseMiddle`, `MouseRight`, `MouseWheelUp`, `MouseWheelDown`, `MouseNone`
8. **MouseAction Constants**:
   - `MousePress`, `MouseRelease`, `MouseDrag`
9. **KeyMap Type**:
   - `tui.KeyMap` — `[]KeyBinding`
   - `KeyBinding` struct: `Pattern KeyPattern`, `Handler func(KeyEvent)`, `Stop bool`
   - `KeyPattern` struct: `Key`, `Rune`, `AnyRune`, `Mod`, `RequireNoMods`
10. **KeyMap Helpers**:
    - `OnKey(Key, func(KeyEvent)) KeyBinding`
    - `OnKeyStop(Key, func(KeyEvent)) KeyBinding`
    - `OnRune(rune, func(KeyEvent)) KeyBinding`
    - `OnRuneStop(rune, func(KeyEvent)) KeyBinding`
    - `OnRunes(func(KeyEvent)) KeyBinding`
    - `OnRunesStop(func(KeyEvent)) KeyBinding`

---

#### Reference 07: Focus ✅
**File**: `docs/content/reference/focus.md`

**Purpose**: Complete reference for focus management types.

**Sections**:
1. **Focusable Interface**:
   - `IsFocusable() bool`, `HandleEvent(Event) bool`, `Focus()`, `Blur()`
2. **FocusManager**:
   - `tui.NewFocusManager() *FocusManager`
   - `Register(Focusable)`, `Unregister(Focusable)`
   - `Focused() Focusable`, `SetFocus(Focusable)`
   - `Next()`, `Prev()`
   - `Dispatch(Event) bool`
3. **FocusGroup**:
   - `tui.NewFocusGroup(members ...*State[bool]) (*FocusGroup, error)`
   - `tui.MustNewFocusGroup(members ...*State[bool]) *FocusGroup`
   - `Next()`, `Prev()`, `Current() int`
   - `KeyMap() KeyMap` — returns Tab/Shift-Tab bindings

---

#### Reference 08: Watchers ✅
**File**: `docs/content/reference/watchers.md`

**Purpose**: Complete reference for Watcher types.

**Sections**:
1. **Watcher Interface**:
   - `Start(eventQueue chan<- func(), stopCh <-chan struct{})` — starts the watcher
2. **OnTimer**:
   - `tui.OnTimer(interval time.Duration, handler func()) Watcher`
   - Fires handler at the given interval
3. **Channel Watcher**:
   - `tui.Watch[T](ch <-chan T, handler func(T)) Watcher`
   - `tui.NewChannelWatcher[T](ch <-chan T, fn func(T)) *ChannelWatcher[T]`
   - Receives from channel and calls handler on UI thread
4. **WatcherProvider Interface**:
   - `Watchers() []Watcher` — return watchers for this component

---

#### Reference 09: Refs ✅
**File**: `docs/content/reference/refs.md`

**Purpose**: Complete reference for Ref, RefList, RefMap, and click handling.

**Sections**:
1. **Ref** — Single element reference:
   - `tui.NewRef() *Ref`
   - `ref.Set(*Element)`, `ref.El() *Element`, `ref.IsSet() bool`
   - Usage in .gsx: `ref={myRef}`
2. **RefList** — Ordered collection (for loops):
   - `tui.NewRefList() *RefList`
   - `refs.Append(*Element)`, `refs.All() []*Element`, `refs.At(i int) *Element`, `refs.Len() int`
3. **RefMap[K]** — Keyed collection:
   - `tui.NewRefMap[K comparable]() *RefMap[K]`
   - `refMap.Put(key, *Element)`, `refMap.Get(key) *Element`, `refMap.All() map[K]*Element`, `refMap.Len() int`
4. **Click Handling**:
   - `tui.Click(ref *Ref, fn func()) ClickBinding` — create a click binding
   - `tui.HandleClicks(me MouseEvent, bindings ...ClickBinding) bool` — process click against refs
   - `ClickBinding` struct: `Ref *Ref`, `Fn func()`

---

#### Reference 10: Buffer and Rendering ✅
**File**: `docs/content/reference/buffer.md`

**Purpose**: Complete reference for Buffer, Cell, and rendering functions.

**Sections**:
1. **Buffer** — Double-buffered character grid:
   - `tui.NewBuffer(width, height int) *Buffer`
   - Query: `Width()`, `Height()`, `Size()`, `Rect()`
   - Read: `Cell(x, y int) Cell`
   - Write: `SetCell(x, y, Cell)`, `SetRune(x, y, rune, Style)`, `SetString(x, y, string, Style) int`, `SetStringClipped(x, y, string, Style, Rect) int`
   - Gradient: `SetStringGradient(x, y, string, Gradient, Style) int`, `FillGradient(Rect, rune, Gradient, Style)`
   - Fill: `Fill(Rect, rune, Style)`, `Clear()`, `ClearRect(Rect)`
   - Diff: `Diff() []CellChange`, `Swap()`
   - Output: `String() string`, `StringTrimmed() string`
   - Resize: `Resize(width, height int)`
2. **Cell** — Single character:
   - `tui.NewCell(rune, Style) Cell`, `tui.NewCellWithWidth(rune, Style, uint8) Cell`
   - Fields: `Rune rune`, `Style Style`, `Width uint8`
   - Methods: `IsContinuation() bool`, `Equal(Cell) bool`, `IsEmpty() bool`
   - `tui.RuneWidth(rune) int` — character width (CJK support)
3. **CellChange** — Diff entry:
   - Fields: `X int`, `Y int`, `Cell Cell`
4. **Render Functions**:
   - `tui.Render(Terminal, *Buffer)` — diff-based render
   - `tui.RenderFull(Terminal, *Buffer)` — full redraw
   - `tui.RenderTree(*Buffer, *Element)` — render element tree to buffer

---

#### Reference 11: Terminal ✅
**File**: `docs/content/reference/terminal.md`

**Purpose**: Complete reference for the Terminal interface and implementations.

**Sections**:
1. **Terminal Interface** — All methods documented:
   - `Size()`, `Flush([]CellChange)`, `Clear()`, `ClearToEnd()`
   - `SetCursor(x, y)`, `HideCursor()`, `ShowCursor()`
   - `EnterRawMode()`, `ExitRawMode()`
   - `EnterAltScreen()`, `ExitAltScreen()`
   - `EnableMouse()`, `DisableMouse()`
   - `Caps() Capabilities`
   - `WriteDirect([]byte) (int, error)`
2. **ANSITerminal** — Real terminal implementation:
   - `tui.NewANSITerminal(out io.Writer, in io.Reader) (*ANSITerminal, error)`
   - `tui.NewANSITerminalWithCaps(out, in, Capabilities) *ANSITerminal`
   - Additional methods: `SetCaps()`, `ResetStyle()`, `Writer()`, `BeginSyncUpdate()`, `EndSyncUpdate()`
3. **BufferedWriter**:
   - `tui.NewBufferedWriter(out io.Writer) *BufferedWriter`
   - `Write([]byte) (int, error)`, `Flush() error`
4. **EventReader** — Input reading:
   - `tui.NewEventReader(in *os.File) (EventReader, error)`
   - `EventReader` interface: `PollEvent(timeout) (Event, bool)`, `Close() error`
   - `InterruptibleReader` interface: adds `EnableInterrupt()`, `Interrupt()`

---

#### Reference 12: Built-in Components ✅
**File**: `docs/content/reference/built-in-components.md`

**Purpose**: Complete reference for TextArea and any other built-in widget components.

**Sections**:
1. **TextArea** — Multi-line text input:
   - `tui.NewTextArea(opts ...TextAreaOption) *TextArea`
   - Implements: `Component`, `KeyListener`, `WatcherProvider`, `Focusable`, `AppBinder`
   - Methods: `Text() string`, `SetText(string)`, `Clear()`, `Height() int`
   - Focus: `IsFocusable() bool`, `Focus()`, `Blur()`
   - App binding: `BindApp(app *App)`
2. **TextAreaOption Functions**:
   - `WithTextAreaWidth(int)` — fixed width
   - `WithTextAreaMaxHeight(int)` — max rows
   - `WithTextAreaBorder(BorderStyle)` — border style
   - `WithTextAreaTextStyle(Style)` — text style
   - `WithTextAreaPlaceholder(string)` — placeholder text
   - `WithTextAreaPlaceholderStyle(Style)` — placeholder style
   - `WithTextAreaCursor(rune)` — cursor character
   - `WithTextAreaSubmitKey(Key)` — submit trigger key
   - `WithTextAreaOnSubmit(func(string))` — submit handler

---

#### Reference 13: Component Interfaces ✅
**File**: `docs/content/reference/interfaces.md`

**Purpose**: Complete reference for all interfaces that users can implement.

**Sections**:
1. **Component** — `Render(app *App) *Element`
2. **Renderable** — `Render(buf *Buffer, width, height int)`, `MarkDirty()`, `IsDirty() bool`
3. **Viewable** — `GetRoot() Renderable`, `GetWatchers() []Watcher`
4. **KeyListener** — `KeyMap() KeyMap`
5. **MouseListener** — `HandleMouse(MouseEvent) bool`
6. **WatcherProvider** — `Watchers() []Watcher`
7. **Initializer** — `Init() func()` (returns cleanup function)
8. **AppBinder** — `BindApp(app *App)`
9. **PropsUpdater** — `UpdateProps(fresh Component)`
10. **Focusable** — `IsFocusable() bool`, `HandleEvent(Event) bool`, `Focus()`, `Blur()`
11. **Watcher** — `Start(eventQueue chan<- func(), stopCh <-chan struct{})`
12. **Terminal** — Full interface (see Terminal Reference)
13. **EventReader** — `PollEvent(timeout) (Event, bool)`, `Close() error`
14. **InterruptibleReader** — extends EventReader with `EnableInterrupt()`, `Interrupt()`

For each interface: signature, when to implement, example implementation.

---

#### Reference 14: GSX Syntax Reference ✅
**File**: `docs/content/reference/gsx-syntax.md`

**Purpose**: Complete reference for all .gsx syntax — elements, attributes, control flow, and Tailwind classes.

**Sections**:
1. **File Structure** — Package, imports, components, functions
2. **Element Reference** — Each built-in element with its behavior:
   - `<div>` — Block flex container, default direction Row
   - `<span>` — Inline text container
   - `<p>` — Paragraph with text wrapping
   - `<ul>` — Unordered list container
   - `<li>` — List item with bullet
   - `<button>` — Clickable element
   - `<input />` — Text input (self-closing), attributes: `value`, `placeholder`
   - `<table>` — Table container
   - `<progress />` — Progress bar (self-closing), attributes: `value`, `max`
   - `<hr />` — Horizontal rule (self-closing)
   - `<br />` — Line break (self-closing)
3. **Attribute Reference** — Complete list organized by category (same as Framework Context above but with detailed types and descriptions)
4. **Control Flow** — `@if`, `@else`, `@for`, `@let` with formal syntax
5. **Tailwind Class Reference** — Complete class listing (same as Framework Context above) with the generated Option function each maps to
6. **Component Syntax**:
   - Pure: `templ Name(params) { ... }`
   - Struct method: `templ (s *Struct) MethodName() { ... }`
   - Calling: `<ComponentName param={value}>children</ComponentName>`
   - Children: `{children...}`
7. **Go Expressions** — `{expr}` syntax in text and attributes

---

#### Reference 15: Testing Utilities ✅
**File**: `docs/content/reference/testing.md`

**Purpose**: Complete reference for MockTerminal and MockEventReader.

**Sections**:
1. **MockTerminal**:
   - `tui.NewMockTerminal(width, height int) *MockTerminal`
   - Implements `Terminal` interface
   - Test helpers: `CellAt(x, y) Cell`, `String() string`, `StringTrimmed() string`
   - Cursor: `Cursor() (x, y int)`, `IsCursorHidden() bool`
   - Mode checks: `IsInRawMode() bool`, `IsInAltScreen() bool`, `IsMouseEnabled() bool`
   - Alt screen: `AltScreenEnterCount() int`, `AltScreenExitCount() int`
   - State: `Reset()`, `Resize(width, height int)`, `SetCell(x, y, Cell)`, `SetCaps(Capabilities)`
2. **MockEventReader**:
   - `tui.NewMockEventReader(events ...Event) *MockEventReader`
   - Implements `EventReader` and `InterruptibleReader`
   - Methods: `AddEvents(events...)`, `Remaining() int`, `Reset()`

---

#### Reference 16: CLI Reference ✅
**File**: `docs/content/reference/cli.md`

**Purpose**: Complete reference for the `tui` command-line tool.

**Sections**:
1. **Installation** — `go install github.com/grindlemire/go-tui/cmd/tui@latest`
2. **Commands**:
   - `tui generate [path...]` — Generate Go code from .gsx files
   - `tui check [path...]` — Validate .gsx files without generating
   - `tui fmt [path...]` — Format .gsx files
   - `tui fmt --check [path...]` — Check formatting without modifying
   - `tui fmt --stdout [path...]` — Write formatted output to stdout
   - `tui lsp` — Start language server (stdio)
   - `tui lsp --log FILE` — Start LSP with logging
   - `tui version` — Print version
   - `tui help` — Show help
3. **Editor Integration** — LSP setup for VS Code, Neovim, etc.

---

## File Tree

```
docs/content/
├── guides/
│   ├── 01-getting-started.md
│   ├── 02-gsx-syntax.md
│   ├── 03-styling.md
│   ├── 04-layout.md
│   ├── 05-state.md
│   ├── 06-components.md
│   ├── 07-events.md
│   ├── 08-scrolling.md
│   ├── 09-watchers.md
│   ├── 10-testing.md
│   ├── 11-multi-component.md
│   ├── 12-inline-mode.md
│   ├── 13-focus.md
│   └── 14-dashboard.md
└── reference/
    ├── app.md
    ├── element.md
    ├── layout.md
    ├── styling.md
    ├── state.md
    ├── events.md
    ├── focus.md
    ├── watchers.md
    ├── refs.md
    ├── buffer.md
    ├── terminal.md
    ├── built-in-components.md
    ├── interfaces.md
    ├── gsx-syntax.md
    ├── testing.md
    └── cli.md
```

## Execution Plan

Each document is independent. An agent writing any single document needs:
1. This spec file (for the document's section outline and the Framework Context)
2. Access to the source code files listed in "Source Code Locations" for API accuracy
3. Access to the `examples/` directory for working code examples

Documents can be written in any order. The guides reference each other via cross-links, but each is self-contained enough to be written independently.

For each document, the agent should:
1. Read the relevant source files to verify API signatures
2. Write the document following the section outline
3. Include complete, runnable code examples (not fragments)
4. Add cross-reference links to related docs
5. Save to the specified file path
