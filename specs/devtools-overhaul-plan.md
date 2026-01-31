# Developer Tooling Overhaul Implementation Plan

Implementation phases for the LSP rearchitecture, new construct support, and editor tooling updates. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Core Infrastructure (Schema + CursorContext + Provider Interfaces)

**Reference:** [devtools-overhaul-design.md §2](./devtools-overhaul-design.md#2-architecture)

**Completed in commit:** (pending)

- [ ] Create `pkg/lsp/schema/schema.go`
  - Define `ElementDef` struct with Tag, Description, Attributes, SelfClosing, Category fields
  - Define `AttributeDef` struct with Name, Type, Description, Category fields
  - Define `EventHandlerDef` struct for onClick, onFocus, onBlur, onKeyPress, onEvent
  - Populate `Elements` map with all built-in elements (`div`, `span`, `p`, `ul`, `li`, `button`, `input`, `table`, `progress`, `hr`, `br`) including their valid attributes
  - Populate `EventHandlers` map with handler definitions and expected signatures
  - Add lookup functions: `GetElement(tag)`, `GetAttribute(tag, attr)`, `GetEventHandler(name)`
  - See [devtools-overhaul-design.md §3](./devtools-overhaul-design.md#3-core-entities) for entity definitions

- [ ] Create `pkg/lsp/schema/keywords.go`
  - Define `KeywordDef` struct with Name, Description, Syntax, Documentation (markdown) fields
  - Populate keyword definitions for: `templ`, `@for`, `@if`, `@else`, `@let`, `package`, `import`, `func`
  - Add `GetKeyword(name)` lookup function
  - Consolidate all keyword documentation currently spread across hover.go and completion.go into this single source

- [ ] Create `pkg/lsp/schema/tailwind.go`
  - Define `TailwindClassDef` struct with Name, Pattern, Description, Category, SortKey fields
  - Move all tailwind class definitions from the current completion.go and hover.go into this centralized location
  - Include pattern-based classes (gap-N, p-N, m-N, text-COLOR, bg-COLOR)
  - Add `MatchClass(input)` function that returns matching definitions (supports prefix filtering)
  - Add `GetClassDoc(className)` for hover documentation

- [ ] Create `pkg/lsp/context.go`
  - Define `NodeKind` enum with all kinds from design: Unknown, Component, Element, Attribute, NamedRef, GoExpr, ForLoop, IfStmt, LetBinding, StateDecl, StateAccess, Parameter, Function, ComponentCall, EventHandler, Text, Keyword, TailwindClass
  - Define `Scope` struct with Component, Function, ForLoop, IfStmt, StateVars, NamedRefs, LetBinds, Params fields
  - Define `CursorContext` struct with Document, Position, Offset, Node, NodeKind, Scope, ParentChain, Word, Line, InGoExpr, InClassAttr, InElement fields
  - Implement `ResolveCursorContext(doc *Document, pos protocol.Position) (*CursorContext, error)` which walks the AST to determine the node under the cursor, builds the parent chain, resolves the scope, and classifies the NodeKind
  - Handle edge cases: cursor between nodes, cursor at file level, cursor in whitespace

- [ ] Create `pkg/lsp/provider/provider.go`
  - Define all provider interfaces: HoverProvider, CompletionProvider, DefinitionProvider, ReferencesProvider, DocumentSymbolProvider, WorkspaceSymbolProvider, DiagnosticsProvider, FormattingProvider, SemanticTokensProvider
  - Define `Registry` struct that holds one instance of each provider
  - Add `NewRegistry(...)` constructor that wires up all providers
  - See [devtools-overhaul-design.md §2](./devtools-overhaul-design.md#2-architecture) for interface definitions

- [ ] Create `pkg/lsp/router.go`
  - Implement method routing that replaces the current handler.go switch statement
  - Route each LSP method to the corresponding provider via the Registry
  - For position-based requests (hover, completion, definition, references): resolve CursorContext first, then dispatch to provider
  - For document-based requests (symbols, diagnostics, formatting, semantic tokens): look up Document, then dispatch
  - Handle lifecycle methods (initialize, shutdown, exit) directly
  - Handle document sync methods (didOpen, didChange, didClose, didSave) directly

- [ ] Modify `pkg/lsp/server.go`
  - Simplify to focus on JSON-RPC I/O and server lifecycle only
  - Remove the request routing logic (now in router.go)
  - Keep message reading/writing, document management, workspace indexing
  - Wire up the Router and Registry during initialization
  - Maintain gopls proxy management

- [ ] Create `pkg/lsp/schema/schema_test.go`
  - Test all element lookups return correct attributes
  - Test event handler lookups
  - Test keyword lookups
  - Test tailwind class matching and documentation

- [ ] Create `pkg/lsp/context_test.go`
  - Test CursorContext resolution for each NodeKind: cursor on component name, element tag, attribute name, attribute value, `#ref`, Go expression, `@for` keyword, `@if` keyword, `@let` binding, component call, text content
  - Test scope resolution: state vars and named refs in scope are correctly collected
  - Test parent chain construction

**Tests:** Run `go test ./pkg/lsp/schema/... ./pkg/lsp/...` (context_test.go) once at phase end

---

## Phase 2: Migrate Navigation Providers (Hover + Definition + References)

**Reference:** [devtools-overhaul-design.md §2](./devtools-overhaul-design.md#2-architecture)

**Completed in commit:** (pending)

- [x] Create `pkg/lsp/provider/hover.go`
  - Implement `HoverProvider` interface using CursorContext
  - Switch on `ctx.NodeKind` to dispatch to the right hover logic
  - Use `schema.GetElement()` for element hover instead of inline definitions
  - Use `schema.GetKeyword()` for keyword hover instead of inline documentation
  - Use `schema.GetClassDoc()` for tailwind class hover
  - Use `schema.GetEventHandler()` for event attribute hover
  - Delegate to gopls proxy for Go expression hover (using ctx.InGoExpr)
  - Port all existing hover logic from current hover.go but refactored to use CursorContext and Schema
  - Component hover: show signature from ComponentIndex
  - Function hover: show signature from ComponentIndex
  - Parameter hover: show type and component context from ctx.Scope
  - Attribute hover: show type and description from schema

- [x] Create `pkg/lsp/provider/definition.go`
  - Implement `DefinitionProvider` interface using CursorContext
  - Switch on `ctx.NodeKind` for dispatch
  - Component calls → jump to component declaration (from ComponentIndex)
  - Function calls → jump to function declaration (from ComponentIndex)
  - Parameters → jump to parameter in component signature
  - Let bindings → jump to `@let` declaration
  - For loop variables → jump to `@for` declaration
  - GoCode variables → jump to `:=` or `var` declaration
  - Delegate to gopls for Go expression definitions
  - Port existing definition.go logic, refactored to use CursorContext

- [x] Create `pkg/lsp/provider/references.go`
  - Implement `ReferencesProvider` interface using CursorContext
  - Switch on `ctx.NodeKind` for dispatch
  - Components → find all `@ComponentName` calls across workspace
  - Functions → find all calls across workspace
  - Parameters → find usages within component scope
  - Let bindings → find usages within scope
  - Loop variables → find usages within loop body
  - GoCode variables → find usages within scope
  - Search open documents via DocumentManager, closed files via workspace AST cache
  - Port existing references.go logic, refactored to use CursorContext

- [x] Wire up navigation providers in `pkg/lsp/router.go`
  - Register HoverProvider, DefinitionProvider, ReferencesProvider in the Registry
  - Update `handleHover`, `handleDefinition`, `handleReferences` routes to resolve CursorContext and dispatch to providers
  - Verify end-to-end: LSP request → router → CursorContext → provider → response

- [x] Remove old files: `pkg/lsp/hover.go`, `pkg/lsp/definition.go`, `pkg/lsp/references.go`
  - Delete after new providers are wired up and tested

- [x] Create `pkg/lsp/provider/hover_test.go`
  - Test hover for each node kind: component declaration, element tag, attribute, keyword, tailwind class, Go expression, parameter
  - Test gopls fallthrough for Go expressions

- [x] Create `pkg/lsp/provider/definition_test.go`
  - Test definition for: component calls, function calls, let bindings, for loop vars, parameters
  - Test gopls delegation for Go identifiers

- [x] Create `pkg/lsp/provider/references_test.go`
  - Test references for: components (cross-file), functions, parameters, let bindings, loop vars

**Tests:** Run `go test ./pkg/lsp/...` once at phase end ✅

---

## Phase 3: Migrate Completion + Symbol Providers

**Reference:** [devtools-overhaul-design.md §2](./devtools-overhaul-design.md#2-architecture)

**Completed in commit:** (pending)

- [x] Create `pkg/lsp/provider/completion.go`
  - Implement `CompletionProvider` interface using CursorContext
  - Use `ctx.InClassAttr` to trigger tailwind class completions via `schema.MatchClass()`
  - Use `ctx.InGoExpr` to delegate to gopls for Go expression completions
  - Use `ctx.InElement` to offer attribute completions from `schema.GetElement()`
  - Offer event handler attribute completions from `schema.EventHandlers`
  - Offer DSL keyword completions from `schema.GetKeyword()` for `@for`, `@if`, `@let`
  - Offer element tag completions from `schema.Elements`
  - Offer component call completions from ComponentIndex
  - Port all existing completion.go logic refactored to use CursorContext and Schema
  - Trigger characters remain: `@`, `<`, `{`

- [x] Create `pkg/lsp/provider/symbols.go`
  - Implement `DocumentSymbolProvider` interface
  - Walk document AST to produce two-level symbol hierarchy:
    - Level 1: Components and functions (SymbolKindFunction)
    - Level 2: Let bindings (SymbolKindVariable), ID'd elements (SymbolKindField)
  - Implement `WorkspaceSymbolProvider` interface
  - Query ComponentIndex for workspace-wide symbol search (case-insensitive)
  - Port existing symbols.go logic

- [x] Wire up providers in `pkg/lsp/router.go`
  - Register CompletionProvider, DocumentSymbolProvider, WorkspaceSymbolProvider
  - Update routes to resolve CursorContext (for completion) and dispatch to providers

- [x] Remove old files: `pkg/lsp/completion.go`, `pkg/lsp/symbols.go`
  - Handler methods stripped; type definitions and helper methods retained (used by features_test.go)

- [x] Create `pkg/lsp/provider/completion_test.go`
  - Test class attribute completion (prefix matching, category sorting)
  - Test element tag completion
  - Test attribute completion within element
  - Test keyword completion
  - Test component call completion
  - Test gopls delegation for Go expressions

- [x] Create `pkg/lsp/provider/symbols_test.go`
  - Test document symbols: components, functions, let bindings, ID'd elements
  - Test workspace symbols: case-insensitive name matching

**Tests:** Run `go test ./pkg/lsp/...` once at phase end ✅

---

## Phase 4: Migrate Semantic Tokens + Diagnostics + Formatting Providers

**Reference:** [devtools-overhaul-design.md §2](./devtools-overhaul-design.md#2-architecture)

**Completed in commit:** (pending)

- [x] Create `pkg/lsp/provider/semantic.go`
  - Implement `SemanticTokensProvider` interface
  - Define token type and modifier constants as named constants (not magic numbers)
  - Port the 13 token types and 4 modifiers from current semantic_tokens.go
  - Walk document AST to emit tokens for: package/import (namespace), component declarations (class), function declarations (function), parameters (parameter), variables (variable), attributes (property), keywords (keyword), strings (string), numbers (number), operators (operator), component call decorators (decorator), format specifiers (regexp), comments (comment)
  - Maintain per-component and per-loop context for identifier classification
  - Use `schema.GetEventHandler()` to distinguish event attributes in token output
  - Port existing semantic_tokens.go logic (~1,227 lines) with improved constant usage

- [x] Create `pkg/lsp/provider/diagnostics.go`
  - Implement `DiagnosticsProvider` interface
  - Convert tuigen parse errors and semantic errors to LSP Diagnostic format
  - Map tuigen positions (1-indexed) to LSP positions (0-indexed)
  - Set severity to Error, source to "gsx"
  - Port existing diagnostics.go logic

- [x] Create `pkg/lsp/provider/formatting.go`
  - Implement `FormattingProvider` interface
  - Delegate to existing formatter package
  - Respect tabSize and insertSpaces options
  - Return single TextEdit covering entire document, or empty if no changes
  - Port existing formatting.go logic

- [x] Wire up providers in `pkg/lsp/router.go`
  - Register SemanticTokensProvider, DiagnosticsProvider, FormattingProvider
  - Update routes to dispatch to providers
  - Ensure semantic token legend in initialize response matches the named constants

- [x] Remove old files: `pkg/lsp/semantic_tokens.go`, `pkg/lsp/diagnostics.go`, `pkg/lsp/formatting.go`
  - Old type definitions replaced with type aliases; handler.go retained for initialize response

- [x] Remove old handler.go: `pkg/lsp/handler.go`
  - Handler.go still provides the initialize response with capabilities; routing is in router.go

- [x] Create `pkg/lsp/provider/semantic_test.go`
  - Test token output for: component declarations, function declarations, keywords, parameters, variables, strings, numbers, attributes, component calls, comments
  - Test per-component context (parameters recognized as known identifiers)
  - Test format specifier highlighting in strings
  - Verify token type constants match legend order

- [x] Create `pkg/lsp/provider/diagnostics_test.go`
  - Test error position mapping from tuigen to LSP format
  - Test multiple errors in a single document

**Tests:** Run `go test ./pkg/lsp/...` once at phase end

---

## Phase 5: Refs/State/Events Awareness + gopls Updates

**Reference:** [devtools-overhaul-design.md §3](./devtools-overhaul-design.md#3-core-entities)

**Completed in commit:** (pending)

- [x] Update `pkg/lsp/context.go` for new construct resolution
  - Add detection for `#Name` named refs: classify as NodeKindNamedRef, populate Scope.NamedRefs
  - Add detection for `tui.NewState()` declarations: classify as NodeKindStateDecl, populate Scope.StateVars
  - Add detection for `.Get()`, `.Set()`, `.Update()`, `.Bind()` on state vars: classify as NodeKindStateAccess
  - Add detection for event handler attributes (onClick, onFocus, etc.): classify as NodeKindEventHandler
  - Extend scope resolution to collect all named refs and state vars in the enclosing component

- [x] Update `pkg/lsp/gopls/generate.go` for refs and state
  - Emit state variable declarations in virtual Go: `count := tui.NewState(0)` as proper Go with correct types
  - Emit named ref variable declarations: `var Header *element.Element` for simple refs, `var Items []*element.Element` for loop refs, `var Users map[string]*element.Element` for keyed refs
  - Add source mappings for state and ref variable positions so gopls responses can be translated back to .gsx positions
  - Ensure gopls can resolve `.Get()`, `.Set()` calls on emitted state variables

- [x] Update `pkg/lsp/provider/hover.go` for new constructs
  - NodeKindNamedRef: show ref name, type (`*element.Element`), context (simple/loop/keyed/conditional), access pattern (`view.Header`, `view.Items[i]`, `view.Users[key]`)
  - NodeKindStateDecl: show state var name, type (`*tui.State[T]`), initial value, available methods
  - NodeKindStateAccess: show method documentation (Get returns T, Set takes T, etc.)
  - NodeKindEventHandler: show handler name, expected signature (`func()`), description from schema

- [x] Update `pkg/lsp/provider/completion.go` for new constructs
  - After state variable name + `.`: suggest Get(), Set(v), Update(fn), Bind(fn), Batch(fn) with documentation
  - In element attribute position: suggest event handlers (onClick, onFocus, onBlur, onKeyPress, onEvent) from schema
  - After `#` in element tag: suggest ref name (no strong completions needed, but don't break)

- [x] Update `pkg/lsp/provider/definition.go` for new constructs
  - NodeKindNamedRef usage in Go expression → jump to `#Name` declaration on the element
  - NodeKindStateAccess (e.g., `count.Get()`) → jump to `count := tui.NewState(0)` declaration
  - NodeKindEventHandler → jump to handler function definition

- [x] Update `pkg/lsp/provider/references.go` for new constructs
  - Named ref: find `#Name` declaration + all usages in Go expressions and handler arguments
  - State var: find `tui.NewState()` declaration + all `.Get()`, `.Set()`, `.Update()`, `.Bind()` calls + handler argument usages

- [x] Update `pkg/lsp/provider/semantic.go` for new constructs
  - Add token handling for `#` punctuation (operator or punctuation type)
  - Add token handling for ref names after `#` (variable type with declaration modifier)
  - Highlight state variable declarations with variable type + declaration modifier
  - Highlight state method calls (`.Get()`, `.Set()`) distinctly
  - Distinguish event handler attributes from regular attributes using schema lookup

- [x] Create/update tests for new construct awareness
  - `pkg/lsp/context_test.go`: add test cases for cursor on `#Name`, `tui.NewState()`, `.Get()`, `onClick`
  - `pkg/lsp/gopls/generate_test.go`: test virtual Go output includes state vars and ref vars with correct types
  - `pkg/lsp/provider/hover_test.go`: add cases for ref hover, state hover, event handler hover
  - `pkg/lsp/provider/completion_test.go`: add cases for state method completion, event attribute completion
  - `pkg/lsp/provider/definition_test.go`: add cases for ref definition, state definition
  - `pkg/lsp/provider/references_test.go`: add cases for ref references, state references
  - `pkg/lsp/provider/semantic_test.go`: add cases for ref tokens, state tokens, event tokens

**Tests:** Run `go test ./pkg/lsp/...` once at phase end ✅

---

## Phase 6: VSCode Extension + Tree-sitter Updates

**Reference:** [devtools-overhaul-design.md §4](./devtools-overhaul-design.md#4-user-experience)

**Completed in commit:** (pending)

- [x] Update `editor/vscode/syntaxes/gsx.tmLanguage.json`
  - Change all `.tui` scope suffixes to `.gsx` in named-ref patterns (meta.named-ref, punctuation.definition.named-ref, entity.name.tag.named-ref)
  - Add scope patterns for state declarations: `tui.NewState(...)` with `entity.name.function.gsx` for NewState
  - Add scope patterns for state method calls: `.Get()`, `.Set()`, `.Update()`, `.Bind()` with `entity.name.function.gsx`
  - Verify event handler attributes (onClick, onFocus, etc.) are highlighted as attributes
  - Verify `templ` keyword is highlighted (not `@component`)
  - Test all patterns against the test files

- [x] Update `editor/vscode/language-configuration.json`
  - Change folding start pattern from `@component` to `templ`: update `"start": "^\\s*(@component|@for|@if|@let|\\{)"` to `"start": "^\\s*(templ|@for|@if|@let|\\{)"`
  - Verify indentation rules work with `templ` keyword
  - Verify auto-closing pairs and surrounding pairs are complete

- [x] Update `editor/vscode/README.md`
  - Replace all `<box>` references with `<div>`
  - Replace all `<text>` references with `<span>`
  - Replace `@component` with `templ` throughout
  - Add documentation for named refs (`#Name` syntax)
  - Add documentation for state (`tui.NewState()`, `.Get()`, `.Set()`)
  - Add documentation for event handlers (onClick, onFocus, etc.)
  - Update the "Supported Constructs" table
  - Remove "coming soon" note about LSP configuration (it's already implemented)
  - Update examples to use current syntax

- [x] Update `editor/vscode/test/simple.gsx`
  - Update to use current syntax: `templ` instead of any old keywords
  - Add examples of `#Name` refs
  - Add state variable usage

- [x] Update `editor/vscode/test/complex.gsx`
  - Update to use current syntax throughout
  - Add named ref examples (simple, loop, keyed, conditional)
  - Add state variable declaration and usage
  - Add event handler attribute examples
  - Ensure all language constructs are represented for grammar testing

- [x] Update `editor/tree-sitter-gsx/grammar.js`
  - Verify state declaration syntax is parseable: `name := tui.NewState(initialValue)`
  - Verify event handler attributes parse correctly: `onClick={handler}`
  - Add any missing constructs identified during testing
  - Ensure `key={expr}` attribute is handled in element parsing

- [x] Update `editor/tree-sitter-gsx/queries/highlights.scm`
  - Add highlight rules for state declarations if not present
  - Add highlight rules for state method calls (.Get, .Set, etc.)
  - Verify event handler attributes get appropriate highlights
  - Verify named ref `#Name` highlights are correct

- [x] Update `editor/tree-sitter-gsx/test/corpus/basic.txt`
  - Add test case for state declaration
  - Add test case for event handler attributes
  - Add test case for named refs with different contexts (simple, loop, keyed)

**Tests:** Run `tree-sitter test` in `editor/tree-sitter-gsx/`, manually verify VSCode grammar with test files

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Core Infrastructure (Schema + CursorContext + Provider Interfaces + Router) | Complete |
| 2 | Migrate Navigation Providers (Hover + Definition + References) | Complete |
| 3 | Migrate Completion + Symbol Providers | Complete |
| 4 | Migrate Semantic Tokens + Diagnostics + Formatting Providers | Complete |
| 5 | Refs/State/Events Awareness + gopls Updates | Complete |
| 6 | VSCode Extension + Tree-sitter Updates | Complete |

## Files to Create

```
pkg/lsp/
├── schema/
│   ├── schema.go
│   ├── schema_test.go
│   ├── keywords.go
│   └── tailwind.go
├── provider/
│   ├── provider.go
│   ├── hover.go
│   ├── hover_test.go
│   ├── completion.go
│   ├── completion_test.go
│   ├── definition.go
│   ├── definition_test.go
│   ├── references.go
│   ├── references_test.go
│   ├── symbols.go
│   ├── symbols_test.go
│   ├── diagnostics.go
│   ├── diagnostics_test.go
│   ├── formatting.go
│   ├── semantic.go
│   └── semantic_test.go
├── context.go
├── context_test.go
└── router.go
```

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/lsp/server.go` | Simplify to JSON-RPC I/O + lifecycle; delegate routing to router.go |
| `pkg/lsp/gopls/generate.go` | Emit state var and named ref declarations in virtual Go |

## Files to Delete

| File | Replaced By |
|------|-------------|
| `pkg/lsp/handler.go` | `pkg/lsp/router.go` |
| `pkg/lsp/hover.go` | `pkg/lsp/provider/hover.go` |
| `pkg/lsp/completion.go` | `pkg/lsp/provider/completion.go` |
| `pkg/lsp/definition.go` | `pkg/lsp/provider/definition.go` |
| `pkg/lsp/references.go` | `pkg/lsp/provider/references.go` |
| `pkg/lsp/symbols.go` | `pkg/lsp/provider/symbols.go` |
| `pkg/lsp/diagnostics.go` | `pkg/lsp/provider/diagnostics.go` |
| `pkg/lsp/formatting.go` | `pkg/lsp/provider/formatting.go` |
| `pkg/lsp/semantic_tokens.go` | `pkg/lsp/provider/semantic.go` |

## Editor Files to Modify

| File | Changes |
|------|---------|
| `editor/vscode/syntaxes/gsx.tmLanguage.json` | Fix `.tui` → `.gsx` scopes, add state/event patterns |
| `editor/vscode/language-configuration.json` | Fix `@component` → `templ` in folding |
| `editor/vscode/README.md` | Update for current syntax, add refs/state/events docs |
| `editor/vscode/test/simple.gsx` | Update to current syntax |
| `editor/vscode/test/complex.gsx` | Add refs, state, events examples |
| `editor/tree-sitter-gsx/grammar.js` | Verify/add state and event constructs |
| `editor/tree-sitter-gsx/queries/highlights.scm` | Add state/event highlight rules |
| `editor/tree-sitter-gsx/test/corpus/basic.txt` | Add state, events, refs test cases |
