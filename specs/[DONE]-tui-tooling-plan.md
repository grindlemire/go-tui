# TUI Tooling Implementation Plan

Implementation phases for TUI developer tooling. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: VS Code Extension (TextMate Grammar)

**Reference:** [tui-tooling-design.md §2](./tui-tooling-design.md#2-architecture), [§5.1](./tui-tooling-design.md#51-textmate-grammar-vs-code)

**Status:** Complete

- [x] Create `editor/vscode/package.json`
  - Extension manifest with name, version, publisher
  - Declare `.tui` language contribution
  - Register TextMate grammar for `source.tui`
  - Configure file icons and associations

- [x] Create `editor/vscode/language-configuration.json`
  - Bracket pairs: `{}`, `()`, `[]`, `<>`
  - Auto-closing pairs for quotes and brackets
  - Comment configuration: `//` and `/* */`
  - Folding markers

- [x] Create `editor/vscode/syntaxes/tui.tmLanguage.json`
  - `keyword.control.tui` for `@component`, `@for`, `@if`, `@else`, `@let`
  - `entity.name.function.tui` for component names
  - `entity.name.tag.tui` for element tags (`<box>`, `</text>`)
  - `entity.other.attribute-name.tui` for attribute names
  - `string.quoted.double.tui` for string literals
  - `constant.numeric.tui` for numbers
  - `comment.line.double-slash.tui` and `comment.block.tui`
  - `meta.embedded.expression.go` for Go expressions in `{}`
  - Grammar injection for Go code inside braces

- [x] Create `editor/vscode/README.md`
  - Installation instructions
  - Feature overview with screenshots placeholders

- [x] Add tests: Create `editor/vscode/test/` with sample `.tui` files
  - `simple.tui` - basic component
  - `complex.tui` - all DSL features

**Tests:** Manual verification in VS Code: open test files, verify all tokens are colored correctly

---

## Phase 2: Tree-sitter Grammar

**Reference:** [tui-tooling-design.md §5.2](./tui-tooling-design.md#52-tree-sitter-grammar)

**Status:** Complete

- [x] Create `editor/tree-sitter-tui/package.json`
  - Tree-sitter grammar package configuration
  - Build scripts for generating parser

- [x] Create `editor/tree-sitter-tui/grammar.js`
  - `source_file` root rule with package, imports, declarations
  - `package_clause` for `package name`
  - `import_declaration` for single and grouped imports
  - `component_declaration` for `@component Name(params) { body }`
  - `element` for `<tag attrs>children</tag>` and `<tag />`
  - `attribute` for `name=value` and `name={expr}`
  - `for_statement` for `@for i, v := range items { }`
  - `if_statement` for `@if condition { } @else { }`
  - `let_binding` for `@let name = <element>`
  - `component_call` for `@ComponentName(args)`
  - `go_expression` for `{expr}` blocks
  - `comment` for `//` and `/* */`

- [x] Create `editor/tree-sitter-tui/queries/highlights.scm`
  - Map Tree-sitter node types to highlight groups
  - `@keyword` for DSL keywords
  - `@function` for component names
  - `@tag` for element tags
  - `@property` for attributes
  - `@string`, `@number`, `@comment`

- [x] Create `editor/tree-sitter-tui/queries/injections.scm`
  - Inject Go language for `go_expression` nodes
  - Inject Go for import paths (string highlighting)

- [x] Create `editor/tree-sitter-tui/test/corpus/` test files
  - Test cases for each grammar rule
  - Expected parse tree output

**Tests:** Run `tree-sitter test` to verify grammar parses all constructs correctly

---

## Phase 3: Code Formatter

**Reference:** [tui-tooling-design.md §5.5](./tui-tooling-design.md#55-formatter-rules)

**Status:** Complete

- [x] Create `pkg/formatter/formatter.go`
  - `type Formatter struct` with configuration options
  - `func (f *Formatter) Format(source string) (string, error)`
  - Parse source using `tuigen.NewParser`
  - Call printer to generate formatted output
  - Return error if source has parse errors

- [x] Create `pkg/formatter/printer.go`
  - `type Printer struct` with output buffer, indent tracking
  - `func (p *Printer) PrintFile(file *tuigen.File) string`
  - `printPackage`, `printImports`, `printComponent`
  - `printElement` with smart line-breaking for attributes
  - `printForLoop`, `printIfStmt`, `printLetBinding`
  - Consistent indentation (tabs by default)
  - Line width aware formatting (break long attribute lists)

- [x] Create `cmd/tui/fmt.go`
  - `tui fmt <files...>` command implementation
  - `--stdout` flag to print instead of modifying
  - `--check` flag for CI (exit 1 if not formatted)
  - `./...` glob pattern support for directories
  - Parallel file processing for performance

- [x] Create `pkg/formatter/formatter_test.go`
  - Table-driven tests for formatting scenarios
  - Round-trip tests: format(format(x)) == format(x)
  - Test preservation of comments
  - Test multi-line attribute formatting

**Tests:** `go test ./pkg/formatter/...` - all formatter tests pass

---

## Phase 4: LSP Server Foundation

**Reference:** [tui-tooling-design.md §2](./tui-tooling-design.md#2-architecture), [§5.3](./tui-tooling-design.md#53-lsp-capabilities)

**Status:** Complete

- [x] Create `pkg/lsp/server.go`
  - `type Server struct` with connection, document manager, component index
  - `func NewServer() *Server`
  - `func (s *Server) Run(ctx context.Context) error` - main loop
  - JSON-RPC 2.0 message handling over stdio
  - Use `go.lsp.dev/protocol` for LSP types

- [x] Create `pkg/lsp/handler.go`
  - LSP method router
  - `initialize` / `initialized` handlers
  - `shutdown` / `exit` handlers
  - `textDocument/didOpen`, `didChange`, `didClose` handlers
  - Capability negotiation (advertise supported features)

- [x] Create `pkg/lsp/document.go`
  - `type Document struct` with URI, content, version, AST, errors
  - `type DocumentManager struct` tracking open documents
  - `func (dm *DocumentManager) Open(uri, content string)`
  - `func (dm *DocumentManager) Update(uri string, changes []Change)`
  - `func (dm *DocumentManager) Close(uri string)`
  - Re-parse on every change (incremental parsing later)

- [x] Create `pkg/lsp/diagnostics.go`
  - `func (s *Server) publishDiagnostics(uri string)`
  - Convert `tuigen.Error` to LSP `Diagnostic`
  - Map severity (Error, Warning, Information, Hint)
  - Publish diagnostics after every document change

- [x] Create `cmd/tui/lsp.go`
  - `tui lsp` subcommand
  - `--log` flag for debug logging to file
  - Start server on stdio

- [x] Create `pkg/lsp/server_test.go`
  - Test server initialization
  - Test document open/change/close lifecycle
  - Test diagnostics published on parse errors

**Tests:** `go test ./pkg/lsp/...` - server lifecycle and diagnostics tests pass

---

## Phase 5: LSP Features (Definition, Hover, Completion, Symbols)

**Reference:** [tui-tooling-design.md §5.3](./tui-tooling-design.md#53-lsp-capabilities)

**Status:** Complete

- [x] Create `pkg/lsp/index.go`
  - `type ComponentIndex struct` mapping component names to locations
  - `func (idx *ComponentIndex) Add(uri string, comp *tuigen.Component)`
  - `func (idx *ComponentIndex) Remove(uri string)`
  - `func (idx *ComponentIndex) Lookup(name string) (Location, bool)`
  - Rebuild index on document changes

- [x] Create `pkg/lsp/definition.go`
  - `textDocument/definition` handler
  - Detect if cursor is on `@ComponentName` call
  - Look up component in index
  - Return location of `@component` definition

- [x] Create `pkg/lsp/hover.go`
  - `textDocument/hover` handler
  - Hover on `@ComponentName` shows signature: `func Name(params) *element.Element`
  - Hover on element tags shows available attributes
  - Hover on attributes shows type information

- [x] Create `pkg/lsp/completion.go`
  - `textDocument/completion` handler
  - After `@` - suggest component names from index
  - After `<` - suggest element tags (`box`, `text`, etc.)
  - After element tag - suggest attributes for that element
  - Inside `{` - defer to gopls proxy (Phase 6)

- [x] Create `pkg/lsp/symbols.go`
  - `textDocument/documentSymbol` handler
  - Return list of components in current file
  - `workspace/symbol` handler
  - Search components across all indexed files

- [x] Create `pkg/lsp/features_test.go`
  - Test go-to-definition finds component
  - Test hover returns correct signature
  - Test completion returns components after `@`
  - Test document symbols lists all components

**Tests:** `go test ./pkg/lsp/...` - all LSP feature tests pass

---

## Phase 6: gopls Proxy for Go Intelligence

**Reference:** [tui-tooling-design.md §5.4](./tui-tooling-design.md#54-gopls-proxy)

**Status:** Complete

- [x] Create `pkg/lsp/gopls/proxy.go`
  - `type GoplsProxy struct` managing gopls subprocess
  - `func NewGoplsProxy() (*GoplsProxy, error)` - start gopls
  - `func (p *GoplsProxy) Initialize(rootURI string) error`
  - `func (p *GoplsProxy) Completion(uri string, pos Position) ([]CompletionItem, error)`
  - `func (p *GoplsProxy) Hover(uri string, pos Position) (*Hover, error)`
  - `func (p *GoplsProxy) Definition(uri string, pos Position) ([]Location, error)`
  - JSON-RPC communication over stdin/stdout pipes

- [x] Create `pkg/lsp/gopls/generate.go`
  - `func GenerateVirtualGo(file *tuigen.File) (string, *SourceMap)`
  - Generate valid Go code from `.tui` AST
  - Preserve Go expressions, imports, package declaration
  - Generate dummy variable assignments for expressions (to preserve positions)
  - Return source map for position translation

- [x] Create `pkg/lsp/gopls/mapping.go`
  - `type SourceMap struct` with bidirectional mappings
  - `func (sm *SourceMap) TuiToGo(line, col int) (int, int)`
  - `func (sm *SourceMap) GoToTui(line, col int) (int, int)`
  - Handle multi-line Go expressions
  - Cache virtual files and source maps per document

- [x] Update `pkg/lsp/completion.go`
  - Inside `{expr}` - call `gopls.Completion()` via proxy
  - Translate positions using source map
  - Merge gopls results with TUI-specific completions

- [x] Update `pkg/lsp/hover.go`
  - Inside `{expr}` - call `gopls.Hover()` via proxy
  - Translate positions and return Go type information

- [x] Update `pkg/lsp/definition.go`
  - Inside `{expr}` - call `gopls.Definition()` via proxy
  - Handle jumps to Go standard library / dependencies

- [x] Create `pkg/lsp/gopls/proxy_test.go`
  - Test gopls subprocess lifecycle
  - Test position mapping accuracy
  - Test completion inside Go expressions
  - Integration test: full round-trip completion

**Tests:** `go test ./pkg/lsp/gopls/...` - gopls proxy tests pass, integration tests verify end-to-end

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | VS Code Extension (TextMate Grammar) | Complete |
| 2 | Tree-sitter Grammar | Complete |
| 3 | Code Formatter (`tui fmt`) | Complete |
| 4 | LSP Server Foundation | Complete |
| 5 | LSP Features (Definition, Hover, Completion, Symbols) | Complete |
| 6 | gopls Proxy for Go Intelligence | Complete |

## Files to Create

```
go-tui/
├── cmd/tui/
│   ├── fmt.go                    # Phase 3
│   └── lsp.go                    # Phase 4
├── pkg/
│   ├── formatter/                # Phase 3
│   │   ├── formatter.go
│   │   ├── formatter_test.go
│   │   └── printer.go
│   └── lsp/                      # Phases 4-6
│       ├── server.go
│       ├── server_test.go
│       ├── handler.go
│       ├── document.go
│       ├── diagnostics.go
│       ├── index.go
│       ├── definition.go
│       ├── hover.go
│       ├── completion.go
│       ├── symbols.go
│       ├── features_test.go
│       └── gopls/                # Phase 6
│           ├── proxy.go
│           ├── proxy_test.go
│           ├── generate.go
│           └── mapping.go
└── editor/
    ├── vscode/                   # Phase 1
    │   ├── package.json
    │   ├── language-configuration.json
    │   ├── README.md
    │   ├── syntaxes/
    │   │   └── tui.tmLanguage.json
    │   └── test/
    │       ├── simple.tui
    │       └── complex.tui
    └── tree-sitter-tui/          # Phase 2
        ├── package.json
        ├── grammar.js
        ├── queries/
        │   ├── highlights.scm
        │   └── injections.scm
        └── test/
            └── corpus/
```

## Files to Modify

| File | Changes |
|------|---------|
| `cmd/tui/main.go` | Add `fmt` and `lsp` subcommands |
| `go.mod` | Add LSP dependencies (`go.lsp.dev/protocol`, etc.) |
