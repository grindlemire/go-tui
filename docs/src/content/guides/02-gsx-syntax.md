# GSX Syntax

## Overview

`.gsx` files are Go files extended with a templ-like syntax for declaring UI. They look like regular Go with the same package declarations, imports, and type definitions, but they add a `templ` keyword for defining components that return element trees. The `tui generate` command reads `.gsx` files and produces standard `_gsx.go` files that call the `tui` package API. You never need to edit the generated files.

## Package and Imports

A `.gsx` file starts with a package declaration and imports, just like any Go file:

```gsx
package main

import (
    "fmt"
    "strings"

    tui "github.com/grindlemire/go-tui"
)
```

The convention is to alias the import as `tui`. This keeps element option calls readable (`tui.NewState`, `tui.KeyEscape`, etc.) and is the style used throughout the framework's own examples.

Everything else in the file — type declarations, constants, variables, helper functions — follows normal Go syntax.

## Pure Components

Pure components are stateless functions declared with the `templ` keyword. They take parameters, return an element tree, and have no lifecycle of their own.

```gsx
templ Greeting(name string) {
    <span class="text-cyan font-bold">{"Hello, " + name}</span>
}
```

This generates a function called `Greeting` that accepts `name string` and returns a `GreetingView` containing the element tree. You call it from other templates using the `@` prefix with positional arguments:

```gsx
templ App() {
    <div class="flex-col gap-1">
        @Greeting("Alice")
        @Greeting("Bob")
    </div>
}
```

### Children Slot

Both pure and struct components can accept nested content via the `{children...}` placeholder.

In a pure component, children arrive as a function parameter:

```gsx
templ Card(title string) {
    <div class="border-rounded p-1 flex-col gap-1">
        <span class="text-gradient-cyan-magenta font-bold">{title}</span>
        <hr class="border-single" />
        {children...}
    </div>
}
```

The caller passes children by nesting elements inside the component call:

```gsx
templ Dashboard() {
    @Card("System Info") {
        <span>Version: 1.2.0</span>
        <span>Uptime: 3d 14h</span>
    }
}
```

Struct components use the same `{children...}` syntax. Add a `children []*tui.Element` field to the struct and accept it in the constructor:

```gsx
type panel struct {
    title    string
    children []*tui.Element
}

func NewPanel(title string, children []*tui.Element) *panel {
    return &panel{title: title, children: children}
}

templ (p *panel) Render() {
    <div class="border-rounded p-1 flex-col gap-1">
        <span class="font-bold">{p.title}</span>
        {children...}
    </div>
}
```

The caller syntax is the same as with pure components:

```gsx
templ (a *myApp) Render() {
    @NewPanel("Items") {
        <span>First</span>
        <span>Second</span>
    }
}
```

The generated code passes children through the constructor. On re-renders, `UpdateProps` copies the fresh children to the cached instance automatically.

### When to Use Pure Components

Use pure components for reusable visual elements that don't need their own state: cards, badges, headers, layout wrappers, styled containers. They're the go-tui equivalent of a React functional component with no hooks.

## Struct Components

Struct components carry state, handle input, and support lifecycle hooks. They're defined in three parts: a struct, a constructor, and a `templ` render method.

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

type counter struct {
    count *tui.State[int]
}

func Counter() *counter {
    return &counter{
        count: tui.NewState(0),
    }
}

func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('+', func(ke tui.KeyEvent) {
            c.count.Update(func(v int) int { return v + 1 })
        }),
        tui.OnRune('-', func(ke tui.KeyEvent) {
            c.count.Update(func(v int) int { return v - 1 })
        }),
    }
}

templ (c *counter) Render() {
    <div class="flex-col items-center justify-center h-full gap-1">
        <span class="font-bold text-cyan">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
        <span class="font-dim">Press + / - to change, Esc to quit</span>
    </div>
}
```

The `templ (c *counter) Render()` syntax declares a method on the struct that returns a `*tui.Element`. The method name must be `Render`. In the `.gsx` file you write no parameters, but the generated Go code adds `app *tui.App` automatically.

### The Component Interface

At minimum, a struct component must implement `Render(app *App) *Element`. The `templ` method syntax generates this for you. Beyond that, several optional interfaces add behavior:

| Interface | Method | Purpose |
|-----------|--------|---------|
| `KeyListener` | `KeyMap() tui.KeyMap` | Keyboard bindings |
| `MouseListener` | `HandleMouse(MouseEvent) bool` | Mouse click/scroll handling |
| `WatcherProvider` | `Watchers() []tui.Watcher` | Timers and channel listeners |
| `Initializer` | `Init() func()` | Setup; returns a cleanup function |
| `AppBinder` | `BindApp(app *App)` | Receive the App instance |
| `PropsUpdater` | `UpdateProps(fresh Component)` | Handle prop changes on re-mount |

You don't need to implement all of these. Use only what your component needs. The [Components](components) guide covers each in detail.

## Elements

Elements are HTML-like tags that map to `tui.Element` instances. There are two kinds: container elements that can hold children, and self-closing (void) elements that cannot.

### Container Elements

| Element | Description |
|---------|-------------|
| `<div>` | Block flex container (the workhorse for layout) |
| `<span>` | Inline text container |
| `<p>` | Paragraph with text wrapping |
| `<ul>` | Unordered list container |
| `<li>` | List item (renders with a bullet) |
| `<button>` | Clickable element |
| `<table>` | Table container |

### Self-Closing (Void) Elements

These must use the `/>` closing syntax and cannot have children:

| Element | Description |
|---------|-------------|
| `<input />` | Text input field |
| `<progress />` | Progress bar |
| `<hr />` | Horizontal rule |
| `<br />` | Line break |

### Nesting

Elements nest naturally:

```gsx
<div class="flex-col gap-1 p-1">
    <div class="flex justify-between">
        <span class="font-bold">Title</span>
        <span class="font-dim">Subtitle</span>
    </div>
    <hr />
    <ul class="flex-col">
        <li><span>First item</span></li>
        <li><span>Second item</span></li>
    </ul>
