# go-tui Project Guidelines

A declarative terminal UI framework for Go with templ-like syntax and flexbox layout.

## Git Commits IMPORTANT

Use `gcommit -m ""` for all commits to ensure proper signing.

ONLY EVER COMMIT USING THIS APPROACH

## Project Overview

go-tui allows defining UIs in `.gsx` files that compile to type-safe Go code. The framework provides:

- Declarative component syntax (similar to templ/JSX)
- Pure Go flexbox layout engine (no CGO)
- Minimal external dependencies
- Composable widget system

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
Character Buffer (2D grid)
        │
        ▼
Terminal (ANSI escape sequences)
```

All public API types live in the root `tui` package. Internal packages (`internal/layout`,
`internal/tuigen`, `internal/formatter`, `internal/lsp`, `internal/debug`) are not importable
by external consumers.

## Directory Structure

```
go-tui/                          # Root package "tui" — single public API
├── doc.go                       # Package documentation
├── layout.go                    # Re-exports from internal/layout
├── app.go                       # Application loop
├── app_options.go               # App option functions
├── app_lifecycle.go             # Close, PrintAbove
├── app_events.go                # Dispatch, event handling
├── app_render.go                # Render methods
├── app_loop.go                  # Run, Stop, QueueUpdate
├── element.go                   # Element struct, New()
├── element_options.go           # Element option functions
├── element_layout.go            # Layoutable interface impl
├── element_tree.go              # Child management
├── element_accessors.go         # Getters/setters
├── element_focus.go             # Focus handling
├── element_render.go            # Element rendering
├── element_scroll.go            # Scroll support
├── buffer.go                    # Character buffer
├── border.go                    # Border styles
├── cell.go                      # Cell type
├── color.go                     # Color definitions
├── escape.go                    # ANSI escape sequences
├── event.go                     # Event types
├── focus.go                     # Focus management
├── key.go                       # Key parsing
├── parse.go                     # Input parsing
├── reader.go                    # Event reading
├── render.go                    # Tree rendering
├── state.go                     # Reactive state
├── style.go                     # Styling
├── terminal.go                  # Terminal interface
├── watcher.go                   # State watchers
├── generate.go                  # go:generate directive
│
├── cmd/tui/                     # CLI tool
│   ├── main.go                  # Entry point
│   ├── generate.go              # tui generate command
│   ├── check.go                 # tui check command
│   ├── fmt.go                   # tui fmt command
│   └── lsp.go                   # tui lsp command
│
├── internal/
│   ├── layout/                  # Flexbox layout engine
│   │   ├── calculate.go         # Layout algorithm
│   │   ├── layoutable.go        # Layoutable interface
│   │   ├── style.go             # Layout style types
│   │   ├── value.go             # Dimension values
│   │   └── flex.go              # Flexbox implementation
│   ├── tuigen/                  # DSL compiler (.gsx → Go)
│   │   ├── lexer.go             # Tokenizer
│   │   ├── parser.go            # Parser
│   │   ├── ast.go               # AST types
│   │   ├── analyzer.go          # Semantic analysis
│   │   ├── generator.go         # Go code generator
│   │   └── tailwind.go          # Tailwind-style classes
│   ├── formatter/               # Code formatter
│   │   ├── formatter.go         # Formatting logic
│   │   └── printer.go           # Pretty printer
│   ├── lsp/                     # Language server
│   │   ├── server.go            # LSP server
│   │   ├── document.go          # Document management
│   │   ├── context.go           # Cursor context
│   │   ├── index.go             # Symbol indexing
│   │   ├── provider/            # LSP feature providers
│   │   ├── gopls/               # gopls integration
│   │   └── schema/              # Element/attribute schema
│   └── debug/                   # Debug logging
│
├── editor/
│   ├── tree-sitter-gsx/         # Tree-sitter grammar
│   └── vscode/                  # VSCode extension
└── examples/                    # Example applications
```

## CLI Commands

```bash
tui generate ./...       # Generate Go code from .gsx files
tui check ./...          # Check .gsx files without generating
tui fmt ./...            # Format .gsx files
tui fmt --check ./...    # Check formatting without modifying
tui lsp                  # Start language server (stdio)
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
go test -run TestParser ./...        # Run specific test
```

## Building

```bash
go build -o tui ./cmd/tui        # Build CLI
./tui generate ./examples/...    # Generate example code
```

## Layout System

The layout engine implements CSS flexbox with:

- `Row` and `Column` directions
- `JustifyContent`: Start, Center, End, SpaceBetween, SpaceAround
- `AlignItems`: Start, Center, End, Stretch
- Padding, margin, and gap
- Min/max width/height constraints
- Percentage and fixed values

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
```

## Editor Support

### VSCode

The `editor/vscode/` directory contains a VSCode extension providing:

- Syntax highlighting
- Basic language support

### Tree-sitter

The `editor/tree-sitter-gsx/` directory contains a tree-sitter grammar for:

- Accurate syntax parsing
- Integration with editors supporting tree-sitter

### LSP

The `tui lsp` command starts a Language Server providing:

- Real-time diagnostics
- Error reporting as you type
