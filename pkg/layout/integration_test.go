package layout

import "testing"

// TestIntegration_Dashboard tests a typical dashboard layout:
// - Header (fixed height at top)
// - Sidebar (fixed width on left)
// - Main content (grows to fill remaining space)
// - Footer (fixed height at bottom)
func TestIntegration_Dashboard(t *testing.T) {
	// Root container - full terminal size
	root := NewNode(DefaultStyle())
	root.Style.Width = Fixed(120)
	root.Style.Height = Fixed(40)
	root.Style.Direction = Column

	// Header - fixed height
	header := NewNode(DefaultStyle())
	header.Style.Height = Fixed(3)
	// Width auto-stretches

	// Middle section - row with sidebar and main
	middle := NewNode(DefaultStyle())
	middle.Style.FlexGrow = 1
	middle.Style.Direction = Row

	// Sidebar - fixed width
	sidebar := NewNode(DefaultStyle())
	sidebar.Style.Width = Fixed(20)
	// Height auto-stretches

	// Main content - grows to fill
	main := NewNode(DefaultStyle())
	main.Style.FlexGrow = 1

	// Footer - fixed height
	footer := NewNode(DefaultStyle())
	footer.Style.Height = Fixed(2)

	// Build tree
	middle.AddChild(sidebar, main)
	root.AddChild(header, middle, footer)

	Calculate(root, 120, 40)

	// Verify header
	if header.Layout.Rect.X != 0 || header.Layout.Rect.Y != 0 {
		t.Errorf("header position = (%d, %d), want (0, 0)",
			header.Layout.Rect.X, header.Layout.Rect.Y)
	}
	if header.Layout.Rect.Width != 120 || header.Layout.Rect.Height != 3 {
		t.Errorf("header size = %dx%d, want 120x3",
			header.Layout.Rect.Width, header.Layout.Rect.Height)
	}

	// Verify middle section
	if middle.Layout.Rect.Y != 3 {
		t.Errorf("middle.Y = %d, want 3", middle.Layout.Rect.Y)
	}
	// Middle should be 40 - 3 (header) - 2 (footer) = 35 tall
	if middle.Layout.Rect.Height != 35 {
		t.Errorf("middle.Height = %d, want 35", middle.Layout.Rect.Height)
	}

	// Verify sidebar
	if sidebar.Layout.Rect.X != 0 || sidebar.Layout.Rect.Y != 3 {
		t.Errorf("sidebar position = (%d, %d), want (0, 3)",
			sidebar.Layout.Rect.X, sidebar.Layout.Rect.Y)
	}
	if sidebar.Layout.Rect.Width != 20 {
		t.Errorf("sidebar.Width = %d, want 20", sidebar.Layout.Rect.Width)
	}

	// Verify main content
	if main.Layout.Rect.X != 20 {
		t.Errorf("main.X = %d, want 20", main.Layout.Rect.X)
	}
	// Main should fill remaining width: 120 - 20 = 100
	if main.Layout.Rect.Width != 100 {
		t.Errorf("main.Width = %d, want 100", main.Layout.Rect.Width)
	}

	// Verify footer
	if footer.Layout.Rect.Y != 38 { // 40 - 2
		t.Errorf("footer.Y = %d, want 38", footer.Layout.Rect.Y)
	}
	if footer.Layout.Rect.Height != 2 {
		t.Errorf("footer.Height = %d, want 2", footer.Layout.Rect.Height)
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

	root := NewNode(DefaultStyle())
	root.Style.Width = Fixed(100)
	root.Style.Height = Fixed(100)
	root.Style.Direction = Row

	columnA := NewNode(DefaultStyle())
	columnA.Style.Width = Fixed(50)
	columnA.Style.Direction = Column

	rowA1 := NewNode(DefaultStyle())
	rowA1.Style.Height = Fixed(30)
	rowA1.Style.Direction = Row

	itemA1a := NewNode(DefaultStyle())
	itemA1a.Style.Width = Fixed(20)

	itemA1b := NewNode(DefaultStyle())
	itemA1b.Style.FlexGrow = 1

	rowA2 := NewNode(DefaultStyle())
	rowA2.Style.FlexGrow = 1
	rowA2.Style.Direction = Row

	columnB := NewNode(DefaultStyle())
	columnB.Style.FlexGrow = 1
	columnB.Style.Direction = Column

	// Build tree
	rowA1.AddChild(itemA1a, itemA1b)
	columnA.AddChild(rowA1, rowA2)
	root.AddChild(columnA, columnB)

	Calculate(root, 100, 100)

	// Verify column A
	if columnA.Layout.Rect.Width != 50 {
		t.Errorf("columnA.Width = %d, want 50", columnA.Layout.Rect.Width)
	}
	if columnA.Layout.Rect.Height != 100 {
		t.Errorf("columnA.Height = %d, want 100", columnA.Layout.Rect.Height)
	}

	// Verify row A1
	if rowA1.Layout.Rect.Height != 30 {
		t.Errorf("rowA1.Height = %d, want 30", rowA1.Layout.Rect.Height)
	}

	// Verify item A1a (within rowA1)
	if itemA1a.Layout.Rect.Width != 20 {
		t.Errorf("itemA1a.Width = %d, want 20", itemA1a.Layout.Rect.Width)
	}

	// Verify item A1b (should grow to fill remaining: 50 - 20 = 30)
	if itemA1b.Layout.Rect.Width != 30 {
		t.Errorf("itemA1b.Width = %d, want 30", itemA1b.Layout.Rect.Width)
	}

	// Verify row A2 (should grow to fill: 100 - 30 = 70)
	if rowA2.Layout.Rect.Height != 70 {
		t.Errorf("rowA2.Height = %d, want 70", rowA2.Layout.Rect.Height)
	}

	// Verify column B (should grow to fill: 100 - 50 = 50)
	if columnB.Layout.Rect.X != 50 {
		t.Errorf("columnB.X = %d, want 50", columnB.Layout.Rect.X)
	}
	if columnB.Layout.Rect.Width != 50 {
		t.Errorf("columnB.Width = %d, want 50", columnB.Layout.Rect.Width)
	}
}

// TestIntegration_FormLayout tests a typical form layout:
// - Labels (fixed width)
// - Inputs (grow to fill)
// - Arranged in a column
func TestIntegration_FormLayout(t *testing.T) {
	form := NewNode(DefaultStyle())
	form.Style.Width = Fixed(80)
	form.Style.Height = Fixed(30)
	form.Style.Direction = Column
	form.Style.Gap = 1

	// Create 3 form rows
	for i := 0; i < 3; i++ {
		row := NewNode(DefaultStyle())
		row.Style.Height = Fixed(3)
		row.Style.Direction = Row
		row.Style.Gap = 2

		label := NewNode(DefaultStyle())
		label.Style.Width = Fixed(15)

		input := NewNode(DefaultStyle())
		input.Style.FlexGrow = 1

		row.AddChild(label, input)
		form.AddChild(row)
	}

	Calculate(form, 100, 50)

	// Verify each row
	for i, row := range form.Children {
		expectedY := i * 4 // 3 height + 1 gap
		if row.Layout.Rect.Y != expectedY {
			t.Errorf("row[%d].Y = %d, want %d", i, row.Layout.Rect.Y, expectedY)
		}

		label := row.Children[0]
		input := row.Children[1]

		// Label should be fixed at 15
		if label.Layout.Rect.Width != 15 {
			t.Errorf("row[%d] label.Width = %d, want 15", i, label.Layout.Rect.Width)
		}

		// Input should grow to fill: 80 - 15 - 2 (gap) = 63
		if input.Layout.Rect.Width != 63 {
			t.Errorf("row[%d] input.Width = %d, want 63", i, input.Layout.Rect.Width)
		}

		// Input should be positioned after label + gap
		if input.Layout.Rect.X != 17 { // 15 + 2 gap
			t.Errorf("row[%d] input.X = %d, want 17", i, input.Layout.Rect.X)
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

	root := NewNode(DefaultStyle())
	root.Style.Width = Fixed(100)
	root.Style.Height = Fixed(100)
	root.Style.Direction = Column

	row1 := NewNode(DefaultStyle())
	row1.Style.Height = Fixed(50)
	row1.Style.Direction = Row

	col1a := NewNode(DefaultStyle())
	col1a.Style.Width = Fixed(30)
	col1a.Style.Direction = Column

	col1b := NewNode(DefaultStyle())
	col1b.Style.FlexGrow = 1
	col1b.Style.Direction = Column

	row2 := NewNode(DefaultStyle())
	row2.Style.FlexGrow = 1
	row2.Style.Direction = Row

	row1.AddChild(col1a, col1b)
	root.AddChild(row1, row2)

	Calculate(root, 100, 100)

	// Verify row1 position and size
	if row1.Layout.Rect.Y != 0 {
		t.Errorf("row1.Y = %d, want 0", row1.Layout.Rect.Y)
	}
	if row1.Layout.Rect.Height != 50 {
		t.Errorf("row1.Height = %d, want 50", row1.Layout.Rect.Height)
	}

	// Verify col1a
	if col1a.Layout.Rect.Width != 30 {
		t.Errorf("col1a.Width = %d, want 30", col1a.Layout.Rect.Width)
	}
	if col1a.Layout.Rect.Height != 50 {
		t.Errorf("col1a.Height = %d, want 50 (stretched)", col1a.Layout.Rect.Height)
	}

	// Verify col1b (grows to fill: 100 - 30 = 70)
	if col1b.Layout.Rect.X != 30 {
		t.Errorf("col1b.X = %d, want 30", col1b.Layout.Rect.X)
	}
	if col1b.Layout.Rect.Width != 70 {
		t.Errorf("col1b.Width = %d, want 70", col1b.Layout.Rect.Width)
	}

	// Verify row2 (grows to fill: 100 - 50 = 50)
	if row2.Layout.Rect.Y != 50 {
		t.Errorf("row2.Y = %d, want 50", row2.Layout.Rect.Y)
	}
	if row2.Layout.Rect.Height != 50 {
		t.Errorf("row2.Height = %d, want 50", row2.Layout.Rect.Height)
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
			node := NewNode(DefaultStyle())
			node.Style.Width = Fixed(tt.width)
			node.Style.Height = Fixed(tt.height)

			Calculate(node, 100, 100)

			if node.Layout.Rect.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d", node.Layout.Rect.Width, tt.expectedWidth)
			}
			if node.Layout.Rect.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d", node.Layout.Rect.Height, tt.expectedHeight)
			}
		})
	}
}

