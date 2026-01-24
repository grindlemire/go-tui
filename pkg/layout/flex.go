package layout

// flexItem holds intermediate calculation state for a child.
// This is stack-allocated per layout call, not stored on nodes.
type flexItem struct {
	node      *Node
	baseSize  int
	mainSize  int
	crossSize int
	mainPos   int
	crossPos  int
	grow      float64
	shrink    float64
}

// layoutChildren arranges the children of a node within the given content rect.
// This implements the core flexbox algorithm.
func layoutChildren(node *Node, contentRect Rect) {
	if len(node.Children) == 0 {
		return
	}

	style := node.Style
	isRow := style.Direction == Row

	// Determine main/cross axis dimensions
	mainSize := contentRect.Width
	crossSize := contentRect.Height
	if !isRow {
		mainSize, crossSize = crossSize, mainSize
	}

	// Phase 1: Compute base sizes and flex factors
	// Base size includes the child's content size plus its margin.
	// Margin is part of the child's "outer size" in the flex calculation.
	items := make([]flexItem, len(node.Children))
	totalFixed := 0
	totalGrow := 0.0
	totalShrink := 0.0

	for i, child := range node.Children {
		item := &items[i]
		item.node = child

		// Compute margin on main and cross axes
		var mainMargin, crossMargin int
		if isRow {
			mainMargin = child.Style.Margin.Horizontal()
			crossMargin = child.Style.Margin.Vertical()
		} else {
			mainMargin = child.Style.Margin.Vertical()
			crossMargin = child.Style.Margin.Horizontal()
		}

		// Resolve base content size (or 0 if auto), then add margin
		if isRow {
			item.baseSize = child.Style.Width.Resolve(mainSize, 0) + mainMargin
		} else {
			item.baseSize = child.Style.Height.Resolve(mainSize, 0) + mainMargin
		}

		// Store margin for later use
		_ = crossMargin // Will be used in cross-axis sizing

		item.grow = child.Style.FlexGrow
		item.shrink = child.Style.FlexShrink

		totalFixed += item.baseSize
		totalGrow += item.grow
		totalShrink += item.shrink
	}

	// Account for gaps
	totalGap := style.Gap * max(0, len(node.Children)-1)
	freeSpace := mainSize - totalFixed - totalGap

	// Phase 2: Distribute free space
	if freeSpace > 0 && totalGrow > 0 {
		// Grow items
		for i := range items {
			if items[i].grow > 0 {
				extra := int(float64(freeSpace) * items[i].grow / totalGrow)
				items[i].mainSize = items[i].baseSize + extra
			} else {
				items[i].mainSize = items[i].baseSize
			}
		}
	} else if freeSpace < 0 && totalShrink > 0 {
		// Shrink items
		deficit := -freeSpace
		for i := range items {
			if items[i].shrink > 0 {
				reduction := int(float64(deficit) * items[i].shrink / totalShrink)
				items[i].mainSize = max(0, items[i].baseSize-reduction)
			} else {
				items[i].mainSize = items[i].baseSize
			}
		}
	} else {
		// No flex needed
		for i := range items {
			items[i].mainSize = items[i].baseSize
		}
		freeSpace = max(0, freeSpace) // For justify calculations
	}

	// Phase 3: Apply min/max constraints
	for i, child := range node.Children {
		minMain := resolveMinMain(child.Style, isRow, mainSize)
		maxMain := resolveMaxMain(child.Style, isRow, mainSize)
		items[i].mainSize = clampFlex(items[i].mainSize, minMain, maxMain)
	}

	// Recalculate free space after min/max constraints
	// (needed for justify calculations)
	totalUsed := 0
	for i := range items {
		totalUsed += items[i].mainSize
	}
	freeSpace = mainSize - totalUsed - totalGap

	// Phase 4: Position children along main axis (justify)
	// For Phase 2, we only handle JustifyStart
	offset := calculateJustifyOffset(style.JustifyContent, freeSpace, len(items))
	spacing := calculateJustifySpacing(style.JustifyContent, freeSpace, len(items))

	for i := range items {
		items[i].mainPos = offset
		offset += items[i].mainSize + style.Gap + spacing
	}

	// Phase 5: Cross-axis sizing and alignment
	for i, child := range node.Children {
		align := style.AlignItems
		if child.Style.AlignSelf != nil {
			align = *child.Style.AlignSelf
		}

		// Determine cross-axis size value
		var crossStyleValue Value
		var crossMargin int
		if isRow {
			crossStyleValue = child.Style.Height
			crossMargin = child.Style.Margin.Vertical()
		} else {
			crossStyleValue = child.Style.Width
			crossMargin = child.Style.Margin.Horizontal()
		}

		// Available cross space after margin
		availableCross := crossSize - crossMargin

		if align == AlignStretch && crossStyleValue.IsAuto() {
			// Stretch: fill the available cross axis (minus margin)
			items[i].crossSize = availableCross + crossMargin // Include margin in slot size
			items[i].crossPos = 0
		} else {
			// Non-stretch or explicit size: use the specified value or stretch to available
			var contentCross int
			if crossStyleValue.IsAuto() {
				contentCross = availableCross
			} else {
				contentCross = crossStyleValue.Resolve(availableCross, availableCross)
			}
			// Slot size includes content + margin
			items[i].crossSize = contentCross + crossMargin
			items[i].crossPos = calculateAlignOffset(align, crossSize, items[i].crossSize)
		}
	}

	// Phase 6: Convert to rects and recurse
	for i, child := range node.Children {
		// Compute the slot allocated to this child (before margin)
		var slot Rect
		if isRow {
			slot = Rect{
				X:      contentRect.X + items[i].mainPos,
				Y:      contentRect.Y + items[i].crossPos,
				Width:  items[i].mainSize,
				Height: items[i].crossSize,
			}
		} else {
			slot = Rect{
				X:      contentRect.X + items[i].crossPos,
				Y:      contentRect.Y + items[i].mainPos,
				Width:  items[i].crossSize,
				Height: items[i].mainSize,
			}
		}

		// Apply child's margin: shrink the slot to get the child's border box.
		// The child receives this as 'available' and does NOT re-apply margin.
		childBorderBox := slot.Inset(child.Style.Margin)

		// Recurse—child computes its layout within this border box
		calculateNode(child, childBorderBox)
	}
}

