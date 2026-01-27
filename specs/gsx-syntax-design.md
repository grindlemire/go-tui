# GSX Syntax Specification

**Status:** Draft\
**Version:** 1.0\
**Last Updated:** 2026-01-27

---

## 1. Overview

### Purpose

Rename `.tui` files to `.gsx` (Go Syntax eXtension) and replace the `@component` keyword with standard Go function syntax. Components are differentiated from helper functions by their return type (`Element`) rather than a special keyword.

### Goals

- **Go-native syntax**: Components use `func Name(params) Element { ... }` instead of `@component`
- **File extension change**: `.tui` → `.gsx` for clearer identity
- **Return type differentiation**: `Element` return type = component, any other return type = Go function
- **Cleaner mental model**: "A component is just a function that returns an Element"
- **JSX similarity**: More familiar to developers coming from React/JSX
### Non-Goals

- Changing the internal element tree structure
- Modifying the layout or styling system
- Adding new control flow constructs
- Changing how imports work
- Supporting multiple return types for components

---

## 2. Syntax Comparison

### Current Syntax (.tui)

```tui
package mypackage

import "fmt"

@component Header(title string) {
    <div class="border-single p-1">
        <span class="font-bold">{title}</span>
    </div>
}

@component Counter(count int) {
    @let label = fmt.Sprintf("Count: %d", count)
    <span>{label}</span>
}

func helper(s string) string {
    return fmt.Sprintf("[%s]", s)
}

func sideEffect() {
    fmt.Println("log")
}
```

### New Syntax (.gsx)

```gsx
package mypackage

import "fmt"

func Header(title string) Element {
    <div class="border-single p-1">
        <span class="font-bold">{title}</span>
    </div>
}

func Counter(count int) Element {
    @let label = fmt.Sprintf("Count: %d", count)
    <span>{label}</span>
}

func helper(s string) string {
    return fmt.Sprintf("[%s]", s)
}

func sideEffect() {
    fmt.Println("log")
}
```

### Differentiation Rules

| Function Signature | Classification | Body Parsing |
|--------------------|----------------|--------------|
| `func Name(params) Element { ... }` | Component | XML/DSL |
| `func Name(params) Type { ... }` | Go function | Raw Go |
| `func Name(params) { ... }` | Go function | Raw Go |

---

## 3. Architecture

### Directory Structure (Files Modified)

```
pkg/tuigen/
├── token.go          # Remove TokenComponent, add TokenElement
├── lexer.go          # Update keyword detection
├── parser.go         # Parse func with Element return as component
├── ast.go            # Update Component to use func-style representation
├── analyzer.go       # Detect Element return type for component classification
├── generator.go      # No changes to output (still generates same Go code)

cmd/tui/
├── generate.go       # Change file pattern from *.tui to *.gsx
├── check.go          # Change file pattern from *.tui to *.gsx
├── fmt.go            # Change file pattern from *.tui to *.gsx
├── lsp.go            # Register .gsx file type

pkg/lsp/
├── server.go         # Update file extension handling
├── document.go       # Update file extension detection

pkg/formatter/
├── formatter.go      # Update to emit func syntax instead of @component

editor/
├── tree-sitter-tui/  # Rename to tree-sitter-gsx, update grammar
└── vscode/           # Update file associations and syntax highlighting
```

### Compilation Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  .gsx Source File                                               │
│  package components                                             │
│  func Counter(count int) Element { <span>{count}</span> }       │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Lexer
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Token Stream                                                   │
│  [PACKAGE, IDENT("components"), FUNC, IDENT("Counter"),         │
│   LPAREN, ..., IDENT("Element"), LBRACE, ...]                   │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Parser (detects Element return type)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Abstract Syntax Tree                                           │
│  File{                                                          │
│    Package: "components",                                       │
│    Components: [Component{Name: "Counter", ...}]                │
│  }                                                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Generator (unchanged)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Generated Go Source (unchanged output)                         │
│  func Counter(count int) *element.Element { ... }               │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Entities

