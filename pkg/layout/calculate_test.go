package layout

import "testing"

func TestCalculate_SingleNode_FixedSize(t *testing.T) {
	type tc struct {
		style          Style
		availableW     int
		availableH     int
		expectedWidth  int
		expectedHeight int
	}

	tests := map[string]tc{
		"fixed width and height": {
			style: func() Style {
				s := DefaultStyle()
				s.Width = Fixed(50)
				s.Height = Fixed(30)
				return s
			}(),
			availableW:     100,
			availableH:     100,
			expectedWidth:  50,
			expectedHeight: 30,
		},
		"auto fills available space": {
			style:          DefaultStyle(),
			availableW:     100,
			availableH:     80,
			expectedWidth:  100,
			expectedHeight: 80,
		},
		"percent of available": {
			style: func() Style {
				s := DefaultStyle()
				s.Width = Percent(50)
				s.Height = Percent(25)
				return s
			}(),
			availableW:     200,
			availableH:     100,
			expectedWidth:  100,
			expectedHeight: 25,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			node := newTestNode(tt.style)
			Calculate(node, tt.availableW, tt.availableH)

			if node.layout.Rect.Width != tt.expectedWidth {
				t.Errorf("Layout.Rect.Width = %d, want %d", node.layout.Rect.Width, tt.expectedWidth)
			}
			if node.layout.Rect.Height != tt.expectedHeight {
				t.Errorf("Layout.Rect.Height = %d, want %d", node.layout.Rect.Height, tt.expectedHeight)
			}
			if node.layout.Rect.X != 0 || node.layout.Rect.Y != 0 {
				t.Errorf("Layout.Rect position = (%d, %d), want (0, 0)",
					node.layout.Rect.X, node.layout.Rect.Y)
			}
			if node.IsDirty() {
				t.Error("node should not be dirty after Calculate")
			}
		})
	}
}

func TestCalculate_SingleNode_WithPadding(t *testing.T) {
	style := DefaultStyle()
	style.Width = Fixed(100)
	style.Height = Fixed(80)
	style.Padding = EdgeAll(10)

	node := newTestNode(style)
	Calculate(node, 200, 200)

	// Border box should be the full size
	if node.layout.Rect.Width != 100 || node.layout.Rect.Height != 80 {
		t.Errorf("Layout.Rect = %dx%d, want 100x80",
			node.layout.Rect.Width, node.layout.Rect.Height)
	}

	// Content rect should be inset by padding
	if node.layout.ContentRect.X != 10 || node.layout.ContentRect.Y != 10 {
		t.Errorf("ContentRect position = (%d, %d), want (10, 10)",
			node.layout.ContentRect.X, node.layout.ContentRect.Y)
	}
	if node.layout.ContentRect.Width != 80 || node.layout.ContentRect.Height != 60 {
		t.Errorf("ContentRect size = %dx%d, want 80x60",
			node.layout.ContentRect.Width, node.layout.ContentRect.Height)
	}
}

func TestCalculate_TwoChildren_Row(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(30)
	child1.style.Height = Fixed(50)

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(40)
	child2.style.Height = Fixed(50)

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Child 1 should be at position 0
	if child1.layout.Rect.X != 0 || child1.layout.Rect.Y != 0 {
		t.Errorf("child1 position = (%d, %d), want (0, 0)",
			child1.layout.Rect.X, child1.layout.Rect.Y)
	}
	if child1.layout.Rect.Width != 30 {
		t.Errorf("child1 width = %d, want 30", child1.layout.Rect.Width)
	}

	// Child 2 should be at position 30 (after child1)
	if child2.layout.Rect.X != 30 {
		t.Errorf("child2.X = %d, want 30", child2.layout.Rect.X)
	}
	if child2.layout.Rect.Width != 40 {
		t.Errorf("child2 width = %d, want 40", child2.layout.Rect.Width)
	}
}

