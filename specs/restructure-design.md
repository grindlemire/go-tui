# Codebase Restructure Specification

**Status:** Planned\
**Version:** 1.0\
**Last Updated:** 2025-07-17

---

## 1. Overview

### Purpose

Restructure the go-tui codebase to establish a clean public/private API boundary, reduce the user-facing import surface to a single package, move tooling internals behind Go's `internal/` convention, split oversized files, and reorganize tests for maintainability.

### Goals

- **Single-import public API**: Users import `"github.com/grindlemire/go-tui"` and get everything — `tui.NewApp()`, `tui.New()` (element), `tui.Column`, `tui.BorderSingle`, etc.
- **Internal tooling**: Move `tuigen`, `formatter`, `lsp`, and `debug` to `internal/` — these are implementation details of `cmd/tui`, not user-facing
- **Internal layout engine**: Move the layout algorithm to `internal/layout/` — users access layout types via re-exports in the root package
- **Idiomatic file sizes**: No source file exceeds ~500 lines; split by responsibility
- **Test organization**: No test file exceeds ~500 lines; split by category
- **Clear architecture documentation**: Package-level doc comments explaining how packages fit together

### Non-Goals

- Changing the layout algorithm itself
- Changing the DSL syntax or compiler behavior
- Changing the LSP protocol implementation
- Adding new features — this is purely structural
- Changing the module path (`github.com/grindlemire/go-tui` stays the same)

---

## 2. Architecture

### Current Structure (Problems)

```
pkg/
├── tui/              # GOD PACKAGE: 24 files, ~5,000 lines
│   │                 # Mixes public API with ANSI escapes, raw mode,
│   │                 # input parsing, mock implementations
│   └── element/      # Separate package requiring 2nd import
├── layout/           # 3rd import required for Direction/Justify/etc.
├── tuigen/           # Public but only used by cmd/tui
├── formatter/        # Public but only used by cmd/tui
├── lsp/              # Public but only used by cmd/tui
└── debug/            # Public but only used internally
```

**Issues:**
1. End users need **3 imports** for basic usage: `tui`, `tui/element`, `layout`
2. `pkg/tui` is a god package mixing user API with implementation details
3. Tooling packages (`tuigen`, `formatter`, `lsp`) are public despite being internal-only
4. 14 source files exceed 500 lines; 7 test files exceed 1,000 lines
5. No `internal/` boundary — everything is importable by external code

### Proposed Structure

