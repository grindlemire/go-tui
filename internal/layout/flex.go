package layout

import "math"

// flexItem holds intermediate calculation state for a child.
// This is stack-allocated per layout call, not stored on nodes.
// Positions are stored as float64 to enable precise centering calculations
// that only round at the final stage, preventing jitter during animation.
type flexItem struct {
	node             Layoutable
	baseSize         int
	mainSize         int
	crossSize        int
	mainPos          float64 // float to avoid centering jitter
	crossPos         float64 // float to avoid centering jitter
	grow             float64
	shrink           float64
	wrappedCrossSize int  // set by Phase 3.5 when text wrapping increases cross-axis height
	hasWrappedCross  bool // true if wrappedCrossSize is valid
}

// flexLine represents a single line of flex items in a wrapped layout.
type flexLine struct {
	startIdx  int     // index into items array (inclusive)
	endIdx    int     // index into items array (exclusive)
	crossSize int     // max cross size among items on this line
	crossPos  float64 // line's position on the cross axis
}

// breakIntoLines splits flex items into lines based on available main-axis space.
// Each item's baseSize (including margin) is used for line-break decisions.
// If mainSize is 0 or negative, all items go on one line.
func breakIntoLines(items []flexItem, mainSize, gap int) []flexLine {
	if len(items) == 0 {
		return nil
	}
	if mainSize <= 0 {
		return []flexLine{{startIdx: 0, endIdx: len(items)}}
	}

	var lines []flexLine
	lineStart := 0
	used := 0

	for i := range items {
		itemSize := items[i].baseSize
		gapCost := 0
		if i > lineStart {
			gapCost = gap
		}

		if used+gapCost+itemSize > mainSize && i > lineStart {
			lines = append(lines, flexLine{startIdx: lineStart, endIdx: i})
			lineStart = i
			used = itemSize
		} else {
			used += gapCost + itemSize
		}
	}
	// Flush the last line
	lines = append(lines, flexLine{startIdx: lineStart, endIdx: len(items)})

	return lines
}

// distributeLineMainAxis runs flex grow/shrink distribution (Phase 2),
// min/max constraints (Phase 3), and justify positioning (Phase 4)
// for a single line of items.
func distributeLineMainAxis(items []flexItem, mainSize, gap int, justify Justify, isRow bool) {
	lineItems := len(items)
	if lineItems == 0 {
		return
	}

	// Compute totals for this line
	totalFixed := 0
	totalGrow := 0.0
	totalShrink := 0.0
	for i := range items {
		totalFixed += items[i].baseSize
		totalGrow += items[i].grow
		totalShrink += items[i].shrink
	}

	totalGap := gap * max(0, lineItems-1)
	freeSpace := mainSize - totalFixed - totalGap

	// Phase 2: Distribute free space
	if freeSpace > 0 && totalGrow > 0 {
		for i := range items {
			if items[i].grow > 0 {
				extra := int(float64(freeSpace) * items[i].grow / totalGrow)
				items[i].mainSize = items[i].baseSize + extra
			} else {
				items[i].mainSize = items[i].baseSize
			}
		}
	} else if freeSpace < 0 && totalShrink > 0 {
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
		for i := range items {
			items[i].mainSize = items[i].baseSize
		}
		freeSpace = max(0, freeSpace)
	}

	// Phase 3: Apply min/max constraints
	for i := range items {
		child := items[i].node
		childStyle := child.LayoutStyle()
		childIntrinsicW, childIntrinsicH := child.IntrinsicSize()
		var intrinsicMain int
		if isRow {
			intrinsicMain = childIntrinsicW
		} else {
			intrinsicMain = childIntrinsicH
		}
		minMain := resolveMinMain(childStyle, isRow, mainSize, intrinsicMain)
		maxMain := resolveMaxMain(childStyle, isRow, mainSize)
		items[i].mainSize = clampFlex(items[i].mainSize, minMain, maxMain)
	}

	// Recalculate free space after constraints
	totalUsed := 0
	for i := range items {
		totalUsed += items[i].mainSize
	}
	freeSpace = mainSize - totalUsed - totalGap

	// Phase 4: Position children along main axis (justify)
	offset := calculateJustifyOffset(justify, freeSpace, lineItems)
	spacing := calculateJustifySpacing(justify, freeSpace, lineItems)

	for i := range items {
		items[i].mainPos = offset
		offset += float64(items[i].mainSize) + float64(gap) + spacing
	}
}

