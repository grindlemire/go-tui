package tui

import (
	"testing"
)

func TestNewRect(t *testing.T) {
	r := NewRect(5, 10, 20, 15)

	if r.X != 5 {
		t.Errorf("NewRect().X = %d, want 5", r.X)
	}
	if r.Y != 10 {
		t.Errorf("NewRect().Y = %d, want 10", r.Y)
	}
	if r.Width != 20 {
		t.Errorf("NewRect().Width = %d, want 20", r.Width)
	}
	if r.Height != 15 {
		t.Errorf("NewRect().Height = %d, want 15", r.Height)
	}
}

func TestRect_RightBottom(t *testing.T) {
	type tc struct {
		rect   Rect
		right  int
		bottom int
	}

	tests := map[string]tc{
		"standard rect": {
			rect:   NewRect(5, 10, 20, 15),
			right:  25,
			bottom: 25,
		},
		"zero position": {
			rect:   NewRect(0, 0, 10, 10),
			right:  10,
			bottom: 10,
		},
		"negative position": {
			rect:   NewRect(-5, -5, 10, 10),
			right:  5,
			bottom: 5,
		},
		"zero size": {
			rect:   NewRect(5, 5, 0, 0),
			right:  5,
			bottom: 5,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.rect.Right(); got != tt.right {
				t.Errorf("Right() = %d, want %d", got, tt.right)
			}
			if got := tt.rect.Bottom(); got != tt.bottom {
				t.Errorf("Bottom() = %d, want %d", got, tt.bottom)
			}
		})
	}
}

func TestRect_Area(t *testing.T) {
	type tc struct {
		rect Rect
		area int
	}

	tests := map[string]tc{
		"standard rect": {
			rect: NewRect(0, 0, 10, 5),
			area: 50,
		},
		"zero width": {
			rect: NewRect(0, 0, 0, 10),
			area: 0,
		},
		"zero height": {
			rect: NewRect(0, 0, 10, 0),
			area: 0,
		},
		"negative width": {
			rect: NewRect(0, 0, -5, 10),
			area: 0,
		},
		"negative height": {
			rect: NewRect(0, 0, 10, -5),
			area: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.rect.Area(); got != tt.area {
				t.Errorf("Area() = %d, want %d", got, tt.area)
			}
		})
	}
}

func TestRect_IsEmpty(t *testing.T) {
	type tc struct {
		rect    Rect
		isEmpty bool
	}

	tests := map[string]tc{
		"standard rect": {
			rect:    NewRect(0, 0, 10, 5),
			isEmpty: false,
		},
		"zero width": {
			rect:    NewRect(0, 0, 0, 10),
			isEmpty: true,
		},
		"zero height": {
			rect:    NewRect(0, 0, 10, 0),
			isEmpty: true,
		},
		"negative width": {
			rect:    NewRect(0, 0, -5, 10),
			isEmpty: true,
		},
		"negative height": {
			rect:    NewRect(0, 0, 10, -5),
			isEmpty: true,
		},
		"zero rect": {
			rect:    Rect{},
			isEmpty: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.rect.IsEmpty(); got != tt.isEmpty {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.isEmpty)
			}
		})
	}
}

func TestRect_Contains(t *testing.T) {
	type tc struct {
		rect     Rect
		x, y     int
		contains bool
	}

	r := NewRect(10, 20, 30, 40)

	tests := map[string]tc{
		"point inside": {
			rect:     r,
			x:        20,
			y:        30,
			contains: true,
		},
		"top-left corner (inside)": {
			rect:     r,
			x:        10,
			y:        20,
			contains: true,
		},
		"right edge (outside)": {
			rect:     r,
			x:        40,
			y:        30,
			contains: false,
		},
		"bottom edge (outside)": {
			rect:     r,
			x:        20,
			y:        60,
			contains: false,
		},
		"bottom-right corner (outside)": {
			rect:     r,
			x:        40,
			y:        60,
			contains: false,
		},
		"point left of rect": {
			rect:     r,
			x:        5,
			y:        30,
			contains: false,
		},
		"point above rect": {
			rect:     r,
			x:        20,
			y:        10,
			contains: false,
		},
		"point right of rect": {
			rect:     r,
			x:        50,
			y:        30,
			contains: false,
		},
		"point below rect": {
			rect:     r,
			x:        20,
			y:        70,
			contains: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.rect.Contains(tt.x, tt.y); got != tt.contains {
				t.Errorf("Contains(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.contains)
			}
		})
	}
}

