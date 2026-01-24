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
			node := NewNode(tt.style)
			Calculate(node, tt.availableW, tt.availableH)

			if node.Layout.Rect.Width != tt.expectedWidth {
				t.Errorf("Layout.Rect.Width = %d, want %d", node.Layout.Rect.Width, tt.expectedWidth)
			}
			if node.Layout.Rect.Height != tt.expectedHeight {
				t.Errorf("Layout.Rect.Height = %d, want %d", node.Layout.Rect.Height, tt.expectedHeight)
			}
			if node.Layout.Rect.X != 0 || node.Layout.Rect.Y != 0 {
				t.Errorf("Layout.Rect position = (%d, %d), want (0, 0)",
					node.Layout.Rect.X, node.Layout.Rect.Y)
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

	node := NewNode(style)
	Calculate(node, 200, 200)

	// Border box should be the full size
	if node.Layout.Rect.Width != 100 || node.Layout.Rect.Height != 80 {
		t.Errorf("Layout.Rect = %dx%d, want 100x80",
			node.Layout.Rect.Width, node.Layout.Rect.Height)
	}

	// Content rect should be inset by padding
	if node.Layout.ContentRect.X != 10 || node.Layout.ContentRect.Y != 10 {
		t.Errorf("ContentRect position = (%d, %d), want (10, 10)",
			node.Layout.ContentRect.X, node.Layout.ContentRect.Y)
	}
	if node.Layout.ContentRect.Width != 80 || node.Layout.ContentRect.Height != 60 {
		t.Errorf("ContentRect size = %dx%d, want 80x60",
			node.Layout.ContentRect.Width, node.Layout.ContentRect.Height)
	}
}

func TestCalculate_TwoChildren_Row(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(50)
	parent.Style.Direction = Row

	child1 := NewNode(DefaultStyle())
	child1.Style.Width = Fixed(30)
	child1.Style.Height = Fixed(50)

	child2 := NewNode(DefaultStyle())
	child2.Style.Width = Fixed(40)
	child2.Style.Height = Fixed(50)

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Child 1 should be at position 0
	if child1.Layout.Rect.X != 0 || child1.Layout.Rect.Y != 0 {
		t.Errorf("child1 position = (%d, %d), want (0, 0)",
			child1.Layout.Rect.X, child1.Layout.Rect.Y)
	}
	if child1.Layout.Rect.Width != 30 {
		t.Errorf("child1 width = %d, want 30", child1.Layout.Rect.Width)
	}

	// Child 2 should be at position 30 (after child1)
	if child2.Layout.Rect.X != 30 {
		t.Errorf("child2.X = %d, want 30", child2.Layout.Rect.X)
	}
	if child2.Layout.Rect.Width != 40 {
		t.Errorf("child2 width = %d, want 40", child2.Layout.Rect.Width)
	}
}

func TestCalculate_TwoChildren_Column(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(100)
	parent.Style.Direction = Column

	child1 := NewNode(DefaultStyle())
	child1.Style.Width = Fixed(100)
	child1.Style.Height = Fixed(30)

	child2 := NewNode(DefaultStyle())
	child2.Style.Width = Fixed(100)
	child2.Style.Height = Fixed(40)

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Child 1 should be at position 0
	if child1.Layout.Rect.X != 0 || child1.Layout.Rect.Y != 0 {
		t.Errorf("child1 position = (%d, %d), want (0, 0)",
			child1.Layout.Rect.X, child1.Layout.Rect.Y)
	}
	if child1.Layout.Rect.Height != 30 {
		t.Errorf("child1 height = %d, want 30", child1.Layout.Rect.Height)
	}

	// Child 2 should be at Y position 30 (after child1)
	if child2.Layout.Rect.Y != 30 {
		t.Errorf("child2.Y = %d, want 30", child2.Layout.Rect.Y)
	}
	if child2.Layout.Rect.Height != 40 {
		t.Errorf("child2 height = %d, want 40", child2.Layout.Rect.Height)
	}
}

func TestCalculate_FlexGrow(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(50)
	parent.Style.Direction = Row

	// Fixed child
	fixed := NewNode(DefaultStyle())
	fixed.Style.Width = Fixed(30)
	fixed.Style.Height = Fixed(50)

	// Growing child
	growing := NewNode(DefaultStyle())
	growing.Style.Width = Fixed(0) // Start at 0
	growing.Style.Height = Fixed(50)
	growing.Style.FlexGrow = 1

	parent.AddChild(fixed, growing)
	Calculate(parent, 200, 200)

	// Fixed child should stay at 30
	if fixed.Layout.Rect.Width != 30 {
		t.Errorf("fixed width = %d, want 30", fixed.Layout.Rect.Width)
	}

	// Growing child should expand to fill remaining space (100 - 30 = 70)
	if growing.Layout.Rect.Width != 70 {
		t.Errorf("growing width = %d, want 70", growing.Layout.Rect.Width)
	}
	if growing.Layout.Rect.X != 30 {
		t.Errorf("growing.X = %d, want 30", growing.Layout.Rect.X)
	}
}