// recomputeTextWrapping runs Phase 3.5 for a slice of items.
// It calls HeightForWidth to determine if text wrapping changes cross-axis sizes.
func recomputeTextWrapping(items []flexItem, parentStyle Style, isRow bool, mainSize, crossSize int) {
	for i := range items {
		child := items[i].node
		childStyle := child.LayoutStyle()

		var childWidth int
		if isRow {
			childWidth = items[i].mainSize - childStyle.Margin.Horizontal()
		} else {
			align := parentStyle.AlignItems
			if childStyle.AlignSelf != nil {
				align = *childStyle.AlignSelf
			}
			crossStyleValue := childStyle.Width
			crossMargin := childStyle.Margin.Horizontal()
			if align == AlignStretch && crossStyleValue.IsAuto() {
				childWidth = crossSize - crossMargin
			} else if crossStyleValue.IsAuto() {
				childWidth, _ = child.IntrinsicSize()
			} else {
				childWidth = crossStyleValue.Resolve(crossSize-crossMargin, 0)
			}
		}

		wrappedHeight := child.HeightForWidth(childWidth)
		_, intrinsicH := child.IntrinsicSize()

		if wrappedHeight > intrinsicH {
			if isRow {
				items[i].wrappedCrossSize = wrappedHeight
				items[i].hasWrappedCross = true
			} else {
				mainMargin := childStyle.Margin.Vertical()
				newMainSize := wrappedHeight + mainMargin
				items[i].mainSize = clampFlex(newMainSize,
					resolveMinMain(childStyle, isRow, mainSize, wrappedHeight),
					resolveMaxMain(childStyle, isRow, mainSize))
			}
		}
	}
}

// calculateContentOffset returns the initial cross-axis offset for line distribution.
func calculateContentOffset(ac AlignContent, freeSpace, lineCount int) float64 {
	if freeSpace <= 0 || lineCount == 0 {
		return 0
	}

	fs := float64(freeSpace)
	lc := float64(lineCount)

	switch ac {
	case ContentEnd:
		return fs
	case ContentCenter:
		return fs / 2.0
	case ContentSpaceAround:
		if lineCount > 0 {
			return fs / (lc * 2.0)
		}
		return 0
	default: // ContentStart, ContentSpaceBetween, ContentStretch
		return 0
	}
}

// calculateContentSpacing returns the extra spacing between lines.
func calculateContentSpacing(ac AlignContent, freeSpace, lineCount int) float64 {
	if freeSpace <= 0 || lineCount <= 1 {
		return 0
	}

	fs := float64(freeSpace)
	lc := float64(lineCount)

	switch ac {
	case ContentSpaceBetween:
		return fs / (lc - 1.0)
	case ContentSpaceAround:
		return fs / lc
	default:
		return 0
	}
}

