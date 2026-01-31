package layout

import "testing"

func TestRect_Translate(t *testing.T) {
	type tc struct {
		rect     Rect
		dx, dy   int
		expected Rect
	}

	tests := map[string]tc{
		"positive translation": {
			rect:     NewRect(10, 20, 30, 40),
			dx:       5,
			dy:       15,
			expected: NewRect(15, 35, 30, 40),
		},
		"negative translation": {
			rect:     NewRect(10, 20, 30, 40),
			dx:       -5,
			dy:       -10,
			expected: NewRect(5, 10, 30, 40),
		},
		"no translation": {
			rect:     NewRect(10, 20, 30, 40),
			dx:       0,
			dy:       0,
			expected: NewRect(10, 20, 30, 40),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.rect.Translate(tt.dx, tt.dy)
			if got != tt.expected {
				t.Errorf("Translate(%d, %d) = %+v, want %+v", tt.dx, tt.dy, got, tt.expected)
			}
		})
	}
}

func TestRect_Intersect(t *testing.T) {
	type tc struct {
		a, b     Rect
		expected Rect
	}

	tests := map[string]tc{
		"overlapping rects": {
			a:        NewRect(0, 0, 20, 20),
			b:        NewRect(10, 10, 20, 20),
			expected: NewRect(10, 10, 10, 10),
		},
		"same rect": {
			a:        NewRect(10, 10, 20, 20),
			b:        NewRect(10, 10, 20, 20),
			expected: NewRect(10, 10, 20, 20),
		},
		"one inside other": {
			a:        NewRect(0, 0, 100, 100),
			b:        NewRect(20, 20, 30, 30),
			expected: NewRect(20, 20, 30, 30),
		},
		"adjacent horizontal (no overlap)": {
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(10, 0, 10, 10),
			expected: Rect{},
		},
		"disjoint": {
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(50, 50, 10, 10),
			expected: Rect{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.a.Intersect(tt.b)
			if got != tt.expected {
				t.Errorf("Intersect() = %+v, want %+v", got, tt.expected)
			}
			// Test commutativity
			got2 := tt.b.Intersect(tt.a)
			if got2 != tt.expected {
				t.Errorf("Intersect() (reversed) = %+v, want %+v", got2, tt.expected)
			}
		})
	}
}

func TestRect_Union(t *testing.T) {
	type tc struct {
		a, b     Rect
		expected Rect
	}

	tests := map[string]tc{
		"overlapping rects": {
			a:        NewRect(0, 0, 20, 20),
			b:        NewRect(10, 10, 20, 20),
			expected: NewRect(0, 0, 30, 30),
		},
		"disjoint rects": {
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(20, 20, 10, 10),
			expected: NewRect(0, 0, 30, 30),
		},
		"one inside other": {
			a:        NewRect(0, 0, 100, 100),
			b:        NewRect(20, 20, 30, 30),
			expected: NewRect(0, 0, 100, 100),
		},
		"one empty": {
			a:        NewRect(10, 10, 20, 20),
			b:        Rect{},
			expected: NewRect(10, 10, 20, 20),
		},
		"both empty": {
			a:        Rect{},
			b:        Rect{},
			expected: Rect{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.a.Union(tt.b)
			if got != tt.expected {
				t.Errorf("Union() = %+v, want %+v", got, tt.expected)
			}
			// Test commutativity
			got2 := tt.b.Union(tt.a)
			if got2 != tt.expected {
				t.Errorf("Union() (reversed) = %+v, want %+v", got2, tt.expected)
			}
		})
	}
}

func TestRect_Clamp(t *testing.T) {
	type tc struct {
		rect      Rect
		x, y      int
		expectedX int
		expectedY int
	}

	r := NewRect(10, 20, 30, 40)

	tests := map[string]tc{
		"point inside": {
			rect:      r,
			x:         20,
			y:         30,
			expectedX: 20,
			expectedY: 30,
		},
		"point left of rect": {
			rect:      r,
			x:         5,
			y:         30,
			expectedX: 10,
			expectedY: 30,
		},
		"point above rect": {
			rect:      r,
			x:         20,
			y:         10,
			expectedX: 20,
			expectedY: 20,
		},
		"point right of rect": {
			rect:      r,
			x:         50,
			y:         30,
			expectedX: 39, // Right edge - 1
			expectedY: 30,
		},
		"point below rect": {
			rect:      r,
			x:         20,
			y:         70,
			expectedX: 20,
			expectedY: 59, // Bottom edge - 1
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotX, gotY := tt.rect.Clamp(tt.x, tt.y)
			if gotX != tt.expectedX || gotY != tt.expectedY {
				t.Errorf("Clamp(%d, %d) = (%d, %d), want (%d, %d)",
					tt.x, tt.y, gotX, gotY, tt.expectedX, tt.expectedY)
			}
		})
	}
}

