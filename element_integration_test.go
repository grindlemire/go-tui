package tui

import (
	"testing"
)

// TestIntegration_BasicFlow tests the complete flow: New → AddChild → Render
func TestIntegration_BasicFlow(t *testing.T) {
	// Create root element
	root := New(
		WithSize(80, 24),
		WithDirection(Column),
	)

	// Add a child panel
	panel := New(
		WithSize(40, 10),
		WithBorder(BorderSingle),
	)

	root.AddChild(panel)

	// Render to buffer
	buf := NewBuffer(80, 24)
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
		WithDirection(Column),
	)

	topRow := New(
		WithHeight(20),
		WithDirection(Row),
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
	buf := NewBuffer(100, 50)
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
				WithDirection(Row),
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

			buf := NewBuffer(tt.parentWidth, 50)
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
		WithDirection(Column),
		WithJustify(JustifyCenter),
		WithAlign(AlignCenter),
	)

	// Panel with border
	panel := New(
		WithSize(40, 10),
		WithBorder(BorderRounded),
		WithDirection(Column),
		WithPadding(1),
		WithJustify(JustifyCenter),
		WithAlign(AlignCenter),
	)

	// Text element inside panel using new WithText API
	// WithText sets intrinsic width to text width and height to 1
	title := New(
		WithText("Hello World"),
		WithTextStyle(NewStyle().Bold()),
		WithTextAlign(TextAlignCenter),
		WithSize(38, 1), // Override intrinsic size for centering to work
	)

	panel.AddChild(title)
	root.AddChild(panel)

	// Render elements first (for layout and borders)
	// Text is now rendered automatically as part of RenderTree
	buf := NewBuffer(80, 24)
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
	bg := NewStyle().Background(Blue)
	border := NewStyle().Foreground(Red)

	panel := New(
		WithSize(10, 5),
		WithBorder(BorderSingle),
		WithBorderStyle(border),
		WithBackground(bg),
	)

	buf := NewBuffer(20, 10)
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
	if borderCell.Style.Fg != Red {
		t.Errorf("border foreground = %d, want %d (Red)", borderCell.Style.Fg, Red)
	}
}

