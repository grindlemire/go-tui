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

// Chars returns the box-drawing characters for this border style.
func (b BorderStyle) Chars() BorderChars {
	switch b {
	case BorderSingle:
		return BorderChars{
			TopLeft:     '┌',
			Top:         '─',
			TopRight:    '┐',
			Left:        '│',
			Right:       '│',
			BottomLeft:  '└',
			Bottom:      '─',
			BottomRight: '┘',
		}
	case BorderDouble:
		return BorderChars{
			TopLeft:     '╔',
			Top:         '═',
			TopRight:    '╗',
			Left:        '║',
			Right:       '║',
			BottomLeft:  '╚',
			Bottom:      '═',
			BottomRight: '╝',
		}
	case BorderRounded:
		return BorderChars{
			TopLeft:     '╭',
			Top:         '─',
			TopRight:    '╮',
			Left:        '│',
			Right:       '│',
			BottomLeft:  '╰',
			Bottom:      '─',
			BottomRight: '╯',
		}
	case BorderThick:
		return BorderChars{
			TopLeft:     '┏',
			Top:         '━',
			TopRight:    '┓',
			Left:        '┃',
			Right:       '┃',
			BottomLeft:  '┗',
			Bottom:      '━',
			BottomRight: '┛',
		}
	default:
		// BorderNone or unknown - return spaces
		return BorderChars{
			TopLeft:     ' ',
			Top:         ' ',
			TopRight:    ' ',
			Left:        ' ',
			Right:       ' ',
			BottomLeft:  ' ',
			Bottom:      ' ',
			BottomRight: ' ',
		}
	}
}

// DrawBox draws a box border on the buffer at the specified rectangle.
// The box is drawn using the specified border style and style (colors/attributes).
// If the rectangle is smaller than 2x2, the function does nothing.
func DrawBox(buf *Buffer, rect Rect, border BorderStyle, style Style) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if border == BorderNone {
		return
	}

	chars := border.Chars()

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

// DrawBoxGradient draws a box border with a gradient applied around the perimeter.
// The gradient is applied based on its direction:
// - Horizontal: left to right along top/bottom edges, top to bottom along left/right edges
// - Vertical: top to bottom along all edges
// - DiagonalDown: top-left to bottom-right
// - DiagonalUp: bottom-left to top-right
func DrawBoxGradient(buf *Buffer, rect Rect, border BorderStyle, g Gradient, baseStyle Style) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if border == BorderNone {
		return
	}

	chars := border.Chars()

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
	style.Fg = g.At(getPerimeterT(left, top))
	buf.SetRune(left, top, chars.TopLeft, style)

	style.Fg = g.At(getPerimeterT(right, top))
	buf.SetRune(right, top, chars.TopRight, style)

	style.Fg = g.At(getPerimeterT(left, bottom))
	buf.SetRune(left, bottom, chars.BottomLeft, style)

	style.Fg = g.At(getPerimeterT(right, bottom))
	buf.SetRune(right, bottom, chars.BottomRight, style)

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

// DrawBoxClipped draws a box border clipped to the given clipRect.
// Positions are computed from the full rect, but only characters within
// clipRect are actually drawn. This enables partial border rendering
// when an element is partially scrolled out of view.
func DrawBoxClipped(buf *Buffer, rect Rect, border BorderStyle, style Style, clipRect Rect) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if border == BorderNone {
		return
	}

	chars := border.Chars()

	left := rect.X
	right := rect.Right() - 1
	top := rect.Y
	bottom := rect.Bottom() - 1

	// Draw corners (only if within clip region)
	if clipRect.Contains(left, top) {
		buf.SetRune(left, top, chars.TopLeft, style)
	}
	if clipRect.Contains(right, top) {
		buf.SetRune(right, top, chars.TopRight, style)
	}
	if clipRect.Contains(left, bottom) {
		buf.SetRune(left, bottom, chars.BottomLeft, style)
	}
	if clipRect.Contains(right, bottom) {
		buf.SetRune(right, bottom, chars.BottomRight, style)
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

// DrawBoxGradientClipped draws a gradient box border clipped to the given clipRect.
// Positions and gradient colors are computed from the full rect, but only
// characters within clipRect are actually drawn.
func DrawBoxGradientClipped(buf *Buffer, rect Rect, border BorderStyle, g Gradient, baseStyle Style, clipRect Rect) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if border == BorderNone {
		return
	}

	chars := border.Chars()

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
	if clipRect.Contains(left, top) {
		style.Fg = g.At(getPerimeterT(left, top))
		buf.SetRune(left, top, chars.TopLeft, style)
	}
	if clipRect.Contains(right, top) {
		style.Fg = g.At(getPerimeterT(right, top))
		buf.SetRune(right, top, chars.TopRight, style)
	}
	if clipRect.Contains(left, bottom) {
		style.Fg = g.At(getPerimeterT(left, bottom))
		buf.SetRune(left, bottom, chars.BottomLeft, style)
	}
	if clipRect.Contains(right, bottom) {
		style.Fg = g.At(getPerimeterT(right, bottom))
		buf.SetRune(right, bottom, chars.BottomRight, style)
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

// DrawBoxWithTitle draws a box border with a title in the top border.
// The title is centered in the top border and truncated if too long.
// If the rectangle is smaller than 2x2, the function does nothing.
func DrawBoxWithTitle(buf *Buffer, rect Rect, border BorderStyle, title string, style Style) {
	if rect.Width < 2 || rect.Height < 2 {
		return
	}
	if border == BorderNone {
		return
	}

	// First draw the box
	DrawBox(buf, rect, border, style)

	// Now add the title if there's room
	if len(title) == 0 {
		return
	}

	// Calculate available space for title (leave at least 1 char on each side for corners)
	availableWidth := rect.Width - 2
	if availableWidth <= 0 {
		return
	}

	// Truncate title if needed
	titleRunes := []rune(title)
	titleWidth := 0
	truncatedRunes := make([]rune, 0, len(titleRunes))

	for _, r := range titleRunes {
		w := RuneWidth(r)
		if titleWidth+w > availableWidth {
			break
		}
		truncatedRunes = append(truncatedRunes, r)
		titleWidth += w
	}

	if len(truncatedRunes) == 0 {
		return
	}

	// Center the title in the available space
	startX := rect.X + 1 + (availableWidth-titleWidth)/2

	// Draw the title
	x := startX
	for _, r := range truncatedRunes {
		buf.SetRune(x, rect.Y, r, style)
		x += RuneWidth(r)
	}
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
