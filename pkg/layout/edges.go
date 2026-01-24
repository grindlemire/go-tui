package layout

// Edges represents values for four sides of a box.
type Edges struct {
	Top, Right, Bottom, Left int
}

// EdgeAll creates Edges with the same value on all sides.
func EdgeAll(n int) Edges {
	return Edges{Top: n, Right: n, Bottom: n, Left: n}
}

// EdgeSymmetric creates Edges with vertical (top/bottom) and horizontal (left/right) values.
func EdgeSymmetric(v, h int) Edges {
	return Edges{Top: v, Right: h, Bottom: v, Left: h}
}

// EdgeTRBL creates Edges following CSS order: Top, Right, Bottom, Left.
func EdgeTRBL(t, r, b, l int) Edges {
	return Edges{Top: t, Right: r, Bottom: b, Left: l}
}

// Horizontal returns the sum of Left and Right.
func (e Edges) Horizontal() int {
	return e.Left + e.Right
}

// Vertical returns the sum of Top and Bottom.
func (e Edges) Vertical() int {
	return e.Top + e.Bottom
}

// IsZero returns true if all edge values are zero.
func (e Edges) IsZero() bool {
	return e.Top == 0 && e.Right == 0 && e.Bottom == 0 && e.Left == 0
}