func TestCalculate_TwoChildren_Column(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Column

	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(100)
	child1.style.Height = Fixed(30)

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(100)
	child2.style.Height = Fixed(40)

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Child 1 should be at position 0
	if child1.layout.Rect.X != 0 || child1.layout.Rect.Y != 0 {
		t.Errorf("child1 position = (%d, %d), want (0, 0)",
			child1.layout.Rect.X, child1.layout.Rect.Y)
	}
	if child1.layout.Rect.Height != 30 {
		t.Errorf("child1 height = %d, want 30", child1.layout.Rect.Height)
	}

	// Child 2 should be at Y position 30 (after child1)
	if child2.layout.Rect.Y != 30 {
		t.Errorf("child2.Y = %d, want 30", child2.layout.Rect.Y)
	}
	if child2.layout.Rect.Height != 40 {
		t.Errorf("child2 height = %d, want 40", child2.layout.Rect.Height)
	}
}

func TestCalculate_FlexGrow(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Fixed child
	fixed := newTestNode(DefaultStyle())
	fixed.style.Width = Fixed(30)
	fixed.style.Height = Fixed(50)

	// Growing child
	growing := newTestNode(DefaultStyle())
	growing.style.Width = Fixed(0) // Start at 0
	growing.style.Height = Fixed(50)
	growing.style.FlexGrow = 1

	parent.AddChild(fixed, growing)
	Calculate(parent, 200, 200)

	// Fixed child should stay at 30
	if fixed.layout.Rect.Width != 30 {
		t.Errorf("fixed width = %d, want 30", fixed.layout.Rect.Width)
	}

	// Growing child should expand to fill remaining space (100 - 30 = 70)
	if growing.layout.Rect.Width != 70 {
		t.Errorf("growing width = %d, want 70", growing.layout.Rect.Width)
	}
	if growing.layout.Rect.X != 30 {
		t.Errorf("growing.X = %d, want 30", growing.layout.Rect.X)
	}
}

func TestCalculate_FlexGrow_ProportionalDistribution(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Two growing children with different flex values
	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(0)
	child1.style.Height = Fixed(50)
	child1.style.FlexGrow = 1

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(0)
	child2.style.Height = Fixed(50)
	child2.style.FlexGrow = 3

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Child1 should get 1/4 of space (25), child2 should get 3/4 (75)
	if child1.layout.Rect.Width != 25 {
		t.Errorf("child1 width = %d, want 25", child1.layout.Rect.Width)
	}
	if child2.layout.Rect.Width != 75 {
		t.Errorf("child2 width = %d, want 75", child2.layout.Rect.Width)
	}
}

func TestCalculate_FlexShrink(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Two children that are too wide for the container
	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(80)
	child1.style.Height = Fixed(50)
	child1.style.FlexShrink = 1

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(80)
	child2.style.Height = Fixed(50)
	child2.style.FlexShrink = 1

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Total is 160, container is 100, deficit is 60
	// Each should shrink by 30 (equal shrink factors)
	if child1.layout.Rect.Width != 50 {
		t.Errorf("child1 width = %d, want 50", child1.layout.Rect.Width)
	}
	if child2.layout.Rect.Width != 50 {
		t.Errorf("child2 width = %d, want 50", child2.layout.Rect.Width)
	}
}

func TestCalculate_FlexShrink_ProportionalDistribution(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Two children that are too wide for the container
	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(80)
	child1.style.Height = Fixed(50)
	child1.style.FlexShrink = 1 // Will shrink less

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(80)
	child2.style.Height = Fixed(50)
	child2.style.FlexShrink = 3 // Will shrink more

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Total is 160, container is 100, deficit is 60
	// child1 shrinks by 60 * 1/4 = 15 -> 65
	// child2 shrinks by 60 * 3/4 = 45 -> 35
	if child1.layout.Rect.Width != 65 {
		t.Errorf("child1 width = %d, want 65", child1.layout.Rect.Width)
	}
	if child2.layout.Rect.Width != 35 {
		t.Errorf("child2 width = %d, want 35", child2.layout.Rect.Width)
	}
}

func TestCalculate_DirtyTracking(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(50)
	child.style.Height = Fixed(50)

	parent.AddChild(child)

	// First calculation
	Calculate(parent, 200, 200)

	if parent.IsDirty() || child.IsDirty() {
		t.Error("nodes should not be dirty after Calculate")
	}

	// Store original layout
	originalChildRect := child.layout.Rect

	// Calculate again - should be a no-op since nodes are clean
	Calculate(parent, 200, 200)

	if child.layout.Rect != originalChildRect {
		t.Error("clean node layout should not change")
	}

	// Modify child style
	child.SetStyle(child.style) // This marks it dirty

	if !child.IsDirty() {
		t.Error("child should be dirty after SetStyle")
	}
	if !parent.IsDirty() {
		t.Error("parent should be dirty (propagated from child)")
	}
}

