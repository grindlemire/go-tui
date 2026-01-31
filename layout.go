// layout.go re-exports layout types from internal/layout.
// Any changes to internal/layout types must be mirrored here.
package tui

import "github.com/grindlemire/go-tui/internal/layout"

// Direction specifies the main axis for laying out children.
type Direction = layout.Direction

const (
	Row    = layout.Row
	Column = layout.Column
)

// Justify specifies how children are distributed along the main axis.
type Justify = layout.Justify

const (
	JustifyStart        = layout.JustifyStart
	JustifyEnd          = layout.JustifyEnd
	JustifyCenter       = layout.JustifyCenter
	JustifySpaceBetween = layout.JustifySpaceBetween
	JustifySpaceAround  = layout.JustifySpaceAround
	JustifySpaceEvenly  = layout.JustifySpaceEvenly
)

// Align specifies how children are aligned along the cross axis.
type Align = layout.Align

const (
	AlignStart   = layout.AlignStart
	AlignEnd     = layout.AlignEnd
	AlignCenter  = layout.AlignCenter
	AlignStretch = layout.AlignStretch
)

// Value represents a dimension value (fixed, percent, or auto).
type Value = layout.Value

// Unit specifies how a Value is interpreted.
type Unit = layout.Unit

const (
	UnitAuto    = layout.UnitAuto
	UnitFixed   = layout.UnitFixed
	UnitPercent = layout.UnitPercent
)

// LayoutStyle holds the layout properties for a node.
type LayoutStyle = layout.Style

// Rect represents a rectangle with position and dimensions.
type Rect = layout.Rect

// Edges represents spacing on four sides (top, right, bottom, left).
type Edges = layout.Edges

// Size represents a width/height pair.
type Size = layout.Size

// Point represents an x/y coordinate.
type Point = layout.Point

// LayoutResult holds the computed layout for a node.
type LayoutResult = layout.Layout

// Layoutable is the interface that nodes must implement for layout calculation.
type Layoutable = layout.Layoutable

// Fixed creates a Value with a fixed character count.
func Fixed(n int) Value {
	return layout.Fixed(n)
}

// Percent creates a Value representing a percentage of available space.
func Percent(p float64) Value {
	return layout.Percent(p)
}

// Auto creates a Value that sizes to content.
func Auto() Value {
	return layout.Auto()
}

// DefaultLayoutStyle returns a Style with default values.
func DefaultLayoutStyle() LayoutStyle {
	return layout.DefaultStyle()
}

// NewRect creates a new Rect with the given position and dimensions.
func NewRect(x, y, width, height int) Rect {
	return layout.NewRect(x, y, width, height)
}

// EdgeAll creates Edges with the same value on all sides.
func EdgeAll(n int) Edges {
	return layout.EdgeAll(n)
}

// EdgeSymmetric creates Edges with vertical (top/bottom) and horizontal (left/right) values.
func EdgeSymmetric(v, h int) Edges {
	return layout.EdgeSymmetric(v, h)
}

// EdgeTRBL creates Edges following CSS order: Top, Right, Bottom, Left.
func EdgeTRBL(t, r, b, l int) Edges {
	return layout.EdgeTRBL(t, r, b, l)
}

// Calculate performs flexbox layout on the given tree.
func Calculate(root Layoutable, availableWidth, availableHeight int) {
	layout.Calculate(root, availableWidth, availableHeight)
}

// InsetRect returns a new Rect inset by the given amounts on each edge.
// The order follows CSS convention: top, right, bottom, left.
// This is a convenience function that wraps Rect.Inset(Edges).
func InsetRect(r Rect, top, right, bottom, left int) Rect {
	return r.Inset(layout.EdgeTRBL(top, right, bottom, left))
}

// InsetUniform returns a new Rect inset by n on all edges.
// This is a convenience function that wraps Rect.Inset(Edges).
func InsetUniform(r Rect, n int) Rect {
	return r.Inset(layout.EdgeAll(n))
}
