# Layout Reference

## Overview

go-tui uses a CSS flexbox layout engine to position elements on screen. Every `<div>` acts as a flex container, arranging its children along a main axis (horizontal by default) with control over alignment, spacing, and sizing. The layout engine runs automatically when state changes, but you can also trigger it manually for testing or advanced use cases.

All layout types live in the root `tui` package (re-exported from `internal/layout`). You interact with them through element options in `.gsx` files or programmatically via `tui.New()`.

```gsx
// Flexbox in action: vertical layout with centered children
<div class="flex-col items-center gap-1 p-1 border-rounded">
    <span class="font-bold">Title</span>
    <span>Content goes here</span>
</div>
```

```go
// Programmatic equivalent
root := tui.New(
    tui.WithDirection(tui.Column),
    tui.WithAlign(tui.AlignCenter),
    tui.WithGap(1),
    tui.WithPadding(1),
    tui.WithBorder(tui.BorderRounded),
)
title := tui.New(
    tui.WithText("Title"),
    tui.WithTextStyle(tui.NewStyle().Bold()),
)
body := tui.New(tui.WithText("Content goes here"))
root.AddChild(title, body)
```

## Direction

`Direction` controls which axis children are laid out along. It is a `uint8` enum.

| Constant | Value | Description |
|----------|-------|-------------|
| `Row` | `0` | Children laid out left-to-right (default) |
| `Column` | `1` | Children laid out top-to-bottom |

```go
tui.New(tui.WithDirection(tui.Column)) // vertical stack
tui.New(tui.WithDirection(tui.Row))    // horizontal row (default)
```

In `.gsx`, use the Tailwind classes `flex-col` for column and `flex` or `flex-row` for row.

## Justify

`Justify` controls how children are distributed along the main axis (the direction axis). It is a `uint8` enum.

| Constant | Value | Description |
|----------|-------|-------------|
| `JustifyStart` | `0` | Pack children at the start (default) |
| `JustifyEnd` | `1` | Pack children at the end |
| `JustifyCenter` | `2` | Center children |
| `JustifySpaceBetween` | `3` | Equal space between children, no space at edges |
| `JustifySpaceAround` | `4` | Equal space around each child (half-space at edges) |
| `JustifySpaceEvenly` | `5` | Equal space between children and at edges |

```go
tui.New(tui.WithJustify(tui.JustifyCenter))       // center children
tui.New(tui.WithJustify(tui.JustifySpaceBetween))  // spread children apart
```

Visual comparison for three children `[A] [B] [C]` in a row:

```
JustifyStart:        [A][B][C]
JustifyEnd:                      [A][B][C]
JustifyCenter:            [A][B][C]
JustifySpaceBetween: [A]       [B]       [C]
JustifySpaceAround:   [A]    [B]    [C]
JustifySpaceEvenly:    [A]    [B]    [C]
```

In `.gsx`, use `justify-start`, `justify-center`, `justify-end`, `justify-between`, `justify-around`, or `justify-evenly`.

## Align

`Align` controls how children are positioned on the cross axis (perpendicular to the direction). It is a `uint8` enum.

| Constant | Value | Description |
|----------|-------|-------------|
| `AlignStart` | `0` | Align to the start of the cross axis |
| `AlignEnd` | `1` | Align to the end of the cross axis |
| `AlignCenter` | `2` | Center on the cross axis |
| `AlignStretch` | `3` | Stretch to fill the cross axis (default) |

```go
tui.New(tui.WithAlign(tui.AlignCenter))  // center children on cross axis
tui.New(tui.WithAlign(tui.AlignStretch)) // stretch to fill (default)
```

In `.gsx`, use `items-start`, `items-end`, `items-center`, or `items-stretch`.

Individual children can override the parent's alignment with `AlignSelf`:

```go
tui.New(tui.WithAlignSelf(tui.AlignCenter)) // this child centers itself
```