func TestCalculate_CleanSubtreeSkipped(t *testing.T) {
	// Create a tree where we can verify that clean subtrees are skipped
	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(200)
	root.style.Height = Fixed(100)

	left := newTestNode(DefaultStyle())
	left.style.Width = Fixed(100)
	left.style.Height = Fixed(100)

	right := newTestNode(DefaultStyle())
	right.style.Width = Fixed(100)
	right.style.Height = Fixed(100)

	root.AddChild(left, right)

	// Initial calculation
	Calculate(root, 300, 200)

	// Clear dirty flags
	root.dirty = false
	left.dirty = false
	right.dirty = false

	// Mark only left subtree dirty
	left.markDirty()

	// Store right's layout
	rightRect := right.layout.Rect

	// Calculate should skip right subtree
	Calculate(root, 300, 200)

	// Right should still have the same layout (wasn't recalculated)
	// Note: This test verifies the dirty flag works, not that layout is literally skipped
	// (since we can't easily measure "not recalculated" without instrumentation)
	if right.layout.Rect != rightRect {
		t.Error("clean right subtree should maintain its layout")
	}
	if right.IsDirty() {
		t.Error("clean right subtree should remain clean")
	}
}

func TestCalculate_NilNode(t *testing.T) {
	// Should not panic
	Calculate(nil, 100, 100)
}

func TestCalculate_EmptyChildren(t *testing.T) {
	node := newTestNode(DefaultStyle())
	node.style.Width = Fixed(100)
	node.style.Height = Fixed(100)

	// Should not panic with no children
	Calculate(node, 200, 200)

	if node.layout.Rect.Width != 100 || node.layout.Rect.Height != 100 {
		t.Errorf("Layout = %dx%d, want 100x100",
			node.layout.Rect.Width, node.layout.Rect.Height)
	}
}

func TestCalculate_NestedContainers(t *testing.T) {
	// Root is a row, child is a column
	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(200)
	root.style.Height = Fixed(100)
	root.style.Direction = Row

	column := newTestNode(DefaultStyle())
	column.style.Width = Fixed(100)
	column.style.Height = Fixed(100)
	column.style.Direction = Column

	grandchild1 := newTestNode(DefaultStyle())
	grandchild1.style.Width = Fixed(100)
	grandchild1.style.Height = Fixed(40)

	grandchild2 := newTestNode(DefaultStyle())
	grandchild2.style.Width = Fixed(100)
	grandchild2.style.Height = Fixed(60)

	column.AddChild(grandchild1, grandchild2)
	root.AddChild(column)

	Calculate(root, 300, 200)

	// Column should be positioned at (0, 0)
	if column.layout.Rect.X != 0 || column.layout.Rect.Y != 0 {
		t.Errorf("column position = (%d, %d), want (0, 0)",
			column.layout.Rect.X, column.layout.Rect.Y)
	}

	// Grandchild1 should be at (0, 0) within the column
	if grandchild1.layout.Rect.X != 0 || grandchild1.layout.Rect.Y != 0 {
		t.Errorf("grandchild1 position = (%d, %d), want (0, 0)",
			grandchild1.layout.Rect.X, grandchild1.layout.Rect.Y)
	}

	// Grandchild2 should be at (0, 40) within the column
	if grandchild2.layout.Rect.X != 0 || grandchild2.layout.Rect.Y != 40 {
		t.Errorf("grandchild2 position = (%d, %d), want (0, 40)",
			grandchild2.layout.Rect.X, grandchild2.layout.Rect.Y)
	}
}

func TestCalculate_AlignStretch(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(80)
	parent.style.Direction = Row
	parent.style.AlignItems = AlignStretch

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(30)
	// Height is Auto - should stretch

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// Child should stretch to fill cross axis (height)
	if child.layout.Rect.Height != 80 {
		t.Errorf("child height = %d, want 80 (stretched)", child.layout.Rect.Height)
	}
}

