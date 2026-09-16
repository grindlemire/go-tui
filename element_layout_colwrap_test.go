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
// min(intrinsic, content width), matching Phase 5. Only observable when the
// child's height depends on its own width: a percent-width text with min-w-0
// wraps at the row's intrinsic 29 columns but not at the full 60.
func TestColumnHeightForWidthClampsNonStretchChild(t *testing.T) {
	text := New(WithText(tableWrapCellText), WithWidthPercent(50), WithMinWidth(0))
	row := New(WithDisplay(DisplayFlex), WithDirection(Row))
	row.AddChild(text)
	inner := New(WithDisplay(DisplayFlex), WithDirection(Column), WithAlign(AlignStart))
	inner.AddChild(row)
	sentinel := New(WithText("SENTINEL"))
	outer := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(60))
	outer.AddChild(inner)
	outer.AddChild(sentinel)

	if got := inner.HeightForWidth(60); got != 3 {
		t.Errorf("inner.HeightForWidth(60) = %d, want 3", got)
	}

	outer.Calculate(60, 24)

	// The row is clamped to its intrinsic 29 columns, so the text gets 14
	// and wraps to 3 rows; that height must reach the outer column.
	if got := text.Rect(); got.Width != 14 || got.Height != 3 {
		t.Errorf("text rect = %dx%d, want 14x3", got.Width, got.Height)
	}
	if got := row.Rect().Height; got != 3 {
		t.Errorf("row height = %d, want 3", got)
	}
	if got := inner.Rect().Height; got != 3 {
		t.Errorf("inner column height = %d, want 3", got)
	}
	if got := sentinel.Rect().Y; got != 3 {
		t.Errorf("sentinel Y = %d, want 3 (below the inner column)", got)
	}
}