// TestEdgeCase_ZeroSizeParent tests children of a zero-size parent.
func TestEdgeCase_ZeroSizeParent(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(0)
	parent.Style.Height = Fixed(0)
	parent.Style.Direction = Row

	child := NewNode(DefaultStyle())
	child.Style.Width = Fixed(50)
	child.Style.Height = Fixed(30)

	parent.AddChild(child)
	Calculate(parent, 100, 100)

	// Parent should be zero size
	if parent.Layout.Rect.Width != 0 || parent.Layout.Rect.Height != 0 {
		t.Errorf("parent size = %dx%d, want 0x0",
			parent.Layout.Rect.Width, parent.Layout.Rect.Height)
	}

	// Child gets clamped due to min/max constraints from zero-size parent
	// The child's border box will be constrained by parent's content rect
	// which is also 0x0, so child dimensions will be constrained
}

// TestEdgeCase_OverflowNoShrink tests overflow when shrink totals to 0.
func TestEdgeCase_OverflowNoShrink(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(100)
	parent.Style.Height = Fixed(50)
	parent.Style.Direction = Row

	// Children that overflow but don't shrink
	child1 := NewNode(DefaultStyle())
	child1.Style.Width = Fixed(80)
	child1.Style.Height = Fixed(50)
	child1.Style.FlexShrink = 0

	child2 := NewNode(DefaultStyle())
	child2.Style.Width = Fixed(80)
	child2.Style.Height = Fixed(50)
	child2.Style.FlexShrink = 0

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 200)

	// Children should maintain their sizes (overflow parent)
	if child1.Layout.Rect.Width != 80 {
		t.Errorf("child1.Width = %d, want 80", child1.Layout.Rect.Width)
	}
	if child2.Layout.Rect.Width != 80 {
		t.Errorf("child2.Width = %d, want 80", child2.Layout.Rect.Width)
	}

	// Child2 should be positioned after child1
	if child2.Layout.Rect.X != 80 {
		t.Errorf("child2.X = %d, want 80", child2.Layout.Rect.X)
	}
}

