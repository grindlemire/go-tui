# DSL & Code Generation Specification

**Status:** Draft\
**Version:** 1.2\
**Last Updated:** 2026-01-25

---

## 1. Overview

### Purpose

The DSL provides a declarative, templ-inspired syntax for building go-tui element trees. Unlike frameworks like Bubbletea or React, the DSL does NOT provide state management, event loops, or application structure — it is **purely syntactic sugar for element construction**.

Users write `.tui` files that compile to pure Go functions. The generated code uses the existing `element.Element` API with no runtime framework. This approach provides:

- Ergonomic element tree construction
- Type-safe Go code generation
- Zero runtime overhead
- Complete freedom in application architecture

### Goals

- **Templ-style syntax**: `@component`, `@for`, `@if`, `{expressions}`
- **Semantic element tags**: `<box>`, `<text>`, `<scrollable>`, `<button>`, `<input>`
- **Direct mapping to Element API**: Each tag maps to `element.New()` with options
- **Go expression embedding**: Full Go expressions via `{expr}`
- **Element references**: `@let` for capturing elements for later mutation
- **Component composition**: `@ComponentName(args) { children }` with `{children...}` slots
- **Standard Go imports**: Familiar import syntax at top of file
- **Zero framework**: No runtime, no state management, no magic

### Non-Goals

- State management (users choose their own approach)
- Event loop management (users write their own loops)
- Application structure (users design their architecture)
- Hot reloading (standard Go rebuild workflow)
- CSS-like selectors or styling language

---

## 2. Architecture

### Directory Structure

```
pkg/tuigen/
├── token.go          # Token types for lexer
├── lexer.go          # Lexer/tokenizer
├── lexer_test.go
├── ast.go            # Abstract Syntax Tree types
├── parser.go         # Parser (tokens → AST)
├── parser_test.go
├── analyzer.go       # Semantic analysis (type checking, import resolution)
├── analyzer_test.go
├── generator.go      # Code generator (AST → Go source)
├── generator_test.go
└── errors.go         # Error types with line/column info

cmd/tui/
├── main.go           # CLI entry point
├── generate.go       # generate subcommand
└── watch.go          # watch mode (optional, deferred)
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `token.go` | Token types: keywords, operators, literals, identifiers |
| `lexer.go` | Convert `.tui` source to token stream |
| `ast.go` | AST nodes: File, Component, Element, Expression, ControlFlow, ComponentCall |
| `parser.go` | Parse tokens into AST with error recovery |
| `analyzer.go` | Validate AST, resolve types, check imports |
| `generator.go` | Generate Go source code from validated AST |
| `errors.go` | Structured errors with source locations |

### Compilation Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  .tui Source File                                               │
│  package components                                             │
│  import "fmt"                                                   │
│  @component Counter(count int) { <text>{count}</text> }         │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Lexer
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Token Stream                                                   │
│  [PACKAGE, IDENT("components"), IMPORT, STRING("fmt"),          │
│   AT_COMPONENT, IDENT("Counter"), LPAREN, ...]                  │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Parser
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Abstract Syntax Tree                                           │
│  File{                                                          │
│    Package: "components",                                       │
│    Imports: [...],                                              │
│    Components: [Component{Name: "Counter", ...}]                │
│  }                                                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Analyzer
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Validated AST                                                  │
│  - Imports resolved                                             │
│  - Types checked                                                │
│  - Element tags validated                                       │
│  - Go expressions extracted                                     │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Generator
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Generated Go Source                                            │
│  // Code generated by tui generate. DO NOT EDIT.                │
│  package components                                             │
│  import "fmt"                                                   │
│  func Counter(count int) *element.Element { ... }               │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Core Entities

### 3.1 Token Types

```go
// pkg/tuigen/token.go

type TokenType int