In `.gsx`, use `self-start`, `self-center`, `self-end`, or `self-stretch`.

## FlexWrap

`FlexWrap` controls whether children wrap onto new lines when they overflow the main axis. It is a `uint8` enum.

| Constant | Value | Description |
|----------|-------|-------------|
| `WrapNone` | `0` | Children stay on one line (default) |
| `Wrap` | `1` | Children wrap to new lines when they overflow |
| `WrapReverse` | `2` | Children wrap in reverse order |

```go
tui.New(tui.WithFlexWrap(tui.Wrap)) // enable wrapping
```

In `.gsx`, use `flex-wrap`, `flex-wrap-reverse`, or `flex-nowrap`.

With wrapping enabled, each line handles grow, shrink, and justify separately.

## AlignContent

`AlignContent` controls how wrapped lines are spaced along the cross axis. It requires `FlexWrap` to be enabled and at least two lines to have any effect. It is a `uint8` enum.

| Constant | Value | Description |
|----------|-------|-------------|
| `ContentStart` | `0` | Pack lines at the start (default) |
| `ContentEnd` | `1` | Pack lines at the end |
| `ContentCenter` | `2` | Center lines |
| `ContentStretch` | `3` | Stretch lines to fill the cross axis |
| `ContentSpaceBetween` | `4` | First line at start, last at end, even spacing between |
| `ContentSpaceAround` | `5` | Equal spacing around each line |

```go
tui.New(
    tui.WithFlexWrap(tui.Wrap),
    tui.WithAlignContent(tui.ContentCenter),
)
```

In `.gsx`, use `content-start`, `content-end`, `content-center`, `content-stretch`, `content-between`, or `content-around`.

## Value

`Value` represents a dimension that can be fixed, percentage-based, or automatic. The layout engine resolves values against available space during calculation.

### Constructors

```go
func Fixed(n int) Value
```

Creates a value representing an absolute number of terminal cells (columns for width, rows for height).

```go
tui.Fixed(40) // exactly 40 characters wide
```

---

```go
func Percent(p float64) Value
```

Creates a value representing a percentage of the parent's available space. Uses a 0-100 scale.

```go
tui.Percent(50)  // half of available space
tui.Percent(100) // full available space
```

---

```go
func Auto() Value
```

Creates a value that sizes to content. The layout engine computes the actual size from the element's intrinsic dimensions or flex properties. This is the default for both width and height.

```go
tui.Auto() // size determined by content
```

### Value Fields

```go
type Value struct {
    Amount float64
    Unit   Unit
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Amount` | `float64` | The numeric value (cell count for Fixed, 0-100 for Percent, unused for Auto) |
| `Unit` | `Unit` | How the value is interpreted |

### Unit Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `UnitAuto` | `0` | Size determined by content or flex |
| `UnitFixed` | `1` | Absolute terminal cells |
| `UnitPercent` | `2` | Percentage of parent's available space |

### Value Methods

#### Resolve

```go
func (v Value) Resolve(available, fallback int) int
```

Computes the actual integer value given the available space. For `UnitFixed`, returns the amount directly. For `UnitPercent`, computes `available * amount / 100`. For `UnitAuto`, returns the fallback value.

```go
v := tui.Percent(50)
cells := v.Resolve(80, 0) // returns 40 (50% of 80)

v = tui.Fixed(30)
cells = v.Resolve(80, 0) // returns 30

v = tui.Auto()
cells = v.Resolve(80, 12) // returns 12 (the fallback)
```

#### IsAuto

```go
func (v Value) IsAuto() bool
```

Returns `true` if the value has unit `UnitAuto`.

## LayoutStyle

`LayoutStyle` (aliased from `internal/layout.Style`) holds all layout properties for a node. You rarely construct this directly; instead, use element options like `WithWidth`, `WithDirection`, `WithPadding`, etc. However, it's useful for reading computed styles or building custom `Layoutable` implementations.

