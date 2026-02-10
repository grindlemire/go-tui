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
func (e *Element) LayoutChildren() []Layoutable {
	result := make([]Layoutable, len(e.children))
	for i, child := range e.children {
		result[i] = child
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
	var intrinsicW, intrinsicH int

	for i, child := range e.children {
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

		// Add gap between children (not before first)
		if i > 0 {
			if isRow {
				intrinsicW += e.style.Gap
			} else {
				intrinsicH += e.style.Gap
			}
		}
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
	// Signal to the owning app that UI needs re-rendering.
	if e.app != nil {
		e.app.MarkDirty()
		return
	}
	if app := DefaultApp(); app != nil {
		app.MarkDirty()
		return
	}
	panic("tui.Element.MarkDirty requires app context; ensure element is attached to an app")
}