### 4.1 Token Changes

```go
// pkg/tuigen/token.go

// REMOVE:
// TokenAtComponent   // @component

// The TokenFunc already exists and will be used for both components and helper functions.
// No new tokens needed - we differentiate by parsing the return type.

// Also remove from tokenNames map:
// TokenAtComponent: "@component",
```

### 4.2 Lexer Changes

Remove the `@component` keyword handling from the lexer's `scanAtKeyword` function:

```go
// pkg/tuigen/lexer.go

// In scanAtKeyword(), REMOVE this case:
// case "component":
//     return l.makeToken(TokenAtComponent, "@component")

// The remaining @-prefixed keywords (@let, @for, @if, @else, @ComponentCall) are unchanged.
```

### 4.3 AST Changes

The `Component` struct remains largely the same, but is now parsed from `func` syntax:

```go
// pkg/tuigen/ast.go

// Component represents a func Name(params) Element { body } definition
type Component struct {
    Name            string
    Params          []*Param
    ReturnType      string   // "Element" for components, determines parsing mode
    Body            []Node   // Elements, GoCode, ControlFlow, etc.
    AcceptsChildren bool     // true if body contains {children...}
    Position        Position
}
```

### 4.4 Parser Changes

The parser must:

1. Parse `func` keyword
2. Parse function name
3. Parse parameters
4. Parse return type (if present)
5. **Decision point**: If return type is `Element`, parse body as component DSL; otherwise, capture as raw Go

```go
// pkg/tuigen/parser.go

func (p *Parser) parseTopLevel() Node {
    switch p.cur.Type {
    case TokenFunc:
        return p.parseFuncOrComponent()
    // ... other cases
    }
}

func (p *Parser) parseFuncOrComponent() Node {
    p.expect(TokenFunc)
    name := p.expect(TokenIdent).Literal
    params := p.parseParams()

    // Check for return type
    returnType := ""
    if p.cur.Type == TokenIdent {
        returnType = p.cur.Literal
        p.advance()
    }

    // Decision: Element return type means component
    if returnType == "Element" {
        return p.parseComponentBody(name, params)
    }

    // Otherwise, capture as raw Go function
    return p.parseGoFunc(name, params, returnType)
}

func (p *Parser) parseComponentBody(name string, params []*Param) *Component {
    p.expect(TokenLBrace)

    body := []Node{}
    for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
        body = append(body, p.parseBodyNode())
    }

    p.expect(TokenRBrace)

    return &Component{
        Name:       name,
        Params:     params,
        ReturnType: "Element",
        Body:       body,
    }
}
```

### 4.5 Formatter Changes

The formatter emits `func Name(params) Element` instead of `@component Name(params)`:

```go
// pkg/formatter/printer.go

func (p *Printer) printComponent(comp *Component) {
    // OLD: p.printf("@component %s(", comp.Name)
    // NEW:
    p.printf("func %s(", comp.Name)
    p.printParams(comp.Params)
    p.print(") Element {\n")

    p.indent++
    for _, node := range comp.Body {
        p.printNode(node)
    }
    p.indent--

    p.print("}\n")
}
```

### 4.6 CLI Changes

Update file discovery patterns:

```go
// cmd/tui/generate.go

func findSourceFiles(patterns []string) ([]string, error) {
    // OLD: match "*.tui"
    // NEW: match "*.gsx"
    for _, pattern := range patterns {
        if pattern == "./..." {
            // Recursively find all .gsx files
            filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
                if strings.HasSuffix(path, ".gsx") {
                    files = append(files, path)
                }
                return nil
            })
        }
    }
}
```

### 4.7 Output Naming

| Input | Output |
|-------|--------|
| `header.gsx` | `header_gsx.go` |
| `components.gsx` | `components_gsx.go` |
| `my-app.gsx` | `my_app_gsx.go` |

---

## 5. DSL Syntax Reference

### 5.1 File Structure