func TestCalculate_FlexGrow_ProportionalDistribution(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(50)
	parent.Style.Direction = Row

	// Two growing children with different flex values
	child1 := NewNode(DefaultStyle())
	child1.Style.Width = Fixed(0)
	child1.Style.Height = Fixed(50)
	child1.Style.FlexGrow = 1

	child2 := NewNode(DefaultStyle())
	child2.Style.Width = Fixed(0)
	child2.Style.Height = Fixed(50)
	child2.Style.FlexGrow = 3

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Child1 should get 1/4 of space (25), child2 should get 3/4 (75)
	if child1.Layout.Rect.Width != 25 {
		t.Errorf("child1 width = %d, want 25", child1.Layout.Rect.Width)
	}
	if child2.Layout.Rect.Width != 75 {
		t.Errorf("child2 width = %d, want 75", child2.Layout.Rect.Width)
	}
}

func TestCalculate_FlexShrink(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(50)
	parent.Style.Direction = Row

	// Two children that are too wide for the container
	child1 := NewNode(DefaultStyle())
	child1.Style.Width = Fixed(80)
	child1.Style.Height = Fixed(50)
	child1.Style.FlexShrink = 1

	child2 := NewNode(DefaultStyle())
	child2.Style.Width = Fixed(80)
	child2.Style.Height = Fixed(50)
	child2.Style.FlexShrink = 1

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Total is 160, container is 100, deficit is 60
	// Each should shrink by 30 (equal shrink factors)
	if child1.Layout.Rect.Width != 50 {
		t.Errorf("child1 width = %d, want 50", child1.Layout.Rect.Width)
	}
	if child2.Layout.Rect.Width != 50 {
		t.Errorf("child2 width = %d, want 50", child2.Layout.Rect.Width)
	}
}

func TestCalculate_FlexShrink_ProportionalDistribution(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(50)
	parent.Style.Direction = Row

	// Two children that are too wide for the container
	child1 := NewNode(DefaultStyle())
	child1.Style.Width = Fixed(80)
	child1.Style.Height = Fixed(50)
	child1.Style.FlexShrink = 1 // Will shrink less

	child2 := NewNode(DefaultStyle())
	child2.Style.Width = Fixed(80)
	child2.Style.Height = Fixed(50)
	child2.Style.FlexShrink = 3 // Will shrink more

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Total is 160, container is 100, deficit is 60
	// child1 shrinks by 60 * 1/4 = 15 -> 65
	// child2 shrinks by 60 * 3/4 = 45 -> 35
	if child1.Layout.Rect.Width != 65 {
		t.Errorf("child1 width = %d, want 65", child1.Layout.Rect.Width)
	}
	if child2.Layout.Rect.Width != 35 {
		t.Errorf("child2 width = %d, want 35", child2.Layout.Rect.Width)
	}
}

func TestCalculate_DirtyTracking(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(100)

	child := NewNode(DefaultStyle())
	child.Style.Width = Fixed(50)
	child.Style.Height = Fixed(50)

	parent.AddChild(child)

	// First calculation
	Calculate(parent, 200, 200)

	if parent.IsDirty() || child.IsDirty() {
		t.Error("nodes should not be dirty after Calculate")
	}

	// Store original layout
	originalChildRect := child.Layout.Rect

	// Calculate again - should be a no-op since nodes are clean
	Calculate(parent, 200, 200)

	if child.Layout.Rect != originalChildRect {
		t.Error("clean node layout should not change")
	}

	// Modify child style
	child.SetStyle(child.Style) // This marks it dirty

	if !child.IsDirty() {
		t.Error("child should be dirty after SetStyle")
	}
	if !parent.IsDirty() {
		t.Error("parent should be dirty (propagated from child)")
	}
}

