package tui

import "testing"

func TestEscBuilder_ScreenSequences(t *testing.T) {
	type tc struct {
		fn       func(*escBuilder)
		expected string
	}

	tests := map[string]tc{
		"begin sync update": {
			fn:       func(e *escBuilder) { e.BeginSyncUpdate() },
			expected: "\x1b[?2026h",
		},
		"end sync update": {
			fn:       func(e *escBuilder) { e.EndSyncUpdate() },
			expected: "\x1b[?2026l",
		},
		"clear to end of screen": {
			fn:       func(e *escBuilder) { e.ClearToEndOfScreen() },
			expected: "\x1b[J",
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

func TestEscBuilder_Mouse(t *testing.T) {
	type tc struct {
		fn       func(*escBuilder)
		expected string
	}

	tests := map[string]tc{
		"enable mouse emits X10 then SGR": {
			fn:       func(e *escBuilder) { e.EnableMouse() },
			expected: "\x1b[?1000h\x1b[?1006h",
		},
		"disable mouse emits SGR then X10": {
			fn:       func(e *escBuilder) { e.DisableMouse() },
			expected: "\x1b[?1006l\x1b[?1000l",
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

func TestEscBuilder_KittyKeyboard(t *testing.T) {
	type tc struct {
		fn       func(*escBuilder)
		expected string
	}

	tests := map[string]tc{
		"push flags 1": {
			fn:       func(e *escBuilder) { e.KittyKeyboardPush(1) },
			expected: "\x1b[>1u",
		},
		"push flags 0": {
			fn:       func(e *escBuilder) { e.KittyKeyboardPush(0) },
			expected: "\x1b[>0u",
		},
		"push flags 31": {
			fn:       func(e *escBuilder) { e.KittyKeyboardPush(31) },
			expected: "\x1b[>31u",
		},
		"pop": {
			fn:       func(e *escBuilder) { e.KittyKeyboardPop() },
			expected: "\x1b[<u",
		},
		"query": {
			fn:       func(e *escBuilder) { e.KittyKeyboardQuery() },
			expected: "\x1b[?u",
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