func TestCalculate_WithMargin(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(50)
	child.style.Height = Fixed(50)
	child.style.Margin = EdgeAll(10)

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// Child border box should be inset by margin
	if child.layout.Rect.X != 10 || child.layout.Rect.Y != 10 {
		t.Errorf("child position = (%d, %d), want (10, 10)",
			child.layout.Rect.X, child.layout.Rect.Y)
	}
	// Child dimensions should account for margin being applied
	if child.layout.Rect.Width != 50 || child.layout.Rect.Height != 50 {
		t.Errorf("child size = %dx%d, want 50x50",
			child.layout.Rect.Width, child.layout.Rect.Height)
	}
}

func TestCalculate_WithGap(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row
	parent.style.Gap = 10

	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(20)
	child1.style.Height = Fixed(50)

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(20)
	child2.style.Height = Fixed(50)

	child3 := newTestNode(DefaultStyle())
	child3.style.Width = Fixed(20)
	child3.style.Height = Fixed(50)

	parent.AddChild(child1, child2, child3)
	Calculate(parent, 200, 200)

	// Children should be spaced with gaps
	if child1.layout.Rect.X != 0 {
		t.Errorf("child1.X = %d, want 0", child1.layout.Rect.X)
	}
	if child2.layout.Rect.X != 30 { // 20 + 10 gap
		t.Errorf("child2.X = %d, want 30", child2.layout.Rect.X)
	}
	if child3.layout.Rect.X != 60 { // 20 + 10 + 20 + 10 gap
		t.Errorf("child3.X = %d, want 60", child3.layout.Rect.X)
	}
}

// Tests for all Justify modes
func TestCalculate_JustifyModes(t *testing.T) {
	type tc struct {
		justify    Justify
		expectedX1 int // First child X position
		expectedX2 int // Second child X position
		expectedX3 int // Third child X position
	}

	// Container: 100 wide, Children: 20 each = 60 total, Free space: 40
	tests := map[string]tc{
		"JustifyStart": {
			justify:    JustifyStart,
			expectedX1: 0,
			expectedX2: 20,
			expectedX3: 40,
		},
		"JustifyEnd": {
			justify:    JustifyEnd,
			expectedX1: 40, // Free space at start
			expectedX2: 60,
			expectedX3: 80,
		},
		"JustifyCenter": {
			justify:    JustifyCenter,
			expectedX1: 20, // Free space / 2 = 20
			expectedX2: 40,
			expectedX3: 60,
		},
		"JustifySpaceBetween": {
			// Free space (40) / (items-1) = 40/2 = 20 between each
			justify:    JustifySpaceBetween,
			expectedX1: 0,
			expectedX2: 40, // 20 + 20
			expectedX3: 80, // 20 + 20 + 20 + 20
		},
		"JustifySpaceAround": {
			// Free space (40) / items = 40/3 ≈ 13.33 around each
			// Start offset: 40 / (3*2) = 6.67 → rounds to 7
			// Between spacing: 40 / 3 = 13.33 → rounds to 13
			justify:    JustifySpaceAround,
			expectedX1: 7,
			expectedX2: 40, // 7 + 20 + 13
			expectedX3: 73, // 40 + 20 + 13
		},
		"JustifySpaceEvenly": {
			// Free space (40) / (items+1) = 40/4 = 10 everywhere
			justify:    JustifySpaceEvenly,
			expectedX1: 10,
			expectedX2: 40, // 10 + 20 + 10
			expectedX3: 70, // 40 + 20 + 10
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := newTestNode(DefaultStyle())
			parent.style.Width = Fixed(100)
			parent.style.Height = Fixed(50)
			parent.style.Direction = Row
			parent.style.JustifyContent = tt.justify

			child1 := newTestNode(DefaultStyle())
			child1.style.Width = Fixed(20)
			child1.style.Height = Fixed(50)

			child2 := newTestNode(DefaultStyle())
			child2.style.Width = Fixed(20)
			child2.style.Height = Fixed(50)

			child3 := newTestNode(DefaultStyle())
			child3.style.Width = Fixed(20)
			child3.style.Height = Fixed(50)

			parent.AddChild(child1, child2, child3)
			Calculate(parent, 200, 200)

			if child1.layout.Rect.X != tt.expectedX1 {
				t.Errorf("child1.X = %d, want %d", child1.layout.Rect.X, tt.expectedX1)
			}
			if child2.layout.Rect.X != tt.expectedX2 {
				t.Errorf("child2.X = %d, want %d", child2.layout.Rect.X, tt.expectedX2)
			}
			if child3.layout.Rect.X != tt.expectedX3 {
				t.Errorf("child3.X = %d, want %d", child3.layout.Rect.X, tt.expectedX3)
			}
		})
	}
}

