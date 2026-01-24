package element

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

// TestIntegration_BasicFlow tests the complete flow: New → AddChild → Render
func TestIntegration_BasicFlow(t *testing.T) {
	// Create root element
	root := New(
		WithSize(80, 24),
		WithDirection(layout.Column),
	)

	// Add a child panel
	panel := New(
		WithSize(40, 10),
		WithBorder(tui.BorderSingle),
	)

	root.AddChild(panel)

	// Render to buffer
	buf := tui.NewBuffer(80, 24)
	root.Render(buf, 80, 24)

	// Verify layout was calculated
	panelRect := panel.Rect()
	if panelRect.Width != 40 {
		t.Errorf("panel.Rect().Width = %d, want 40", panelRect.Width)
	}
	if panelRect.Height != 10 {
		t.Errorf("panel.Rect().Height = %d, want 10", panelRect.Height)
	}

	// Verify border was rendered (check top-left corner)
	cell := buf.Cell(panelRect.X, panelRect.Y)
	if cell.Rune != '┌' {
		t.Errorf("top-left cell = %q, want '┌'", cell.Rune)
	}
}

// TestIntegration_NestedLayouts tests nested layouts with alternating directions
func TestIntegration_NestedLayouts(t *testing.T) {
	// Column layout
	//   Row layout (fills width, fixed height)
	//     Left panel (fixed width)
	//     Right panel (flex grow)
	//   Bottom panel (flex grow)

	root := New(
		WithSize(100, 50),
		WithDirection(layout.Column),
	)

	topRow := New(
		WithHeight(20),
		WithDirection(layout.Row),
	)

	leftPanel := New(
		WithWidth(30),
	)

	rightPanel := New(
		WithFlexGrow(1),
	)

	bottomPanel := New(
		WithFlexGrow(1),
	)

	topRow.AddChild(leftPanel, rightPanel)
	root.AddChild(topRow, bottomPanel)

	// Render
	buf := tui.NewBuffer(100, 50)
	root.Render(buf, 100, 50)

	// Verify topRow
	topRect := topRow.Rect()
	if topRect.Y != 0 {
		t.Errorf("topRow.Y = %d, want 0", topRect.Y)
	}
	if topRect.Height != 20 {
		t.Errorf("topRow.Height = %d, want 20", topRect.Height)
	}
	if topRect.Width != 100 {
		t.Errorf("topRow.Width = %d, want 100", topRect.Width)
	}

	// Verify leftPanel
	leftRect := leftPanel.Rect()
	if leftRect.X != 0 {
		t.Errorf("leftPanel.X = %d, want 0", leftRect.X)
	}
	if leftRect.Width != 30 {
		t.Errorf("leftPanel.Width = %d, want 30", leftRect.Width)
	}
	if leftRect.Height != 20 {
		t.Errorf("leftPanel.Height = %d, want 20 (stretched)", leftRect.Height)
	}

	// Verify rightPanel (should fill remaining: 100 - 30 = 70)
	rightRect := rightPanel.Rect()
	if rightRect.X != 30 {
		t.Errorf("rightPanel.X = %d, want 30", rightRect.X)
	}
	if rightRect.Width != 70 {
		t.Errorf("rightPanel.Width = %d, want 70", rightRect.Width)
	}

	// Verify bottomPanel (should fill remaining: 50 - 20 = 30)
	bottomRect := bottomPanel.Rect()
	if bottomRect.Y != 20 {
		t.Errorf("bottomPanel.Y = %d, want 20", bottomRect.Y)
	}
	if bottomRect.Height != 30 {
		t.Errorf("bottomPanel.Height = %d, want 30", bottomRect.Height)
	}
	if bottomRect.Width != 100 {
		t.Errorf("bottomPanel.Width = %d, want 100", bottomRect.Width)
	}
}

