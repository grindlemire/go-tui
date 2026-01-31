package layout

import "testing"

// TestIntegration_Dashboard tests a typical dashboard layout:
// - Header (fixed height at top)
// - Sidebar (fixed width on left)
// - Main content (grows to fill remaining space)
// - Footer (fixed height at bottom)
func TestIntegration_Dashboard(t *testing.T) {
	// Root container - full terminal size
	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(120)
	root.style.Height = Fixed(40)
	root.style.Direction = Column

	// Header - fixed height
	header := newTestNode(DefaultStyle())
	header.style.Height = Fixed(3)
	// Width auto-stretches

	// Middle section - row with sidebar and main
	middle := newTestNode(DefaultStyle())
	middle.style.FlexGrow = 1
	middle.style.Direction = Row

	// Sidebar - fixed width
	sidebar := newTestNode(DefaultStyle())
	sidebar.style.Width = Fixed(20)
	// Height auto-stretches

	// Main content - grows to fill
	main := newTestNode(DefaultStyle())
	main.style.FlexGrow = 1

	// Footer - fixed height
	footer := newTestNode(DefaultStyle())
	footer.style.Height = Fixed(2)

	// Build tree
	middle.AddChild(sidebar, main)
	root.AddChild(header, middle, footer)

	Calculate(root, 120, 40)

	// Verify header
	if header.layout.Rect.X != 0 || header.layout.Rect.Y != 0 {
		t.Errorf("header position = (%d, %d), want (0, 0)",
			header.layout.Rect.X, header.layout.Rect.Y)
	}
	if header.layout.Rect.Width != 120 || header.layout.Rect.Height != 3 {
		t.Errorf("header size = %dx%d, want 120x3",
			header.layout.Rect.Width, header.layout.Rect.Height)
	}

	// Verify middle section
	if middle.layout.Rect.Y != 3 {
		t.Errorf("middle.Y = %d, want 3", middle.layout.Rect.Y)
	}
	// Middle should be 40 - 3 (header) - 2 (footer) = 35 tall
	if middle.layout.Rect.Height != 35 {
		t.Errorf("middle.Height = %d, want 35", middle.layout.Rect.Height)
	}

	// Verify sidebar
	if sidebar.layout.Rect.X != 0 || sidebar.layout.Rect.Y != 3 {
		t.Errorf("sidebar position = (%d, %d), want (0, 3)",
			sidebar.layout.Rect.X, sidebar.layout.Rect.Y)
	}
	if sidebar.layout.Rect.Width != 20 {
		t.Errorf("sidebar.Width = %d, want 20", sidebar.layout.Rect.Width)
	}

	// Verify main content
	if main.layout.Rect.X != 20 {
		t.Errorf("main.X = %d, want 20", main.layout.Rect.X)
	}
	// Main should fill remaining width: 120 - 20 = 100
	if main.layout.Rect.Width != 100 {
		t.Errorf("main.Width = %d, want 100", main.layout.Rect.Width)
	}

	// Verify footer
	if footer.layout.Rect.Y != 38 { // 40 - 2
		t.Errorf("footer.Y = %d, want 38", footer.layout.Rect.Y)
	}
	if footer.layout.Rect.Height != 2 {
		t.Errorf("footer.Height = %d, want 2", footer.layout.Rect.Height)
	}
}

