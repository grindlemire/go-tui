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

	// 1-3. Compute grid dimensions and column widths (shrunk to fit).
	numCols, colWidths := tableColumnWidths(rows, contentRect.Width)
	if numCols == 0 {
		return
	}

	// 4. Compute row heights from cells at their final column widths.
	rowHeights := tableRowHeights(rows, colWidths, contentRect.Height)

	// 5. Position rows top-to-bottom, cells left-to-right at column offsets.
	// Precompute column X offsets with 1-character gap between columns.
	colOffsets := make([]float64, numCols)
	offset := 0.0
	for ci := range numCols {
		colOffsets[ci] = offset
		offset += float64(colWidths[ci]) + 1 // +1 for inter-column gap
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

// tableColumnWidths computes the number of columns and the final column
// widths for a table. Each column starts at the max intrinsic (or explicit)
// cell width in that column; auto columns are then shrunk proportionally when
// the total exceeds availableWidth. Widths include cell padding.
func tableColumnWidths(rows []Layoutable, availableWidth int) (numCols int, colWidths []int) {
	for _, row := range rows {
		cells := row.LayoutChildren()
		if len(cells) > numCols {
			numCols = len(cells)
		}
	}
	if numCols == 0 {
		return 0, nil
	}

	colWidths = make([]int, numCols)
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
				cellWidth = cellStyle.Width.Resolve(availableWidth, intrW)
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

	// Shrink auto columns proportionally if total > available width.
	// Include 1-character gap between columns in total width.
	columnGap := max(0, numCols-1) // 1 char gap between each pair of columns
	totalWidth := columnGap
	for _, w := range colWidths {
		totalWidth += w
	}

	if totalWidth > availableWidth {
		// Compute total auto-column width for proportional shrinking
		totalAutoWidth := 0
		for ci, w := range colWidths {
			if colIsAuto[ci] {
				totalAutoWidth += w
			}
		}

		overflow := totalWidth - availableWidth
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

	return numCols, colWidths
}

// tableRowHeights computes per-row heights from cells measured at their final
// column widths. Explicit heights on a <tr> or cell win; auto-height cells
// grow to their wrapped height when the shrunk column forces text to wrap
// (issue #127).
func tableRowHeights(rows []Layoutable, colWidths []int, availableHeight int) []int {
	rowHeights := make([]int, len(rows))
	for ri, row := range rows {
		rowStyle := row.LayoutStyle()
		if !rowStyle.Height.IsAuto() {
			// Explicit row height overrides cell-based calculation
			rowHeights[ri] = rowStyle.Height.Resolve(availableHeight, 1)
			continue
		}

		cells := row.LayoutChildren()
		maxH := 1 // minimum row height is 1
		for ci, cell := range cells {
			cellStyle := cell.LayoutStyle()
			_, intrH := cell.IntrinsicSize()

			var cellHeight int
			if !cellStyle.Height.IsAuto() {
				cellHeight = cellStyle.Height.Resolve(availableHeight, intrH)
			} else if ci < len(colWidths) {
				// HeightForWidth equals intrH when nothing wraps, so this
				// only grows rows whose cells wrap at the shrunk width.
				cellHeight = max(intrH, cell.HeightForWidth(colWidths[ci]))
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
	return rowHeights
}

// TableHeightForWidth measures the content height a table needs at the given
// content width: column widths are resolved (including proportional
// shrinking) and rows are measured at those final widths. availableHeight is
// indefinite in this context, so explicit percent heights resolve against 0,
// matching TableIntrinsicSize.
func TableHeightForWidth(table Layoutable, contentWidth int) int {
	rows := table.LayoutChildren()
	numCols, colWidths := tableColumnWidths(rows, contentWidth)
	if numCols == 0 {
		return 0
	}
	total := 0
	for _, h := range tableRowHeights(rows, colWidths, 0) {
		total += h
	}
	return total
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

	// Sum column widths + inter-column gaps
	for _, w := range colWidths {
		width += w
	}
	if numCols > 1 {
		width += numCols - 1 // 1 char gap between each pair of columns
	}

	return width, height
}
