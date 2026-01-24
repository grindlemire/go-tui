package layout

// Rect represents a rectangle with integer coordinates.
// X and Y are the top-left corner; Width and Height are dimensions.
type Rect struct {
	X, Y          int
	Width, Height int
}

// NewRect creates a new Rect with the given position and dimensions.
func NewRect(x, y, width, height int) Rect {
	return Rect{X: x, Y: y, Width: width, Height: height}
}

// Right returns the x-coordinate of the right edge (exclusive).
func (r Rect) Right() int {
	return r.X + r.Width
}

// Bottom returns the y-coordinate of the bottom edge (exclusive).
func (r Rect) Bottom() int {
	return r.Y + r.Height
}

// IsEmpty returns true if the rectangle has zero or negative area.
func (r Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Area returns the area of the rectangle.
func (r Rect) Area() int {
	if r.Width <= 0 || r.Height <= 0 {
		return 0
	}
	return r.Width * r.Height
}

// Contains returns true if the point (x, y) is inside the rectangle.
// Points on the left and top edges are inside; points on the right and bottom edges are outside.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.Right() && y >= r.Y && y < r.Bottom()
}

// ContainsRect returns true if the other rectangle is fully contained within this rectangle.
func (r Rect) ContainsRect(other Rect) bool {
	if other.IsEmpty() {
		return true
	}
	if r.IsEmpty() {
		return false
	}
	return other.X >= r.X && other.Y >= r.Y &&
		other.Right() <= r.Right() && other.Bottom() <= r.Bottom()
}

// Inset returns a new Rect inset by the given Edges.
// Positive values shrink the rectangle; negative values expand it.
func (r Rect) Inset(edges Edges) Rect {
	return Rect{
		X:      r.X + edges.Left,
		Y:      r.Y + edges.Top,
		Width:  r.Width - edges.Left - edges.Right,
		Height: r.Height - edges.Top - edges.Bottom,
	}
}

// Outset returns a new Rect expanded outward by the given Edges.
// Positive values expand the rectangle; negative values shrink it.
func (r Rect) Outset(edges Edges) Rect {
	return Rect{
		X:      r.X - edges.Left,
		Y:      r.Y - edges.Top,
		Width:  r.Width + edges.Left + edges.Right,
		Height: r.Height + edges.Top + edges.Bottom,
	}
}

// Translate returns a new Rect moved by (dx, dy).
func (r Rect) Translate(dx, dy int) Rect {
	return Rect{X: r.X + dx, Y: r.Y + dy, Width: r.Width, Height: r.Height}
}

// Intersect returns the intersection of two rectangles.
// If the rectangles don't overlap, returns an empty Rect.
func (r Rect) Intersect(other Rect) Rect {
	x := max(r.X, other.X)
	y := max(r.Y, other.Y)
	right := min(r.Right(), other.Right())
	bottom := min(r.Bottom(), other.Bottom())

	width := right - x
	height := bottom - y

	if width <= 0 || height <= 0 {
		return Rect{}
	}

	return Rect{X: x, Y: y, Width: width, Height: height}
}

// Union returns the smallest rectangle that contains both rectangles.
// If either rectangle is empty, returns the other rectangle.
func (r Rect) Union(other Rect) Rect {
	if r.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return r
	}

	x := min(r.X, other.X)
	y := min(r.Y, other.Y)
	right := max(r.Right(), other.Right())
	bottom := max(r.Bottom(), other.Bottom())

	return Rect{X: x, Y: y, Width: right - x, Height: bottom - y}
}

// Intersects returns true if the two rectangles overlap.
// Touching edges do not count as overlapping.
func (r Rect) Intersects(other Rect) bool {
	return !r.Intersect(other).IsEmpty()
}

// Clamp constrains a point to be within the rectangle bounds.
// Returns the clamped (x, y) coordinates.
func (r Rect) Clamp(x, y int) (int, int) {
	if r.IsEmpty() {
		return r.X, r.Y
	}

	// Clamp x to [r.X, r.Right()-1]
	if x < r.X {
		x = r.X
	} else if x >= r.Right() {
		x = r.Right() - 1
	}

	// Clamp y to [r.Y, r.Bottom()-1]
	if y < r.Y {
		y = r.Y
	} else if y >= r.Bottom() {
		y = r.Bottom() - 1
	}

	return x, y
}
