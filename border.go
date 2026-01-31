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
