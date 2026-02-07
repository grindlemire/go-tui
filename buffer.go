package tui

import "strings"

// Buffer is a double-buffered 2D grid of cells.
// Writes go to the back buffer; Flush() computes the diff and swaps buffers.
type Buffer struct {
	front  []Cell // Currently displayed state
	back   []Cell // State being built
	width  int
	height int
}

// CellChange represents a single cell that differs between front and back buffers.
type CellChange struct {
	X, Y int
	Cell Cell
}

// NewBuffer creates a new double-buffered grid of the specified dimensions.
// Both buffers are initialized with spaces and default styling.
func NewBuffer(width, height int) *Buffer {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	size := width * height
	front := make([]Cell, size)
	back := make([]Cell, size)

	// Initialize with spaces (default style is zero value)
	defaultCell := NewCell(' ', NewStyle())
	for i := range front {
		front[i] = defaultCell
		back[i] = defaultCell
	}

	return &Buffer{
		front:  front,
		back:   back,
		width:  width,
		height: height,
	}
}

// Width returns the buffer width in columns.
func (b *Buffer) Width() int {
	return b.width
}

// Height returns the buffer height in rows.
func (b *Buffer) Height() int {
	return b.height
}

// Size returns the buffer dimensions (width, height).
func (b *Buffer) Size() (width, height int) {
	return b.width, b.height
}

// Rect returns the buffer bounds as a Rect starting at (0, 0).
func (b *Buffer) Rect() Rect {
	return NewRect(0, 0, b.width, b.height)
}

// idx converts (x, y) coordinates to a flat index.
// Returns -1 if out of bounds.
func (b *Buffer) idx(x, y int) int {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return -1
	}
	return y*b.width + x
}

// Cell returns the cell at position (x, y) from the back buffer.
// Returns an empty Cell if the position is out of bounds.
func (b *Buffer) Cell(x, y int) Cell {
	idx := b.idx(x, y)
	if idx < 0 {
		return Cell{}
	}
	return b.back[idx]
}

// SetCell sets the cell at position (x, y) in the back buffer.
// Does nothing if the position is out of bounds.
func (b *Buffer) SetCell(x, y int, c Cell) {
	idx := b.idx(x, y)
	if idx < 0 {
		return
	}
	b.back[idx] = c
}

// SetRune sets a rune at position (x, y) with the given style.
// Handles wide characters by setting continuation cells.
// Properly clears overlapped wide characters.
func (b *Buffer) SetRune(x, y int, r rune, style Style) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}

	width := RuneWidth(r)
	currentCell := b.Cell(x, y)

	// If target position is a continuation cell, clear the originating wide char
	if currentCell.IsContinuation() {
		b.clearWideCharAt(x, y)
	}

	// If target position is the START of a wide character, clear its continuation
	if currentCell.Width == 2 && x+1 < b.width {
		b.SetCell(x+1, y, NewCell(' ', NewStyle()))
	}

	// If placing a wide char would overlap an existing wide char at x+1, clear it
	if width == 2 && x+1 < b.width {
		next := b.Cell(x+1, y)
		// If next cell is the start of a wide char (width 2), clear it and its continuation
		if next.Width == 2 {
			b.clearWideCharAt(x+1, y)
		}
		// If next cell is a continuation, clear its originating wide char
		if next.IsContinuation() {
			b.clearWideCharAt(x+1, y)
		}
	}

	// Handle edge case: wide char at last column - can't fit, skip it
	if width == 2 && x+1 >= b.width {
		// Place a space instead since the wide char can't fit
		b.SetCell(x, y, NewCell(' ', style))
		return
	}

	// Set the primary cell
	b.SetCell(x, y, NewCellWithWidth(r, style, uint8(width)))

	// Set continuation cell for wide characters
	if width == 2 {
		b.SetCell(x+1, y, NewCellWithWidth(0, style, 0))
	}
}

// clearWideCharAt clears a wide character that includes position (x, y).
// If (x, y) is a continuation cell, finds and clears the originating cell.
// If (x, y) is a wide char start, clears it and its continuation.
func (b *Buffer) clearWideCharAt(x, y int) {
	cell := b.Cell(x, y)
	defaultCell := NewCell(' ', NewStyle())

	if cell.IsContinuation() {
		// This is a continuation - the wide char starts at x-1
		if x > 0 {
			b.SetCell(x-1, y, defaultCell)
		}
		b.SetCell(x, y, defaultCell)
	} else if cell.Width == 2 {
		// This is the start of a wide char
		b.SetCell(x, y, defaultCell)
		if x+1 < b.width {
			b.SetCell(x+1, y, defaultCell)
		}
	}
}

