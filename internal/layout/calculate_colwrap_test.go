package layout

import "testing"

// Issue #130: in a column, a non-stretch auto-width child is measured for
// wrapping at the width it will actually receive, min(intrinsic, available),
// so the main-axis size grows for the wrapping the clamp forces.
func TestColumnNonStretchChildWrapsAtClampedWidth(t *testing.T) {
	type tc struct {
		parentWidth int
		alignItems  Align
		alignSelf   *Align
		// Width the child must be measured at: min(intrinsic 32, available).
		wantMeasureWidth int
		wantWidth        int
		wantHeight       int
	}

	start := AlignStart
	tests := map[string]tc{
		"items-start clamps and wraps": {
			parentWidth:      20,
			alignItems:       AlignStart,
			wantMeasureWidth: 20,
			wantWidth:        20,
			wantHeight:       2,
		},
		"self-start clamps and wraps": {
			parentWidth:      20,
			alignItems:       AlignStretch,
			alignSelf:        &start,
			wantMeasureWidth: 20,
			wantWidth:        20,
			wantHeight:       2,
		},
		"stretch still wraps": {
			parentWidth:      20,
			alignItems:       AlignStretch,
			wantMeasureWidth: 20,
			wantWidth:        20,
			wantHeight:       2,
		},
		"wide parent keeps intrinsic size": {
			parentWidth:      40,
			alignItems:       AlignStart,
			wantMeasureWidth: 32,
			wantWidth:        32,
			wantHeight:       1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := newTestNode(DefaultStyle())
			parent.style.Display = DisplayFlex
			parent.style.Direction = Column
			parent.style.AlignItems = tt.alignItems
			parent.style.Width = Fixed(tt.parentWidth)
			parent.style.Height = Fixed(24)

			// Simulates a table or text that is one row wide at 32 columns
			// and wraps to two rows once narrower than that.
			child := newTestNode(DefaultStyle())
			child.style.AlignSelf = tt.alignSelf
			child.SetIntrinsicSize(32, 1)
			gotWidth := 0
			child.heightForWidth = func(width int) int {
				gotWidth = width
				if width < 32 {
					return 2
				}
				return 1
			}
			sibling := newTestNode(DefaultStyle())
			sibling.SetIntrinsicSize(5, 1)
			parent.AddChild(child, sibling)

			Calculate(parent, tt.parentWidth, 24)

			if gotWidth != tt.wantMeasureWidth {
				t.Errorf("measured at width %d, want %d", gotWidth, tt.wantMeasureWidth)
			}
			rect := child.layout.Rect
			if rect.Width != tt.wantWidth {
				t.Errorf("child width = %d, want %d", rect.Width, tt.wantWidth)
			}
			if rect.Height != tt.wantHeight {
				t.Errorf("child height = %d, want %d", rect.Height, tt.wantHeight)
			}
			if got := sibling.layout.Rect.Y; got != tt.wantHeight {
				t.Errorf("sibling Y = %d, want %d (below the child)", got, tt.wantHeight)
			}
		})
	}
}
