# go-tui Project Guidelines

A declarative terminal UI framework for Go with templ-like syntax and flexbox layout.

## Git Commits IMPORTANT

Use `gcommit -m ""` for all commits to ensure proper signing.

ONLY EVER COMMIT USING THIS APPROACH

## Project Overview

go-tui allows defining UIs in `.gsx` files that compile to type-safe Go code. The framework provides:

- Declarative component syntax (similar to templ/JSX)
- Pure Go flexbox layout engine (no CGO)
- Minimal external dependencies (golang.org/x/{mod,sync,sys,tools})
- Composable widget system
- Reactive state management with generic `State[T]`
- Language server, formatter, and tree-sitter grammar for editor support

## Architecture

```
.gsx files (declarative syntax)
        │ tui generate (internal/tuigen)
        ▼
Generated Go code (*_gsx.go)
        │ imports tui "github.com/grindlemire/go-tui"
        ▼
Widget Tree + Layout Engine (internal/layout)
        │
        ▼
Character Buffer (double-buffered 2D grid)
        │
        ▼
Terminal (ANSI escape sequences)
```

All public API types live in the root `tui` package. Internal packages (`internal/layout`,
`internal/tuigen`, `internal/formatter`, `internal/lsp`, `internal/debug`) are not importable
by external consumers.

## Where to Look (By Task)

Use this section to quickly find the right files for a given change.

### Changing the public API (Element options, App config, types)

- `element.go` — Element struct definition, TextAlign/ScrollMode enums
- `element_options.go` — Option funcs: WithWidth, WithHeight, WithFlexGrow, WithDirection, WithBorder, etc.
- `element_options_auto.go` — WithWidthAuto(), WithHeightAuto()
- `element_accessors.go` — Getters/setters: SetText, SetBorder, SetStyle, Background, etc.
- `app_options.go` — AppOption funcs: WithFrameRate, WithMouseEnabled, WithInlineHeight, etc.
- `layout.go` — Re-exports from internal/layout (Direction, Justify, Align, Value, etc.)

### Changing layout behavior (flexbox algorithm)

- `internal/layout/calculate.go` — Main layout algorithm, `Calculate()` entry point
- `internal/layout/flex.go` — Core flexbox: `layoutChildren()` 6-phase algorithm, justify/align offsets
- `internal/layout/style.go` — LayoutStyle struct, Direction/Justify/Align constants
- `internal/layout/value.go` — Value type: Auto/Fixed/Percent dimension units
- `internal/layout/edges.go` — Edges struct (TRBL spacing), EdgeAll(), EdgeSymmetric()
- `internal/layout/layoutable.go` — Layoutable interface (LayoutStyle, LayoutChildren, SetLayout, etc.)
- `element_layout.go` — Element's implementation of Layoutable interface

### Changing rendering / terminal output

- `element_render.go` — RenderTree(), renderElement(): border drawing, text rendering, child recursion
- `render.go` — Render(), RenderFull(), RenderRegion(): diff computation, full redraw
- `buffer.go` — Double-buffered character grid: front/back, Diff(), Swap(), SetCell(), Fill()
- `cell.go` — Cell struct: Rune, Style, Width (CJK support)
- `escape.go` — escBuilder: ANSI escape sequence construction (cursor, colors, styles)
- `terminal.go` — Terminal interface definition
- `terminal_ansi.go` — ANSITerminal implementation (ANSI escape output, capabilities)
- `terminal_unix.go` — Unix raw mode via termios syscalls
- `caps.go` — DetectCapabilities(): TERM/COLORTERM env var detection

### Changing event handling / input

- `event.go` — Event interface, KeyEvent, MouseEvent, ResizeEvent types
- `key.go` — Key type: special keys (Escape, Enter, Tab, arrows, Ctrl+A-Z, function keys)
- `parse.go` — parseInput(): CSI/SS3 sequence parsing, mouse SGR, UTF-8, modifiers
- `reader.go` — EventReader/InterruptibleReader interfaces, stdinReader, PollEvent()
- `reader_unix.go` — Unix-specific: getTerminalSizeForReader(), selectWithTimeout()
- `app_events.go` — App.Dispatch(): routes ResizeEvent, MouseEvent, key events via FocusManager

### Changing focus management

- `focus.go` — Focusable interface, FocusManager: Register(), Next(), Prev(), SetFocus(), Dispatch()
- `element_focus.go` — Element focus API: IsFocusable, Focus(), Blur(), HandleEvent(), SetOnEvent()
- `element_watchers.go` — Focus tree discovery: WalkFocusables(), SetOnFocusableAdded()

### Changing reactive state / watchers

