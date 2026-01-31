package tui

import (
	"testing"
)

func TestEscBuilder_SetStyle_ANSIForeground(t *testing.T) {
	type tc struct {
		color    Color
		caps     Capabilities
		expected string
	}

	tests := map[string]tc{
		"basic red": {
			color:    Red,
			caps:     Capabilities{Colors: Color16},
			expected: "\x1b[0;31m",
		},
		"basic green": {
			color:    Green,
			caps:     Capabilities{Colors: Color16},
			expected: "\x1b[0;32m",
		},
		"bright red": {
			color:    BrightRed,
			caps:     Capabilities{Colors: Color16},
			expected: "\x1b[0;91m",
		},
		"bright white": {
			color:    BrightWhite,
			caps:     Capabilities{Colors: Color16},
			expected: "\x1b[0;97m",
		},
		"color 256 index 200": {
			color:    ANSIColor(200),
			caps:     Capabilities{Colors: Color256},
			expected: "\x1b[0;38;5;200m",
		},
		"color 256 index 42": {
			color:    ANSIColor(42),
			caps:     Capabilities{Colors: Color256},
			expected: "\x1b[0;38;5;42m",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			style := NewStyle().Foreground(tt.color)
			e.SetStyle(style, tt.caps)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("SetStyle(fg=%v) = %q, want %q", tt.color, e.Bytes(), tt.expected)
			}
		})
	}
}

func TestEscBuilder_SetStyle_RGBForeground(t *testing.T) {
	type tc struct {
		r, g, b  uint8
		caps     Capabilities
		expected string
	}

	tests := map[string]tc{
		"true color red": {
			r: 255, g: 0, b: 0,
			caps:     Capabilities{Colors: ColorTrue, TrueColor: true},
			expected: "\x1b[0;38;2;255;0;0m",
		},
		"true color green": {
			r: 0, g: 255, b: 0,
			caps:     Capabilities{Colors: ColorTrue, TrueColor: true},
			expected: "\x1b[0;38;2;0;255;0m",
		},
		"true color mixed": {
			r: 128, g: 64, b: 192,
			caps:     Capabilities{Colors: ColorTrue, TrueColor: true},
			expected: "\x1b[0;38;2;128;64;192m",
		},
		"fallback to 256 when no true color": {
			r: 255, g: 0, b: 0,
			caps:     Capabilities{Colors: Color256, TrueColor: false},
			expected: "\x1b[0;38;5;196m", // Red approximation in 256 color
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			style := NewStyle().Foreground(RGBColor(tt.r, tt.g, tt.b))
			e.SetStyle(style, tt.caps)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("SetStyle(fg=RGB(%d,%d,%d)) = %q, want %q", tt.r, tt.g, tt.b, e.Bytes(), tt.expected)
			}
		})
	}
}

func TestEscBuilder_SetStyle_Background(t *testing.T) {
	type tc struct {
		color    Color
		caps     Capabilities
		expected string
	}

	tests := map[string]tc{
		"basic blue background": {
			color:    Blue,
			caps:     Capabilities{Colors: Color16},
			expected: "\x1b[0;44m",
		},
		"bright cyan background": {
			color:    BrightCyan,
			caps:     Capabilities{Colors: Color16},
			expected: "\x1b[0;106m",
		},
		"256 color background": {
			color:    ANSIColor(128),
			caps:     Capabilities{Colors: Color256},
			expected: "\x1b[0;48;5;128m",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			style := NewStyle().Background(tt.color)
			e.SetStyle(style, tt.caps)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("SetStyle(bg=%v) = %q, want %q", tt.color, e.Bytes(), tt.expected)
			}
		})
	}
}

func TestEscBuilder_SetStyle_RGBBackground(t *testing.T) {
	e := newEscBuilder(64)
	caps := Capabilities{Colors: ColorTrue, TrueColor: true}
	style := NewStyle().Background(RGBColor(100, 150, 200))
	e.SetStyle(style, caps)
	expected := "\x1b[0;48;2;100;150;200m"
	if string(e.Bytes()) != expected {
		t.Errorf("SetStyle(bg=RGB) = %q, want %q", e.Bytes(), expected)
	}
}

func TestEscBuilder_SetStyle_Combined(t *testing.T) {
	type tc struct {
		style    Style
		caps     Capabilities
		expected string
	}

	tests := map[string]tc{
		"bold red on blue": {
			style:    NewStyle().Bold().Foreground(Red).Background(Blue),
			caps:     Capabilities{Colors: Color16},
			expected: "\x1b[0;1;31;44m",
		},
		"italic with 256 colors": {
			style:    NewStyle().Italic().Foreground(ANSIColor(196)).Background(ANSIColor(21)),
			caps:     Capabilities{Colors: Color256},
			expected: "\x1b[0;3;38;5;196;48;5;21m",
		},
		"full style with true color": {
			style:    NewStyle().Bold().Underline().Foreground(RGBColor(255, 100, 50)).Background(RGBColor(0, 0, 128)),
			caps:     Capabilities{Colors: ColorTrue, TrueColor: true},
			expected: "\x1b[0;1;4;38;2;255;100;50;48;2;0;0;128m",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(128)
			e.SetStyle(tt.style, tt.caps)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("SetStyle() = %q, want %q", e.Bytes(), tt.expected)
			}
		})
	}
}

func TestEscBuilder_SetStyle_CapabilityFallback(t *testing.T) {
	type tc struct {
		style    Style
		caps     Capabilities
		desc     string
		contains string
	}

	tests := map[string]tc{
		"RGB falls back to 256 color": {
			style:    NewStyle().Foreground(RGBColor(255, 0, 0)),
			caps:     Capabilities{Colors: Color256, TrueColor: false},
			desc:     "should use 38;5 instead of 38;2",
			contains: "38;5;",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			e.SetStyle(tt.style, tt.caps)
			result := string(e.Bytes())
			if !containsSubstring(result, tt.contains) {
				t.Errorf("SetStyle() = %q, should contain %q (%s)", result, tt.contains, tt.desc)
			}
		})
	}
}

func TestEscBuilder_WriteRune(t *testing.T) {
	type tc struct {
		r        rune
		expected string
	}

	tests := map[string]tc{
		"ASCII letter": {
			r:        'A',
			expected: "A",
		},
		"space": {
			r:        ' ',
			expected: " ",
		},
		"CJK character": {
			r:        '中',
			expected: "中",
		},
		"emoji": {
			r:        '😀',
			expected: "😀",
		},
		"box drawing": {
			r:        '┌',
			expected: "┌",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := newEscBuilder(64)
			e.WriteRune(tt.r)
			if string(e.Bytes()) != tt.expected {
				t.Errorf("WriteRune(%q) = %q, want %q", tt.r, e.Bytes(), tt.expected)
			}
		})
	}
}

func TestEscBuilder_Reset(t *testing.T) {
	e := newEscBuilder(64)
	e.WriteString("hello")
	if e.Len() != 5 {
		t.Errorf("Len() = %d, want 5", e.Len())
	}

	e.Reset()
	if e.Len() != 0 {
		t.Errorf("after Reset(), Len() = %d, want 0", e.Len())
	}

	e.WriteString("world")
	if string(e.Bytes()) != "world" {
		t.Errorf("after Reset() and write, got %q, want %q", e.Bytes(), "world")
	}
}

func TestEscBuilder_WriteString(t *testing.T) {
	e := newEscBuilder(64)
	e.WriteString("Hello, World!")
	if string(e.Bytes()) != "Hello, World!" {
		t.Errorf("WriteString() = %q, want %q", e.Bytes(), "Hello, World!")
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