### Fields

```go
type LayoutStyle struct {
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
    Gap            int          // Space between children (main axis only)
    FlexWrap       FlexWrap     // Whether children wrap to new lines
    AlignContent   AlignContent // How wrapped lines are distributed on the cross axis

    // Flex item properties
    FlexGrow   float64      // How much to grow relative to siblings
    FlexShrink float64      // How much to shrink relative to siblings (default 1)
    AlignSelf  *Align       // Override parent's AlignItems (nil = inherit)

    // Spacing
    Padding Edges
    Margin  Edges
}
```

### DefaultLayoutStyle

```go
func DefaultLayoutStyle() LayoutStyle
```

Returns a `LayoutStyle` with the following defaults:

| Property | Default |
|----------|---------|
| `Width` | `Auto()` |
| `Height` | `Auto()` |
| `MinWidth` | `Auto()` (intrinsic size, matching CSS flexbox `min-width: auto`) |
| `MinHeight` | `Auto()` (intrinsic size) |
| `MaxWidth` | `Auto()` (no maximum) |
| `MaxHeight` | `Auto()` (no maximum) |
| `Direction` | `Row` |
| `JustifyContent` | `JustifyStart` |
| `AlignItems` | `AlignStretch` |
| `FlexWrap` | `WrapNone` |
| `AlignContent` | `ContentStart` |
| `FlexGrow` | `0` |
| `FlexShrink` | `1.0` |
| `AlignSelf` | `nil` (inherit from parent) |
| `Gap` | `0` |
| `Padding` | all zeros |
| `Margin` | all zeros |

## Rect

`Rect` represents a rectangle with integer coordinates. After layout runs, every element has a `Rect` (border box) and a `ContentRect` (inner area where children are placed).

### Fields

```go
type Rect struct {
    X, Y          int
    Width, Height int
}
```

`X` and `Y` are the top-left corner. `Width` and `Height` are the dimensions.

### NewRect

```go
func NewRect(x, y, width, height int) Rect
```

Creates a new Rect.

```go
r := tui.NewRect(5, 10, 40, 20) // x=5, y=10, 40 wide, 20 tall
```

### Edge Queries

```go
func (r Rect) Right() int
func (r Rect) Bottom() int
```

`Right()` returns `X + Width` (exclusive right edge). `Bottom()` returns `Y + Height` (exclusive bottom edge).

### State Queries

```go
func (r Rect) IsEmpty() bool
func (r Rect) Area() int
```

`IsEmpty()` returns `true` if width or height is zero or negative. `Area()` returns width times height (0 if empty).

### Hit Testing

```go
func (r Rect) Contains(x, y int) bool
func (r Rect) ContainsRect(other Rect) bool
```

`Contains` checks if a point falls inside the rectangle. Points on the left and top edges are inside; points on the right and bottom edges are outside. `ContainsRect` checks if another rectangle is fully within this one. An empty `other` is always contained.

```go
r := tui.NewRect(0, 0, 80, 24)
r.Contains(0, 0)    // true (top-left is inside)
r.Contains(80, 0)   // false (right edge is outside)
r.Contains(79, 23)  // true
```

### Geometric Operations

```go
func (r Rect) Inset(edges Edges) Rect
func (r Rect) Outset(edges Edges) Rect
func (r Rect) Translate(dx, dy int) Rect
func (r Rect) Intersect(other Rect) Rect
func (r Rect) Union(other Rect) Rect
func (r Rect) Intersects(other Rect) bool
```

| Method | Description |
|--------|-------------|
| `Inset(edges)` | Shrink the rectangle inward by the given edges. Positive values shrink; negative expand |
| `Outset(edges)` | Expand the rectangle outward by the given edges. The inverse of Inset |
| `Translate(dx, dy)` | Move the rectangle by `(dx, dy)` without changing size |
| `Intersect(other)` | Return the overlapping region. Returns an empty Rect if they don't overlap |
| `Union(other)` | Return the smallest rectangle containing both. If either is empty, returns the other |
| `Intersects(other)` | Return `true` if the rectangles overlap. Touching edges don't count |