func TestCalculate_JustifyModes_SingleChild(t *testing.T) {
	type tc struct {
		justify   Justify
		expectedX int
	}

	// Container: 100 wide, Child: 20, Free space: 80
	tests := map[string]tc{
		"JustifyStart":        {justify: JustifyStart, expectedX: 0},
		"JustifyEnd":          {justify: JustifyEnd, expectedX: 80},
		"JustifyCenter":       {justify: JustifyCenter, expectedX: 40},
		"JustifySpaceBetween": {justify: JustifySpaceBetween, expectedX: 0},     // No between with 1 item
		"JustifySpaceAround":  {justify: JustifySpaceAround, expectedX: 40},     // 80 / (1*2) = 40
		"JustifySpaceEvenly":  {justify: JustifySpaceEvenly, expectedX: 40},     // 80 / 2 = 40
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := newTestNode(DefaultStyle())
			parent.style.Width = Fixed(100)
			parent.style.Height = Fixed(50)
			parent.style.Direction = Row
			parent.style.JustifyContent = tt.justify

			child := newTestNode(DefaultStyle())
			child.style.Width = Fixed(20)
			child.style.Height = Fixed(50)

			parent.AddChild(child)
			Calculate(parent, 200, 200)

			if child.layout.Rect.X != tt.expectedX {
				t.Errorf("child.X = %d, want %d", child.layout.Rect.X, tt.expectedX)
			}
		})
	}
}

func TestCalculate_JustifyModes_Column(t *testing.T) {
	// Test justify in Column direction (Y positions)
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(50)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Column
	parent.style.JustifyContent = JustifyEnd

	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(50)
	child1.style.Height = Fixed(20)

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(50)
	child2.style.Height = Fixed(20)

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Free space: 100 - 40 = 60, JustifyEnd pushes children to end
	if child1.layout.Rect.Y != 60 {
		t.Errorf("child1.Y = %d, want 60", child1.layout.Rect.Y)
	}
	if child2.layout.Rect.Y != 80 {
		t.Errorf("child2.Y = %d, want 80", child2.layout.Rect.Y)
	}
}

// Tests for all Align modes
func TestCalculate_AlignModes(t *testing.T) {
	type tc struct {
		align     Align
		childH    int // Child height (explicit, not auto)
		expectedY int
	}

	// Container: 80 high, Children have explicit heights
	tests := map[string]tc{
		"AlignStart": {
			align:     AlignStart,
			childH:    30,
			expectedY: 0,
		},
		"AlignEnd": {
			align:     AlignEnd,
			childH:    30,
			expectedY: 50, // 80 - 30
		},
		"AlignCenter": {
			align:     AlignCenter,
			childH:    30,
			expectedY: 25, // (80 - 30) / 2
		},
		"AlignStretch_explicit": {
			// With explicit height, stretch doesn't change the height
			align:     AlignStretch,
			childH:    30,
			expectedY: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent := newTestNode(DefaultStyle())
			parent.style.Width = Fixed(100)
			parent.style.Height = Fixed(80)
			parent.style.Direction = Row
			parent.style.AlignItems = tt.align

			child := newTestNode(DefaultStyle())
			child.style.Width = Fixed(30)
			child.style.Height = Fixed(tt.childH)

			parent.AddChild(child)
			Calculate(parent, 200, 200)

			if child.layout.Rect.Y != tt.expectedY {
				t.Errorf("child.Y = %d, want %d", child.layout.Rect.Y, tt.expectedY)
			}
			if child.layout.Rect.Height != tt.childH {
				t.Errorf("child.Height = %d, want %d", child.layout.Rect.Height, tt.childH)
			}
		})
	}
}

