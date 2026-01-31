package layout

import "testing"

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

func TestCalculate_IntrinsicSize_WithGap(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(200)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row
	parent.style.Gap = 10

	child1 := newTestNode(DefaultStyle())
	child1.SetIntrinsicSize(40, 30)

	child2 := newTestNode(DefaultStyle())
	child2.SetIntrinsicSize(50, 25)

	parent.AddChild(child1, child2)
	Calculate(parent, 200, 100)

	// Children should be positioned with gap between them
	if child1.layout.Rect.X != 0 {
		t.Errorf("child1.X = %d, want 0", child1.layout.Rect.X)
	}
	if child2.layout.Rect.X != 50 { // 40 (child1 width) + 10 (gap)
		t.Errorf("child2.X = %d, want 50", child2.layout.Rect.X)
	}
}

func TestCalculate_IntrinsicSize_FlexGrowFromIntrinsic(t *testing.T) {
	// FlexGrow should add to intrinsic base size, not replace it
	parent := newTestNode(DefaultStyle())
	parent.style.Width = Fixed(100)
	parent.style.Height = Fixed(100)
	parent.style.Direction = Row

	child1 := newTestNode(DefaultStyle())
	child1.SetIntrinsicSize(20, 50)
	child1.style.FlexGrow = 1

	child2 := newTestNode(DefaultStyle())
	child2.SetIntrinsicSize(30, 50)
	child2.style.FlexGrow = 1

	parent.AddChild(child1, child2)
	Calculate(parent, 100, 100)

	// Total intrinsic = 20 + 30 = 50
	// Free space = 100 - 50 = 50
	// Each gets 25 extra (equal grow)
	// child1 = 20 + 25 = 45
	// child2 = 30 + 25 = 55
	if child1.layout.Rect.Width != 45 {
		t.Errorf("child1.Width = %d, want 45 (intrinsic + grow)", child1.layout.Rect.Width)
	}
	if child2.layout.Rect.Width != 55 {
		t.Errorf("child2.Width = %d, want 55 (intrinsic + grow)", child2.layout.Rect.Width)
	}
}
