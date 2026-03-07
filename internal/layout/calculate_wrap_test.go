package layout

import "testing"

func TestFlexWrap_BasicLineBreak(t *testing.T) {
	type tc struct {
		parentWidth    int
		childWidths    []int
		childHeight    int
		expectedLines  int // number of lines (verified by Y positions)
		expectedYs     []int
		expectedXs     []int
		expectedWidths []int
	}

	tests := map[string]tc{
		"all fit on one line": {
			parentWidth:    100,
			childWidths:    []int{20, 30, 40},
			childHeight:    10,
			expectedLines:  1,
			expectedYs:     []int{0, 0, 0},
			expectedXs:     []int{0, 20, 50},
			expectedWidths: []int{20, 30, 40},
		},
		"two lines": {
			parentWidth:    60,
			childWidths:    []int{30, 30, 30},
			childHeight:    10,
			expectedLines:  2,
			expectedYs:     []int{0, 0, 10},
			expectedXs:     []int{0, 30, 0},
			expectedWidths: []int{30, 30, 30},
		},
		"three lines": {
			parentWidth:    30,
			childWidths:    []int{20, 20, 20},
			childHeight:    10,
			expectedLines:  3,
			expectedYs:     []int{0, 10, 20},
			expectedXs:     []int{0, 0, 0},
			expectedWidths: []int{20, 20, 20},
		},
		"single item per line": {
			parentWidth:    15,
			childWidths:    []int{20, 20},
			childHeight:    10,
			expectedLines:  2,
			expectedYs:     []int{0, 10},
			expectedXs:     []int{0, 0},
			expectedWidths: []int{20, 20},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := newTestNode(DefaultStyle())
			parent.style.Width = Fixed(tt.parentWidth)
			parent.style.Height = Fixed(100)
			parent.style.Display = DisplayFlex
			parent.style.Direction = Row
			parent.style.FlexWrap = Wrap
			parent.style.AlignItems = AlignStart

			children := make([]*testNode, len(tt.childWidths))
			for i, w := range tt.childWidths {
				child := newTestNode(DefaultStyle())
				child.style.Width = Fixed(w)
				child.style.Height = Fixed(tt.childHeight)
				child.style.FlexShrink = 0 // Don't shrink; wrap instead
				children[i] = child
			}
			parent.AddChild(children...)

			Calculate(parent, tt.parentWidth, 100)

			for i, child := range children {
				if child.layout.Rect.X != tt.expectedXs[i] {
					t.Errorf("child[%d].X = %d, want %d", i, child.layout.Rect.X, tt.expectedXs[i])
				}
				if child.layout.Rect.Y != tt.expectedYs[i] {
					t.Errorf("child[%d].Y = %d, want %d", i, child.layout.Rect.Y, tt.expectedYs[i])
				}
				if child.layout.Rect.Width != tt.expectedWidths[i] {
					t.Errorf("child[%d].Width = %d, want %d", i, child.layout.Rect.Width, tt.expectedWidths[i])
				}
			}
		})
	}
}

func TestFlexWrap_NoWrapPreservesExistingBehavior(t *testing.T) {
	// With WrapNone (default), children that overflow should shrink, not wrap
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(50)
	parent.style.Height = Fixed(50)
	parent.style.Display = DisplayFlex
	parent.style.Direction = Row

	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(30)
	child1.style.Height = Fixed(10)
	child1.style.MinWidth = Fixed(0) // Allow shrink below intrinsic

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(30)
	child2.style.Height = Fixed(10)
	child2.style.MinWidth = Fixed(0)

	parent.AddChild(child1, child2)
	Calculate(parent, 50, 50)

	// Both should be on the same Y (no wrapping)
	if child1.layout.Rect.Y != 0 {
		t.Errorf("child1.Y = %d, want 0", child1.layout.Rect.Y)
	}
	if child2.layout.Rect.Y != 0 {
		t.Errorf("child2.Y = %d, want 0", child2.layout.Rect.Y)
	}
}

func TestFlexWrap_GrowWithinLine(t *testing.T) {
	// Items with flex-grow should expand to fill their line
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Display = DisplayFlex
	parent.style.Direction = Row
	parent.style.FlexWrap = Wrap
	parent.style.AlignItems = AlignStart

	// Line 1: 40 + grow item
	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(40)
	child1.style.Height = Fixed(10)
	child1.style.FlexShrink = 0

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(40)
	child2.style.Height = Fixed(10)
	child2.style.FlexGrow = 1
	child2.style.FlexShrink = 0

	// Line 2: forces wrap (40+40+40 > 100)
	child3 := newTestNode(DefaultStyle())
	child3.style.Width = Fixed(40)
	child3.style.Height = Fixed(10)
	child3.style.FlexShrink = 0

	parent.AddChild(child1, child2, child3)
	Calculate(parent, 100, 100)

	// child2 should grow to fill remaining space on line 1 (100 - 40 = 60)
	if child2.layout.Rect.Width != 60 {
		t.Errorf("child2.Width = %d, want 60", child2.layout.Rect.Width)
	}

	// child3 should be on line 2
	if child3.layout.Rect.Y != 10 {
		t.Errorf("child3.Y = %d, want 10", child3.layout.Rect.Y)
	}
}

