# Phase 2: Layout Engine Implementation Plan

Implementation phases for the layout engine. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Foundation Types & Migration

**Reference:** [phase2-layout-design.md §3.1-3.2](./phase2-layout-design.md#31-rect)
**Review:** false

**Completed in commit:** 7c8c873

- [x] Create `pkg/layout/rect.go`
  - Define `Rect` struct with `X, Y, Width, Height int`
  - Implement `NewRect(x, y, width, height int) Rect`
  - Implement edge accessors: `Right()`, `Bottom()`
  - Implement queries: `IsEmpty()`, `Area()`, `Contains(x, y)`, `ContainsRect(other)`
  - Implement transformations: `Inset(Edges)`, `Outset(Edges)`, `Translate(dx, dy)`, `Intersect(other)`, `Union(other)`, `Clamp(x, y)`
  - See [phase2-layout-design.md §3.1](./phase2-layout-design.md#31-rect)

- [x] Create `pkg/layout/point.go`
  - Define `Point` struct with `X, Y int`
  - Implement `Add(other)`, `Sub(other)`, `In(Rect)`
  - See [phase2-layout-design.md §3.1.1](./phase2-layout-design.md#311-point)

- [x] Create `pkg/layout/edges.go`
  - Define `Edges` struct with `Top, Right, Bottom, Left int`
  - Implement constructors: `EdgeAll(n)`, `EdgeSymmetric(v, h)`, `EdgeTRBL(t, r, b, l)`
  - Implement methods: `Horizontal()`, `Vertical()`, `IsZero()`
  - See [phase2-layout-design.md §3.2](./phase2-layout-design.md#32-edges)

- [x] Create `pkg/layout/value.go`
  - Define `Unit` enum: `UnitAuto`, `UnitFixed`, `UnitPercent`
  - Define `Value` struct with `Amount float64`, `Unit Unit`
  - Implement constructors: `Auto()`, `Fixed(n)`, `Percent(p)`
  - Implement `Resolve(available, fallback int) int`
  - Implement `IsAuto() bool`
  - See [phase2-layout-design.md §3.3](./phase2-layout-design.md#33-value-dimension-system)

- [x] Migrate `pkg/tui` to use `layout.Rect`
  - Update `pkg/tui/rect.go` to re-export `layout.Rect` as type alias
  - Update all tui files importing Rect to work with layout.Rect
  - Remove duplicate Rect methods from tui (keep in layout)
  - Ensure all existing tests pass

- [x] Create `pkg/layout/rect_test.go` and `pkg/layout/value_test.go`
  - Table-driven tests for Rect methods
  - Table-driven tests for Value.Resolve with Fixed/Percent/Auto

**Tests:** Run `go test ./pkg/layout/... ./pkg/tui/...` once at phase end ✓

---

## Phase 2: Node & Basic Flex Algorithm

**Reference:** [phase2-layout-design.md §3.4-3.5, §4](./phase2-layout-design.md#34-style)

**Completed in commit:** c6a7e49

- [x] Create `pkg/layout/style.go`
  - Define `Direction` enum: `Row`, `Column`
  - Define `Justify` enum: `JustifyStart`, `JustifyEnd`, `JustifyCenter`, `JustifySpaceBetween`, `JustifySpaceAround`, `JustifySpaceEvenly`
  - Define `Align` enum: `AlignStart`, `AlignEnd`, `AlignCenter`, `AlignStretch`
  - Define `Style` struct with all fields from design
  - Implement `DefaultStyle() Style`
  - See [phase2-layout-design.md §3.4](./phase2-layout-design.md#34-style)

- [x] Create `pkg/layout/node.go`
  - Define `Layout` struct with `Rect`, `ContentRect`
  - Define `Node` struct with `Style`, `Children`, `Layout`, `dirty`, `parent`
  - Implement `NewNode(style Style) *Node`
  - Implement `AddChild(children ...*Node)` — sets parent, marks dirty
  - Implement `RemoveChild(child *Node) bool` — clears parent, marks dirty
  - Implement `SetStyle(style Style)` — marks dirty
  - Implement `MarkDirty()` — propagates up parent chain
  - Implement `IsDirty() bool`
  - See [phase2-layout-design.md §3.5](./phase2-layout-design.md#35-node)

- [x] Create `pkg/layout/calculate.go`
  - Implement `Calculate(node *Node, availableWidth, availableHeight int)`
  - Implement internal `calculateNode(node *Node, available Rect)`
  - Handle dirty check early return
  - Compute border box from style constraints
  - Compute content rect (border box minus padding)
  - Call `layoutChildren` for nodes with children
  - Store Layout and clear dirty flag
  - See [phase2-layout-design.md §4.2](./phase2-layout-design.md#42-algorithm-overview)

- [x] Create `pkg/layout/flex.go`
  - Define internal `flexItem` struct
  - Implement `layoutChildren(node *Node, contentRect Rect)`
  - Phase 1: Compute base sizes from Width/Height values
  - Phase 2: Distribute free space with FlexGrow/FlexShrink
  - Phase 6: Convert to rects (apply child margin) and recurse
  - Initially only handle `JustifyStart` and `AlignStretch`
  - See [phase2-layout-design.md §4.3](./phase2-layout-design.md#43-flexbox-calculation)

- [x] Create `pkg/layout/node_test.go`
  - Test dirty propagation through tree
  - Test AddChild/RemoveChild parent management
  - Test SetStyle marks dirty

- [x] Create `pkg/layout/calculate_test.go`
  - Test single node sizing (Fixed width/height)
  - Test single node with padding
  - Test two children in Row direction with fixed sizes
  - Test two children in Column direction with fixed sizes
  - Test FlexGrow distributes extra space
  - Test FlexShrink handles overflow
  - Test dirty node is recalculated, clean node is skipped

**Tests:** Run `go test ./pkg/layout/...` once at phase end

---

## Phase 3: Complete Flex Features

**Reference:** [phase2-layout-design.md §4.3-4.5](./phase2-layout-design.md#43-flexbox-calculation)

**Completed in commit:** 331623d

- [x] Implement all Justify modes in `pkg/layout/flex.go`
  - `JustifyStart` — pack at start (already done)
  - `JustifyEnd` — pack at end
  - `JustifyCenter` — center children
  - `JustifySpaceBetween` — space between, none at edges
  - `JustifySpaceAround` — space around each child
  - `JustifySpaceEvenly` — equal space everywhere
  - Add helper: `calculateJustifyOffset(justify, freeSpace, itemCount) int`
  - Add helper: `calculateJustifySpacing(justify, freeSpace, itemCount) int`
  - See [phase2-layout-design.md §4.4](./phase2-layout-design.md#44-justify-content-distribution)

- [x] Implement all Align modes in `pkg/layout/flex.go`
  - `AlignStart` — align to start of cross axis
  - `AlignEnd` — align to end of cross axis
  - `AlignCenter` — center on cross axis
  - `AlignStretch` — fill cross axis (already done)
  - Implement `AlignSelf` override (check `child.Style.AlignSelf != nil`)
  - Add helper: `calculateAlignOffset(align, crossSize, itemSize) int`
  - See [phase2-layout-design.md §4.5](./phase2-layout-design.md#45-align-itemsself-positioning)

- [x] Implement Gap support in `pkg/layout/flex.go`
  - Account for `style.Gap * (len(children) - 1)` in free space calculation
  - Add gap between children when positioning

- [x] Implement Min/Max constraints in `pkg/layout/flex.go`
  - Phase 3: Apply min/max after flex distribution
  - Add helpers: `resolveMin(style, isRow, available) int`
  - Add helpers: `resolveMax(style, isRow, available) int`
  - Clamp: `finalSize = clamp(flexSize, minSize, maxSize)`
  - If min > max, min wins
  - See [phase2-layout-design.md §6.4](./phase2-layout-design.md#64-minmax-constraints)

- [x] Implement Percent value resolution
  - Percentages resolve against parent's content area
  - Test nested percentages

- [x] Extend `pkg/layout/calculate_test.go`
  - Test all 6 Justify modes with 1, 2, 3+ children
  - Test all 4 Align modes
  - Test AlignSelf overriding AlignItems
  - Test Gap between children
  - Test Min/Max constraints clamping flex results
  - Test Percent values at various nesting levels

**Tests:** Run `go test ./pkg/layout/...` once at phase end ✓

---

## Phase 4: Edge Cases, Integration & Polish

**Reference:** [phase2-layout-design.md §6, §8](./phase2-layout-design.md#6-edge-cases-and-constraints)

**Completed in commit:** 1e4564b

- [x] Handle edge cases in `pkg/layout/flex.go` and `pkg/layout/calculate.go`
  - Zero/negative dimensions clamped to >= 0
  - Empty nodes (no children) size from Width/Height or collapse to 0x0
  - Overflow when shrink totals 0 (children clipped, no negative sizes)
  - Empty child list (no-op for layoutChildren)
  - See [phase2-layout-design.md §6](./phase2-layout-design.md#6-edge-cases-and-constraints)

- [x] Create integration tests in `pkg/layout/integration_test.go`
  - Dashboard layout: header (fixed), sidebar (fixed), main (grow), footer (fixed)
  - Nested flex: row containing columns containing rows
  - Form layout: labels (fixed) + inputs (grow) in column
  - Mixed direction: alternating Row/Column at different levels

- [x] Verify incremental layout efficiency
  - Test that modifying one leaf only recalculates path to root
  - Add calculation counter to verify clean subtrees skipped
  - Test that reading Layout does not mark dirty

- [x] Create benchmarks in `pkg/layout/benchmark_test.go`
  - Benchmark full tree layout (10, 100, 1000 nodes)
  - Benchmark incremental layout (single node change)
  - Verify allocation count per layout pass
  - Document results in comments

- [x] Final verification
  - Run `go test ./pkg/layout/... ./pkg/tui/...` — all tests pass ✓
  - Run `go vet ./pkg/layout/...` — no issues ✓
  - Verify `pkg/layout` has zero imports from `pkg/tui` ✓

**Tests:** Run full test suite and benchmarks at phase end ✓

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Foundation types (Rect, Point, Edges, Value) + tui migration | Complete |
| 2 | Node, Style, basic flex algorithm (Row/Column, Grow/Shrink) | Complete |
| 3 | All Justify/Align modes, Gap, Min/Max, Percent | Complete |
| 4 | Edge cases, integration tests, benchmarks, polish | Complete |

## Files to Create

```
pkg/layout/
├── rect.go
├── rect_test.go
├── point.go
├── edges.go
├── value.go
├── value_test.go
├── style.go
├── node.go
├── node_test.go
├── calculate.go
├── calculate_test.go
├── flex.go
├── integration_test.go
└── benchmark_test.go
```

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/tui/rect.go` | Convert to type alias for `layout.Rect`, remove duplicate methods |
| `pkg/tui/*.go` | Update imports to use `layout.Rect` where needed |
