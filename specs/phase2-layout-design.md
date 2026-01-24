# Phase 2: Layout Engine Specification

**Status:** Planned
**Version:** 1.1
**Last Updated:** 2025-01-24

---

## 1. Overview

### Purpose

The layout engine computes absolute positions and sizes for a tree of nodes based on flexbox-inspired style properties. It is a pure Go implementation with no external dependencies, designed as a standalone package that can be used independently of the TUI rendering system.

The engine follows a clear separation of concerns: users build a tree of nodes with style properties, the layout engine computes rectangles, and renderers read those rectangles. The layout pass performs minimal allocations—it reads styles and writes computed layouts to existing nodes, with only a small per-node scratch allocation for flex item calculations.

### Goals

- **Minimal-allocation layout pass**: Operate on existing nodes with only scratch allocations for flex calculations
- **Incremental layout**: Track dirty nodes and only recalculate affected subtrees
- **Standalone package**: `pkg/layout` imports nothing from `pkg/tui`—enables independent testing and reuse
- **Predictable behavior**: Prioritize consistency and debuggability over CSS spec compliance
- **Terminal-optimized flexbox**: Core flexbox properties adapted for integer-based terminal coordinates

### Non-Goals

- Full CSS Flexbox specification compliance
- `flex-wrap` (use explicit containers or scroll widgets instead)
- `order` property (reorder the tree directly)
- `flex-basis: auto` with intrinsic sizing (require explicit dimensions)
- Float/percentage sub-pixel precision (all values resolve to integers)
- Absolute/fixed positioning (all layout is relative/flex)

---

## 2. Architecture

### Directory Structure

```
pkg/layout/
├── rect.go        # Rect type (canonical geometry primitive)
├── point.go       # Point type for coordinates
├── edges.go       # Edges type for padding/margin
├── value.go       # Value type with unit system
├── style.go       # Layout style properties
├── node.go        # Node tree with dirty tracking
├── calculate.go   # Layout algorithm entry point
├── flex.go        # Flexbox calculation logic
└── layout_test.go # Comprehensive test suite
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `rect.go` | Canonical geometry primitive—tui imports this |
| `point.go` | 2D coordinate point |
| `edges.go` | Four-sided values for padding/margin |
| `value.go` | Dimension values with Fixed/Percent/Auto units |
| `style.go` | Flexbox properties (direction, justify, align, etc.) |
| `node.go` | Tree node with style, children, computed layout, dirty flag |
| `calculate.go` | Public API: `Calculate(node, width, height)` |
| `flex.go` | Internal flexbox algorithm implementation |

### Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│  User Code                                                  │
│  root := layout.NewNode(style)                              │
│  root.AddChild(child1, child2)                              │
│  child1.SetStyle(newStyle)  // marks dirty                  │
└─────────────────────────────┬───────────────────────────────┘
                              │ Build tree, set styles
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Node Tree (user-owned)                                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Node { Style, Children, Layout (computed), dirty }   │   │
│  │   ├── Node { ... }                                   │   │
│  │   └── Node { ... }                                   │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────┬───────────────────────────────┘
                              │ layout.Calculate(root, w, h)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Layout Engine (zero-alloc pass)                            │
│  1. Check dirty flags                                       │
│  2. Skip clean subtrees                                     │
│  3. Compute Layout rect for dirty nodes                     │
│  4. Clear dirty flags                                       │
└─────────────────────────────┬───────────────────────────────┘
                              │ Layout.Rect populated
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Renderer reads node.Layout.Rect                            │
│  Draws content at computed position                         │
└─────────────────────────────────────────────────────────────┘
```

### Integration with pkg/tui

The layout package owns all geometry primitives. The tui package imports and uses them directly—no duplication:

```go
// In pkg/tui
import "github.com/grindlemire/go-tui/pkg/layout"

// Use layout.Rect directly throughout tui
func (b *Buffer) Fill(rect layout.Rect, r rune, style Style)
func (b *Buffer) SetString(x, y int, s string, style Style)
```

The existing `tui.Rect` will be replaced by importing `layout.Rect`. This is a breaking change to Phase 1 code, but maintains a single source of truth for geometry types.

Widgets build a `layout.Node` tree, call `layout.Calculate()`, then use the computed `Layout.Rect` for rendering.

