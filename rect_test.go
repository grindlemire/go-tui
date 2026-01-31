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

