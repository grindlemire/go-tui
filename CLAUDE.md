# go-tui Project Guidelines

A declarative terminal UI framework for Go with templ-like syntax and flexbox layout.

## Git Commits IMPORTANT

Use `gcommit -m ""` for all commits to ensure proper signing.

ONLY EVER COMMIT USING THIS APPROACH

All commit messages MUST use conventional commit format. This is required for
automated releases via release-please.

Format: `<type>: <description>` or `<type>(<scope>): <description>`

| Prefix | Version Bump | Use When |
|--------|-------------|----------|
| `feat:` | minor (0.1.0 → 0.2.0) | Adding new functionality |
| `fix:` | patch (0.1.0 → 0.1.1) | Fixing a bug |
| `perf:` | patch | Performance improvements |
| `refactor:` | patch | Code changes that don't add features or fix bugs |
| `docs:` | patch | Documentation only |
| `test:` | patch | Adding or updating tests |
| `chore:` | patch | Maintenance, dependencies, tooling |
| `ci:` | patch | CI/CD changes |
| `build:` | patch | Build system changes |
| `revert:` | patch | Reverting a previous commit |

For BREAKING CHANGES (major bump, e.g. 0.1.0 → 1.0.0), add `!` after the type:
`feat!: remove deprecated API` or include `BREAKING CHANGE:` in the commit body.

Examples:
```
gcommit -m "feat: add table element support"
gcommit -m "fix(layout): correct flexbox gap calculation"
gcommit -m "feat!: change State API to require type parameter"
gcommit -m "chore: update golang.org/x dependencies"
```

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

- `element.go` — Element struct definition, TextAlign/ScrollMode/OverflowMode enums
- `element_options.go` — Option funcs: WithWidth, WithHeight, WithFlexGrow, WithDirection, WithBorder, WithScrollable, WithTruncate, WithHidden, WithOverflow, WithTextGradient, WithBackgroundGradient, WithBorderGradient, etc.
- `element_options_auto.go` — WithWidthAuto(), WithHeightAuto()
- `element_accessors.go` — Getters/setters: SetText, SetBorder, SetStyle, Background, etc.
- `element_tree.go` — Tree manipulation: AddChild, RemoveChild, RemoveAllChildren
- `element_scroll.go` — Scroll methods: ScrollTo, ScrollOffset, MaxScroll, ViewportSize
- `app_options.go` — AppOption funcs: WithFrameRate, WithMouseEnabled, WithInlineHeight, WithGlobalKeyHandler, WithInputLatency, WithEventQueueSize, etc.
- `layout.go` — Re-exports from internal/layout (Direction, Justify, Align, Value, etc.)
- `color.go` — Color type, ANSIColor(), RGBColor(), HexColor(), Gradient, GradientDirection
- `style.go` — Style type with chainable methods (Bold, Dim, Italic, Foreground, etc.)
- `ref.go` — Ref, RefList, RefMap[K] for element references
- `click.go` — Click(), HandleClicks() for ref-based mouse hit testing
- `keymap.go` — KeyMap, KeyBinding, KeyPattern, OnKey(), OnRune(), etc.
- `component.go` — Component, KeyListener, MouseListener, Initializer, WatcherProvider interfaces
- `events.go` — Events[T] generic event bus for cross-component communication
- `mount.go` — Component caching and lifecycle (Mount, PropsUpdater)

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
- `render_element.go` — renderElementToBuffer, bufferRowToANSI: standalone element-to-ANSI rendering
- `escape.go` — escBuilder: ANSI escape sequence construction (cursor, colors, styles)
- `terminal.go` — Terminal interface definition
- `terminal_ansi.go` — ANSITerminal implementation (ANSI escape output, capabilities)
- `terminal_unix.go` — Unix raw mode via termios syscalls
- `terminal_windows.go` — Windows raw mode
- `caps.go` — DetectCapabilities(): TERM/COLORTERM env var detection
- `border.go` — BorderStyle type and BorderChars definitions
- `inline_ansi.go` — ANSI escape handling for inline mode
- `inline_session.go` — Inline session management
- `inline_wrap.go` — Inline content wrapping

### Changing event handling / input

