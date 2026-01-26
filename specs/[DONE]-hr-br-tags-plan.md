# HR and BR Tags Implementation Plan

Implementation phases for hr/br tags. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Analyzer and Generator Changes

**Reference:** [hr-br-tags-design.md §3.1, §3.7, §3.8](./hr-br-tags-design.md#31-void-element-set)

**Review:** false

**Completed in commit:** (pending)

- [x] Modify `pkg/tuigen/analyzer.go`
  - Add `hr` and `br` to `knownTags` map (line ~35)
  - Add `voidElements` map with `hr`, `br`, `input`
  - Update `analyzeElement()` to validate void elements have no children
  - See [design §3.1](./hr-br-tags-design.md#31-void-element-set) and [design §3.8](./hr-br-tags-design.md#38-analyzer-validation)

- [x] Modify `pkg/tuigen/generator.go`
  - Update `buildElementOptions()` switch statement to handle `hr` and `br` tags
  - `hr` → append `element.WithHR()`
  - `br` → append `element.WithWidth(0)`, `element.WithHeight(1)`
  - See [design §3.7](./hr-br-tags-design.md#37-generator-changes)

- [x] Add analyzer tests in `pkg/tuigen/analyzer_test.go`
  - `TestAnalyzeHRValid`: `<hr/>` passes validation
  - `TestAnalyzeBRValid`: `<br/>` passes validation
  - `TestAnalyzeVoidWithChildren`: `<hr>text</hr>` produces error

- [x] Add generator tests in `pkg/tuigen/generator_test.go`
  - `TestGenerateHR`: verify `<hr/>` generates `element.WithHR()`
  - `TestGenerateBR`: verify `<br/>` generates width/height options
  - `TestGenerateHRWithBorder`: verify `<hr class="border-double"/>` includes border
  - `TestGenerateHRWithTextClass`: verify `<hr class="text-cyan"/>` generates TextStyle

**Tests:** Run `go test ./pkg/tuigen/...` at phase end ✓

---

## Phase 2: Element Options and Rendering

**Reference:** [hr-br-tags-design.md §3.2, §3.3, §3.4, §3.5, §3.6](./hr-br-tags-design.md#32-element-struct-changes)

**Completed in commit:** (pending)

- [x] Modify `pkg/tui/element/element.go`
  - Add `hr bool` field to `Element` struct
  - Add `IsHR() bool` method
  - Update `IntrinsicSize()` to return `(0, 1)` for HR elements
  - See [design §3.2](./hr-br-tags-design.md#32-element-struct-changes) and [design §3.3](./hr-br-tags-design.md#33-intrinsicsize-for-hrbr)

- [x] Modify `pkg/tui/element/options.go`
  - Add `WithHR()` option function
  - Set `e.hr = true`, `e.style.Height = layout.Fixed(1)`, `e.style.AlignSelf = &layout.AlignStretch`
  - See [design §3.4](./hr-br-tags-design.md#34-withhr-option)

- [x] Modify `pkg/tui/element/render.go`
  - Add `hrCharacter(border tui.BorderStyle) rune` helper function
  - Add `renderHR(buf *tui.Buffer, e *Element)` function
  - Update `renderElement()` to check `e.hr` and call `renderHR()` early return
  - See [design §3.5](./hr-br-tags-design.md#35-hr-character-mapping) and [design §3.6](./hr-br-tags-design.md#36-hr-rendering)

- [x] Add render tests in `pkg/tui/element/render_test.go`
  - `TestRenderHRDefault`: HR draws `─` characters
  - `TestRenderHRDouble`: HR with BorderDouble draws `═`
  - `TestRenderHRThick`: HR with BorderThick draws `━`
  - `TestRenderHRWithColor`: HR respects textStyle for color

**Tests:** Run `go test ./pkg/tui/element/...` at phase end ✓

---

## Phase 3: LSP Completions and Integration

**Reference:** [hr-br-tags-design.md §3.9](./hr-br-tags-design.md#39-lsp-completion)

**Review:** false

**Completed in commit:** (pending)

- [x] Modify `pkg/lsp/completion.go`
  - Add `hr` completion item to `getElementCompletions()` with proper InsertText/Documentation
  - Add `br` completion item to `getElementCompletions()` with proper InsertText/Documentation
  - See [design §3.9](./hr-br-tags-design.md#39-lsp-completion)

- [x] Create integration test file (optional)
  - Verify full flow: `.tui` file → parse → analyze → generate → compile → render
  - Test `<hr/>`, `<hr class="border-double text-cyan"/>`, `<br/>`

- [x] Run all tests
  - `go test ./...`
  - Verify no regressions

**Tests:** Run `go test ./...` at phase end ✓

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Analyzer and generator changes + tests | Complete |
| 2 | Element struct, options, rendering + tests | Complete |
| 3 | LSP completions and integration | Complete |

## Files to Create

None - all changes are modifications to existing files.

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/tuigen/analyzer.go` | Add hr/br to knownTags, add voidElements, validate void children |
| `pkg/tuigen/generator.go` | Handle hr/br in buildElementOptions() |
| `pkg/tuigen/analyzer_test.go` | Add void element validation tests |
| `pkg/tuigen/generator_test.go` | Add hr/br generation tests |
| `pkg/tui/element/element.go` | Add hr field, IsHR(), update IntrinsicSize() |
| `pkg/tui/element/options.go` | Add WithHR() option |
| `pkg/tui/element/render.go` | Add hrCharacter(), renderHR(), update renderElement() |
| `pkg/tui/element/render_test.go` | Add HR rendering tests |
| `pkg/lsp/completion.go` | Add hr/br to getElementCompletions() |