const (
    // Special
    TokenEOF TokenType = iota
    TokenError
    TokenNewline
    TokenWhitespace

    // Keywords
    TokenPackage     // package
    TokenImport      // import
    TokenComponent   // @component
    TokenLet         // @let
    TokenFor         // @for
    TokenIf          // @if
    TokenElse        // @else
    TokenAtCall      // @ComponentName (uppercase, component call)
    TokenRange       // range

    // Literals
    TokenIdent       // identifier
    TokenInt         // 123
    TokenFloat       // 1.23
    TokenString      // "..."
    TokenRawString   // `...`

    // Operators and Punctuation
    TokenLParen      // (
    TokenRParen      // )
    TokenLBrace      // {
    TokenRBrace      // }
    TokenLAngle      // <
    TokenRAngle      // >
    TokenLBracket    // [
    TokenRBracket    // ]
    TokenSlash       // /
    TokenEquals      // =
    TokenComma       // ,
    TokenDot         // .
    TokenColon       // :
    TokenSemicolon   // ;
    TokenColonEquals // :=
    TokenAmpersand   // &
    TokenPipe        // |
    TokenStar        // *
    TokenPlus        // +
    TokenMinus       // -

    // Composite
    TokenGoExpr      // Go expression inside {}
)

type Token struct {
    Type     TokenType
    Literal  string
    Line     int
    Column   int
    StartPos int       // byte offset in source for raw capture
}
```

**Raw Source Capture**: The lexer supports capturing raw source text for complex Go constructs:
- `SourcePos() int` - Current byte position in source
- `SourceRange(start, end int) string` - Extract substring from source
- `ReadBalancedBracesFrom(pos int) (string, error)` - Read balanced `{...}` content from position
- `ReadUntilBrace() string` - Read until next `{`

This enables the parser to capture Go expressions, conditions, and types as raw source
instead of reconstructing them from tokens, preserving original formatting.

### 3.2 AST Types

```go
// pkg/tuigen/ast.go

// File represents a .tui source file
type File struct {
    Package    string
    Imports    []Import
    Components []*Component
    Funcs      []*GoFunc     // top-level Go functions
}

// Import represents a Go import
type Import struct {
    Alias string // optional alias
    Path  string
}

// Component represents a @component definition
type Component struct {
    Name            string
    Params          []*Param
    ReturnType      string   // defaults to "*element.Element"
    Body            []Node   // Elements, GoCode, ControlFlow, etc.
    AcceptsChildren bool     // true if body contains {children...}
    Position        Position
}

// Param represents a function parameter
type Param struct {
    Name string
    Type string
}

// Node is the interface for all AST nodes
type Node interface {
    node()
    Position() Position
}

// Element represents a <tag attr=value>children</tag>
type Element struct {
    Tag        string            // "box", "text", etc.
    Attributes []Attribute       // key=value pairs
    Children   []Node            // Elements, Expressions, ControlFlow
    SelfClose  bool              // <tag /> vs <tag></tag>
    Pos        Position
}

// Attribute represents tag=value or tag={expr}
type Attribute struct {
    Name  string
    Value Node   // StringLit, IntLit, or GoExpr
    Pos   Position
}

// GoExpr represents a Go expression {expr}
type GoExpr struct {
    Code string
    Pos  Position
}

// StringLit represents "string"
type StringLit struct {
    Value string
    Pos   Position
}

// IntLit represents 123
type IntLit struct {
    Value int64
    Pos   Position
}

// FloatLit represents 1.23
type FloatLit struct {
    Value float64
    Pos   Position
}

// LetBinding represents @let name = <element>
type LetBinding struct {
    Name    string
    Element *Element
    Pos     Position
}

// ForLoop represents @for i, v := range items { ... }
type ForLoop struct {
    Index    string   // optional, "_" if not used
    Value    string   // loop variable
    Iterable string   // Go expression for the iterable
    Body     []Node   // Elements and other nodes
    Pos      Position
}

// IfStmt represents @if condition { ... } @else { ... }
type IfStmt struct {
    Condition string   // Go expression (raw source)
    Then      []Node
    Else      []Node   // optional
    Position  Position
}

// GoCode represents raw Go statements in component bodies
// e.g., fmt.Printf("debug"), x := compute(), defer cleanup()
type GoCode struct {
    Code     string   // raw Go statement
    Position Position
}

// ComponentCall represents @ComponentName(args) { children }
type ComponentCall struct {
    Name     string   // component name (e.g., "Card", "Header")
    Args     string   // raw Go expression for arguments
    Children []Node   // child elements (may be empty if no children block)
    Position Position
}

// ChildrenSlot represents {children...} placeholder in a component body
type ChildrenSlot struct {
    Position Position
}

// GoFunc represents a top-level Go function definition
type GoFunc struct {
    Code     string   // raw function source including "func" keyword
    Position Position
}

// TextContent represents literal text inside elements
type TextContent struct {
    Text     string
    Position Position
}

// BoolLit represents true/false
type BoolLit struct {
    Value    bool
    Position Position
}

// Position tracks source location for errors
type Position struct {
    File   string
    Line   int
    Column int
}
```

### 3.3 Semantic Element Types

The DSL supports these semantic element tags, each mapping to specific Element options:

| Tag | Generated Options | Description |
|-----|-------------------|-------------|
| `<box>` | Default Element | Container with flexbox layout |
| `<text>` | WithText() | Text content with intrinsic sizing |
| `<scrollable>` | WithScrollable() | Scrollable container |
| `<button>` | WithOnEvent() for Enter, focusable | Interactive button |
| `<input>` | (Future) Text input field | User text entry |
| `<list>` | (Future) Scrollable list | Item selection |
| `<table>` | (Future) Data grid | Tabular data |
| `<progress>` | (Future) Progress bar | Progress indication |

### 3.4 Attribute Mapping

Attributes map to Element options:

```go
// pkg/tuigen/generator.go

var attributeToOption = map[string]string{
    // Dimensions
    "width":         "element.WithWidth(%s)",
    "widthPercent":  "element.WithWidthPercent(%s)",
    "height":        "element.WithHeight(%s)",
    "heightPercent": "element.WithHeightPercent(%s)",
    "minWidth":      "element.WithMinWidth(%s)",
    "minHeight":     "element.WithMinHeight(%s)",
    "maxWidth":      "element.WithMaxWidth(%s)",
    "maxHeight":     "element.WithMaxHeight(%s)",

    // Flex container
    "direction":     "element.WithDirection(%s)",
    "justify":       "element.WithJustify(%s)",
    "align":         "element.WithAlign(%s)",
    "gap":           "element.WithGap(%s)",

    // Flex item
    "flexGrow":      "element.WithFlexGrow(%s)",
    "flexShrink":    "element.WithFlexShrink(%s)",
    "alignSelf":     "element.WithAlignSelf(%s)",

    // Spacing
    "padding":       "element.WithPadding(%s)",
    "margin":        "element.WithMargin(%s)",

    // Visual
    "border":        "element.WithBorder(%s)",
    "borderStyle":   "element.WithBorderStyle(%s)",
    "background":    "element.WithBackground(%s)",

    // Text
    "text":          "element.WithText(%s)",
    "textStyle":     "element.WithTextStyle(%s)",
    "textAlign":     "element.WithTextAlign(%s)",

    // Focus
    "onFocus":       "element.WithOnFocus(%s)",
    "onBlur":        "element.WithOnBlur(%s)",
    "onEvent":       "element.WithOnEvent(%s)",

    // Scroll
    "scrollable":    "element.WithScrollable(%s)",
}
```

---

## 4. DSL Syntax

### 4.1 File Structure

```
// package declaration (required, first line)
package myapp

// imports (optional, standard Go syntax)
import (
    "fmt"
    "time"

    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
)

// components (one or more)
@component ComponentName(param1 Type1, param2 Type2) {
    // element tree
}
```

### 4.2 Component Definition

```
@component Name(params...) {
    <element>children</element>
}

// Examples:

// No parameters
@component Header() {
    <box border={tui.BorderSingle}>
        <text>My App</text>
    </box>
}

// With parameters
@component Button(label string, onClick func()) {
    <box
        border={tui.BorderSingle}
        padding={1}
        onEvent={func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                onClick()
                return true
            }
            return false
        }}
    >
        <text>{label}</text>
    </box>
}
```

### 4.3 Element Tags

```
// Self-closing
<text />

