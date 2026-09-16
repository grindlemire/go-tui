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

// Issue #164: computeBorderBox applies MinWidth/MaxWidth after the slot is
// assigned. The nested column pins both wrap measurements at that width:
// recomputeTextWrapping for the child and Element.HeightForWidth for inner.
func TestColumnChildMeasuredAtMinMaxWidth(t *testing.T) {
	type tc struct {
		containerWidth int
		innerOpts      []Option
		childOpts      []Option
		wantChildWidth int
		wantHeight     int
	}

	tests := map[string]tc{
		"max-w-15 items-start": {
			containerWidth: 20,
			innerOpts:      []Option{WithAlign(AlignStart)},
			childOpts:      []Option{WithMaxWidth(15)},
			wantChildWidth: 15,
			wantHeight:     3,
		},
		"max-w-15 stretch default": {
			containerWidth: 20,
			childOpts:      []Option{WithMaxWidth(15)},
			wantChildWidth: 15,
			wantHeight:     3,
		},
		"max-w-10 items-start": {
			containerWidth: 20,
			innerOpts:      []Option{WithAlign(AlignStart)},
			childOpts:      []Option{WithMaxWidth(10)},
			wantChildWidth: 10,
			wantHeight:     5,
		},
		// A min width above the slot overflows the column; the child must
		// be measured at the overflowing width, not the narrower slot.
		"min-w-25 overflows 12 wide items-start": {
			containerWidth: 12,
			innerOpts:      []Option{WithAlign(AlignStart)},
			childOpts:      []Option{WithMinWidth(25)},
			wantChildWidth: 25,
			wantHeight:     2,
		},
		"min-w-25 overflows 12 wide stretch default": {
			containerWidth: 12,
			childOpts:      []Option{WithMinWidth(25)},
			wantChildWidth: 25,
			wantHeight:     2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			child := New(append([]Option{WithText(tableWrapCellText)}, tt.childOpts...)...)
			innerOpts := append([]Option{WithDisplay(DisplayFlex), WithDirection(Column)}, tt.innerOpts...)
			inner := New(innerOpts...)
			inner.AddChild(child)
			sentinel := New(WithText("SENTINEL"))
			outer := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(tt.containerWidth))
			outer.AddChild(inner)
			outer.AddChild(sentinel)

			if got := inner.HeightForWidth(tt.containerWidth); got != tt.wantHeight {
				t.Errorf("inner.HeightForWidth(%d) = %d, want %d", tt.containerWidth, got, tt.wantHeight)
			}

			outer.Calculate(tt.containerWidth, 24)

			rect := child.Rect()
			if rect.Width != tt.wantChildWidth {
				t.Errorf("child width = %d, want %d", rect.Width, tt.wantChildWidth)
			}
			if rect.Height != tt.wantHeight {
				t.Errorf("child height = %d, want %d", rect.Height, tt.wantHeight)
			}
			if got := inner.Rect().Height; got != tt.wantHeight {
				t.Errorf("inner column height = %d, want %d", got, tt.wantHeight)
			}
			if got := sentinel.Rect().Y; got != tt.wantHeight {
				t.Errorf("sentinel Y = %d, want %d (below the inner column)", got, tt.wantHeight)
			}
		})
	}
}

// Issue #165: the column branch of HeightForWidth measured every child at the
// full content width and ignored margins in both directions, so a nested
// column reported less height than layout gives its children.
func TestColumnHeightForWidthIncludesChildMargins(t *testing.T) {
	type tc struct {
		innerOpts     []Option
		childOpts     []Option
		wantInnerH    int
		wantChildRect Rect
		wantSentinelY int
	}

	tests := map[string]tc{
		"horizontal margins items-start": {
			innerOpts:     []Option{WithAlign(AlignStart)},
			childOpts:     []Option{WithMarginTRBL(0, 2, 0, 2)},
			wantInnerH:    3,
			wantChildRect: Rect{X: 2, Y: 0, Width: 16, Height: 3},
			wantSentinelY: 3,
		},
		"horizontal margins stretch default": {
			childOpts:     []Option{WithMarginTRBL(0, 2, 0, 2)},
			wantInnerH:    3,
			wantChildRect: Rect{X: 2, Y: 0, Width: 16, Height: 3},
			wantSentinelY: 3,
		},
		"vertical margins add rows": {
			childOpts:     []Option{WithMarginTRBL(1, 0, 1, 0)},
			wantInnerH:    4,
			wantChildRect: Rect{X: 0, Y: 1, Width: 20, Height: 2},
			wantSentinelY: 4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			text := New(append([]Option{WithText(tableWrapCellText)}, tt.childOpts...)...)
			innerOpts := append([]Option{WithDisplay(DisplayFlex), WithDirection(Column)}, tt.innerOpts...)
			inner := New(innerOpts...)
			inner.AddChild(text)
			sentinel := New(WithText("SENTINEL"))
			outer := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(20))
			outer.AddChild(inner)
			outer.AddChild(sentinel)

			if got := inner.HeightForWidth(20); got != tt.wantInnerH {
				t.Errorf("inner.HeightForWidth(20) = %d, want %d", got, tt.wantInnerH)
			}

			outer.Calculate(20, 24)

			if got := text.Rect(); got != tt.wantChildRect {
				t.Errorf("text rect = %+v, want %+v", got, tt.wantChildRect)
			}
			if got := inner.Rect().Height; got != tt.wantInnerH {
				t.Errorf("inner column height = %d, want %d", got, tt.wantInnerH)
			}
			if got := sentinel.Rect().Y; got != tt.wantSentinelY {
				t.Errorf("sentinel Y = %d, want %d (below the inner column)", got, tt.wantSentinelY)
			}
		})
	}
}