---

## 3. Core Entities

### 3.1 Rect

The canonical geometry primitive for the entire codebase. Defined in `pkg/layout`; `pkg/tui` imports it.

```go
// pkg/layout/rect.go

// Rect represents a rectangle with integer coordinates.
// X and Y are the top-left corner; Width and Height are dimensions.
type Rect struct {
    X, Y          int
    Width, Height int
}

// Constructors
func NewRect(x, y, width, height int) Rect

// Edge accessors
func (r Rect) Right() int              // X + Width
func (r Rect) Bottom() int             // Y + Height

// Queries
func (r Rect) IsEmpty() bool           // Width <= 0 || Height <= 0
func (r Rect) Area() int
func (r Rect) Contains(x, y int) bool
func (r Rect) ContainsRect(other Rect) bool

// Transformations
func (r Rect) Inset(edges Edges) Rect   // Shrink inward (positive values shrink)
func (r Rect) Outset(edges Edges) Rect  // Expand outward (positive values expand)
func (r Rect) Translate(dx, dy int) Rect
func (r Rect) Intersect(other Rect) Rect
func (r Rect) Union(other Rect) Rect
func (r Rect) Clamp(x, y int) (int, int)
```

**Design Decision**: `pkg/layout` owns `Rect`; `pkg/tui` imports it. No duplication. This establishes layout as the foundational geometry package. The existing `tui.Rect` will be removed and replaced with imports from layout.

### 3.1.1 Point

Simple 2D coordinate type for completeness.

```go
// pkg/layout/point.go

// Point represents an (X, Y) coordinate.
type Point struct {
    X, Y int
}

func (p Point) Add(other Point) Point
func (p Point) Sub(other Point) Point
func (p Point) In(r Rect) bool
```

### 3.2 Edges

Four-sided values for padding and margin.

```go
// pkg/layout/edges.go

// Edges represents values for four sides of a box.
type Edges struct {
    Top, Right, Bottom, Left int
}

// Convenience constructors
func EdgeAll(n int) Edges              // Same value all sides
func EdgeSymmetric(v, h int) Edges     // Vertical, Horizontal
func EdgeTRBL(t, r, b, l int) Edges    // CSS order: Top Right Bottom Left

// Methods
func (e Edges) Horizontal() int        // Left + Right
func (e Edges) Vertical() int          // Top + Bottom
func (e Edges) IsZero() bool
```

### 3.3 Value (Dimension System)

Dimensions can be fixed pixels, percentages, or auto-computed.

```go
// pkg/layout/value.go

// Unit specifies how a Value is interpreted.
type Unit uint8

const (
    UnitAuto    Unit = iota  // Size determined by content/flex
    UnitFixed                // Absolute terminal cells
    UnitPercent              // Percentage of parent's available space
)

// Value represents a dimension that can be fixed, percentage, or auto.
type Value struct {
    Amount float64
    Unit   Unit
}

// Constructors (prefer these over struct literals)
func Auto() Value
func Fixed(n int) Value
func Percent(p float64) Value  // 0-100 scale (50.0 = 50%)

// Resolve computes the actual integer value given available space.
// For UnitAuto, returns the fallback value.
func (v Value) Resolve(available, fallback int) int

// IsAuto returns true if this value should be computed from content/flex.
func (v Value) IsAuto() bool
```

**Design Decision**: Use `float64` for Amount to support fractional percentages. Resolution always rounds to integers (terminal cells are discrete).

### 3.4 Style

Layout properties for a node. Separate from `tui.Style` which handles visual appearance (colors, bold, etc.).