// With content
<text>Hello World</text>

// With expression content
<text>{fmt.Sprintf("Count: %d", count)}</text>

// With attributes
<box direction={layout.Column} gap={1} padding={2}>
    <text>Child 1</text>
    <text>Child 2</text>
</box>

// String attributes
<text textAlign={element.TextAlignCenter}>Centered</text>

// Integer attributes
<box width={40} height={10} />

// Go expression attributes
<box borderStyle={tui.NewStyle().Foreground(tui.Cyan)} />
```

### 4.4 Control Flow

```
// For loops
@for i, item := range items {
    <text>{fmt.Sprintf("%d: %s", i, item)}</text>
}

// For with index only
@for i := range items {
    <text>{strconv.Itoa(i)}</text>
}

// For with value only (use _ for index)
@for _, item := range items {
    <text>{item}</text>
}

// Conditional
@if showHeader {
    <Header title={title} />
}

// Conditional with else
@if isLoading {
    <text>Loading...</text>
} @else {
    <text>{data}</text>
}

// Conditional with else-if
@if err != nil {
    <text textStyle={errorStyle}>{err.Error()}</text>
} @else if len(items) == 0 {
    <text>No items</text>
} @else {
    <ItemList items={items} />
}
```

### 4.5 Element References (@let)

```
// Capture element for later mutation
@component Counter() {
    @let countText = <text>Count: 0</text>
    count := 0

    <box direction={layout.Column}>
        {countText}
        <box
            onEvent={func(e tui.Event) bool {
                if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                    count++
                    countText.SetText(fmt.Sprintf("Count: %d", count))
                    return true
                }
                return false
            }}
        >
            <text>Increment</text>
        </box>
    </box>
}
```

### 4.6 Embedded Go Code

Raw Go statements can be included directly in component bodies. The parser captures
complete statements from the start of the line until newline or semicolon (at bracket depth 0).

```
@component DynamicList(items []string) {
    // Variable assignments
    filtered := filterItems(items)
    count := len(filtered)

    // Function calls
    fmt.Printf("Rendering %d items\n", count)
    log.Debug("list render started")

    // Control flow (captured as single statements)
    if count == 0 { return emptyView() }

    // Defer, go, etc.
    defer cleanup()
    go backgroundTask()

    // For loops with bodies
    for i := 0; i < count; i++ { process(i) }

    // The element tree
    <box direction={layout.Column}>
        @for _, item := range filtered {
            <text>{item}</text>
        }
    </box>
}
```

**Supported Go statement starters:**
- Identifiers: `x := 1`, `fmt.Println()`, `myFunc()`
- Keywords: `if`, `for`, `switch`, `go`, `defer`, `return`
- Multi-line statements: Bracket depth is tracked, so statements spanning lines are captured correctly

**Note:** The `return <element>` syntax is not yet supported. Use DSL control flow (`@if`) for conditional rendering.

### 4.7 Component Calls with Children

Components can accept children using the `{children...}` slot, similar to templ's children pattern.

**Defining a component with children slot:**
```
@component Card(title string) {
    <box border={tui.BorderSingle}>
        <text bold>{title}</text>
        {children...}
    </box>
}
```

**Calling a component with children:**
```
@component App() {
    @Card("My Card") {
        <text>Line 1</text>
        <text>Line 2</text>
    }
}
```

**Calling a component without children:**
```
@component App() {
    @Header("Welcome")
}
```

**Rules:**
- Component calls use `@ComponentName(args)` syntax (uppercase first letter)
- Children block is optional: `@Header("title")` vs `@Card("title") { ... }`
- Components that accept children must have `{children...}` in their body
- Calling a component with children that doesn't have `{children...}` is an error
- Only one `{children...}` slot per component (multiple slots not supported)

**Children in control flow:**
```
@component ConditionalWrapper(show bool) {
    @if show {
        <box border={tui.BorderSingle}>
            {children...}
        </box>
    } @else {
        {children...}
    }
}
```

---

## 5. Code Generation

### 5.1 Generated File Structure

```go
// Code generated by tui generate. DO NOT EDIT.
// Source: components.tui