```go
r := tui.NewRect(0, 0, 80, 24)

// Shrink by 2 on all sides
inner := r.Inset(tui.EdgeAll(2)) // {X:2, Y:2, Width:76, Height:20}

// Move right by 10
shifted := r.Translate(10, 0) // {X:10, Y:0, Width:80, Height:24}
```

### Clamp

```go
func (r Rect) Clamp(x, y int) (int, int)
```

Constrains a point to be within the rectangle bounds. Returns the clamped coordinates.

```go
r := tui.NewRect(0, 0, 80, 24)
x, y := r.Clamp(100, 30) // returns 79, 23
```

### Convenience Functions

```go
func InsetRect(r Rect, top, right, bottom, left int) Rect
func InsetUniform(r Rect, n int) Rect
```

`InsetRect` insets a Rect by TRBL amounts (CSS order). `InsetUniform` insets by the same amount on all edges. Both wrap `Rect.Inset()`.

```go
r := tui.NewRect(0, 0, 80, 24)
inner := tui.InsetUniform(r, 1)             // {X:1, Y:1, Width:78, Height:22}
inner = tui.InsetRect(r, 2, 1, 2, 1)        // {X:1, Y:2, Width:78, Height:20}
```

## Edges

`Edges` represents spacing values on the four sides of a box. Used for padding and margin.

### Fields

```go
type Edges struct {
    Top, Right, Bottom, Left int
}
```

### Constructors

```go
func EdgeAll(n int) Edges
```

Same value on all four sides.

```go
tui.EdgeAll(2) // {Top:2, Right:2, Bottom:2, Left:2}
```

---

```go
func EdgeSymmetric(v, h int) Edges
```

Vertical (top and bottom) and horizontal (left and right) values.

```go
tui.EdgeSymmetric(1, 2) // {Top:1, Right:2, Bottom:1, Left:2}
```

---

```go
func EdgeTRBL(t, r, b, l int) Edges
```

Per-side values in CSS order: Top, Right, Bottom, Left.

```go
tui.EdgeTRBL(1, 2, 3, 4) // {Top:1, Right:2, Bottom:3, Left:4}
```

### Methods

```go
func (e Edges) Horizontal() int
func (e Edges) Vertical() int
func (e Edges) IsZero() bool
```

| Method | Returns |
|--------|---------|
| `Horizontal()` | `Left + Right` |
| `Vertical()` | `Top + Bottom` |
| `IsZero()` | `true` if all four values are zero |

```go
e := tui.EdgeTRBL(1, 2, 1, 2)
e.Horizontal() // 4
e.Vertical()   // 2
e.IsZero()     // false
```

## Size

`Size` represents dimensions without a position.

### Fields

```go
type Size struct {
    Width, Height int
}
```

## Point

`Point` represents an (X, Y) coordinate.

### Fields

```go
type Point struct {
    X, Y int
}
```

### Methods

```go
func (p Point) Add(other Point) Point
func (p Point) Sub(other Point) Point
func (p Point) In(r Rect) bool
```

| Method | Description |
|--------|-------------|
| `Add(other)` | Return a new Point offset by `other` |
| `Sub(other)` | Return a new Point with `other` subtracted |
| `In(r)` | Return `true` if the point is inside the rectangle |

```go
a := tui.Point{X: 10, Y: 5}
b := tui.Point{X: 3, Y: 2}
c := a.Add(b) // {X: 13, Y: 7}

r := tui.NewRect(0, 0, 80, 24)
a.In(r) // true
```

## LayoutResult

`LayoutResult` (aliased from `internal/layout.Layout`) holds the computed position and size after the layout engine runs. Every element stores one of these after `Calculate` completes.

