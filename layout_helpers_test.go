package tui

import "testing"

func TestEdgeSymmetric(t *testing.T) {
	type tc struct {
		v, h int
		want Edges
	}

	tests := map[string]tc{
		"vertical and horizontal differ": {
			v:    2,
			h:    5,
			want: Edges{Top: 2, Right: 5, Bottom: 2, Left: 5},
		},
		"equal values": {
			v:    3,
			h:    3,
			want: Edges{Top: 3, Right: 3, Bottom: 3, Left: 3},
		},
		"zero values": {
			v:    0,
			h:    0,
			want: Edges{},
		},
		"vertical only": {
			v:    4,
			h:    0,
			want: Edges{Top: 4, Bottom: 4},
		},
		"negative values": {
			v:    -1,
			h:    -2,
			want: Edges{Top: -1, Right: -2, Bottom: -1, Left: -2},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := EdgeSymmetric(tt.v, tt.h)
			if got != tt.want {
				t.Errorf("EdgeSymmetric(%d, %d) = %+v, want %+v", tt.v, tt.h, got, tt.want)
			}
		})
	}
}

func TestInsetRect(t *testing.T) {
	type tc struct {
		r                        Rect
		top, right, bottom, left int
		want                     Rect
	}

	tests := map[string]tc{
		"uniform inset": {
			r:   NewRect(0, 0, 10, 10),
			top: 1, right: 1, bottom: 1, left: 1,
			want: NewRect(1, 1, 8, 8),
		},
		"asymmetric inset": {
			r:   NewRect(5, 5, 20, 10),
			top: 1, right: 2, bottom: 3, left: 4,
			want: NewRect(9, 6, 14, 6),
		},
		"zero inset returns original": {
			r:   NewRect(2, 3, 7, 8),
			top: 0, right: 0, bottom: 0, left: 0,
			want: NewRect(2, 3, 7, 8),
		},
		"negative inset expands": {
			r:   NewRect(5, 5, 10, 10),
			top: -1, right: -2, bottom: -3, left: -4,
			want: NewRect(1, 4, 16, 14),
		},
		"inset larger than rect goes negative": {
			r:   NewRect(0, 0, 4, 4),
			top: 3, right: 3, bottom: 3, left: 3,
			want: NewRect(3, 3, -2, -2),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := InsetRect(tt.r, tt.top, tt.right, tt.bottom, tt.left)
			if got != tt.want {
				t.Errorf("InsetRect(%+v, %d, %d, %d, %d) = %+v, want %+v",
					tt.r, tt.top, tt.right, tt.bottom, tt.left, got, tt.want)
			}
		})
	}
}
