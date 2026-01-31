# Phase 4 Implementation Prompt

Implement Phase 4 of the go-tui restructure: "Split Oversized Source Files" as defined in `specs/restructure-plan.md` lines 166-260.

## Rules

1. **Pure file reorganization** — no logic changes, no renaming, no refactoring. Cut declarations from the source file and paste them into the new file verbatim.
2. **Same package** — every new file uses the same `package` declaration as the original. No import path changes anywhere in the codebase.
3. **Correct imports per file** — each new file gets only the `import` block it actually needs. Remove unused imports from the slimmed-down original.
4. **No test changes** — since everything stays in the same package, all existing tests compile and pass without modification.
5. **Target ≤500 lines per file** — but don't force awkward splits just to hit a number. Natural groupings matter more.
6. **Verification after each split** — run `go build ./...` after each file is split. Run `go test ./...` after all splits are complete.
7. **Commit with `gcommit -m "message"`** — never `git commit`. Single commit at the end.

## Execution Order

Work through the 15 splits below. You can do independent splits in parallel (different packages don't interact). The splits within the same package should be done sequentially to avoid conflicts.

**Group A (root package):** `app.go`, `element.go` — do sequentially
**Group B (internal/tuigen):** `parser.go`, `generator.go`, `analyzer.go`, `lexer.go`, `tailwind.go` — do sequentially
**Group C (internal/lsp):** `context.go` — standalone
**Group D (internal/lsp/provider):** `semantic.go`, `references.go`, `definition.go`, `completion.go` — do sequentially
**Group E (internal/formatter):** `printer.go` — standalone
**Group F (internal/lsp/gopls):** `proxy.go`, `generate.go` — do sequentially

Groups A-F are independent and can be done in parallel.

---

## Split 1: `app.go` (913 lines) → 6 files

### `app.go` — App struct, interfaces, constructors, accessors (~430 lines)
Keep:
- L1-14: package, imports, `const InputLatencyBlocking`
- L114-188: interfaces (`Renderable`, `focusableTreeWalker`, `watcherTreeWalker`, `mouseHitTester`, `Viewable`), `type App struct`, `var currentApp`
- L199-429: `func NewApp`, `func NewAppWithReader`
- L528-640: `func (a *App) SetRoot` through `func (a *App) PollEvent` (all simple accessors/setters)

### `app_options.go` — AppOption type and all With* functions (~100 lines)
Move:
- L16-17: `type AppOption func(*App)`
- L19-112: `WithInputLatency`, `WithFrameRate`, `WithEventQueueSize`, `WithGlobalKeyHandler`, `WithRoot`, `WithoutMouse`, `WithCursor`, `WithInlineHeight`

### `app_lifecycle.go` — Close, PrintAbove, package-level Stop (~100 lines)
Move:
- L190-197: `func Stop()` (package-level)
- L431-526: `func (a *App) Close`, `func (a *App) PrintAbove`, `func (a *App) PrintAboveln`, `func (a *App) printAboveRaw`

### `app_events.go` — Dispatch, readInputEvents (~75 lines)
Move:
- L642-688: `func (a *App) Dispatch`
- L886-913: `func (a *App) readInputEvents`

### `app_render.go` — Render, renderInline, RenderFull (~100 lines)
Move:
- L690-789: `func (a *App) Render`, `func (a *App) renderInline`, `func (a *App) RenderFull`

### `app_loop.go` — Run, Stop (method), QueueUpdate (~95 lines)
Move:
- L791-884: `func (a *App) Run`, `func (a *App) Stop` (method), `func (a *App) QueueUpdate`

---

## Split 2: `element.go` (707 lines) → 6 files

### `element.go` — Element struct, New(), type definitions (~115 lines)
Keep:
- L1-10: package, imports, var assertions
- L12-36: `type TextAlign`, `type ScrollMode` with const blocks
- L38-101: `type Element struct`, `var _ Layoutable`
- L103-114: `func New`

### `element_layout.go` — Layoutable interface impl (~140 lines)
Move:
- L118-165: `LayoutStyle`, `LayoutChildren`, `SetLayout`, `GetLayout`, `IsDirty`, `SetDirty`, `IsHR`
- L167-250: `IntrinsicSize`
- L321-334: `Calculate`, `Rect`, `ContentRect`
- L336-344: `MarkDirty`

### `element_tree.go` — Tree structure (~70 lines)
Move:
- L254-319: `AddChild`, `notifyChildAdded`, `SetOnChildAdded`, `RemoveChild`, `RemoveAllChildren`, `Children`, `Parent`

### `element_accessors.go` — Property getters/setters (~85 lines)
Move:
- L346-428: `SetStyle`, `Style`, `Border`, `SetBorder`, `BorderStyle`, `SetBorderStyle`, `Background`, `SetBackground`, `Text`, `SetText`, `TextStyle`, `SetTextStyle`, `TextAlign` (method), `SetTextAlign`, `stringWidth`

### `element_focus.go` — Focus/Blur, event handling (~200 lines)
Move:
- L432-623: `IsFocusable`, `IsFocused`, `Focus`, `Blur`, `SetFocusable`, `SetOnKeyPress`, `SetOnClick`, `SetOnEvent`, `SetOnFocus`, `SetOnBlur`, `HandleEvent`, `handleScrollEvent`

### `element_watchers.go` — Watchers, tree walking, hit testing (~85 lines)
Move:
- L627-707: `SetOnFocusableAdded`, `WalkFocusables`, `SetOnUpdate`, `AddWatcher`, `Watchers`, `WalkWatchers`, `ElementAt`, `ElementAtPoint`

---

## Split 3: `internal/tuigen/parser.go` (1551 lines) → 5 files

### `parser.go` — Parser struct, token navigation, comments, file/package/import parsing (~370 lines)
Keep:
- L1-367: package, imports, `type Parser`, `NewParser`, `Errors`, `advance`, `advanceSkipNewlines`, `skipNewlines`, `position`, `expect`, `expectSkipNewlines`, `synchronize`, comment handling functions, `groupComments`, `getLeadingCommentGroup`, `getTrailingCommentOnLine`, `ParseFile`, `parsePackage`, `parseImports`, `parseSingleImport`

### `parser_component.go` — Component/function parsing (~340 lines)
Move:
- L369-704: `parseFuncOrComponent`, `parseGoDecl`, `captureRawGoFunc`, `parseTempl`, `parseParams`, `parseParam`, `parseType`

### `parser_element.go` — Element tags, attributes, children (~450 lines)
Move:
- L706-1155: `parseComponentBody`, `parseComponentBodyWithOrphans`, `attachLeadingComments`, `parseBodyNode`, `parseElement`, `parseAttributes`, `parseAttribute`, `parseChildren`, `setBlankLineBefore`

### `parser_control.go` — @let, @for, @if (~200 lines)
Move:
- L1232-1426: `parseLet`, `parseFor`, `parseIf`

### `parser_expr.go` — Go expressions, text content, component calls (~170 lines)
Move:
- L1157-1230: `parseGoExprNode`, `parseGoStatement`
- L1428-1551: `parseComponentCall`, `parseGoExprOrChildrenSlot`, `isTextToken`, `isWordToken`

---

## Split 4: `internal/tuigen/generator.go` (1301 lines) → 5 files

### `generator.go` — Generator struct, file-level generation, utilities (~360 lines)
Keep:
- L1-10: package, imports
- L12-59: `type deferredWatcher`, `type deferredHandler`, `type Generator`, `NewGenerator`
- L61-156: `Generate`, `generateHeader`, `generatePackage`, `generateImports`
- L1120-1301: `nextVar`, `write`, `writef`, `writeln`, `writeIndent`, `GenerateString`, `ParseAndGenerate`, `parseAndGenerateSkipImports`, `parseAndGenerate`, `textElementWithOptions`, `skipTextChildren`, `GenerateToBuffer`, `generateStateBindings`, `generateBinding`, `getSetterForAttribute`

### `generator_component.go` — Component function + view struct generation (~210 lines)
Move:
- L158-365: `generateComponent`, `generateViewStruct`

### `generator_element.go` — Element creation, options, attributes (~310 lines)
Move:
- L367-672: `generateElement`, `generateElementWithRefs`, `type elementOptions`, `buildElementOptions`, `getClassAttributeValue`, `extractTextContent`, `var handlerAttributes`, `var watcherAttributes`, `var attributeToOption`, `generateAttributeOption`, `generateAttributeValue`
- L1191-1222: `textElementWithOptions`, `skipTextChildren` (note: these helper funcs may need to stay in generator.go if other files use them — check references)

**IMPORTANT:** `textElementWithOptions` and `skipTextChildren` (L1191-1222) are used by `generateElementWithRefs` in this file. Move them here, not in generator.go.

### `generator_control.go` — For loop, if statement, let binding generation (~210 lines)
Move:
- L717-863: `generateLetBinding`, `generateForLoop`, `generateForLoopWithRefs`, `generateIfStmt`, `generateIfStmtWithRefs`

### `generator_children.go` — Children rendering, body dispatch, slice building (~260 lines)
Move:
- L674-715: `generateChildren`, `generateChildrenWithRefs`
- L865-927: `generateBodyNode`, `generateBodyNodeWithRefs`, `generateGoCode`, `generateGoFunc`, `generateGoDecl`
- L929-1118: `generateComponentCall`, `generateComponentCallWithRefs`, `generateForLoopForSlice`, `generateIfStmtForSlice`

---

## Split 5: `internal/tuigen/analyzer.go` (1112 lines) → 4 files

### `analyzer.go` — Analyzer struct, known data, main Analyze, component validation (~500 lines)
Keep:
- L1-238: package, imports, types (`StateVar`, `StateBinding`, `NamedRef`, `Analyzer`), `NewAnalyzer`, all `var known*` maps, all regexps, `var attributeSimilar`, `func Analyze`
- L481-757: `Errors`, `analyzeComponent`, `analyzeNode`, `analyzeElement`, `analyzeAttribute`, `analyzeLetBinding`, `analyzeForLoop`, `analyzeIfStmt`, `analyzeComponentCall`, `analyzeGoExpr`, `analyzeGoCode`, `addMissingImports`, `AnalyzeFile`, `ValidateElement`, `ValidateAttribute`, `SuggestAttribute`

### `analyzer_refs.go` — Named ref validation, inference, let-binding transformation (~240 lines)
Move:
- L240-479: `validateNamedRefs`, `isValidRefName`, `inferKeyType`, `CollectNamedRefs`, `collectLetBindings`, `containsChildrenSlot`, `transformElementRefs`, `transformNode`, `isSimpleIdentifier`, `isIdentLetter`, `isIdentDigit`

### `analyzer_imports.go` — Import management (~20 lines)
Move:
- L703-721: `addMissingImports`

**NOTE:** `addMissingImports` is only ~20 lines. If this feels too small for its own file, merge it into `analyzer.go` and skip creating `analyzer_imports.go`. The spec lists it but practical judgment says keep it in `analyzer.go`.

### `analyzer_state.go` — State detection, binding detection, deps parsing (~260 lines)
Move:
- L759-1112: `DetectStateVars`, `parseStateDeclarations`, `inferTypeFromExpr`, `DetectStateBindings`, `parseExplicitDeps`, `detectGetCalls`

---

## Split 6: `internal/tuigen/lexer.go` (924 lines) → 4 files

### `lexer.go` — Lexer struct, init, Next(), position tracking (~280 lines)
Keep:
- L1-6: package, imports
- L8-276: `type Lexer`, `NewLexer`, `Errors`, `readChar`, `peekChar`, `startToken`, `makeToken`, `position`, `Next`

### `lexer_strings.go` — String/rune/number literal reading (~160 lines)
Move:
- L439-589: `readString`, `readRune`, `readRawString`, `readNumber`

### `lexer_goexpr.go` — Go expression reading, balanced braces, peek (~230 lines)
Move:
- L591-685: `ReadGoExpr`, `skipStringInExpr`, `skipRawStringInExpr`, `skipCharInExpr`
- L698-725: `PeekToken`
- L761-924: `ReadBalancedBraces`, `ReadUntilBrace`, `ReadBalancedBracesFrom`

### `lexer_utils.go` — Comments, identifiers, whitespace, utilities (~160 lines)
Move:
- L278-437: `skipWhitespaceAndCollectComments`, `skipWhitespaceOnly`, `collectLineComment`, `collectBlockComment`, `hadBlankLineBefore`, `ConsumeComments`, `readIdentifier`, `readAtKeyword`
- L687-759: `isLetter`, `isDigit`, `CurrentChar`, `SkipWhitespace`, `SourcePos`, `SourceRange`

---

## Split 7: `internal/tuigen/tailwind.go` (929 lines) → 4 files

### `tailwind.go` — Core parsing functions + accumulator types (~365 lines)
Keep:
- L1-7: package, imports
- L146-303: `PaddingAccumulator` (type + methods), `MarginAccumulator` (type + methods), `IndividualSpacingResult`, `parseIndividualSpacing`
- L305-508: `ParseTailwindClass`, `TailwindParseResult`, `ParseTailwindClasses`, `BuildTextStyleOption`

### `tailwind_data.go` — Static class map, regex patterns (~140 lines)
Move:
- L10-143: `type TailwindMapping`, `var tailwindClasses`, regex pattern vars

### `tailwind_validation.go` — Validation, fuzzy matching, Levenshtein (~230 lines)
Move:
- L510-515: `type TailwindValidationResult`
- L525-532: `type TailwindClassWithPosition`
- L534-738: `var similarClasses`, `levenshteinDistance`, `getAllKnownClassNames`, `findSimilarClass`, `ValidateTailwindClass`, `ParseTailwindClassesWithPositions`

### `tailwind_autocomplete.go` — AllTailwindClasses documentation data (~190 lines)
Move:
- L517-523: `type TailwindClassInfo`
- L740-929: `func AllTailwindClasses`

---

## Split 8: `internal/lsp/provider/semantic.go` (1382 lines) → 3 files

### `semantic.go` — Types, constants, provider, encoding (~250 lines)
Keep:
- L1-76: package, imports, regexps, const blocks, types (`SemanticTokens`, `SemanticToken`, `FunctionNameChecker`), `type semanticTokensProvider`, `NewSemanticTokensProvider`
- L78-226: `SemanticTokensFull`, `collectSemanticTokens`
- L1205-1260: `EncodeSemanticTokens`, `isDigit`, `isHexDigit`, `isWordStartChar`, `isWordCharByte`, `isValidIdentifier`

### `semantic_nodes.go` — AST node processing, comments (~530 lines)
Move:
- L228-489: `collectTokensFromNodes`, `collectTokensFromNode`
- L768-821: `emitStringWithFormatSpecifiers`
- L883-1203: `findElseKeyword`, `collectAllCommentTokens`, `collectComponentCommentTokens`, `collectNodeCommentTokens`, `collectCommentGroupTokens`, `collectCommentToken`, `emitGoTypeTokens`

### `semantic_gocode.go` — Go expression tokenization, variable extraction (~400 lines)
Move:
- L491-766: `collectVariableTokensInCode`, `collectTokensInGoCodeDirect`, `collectTokensInGoCode`
- L823-881: `collectTokensFromFuncBody`
- L1262-1382: `type varDecl`, `extractVarDeclarationsWithPositions`, `type funcParam`, `parseFuncSignatureForTokens`

---

## Split 9: `internal/lsp/provider/references.go` (874 lines) → 2 files

### `references.go` — Main provider, reference dispatch (~565 lines)
Keep:
- L1-565: package, imports, `type referencesProvider`, `NewReferencesProvider`, `References`, `findComponentReferences`, `findFunctionReferences`, `findParamReferences`, `findFuncParamReferences`, `findLocalVariableReferences`, `findLoopVariableReferences`, `findGoCodeVariableReferences`, `findNamedRefReferences`, `findNamedRefDeclInNodes`, `findStateVarReferences`, `findStateVarDeclInNodes`

### `references_search.go` — Cross-file search, workspace scanning (~310 lines)
Move:
- L567-874: `searchWorkspaceForComponentRefs`, `searchWorkspaceForFunctionRefs`, `findComponentCallsInNodes`, `findFunctionCallsInNodes`, `findFuncCallInCode`, `findVariableUsagesInNodes`, `findVariableUsagesInNodesExcluding`, `findVariableInCode`, `findVariableInCodeExcluding`, `indexWholeWord`

---

## Split 10: `internal/lsp/context.go` (837 lines) → 2 files

### `context.go` — CursorContext struct, NodeKind, Scope, entry point (~150 lines)
Keep:
- L1-146: package, imports, `type NodeKind`, const block, `NodeKind.String`, `type Scope`, `type CursorContext`, `ResolveCursorContext`

### `context_resolve.go` — AST walking, classification, scope building (~690 lines)
Move:
- L148-837: `resolveFromAST`, `resolveInNodes`, `resolveInNode`, `resolveInNodeInner`, `resolveInElement`, `resolveInForLoop`, `resolveInIfStmt`, `resolveInLetBinding`, `resolveInComponentCall`, `classifyGoExpr`, `classifyGoCode`, `classifyFromText`, `collectScopeFromBody`, `collectScopeFromBodyInner`, `getLineText`, `getWordAtOffset`, `isOffsetInGoExpr`, `const maxClassAttrSearchDistance`, `isOffsetInClassAttr`, `findComponentEndLine`, `isOffsetInElementTag`, `findFuncParamAtColumn`

---

## Split 11: `internal/lsp/provider/definition.go` (741 lines) → 2 files

### `definition.go` — Main provider, definition dispatch (~505 lines)
Keep:
- L1-503: package, imports, `type definitionProvider`, `NewDefinitionProvider`, `Definition`, `definitionComponentCall`, `definitionNamedRefFromScope`, `definitionNamedRef`, `findNamedRefPosition`, `definitionStateVar`, `definitionEventHandler`, `definitionParameter`, `findLetBindingDefinition`, `findLoopVariableDefinition`, `findGoCodeVariableDefinition`, `findComponentInAST`, `findFuncInAST`, `getGoplsDefinition`

### `definition_search.go` — Cross-file search helpers (~240 lines)
Move:
- L505-741: `findLetBindingInNodes`, `findForLoopWithVariable`, `findGoCodeWithVariable`, `containsVarDecl`, `isWordBoundary`, `findVarDeclPosition`, `indexWholeWordIn`, `parseFuncName`

---

## Split 12: `internal/lsp/provider/completion.go` (587 lines) → 2 files

### `completion.go` — Main provider, types, dispatch (~190 lines)
Keep:
- L1-189: package, imports, types (`CompletionList`, `CompletionItem`, `CompletionItemKind`, `CompletionContext`), const block, `type completionProvider`, `NewCompletionProvider`, `Complete`, `triggerChar`, `const maxClassAttrSearchDistance`, `classPrefix`

### `completion_items.go` — Completion item builders (~400 lines)
Move:
- L191-587: `getComponentCompletions`, `getDSLKeywordCompletions`, `getElementCompletions`, `getAttributeCompletions`, `getContextualCompletions`, `enclosingTagFromText`, `getStateMethodCompletions`, `getTailwindCompletions`, `sortCompletionsByCategory`, `getGoplsCompletions`

---

## Split 13: `internal/formatter/printer.go` (852 lines) → 4 files

### `printer.go` — Printer struct, file/package/component printing, node dispatch (~220 lines)
Keep:
- L1-22: package, imports, `type printer`, `newPrinter`
- L25-217: `PrintFile`, `printPackage`, `printImports`, `printComponent`, `printBody`, `hasBlankLineBefore`, `printNode`
- L617-631: `write`, `newline`, `writeIndent`

### `printer_elements.go` — Element printing with attributes (~160 lines)
Move:
- L219-378: `printElement`, `printAttribute`, `printAttrValue`, `canStructurallyInline`, `printChildrenInline`

### `printer_control.go` — Control flow + component call printing (~240 lines)
Move:
- L380-615: `printForLoop`, `printIfStmt`, `printElseBranch`, `printLetBinding`, `printComponentCall`, `printGoFunc`

### `printer_comments.go` — Comment formatting and printing (~220 lines)
Move:
- L633-852: `escapeString`, `formatBlockComment`, `formatLineComment`, `formatComment`, `formatInlineBlockComments`, `printCommentGroup`, `printLeadingComments`, `printTrailingComment`, `printOrphanComments`

---

## Split 14: `internal/lsp/gopls/proxy.go` (564 lines) → 2 files

### `proxy.go` — GoplsProxy struct, types, lifecycle, communication (~330 lines)
Keep:
- L1-195: package, imports, all type definitions, `NewGoplsProxy`
- L197-246: `Initialize`, `Shutdown`
- L436-564: `send`, `readResponses`, `readMessage`, `TuiURIToGoURI`, `GoURIToTuiURI`, `IsVirtualGoFile`, `GetVirtualFilePath`

### `proxy_requests.go` — LSP request methods (~200 lines)
Move:
- L248-434: `OpenVirtualFile`, `UpdateVirtualFile`, `CloseVirtualFile`, `Completion`, `Hover`, `Definition`, `call`, `notify`

---

## Split 15: `internal/lsp/gopls/generate.go` (557 lines) → 2 files

### `generate.go` — Virtual Go file generation core (~460 lines)
Keep:
- L1-143: package, imports, `stateNewStateRegex`, `GenerateVirtualGo`, `type generator`, `generate`, `generateComponent`
- L248-557: `generateNodes`, `generateNode`, `generateElement`, `generateGoExpr`, `generateGoCode`, `generateRawGoExpr`, `generateForLoop`, `generateIfStmt`, `generateLetBinding`, `generateComponentCall`, `generateFunc`, `writeLine`

### `generate_state.go` — State variable and named ref emission (~100 lines)
Move:
- L145-246: `emitStateVarDeclarations`, `emitNamedRefDeclarations`, `emitNamedRefFromNodes`

---

## Final Verification

After all 15 splits:

```bash
go build ./...           # Must pass — no compilation errors
go test ./...            # Must pass — no test changes needed
```

Then update `specs/restructure-plan.md`:
- Mark all Phase 4 checkboxes as `[x]`
- Fill in the commit hash on the "Completed in commit" line

Commit with:
```bash
gcommit -m "refactor: Phase 4 - split oversized source files (≤500 lines each)"
```
