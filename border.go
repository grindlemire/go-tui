package tui

// BorderStyle represents different styles of box borders.
type BorderStyle int

const (
	// BorderNone indicates no border should be drawn.
	BorderNone BorderStyle = iota
	// BorderSingle uses single-line box-drawing characters (─, │, ┌, etc.)
	BorderSingle
	// BorderDouble uses double-line box-drawing characters (═, ║, ╔, etc.)
	BorderDouble
	// BorderRounded uses rounded corner characters (─, │, ╭, ╮, ╰, ╯)
	BorderRounded
	// BorderThick uses thick/heavy box-drawing characters (━, ┃, ┏, etc.)
	BorderThick
)

// Borders represents border style values for the four sides of a box.
type Borders struct {
	Top, Right, Bottom, Left BorderStyle
}

// BorderChars holds the characters used to draw a box border.
type BorderChars struct {
	TopLeft     rune
	Top         rune
	TopRight    rune
	Left        rune
	Right       rune
	BottomLeft  rune
	Bottom      rune
	BottomRight rune
}

func (c *BorderChars) sideChars(side rune, style BorderStyle, styles *Borders) {
	if style == BorderNone {
		return
	}

	switch side {
	case 't':
		switch styles.Top {
		case BorderSingle:
			c.Top = '─'
			if styles.Left != BorderNone {
				c.TopLeft = '┌'
			}
			if styles.Right != BorderNone {
				c.TopRight = '┐'
			}
		case BorderDouble:
			c.Top = '═'
			if styles.Left != BorderNone {
				c.TopLeft = '╔'
			}
			if styles.Right != BorderNone {
				c.TopRight = '╗'
			}
		case BorderRounded:
			c.Top = '─'
			if styles.Left != BorderNone {
				c.TopLeft = '╭'
			}
			if styles.Right != BorderNone {
				c.TopRight = '╮'
			}
		case BorderThick:
			c.Top = '━'
			if styles.Left != BorderNone {
				c.TopLeft = '┏'
			}
			if styles.Right != BorderNone {
				c.TopRight = '┓'
			}
		}
	case 'r':
		switch styles.Right {
		case BorderSingle:
			c.Right = '│'
		case BorderDouble:
			c.Right = '║'
		case BorderRounded:
			c.Right = '│'
		case BorderThick:
			c.Right = '┃'
		}
	case 'b':
		switch styles.Bottom {
		case BorderSingle:
			c.Bottom = '─'
			if styles.Left != BorderNone {
				c.BottomLeft = '└'
			}
			if styles.Right != BorderNone {
				c.BottomRight = '┘'
			}
		case BorderDouble:
			c.Bottom = '═'
			if styles.Left != BorderNone {
				c.BottomLeft = '╚'
			}
			if styles.Right != BorderNone {
				c.BottomRight = '╝'
			}
		case BorderRounded:
			c.Bottom = '─'
			if styles.Left != BorderNone {
				c.BottomLeft = '╰'
			}
			if styles.Right != BorderNone {
				c.BottomRight = '╯'
			}
		case BorderThick:
			c.Bottom = '━'
			if styles.Left != BorderNone {
				c.BottomLeft = '┗'
			}
			if styles.Right != BorderNone {
				c.BottomRight = '┛'
			}
		}
	case 'l':
		switch styles.Left {
		case BorderSingle:
			c.Left = '│'
		case BorderDouble:
			c.Left = '║'
		case BorderRounded:
			c.Left = '│'
		case BorderThick:
			c.Left = '┃'
		}
	}
}

// Chars returns the box-drawing characters for their border styles.
func (b Borders) Chars() BorderChars {
	c := &BorderChars{
		TopLeft:     ' ',
		Top:         ' ',
		TopRight:    ' ',
		Left:        ' ',
		Right:       ' ',
		BottomLeft:  ' ',
		Bottom:      ' ',
		BottomRight: ' ',
	}

	c.sideChars('t', b.Top, &b)
	c.sideChars('r', b.Right, &b)
	c.sideChars('b', b.Bottom, &b)
	c.sideChars('l', b.Left, &b)

	return *c
}

// BorderAll creates Borders with the same border style for all sides.
func BorderAll(border BorderStyle) Borders {
	return Borders{Top: border, Right: border, Bottom: border, Left: border}
}

// BorderTRBL creates border styles following CSS order: Top, Right, Bottom, Left.
func BorderTRBL(t, r, b, l BorderStyle) Borders {
	return Borders{Top: t, Right: r, Bottom: b, Left: l}
}