</div>
```

## Attributes

Attributes set element properties. There are three forms:

### String Attributes

Quoted strings for `class`, `id`, and other string-typed properties:

```gsx
<div class="flex-col gap-2" id="main-panel">
    <span class="font-bold text-cyan">Title</span>
</div>
```

### Go Expression Attributes

Curly braces for any Go expression:

```gsx
<div width={42} height={10} flexGrow={1.5}>
    <span textStyle={tui.NewStyle().Bold().Foreground(tui.ANSIColor(tui.Cyan))}>
        Styled text
    </span>
</div>
```

This works for integers, floats, booleans, function calls, struct literals — anything that produces a value of the right type.

### Boolean Attributes

A bare attribute name is shorthand for `true`:

```gsx
<button disabled>Can't click</button>
<!-- equivalent to: -->
<button disabled={true}>Can't click</button>
```

### Ref Attributes

Bind an element to a reference variable for later access (scroll control, click detection, etc.):

```gsx
<div ref={myRef} class="flex-col">
    content
</div>
```

See the [Event Handling](events) guide for how refs work with click handling.

### Attribute Reference

Here's the full set of supported attributes, grouped by purpose:

**Identity**: `id`, `class`, `disabled`, `ref`, `deps`

**Dimensions**: `width`, `widthPercent`, `height`, `heightPercent`, `minWidth`, `minHeight`, `maxWidth`, `maxHeight`

**Flex Container**: `direction`, `justify`, `align`, `gap`

**Flex Item**: `flexGrow`, `flexShrink`, `alignSelf`

**Spacing**: `padding`, `margin`

**Visual**: `border`, `borderStyle`, `background`, `text`, `textStyle`, `textAlign`

**Focus**: `focusable`, `onFocus`, `onBlur`

**Scroll**: `scrollable`, `scrollOffset`, `scrollbarStyle`, `scrollbarThumbStyle`

**Input-specific**: `value`, `placeholder`

**Progress-specific**: `value`, `max`

## Go Expressions

Curly braces embed Go expressions as text content or attribute values.

### Text Content

Any Go expression inside `{...}` is rendered as text:

```gsx
<span>{fmt.Sprintf("Count: %d", c.count.Get())}</span>
<span>{"Hello, " + name}</span>
<span>{len(items)}</span>
```

### Computed Classes

Classes can be built from Go expressions too:

```gsx
<span class={statusClass(isOnline)}>Status</span>
```

Where `statusClass` is a regular Go function:

```go
func statusClass(online bool) string {
    if online {
        return "text-green font-bold"
    }
    return "text-red font-dim"
}
```

### Method Calls

Call methods on the receiver or on state variables:

```gsx
<span textStyle={s.getHeaderStyle()}>{s.count.Get()}</span>
```

## Control Flow

Three directives control rendering logic: `@if`, `@for`, and `@let`.

### @if / @else

Conditionally render elements based on a Go boolean expression:

```gsx
@if s.count.Get() > 0 {
    <span class="text-green">Positive</span>
} @else {
    <span class="text-red">Zero or negative</span>
}
```

You can chain conditions with `@else @if`:

```gsx
@if s.count.Get() > 10 {
    <span class="text-green font-bold">High</span>
} @else @if s.count.Get() > 0 {
    <span class="text-yellow">Low</span>
} @else {
    <span class="text-red">Zero</span>
}
```

### @for

Loop over slices, maps, or any Go iterable with `range`:

```gsx
@for i, item := range items {
    <span>{fmt.Sprintf("%d. %s", i+1, item)}</span>
}
```

You can ignore the index with `_`:

```gsx
@for _, item := range items {
    <span>{item}</span>
}
```

Loops and conditionals nest freely:

```gsx
@for i, item := range items {
    @if i == s.selected.Get() {
        <span class="text-cyan font-bold">{"> " + item}</span>
    } @else {
        <span>{"  " + item}</span>
    }
}
```

### @let

Bind an element to a local variable to avoid repeating complex expressions:

```gsx
@let countText = <span class="font-bold">{fmt.Sprintf("%d", s.count.Get())}</span>
<div class="flex gap-1">
    <span>Count:</span>
    {countText}
