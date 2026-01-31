package layout

import "testing"

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

// --- Intrinsic Sizing Tests ---
// These tests verify that Auto-sized elements with intrinsic content
// (like text) get proper sizes instead of collapsing to 0.

func TestCalculate_IntrinsicSize_AutoChildUsesIntrinsic(t *testing.T) {
	// This test would have caught the original DSL counter bug:
	// An Auto-sized child with intrinsic content should use that as base size,
	// not collapse to 0.

	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Column

	// Child has Auto size but intrinsic content (like a text element)
	child := newTestNode(DefaultStyle())
	// Auto width and height (default)
	child.SetIntrinsicSize(40, 10) // Simulates text "hello world" or similar

	parent.AddChild(child)
	Calculate(parent, 100, 100)

	// Child should have its intrinsic dimensions as base size
	// With AlignStretch (default), it stretches on cross axis but uses intrinsic on main
	if child.layout.Rect.Height != 10 {
		t.Errorf("child.Height = %d, want 10 (intrinsic)", child.layout.Rect.Height)
	}
	// Width stretches to fill parent (AlignStretch is default)
	if child.layout.Rect.Width != 100 {
		t.Errorf("child.Width = %d, want 100 (stretched)", child.layout.Rect.Width)
	}
}

func TestCalculate_IntrinsicSize_RowWithIntrinsicChildren(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(200)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Two Auto-sized children with intrinsic sizes
	child1 := newTestNode(DefaultStyle())
	child1.SetIntrinsicSize(30, 20)

	child2 := newTestNode(DefaultStyle())
	child2.SetIntrinsicSize(50, 15)

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 50)

	// Children should have their intrinsic widths
	if child1.layout.Rect.Width != 30 {
		t.Errorf("child1.Width = %d, want 30 (intrinsic)", child1.layout.Rect.Width)
	}
	if child2.layout.Rect.Width != 50 {
		t.Errorf("child2.Width = %d, want 50 (intrinsic)", child2.layout.Rect.Width)
	}

	// Positions should be sequential
	if child1.layout.Rect.X != 0 {
		t.Errorf("child1.X = %d, want 0", child1.layout.Rect.X)
	}
	if child2.layout.Rect.X != 30 {
		t.Errorf("child2.X = %d, want 30", child2.layout.Rect.X)
	}
}

func TestCalculate_IntrinsicSize_ColumnWithIntrinsicChildren(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(200)
	parent.style.Direction = Column

	// Three Auto-sized children with intrinsic sizes (simulating text elements)
	child1 := newTestNode(DefaultStyle())
	child1.SetIntrinsicSize(80, 10)

	child2 := newTestNode(DefaultStyle())
	child2.SetIntrinsicSize(60, 15)

	child3 := newTestNode(DefaultStyle())
	child3.SetIntrinsicSize(70, 20)

	parent.AddChild(child1, child2, child3)
	Calculate(parent, 100, 200)

	// Children should have their intrinsic heights
	if child1.layout.Rect.Height != 10 {
		t.Errorf("child1.Height = %d, want 10 (intrinsic)", child1.layout.Rect.Height)
	}
	if child2.layout.Rect.Height != 15 {
		t.Errorf("child2.Height = %d, want 15 (intrinsic)", child2.layout.Rect.Height)
	}
	if child3.layout.Rect.Height != 20 {
		t.Errorf("child3.Height = %d, want 20 (intrinsic)", child3.layout.Rect.Height)
	}

	// Positions should be sequential
	if child1.layout.Rect.Y != 0 {
		t.Errorf("child1.Y = %d, want 0", child1.layout.Rect.Y)
	}
	if child2.layout.Rect.Y != 10 {
		t.Errorf("child2.Y = %d, want 10", child2.layout.Rect.Y)
	}
	if child3.layout.Rect.Y != 25 {
		t.Errorf("child3.Y = %d, want 25", child3.layout.Rect.Y)
	}
}

func TestCalculate_IntrinsicSize_JustifyCenterWithIntrinsic(t *testing.T) {
	// This tests centering Auto-sized children based on their intrinsic size
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Column
	parent.style.JustifyContent = JustifyCenter

	child := newTestNode(DefaultStyle())
	child.SetIntrinsicSize(60, 20)

	parent.AddChild(child)
	Calculate(parent, 100, 100)

	// Child should be centered vertically based on intrinsic height
	// Remaining space = 100 - 20 = 80, centered offset = 40
	if child.layout.Rect.Y != 40 {
		t.Errorf("child.Y = %d, want 40 (centered)", child.layout.Rect.Y)
	}
	if child.layout.Rect.Height != 20 {
		t.Errorf("child.Height = %d, want 20 (intrinsic)", child.layout.Rect.Height)
	}
}