- `state.go` — State[T] generic type: Bind(), Set(), Get(), Batch(), dirty marking
- `watcher.go` — Watcher interface, Watch() helper, TimerWatcher, TickerWatcher
- `element_watchers.go` — AddWatcher(), Watchers(), WalkWatchers(), SetOnUpdate()
- `dirty.go` — Global dirty flag: MarkDirty(), checkAndClearDirty()

### Changing the app lifecycle / main loop

- `app.go` — App struct, Renderable interface, tree walkers
- `app_loop.go` — Run() main event loop, frame timing, signal handling
- `app_lifecycle.go` — Close(), Stop(), PrintAbove() for inline mode
- `app_render.go` — App.Render(): buffer management, dirty checking, inline vs full-screen

### Changing the .gsx compiler (code generation)

The compiler pipeline is: **Lexer → Parser → Analyzer → Generator**

- `internal/tuigen/token.go` — TokenType enum (79 types), Token struct, Position
- **Lexer** (`internal/tuigen/`):
  - `lexer.go` — Main Lexer struct, NewLexer(), Next()
  - `lexer_utils.go` — Helper functions (readIdent, readString, isLetter)
  - `lexer_goexpr.go` — Go expression lexing inside `{...}`
  - `lexer_strings.go` — String literal parsing (double-quoted, backtick)
- **Parser** (`internal/tuigen/`):
  - `parser.go` — Parser struct, NewParser(), ParseFile(), error synchronization
  - `parser_component.go` — Component definition and parameter parsing
  - `parser_element.go` — Element parsing: attributes, children, self-closing tags
  - `parser_control.go` — Control flow: @for, @if, @else, @let
  - `parser_expr.go` — Go expression parsing within elements/attributes
- **AST** (`internal/tuigen/`):
  - `ast.go` — 20+ node types: File, Component, Element, Attribute, ForLoop, IfStmt, LetBinding, GoExpr, ComponentCall, ChildrenSlot, GoFunc, GoDecl, etc.
- **Analyzer** (`internal/tuigen/`):
  - `analyzer.go` — Semantic analysis: validates tags, attributes, imports
  - `analyzer_refs.go` — Named reference tracking and validation
  - `analyzer_state.go` — State variable and binding tracking
- **Generator** (`internal/tuigen/`):
  - `generator.go` — Generator struct, NewGenerator(), Generate(), imports processing
  - `generator_element.go` — Element code generation with option building, ref handling
  - `generator_children.go` — Child element generation, text handling
  - `generator_component.go` — Component function generation
  - `generator_control.go` — Control flow code generation (@for, @if, @let)
- **Tailwind** (`internal/tuigen/`):
  - `tailwind.go` — ParseTailwindClasses(): converts class strings to element options
  - `tailwind_validation.go` — Class syntax validation
  - `tailwind_data.go` — Class definitions and mappings
  - `tailwind_autocomplete.go` — Autocomplete suggestions for classes
- `errors.go` — Error struct with position/message/hint, ErrorList

### Changing the formatter

- `internal/formatter/formatter.go` — Formatter struct, New(), Format(), FormatWithResult()
- `internal/formatter/printer.go` — printer struct: generates formatted .gsx from AST
- `internal/formatter/printer_comments.go` — Comment formatting (leading/trailing/orphan)
- `internal/formatter/printer_control.go` — Control structure formatting (@if, @for, @let)
- `internal/formatter/printer_elements.go` — Element/attribute formatting, multi-line layout
- `internal/formatter/imports.go` — Import fixing via goimports integration

### Changing the LSP (language server)

- **Core** (`internal/lsp/`):
  - `server.go` — Server struct, JSON-RPC communication, document management
  - `router.go` — Request dispatch to handlers and providers
  - `handler.go` — LSP lifecycle: initialize, shutdown, exit
  - `document.go` — Document/DocumentManager for open .gsx file tracking
  - `context.go` — CursorContext: resolved AST info at cursor position
  - `context_resolve.go` — Cursor position resolution logic
  - `index.go` — ComponentIndex: workspace-level symbol indexing
  - `providers.go` — Provider interface definitions, Registry struct
  - `provider_adapters.go` — Adapters between lsp and provider package types
- **Feature files** (`internal/lsp/`):
  - `hover.go`, `completion.go`, `definition.go`, `references.go`
  - `diagnostics.go`, `formatting.go`, `semantic_tokens.go`, `symbols.go`