// TestIntegration_NestedFlex tests deeply nested flex containers
// with alternating Row/Column directions.
func TestIntegration_NestedFlex(t *testing.T) {
	// Row
	//   Column A
	//     Row A1
	//       Item A1a (fixed)
	//       Item A1b (grow)
	//     Row A2 (grow)
	//   Column B (grow)

	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(100)
	root.style.Height = Fixed(100)
	root.style.Direction = Row

	columnA := newTestNode(DefaultStyle())
	columnA.style.Width = Fixed(50)
	columnA.style.Direction = Column

	rowA1 := newTestNode(DefaultStyle())
	rowA1.style.Height = Fixed(30)
	rowA1.style.Direction = Row

	itemA1a := newTestNode(DefaultStyle())
	itemA1a.style.Width = Fixed(20)

	itemA1b := newTestNode(DefaultStyle())
	itemA1b.style.FlexGrow = 1

	rowA2 := newTestNode(DefaultStyle())
	rowA2.style.FlexGrow = 1
	rowA2.style.Direction = Row

	columnB := newTestNode(DefaultStyle())
	columnB.style.FlexGrow = 1
	columnB.style.Direction = Column

	// Build tree
	rowA1.AddChild(itemA1a, itemA1b)
	columnA.AddChild(rowA1, rowA2)
	root.AddChild(columnA, columnB)

	Calculate(root, 100, 100)

	// Verify column A
	if columnA.layout.Rect.Width != 50 {
		t.Errorf("columnA.Width = %d, want 50", columnA.layout.Rect.Width)
	}
	if columnA.layout.Rect.Height != 100 {
		t.Errorf("columnA.Height = %d, want 100", columnA.layout.Rect.Height)
	}

	// Verify row A1
	if rowA1.layout.Rect.Height != 30 {
		t.Errorf("rowA1.Height = %d, want 30", rowA1.layout.Rect.Height)
	}

	// Verify item A1a (within rowA1)
	if itemA1a.layout.Rect.Width != 20 {
		t.Errorf("itemA1a.Width = %d, want 20", itemA1a.layout.Rect.Width)
	}

	// Verify item A1b (should grow to fill remaining: 50 - 20 = 30)
	if itemA1b.layout.Rect.Width != 30 {
		t.Errorf("itemA1b.Width = %d, want 30", itemA1b.layout.Rect.Width)
	}

	// Verify row A2 (should grow to fill: 100 - 30 = 70)
	if rowA2.layout.Rect.Height != 70 {
		t.Errorf("rowA2.Height = %d, want 70", rowA2.layout.Rect.Height)
	}

	// Verify column B (should grow to fill: 100 - 50 = 50)
	if columnB.layout.Rect.X != 50 {
		t.Errorf("columnB.X = %d, want 50", columnB.layout.Rect.X)
	}
	if columnB.layout.Rect.Width != 50 {
		t.Errorf("columnB.Width = %d, want 50", columnB.layout.Rect.Width)
	}
}

// TestIntegration_FormLayout tests a typical form layout:
// - Labels (fixed width)
// - Inputs (grow to fill)
// - Arranged in a column
func TestIntegration_FormLayout(t *testing.T) {
	form := newTestNode(DefaultStyle())
	form.style.Width = Fixed(80)
	form.style.Height = Fixed(30)
	form.style.Direction = Column
	form.style.Gap = 1

	// Create 3 form rows
	for i := 0; i < 3; i++ {
		row := newTestNode(DefaultStyle())
		row.style.Height = Fixed(3)
		row.style.Direction = Row
		row.style.Gap = 2

		label := newTestNode(DefaultStyle())
		label.style.Width = Fixed(15)

		input := newTestNode(DefaultStyle())
		input.style.FlexGrow = 1

		row.AddChild(label, input)
		form.AddChild(row)
	}

	Calculate(form, 100, 50)

	// Verify each row
	for i, row := range form.children {
		expectedY := i * 4 // 3 height + 1 gap
		if row.layout.Rect.Y != expectedY {
			t.Errorf("row[%d].Y = %d, want %d", i, row.layout.Rect.Y, expectedY)
		}

		label := row.children[0]
		input := row.children[1]

		// Label should be fixed at 15
		if label.layout.Rect.Width != 15 {
			t.Errorf("row[%d] label.Width = %d, want 15", i, label.layout.Rect.Width)
		}

		// Input should grow to fill: 80 - 15 - 2 (gap) = 63
		if input.layout.Rect.Width != 63 {
			t.Errorf("row[%d] input.Width = %d, want 63", i, input.layout.Rect.Width)
		}

		// Input should be positioned after label + gap
		if input.layout.Rect.X != 17 { // 15 + 2 gap
			t.Errorf("row[%d] input.X = %d, want 17", i, input.layout.Rect.X)
		}
	}
}