// calculateJustifyOffset returns the initial offset for positioning children
// based on the justify mode and available free space.
func calculateJustifyOffset(justify Justify, freeSpace, itemCount int) int {
	if freeSpace <= 0 || itemCount == 0 {
		return 0
	}

	switch justify {
	case JustifyEnd:
		return freeSpace
	case JustifyCenter:
		return freeSpace / 2
	case JustifySpaceAround:
		if itemCount > 0 {
			return freeSpace / (itemCount * 2)
		}
		return 0
	case JustifySpaceEvenly:
		return freeSpace / (itemCount + 1)
	default: // JustifyStart, JustifySpaceBetween
		return 0
	}
}

// calculateJustifySpacing returns the extra spacing between children
// based on the justify mode and available free space.
func calculateJustifySpacing(justify Justify, freeSpace, itemCount int) int {
	if freeSpace <= 0 || itemCount <= 1 {
		return 0
	}

	switch justify {
	case JustifySpaceBetween:
		return freeSpace / (itemCount - 1)
	case JustifySpaceAround:
		return freeSpace / itemCount
	case JustifySpaceEvenly:
		return freeSpace / (itemCount + 1)
	default: // JustifyStart, JustifyEnd, JustifyCenter
		return 0
	}
}

// calculateAlignOffset returns the offset for positioning a child on the cross axis.
func calculateAlignOffset(align Align, crossSize, itemSize int) int {
	switch align {
	case AlignEnd:
		return crossSize - itemSize
	case AlignCenter:
		return (crossSize - itemSize) / 2
	default: // AlignStart, AlignStretch
		return 0
	}
}

// resolveMinMain resolves the minimum size constraint for the main axis.
func resolveMinMain(style Style, isRow bool, available int) int {
	if isRow {
		return style.MinWidth.Resolve(available, 0)
	}
	return style.MinHeight.Resolve(available, 0)
}

// resolveMaxMain resolves the maximum size constraint for the main axis.
// Returns a large value if no max is set (UnitAuto).
func resolveMaxMain(style Style, isRow bool, available int) int {
	if isRow {
		if style.MaxWidth.IsAuto() {
			return available // No constraint - use available as upper bound
		}
		return style.MaxWidth.Resolve(available, available)
	}
	if style.MaxHeight.IsAuto() {
		return available // No constraint - use available as upper bound
	}
	return style.MaxHeight.Resolve(available, available)
}

// clampFlex restricts v to the range [minVal, maxVal].
// If minVal > maxVal, minVal wins (matches CSS behavior).
func clampFlex(v, minVal, maxVal int) int {
	if v < minVal {
		return minVal
	}
	if maxVal >= minVal && v > maxVal {
		return maxVal
	}
	return v
}