// All returns true if all border styles match the given border style.
func (b Borders) All(s BorderStyle) bool {
	return b.Top == s && b.Right == s && b.Bottom == s && b.Left == s
}

// DrawBox draws box borders on the buffer at the specified rectangle.
// The box is drawn using the specified borders styles and style (colors/attributes).
// If the rectangle is smaller than 2x2, the function does nothing.
func DrawBox(buf *Buffer, rect Rect, borders Borders, style Style) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if borders.All(BorderNone) {
		return
	}

	chars := borders.Chars()

	// Clip rect to buffer bounds
	bufRect := buf.Rect()
	rect = rect.Intersect(bufRect)
	if rect.IsEmpty() || rect.Width < 2 || rect.Height < 2 {
		return
	}

	left := rect.X
	right := rect.Right() - 1
	top := rect.Y
	bottom := rect.Bottom() - 1

	// Draw corners
	buf.SetRune(left, top, chars.TopLeft, style)
	buf.SetRune(right, top, chars.TopRight, style)
	buf.SetRune(left, bottom, chars.BottomLeft, style)
	buf.SetRune(right, bottom, chars.BottomRight, style)

	// Draw top and bottom edges
	for x := left + 1; x < right; x++ {
		buf.SetRune(x, top, chars.Top, style)
		buf.SetRune(x, bottom, chars.Bottom, style)
	}

	// Draw left and right edges
	for y := top + 1; y < bottom; y++ {
		buf.SetRune(left, y, chars.Left, style)
		buf.SetRune(right, y, chars.Right, style)
	}
}

// DrawBoxGradient draws box borders with a gradient applied around the perimeter.
// The gradient is applied based on its direction:
// - Horizontal: left to right along top/bottom edges, top to bottom along left/right edges
// - Vertical: top to bottom along all edges
// - DiagonalDown: top-left to bottom-right
// - DiagonalUp: bottom-left to top-right
func DrawBoxGradient(buf *Buffer, rect Rect, borders Borders, g Gradient, baseStyle Style) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if borders.All(BorderNone) {
		return
	}

	chars := borders.Chars()

	// Clip rect to buffer bounds
	bufRect := buf.Rect()
	rect = rect.Intersect(bufRect)
	if rect.IsEmpty() || rect.Width < 2 || rect.Height < 2 {
		return
	}

	left := rect.X
	right := rect.Right() - 1
	top := rect.Y
	bottom := rect.Bottom() - 1
	width := float64(rect.Width)
	height := float64(rect.Height)
	perimeter := 2*width + 2*height - 4 // Subtract 4 for corners counted twice

	// Helper to calculate t along the perimeter, mirrored so the gradient
	// goes Start→End over the first half and End→Start over the second half.
	// This avoids a jarring color discontinuity where the perimeter wraps.
	getPerimeterT := func(x, y int) float64 {
		// Calculate position along perimeter: start at top-left, go clockwise
		var pos float64
		if y == top {
			// Top edge
			pos = float64(x - left)
		} else if x == right {
			// Right edge
			pos = width - 1 + float64(y-top)
		} else if y == bottom {
			// Bottom edge (right to left)
			pos = width - 1 + height - 1 + float64(right-x)
		} else {
			// Left edge (bottom to top)
			pos = width - 1 + height - 1 + width - 1 + float64(bottom-y)
		}
		t := pos / perimeter
		// Mirror: 0→1 for first half, 1→0 for second half
		if t <= 0.5 {
			return 2 * t
		}
		return 2 * (1 - t)
	}

	// Draw corners with gradient
	style := baseStyle
	if borders.Top != BorderNone {
		if borders.Left != BorderNone {
			style.Fg = g.At(getPerimeterT(left, top))
			buf.SetRune(left, top, chars.TopLeft, style)
		}

		if borders.Right != BorderNone {
			style.Fg = g.At(getPerimeterT(right, top))
			buf.SetRune(right, top, chars.TopRight, style)
		}
	}
	if borders.Bottom != BorderNone {
		if borders.Left != BorderNone {
			style.Fg = g.At(getPerimeterT(left, bottom))
			buf.SetRune(left, bottom, chars.BottomLeft, style)
		}

		if borders.Right != BorderNone {
			style.Fg = g.At(getPerimeterT(right, bottom))
			buf.SetRune(right, bottom, chars.BottomRight, style)
		}
	}

	// Draw top and bottom edges with gradient
	for x := left + 1; x < right; x++ {
		style.Fg = g.At(getPerimeterT(x, top))
		buf.SetRune(x, top, chars.Top, style)

		style.Fg = g.At(getPerimeterT(x, bottom))
		buf.SetRune(x, bottom, chars.Bottom, style)
	}

	// Draw left and right edges with gradient
	for y := top + 1; y < bottom; y++ {
		style.Fg = g.At(getPerimeterT(left, y))
		buf.SetRune(left, y, chars.Left, style)

		style.Fg = g.At(getPerimeterT(right, y))
		buf.SetRune(right, y, chars.Right, style)
	}
}

