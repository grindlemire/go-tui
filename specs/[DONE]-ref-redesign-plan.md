# Explicit Refs & Handler Self-Inject Implementation Plan

Implementation phases for the ref redesign. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Core Types & Handler Self-Inject (tui package) ✅

**Reference:** [ref-redesign-design.md §3](./ref-redesign-design.md#3-core-entities)

**Completed in commit:** (pending review)

- [x] Create `ref.go` — new ref types in root `tui` package
  - `Ref` struct with `Set(*Element)`, `El() *Element`, `IsSet() bool`
  - `RefList` struct with `Append(*Element)`, `All() []*Element`, `At(i int) *Element`, `Len() int`
  - `RefMap[K comparable]` struct with `Put(K, *Element)`, `Get(K) *Element`, `All() map[K]*Element`, `Len() int`
  - `NewRef()`, `NewRefList()`, `NewRefMap[K]()` constructors
  - All methods thread-safe with `sync.RWMutex`
  - See [ref-redesign-design.md §3.1](./ref-redesign-design.md#31-ref-types-new-tuirefgo)

- [x] Create `ref_test.go` — tests for all ref types
  - `TestRef_SetAndGet` — basic set/get round-trip
  - `TestRef_IsSet` — nil before set, true after
  - `TestRef_NilBeforeSet` — `El()` returns nil before `Set()`
  - `TestRefList_AppendAndAll` — append multiple, retrieve all
  - `TestRefList_At` — valid index, out-of-bounds returns nil
  - `TestRefList_Len` — count tracking
  - `TestRefMap_PutAndGet` — key-based store/retrieve
  - `TestRefMap_All` — full map retrieval
  - `TestRefMap_GetMissing` — missing key returns nil
  - `TestRefMap_Len` — count tracking
  - Use table-driven test pattern with `tc` struct

- [x] Modify `element.go` — change handler field signatures
  - `onKeyPress func(KeyEvent)` → `onKeyPress func(*Element, KeyEvent)`
  - `onClick func()` → `onClick func(*Element)`
  - `onEvent func(Event) bool` → `onEvent func(*Element, Event) bool`
  - `onFocus func()` → `onFocus func(*Element)`
  - `onBlur func()` → `onBlur func(*Element)`
  - Leave `onRender`, `onChildAdded`, `onFocusableAdded`, `onUpdate` unchanged
  - See [ref-redesign-design.md §3.2](./ref-redesign-design.md#32-changed-handler-signatures)

- [x] Modify `element_options.go` — update `WithOn*` option signatures
  - `WithOnKeyPress(fn func(KeyEvent))` → `WithOnKeyPress(fn func(*Element, KeyEvent))`
  - `WithOnClick(fn func())` → `WithOnClick(fn func(*Element))`
  - `WithOnEvent(fn func(Event) bool)` → `WithOnEvent(fn func(*Element, Event) bool)`
  - `WithOnFocus(fn func())` → `WithOnFocus(fn func(*Element))`
  - `WithOnBlur(fn func())` → `WithOnBlur(fn func(*Element))`

- [x] Modify `element_focus.go` — update `Set*` methods and event dispatch
  - Update `SetOnKeyPress`, `SetOnClick`, `SetOnEvent`, `SetOnFocus`, `SetOnBlur` signatures
  - Update `HandleEvent()` dispatch to pass `e` (self) as first arg to all handler calls
  - Update `Focus()` to pass `e` to `onFocus(e)`
  - Update `Blur()` to pass `e` to `onBlur(e)`

- [x] Update all handler-related test files for new signatures
  - `element_tree_test.go` — `HandleEvent` tests with `SetOnEvent` lambdas
  - `element_options_test.go` — any `WithOn*` usage
  - `focus_test.go` — `HandleEvent` mocks
  - `focus_dispatch_test.go` — focus transition handler lambdas
  - `element_accessors_test.go` — if any handler references

**Tests:** Run `go test ./...` at root — all root package tests must pass

---

## Phase 2: Compiler Pipeline (internal/tuigen) ✅

**Reference:** [ref-redesign-design.md §6](./ref-redesign-design.md#6-detailed-design)

**Completed in commit:** (pending review)

- [x] Modify `internal/tuigen/token.go` — remove `TokenHash`
  - Remove `TokenHash` from token type enum
  - See [ref-redesign-design.md §6.1](./ref-redesign-design.md#61-lexer-changes-internaltuigentokengo-lexergo)

- [x] Modify `internal/tuigen/lexer.go` — remove `#` handling
  - Remove the `case '#'` branch from the lexer scan switch
  - `#` in element position becomes a syntax error

- [x] Modify `internal/tuigen/ast.go` — replace `NamedRef` with `RefExpr`
  - Remove `NamedRef string` field from `Element` struct
  - Add `RefExpr *GoExpr` field (expression from `ref={expr}`)
  - Keep `RefKey *GoExpr` field unchanged
  - See [ref-redesign-design.md §3.3](./ref-redesign-design.md#33-ast-changes-internaltuigenastgo)

- [x] Modify `internal/tuigen/parser_element.go` — remove `#Name` parsing
  - Remove the block that detects `TokenHash` after tag name and populates `NamedRef`
  - No new parser changes needed — `ref={}` is parsed as a regular attribute

- [x] Modify `internal/tuigen/analyzer.go` — replace `NamedRef` struct with `RefInfo`
  - Add `RefInfo` struct with `Name`, `ExportName`, `RefKind`, etc.
  - Add `RefKind` enum: `RefSingle`, `RefList`, `RefMap`
  - Update `NamedRef` references throughout to use `RefInfo`
  - Add `"ref"` to `knownAttributes` map
  - See [ref-redesign-design.md §3.4](./ref-redesign-design.md#34-analyzer-refinfo-internaltuigenanalyzergo)

- [x] Rewrite `internal/tuigen/analyzer_refs.go` — new ref validation
  - Replace `validateNamedRefs` with `validateRefs`
  - Scan element attributes for `ref={expr}`, extract and store in `elem.RefExpr`
  - Remove `ref` from attribute list (like `key` is extracted today)
  - Validate ref expression is a simple identifier
  - Determine ref kind from context (loop → RefList/RefMap, else → RefSingle)
  - Capitalize variable name for `ExportName` (View struct field)
  - Remove `isValidRefName` uppercase requirement — ref names are normal Go variables
  - Keep duplicate name and reserved name ("Root") validation
  - See [ref-redesign-design.md §6.3](./ref-redesign-design.md#63-analyzer-changes-internaltuigenanalyzer_refsgo)

- [x] Modify `internal/tuigen/generator_element.go` — handlers as options, ref binding
  - Change `handlerAttributes` map values from `"SetOnKeyPress"` → `"tui.WithOnKeyPress"` etc.
  - Move handler attribute processing from `handlers` collection to `options` (inline `With*` options)
  - Remove `deferredHandler` struct and handler field from `elementOptions`
  - After element creation, if `elem.RefExpr` is set, emit binding call:
    - `RefSingle`: `content.Set(__tui_3)`
    - `RefList`: `items.Append(__tui_5)`
    - `RefMap`: `users.Put(key, __tui_5)`
  - Keep deferred watcher handling unchanged
  - See [ref-redesign-design.md §6.4](./ref-redesign-design.md#64-generator-changes-internaltuigengenerator_componentgo-generator_elementgo)

- [x] Modify `internal/tuigen/generator_component.go` — remove forward decls, update view struct
  - Remove the block that forward-declares `var Content *tui.Element` / `var Items []*tui.Element`
  - Remove `g.deferredHandlers` field, deferred handler emission block
  - Update view struct generation to use `RefInfo` with `ExportName`
  - View struct field resolution: `content.El()`, `items.All()`, `users.All()`
  - Update `generateViewStruct` to use `RefInfo` instead of `NamedRef`

- [x] Modify `internal/tuigen/generator.go` — clean up deferred handler types
  - Remove `deferredHandler` struct definition
  - Remove `deferredHandlers []deferredHandler` field from Generator
  - Keep `deferredWatcher` struct and `deferredWatchers` field

- [x] Modify `internal/formatter/printer_elements.go` — remove `#Name` formatting, add `ref={}` output
  - Remove `NamedRef` handling in element printing (multiline and single-line modes)
  - Added `RefExpr` output in both single-line and multi-line modes

- [x] Modify `internal/formatter/printer_control.go` — remove any `NamedRef` references, add `ref={}` output
  - Removed `NamedRef` references in control structure formatting
  - Added `RefExpr` output in let binding element printing

- [x] Update compiler test files
  - `internal/tuigen/parser_element_test.go` — remove `#Name` tests, add `ref={}` attribute tests
  - `internal/tuigen/analyzer_refs_test.go` — rewrite for `ref={}` validation (lowercase names, ref kinds, duplicate detection)
  - `internal/tuigen/generator_test.go` — update generated output expectations for `With*` options and ref binding
  - `internal/formatter/formatter_test.go` — update 3 formatter tests from `#Name` to `ref={}` syntax

**Tests:** Run `go test ./internal/tuigen/... ./internal/formatter/...` — all compiler + formatter tests must pass

---

## Phase 3: Editor Support (tree-sitter + VSCode) ✅

**Reference:** [ref-redesign-design.md §6.7](./ref-redesign-design.md#67-editor-changes)

**Completed in commit:** (pending review)

- [x] Modify `editor/tree-sitter-gsx/grammar.js` — remove `named_ref` rule
  - Remove `named_ref: $ => seq('#', $.identifier)` rule
  - Remove `optional(field('named_ref', $.named_ref))` from `self_closing_element`
  - Remove `optional(field('named_ref', $.named_ref))` from `element_with_children`
  - `ref={expr}` handled by existing `attribute` rule automatically

- [x] Regenerate tree-sitter parser
  - Run `tree-sitter generate` in `editor/tree-sitter-gsx/`
  - This regenerates `src/parser.c`, `src/grammar.json`, `src/node-types.json`

- [x] Modify `editor/tree-sitter-gsx/queries/highlights.scm` — update highlighting
  - Remove `named_ref` highlighting block (`#` + identifier pattern)
  - Optionally add `ref` attribute name highlight with special scope

- [x] Update `editor/tree-sitter-gsx/test/corpus/basic.txt` — update test cases
  - Update "Named ref on element" test → "Ref attribute on element" with `ref={name}`
  - Update "Named ref on self-closing element" test
  - Update "Element with named ref and attributes" test
  - Ensure all updated tests match expected parse tree output

- [x] Modify `editor/vscode/syntaxes/gsx.tmLanguage.json` — update syntax highlighting
  - Remove `named-ref` pattern (the `(#)([A-Z][a-zA-Z0-9_]*)` match)
  - Remove `#named-ref` include from `element-open-tag` patterns
  - Add `ref` attribute pattern to `attributes` section with appropriate scopes

- [x] Update VSCode test files (if any reference `#Name`)
  - Check `editor/vscode/test/simple.gsx` and `editor/vscode/test/complex.gsx`
  - Replace `#Name` with `ref={name}` syntax

**Tests:** Run `tree-sitter test` in `editor/tree-sitter-gsx/` — all corpus tests must pass

---

## Phase 4: LSP Updates (internal/lsp)

**Reference:** [ref-redesign-design.md §6.6](./ref-redesign-design.md#66-lsp-changes)

**Completed in commit:** (pending review)

- [x] Modify `internal/lsp/context.go` — remove `NodeKindNamedRef`
  - Remove `NodeKindNamedRef` from `NodeKind` enum
  - Add new node kind for ref attribute value if needed (or reuse existing attribute kinds)
  - Update `Scope` struct: remove or rename `NamedRefs` field

- [x] Modify `internal/lsp/context_resolve.go` — detect `ref={}` instead of `#Name`
  - Remove logic that detects cursor on `#Name` syntax
  - Add logic to detect cursor on `ref={ident}` attribute value
  - Return appropriate node kind for ref attribute context

- [x] Modify `internal/lsp/schema/schema.go` — add `ref` attribute
  - Add `ref` to generic element attributes with type `"expression"` and description

- [x] Modify `internal/lsp/provider/definition.go` — ref definition navigation
  - Remove `definitionNamedRef()` and `definitionNamedRefFromScope()` functions
  - Add logic to navigate from `ref={content}` to the variable declaration site
  - Support navigating from ref usage in Go expressions to ref declaration

- [x] Modify `internal/lsp/provider/references.go` — ref reference finding
  - Remove `NodeKindNamedRef` case handling
  - Add logic to find `ref={identifier}` in element attributes
  - Find the ref variable declaration and all usages

- [x] Modify `internal/lsp/provider/hover.go` — ref hover info
  - Remove `NodeKindNamedRef` case
  - Add hover info for `ref={content}`: "Element reference — binds this element to the `content` ref variable"

- [x] Modify `internal/lsp/provider/semantic_nodes.go` — remove `#Name` tokens
  - Remove semantic token generation for `#` punctuation and named ref identifier
  - Add semantic token for `ref` attribute value (variable reference highlighting)

- [x] Modify `internal/lsp/provider_adapters.go` — update named ref handling
  - Update any adapter code that translates between `NamedRef` and LSP types
  - Use `RefInfo` / new ref attribute representation

- [x] Modify `internal/lsp/gopls/generate.go` — update virtual Go generation
  - Remove named ref variable generation for gopls
  - Update to generate ref variable declarations if needed for type inference

- [x] Update LSP test files
  - `internal/lsp/context_test.go` — update cursor context tests for `ref={}` syntax
  - `internal/lsp/context_scope_test.go` — update scope resolution tests
  - `internal/lsp/provider/definition_test.go` — update go-to-definition tests
  - `internal/lsp/provider/references_test.go` — update find-references tests
  - `internal/lsp/provider/hover_test.go` — update hover tests
  - `internal/lsp/provider/semantic_nodes_test.go` — update semantic token tests
  - `internal/lsp/gopls/generate_test.go` — update virtual Go generation tests

**Tests:** Run `go test ./internal/lsp/...` — all LSP tests must pass

---

## Phase 5: Examples & Integration

**Reference:** [ref-redesign-design.md §4](./ref-redesign-design.md#4-user-experience)

**Review:** false

**Completed in commit:** (pending review)

- [x] Update `examples/08-focus/focus.gsx`
  - Removed `#BoxA`, `#BoxB`, `#BoxC` — self-inject eliminates need for refs
  - Updated `onFocusBox`/`onBlurBox` to use self-inject `func(*tui.Element)` signature

- [x] Update `examples/09-scrollable/scrollable.gsx`
  - Removed `#Content` — self-inject eliminates need for ref
  - Converted `handleScrollKeys` and `handleMouseScroll` to plain functions with self-inject

- [x] Update `examples/10-refs/refs.gsx`
  - Added `tui.NewRef()` for `counter`, `incrementBtn`, `decrementBtn`, `status`
  - Replaced `#Name` with `ref={name}` attributes
  - Updated `handleIncrement`/`handleDecrement` to `func(*tui.Element)` signature

- [x] Update `examples/10-refs/main.go`
  - Updated doc comments

- [x] Update `examples/11-streaming/streaming.gsx`
  - Added `content := tui.NewRef()` for cross-element access
  - Replaced `#Content` with `ref={content}`
  - Converted `handleScrollKeys` to self-inject, `addLine` to use `*tui.Ref`

- [x] Update `examples/refs-demo/refs.gsx`
  - Added `tui.NewRef()` for `header`, `content`, `warning`, `statusBar`
  - Added `tui.NewRefList()` for `itemRefs` and `userRefs`
  - Replaced all `#Name` with `ref={name}` attributes

- [x] Update `examples/refs-demo/main.go`
  - Updated view struct field access: `Items` → `ItemRefs`, `Users` → `UserRefs`
  - Updated doc comments

- [x] Update `examples/streaming-dsl/streaming.gsx`
  - Added `content := tui.NewRef()` for cross-element access
  - Converted `handleScrollKeys`/`handleEvent` to self-inject
  - Updated `addLine` to use `*tui.Ref`

- [x] Update `examples/streaming-dsl/main.go`
  - Updated doc comments

- [x] Update additional examples with handler signature changes
  - `examples/06-interactive/interactive.gsx` — updated onClick/onKeyPress handler return types
  - `examples/07-keyboard/keyboard.gsx` — updated onKeyPress handler return type
  - `examples/counter-state/counter.gsx` — updated onClick/onKeyPress handler return types
  - `examples/focus/main.go` — updated WithOnFocus/WithOnBlur inline closures
  - `examples/state/state_tui.go` — updated handler signatures in generated file

- [x] Regenerate all `*_gsx.go` files
  - Built CLI and regenerated all 17 `.gsx` files
  - Verified generated files use new patterns: `With*` inline options, `ref.Set()`, `ref.El()`

- [x] Run full test suite and verify examples build
  - `go test ./...` — all tests pass
  - `go build ./examples/...` — all examples build successfully

**Tests:** Run `go test ./...` — full project test suite must pass. All examples must build.

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Core types (Ref, RefList, RefMap) + handler self-inject in tui package | ✅ Complete |
| 2 | Compiler pipeline: lexer, parser, AST, analyzer, generator, formatter | ✅ Complete |
| 3 | Editor support: tree-sitter grammar + VSCode syntax highlighting | ✅ Complete |
| 4 | LSP updates: context, providers, schema, gopls | ✅ Complete |
| 5 | Update all 6 examples, regenerate, full integration test | ✅ Complete |

## Files to Create

```
ref.go                    # Ref, RefList, RefMap types
ref_test.go               # Ref type tests
```

## Files to Modify

| File | Changes |
|------|---------|
| `element.go` | Handler field signatures gain `*Element` |
| `element_options.go` | `WithOn*` option signatures gain `*Element` |
| `element_focus.go` | `Set*` methods + `HandleEvent` dispatch pass self |
| `element_tree_test.go` | Update handler lambdas |
| `element_options_test.go` | Update `WithOn*` usage |
| `focus_test.go` | Update `HandleEvent` mocks |
| `focus_dispatch_test.go` | Update focus handler lambdas |
| `internal/tuigen/token.go` | Remove `TokenHash` |
| `internal/tuigen/lexer.go` | Remove `#` handling |
| `internal/tuigen/ast.go` | `NamedRef` → `RefExpr` |
| `internal/tuigen/parser_element.go` | Remove `#Name` parsing |
| `internal/tuigen/analyzer.go` | `NamedRef` → `RefInfo` struct |
| `internal/tuigen/analyzer_refs.go` | Rewrite ref validation |
| `internal/tuigen/generator.go` | Remove deferred handler types |
| `internal/tuigen/generator_element.go` | Handlers as options, ref binding |
| `internal/tuigen/generator_component.go` | Remove forward decls, update view struct |
| `internal/tuigen/parser_element_test.go` | Update parser tests |
| `internal/tuigen/analyzer_refs_test.go` | Rewrite ref validation tests |
| `internal/tuigen/generator_test.go` | Update generated output tests |
| `internal/formatter/printer_elements.go` | Remove `NamedRef` formatting |
| `internal/formatter/printer_control.go` | Remove `NamedRef` references |
| `internal/lsp/context.go` | Remove `NodeKindNamedRef` |
| `internal/lsp/context_resolve.go` | Detect `ref={}` instead of `#Name` |
| `internal/lsp/schema/schema.go` | Add `ref` attribute |
| `internal/lsp/provider/definition.go` | Ref definition navigation |
| `internal/lsp/provider/references.go` | Ref reference finding |
| `internal/lsp/provider/hover.go` | Ref hover info |
| `internal/lsp/provider/semantic_nodes.go` | Remove `#Name` tokens |
| `internal/lsp/provider_adapters.go` | Update ref handling |
| `internal/lsp/gopls/generate.go` | Update virtual Go generation |
| `internal/lsp/context_test.go` | Update context tests |
| `internal/lsp/context_scope_test.go` | Update scope tests |
| `internal/lsp/provider/definition_test.go` | Update definition tests |
| `internal/lsp/provider/references_test.go` | Update references tests |
| `internal/lsp/provider/hover_test.go` | Update hover tests |
| `internal/lsp/provider/semantic_nodes_test.go` | Update semantic tests |
| `internal/lsp/gopls/generate_test.go` | Update generate tests |
| `editor/tree-sitter-gsx/grammar.js` | Remove `named_ref` rule |
| `editor/tree-sitter-gsx/queries/highlights.scm` | Update highlighting |
| `editor/tree-sitter-gsx/test/corpus/basic.txt` | Update test cases |
| `editor/vscode/syntaxes/gsx.tmLanguage.json` | Remove `#Name`, add `ref={}` |
| `examples/08-focus/focus.gsx` | Replace `#Name` with `ref={}` |
| `examples/09-scrollable/scrollable.gsx` | Replace `#Name` with `ref={}` |
| `examples/10-refs/refs.gsx` | Replace `#Name` with `ref={}` |
| `examples/10-refs/main.go` | Update view struct access |
| `examples/11-streaming/streaming.gsx` | Replace `#Name` with `ref={}` |
| `examples/refs-demo/refs.gsx` | Replace `#Name` with `ref={}` |
| `examples/refs-demo/main.go` | Update view struct access |
| `examples/streaming-dsl/streaming.gsx` | Replace `#Name` with `ref={}` |
| `examples/streaming-dsl/main.go` | Update view struct access |
