# GSX Syntax Implementation Plan

Implementation phases for GSX syntax changes. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Core Syntax Changes (Compiler & CLI)

**Reference:** [gsx-syntax-design.md §4](./gsx-syntax-design.md#4-core-entities)

**Status:** ✅ Complete

- [x] Modify `pkg/tuigen/token.go`
  - Remove `TokenAtComponent` from the `TokenType` constants
  - Remove `TokenAtComponent: "@component"` from `tokenNames` map
  - See [gsx-syntax-design.md §4.1](./gsx-syntax-design.md#41-token-changes)

- [x] Modify `pkg/tuigen/lexer.go`
  - Remove `case "component":` branch from `scanAtKeyword()` function
  - See [gsx-syntax-design.md §4.2](./gsx-syntax-design.md#42-lexer-changes)

- [x] Modify `pkg/tuigen/parser.go`
  - Update `synchronize()` to remove `TokenAtComponent` check (line ~88)
  - Update `Parse()` switch to handle `TokenFunc` for both components and Go functions
  - Create new `parseFuncOrComponent()` method that:
    - Parses `func` keyword, name, parameters
    - Checks return type - if exactly `Element`, parse as component DSL
    - Otherwise delegate to `parseGoFunc()`
  - Remove or repurpose existing `parseComponent()` method
  - See [gsx-syntax-design.md §4.4](./gsx-syntax-design.md#44-parser-changes)

- [x] Modify `pkg/formatter/printer.go`
  - Update `printComponent()` to emit `func Name(params) Element {` instead of `@component Name(params) {`
  - See [gsx-syntax-design.md §4.5](./gsx-syntax-design.md#45-formatter-changes)

- [x] Modify `cmd/tui/generate.go`
  - Change file pattern from `.tui` to `.gsx` in `collectTuiFiles()` (rename to `collectGsxFiles()`)
  - Update `outputFileName()` to produce `*_gsx.go` instead of `*_tui.go`
  - Update comments and error messages
  - See [gsx-syntax-design.md §4.6](./gsx-syntax-design.md#46-cli-changes)

- [x] Modify `cmd/tui/check.go`
  - Change file pattern from `.tui` to `.gsx`
  - Update comments and error messages

- [x] Modify `cmd/tui/fmt.go`
  - Change file pattern from `.tui` to `.gsx`
  - Update comments

- [x] Modify `cmd/tui/main.go`
  - Update usage help text to reference `.gsx` instead of `.tui`

- [x] Rename test fixtures in `cmd/tui/testdata/`
  - `simple.tui` → `simple.gsx` (convert syntax)
  - `complex.tui` → `complex.gsx` (convert syntax)
  - `other.tui` → `other.gsx` (convert syntax)
  - `simple_tui.go` → `simple_gsx.go` (update source comment)
  - `complex_tui.go` → `complex_gsx.go` (update source comment)
  - `other_tui.go` → `other_gsx.go` (update source comment)

- [x] Update test files in `pkg/tuigen/`
  - `lexer_test.go` - remove/update `@component` test cases
  - `parser_test.go` - update to use `func ... Element` syntax
  - `parser_comment_test.go` - update component syntax in test cases
  - `analyzer_test.go` - update component syntax in test cases
  - `generator_test.go` - update component syntax in test cases

- [x] Update test files in `pkg/formatter/`
  - `formatter_test.go` - update to use `func ... Element` syntax
  - `formatter_comment_test.go` - update component syntax in test cases

- [x] Update test files in `pkg/lsp/` (needed for tests to pass)
  - `server_test.go` - update to use `func ... Element` syntax
  - `features_test.go` - update to use `func ... Element` syntax
  - `semantic_tokens_comment_test.go` - update to use `func ... Element` syntax

**Tests:** Run `go test ./pkg/tuigen/... ./cmd/tui/...` at phase end ✅ All pass

---

## Phase 2: LSP & Editor Support

**Reference:** [gsx-syntax-design.md §7-9](./gsx-syntax-design.md#7-lsp-changes)

**Status:** ✅ Complete

- [x] Modify `pkg/lsp/server.go`
  - Update package comment to reference `.gsx` files
  - Update any `.tui` string literals

- [x] Modify `pkg/lsp/document.go`
  - Update comment referencing `.tui` file

- [x] Modify `pkg/lsp/handler.go`
  - Update `indexWorkspace()` to find `.gsx` files instead of `.tui`
  - Update file extension check from `.tui` to `.gsx`

- [x] Modify `pkg/lsp/definition.go`
  - Update comments referencing `.tui` to `.gsx`

- [x] Modify `pkg/lsp/completion.go`
  - Update comments referencing `.tui` to `.gsx`

- [x] Modify `pkg/lsp/hover.go`
  - Update comments referencing `.tui` to `.gsx`

- [x] Modify `pkg/lsp/gopls/proxy.go`
  - Update `TuiURIToGoURI()` to handle `.gsx` → `_gsx_generated.go`
  - Update `GoURIToTuiURI()` to handle `_gsx_generated.go` → `.gsx`
  - Update `DiskPath()` for `.gsx` extension
  - Update comments

- [x] Modify `pkg/lsp/gopls/mapping.go`
  - Update comments referencing `.tui` to `.gsx`

- [x] Modify `pkg/lsp/gopls/generate.go`
  - Update comments referencing `.tui` to `.gsx`

- [x] Update LSP test files (done in Phase 1 to pass tests)
  - `pkg/lsp/server_test.go` - change `file:///test.tui` to `file:///test.gsx`
  - `pkg/lsp/features_test.go` - change all `.tui` URIs to `.gsx`
  - `pkg/lsp/semantic_tokens_comment_test.go` - change all `.tui` URIs to `.gsx`
  - `pkg/lsp/gopls/proxy_test.go` - update URI tests for `.gsx`
  - Update test content to use `func ... Element` syntax

- [x] Rename `editor/tree-sitter-tui/` → `editor/tree-sitter-gsx/`
  - Rename entire directory

- [x] Modify `editor/tree-sitter-gsx/grammar.js`
  - Change grammar name from `'tui'` to `'gsx'`
  - Update `component_declaration` rule from `@component` to `func ... Element` pattern
  - See [gsx-syntax-design.md §8](./gsx-syntax-design.md#8-tree-sitter-grammar-changes)

- [x] Update tree-sitter bindings
  - `bindings/c/tree-sitter-tui.h` → `tree-sitter-gsx.h`
  - `bindings/c/tree-sitter-tui.pc.in` - update name
  - `bindings/go/binding.go` - update package references
  - `bindings/node/index.js`, `index.d.ts` - update references
  - `bindings/python/tree_sitter_tui/` → `tree_sitter_gsx/`
  - `bindings/swift/TreeSitterTui/tui.h` - update
  - `Cargo.toml`, `Package.swift`, `setup.py`, `pyproject.toml` - update names

- [x] Modify `editor/vscode/package.json`
  - Change `name` from `tui-language` to `gsx-language`
  - Change `displayName` to `GSX Language Support`
  - Change language `id` from `tui` to `gsx`
  - Change `extensions` from `[".tui"]` to `[".gsx"]`
  - Update `activationEvents` from `onLanguage:tui` to `onLanguage:gsx`
  - Change grammar `scopeName` from `source.tui` to `source.gsx`
  - Change grammar `path` to `./syntaxes/gsx.tmLanguage.json`
  - See [gsx-syntax-design.md §9](./gsx-syntax-design.md#9-vscode-extension-changes)

- [x] Rename and modify `editor/vscode/syntaxes/tui.tmLanguage.json` → `gsx.tmLanguage.json`
  - Change `name` from `TUI` to `GSX`
  - Change `scopeName` from `source.tui` to `source.gsx`
  - Update `component-declaration` pattern from `@component` to `func ... Element`
  - Replace all `.tui` scope suffixes with `.gsx`

- [x] Rename test files in `editor/vscode/test/`
  - `simple.tui` → `simple.gsx` (convert syntax)
  - `complex.tui` → `complex.gsx` (convert syntax)
  - `simple_tui.go` → `simple_gsx.go` (update source comment)

- [x] Update `editor/vscode/README.md`
  - Update all `.tui` references to `.gsx`
  - Update syntax examples

- [x] Update `editor/vscode/src/extension.ts`
  - Update language configuration to use `gsx` namespace
  - Update file watcher pattern from `**/*.tui` to `**/*.gsx`

**Tests:** Run `go test ./pkg/lsp/...` ✅ All pass

---

## Phase 3: Documentation & Examples

**Reference:** [gsx-syntax-design.md §10](./gsx-syntax-design.md#10-complexity-assessment)

**Status:** ✅ Complete

- [x] Modify `CLAUDE.md`
  - Update all `.tui` references to `.gsx`
  - Update all `@component` syntax examples to `func ... Element`
  - Update CLI command examples (`tui generate ./...` now finds `.gsx`)
  - Update directory structure comments
  - Update `.tui File Syntax` section header to `.gsx File Syntax`

- [x] Rename and convert `examples/dsl-counter/counter.tui` → `counter.gsx`
  - Rename file
  - Convert `@component` to `func ... Element` syntax

- [x] Rename and convert `examples/streaming-dsl/streaming.tui` → `streaming.gsx`
  - Rename file
  - Convert `@component` to `func ... Element` syntax

- [x] Rename and convert `examples/state/state.tui` → `state.gsx`
  - Removed - state.tui used unimplemented features (`#Content` refs)
  - The state example is an aspirational design spec for future features

- [x] Verify all examples compile and run
  - Run `go build ./examples/dsl-counter/...` ✅
  - Run `go build ./examples/streaming-dsl/...` ✅
  - Run `tui generate ./examples/dsl-counter/...` ✅
  - Run `tui generate ./examples/streaming-dsl/...` ✅

- [x] Update `editor/vscode/README.md`
  - Already updated in Phase 2

- [x] Update main.go files to reference `.gsx` instead of `.tui`
  - `examples/dsl-counter/main.go` - updated go:generate and comments
  - `examples/streaming-dsl/main.go` - updated go:generate and comments
  - `examples/state/main.go` - updated comments

**Tests:** Run `go test ./pkg/... ./cmd/...` - all pass ✅

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Core Syntax Changes (token, lexer, parser, formatter, CLI, tests) | ✅ Complete |
| 2 | LSP & Editor Support (LSP server, tree-sitter, VSCode extension) | ✅ Complete |
| 3 | Documentation & Examples (CLAUDE.md, examples conversion) | ✅ Complete |

## Files to Create

None - this is a refactoring change.

## Files to Modify

### Phase 1
| File | Changes |
|------|---------|
| `pkg/tuigen/token.go` | Remove `TokenAtComponent` |
| `pkg/tuigen/lexer.go` | Remove `@component` handling |
| `pkg/tuigen/parser.go` | Detect `func ... Element` as component |
| `pkg/formatter/printer.go` | Emit `func ... Element` syntax |
| `cmd/tui/generate.go` | `.tui` → `.gsx`, `_tui.go` → `_gsx.go` |
| `cmd/tui/check.go` | `.tui` → `.gsx` |
| `cmd/tui/fmt.go` | `.tui` → `.gsx` |
| `cmd/tui/main.go` | Update help text |
| `cmd/tui/testdata/*.tui` | Rename to `.gsx`, convert syntax |
| `pkg/tuigen/*_test.go` | Update test cases |

### Phase 2
| File | Changes |
|------|---------|
| `pkg/lsp/*.go` | `.tui` → `.gsx` references |
| `pkg/lsp/gopls/*.go` | `.tui` → `.gsx`, `_tui_generated` → `_gsx_generated` |
| `pkg/lsp/*_test.go` | Update URIs and syntax |
| `editor/tree-sitter-tui/` | Rename directory, update grammar |
| `editor/vscode/package.json` | Language ID and extension |
| `editor/vscode/syntaxes/*.json` | Rename, update scopes and patterns |

### Phase 3
| File | Changes |
|------|---------|
| `CLAUDE.md` | All `.tui` → `.gsx`, syntax examples |
| `examples/*/*.tui` | Rename to `.gsx`, convert syntax |
| `editor/vscode/README.md` | Documentation updates |