// layoutChildren arranges the children of a node within the given content rect.
// This implements the core flexbox algorithm with flex-wrap support.
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
	if style.Display == DisplayBlock {
		isRow = false
	}

	mainSize := contentRect.Width
	crossSize := contentRect.Height
	if !isRow {
		mainSize, crossSize = crossSize, mainSize
	}

	// Phase 1: Compute base sizes and flex factors
	items := make([]flexItem, len(children))
	for i, child := range children {
		item := &items[i]
		item.node = child

		childStyle := child.LayoutStyle()
		var mainMargin int
		if isRow {
			mainMargin = childStyle.Margin.Horizontal()
		} else {
			mainMargin = childStyle.Margin.Vertical()
		}

		childIntrinsicW, childIntrinsicH := child.IntrinsicSize()
		if isRow {
			item.baseSize = childStyle.Width.Resolve(mainSize, childIntrinsicW) + mainMargin
		} else {
			item.baseSize = childStyle.Height.Resolve(mainSize, childIntrinsicH) + mainMargin
		}

		item.grow = childStyle.FlexGrow
		item.shrink = childStyle.FlexShrink
	}

	// Determine lines
	var lines []flexLine
	if style.FlexWrap == WrapNone {
		lines = []flexLine{{startIdx: 0, endIdx: len(items)}}
	} else {
		lines = breakIntoLines(items, mainSize, style.Gap)
	}

	// Reverse line order for WrapReverse
	if style.FlexWrap == WrapReverse {
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
	}

	// Per-line: distribute main axis, apply text wrapping, compute cross sizes
	for l := range lines {
		line := &lines[l]
		lineItems := items[line.startIdx:line.endIdx]

		// Phases 2-4: main axis distribution and positioning
		distributeLineMainAxis(lineItems, mainSize, style.Gap, style.JustifyContent, isRow)

		// Phase 3.5: text wrapping recomputation
		recomputeTextWrapping(lineItems, style, isRow, mainSize, crossSize)

		// Re-position main axis after text wrapping may have changed sizes (column mode)
		if !isRow {
			totalUsed := 0
			for i := range lineItems {
				totalUsed += lineItems[i].mainSize
			}
			totalGap := style.Gap * max(0, len(lineItems)-1)
			freeSpace := mainSize - totalUsed - totalGap
			offset := calculateJustifyOffset(style.JustifyContent, freeSpace, len(lineItems))
			spacing := calculateJustifySpacing(style.JustifyContent, freeSpace, len(lineItems))
			for i := range lineItems {
				lineItems[i].mainPos = offset
				offset += float64(lineItems[i].mainSize) + float64(style.Gap) + spacing
			}
		}

		// Compute line cross size (max of all items' cross sizes on this line)
		maxCross := 0
		for i := range lineItems {
			child := lineItems[i].node
			childStyle := child.LayoutStyle()

			var crossIntrinsic int
			childIntrinsicW, childIntrinsicH := child.IntrinsicSize()
			if isRow {
				if lineItems[i].hasWrappedCross {
					crossIntrinsic = lineItems[i].wrappedCrossSize
				} else {
					crossIntrinsic = childIntrinsicH
				}
				crossIntrinsic += childStyle.Margin.Vertical()
			} else {
				crossIntrinsic = childIntrinsicW + childStyle.Margin.Horizontal()
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

			itemCross := crossIntrinsic
			if !crossStyleValue.IsAuto() {
				itemCross = crossStyleValue.Resolve(crossSize-crossMargin, 0) + crossMargin
			}
			if itemCross > maxCross {
				maxCross = itemCross
			}
		}
		line.crossSize = maxCross
	}

	// Phase 5.5: Distribute lines along cross axis
	if style.FlexWrap != WrapNone && len(lines) > 1 {
		totalLineCross := 0
		for l := range lines {
			totalLineCross += lines[l].crossSize
		}
		freeCrossSpace := crossSize - totalLineCross

		// ContentStretch: distribute extra space equally among lines
		if style.AlignContent == ContentStretch && freeCrossSpace > 0 {
			extra := freeCrossSpace / len(lines)
			remainder := freeCrossSpace % len(lines)
			for l := range lines {
				lines[l].crossSize += extra
				if l < remainder {
					lines[l].crossSize++
				}
			}
			freeCrossSpace = 0
		}

		offset := calculateContentOffset(style.AlignContent, freeCrossSpace, len(lines))
		spacing := calculateContentSpacing(style.AlignContent, freeCrossSpace, len(lines))

		for l := range lines {
			lines[l].crossPos = offset
			offset += float64(lines[l].crossSize) + spacing
		}
	} else if len(lines) == 1 {
		// Single line: use full cross size
		lines[0].crossSize = crossSize
		lines[0].crossPos = 0
	}

	// Phase 5 + 6: Cross-axis alignment and rect conversion per line
	for l := range lines {
		line := &lines[l]
		lineItems := items[line.startIdx:line.endIdx]
		lineChildren := children[line.startIdx:line.endIdx]
		lineCross := line.crossSize

		// Phase 5: Cross-axis sizing and alignment within the line
		for i := range lineItems {
			child := lineItems[i].node
			childStyle := child.LayoutStyle()
			align := style.AlignItems
			if childStyle.AlignSelf != nil {
				align = *childStyle.AlignSelf
			}

			var crossStyleValue Value
			var crossMargin int
			var crossIntrinsic int
			childIntrinsicW, childIntrinsicH := child.IntrinsicSize()
			if isRow {
				crossStyleValue = childStyle.Height
				crossMargin = childStyle.Margin.Vertical()
				if lineItems[i].hasWrappedCross {
					crossIntrinsic = lineItems[i].wrappedCrossSize
				} else {
					crossIntrinsic = childIntrinsicH
				}
			} else {
				crossStyleValue = childStyle.Width
				crossMargin = childStyle.Margin.Horizontal()
				crossIntrinsic = childIntrinsicW
			}

			availableCross := lineCross - crossMargin

			if align == AlignStretch && crossStyleValue.IsAuto() {
				lineItems[i].crossSize = availableCross + crossMargin
				lineItems[i].crossPos = 0
			} else {
				var contentCross int
				if crossStyleValue.IsAuto() {
					contentCross = min(crossIntrinsic, availableCross)
				} else {
					contentCross = crossStyleValue.Resolve(availableCross, crossIntrinsic)
				}
				lineItems[i].crossSize = contentCross + crossMargin
				lineItems[i].crossPos = calculateAlignOffset(align, lineCross, lineItems[i].crossSize)
			}
		}

		// Phase 6: Convert to rects and recurse
		for i := range lineItems {
			child := lineChildren[i]
			childStyle := child.LayoutStyle()

			var childAbsX, childAbsY float64
			if isRow {
				childAbsX = parentAbsX + lineItems[i].mainPos
				childAbsY = parentAbsY + line.crossPos + lineItems[i].crossPos
			} else {
				childAbsX = parentAbsX + line.crossPos + lineItems[i].crossPos
				childAbsY = parentAbsY + lineItems[i].mainPos
			}

			var slot Rect
			if isRow {
				slot = Rect{
					X:      int(math.Round(childAbsX)),
					Y:      int(math.Round(childAbsY)),
					Width:  lineItems[i].mainSize,
					Height: lineItems[i].crossSize,
				}
			} else {
				slot = Rect{
					X:      int(math.Round(childAbsX)),
					Y:      int(math.Round(childAbsY)),
					Width:  lineItems[i].crossSize,
					Height: lineItems[i].mainSize,
				}
			}

			if isRow {
				childAbsX += float64(childStyle.Margin.Left)
				childAbsY += float64(childStyle.Margin.Top)
			} else {
				childAbsX += float64(childStyle.Margin.Left)
				childAbsY += float64(childStyle.Margin.Top)
			}
			childBorderBox := slot.Inset(childStyle.Margin)

			child.SetDirty(true)
			calculateNode(child, childBorderBox, childAbsX, childAbsY)
		}
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
// When the min value is Auto, uses the child's intrinsic size as the minimum,
// matching CSS flexbox min-width:auto / min-height:auto behavior.
func resolveMinMain(style Style, isRow bool, available int, intrinsic int) int {
	if isRow {
		return style.MinWidth.Resolve(available, intrinsic)
	}
	return style.MinHeight.Resolve(available, intrinsic)
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