```go
// pkg/layout/style.go

// Direction specifies the main axis for laying out children.
type Direction uint8

const (
    Row    Direction = iota  // Children laid out left-to-right
    Column                   // Children laid out top-to-bottom
)

// Justify specifies how children are distributed along the main axis.
type Justify uint8

const (
    JustifyStart        Justify = iota  // Pack at start
    JustifyEnd                          // Pack at end
    JustifyCenter                       // Center children
    JustifySpaceBetween                 // Even space between, none at edges
    JustifySpaceAround                  // Even space around each child
    JustifySpaceEvenly                  // Equal space between and at edges
)

// Align specifies how children are positioned on the cross axis.
type Align uint8

const (
    AlignStart   Align = iota  // Align to start of cross axis
    AlignEnd                   // Align to end of cross axis
    AlignCenter                // Center on cross axis
    AlignStretch               // Stretch to fill cross axis
)

// Style contains all layout properties for a node.
type Style struct {
    // Sizing
    Width     Value
    Height    Value
    MinWidth  Value
    MinHeight Value
    MaxWidth  Value
    MaxHeight Value

    // Flex container properties
    Direction      Direction
    JustifyContent Justify
    AlignItems     Align
    Gap            int  // Space between children (main axis only)

    // Flex item properties
    FlexGrow   float64  // How much to grow relative to siblings
    FlexShrink float64  // How much to shrink relative to siblings (default 1)
    AlignSelf  *Align   // Override parent's AlignItems (nil = inherit)

    // Spacing
    Padding Edges
    Margin  Edges
}

// DefaultStyle returns a Style with sensible defaults.
func DefaultStyle() Style {
    return Style{
        Width:      Auto(),
        Height:     Auto(),
        MinWidth:   Fixed(0),
        MinHeight:  Fixed(0),
        MaxWidth:   Auto(),  // No maximum
        MaxHeight:  Auto(),  // No maximum
        Direction:  Row,
        AlignItems: AlignStretch,
        FlexShrink: 1.0,
    }
}
```

**Design Decisions**:

1. **AlignSelf is a pointer**: `nil` means "inherit from parent's AlignItems". A concrete value overrides.
2. **FlexShrink defaults to 1.0**: Matches CSS behavior—items shrink by default when container is too small.
3. **Gap is main-axis only**: Consistent with CSS `gap` in flex containers. Cross-axis gap would require wrap support.
4. **No flex-basis**: Use `Width`/`Height` directly. Auto-sizing from content is not supported (explicit dimensions required).

### 3.5 Node

The tree structure with dirty tracking for incremental layout.

```go
// pkg/layout/node.go

// Layout holds the computed position and size after layout calculation.
type Layout struct {
    // Rect is the border box—the space allocated by the parent after
    // applying this node's margin. Use for hit testing and bounds.
    Rect Rect

    // ContentRect is Rect minus padding—the area where children are placed.
    // Use for rendering content and positioning children.
    ContentRect Rect
}

// Node represents an element in the layout tree.
type Node struct {
    // Configuration (user-set)
    Style    Style
    Children []*Node

    // Computed (set by layout engine)
    Layout Layout

    // Internal state
    dirty  bool   // Needs recalculation
    parent *Node  // Back-pointer for dirty propagation
}

// NewNode creates a new node with the given style.
func NewNode(style Style) *Node

// AddChild appends a child and marks this node dirty.
func (n *Node) AddChild(children ...*Node)

// RemoveChild removes a child by pointer and marks dirty.
func (n *Node) RemoveChild(child *Node) bool

// SetStyle updates the style and marks the node dirty.
func (n *Node) SetStyle(style Style)

// MarkDirty marks this node and all ancestors as needing recalculation.
func (n *Node) MarkDirty()

// IsDirty returns whether this node needs recalculation.
func (n *Node) IsDirty() bool
```

**Dirty Tracking Strategy**:

When a node is marked dirty:
1. Set its `dirty` flag to `true`
2. Walk up the parent chain, marking each ancestor dirty
3. Stop if an ancestor is already dirty (optimization)

During layout calculation:
1. If a node is clean and has no dirty descendants, skip it entirely
2. If dirty, recalculate and clear the flag
3. After calculation, all reachable nodes are clean

```
Before MarkDirty(C):          After MarkDirty(C):
       A (clean)                     A (DIRTY)
      / \                           / \
     B   D (clean)                 B   D (clean)
    /                             / (DIRTY)
   C (target)                    C (DIRTY)
```

---

## 4. Layout Algorithm

### 4.1 Public API

```go
// pkg/layout/calculate.go

// Calculate performs layout calculation on the tree rooted at node.
// The node and all descendants will have their Layout field populated.
// Only dirty nodes are recalculated (incremental layout).
//
// availableWidth and availableHeight specify the root constraint
// (typically the terminal size).
func Calculate(node *Node, availableWidth, availableHeight int)
```

### 4.2 Algorithm Overview