// DrawBoxClipped draws box borders clipped to the given clipRect.
// Positions are computed from the full rect, but only characters within
// clipRect are actually drawn. This enables partial borders rendering
// when an element is partially scrolled out of view.
func DrawBoxClipped(buf *Buffer, rect Rect, borders Borders, style Style, clipRect Rect) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if borders.All(BorderNone) {
		return
	}

	chars := borders.Chars()

	left := rect.X
	right := rect.Right() - 1
	top := rect.Y
	bottom := rect.Bottom() - 1

	// Draw corners (only if within clip region)
	if borders.Top != BorderNone {
		if borders.Left != BorderNone && clipRect.Contains(left, top) {
			buf.SetRune(left, top, chars.TopLeft, style)
		}
		if borders.Right != BorderNone && clipRect.Contains(right, top) {
			buf.SetRune(right, top, chars.TopRight, style)
		}
	}
	if borders.Bottom != BorderNone {
		if borders.Left != BorderNone && clipRect.Contains(left, bottom) {
			buf.SetRune(left, bottom, chars.BottomLeft, style)
		}
		if borders.Right != BorderNone && clipRect.Contains(right, bottom) {
			buf.SetRune(right, bottom, chars.BottomRight, style)
		}
	}

	// Draw top and bottom edges
	for x := left + 1; x < right; x++ {
		if clipRect.Contains(x, top) {
			buf.SetRune(x, top, chars.Top, style)
		}
		if clipRect.Contains(x, bottom) {
			buf.SetRune(x, bottom, chars.Bottom, style)
		}
	}

	// Draw left and right edges
	for y := top + 1; y < bottom; y++ {
		if clipRect.Contains(left, y) {
			buf.SetRune(left, y, chars.Left, style)
		}
		if clipRect.Contains(right, y) {
			buf.SetRune(right, y, chars.Right, style)
		}
	}
}

// DrawBoxGradientClipped draws gradient box borders clipped to the given clipRect.
// Positions and gradient colors are computed from the full rect, but only
// characters within clipRect are actually drawn.
func DrawBoxGradientClipped(buf *Buffer, rect Rect, borders Borders, g Gradient, baseStyle Style, clipRect Rect) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if borders.All(BorderNone) {
		return
	}

	chars := borders.Chars()

	left := rect.X
	right := rect.Right() - 1
	top := rect.Y
	bottom := rect.Bottom() - 1
	width := float64(rect.Width)
	height := float64(rect.Height)
	perimeter := 2*width + 2*height - 4

	// Mirrored perimeter t: Start→End over first half, End→Start over second half.
	getPerimeterT := func(x, y int) float64 {
		var pos float64
		if y == top {
			pos = float64(x - left)
		} else if x == right {
			pos = width - 1 + float64(y-top)
		} else if y == bottom {
			pos = width - 1 + height - 1 + float64(right-x)
		} else {
			pos = width - 1 + height - 1 + width - 1 + float64(bottom-y)
		}
		t := pos / perimeter
		if t <= 0.5 {
			return 2 * t
		}
		return 2 * (1 - t)
	}

	style := baseStyle

	// Draw corners with gradient (only if within clip region)
	if borders.Top != BorderNone {
		if borders.Left != BorderNone && clipRect.Contains(left, top) {
			style.Fg = g.At(getPerimeterT(left, top))
			buf.SetRune(left, top, chars.TopLeft, style)
		}
		if borders.Right != BorderNone && clipRect.Contains(right, top) {
			style.Fg = g.At(getPerimeterT(right, top))
			buf.SetRune(right, top, chars.TopRight, style)
		}
	}
	if borders.Bottom != BorderNone {
		if borders.Left != BorderNone && clipRect.Contains(left, bottom) {
			style.Fg = g.At(getPerimeterT(left, bottom))
			buf.SetRune(left, bottom, chars.BottomLeft, style)
		}
		if borders.Right != BorderNone && clipRect.Contains(right, bottom) {
			style.Fg = g.At(getPerimeterT(right, bottom))
			buf.SetRune(right, bottom, chars.BottomRight, style)
		}
	}

	// Draw top and bottom edges with gradient
	for x := left + 1; x < right; x++ {
		if clipRect.Contains(x, top) {
			style.Fg = g.At(getPerimeterT(x, top))
			buf.SetRune(x, top, chars.Top, style)
		}
		if clipRect.Contains(x, bottom) {
			style.Fg = g.At(getPerimeterT(x, bottom))
			buf.SetRune(x, bottom, chars.Bottom, style)
		}
	}

	// Draw left and right edges with gradient
	for y := top + 1; y < bottom; y++ {
		if clipRect.Contains(left, y) {
			style.Fg = g.At(getPerimeterT(left, y))
			buf.SetRune(left, y, chars.Left, style)
		}
		if clipRect.Contains(right, y) {
			style.Fg = g.At(getPerimeterT(right, y))
			buf.SetRune(right, y, chars.Right, style)
		}
	}
}

