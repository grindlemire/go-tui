package tui

// --- Implement Layoutable interface ---

// LayoutStyle returns the layout style properties for this element.
// If the element has a border, padding is increased to account for border width.
func (e *Element) LayoutStyle() LayoutStyle {
	style := e.style
	// Add padding for border (HR uses border field for line style, not actual border)
	if e.border != BorderNone && !e.hr {
		// Border takes 1 character on each side
		style.Padding.Top += 1
		style.Padding.Right += 1
		style.Padding.Bottom += 1
		style.Padding.Left += 1
	}
	return style
}

// LayoutChildren returns the children to be laid out.
// Hidden children are excluded from layout.
func (e *Element) LayoutChildren() []Layoutable {
	result := make([]Layoutable, 0, len(e.children))
	for _, child := range e.children {
		if !child.hidden && !child.overlay {
			result = append(result, child)
		}
	}
	return result
}

// SetLayout is called by the layout engine to store computed layout.
func (e *Element) SetLayout(l LayoutResult) {
	e.layout = l
}

// GetLayout returns the last computed layout.
func (e *Element) GetLayout() LayoutResult {
	return e.layout
}

// IsDirty returns whether this element needs layout recalculation.
func (e *Element) IsDirty() bool {
	return e.dirty
}

// SetDirty marks this element as needing recalculation or not.
func (e *Element) SetDirty(dirty bool) {
	e.dirty = dirty
}

// IsHR returns whether this element is a horizontal rule.
func (e *Element) IsHR() bool {
	return e.hr
}

// IntrinsicSize returns the natural content-based dimensions of this element.
// For text elements, returns the text width and height (1 line).
// For containers, returns the computed intrinsic size based on children.
func (e *Element) IntrinsicSize() (width, height int) {
	// HR has intrinsic height of 1, but 0 intrinsic width.
	// The 0 width is intentional - HR relies on AlignSelf=Stretch (set by WithHR)
	// to fill the container width, similar to how block elements work in CSS.
	if e.hr {
		return 0, 1
	}

	// Scrollable elements have 0 intrinsic size in their scroll direction.
	// They rely on flexGrow or explicit sizing to get space, then scroll their content.
	// This prevents content from pushing other elements out of the layout.
	if e.scrollMode != ScrollNone {
		// Return 0 for scrollable dimensions - the element will use available space
		return 0, 0
	}

	// Table elements compute intrinsic size from column widths and row heights
	if e.tag == "table" {
		w, h := TableIntrinsicSize(e)
		w += e.style.Padding.Horizontal()
		h += e.style.Padding.Vertical()
		if e.border != BorderNone {
			w += 2
			h += 2
		}
		return w, h
	}

	// Text content has explicit intrinsic size
	if e.text != "" {
		textWidth := stringWidth(e.text)
		textHeight := 1
		// Add padding to get the element's intrinsic size
		width = textWidth + e.style.Padding.Horizontal()
		height = textHeight + e.style.Padding.Vertical()
		// Add border if present (borders take 1 cell on each side)
		if e.border != BorderNone {
			width += 2
			height += 2
		}
		return width, height
	}

	// For containers without text, compute from children
	if len(e.children) == 0 {
		// Empty container has no intrinsic size
		return 0, 0
	}

	// Compute intrinsic size from children
	isRow := e.style.Direction == Row
	// Block mode forces column direction regardless of Direction setting
	if e.style.Display == DisplayBlock {
		isRow = false
	}
	var intrinsicW, intrinsicH int
	visibleIdx := 0

	for _, child := range e.children {
		if child.hidden || child.overlay {
			continue
		}
		childW, childH := child.IntrinsicSize()
		childStyle := child.LayoutStyle()
		marginH := childStyle.Margin.Horizontal()
		marginV := childStyle.Margin.Vertical()

		if isRow {
			intrinsicW += childW + marginH
			if childH+marginV > intrinsicH {
				intrinsicH = childH + marginV
			}
		} else {
			if childW+marginH > intrinsicW {
				intrinsicW = childW + marginH
			}
			intrinsicH += childH + marginV
		}

		// Add gap between visible children (not before first)
		if visibleIdx > 0 {
			if isRow {
				intrinsicW += e.style.Gap
			} else {
				intrinsicH += e.style.Gap
			}
		}
		visibleIdx++
	}

	// Add padding
	intrinsicW += e.style.Padding.Horizontal()
	intrinsicH += e.style.Padding.Vertical()

	// Add border if present
	if e.border != BorderNone {
		intrinsicW += 2
		intrinsicH += 2
	}

	return intrinsicW, intrinsicH
}