```
github.com/grindlemire/go-tui/
│
├── (root package: "tui") ← SINGLE PUBLIC IMPORT
│   ├── doc.go                    # Package documentation
│   │
│   │  ── App ──
│   ├── app.go                    # App type, NewApp, Run, Close
│   ├── app_options.go            # AppOption funcs (WithRoot, WithFrameRate, etc.)
│   │
│   │  ── Element ──
│   ├── element.go                # Element struct, New(), core methods
│   ├── element_tree.go           # AddChild, RemoveChild, tree walking
│   ├── element_options.go        # With* option functions
│   ├── element_render.go         # Render method, tree rendering
│   ├── element_scroll.go         # Scroll methods and types
│   │
│   │  ── Visual Types ──
│   ├── style.go                  # Style, Attr
│   ├── color.go                  # Color type, named colors
│   ├── border.go                 # BorderStyle, border chars
│   │
│   │  ── Layout Types (re-exported) ──
│   ├── layout.go                 # Type aliases + re-exports from internal/layout
│   │
│   │  ── Events & Input ──
│   ├── event.go                  # Event, KeyEvent, MouseEvent, ResizeEvent
│   ├── key.go                    # Key constants, Mod type
│   │
│   │  ── Reactive State ──
│   ├── state.go                  # State[T], NewState, Batch
│   ├── watcher.go                # Watcher, Watch, OnTimer
│   │
│   │  ── Focus ──
│   ├── focus.go                  # Focusable interface, FocusManager
│   │
│   │  ── Rendering Infrastructure ──
│   ├── buffer.go                 # Buffer (double-buffered cell grid)
│   ├── cell.go                   # Cell, CellChange
│   ├── rect.go                   # Rect/Edges (aliases to internal/layout)
│   ├── render.go                 # Render, RenderFull, RenderRegion
│   │
│   │  ── Terminal (interface public, implementation private) ──
│   ├── terminal.go               # Terminal interface, Capabilities
│   ├── terminal_ansi.go          # ANSITerminal (unexported constructor used by App)
│   ├── terminal_unix.go          # Unix raw mode
│   │
│   │  ── Input (private implementation) ──
│   ├── escape.go                 # ANSI escape sequence builder
│   ├── reader.go                 # EventReader
│   ├── reader_unix.go            # Unix-specific reader
│   ├── parse.go                  # ANSI input sequence parser
│   ├── caps.go                   # Terminal capability detection
│   └── dirty.go                  # Dirty tracking utility
│
├── internal/
│   ├── layout/                   # Pure flexbox engine (zero external deps)
│   │   ├── calculate.go          # Calculate() — main entry point
│   │   ├── flex.go               # Flex distribution algorithm
│   │   ├── layoutable.go         # Layoutable interface
│   │   ├── style.go              # Style, Direction, Justify, Align
│   │   ├── value.go              # Value (Fixed, Percent, Auto)
│   │   ├── rect.go               # Rect, Size
│   │   ├── edges.go              # Edges type
│   │   ├── point.go              # Point type
│   │   └── layout.go             # Layout result type
│   │
│   ├── tuigen/                   # DSL compiler (.gsx → Go)
│   │   ├── ast.go                # AST node types
│   │   ├── token.go              # Token types
│   │   ├── errors.go             # Compiler errors
│   │   ├── lexer.go              # Core lexer loop
│   │   ├── lexer_string.go       # String/template lexing
│   │   ├── lexer_comment.go      # Comment lexing
│   │   ├── parser.go             # Core parser, file/imports
│   │   ├── parser_element.go     # Element/attribute parsing
│   │   ├── parser_control.go     # @if, @for, @let parsing
│   │   ├── parser_comment.go     # Comment attachment
│   │   ├── analyzer.go           # Semantic analysis core
│   │   ├── analyzer_types.go     # Type checking, import resolution
│   │   ├── generator.go          # Code gen core, file structure
│   │   ├── generator_element.go  # Element/option generation
│   │   ├── generator_view.go     # View struct generation
│   │   ├── tailwind.go           # Tailwind class → options
│   │   └── tailwind_data.go      # Class lookup tables (if large)
│   │
│   ├── formatter/                # Code formatter
│   │   ├── formatter.go          # Format() entry point
│   │   ├── printer.go            # Core printing logic
│   │   ├── printer_element.go    # Element printing
│   │   ├── printer_control.go    # Control flow printing
│   │   ├── printer_comment.go    # Comment printing
│   │   └── imports.go            # Import management
│   │
│   ├── lsp/                      # Language server
│   │   ├── server.go             # LSP server lifecycle
│   │   ├── handler.go            # Request dispatch
│   │   ├── document.go           # Document management
│   │   ├── diagnostics.go        # Error reporting
│   │   ├── completion.go         # Auto-completion
│   │   ├── completion_attr.go    # Attribute completions
│   │   ├── definition.go         # Go-to-definition
│   │   ├── hover.go              # Hover info (core)
│   │   ├── hover_attribute.go    # Hover for attributes
│   │   ├── hover_element.go      # Hover for elements
│   │   ├── references.go         # Find references
│   │   ├── semantic_tokens.go    # Semantic token core
│   │   ├── semantic_tokens_go.go # Go expression tokens
│   │   ├── symbols.go            # Document symbols
│   │   ├── formatting.go         # Format integration
│   │   ├── index.go              # Symbol index
│   │   └── gopls/                # gopls proxy (unchanged)
│   │       ├── proxy.go
│   │       ├── mapping.go
│   │       └── generate.go
│   │
│   └── debug/                    # Debug logging
│       └── debug.go
│
├── cmd/tui/                      # CLI (unchanged structure)
│   ├── main.go
│   ├── generate.go
│   ├── check.go
│   ├── fmt.go
│   └── lsp.go
│
├── editor/                       # Editor support (unchanged)
├── examples/                     # Updated imports
└── specs/                        # Design docs
```