func TestCalculate_AlignStretch_AutoHeight(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(80)
	parent.style.Direction = Row
	parent.style.AlignItems = AlignStretch

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(30)
	// Height is Auto - should stretch

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	if child.layout.Rect.Y != 0 {
		t.Errorf("child.Y = %d, want 0", child.layout.Rect.Y)
	}
	if child.layout.Rect.Height != 80 {
		t.Errorf("child.Height = %d, want 80 (stretched)", child.layout.Rect.Height)
	}
}

func TestCalculate_AlignModes_Column(t *testing.T) {
	// In Column direction, cross axis is X
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(80)
	parent.style.Direction = Column
	parent.style.AlignItems = AlignEnd

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(30)
	child.style.Height = Fixed(20)

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// AlignEnd in Column means X is at end
	if child.layout.Rect.X != 70 { // 100 - 30
		t.Errorf("child.X = %d, want 70", child.layout.Rect.X)
	}
}

func TestCalculate_AlignSelf_Override(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(80)
	parent.style.Direction = Row
	parent.style.AlignItems = AlignStart

	// First child inherits AlignStart
	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(30)
	child1.style.Height = Fixed(30)

	// Second child overrides with AlignEnd
	alignEnd := AlignEnd
	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(30)
	child2.style.Height = Fixed(30)
	child2.style.AlignSelf = &alignEnd

	// Third child overrides with AlignCenter
	alignCenter := AlignCenter
	child3 := newTestNode(DefaultStyle())
	child3.style.Width = Fixed(30)
	child3.style.Height = Fixed(30)
	child3.style.AlignSelf = &alignCenter

	parent.AddChild(child1, child2, child3)
	Calculate(parent, 200, 200)

	if child1.layout.Rect.Y != 0 {
		t.Errorf("child1.Y = %d, want 0 (AlignStart)", child1.layout.Rect.Y)
	}
	if child2.layout.Rect.Y != 50 { // 80 - 30
		t.Errorf("child2.Y = %d, want 50 (AlignEnd)", child2.layout.Rect.Y)
	}
	if child3.layout.Rect.Y != 25 { // (80 - 30) / 2
		t.Errorf("child3.Y = %d, want 25 (AlignCenter)", child3.layout.Rect.Y)
	}
}

// Tests for Min/Max constraints
func TestCalculate_MinWidth_Constraint(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Child would naturally shrink, but has MinWidth
	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(30)
	child.style.Height = Fixed(50)
	child.style.MinWidth = Fixed(40)

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// MinWidth should enforce minimum
	if child.layout.Rect.Width < 40 {
		t.Errorf("child.Width = %d, want >= 40 (MinWidth)", child.layout.Rect.Width)
	}
}

func TestCalculate_MaxWidth_Constraint(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Growing child with MaxWidth limit
	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(0)
	child.style.Height = Fixed(50)
	child.style.FlexGrow = 1
	child.style.MaxWidth = Fixed(60)

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// MaxWidth should cap the growth
	if child.layout.Rect.Width > 60 {
		t.Errorf("child.Width = %d, want <= 60 (MaxWidth)", child.layout.Rect.Width)
	}
}

func TestCalculate_MinMax_FlexGrow(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Two growing children, one has max constraint
	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(0)
	child1.style.Height = Fixed(50)
	child1.style.FlexGrow = 1
	child1.style.MaxWidth = Fixed(30)

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(0)
	child2.style.Height = Fixed(50)
	child2.style.FlexGrow = 1

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// child1 should be capped at 30
	if child1.layout.Rect.Width != 30 {
		t.Errorf("child1.Width = %d, want 30 (MaxWidth)", child1.layout.Rect.Width)
	}
	// child2 gets remaining space
	if child2.layout.Rect.Width != 50 {
		t.Errorf("child2.Width = %d, want 50", child2.layout.Rect.Width)
	}
}

func TestCalculate_MinMax_FlexShrink(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Two children that need to shrink, one has min constraint
	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(80)
	child1.style.Height = Fixed(50)
	child1.style.FlexShrink = 1
	child1.style.MinWidth = Fixed(60)

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(80)
	child2.style.Height = Fixed(50)
	child2.style.FlexShrink = 1

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// child1 should not shrink below MinWidth
	if child1.layout.Rect.Width < 60 {
		t.Errorf("child1.Width = %d, want >= 60 (MinWidth)", child1.layout.Rect.Width)
	}
}

