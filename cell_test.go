package tui

import (
	"testing"
)

func TestNewCell(t *testing.T) {
	type tc struct {
		r             rune
		style         Style
		expectedWidth uint8
	}

	tests := map[string]tc{
		"ASCII letter": {
			r:             'A',
			style:         NewStyle(),
			expectedWidth: 1,
		},
		"ASCII space": {
			r:             ' ',
			style:         NewStyle().Bold(),
			expectedWidth: 1,
		},
		"CJK character": {
			r:             '你',
			style:         NewStyle(),
			expectedWidth: 2,
		},
		"emoji": {
			r:             '😀',
			style:         NewStyle(),
			expectedWidth: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := NewCell(tt.r, tt.style)
			if c.Rune != tt.r {
				t.Errorf("NewCell().Rune = %q, want %q", c.Rune, tt.r)
			}
			if !c.Style.Equal(tt.style) {
				t.Errorf("NewCell().Style doesn't match expected style")
			}
			if c.Width != tt.expectedWidth {
				t.Errorf("NewCell(%q).Width = %d, want %d", tt.r, c.Width, tt.expectedWidth)
			}
		})
	}
}

func TestNewCellWithWidth(t *testing.T) {
	style := NewStyle().Foreground(Red)

	// Test explicit width for continuation cell
	c := NewCellWithWidth(0, style, 0)
	if c.Rune != 0 {
		t.Errorf("NewCellWithWidth().Rune = %q, want 0", c.Rune)
	}
	if c.Width != 0 {
		t.Errorf("NewCellWithWidth().Width = %d, want 0", c.Width)
	}
	if !c.Style.Equal(style) {
		t.Error("NewCellWithWidth().Style doesn't match")
	}

	// Test explicit width override
	c2 := NewCellWithWidth('A', style, 2)
	if c2.Width != 2 {
		t.Errorf("NewCellWithWidth('A', _, 2).Width = %d, want 2", c2.Width)
	}
}

func TestCell_IsContinuation(t *testing.T) {
	type tc struct {
		cell           Cell
		isContinuation bool
	}

	tests := map[string]tc{
		"regular ASCII cell": {
			cell:           NewCell('A', NewStyle()),
			isContinuation: false,
		},
		"wide character cell": {
			cell:           NewCell('你', NewStyle()),
			isContinuation: false,
		},
		"continuation cell": {
			cell:           NewCellWithWidth(0, NewStyle(), 0),
			isContinuation: true,
		},
		"zero rune but width 1": {
			cell:           NewCellWithWidth(0, NewStyle(), 1),
			isContinuation: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cell.IsContinuation(); got != tt.isContinuation {
				t.Errorf("IsContinuation() = %v, want %v", got, tt.isContinuation)
			}
		})
	}
}

func TestCell_Equal(t *testing.T) {
	type tc struct {
		a, b  Cell
		equal bool
	}

	styleRed := NewStyle().Foreground(Red)
	styleBlue := NewStyle().Foreground(Blue)

	tests := map[string]tc{
		"identical cells": {
			a:     NewCell('A', NewStyle()),
			b:     NewCell('A', NewStyle()),
			equal: true,
		},
		"different rune": {
			a:     NewCell('A', NewStyle()),
			b:     NewCell('B', NewStyle()),
			equal: false,
		},
		"different style": {
			a:     NewCell('A', styleRed),
			b:     NewCell('A', styleBlue),
			equal: false,
		},
		"different width": {
			a:     NewCellWithWidth('A', NewStyle(), 1),
			b:     NewCellWithWidth('A', NewStyle(), 2),
			equal: false,
		},
		"wide characters equal": {
			a:     NewCell('好', styleRed),
			b:     NewCell('好', styleRed),
			equal: true,
		},
		"continuation cells equal": {
			a:     NewCellWithWidth(0, NewStyle(), 0),
			b:     NewCellWithWidth(0, NewStyle(), 0),
			equal: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.equal {
				t.Errorf("Equal() = %v, want %v", got, tt.equal)
			}
			// Test symmetry
			if got := tt.b.Equal(tt.a); got != tt.equal {
				t.Errorf("Equal() (reversed) = %v, want %v", got, tt.equal)
			}
		})
	}
}

func TestCell_IsEmpty(t *testing.T) {
	type tc struct {
		cell    Cell
		isEmpty bool
	}

	tests := map[string]tc{
		"space with default style": {
			cell:    NewCell(' ', NewStyle()),
			isEmpty: true,
		},
		"space with style": {
			cell:    NewCell(' ', NewStyle().Bold()),
			isEmpty: false,
		},
		"space with foreground color": {
			cell:    NewCell(' ', NewStyle().Foreground(Red)),
			isEmpty: false,
		},
		"zero rune": {
			cell:    NewCellWithWidth(0, NewStyle(), 1),
			isEmpty: true,
		},
		"zero rune continuation": {
			cell:    NewCellWithWidth(0, NewStyle(), 0),
			isEmpty: true,
		},
		"regular character": {
			cell:    NewCell('A', NewStyle()),
			isEmpty: false,
		},
		"wide character": {
			cell:    NewCell('你', NewStyle()),
			isEmpty: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cell.IsEmpty(); got != tt.isEmpty {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.isEmpty)
			}
		})
	}
}