### Fields

```go
type LayoutResult struct {
    Rect        Rect     // Border box: full area including border and padding
    ContentRect Rect     // Content area: inside border and padding, where children go
    AbsoluteX   float64  // Float position before rounding (for jitter-free animation)
    AbsoluteY   float64  // Float position before rounding (for jitter-free animation)
}
```

`Rect` is the border box, the space allocated by the parent after margin. Use it for hit testing and bounds checking. `ContentRect` is the area inside border and padding where children are placed.

`AbsoluteX` and `AbsoluteY` store the true floating-point position before integer rounding. The layout engine tracks float positions through the tree and rounds only once at the end, preventing cumulative rounding errors that cause visual jitter during animations.

## Layoutable

`Layoutable` is the interface that nodes must implement to participate in layout calculation. `Element` implements this interface. You can also implement it on custom types to plug into the layout engine directly.

```go
type Layoutable interface {
    LayoutStyle() Style
    LayoutChildren() []Layoutable
    SetLayout(Layout)
    GetLayout() Layout
    IsDirty() bool
    SetDirty(dirty bool)
    IntrinsicSize() (width, height int)
}
```

| Method | Description |
|--------|-------------|
| `LayoutStyle()` | Returns the layout properties for this node |
| `LayoutChildren()` | Returns the children to be laid out (should exclude hidden children) |
| `SetLayout(Layout)` | Called by the engine to store computed layout |
| `GetLayout()` | Returns the last computed layout |
| `IsDirty()` | Whether this node needs recalculation |
| `SetDirty(bool)` | Mark or clear the dirty flag |
| `IntrinsicSize()` | Natural content-based dimensions. Used as the base size for auto-sized elements |

## Calculate

```go
func Calculate(root Layoutable, availableWidth, availableHeight int)
```

Runs the flexbox layout algorithm on the tree rooted at `root`. After this call, every node in the tree has its `LayoutResult` populated with computed positions and sizes. Only dirty nodes are recalculated (incremental layout).

The `availableWidth` and `availableHeight` are the outer constraints, typically the terminal dimensions.

```go
root := tui.New(
    tui.WithDirection(tui.Column),
    tui.WithText("Hello"),
)
tui.Calculate(root, 80, 24)

r := root.Rect()           // computed border box
cr := root.ContentRect()   // computed content area
```

During normal application use, the framework calls `Calculate` automatically before each render. You only need to call it directly when working with elements outside the app lifecycle (testing, offline layout computation, etc.).

### How the Algorithm Works

The layout engine implements CSS flexbox in several phases:

1. **Compute base sizes** -- Resolve each child's main-axis size from its style (fixed, percent, or intrinsic) and record its flex grow/shrink factor.
2. **Break into lines** -- If `FlexWrap` is enabled, split children into lines based on main-axis overflow. Each line runs phases 3 and 4 independently.
3. **Distribute free space** -- If there's leftover space, grow children proportional to their `FlexGrow`. If children overflow, shrink them proportional to their `FlexShrink`.
4. **Apply min/max constraints** -- Clamp each child's computed size to its min/max bounds. If clamping changes a child's size, redistribute the difference among remaining flexible children.
5. **Position on main axis** -- Place children along the main axis using the parent's `JustifyContent` setting to compute offsets and spacing.
6. **Distribute lines on cross axis** -- When wrapping produces multiple lines, use `AlignContent` to distribute them (start, end, center, stretch, space-between, space-around).
7. **Cross-axis sizing and alignment** -- Compute each child's cross-axis size. Stretch children if `AlignItems` is `AlignStretch` (unless overridden by `AlignSelf`). Position using the alignment setting.
8. **Recurse** -- Convert float positions to integer rects and recurse into each child's subtree.

The engine uses Yoga-style float rounding (described in [LayoutResult](#layoutresult)) to avoid jitter during animations.