func TestCalculate_MinHeight_Column(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(50)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Column

	// Child with MinHeight
	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(50)
	child.style.Height = Fixed(20)
	child.style.MinHeight = Fixed(40)

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// MinHeight should enforce minimum
	if child.layout.Rect.Height < 40 {
		t.Errorf("child.Height = %d, want >= 40 (MinHeight)", child.layout.Rect.Height)
	}
}

func TestCalculate_MinMax_MinWins(t *testing.T) {
	// When min > max, min should win (CSS behavior)
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(50)
	child.style.Height = Fixed(50)
	child.style.MinWidth = Fixed(60) // Min > Max
	child.style.MaxWidth = Fixed(40)

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// Min should win
	if child.layout.Rect.Width != 60 {
		t.Errorf("child.Width = %d, want 60 (MinWidth wins over MaxWidth)", child.layout.Rect.Width)
	}
}

// Tests for Percent values
func TestCalculate_PercentWidth(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(200)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row

	child := newTestNode(DefaultStyle())
	child.style.Width = Percent(50)
	child.style.Height = Fixed(100)

	parent.AddChild(child)
	Calculate(parent, 300, 300)

	// 50% of parent's content width (200) = 100
	if child.layout.Rect.Width != 100 {
		t.Errorf("child.Width = %d, want 100 (50%% of 200)", child.layout.Rect.Width)
	}
}

func TestCalculate_PercentHeight(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(200)
	parent.style.Direction = Column

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(100)
	child.style.Height = Percent(25)

	parent.AddChild(child)
	Calculate(parent, 300, 300)

	// 25% of parent's content height (200) = 50
	if child.layout.Rect.Height != 50 {
		t.Errorf("child.Height = %d, want 50 (25%% of 200)", child.layout.Rect.Height)
	}
}

func TestCalculate_NestedPercent(t *testing.T) {
	// Root -> Parent (50%) -> Child (50%)
	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(200)
	root.style.Height = Fixed(100)
	root.style.Direction = Row

	parent := newTestNode(DefaultStyle())
	parent.style.Width = Percent(50) // 100
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row

	child := newTestNode(DefaultStyle())
	child.style.Width = Percent(50) // 50% of 100 = 50
	child.style.Height = Fixed(100)

	parent.AddChild(child)
	root.AddChild(parent)
	Calculate(root, 300, 300)

	// Parent should be 50% of root = 100
	if parent.layout.Rect.Width != 100 {
		t.Errorf("parent.Width = %d, want 100 (50%% of 200)", parent.layout.Rect.Width)
	}

	// Child should be 50% of parent = 50
	if child.layout.Rect.Width != 50 {
		t.Errorf("child.Width = %d, want 50 (50%% of 100)", child.layout.Rect.Width)
	}
}

func TestCalculate_PercentMinMax(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(200)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(0)
	child.style.Height = Fixed(100)
	child.style.FlexGrow = 1
	child.style.MaxWidth = Percent(30) // 30% of 200 = 60

	parent.AddChild(child)
	Calculate(parent, 300, 300)

	// MaxWidth as percent should cap growth
	if child.layout.Rect.Width > 60 {
		t.Errorf("child.Width = %d, want <= 60 (30%% max)", child.layout.Rect.Width)
	}
}

// Incremental layout efficiency tests