func TestRuneWidth_ASCII(t *testing.T) {
	// ASCII letters and numbers should be width 1
	asciiChars := []rune{'a', 'z', 'A', 'Z', '0', '9', '!', '@', '#', ' ', '\t'}

	for _, r := range asciiChars {
		if w := RuneWidth(r); w != 1 {
			t.Errorf("RuneWidth(%q) = %d, want 1", r, w)
		}
	}
}

func TestRuneWidth_CJK(t *testing.T) {
	// CJK characters should be width 2
	cjkChars := []rune{
		'你', '好', '中', '文', // Chinese
		'日', '本', '語', // Japanese kanji
		'あ', 'い', 'う', // Hiragana
		'ア', 'イ', 'ウ', // Katakana
		'한', '글', // Korean Hangul
	}

	for _, r := range cjkChars {
		if w := RuneWidth(r); w != 2 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 2", r, r, w)
		}
	}
}

func TestRuneWidth_Emoji(t *testing.T) {
	// Common emoji should be width 2
	emojis := []rune{
		'😀', '😁', '🎉', '🚀', '💻', '🌟',
		// BMP emoji that were missing from emojiWideRanges
		'✨', // U+2728 Sparkles
		'⭐', // U+2B50 Star
		'☕', // U+2615 Hot Beverage
		'✅', // U+2705 Check Mark
		'❤', // U+2764 Heavy Black Heart
		'☀', // U+2600 Sun
		'⚡', // U+26A1 High Voltage
	}

	for _, r := range emojis {
		if w := RuneWidth(r); w != 2 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 2", r, r, w)
		}
	}
}

func TestRuneWidth_BoxDrawing(t *testing.T) {
	// Box drawing characters should be width 1
	boxChars := []rune{
		'─', '│', '┌', '┐', '└', '┘', '├', '┤', '┬', '┴', '┼',
		'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬',
		'╭', '╮', '╯', '╰', // Rounded corners
	}

	for _, r := range boxChars {
		if w := RuneWidth(r); w != 1 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 1", r, r, w)
		}
	}
}

func TestRuneWidth_Latin(t *testing.T) {
	// Extended Latin characters should be width 1
	latinChars := []rune{
		'é', 'è', 'ê', 'ë', // French accents
		'ñ', 'ü', 'ö', 'ä', // Spanish/German
		'ø', 'æ', 'å', // Nordic
		'ß', // German eszett
	}

	for _, r := range latinChars {
		if w := RuneWidth(r); w != 1 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 1", r, r, w)
		}
	}
}

func TestRuneWidth_Fullwidth(t *testing.T) {
	// Fullwidth ASCII variants should be width 2
	fullwidthChars := []rune{
		'Ａ', 'Ｂ', 'Ｃ', // Fullwidth Latin
		'０', '１', '２', // Fullwidth digits
	}

	for _, r := range fullwidthChars {
		if w := RuneWidth(r); w != 2 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 2", r, r, w)
		}
	}
}

func TestRuneWidth_RegionalIndicators(t *testing.T) {
	// Regional indicator symbols are used for flag emoji and are rendered wide.
	indicators := []rune{
		'\U0001F1FA', // REGIONAL INDICATOR SYMBOL LETTER U
		'\U0001F1F8', // REGIONAL INDICATOR SYMBOL LETTER S
		'\U0001F1EF', // REGIONAL INDICATOR SYMBOL LETTER J
		'\U0001F1F5', // REGIONAL INDICATOR SYMBOL LETTER P
	}

	for _, r := range indicators {
		if w := RuneWidth(r); w != 2 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 2", r, r, w)
		}
	}
}

func TestRuneWidth_CJKCompatibilityForms(t *testing.T) {
	// Vertical presentation and compatibility punctuation are wide.
	chars := []rune{
		'\uFE10', // PRESENTATION FORM FOR VERTICAL COMMA
		'\uFE31', // PRESENTATION FORM FOR VERTICAL EM DASH
		'\uFE44', // PRESENTATION FORM FOR VERTICAL RIGHT WHITE CORNER BRACKET
	}

	for _, r := range chars {
		if w := RuneWidth(r); w != 2 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 2", r, r, w)
		}
	}
}

