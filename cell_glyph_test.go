package tui

// cellGlyph reconstructs a cell's full glyph (its base rune followed by any
// combining runes) for test assertions on multi-rune grapheme clusters. A blank
// or continuation cell has no base rune (Rune == 0) and reads as an empty glyph,
// not a NUL.
func cellGlyph(c Cell) string {
	if c.Rune == 0 {
		return c.Combining
	}
	return string(c.Rune) + c.Combining
}