- **Provider implementations** (`internal/lsp/provider/`):
  - `provider.go` — Provider interface definitions, type aliases
  - `completion.go` / `completion_items.go` — Completion provider
  - `hover.go` — Hover documentation provider
  - `definition.go` / `definition_search.go` — Go-to-definition
  - `references.go` / `references_search.go` — Find references
  - `diagnostics.go` — Diagnostic reporting
  - `formatting.go` — Document formatting
  - `semantic.go` / `semantic_gocode.go` / `semantic_nodes.go` — Semantic tokens
  - `symbols.go` — Document and workspace symbols
- **gopls proxy** (`internal/lsp/gopls/`):
  - `proxy.go` — GoplsProxy: subprocess communication over JSON-RPC
  - `proxy_requests.go` — JSON-RPC request types
  - `mapping.go` — SourceMap: position translation between .gsx and generated .go
  - `generate.go` / `generate_state.go` — Generate virtual Go from .gsx AST for gopls
- **Schema** (`internal/lsp/schema/`):
  - `schema.go` — Element definitions (div, span, p, ul, li, button, input, table, progress, hr, br)
  - `keywords.go` — Keyword definitions (templ, @for, @if, @else, @let)
  - `tailwind.go` — Tailwind class documentation and matching
- **Logging** (`internal/lsp/log/`):
  - `log.go` — Logging with component prefixes (Server, Gopls, Generate, Mapping)

### Changing the CLI tool

- `cmd/tui/main.go` — Entry point, version, CLI dispatcher
- `cmd/tui/generate.go` — `tui generate` command
- `cmd/tui/check.go` — `tui check` command
- `cmd/tui/fmt.go` — `tui fmt` command (supports --stdout, --check)
- `cmd/tui/lsp.go` — `tui lsp` command (supports --log FILE)

### Adding or modifying element types / attributes

- `internal/lsp/schema/schema.go` — Element definitions and attribute schemas
- `internal/tuigen/analyzer.go` — Attribute validation rules
- `internal/tuigen/tailwind.go` — Tailwind class → option mapping
- `element_options.go` — New Option funcs for generated code to call

### Writing tests

- `mock_terminal.go` — MockTerminal: captures operations, maintains internal cell buffer
- `mock_reader.go` — MockEventReader: returns queued events in order
- `cmd/tui/testdata/` — .gsx fixtures and their expected generated .go output

## CLI Commands

```bash
tui generate [path...]       # Generate Go code from .gsx files
tui check [path...]          # Check .gsx files without generating
tui fmt [path...]            # Format .gsx files
tui fmt --check [path...]    # Check formatting without modifying
tui fmt --stdout [path...]   # Write formatted output to stdout
tui lsp                      # Start language server (stdio)
tui lsp --log FILE           # Start language server with logging
tui version                  # Print version
tui help                     # Show help
```

## .gsx File Syntax

```gsx
package mypackage

import (
    "fmt"
)

// Component definition (returns Element)
templ Header(title string) {
    <div class="border-single p-1">
        <span class="font-bold">{title}</span>
    </div>
}

// Conditionals
templ Conditional(show bool) {
    <div class="flex-col">
        @if show {
            <span>Visible</span>
        } @else {
            <span>Hidden</span>
        }
    </div>
}

// Loops
templ List(items []string) {
    <div class="flex-col gap-1">
        @for i, item := range items {
            <span>{fmt.Sprintf("%d: %s", i, item)}</span>
        }
    </div>
}

// Local bindings
templ Counter(count int) {
    @let label = fmt.Sprintf("Count: %d", count)
    <span>{label}</span>
}

// Helper functions (regular Go - no Element return type)
func helper(s string) string {
    return fmt.Sprintf("[%s]", s)
}
```

### Built-in Elements (HTML-style)

| Element | Description |
|---------|-------------|
| `<div>` | Block container with flexbox layout |
| `<span>` | Inline text/content container |
| `<p>` | Paragraph text |
| `<ul>` | Unordered list container |
| `<li>` | List item |
| `<button>` | Clickable button |
| `<input>` | Text input field |
| `<table>` | Table container |
| `<progress>` | Progress bar |
| `<hr>` | Horizontal rule (self-closing) |
| `<br>` | Line break (self-closing) |

### Common Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | `string` | Unique identifier |
| `class` | `string` | Tailwind-style classes |
| `disabled` | `bool` | Disable interaction |

### Layout Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `width` | `int` | Fixed width in characters |
| `widthPercent` | `int` | Width as percentage |
| `height` | `int` | Fixed height in rows |
| `heightPercent` | `int` | Height as percentage |
| `minWidth` | `int` | Minimum width |
| `minHeight` | `int` | Minimum height |
| `maxWidth` | `int` | Maximum width |
| `maxHeight` | `int` | Maximum height |
| `direction` | `tui.Direction` | Flex direction |
| `justify` | `tui.Justify` | Main axis alignment |
| `align` | `tui.Align` | Cross axis alignment |
| `gap` | `int` | Gap between children |
| `flexGrow` | `float64` | Flex grow factor |
| `flexShrink` | `float64` | Flex shrink factor |
| `padding` | `int` | Padding on all sides |
| `margin` | `int` | Margin on all sides |

