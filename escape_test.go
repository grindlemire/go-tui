package tui

import (
	"testing"
)

func TestEscBuilder_MoveTo(t *testing.T) {
	type tc struct {
		x, y     int
		expected string
	}

	tests := map[string]tc{
		"origin": {
			x:        0,
			y:        0,
			expected: "\x1b[1;1H",
		},
		"position 5,3": {
			x:        5,
			y:        3,
			expected: "\x1b[4;6H", // 1-indexed: row 4, col 6
		},
		"large position": {
			x:        99,
			y:        49,
			expected: "\x1b[50;100H",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			e.MoveTo(tt.x, tt.y)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("MoveTo(%d, %d) = %q, want %q", tt.x, tt.y, e.Bytes(), tt.expected)
			}
		})
	}
}

func TestEscBuilder_CursorMovement(t *testing.T) {
	type tc struct {
		fn       func(*escBuilder, int)
		n        int
		expected string
	}

	tests := map[string]tc{
		"move up 1": {
			fn:       func(e *escBuilder, n int) { e.MoveUp(n) },
			n:        1,
			expected: "\x1b[A",
		},
		"move up 5": {
			fn:       func(e *escBuilder, n int) { e.MoveUp(n) },
			n:        5,
			expected: "\x1b[5A",
		},
		"move down 1": {
			fn:       func(e *escBuilder, n int) { e.MoveDown(n) },
			n:        1,
			expected: "\x1b[B",
		},
		"move down 3": {
			fn:       func(e *escBuilder, n int) { e.MoveDown(n) },
			n:        3,
			expected: "\x1b[3B",
		},
		"move right 1": {
			fn:       func(e *escBuilder, n int) { e.MoveRight(n) },
			n:        1,
			expected: "\x1b[C",
		},
		"move right 10": {
			fn:       func(e *escBuilder, n int) { e.MoveRight(n) },
			n:        10,
			expected: "\x1b[10C",
		},
		"move left 1": {
			fn:       func(e *escBuilder, n int) { e.MoveLeft(n) },
			n:        1,
			expected: "\x1b[D",
		},
		"move left 7": {
			fn:       func(e *escBuilder, n int) { e.MoveLeft(n) },
			n:        7,
			expected: "\x1b[7D",
		},
		"move up 0 does nothing": {
			fn:       func(e *escBuilder, n int) { e.MoveUp(n) },
			n:        0,
			expected: "",
		},
		"move down negative does nothing": {
			fn:       func(e *escBuilder, n int) { e.MoveDown(n) },
			n:        -1,
			expected: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			tt.fn(e, tt.n)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("got %q, want %q", e.Bytes(), tt.expected)
			}
		})
	}
}

func TestEscBuilder_ClearScreen(t *testing.T) {
	e := newEscBuilder(64)
	e.ClearScreen()
	expected := "\x1b[2J"
	if string(e.Bytes()) != expected {
		t.Errorf("ClearScreen() = %q, want %q", e.Bytes(), expected)
	}
}

func TestEscBuilder_ClearLine(t *testing.T) {
	e := newEscBuilder(64)
	e.ClearLine()
	expected := "\x1b[2K"
	if string(e.Bytes()) != expected {
		t.Errorf("ClearLine() = %q, want %q", e.Bytes(), expected)
	}
}

func TestEscBuilder_HideCursor(t *testing.T) {
	e := newEscBuilder(64)
	e.HideCursor()
	expected := "\x1b[?25l"
	if string(e.Bytes()) != expected {
		t.Errorf("HideCursor() = %q, want %q", e.Bytes(), expected)
	}
}

func TestEscBuilder_ShowCursor(t *testing.T) {
	e := newEscBuilder(64)
	e.ShowCursor()
	expected := "\x1b[?25h"
	if string(e.Bytes()) != expected {
		t.Errorf("ShowCursor() = %q, want %q", e.Bytes(), expected)
	}
}

func TestEscBuilder_AltScreen(t *testing.T) {
	type tc struct {
		fn       func(*escBuilder)
		expected string
	}

	tests := map[string]tc{
		"enter alt screen": {
			fn:       func(e *escBuilder) { e.EnterAltScreen() },
			expected: "\x1b[?1049h",
		},
		"exit alt screen": {
			fn:       func(e *escBuilder) { e.ExitAltScreen() },
			expected: "\x1b[?1049l",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			tt.fn(e)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("got %q, want %q", e.Bytes(), tt.expected)
			}
		})
	}
}

func TestEscBuilder_ResetStyle(t *testing.T) {
	e := newEscBuilder(64)
	e.ResetStyle()
	expected := "\x1b[0m"
	if string(e.Bytes()) != expected {
		t.Errorf("ResetStyle() = %q, want %q", e.Bytes(), expected)
	}
}

func TestEscBuilder_SetStyle_Bold(t *testing.T) {
	e := newEscBuilder(64)
	caps := Capabilities{Colors: Color256, TrueColor: false}
	style := NewStyle().Bold()
	e.SetStyle(style, caps)
	expected := "\x1b[0;1m"
	if string(e.Bytes()) != expected {
		t.Errorf("SetStyle(bold) = %q, want %q", e.Bytes(), expected)
	}
}

func TestEscBuilder_SetStyle_Attributes(t *testing.T) {
	type tc struct {
		style    Style
		expected string
	}

	tests := map[string]tc{
		"bold": {
			style:    NewStyle().Bold(),
			expected: "\x1b[0;1m",
		},
		"dim": {
			style:    NewStyle().Dim(),
			expected: "\x1b[0;2m",
		},
		"italic": {
			style:    NewStyle().Italic(),
			expected: "\x1b[0;3m",
		},
		"underline": {
			style:    NewStyle().Underline(),
			expected: "\x1b[0;4m",
		},
		"blink": {
			style:    NewStyle().Blink(),
			expected: "\x1b[0;5m",
		},
		"reverse": {
			style:    NewStyle().Reverse(),
			expected: "\x1b[0;7m",
		},
		"strikethrough": {
			style:    NewStyle().Strikethrough(),
			expected: "\x1b[0;9m",
		},
		"bold and italic": {
			style:    NewStyle().Bold().Italic(),
			expected: "\x1b[0;1;3m",
		},
		"all attributes": {
			style:    NewStyle().Bold().Dim().Italic().Underline().Blink().Reverse().Strikethrough(),
			expected: "\x1b[0;1;2;3;4;5;7;9m",
		},
		"no attributes": {
			style:    NewStyle(),
			expected: "\x1b[0m",
		},
	}

	caps := Capabilities{Colors: Color256}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			e.SetStyle(tt.style, caps)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("SetStyle(%v) = %q, want %q", tt.style, e.Bytes(), tt.expected)
			}
		})
	}
}