### Dependency Graph

```
                    ┌──────────────────────┐
                    │   cmd/tui (CLI)      │
                    └──────────┬───────────┘
                               │ imports
            ┌──────────────────┼──────────────────┐
            ▼                  ▼                   ▼
   internal/tuigen    internal/formatter    internal/lsp
            │                  │                   │
            └──────────────────┴───────────────────┘
                               │ imports
                               ▼
                    ┌──────────────────────┐
                    │  root: package tui   │ ◄── USER IMPORTS THIS
                    │  (public API)        │
                    └──────────┬───────────┘
                               │ imports
                               ▼
                    ┌──────────────────────┐
                    │  internal/layout     │
                    │  (pure flexbox)      │
                    └──────────────────────┘
                               │ imports
                               ▼
                          (stdlib only)
```

**Key dependency rules:**
- `internal/layout` → stdlib only (pure algorithm, zero deps)
- Root `tui` → `internal/layout` + stdlib
- `internal/tuigen` → root `tui` (for import path references in generated code) + stdlib
- `internal/formatter` → `internal/tuigen` (AST types)
- `internal/lsp` → `internal/tuigen`, `internal/formatter`
- `cmd/tui` → `internal/tuigen`, `internal/formatter`, `internal/lsp`
- User code → root `tui` only (single import)

---

## 3. Core Entities

### Public API Surface (root `tui` package)

The root package exports everything an end user or generated code needs:

```go
// App lifecycle
type App struct { ... }
func NewApp(opts ...AppOption) (*App, error)
type AppOption func(*App) error
func Stop()

// Element construction
type Element struct { ... }
func New(opts ...Option) *Element
type Option func(*Element)

// Element options (subset — many With* functions)
func WithText(s string) Option
func WithBorder(style BorderStyle) Option
func WithDirection(d Direction) Option
func WithWidth(n int) Option
// ... ~30 more options

// Visual types
type Style struct { Fg Color; Bg Color; Attrs Attr }
type Color struct { ... }
type BorderStyle uint8
const (BorderNone, BorderSingle, BorderDouble, BorderRounded, BorderThick)
type Attr uint8
const (AttrBold, AttrDim, AttrItalic, ...)

// Layout types (re-exported from internal/layout via type aliases)
type Direction = layout.Direction
const (Row, Column)
type Justify = layout.Justify
const (JustifyStart, JustifyCenter, JustifyEnd, ...)
type Align = layout.Align
const (AlignStart, AlignCenter, AlignEnd, AlignStretch)
type Value = layout.Value
func Fixed(n int) Value
func Percent(n float64) Value
func Auto() Value

// Events
type Event interface{}
type KeyEvent struct { Key Key; Rune rune; Mod Modifiers }
type MouseEvent struct { ... }
type ResizeEvent struct { Width, Height int }
type Key int
const (KeyEscape, KeyEnter, KeyTab, KeyBackspace, ...)

// Reactive state
type State[T any] struct { ... }
func NewState[T any](initial T) *State[T]
func Batch(fn func())

// Watchers
type Watcher interface { Start(chan<- func(), <-chan struct{}) }
func Watch[T any](ch <-chan T, handler func(T)) Watcher
func OnTimer(interval time.Duration, handler func()) Watcher

// Focus
type Focusable interface { ... }
type FocusManager struct { ... }

// Rendering
type Buffer struct { ... }
func NewBuffer(w, h int) *Buffer
type Rect = layout.Rect
func Render(term Terminal, buf *Buffer)
```

### Layout Types Re-export Pattern

```go
// layout.go — in root tui package
package tui

import "github.com/grindlemire/go-tui/internal/layout"

// Direction specifies the main axis for laying out children.
type Direction = layout.Direction

const (
    Row    = layout.Row
    Column = layout.Column
)

// Justify specifies how children are distributed along the main axis.
type Justify = layout.Justify

const (
    JustifyStart        = layout.JustifyStart
    JustifyEnd          = layout.JustifyEnd
    JustifyCenter       = layout.JustifyCenter
    JustifySpaceBetween = layout.JustifySpaceBetween
    JustifySpaceAround  = layout.JustifySpaceAround
    JustifySpaceEvenly  = layout.JustifySpaceEvenly
)

// ... Align, Value, Rect, Edges, etc.
```