// SetString writes a string starting at position (x, y) with the given style.
// Returns the total display width consumed (handles wide characters).
// Stops at buffer edge without wrapping.
func (b *Buffer) SetString(x, y int, s string, style Style) int {
	if y < 0 || y >= b.height {
		return 0
	}

	totalWidth := 0
	curX := x

	for _, r := range s {
		if curX >= b.width {
			break
		}
		if curX < 0 {
			// Skip characters before the visible area
			curX += RuneWidth(r)
			continue
		}

		width := RuneWidth(r)

		// Check if wide char fits
		if width == 2 && curX+1 >= b.width {
			// Wide char doesn't fit, stop here
			break
		}

		b.SetRune(curX, y, r, style)
		curX += width
		totalWidth += width
	}

	return totalWidth
}

// SetStringClipped writes a string clipped to a rectangle.
// Characters outside clipRect are not rendered.
// Returns the total display width of rendered characters.
func (b *Buffer) SetStringClipped(x, y int, s string, style Style, clipRect Rect) int {
	if y < clipRect.Y || y >= clipRect.Bottom() {
		return 0
	}

	totalWidth := 0
	curX := x

	for _, r := range s {
		width := RuneWidth(r)

		// Skip if entirely before clip region
		if curX+width <= clipRect.X {
			curX += width
			continue
		}

		// Stop if past clip region
		if curX >= clipRect.Right() {
			break
		}

		// Render if within clip (also check buffer bounds)
		if curX >= clipRect.X && curX < clipRect.Right() {
			// For wide characters, ensure both cells fit in clip region
			if width == 2 && curX+1 >= clipRect.Right() {
				// Wide char doesn't fit, skip it
				curX += width
				continue
			}
			b.SetRune(curX, y, r, style)
			totalWidth += width
		}

		curX += width
	}

	return totalWidth
}

// Fill fills a rectangle with the given rune and style.
// Handles wide characters appropriately.
func (b *Buffer) Fill(rect Rect, r rune, style Style) {
	// Intersect with buffer bounds
	rect = rect.Intersect(b.Rect())
	if rect.IsEmpty() {
		return
	}

	width := RuneWidth(r)

	for y := rect.Y; y < rect.Bottom(); y++ {
		for x := rect.X; x < rect.Right(); {
			if width == 2 && x+1 >= rect.Right() {
				// Wide char doesn't fit in remaining space, fill with space
				b.SetRune(x, y, ' ', style)
				x++
			} else {
				b.SetRune(x, y, r, style)
				x += width
			}
		}
	}
}

// SetStringGradient writes a string with a gradient applied per-character.
// The gradient is applied horizontally along the string.
// Returns the total display width consumed (handles wide characters).
func (b *Buffer) SetStringGradient(x, y int, s string, g Gradient, baseStyle Style) int {
	if y < 0 || y >= b.height {
		return 0
	}

	runes := []rune(s)
	if len(runes) == 0 {
		return 0
	}

	totalWidth := 0
	curX := x

	for i, r := range runes {
		if curX >= b.width {
			break
		}
		if curX < 0 {
			// Skip characters before the visible area
			curX += RuneWidth(r)
			continue
		}

		width := RuneWidth(r)

		// Check if wide char fits
		if width == 2 && curX+1 >= b.width {
			// Wide char doesn't fit, stop here
			break
		}

		// Calculate gradient position t in [0, 1]
		t := float64(i) / float64(len(runes)-1)
		if len(runes) == 1 {
			t = 0
		}

		// Get gradient color and apply to style
		gradColor := g.At(t)
		style := baseStyle
		style.Fg = gradColor

		b.SetRune(curX, y, r, style)
		curX += width
		totalWidth += width
	}

	return totalWidth
}

// FillGradient fills a rectangle with a gradient background.
// The gradient direction determines how it's applied:
// - Horizontal: left to right
// - Vertical: top to bottom
// - DiagonalDown: top-left to bottom-right
// - DiagonalUp: bottom-left to top-right
func (b *Buffer) FillGradient(rect Rect, r rune, g Gradient, baseStyle Style) {
	// Intersect with buffer bounds
	rect = rect.Intersect(b.Rect())
	if rect.IsEmpty() {
		return
	}

	width := RuneWidth(r)
	rectWidth := float64(rect.Width)
	rectHeight := float64(rect.Height)

	// Avoid division by zero
	if rectWidth <= 0 {
		rectWidth = 1
	}
	if rectHeight <= 0 {
		rectHeight = 1
	}

	for y := rect.Y; y < rect.Bottom(); y++ {
		for x := rect.X; x < rect.Right(); {
			if width == 2 && x+1 >= rect.Right() {
				// Wide char doesn't fit in remaining space, fill with space
				style := baseStyle
				var t float64
				switch g.Direction {
				case GradientHorizontal:
					t = float64(x-rect.X) / rectWidth
				case GradientVertical:
					t = float64(y-rect.Y) / rectHeight
				case GradientDiagonalDown:
					tx := float64(x-rect.X) / rectWidth
					ty := float64(y-rect.Y) / rectHeight
					t = (tx + ty) / 2
				case GradientDiagonalUp:
					tx := float64(x-rect.X) / rectWidth
					ty := float64(rect.Bottom()-1-y-rect.Y) / rectHeight
					t = (tx + ty) / 2
				default:
					t = float64(x-rect.X) / rectWidth
				}
				style.Bg = g.At(t)
				b.SetRune(x, y, ' ', style)
				x++
			} else {
				// Calculate gradient position based on direction
				var t float64
				switch g.Direction {
				case GradientHorizontal:
					t = float64(x-rect.X) / rectWidth
				case GradientVertical:
					t = float64(y-rect.Y) / rectHeight
				case GradientDiagonalDown:
					tx := float64(x-rect.X) / rectWidth
					ty := float64(y-rect.Y) / rectHeight
					t = (tx + ty) / 2
				case GradientDiagonalUp:
					tx := float64(x-rect.X) / rectWidth
					ty := float64(rect.Bottom()-1-y-rect.Y) / rectHeight
					t = (tx + ty) / 2
				default:
					t = float64(x-rect.X) / rectWidth
				}

				// Get gradient color and apply to style
				gradColor := g.At(t)
				style := baseStyle
				style.Bg = gradColor

				b.SetRune(x, y, r, style)
				x += width
			}
		}
	}
}