func TestRuneWidth_ZeroWidthCategoriesFallback(t *testing.T) {
	// These are logically zero-width, but this buffer reserves width 0 for
	// continuation cells only, so they remain width 1.
	chars := []rune{
		'\u0301', // COMBINING ACUTE ACCENT
		'\u200D', // ZERO WIDTH JOINER
		'\uFE0F', // VARIATION SELECTOR-16
	}

	for _, r := range chars {
		if w := RuneWidth(r); w != 1 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 1", r, r, w)
		}
	}
}

func TestCell_ZeroValue(t *testing.T) {
	var c Cell

	// Zero value cell
	if c.Rune != 0 {
		t.Errorf("zero value Cell.Rune = %q, want 0", c.Rune)
	}
	if c.Width != 0 {
		t.Errorf("zero value Cell.Width = %d, want 0", c.Width)
	}
	// Zero value is a continuation cell
	if !c.IsContinuation() {
		t.Error("zero value Cell should be continuation")
	}
	// Zero value is empty
	if !c.IsEmpty() {
		t.Error("zero value Cell should be empty")
	}
}

func TestCell_EqualConsidersLink(t *testing.T) {
	a := NewCell('x', NewStyle())
	b := NewCell('x', NewStyle())
	b.Link = "https://example.com"
	if a.Equal(b) {
		t.Error("cells differing only in Link should not be Equal")
	}
	b.Link = ""
	if !a.Equal(b) {
		t.Error("cells with identical fields (empty Link) should be Equal")
	}
}

func TestBuffer_DiffDetectsLinkChange(t *testing.T) {
	buf := NewBuffer(3, 1)
	buf.Swap() // front == back, no diff
	c := NewCell('a', NewStyle())
	c.Link = "https://example.com"
	buf.SetCell(0, 0, c)
	changes := buf.Diff()
	if len(changes) != 1 || changes[0].Cell.Link != "https://example.com" {
		t.Fatalf("expected one change carrying the link, got %+v", changes)
	}
}