The algorithm uses a two-pass approach for each node:

**Pass 1 - Measure**: Determine how much space each child needs
**Pass 2 - Arrange**: Position children within the available space

**Margin Handling**: Margins are applied by the *parent* when positioning children. A child receives its border box (post-margin space) as `available`. The child's `computeBorderBox` only applies width/height/min/max constraints—it does NOT re-apply margin.

```go
// Pseudocode for the core algorithm
func calculateNode(node *Node, available Rect) {
    // Dirty propagates up, so a clean node guarantees a clean subtree
    if !node.dirty {
        return
    }

    // 1. Compute this node's border box within available space
    //    (applies width/height/min/max constraints, NOT margin—margin was
    //    already applied by parent when computing 'available')
    borderBox := computeBorderBox(node.Style, available)

    // 2. Compute content rect (border box minus padding)
    contentRect := borderBox.Inset(node.Style.Padding)

    // 3. Layout children within content rect
    layoutChildren(node, contentRect)

    // 4. Store computed layout
    node.Layout = Layout{
        Rect:        borderBox,
        ContentRect: contentRect,
    }

    // 5. Clear dirty flag
    node.dirty = false
}
```

### 4.3 Flexbox Calculation

The flexbox algorithm handles the main complexity:

```go
// pkg/layout/flex.go

func layoutChildren(node *Node, contentRect Rect) {
    if len(node.Children) == 0 {
        return
    }

    style := node.Style
    isRow := style.Direction == Row

    // Determine main/cross axis dimensions
    mainSize := contentRect.Width
    crossSize := contentRect.Height
    if !isRow {
        mainSize, crossSize = crossSize, mainSize
    }

    // Phase 1: Compute base sizes and flex factors
    items := make([]flexItem, len(node.Children))
    totalFixed := 0
    totalGrow := 0.0
    totalShrink := 0.0

    for i, child := range node.Children {
        item := &items[i]
        item.node = child

        // Resolve fixed size (or 0 if auto)
        if isRow {
            item.baseSize = child.Style.Width.Resolve(mainSize, 0)
        } else {
            item.baseSize = child.Style.Height.Resolve(mainSize, 0)
        }

        item.grow = child.Style.FlexGrow
        item.shrink = child.Style.FlexShrink

        totalFixed += item.baseSize
        totalGrow += item.grow
        totalShrink += item.shrink
    }

    // Account for gaps
    totalGap := style.Gap * (len(node.Children) - 1)
    freeSpace := mainSize - totalFixed - totalGap

    // Phase 2: Distribute free space
    if freeSpace > 0 && totalGrow > 0 {
        // Grow items
        for i := range items {
            if items[i].grow > 0 {
                extra := int(float64(freeSpace) * items[i].grow / totalGrow)
                items[i].mainSize = items[i].baseSize + extra
            } else {
                items[i].mainSize = items[i].baseSize
            }
        }
    } else if freeSpace < 0 && totalShrink > 0 {
        // Shrink items
        deficit := -freeSpace
        for i := range items {
            if items[i].shrink > 0 {
                reduction := int(float64(deficit) * items[i].shrink / totalShrink)
                items[i].mainSize = max(0, items[i].baseSize - reduction)
            } else {
                items[i].mainSize = items[i].baseSize
            }
        }
    } else {
        for i := range items {
            items[i].mainSize = items[i].baseSize
        }
        freeSpace = max(0, freeSpace)  // For justify calculations
    }

    // Phase 3: Apply min/max constraints
    for i, child := range node.Children {
        minMain := resolveMin(child.Style, isRow, mainSize)
        maxMain := resolveMax(child.Style, isRow, mainSize)
        items[i].mainSize = clamp(items[i].mainSize, minMain, maxMain)
    }

    // Phase 4: Position children (justify)
    offset := calculateJustifyOffset(style.JustifyContent, freeSpace, len(items))
    spacing := calculateJustifySpacing(style.JustifyContent, freeSpace, len(items))

    for i := range items {
        items[i].mainPos = offset
        offset += items[i].mainSize + style.Gap + spacing
    }

    // Phase 5: Cross-axis sizing and alignment
    for i, child := range node.Children {
        align := style.AlignItems
        if child.Style.AlignSelf != nil {
            align = *child.Style.AlignSelf
        }

        if align == AlignStretch {
            items[i].crossSize = crossSize
            items[i].crossPos = 0
        } else {
            // Resolve explicit cross size or use crossSize as fallback
            if isRow {
                items[i].crossSize = child.Style.Height.Resolve(crossSize, crossSize)
            } else {
                items[i].crossSize = child.Style.Width.Resolve(crossSize, crossSize)
            }
            items[i].crossPos = calculateAlignOffset(align, crossSize, items[i].crossSize)
        }
    }

    // Phase 6: Convert to rects and recurse
    for i, child := range node.Children {
        // Compute the slot allocated to this child (before margin)
        var slot Rect
        if isRow {
            slot = Rect{
                X:      contentRect.X + items[i].mainPos,
                Y:      contentRect.Y + items[i].crossPos,
                Width:  items[i].mainSize,
                Height: items[i].crossSize,
            }
        } else {
            slot = Rect{
                X:      contentRect.X + items[i].crossPos,
                Y:      contentRect.Y + items[i].mainPos,
                Width:  items[i].crossSize,
                Height: items[i].mainSize,
            }
        }

        // Apply child's margin: shrink the slot to get the child's border box.
        // The child receives this as 'available' and does NOT re-apply margin.
        childBorderBox := slot.Inset(child.Style.Margin)

        // Recurse—child computes its layout within this border box
        calculateNode(child, childBorderBox)
    }
}

// flexItem holds intermediate calculation state for a child.
// This is stack-allocated per layout call, not stored on nodes.
type flexItem struct {
    node      *Node
    baseSize  int
    mainSize  int
    crossSize int
    mainPos   int
    crossPos  int
    grow      float64
    shrink    float64
}
```

