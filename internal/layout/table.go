package layout

import "math"

// layoutTable performs table-specific layout for a <table> element.
// It arranges children as rows (tr) containing cells (td/th) in a grid,
// computing column widths from the maximum intrinsic width of cells in each column
// and row heights from the maximum intrinsic height of cells in each row.
func layoutTable(table Layoutable, contentRect Rect, parentAbsX, parentAbsY float64) {
	rows := table.LayoutChildren()
	if len(rows) == 0 {
		return
	}

	// 1. Collect grid dimensions: determine number of columns and gather cells.
	numCols := 0
	for _, row := range rows {
		cells := row.LayoutChildren()
		if len(cells) > numCols {
			numCols = len(cells)
		}
	}
	if numCols == 0 {
		return
	}

	// 2. Compute column widths: max intrinsic width per column.
	// An explicit (non-Auto) width on a cell overrides its intrinsic width.
	colWidths := make([]int, numCols)
	colIsAuto := make([]bool, numCols) // track which columns are auto-sized
	for i := range colIsAuto {
		colIsAuto[i] = true
	}

	for _, row := range rows {
		cells := row.LayoutChildren()
		for ci, cell := range cells {
			cellStyle := cell.LayoutStyle()
			intrW, _ := cell.IntrinsicSize()

			var cellWidth int
			if !cellStyle.Width.IsAuto() {
				// Explicit width overrides intrinsic
				cellWidth = cellStyle.Width.Resolve(contentRect.Width, intrW)
				colIsAuto[ci] = false
			} else {
				cellWidth = intrW
			}

			// Include cell padding in column width calculation
			cellWidth += cellStyle.Padding.Horizontal()

			if cellWidth > colWidths[ci] {
				colWidths[ci] = cellWidth
			}
		}
	}

	// 3. Shrink auto columns proportionally if total > available width.
	totalWidth := 0
	for _, w := range colWidths {
		totalWidth += w
	}

	if totalWidth > contentRect.Width {
		// Compute total auto-column width for proportional shrinking
		totalAutoWidth := 0
		for ci, w := range colWidths {
			if colIsAuto[ci] {
				totalAutoWidth += w
			}
		}

		overflow := totalWidth - contentRect.Width
		if totalAutoWidth > 0 && overflow > 0 {
			// Shrink auto columns proportionally
			shrunk := 0
			lastAutoCol := -1
			for ci := range colWidths {
				if colIsAuto[ci] {
					lastAutoCol = ci
				}
			}

			for ci := range colWidths {
				if colIsAuto[ci] {
					reduction := int(float64(overflow) * float64(colWidths[ci]) / float64(totalAutoWidth))
					if ci == lastAutoCol {
						// Give the remainder to the last auto column to avoid rounding errors
						reduction = overflow - shrunk
					}
					colWidths[ci] = max(1, colWidths[ci]-reduction)
					shrunk += reduction
				}
			}
		}
	}

	// 4. Compute row heights: max intrinsic height per row.
	// Explicit h-N on <tr> overrides the computed max.
	rowHeights := make([]int, len(rows))
	for ri, row := range rows {
		rowStyle := row.LayoutStyle()
		if !rowStyle.Height.IsAuto() {
			// Explicit row height overrides cell-based calculation
			rowHeights[ri] = rowStyle.Height.Resolve(contentRect.Height, 1)
			continue
		}

		cells := row.LayoutChildren()
		maxH := 1 // minimum row height is 1
		for _, cell := range cells {
			cellStyle := cell.LayoutStyle()
			_, intrH := cell.IntrinsicSize()

			var cellHeight int
			if !cellStyle.Height.IsAuto() {
				cellHeight = cellStyle.Height.Resolve(contentRect.Height, intrH)
			} else {
				cellHeight = intrH
			}

			// Include cell padding in row height calculation
			cellHeight += cellStyle.Padding.Vertical()

			if cellHeight > maxH {
				maxH = cellHeight
			}
		}
		rowHeights[ri] = maxH
	}

	// 5. Position rows top-to-bottom, cells left-to-right at column offsets.
	// Precompute column X offsets.
	colOffsets := make([]float64, numCols)
	offset := 0.0
	for ci := range numCols {
		colOffsets[ci] = offset
		offset += float64(colWidths[ci])
	}

	rowAbsY := parentAbsY
	for ri, row := range rows {
		rowH := rowHeights[ri]

		// Set row layout
		rowAbsX := parentAbsX
		rowRect := Rect{
			X:      int(math.Round(rowAbsX)),
			Y:      int(math.Round(rowAbsY)),
			Width:  contentRect.Width,
			Height: rowH,
		}
		row.SetLayout(Layout{
			Rect:        rowRect,
			ContentRect: rowRect, // rows have no padding of their own
			AbsoluteX:   rowAbsX,
			AbsoluteY:   rowAbsY,
		})
		row.SetDirty(false)

		// Position cells within this row
		cells := row.LayoutChildren()
		for ci, cell := range cells {
			cellW := colWidths[ci]
			cellAbsX := parentAbsX + colOffsets[ci]
			cellAbsY := rowAbsY

			cellStyle := cell.LayoutStyle()

			// Border box for the cell
			cellBorderBox := Rect{
				X:      int(math.Round(cellAbsX)),
				Y:      int(math.Round(cellAbsY)),
				Width:  cellW,
				Height: rowH,
			}

			// Content rect: border box minus padding
			cellContentAbsX := cellAbsX + float64(cellStyle.Padding.Left)
			cellContentAbsY := cellAbsY + float64(cellStyle.Padding.Top)
			cellContentRect := Rect{
				X:      int(math.Round(cellContentAbsX)),
				Y:      int(math.Round(cellContentAbsY)),
				Width:  cellW - cellStyle.Padding.Horizontal(),
				Height: rowH - cellStyle.Padding.Vertical(),
			}

			// Clamp content dimensions to non-negative
			if cellContentRect.Width < 0 {
				cellContentRect.Width = 0
			}
			if cellContentRect.Height < 0 {
				cellContentRect.Height = 0
			}

			cell.SetLayout(Layout{
				Rect:        cellBorderBox,
				ContentRect: cellContentRect,
				AbsoluteX:   cellAbsX,
				AbsoluteY:   cellAbsY,
			})
			cell.SetDirty(false)

			// 6. Recurse into cell children using the flex layout
			cellChildren := cell.LayoutChildren()
			if len(cellChildren) > 0 {
				layoutChildren(cell, cellContentRect, cellContentAbsX, cellContentAbsY)
			}
		}

		rowAbsY += float64(rowH)
	}
}

