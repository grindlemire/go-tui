# go-tui — Complete Reference for LLMs

> Declarative terminal UI framework for Go with templ-like syntax and flexbox layout.
> Pure Go (no CGO), minimal dependencies, generates type-safe code from `.gsx` templates.

## Installation

```bash
go get github.com/grindlemire/go-tui
go install github.com/grindlemire/go-tui/cmd/tui@latest
```

## CLI Commands

```bash
tui generate [path...]       # Generate Go code from .gsx files
tui check [path...]          # Validate .gsx files without writing output
tui fmt [path...]            # Format .gsx files
tui fmt --check [path...]    # Check formatting without modifying
tui lsp                      # Start language server (stdio)
```

## Architecture

```
.gsx files → tui generate → _gsx.go files → go build → binary
```

At runtime: Event Loop → Layout Engine (flexbox) → Double-buffered Render → ANSI Terminal Output

## GSX Syntax

`.gsx` files are Go files with a `templ` keyword for declaring UI components. They compile to standard Go code.

### Pure Components (stateless)

```gsx
package main

import tui "github.com/grindlemire/go-tui"

templ Greeting(name string) {
    <span class="text-cyan font-bold">{"Hello, " + name}</span>
}

// With children slot
templ Card(title string) {
    <div class="border-rounded p-1 flex-col gap-1">
        <span class="font-bold">{title}</span>
        {children...}
    </div>
}

// Usage
templ App() {
    @Card("Info") {
        @Greeting("Alice")
    }
}
```

### Struct Components (stateful)

```gsx
type counter struct {
    count *tui.State[int]
}

func Counter() *counter {
    return &counter{count: tui.NewState(0)}
}

func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('+', func(ke tui.KeyEvent) {
            c.count.Update(func(v int) int { return v + 1 })
        }),
    }
}

templ (c *counter) Render() {
    <div class="flex-col items-center justify-center h-full">
        <span class="text-cyan font-bold">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
    </div>
}
```

### Struct Components with Children

```gsx
type panel struct {
    title    string
    children []*tui.Element
}

func NewPanel(title string, children []*tui.Element) *panel {
    return &panel{title: title, children: children}
}

templ (p *panel) Render() {
    <div class="border-rounded p-1 flex-col">
        <span class="font-bold">{p.title}</span>
        {children...}
    </div>
}
```

### Control Flow

```gsx
// Conditionals
@if condition {
    <span>True</span>
} @else @if otherCondition {
    <span>Other</span>
} @else {
    <span>Default</span>
}

// Loops
@for i, item := range items {
    <span>{fmt.Sprintf("%d: %s", i, item)}</span>
}

// Local element binding
@let badge = <span class="font-bold">{label}</span>
<div>{badge}</div>
```

### Go Expressions

```gsx
<span>{fmt.Sprintf("Count: %d", c.count.Get())}</span>
<span class={computedClass(isActive)}>Dynamic class</span>
<div width={42} height={10} flexGrow={1.5}>content</div>
```

## main.go Pattern

```go
package main

import (
    "fmt"
    "os"
    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(MyComponent()),
        // Optional:
        // tui.WithMouse(),
        // tui.WithFrameRate(60),
        // tui.WithInlineHeight(10),
        // tui.WithGlobalKeyHandler(func(ke tui.KeyEvent) bool { return false }),
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

## Single-Frame Printing

For CLI tools that render styled output once and exit, without starting an interactive app:

```go
// Print to stdout (auto-detects terminal width, falls back to 80)
tui.Print(MyComponent("hello"))

// Return as ANSI string (no trailing newline)
s := tui.Sprint(view, tui.WithPrintWidth(80))

