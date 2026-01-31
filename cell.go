package tui

// Cell represents a single character cell in the terminal buffer.
// Wide characters (CJK, emoji) occupy multiple cells; the first cell holds
// the rune, subsequent cells are marked as continuations.
type Cell struct {
	Rune  rune  // The character (0 for continuation cells)
	Style Style // Visual styling
	Width uint8 // Display width (1 or 2; 0 for continuation)
}

// NewCell creates a new Cell with automatic width detection.
func NewCell(r rune, style Style) Cell {
	return Cell{
		Rune:  r,
		Style: style,
		Width: uint8(RuneWidth(r)),
	}
}

// NewCellWithWidth creates a new Cell with an explicit width.
// Use this for continuation cells (width 0) or when width is already known.
func NewCellWithWidth(r rune, style Style, width uint8) Cell {
	return Cell{
		Rune:  r,
		Style: style,
		Width: width,
	}
}

// IsContinuation returns true if this cell is a continuation of a wide character.
// Continuation cells have Width == 0 and are placed after the primary cell.
func (c Cell) IsContinuation() bool {
	return c.Width == 0
}

// Equal returns true if both cells are identical.
func (c Cell) Equal(other Cell) bool {
	return c.Rune == other.Rune && c.Style.Equal(other.Style) && c.Width == other.Width
}

// IsEmpty returns true if this cell represents an empty/blank cell.
// A cell is empty if it's a space (or zero rune) with default styling.
func (c Cell) IsEmpty() bool {
	// Zero rune with any style is considered empty
	if c.Rune == 0 {
		return true
	}
	// Space with default style is considered empty
	if c.Rune == ' ' {
		return c.Style.Equal(NewStyle())
	}
	return false
}

// RuneWidth returns the display width of a rune in terminal cells.
// Returns 1 for most characters, 2 for wide characters (CJK, most emoji).
// This uses simple range checks for common cases to avoid dependency on
// full Unicode tables.
func RuneWidth(r rune) int {
	// Control characters and zero have no width, but we return 1 for simplicity
	// since we need at least 1 cell to represent them
	if r < 32 {
		return 1
	}

	// ASCII printable characters are width 1
	if r < 127 {
		return 1
	}

	// Latin-1 Supplement through various European scripts
	if r < 0x1100 {
		return 1
	}

	// Hangul Jamo (Korean)
	if r >= 0x1100 && r <= 0x115F {
		return 2
	}

	// Various symbols and diacritical marks
	if r >= 0x1160 && r <= 0x2328 {
		return 1
	}

	// Miscellaneous symbols that are narrow
	if r >= 0x2329 && r <= 0x232A {
		return 2
	}

	if r >= 0x232B && r <= 0x2E7F {
		return 1
	}

	// CJK Radicals Supplement through Yi Radicals
	// This covers:
	// - CJK Radicals Supplement (2E80-2EFF)
	// - Kangxi Radicals (2F00-2FDF)
	// - CJK Symbols and Punctuation (3000-303F)
	// - Hiragana (3040-309F)
	// - Katakana (30A0-30FF)
	// - Bopomofo (3100-312F)
	// - Hangul Compatibility Jamo (3130-318F)
	// - Kanbun (3190-319F)
	// - Bopomofo Extended (31A0-31BF)
	// - CJK Strokes (31C0-31EF)
	// - Katakana Phonetic Extensions (31F0-31FF)
	// - Enclosed CJK Letters and Months (3200-32FF)
	// - CJK Compatibility (3300-33FF)
	// - CJK Unified Ideographs Extension A (3400-4DBF)
	// - Yijing Hexagram Symbols (4DC0-4DFF)
	// - CJK Unified Ideographs (4E00-9FFF)
	// - Yi Syllables (A000-A48F)
	// - Yi Radicals (A490-A4CF)
	if r >= 0x2E80 && r <= 0xA4CF {
		return 2
	}

	// Hangul Syllables
	if r >= 0xAC00 && r <= 0xD7A3 {
		return 2
	}

	// CJK Compatibility Ideographs
	if r >= 0xF900 && r <= 0xFAFF {
		return 2
	}

	// Fullwidth ASCII variants and Halfwidth CJK punctuation
	// Fullwidth characters are width 2
	if r >= 0xFF00 && r <= 0xFF60 {
		return 2
	}

	// Halfwidth Katakana and Hangul
	if r >= 0xFF61 && r <= 0xFFDC {
		return 1
	}

	// Fullwidth symbol variants
	if r >= 0xFFE0 && r <= 0xFFE6 {
		return 2
	}

	// CJK Unified Ideographs Extension B through Extension G and beyond
	if r >= 0x20000 && r <= 0x3FFFF {
		return 2
	}

	// Emoji and pictographs (most are wide)
	// Miscellaneous Symbols and Pictographs
	if r >= 0x1F300 && r <= 0x1F9FF {
		return 2
	}

	// Supplemental Symbols and Pictographs
	if r >= 0x1FA00 && r <= 0x1FAFF {
		return 2
	}

	// Default to width 1 for anything else
	return 1
}
