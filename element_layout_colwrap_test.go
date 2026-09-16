package tui

import "testing"

// Issue #130: a non-stretch auto-width child in a column is laid out at
// min(intrinsic, available) but was measured for wrapping at its unclamped
// intrinsic width, so the height allocated to it skipped the wrap growth.
func TestColumnNonStretchChildMeasuredAtClampedWidth(t *testing.T) {
	type tc struct {
		containerWidth int
		containerOpts  []Option
		childOpts      []Option
		useTable       bool
		wantChildWidth int
		wantHeight     int
	}

	tests := map[string]tc{
		"table clamped by items-start": {
			containerWidth: 20,
			containerOpts:  []Option{WithAlign(AlignStart)},
			useTable:       true,
			wantChildWidth: 20,
			wantHeight:     2,
		},
		"text clamped by items-start": {
			containerWidth: 20,
			containerOpts:  []Option{WithAlign(AlignStart)},
			wantChildWidth: 20,
			wantHeight:     2,
		},
		"text clamped by self-start": {
			containerWidth: 20,
			childOpts:      []Option{WithAlignSelf(AlignStart)},
			wantChildWidth: 20,
			wantHeight:     2,
		},
		"table clamped by self-center": {
			containerWidth: 20,
			childOpts:      []Option{WithAlignSelf(AlignCenter)},
			useTable:       true,
			wantChildWidth: 20,
			wantHeight:     2,
		},
		"stretch default still wraps": {
			containerWidth: 20,
			useTable:       true,
			wantChildWidth: 20,
			wantHeight:     2,
		},
		"container wider than intrinsic does not clamp": {
			containerWidth: 40,
			containerOpts:  []Option{WithAlign(AlignStart)},
			useTable:       true,
			wantChildWidth: 32,
			wantHeight:     1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var child, cell *Element
			if tt.useTable {
				child, cell = newWrapTestTable()
				for _, opt := range tt.childOpts {
					opt(child)
				}
			} else {
				child = New(append([]Option{WithText(tableWrapCellText)}, tt.childOpts...)...)
			}
			sentinel := New(WithText("SENTINEL"))

			rootOpts := append([]Option{
				WithDisplay(DisplayFlex),
				WithDirection(Column),
				WithWidth(tt.containerWidth),
			}, tt.containerOpts...)
			root := New(rootOpts...)
			root.AddChild(child)
			root.AddChild(sentinel)
			root.Calculate(tt.containerWidth, 24)

			rect := child.Rect()
			if rect.Width != tt.wantChildWidth {
				t.Errorf("child width = %d, want %d", rect.Width, tt.wantChildWidth)
			}
			if rect.Height != tt.wantHeight {
				t.Errorf("child height = %d, want %d", rect.Height, tt.wantHeight)
			}
			if got := sentinel.Rect().Y; got != rect.Y+tt.wantHeight {
				t.Errorf("sentinel Y = %d, want %d (below the child)", got, rect.Y+tt.wantHeight)
			}
			if cell != nil {
				cr := cell.Rect()
				if cr.Y+cr.Height > rect.Y+rect.Height {
					t.Errorf("cell bottom %d overflows table bottom %d", cr.Y+cr.Height, rect.Y+rect.Height)
				}
			}
		})
	}
}

// HeightForWidth on a column container measures each non-stretch child at
// the width it will receive, min(intrinsic, content width), matching layout.
func TestColumnHeightForWidthClampsNonStretchChild(t *testing.T) {
	type tc struct {
		width int
		want  int
	}

	tests := map[string]tc{
		"narrower than intrinsic wraps": {width: 20, want: 2},
		"wider than intrinsic fits":     {width: 40, want: 1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			table, _ := newWrapTestTable()
			container := New(WithDisplay(DisplayFlex), WithDirection(Column), WithAlign(AlignStart))
			container.AddChild(table)

			if got := container.HeightForWidth(tt.width); got != tt.want {
				t.Errorf("HeightForWidth(%d) = %d, want %d", tt.width, got, tt.want)
			}
		})
	}
}
