package layout

import "math"

// flexItem holds intermediate calculation state for a child.
// This is stack-allocated per layout call, not stored on nodes.
// Positions are stored as float64 to enable precise centering calculations
// that only round at the final stage, preventing jitter during animation.
type flexItem struct {
	node      Layoutable
	baseSize  int
	mainSize  int
	crossSize int
	mainPos   float64 // float to avoid centering jitter
	crossPos  float64 // float to avoid centering jitter
	grow      float64
	shrink    float64
}

// layoutChildren arranges the children of a node within the given content rect.
// This implements the core flexbox algorithm.
//
// parentAbsX and parentAbsY are the parent's absolute float positions (content rect origin).
// These are used for Yoga-style rounding: we compute each child's absolute float position
// and only round once when creating the final integer Rect.
func layoutChildren(node Layoutable, contentRect Rect, parentAbsX, parentAbsY float64) {
	children := node.LayoutChildren()
	if len(children) == 0 {
		return
	}

	style := node.LayoutStyle()
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
	items := make([]flexItem, len(children))
	totalFixed := 0
	totalGrow := 0.0
	totalShrink := 0.0

	for i, child := range children {
		item := &items[i]
		item.node = child

		childStyle := child.LayoutStyle()

		// Compute margin on main and cross axes
		var mainMargin, crossMargin int
		if isRow {
			mainMargin = childStyle.Margin.Horizontal()
			crossMargin = childStyle.Margin.Vertical()
		} else {
			mainMargin = childStyle.Margin.Vertical()
			crossMargin = childStyle.Margin.Horizontal()
		}

		// Resolve base content size (or 0 if auto), then add margin
		if isRow {
			item.baseSize = childStyle.Width.Resolve(mainSize, 0) + mainMargin
		} else {
			item.baseSize = childStyle.Height.Resolve(mainSize, 0) + mainMargin
		}

		// Store margin for later use
		_ = crossMargin // Will be used in cross-axis sizing

		item.grow = childStyle.FlexGrow
		item.shrink = childStyle.FlexShrink

		totalFixed += item.baseSize
		totalGrow += item.grow
		totalShrink += item.shrink
	}

	// Account for gaps
	totalGap := style.Gap * max(0, len(children)-1)
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
	for i, child := range children {
		childStyle := child.LayoutStyle()
		minMain := resolveMinMain(childStyle, isRow, mainSize)
		maxMain := resolveMaxMain(childStyle, isRow, mainSize)
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
	// Use float64 for offset to enable precise centering that only rounds at final stage
	offset := calculateJustifyOffset(style.JustifyContent, freeSpace, len(items))
	spacing := calculateJustifySpacing(style.JustifyContent, freeSpace, len(items))

	for i := range items {
		items[i].mainPos = offset
		offset += float64(items[i].mainSize) + float64(style.Gap) + spacing
	}

	// Phase 5: Cross-axis sizing and alignment
	for i, child := range children {
		childStyle := child.LayoutStyle()
		align := style.AlignItems
		if childStyle.AlignSelf != nil {
			align = *childStyle.AlignSelf
		}

		// Determine cross-axis size value
		var crossStyleValue Value
		var crossMargin int
		if isRow {
			crossStyleValue = childStyle.Height
			crossMargin = childStyle.Margin.Vertical()
		} else {
			crossStyleValue = childStyle.Width
			crossMargin = childStyle.Margin.Horizontal()
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
	// Yoga-style rounding: compute each child's ABSOLUTE float position,
	// then round once to get the integer Rect. This prevents jitter because
	// fractional parts accumulate correctly before rounding.
	for i, child := range children {
		childStyle := child.LayoutStyle()

		// Compute child's ABSOLUTE float position (parent float + relative offset)
		var childAbsX, childAbsY float64
		if isRow {
			childAbsX = parentAbsX + items[i].mainPos
			childAbsY = parentAbsY + items[i].crossPos
		} else {
			childAbsX = parentAbsX + items[i].crossPos
			childAbsY = parentAbsY + items[i].mainPos
		}

		// Round absolute position to get integer slot
		var slot Rect
		if isRow {
			slot = Rect{
				X:      int(math.Round(childAbsX)),
				Y:      int(math.Round(childAbsY)),
				Width:  items[i].mainSize,
				Height: items[i].crossSize,
			}
		} else {
			slot = Rect{
				X:      int(math.Round(childAbsX)),
				Y:      int(math.Round(childAbsY)),
				Width:  items[i].crossSize,
				Height: items[i].mainSize,
			}
		}

		// Apply child's margin: shrink the slot to get the child's border box.
		// Also adjust float position to account for margin.
		if isRow {
			childAbsX += float64(childStyle.Margin.Left)
			childAbsY += float64(childStyle.Margin.Top)
		} else {
			childAbsX += float64(childStyle.Margin.Left)
			childAbsY += float64(childStyle.Margin.Top)
		}
		childBorderBox := slot.Inset(childStyle.Margin)

		// Force child to recalculate since parent layout changed.
		child.SetDirty(true)

		// Recurse with FLOAT position for jitter-free child positioning
		calculateNode(child, childBorderBox, childAbsX, childAbsY)
	}
}

// calculateJustifyOffset returns the initial offset for positioning children
// based on the justify mode and available free space.
// Returns float64 to enable precise centering that only rounds at the final stage.
func calculateJustifyOffset(justify Justify, freeSpace, itemCount int) float64 {
	if freeSpace <= 0 || itemCount == 0 {
		return 0
	}

	fs := float64(freeSpace)
	ic := float64(itemCount)

	switch justify {
	case JustifyEnd:
		return fs
	case JustifyCenter:
		return fs / 2.0
	case JustifySpaceAround:
		if itemCount > 0 {
			return fs / (ic * 2.0)
		}
		return 0
	case JustifySpaceEvenly:
		return fs / (ic + 1.0)
	default: // JustifyStart, JustifySpaceBetween
		return 0
	}
}

// calculateJustifySpacing returns the extra spacing between children
// based on the justify mode and available free space.
// Returns float64 to enable precise spacing that only rounds at the final stage.
func calculateJustifySpacing(justify Justify, freeSpace, itemCount int) float64 {
	if freeSpace <= 0 || itemCount <= 1 {
		return 0
	}

	fs := float64(freeSpace)
	ic := float64(itemCount)

	switch justify {
	case JustifySpaceBetween:
		return fs / (ic - 1.0)
	case JustifySpaceAround:
		return fs / ic
	case JustifySpaceEvenly:
		return fs / (ic + 1.0)
	default: // JustifyStart, JustifyEnd, JustifyCenter
		return 0
	}
}

// calculateAlignOffset returns the offset for positioning a child on the cross axis.
// Returns float64 to enable precise centering that only rounds at the final stage.
func calculateAlignOffset(align Align, crossSize, itemSize int) float64 {
	switch align {
	case AlignEnd:
		return float64(crossSize - itemSize)
	case AlignCenter:
		return float64(crossSize-itemSize) / 2.0
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