// TestRuneWidth_EmojiRangeValidation verifies every code point in the
// emojiWideRanges list is actually width 2, and that common non-emoji
// characters in the same Unicode blocks are NOT accidentally width 2.
func TestRuneWidth_EmojiRangeValidation(t *testing.T) {
	// Well-known emoji that MUST be width 2
	mustBeWide := []rune{
		0x00A9, 0x00AE, 0x203C, 0x2049, 0x2122, 0x2139,
		0x2194, 0x2195, 0x2196, 0x2197, 0x2198, 0x2199,
		0x21A9, 0x21AA, 0x231A, 0x231B, 0x2328, 0x23CF,
		0x23E9, 0x23EA, 0x23EB, 0x23EC, 0x23ED, 0x23EE,
		0x23EF, 0x23F0, 0x23F8, 0x23F9, 0x23FA,
		0x24C2, 0x25AA, 0x25AB, 0x25B6, 0x25C0,
		0x25FB, 0x25FC, 0x25FD, 0x25FE,
		0x2600, 0x2601, 0x260E, 0x2611, 0x2614, 0x2615,
		0x2618, 0x261D, 0x2620, 0x2622, 0x2623, 0x2626,
		0x262A, 0x262E, 0x262F, 0x2638, 0x2639, 0x263A,
		0x2640, 0x2642,
		0x2648, 0x2649, 0x264A, 0x264B, 0x264C, 0x264D,
		0x264E, 0x264F, 0x2650, 0x2651, 0x2652, 0x2653,
		0x265F, 0x2660, 0x2663, 0x2665, 0x2666, 0x2668,
		0x267B, 0x267E, 0x267F,
		0x2692, 0x2693, 0x2694, 0x2695, 0x2696, 0x2697,
		0x2699, 0x269B, 0x269C,
		0x26A0, 0x26A1, 0x26A7, 0x26AA, 0x26AB,
		0x26B0, 0x26B1, 0x26BD, 0x26BE, 0x26C4, 0x26C5,
		0x26C8, 0x26CE, 0x26CF, 0x26D1, 0x26D3, 0x26D4,
		0x26E9, 0x26EA,
		0x26F0, 0x26F1, 0x26F2, 0x26F3, 0x26F4, 0x26F5,
		0x26F7, 0x26F8, 0x26F9, 0x26FA, 0x26FD,
		0x2702, 0x2705, 0x2708, 0x2709, 0x270A, 0x270B,
		0x270C, 0x270D, 0x270F, 0x2712, 0x2714, 0x2716,
		0x271D, 0x2721, 0x2728,
		0x2733, 0x2734, 0x2744, 0x2747, 0x274C, 0x274E,
		0x2753, 0x2754, 0x2755, 0x2757, 0x2763, 0x2764,
		0x2795, 0x2796, 0x2797, 0x27A1, 0x27B0, 0x27BF,
		0x2934, 0x2935, 0x2B05, 0x2B06, 0x2B07,
		0x2B1B, 0x2B1C, 0x2B50, 0x2B55,
	}
	for _, r := range mustBeWide {
		if w := RuneWidth(r); w != 2 {
			t.Errorf("RuneWidth(U+%04X) = %d, want 2 (emoji)", r, w)
		}
	}

	// Characters in the same blocks that should NOT be width 2
	mustBeNarrow := []rune{
		0x231C, 0x231D, 0x231E, 0x231F, // Corner brackets
		0x2320, 0x2321, // Integral symbols
		0x2605, 0x2606, // Black/white stars
		0x2607, 0x2608, 0x2609, 0x260A, 0x260B, 0x260C, 0x260D, // Astrological
		0x260F, 0x2610, // Telephone, ballot box
		0x2612, 0x2613, // Ballot X, saltire
		0x2616, 0x2617, // Shogi pieces
		0x2619, 0x261A, 0x261B, 0x261C, // Floral, pointing
		0x261E, 0x261F, // Pointing
		0x2621, 0x2624, 0x2625, // Caution, caduceus, ankh
		0x2627, 0x2628, 0x2629, // Crosses
		0x262B, 0x262C, 0x262D, // Farsi, adi shakti, hammer
		0x2630, 0x2631, 0x2632, 0x2633, 0x2634, 0x2635, 0x2636, 0x2637, // Trigrams
		0x2641, 0x2643, 0x2644, 0x2645, 0x2646, 0x2647, // Planets
		0x2654, 0x2655, 0x2656, 0x2657, 0x2658, 0x2659, 0x265A, 0x265B, 0x265C, 0x265D, 0x265E, // Chess
		0x2661, 0x2662, // White heart, diamond
		0x2664, 0x2667, // White spade, club
		0x2669, 0x266A, 0x266B, 0x266C, 0x266D, 0x266E, 0x266F, // Music
		0x2670, 0x2671, // Crosses
		0x267C, 0x267D, // Recycling variants
		0x2690, 0x2691, // Flags
		0x2698,                 // Flower
		0x269D, 0x269E, 0x269F, // Stars/lines
		0x26A2, 0x26A3, 0x26A4, 0x26A5, 0x26A6, // Gender symbols
		0x26A8, 0x26A9, // Gender symbols
		0x26AC, 0x26AD, 0x26AE, 0x26AF, // Marriage/divorce
		0x26B2,                                                                         // Neuter
		0x26B3, 0x26B4, 0x26B5, 0x26B6, 0x26B7, 0x26B8, 0x26B9, 0x26BA, 0x26BB, 0x26BC, // Astrological
		0x26C0, 0x26C1, 0x26C2, 0x26C3, // Draughts
		0x26C6, 0x26C7, // Rain, snowman
		0x26C9, 0x26CA, 0x26CB, 0x26CC, 0x26CD, // Shogi, lanes
		0x26D0, 0x26D2, // Car sliding, car crossing
		0x26D5, 0x26D6, 0x26D7, 0x26D8, 0x26D9, 0x26DA, 0x26DB, 0x26DC, 0x26DD, 0x26DE, 0x26DF, // Road symbols
		0x26E0, 0x26E1, 0x26E2, 0x26E3, 0x26E4, 0x26E5, 0x26E6, 0x26E7, 0x26E8, // Misc
		0x26EC, 0x26ED, 0x26EE, 0x26EF, // Misc
		0x2700, 0x2701, // Scissors variants
		0x2703, 0x2704, // Scissors variants
	}
	for _, r := range mustBeNarrow {
		if w := RuneWidth(r); w != 1 {
			t.Errorf("RuneWidth(U+%04X) = %d, want 1 (not emoji)", r, w)
		}
	}

	// Verify no overlaps with eastAsianWideRanges
	eastAsian := []runeRange{
		{0x1100, 0x115F},
		{0x2329, 0x232A},
		{0x2E80, 0x303E},
		{0x3040, 0xA4CF},
		{0xAC00, 0xD7A3},
		{0xF900, 0xFAFF},
		{0xFE10, 0xFE19},
		{0xFE30, 0xFE6F},
		{0xFF00, 0xFF60},
		{0xFFE0, 0xFFE6},
		{0x1B000, 0x1B12F},
		{0x1B130, 0x1B167},
		{0x20000, 0x2FFFD},
		{0x30000, 0x3FFFD},
	}
	emoji := emojiWideRanges
	for _, er := range emoji {
		for _, ea := range eastAsian {
			if er.max >= ea.min && er.min <= ea.max {
				t.Errorf("emojiWideRanges U+%04X-U+%04X overlaps eastAsianWideRanges U+%04X-U+%04X",
					er.min, er.max, ea.min, ea.max)
			}
		}
	}
}
