# Codebase Restructure Implementation Plan

Implementation phases for the go-tui restructure. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Move Layout to Internal + Create Root Re-exports

**Reference:** [restructure-design.md §2](./restructure-design.md#2-architecture)

**Completed in commit:** 93fd701

- [x] Move `pkg/layout/*.go` to `internal/layout/`
  - Move all source files: `calculate.go`, `flex.go`, `layoutable.go`, `style.go`, `value.go`, `rect.go`, `edges.go`, `point.go`, `layout.go`
  - Move all test files: `calculate_test.go`, `integration_test.go`, `layoutable_test.go`, `rect_test.go`, `value_test.go`, `benchmark_test.go`
  - Update package-internal imports (layout has no internal deps — just verify `package layout` declarations)

- [x] Create `layout.go` in module root (package `tui`)
  - Re-export all layout types via type aliases: `Direction`, `Justify`, `Align`, `Value`, `Style` (as `LayoutStyle`), `Edges`, `Rect`, `Size`, `Point`, `Layout` (as `LayoutResult`)
  - Re-export all constants: `Row`, `Column`, `JustifyStart` through `JustifySpaceEvenly`, `AlignStart` through `AlignStretch`
  - Re-export constructors: `Fixed()`, `Percent()`, `Auto()`, `DefaultStyle()` (as `DefaultLayoutStyle()`), `NewRect()`, `EdgeAll()`, `EdgeSymmetric()`, `EdgeTRBL()`
  - Add header comment: `// layout.go re-exports layout types from internal/layout. Any changes to internal/layout types must be mirrored here.`
  - See [restructure-design.md §3](./restructure-design.md#3-core-entities) for the re-export pattern

- [x] Create backward-compat shim at `pkg/layout/compat.go`
  - `package layout` that re-exports everything from `internal/layout` via type aliases
  - This is TEMPORARY — allows examples and generated code to keep using `pkg/layout` until Phase 6
  - Add comment: `// DEPRECATED: This package is a temporary compatibility shim. Use "github.com/grindlemire/go-tui" instead.`

- [x] Update `pkg/tui/*.go` imports
  - Change `"github.com/grindlemire/go-tui/pkg/layout"` → `"github.com/grindlemire/go-tui/internal/layout"` in all pkg/tui source and test files
  - Files: `rect.go` (only file importing layout; `app.go` does not import layout)

- [x] Update `pkg/tui/element/*.go` imports
  - Change `"github.com/grindlemire/go-tui/pkg/layout"` → `"github.com/grindlemire/go-tui/internal/layout"` in all element source and test files
  - Files: `element.go`, `options.go`, `options_auto.go`, `render.go`, `scroll.go`, `element_test.go`, `options_test.go`, `render_test.go`, `scrollbox_test.go`, `integration_test.go`

- [x] Update `pkg/tuigen/*.go` imports (if any reference pkg/layout)
  - Checked `analyzer.go`, `generator.go`, `tailwind.go` — none have actual Go imports of `pkg/layout` (only string literals for generated code paths). No changes needed.

- [x] Create `doc.go` in module root
  - Package comment: brief description of the tui package as the public API
  - `package tui`

**Tests:** Run `go test ./internal/layout/... ./pkg/tui/... ./pkg/tui/element/... ./pkg/tuigen/...` — all pass

---

## Phase 2: Merge pkg/tui + pkg/tui/element Into Root Package

**Reference:** [restructure-design.md §2](./restructure-design.md#2-architecture)

**Completed in commit:** 93fd701

- [x] Move `pkg/tui/element/*.go` source files to module root
  - `element.go` → root `element.go` (change `package element` → `package tui`)
  - `options.go` → root `element_options.go`
  - `options_auto.go` → root `element_options_auto.go`
  - `render.go` → root `element_render.go`
  - `scroll.go` → root `element_scroll.go`
  - Remove `import "github.com/grindlemire/go-tui/pkg/tui"` — types like `tui.Style`, `tui.BorderStyle`, `tui.Buffer` become just `Style`, `BorderStyle`, `Buffer`
  - Remove `import "github.com/grindlemire/go-tui/internal/layout"` — already available via root's own import
  - Remove `import "github.com/grindlemire/go-tui/pkg/debug"` — update to `internal/debug` path

- [x] Move `pkg/tui/element/*_test.go` files to module root
  - `element_test.go` → root `element_test.go`
  - `options_test.go` → root `element_options_test.go`
  - `render_test.go` → root `element_render_test.go`
  - `scrollbox_test.go` → root `element_scrollbox_test.go`
  - `integration_test.go` → root `element_integration_test.go`
  - Remove cross-package test imports (`"github.com/grindlemire/go-tui/pkg/tui"`, `"github.com/grindlemire/go-tui/pkg/layout"`)
  - Tests now reference types directly (same package)

- [x] Move `pkg/tui/*.go` source files to module root
  - Move ALL source files: `app.go`, `border.go`, `buffer.go`, `caps.go`, `cell.go`, `color.go`, `dirty.go`, `escape.go`, `event.go`, `focus.go`, `key.go`, `mock_reader.go`, `mock_terminal.go`, `parse.go`, `reader.go`, `reader_unix.go`, `render.go`, `state.go`, `style.go`, `terminal.go`, `terminal_ansi.go`, `terminal_unix.go`, `watcher.go`
  - Change `package tui` is already correct (root package is also `tui`)
  - Update `import "github.com/grindlemire/go-tui/pkg/layout"` → `"github.com/grindlemire/go-tui/internal/layout"` (if not done in Phase 1 — some files may still reference old path)
  - Update `import "github.com/grindlemire/go-tui/pkg/debug"` → `"github.com/grindlemire/go-tui/internal/debug"` in all files
  - Remove root `rect.go` duplicate (Phase 1 created one; `pkg/tui/rect.go` has the same content — keep the Phase 1 version which already points to internal/layout)

- [x] Move `pkg/tui/*_test.go` files to module root
  - Move ALL test files, keeping original names
  - Rename conflicts: `integration_test.go` → `app_integration_test.go` (since element's is `element_integration_test.go`)
  - Update test imports to remove `"github.com/grindlemire/go-tui/pkg/layout"` and `"github.com/grindlemire/go-tui/pkg/tui/element"`

- [x] Create backward-compat shims (TEMPORARY)
  - `pkg/tui/compat.go`: `package tui` that re-exports key types from root via aliases
  - `pkg/tui/element/compat.go`: `package element` that re-exports Element, New, Option, and all With* from root
  - These allow examples and generated code to keep working until Phase 6

- [x] Update `cmd/tui/*.go` imports
  - Change `"github.com/grindlemire/go-tui/pkg/tui"` → `"github.com/grindlemire/go-tui"`
  - Change `"github.com/grindlemire/go-tui/pkg/tui/element"` → remove (same package now)
  - Change `"github.com/grindlemire/go-tui/pkg/layout"` → remove (re-exported from root)

- [x] Delete empty directories
  - Delete `pkg/tui/element/` contents (except compat.go shim)
  - Verify `pkg/tui/` only has compat shims left

**Tests:** Run `go test ./...` — all pass (examples use shims)

---

## Phase 3: Move Tooling to Internal

**Reference:** [restructure-design.md §2](./restructure-design.md#2-architecture)

**Completed in commit:** 89c986a

- [x] Move `pkg/debug/` → `internal/debug/`
  - Move `debug.go`
  - Update ALL consumers: root package files (`state.go`, `watcher.go`, `focus.go`, etc.)
  - Update `import "github.com/grindlemire/go-tui/pkg/debug"` → `"github.com/grindlemire/go-tui/internal/debug"`

- [x] Move `pkg/tuigen/` → `internal/tuigen/`
  - Move all source files: `ast.go`, `token.go`, `errors.go`, `lexer.go`, `parser.go`, `analyzer.go`, `generator.go`, `tailwind.go`
  - Move all test files
  - Move `cmd/tui/testdata/` if it references tuigen
  - Package declaration stays `package tuigen`

- [x] Move `pkg/formatter/` → `internal/formatter/`
  - Move all source files: `formatter.go`, `printer.go`, `imports.go`
  - Move all test files: `formatter_test.go`, `formatter_comment_test.go`
  - Update `import "github.com/grindlemire/go-tui/pkg/tuigen"` → `"github.com/grindlemire/go-tui/internal/tuigen"` in formatter files

- [x] Move `pkg/lsp/` → `internal/lsp/`
  - Move all source files and subdirectories: `gopls/`, `log/`, `provider/`, `schema/`
  - Core files: `server.go`, `router.go`, `handler.go`, `context.go`, `context_test.go`, `document.go`, `index.go`, `provider_adapters.go`, `providers.go`
  - Legacy adapters (thin delegators): `completion.go`, `definition.go`, `diagnostics.go`, `formatting.go`, `hover.go`, `references.go`, `semantic_tokens.go`, `symbols.go`
  - Test files: `features_test.go`, `server_test.go`, `semantic_tokens_comment_test.go`
  - Move all test files
  - Update imports: `pkg/tuigen` → `internal/tuigen`, `pkg/formatter` → `internal/formatter`
  - Update imports within `provider/*.go` and `schema/*.go` files if they reference `pkg/tuigen` or `pkg/formatter`

- [x] Update `cmd/tui/*.go` imports
  - `generate.go`: `"github.com/grindlemire/go-tui/pkg/tuigen"` → `"github.com/grindlemire/go-tui/internal/tuigen"`
  - `check.go`: same tuigen import update
  - `fmt.go`: `"github.com/grindlemire/go-tui/pkg/formatter"` → `"github.com/grindlemire/go-tui/internal/formatter"`
  - `lsp.go`: `"github.com/grindlemire/go-tui/pkg/lsp"` → `"github.com/grindlemire/go-tui/internal/lsp"`

- [x] Update code generator output paths
  - In `internal/tuigen/generator.go`, change emitted import paths:
    - `"github.com/grindlemire/go-tui/pkg/tui"` → `"github.com/grindlemire/go-tui"`
    - `"github.com/grindlemire/go-tui/pkg/tui/element"` → remove (merged into root)
    - `"github.com/grindlemire/go-tui/pkg/layout"` → remove (re-exported from root)
  - Update generated code references: `element.New(` → `tui.New(`, `element.With*` → `tui.With*`, `layout.Column` → `tui.Column`, etc.
  - Update view struct: `*element.Element` → `*tui.Element`

- [x] Update `internal/tuigen/analyzer.go` import path references
  - String literals referencing `"github.com/grindlemire/go-tui/pkg/tui"` → `"github.com/grindlemire/go-tui"`

- [x] Delete `pkg/` directory entirely
  - Remove all shim/compat files created in Phases 1-2
  - Delete `pkg/layout/`, `pkg/tui/`, `pkg/tui/element/`, `pkg/debug/`, `pkg/tuigen/`, `pkg/formatter/`, `pkg/lsp/`
  - The `pkg/` directory should no longer exist

- [x] Update all examples/ and editor/ files to use new import paths
  - Replace `pkg/tui` → root import, remove `pkg/layout` and `pkg/tui/element` imports
  - Update `element.*` → `tui.*` and `layout.*` → `tui.*` code references
  - Update `pkg/debug` → `internal/debug`

**Tests:** `go build ./...` and `go test ./...` — all pass.

---

## Phase 4: Split Oversized Source Files

**Reference:** [restructure-design.md §2](./restructure-design.md#2-architecture)

**Completed in commit:** 31f7619

All splits are pure file reorganization — no logic changes. Target: every source file <=500 lines.

- [x] Split root `app.go` (~913 lines) into:
  - `app.go` — App struct, NewApp constructor, NewAppWithReader
  - `app_options.go` — AppOption type and all With* option functions (~112 lines)
  - `app_lifecycle.go` — Close, PrintAbove, printAboveRaw
  - `app_events.go` — Dispatch, event handling, readInputEvents
  - `app_render.go` — Render, renderInline, RenderFull methods
  - `app_loop.go` — Run, Stop, QueueUpdate

- [x] Split root `element.go` (~713 lines) into:
  - `element.go` — Element struct, New(), type definitions (TextAlign, ScrollMode)
  - `element_layout.go` — Layoutable interface impl (LayoutStyle, LayoutChildren, IntrinsicSize)
  - `element_tree.go` — AddChild, RemoveChild, Children, Parent, tree walking, notifyChildAdded
  - `element_accessors.go` — Getters/setters for style, border, background, text, focus properties
  - `element_focus.go` — Focus/Blur, HandleEvent, handleScrollEvent, WalkFocusables
  - `element_watchers.go` — SetOnUpdate, AddWatcher, WalkWatchers, ElementAtPoint

- [x] Split `internal/tuigen/parser.go` (~1537 lines) into:
  - `parser.go` — Parser struct, initialization, token navigation, comment handling, file/package/import parsing
  - `parser_component.go` — Component and function signature parsing, templ detection
  - `parser_element.go` — Element tag parsing, attributes, inline children
  - `parser_control.go` — @let, @for, @if parsing and related helpers
  - `parser_expr.go` — Go expression parsing, text content, component calls

- [x] Split `internal/tuigen/generator.go` (~1312 lines) into:
  - `generator.go` — Generator struct, file/package/import generation, utility methods
  - `generator_component.go` — Component function generation, view struct generation
  - `generator_element.go` — Element creation, option building, attribute-to-option mapping
  - `generator_control.go` — For loop, if statement, let binding generation
  - `generator_children.go` — Children rendering, body dispatch, slice-building context

- [x] Split `internal/tuigen/analyzer.go` (~1133 lines) into:
  - `analyzer.go` — Analyzer struct, known attributes/tags, main Analyze method, component validation
  - `analyzer_refs.go` — Named ref validation, inference, let-binding transformation
  - `analyzer_imports.go` — Import management, missing import insertion
  - `analyzer_state.go` — State variable detection, binding detection, deps parsing

- [x] Split `internal/tuigen/lexer.go` (~924 lines) into:
  - `lexer.go` — Lexer struct, initialization, main Next() method, position tracking
  - `lexer_strings.go` — String, rune, raw string literal reading
  - `lexer_goexpr.go` — Balanced brace reading for Go expressions, peek variants
  - `lexer_utils.go` — Comment collection, identifier reading, number literals, utility helpers

- [x] Split `internal/tuigen/tailwind.go` (~929 lines) into:
  - `tailwind.go` — ParseTailwindClass, ParseTailwindClasses, BuildTextStyleOption
  - `tailwind_data.go` — Static class map, regex patterns, accumulator types
  - `tailwind_validation.go` — Validation, fuzzy matching, Levenshtein distance
  - `tailwind_autocomplete.go` — AllTailwindClasses documentation data

- [x] Split `internal/lsp/provider/semantic.go` (~1382 lines) into:
  - `semantic.go` — Types, constants, main SemanticTokensProvider, encoding
  - `semantic_nodes.go` — AST node processing and dispatch
  - `semantic_gocode.go` — Go expression tokenization, variable extraction

- [x] Split `internal/lsp/provider/references.go` (~874 lines) into:
  - `references.go` — Main ReferencesProvider, reference dispatch
  - `references_search.go` — Cross-file search, workspace scanning

- [x] Split `internal/lsp/context.go` (~837 lines) into:
  - `context.go` — CursorContext struct, NodeKind enum, Scope struct, resolve entry point
  - `context_resolve.go` — AST walking, node classification, scope building

- [x] Split `internal/lsp/provider/definition.go` (~741 lines) into:
  - `definition.go` — Main DefinitionProvider, definition dispatch
  - `definition_search.go` — Cross-file definition search, gopls delegation

- [x] Split `internal/lsp/provider/completion.go` (~587 lines) into:
  - `completion.go` — Main CompletionProvider, completion dispatch
  - `completion_items.go` — Completion item builders, attribute/event completions

- [x] Split `internal/formatter/printer.go` (~852 lines) into:
  - `printer.go` — Printer struct, PrintFile, package/component printing, node dispatch
  - `printer_elements.go` — Element printing with attributes and inline children
  - `printer_control.go` — @for, @if, @let, component call printing
  - `printer_comments.go` — Comment formatting and printing methods

- [x] Split `internal/lsp/gopls/proxy.go` (~564 lines) into:
  - `proxy.go` — GoplsProxy struct, lifecycle, communication
  - `proxy_requests.go` — Request forwarding, response handling

- [x] Split `internal/lsp/gopls/generate.go` (~557 lines) into:
  - `generate.go` — Virtual Go file generation core
  - `generate_state.go` — State variable and named ref emission

- [x] Note: The following LSP root-level files are now thin adapters (≤195 lines each) after the devtools overhaul moved logic into `provider/`. No splits needed:
  - `semantic_tokens.go` (12 lines), `hover.go` (33 lines), `references.go` (14 lines), `definition.go` (13 lines), `formatting.go` (14 lines), `diagnostics.go` (69 lines), `completion.go` (195 lines), `symbols.go` (49 lines), `handler.go` (466 lines)

**Tests:** Run `go test ./... ` (excluding examples) — all pass, no logic changes

---

## Phase 5: Split Root Package Test Files

**Reference:** [restructure-design.md §2](./restructure-design.md#2-architecture)

**Completed in commit:** (in progress)

Pure test file reorganization for the root package. Target: every test file <=500 lines. Tests split by topic to match source file splits from Phase 4.

**NOTE:** `app_test.go` (223 lines) and `element_test.go` (452 lines) are already ≤500 — no split needed. `formatter_test.go` is in `internal/formatter/`, not root — listed in error. `parse_test.go` (658 lines) was missing from this plan but needs splitting.

- [x] ~~Split root `app_test.go` (~956 lines)~~ — Already ≤500 lines (223). Previously split in Phase 4.

- [x] ~~Split root `element_test.go` (~1532 lines)~~ — Already ≤500 lines (452). Previously split in Phase 4.

- [x] Split root `buffer_test.go` (837→421 lines) into:
  - `buffer_test.go` (421 lines) — Buffer creation, fill, clear, resize
  - `buffer_text_test.go` (313 lines) — SetRune, SetString, wide character tests
  - `buffer_diff_test.go` (113 lines) — Diff, swap, change detection tests

- [x] Split root `state_test.go` (~828 lines) into:
  - `state_test.go` — NewState, Get, Set, MarksDirty (lines 1-121)
  - `state_binding_test.go` — Bind, unbind, concurrent, cleanup tests (lines 122-475)
  - `state_batch_test.go` — Batch, nested batch, coalescing, panic recovery (lines 477-828)

- [x] ~~Split root `formatter_test.go` (~823 lines)~~ — NOT in root package (lives in `internal/formatter/`). Remove from Phase 5.

- [x] Split root `parse_test.go` (658 lines) into:
  - `parse_test.go` — Keyboard/control/function key parsing (lines 1-494)
  - `parse_mouse_test.go` — Mouse SGR parsing, mouse event tests (lines 496-658)

- [x] Split root `focus_test.go` (~618 lines) into:
  - `focus_test.go` — FocusManager, Register, Next, Prev, SetFocus (lines 1-408)
  - `focus_dispatch_test.go` — Dispatch, blur, skip, empty tests (lines 410-618)

- [x] Split root `rect_test.go` (~600 lines) into:
  - `rect_test.go` — NewRect, RightBottom, Area, IsEmpty, Contains, ContainsRect (lines 1-283)
  - `rect_ops_test.go` — Inset, Intersect, Union, Translate, Clamp, Immutability (lines 285-601)

- [x] Split root `escape_test.go` (~539 lines) into:
  - `escape_test.go` — MoveTo, cursor, clear, hide/show, alt screen, reset, bold, attributes (lines 1-257)
  - `escape_style_test.go` — Color tests, combined styles, WriteRune, Reset, containsSubstring (lines 259-539)

- [x] Split root `element_integration_test.go` (586 lines) into:
  - `element_integration_test.go` — BasicFlow through BackgroundAndBorder (lines 1-326)
  - `element_integration_render_test.go` — DeepNesting, Centering, RenderOutput, Culling, Gap, TextAlignment (lines 328-587)

- [x] Split root `element_render_test.go` (526 lines) into:
  - `element_render_test.go` — Background, border, nested, cull, text, alignment, stringWidth, checkString (lines 1-385)
  - `element_render_hr_test.go` — HR rendering tests (lines 387-527)

**Tests:** Run `go test .` — all root package tests pass, no logic changes

---

## Phase 6: Split internal/tuigen Test Files

**Reference:** [restructure-design.md §2](./restructure-design.md#2-architecture)

**Completed in commit:** (done — see Phase 5 commit)

Pure test file reorganization for the `internal/tuigen` package. Target: every test file <=500 lines.

- [x] Split `internal/tuigen/analyzer_test.go` (~1896 lines) into:
  - `analyzer_test.go` — Basic analysis, component validation tests
  - `analyzer_refs_test.go` — Named ref, let binding, ref inference tests
  - `analyzer_state_test.go` — State detection, binding analysis tests
  - `analyzer_error_test.go` — Error cases, invalid syntax tests

- [x] Split `internal/tuigen/generator_test.go` (~1794 lines) into:
  - `generator_test.go` — Basic generation, file structure tests
  - `generator_element_test.go` — Element generation, options, attributes
  - `generator_control_test.go` — For loop, if statement, let binding generation
  - `generator_component_test.go` — View struct, component call generation

- [x] Split `internal/tuigen/parser_test.go` (~1720 lines) into:
  - `parser_test.go` — File, package, import parsing
  - `parser_element_test.go` — Element and attribute parsing
  - `parser_control_test.go` — @if, @for, @let parsing
  - `parser_component_test.go` — Component and function parsing

- [x] Split `internal/tuigen/tailwind_test.go` (~1572 lines) into:
  - `tailwind_test.go` — Single class parsing tests
  - `tailwind_batch_test.go` — Multi-class parsing, accumulator tests
  - `tailwind_extended_test.go` — Extended single-class parsing (sizing, flex, borders)
  - `tailwind_validation_test.go` — Validation, fuzzy match, error tests

- [x] Split `internal/tuigen/lexer_test.go` (~873 lines) into:
  - `lexer_test.go` — Basic token, punctuation, keyword tests
  - `lexer_strings_test.go` — String literal, rune, raw string tests
  - `lexer_goexpr_test.go` — Go expression, balanced brace tests

- [x] Split remaining `internal/tuigen` test files >500 lines similarly by topic
  - `parser_comment_test.go` (546 lines) → `parser_comment_test.go` + `parser_comment_context_test.go`

**Tests:** Run `go test ./internal/tuigen/...` — all tuigen tests pass, no logic changes

---

## Phase 7: Split internal/lsp and internal/layout Test Files

**Reference:** [restructure-design.md §2](./restructure-design.md#2-architecture)

**Completed in commit:** (pending)

Pure test file reorganization for the `internal/lsp` and `internal/layout` packages. Target: every test file <=500 lines.

- [x] Split `internal/layout/calculate_test.go` (~1538 lines) into:
  - `calculate_test.go` — Single node, fixed size, percent tests
  - `calculate_flex_test.go` — FlexGrow, FlexShrink, gap tests
  - `calculate_align_test.go` — Justify, align, padding, margin tests
  - `calculate_minmax_test.go` — MinWidth, MaxWidth, constraint tests
  - `calculate_intrinsic_test.go` — Intrinsic sizing tests

- [x] Split `internal/layout/rect_test.go` (~712 lines) into:
  - `rect_test.go` (350 lines) — Construction, accessors
  - `rect_ops_test.go` (365 lines) — Intersection, union, contains

- [x] Split `internal/lsp/features_test.go` (~855 lines) into:
  - `features_test.go` (143 lines) — Basic LSP feature tests
  - `features_completion_test.go` (288 lines) — Completion-specific tests
  - `features_hover_test.go` (435 lines) — Hover-specific tests

- [x] Split `internal/lsp/provider/semantic_test.go` (~847 lines) into:
  - `semantic_test.go` (417 lines) — Basic semantic token tests, constant verification
  - `semantic_nodes_test.go` (435 lines) — AST node token output tests

- [x] Split `internal/lsp/context_test.go` (~775 lines) into:
  - `context_test.go` (184 lines) — Basic CursorContext resolution, NodeKind classification
  - `context_scope_test.go` (400 lines) — Scope resolution, state vars, named refs in scope
  - `context_helpers_test.go` (200 lines) — Text helper function tests (getLineText, getWordAtOffset, isOffsetIn*)

- [x] Split `internal/lsp/gopls/proxy_test.go` (~640 lines) into:
  - `proxy_test.go` (361 lines) — SourceMap, VirtualFileCache, URI conversion tests
  - `proxy_requests_test.go` (285 lines) — GenerateVirtualGo, SourceMap advanced tests

- [x] Split remaining `internal/lsp` test files >500 lines similarly by topic
  - `internal/lsp/server_test.go` (549 lines) → `server_test.go` (434 lines) + `server_document_test.go` (121 lines)
  - `internal/lsp/semantic_tokens_comment_test.go` (~488 lines) — already ≤500, no split needed

**Tests:** Run `go test ./internal/layout/... ./internal/lsp/...` — all tests pass, no logic changes

---

## Phase 8: Update Examples, Generator Output, and Documentation

**Reference:** [restructure-design.md §4](./restructure-design.md#4-user-experience)

**Completed in commit:** PLACEHOLDER

- [x] Update ALL example imports to use root package
  - Already completed in Phase 3. All 24 examples use `tui "github.com/grindlemire/go-tui"`.
  - No references to `pkg/tui`, `pkg/tui/element`, or `pkg/layout` remain.

- [x] Regenerate ALL `*_gsx.go` files
  - Ran `go run ./cmd/tui generate ./examples/...`
  - All 17 generated files use single `"github.com/grindlemire/go-tui"` import
  - Generated code uses `tui.New(`, `tui.With*`, `tui.Column`, etc.

- [x] Update `cmd/tui/testdata/` files
  - Already completed in Phase 3. All testdata files use correct import paths.

- [x] Update `editor/vscode/test/simple_gsx.go`
  - Already completed in Phase 3. Uses `tui "github.com/grindlemire/go-tui"`.

- [x] Update `CLAUDE.md`
  - Updated Directory Structure to reflect root package + internal/ layout
  - Updated Architecture diagram with internal package annotations
  - Replaced `layout.Direction/Justify/Align` → `tui.Direction/Justify/Align` in attribute tables
  - Updated Key Types section: `layout.*` → `tui.*`
  - Updated `go test` commands to use `internal/` paths

- [x] Update `generate.go` at module root
  - Verified: go:generate directive (`go run ./cmd/tui generate ./...`) is correct and functional

- [x] Add `doc.go` files with package documentation
  - Root `doc.go`: already existed
  - `internal/layout/doc.go`: created — flexbox engine description
  - `internal/tuigen/doc.go`: created — DSL compiler pipeline description
  - `internal/formatter/doc.go`: created — code formatter description
  - `internal/lsp/doc.go`: created — LSP server description
  - `internal/debug/doc.go`: created — debug logging description

- [x] Final verification
  - `go build ./...` — all packages and examples build ✓
  - `go test ./...` — all tests pass ✓
  - `go vet ./...` — no issues ✓
  - `pkg/` directory no longer exists ✓
  - **Note:** 6 files from previous phases remain >500 lines (not addressed here):
    - `internal/formatter/formatter_test.go` (928), `internal/lsp/context_resolve.go` (699),
      `internal/lsp/provider/semantic_nodes.go` (648), `internal/formatter/formatter_comment_test.go` (631),
      `internal/lsp/provider/references.go` (563), `internal/tuigen/analyzer.go` (514)

**Tests:** `go test ./...` — ALL tests pass including examples

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Move layout to internal, create root re-exports + compat shim | Done |
| 2 | Merge pkg/tui + pkg/tui/element into root package | Done |
| 3 | Move tuigen, formatter, lsp, debug to internal | Done |
| 4 | Split oversized source files (<=500 lines each) | Done |
| 5 | Split root package test files (<=500 lines each) | Done |
| 6 | Split internal/tuigen test files (<=500 lines each) | Done |
| 7 | Split internal/lsp + internal/layout test files (<=500 lines each) | Done |
| 8 | Update examples, regenerate code, update docs | Done |

## Files to Create

```
(root: package tui)
├── doc.go
├── layout.go                     # Re-exports from internal/layout
├── element.go                    # Merged from pkg/tui/element/
├── element_options.go
├── element_options_auto.go
├── element_render.go
├── element_scroll.go
├── element_layout.go             # Split from element.go
├── element_tree.go               # Split from element.go
├── element_accessors.go          # Split from element.go
├── element_focus.go              # Split from element.go
├── element_watchers.go           # Split from element.go
├── app.go                        # Moved from pkg/tui/
├── app_options.go                # Split from app.go
├── app_lifecycle.go              # Split from app.go
├── app_events.go                 # Split from app.go
├── app_render.go                 # Split from app.go
├── app_loop.go                   # Split from app.go
├── (all other files from pkg/tui/ and pkg/tui/element/ — unchanged names)
│
├── internal/
│   ├── layout/
│   │   ├── doc.go
│   │   └── (all files from pkg/layout/)
│   ├── tuigen/
│   │   ├── doc.go
│   │   ├── parser.go             # Split original
│   │   ├── parser_component.go
│   │   ├── parser_element.go
│   │   ├── parser_control.go
│   │   ├── parser_expr.go
│   │   ├── generator.go          # Split original
│   │   ├── generator_component.go
│   │   ├── generator_element.go
│   │   ├── generator_control.go
│   │   ├── generator_children.go
│   │   ├── analyzer.go           # Split original
│   │   ├── analyzer_refs.go
│   │   ├── analyzer_imports.go
│   │   ├── analyzer_state.go
│   │   ├── lexer.go              # Split original
│   │   ├── lexer_strings.go
│   │   ├── lexer_goexpr.go
│   │   ├── lexer_utils.go
│   │   ├── tailwind.go           # Split original
│   │   ├── tailwind_data.go
│   │   ├── tailwind_validation.go
│   │   ├── tailwind_autocomplete.go
│   │   └── (test files — split similarly)
│   ├── formatter/
│   │   ├── doc.go
│   │   ├── printer.go            # Split original
│   │   ├── printer_elements.go
│   │   ├── printer_control.go
│   │   ├── printer_comments.go
│   │   └── (other files unchanged)
│   ├── lsp/
│   │   ├── doc.go
│   │   ├── server.go             # LSP server lifecycle
│   │   ├── router.go             # Method routing with provider dispatch
│   │   ├── handler.go            # Initialize response, capabilities
│   │   ├── context.go            # CursorContext (split)
│   │   ├── context_resolve.go    # AST walking, node classification
│   │   ├── document.go           # Document management
│   │   ├── index.go              # Workspace symbol index
│   │   ├── provider_adapters.go  # Adapter layer: router → providers
│   │   ├── providers.go          # Provider initialization
│   │   ├── (thin legacy adapters: completion.go, definition.go, etc.)
│   │   ├── schema/
│   │   │   ├── schema.go         # Elements, attributes, type defs
│   │   │   ├── keywords.go       # DSL keywords and documentation
│   │   │   └── tailwind.go       # Tailwind class defs and docs
│   │   ├── provider/
│   │   │   ├── provider.go       # Interfaces and registry
│   │   │   ├── hover.go          # Hover provider
│   │   │   ├── completion.go     # Completion provider (split)
│   │   │   ├── completion_items.go
│   │   │   ├── definition.go     # Definition provider (split)
│   │   │   ├── definition_search.go
│   │   │   ├── references.go     # References provider (split)
│   │   │   ├── references_search.go
│   │   │   ├── symbols.go        # Symbol providers
│   │   │   ├── diagnostics.go    # Diagnostics provider
│   │   │   ├── formatting.go     # Formatting provider
│   │   │   ├── semantic.go       # Semantic tokens (split)
│   │   │   ├── semantic_nodes.go
│   │   │   ├── semantic_gocode.go
│   │   │   └── (test files)
│   │   ├── gopls/
│   │   │   ├── proxy.go          # Subprocess communication (split)
│   │   │   ├── proxy_requests.go
│   │   │   ├── generate.go       # Virtual Go generation (split)
│   │   │   ├── generate_state.go
│   │   │   └── mapping.go
│   │   └── log/
│   │       └── log.go
│   └── debug/
│       ├── doc.go
│       └── debug.go
```

## Files to Modify

| File | Changes |
|------|---------|
| `cmd/tui/generate.go` | Import paths: pkg/tuigen → internal/tuigen |
| `cmd/tui/check.go` | Import paths: pkg/tuigen → internal/tuigen |
| `cmd/tui/fmt.go` | Import paths: pkg/formatter → internal/formatter |
| `cmd/tui/lsp.go` | Import paths: pkg/lsp → internal/lsp |
| `cmd/tui/main.go` | Import paths if needed |
| `internal/tuigen/generator.go` | Emit new root import path in generated code |
| `internal/tuigen/analyzer.go` | Update import path string literals |
| `examples/*/main.go` | All imports → single root import |
| `examples/*/*_gsx.go` | Regenerated with new imports |
| `CLAUDE.md` | Architecture, examples, directory structure |
| `generate.go` | Verify still works |

## Files to Delete

| File/Directory | Reason |
|------|---------|
| `pkg/` (entire directory) | Replaced by root package + internal/ |