func TestRect_Clamp_EmptyRect(t *testing.T) {
	empty := Rect{}
	x, y := empty.Clamp(10, 20)

	if x != 0 || y != 0 {
		t.Errorf("Clamp on empty rect = (%d, %d), want (0, 0)", x, y)
	}
}

func TestRect_Immutability(t *testing.T) {
	original := NewRect(10, 10, 20, 20)

	// All methods should return new Rects, not modify original
	_ = original.Inset(EdgeAll(5))
	_ = original.Outset(EdgeAll(5))
	_ = original.Intersect(NewRect(0, 0, 100, 100))
	_ = original.Union(NewRect(50, 50, 20, 20))
	_ = original.Translate(10, 10)

	// Original should be unchanged
	if original.X != 10 || original.Y != 10 || original.Width != 20 || original.Height != 20 {
		t.Error("original rect was modified by method calls")
	}
}

func TestEdges(t *testing.T) {
	type tc struct {
		edges      Edges
		horizontal int
		vertical   int
		isZero     bool
	}

	tests := map[string]tc{
		"EdgeAll": {
			edges:      EdgeAll(5),
			horizontal: 10,
			vertical:   10,
			isZero:     false,
		},
		"EdgeSymmetric": {
			edges:      EdgeSymmetric(10, 20),
			horizontal: 40,
			vertical:   20,
			isZero:     false,
		},
		"EdgeTRBL": {
			edges:      EdgeTRBL(1, 2, 3, 4),
			horizontal: 6,
			vertical:   4,
			isZero:     false,
		},
		"zero edges": {
			edges:      Edges{},
			horizontal: 0,
			vertical:   0,
			isZero:     true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.edges.Horizontal(); got != tt.horizontal {
				t.Errorf("Horizontal() = %d, want %d", got, tt.horizontal)
			}
			if got := tt.edges.Vertical(); got != tt.vertical {
				t.Errorf("Vertical() = %d, want %d", got, tt.vertical)
			}
			if got := tt.edges.IsZero(); got != tt.isZero {
				t.Errorf("IsZero() = %v, want %v", got, tt.isZero)
			}
		})
	}
}

func TestRect_Intersects(t *testing.T) {
	type tc struct {
		a, b       Rect
		intersects bool
	}

	tests := map[string]tc{
		"overlapping rects": {
			a:          NewRect(0, 0, 20, 20),
			b:          NewRect(10, 10, 20, 20),
			intersects: true,
		},
		"same rect": {
			a:          NewRect(10, 10, 20, 20),
			b:          NewRect(10, 10, 20, 20),
			intersects: true,
		},
		"one inside other": {
			a:          NewRect(0, 0, 100, 100),
			b:          NewRect(20, 20, 30, 30),
			intersects: true,
		},
		"adjacent horizontal (touching edges)": {
			a:          NewRect(0, 0, 10, 10),
			b:          NewRect(10, 0, 10, 10),
			intersects: false,
		},
		"adjacent vertical (touching edges)": {
			a:          NewRect(0, 0, 10, 10),
			b:          NewRect(0, 10, 10, 10),
			intersects: false,
		},
		"disjoint": {
			a:          NewRect(0, 0, 10, 10),
			b:          NewRect(50, 50, 10, 10),
			intersects: false,
		},
		"empty rect": {
			a:          NewRect(0, 0, 10, 10),
			b:          Rect{},
			intersects: false,
		},
		"both empty": {
			a:          Rect{},
			b:          Rect{},
			intersects: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.a.Intersects(tt.b)
			if got != tt.intersects {
				t.Errorf("Intersects() = %v, want %v", got, tt.intersects)
			}
			// Test commutativity
			got2 := tt.b.Intersects(tt.a)
			if got2 != tt.intersects {
				t.Errorf("Intersects() (reversed) = %v, want %v", got2, tt.intersects)
			}
		})
	}
}

func TestPoint(t *testing.T) {
	p1 := Point{X: 10, Y: 20}
	p2 := Point{X: 5, Y: 15}

	// Test Add
	sum := p1.Add(p2)
	if sum.X != 15 || sum.Y != 35 {
		t.Errorf("Add() = {%d, %d}, want {15, 35}", sum.X, sum.Y)
	}

	// Test Sub
	diff := p1.Sub(p2)
	if diff.X != 5 || diff.Y != 5 {
		t.Errorf("Sub() = {%d, %d}, want {5, 5}", diff.X, diff.Y)
	}

	// Test In
	rect := NewRect(0, 0, 50, 50)
	if !p1.In(rect) {
		t.Error("Point should be inside rect")
	}

	outsidePoint := Point{X: 100, Y: 100}
	if outsidePoint.In(rect) {
		t.Error("Point should be outside rect")
	}
}