---

## 4. User Experience

### Before (3 imports)

```go
import (
    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

func main() {
    app, _ := tui.NewApp()
    root := element.New(
        element.WithDirection(layout.Column),
        element.WithBorder(tui.BorderRounded),
        element.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan)),
    )
    // ...
}
```

### After (1 import)

```go
import "github.com/grindlemire/go-tui"

func main() {
    app, _ := tui.NewApp()
    root := tui.New(
        tui.WithDirection(tui.Column),
        tui.WithBorder(tui.BorderRounded),
        tui.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan)),
    )
    // ...
}
```

### Generated Code (Before vs After)

**Before:**
```go
import (
    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

type HelloView struct {
    Root     *element.Element
    watchers []tui.Watcher
}

func Hello() HelloView {
    __tui_0 := element.New(
        element.WithDirection(layout.Column),
        element.WithTextStyle(tui.NewStyle().Bold()),
    )
    // ...
}
```

**After:**
```go
import "github.com/grindlemire/go-tui"

type HelloView struct {
    Root     *tui.Element
    watchers []tui.Watcher
}

func Hello() HelloView {
    __tui_0 := tui.New(
        tui.WithDirection(tui.Column),
        tui.WithTextStyle(tui.NewStyle().Bold()),
    )
    // ...
}
```

---

## 5. Complexity Assessment

This restructuring touches every package in the project. It requires:

1. Moving files between directories (internal/ boundary)
2. Merging 3 packages into 1 root package (tui + element + layout types)
3. Splitting 14+ oversized source files
4. Splitting 7+ oversized test files
5. Updating the code generator to emit new import paths
6. Updating all examples
7. Updating cmd/tui imports
8. Updating CLAUDE.md

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| **Large** | **5-6** | **Cross-cutting feature, new subsystem** |

**Assessed Size:** Large\
**Recommended Phases:** 6\
**Rationale:** This is a cross-cutting structural change touching every package, every import path, the code generator output, all examples, and all tests. However, no logic changes — it's purely mechanical moves, renames, splits, and re-exports. Six phases allows clean incremental progress: (1) internal/layout, (2) merge element+tui into root, (3) move tooling to internal/, (4) split oversized files, (5) split oversized tests, (6) update generator/examples/docs.

> **IMPORTANT:** User must approve the complexity assessment before proceeding to implementation plan. The plan MUST use the approved number of phases.

---

## 6. Success Criteria

1. **Single import**: End users and generated code import only `"github.com/grindlemire/go-tui"`, accessing all types as `tui.X`
2. **Internal boundary**: `tuigen`, `formatter`, `lsp`, `debug`, and `layout` are under `internal/` — external code cannot import them
3. **No source file >500 lines**: All `.go` files (excluding generated code and test fixtures) stay under 500 lines
4. **No test file >500 lines**: All `*_test.go` files stay under 500 lines
5. **All tests pass**: `go test ./...` passes with no regressions
6. **All examples build and run**: Updated imports compile and behave identically
7. **Generated code works**: `tui generate` produces valid code using the new single-import paths
8. **Package docs**: Every package has a `doc.go` or package comment explaining its purpose and how it fits into the architecture
9. **No circular dependencies**: Import graph remains acyclic
10. **CLAUDE.md updated**: Architecture docs reflect the new structure

---

## 7. Open Questions

1. ~~Single import vs multi-import?~~ → Single import (user confirmed)
2. ~~Clean break vs incremental?~~ → Clean break (user confirmed)
3. ~~Module root vs pkg/ for public API?~~ → Module root (user confirmed)
4. ~~Test split threshold?~~ → Aggressive: split any file >500 lines (user confirmed)
5. Should `mock_terminal.go` and `mock_reader.go` stay in the root package or move to an `internal/testutil` package? → Keep in root as unexported types (they implement public interfaces and are useful for user tests)
6. Should the `internal/layout/` package keep its own test files or should layout tests live in the root package? → Keep in `internal/layout/` — the layout engine is independently testable and its tests don't need the rest of the framework
