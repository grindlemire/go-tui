# go-tui Design Document

A declarative terminal UI framework for Go with templ-like syntax and flexbox layout.

## Goals

1. **Declarative component syntax** — Define UIs in `.tui` files that compile to type-safe Go code
2. **Flexbox layout engine** — Pure Go implementation, no CGO dependencies
3. **Minimal external dependencies** — Only depend on standard library where possible
4. **Composable widgets** — Build complex UIs from simple, reusable components
5. **Clean separation of concerns** — Layout, rendering, and state management are distinct layers

## Non-Goals

- Full CSS compatibility (only flexbox subset)
- Mouse support in v1 (keyboard-first)
- Animation framework (can be added later)
- Cross-platform terminal abstraction (raw ANSI initially, can wrap later)

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│  .tui files (declarative syntax)                            │
│  @component Header(title string) {                          │
│    <box border="single" padding="1">                        │
│      <text bold>{title}</text>                              │
│    </box>                                                   │
│  }                                                          │
└─────────────────────┬───────────────────────────────────────┘
                      │ tui generate (code gen)
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Generated Go code                                          │
│  func Header(title string) tui.Widget { ... }               │
└─────────────────────┬───────────────────────────────────────┘
                      │ Build widget tree
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Widget Tree + Layout Engine                                │
│  Nodes with style properties → Computed rects               │
└─────────────────────┬───────────────────────────────────────┘
                      │ Render to buffer
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Character Buffer (2D grid of styled cells)                 │
└─────────────────────┬───────────────────────────────────────┘
                      │ Flush to terminal
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Terminal (ANSI escape sequences)                           │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Terminal Rendering Foundation

**Goal:** Render styled text and shapes at arbitrary positions.

### 1.1 Core Types

```go
// pkg/tui/style.go
type Color uint8 // ANSI 256 colors, 0 = default

type Style struct {
    Fg        Color
    Bg        Color
    Bold      bool
    Italic    bool
    Underline bool
    Dim       bool
    Reverse   bool
}

// pkg/tui/cell.go
type Cell struct {
    Rune  rune
    Style Style
}

// pkg/tui/rect.go
type Rect struct {
    X, Y          int
    Width, Height int
}

func (r Rect) Inner(padding int) Rect
func (r Rect) Contains(x, y int) bool
```

### 1.2 Buffer

```go
// pkg/tui/buffer.go
type Buffer struct {
    cells         [][]Cell
    width, height int
}

func NewBuffer(width, height int) *Buffer
func (b *Buffer) SetCell(x, y int, r rune, style Style)
func (b *Buffer) SetString(x, y int, s string, style Style)
func (b *Buffer) Fill(rect Rect, r rune, style Style)
func (b *Buffer) Clear()
func (b *Buffer) Resize(width, height int)
```

### 1.3 Terminal Abstraction

```go
// pkg/tui/terminal.go
type Terminal interface {
    Size() (width, height int)
    Flush(buf *Buffer)
    Clear()
    SetCursor(x, y int)
    HideCursor()
    ShowCursor()
    EnableRawMode() error
    DisableRawMode() error
}

// pkg/tui/terminal_ansi.go
type ANSITerminal struct {
    out    io.Writer
    in     io.Reader
    oldState *term.State // from golang.org/x/term if needed, or raw syscall
}
```

### 1.4 Deliverable

- Draw colored text and box-drawing characters at any position
- Test: render a bordered box with title to terminal

---

## Phase 2: Layout Engine

**Goal:** Pure Go flexbox implementation that computes positions from style properties.

### 2.1 Layout Types

