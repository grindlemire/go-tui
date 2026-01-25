# TUI Tooling Specification

**Status:** Planned\
**Version:** 1.0\
**Last Updated:** 2025-01-25

---

## 1. Overview

### Purpose

Build comprehensive developer tooling for `.tui` files that provides a first-class IDE experience. This includes syntax highlighting, language server protocol (LSP) support with Go code intelligence, code formatting, and editor extensions for VS Code and Neovim/Tree-sitter editors.

### Goals

- Provide syntax highlighting for `.tui` files in VS Code and Tree-sitter editors (Neovim, Helix, Zed)
- Build an LSP server (`tui lsp`) with diagnostics, go-to-definition, hover, and auto-completion
- Proxy to gopls for Go expression intelligence inside `{expr}` blocks
- Implement a code formatter (`tui fmt`) for consistent `.tui` file styling
- Package everything in the go-tui monorepo for easy distribution

### Non-Goals

- Support for other editors (Sublime, Atom, Emacs) in this initial version
- Refactoring tools beyond basic rename support
- Debugging support (step through generated Go code)
- Language injection for CSS-like styling (future consideration)

---

## 2. Architecture

### Directory Structure

```
go-tui/
├── cmd/tui/
│   ├── main.go              # CLI entry point
│   ├── lsp.go               # tui lsp subcommand
│   └── fmt.go               # tui fmt subcommand
├── pkg/
│   ├── tuigen/              # (existing) lexer, parser, generator
│   ├── lsp/                 # LSP server implementation
│   │   ├── server.go        # Main LSP server
│   │   ├── handler.go       # LSP method handlers
│   │   ├── document.go      # Document state management
│   │   ├── diagnostics.go   # Error/warning reporting
│   │   ├── completion.go    # Auto-completion
│   │   ├── definition.go    # Go-to-definition
│   │   ├── hover.go         # Hover information
│   │   ├── symbols.go       # Document/workspace symbols
│   │   └── gopls/           # gopls proxy
│   │       ├── proxy.go     # gopls communication
│   │       └── mapping.go   # Position mapping .tui <-> .go
│   └── formatter/           # Code formatter
│       ├── formatter.go     # Main formatter logic
│       └── printer.go       # Pretty printer
├── editor/
│   ├── vscode/              # VS Code extension
│   │   ├── package.json
│   │   ├── syntaxes/
│   │   │   └── tui.tmLanguage.json
│   │   └── language-configuration.json
│   └── tree-sitter-tui/     # Tree-sitter grammar
│       ├── grammar.js
│       ├── package.json
│       ├── src/             # Generated parser
│       └── queries/
│           ├── highlights.scm
│           └── injections.scm
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `cmd/tui/lsp.go` | CLI subcommand to start the LSP server |
| `cmd/tui/fmt.go` | CLI subcommand for code formatting |
| `pkg/lsp/` | LSP server with diagnostics, completion, hover, definition |
| `pkg/lsp/gopls/` | Proxy layer to gopls for Go expression intelligence |
| `pkg/formatter/` | Pretty printer for consistent `.tui` formatting |
| `editor/vscode/` | VS Code extension with TextMate grammar |
| `editor/tree-sitter-tui/` | Tree-sitter grammar for Neovim/Helix/Zed |

### Data Flow Diagram

```
                                    ┌─────────────────────────────────────┐
                                    │           VS Code / Editor          │
                                    └─────────────────────────────────────┘
                                                      │
                                                      │ LSP Protocol (JSON-RPC)
                                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              tui lsp                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Diagnostics │  │ Completion  │  │   Hover     │  │  Go-to-Definition   │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
│         │                │                │                    │            │
│         └────────────────┴────────────────┴────────────────────┘            │
│                                    │                                         │
│                          ┌─────────┴─────────┐                              │
│                          │  Document Manager │                              │
│                          │   (parse .tui)    │                              │
│                          └─────────┬─────────┘                              │
│                                    │                                         │
│         ┌──────────────────────────┼──────────────────────────┐             │
│         │                          │                          │             │
│         ▼                          ▼                          ▼             │
│  ┌─────────────┐          ┌─────────────┐            ┌─────────────────┐   │
│  │   Lexer/    │          │  Component  │            │  gopls Proxy    │   │
│  │   Parser    │          │   Index     │            │  (Go intellisense) │ │
│  │  (tuigen)   │          │             │            └────────┬────────┘   │
│  └─────────────┘          └─────────────┘                     │             │
│                                                               ▼             │
│                                                    ┌─────────────────┐      │
│                                                    │     gopls       │      │
│                                                    └─────────────────┘      │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Core Entities