// DrawBoxWithTitle draws the borders of a box with a title in the top border.
// The title is aligned (default: center) and truncated if too long.
// If the rectangle is smaller than 2x2, the function does nothing.
func DrawBoxWithTitle(buf *Buffer, rect Rect, borders Borders, title string, style Style, align ...TextAlign) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if borders.All(BorderNone) {
		return
	}

	// First draw the box
	DrawBox(buf, rect, borders, style)

	// Draw the title
	drawBoxTitle(buf, rect, title, style, align...)
}

// titleCluster is one grapheme cluster of a border title with its display width.
type titleCluster struct {
	text  string
	width int
}

// drawBoxTitle draws an aligned title string on the top border line of the box.
// drawBoxTitle does not draw the box itself — use DrawBoxWithTitle for that.
func drawBoxTitle(buf *Buffer, rect Rect, title string, style Style, align ...TextAlign) {
	titleAlign := TextAlignCenter
	if len(align) > 0 {
		titleAlign = align[0]
	}
	clusters, startX, ok := prepareTitle(title, rect, titleAlign)
	if !ok {
		return
	}
	x := startX
	for _, c := range clusters {
		buf.setCluster(x, rect.Y, c.text, c.width, style, "")
		x += c.width
	}
}

// drawBoxTitleClipped draws an aligned title string on the top border line,
// skipping any cluster outside the given clip rectangle.
func drawBoxTitleClipped(buf *Buffer, rect Rect, title string, style Style, clipRect Rect, align ...TextAlign) {
	// Reject if the title row is outside the clip rect's Y range.
	if rect.Y < clipRect.Y || rect.Y >= clipRect.Y+clipRect.Height {
		return
	}
	titleAlign := TextAlignCenter
	if len(align) > 0 {
		titleAlign = align[0]
	}
	clusters, startX, ok := prepareTitle(title, rect, titleAlign)
	if !ok {
		return
	}
	x := startX
	for _, c := range clusters {
		if x >= clipRect.X && x+c.width <= clipRect.X+clipRect.Width {
			buf.setCluster(x, rect.Y, c.text, c.width, style, "")
		}
		x += c.width
	}
}

// prepareTitle truncates and aligns a title for the given rectangle, breaking
// on grapheme-cluster boundaries so a cluster is never split at the clip edge.
func prepareTitle(title string, rect Rect, align TextAlign) (clusters []titleCluster, startX int, ok bool) {
	availableWidth := rect.Width - 2
	if availableWidth <= 0 {
		return nil, 0, false
	}
	titleWidth := 0
	rest := title
	for len(rest) > 0 {
		cluster, w, size := nextCluster(rest)
		if size == 0 {
			break
		}
		if titleWidth+w > availableWidth {
			break
		}
		clusters = append(clusters, titleCluster{text: cluster, width: w})
		titleWidth += w
		rest = rest[size:]
	}
	if len(clusters) == 0 {
		return nil, 0, false
	}
	switch align {
	case TextAlignLeft:
		startX = rect.X + 1
	case TextAlignRight:
		startX = rect.X + rect.Width - 1 - titleWidth
	default:
		startX = rect.X + 1 + (availableWidth-titleWidth)/2
	}
	return clusters, startX, true
}

// FillBox fills the interior of a box (excluding the border) with a character and style.
// This is useful for clearing the interior before drawing content.
func FillBox(buf *Buffer, rect Rect, r rune, style Style) {
	if rect.Width <= 2 || rect.Height <= 2 {
		return
	}

	interior := rect.Inset(EdgeAll(1))
	buf.Fill(interior, r, style)
}