```go
// pkg/layout/value.go
type Unit int

const (
    UnitAuto Unit = iota
    UnitFixed
    UnitPercent
)

type Value struct {
    Amount float64
    Unit   Unit
}

func Fixed(n int) Value
func Percent(n float64) Value
func Auto() Value

// pkg/layout/style.go
type Direction int

const (
    Row Direction = iota
    Column
)

type Justify int

const (
    JustifyStart Justify = iota
    JustifyCenter
    JustifyEnd
    JustifySpaceBetween
    JustifySpaceAround
)

type Align int

const (
    AlignStart Align = iota
    AlignCenter
    AlignEnd
    AlignStretch
)

type Style struct {
    Width         Value
    Height        Value
    MinWidth      Value
    MinHeight     Value
    MaxWidth      Value
    MaxHeight     Value

    FlexDirection Direction
    FlexGrow      float64
    FlexShrink    float64

    JustifyContent Justify
    AlignItems     Align

    Padding Edges
    Margin  Edges
    Gap     int
}

type Edges struct {
    Top, Right, Bottom, Left int
}
```

### 2.2 Layout Node

```go
// pkg/layout/node.go
type Node struct {
    Style    Style
    Children []*Node

    // Computed after layout
    Layout Rect
}

func (n *Node) AddChild(child *Node)
func (n *Node) Calculate(availableWidth, availableHeight int)
```

### 2.3 Layout Algorithm

Implement in stages:

1. **Column layout with fixed heights and FlexGrow**
2. **Row layout** (same algorithm, swapped axes)
3. **JustifyContent** (distribute space along main axis)
4. **AlignItems** (position on cross axis)
5. **Percentage values**
6. **Gap between children**
7. **Padding and margin**
8. **Min/max constraints**

```go
// pkg/layout/calculate.go
func calculate(node *Node, available Rect) {
    // 1. Resolve fixed sizes
    // 2. Calculate flex basis for each child
    // 3. Distribute remaining space according to FlexGrow
    // 4. Handle shrinking if overflow
    // 5. Apply justify/align
    // 6. Recurse into children
}
```

### 2.4 Deliverable

- Layout engine passes tests for all flexbox features
- Test: nested containers with mixed sizing lay out correctly

---

## Phase 3: Widget System

**Goal:** Composable widgets that know how to render themselves.

### 3.1 Widget Interface

```go
// pkg/tui/widget.go
type Widget interface {
    // Build returns the layout node for this widget
    // Called during tree construction
    Build() *layout.Node

    // Render draws the widget content into the buffer
    // Called after layout is computed
    Render(buf *Buffer, rect Rect)
}
```

### 3.2 Core Widgets

```go
// pkg/tui/widgets/box.go
type Box struct {
    Style    layout.Style
    Border   BorderStyle
    Title    string
    Children []Widget
}

func (b *Box) Build() *layout.Node
func (b *Box) Render(buf *Buffer, rect Rect)

// pkg/tui/widgets/text.go
type Text struct {
    Content string
    Style   Style
}

func (t *Text) Build() *layout.Node
func (t *Text) Render(buf *Buffer, rect Rect)

// pkg/tui/widgets/list.go
type List struct {
    Items    []string
    Selected int
    Style    layout.Style
}
```

### 3.3 Border Styles

```go
// pkg/tui/border.go
type BorderStyle int

const (
    BorderNone BorderStyle = iota
    BorderSingle  // ┌─┐│└─┘
    BorderDouble  // ╔═╗║╚═╝
    BorderRounded // ╭─╮│╰─╯
    BorderThick   // ┏━┓┃┗━┛
)

type BorderChars struct {
    TopLeft, Top, TopRight    rune
    Left, Right               rune
    BottomLeft, Bottom, BottomRight rune
}
```

### 3.4 Tree Rendering

```go
// pkg/tui/render.go
func RenderTree(root Widget, buf *Buffer, available Rect) {
    // 1. Build layout tree from widget tree
    layoutRoot := buildLayoutTree(root)

    // 2. Calculate layout
    layoutRoot.Calculate(available.Width, available.Height)

    // 3. Render each widget at its computed rect
    renderWidget(root, layoutRoot, buf)
}
```

### 3.5 Deliverable

- Build UIs programmatically in Go with flexbox layout
- Test: dashboard with header, sidebar, main content area

---

## Phase 4: Event System

**Goal:** Handle keyboard input and manage focus.

### 4.1 Event Types