package myapp

import (
    "fmt"

    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

// Header renders the application header.
func Header() *element.Element {
    __tui_0 := element.New(
        element.WithBorder(tui.BorderSingle),
    )
    __tui_1 := element.New(
        element.WithText("My App"),
    )
    __tui_0.AddChild(__tui_1)
    return __tui_0
}
```

### 5.2 Variable Naming

Generated code uses deterministic variable names:
- `__tui_0`, `__tui_1`, etc. for anonymous elements
- User-defined names from `@let` bindings are preserved
- Prefixed with `__tui_` to avoid conflicts with user code

### 5.3 Element Tag Mapping

| Tag | Generation Strategy |
|-----|---------------------|
| `<box>` | `element.New(options...)` |
| `<text>` | `element.New(element.WithText(content), options...)` |
| `<text>literal</text>` | WithText with literal string |
| `<text>{expr}</text>` | WithText with expression |
| `<scrollable>` | `element.New(element.WithScrollable(mode), options...)` |
| `<button>` | element.New with onEvent for KeyEnter |

### 5.4 Control Flow Generation

```
// Source:
@for i, item := range items {
    <text>{item}</text>
}

// Generated:
for i, item := range items {
    _ = i // silence unused warning if not used
    __tui_child := element.New(
        element.WithText(item),
    )
    __tui_parent.AddChild(__tui_child)
}
```

### 5.5 @let Generation

```
// Source:
@let counter = <text>0</text>

// Generated:
counter := element.New(
    element.WithText("0"),
)
```

### 5.6 Component Call Generation

**Components with children parameter:**
```
// Source:
@component Card(title string) {
    <box border={tui.BorderSingle}>
        <text>{title}</text>
        {children...}
    </box>
}

// Generated:
func Card(title string, children []*element.Element) *element.Element {
    __tui_0 := element.New(
        element.WithBorder(tui.BorderSingle),
    )
    __tui_1 := element.New(
        element.WithText(title),
    )
    __tui_0.AddChild(__tui_1)
    for _, __child := range children {
        __tui_0.AddChild(__child)
    }
    return __tui_0
}
```

**Calling components without children:**
```
// Source:
@Header("Welcome")

// Generated:
__tui_0 := Header("Welcome")
```

**Calling components with children:**
```
// Source:
@Card("My Card") {
    <text>Line 1</text>
    <text>Line 2</text>
}

// Generated:
__card_children := []*element.Element{}
__tui_1 := element.New(
    element.WithText("Line 1"),
)
__card_children = append(__card_children, __tui_1)
__tui_2 := element.New(
    element.WithText("Line 2"),
)
__card_children = append(__card_children, __tui_2)
__tui_0 := Card("My Card", __card_children)
```

**Children parameter position:** Always last, after all user-defined parameters.

---

## 6. CLI Tool

### 6.1 Commands

```bash
# Generate Go code from .tui files
tui generate [path...]

# Generate with verbose output
tui generate -v ./...

# Watch mode (future)
tui generate --watch ./...

# Check syntax without generating
tui check [path...]

# Format .tui files (future)
tui fmt [path...]
```

### 6.2 File Discovery

- `tui generate ./...` — Recursively find all `.tui` files
- `tui generate ./components` — Process directory
- `tui generate header.tui footer.tui` — Process specific files

### 6.3 Output Naming

| Input | Output |
|-------|--------|
| `header.tui` | `header_tui.go` |
| `components.tui` | `components_tui.go` |
| `my-app.tui` | `my_app_tui.go` |

### 6.4 Error Reporting

```
components.tui:15:8: error: unknown attribute "colour" (did you mean "color"?)
  <box colour="red">
       ^~~~~~~

components.tui:23:4: error: unclosed element <box>
  <box direction={layout.Column}>
  ^~~~
```

---

## 7. User Experience

### 7.1 Complete Example

This example demonstrates a realistic todo-list app where:
- `selectedIndex` state is passed to components for highlighting
- Selection changes trigger a UI rebuild
- Items can be toggled complete/incomplete

```
// components.tui
package app

import (
    "fmt"

    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
)

// Item is the data model
type Item struct {
    Name   string
    Done   bool
}

// Dashboard renders the full todo list UI.
// selectedIndex: which item is highlighted (-1 for none)
// onMove: callback when user moves selection (delta: -1 or +1)
// onToggle: callback when user toggles current item
@component Dashboard(items []Item, selectedIndex int, onMove func(int), onToggle func()) {
    <box direction={layout.Column} padding={1}>
        <Header title="Todo List" count={countDone(items)} total={len(items)} />

        <box direction={layout.Column} flexGrow={1}>
            @for i, item := range items {
                <ItemRow
                    item={item}
                    selected={i == selectedIndex}
                    onSelect={func() { onToggle() }}
                />
            }
        </box>

        <Footer />
    </box>
}

// Helper function (pure Go, included in .tui file)
func countDone(items []Item) int {
    count := 0
    for _, item := range items {
        if item.Done {
            count++
        }
    }
    return count
}

@component Header(title string, count int, total int) {
    <box
        border={tui.BorderRounded}
        borderStyle={tui.NewStyle().Foreground(tui.Cyan)}
        padding={1}
        direction={layout.Row}
        justify={layout.JustifySpaceBetween}
    >
        <text textStyle={tui.NewStyle().Bold()}>{title}</text>
        <text textStyle={tui.NewStyle().Foreground(tui.Green)}>
            {fmt.Sprintf("%d/%d done", count, total)}
        </text>
    </box>
}

// ItemRow renders a single todo item.
// selected: true if this row is highlighted
@component ItemRow(item Item, selected bool, onSelect func()) {
    // Compute styles based on state
    borderStyle := tui.Style{}
    textStyle := tui.Style{}

    @if selected {
        borderStyle = tui.NewStyle().Foreground(tui.Yellow)
    }

    @if item.Done {
        textStyle = tui.NewStyle().Foreground(tui.BrightBlack).Strikethrough()
    }

    <box
        direction={layout.Row}
        gap={2}
        border={tui.BorderSingle}
        borderStyle={borderStyle}
        padding={0, 1, 0, 1}
        onEvent={func(e tui.Event) bool {
            if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                onSelect()
                return true
            }
            return false
        }}
    >
        <text width={3}>
            @if item.Done {
                {"\u2713"}  // checkmark
            } @else {
                {" "}
            }
        </text>
        <text flexGrow={1} textStyle={textStyle}>{item.Name}</text>
    </box>
}

