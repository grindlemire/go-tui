package tui

import "github.com/grindlemire/go-tui/pkg/layout"

// Rect is a type alias for layout.Rect.
// This allows the tui package to use the canonical geometry primitive
// from the layout package while maintaining API compatibility.
type Rect = layout.Rect

// Edges is a type alias for layout.Edges.
type Edges = layout.Edges

// NewRect creates a new Rect with the given position and dimensions.
// This is a convenience wrapper for layout.NewRect.
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