```gsx
// package declaration (required, first line)
package myapp

// imports (optional, standard Go syntax)
import (
    "fmt"
    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
)

// Component: func with Element return type
func Header(title string) Element {
    <div class="border-single">
        <span>{title}</span>
    </div>
}

// Helper function: func with Go return type
func formatTitle(s string) string {
    return fmt.Sprintf("[%s]", s)
}

// Side-effect function: func with no return type
func logDebug() {
    fmt.Println("debug")
}
```

### 5.2 Component Definition

```gsx
// No parameters
func Header() Element {
    <div class="border-single p-1">
        <span>My App</span>
    </div>
}

// With parameters
func Button(label string, onClick func()) Element {
    <button onClick={onClick}>
        <span>{label}</span>
    </button>
}

// With children slot
func Card(title string) Element {
    <div class="border-rounded p-1">
        <span class="font-bold">{title}</span>
        {children...}
    </div>
}
```

### 5.3 Control Flow (Unchanged)

```gsx
func List(items []string) Element {
    <div class="flex-col">
        @for i, item := range items {
            <span>{fmt.Sprintf("%d: %s", i, item)}</span>
        }
    </div>
}

func Conditional(show bool) Element {
    <div>
        @if show {
            <span>Visible</span>
        } @else {
            <span>Hidden</span>
        }
    </div>
}
```

### 5.4 Local Bindings (Unchanged)

```gsx
func Counter(count int) Element {
    @let label = fmt.Sprintf("Count: %d", count)
    <span>{label}</span>
}

func Interactive() Element {
    @let display = <span>Initial</span>

    <div class="flex-col">
        {display}
        <button onClick={func() { display.SetText("Updated") }}>
            Update
        </button>
    </div>
}
```

### 5.5 Component Calls (Unchanged)

```gsx
func App() Element {
    <div class="flex-col">
        @Header("Welcome")

        @Card("My Card") {
            <span>Card content</span>
        }
    </div>
}
```

---

## 6. Edge Cases

### 6.1 Ambiguous Return Types

| Scenario | Resolution |
|----------|------------|
| `func Foo() Element` | Component - body parsed as DSL |
| `func Foo() *Element` | Go function - pointer type is not `Element` |
| `func Foo() element.Element` | Go function - qualified type is not bare `Element` |
| `func Foo() tui.Element` | Go function - qualified type is not bare `Element` |
| `func Foo() (Element, error)` | Go function - multiple return values |

The rule is simple: **exactly `Element` as the sole return type** means component.

### 6.2 Element Type Aliasing

If users want to use a different name, they must use the bare `Element` identifier:

```gsx
// This works - bare Element identifier
func Header() Element {
    <div>...</div>
}

// This is a Go function - will be captured as raw Go
func Header() MyElement {
    return element.New(...)  // User must write Go
}
```

### 6.3 Capitalization Convention

Following Go conventions:
- **Uppercase** function names = exported components
- **Lowercase** function names = unexported (package-private) components

```gsx
// Exported component
func Header() Element {
    <div>...</div>
}

// Unexported component
func internalHelper() Element {
    <div>...</div>
}
```

---

## 7. LSP Changes

### 7.1 File Extension Registration

```go
// pkg/lsp/server.go

func (s *Server) Initialize(params *InitializeParams) (*InitializeResult, error) {
    return &InitializeResult{
        Capabilities: ServerCapabilities{
            // ...
            TextDocumentSync: &TextDocumentSyncOptions{
                OpenClose: true,
                Change:    TextDocumentSyncKindFull,
            },
        },
    }, nil
}

// Document filtering
func isGSXFile(uri string) bool {
    return strings.HasSuffix(uri, ".gsx")
}
```

### 7.2 Semantic Token Updates

The `@component` decorator token type is removed. Functions returning `Element` will be highlighted as function declarations.

---

## 8. Tree-sitter Grammar Changes

### 8.1 Directory Rename

```
editor/tree-sitter-tui/  →  editor/tree-sitter-gsx/
```

### 8.2 Grammar Updates