func TestRect_ContainsRect(t *testing.T) {
	type tc struct {
		outer    Rect
		inner    Rect
		contains bool
	}

	tests := map[string]tc{
		"fully contained": {
			outer:    NewRect(0, 0, 100, 100),
			inner:    NewRect(10, 10, 20, 20),
			contains: true,
		},
		"same rect": {
			outer:    NewRect(10, 10, 20, 20),
			inner:    NewRect(10, 10, 20, 20),
			contains: true,
		},
		"partial overlap left": {
			outer:    NewRect(10, 10, 20, 20),
			inner:    NewRect(5, 15, 10, 10),
			contains: false,
		},
		"partial overlap right": {
			outer:    NewRect(10, 10, 20, 20),
			inner:    NewRect(25, 15, 10, 10),
			contains: false,
		},
		"partial overlap top": {
			outer:    NewRect(10, 10, 20, 20),
			inner:    NewRect(15, 5, 10, 10),
			contains: false,
		},
		"partial overlap bottom": {
			outer:    NewRect(10, 10, 20, 20),
			inner:    NewRect(15, 25, 10, 10),
			contains: false,
		},
		"disjoint": {
			outer:    NewRect(0, 0, 10, 10),
			inner:    NewRect(20, 20, 10, 10),
			contains: false,
		},
		"empty inner": {
			outer:    NewRect(0, 0, 10, 10),
			inner:    NewRect(5, 5, 0, 0),
			contains: true,
		},
		"empty outer": {
			outer:    NewRect(0, 0, 0, 0),
			inner:    NewRect(0, 0, 10, 10),
			contains: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.outer.ContainsRect(tt.inner); got != tt.contains {
				t.Errorf("ContainsRect() = %v, want %v", got, tt.contains)
			}
		})
	}
}

func TestRect_Inset(t *testing.T) {
	type tc struct {
		rect                       Rect
		top, right, bottom, left   int
		expectedX, expectedY       int
		expectedWidth, expectedHeight int
	}

	tests := map[string]tc{
		"uniform positive inset": {
			rect:           NewRect(10, 10, 100, 100),
			top:            5,
			right:          5,
			bottom:         5,
			left:           5,
			expectedX:      15,
			expectedY:      15,
			expectedWidth:  90,
			expectedHeight: 90,
		},
		"different insets": {
			rect:           NewRect(0, 0, 100, 100),
			top:            10,
			right:          20,
			bottom:         30,
			left:           40,
			expectedX:      40,
			expectedY:      10,
			expectedWidth:  40,
			expectedHeight: 60,
		},
		"negative insets (expand)": {
			rect:           NewRect(10, 10, 50, 50),
			top:            -5,
			right:          -5,
			bottom:         -5,
			left:           -5,
			expectedX:      5,
			expectedY:      5,
			expectedWidth:  60,
			expectedHeight: 60,
		},
		"inset to zero": {
			rect:           NewRect(0, 0, 10, 10),
			top:            5,
			right:          5,
			bottom:         5,
			left:           5,
			expectedX:      5,
			expectedY:      5,
			expectedWidth:  0,
			expectedHeight: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.rect.Inset(tt.top, tt.right, tt.bottom, tt.left)
			if got.X != tt.expectedX || got.Y != tt.expectedY ||
				got.Width != tt.expectedWidth || got.Height != tt.expectedHeight {
				t.Errorf("Inset() = {%d, %d, %d, %d}, want {%d, %d, %d, %d}",
					got.X, got.Y, got.Width, got.Height,
					tt.expectedX, tt.expectedY, tt.expectedWidth, tt.expectedHeight)
			}
		})
	}
}

func TestRect_InsetUniform(t *testing.T) {
	r := NewRect(10, 10, 100, 100)
	got := r.InsetUniform(10)

	if got.X != 20 || got.Y != 20 || got.Width != 80 || got.Height != 80 {
		t.Errorf("InsetUniform(10) = {%d, %d, %d, %d}, want {20, 20, 80, 80}",
			got.X, got.Y, got.Width, got.Height)
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
		"adjacent vertical (no overlap)": {
			a:        NewRect(0, 0, 10, 10),
			b:        NewRect(0, 10, 10, 10),
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
		"same rect": {
			a:        NewRect(10, 10, 20, 20),
			b:        NewRect(10, 10, 20, 20),
			expected: NewRect(10, 10, 20, 20),
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

func TestRect_Clamp(t *testing.T) {
	type tc struct {
		rect       Rect
		x, y       int
		expectedX  int
		expectedY  int
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
		"point outside all corners": {
			rect:      r,
			x:         100,
			y:         100,
			expectedX: 39,
			expectedY: 59,
		},
		"point at exact right edge": {
			rect:      r,
			x:         40,
			y:         30,
			expectedX: 39,
			expectedY: 30,
		},
		"point at exact bottom edge": {
			rect:      r,
			x:         20,
			y:         60,
			expectedX: 20,
			expectedY: 59,
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
	_ = original.Inset(5, 5, 5, 5)
	_ = original.InsetUniform(5)
	_ = original.Intersect(NewRect(0, 0, 100, 100))
	_ = original.Union(NewRect(50, 50, 20, 20))
	_ = original.Translate(10, 10)

	// Original should be unchanged
	if original.X != 10 || original.Y != 10 || original.Width != 20 || original.Height != 20 {
		t.Error("original rect was modified by method calls")
	}
}