### LSP Server Types

```go
// Server represents the TUI LSP server.
type Server struct {
    conn     *jsonrpc2.Conn
    docs     *DocumentManager
    index    *ComponentIndex
    gopls    *GoplsProxy
}

// Document represents an open .tui file.
type Document struct {
    URI      string
    Content  string
    Version  int
    AST      *tuigen.File
    Errors   []tuigen.Error
}

// ComponentIndex tracks all components across the workspace.
type ComponentIndex struct {
    // component name -> location
    Components map[string]Location
    // file URI -> components defined in that file
    FileComponents map[string][]string
}

// Location represents a source location.
type Location struct {
    URI   string
    Range Range
}
```

### Formatter Types

```go
// Formatter formats .tui source code.
type Formatter struct {
    IndentWidth int    // default: tab
    MaxLineWidth int   // default: 100
}

// FormatResult contains the formatted output.
type FormatResult struct {
    Content string
    Changed bool
}
```

---

## 4. User Experience

### CLI Commands

```bash
# Start LSP server (typically called by editor)
$ tui lsp
# LSP server starts on stdio, ready for JSON-RPC

# Format a file in-place
$ tui fmt counter.tui
Formatted: counter.tui

# Format and print to stdout (don't modify file)
$ tui fmt --stdout counter.tui
package main
...

# Format all .tui files in directory
$ tui fmt ./...
Formatted: components/header.tui
Formatted: components/footer.tui
Formatted: main.tui

# Check formatting without modifying (CI mode)
$ tui fmt --check ./...
ERROR: components/header.tui is not formatted
```

### VS Code Integration

After installing the extension:

1. **Syntax Highlighting**: Keywords (`@component`, `@for`, `@if`), tags (`<box>`, `<text>`), attributes, Go expressions all colorized
2. **Diagnostics**: Red squiggles for syntax errors, undefined components
3. **Hover**: Hover over `@ComponentName` to see its signature
4. **Go-to-Definition**: Cmd+Click on component call jumps to `@component` definition
5. **Auto-completion**: Type `@` to see component names, type `<` to see element tags
6. **Format on Save**: Automatically formats `.tui` files on save (configurable)

### Neovim Integration (Tree-sitter)

After installing the grammar:

1. **Syntax Highlighting**: Full semantic highlighting via Tree-sitter queries
2. **Go Injection**: Go expressions inside `{}` highlighted as Go code
3. **LSP features**: Via `nvim-lspconfig` pointing to `tui lsp`

---

## 5. Feature Details

### 5.1 TextMate Grammar (VS Code)

The TextMate grammar captures:

| Scope | Pattern | Example |
|-------|---------|---------|
| `keyword.control.tui` | `@component`, `@for`, `@if`, `@else`, `@let` | `@component Header` |
| `entity.name.function.tui` | Component name after `@component` | `@component Header` |
| `entity.name.tag.tui` | Element tags | `<box>`, `</text>` |
| `entity.other.attribute-name.tui` | Attribute names | `border=`, `padding=` |
| `string.quoted.double.tui` | String values | `"hello"` |
| `meta.embedded.expression.go` | Go expressions | `{fmt.Sprintf(...)}` |
| `comment.line.tui` | Line comments | `// comment` |
| `comment.block.tui` | Block comments | `/* comment */` |

### 5.2 Tree-sitter Grammar

The Tree-sitter grammar defines node types:

```javascript
module.exports = grammar({
  name: 'tui',

  rules: {
    source_file: $ => seq(
      $.package_clause,
      optional($.import_declarations),
      repeat($._top_level_declaration)
    ),

    component_declaration: $ => seq(
      '@component',
      $.identifier,
      $.parameter_list,
      $.component_body
    ),

    element: $ => choice(
      $.self_closing_element,
      $.element_with_children
    ),

    go_expression: $ => seq('{', /[^}]+/, '}'),
    // ... more rules
  }
});
```

### 5.3 LSP Capabilities

| Capability | Method | Description |
|------------|--------|-------------|
| Diagnostics | `textDocument/publishDiagnostics` | Parse errors, undefined components |
| Completion | `textDocument/completion` | Components, elements, attributes |
| Hover | `textDocument/hover` | Component signatures, attribute docs |
| Definition | `textDocument/definition` | Jump to component definition |
| References | `textDocument/references` | Find all usages of a component |
| Document Symbols | `textDocument/documentSymbol` | Outline of components in file |
| Workspace Symbols | `workspace/symbol` | Search components across workspace |
| Formatting | `textDocument/formatting` | Format entire document |
| Range Formatting | `textDocument/rangeFormatting` | Format selection |

