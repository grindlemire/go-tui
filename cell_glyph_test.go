package tui

// cellGlyph reconstructs a cell's full glyph (its base rune followed by any
// combining runes) for test assertions on multi-rune grapheme clusters.
func cellGlyph(c Cell) string { return string(c.Rune) + c.Combining }
