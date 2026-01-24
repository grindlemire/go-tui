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
		'你', '好', '中', '文',  // Chinese
		'日', '本', '語',       // Japanese kanji
		'あ', 'い', 'う',       // Hiragana
		'ア', 'イ', 'ウ',       // Katakana
		'한', '글',            // Korean Hangul
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
		'é', 'è', 'ê', 'ë',  // French accents
		'ñ', 'ü', 'ö', 'ä',  // Spanish/German
		'ø', 'æ', 'å',       // Nordic
		'ß',                 // German eszett
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
		'Ａ', 'Ｂ', 'Ｃ',  // Fullwidth Latin
		'０', '１', '２',  // Fullwidth digits
	}

	for _, r := range fullwidthChars {
		if w := RuneWidth(r); w != 2 {
			t.Errorf("RuneWidth(%q U+%04X) = %d, want 2", r, r, w)
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