```go
// pkg/tui/event.go
type Event interface {
    isEvent()
}

type KeyEvent struct {
    Key  Key
    Rune rune
    Mod  Modifier
}

type ResizeEvent struct {
    Width, Height int
}

type Key int

const (
    KeyNone Key = iota
    KeyEscape
    KeyEnter
    KeyTab
    KeyBackspace
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    // ... etc
)

type Modifier int

const (
    ModNone Modifier = 0
    ModCtrl Modifier = 1 << iota
    ModAlt
    ModShift
)
```

### 4.2 Event Reading

```go
// pkg/tui/input.go
type EventReader interface {
    Read() (Event, error)
}

type stdinReader struct {
    buf []byte
}

func (r *stdinReader) Read() (Event, error) {
    // Parse ANSI escape sequences into events
}
```

### 4.3 Focusable Widgets

```go
// pkg/tui/focus.go
type Focusable interface {
    Widget
    IsFocusable() bool
    HandleEvent(event Event) (handled bool)
}

type FocusManager struct {
    root     Widget
    focused  Focusable
    focusOrder []Focusable
}

func (f *FocusManager) Next()
func (f *FocusManager) Prev()
func (f *FocusManager) Dispatch(event Event) bool
```

### 4.4 Application Loop

```go
// pkg/tui/app.go
type App struct {
    terminal Terminal
    root     Widget
    focus    *FocusManager
    events   EventReader
}

func (a *App) Run() error {
    for {
        // 1. Read event
        event, err := a.events.Read()

        // 2. Handle resize
        if resize, ok := event.(ResizeEvent); ok {
            a.terminal.Resize(resize.Width, resize.Height)
        }

        // 3. Dispatch to focused widget
        a.focus.Dispatch(event)

        // 4. Re-render
        buf := NewBuffer(...)
        RenderTree(a.root, buf, ...)
        a.terminal.Flush(buf)
    }
}
```

### 4.5 Deliverable

- Navigate between focusable widgets with Tab/Shift+Tab
- Widgets respond to keyboard input
- Test: form with multiple input fields

---

## Phase 5: Syntax and Code Generation

**Goal:** Define UIs in `.tui` files that compile to Go.

### 5.1 Syntax Design

```
// components/header.tui

@component Header(title string) {
    <box border="single" padding="1" flexDirection="row" justifyContent="space-between">
        <text bold>{title}</text>
        <text dim>v1.0.0</text>
    </box>
}

@component App() {
    <box flexDirection="column" height="100%">
        <Header title="My Application" />
        <box flexGrow="1">
            {children}
        </box>
        <StatusBar />
    </box>
}
```

### 5.2 Lexer

```go
// pkg/tuigen/lexer.go
type TokenType int

const (
    TokenEOF TokenType = iota
    TokenComponent    // @component
    TokenIdent        // identifier
    TokenString       // "..."
    TokenNumber       // 123
    TokenLBrace       // {
    TokenRBrace       // }
    TokenLAngle       // <
    TokenRAngle       // >
    TokenSlash        // /
    TokenEquals       // =
    TokenGoCode       // Go expression inside {}
)

type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Column  int
}

type Lexer struct {
    input   string
    pos     int
    line    int
    column  int
}

func (l *Lexer) Next() Token
```

### 5.3 Parser

```go
// pkg/tuigen/parser.go
type ComponentDef struct {
    Name   string
    Params []Param
    Body   Element
}

type Param struct {
    Name string
    Type string
}

type Element struct {
    Tag        string           // "box", "text", or component name
    Attributes map[string]Value // static or Go expression
    Children   []Node           // Element or GoExpr
}

type GoExpr struct {
    Code string
}

type Parser struct {
    lexer  *Lexer
    current Token
}

func (p *Parser) ParseFile() ([]ComponentDef, error)
```

### 5.4 Code Generator

