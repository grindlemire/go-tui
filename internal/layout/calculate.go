package layout

import "math"

// Calculate performs layout calculation on the tree rooted at root.
// The root and all descendants will have their Layout field populated.
// Only dirty nodes are recalculated (incremental layout).
//
// availableWidth and availableHeight specify the root constraint
// (typically the terminal size).
func Calculate(root Layoutable, availableWidth, availableHeight int) {
	if root == nil {
		return
	}

	// For the root node, resolve its width/height constraints against
	// the available space. This is different from child nodes, which
	// receive their size from the parent's flex calculations.
	style := root.LayoutStyle()
	width := style.Width.Resolve(availableWidth, availableWidth)
	height := style.Height.Resolve(availableHeight, availableHeight)

	available := NewRect(0, 0, width, height)
	calculateNode(root, available, 0.0, 0.0)
}

// calculateNode computes the layout for a single node within the available space.
// The available rect represents the border box space allocated by the parent
// (after the parent has already applied this node's margin).
//
// absoluteX and absoluteY are the true float positions passed from the parent.
// This enables Yoga-style rounding: we track float positions through the tree
// and only round once when computing the final integer Rect.
func calculateNode(node Layoutable, available Rect, absoluteX, absoluteY float64) {
	// Dirty propagates up, so a clean node guarantees a clean subtree
	if !node.IsDirty() {
		return
	}

	style := node.LayoutStyle()

	// 1. Compute this node's border box dimensions (width/height only)
	borderBox := computeBorderBox(style, available)

	// Pre-sizing pass for flex-wrap containers with auto cross-axis size.
	// Determines how many wrap lines are needed and adjusts the border box
	// cross dimension before running the full layout.
	if style.FlexWrap != WrapNone && len(node.LayoutChildren()) > 0 && node.Tag() != "table" {
		isRow := style.Direction == Row
		if style.Display == DisplayBlock {
			isRow = false
		}

		// Check if cross-axis is Auto
		crossIsAuto := false
		if isRow {
			crossIsAuto = style.Height.IsAuto()
		} else {
			crossIsAuto = style.Width.IsAuto()
		}

		if crossIsAuto {
			contentWidth := borderBox.Width - style.Padding.Horizontal()
			contentHeight := borderBox.Height - style.Padding.Vertical()

			mainSz := contentWidth
			crossSz := contentHeight
			if !isRow {
				mainSz, crossSz = crossSz, mainSz
			}

			// Quick Phase 1: compute base sizes
			children := node.LayoutChildren()
			preItems := make([]flexItem, len(children))
			for i, child := range children {
				childStyle := child.LayoutStyle()
				var mainMargin int
				if isRow {
					mainMargin = childStyle.Margin.Horizontal()
				} else {
					mainMargin = childStyle.Margin.Vertical()
				}
				childIntrinsicW, childIntrinsicH := child.IntrinsicSize()
				if isRow {
					preItems[i].baseSize = childStyle.Width.Resolve(mainSz, childIntrinsicW) + mainMargin
				} else {
					preItems[i].baseSize = childStyle.Height.Resolve(mainSz, childIntrinsicH) + mainMargin
				}
			}

			// Break into lines
			preLines := breakIntoLines(preItems, mainSz, style.Gap)

			// Measure cross size per line
			totalCross := 0
			for _, pl := range preLines {
				maxCross := 0
				for j := pl.startIdx; j < pl.endIdx; j++ {
					child := children[j]
					childStyle := child.LayoutStyle()
					childIntrinsicW, childIntrinsicH := child.IntrinsicSize()

					var cross int
					if isRow {
						childWidth := preItems[j].baseSize - childStyle.Margin.Horizontal()
						wrappedH := child.HeightForWidth(childWidth)
						if wrappedH > childIntrinsicH {
							cross = wrappedH
						} else {
							cross = childIntrinsicH
						}
						cross += childStyle.Margin.Vertical()
					} else {
						cross = childIntrinsicW + childStyle.Margin.Horizontal()
					}

					// Use explicit cross size if set
					var crossStyleValue Value
					var crossMargin int
					if isRow {
						crossStyleValue = childStyle.Height
						crossMargin = childStyle.Margin.Vertical()
					} else {
						crossStyleValue = childStyle.Width
						crossMargin = childStyle.Margin.Horizontal()
					}
					if !crossStyleValue.IsAuto() {
						cross = crossStyleValue.Resolve(crossSz-crossMargin, 0) + crossMargin
					}

					if cross > maxCross {
						maxCross = cross
					}
				}
				totalCross += maxCross
			}

			// Adjust border box
			if isRow {
				borderBox.Height = totalCross + style.Padding.Vertical()
			} else {
				borderBox.Width = totalCross + style.Padding.Horizontal()
			}
		}
	}

	// 2. Set border box position from the rounded absolute float position
	borderBox.X = int(math.Round(absoluteX))
	borderBox.Y = int(math.Round(absoluteY))

	// 3. Compute content rect position from float position (then round)
	contentAbsX := absoluteX + float64(style.Padding.Left)
	contentAbsY := absoluteY + float64(style.Padding.Top)
	contentRect := Rect{
		X:      int(math.Round(contentAbsX)),
		Y:      int(math.Round(contentAbsY)),
		Width:  borderBox.Width - style.Padding.Horizontal(),
		Height: borderBox.Height - style.Padding.Vertical(),
	}

	// 4. Layout children within content rect, passing float positions
	children := node.LayoutChildren()
	if len(children) > 0 {
		if node.Tag() == "table" {
			layoutTable(node, contentRect, contentAbsX, contentAbsY)
		} else {
			layoutChildren(node, contentRect, contentAbsX, contentAbsY)
		}
	}

	// 5. Store computed layout with float positions for child calculations
	node.SetLayout(Layout{
		Rect:        borderBox,
		ContentRect: contentRect,
		AbsoluteX:   absoluteX,
		AbsoluteY:   absoluteY,
	})

	// 6. Clear dirty flag
	node.SetDirty(false)
}

// computeBorderBox calculates the border box dimensions for a node.
// The available rect is the space allocated by the parent (after margin and flex).
// For flex children, the available rect already contains the flex-computed size,
// so this function just uses the available dimensions directly.
// Only min/max constraints are applied; Width/Height were already used by the
// flex algorithm to compute the slot size.
func computeBorderBox(style Style, available Rect) Rect {
	// Start with available dimensions (flex-computed or parent-allocated)
	width := available.Width
	height := available.Height

	// Apply min/max width constraints
	minWidth := style.MinWidth.Resolve(available.Width, 0)
	maxWidth := style.MaxWidth.Resolve(available.Width, available.Width)
	width = clamp(width, minWidth, maxWidth)

	// Apply min/max height constraints
	minHeight := style.MinHeight.Resolve(available.Height, 0)
	maxHeight := style.MaxHeight.Resolve(available.Height, available.Height)
	height = clamp(height, minHeight, maxHeight)

	// Clamp to non-negative
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return Rect{
		X:      available.X,
		Y:      available.Y,
		Width:  width,
		Height: height,
	}
}

// clamp restricts v to the range [minVal, maxVal].
// If minVal > maxVal, minVal wins (matches CSS behavior).
func clamp(v, minVal, maxVal int) int {
	if v < minVal {
		return minVal
	}
	if maxVal >= minVal && v > maxVal {
		return maxVal
	}
	return v
}