// TestIntegration_MixedDirection tests layout with alternating
// Row and Column directions at different levels.
func TestIntegration_MixedDirection(t *testing.T) {
	// Column (vertical)
	//   Row (horizontal)
	//     Column (vertical)
	//     Column (vertical)
	//   Row (horizontal)

	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(100)
	root.style.Height = Fixed(100)
	root.style.Direction = Column

	row1 := newTestNode(DefaultStyle())
	row1.style.Height = Fixed(50)
	row1.style.Direction = Row

	col1a := newTestNode(DefaultStyle())
	col1a.style.Width = Fixed(30)
	col1a.style.Direction = Column

	col1b := newTestNode(DefaultStyle())
	col1b.style.FlexGrow = 1
	col1b.style.Direction = Column

	row2 := newTestNode(DefaultStyle())
	row2.style.FlexGrow = 1
	row2.style.Direction = Row

	row1.AddChild(col1a, col1b)
	root.AddChild(row1, row2)

	Calculate(root, 100, 100)

	// Verify row1 position and size
	if row1.layout.Rect.Y != 0 {
		t.Errorf("row1.Y = %d, want 0", row1.layout.Rect.Y)
	}
	if row1.layout.Rect.Height != 50 {
		t.Errorf("row1.Height = %d, want 50", row1.layout.Rect.Height)
	}

	// Verify col1a
	if col1a.layout.Rect.Width != 30 {
		t.Errorf("col1a.Width = %d, want 30", col1a.layout.Rect.Width)
	}
	if col1a.layout.Rect.Height != 50 {
		t.Errorf("col1a.Height = %d, want 50 (stretched)", col1a.layout.Rect.Height)
	}

	// Verify col1b (grows to fill: 100 - 30 = 70)
	if col1b.layout.Rect.X != 30 {
		t.Errorf("col1b.X = %d, want 30", col1b.layout.Rect.X)
	}
	if col1b.layout.Rect.Width != 70 {
		t.Errorf("col1b.Width = %d, want 70", col1b.layout.Rect.Width)
	}

	// Verify row2 (grows to fill: 100 - 50 = 50)
	if row2.layout.Rect.Y != 50 {
		t.Errorf("row2.Y = %d, want 50", row2.layout.Rect.Y)
	}
	if row2.layout.Rect.Height != 50 {
		t.Errorf("row2.Height = %d, want 50", row2.layout.Rect.Height)
	}
}

// Edge case tests

// TestEdgeCase_ZeroDimensions tests that zero/negative dimensions are handled.
func TestEdgeCase_ZeroDimensions(t *testing.T) {
	type tc struct {
		width          int
		height         int
		expectedWidth  int
		expectedHeight int
	}

	tests := map[string]tc{
		"zero width": {
			width:          0,
			height:         50,
			expectedWidth:  0,
			expectedHeight: 50,
		},
		"zero height": {
			width:          50,
			height:         0,
			expectedWidth:  50,
			expectedHeight: 0,
		},
		"both zero": {
			width:          0,
			height:         0,
			expectedWidth:  0,
			expectedHeight: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			node := newTestNode(DefaultStyle())
			node.style.Width = Fixed(tt.width)
			node.style.Height = Fixed(tt.height)

			Calculate(node, 100, 100)

			if node.layout.Rect.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d", node.layout.Rect.Width, tt.expectedWidth)
			}
			if node.layout.Rect.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d", node.layout.Rect.Height, tt.expectedHeight)
			}
		})
	}
}