### 5.4 gopls Proxy

For Go expression intelligence, the LSP:

1. **Generates virtual `.go` file** from the `.tui` file (like templ does)
2. **Maintains position mapping** between `.tui` and virtual `.go` positions
3. **Proxies requests** to gopls for the virtual file
4. **Translates responses** back to `.tui` positions

Example mapping:
```
.tui file:                          virtual .go file:
---------                           ------------------
@component Counter(count int) {     func Counter(count int) *element.Element {
  <text>{fmt.Sprintf("%d", count)}</text>    _ = fmt.Sprintf("%d", count)
}                                   }
```

When user requests completion at `fmt.Spr|` in the `.tui` file:
1. Map position to virtual `.go` position
2. Send `textDocument/completion` to gopls
3. Receive completion items from gopls
4. Map response positions back to `.tui`

### 5.5 Formatter Rules

The formatter enforces:

1. **Indentation**: Tabs for nesting (configurable)
2. **Element formatting**:
   - Self-closing elements on one line if short: `<text>{x}</text>`
   - Multi-line for complex children
3. **Attribute formatting**:
   - Single attribute inline: `<box padding={1}>`
   - Multiple attributes: one per line if exceeds line width
4. **Consistent spacing**:
   - Space after `@for`, `@if`, `@let` keywords
   - No space before `=` in attributes

---

## 6. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Large\
**Recommended Phases:** 6\
**Rationale:**

This is a cross-cutting feature building an entirely new tooling subsystem:
1. VS Code extension (TextMate grammar, extension packaging)
2. Tree-sitter grammar (grammar.js, queries)
3. LSP server foundation (protocol, document management)
4. LSP features (diagnostics, completion, hover, definition)
5. gopls proxy for Go intelligence
6. Code formatter

Each component is substantial and has dependencies. The LSP alone requires understanding the LSP protocol, maintaining document state, indexing components, and proxying to gopls.

> **IMPORTANT:** User must approve the complexity assessment and 6-phase count before proceeding to implementation plan.

---

## 7. Success Criteria

1. **Syntax highlighting works** in VS Code with all DSL constructs properly colored
2. **Tree-sitter grammar** provides highlighting in Neovim/Helix
3. **Diagnostics** show parse errors in real-time as user types
4. **Go-to-definition** jumps to `@component` definitions from calls
5. **Hover** shows component signatures with parameter types
6. **Completion** offers component names, element tags, and attribute names
7. **Go expressions** have full intellisense via gopls proxy
8. **Formatter** produces consistent output that round-trips through parse
9. **Format on save** works in VS Code

---

## 8. Technical Considerations

### 8.1 LSP Library Choice

Two main options for Go LSP servers:

| Library | Pros | Cons |
|---------|------|------|
| `golang.org/x/tools/gopls/internal/lsp/protocol` | Used by gopls, battle-tested | Internal package, API may change |
| `github.com/tliron/glsp` | Public API, documented | Less battle-tested |
| `go.lsp.dev/protocol` | Standard LSP types | Need to implement server layer |

**Recommendation:** Use `go.lsp.dev/protocol` for types and build a thin server layer. This provides stability while keeping the implementation maintainable.

### 8.2 gopls Integration

Two approaches for gopls integration:

1. **Subprocess + Pipe**: Start gopls as subprocess, communicate via JSON-RPC
2. **In-process**: Import gopls packages directly (complex, tied to gopls version)

**Recommendation:** Subprocess approach (how templ does it). More robust and version-independent.

### 8.3 Position Mapping

The `.tui` to virtual `.go` position mapping is critical for gopls proxy:

```go
// SourceMap tracks position mappings between .tui and generated .go
type SourceMap struct {
    // For each line in .go, track corresponding .tui position
    Mappings []Mapping
}

type Mapping struct {
    GoLine, GoCol     int
    TuiLine, TuiCol   int
    Length            int
}
```

---

## 9. Open Questions

1. **Should completion include Go stdlib?** → Yes, via gopls proxy
2. **How to handle multi-file projects?** → Workspace indexing on initialization
3. **Should formatter be configurable?** → Start with sensible defaults, add config later
4. **Support for semantic tokens?** → Nice to have, defer to v2

---

## 10. References

- [templ LSP implementation](https://github.com/a-h/templ/tree/main/cmd/templ/lspcmd)
- [VS Code extension guide](https://code.visualstudio.com/api)
- [Tree-sitter documentation](https://tree-sitter.github.io/tree-sitter/)
- [LSP specification](https://microsoft.github.io/language-server-protocol/)
- [gopls design](https://github.com/golang/tools/tree/master/gopls)