// Clear clears the entire back buffer to spaces with default style.
func (b *Buffer) Clear() {
	b.ClearRect(b.Rect())
}

// ClearRect clears a rectangular region to spaces with default style.
func (b *Buffer) ClearRect(rect Rect) {
	// Intersect with buffer bounds
	rect = rect.Intersect(b.Rect())
	if rect.IsEmpty() {
		return
	}

	defaultCell := NewCell(' ', NewStyle())

	for y := rect.Y; y < rect.Bottom(); y++ {
		for x := rect.X; x < rect.Right(); x++ {
			// First, handle any wide character cleanup at the edges
			cell := b.Cell(x, y)
			if cell.IsContinuation() && x == rect.X {
				// Clearing starts at a continuation - clear the originating char too
				if x > 0 {
					b.SetCell(x-1, y, defaultCell)
				}
			}
			if cell.Width == 2 && x+1 == rect.Right() {
				// Clearing ends at a wide char - also clear the continuation
				if x+1 < b.width {
					b.SetCell(x+1, y, defaultCell)
				}
			}
			b.SetCell(x, y, defaultCell)
		}
	}
}

// Diff returns all cells that changed between front and back buffers.
// Cells are returned in row-major order (top-to-bottom, left-to-right)
// which optimizes terminal output by minimizing cursor moves.
func (b *Buffer) Diff() []CellChange {
	changes := make([]CellChange, 0, b.width) // Pre-allocate one row
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			idx := y*b.width + x
			if !b.back[idx].Equal(b.front[idx]) {
				changes = append(changes, CellChange{X: x, Y: y, Cell: b.back[idx]})
			}
		}
	}
	return changes
}

// Swap copies the back buffer to the front buffer.
// Call this after flushing changes to the terminal.
func (b *Buffer) Swap() {
	copy(b.front, b.back)
}

// String renders the back buffer to a string for debugging.
// Each row is separated by a newline. Continuation cells (from wide characters) are skipped.
func (b *Buffer) String() string {
	var sb strings.Builder
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			cell := b.back[y*b.width+x]
			if cell.IsContinuation() {
				continue // Skip continuation cells
			}
			if cell.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune)
			}
		}
		if y < b.height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// StringTrimmed returns the back buffer content with trailing spaces removed from each line.
func (b *Buffer) StringTrimmed() string {
	var sb strings.Builder
	for y := 0; y < b.height; y++ {
		var line strings.Builder
		for x := 0; x < b.width; x++ {
			cell := b.back[y*b.width+x]
			if cell.IsContinuation() {
				continue
			}
			if cell.Rune == 0 {
				line.WriteRune(' ')
			} else {
				line.WriteRune(cell.Rune)
			}
		}
		sb.WriteString(strings.TrimRight(line.String(), " "))
		if y < b.height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// Resize changes the buffer dimensions, preserving content where possible.
// Content in the overlapping region is preserved; new areas are cleared.
func (b *Buffer) Resize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	if width == b.width && height == b.height {
		return
	}

	newSize := width * height
	newFront := make([]Cell, newSize)
	newBack := make([]Cell, newSize)

	// Initialize with spaces
	defaultCell := NewCell(' ', NewStyle())
	for i := range newFront {
		newFront[i] = defaultCell
		newBack[i] = defaultCell
	}

	// Copy overlapping content
	copyWidth := min(width, b.width)
	copyHeight := min(height, b.height)

	for y := 0; y < copyHeight; y++ {
		for x := 0; x < copyWidth; x++ {
			oldIdx := y*b.width + x
			newIdx := y*width + x
			newFront[newIdx] = b.front[oldIdx]
			newBack[newIdx] = b.back[oldIdx]
		}
	}

	b.front = newFront
	b.back = newBack
	b.width = width
	b.height = height
}