### 4.4 Justify Content Distribution

```
JustifyStart:        [A][B][C]...............
JustifyEnd:          ...............[A][B][C]
JustifyCenter:       .......[A][B][C].......
JustifySpaceBetween: [A]......[B]......[C]
JustifySpaceAround:  ..[A]....[B]....[C]..
JustifySpaceEvenly:  ...[A]...[B]...[C]...
```

### 4.5 Align Items/Self Positioning

```
Cross axis (for Row direction = vertical):

AlignStart:   [Content]
              ─────────────

AlignCenter:  ─────────────
              [Content]
              ─────────────

AlignEnd:     ─────────────
                   [Content]

AlignStretch: [    Content fills entire cross axis    ]
```

---

## 5. Incremental Layout

### 5.1 Dirty Propagation

```go
// MarkDirty marks this node and propagates up the tree.
func (n *Node) MarkDirty() {
    for node := n; node != nil && !node.dirty; node = node.parent {
        node.dirty = true
    }
}
```

### 5.2 Clean Subtree Skipping

During calculation, we can skip entire subtrees with a simple check:

```go
func calculateNode(node *Node, available Rect) {
    // Dirty propagates up, so a clean node guarantees a clean subtree.
    // No need to check descendants—if any were dirty, this node would be too.
    if !node.dirty {
        return
    }

    // ... perform layout ...

    node.dirty = false
}
```

**Important**: When `MarkDirty()` is called on a node, it propagates up to the root. Therefore, a clean node (`!dirty`) guarantees all descendants are also clean. No `hasDirtyDescendant()` check is needed.

### 5.3 What Triggers Dirty

| Action | Effect |
|--------|--------|
| `SetStyle(s)` | Node marked dirty |
| `AddChild(c)` | Node marked dirty |
| `RemoveChild(c)` | Node marked dirty |
| `MarkDirty()` | Explicit dirty marking |

Importantly, reading `node.Layout` does NOT mark dirty. The layout is stable until explicitly changed.

---

## 6. Edge Cases and Constraints

### 6.1 Zero/Negative Dimensions

- Widths and heights are clamped to `>= 0`
- A node with zero size still exists in the tree but renders nothing
- Children of a zero-size parent get zero-size rects

### 6.2 Overflow

When children exceed parent size:
1. `FlexShrink` distributes the deficit proportionally
2. If shrink totals to 0, children are clipped (no negative sizes)
3. Clipping is the renderer's responsibility—layout produces oversized rects

### 6.3 Percentage Resolution

