# DSL & Code Generation Implementation Plan

Implementation phases for the DSL and code generation system. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Core Types & Lexer

**Reference:** [dsl-codegen-design.md §3.1-3.2](./dsl-codegen-design.md#31-token-types)

**Completed in commit:** (pending)

- [ ] Create `pkg/tuigen/token.go`
  - Define `TokenType` enum with all token types
  - Define `Token` struct with Type, Literal, Line, Column
  - Define `Position` struct for error locations
  - Implement `String()` for debugging

- [ ] Create `pkg/tuigen/errors.go`
  - Define `Error` struct with Position, Message, Hint
  - Implement `Error()` interface for error formatting
  - Implement `ErrorList` for collecting multiple errors

- [ ] Create `pkg/tuigen/lexer.go`
  - Implement `Lexer` struct with source, position tracking
  - Implement `NewLexer(filename, source string) *Lexer`
  - Implement `Next() Token` main tokenization method
  - Handle whitespace, newlines, comments
  - Handle keywords: `package`, `import`, `func`, `return`, `if`, `else`, `for`, `range`
  - Handle DSL keywords: `@component`, `@let`, `@for`, `@if`, `@else`
  - Handle identifiers and literals (int, float, string, raw string)
  - Handle operators and punctuation
  - Handle Go expressions inside `{...}` with brace balancing
  - Handle XML-like tokens: `<`, `</`, `/>`, `>`

- [ ] Create `pkg/tuigen/lexer_test.go`
  - Test basic tokens (keywords, identifiers, literals)
  - Test operators and punctuation
  - Test Go expression extraction with nested braces
  - Test string literals with escapes
  - Test multi-line input with position tracking
  - Test error cases (unclosed strings, invalid characters)

**Tests:** Run `go test ./pkg/tuigen/...` at phase end

---

## Phase 2: Parser & AST

**Reference:** [dsl-codegen-design.md §3.2](./dsl-codegen-design.md#32-ast-types)

**Completed in commit:** (pending)

- [ ] Create `pkg/tuigen/ast.go`
  - Define `Node` interface with `node()` marker and `Position()` method
  - Define `File` struct (Package, Imports, Components, Funcs)
  - Define `Import` struct (Alias, Path)
  - Define `Component` struct (Name, Params, Body, Pos)
  - Define `Param` struct (Name, Type)
  - Define `Element` struct (Tag, Attributes, Children, SelfClose, Pos)
  - Define `Attribute` struct (Name, Value, Pos)
  - Define `GoExpr` struct (Code, Pos)
  - Define `StringLit`, `IntLit`, `FloatLit` structs
  - Define `LetBinding` struct (Name, Element, Pos)
  - Define `ForLoop` struct (Index, Value, Iterable, Body, Pos)
  - Define `IfStmt` struct (Condition, Then, Else, Pos)
  - Define `GoCode` struct for embedded Go code blocks
  - Implement `node()` for all types

- [ ] Create `pkg/tuigen/parser.go`
  - Implement `Parser` struct with lexer, current token, peek token, errors
  - Implement `NewParser(lexer *Lexer) *Parser`
  - Implement `ParseFile() (*File, error)` entry point
  - Implement `parsePackage() string`
  - Implement `parseImports() []Import`
  - Implement `parseComponent() *Component`
  - Implement `parseParams() []Param`
  - Implement `parseElement() *Element`
  - Implement `parseAttribute() *Attribute`
  - Implement `parseChildren() []Node`
  - Implement `parseLet() *LetBinding`
  - Implement `parseFor() *ForLoop`
  - Implement `parseIf() *IfStmt`
  - Implement `parseGoExpr() *GoExpr`
  - Implement error recovery for common mistakes

- [ ] Create `pkg/tuigen/parser_test.go`
  - Test package and import parsing
  - Test simple component parsing
  - Test element parsing (self-closing, with children)
  - Test attribute parsing (literals, expressions)
  - Test @let binding parsing
  - Test @for loop parsing
  - Test @if/@else parsing
  - Test nested elements
  - Test Go expression parsing in attributes and content
  - Test error messages for invalid syntax

**Tests:** Run `go test ./pkg/tuigen/...` at phase end

---

## Phase 3: Code Generator

**Reference:** [dsl-codegen-design.md §5](./dsl-codegen-design.md#5-code-generation)

**Completed in commit:** (pending)

- [ ] Create `pkg/tuigen/generator.go`
  - Implement `Generator` struct with output buffer, indent level, var counter
  - Implement `NewGenerator() *Generator`
  - Implement `Generate(file *File) ([]byte, error)` entry point
  - Implement `generateHeader()` with "DO NOT EDIT" comment
  - Implement `generatePackage(pkg string)`
  - Implement `generateImports(imports []Import)`
  - Implement `generateComponent(c *Component)`
  - Implement `generateElement(e *Element) string` returns var name
  - Implement `generateAttribute(attr *Attribute) string` returns option code
  - Implement `generateLetBinding(l *LetBinding)`
  - Implement `generateForLoop(f *ForLoop)`
  - Implement `generateIfStmt(i *IfStmt)`
  - Implement `generateGoExpr(g *GoExpr) string`
  - Implement `generateChildren(parent string, children []Node)`
  - Implement attribute-to-option mapping table
  - Implement element tag-to-options mapping (box, text, scrollable, button)
  - Handle proper Go formatting of output

- [ ] Create `pkg/tuigen/generator_test.go`
  - Test simple component generation
  - Test element generation with attributes
  - Test nested element generation
  - Test @let binding generation
  - Test @for loop generation
  - Test @if/@else generation
  - Test text element content generation
  - Test attribute expression generation
  - Test import propagation
  - Verify generated code compiles (go build check)

- [ ] Create `pkg/tuigen/analyzer.go`
  - Implement `Analyzer` struct
  - Implement `Analyze(file *File) []error`
  - Validate element tags are known
  - Validate attributes are valid for element type
  - Validate required imports are present
  - Add missing standard imports (element, layout, tui)
  - Warn on unused @let bindings

- [ ] Create `pkg/tuigen/analyzer_test.go`
  - Test unknown element tag detection
  - Test unknown attribute detection
  - Test import validation
  - Test automatic import insertion

**Tests:** Run `go test ./pkg/tuigen/...` at phase end

---

## Phase 4: CLI & Integration

**Reference:** [dsl-codegen-design.md §6](./dsl-codegen-design.md#6-cli-tool)

**Completed in commit:** (pending)

- [ ] Create `cmd/tui/main.go`
  - Implement CLI entry point
  - Parse command-line arguments
  - Dispatch to subcommands

- [ ] Create `cmd/tui/generate.go`
  - Implement `generate` subcommand
  - Implement `-v` verbose flag
  - Implement file/directory argument handling
  - Implement `./...` recursive file discovery
  - Implement `.tui` file filtering
  - Implement output file naming (`foo.tui` → `foo_tui.go`)
  - Implement error collection and reporting
  - Format output with gofmt

- [ ] Create `cmd/tui/check.go`
  - Implement `check` subcommand
  - Parse and analyze without generating
  - Report errors with line numbers

- [ ] Create integration test
  - Create `testdata/simple.tui` with basic component
  - Create `testdata/complex.tui` with loops, conditionals, @let
  - Run `tui generate testdata/`
  - Verify `testdata/simple_tui.go` exists and compiles
  - Verify `testdata/complex_tui.go` exists and compiles
  - Run generated code to verify runtime behavior

- [ ] Create end-to-end example
  - Create `examples/dsl-counter/counter.tui`
  - Create `examples/dsl-counter/main.go` using generated component
  - Document workflow in README

**Tests:** Run `go test ./cmd/tui/...` and integration tests at phase end

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Core Types & Lexer | Pending |
| 2 | Parser & AST | Pending |
| 3 | Code Generator | Pending |
| 4 | CLI & Integration | Pending |

## Files to Create

```
pkg/tuigen/
├── token.go          # Token types
├── errors.go         # Error handling
├── lexer.go          # Lexer
├── lexer_test.go
├── ast.go            # AST types
├── parser.go         # Parser
├── parser_test.go
├── generator.go      # Code generator
├── generator_test.go
├── analyzer.go       # Semantic analysis
└── analyzer_test.go

cmd/tui/
├── main.go           # CLI entry point
├── generate.go       # generate subcommand
└── check.go          # check subcommand

examples/dsl-counter/
├── counter.tui       # Example DSL component
├── main.go           # Example usage
└── README.md         # Documentation
```

## Files to Modify

| File | Changes |
|------|---------|
| `go.mod` | May need to add dependencies for CLI flags |

## Dependencies

The implementation should use only standard library where possible:

- `go/format` for formatting generated code
- `flag` or similar for CLI argument parsing
- No external lexer/parser generators
