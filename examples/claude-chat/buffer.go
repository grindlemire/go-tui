package main

// TextBuffer manages text with cursor position
type TextBuffer struct {
	text   []rune
	cursor int
}

func NewTextBuffer() *TextBuffer {
	return &TextBuffer{}
}

func (b *TextBuffer) Insert(r rune) {
	b.text = append(b.text[:b.cursor], append([]rune{r}, b.text[b.cursor:]...)...)
	b.cursor++
}

func (b *TextBuffer) Backspace() {
	if b.cursor > 0 {
		b.text = append(b.text[:b.cursor-1], b.text[b.cursor:]...)
		b.cursor--
	}
}

func (b *TextBuffer) Delete() {
	if b.cursor < len(b.text) {
		b.text = append(b.text[:b.cursor], b.text[b.cursor+1:]...)
	}
}

func (b *TextBuffer) Left() {
	if b.cursor > 0 {
		b.cursor--
	}
}

func (b *TextBuffer) Right() {
	if b.cursor < len(b.text) {
		b.cursor++
	}
}

func (b *TextBuffer) Home() {
	b.cursor = 0
}

func (b *TextBuffer) End() {
	b.cursor = len(b.text)
}

func (b *TextBuffer) Clear() {
	b.text = nil
	b.cursor = 0
}

func (b *TextBuffer) String() string {
	return string(b.text)
}

// RenderWithCursor returns text with cursor marker at position
func (b *TextBuffer) RenderWithCursor() string {
	before := string(b.text[:b.cursor])
	after := string(b.text[b.cursor:])
	return "> " + before + "\u2588" + after
}

// GetDisplayLines returns wrapped lines for display
func (b *TextBuffer) GetDisplayLines(width int) []string {
	display := b.RenderWithCursor()
	// Account for border (2) and padding (2) = 4
	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}
	return WrapText(display, contentWidth)
}
