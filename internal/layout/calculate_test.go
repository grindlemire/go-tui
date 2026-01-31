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
