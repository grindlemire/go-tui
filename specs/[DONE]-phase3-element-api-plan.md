# Phase 3: Element API Implementation Plan

Implementation phases for the Element API. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Layout Interface + Element Core

**Reference:** [phase3-element-api-design.md §3.1, §3.3, §4](./phase3-element-api-design.md#31-layoutable-interface-layout-package)

**Status:** COMPLETE

- [x] Create `pkg/layout/layoutable.go`
  - Define `Layoutable` interface with all methods
  - See [design §3.1](./phase3-element-api-design.md#31-layoutable-interface-layout-package)
  ```go
  type Layoutable interface {
      LayoutStyle() Style
      LayoutChildren() []Layoutable
      SetLayout(Layout)
      GetLayout() Layout
      IsDirty() bool
      SetDirty(bool)
  }
  ```

- [x] Create `pkg/layout/layout.go`
  - Move `Layout` struct from `node.go` to its own file
  - Keep `Rect` and `ContentRect` fields
  - See [design §3.2](./phase3-element-api-design.md#32-layout-result-type)

- [x] Refactor `pkg/layout/calculate.go`
  - Change `Calculate(node *Node, ...)` to `Calculate(root Layoutable, ...)`
  - Update `calculateNode` to work with `Layoutable` interface
  - Replace `node.Style` with `node.LayoutStyle()`
  - Replace `node.Children` with `node.LayoutChildren()`
  - Replace `node.Layout = ...` with `node.SetLayout(...)`
  - Replace dirty checks with interface methods
  - See [design §4.1](./phase3-element-api-design.md#41-calculate-function-refactored)

- [x] Refactor `pkg/layout/flex.go`
  - Change `flexItem.node` from `*Node` to `Layoutable`
  - Update `layoutChildren` to accept `Layoutable` parent
  - Use `child.LayoutStyle()` instead of `child.Style`
  - See [design §4.2](./phase3-element-api-design.md#42-flex-algorithm-refactored)

- [x] Remove `pkg/layout/node.go`
  - Delete the file entirely
  - Node type is replaced by Layoutable interface

- [x] Update `pkg/layout/node_test.go` → `pkg/layout/layoutable_test.go`
  - Create a minimal test implementation of Layoutable
  - Port relevant tests to use the test implementation
  - Ensure all existing layout behaviors still work

- [x] Create `pkg/tui/element/element.go`
  - Define `Element` struct with children, style, layout, dirty, visual properties
  - Implement `Layoutable` interface methods
  - Implement `New(opts ...Option) *Element`
  - Implement `AddChild`, `RemoveChild`, `Children`
  - Implement `MarkDirty` with parent propagation
  - Implement `Calculate`, `Rect`, `ContentRect`
  - See [design §3.3](./phase3-element-api-design.md#33-element-implements-layoutable)

- [x] Create `pkg/tui/element/options.go`
  - Define `Option` type as `func(*Element)`
  - Implement all dimension options: `WithWidth`, `WithWidthPercent`, `WithHeight`, `WithHeightPercent`, `WithSize`, `WithMinWidth`, `WithMinHeight`, `WithMaxWidth`, `WithMaxHeight`
  - Implement flex container options: `WithDirection`, `WithJustify`, `WithAlign`, `WithGap`
  - Implement flex item options: `WithFlexGrow`, `WithFlexShrink`, `WithAlignSelf`
  - Implement spacing options: `WithPadding`, `WithPaddingTRBL`, `WithMargin`, `WithMarginTRBL`
  - Implement visual options: `WithBorder`, `WithBorderStyle`, `WithBackground`
  - See [design §3.4](./phase3-element-api-design.md#34-option-functional-options)

- [x] Create `pkg/tui/element/element_test.go`
  - Test `New()` creates Element with default Auto dimensions
  - Test all `WithX` options correctly set style properties
  - Test `AddChild` / `RemoveChild` / `Children`
  - Test `MarkDirty` propagates to ancestors
  - Test `Calculate` computes correct layout

- [x] Create `pkg/tui/element/options_test.go`
  - Test each option function individually
  - Verify options compose correctly

**Tests:** All tests pass (`go test ./pkg/layout/... ./pkg/tui/element/...`)

---

## Phase 2: Text Element + Rendering

**Reference:** [phase3-element-api-design.md §3.5, §5](./phase3-element-api-design.md#35-text-element)

**Status:** COMPLETE

- [x] Create `pkg/tui/element/text.go`
  - Define `Text` struct embedding `*Element`
  - Add `content`, `contentStyle`, `align` fields
  - Define `TextAlign` type and constants
  - Implement `NewText(content string, opts ...TextOption) *Text`
  - Implement `SetContent`, `Content` methods
  - Define `TextOption` type
  - Implement `WithTextStyle`, `WithTextAlign`, `WithElementOption`
  - See [design §3.5](./phase3-element-api-design.md#35-text-element)

- [x] Create `pkg/tui/element/render.go`
  - Implement `RenderTree(buf *tui.Buffer, root *Element)`
  - Implement `renderElement` that draws background, border, recurses children
  - Implement `renderText` that draws text content with alignment
  - Handle buffer bounds checking (skip elements outside visible area)
  - See [design §5.1](./phase3-element-api-design.md#51-render-implementation)

- [x] Add `Render` method to Element
  - Implement `func (e *Element) Render(buf *tui.Buffer, width, height int)`
  - Call `Calculate` if dirty, then `RenderTree`
  - See [design §5.2](./phase3-element-api-design.md#52-elementrender-method)

- [x] Add `Intersects` method to `layout.Rect` if missing
  - Needed for render culling optimization
  - `func (r Rect) Intersects(other Rect) bool`

- [x] Create `pkg/tui/element/text_test.go`
  - Test `NewText` creates Text with content
  - Test `SetContent` / `Content`
  - Test text options work correctly
  - Test text alignment rendering

- [x] Create `pkg/tui/element/render_test.go`
  - Test `RenderTree` draws elements correctly
  - Test background fills interior
  - Test border draws at correct positions
  - Test nested elements render in correct order
  - Test culling works (elements outside buffer not drawn)

**Tests:** All tests pass (`go test ./pkg/tui/element/...`)

---

## Phase 3: Integration + Examples

**Reference:** [phase3-element-api-design.md §6](./phase3-element-api-design.md#6-user-experience)

**Status:** COMPLETE

- [x] Create `pkg/tui/element/integration_test.go`
  - Test complete flow: New → AddChild → Render
  - Test nested layouts (row inside column, etc.)
  - Test flex grow/shrink behavior
  - Test mixed Element and Text children
  - Compare rendered output against expected snapshots

- [x] Update `examples/dashboard/main.go`
  - Rewrite using Element API
  - Demonstrate significant code reduction
  - Show border, text, centering
  - See [design §6.1](./phase3-element-api-design.md#61-complete-example)

- [x] Update/remove `pkg/layout/integration_test.go`
  - Tests already use testNode (test Layoutable implementation)
  - No changes needed - tests remain valid

- [x] Verify all existing tests pass
  - Run `go test ./...` from repo root
  - All tests pass

- [x] Clean up old Node references
  - Search for any remaining `layout.Node` usage
  - Updated stale comment in layoutable_test.go

**Tests:** All tests pass (`go test ./...`)

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Layout interface refactor + Element core + options | COMPLETE |
| 2 | Text element + rendering implementation | COMPLETE |
| 3 | Integration tests + examples update | COMPLETE |

## Files Created

```
pkg/layout/
├── layoutable.go       # NEW: Layoutable interface
├── layout.go           # NEW: Layout struct (moved from node.go)
├── layoutable_test.go  # NEW: Interface tests
├── calculate.go        # MODIFIED: Use Layoutable
├── flex.go             # MODIFIED: Use Layoutable
└── node.go             # DELETED

pkg/tui/element/
├── element.go          # NEW
├── element_test.go     # NEW
├── options.go          # NEW
├── options_test.go     # NEW
├── text.go             # NEW
├── text_test.go        # NEW
├── render.go           # NEW
├── render_test.go      # NEW
└── integration_test.go # NEW
```

## Files Modified

| File | Changes |
|------|---------|
| `pkg/layout/calculate.go` | Changed to work with Layoutable interface |
| `pkg/layout/flex.go` | Changed flexItem.node to Layoutable |
| `pkg/layout/rect.go` | Intersects method already existed |
| `examples/dashboard/main.go` | Rewritten with Element API |
