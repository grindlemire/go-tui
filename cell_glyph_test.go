package tui

import "unicode/utf8"

// cellGlyph reconstructs the full glyph of a cell (base rune + combining string).
// Returns the empty string for empty/continuation cells (Rune == 0).
func cellGlyph(c Cell) string {
	if c.Rune == 0 {
		return ""
	}
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], c.Rune)
	s := string(buf[:n])
	if c.Combining != "" {
		s += c.Combining
	}
	return s
}
