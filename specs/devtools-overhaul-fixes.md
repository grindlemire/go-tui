# Devtools Overhaul - Post-Implementation Fixes

Code review findings from the implementation of devtools-overhaul-plan.md. Build, vet, and all tests pass clean.

---

## HIGH - Bugs Producing Incorrect Results

### 1. Indent calculation bug in gopls virtual Go generation
- **File:** `pkg/lsp/gopls/generate.go` lines 379, 410, 414, 479
- **Bug:** Nested body indent calculated as `len(indent)/4+1`. Since indent strings use `strings.Repeat("\t", level)` (1 byte per tab), dividing by 4 produces wrong results. At level 1: `1/4+1=1` (no increase); at level 2: `2/4+1=1` (decreases). Should be `len(indent)+1`.
- **Impact:** Flattens all nested for-loops, if-statements, and component call children in virtual Go. Breaks source-map column calculations for nested expressions.
- **Fix:** Replace `len(indent)/4+1` with `len(indent)+1` at all four sites.
- [x] Fixed

### 2. Function reference search uses substring matching without word boundaries
- **File:** `pkg/lsp/provider/references.go:508`
- **Bug:** `strings.Contains(n.Code, name+"(")` matches substrings. Searching for function `f` would match `fmt.Printf(`.
- **Impact:** False positive references returned for short function names.
- **Fix:** Extracted `findFuncCallInCode` helper with word-boundary check. Replaced all inline `strings.Contains`/`strings.Index` usages.
- [x] Fixed

### 3. Function reference search only finds first occurrence per code block
- **File:** `pkg/lsp/provider/references.go:508-519`
- **Bug:** `strings.Index` returns only the first match. `helper(a) + helper(b)` only reports the first call.
- **Impact:** Missing references when a function is called multiple times in one expression.
- **Fix:** Use a loop to find all occurrences with word-boundary checking.
- [x] Fixed

### 4. Definition provider range mapping uses `||` instead of `&&`
- **File:** `pkg/lsp/provider/definition.go:376`
- **Bug:** `if startFound || endFound` allows returning a location when only one end maps. If only `startFound`, end defaults to `(0,0)`.
- **Impact:** Broken definition ranges when only partial position mapping succeeds.
- **Fix:** Change `||` to `&&`.
- [x] Fixed

### 5. Component references don't search workspace ASTs
- **File:** `pkg/lsp/provider/references.go:95-115`
- **Bug:** `findComponentReferences` only searches open documents. Unlike `findFunctionReferences` (which calls `searchWorkspaceForFunctionRefs`), component references in closed files are missed.
- **Impact:** Incomplete find-references results for components when files aren't open.
- **Fix:** Add workspace AST search analogous to `searchWorkspaceForFunctionRefs`.
- [x] Fixed

### 6. Number literal tokenizer over-consumes `+`/`-` operators
- **File:** `pkg/lsp/provider/semantic.go:564-571`
- **Bug:** Number parsing loop consumes `+` and `-` unconditionally. In `1+2`, it tokenizes `1+2` as one number. Should only consume `+`/`-` after `e`/`E` in scientific notation.
- **Impact:** Incorrect semantic token boundaries for arithmetic expressions.
- **Fix:** Track whether previous character was `e`/`E` before consuming `+`/`-`.
- [x] Fixed

### 7. Format specifiers use `TokenTypeNumber` instead of `TokenTypeRegexp`
- **File:** `pkg/lsp/provider/semantic.go:718-724`
- **Bug:** `emitStringWithFormatSpecifiers` emits `%s`, `%d` etc as `TokenTypeNumber`. The `TokenTypeRegexp` constant (12) exists specifically for format specifiers.
- **Impact:** Format specifiers get wrong highlighting color (number color instead of regexp/purple).
- **Fix:** Change `TokenTypeNumber` to `TokenTypeRegexp` on line 722.
- [x] Fixed

### 8. GoCode nodes silently dropped in virtual Go generation
- **File:** `pkg/lsp/gopls/generate.go:247-267`
- **Bug:** `generateNode` switch has no `case *tuigen.GoCode:`. Non-`tui.NewState` GoCode in the component body is silently dropped from virtual Go.
- **Impact:** Go expressions in GoCode nodes (like `visible := true`) get no source mapping, breaking gopls integration for those lines.
- **Fix:** Add a `case *tuigen.GoCode:` that emits the code with source mapping.
- [x] Fixed

### 9. Tree-sitter generated files are stale
- **File:** `editor/tree-sitter-gsx/src/parser.c`, `grammar.json`, `node-types.json`
- **Bug:** Generated parser files still reference `@component` (old syntax). The `grammar.js` uses `templ`. Need `tree-sitter generate`.
- **Impact:** Tree-sitter parser doesn't match the grammar definition.
- **Fix:** Run `tree-sitter generate` to regenerate.
- [x] Fixed