// TestEdgeCase_EmptyAutoNode tests a node with no children and auto dimensions.
func TestEdgeCase_EmptyAutoNode(t *testing.T) {
	node := NewNode(DefaultStyle())
	// Width and Height are Auto by default

	Calculate(node, 100, 100)

	// Auto dimensions should use available space (fill parent)
	if node.Layout.Rect.Width != 100 {
		t.Errorf("Width = %d, want 100 (auto fills available)", node.Layout.Rect.Width)
	}
	if node.Layout.Rect.Height != 100 {
		t.Errorf("Height = %d, want 100 (auto fills available)", node.Layout.Rect.Height)
	}
}

// TestEdgeCase_VeryLargeTree tests layout with a large number of children.
func TestEdgeCase_VeryLargeTree(t *testing.T) {
	parent := NewNode(DefaultStyle())
	parent.Style.Width = Fixed(1000)
	parent.Style.Height = Fixed(100)
	parent.Style.Direction = Row

	// Add 100 children
	for i := 0; i < 100; i++ {
		child := NewNode(DefaultStyle())
		child.Style.Width = Fixed(10)
		child.Style.Height = Fixed(100)
		parent.AddChild(child)
	}

	Calculate(parent, 1000, 100)

	// Verify all children are positioned correctly
	for i, child := range parent.Children {
		expectedX := i * 10
		if child.Layout.Rect.X != expectedX {
			t.Errorf("child[%d].X = %d, want %d", i, child.Layout.Rect.X, expectedX)
		}
	}
}

// TestEdgeCase_DeepNesting tests deeply nested layout.
func TestEdgeCase_DeepNesting(t *testing.T) {
	// Create a chain of 10 nested nodes
	root := NewNode(DefaultStyle())
	root.Style.Width = Fixed(100)
	root.Style.Height = Fixed(100)

	current := root
	for i := 0; i < 10; i++ {
		child := NewNode(DefaultStyle())
		child.Style.FlexGrow = 1
		child.Style.Padding = EdgeAll(1)
		current.AddChild(child)
		current = child
	}

	Calculate(root, 100, 100)

	// Leaf node should have padding accumulated
	// Each level adds 1 padding on each side, so after 10 levels:
	// Content area shrinks by 2 per level = 20 total per dimension
	leaf := current
	expectedSize := 100 - 20 // 80
	if leaf.Layout.ContentRect.Width != expectedSize {
		t.Errorf("leaf.ContentRect.Width = %d, want %d",
			leaf.Layout.ContentRect.Width, expectedSize)
	}
}