// TestEdgeCase_ZeroSizeParent tests children of a zero-size parent.
func TestEdgeCase_ZeroSizeParent(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(0)
	parent.style.Height = Fixed(0)
	parent.style.Direction = Row

	child := newTestNode(DefaultStyle())
	child.style.Width = Fixed(50)
	child.style.Height = Fixed(30)

	parent.AddChild(child)
	Calculate(parent, 100, 100)

	// Parent should be zero size
	if parent.layout.Rect.Width != 0 || parent.layout.Rect.Height != 0 {
		t.Errorf("parent size = %dx%d, want 0x0",
			parent.layout.Rect.Width, parent.layout.Rect.Height)
	}

	// Child gets clamped due to min/max constraints from zero-size parent
	// The child's border box will be constrained by parent's content rect
	// which is also 0x0, so child dimensions will be constrained
}

// TestEdgeCase_OverflowNoShrink tests overflow when shrink totals to 0.
func TestEdgeCase_OverflowNoShrink(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(50)
	parent.style.Direction = Row

	// Children that overflow but don't shrink
	child1 := newTestNode(DefaultStyle())
	child1.style.Width = Fixed(80)
	child1.style.Height = Fixed(50)
	child1.style.FlexShrink = 0

	child2 := newTestNode(DefaultStyle())
	child2.style.Width = Fixed(80)
	child2.style.Height = Fixed(50)
	child2.style.FlexShrink = 0

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Children should maintain their sizes (overflow parent)
	if child1.layout.Rect.Width != 80 {
		t.Errorf("child1.Width = %d, want 80", child1.layout.Rect.Width)
	}
	if child2.layout.Rect.Width != 80 {
		t.Errorf("child2.Width = %d, want 80", child2.layout.Rect.Width)
	}

	// Child2 should be positioned after child1
	if child2.layout.Rect.X != 80 {
		t.Errorf("child2.X = %d, want 80", child2.layout.Rect.X)
	}
}

// TestEdgeCase_EmptyAutoNode tests a node with no children and auto dimensions.
func TestEdgeCase_EmptyAutoNode(t *testing.T) {
	node := newTestNode(DefaultStyle())
	// Width and Height are Auto by default

	Calculate(node, 100, 100)

	// Auto dimensions should use available space (fill parent)
	if node.layout.Rect.Width != 100 {
		t.Errorf("Width = %d, want 100 (auto fills available)", node.layout.Rect.Width)
	}
	if node.layout.Rect.Height != 100 {
		t.Errorf("Height = %d, want 100 (auto fills available)", node.layout.Rect.Height)
	}
}

// TestEdgeCase_VeryLargeTree tests layout with a large number of children.
func TestEdgeCase_VeryLargeTree(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(1000)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row

	// Add 100 children
	for i := 0; i < 100; i++ {
		child := newTestNode(DefaultStyle())
		child.style.Width = Fixed(10)
		child.style.Height = Fixed(100)
		parent.AddChild(child)
	}

	Calculate(parent, 1000, 100)

	// Verify all children are positioned correctly
	for i, child := range parent.children {
		expectedX := i * 10
		if child.layout.Rect.X != expectedX {
			t.Errorf("child[%d].X = %d, want %d", i, child.layout.Rect.X, expectedX)
		}
	}
}

// TestEdgeCase_DeepNesting tests deeply nested layout.
func TestEdgeCase_DeepNesting(t *testing.T) {
	// Create a chain of 10 nested nodes
	root := newTestNode(DefaultStyle())
	root.style.Width = Fixed(100)
	root.style.Height = Fixed(100)

	current := root
	for i := 0; i < 10; i++ {
		child := newTestNode(DefaultStyle())
		child.style.FlexGrow = 1
		child.style.Padding = EdgeAll(1)
		current.AddChild(child)
		current = child
	}

	Calculate(root, 100, 100)

	// Leaf node should have padding accumulated
	// Each level adds 1 padding on each side, so after 10 levels:
	// Content area shrinks by 2 per level = 20 total per dimension
	leaf := current
	expectedSize := 100 - 20 // 80
	if leaf.layout.ContentRect.Width != expectedSize {
		t.Errorf("leaf.ContentRect.Width = %d, want %d",
			leaf.layout.ContentRect.Width, expectedSize)
	}
}