// HeightForWidth returns the height this element needs given an assigned width.
// For text elements with wrapping enabled, computes the wrapped text height.
// For column containers with Auto height, recursively computes from children.
// Scrollable elements and elements with explicit heights are not affected.
func (e *Element) HeightForWidth(width int) int {
	// Scrollable elements have fixed viewport — don't expand based on content.
	if e.scrollMode != ScrollNone {
		_, h := e.IntrinsicSize()
		return h
	}

	// Text elements with wrapping
	if e.text != "" && !e.noWrap {
		contentWidth := width - e.style.Padding.Horizontal()
		if e.border != BorderNone {
			contentWidth -= 2
		}
		if contentWidth <= 0 {
			h := e.style.Padding.Vertical()
			if e.border != BorderNone {
				h += 2
			}
			return h
		}
		lines := wrapText(e.text, contentWidth)
		h := len(lines) + e.style.Padding.Vertical()
		if e.border != BorderNone {
			h += 2
		}
		return h
	}

	// Column containers: recursively compute from children.
	// Each child gets the full content width (correct for AlignStretch + Auto width,
	// which is the default). This propagates text wrapping heights up the tree.
	isColumn := e.style.Direction == Column || e.style.Display == DisplayBlock
	if len(e.children) > 0 && isColumn {
		contentWidth := width - e.style.Padding.Horizontal()
		if e.border != BorderNone {
			contentWidth -= 2
		}
		totalH := 0
		visibleIdx := 0
		for _, child := range e.children {
			if child.hidden || child.overlay {
				continue
			}
			childH := child.HeightForWidth(contentWidth)
			totalH += childH
			if visibleIdx > 0 {
				totalH += e.style.Gap
			}
			visibleIdx++
		}
		totalH += e.style.Padding.Vertical()
		if e.border != BorderNone {
			totalH += 2
		}
		return totalH
	}

	// Row containers: recursively compute max child height.
	// Children share the width via flex, so we approximate by giving each child
	// its intrinsic width or a fair share. For text wrapping, the key case is
	// row children with explicit or flex-computed widths which Phase 3.5 handles
	// directly. Here we just find the max child height for the cross-axis.
	if len(e.children) > 0 {
		contentWidth := width - e.style.Padding.Horizontal()
		if e.border != BorderNone {
			contentWidth -= 2
		}
		maxH := 0
		for _, child := range e.children {
			if child.hidden || child.overlay {
				continue
			}
			// For row children, approximate: give each child the full width
			// (overestimate). Phase 3.5 handles precise per-child widths.
			childH := child.HeightForWidth(contentWidth)
			if childH > maxH {
				maxH = childH
			}
		}
		maxH += e.style.Padding.Vertical()
		if e.border != BorderNone {
			maxH += 2
		}
		return maxH
	}

	// Default: intrinsic height
	_, h := e.IntrinsicSize()
	return h
}

// Tag returns the element tag for layout dispatch.
func (e *Element) Tag() string {
	return e.tag
}

// Calculate computes layout for this Element and all descendants.
func (e *Element) Calculate(availableWidth, availableHeight int) {
	Calculate(e, availableWidth, availableHeight)
}

// Rect returns the computed border box.
func (e *Element) Rect() Rect {
	return e.layout.Rect
}

// ContentRect returns the computed content area.
func (e *Element) ContentRect() Rect {
	return e.layout.ContentRect
}

// MarkDirty marks this Element and ancestors as needing recalculation.
// Also marks the owning app as dirty so the app knows to re-render.
func (e *Element) MarkDirty() {
	for elem := e; elem != nil && !elem.dirty; elem = elem.parent {
		elem.dirty = true
	}
	if e.app != nil {
		e.app.MarkDirty()
	}
	// nil app: element-level dirty flags are set; app-level dirty
	// will be set when the element is attached via setAppRecursive.
}
