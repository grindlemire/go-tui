package layout

import "testing"

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