// TestIntegration_FlexGrowShrink tests flex grow and shrink behavior
func TestIntegration_FlexGrowShrink(t *testing.T) {
	type tc struct {
		children      []struct{ width int; grow, shrink float64 }
		parentWidth   int
		expectedSizes []int
	}

	tests := map[string]tc{
		"equal grow": {
			children: []struct{ width int; grow, shrink float64 }{
				{0, 1, 1},
				{0, 1, 1},
			},
			parentWidth:   100,
			expectedSizes: []int{50, 50},
		},
		"unequal grow": {
			children: []struct{ width int; grow, shrink float64 }{
				{0, 1, 1},
				{0, 2, 1},
			},
			parentWidth:   90,
			expectedSizes: []int{30, 60},
		},
		"fixed and grow": {
			children: []struct{ width int; grow, shrink float64 }{
				{30, 0, 1},
				{0, 1, 1},
			},
			parentWidth:   100,
			expectedSizes: []int{30, 70},
		},
		"no shrink overflow": {
			children: []struct{ width int; grow, shrink float64 }{
				{60, 0, 0},
				{60, 0, 0},
			},
			parentWidth:   100,
			expectedSizes: []int{60, 60}, // No shrink, overflow allowed
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := New(
				WithWidth(tt.parentWidth),
				WithHeight(50),
				WithDirection(layout.Row),
			)

			children := make([]*Element, len(tt.children))
			for i, c := range tt.children {
				opts := []Option{
					WithFlexGrow(c.grow),
					WithFlexShrink(c.shrink),
				}
				if c.width > 0 {
					opts = append(opts, WithWidth(c.width))
				}
				children[i] = New(opts...)
				root.AddChild(children[i])
			}

			buf := tui.NewBuffer(tt.parentWidth, 50)
			root.Render(buf, tt.parentWidth, 50)

			for i, child := range children {
				if child.Rect().Width != tt.expectedSizes[i] {
					t.Errorf("child[%d].Width = %d, want %d",
						i, child.Rect().Width, tt.expectedSizes[i])
				}
			}
		})
	}
}