// TableIntrinsicSize computes the intrinsic size of a table.
// Width = sum of max column widths, Height = sum of max row heights.
func TableIntrinsicSize(table Layoutable) (width, height int) {
	rows := table.LayoutChildren()
	if len(rows) == 0 {
		return 0, 0
	}

	// Determine number of columns
	numCols := 0
	for _, row := range rows {
		cells := row.LayoutChildren()
		if len(cells) > numCols {
			numCols = len(cells)
		}
	}
	if numCols == 0 {
		return 0, 0
	}

	// Compute column widths (max intrinsic width per column)
	colWidths := make([]int, numCols)
	for _, row := range rows {
		cells := row.LayoutChildren()
		for ci, cell := range cells {
			cellStyle := cell.LayoutStyle()
			intrW, _ := cell.IntrinsicSize()

			var cellWidth int
			if !cellStyle.Width.IsAuto() {
				cellWidth = cellStyle.Width.Resolve(0, intrW)
			} else {
				cellWidth = intrW
			}
			cellWidth += cellStyle.Padding.Horizontal()

			if cellWidth > colWidths[ci] {
				colWidths[ci] = cellWidth
			}
		}
	}

	// Compute row heights (max intrinsic height per row)
	for ri, row := range rows {
		cells := row.LayoutChildren()
		maxH := 1 // minimum row height is 1
		for _, cell := range cells {
			cellStyle := cell.LayoutStyle()
			_, intrH := cell.IntrinsicSize()

			var cellHeight int
			if !cellStyle.Height.IsAuto() {
				cellHeight = cellStyle.Height.Resolve(0, intrH)
			} else {
				cellHeight = intrH
			}
			cellHeight += cellStyle.Padding.Vertical()

			if cellHeight > maxH {
				maxH = cellHeight
			}
		}
		height += maxH
		_ = ri
	}

	// Sum column widths
	for _, w := range colWidths {
		width += w
	}

	return width, height
}