func TestCalculate_IntrinsicSize_AlignCenterWithIntrinsic(t *testing.T) {
	// This tests cross-axis centering of Auto-sized children
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Column
	parent.style.AlignItems = AlignCenter

	child := newTestNode(DefaultStyle())
	child.SetIntrinsicSize(40, 20) // 40 wide, 20 tall

	parent.AddChild(child)
	Calculate(parent, 100, 100)

	// Child should be centered horizontally based on intrinsic width
	// Remaining width = 100 - 40 = 60, centered offset = 30
	if child.layout.Rect.X != 30 {
		t.Errorf("child.X = %d, want 30 (centered)", child.layout.Rect.X)
	}
	if child.layout.Rect.Width != 40 {
		t.Errorf("child.Width = %d, want 40 (intrinsic)", child.layout.Rect.Width)
	}
}

func TestCalculate_IntrinsicSize_NestedContainers(t *testing.T) {
	// This simulates the DSL counter structure:
	// root (fixed size, Column, JustifyCenter, AlignCenter)
	//   └── container (Auto size, Column, AlignStart)
	//         ├── box (Auto size, AlignStart)
	//         │     └── text (intrinsic: 15x1)
	//         └── text (intrinsic: 30x1)

	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(80)
	root.style.Height = Fixed(24)
	root.style.Direction = Column
	root.style.JustifyContent = JustifyCenter
	root.style.AlignItems = AlignCenter

	container := newTestNode(DefaultStyle())
	container.style.Direction = Column
	container.style.AlignItems = AlignStart // Prevent children from stretching
	// Auto size - should get intrinsic size from children

	box := newTestNode(DefaultStyle())
	box.style.Direction = Column
	box.style.AlignItems = AlignStart // Prevent text1 from stretching
	box.style.Padding = EdgeAll(1)

	text1 := newTestNode(DefaultStyle())
	text1.SetIntrinsicSize(15, 1) // "Counter Example"

	text2 := newTestNode(DefaultStyle())
	text2.SetIntrinsicSize(30, 1) // "Press +/- to change, q to quit"

	box.AddChild(text1)
	container.AddChild(box)
	container.AddChild(text2)
	root.AddChild(container)

	Calculate(root, 80, 24)

	// text1 should have its intrinsic size
	if text1.layout.Rect.Width != 15 || text1.layout.Rect.Height != 1 {
		t.Errorf("text1 size = %dx%d, want 15x1",
			text1.layout.Rect.Width, text1.layout.Rect.Height)
	}

	// box should wrap text1 plus padding (1 on each side)
	// width = 15 + 2 = 17, height = 1 + 2 = 3
	if box.layout.Rect.Width != 17 {
		t.Errorf("box.Width = %d, want 17 (text + padding)", box.layout.Rect.Width)
	}
	if box.layout.Rect.Height != 3 {
		t.Errorf("box.Height = %d, want 3 (text + padding)", box.layout.Rect.Height)
	}

	// container should have height = box height + text2 height = 3 + 1 = 4
	// container width = max(box, text2) = max(17, 30) = 30
	if container.layout.Rect.Height != 4 {
		t.Errorf("container.Height = %d, want 4 (box + text2)", container.layout.Rect.Height)
	}
	if container.layout.Rect.Width != 30 {
		t.Errorf("container.Width = %d, want 30 (max child width)", container.layout.Rect.Width)
	}

	// Container should be centered in root
	// Vertical: (24 - 4) / 2 = 10
	// Horizontal: (80 - 30) / 2 = 25
	if container.layout.Rect.Y != 10 {
		t.Errorf("container.Y = %d, want 10 (centered)", container.layout.Rect.Y)
	}
	if container.layout.Rect.X != 25 {
		t.Errorf("container.X = %d, want 25 (centered)", container.layout.Rect.X)
	}
}

func TestCalculate_IntrinsicSize_ZeroIntrinsicStillWorks(t *testing.T) {
	// Elements with no intrinsic size (like empty containers) should still work
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row

	child := newTestNode(DefaultStyle())
	// No intrinsic size set, no explicit size - relies on FlexGrow
	child.style.FlexGrow = 1

	parent.AddChild(child)
	Calculate(parent, 100, 100)

	// Child should fill available space via FlexGrow
	if child.layout.Rect.Width != 100 {
		t.Errorf("child.Width = %d, want 100 (filled via grow)", child.layout.Rect.Width)
	}
}