- `event.go` — Event interface, KeyEvent, MouseEvent (MouseButton, MouseAction), ResizeEvent types
- `key.go` — Key type: special keys (Escape, Enter, Tab, arrows, Ctrl+A-Z, function keys)
- `keymap.go` — KeyMap, KeyBinding, KeyPattern; helpers: OnKey(), OnKeyStop(), OnRune(), OnRuneStop(), OnRunes(), OnRunesStop()
- `dispatch.go` — Key dispatch table: priority-ordered key binding dispatch by tree position
- `parse.go` — parseInput(): CSI/SS3 sequence parsing, mouse SGR, UTF-8, modifiers
- `reader.go` — EventReader/InterruptibleReader interfaces, stdinReader, PollEvent()
- `reader_types.go` — EventReader and InterruptibleReader interface definitions
- `reader_unix.go` — Unix-specific: getTerminalSizeForReader(), selectWithTimeout()
- `reader_windows.go` — Windows-specific terminal input
- `click.go` — ClickBinding, Click(), HandleClicks() for ref-based mouse hit testing
- `app_events.go` — App.Dispatch(): routes ResizeEvent, MouseEvent, key events via FocusManager

### Changing focus management

- `focus.go` — Focusable interface, FocusManager: Register(), Next(), Prev(), SetFocus(), Dispatch()
- `focus_group.go` — FocusGroup for Tab/Shift+Tab cycling between components
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
- `app_lifecycle.go` — Close(), Stop(), PrintAbove(), PrintAboveElement(), StreamAbove() for inline mode
- `app_render.go` — App.Render(): buffer management, dirty checking, inline vs full-screen
- `app_screen.go` — Alternate screen mode API (EnterAlternateScreen, ExitAlternateScreen)
- `app_inline_startup.go` — Inline screen mode initialization and startup policy
- `mount.go` — Component caching with mark-and-sweep lifecycle (Mount, PropsUpdater)

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

