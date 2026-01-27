# Named Element References Implementation Plan

Implementation phases for named element refs (`#Name` syntax). Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Lexer, AST, Parser, and Analyzer

**Reference:** [named-element-refs-design.md §3](./named-element-refs-design.md#3-core-entities)

**Completed in commit:** (pending)

- [ ] Modify `pkg/tuigen/token.go`
  - Add `TokenHash` constant for `#` character
  - See [design §3.1](./named-element-refs-design.md#31-lexer-changes)

- [ ] Modify `pkg/tuigen/lexer.go`
  - Add case for `#` in `Next()` to emit `TokenHash`
  - Token should capture position (line, column) for error reporting

- [ ] Modify `pkg/tuigen/ast.go`
  - Add `NamedRef string` field to `Element` struct
  - Add `RefKey *Expression` field for keyed refs (map-based)
  - See [design §3.2](./named-element-refs-design.md#32-ast-changes)

- [ ] Modify `pkg/tuigen/parser.go`
  - After parsing tag name, check for `TokenHash`
  - If found, consume `#` and expect `TokenIdent` for ref name
  - Store in `Element.NamedRef`
  - Parse `key={expr}` attribute into `Element.RefKey`
  - See [design §3.3](./named-element-refs-design.md#33-parser-changes)

- [ ] Modify `pkg/tuigen/analyzer.go`
  - Add `NamedRef` struct to track ref metadata (name, element, InLoop, InConditional, KeyExpr, KeyType)
  - Add `validateNamedRefs()` function that walks AST
  - Track loop/conditional context during traversal
  - Validate: must be valid Go identifier starting with uppercase
  - Validate: `Root` is reserved name (error if used)
  - Validate: names unique within component (including across conditional branches)
  - Validate: `key` attribute only valid inside `@for` loops
  - Infer key type from expression for map-based refs
  - Return collected refs with context flags for generator
  - See [design §3.4](./named-element-refs-design.md#34-analyzer-changes)

- [ ] Add tests to `pkg/tuigen/lexer_test.go`
  - Test `#` token is lexed correctly
  - Test `#Name` followed by attributes

- [ ] Add tests to `pkg/tuigen/parser_test.go`
  - Test `<div #Content>` parses with NamedRef="Content"
  - Test `<span #Title class="bold">` parses ref and attributes
  - Test `<div #Items key={item.ID}>` parses ref and key
  - Test self-closing `<div #Spacer />`

- [ ] Add tests to `pkg/tuigen/analyzer_test.go`
  - Test valid ref names pass validation
  - Test invalid ref name (lowercase, starts with number) produces error
  - Test `#Root` produces reserved name error
  - Test duplicate ref names produce error with position info
  - Test ref inside `@for` is marked `InLoop=true`
  - Test ref inside `@if` is marked `InConditional=true`
  - Test ref inside nested `@for` loops is still `InLoop=true` (flat)
  - Test `key={expr}` inside loop is valid
  - Test `key={expr}` outside loop produces error

- [ ] Update `editor/tree-sitter-tui/grammar.js`
  - Add `#` token recognition in element rule
  - Parse identifier after `#` as named_ref
  - Update element rule: `<tag #Name ...>`

- [ ] Update `editor/vscode/syntaxes/tui.tmLanguage.json`
  - Add pattern to highlight `#Name` (e.g., as entity.name.tag or variable)
  - Match pattern: `#[A-Z][a-zA-Z0-9]*`

**Tests:** Run `go test ./pkg/tuigen/... -run "Lexer|Parser|Analyzer"` once at phase end

---

## Phase 2: Generator Struct Returns

**Reference:** [named-element-refs-design.md §3.5](./named-element-refs-design.md#35-generator-changes)

**Completed in commit:** (pending)

- [ ] Modify `pkg/tuigen/generator.go` - struct generation
  - Add `generateViewStruct()` function
  - Always generate `ComponentNameView` struct for all components
  - Struct always has `Root *element.Element` field
  - Add fields for each named ref:
    - Outside loops/conditionals: `*element.Element`
    - Inside `@for` without key: `[]*element.Element`
    - Inside `@for` with `key={expr}`: `map[KeyType]*element.Element`
    - Inside `@if`: `*element.Element` with `// may be nil` comment
  - See [design §3.5](./named-element-refs-design.md#35-generator-changes)

- [ ] Modify `pkg/tuigen/generator.go` - function signature
  - Change return type from `*element.Element` to `ComponentNameView`
  - Pre-declare `var view ComponentNameView` at function start for closure capture
  - Declare slice/map variables at function scope for loop refs
  - Declare pointer variables at function scope for conditional refs

- [ ] Modify `pkg/tuigen/generator.go` - element generation
  - When generating element with `NamedRef`, use ref name as variable name
  - Inside loops: append to slice or set in map (based on key presence)
  - Inside conditionals: assign to pre-declared variable

- [ ] Modify `pkg/tuigen/generator.go` - return statement
  - Populate `view` struct before returning
  - Include `Root` and all named refs
  - Return `view` instead of root element

- [ ] Modify `pkg/tuigen/generator.go` - ref on root element
  - When root element has `#Name`, both `Root` and `Name` point to same element
  - See [design Q8](./named-element-refs-design.md#q8-can-refs-be-placed-on-the-root-element)

- [ ] Add tests to `pkg/tuigen/generator_test.go`
  - Test component without refs generates struct with only `Root`
  - Test component with `#Name` generates struct with `Root` and `Name` fields
  - Test multiple refs generate multiple fields
  - Test ref inside `@for` generates slice field `[]*element.Element`
  - Test ref inside `@for` with `key={expr}` generates map field
  - Test ref inside `@if` generates field that may be nil
  - Test ref on root element generates both `Root` and named field pointing to same element
  - Test nested refs at various depths are captured
  - Test generated code compiles (go build check)
  - Test `var view` is declared before element creation (for closure capture)

- [ ] Update `examples/streaming-dsl/` to use named refs
  - Update `streaming.tui` to use `#Content` pattern
  - Update `main.go` to use `view.Content` instead of passing refs
  - Verify example compiles and runs

**Tests:** Run `go test ./pkg/tuigen/... -run Generator` and `go build ./examples/...` once at phase end

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Lexer, AST, Parser, Analyzer for #Name syntax | Pending |
| 2 | Generator struct returns and code generation | Pending |

## Files to Modify

```
pkg/tuigen/
├── token.go           # Add TokenHash
├── lexer.go           # Lex # token
├── ast.go             # Add NamedRef, RefKey fields
├── parser.go          # Parse #Name syntax
├── analyzer.go        # Validate refs, track context
├── generator.go       # Generate struct returns
├── lexer_test.go      # Tests for # token
├── parser_test.go     # Tests for #Name parsing
├── analyzer_test.go   # Tests for ref validation
└── generator_test.go  # Tests for struct generation

editor/
├── tree-sitter-tui/
│   └── grammar.js     # Parse #Name in elements
└── vscode/
    └── syntaxes/
        └── tui.tmLanguage.json  # Highlight #Name

examples/
└── streaming-dsl/
    ├── streaming.tui  # Use #Name pattern
    └── main.go        # Use view.Content
```

## Files Unchanged

| File | Reason |
|------|--------|
| `pkg/tuigen/tailwind.go` | No changes needed for named refs |
| `pkg/tuigen/errors.go` | Existing error types sufficient |
| `pkg/formatter/` | Formatter can print NamedRef once AST supports it |
| `pkg/lsp/` | LSP support deferred (noted in design §9) |
