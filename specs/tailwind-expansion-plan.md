# Tailwind Expansion Implementation Plan

Implementation phases for Tailwind expansion. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Expand Tailwind Class Mappings ✓

**Reference:** [tailwind-expansion-design.md §3](./tailwind-expansion-design.md#3-core-entities)

**Status:** Complete

- [x] Create `pkg/tui/element/options_auto.go`
  - Add `WithWidthAuto()` option that sets `e.style.Width = layout.Auto()`
  - Add `WithHeightAuto()` option that sets `e.style.Height = layout.Auto()`
  - Keep minimal, just these two functions

- [x] Modify `pkg/tuigen/tailwind.go` - Add new regex patterns
  - Add `widthFractionPattern = regexp.MustCompile("^w-(\\d+)/(\\d+)$")`
  - Add `heightFractionPattern = regexp.MustCompile("^h-(\\d+)/(\\d+)$")`
  - Add `widthKeywordPattern = regexp.MustCompile("^w-(full|auto)$")`
  - Add `heightKeywordPattern = regexp.MustCompile("^h-(full|auto)$")`
  - Add individual padding patterns: `ptPattern`, `prPattern`, `pbPattern`, `plPattern`
  - Add individual margin patterns: `mtPattern`, `mrPattern`, `mbPattern`, `mlPattern`, `mxPattern`, `myPattern`
  - Add `flexGrowPattern = regexp.MustCompile("^flex-grow-(\\d+)$")`
  - Add `flexShrinkPattern = regexp.MustCompile("^flex-shrink-(\\d+)$")`

- [x] Modify `pkg/tuigen/tailwind.go` - Add new static mappings
  - Add width/height keywords: `w-full`, `w-auto`, `h-full`, `h-auto`
  - Add flex utilities: `justify-evenly`, `justify-around`, `items-stretch`
  - Add self-alignment: `self-start`, `self-end`, `self-center`, `self-stretch`
  - Add text alignment: `text-left`, `text-center`, `text-right`
  - Add border colors: `border-red`, `border-green`, `border-blue`, `border-cyan`, `border-magenta`, `border-yellow`, `border-white`, `border-black`

- [x] Modify `pkg/tuigen/tailwind.go` - Implement `PaddingAccumulator` and `MarginAccumulator`
  - Add `PaddingAccumulator` struct with `Top, Right, Bottom, Left int` and `HasTop, HasRight, HasBottom, HasLeft bool`
  - Add `MarginAccumulator` struct with same fields
  - Add `(p *PaddingAccumulator) Merge(side string, value int)` method
  - Add `(p *PaddingAccumulator) ToOption() string` method - generates `WithPaddingTRBL(T, R, B, L)`
  - Same methods for `MarginAccumulator`

- [x] Modify `pkg/tuigen/tailwind.go` - Update `ParseTailwindClass()`
  - Handle width fraction patterns (w-1/2, w-2/3, etc.) → `WithWidthPercent(percent)`
  - Handle height fraction patterns → `WithHeightPercent(percent)`
  - Handle width/height keywords (w-full → 100%, w-auto → Auto option)
  - Handle individual padding patterns (pt-N, pr-N, pb-N, pl-N) → return marker for accumulation
  - Handle individual margin patterns (mt-N, mr-N, mb-N, ml-N, mx-N, my-N) → return marker for accumulation
  - Handle flex-grow-N and flex-shrink-N patterns

- [x] Modify `pkg/tuigen/tailwind.go` - Update `ParseTailwindClasses()`
  - Add padding/margin accumulation logic
  - Collect individual side classes during iteration
  - After loop, call accumulators' `ToOption()` if any sides were set
  - Ensure existing `p-N`, `px-N`, `py-N`, `m-N` continue to work
  - Note: when both `p-2` and `pt-4` are used, `p-2` sets all sides, then `pt-4` overrides top only

- [x] Add tests in `pkg/tuigen/tailwind_test.go`
  - Test width fractions: `w-1/2` → 50%, `w-1/3` → 33.33%, `w-2/3` → 66.67%
  - Test height fractions: `h-1/2`, `h-1/4`, `h-3/4`
  - Test keywords: `w-full` → 100%, `w-auto`, `h-full`, `h-auto`
  - Test individual padding: `pt-2`, `pr-3`, `pb-4`, `pl-1`
  - Test padding accumulation: `pt-2 pb-4` → single `WithPaddingTRBL(2, 0, 4, 0)`
  - Test individual margins with accumulation
  - Test new flex utilities: `self-start`, `justify-evenly`, `items-stretch`
  - Test border colors: `border-red`, `border-cyan`
  - Test text alignment: `text-center`, `text-right`
  - Test flex patterns: `flex-grow-0`, `flex-grow-2`, `flex-shrink-0`

- [x] Add test in `pkg/tui/element/options_test.go`
  - Test `WithWidthAuto()` sets `style.Width` to `layout.Auto()`
  - Test `WithHeightAuto()` sets `style.Height` to `layout.Auto()`

**Tests:** Run `go test ./pkg/tuigen/... ./pkg/tui/element/...` once at phase end

---

## Phase 2: Add Validation and Similarity Matching ✓

**Reference:** [tailwind-expansion-design.md §3.1](./tailwind-expansion-design.md#31-technical-design-details)

**Status:** Complete

- [x] Modify `pkg/tuigen/tailwind.go` - Add validation types
  - Add `TailwindValidationResult` struct: `Valid bool`, `Class string`, `Suggestion string`
  - Add `TailwindClassInfo` struct: `Name string`, `Category string`, `Description string`, `Example string`
  - Add `TailwindClassWithPosition` struct: `Class string`, `StartCol int`, `EndCol int`, `Valid bool`, `Suggestion string`

- [x] Modify `pkg/tuigen/tailwind.go` - Add `similarClasses` map
  - Map common typos/alternatives to correct class names
  - Include: `flex-column` → `flex-col`, `bold` → `font-bold`, `center` → `text-center`, etc.
  - See design document §3 for full list

- [x] Modify `pkg/tuigen/tailwind.go` - Add similarity matching function
  - Add `findSimilarClass(class string) string` function
  - First check exact match in `similarClasses` map
  - Then use Levenshtein distance for fuzzy matching against all known classes
  - Return best match if distance ≤ 3, otherwise empty string

- [x] Modify `pkg/tuigen/tailwind.go` - Add `ValidateTailwindClass()` function
  - Attempt to parse the class with `ParseTailwindClass()`
  - If valid, return `TailwindValidationResult{Valid: true, Class: class}`
  - If invalid, call `findSimilarClass()` and return with suggestion

- [x] Modify `pkg/tuigen/tailwind.go` - Add `ParseTailwindClassesWithPositions()` function
  - Parse class string while tracking character positions
  - For each whitespace-separated class:
    - Record `StartCol` as offset from attribute value start
    - Record `EndCol` as `StartCol + len(class)`
    - Call `ValidateTailwindClass()` to get validity and suggestion
  - Return slice of `TailwindClassWithPosition`

- [x] Modify `pkg/tuigen/tailwind.go` - Add `AllTailwindClasses()` function
  - Return slice of `TailwindClassInfo` for all known classes
  - Include static classes from `tailwindClasses` map
  - Include pattern-based classes with examples (e.g., `gap-N` with description)
  - Categorize: "layout", "spacing", "typography", "visual", "flex"

- [x] Add Levenshtein distance helper in `pkg/tuigen/tailwind.go`
  - Add `levenshteinDistance(a, b string) int` function
  - Standard dynamic programming implementation
  - Used for fuzzy class name matching

- [x] Add tests in `pkg/tuigen/tailwind_test.go`
  - Test `ValidateTailwindClass()` with valid classes
  - Test `ValidateTailwindClass()` with invalid classes, check suggestions
  - Test `findSimilarClass()` with typos: `flex-columns` → `flex-col`
  - Test `ParseTailwindClassesWithPositions()` position tracking
  - Test `AllTailwindClasses()` returns expected count and categories
  - Test Levenshtein distance function

**Tests:** Run `go test ./pkg/tuigen/...` once at phase end

---

## Phase 3: Integrate Validation into Analyzer and LSP Diagnostics ✓

**Reference:** [tailwind-expansion-design.md §2](./tailwind-expansion-design.md#2-architecture)

**Status:** Complete

- [x] Modify `pkg/tuigen/analyzer.go` - Add class validation
  - In `analyzeAttribute()` when `attr.Name == "class"`:
    - Call `ParseTailwindClassesWithPositions()` on the class value
    - For each invalid class, create an error with:
      - Position adjusted to the class's location within the attribute
      - Message: `unknown Tailwind class "X"`
      - Hint: `did you mean "Y"?` if suggestion exists
  - Track attribute position for error reporting

- [x] Modify `pkg/tuigen/ast.go` - Add position tracking to Attribute
  - Ensure `Attribute` struct has `ValuePosition Position` field for the start of the value
  - This is needed to calculate individual class positions within the value string

- [x] Modify `pkg/tuigen/parser.go` - Track attribute value positions
  - When parsing string literal attribute values, record the position
  - Store in `Attribute.ValuePosition`

- [x] Modify `pkg/tuigen/errors.go` - Extend error for class validation
  - Added `EndPos Position` field to Error struct for range-based highlighting
  - Added `NewErrorWithRange()` and `NewErrorWithRangeAndHint()` helper functions

- [x] Modify `pkg/lsp/diagnostics.go` - Handle class validation errors
  - Added `TuigenPosToRangeWithEnd()` function for precise range calculation
  - Updated `publishDiagnostics()` to use EndPos when available for range-based highlighting

- [x] Modify `pkg/lsp/document.go` - Ensure class errors propagate
  - Added analyzer invocation in `parseDocument()` to collect semantic errors
  - Class validation errors from analyzer now appear in `doc.Errors`
  - LSP publishes these via `publishDiagnostics()`

- [x] Add integration tests
  - Test that `.tui` file with `class="flex-columns"` produces error diagnostic
  - Test that error message includes "did you mean flex-col?"
  - Test that error range covers only the invalid class (EndPos is set)
  - Test that valid classes don't produce errors

**Tests:** Run `go test ./pkg/tuigen/... ./pkg/lsp/...` once at phase end ✓

---

## Phase 4: Add Class Autocomplete to LSP

**Reference:** [tailwind-expansion-design.md §4](./tailwind-expansion-design.md#4-user-experience)

**Completed in commit:** (pending)

- [ ] Modify `pkg/lsp/completion.go` - Add class attribute detection
  - Add `isInClassAttribute(doc *Document, pos Position) (bool, string)` function
    - Search backwards from cursor for `class="`
    - If found and cursor is before closing quote, return true
    - Extract partial class prefix (text after last space before cursor)
    - Return (true, prefix) or (false, "")

- [ ] Modify `pkg/lsp/completion.go` - Add class completions
  - Add `getTailwindCompletions(prefix string) []CompletionItem` function
  - Call `tuigen.AllTailwindClasses()` to get all available classes
  - Filter by prefix if provided
  - Convert each `TailwindClassInfo` to `CompletionItem`:
    - Label: class name
    - Kind: `CompletionItemKindConstant` or `CompletionItemKindValue`
    - Detail: category (e.g., "layout", "spacing")
    - Documentation: description + example usage
    - InsertText: class name (just the class, user adds space)
    - FilterText: class name for filtering

- [ ] Modify `pkg/lsp/completion.go` - Update `handleCompletion()`
  - Add check for class attribute context before other contextual completions
  - If `isInClassAttribute()` returns true:
    - Call `getTailwindCompletions(prefix)`
    - Return immediately with class completions
  - This should take priority over other completion types

- [ ] Modify `pkg/lsp/completion.go` - Group completions by category
  - Sort completions: layout → flex → spacing → typography → visual
  - Or alphabetically within categories
  - Ensure consistent ordering for user experience

- [ ] Add tests in `pkg/lsp/completion_test.go` (if exists, else create)
  - Test `isInClassAttribute()` detects class context
  - Test `isInClassAttribute()` extracts prefix correctly
  - Test `isInClassAttribute()` returns false outside class attribute
  - Test `getTailwindCompletions("")` returns all classes
  - Test `getTailwindCompletions("flex")` filters to flex-related classes
  - Test completion items have correct documentation

- [ ] Manual testing checklist
  - Open `.tui` file in VSCode with LSP running
  - Type `class="` and verify completions appear
  - Type `class="flex-` and verify filtered completions
  - Verify each completion has description
  - Type invalid class, verify red underline appears
  - Verify "did you mean" hint shows on hover/in problems panel

**Tests:** Run `go test ./pkg/lsp/...` once at phase end

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Expand Tailwind class mappings with new patterns and accumulation | ✓ Complete |
| 2 | Add validation, similarity matching, and class info registry | ✓ Complete |
| 3 | Integrate validation into analyzer and LSP diagnostics | ✓ Complete |
| 4 | Add class autocomplete to LSP completion handler | Pending |

## Files to Create

```
pkg/tui/element/
└── options_auto.go        # WithWidthAuto, WithHeightAuto
```

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/tuigen/tailwind.go` | New patterns, static mappings, accumulators, validation, similarity, AllTailwindClasses |
| `pkg/tuigen/tailwind_test.go` | Tests for all new functionality |
| `pkg/tuigen/analyzer.go` | Class validation integration |
| `pkg/tuigen/ast.go` | ValuePosition on Attribute |
| `pkg/tuigen/parser.go` | Track attribute value positions |
| `pkg/tuigen/errors.go` | Added EndPos field and range-based error functions |
| `pkg/lsp/completion.go` | Class attribute detection, Tailwind completions |
| `pkg/lsp/diagnostics.go` | Precise range calculation for class errors |
| `pkg/lsp/document.go` | Ensure class errors propagate |
| `pkg/tui/element/options_auto.go` | New file for auto width/height |
| `pkg/tui/element/options_test.go` | Tests for auto options |