```go
// pkg/tuigen/generate.go
type Generator struct {
    pkg string
}

func (g *Generator) Generate(components []ComponentDef) ([]byte, error) {
    // Output:
    // func Header(title string) tui.Widget {
    //     return &tui.Box{
    //         Border: tui.BorderSingle,
    //         Padding: layout.Edges{Top: 1, Right: 1, Bottom: 1, Left: 1},
    //         Style: layout.Style{FlexDirection: layout.Row, JustifyContent: layout.SpaceBetween},
    //         Children: []tui.Widget{
    //             &tui.Text{Content: title, Style: tui.Style{Bold: true}},
    //             &tui.Text{Content: "v1.0.0", Style: tui.Style{Dim: true}},
    //         },
    //     }
    // }
}
```

### 5.5 CLI Tool

```go
// cmd/tui/main.go
// tui generate [--watch] [path]

func main() {
    // Find .tui files
    // Parse each file
    // Generate *_tui.go alongside each .tui file
    // Optionally watch for changes
}
```

### 5.6 Deliverable

- `tui generate` compiles `.tui` files to Go
- Full round-trip: `.tui` → Go → running TUI
- Test: example app using generated components

---

## Phase 6: Polish and Additional Widgets

### 6.1 More Widgets

- `Input` — text input with cursor
- `Select` — dropdown selection
- `Table` — data grid
- `Scrollable` — container with scrolling
- `Tabs` — tabbed container
- `Progress` — progress bar
- `Spinner` — loading indicator

### 6.2 Theming

```go
type Theme struct {
    Primary   Style
    Secondary Style
    Border    Style
    Text      Style
    Dim       Style
    Error     Style
    Success   Style
}

var DefaultTheme = Theme{...}
var DarkTheme = Theme{...}
```

### 6.3 Developer Experience

- Clear error messages from lexer/parser with line numbers
- Hot reload in watch mode
- Example applications
- Documentation

---

## Directory Structure

```
go-tui/
├── cmd/
│   └── tui/              # CLI tool (tui generate)
│       └── main.go
├── pkg/
│   ├── tui/              # Core TUI package
│   │   ├── app.go        # Application loop
│   │   ├── buffer.go     # Character buffer
│   │   ├── cell.go       # Cell type
│   │   ├── border.go     # Border styles
│   │   ├── event.go      # Event types
│   │   ├── input.go      # Event reading
│   │   ├── focus.go      # Focus management
│   │   ├── rect.go       # Rectangle type
│   │   ├── render.go     # Tree rendering
│   │   ├── style.go      # Styling
│   │   ├── terminal.go   # Terminal interface
│   │   ├── widget.go     # Widget interface
│   │   └── widgets/      # Built-in widgets
│   │       ├── box.go
│   │       ├── text.go
│   │       └── list.go
│   ├── layout/           # Layout engine
│   │   ├── calculate.go  # Layout algorithm
│   │   ├── node.go       # Layout node
│   │   ├── style.go      # Layout style types
│   │   └── value.go      # Dimension values
│   └── tuigen/           # Code generator
│       ├── lexer.go
│       ├── parser.go
│       ├── ast.go
│       └── generate.go
├── examples/
│   ├── hello/            # Minimal example
│   ├── dashboard/        # Complex layout example
│   └── form/             # Input handling example
├── design.md
├── go.mod
└── go.sum
```

---

## Dependencies

**Required:**
- None initially (pure stdlib for terminal I/O)

**Optional (for better terminal handling):**
- `golang.org/x/term` — terminal raw mode (can be replaced with syscalls)

**Development only:**
- Testing framework (stdlib `testing`)

---

## Build Order Summary

| Phase | Focus | Output |
|-------|-------|--------|
| 1 | Terminal rendering | Buffer, Terminal, Cell, Rect |
| 2 | Layout engine | Node, Style, Calculate() |
| 3 | Widget system | Widget interface, Box, Text |
| 4 | Event system | Events, Focus, App loop |
| 5 | Code generation | Lexer, Parser, Generator, CLI |
| 6 | Polish | More widgets, theming, docs |

Each phase builds on the previous. Do not skip ahead—the later phases need the foundation to be solid.

---

## Next Steps

Begin with Phase 1.1: implement `Cell`, `Style`, and `Rect` types. Write tests. Then move to `Buffer`.
