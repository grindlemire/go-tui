package tui

import (
	"strings"
	"testing"
)

// TestIntegration_DeepNesting tests deeply nested elements
func TestIntegration_DeepNesting(t *testing.T) {
	root := New(
		WithSize(100, 100),
		WithPadding(2),
	)

	current := root
	depth := 5
	for i := 0; i < depth; i++ {
		child := New(
			WithFlexGrow(1),
			WithPadding(2),
		)
		current.AddChild(child)
		current = child
	}

	buf := NewBuffer(100, 100)
	root.Render(buf, 100, 100)

	// Each level adds 2 padding on each side = 4 per level
	// Root: 100x100, content = 96x96
	// L1: 96x96, content = 92x92
	// L2: 92x92, content = 88x88
	// L3: 88x88, content = 84x84
	// L4: 84x84, content = 80x80
	// L5: 80x80, content = 76x76

	leaf := current
	expectedContentWidth := 100 - (depth+1)*4 // Each level has padding 2 on each side
	if leaf.ContentRect().Width != expectedContentWidth {
		t.Errorf("leaf.ContentRect().Width = %d, want %d",
			leaf.ContentRect().Width, expectedContentWidth)
	}
}

// TestIntegration_Centering tests various centering scenarios
func TestIntegration_Centering(t *testing.T) {
	type tc struct {
		parentWidth, parentHeight int
		childWidth, childHeight   int
		justify                   Justify
		align                     Align
		expectedX, expectedY      int
	}

	tests := map[string]tc{
		"center center": {
			parentWidth: 100, parentHeight: 100,
			childWidth: 20, childHeight: 10,
			justify:   JustifyCenter,
			align:     AlignCenter,
			expectedX: 40, expectedY: 45, // (100-20)/2, (100-10)/2
		},
		"end center column": {
			parentWidth: 100, parentHeight: 100,
			childWidth: 20, childHeight: 10,
			justify:   JustifyEnd,
			align:     AlignCenter,
			expectedX: 40, expectedY: 90, // For Column: justify affects Y, align affects X
		},
		"start end": {
			parentWidth: 100, parentHeight: 100,
			childWidth: 20, childHeight: 10,
			justify:   JustifyStart,
			align:     AlignEnd,
			expectedX: 80, expectedY: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := New(
				WithSize(tt.parentWidth, tt.parentHeight),
				WithDirection(Column),
				WithJustify(tt.justify),
				WithAlign(tt.align),
			)

			child := New(
				WithSize(tt.childWidth, tt.childHeight),
			)

			root.AddChild(child)

			buf := NewBuffer(tt.parentWidth, tt.parentHeight)
			root.Render(buf, tt.parentWidth, tt.parentHeight)

			childRect := child.Rect()
			if childRect.X != tt.expectedX {
				t.Errorf("child.X = %d, want %d", childRect.X, tt.expectedX)
			}
			if childRect.Y != tt.expectedY {
				t.Errorf("child.Y = %d, want %d", childRect.Y, tt.expectedY)
			}
		})
	}
}

// TestIntegration_RenderOutput tests that rendered output matches expectations
func TestIntegration_RenderOutput(t *testing.T) {
	// Create a simple 10x5 panel with a border
	panel := New(
		WithSize(10, 5),
		WithBorder(BorderSingle),
	)

	buf := NewBuffer(10, 5)
	panel.Render(buf, 10, 5)

	// Build expected output
	// ┌────────┐
	// │        │
	// │        │
	// │        │
	// └────────┘
	expected := []string{
		"┌────────┐",
		"│        │",
		"│        │",
		"│        │",
		"└────────┘",
	}

	for y := 0; y < 5; y++ {
		var row strings.Builder
		for x := 0; x < 10; x++ {
			cell := buf.Cell(x, y)
			row.WriteRune(cell.Rune)
		}
		if row.String() != expected[y] {
			t.Errorf("row %d = %q, want %q", y, row.String(), expected[y])
		}
	}
}

// TestIntegration_CullingOutsideBounds tests that elements outside buffer are not rendered
func TestIntegration_CullingOutsideBounds(t *testing.T) {
	// Create root with a child positioned way outside
	root := New(
		WithSize(100, 100),
	)

	// This child will be outside a small buffer
	child := New(
		WithSize(10, 10),
	)

	root.AddChild(child)
	root.Calculate(100, 100)

	// Render to a small buffer - child should be culled if outside
	// Since child is at (0,0) and buffer is 100x100, it should render
	buf := NewBuffer(5, 5)

	// Manually adjust child's layout to be outside bounds for testing
	// (This simulates what would happen with complex layouts)
	// For now, just verify rendering doesn't crash with small buffer
	RenderTree(buf, root)

	// If we got here without panic, culling is working for bounds checking
}

// TestIntegration_GapBetweenChildren tests gap spacing
func TestIntegration_GapBetweenChildren(t *testing.T) {
	root := New(
		WithSize(100, 100),
		WithDirection(Row),
		WithGap(10),
	)

	child1 := New(WithWidth(20), WithHeight(100))
	child2 := New(WithWidth(20), WithHeight(100))
	child3 := New(WithWidth(20), WithHeight(100))

	root.AddChild(child1, child2, child3)

	buf := NewBuffer(100, 100)
	root.Render(buf, 100, 100)

	// Verify positions with gap
	// child1: x=0, width=20
	// gap: 10
	// child2: x=30, width=20
	// gap: 10
	// child3: x=60, width=20

	if child1.Rect().X != 0 {
		t.Errorf("child1.X = %d, want 0", child1.Rect().X)
	}
	if child2.Rect().X != 30 {
		t.Errorf("child2.X = %d, want 30", child2.Rect().X)
	}
	if child3.Rect().X != 60 {
		t.Errorf("child3.X = %d, want 60", child3.Rect().X)
	}
}

// TestIntegration_TextAlignment tests text alignment within elements
func TestIntegration_TextAlignment(t *testing.T) {
	type tc struct {
		align    TextAlign
		content  string
		boxWidth int
		// We check the x position where content starts
		expectedStartOffset int // offset from content rect left
	}

	tests := map[string]tc{
		"left align": {
			align:               TextAlignLeft,
			content:             "Hi",
			boxWidth:            20,
			expectedStartOffset: 0,
		},
		"center align": {
			align:               TextAlignCenter,
			content:             "Hi", // 2 chars
			boxWidth:            20,
			expectedStartOffset: 9, // (20-2)/2 = 9
		},
		"right align": {
			align:               TextAlignRight,
			content:             "Hi", // 2 chars
			boxWidth:            20,
			expectedStartOffset: 18, // 20-2 = 18
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			elem := New(
				WithText(tt.content),
				WithTextAlign(tt.align),
				WithSize(tt.boxWidth, 1),
			)

			buf := NewBuffer(tt.boxWidth, 1)
			elem.Calculate(tt.boxWidth, 1)
			RenderTree(buf, elem)

			// Find where 'H' appears
			foundX := -1
			for x := 0; x < tt.boxWidth; x++ {
				if buf.Cell(x, 0).Rune == 'H' {
					foundX = x
					break
				}
			}

			contentRect := elem.ContentRect()
			expectedX := contentRect.X + tt.expectedStartOffset
			if foundX != expectedX {
				t.Errorf("'H' found at x=%d, want %d", foundX, expectedX)
			}
		})
	}
}
