package layout

// Point represents an (X, Y) coordinate.
type Point struct {
	X, Y int
}

// Add returns a new Point offset by other.
func (p Point) Add(other Point) Point {
	return Point{X: p.X + other.X, Y: p.Y + other.Y}
}

// Sub returns a new Point with other subtracted.
func (p Point) Sub(other Point) Point {
	return Point{X: p.X - other.X, Y: p.Y - other.Y}
}

// In returns true if the point is inside the given rectangle.
func (p Point) In(r Rect) bool {
	return r.Contains(p.X, p.Y)
}