- Percentages resolve against the parent's **content** area (after padding)
- Nested percentages are relative to their immediate parent
- `Percent(100)` in a padded container fills the content area, not the outer rect

### 6.4 Min/Max Constraints

Applied after flex distribution:
```go
finalSize = clamp(flexComputedSize, minSize, maxSize)
```

If min > max, min wins (matches CSS behavior).

### 6.5 Empty Nodes

- A node with no children has its size determined by `Width`/`Height` style
- If both are `Auto`, the node collapses to 0x0 (no intrinsic sizing)

---

## 7. Performance Considerations

### 7.1 Minimal-Allocation Layout Pass

The layout pass minimizes allocations:
- **One allocation per node with children**: `make([]flexItem, len(children))` in `layoutChildren`
- Nodes and their children slices are pre-allocated by the user
- No allocations for leaf nodes

For strict zero-allocation requirements (not needed for v1), a stack buffer pattern can be used:
```go
var buf [16]flexItem  // Stack-allocated for small child counts
items := buf[:len(node.Children)]
if len(node.Children) > 16 {
    items = make([]flexItem, len(node.Children))  // Heap fallback
}
```

This optimization is deferred—the current approach prioritizes clarity over micro-optimization.

### 7.2 Incremental Efficiency

For a tree of N nodes with K dirty:
- Best case: O(K) when dirty nodes are leaves
- Worst case: O(N) when root is dirty
- Typical case: O(K + path-to-root) for localized changes

### 7.3 Cache Friendliness

- Node struct keeps hot fields (Style, Layout) close together
- Children slice is a single pointer dereference
- Dirty flag checked before any computation

---

## 8. Testing Strategy

### 8.1 Unit Test Categories

| Category | Description |
|----------|-------------|
| Value resolution | Fixed, Percent, Auto with various available spaces |
| Single node | Sizing with min/max, padding, margin |
| Two children | Row/Column, grow/shrink distribution |
| Justify | All 6 justify modes with various child counts |
| Align | All 4 align modes plus AlignSelf override |
| Nested | Multi-level trees with mixed directions |
| Dirty tracking | Incremental recalculation verification |
| Edge cases | Zero sizes, overflow, constraint conflicts |

### 8.2 Golden Tests

Complex layouts with known expected results:
- Dashboard layout (header, sidebar, main, footer)
- Form with labels and inputs
- Nested flex containers

### 8.3 Fuzz Testing

Random tree generation to catch:
- Integer overflow
- Infinite loops
- Negative dimensions

---

## 9. Complexity Assessment

| Factor | Assessment |
|--------|------------|
| New files | 6-7 files in `pkg/layout/` |
| Core algorithm complexity | Moderate—flexbox is well-understood |
| Integration points | Minimal—standalone package |
| Testing scope | Significant—many edge cases |

**Assessed Size:** Medium
**Recommended Phases:** 4
**Rationale:** The layout engine has moderate complexity with well-defined boundaries. Four phases allow for incremental delivery: (1) foundation types, (2) basic flex algorithm, (3) advanced flex features + dirty tracking, (4) integration and comprehensive testing. Each phase is independently testable.

---

## 10. Success Criteria

1. `pkg/layout` compiles with zero imports from `pkg/tui`
2. `pkg/tui` successfully imports and uses `layout.Rect` (replacing duplicate type)
3. `Calculate()` produces correct rects for all flexbox properties listed in Style
4. Layout pass allocates only for `flexItem` slices (verified via benchmarks)
5. Dirty tracking correctly skips clean subtrees (verified via calculation counts)
6. All justify/align combinations produce expected positioning
7. Min/max constraints are respected after flex distribution
8. Nested flex containers (row-in-column, column-in-row) work correctly
9. Edge cases (zero size, overflow, empty nodes) are handled gracefully

---

## 11. Open Questions

1. **Should `Layout` include margin-excluded rect?** → Included as `ContentRect` for renderer convenience
2. **How to handle fractional flex distribution?** → Round to integers; accept small rounding errors
3. **Should AlignSelf support all Align values?** → Yes, same enum reused
4. **Gap on cross axis?** → Deferred; requires wrap support. Use explicit margins instead.
5. **Migration of tui.Rect?** → Phase 1 of implementation will include updating `pkg/tui` to import `layout.Rect`. This is a breaking change but ensures single source of truth.
