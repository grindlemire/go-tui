package layout

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
	calculateNode(root, available)
}

// calculateNode computes the layout for a single node within the available space.
// The available rect represents the border box space allocated by the parent
// (after the parent has already applied this node's margin).
func calculateNode(node Layoutable, available Rect) {
	// Dirty propagates up, so a clean node guarantees a clean subtree
	if !node.IsDirty() {
		return
	}

	style := node.LayoutStyle()

	// 1. Compute this node's border box within available space
	borderBox := computeBorderBox(style, available)

	// 2. Compute content rect (border box minus padding)
	contentRect := borderBox.Inset(style.Padding)

	// 3. Layout children within content rect
	children := node.LayoutChildren()
	if len(children) > 0 {
		layoutChildren(node, contentRect)
	}

	// 4. Store computed layout
	node.SetLayout(Layout{
		Rect:        borderBox,
		ContentRect: contentRect,
	})

	// 5. Clear dirty flag
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