@component Footer() {
    <box padding={1, 0, 0, 0}>
        <text textStyle={tui.NewStyle().Dim()}>
            ↑/↓ move • Enter toggle • Esc quit
        </text>
    </box>
}
```

### 7.2 Using Generated Code: State Store Pattern (Recommended)

The key insight: **create elements once, then mutate them when state changes**.

Instead of rebuilding the entire tree, we use a **state store** that:
1. Holds application state
2. Holds references to elements that need updating
3. Provides methods to update state AND the relevant elements

```go
// store.go - Application state store (you write this in Go)
package main

import (
    "fmt"

    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

// Item is the data model
type Item struct {
    Name string
    Done bool
}

// Store holds application state and element references.
// This is NOT generated - you write this based on your app's needs.
type Store struct {
    // State
    items         []Item
    selectedIndex int

    // Element references (populated when building UI)
    itemRows    []*element.Element
    headerCount *element.Element

    // Styles
    selectedStyle   tui.Style
    unselectedStyle tui.Style
    doneStyle       tui.Style
    normalStyle     tui.Style
}

func NewStore(items []Item) *Store {
    return &Store{
        items:           items,
        selectedIndex:   0,
        selectedStyle:   tui.NewStyle().Foreground(tui.Yellow),
        unselectedStyle: tui.Style{},
        doneStyle:       tui.NewStyle().Foreground(tui.BrightBlack).Strikethrough(),
        normalStyle:     tui.Style{},
    }
}

// MoveSelection changes selection and updates relevant elements.
func (s *Store) MoveSelection(delta int) {
    oldIndex := s.selectedIndex

    s.selectedIndex += delta
    if s.selectedIndex < 0 {
        s.selectedIndex = len(s.items) - 1
    } else if s.selectedIndex >= len(s.items) {
        s.selectedIndex = 0
    }

    // Only update the two affected rows
    if oldIndex != s.selectedIndex {
        s.itemRows[oldIndex].SetBorderStyle(s.unselectedStyle)
        s.itemRows[s.selectedIndex].SetBorderStyle(s.selectedStyle)
    }
}

// ToggleSelected toggles the current item and updates relevant elements.
func (s *Store) ToggleSelected() {
    s.items[s.selectedIndex].Done = !s.items[s.selectedIndex].Done

    // Update the affected row
    row := s.itemRows[s.selectedIndex]
    item := s.items[s.selectedIndex]

    if item.Done {
        // Find the text child and update its style
        children := row.Children()
        if len(children) >= 2 {
            children[0].SetText("✓")
            children[1].SetTextStyle(s.doneStyle)
        }
    } else {
        children := row.Children()
        if len(children) >= 2 {
            children[0].SetText(" ")
            children[1].SetTextStyle(s.normalStyle)
        }
    }

    // Update header count
    s.headerCount.SetText(fmt.Sprintf("%d/%d done", s.countDone(), len(s.items)))
}

func (s *Store) countDone() int {
    count := 0
    for _, item := range s.items {
        if item.Done {
            count++
        }
    }
    return count
}
```

Now the DSL components capture element references for the store:

```
// components.tui
package main

import (
    "fmt"

    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
)

// Dashboard builds the UI and populates store's element references.
@component Dashboard(store *Store) {
    // Capture header count element
    @let headerCount = <text textStyle={tui.NewStyle().Foreground(tui.Green)}>
        {fmt.Sprintf("%d/%d done", store.countDone(), len(store.items))}
    </text>
    store.headerCount = headerCount

    // Build item rows and capture references
    store.itemRows = make([]*element.Element, len(store.items))

    <box direction={layout.Column} padding={1}>
        <Header title="Todo List" countElement={headerCount} />

        <box direction={layout.Column} flexGrow={1}>
            @for i, item := range store.items {
                @let row = <ItemRow
                    item={item}
                    selected={i == store.selectedIndex}
                    store={store}
                />
                store.itemRows[i] = row
                {row}
            }
        </box>

        <Footer />
    </box>
}

@component Header(title string, countElement *element.Element) {
    <box
        border={tui.BorderRounded}
        borderStyle={tui.NewStyle().Foreground(tui.Cyan)}
        padding={1}
        direction={layout.Row}
        justify={layout.JustifySpaceBetween}
    >
        <text textStyle={tui.NewStyle().Bold()}>{title}</text>
        {countElement}
    </box>
}

@component ItemRow(item Item, selected bool, store *Store) {
    borderStyle := tui.Style{}
    @if selected {
        borderStyle = store.selectedStyle
    }

    checkmark := " "
    textStyle := store.normalStyle
    @if item.Done {
        checkmark = "✓"
        textStyle = store.doneStyle
    }

    <box
        direction={layout.Row}
        gap={2}
        border={tui.BorderSingle}
        borderStyle={borderStyle}
        padding={0, 1, 0, 1}
    >
        <text width={3}>{checkmark}</text>
        <text flexGrow={1} textStyle={textStyle}>{item.Name}</text>
    </box>
}

@component Footer() {
    <box padding={1, 0, 0, 0}>
        <text textStyle={tui.NewStyle().Dim()}>
            ↑/↓ move • Enter toggle • Esc quit
        </text>
    </box>
}
```

```go
// main.go
package main

import (
    "log"
    "time"

    "github.com/grindlemire/go-tui/pkg/tui"
)

func main() {
    tuiApp, err := tui.NewApp()
    if err != nil {
        log.Fatal(err)
    }
    defer tuiApp.Close()

    // Create state store
    store := NewStore([]Item{
        {Name: "Buy groceries", Done: false},
        {Name: "Write DSL design", Done: true},
        {Name: "Review pull request", Done: false},
    })

    // Build UI once - elements are created and references stored
    root := Dashboard(store)
    tuiApp.SetRoot(root)

    // Event loop - state changes mutate elements directly, no rebuild
    for {
        event, ok := tuiApp.PollEvent(50 * time.Millisecond)
        if ok {
            switch e := event.(type) {
            case tui.KeyEvent:
                switch e.Key {
                case tui.KeyEscape:
                    return
                case tui.KeyUp:
                    store.MoveSelection(-1) // Mutates elements directly
                case tui.KeyDown:
                    store.MoveSelection(1)  // Mutates elements directly
                case tui.KeyEnter:
                    store.ToggleSelected()  // Mutates elements directly
                }
            }
        }
        tuiApp.Render()
    }
}
```

**Key benefits of this pattern:**
- Elements created once, mutated in place
- Only affected elements are updated
- Store centralizes state AND element references
- State changes are explicit methods, easy to test
- No garbage from rebuilding

### 7.3 Simpler Alternative: Direct Mutation Pattern

For simpler apps, you can use closures and direct element mutation instead of rebuilding:

```
// counter.tui - Uses @let for direct mutation
package app

import (
    "fmt"

    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
)

@component Counter() {
    count := 0

    // Capture the text element so we can mutate it later
    @let countText = <text textStyle={tui.NewStyle().Bold()}>Count: 0</text>

    <box direction={layout.Column} gap={1} padding={2}>
        {countText}

        <box direction={layout.Row} gap={2}>
            <box
                border={tui.BorderSingle}
                padding={0, 2, 0, 2}
                onEvent={func(e tui.Event) bool {
                    if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                        count++
                        countText.SetText(fmt.Sprintf("Count: %d", count))
                        return true
                    }
                    return false
                }}
            >
                <text>+</text>
            </box>

            <box
                border={tui.BorderSingle}
                padding={0, 2, 0, 2}
                onEvent={func(e tui.Event) bool {
                    if ke, ok := e.(tui.KeyEvent); ok && ke.Key == tui.KeyEnter {
                        count--
                        countText.SetText(fmt.Sprintf("Count: %d", count))
                        return true
                    }
                    return false
                }}
            >
                <text>-</text>
            </box>
        </box>
    </box>
}
```

```go
// main.go - Direct mutation is simpler for small apps
func main() {
    tuiApp, _ := tui.NewApp()
    defer tuiApp.Close()

    // No rebuild needed - component manages its own state
    root := app.Counter()
    tuiApp.SetRoot(root)

    for {
        event, ok := tuiApp.PollEvent(50 * time.Millisecond)
        if ok {
            if ke, ok := event.(tui.KeyEvent); ok && ke.Key == tui.KeyEscape {
                return
            }
            tuiApp.Dispatch(event)
        }
        tuiApp.Render()
    }
}
```

---

## 8. Complexity Assessment

| Factor | Assessment |
|--------|------------|
| Lexer | Moderate — escape sequences, multi-line strings, Go expressions, raw source capture |
| Parser | Moderate — recursive descent, error recovery, raw source preservation |
| AST | Low — straightforward tree structure |
| Analyzer | Low — basic type checking, import validation |
| Generator | Moderate — template-based, maintains formatting |
| CLI | Low — standard file discovery pattern |
| Testing | Significant — many edge cases in parsing |

### Implementation Notes

**Raw Source Capture Strategy**: Instead of tokenizing Go expressions and reconstructing
them (which loses formatting and is fragile), the parser captures raw source text:
- Token struct includes `StartPos` for byte offset tracking
- Lexer provides `SourceRange(start, end)` to extract source substrings
- Parser uses this for: Go expressions `{...}`, conditions, iterables, type signatures, raw statements

**Benefits:**
- Preserves original formatting
- Handles arbitrary Go syntax without parser changes
- Avoids edge cases in reconstruction

**Assessed Size:** Medium\
**Recommended Phases:** 4

**Rationale:**
1. **Phase 1: Core Types & Lexer** — Token types, lexer with Go expression support
2. **Phase 2: Parser & AST** — Recursive descent parser, AST construction
3. **Phase 3: Code Generator** — AST to Go source generation
4. **Phase 4: CLI & Integration** — Command-line tool, file discovery, error reporting

---

## 9. Success Criteria

1. `.tui` files parse correctly with helpful error messages
2. Generated Go code compiles without errors
3. Generated code matches expected Element API usage
4. `@for` and `@if` control flow generates correct Go code
5. `@let` bindings capture element references correctly
6. Imports are correctly propagated to generated files
7. `tui generate ./...` recursively processes all .tui files
8. Error messages include file, line, and column numbers
9. Generated code is formatted (gofmt compatible)
10. Dashboard example can be rewritten using DSL with equivalent functionality
11. Component calls with `@ComponentName(args)` generate correct function calls
12. Components with `{children...}` receive children parameter and iterate correctly

---

## 10. Future Extensions (Out of Scope)

These are explicitly deferred to future phases:

1. **Watch mode**: `tui generate --watch` for automatic regeneration
2. **IDE integration**: LSP server for syntax highlighting and autocomplete
3. **Hot reloading**: Live reload without restart
4. **Input widget**: `<input>` for text entry
5. **List widget**: `<list>` for selectable lists
6. **Table widget**: `<table>` for data grids
7. **Theming**: Theme definitions and application
8. **Reactive bindings**: Optional signal-based state management

---

## 11. Open Questions

1. **Should components support generic types?**\
   → Deferred. Start with concrete types, add generics if needed.

2. **How to handle multi-line text content?**\
   → Use Go raw strings inside expressions: `{`\`multi\nline\`}`.

3. **Should attributes support shorthand like `disabled` instead of `disabled={true}`?**\
   → Yes, boolean attributes can be shorthand. Value defaults to `true`.

4. **How to handle attribute values that are method calls vs callbacks?**\
   → Parser doesn't distinguish. Generator treats `onX` attributes as callbacks.

5. **Should there be a standard library of composed widgets?**\
   → Deferred. Users can create their own component libraries.