// Component calls
templ App() {
    <div class="flex-col">
        @Header("Hello")
        @Counter(0)
    </div>
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
| `<textarea>` | Multi-line text input (self-closing, mounted as Component) |
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
| `ref` | expression | Bind element to a `tui.Ref`/`RefList`/`RefMap` variable |
| `deps` | expression | Explicit state dependencies for reactive bindings |

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
| `alignSelf` | `tui.Align` | Override parent's align for this item |
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

### Event & Focus Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `focusable` | `bool` | Whether the element can receive focus |
| `onFocus` | `func()` | Called when the element gains focus |
| `onBlur` | `func()` | Called when the element loses focus |

### Scroll Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `scrollable` | `bool` | Enable scrolling for overflow content |
| `scrollbarStyle` | `tui.Style` | Style for the scrollbar track |
| `scrollbarThumbStyle` | `tui.Style` | Style for the scrollbar thumb |

### Input-specific Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `value` | `string` | Current input value |
| `placeholder` | `string` | Placeholder text when empty |

### Progress-specific Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `value` | `int` | Current progress value (0 to max) |
| `max` | `int` | Maximum progress value |

### Textarea-specific Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `placeholder` | `string` | Placeholder text when empty |
| `width` | `int` | Width in characters (default 40) |
| `maxHeight` | `int` | Maximum height in rows (0 = unlimited) |
| `border` | `tui.BorderStyle` | Border style |
| `textStyle` | `tui.Style` | Text styling |
| `placeholderStyle` | `tui.Style` | Placeholder styling (default: dim) |
| `cursor` | `rune` | Cursor character (default '▌') |
| `submitKey` | `tui.Key` | Submit trigger key (default KeyEnter) |
| `onSubmit` | `func(string)` | Called when submit key is pressed |

### Tailwind-style Classes

Use the `class` attribute for styling:

```gsx
<div class="flex-col gap-2 p-2 border-rounded">
    <span class="font-bold text-cyan">Title</span>
    <span class="font-dim">Subtitle</span>
</div>
```

**Layout Direction**

| Class | Description |
|-------|-------------|
| `flex` / `flex-row` | Display flex row |
| `flex-col` | Display flex column |

**Flex Sizing**

| Class | Description |
|-------|-------------|
| `grow` | Allow element to grow (flex-grow: 1) |
| `grow-0` | Prevent element from growing |
| `shrink` | Allow element to shrink (flex-shrink: 1) |
| `shrink-0` | Prevent element from shrinking |
| `flex-1` | Grow and shrink (takes available space) |
| `flex-auto` | Grow and shrink (respects content size) |
| `flex-initial` | Don't grow, can shrink |
| `flex-none` | Fixed size (no grow or shrink) |
| `flex-grow-N` | Flex grow factor of N |
| `flex-shrink-N` | Flex shrink factor of N |

**Width & Height**

| Class | Description |
|-------|-------------|
| `w-N` | Fixed width of N characters |
| `h-N` | Fixed height of N rows |
| `w-full` / `h-full` | Full width/height (100%) |
| `w-auto` / `h-auto` | Auto size to content |
| `w-N/M` / `h-N/M` | Fractional size (e.g., `w-1/2` for 50%) |
| `min-w-N` | Minimum width of N characters |
| `max-w-N` | Maximum width of N characters |
| `min-h-N` | Minimum height of N rows |
| `max-h-N` | Maximum height of N rows |

**Justify & Align**

| Class | Description |
|-------|-------------|
| `justify-start` | Justify content to start |
| `justify-center` | Justify content to center |
| `justify-end` | Justify content to end |
| `justify-between` | Justify space between |
| `justify-around` | Justify space around |
| `justify-evenly` | Justify space evenly |
| `items-start` | Align items to start |
| `items-center` | Align items to center |
| `items-end` | Align items to end |
| `items-stretch` | Stretch items to fill |
| `self-start` | Align self to start |
| `self-center` | Align self to center |
| `self-end` | Align self to end |
| `self-stretch` | Stretch self to fill |

**Spacing**

| Class | Description |
|-------|-------------|
| `gap-N` | Gap of N characters between children |
| `p-N` | Padding of N on all sides |
| `px-N` / `py-N` | Horizontal / vertical padding |
| `pt-N` / `pr-N` / `pb-N` / `pl-N` | Individual side padding |
| `m-N` | Margin of N on all sides |
| `mx-N` / `my-N` | Horizontal / vertical margin |
| `mt-N` / `mr-N` / `mb-N` / `ml-N` | Individual side margin |

**Borders**

| Class | Description |
|-------|-------------|
| `border` / `border-single` | Single line border |
| `border-double` | Double line border |
| `border-rounded` | Rounded border |
| `border-thick` | Thick border |
| `border-COLOR` | Border color (red, green, blue, cyan, etc.) |
| `border-[#hex]` | Border color from hex (e.g., `border-[#ff6600]`) |
| `border-gradient-C1-C2[-dir]` | Border gradient (directions: h, v, dd, du) |

**Text Styling**

| Class | Description |
|-------|-------------|
| `font-bold` | Bold text |
| `font-dim` / `text-dim` | Dim/faint text |
| `italic` | Italic text |
| `underline` | Underlined text |
| `strikethrough` | Strikethrough text |
| `blink` | Blinking text |
| `reverse` | Reverse video text |
| `text-left` / `text-center` / `text-right` | Text alignment |
| `truncate` | Truncate text with ellipsis on overflow |

**Colors**

| Class | Description |
|-------|-------------|
| `text-COLOR` | Text color (red, green, blue, cyan, magenta, yellow, white, black) |
| `text-bright-COLOR` | Bright text color variant |
| `text-[#hex]` | Text color from hex (e.g., `text-[#ff0000]`) |
| `bg-COLOR` | Background color |
| `bg-bright-COLOR` | Bright background color variant |
| `bg-[#hex]` | Background color from hex |

**Gradients**

| Class | Description |
|-------|-------------|
| `text-gradient-C1-C2[-dir]` | Text gradient between two colors |
| `bg-gradient-C1-C2[-dir]` | Background gradient between two colors |
| `border-gradient-C1-C2[-dir]` | Border gradient between two colors |

Gradient directions: `-h` (horizontal, default), `-v` (vertical), `-dd` (diagonal down), `-du` (diagonal up). Colors: any named color or bright- variant.

**Scroll & Overflow**

| Class | Description |
|-------|-------------|
| `overflow-scroll` | Enable scrolling in both directions |
| `overflow-y-scroll` | Enable vertical scrolling |
| `overflow-x-scroll` | Enable horizontal scrolling |
| `overflow-hidden` | Clip children without scrollbars |
| `scrollbar-COLOR` | Scrollbar track color |
| `scrollbar-thumb-COLOR` | Scrollbar thumb color |
| `scrollbar-[#hex]` / `scrollbar-thumb-[#hex]` | Hex scrollbar colors |

**Other**

| Class | Description |
|-------|-------------|
| `focusable` | Make element focusable |
| `hidden` | Hide element from layout and rendering |

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

// tui.TextAlign
tui.TextAlignLeft    // default
tui.TextAlignCenter
tui.TextAlignRight

// tui.ScrollMode
tui.ScrollNone       // default
tui.ScrollVertical
tui.ScrollHorizontal
tui.ScrollBoth

// tui.OverflowMode
tui.OverflowVisible  // default
tui.OverflowHidden

// tui.Color - terminal colors
tui.Black, tui.Red, tui.Green, tui.Yellow, tui.Blue, tui.Magenta, tui.Cyan, tui.White
tui.BrightBlack, tui.BrightRed, tui.BrightGreen, tui.BrightYellow, tui.BrightBlue, tui.BrightMagenta, tui.BrightCyan, tui.BrightWhite
tui.ANSIColor(index)        // ANSI 256 palette
tui.RGBColor(r, g, b)       // 24-bit true color
tui.HexColor("#ff6600")     // Hex string to Color
tui.DefaultColor()           // Terminal default

// tui.Gradient - color gradients
tui.NewGradient(tui.Red, tui.Blue)                                    // Horizontal gradient
tui.NewGradient(tui.Red, tui.Blue).WithDirection(tui.GradientVertical) // Vertical
// Directions: GradientHorizontal, GradientVertical, GradientDiagonalDown, GradientDiagonalUp

// tui.Style - text styling (chainable)
tui.NewStyle().Bold().Foreground(tui.Red)
tui.NewStyle().Dim().Italic().Underline().Background(tui.Blue)
// Methods: Bold(), Dim(), Italic(), Underline(), Blink(), Reverse(), Strikethrough(), Foreground(Color), Background(Color)

// tui.State[T] - reactive state
count := tui.NewState(0)
count.Set(count.Get() + 1)
count.Bind(func(v int) { /* called on change */ })

// tui.Ref - element references for hit testing
ref := tui.NewRef()           // Single element ref
list := tui.NewRefList()      // Multiple elements from loops
m := tui.NewRefMap[string]()  // Keyed element refs

// tui.Events[T] - cross-component event bus
bus := tui.NewEvents[MyEvent]("topic-name")
bus.Emit(MyEvent{...})
unsub := bus.Subscribe(func(e MyEvent) { ... })
```

## Component Interfaces

Components are structs implementing `Component` (requires `Render(app *App) *Element`).
Additional optional interfaces add capabilities:

| Interface | Method | Description |
|-----------|--------|-------------|
| `Component` | `Render(app *App) *Element` | Required. Returns the element tree. |
| `KeyListener` | `KeyMap() KeyMap` | Keyboard input handling via key bindings |
| `MouseListener` | `HandleMouse(MouseEvent) bool` | Mouse input handling |
| `Initializer` | `Init() func()` | Setup on mount; returned func is cleanup on unmount |
| `WatcherProvider` | `Watchers() []Watcher` | Timers, tickers, channel watchers |
| `PropsUpdater` | `UpdateProps(fresh Component)` | Receive updated props when re-rendered from cache |
| `AppBinder` | `BindApp(app *App)` | Auto-called by mount system for State/Events fields |
| `AppUnbinder` | `UnbindApp()` | Detach app-bound resources on unmount |

## Event Handling

### Key Bindings (KeyMap)

Implement `KeyListener` to handle keyboard input:

```go
func (c *myComponent) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEnter, c.onEnter),         // Specific key, broadcast
        tui.OnKeyStop(tui.KeyEscape, c.onEscape),   // Specific key, stop propagation
        tui.OnRune('q', func(ke tui.KeyEvent) { ... }), // Specific character
        tui.OnRunesStop(c.onTyping),                  // All printable chars, exclusive
    }
}
```

### Mouse Click Handling

Implement `MouseListener` with ref-based hit testing:

```go
func (c *counter) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.incrementBtn, c.increment),
        tui.Click(c.decrementBtn, c.decrement),
    )
}
```

### Component Watchers

Implement `WatcherProvider` for component-level timers/channels:

```go
func (t *timer) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, t.tick),
        tui.NewChannelWatcher(t.dataChan, t.onData),
    }
}
```

### Cross-Component Communication (Events)

Use `Events[T]` for pub/sub between components:

```go
// In component struct
type myComponent struct {
    notifications *tui.Events[Notification]  // created with tui.NewEvents[Notification]("notifications")
}

// Emit from one component
c.notifications.Emit(Notification{Message: "hello"})

// Subscribe in another component
unsub := c.notifications.Subscribe(func(n Notification) { ... })
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
- **Tree Walking**: BFS traversal for component dispatch (key/mouse events, watcher collection); DFS for focus discovery and rendering
- **Component Caching**: Mount system with mark-and-sweep cleanup — components are cached by (parent, index) key and reused across renders
- **Ref System**: Type-safe element references (`Ref`, `RefList`, `RefMap[K]`) for event handling and hit testing
- **Event Bus**: Generic `Events[T]` for topic-based pub/sub between components
- **Key Dispatch**: Priority-ordered key binding dispatch by tree position via `KeyListener`/`KeyMap`

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
