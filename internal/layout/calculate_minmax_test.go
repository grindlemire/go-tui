package layout

import "testing"

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
