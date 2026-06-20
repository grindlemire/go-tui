package tui

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Cell represents a single character cell in the terminal buffer.
// Wide characters (CJK, emoji) and grapheme clusters (flags, ZWJ families,
// skin-tone emoji, accented letters) occupy multiple cells; the first cell
// holds the cluster's glyph, subsequent cells are marked as continuations.
//
// A cell stores its glyph as a base Rune plus an optional Combining string
// holding the remaining runes of a multi-rune cluster. The vast majority of
// cells (all ASCII, all CJK, single-code-point emoji) are a single rune stored
// inline with no allocation. Only genuine multi-rune clusters (flags, ZWJ
// families, decomposed accents, skin-tone, keycaps) populate Combining, which
// newClusterCell clones so the buffer never pins a large source string alive.
type Cell struct {
	Rune      rune   // Base (first) rune of the cluster; 0 for continuation/blank cells
	Combining string // Remaining runes of the cluster ("" for single-rune cells)
	Style     Style  // Visual styling
	Width     uint8  // Display width (1 or 2; 0 for continuation)
	Link      string // Optional OSC 8 hyperlink target ("" = none)
}

// NewCell creates a new Cell with automatic width detection.
func NewCell(r rune, style Style) Cell {
	return Cell{
		Rune:  r,
		Style: style,
		Width: uint8(RuneWidth(r)),
	}
}

// NewCellWithWidth creates a new Cell from a single rune with an explicit width.
// Use this for continuation cells (width 0) or when width is already known.
// A continuation cell (width 0) or a zero rune leaves Rune 0, so the cell reads
// as empty.
func NewCellWithWidth(r rune, style Style, width uint8) Cell {
	if width == 0 || r == 0 {
		return Cell{
			Style: style,
			Width: width,
		}
	}
	return Cell{
		Rune:  r,
		Style: style,
		Width: width,
	}
}

// newClusterCell creates a Cell holding a multi-rune (or single-rune) grapheme
// cluster. The cluster's base rune is stored inline; only the remaining runes of
// a true multi-rune cluster are cloned so the cell owns an independent backing
// array without pinning a larger source string. Pass width 0 for continuation
// cells.
func newClusterCell(text string, width uint8, style Style, link string) Cell {
	if width == 0 || text == "" {
		return Cell{Width: width, Style: style, Link: link} // Rune 0, Combining ""
	}
	r, sz := utf8.DecodeRuneInString(text)
	var comb string
	if sz < len(text) {
		// Only true multi-rune clusters reach here; clone so the cell does not
		// pin the source string alive.
		comb = strings.Clone(text[sz:])
	}
	return Cell{
		Rune:      r,
		Combining: comb,
		Style:     style,
		Width:     width,
		Link:      link,
	}
}

// IsContinuation returns true if this cell is a continuation of a wide character.
// Continuation cells have Width == 0 and are placed after the primary cell.
func (c Cell) IsContinuation() bool {
	return c.Width == 0
}

// Equal returns true if both cells are identical.
func (c Cell) Equal(other Cell) bool {
	return c.Rune == other.Rune && c.Combining == other.Combining && c.Style.Equal(other.Style) && c.Width == other.Width && c.Link == other.Link
}

// IsEmpty returns true if this cell represents an empty/blank cell.
// A cell is empty if it holds no glyph (or a single space) with default styling.
func (c Cell) IsEmpty() bool {
	// No glyph (zero/continuation cell) is considered empty
	if c.Rune == 0 {
		return true
	}
	// Space with default style is considered empty
	if c.Rune == ' ' && c.Combining == "" {
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
	// Emoji_Presentation=Yes BMP emoji — these render at width 2 without VS16.
	{min: 0x231A, max: 0x231B}, // Watch, hourglass
	{min: 0x23E9, max: 0x23EC}, // Fast-forward, rewind, etc.
	{min: 0x23F0, max: 0x23F0}, // Alarm clock
	{min: 0x23F3, max: 0x23F3}, // Hourglass not done
	{min: 0x25FD, max: 0x25FE}, // Medium-small squares
	{min: 0x2614, max: 0x2615}, // Umbrella, hot beverage
	{min: 0x2648, max: 0x2653}, // Zodiac
	{min: 0x267F, max: 0x267F}, // Wheelchair
	{min: 0x2693, max: 0x2693}, // Anchor
	{min: 0x26A1, max: 0x26A1}, // High voltage
	{min: 0x26AA, max: 0x26AB}, // Circles
	{min: 0x26BD, max: 0x26BE}, // Soccer, baseball
	{min: 0x26C4, max: 0x26C5}, // Snowman, sun behind cloud
	{min: 0x26CE, max: 0x26CE}, // Ophiuchus
	{min: 0x26D4, max: 0x26D4}, // No entry
	{min: 0x26EA, max: 0x26EA}, // Church
	{min: 0x26F2, max: 0x26F3}, // Fountain, flag in hole
	{min: 0x26F5, max: 0x26F5}, // Sailboat
	{min: 0x26FA, max: 0x26FA}, // Tent
	{min: 0x26FD, max: 0x26FD}, // Fuel pump
	{min: 0x270A, max: 0x270B}, // Raised fist, raised hand
	{min: 0x2705, max: 0x2705}, // Check mark button
	{min: 0x2728, max: 0x2728}, // Sparkles
	{min: 0x274C, max: 0x274C}, // Cross mark
	{min: 0x274E, max: 0x274E}, // Cross mark in box
	{min: 0x2753, max: 0x2755}, // Question, exclamation marks
	{min: 0x2757, max: 0x2757}, // Exclamation mark
	{min: 0x2795, max: 0x2797}, // Plus, minus, divide
	{min: 0x27B0, max: 0x27B0}, // Curly loop
	{min: 0x27BF, max: 0x27BF}, // Double curly loop
	{min: 0x2B1B, max: 0x2B1C}, // Black/white large squares
	{min: 0x2B50, max: 0x2B50}, // Star
	{min: 0x2B55, max: 0x2B55}, // Hollow red circle
	// SMP emoji (U+1Fxxx) — all are Emoji_Presentation=Yes (emoji-default).
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