// Write to any io.Writer (appends trailing newline)
tui.Fprint(w, view, tui.WithPrintWidth(120))
```

All three accept a `Viewable` (generated `.gsx` views and raw `*Element` values). Same components work with both `Print` and `App.Run()`.

```go
// PrintOption
tui.WithPrintWidth(w int)  // Explicit width; default: auto-detect, fallback 80
```

## Built-in Elements

### Container Elements

| Element | Description | Default Direction |
|---------|-------------|-------------------|
| `<div>` | Block flex container (primary layout element) | Row |
| `<span>` | Inline text container | — |
| `<p>` | Paragraph with automatic text wrapping | — |
| `<ul>` | Unordered list container | Column |
| `<li>` | List item (auto bullet prefix) | — |
| `<button>` | Clickable element | — |
| `<table>` | Table container | Column |

### Self-Closing (Void) Elements

| Element | Description |
|---------|-------------|
| `<input />` | Text input. Attrs: `value`, `placeholder` |
| `<progress />` | Progress bar. Attrs: `value`, `max` |
| `<hr />` | Horizontal rule |
| `<br />` | Line break |

## Attributes

### Layout

| Attribute | Type | Description |
|-----------|------|-------------|
| `width` | `int` | Fixed width in characters |
| `widthPercent` | `int` | Width as percentage |
| `height` | `int` | Fixed height in rows |
| `heightPercent` | `int` | Height as percentage |
| `minWidth` / `maxWidth` | `int` | Min/max width constraints |
| `minHeight` / `maxHeight` | `int` | Min/max height constraints |
| `direction` | `tui.Direction` | `tui.Row` or `tui.Column` |
| `justify` | `tui.Justify` | Main axis alignment |
| `align` | `tui.Align` | Cross axis alignment |
| `alignSelf` | `tui.Align` | Override parent's align |
| `gap` | `int` | Gap between children |
| `flexGrow` | `float64` | Flex grow factor |
| `flexShrink` | `float64` | Flex shrink factor |
| `padding` | `int` | Padding on all sides |
| `margin` | `int` | Margin on all sides |

### Visual

| Attribute | Type | Description |
|-----------|------|-------------|
| `border` | `tui.BorderStyle` | Border style |
| `background` | `tui.Color` | Background color |
| `text` | `string` | Text content |
| `textStyle` | `tui.Style` | Text styling |
| `textAlign` | `string` | `"left"`, `"center"`, `"right"` |
| `borderStyle` | `tui.Style` | Border line styling |

### Identity & Behavior

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | `string` | Unique identifier |
| `class` | `string` | Tailwind-style classes |
| `disabled` | `bool` | Disable interaction |
| `ref` | `*tui.Ref` | Bind to a reference |
| `deps` | expression | Explicit state dependencies |
| `focusable` | `bool` | Enable focus |
| `onFocus` / `onBlur` | `func(*tui.Element)` | Focus callbacks |

### Scroll

| Attribute | Type | Description |
|-----------|------|-------------|
| `scrollable` | `tui.ScrollMode` | `ScrollVertical`, `ScrollHorizontal`, `ScrollBoth` |
| `scrollOffset` | `int, int` | x, y scroll position |
| `scrollbarStyle` | `tui.Style` | Track style |
| `scrollbarThumbStyle` | `tui.Style` | Thumb style |

## Tailwind Classes (Complete Reference)

### Layout Direction

`flex` `flex-row` `flex-col`

### Flex Sizing

`grow` `grow-0` `shrink` `shrink-0` `flex-1` `flex-auto` `flex-initial` `flex-none` `flex-grow-N` `flex-shrink-N`

### Width & Height

`w-N` `h-N` `w-full` `h-full` `w-auto` `h-auto` `w-1/2` `w-1/3` `w-2/3` `h-1/2` `h-1/3` `h-2/3` `min-w-N` `max-w-N` `min-h-N` `max-h-N`

### Justify Content

`justify-start` `justify-center` `justify-end` `justify-between` `justify-around` `justify-evenly`

### Align Items / Self

`items-start` `items-center` `items-end` `items-stretch` `self-start` `self-center` `self-end` `self-stretch`

### Spacing

`gap-N` `p-N` `px-N` `py-N` `pt-N` `pr-N` `pb-N` `pl-N` `m-N` `mx-N` `my-N` `mt-N` `mr-N` `mb-N` `ml-N`

### Borders

`border` `border-single` `border-double` `border-rounded` `border-thick` `border-COLOR` `border-[#hex]` `border-gradient-C1-C2[-dir]`

### Text Styling

`font-bold` `font-dim` `text-dim` `italic` `underline` `strikethrough` `blink` `reverse` `text-left` `text-center` `text-right` `truncate` `wrap` `nowrap`

### Colors

Named: `text-COLOR` `text-bright-COLOR` `bg-COLOR` `bg-bright-COLOR`
Hex: `text-[#hex]` `bg-[#hex]`
Colors: `red` `green` `blue` `cyan` `magenta` `yellow` `white` `black` (plus `bright-` variants)

### Gradients

`text-gradient-C1-C2[-dir]` `bg-gradient-C1-C2[-dir]` `border-gradient-C1-C2[-dir]`
Directions: `-h` (horizontal, default), `-v` (vertical), `-dd` (diagonal down), `-du` (diagonal up)

### Scroll & Overflow

`overflow-scroll` `overflow-y-scroll` `overflow-x-scroll` `overflow-hidden` `scrollbar-COLOR` `scrollbar-thumb-COLOR` `scrollbar-[#hex]` `scrollbar-thumb-[#hex]`

### Other

`focusable` `hidden`

## Key Types

```go
// Dimensions
tui.Fixed(10)        // 10 characters
tui.Percent(50)      // 50% of parent
tui.Auto()           // Size to content

// Borders
tui.BorderNone / tui.BorderSingle / tui.BorderDouble / tui.BorderRounded / tui.BorderThick

// Direction
tui.Row / tui.Column

// Justify
tui.JustifyStart / tui.JustifyCenter / tui.JustifyEnd
tui.JustifySpaceBetween / tui.JustifySpaceAround / tui.JustifySpaceEvenly

// Align
tui.AlignStart / tui.AlignCenter / tui.AlignEnd / tui.AlignStretch

// TextAlign
tui.TextAlignLeft / tui.TextAlignCenter / tui.TextAlignRight

// ScrollMode
tui.ScrollNone / tui.ScrollVertical / tui.ScrollHorizontal / tui.ScrollBoth

// OverflowMode
tui.OverflowVisible / tui.OverflowHidden

// Colors
tui.Black / tui.Red / tui.Green / tui.Yellow / tui.Blue / tui.Magenta / tui.Cyan / tui.White
tui.BrightBlack / tui.BrightRed / tui.BrightGreen / ... / tui.BrightWhite
tui.ANSIColor(index)        // 256 palette
tui.RGBColor(r, g, b)       // 24-bit true color
tui.HexColor("#ff6600")     // Hex string
tui.DefaultColor()           // Terminal default

// Gradients
tui.NewGradient(tui.Red, tui.Blue)                                    // Horizontal
tui.NewGradient(tui.Red, tui.Blue).WithDirection(tui.GradientVertical) // Vertical
// Directions: GradientHorizontal, GradientVertical, GradientDiagonalDown, GradientDiagonalUp

// Style (chainable)
tui.NewStyle().Bold().Dim().Italic().Underline().Strikethrough().Reverse().Blink()
    .Foreground(tui.Cyan).Background(tui.Black)

// State (generic reactive container)
count := tui.NewState(0)
count.Get()                           // Read (safe from any goroutine)
count.Set(5)                          // Write (main loop only)
count.Update(func(v int) int { ... }) // Read-modify-write (main loop only)
unbind := count.Bind(func(v int) { }) // Called on change

// Refs (element references for hit testing)
ref := tui.NewRef()           // Single element
list := tui.NewRefList()      // Loop elements: list.At(i) to bind, list.El(i) to read
m := tui.NewRefMap[string]()  // Keyed elements: m.At(key) to bind, m.El(key) to read

// Events (cross-component pub/sub)
bus := tui.NewEvents[MyEvent]("topic")
bus.Emit(MyEvent{...})
unsub := bus.Subscribe(func(e MyEvent) { ... })
```

## Component Interfaces

All optional except `Component.Render`:

```go
// Required
type Component interface {
    Render(app *App) *Element
}

// Keyboard input
type KeyListener interface {
    KeyMap() KeyMap
}

// Mouse input
type MouseListener interface {
    HandleMouse(MouseEvent) bool
}

// Background timers/channels
type WatcherProvider interface {
    Watchers() []Watcher
}

// Setup/cleanup on mount/unmount
type Initializer interface {
    Init() func()  // returned func is cleanup
}

// Receive updated props on re-render
type PropsUpdater interface {
    UpdateProps(fresh Component)
}

// Auto-generated: binds State/Events fields to app
type AppBinder interface {
    BindApp(app *App)
}
```

## Key Bindings

```go
func (c *myComp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, handler),    // Special key, propagates
        tui.OnKeyStop(tui.KeyEnter, handler), // Special key, stops propagation
        tui.OnRune('q', handler),             // Specific character
        tui.OnRuneStop('/', handler),          // Specific char, stops propagation
        tui.OnRunes(handler),                  // Any printable character
        tui.OnRunesStop(handler),              // Any printable, stops propagation
    }
}
```

**KeyEvent fields:** `ke.Key`, `ke.Rune`, `ke.Mod` (ModCtrl, ModAlt, ModShift), `ke.App()`

**Special keys:** `KeyUp` `KeyDown` `KeyLeft` `KeyRight` `KeyEnter` `KeyTab` `KeyEscape` `KeyBackspace` `KeyDelete` `KeyHome` `KeyEnd` `KeyPageUp` `KeyPageDown` `KeyInsert` `KeyCtrlA`–`KeyCtrlZ` `KeyF1`–`KeyF12`

## Mouse Handling

Requires `tui.WithMouse()` in app options.

```go
func (c *myComp) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.saveBtn, c.onSave),       // Single ref
        tui.Click(c.itemRefs, c.onItemClick),  // RefList
    )
}
```

**MouseEvent fields:** `me.Button` (MouseLeft, MouseRight, MouseWheelUp, MouseWheelDown), `me.Action` (MousePress, MouseRelease, MouseDrag), `me.X`, `me.Y`, `me.App()`

## Watchers

```go
func (c *myComp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, c.tick),           // Timer
        tui.Watch(c.dataCh, c.onData),              // Channel watcher
    }
}
```

Callbacks run on the main event loop, so you can mutate state directly.

## Scroll API

```go
el.ScrollTo(x, y)          // Absolute position (clamped)
el.ScrollBy(dx, dy)        // Relative adjustment (clamped)
el.ScrollToTop()
el.ScrollToBottom()         // With deferred re-scroll after layout
el.ScrollIntoView(child)   // Minimal scroll to make child visible
el.ScrollOffset() (x, y)   // Current position
el.MaxScroll() (maxX, maxY)
el.ContentSize() (w, h)
el.ViewportSize() (w, h)
el.IsAtBottom() bool
```

## Focus Management

```go
// Element-level
element.Focus() / element.Blur() / element.IsFocused()

// App-level cycling
app.FocusNext() / app.FocusPrev() / app.Focused()

// FocusGroup for section-level switching
sidebar := tui.NewState(true)
content := tui.NewState(false)
fg := tui.MustNewFocusGroup(sidebar, content)
fg.Next() / fg.Prev() / fg.Current()
// fg.KeyMap() returns Tab/Shift+Tab bindings
```

## State Batching

```go
ke.App().Batch(func() {
    state1.Set(val1)
    state2.Set(val2)
    // Single re-render after batch completes
})
```

## Thread Safety

| Operation | Main Loop | Any Goroutine |
|-----------|-----------|---------------|
| `state.Get()` | Yes | Yes |
| `state.Set()` / `state.Update()` | Yes | Use `app.QueueUpdate()` |
| `events.Emit()` | Yes | Use `app.QueueUpdate()` |
| `app.QueueUpdate(func() { ... })` | Yes | Yes |

## App Options

```go
tui.WithRootComponent(comp)              // Required: root component
tui.WithMouse()                          // Enable mouse support
tui.WithFrameRate(fps int)               // Default: 60
tui.WithInlineHeight(rows int)           // Inline mode (not fullscreen)
tui.WithGlobalKeyHandler(func(KeyEvent) bool)  // Global key intercept
tui.WithInputLatency(d time.Duration)    // Coalesce rapid input
tui.WithEventQueueSize(n int)            // Event queue buffer size
```

## App Methods

```go
app.Run() error              // Start event loop (blocks)
app.Stop()                   // Stop the event loop
app.Close()                  // Restore terminal
app.StopCh() <-chan struct{} // Closed when app stops
app.Batch(func())            // Batch state updates
app.QueueUpdate(func())      // Schedule work on main loop
app.FocusNext() / app.FocusPrev() / app.Focused()
app.PrintAbove(string)       // Print above inline widget
app.StreamAbove() *StreamWriter  // Stream text char-by-char above inline widget
app.Draw()                   // Manual redraw
app.QueueUpdateDraw(func())  // QueueUpdate + Draw
```

## Inline Mode

For CLI tools that render a widget inline (not fullscreen):

```go
app, _ := tui.NewApp(
    tui.WithRootComponent(myComp),
    tui.WithInlineHeight(5),  // Reserve 5 terminal rows
)
```

### StreamAbove

`app.StreamAbove()` returns a `*StreamWriter` for character-by-character streaming above the inline widget. The writer implements `io.WriteCloser` and adds `WriteStyled` and `WriteGradient` methods. Goroutine-safe. Close it when done to finalize the partial line.

```go
go func() {
    w := app.StreamAbove()
    // Plain write (backward compatible)
    fmt.Fprint(w, "hello ")
    // Styled write
    w.WriteStyled("bold text", tui.NewStyle().Bold().Foreground(tui.Red))
    // Gradient write (per-character color interpolation)
    grad := tui.NewGradient(tui.Cyan, tui.Magenta)
    w.WriteGradient("gradient text", grad)
    // Gradient with base style (bold + gradient foreground)
    w.WriteGradient("bold gradient", grad, tui.NewStyle().Bold())
    w.Close()
}()
```

`WriteStyled(text, style)` wraps text in ANSI style prefix/reset. `WriteGradient(text, gradient, base...)` writes each character with an interpolated gradient foreground; optional base style adds attributes (bold, italic, etc.) and background.

Returns a no-op writer when not in inline mode. Only one stream writer is active at a time; calling `StreamAbove()` again finalizes the previous one. `PrintAbove`/`PrintAboveln` also finalize any active stream before printing.

### PrintAboveElement

`app.PrintAboveElement(el)` renders an element tree at the terminal width and inserts the resulting rows into the inline scrollback as static ANSI text. Works with templ function output. No-op outside inline mode. Must be called from the main event loop. `QueuePrintAboveElement(el)` is the goroutine-safe variant.

`StreamWriter.WriteElement(el)` inserts a rendered element mid-stream. Finalizes the current partial line first.

```go
// Standalone
app.PrintAboveElement(DataTable(rows))

// Mid-stream
w := app.StreamAbove()
w.Write([]byte("Here's the data:\n"))
w.WriteElement(DataTable(rows))
w.Write([]byte("Done.\n"))
w.Close()
```

## Common Layout Patterns

```gsx
// Full-screen with header/content/footer
<div class="flex-col h-full">
    <div class="border-single p-1"><span>Header</span></div>
    <div class="flex-col grow p-1"><span>Content</span></div>
    <div class="border-single p-1"><span>Footer</span></div>
</div>

// Sidebar + main
<div class="flex h-full">
    <div class="w-20 border-single flex-col p-1"><span>Sidebar</span></div>
    <div class="grow flex-col p-1"><span>Main</span></div>
</div>

// Centered card
<div class="flex items-center justify-center h-full">
    <div class="border-rounded p-2 w-40"><span>Centered</span></div>
</div>

// Dashboard grid
<div class="flex-col h-full gap-1">
    <div class="flex gap-1 grow">
        <div class="grow border-rounded p-1"><span>Panel 1</span></div>
        <div class="grow border-rounded p-1"><span>Panel 2</span></div>
    </div>
    <div class="flex gap-1 grow">
        <div class="w-2/3 border-rounded p-1"><span>Wide</span></div>
        <div class="grow border-rounded p-1"><span>Narrow</span></div>
    </div>
</div>
```

## Testing

```go
// MockTerminal captures rendering output
term := tui.NewMockTerminal(80, 24)

// MockEventReader provides scripted events
reader := tui.NewMockEventReader(
    tui.KeyEvent{Key: tui.KeyRune, Rune: 'a'},
    tui.KeyEvent{Key: tui.KeyEscape},
)

// Table-driven test pattern
type tc struct { /* fields */ }
tests := map[string]tc{ "name": { /* values */ } }
for name, tt := range tests {
    t.Run(name, func(t *testing.T) { /* test */ })
}
```

## Element Option Functions (Go API)

When building elements programmatically (outside .gsx):

```go
el := tui.New(
    tui.WithWidth(30),
    tui.WithHeight(10),
    tui.WithWidthPercent(50),
    tui.WithWidthAuto(),
    tui.WithMinWidth(10),
    tui.WithMaxWidth(60),
    tui.WithDirection(tui.Column),
    tui.WithJustify(tui.JustifyCenter),
    tui.WithAlign(tui.AlignCenter),
    tui.WithGap(2),
    tui.WithFlexGrow(1.0),
    tui.WithFlexShrink(0),
    tui.WithPadding(2),
    tui.WithMargin(1),
    tui.WithBorder(tui.BorderRounded),
    tui.WithBackground(tui.Cyan),
    tui.WithText("content"),
    tui.WithTextStyle(tui.NewStyle().Bold()),
    tui.WithTextAlign(tui.TextAlignCenter),
    tui.WithScrollable(tui.ScrollVertical),
    tui.WithFocusable(true),
    tui.WithHidden(true),
    tui.WithTruncate(true),
    tui.WithTextGradient(tui.NewGradient(tui.Cyan, tui.Magenta)),
    tui.WithBackgroundGradient(tui.NewGradient(tui.Blue, tui.Cyan)),
    tui.WithBorderGradient(tui.NewGradient(tui.Cyan, tui.Magenta)),
)
```

## Cross-Component Communication

```gsx
// Producer component
type producer struct {
    bus *tui.Events[string]
}
func Producer() *producer {
    return &producer{bus: tui.NewEvents[string]("my.topic")}
}
func (p *producer) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('e', func(ke tui.KeyEvent) { p.bus.Emit("event!") }),
    }
}

// Consumer component (separate, no direct reference to producer)
type consumer struct {
    bus      *tui.Events[string]
    messages *tui.State[[]string]
}
func Consumer() *consumer {
    c := &consumer{
        bus:      tui.NewEvents[string]("my.topic"), // same topic string
        messages: tui.NewState([]string{}),
    }
    c.bus.Subscribe(func(msg string) {
        c.messages.Set(append(c.messages.Get(), msg))
    })
    return c
}
```

## Buffer / Rendering Internals

- Double-buffered character grid (front/back buffers)
- Diff-based rendering: only changed cells are written to terminal
- Each cell stores: Rune, Style, Width (CJK support)
- ANSI escape sequences for cursor movement, colors, text attributes
- Automatic color capability detection and fallback
