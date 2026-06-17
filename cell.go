package tui

import (
	"strings"
	"unicode"
)

// Cell represents a single character cell in the terminal buffer.
// Wide characters (CJK, emoji) and grapheme clusters (flags, ZWJ families,
// skin-tone emoji, accented letters) occupy multiple cells; the first cell
// holds the cluster's text, subsequent cells are marked as continuations.
//
// Cells own their text: a single rune is stored as string(r) (ASCII is
// staticbytes-backed, no allocation) and a multi-rune cluster sliced from a
// larger source is cloned before storage, so front/back buffers never pin a
// large source string alive.
type Cell struct {
	Text  string // The cluster's display text ("" for continuation/blank cells)
	Style Style  // Visual styling
	Width uint8  // Display width (1 or 2; 0 for continuation)
	Link  string // Optional OSC 8 hyperlink target ("" = none)
}

// NewCell creates a new Cell from a single rune with automatic width detection.
func NewCell(r rune, style Style) Cell {
	return Cell{
		Text:  string(r),
		Style: style,
		Width: uint8(RuneWidth(r)),
	}
}

// NewCellWithWidth creates a new Cell from a single rune with an explicit width.
// Use this for continuation cells (width 0) or when width is already known.
// A continuation cell (width 0) or a zero rune stores empty text, so the cell
// reads as empty.
func NewCellWithWidth(r rune, style Style, width uint8) Cell {
	text := ""
	if width != 0 && r != 0 {
		text = string(r)
	}
	return Cell{
		Text:  text,
		Style: style,
		Width: width,
	}
}

// newClusterCell creates a Cell holding a multi-rune (or single-rune) grapheme
// cluster. The text is cloned when it is a slice of a larger source so the cell
// owns an independent backing array. Pass width 0 for continuation cells.
func newClusterCell(text string, width uint8, style Style, link string) Cell {
	if width != 0 && len(text) > 0 {
		// Clone multi-byte text sliced from a larger source so the cell does not
		// pin the source alive. A single ASCII byte is staticbytes-backed; a
		// single multibyte rune is short enough that cloning is cheap and safe.
		text = strings.Clone(text)
	} else if width == 0 {
		text = ""
	}
	return Cell{
		Text:  text,
		Style: style,
		Width: width,
		Link:  link,
	}
}

// IsContinuation returns true if this cell is a continuation of a wide character.
// Continuation cells have Width == 0 and are placed after the primary cell.
func (c Cell) IsContinuation() bool {
	return c.Width == 0
}

// Equal returns true if both cells are identical.
func (c Cell) Equal(other Cell) bool {
	return c.Text == other.Text && c.Style.Equal(other.Style) && c.Width == other.Width && c.Link == other.Link
}

// IsEmpty returns true if this cell represents an empty/blank cell.
// A cell is empty if its text is empty (or a single space) with default styling.
func (c Cell) IsEmpty() bool {
	// Empty text (zero/continuation cell) is considered empty
	if c.Text == "" {
		return true
	}
	// Space with default style is considered empty
	if c.Text == " " {
		return c.Style.Equal(NewStyle())
	}
	return false
}

// RuneWidth returns the display width of a rune in terminal cells.
// Returns 1 for most characters, 2 for wide characters (CJK/fullwidth, emoji).
//
// Note: this cell model reserves Width==0 for continuation cells only.
// Runes that are logically zero-width (combining marks, variation selectors,
// format controls) are explicitly recognized but still treated as width 1.
func RuneWidth(r rune) int {
	// Keep invalid/control runes narrow so they don't disrupt layout.
	if r < 0 || r > unicode.MaxRune {
		return 1
	}

	// C0 and C1 controls.
	if r < 0x20 || (r >= 0x7F && r < 0xA0) {
		return 1
	}

	// Combining marks and format code points are logically zero-width, but this
	// buffer model uses Width==0 only for continuation cells.
	if isZeroWidthRune(r) {
		return 1
	}

	if inRuneRanges(r, eastAsianWideRanges) || inRuneRanges(r, emojiWideRanges) {
		return 2
	}

	return 1
}

type runeRange struct {
	min rune
	max rune
}

// East Asian wide/fullwidth code point ranges.
var eastAsianWideRanges = []runeRange{
	{min: 0x1100, max: 0x115F},   // Hangul Jamo init. consonants
	{min: 0x2329, max: 0x232A},   // Angle brackets
	{min: 0x2E80, max: 0x303E},   // CJK radicals + punctuation (excluding U+303F)
	{min: 0x3040, max: 0xA4CF},   // Hiragana/Katakana/Bopomofo/CJK/Yi
	{min: 0xAC00, max: 0xD7A3},   // Hangul syllables
	{min: 0xF900, max: 0xFAFF},   // CJK compatibility ideographs
	{min: 0xFE10, max: 0xFE19},   // Vertical forms
	{min: 0xFE30, max: 0xFE6F},   // CJK compatibility forms + small forms
	{min: 0xFF00, max: 0xFF60},   // Fullwidth forms
	{min: 0xFFE0, max: 0xFFE6},   // Fullwidth symbol variants
	{min: 0x1B000, max: 0x1B12F}, // Kana supplement + Kana ext. A
	{min: 0x1B130, max: 0x1B167}, // Kana extended B
	{min: 0x20000, max: 0x2FFFD}, // CJK extensions
	{min: 0x30000, max: 0x3FFFD}, // CJK extensions
}

// Emoji ranges that terminals commonly render as 2-cell glyphs.
var emojiWideRanges = []runeRange{
	{min: 0x1F004, max: 0x1F004}, // Mahjong tile red dragon
	{min: 0x1F0CF, max: 0x1F0CF}, // Playing card black joker
	{min: 0x1F18E, max: 0x1F18E}, // Negative squared AB
	{min: 0x1F191, max: 0x1F19A}, // Squared symbols
	{min: 0x1F1E6, max: 0x1F1FF}, // Regional indicator symbols (flags)
	{min: 0x1F201, max: 0x1F202}, // Squared Katakana words
	{min: 0x1F21A, max: 0x1F21A}, // Squared CJK ideograph
	{min: 0x1F22F, max: 0x1F22F}, // Squared CJK ideograph
	{min: 0x1F232, max: 0x1F23A}, // Squared CJK ideographs
	{min: 0x1F250, max: 0x1F251}, // Circled ideographs
	{min: 0x1F300, max: 0x1F64F}, // Pictographs + emoticons
	{min: 0x1F680, max: 0x1F6FF}, // Transport/map symbols
	{min: 0x1F7E0, max: 0x1F7EB}, // Large colored circles/squares
	{min: 0x1F900, max: 0x1F9FF}, // Supplemental symbols/pictographs
	{min: 0x1FA70, max: 0x1FAFF}, // Symbols/pictographs ext. A
}

func isZeroWidthRune(r rune) bool {
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf, unicode.Variation_Selector, unicode.Join_Control)
}

func inRuneRanges(r rune, ranges []runeRange) bool {
	for _, rr := range ranges {
		if r >= rr.min && r <= rr.max {
			return true
		}
	}
	return false
}