func TestCalculate_CleanSubtreeSkipped(t *testing.T) {
	// Create a tree where we can verify that clean subtrees are skipped
	root := NewNode(DefaultStyle())
	root.Style.Width = Fixed(200)
	root.Style.Height = Fixed(100)

	left := NewNode(DefaultStyle())
	left.Style.Width = Fixed(100)
	left.Style.Height = Fixed(100)

	right := NewNode(DefaultStyle())
	right.Style.Width = Fixed(100)
	right.Style.Height = Fixed(100)

	root.AddChild(left, right)

	// Initial calculation
	Calculate(root, 300, 200)

	// Clear dirty flags
	root.dirty = false
	left.dirty = false
	right.dirty = false

	// Mark only left subtree dirty
	left.MarkDirty()

	// Store right's layout
	rightRect := right.Layout.Rect

	// Calculate should skip right subtree
	Calculate(root, 300, 200)

	// Right should still have the same layout (wasn't recalculated)
	// Note: This test verifies the dirty flag works, not that layout is literally skipped
	// (since we can't easily measure "not recalculated" without instrumentation)
	if right.Layout.Rect != rightRect {
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
	node := NewNode(DefaultStyle())
	node.Style.Width = Fixed(100)
	node.Style.Height = Fixed(100)

	// Should not panic with no children
	Calculate(node, 200, 200)

	if node.Layout.Rect.Width != 100 || node.Layout.Rect.Height != 100 {
		t.Errorf("Layout = %dx%d, want 100x100",
			node.Layout.Rect.Width, node.Layout.Rect.Height)
	}
}

func TestCalculate_NestedContainers(t *testing.T) {
	// Root is a row, child is a column
	root := NewNode(DefaultStyle())
	root.Style.Width = Fixed(200)
	root.Style.Height = Fixed(100)
	root.Style.Direction = Row

	column := NewNode(DefaultStyle())
	column.Style.Width = Fixed(100)
	column.Style.Height = Fixed(100)
	column.Style.Direction = Column

	grandchild1 := NewNode(DefaultStyle())
	grandchild1.Style.Width = Fixed(100)
	grandchild1.Style.Height = Fixed(40)

	grandchild2 := NewNode(DefaultStyle())
	grandchild2.Style.Width = Fixed(100)
	grandchild2.Style.Height = Fixed(60)

	column.AddChild(grandchild1, grandchild2)
	root.AddChild(column)

	Calculate(root, 300, 200)

	// Column should be positioned at (0, 0)
	if column.Layout.Rect.X != 0 || column.Layout.Rect.Y != 0 {
		t.Errorf("column position = (%d, %d), want (0, 0)",
			column.Layout.Rect.X, column.Layout.Rect.Y)
	}

	// Grandchild1 should be at (0, 0) within the column
	if grandchild1.Layout.Rect.X != 0 || grandchild1.Layout.Rect.Y != 0 {
		t.Errorf("grandchild1 position = (%d, %d), want (0, 0)",
			grandchild1.Layout.Rect.X, grandchild1.Layout.Rect.Y)
	}

	// Grandchild2 should be at (0, 40) within the column
	if grandchild2.Layout.Rect.X != 0 || grandchild2.Layout.Rect.Y != 40 {
		t.Errorf("grandchild2 position = (%d, %d), want (0, 40)",
			grandchild2.Layout.Rect.X, grandchild2.Layout.Rect.Y)
	}
}

func TestCalculate_AlignStretch(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(80)
	parent.Style.Direction = Row
	parent.Style.AlignItems = AlignStretch

	child := NewNode(DefaultStyle())
	child.Style.Width = Fixed(30)
	// Height is Auto - should stretch

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// Child should stretch to fill cross axis (height)
	if child.Layout.Rect.Height != 80 {
		t.Errorf("child height = %d, want 80 (stretched)", child.Layout.Rect.Height)
	}
}

func TestCalculate_WithMargin(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(100)
	parent.Style.Direction = Row

	child := NewNode(DefaultStyle())
	child.Style.Width = Fixed(50)
	child.Style.Height = Fixed(50)
	child.Style.Margin = EdgeAll(10)

	parent.AddChild(child)
	Calculate(parent, 200, 200)

	// Child border box should be inset by margin
	if child.Layout.Rect.X != 10 || child.Layout.Rect.Y != 10 {
		t.Errorf("child position = (%d, %d), want (10, 10)",
			child.Layout.Rect.X, child.Layout.Rect.Y)
	}
	// Child dimensions should account for margin being applied
	if child.Layout.Rect.Width != 50 || child.Layout.Rect.Height != 50 {
		t.Errorf("child size = %dx%d, want 50x50",
			child.Layout.Rect.Width, child.Layout.Rect.Height)
	}
}

func TestCalculate_WithGap(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(50)
	parent.Style.Direction = Row
	parent.Style.Gap = 10

	child1 := NewNode(DefaultStyle())
	child1.Style.Width = Fixed(20)
	child1.Style.Height = Fixed(50)

	child2 := NewNode(DefaultStyle())
	child2.Style.Width = Fixed(20)
	child2.Style.Height = Fixed(50)

	child3 := NewNode(DefaultStyle())
	child3.Style.Width = Fixed(20)
	child3.Style.Height = Fixed(50)

	parent.AddChild(child1, child2, child3)
	Calculate(parent, 200, 200)

	// Children should be spaced with gaps
	if child1.Layout.Rect.X != 0 {
		t.Errorf("child1.X = %d, want 0", child1.Layout.Rect.X)
	}
	if child2.Layout.Rect.X != 30 { // 20 + 10 gap
		t.Errorf("child2.X = %d, want 30", child2.Layout.Rect.X)
	}
	if child3.Layout.Rect.X != 60 { // 20 + 10 + 20 + 10 gap
		t.Errorf("child3.X = %d, want 60", child3.Layout.Rect.X)
	}
}