### Visual Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `border` | `tui.BorderStyle` | Border style |
| `borderStyle` | `string` | Border style name |
| `background` | `tui.Color` | Background color |
| `text` | `string` | Text content |
| `textStyle` | `tui.Style` | Text styling |
| `textAlign` | `string` | Text alignment |

### Tailwind-style Classes

Use the `class` attribute for styling:

```gsx
<div class="flex-col gap-2 p-2 border-rounded">
    <span class="font-bold text-cyan">Title</span>
    <span class="font-dim">Subtitle</span>
</div>
```

| Class | Description |
|-------|-------------|
| `flex` | Display flex (row) |
| `flex-col` | Display flex column |
| `gap-N` | Gap of N characters |
| `p-N` | Padding of N |
| `m-N` | Margin of N |
| `border-single` | Single line border |
| `border-double` | Double line border |
| `border-rounded` | Rounded border |
| `border-thick` | Thick border |
| `font-bold` | Bold text |
| `font-dim` | Dim/faint text |
| `font-italic` | Italic text |
| `text-COLOR` | Text color (red, green, blue, cyan, etc.) |
| `bg-COLOR` | Background color |
| `items-center` | Align items center |
| `items-start` | Align items start |
| `items-end` | Align items end |
| `justify-center` | Justify content center |
| `justify-between` | Justify space between |
| `justify-around` | Justify space around |

## Key Types

```go
// tui.Value - dimension specification
tui.Fixed(10)        // 10 characters
tui.Percent(50)      // 50% of available space
tui.Auto()           // Size to content

// tui.BorderStyle
tui.BorderNone
tui.BorderSingle     // ┌─┐│└─┘
tui.BorderDouble     // ╔═╗║╚═╝
tui.BorderRounded    // ╭─╮│╰─╯
tui.BorderThick      // ┏━┓┃┗━┛

// tui.Direction
tui.Row
tui.Column

// tui.Justify
tui.JustifyStart
tui.JustifyCenter
tui.JustifyEnd
tui.JustifySpaceBetween
tui.JustifySpaceAround
tui.JustifySpaceEvenly

// tui.Align
tui.AlignStart
tui.AlignCenter
tui.AlignEnd
tui.AlignStretch

// tui.Style - text styling
tui.Style{}.Bold().Fg(tui.ANSIColor(tui.Red))

// tui.State[T] - reactive state
count := tui.NewState(0)
count.Set(count.Get() + 1)
count.Bind(func(v int) { /* called on change */ })
```

## Layout System

The layout engine (`internal/layout`) implements CSS flexbox with:

- `Row` and `Column` directions
- `JustifyContent`: Start, Center, End, SpaceBetween, SpaceAround, SpaceEvenly
- `AlignItems`: Start, Center, End, Stretch
- Padding, margin, and gap
- Min/max width/height constraints
- Percentage, fixed, and auto values
- Float-precision positioning for jitter-free animation

## Design Patterns

- **Functional Options**: `Element` and `App` use `Option`/`AppOption` funcs for configuration
- **Double Buffering**: `Buffer` maintains front/back grids with diff-based rendering
- **Dirty Flags**: Global dirty flag triggers re-layout/re-render when state changes
- **Reactive State**: `State[T]` with `Bind()` callbacks and `Batch()` for coalescing
- **Interface-based**: `Renderable`, `Layoutable`, `Focusable`, `Watcher`, `Terminal`
- **Tree Walking**: DFS traversal for focus discovery, watcher collection, rendering

## Testing

Use table-driven tests for all unit tests with the following pattern:

```go
type tc struct {
    // test case fields
}

tests := map[string]tc{
    "test name": {
        // test case values
    },
}

for name, tt := range tests {
    t.Run(name, func(t *testing.T) {
        // test logic
    })
}
```

Always define the `tc` struct separately before the test map.

## Running Tests

```bash
go test ./...                        # Run all tests
go test ./internal/tuigen/...        # Run tuigen tests
go test ./internal/lsp/...           # Run LSP tests
go test ./internal/layout/...        # Run layout tests
go test ./internal/formatter/...     # Run formatter tests
go test -run TestParser ./...        # Run specific test
```

## Building

```bash
go build -o tui ./cmd/tui        # Build CLI
./tui generate ./examples/...    # Generate example code
```
