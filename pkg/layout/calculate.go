package layout

// Calculate performs layout calculation on the tree rooted at node.
// The node and all descendants will have their Layout field populated.
// Only dirty nodes are recalculated (incremental layout).
//
// availableWidth and availableHeight specify the root constraint
// (typically the terminal size).
func Calculate(node *Node, availableWidth, availableHeight int) {
	if node == nil {
		return
	}

	// For the root node, resolve its width/height constraints against
	// the available space. This is different from child nodes, which
	// receive their size from the parent's flex calculations.
	width := node.Style.Width.Resolve(availableWidth, availableWidth)
	height := node.Style.Height.Resolve(availableHeight, availableHeight)

	available := NewRect(0, 0, width, height)
	calculateNode(node, available)
}

// calculateNode computes the layout for a single node within the available space.
// The available rect represents the border box space allocated by the parent
// (after the parent has already applied this node's margin).
func calculateNode(node *Node, available Rect) {
	// Dirty propagates up, so a clean node guarantees a clean subtree
	if !node.dirty {
		return
	}

	// 1. Compute this node's border box within available space
	borderBox := computeBorderBox(node.Style, available)

	// 2. Compute content rect (border box minus padding)
	contentRect := borderBox.Inset(node.Style.Padding)

	// 3. Layout children within content rect
	if len(node.Children) > 0 {
		layoutChildren(node, contentRect)
	}

	// 4. Store computed layout
	node.Layout = Layout{
		Rect:        borderBox,
		ContentRect: contentRect,
	}

	// 5. Clear dirty flag
	node.dirty = false
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