```javascript
// grammar.js

module.exports = grammar({
    name: 'gsx',

    rules: {
        source_file: $ => seq(
            $.package_clause,
            repeat($.import_declaration),
            repeat(choice($.component_definition, $.function_definition)),
        ),

        // Component: func Name(params) Element { body }
        component_definition: $ => seq(
            'func',
            $.identifier,
            $.parameter_list,
            'Element',
            $.component_body,
        ),

        // Go function: func name(params) [type] { body }
        function_definition: $ => seq(
            'func',
            $.identifier,
            $.parameter_list,
            optional($.type),
            $.go_block,
        ),

        // ... rest of grammar unchanged
    }
});
```

---

## 9. VSCode Extension Changes

### 9.1 Package.json Updates

```json
{
    "name": "gsx",
    "displayName": "GSX - Go Syntax Extension",
    "contributes": {
        "languages": [{
            "id": "gsx",
            "aliases": ["GSX", "gsx"],
            "extensions": [".gsx"],
            "configuration": "./language-configuration.json"
        }],
        "grammars": [{
            "language": "gsx",
            "scopeName": "source.gsx",
            "path": "./syntaxes/gsx.tmLanguage.json"
        }]
    }
}
```

### 9.2 Syntax Highlighting Updates

Remove `@component` keyword highlighting, add `Element` as a special type when used as return type.

---

## 10. Complexity Assessment

| Factor | Assessment |
|--------|------------|
| Lexer changes | Low — remove one token type |
| Parser changes | Medium — new decision logic for func parsing |
| Formatter changes | Low — straightforward syntax swap |
| CLI changes | Low — file extension pattern change |
| LSP changes | Low — file extension handling |
| Tree-sitter changes | Medium — grammar restructure |
| VSCode extension | Low — configuration updates |
| Testing | Medium — update all test fixtures |
| Documentation | Medium — update CLAUDE.md, examples |

**Assessed Size:** Medium\
**Recommended Phases:** 3

### Phase Breakdown

1. **Phase 1: Core Syntax Changes** (Medium)
   - Update token.go to remove `TokenAtComponent`
   - Update lexer to remove `@component` keyword handling
   - Update parser to detect `func ... Element` as component
   - Update formatter to emit new syntax
   - Change file discovery from `.tui` to `.gsx`
   - Update output naming from `*_tui.go` to `*_gsx.go`
   - Rename test fixtures from `.tui` to `.gsx` (cmd/tui/testdata/, pkg/tuigen/testdata/, etc.)
   - Update unit tests

2. **Phase 2: LSP & Editor Support** (Medium)
   - Update LSP server for `.gsx` files
   - Rename tree-sitter grammar directory
   - Update tree-sitter grammar rules
   - Update VSCode extension

3. **Phase 3: Documentation & Examples** (Small)
   - Update CLAUDE.md
   - Convert all example `.tui` files to `.gsx`
   - Update README and other documentation
   - Verify all examples compile and run

---

## 11. Success Criteria

1. `func Name(params) Element { ... }` is parsed as a component
2. `func name(params) Type { ... }` is captured as a raw Go function
3. `func name(params) { ... }` is captured as a raw Go function (side-effect)
4. `tui generate ./...` finds and processes `.gsx` files
5. `tui fmt` outputs the new syntax correctly
6. LSP provides diagnostics and completion for `.gsx` files
7. Tree-sitter grammar parses `.gsx` files correctly
8. VSCode extension provides syntax highlighting for `.gsx` files
9. All examples compile and run
10. Error messages reference correct line/column in `.gsx` files

---

## 12. Open Questions

1. **Should `Element` be configurable?**\
   → No. The bare `Element` identifier is the canonical marker. Users wanting custom types write regular Go.

2. **Should we support `*Element` as component marker?**\
   → No. Keep the rule simple: exactly `Element` as sole return type.

3. **Should unexported components be allowed?**\
   → Yes. `func helper() Element { ... }` is valid for package-internal components.

4. **What about generic components `func List[T any]() Element`?**\
   → Deferred. Start with non-generic components. Add generics support later if needed.