</div>
```

This is useful when you want to compute an element once and reuse it. The variable is scoped to the rest of the component body after its declaration.

## Helper Functions

Regular Go functions in `.gsx` files work exactly as they do in `.go` files. They're useful for formatting, style computation, and shared logic:

```gsx
package main

import "fmt"

func keyLabel(name string, pressed bool) string {
    if pressed {
        return "* " + name
    }
    return "  " + name
}

func keyStyle(pressed bool) string {
    if pressed {
        return "text-green font-bold"
    }
    return "font-dim"
}

templ KeyIndicator(name string, pressed bool) {
    <span class={keyStyle(pressed)}>{keyLabel(name, pressed)}</span>
}
```

The distinction between a helper function and a component is the `templ` keyword. A `templ` declaration produces an element tree. A `func` declaration is plain Go.

## Calling Components

Components are called with the `@` prefix. Parameters are passed as positional arguments matching the component's parameter list.

### Pure Components

```gsx
templ Badge(label string) {
    <span class="bg-cyan text-black px-1 font-bold">{label}</span>
}

templ StatusLine(label string, value string) {
    <div class="flex gap-1">
        <span class="font-dim">{label}</span>
        <span>{value}</span>
    </div>
}

// Usage:
templ Header() {
    <div class="flex-col gap-1">
        @Badge("v1.0")
        @StatusLine("Status:", "Running")
    </div>
}
```

### Struct Components

Struct components are instantiated through their constructor and passed to the template with `@`:

```gsx
templ (a *app) Render() {
    <div class="flex h-full">
        @Sidebar(a.category)
        @Content(a.category, a.query)
    </div>
}
```

Where `Sidebar` and `Content` are constructors that return struct component instances.

## Code Generation

After writing or editing `.gsx` files, run the code generator to produce the corresponding Go files:

```bash
tui generate ./...
```

This processes all `.gsx` files recursively and writes `_gsx.go` files alongside them. For example, `hello.gsx` produces `hello_gsx.go`. Hyphens in filenames become underscores (`my-app.gsx` becomes `my_app_gsx.go`).

The generated files should not be edited by hand. They're overwritten on every run of `tui generate`.

### Related Commands

| Command | Purpose |
|---------|---------|
| `tui generate [path...]` | Generate Go code from `.gsx` files |
| `tui check [path...]` | Validate `.gsx` files without writing output |
| `tui fmt [path...]` | Format `.gsx` files (like `gofmt` for `.gsx`) |
| `tui fmt --check [path...]` | Check formatting without modifying files |
| `tui fmt --stdout [path...]` | Write formatted output to stdout |

Path arguments accept specific files (`hello.gsx`), directories (`./examples/`), or recursive patterns (`./...`).

See the CLI section above for the full command reference.

## Putting It All Together

Here's a complete `.gsx` file that uses most of the syntax covered above — pure components, a struct component with state, control flow, helper functions, and children slots:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

// Helper function: formats an item label
func itemLabel(index int, name string, selected bool) string {
    prefix := "  "
    if selected {
        prefix = "> "
    }
    return fmt.Sprintf("%s%d. %s", prefix, index+1, name)
}

// Helper function: returns a style class based on selection
func itemClass(selected bool) string {
    if selected {
        return "text-cyan font-bold"
    }
    return ""
}

// Pure component with children slot
templ Panel(title string) {
    <div class="border-rounded p-1 flex-col gap-1" width={32}>
        <span class="font-bold text-gradient-cyan-magenta">{title}</span>
        <hr />
        {children...}
    </div>
}

// Struct component with state and key handling
type listApp struct {
    items    []string
    selected *tui.State[int]
}

func ListApp() *listApp {
    return &listApp{
        items:    []string{"Alpha", "Bravo", "Charlie", "Delta"},
        selected: tui.NewState(0),
    }
}

func (l *listApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) {
            l.selected.Update(func(v int) int {
                if v > 0 {
                    return v - 1
                }
                return v
            })
        }),
        tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) {
            l.selected.Update(func(v int) int {
                if v < len(l.items)-1 {
                    return v + 1
                }
                return v
            })
        }),
    }
}

templ (l *listApp) Render() {
    <div class="flex-col items-center justify-center h-full">
        @Panel("Select an Item") {
            @for i, item := range l.items {
                <span class={itemClass(i == l.selected.Get())}>
                    {itemLabel(i, item, i == l.selected.Get())}
                </span>
            }

            <br />

            @if l.selected.Get() >= 0 {
                <span class="font-dim">{fmt.Sprintf("Selected: %s", l.items[l.selected.Get()])}</span>
            }
        }
    </div>
}
```

With a corresponding `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(ListApp()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

## Next Steps

- [Styling and Colors](styling) — Text styles, colors, borders, and gradients
- [Layout](layout) — Flexbox layout: direction, alignment, spacing, and sizing
- [Components](components) — Component patterns, composition, and lifecycle interfaces