// TestCalculate_IncrementalLayout_LeafChange tests that modifying a leaf
// only recalculates the path to the root.
func TestCalculate_IncrementalLayout_LeafChange(t *testing.T) {
	// Build a tree:
	//     root
	//    /    \
	//  left   right
	//  /  \
	// a    b

	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(200)
	root.style.Height = Fixed(100)
	root.style.Direction = Row

	left := newTestNode(DefaultStyle())
	left.style.Width = Fixed(100)
	left.style.Direction = Column

	a := newTestNode(DefaultStyle())
	a.style.Height = Fixed(50)

	b := newTestNode(DefaultStyle())
	b.style.Height = Fixed(50)

	left.AddChild(a, b)

	right := newTestNode(DefaultStyle())
	right.style.Width = Fixed(100)

	root.AddChild(left, right)

	// Initial calculation
	Calculate(root, 200, 100)

	// Verify all nodes are clean
	if root.IsDirty() || left.IsDirty() || right.IsDirty() || a.IsDirty() || b.IsDirty() {
		t.Error("all nodes should be clean after Calculate")
	}

	// Modify leaf node 'a'
	a.SetStyle(a.style)

	// Verify dirty propagation
	if !a.IsDirty() {
		t.Error("node 'a' should be dirty after SetStyle")
	}
	if !left.IsDirty() {
		t.Error("node 'left' should be dirty (ancestor of 'a')")
	}
	if !root.IsDirty() {
		t.Error("root should be dirty (ancestor of 'a')")
	}
	if right.IsDirty() {
		t.Error("node 'right' should NOT be dirty (not an ancestor of 'a')")
	}

	// Store right's layout before recalculation
	rightLayoutBefore := right.layout.Rect

	// Recalculate
	Calculate(root, 200, 100)

	// Verify right was not recalculated (layout unchanged, still clean)
	if right.layout.Rect != rightLayoutBefore {
		t.Error("right's layout should not have changed")
	}
	if right.IsDirty() {
		t.Error("right should still be clean (wasn't recalculated)")
	}
}

// TestCalculate_IncrementalLayout_ReadDoesNotDirty tests that reading
// Layout does not mark the node dirty.
func TestCalculate_IncrementalLayout_ReadDoesNotDirty(t *testing.T) {
	node := newTestNode(DefaultStyle())
	node.style.Width = Fixed(100)
	node.style.Height = Fixed(100)

	Calculate(node, 200, 200)

	if node.IsDirty() {
		t.Error("node should be clean after Calculate")
	}

	// Read layout
	_ = node.layout.Rect
	_ = node.layout.ContentRect
	_ = node.layout.Rect.Width
	_ = node.layout.Rect.Height

	if node.IsDirty() {
		t.Error("reading Layout should NOT mark node dirty")
	}
}

// TestCalculate_IncrementalLayout_MultipleChanges tests multiple changes
// before recalculation.
func TestCalculate_IncrementalLayout_MultipleChanges(t *testing.T) {
	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(100)
	root.style.Height = Fixed(100)
	root.style.Direction = Row

	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(50)

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(50)

	root.AddChild(child1, child2)

	Calculate(root, 100, 100)

	// Make multiple changes
	child1.SetStyle(func() Style {
		s := child1.style
		s.Width = Fixed(30)
		return s
	}())

	child2.SetStyle(func() Style {
		s := child2.style
		s.Width = Fixed(70)
		return s
	}())

	// All should be dirty
	if !root.IsDirty() {
		t.Error("root should be dirty")
	}

	// Single recalculation should handle all changes
	Calculate(root, 100, 100)

	if child1.layout.Rect.Width != 30 {
		t.Errorf("child1.Width = %d, want 30", child1.layout.Rect.Width)
	}
	if child2.layout.Rect.Width != 70 {
		t.Errorf("child2.Width = %d, want 70", child2.layout.Rect.Width)
	}

	// All should be clean
	if root.IsDirty() || child1.IsDirty() || child2.IsDirty() {
		t.Error("all nodes should be clean after Calculate")
	}
}

// TestCalculate_IncrementalLayout_ParentChange tests that changing a parent
// recalculates children.
func TestCalculate_IncrementalLayout_ParentChange(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row

	child := newTestNode(DefaultStyle())
	child.style.FlexGrow = 1

	parent.AddChild(child)

	Calculate(parent, 100, 100)

	// Child should fill parent
	if child.layout.Rect.Width != 100 {
		t.Errorf("child.Width = %d, want 100", child.layout.Rect.Width)
	}

	// Change parent size
	parent.SetStyle(func() Style {
		s := parent.style
		s.Width = Fixed(200)
		return s
	}())

	// Parent dirty, child should be recalculated
	Calculate(parent, 200, 100)

	// Child should now fill new parent size
	if child.layout.Rect.Width != 200 {
		t.Errorf("child.Width = %d, want 200", child.layout.Rect.Width)
	}
}