---

## MEDIUM - Correctness Issues / Design Problems

### 10. `ParentChain` is never populated
- **File:** `pkg/lsp/context.go:100`
- **Context:** The `CursorContext.ParentChain` field is declared and copied through adapters but never appended to. Providers checking ancestor context always see empty slice.

### 11. Scope `InLoop`/`InConditional` flags always false during scope collection
- **File:** `pkg/lsp/context.go:476-528`
- **Context:** `collectScopeFromBody` runs before `resolveInNodes`, so `ctx.Scope.ForLoop`/`IfStmt` are always nil when named ref flags are set.

### 12. Named ref position mapping points to element start, not `#Name`
- **File:** `pkg/lsp/gopls/generate.go:198-201`
- **Context:** Source map maps `tuiCol = n.Position.Column - 1` (the element's `<` position), not the `#Name` position.

### 13. `input` element schema missing `onClick`
- **File:** `pkg/lsp/schema/schema.go:336-349`
- **Context:** `inputAttrs()` manually lists event handlers but omits `onClick`. Every other interactive element gets it through `eventAttrs()`.

### 14. Schema is stricter than the compiler for attributes
- **File:** `pkg/lsp/schema/schema.go`
- **Context:** The analyzer uses a flat `knownAttributes` map (any attr on any element). Schema assigns per-element attribute sets. LSP won't suggest attrs the compiler accepts.

### 15. `@else` keyword is never tokenized
- **File:** `pkg/lsp/provider/semantic.go:325-343`
- **Context:** `IfStmt` handler emits keyword token for `@if` but not `@else`. AST doesn't store `@else` position.

### 16. References provider ignores `NodeKind` entirely
- **File:** `pkg/lsp/provider/references.go:26-91`
- **Context:** Uses only word-based heuristic matching. A `NodeKindElement` with word matching a component name would be misidentified.

### 17. `onChannel` and `onTimer` not in `EventHandlers` map
- **File:** `pkg/lsp/schema/schema.go`
- **Context:** `watcherAttrs()` defines these as `"event"` category but they're missing from `EventHandlers`. `IsEventHandler()` returns false.

### 18. Tree-sitter grammar missing `map_type`
- **File:** `editor/tree-sitter-gsx/grammar.js`
- **Context:** `type_expression` supports identifier, qualified_type, slice_type but not map types. `map[string]string` can't parse.

### 19. Four token types registered in legend but never emitted
- **File:** `pkg/lsp/provider/semantic.go`
- **Context:** `TokenTypeRegexp` (12), `TokenTypeProperty` (6), `TokenTypeNamespace` (0), `TokenTypeType` (1) registered but unused. Attributes use `TokenTypeFunction` instead of `TokenTypeProperty`.

### 20. Enclosing component detection doesn't check end boundary
- **File:** `pkg/lsp/context.go:194-199`
- **Context:** Selects last component with start line <= cursor, without checking cursor is before closing `}`.

---

## LOW - Code Quality / Test Gaps / Minor Issues

### 21. Duplicate `PositionToOffset` implementations
- `pkg/lsp/document.go:148-174` and `pkg/lsp/provider/provider.go:319-333`

### 22. Duplicate `isWordChar` / `IsWordChar`
- `pkg/lsp/definition.go:11` and `pkg/lsp/provider/provider.go:337`

### 23. `functionNameCheckerAdapter` allocates map on every call
- `pkg/lsp/provider_adapters.go:346-354`

### 24. `classPrefix` lookback is 100 bytes vs `isOffsetInClassAttr` 500
- `pkg/lsp/provider/completion.go:162` vs `pkg/lsp/context.go`

### 25. Dot (`.`) not registered as trigger character
- `pkg/lsp/handler.go:144` — state method completions require manual invocation

### 26. `GoExpr`/`GoCode`/`TextContent` match on line only, not column
- `pkg/lsp/context.go:249-266`

### 27. `findVariableInCode` assumes single-line code blocks
- `pkg/lsp/provider/references.go:694-726`

### 28. `parseFuncName` doesn't handle method receivers
- `pkg/lsp/provider/definition.go:593-605`

### 29. No test for `NodeKindStateAccess`
- `pkg/lsp/context_test.go`

### 30. No formatting provider tests
- `pkg/lsp/provider/formatting.go`

### 31. `nodeToSymbol` doesn't recurse into nested structures
- `pkg/lsp/provider/symbols.go:74`

### 32. `estimateErrorLength` always returns hardcoded 10
- `pkg/lsp/provider/diagnostics.go:87`

### 33. `TailwindClassDef.SortKey` and `.Pattern` fields never used
- `pkg/lsp/schema/tailwind.go:13-14`

### 34. Dead `event-handler-attribute` rule in TextMate grammar
- `editor/vscode/syntaxes/gsx.tmLanguage.json:383-390`

### 35. Gopls `readResponses` leaves pending requests hanging on error
- `pkg/lsp/gopls/proxy.go:460-494`

### 36. Missing test cases in tree-sitter corpus
- `editor/tree-sitter-gsx/test/corpus/basic.txt`

### 37. No tree-sitter highlight rules for event handler attributes or state methods
- `editor/tree-sitter-gsx/queries/highlights.scm`

### 38. `definitionNamedRef` points to element tag, not `#Name`
- **File:** `pkg/lsp/provider/definition.go:148-154`
- **Context:** Related to #12. When going to definition on a named ref usage in a Go expression, the provider returns `elem.Position` (the `<div` tag start), not the `#Name` position on that element. The Range covers the tag name (e.g. `div`) instead of `#Header`. User lands on the wrong text.

### 39. Diagnostics fallback drops `doc.Errors` when provider is nil
- **File:** `pkg/lsp/diagnostics.go:46-47`
- **Context:** When `DiagnosticsProvider` is nil (line 46), `publishDiagnostics` sends an empty diagnostics array. The document's `doc.Errors` (parse errors from tuigen) are silently dropped. If provider registration fails for any reason, the user sees zero errors even on broken files.

### 40. Inconsistent nil-document return types across router handlers
- **File:** `pkg/lsp/router.go`
- **Context:** When `docs.Get(uri)` returns nil, handlers return different things:
  - `dispatchPositional` (hover/completion/definition) returns `nil, nil` (line 233)
  - `handleDocumentSymbol` returns `[]DocumentSymbol{}, nil` (line 147)
  - `handleSemanticTokensFull` returns `&SemanticTokens{Data: []int{}}, nil` (line 208)
  - `handleFormatting` returns `nil, nil` (line 186)
  Not a crash bug, but inconsistent behavior that could confuse LSP clients expecting uniform responses.

### 41. Duplicate loop variable check in definition provider
- **File:** `pkg/lsp/provider/definition.go:42-47` and `103-106`
- **Context:** `findLoopVariableDefinition(ctx, word)` is called twice in the same code path. First at line 42-47 (early return before gopls) and again at line 103-106 (component scope fallback). If the first call returns nil, the second will also return nil (same function, same inputs). Wasteful but not broken.

### 42. Dead Scope nil checks in providers
- **File:** `pkg/lsp/provider/definition.go:90, 158, 189, 218, 270` and others
- **Context:** `ResolveCursorContext` always initializes `Scope` to `&Scope{}` (`context.go:122`), so `ctx.Scope == nil` checks are unreachable dead code. Not harmful but misleading for maintainers.

### 43. Definition tests don't exercise the NodeKind switch
- **File:** `pkg/lsp/provider/definition_test.go`
- **Context:** The definition provider's switch (`definition.go:65-74`) handles `NodeKindComponentCall`, `NodeKindNamedRef`, `NodeKindStateAccess`, `NodeKindStateDecl`, and `NodeKindEventHandler`. Tests only exercise `NodeKindComponentCall` and `NodeKindUnknown` (word-based fallback). The named ref, state, and event handler code paths through the switch have zero test coverage.

### 44. Reference tests use weak count assertions
- **File:** `pkg/lsp/provider/references_test.go`
- **Context:** All reference tests check `len(result) < N` instead of exact counts (e.g. lines 51-52, 91-92, 130-131). A test expecting 2 references passes if 50 are returned. This masks over-reporting bugs.

### 45. Semantic token tests don't verify positions
- **File:** `pkg/lsp/provider/semantic_test.go`
- **Context:** Tests count tokens by type (e.g. "expect 3 keyword tokens") but don't verify line/column positions. A token at the wrong position still passes. The `decodeTokens` helper exists but position validation is never asserted.

---

## Implementation Plan

Issues 1-9 are already fixed. The plan below covers issues 10-45 organized into four phases. Each phase builds on the previous and has clear acceptance criteria.

### Phase 1: CursorContext & Scope Correctness

**Why first:** CursorContext is the foundation. Every provider depends on it producing correct NodeKind, Scope, and position data. Fixing these first means provider fixes in later phases work against correct inputs.

**Issues addressed:** #10, #11, #20, #26, #42

---

#### #10 — Populate `ParentChain` during AST walk

**File:** `pkg/lsp/context.go`, function `resolveFromAST` and its recursive helpers (`resolveInNodes`, `resolveInElement`, `resolveInForLoop`, `resolveInIfStmt`, etc.)

**What to do:** As the AST walk descends into child nodes, append the current node to `ctx.ParentChain` before recursing. When a recursive call returns `true` (found the cursor), leave the chain as-is. When it returns `false`, pop the last element. The result should be a root-to-cursor path.

**Pattern:**
```go
// Before recursing into a child:
ctx.ParentChain = append(ctx.ParentChain, currentNode)
if resolveInNodes(ctx, childBody, line, col) {
    return true // chain stays
}
ctx.ParentChain = ctx.ParentChain[:len(ctx.ParentChain)-1] // pop
```

Apply this pattern in:
- `resolveFromAST` when entering a component body (line ~180)
- `resolveInNodes` when entering ForLoop, IfStmt, LetBinding, Element, ComponentCall children
- `resolveInElement` when entering element children
- `resolveInForLoop` when entering loop body
- `resolveInIfStmt` when entering if/else bodies

**Test:** In `pkg/lsp/context_test.go`, add a test that parses a nested structure (component > for loop > if > element) and verifies `ctx.ParentChain` contains the expected node types from root to cursor.

---

#### #11 — Fix Scope collection ordering so `ForLoop`/`IfStmt` are set

**File:** `pkg/lsp/context.go`

**What's wrong:** `collectScopeFromBody` (called around line 175) populates `Scope.NamedRefs` and `Scope.StateVars`, but it runs before `resolveInNodes` sets `Scope.ForLoop` and `Scope.IfStmt`. Named ref context (simple vs loop vs keyed) is always "simple" because the loop/conditional flags are nil during collection.

**What to do:** Two options:
1. **(Recommended)** Move the scope collection to run AFTER `resolveInNodes` completes. `resolveInNodes` sets `Scope.ForLoop`/`IfStmt` when it finds the cursor is inside one. Then `collectScopeFromBody` can check those fields.
2. Or: Change `collectScopeFromBody` to accept the enclosing ForLoop/IfStmt as parameters and detect whether each named ref is inside a loop/conditional by checking parent nodes during collection.

**Test:** Parse a component with a named ref inside a `@for` loop. Verify `ctx.Scope.ForLoop` is non-nil when the cursor is on the ref.

---

#### #20 — Check component end boundary in enclosing component detection

**File:** `pkg/lsp/context.go`, function `resolveFromAST`, around lines 194-199.

**What's wrong:** The code selects the last component whose `Position.Line <= cursorLine` as the enclosing component. If the cursor is below all components (e.g., in whitespace after the last `}`), it still picks the last component.

**What to do:** After selecting the candidate component, verify the cursor line is before the component's closing brace. The tuigen AST `Component` struct has a body that ends at a certain line. Check either:
- Walk the body to find the last node's end position, or
- Use a heuristic: scan backward from the end of the document for the matching `}` brace.

If the cursor is past the component's body, leave `ctx.Scope.Component` as nil and classify from text instead.

**Test:** Parse a file with two components. Place the cursor on a blank line between them. Verify `ctx.Scope.Component` is nil (or correctly identifies it's not in either component).

---

#### #26 — Use column checking for GoExpr/GoCode/TextContent cursor resolution

**File:** `pkg/lsp/context.go`, around lines 249-266.

**What's wrong:** Node matching for GoExpr, GoCode, and TextContent nodes checks only `line == node.Position.Line` without verifying the cursor column is within the node's character range. On a line with multiple inline expressions (e.g., `<span>{a}</span>{b}`), both GoExprs match by line alone.

**What to do:** For each node type, after the line match, also check that `col >= node.Position.Column` and (if available) `col <= node.Position.Column + len(node.Code)`. If the AST doesn't store end positions, estimate from content length for single-line nodes.

**Test:** Parse a line with two inline Go expressions. Place cursor in the second one. Verify the resolved node is the second expression, not the first.

---

#### #42 — Remove dead `Scope == nil` checks

**File:** `pkg/lsp/provider/definition.go` (lines 90, 158, 189, 218, 270), `pkg/lsp/provider/references.go`, `pkg/lsp/provider/hover.go`

**What to do:** Since `ResolveCursorContext` always sets `Scope` to `&Scope{}` (context.go:122), remove the `if ctx.Scope == nil` guards. These are dead branches. Grep for `ctx.Scope == nil` and `ctx\.Scope != nil` across the provider package and remove the checks, inlining the "non-nil" branch.

**Test:** Existing tests should continue to pass. No new tests needed.

---

**Phase 1 acceptance criteria:**
- `go test ./pkg/lsp/...` passes
- `ParentChain` is populated for nested cursors
- Scope has correct `ForLoop`/`IfStmt` when cursor is inside one
- Component detection excludes cursor positions outside component bodies
- GoExpr/GoCode resolution respects column position

---

### Phase 2: Provider Navigation Fixes

**Why second:** With CursorContext producing correct data, fix the providers that consume it.

**Issues addressed:** #12, #16, #38, #39, #40, #41

---

#### #12 + #38 — Fix named ref position in both gopls generate and definition provider

These are two manifestations of the same problem: named ref positions point to the element tag instead of `#Name`.

**File 1:** `pkg/lsp/gopls/generate.go:198-201`

**What's wrong:** The source map for named ref variable declarations maps `tuiCol` to `n.Position.Column - 1`, which is the `<` character of the element. It should map to the `#` character position.

**What to do:** The tuigen `Element` struct has a `NamedRef` field (the ref name string). To find the `#Name` position within the source line, locate `#` + `elem.NamedRef` on the element's line. Use `strings.Index(lineText, "#"+elem.NamedRef)` relative to the element position. Map `tuiCol` to that index.

**File 2:** `pkg/lsp/provider/definition.go:148-154`, function `definitionNamedRef`

**What's wrong:** Returns `elem.Position` (the element tag position `<div`) as the definition location. The Range covers `elem.Tag` (e.g., `div`).

**What to do:** Find the `#Name` position on the element's source line. The element's `Position` is the start of the tag. Scan the source line for `#` + `elem.NamedRef` and construct the Range around that. You can get the source line from `ctx.Document.Content` using `elem.Position.Line`.

**Pattern for definition.go:**
```go
func (d *definitionProvider) definitionNamedRef(ctx *CursorContext) ([]Location, error) {
    elem, ok := ctx.Node.(*tuigen.Element)
    if !ok || elem == nil || elem.NamedRef == "" {
        return nil, nil
    }

    // Find #Name position on the element's line
    line := getLineText(ctx.Document.Content, elem.Position.Line - 1) // 0-indexed
    hashRef := "#" + elem.NamedRef
    idx := strings.Index(line, hashRef)
    if idx < 0 {
        // Fallback to element position
        idx = elem.Position.Column - 1
    }

    return []Location{{
        URI: ctx.Document.URI,
        Range: Range{
            Start: Position{Line: elem.Position.Line - 1, Character: idx},
            End:   Position{Line: elem.Position.Line - 1, Character: idx + len(hashRef)},
        },
    }}, nil
}
```

**Test:** Create a test with `<div #Header class="...">`. Go-to-definition on a usage of `Header` in a Go expression. Verify the result range covers `#Header`, not `div`.

---

#### #16 — Use NodeKind in references provider dispatch

**File:** `pkg/lsp/provider/references.go:26-91`, function `References`

**What's wrong:** The `References` method ignores `ctx.NodeKind` and relies entirely on word-based heuristics. It tries component lookup, then function lookup, then parameter lookup, etc., in sequence. If an element tag name happens to match a component name, it returns component references instead of nothing.

**What to do:** Add a `switch ctx.NodeKind` at the top of the function (similar to how `definition.go` does it) to dispatch to the correct search. For unknown/fallback kinds, keep the current word-based heuristic.

**Pattern:**
```go
func (r *referencesProvider) References(ctx *CursorContext, includeDecl bool) ([]Location, error) {
    word := ctx.Word
    if word == "" {
        return []Location{}, nil
    }

    switch ctx.NodeKind {
    case NodeKindComponentCall, NodeKindComponent:
        return r.findComponentReferences(ctx, word, includeDecl)
    case NodeKindFunction:
        return r.findFunctionReferences(ctx, word, includeDecl)
    case NodeKindParameter:
        return r.findParameterReferences(ctx, word, includeDecl)
    case NodeKindLetBinding:
        return r.findLetBindingReferences(ctx, word, includeDecl)
    case NodeKindNamedRef:
        return r.findNamedRefReferences(ctx, word, includeDecl)
    case NodeKindStateDecl, NodeKindStateAccess:
        return r.findStateVarReferences(ctx, word, includeDecl)
    }

    // Fallback: word-based heuristic (existing behavior)
    ...
}
```

Keep the existing word-based logic as the `default` fallback for `NodeKindUnknown` and unhandled kinds.

**Test:** Existing tests should continue to pass (they set explicit NodeKinds). Verify that cursoring on an element tag named `Header` (NodeKindElement) does NOT return component references for a component also named `Header`.

---

#### #39 — Fall back to `doc.Errors` when DiagnosticsProvider is nil

**File:** `pkg/lsp/diagnostics.go:46-47`

**What's wrong:** When the provider is nil, diagnostics are silently set to an empty array. The document may have parse errors in `doc.Errors` that should still be shown.

**What to do:** In the `else` branch (line 46-47), convert `doc.Errors` to `[]Diagnostic` using the same position mapping the provider uses (tuigen 1-indexed to LSP 0-indexed):

```go
} else {
    // No provider registered — fall back to inline conversion of parse errors
    for _, e := range doc.Errors {
        diagnostics = append(diagnostics, Diagnostic{
            Range: Range{
                Start: Position{Line: e.Line - 1, Character: e.Column - 1},
                End:   Position{Line: e.Line - 1, Character: e.Column - 1 + 10}, // estimate
            },
            Severity: DiagnosticSeverityError,
            Source:   "gsx",
            Message:  e.Msg,
        })
    }
}
```

Check the `Document` struct to verify `doc.Errors` exists and what fields it has (likely `Line`, `Column`, `Msg` from tuigen).

**Test:** Not critical (this is a fallback path), but you can test by calling `publishDiagnostics` with a nil registry and a document that has errors. Verify the notification includes the errors.

---

#### #40 — Normalize nil-document return types in router

**File:** `pkg/lsp/router.go`

**What to do:** Pick one convention for nil-document returns and apply it consistently. Recommended: return `nil, nil` for all handlers when the document isn't found (this is what most handlers already do). The LSP spec treats `null` results as "no result available," which is correct for a missing document.

Change these two handlers to match:
- `handleDocumentSymbol` (line 147): change `return []DocumentSymbol{}, nil` to `return nil, nil`
- `handleSemanticTokensFull` (line 208): change `return &SemanticTokens{Data: []int{}}, nil` to `return nil, nil`

**Test:** Existing tests should pass. These are only hit when a document isn't tracked.

---

#### #41 — Remove duplicate loop variable check

**File:** `pkg/lsp/provider/definition.go:42-47`

**What to do:** Remove the early loop variable check at lines 42-47. The same check runs later at lines 103-106 within the component scope section. The early check was likely left over from refactoring.

Remove:
```go
// Check for loop variables before gopls (gopls doesn't understand .gsx for loops)
if ctx.Scope != nil && ctx.Scope.Component != nil {
    if loc := d.findLoopVariableDefinition(ctx, word); loc != nil {
        return []Location{*loc}, nil
    }
}
```

The loop variable check at line 103-106 will still run.

**Test:** Existing tests pass. The loop variable definition test should still work.

---

**Phase 2 acceptance criteria:**
- `go test ./pkg/lsp/...` passes
- Go-to-definition on named ref usage lands on `#Name`, not element tag
- References for components don't false-match on element tags with same name
- Parse errors shown even if DiagnosticsProvider is nil
- Router returns consistent types for missing documents

---

### Phase 3: Schema & Semantic Token Correctness

**Why third:** These are self-contained data/highlighting fixes that don't depend on the previous phases.

**Issues addressed:** #13, #15, #17, #19, #25, #34

---

#### #13 — Add `onClick` to input element attributes

**File:** `pkg/lsp/schema/schema.go`, function `inputAttrs()` (around line 336-349)

**What's wrong:** `inputAttrs()` manually lists event handlers but omits `onClick`. All other interactive elements get it via `eventAttrs()`.

**What to do:** Add `eventAttrs()...` to the `inputAttrs()` slice (or add `onClick` explicitly). Check how `button` element does it and mirror that pattern.

**Test:** In `pkg/lsp/schema/schema_test.go`, verify `schema.GetAttribute("input", "onClick")` returns non-nil.

---

#### #15 — Emit `@else` keyword token

**File:** `pkg/lsp/provider/semantic.go`, in the `IfStmt` handler (around lines 325-343)

**What's wrong:** The handler emits a keyword token for `@if` but not `@else`. The AST `IfStmt` struct has an `Else` body but doesn't store the `@else` position.

**What to do:** After processing the `@if` branch, if `ifStmt.Else != nil`, scan the source text between the end of the `@if` body `}` and the start of the `@else` body `{` for the string `@else`. Use the document content with the known line range. Emit a keyword token at that position with length 5 (`@else`).

**Pattern:**
```go
if ifStmt.Else != nil {
    // Find @else between } and { in the source
    // The @else appears after the closing } of the if body
    // Scan lines between if-body end and else-body start
    elseKeywordLine, elseKeywordCol := findElseKeyword(doc.Content, ifStmt)
    if elseKeywordLine >= 0 {
        tokens = append(tokens, SemanticToken{
            Line: elseKeywordLine, StartChar: elseKeywordCol,
            Length: 5, TokenType: TokenTypeKeyword,
        })
    }
}
```

**Test:** In `pkg/lsp/provider/semantic_test.go`, add a test with `@if ... { } @else { }` and verify keyword token count includes the `@else`.

---

#### #17 — Add `onChannel` and `onTimer` to EventHandlers map

**File:** `pkg/lsp/schema/schema.go`, the `EventHandlers` map initialization

**What's wrong:** `watcherAttrs()` creates attributes with category `"event"` for `onChannel` and `onTimer`, but they're not in the `EventHandlers` map. So `schema.IsEventHandler("onChannel")` returns false, and these don't get event handler hover/completion.

**What to do:** Add entries to the `EventHandlers` map:
```go
"onChannel": {Name: "onChannel", Description: "Called when a message is received on the channel", Signature: "func()"},
"onTimer":   {Name: "onTimer", Description: "Called on each timer tick", Signature: "func()"},
```

Check the actual expected signatures in the framework code to get them right.

**Test:** Verify `schema.IsEventHandler("onChannel")` and `schema.IsEventHandler("onTimer")` return true.

---

#### #19 — Use registered token types correctly in semantic tokens

**File:** `pkg/lsp/provider/semantic.go`

**What's wrong:** Four token types are registered in the legend but never emitted:
- `TokenTypeNamespace` (0) — should be used for `package` and `import` keywords
- `TokenTypeType` (1) — should be used for type names in parameter types
- `TokenTypeProperty` (6) — should be used for element attributes (currently uses `TokenTypeFunction`)
- `TokenTypeRegexp` (12) — should be used for format specifiers (fixed in issue #7, verify)

**What to do:** Grep for where attributes and package/import/type tokens are emitted and change their token types:
- Element attributes: change from `TokenTypeFunction` to `TokenTypeProperty`
- `package`/`import` declarations: change to `TokenTypeNamespace`
- Type names in param lists (e.g., `string`, `int`, `layout.Direction`): change to `TokenTypeType`

Do this carefully — changing token types changes highlighting colors. Make sure the semantic token legend order in `handler.go` (the `initialize` response) matches the constant values.

**Test:** Update semantic token tests to expect the new token types.

---

#### #25 — Register `.` as a completion trigger character

**File:** `pkg/lsp/handler.go`, in the `handleInitialize` function, around line 144

**What's wrong:** The `CompletionOptions.TriggerCharacters` list includes `@`, `<`, `{`, `"`, `/` but not `.`. State method completions (`count.Get()`, `count.Set()`) require the user to manually invoke completion because `.` doesn't trigger it.

**What to do:** Add `"."` to the `TriggerCharacters` slice.

**Test:** Manual verification — type `count.` and completion should trigger automatically.

---

#### #34 — Remove dead `event-handler-attribute` rule in TextMate grammar

**File:** `editor/vscode/syntaxes/gsx.tmLanguage.json:383-390`

**What's wrong:** The rule matches `on[A-Z][a-zA-Z]*` inside Go expression content but the scope is redundant (it applies inside `{}` where Go syntax highlighting already applies, not in HTML attribute position).

**What to do:** Remove the dead rule. Verify event handler attributes are still highlighted correctly via the attribute patterns earlier in the grammar (the `entity.other.attribute-name.gsx` pattern).

**Test:** Open `editor/vscode/test/complex.gsx` and visually verify `onClick`, `onFocus` are highlighted as attributes.

---

**Phase 3 acceptance criteria:**
- `go test ./pkg/lsp/...` passes
- `input` element offers `onClick` in completions
- `@else` gets keyword highlighting
- `onChannel`/`onTimer` recognized as event handlers
- Attributes highlighted as properties, not functions
- `.` triggers completion for state methods

---

### Phase 4: Test Coverage & Code Cleanup

**Why last:** These are non-functional improvements. Do after correctness fixes to avoid testing against wrong behavior.

**Issues addressed:** #14, #21, #22, #23, #24, #27, #28, #29, #30, #32, #33, #43, #44, #45

---

#### Test coverage fixes (#29, #30, #43, #44, #45)

**#43 — Add definition tests for NodeKind switch paths**

**File:** `pkg/lsp/provider/definition_test.go`

Add test cases that set explicit NodeKinds and verify the switch in `definition.go:65-74` dispatches correctly:

- `TestDefinition_NamedRef`: Set `NodeKind: NodeKindNamedRef`, `Node` to a `*tuigen.Element` with `NamedRef: "Header"`. Verify returned location points to `#Header`.
- `TestDefinition_StateDecl`: Set `NodeKind: NodeKindStateDecl`, `Word: "count"`, `Scope.StateVars` containing a `count` state var. Verify returned location points to the declaration.
- `TestDefinition_StateAccess`: Set `NodeKind: NodeKindStateAccess`, `Word: "count"`. Same as above.
- `TestDefinition_EventHandler`: Set `NodeKind: NodeKindEventHandler`, `InGoExpr: true`. Verify it delegates to gopls (or returns nil if no proxy).

**#44 — Strengthen reference test assertions**

**File:** `pkg/lsp/provider/references_test.go`

Change all `if len(result) < N` to `if len(result) != N`. For each test, determine the exact expected count from the test fixture and assert it. Example:
```go
// Before:
if len(result) < 2 { t.Fatal("expected at least 2 references") }
// After:
if len(result) != 2 { t.Fatalf("expected 2 references, got %d", len(result)) }
```

**#45 — Add position assertions to semantic token tests**

**File:** `pkg/lsp/provider/semantic_test.go`

For at least the component declaration and keyword tests, use the `decodeTokens` helper to verify specific tokens appear at expected line/column positions. Example:
```go
decoded := decodeTokens(result.Data)
// Verify "templ" keyword is at line 0, col 0
found := false
for _, tok := range decoded {
    if tok.TokenType == TokenTypeKeyword && tok.Line == 0 && tok.StartChar == 0 && tok.Length == 5 {
        found = true
    }
}
if !found { t.Error("expected templ keyword token at 0:0") }
```

**#29 — Add `NodeKindStateAccess` test to context_test.go**

Parse a component with `count := tui.NewState(0)` and a Go expression using `count.Get()`. Place cursor on `.Get()`. Verify `ctx.NodeKind == NodeKindStateAccess`.

**#30 — Add formatting provider test**

**File:** Create `pkg/lsp/provider/formatting_test.go`

Test that `Format()` on a document with inconsistent indentation returns a TextEdit that replaces the full document with formatted content. Use a simple fixture with wrong indentation and verify the result has correct indentation.

---

#### Code cleanup (#21, #22, #23, #24, #27, #28, #32, #33)

These are small, independent cleanups. Do them as a batch:

- **#21 + #22:** Extract `PositionToOffset` and `IsWordChar` into a shared internal package (e.g., `pkg/lsp/lsputil/`) or have one package import from the other. The provider package should define them and the lsp package should call through.
- **#23:** Move the builtin function names to a package-level `var` in `provider_adapters.go` instead of constructing a map on each call.
- **#24:** Align `classPrefix` lookback in `completion.go` (100 bytes) with `isOffsetInClassAttr` in `context.go` (500 bytes). Use the same constant.
- **#27:** Make `findVariableInCode` split on newlines and handle multi-line GoCode blocks.
- **#28:** Update `parseFuncName` to skip `(receiver type)` prefix in function signatures.
- **#32:** Make `estimateErrorLength` actually extract a meaningful length from the error message (e.g., find quoted identifiers, measure the last word).
- **#33:** Remove unused `SortKey` and `Pattern` fields from `TailwindClassDef`, or implement their usage in `MatchClass()`.

---

#### Deferred / Won't Fix

These issues are acknowledged but intentionally deferred:

- **#14 (Schema stricter than compiler):** The schema's per-element attribute sets are more correct than the compiler's flat map. The compiler should eventually be updated to match, not the other way around. Keep the schema as-is.
- **#18 (Tree-sitter missing map_type):** Map types in parameters are rare in .gsx files. Add if someone reports it.
- **#20 (Enclosing component end boundary):** Already addressed in Phase 1.
- **#26 (GoExpr column matching):** Already addressed in Phase 1.
- **#31 (nodeToSymbol recursion):** LSP document symbols work fine with flat component/function lists. Nested symbols can be added later.
- **#35 (Gopls readResponses hanging):** This is a robustness issue in the gopls proxy. The proxy already handles most error cases. Fix if users report hanging.
- **#36 + #37 (Tree-sitter corpus + highlight rules):** Nice to have. The tree-sitter grammar works correctly; these are additional test/highlight coverage.

---

**Phase 4 acceptance criteria:**
- `go test ./pkg/lsp/...` passes
- Definition tests cover all NodeKind switch cases
- Reference tests use exact count assertions
- Semantic token tests verify at least some positions
- StateAccess and formatting have test coverage
- No duplicate utility implementations across packages

---

## Phase Summary

| Phase | Description | Issues |
|-------|-------------|--------|
| 1 | CursorContext & Scope Correctness | #10, #11, #20, #26, #42 |
| 2 | Provider Navigation Fixes | #12, #16, #38, #39, #40, #41 |
| 3 | Schema & Semantic Token Correctness | #13, #15, #17, #19, #25, #34 |
| 4 | Test Coverage & Code Cleanup | #14, #21-24, #27-30, #32, #33, #43-45 |

**After each phase:** Run `go test ./pkg/lsp/...` and verify all tests pass before proceeding.