func TestFlexWrap_AutoCrossAxisGrowth(t *testing.T) {
	type tc struct {
		parentWidth          int
		parentHeight         int // 0 means Auto
		childWidths          []int
		childHeight          int
		expectedParentHeight int
	}

	tests := map[string]tc{
		"auto height grows to fit two lines": {
			parentWidth:          60,
			parentHeight:         0, // Auto
			childWidths:          []int{30, 30, 30},
			childHeight:          10,
			expectedParentHeight: 20, // line1=[30,30], line2=[30] => 2 lines * 10
		},
		"auto height grows to fit three lines": {
			parentWidth:          25,
			parentHeight:         0,
			childWidths:          []int{20, 20, 20},
			childHeight:          10,
			expectedParentHeight: 30, // 3 lines * 10
		},
		"fixed height does not grow": {
			parentWidth:          50,
			parentHeight:         15, // Fixed
			childWidths:          []int{30, 30, 30},
			childHeight:          10,
			expectedParentHeight: 15, // Stays at fixed 15
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := newTestNode(DefaultStyle())
			parent.style.Width = Fixed(tt.parentWidth)
			if tt.parentHeight > 0 {
				parent.style.Height = Fixed(tt.parentHeight)
			}
			// Leave Height as Auto for auto-sizing
			parent.style.Display = DisplayFlex
			parent.style.Direction = Row
			parent.style.FlexWrap = Wrap
			parent.style.AlignItems = AlignStart

			children := make([]*testNode, len(tt.childWidths))
			for i, w := range tt.childWidths {
				child := newTestNode(DefaultStyle())
				child.style.Width = Fixed(w)
				child.style.Height = Fixed(tt.childHeight)
				child.style.FlexShrink = 0
				children[i] = child
			}
			parent.AddChild(children...)

			Calculate(parent, 200, 200)

			if parent.layout.Rect.Height != tt.expectedParentHeight {
				t.Errorf("parent height = %d, want %d", parent.layout.Rect.Height, tt.expectedParentHeight)
			}
		})
	}
}

func TestFlexWrap_AlignContent(t *testing.T) {
	type tc struct {
		alignContent   AlignContent
		expectedLineYs []int // Y position of first item on each line
	}

	// Setup: 50x100 parent, 3 items of width 40, height 10 each.
	// Each item alone takes a full line (40 < 50, but 40+40=80 > 50).
	// 3 lines * 10 = 30 used. 100 - 30 = 70 free cross space.

	tests := map[string]tc{
		"content-start": {
			alignContent:   ContentStart,
			expectedLineYs: []int{0, 10, 20},
		},
		"content-end": {
			alignContent:   ContentEnd,
			expectedLineYs: []int{70, 80, 90},
		},
		"content-center": {
			alignContent:   ContentCenter,
			expectedLineYs: []int{35, 45, 55},
		},
		"content-space-between": {
			alignContent:   ContentSpaceBetween,
			expectedLineYs: []int{0, 45, 90},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := newTestNode(DefaultStyle())
			parent.style.Width = Fixed(50)
			parent.style.Height = Fixed(100)
			parent.style.Display = DisplayFlex
			parent.style.Direction = Row
			parent.style.FlexWrap = Wrap
			parent.style.AlignContent = tt.alignContent
			parent.style.AlignItems = AlignStart

			children := make([]*testNode, 3)
			for i := range children {
				child := newTestNode(DefaultStyle())
				child.style.Width = Fixed(40)
				child.style.Height = Fixed(10)
				child.style.FlexShrink = 0
				children[i] = child
			}
			parent.AddChild(children...)

			Calculate(parent, 200, 200)

			for i, child := range children {
				if child.layout.Rect.Y != tt.expectedLineYs[i] {
					t.Errorf("child[%d].Y = %d, want %d", i, child.layout.Rect.Y, tt.expectedLineYs[i])
				}
			}
		})
	}
}

func TestFlexWrap_WrapReverse(t *testing.T) {
	// WrapReverse should reverse line order on the cross axis
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(50)
	parent.style.Height = Fixed(100)
	parent.style.Display = DisplayFlex
	parent.style.Direction = Row
	parent.style.FlexWrap = WrapReverse
	parent.style.AlignItems = AlignStart

	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(40)
	child1.style.Height = Fixed(10)
	child1.style.FlexShrink = 0

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(40)
	child2.style.Height = Fixed(10)
	child2.style.FlexShrink = 0

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// With WrapReverse, line order is reversed:
	// Line 1 (child1) should be below line 2 (child2)
	if child2.layout.Rect.Y >= child1.layout.Rect.Y {
		t.Errorf("WrapReverse: child2.Y (%d) should be < child1.Y (%d)",
			child2.layout.Rect.Y, child1.layout.Rect.Y)
	}
}

func TestFlexWrap_ColumnDirection(t *testing.T) {
	// Flex-wrap in column direction wraps along the X axis
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(30)
	parent.style.Display = DisplayFlex
	parent.style.Direction = Column
	parent.style.FlexWrap = Wrap
	parent.style.AlignItems = AlignStart

	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(20)
	child1.style.Height = Fixed(20)
	child1.style.FlexShrink = 0

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(20)
	child2.style.Height = Fixed(20)
	child2.style.FlexShrink = 0

	child3 := newTestNode(DefaultStyle())
	child3.style.Width = Fixed(20)
	child3.style.Height = Fixed(20)
	child3.style.FlexShrink = 0

	parent.AddChild(child1, child2, child3)
	Calculate(parent, 100, 30)

	// Column direction with height 30: each child is 20 tall
	// Line 1: child1 (20), can't fit child2 (20+20=40>30)
	// Line 2: child2 (20), can't fit child3
	// Line 3: child3
	// So child1.X=0, child2.X=20, child3.X=40
	if child1.layout.Rect.X != 0 {
		t.Errorf("child1.X = %d, want 0", child1.layout.Rect.X)
	}
	if child2.layout.Rect.X != 20 {
		t.Errorf("child2.X = %d, want 20", child2.layout.Rect.X)
	}
	if child3.layout.Rect.X != 40 {
		t.Errorf("child3.X = %d, want 40", child3.layout.Rect.X)
	}
}