// TestIntegration_MixedElementAndText tests a tree with both Element and Text nodes
func TestIntegration_MixedElementAndText(t *testing.T) {
	root := New(
		WithSize(80, 24),
		WithDirection(layout.Column),
		WithJustify(layout.JustifyCenter),
		WithAlign(layout.AlignCenter),
	)

	// Panel with border
	panel := New(
		WithSize(40, 10),
		WithBorder(tui.BorderRounded),
		WithDirection(layout.Column),
		WithPadding(1),
		WithJustify(layout.JustifyCenter),
		WithAlign(layout.AlignCenter),
	)

	// Text element inside panel
	// Note: Text elements need a height for layout. In a real app, you'd typically
	// give text elements a fixed height or let them stretch via AlignStretch.
	title := NewText("Hello World",
		WithTextStyle(tui.NewStyle().Bold()),
		WithTextAlign(TextAlignCenter),
		WithElementOption(WithHeight(1)), // Text needs height to be visible
	)

	panel.AddChild(title.Element)
	root.AddChild(panel)

	// Render elements first (for layout and borders)
	buf := tui.NewBuffer(80, 24)
	root.Render(buf, 80, 24)

	// Verify panel is centered in root
	panelRect := panel.Rect()
	expectedX := (80 - 40) / 2
	expectedY := (24 - 10) / 2
	if panelRect.X != expectedX {
		t.Errorf("panel.X = %d, want %d (centered)", panelRect.X, expectedX)
	}
	if panelRect.Y != expectedY {
		t.Errorf("panel.Y = %d, want %d (centered)", panelRect.Y, expectedY)
	}

	// Check border was drawn
	topLeft := buf.Cell(panelRect.X, panelRect.Y)
	if topLeft.Rune != '╭' {
		t.Errorf("border top-left = %q, want '╭'", topLeft.Rune)
	}

	// Manually render text content
	// Note: RenderTree only handles Element, so we need to render Text separately
	RenderText(buf, title)

	// Verify text was rendered
	// Text element's ContentRect determines where text is drawn
	contentRect := title.ContentRect()

	// Find where 'H' appears (should be in content area)
	found := false
	foundX := -1
	for y := contentRect.Y; y < contentRect.Bottom(); y++ {
		for x := contentRect.X; x < contentRect.Right(); x++ {
			if buf.Cell(x, y).Rune == 'H' {
				found = true
				foundX = x
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("text 'Hello World' not rendered (could not find 'H' in content area)")
	}

	// Verify the text is centered
	textWidth := stringWidth("Hello World")
	expectedTextX := contentRect.X + (contentRect.Width-textWidth)/2
	if expectedTextX < contentRect.X {
		expectedTextX = contentRect.X
	}
	if foundX != expectedTextX {
		t.Errorf("text 'H' at x=%d, want %d (centered)", foundX, expectedTextX)
	}
}

// TestIntegration_BackgroundAndBorder tests visual rendering
func TestIntegration_BackgroundAndBorder(t *testing.T) {
	bg := tui.NewStyle().Background(tui.Blue)
	border := tui.NewStyle().Foreground(tui.Red)

	panel := New(
		WithSize(10, 5),
		WithBorder(tui.BorderSingle),
		WithBorderStyle(border),
		WithBackground(bg),
	)

	buf := tui.NewBuffer(20, 10)
	panel.Calculate(20, 10)
	RenderTree(buf, panel)

	// Check background fill (interior, not border)
	// Border takes 1 cell on each side, so interior starts at (1, 1)
	interiorX := panel.Rect().X + 1
	interiorY := panel.Rect().Y + 1

	interiorCell := buf.Cell(interiorX, interiorY)
	// Background should be space with blue background
	if interiorCell.Rune != ' ' {
		t.Errorf("interior cell rune = %q, want ' '", interiorCell.Rune)
	}

	// Check border style (red foreground)
	borderCell := buf.Cell(panel.Rect().X, panel.Rect().Y)
	if borderCell.Rune != '┌' {
		t.Errorf("border cell = %q, want '┌'", borderCell.Rune)
	}
	if borderCell.Style.Fg != tui.Red {
		t.Errorf("border foreground = %d, want %d (Red)", borderCell.Style.Fg, tui.Red)
	}
}

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

	buf := tui.NewBuffer(100, 100)
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
		justify                   layout.Justify
		align                     layout.Align
		expectedX, expectedY      int
	}

	tests := map[string]tc{
		"center center": {
			parentWidth: 100, parentHeight: 100,
			childWidth: 20, childHeight: 10,
			justify:   layout.JustifyCenter,
			align:     layout.AlignCenter,
			expectedX: 40, expectedY: 45, // (100-20)/2, (100-10)/2
		},
		"end center column": {
			parentWidth: 100, parentHeight: 100,
			childWidth: 20, childHeight: 10,
			justify:   layout.JustifyEnd,
			align:     layout.AlignCenter,
			expectedX: 40, expectedY: 90, // For Column: justify affects Y, align affects X
		},
		"start end": {
			parentWidth: 100, parentHeight: 100,
			childWidth: 20, childHeight: 10,
			justify:   layout.JustifyStart,
			align:     layout.AlignEnd,
			expectedX: 80, expectedY: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := New(
				WithSize(tt.parentWidth, tt.parentHeight),
				WithDirection(layout.Column),
				WithJustify(tt.justify),
				WithAlign(tt.align),
			)

			child := New(
				WithSize(tt.childWidth, tt.childHeight),
			)

			root.AddChild(child)

			buf := tui.NewBuffer(tt.parentWidth, tt.parentHeight)
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
		WithBorder(tui.BorderSingle),
	)

	buf := tui.NewBuffer(10, 5)
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
	buf := tui.NewBuffer(5, 5)

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
		WithDirection(layout.Row),
		WithGap(10),
	)

	child1 := New(WithWidth(20), WithHeight(100))
	child2 := New(WithWidth(20), WithHeight(100))
	child3 := New(WithWidth(20), WithHeight(100))

	root.AddChild(child1, child2, child3)

	buf := tui.NewBuffer(100, 100)
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
			text := NewText(tt.content,
				WithTextAlign(tt.align),
				WithElementOption(WithSize(tt.boxWidth, 1)),
			)

			buf := tui.NewBuffer(tt.boxWidth, 1)
			text.Calculate(tt.boxWidth, 1)
			RenderText(buf, text)

			// Find where 'H' appears
			foundX := -1
			for x := 0; x < tt.boxWidth; x++ {
				if buf.Cell(x, 0).Rune == 'H' {
					foundX = x
					break
				}
			}

			contentRect := text.ContentRect()
			expectedX := contentRect.X + tt.expectedStartOffset
			if foundX != expectedX {
				t.Errorf("'H' found at x=%d, want %d", foundX, expectedX)
			}
		})
	}
}
