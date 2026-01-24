package tui

import (
	"testing"
)

func TestDefaultColor(t *testing.T) {
	c := DefaultColor()
	if c.Type() != ColorDefault {
		t.Errorf("DefaultColor().Type() = %v, want ColorDefault", c.Type())
	}
	if !c.IsDefault() {
		t.Error("DefaultColor().IsDefault() = false, want true")
	}
}

func TestANSIColor(t *testing.T) {
	tests := []uint8{0, 1, 127, 255}
	for _, idx := range tests {
		c := ANSIColor(idx)
		if c.Type() != ColorANSI {
			t.Errorf("ANSIColor(%d).Type() = %v, want ColorANSI", idx, c.Type())
		}
		if c.IsDefault() {
			t.Errorf("ANSIColor(%d).IsDefault() = true, want false", idx)
		}
		if got := c.ANSI(); got != idx {
			t.Errorf("ANSIColor(%d).ANSI() = %d, want %d", idx, got, idx)
		}
	}
}

func TestRGBColor(t *testing.T) {
	tests := []struct {
		r, g, b uint8
	}{
		{0, 0, 0},
		{255, 255, 255},
		{255, 0, 0},
		{0, 255, 0},
		{0, 0, 255},
		{128, 64, 32},
	}
	for _, tt := range tests {
		c := RGBColor(tt.r, tt.g, tt.b)
		if c.Type() != ColorRGB {
			t.Errorf("RGBColor(%d,%d,%d).Type() = %v, want ColorRGB", tt.r, tt.g, tt.b, c.Type())
		}
		if c.IsDefault() {
			t.Errorf("RGBColor(%d,%d,%d).IsDefault() = true, want false", tt.r, tt.g, tt.b)
		}
		r, g, b := c.RGB()
		if r != tt.r || g != tt.g || b != tt.b {
			t.Errorf("RGBColor(%d,%d,%d).RGB() = %d,%d,%d, want %d,%d,%d",
				tt.r, tt.g, tt.b, r, g, b, tt.r, tt.g, tt.b)
		}
	}
}

func TestHexColor_Valid6Digit(t *testing.T) {
	tests := []struct {
		hex     string
		r, g, b uint8
	}{
		{"#000000", 0, 0, 0},
		{"#FFFFFF", 255, 255, 255},
		{"#ffffff", 255, 255, 255},
		{"#FF0000", 255, 0, 0},
		{"#00FF00", 0, 255, 0},
		{"#0000FF", 0, 0, 255},
		{"#1A2B3C", 26, 43, 60},
		{"1A2B3C", 26, 43, 60}, // without #
	}
	for _, tt := range tests {
		c, err := HexColor(tt.hex)
		if err != nil {
			t.Errorf("HexColor(%q) returned error: %v", tt.hex, err)
			continue
		}
		if c.Type() != ColorRGB {
			t.Errorf("HexColor(%q).Type() = %v, want ColorRGB", tt.hex, c.Type())
			continue
		}
		r, g, b := c.RGB()
		if r != tt.r || g != tt.g || b != tt.b {
			t.Errorf("HexColor(%q).RGB() = %d,%d,%d, want %d,%d,%d",
				tt.hex, r, g, b, tt.r, tt.g, tt.b)
		}
	}
}

func TestHexColor_Valid3Digit(t *testing.T) {
	tests := []struct {
		hex     string
		r, g, b uint8
	}{
		{"#000", 0, 0, 0},
		{"#FFF", 255, 255, 255},
		{"#fff", 255, 255, 255},
		{"#F00", 255, 0, 0},
		{"#0F0", 0, 255, 0},
		{"#00F", 0, 0, 255},
		{"#ABC", 0xAA, 0xBB, 0xCC},
		{"ABC", 0xAA, 0xBB, 0xCC}, // without #
	}
	for _, tt := range tests {
		c, err := HexColor(tt.hex)
		if err != nil {
			t.Errorf("HexColor(%q) returned error: %v", tt.hex, err)
			continue
		}
		if c.Type() != ColorRGB {
			t.Errorf("HexColor(%q).Type() = %v, want ColorRGB", tt.hex, c.Type())
			continue
		}
		r, g, b := c.RGB()
		if r != tt.r || g != tt.g || b != tt.b {
			t.Errorf("HexColor(%q).RGB() = %d,%d,%d, want %d,%d,%d",
				tt.hex, r, g, b, tt.r, tt.g, tt.b)
		}
	}
}

func TestHexColor_Invalid(t *testing.T) {
	invalids := []string{
		"",
		"#",
		"#1",
		"#12",
		"#1234",
		"#12345",
		"#1234567",
		"#GGG",
		"#GGGGGG",
		"#12345G",
		"not-a-color",
	}
	for _, hex := range invalids {
		_, err := HexColor(hex)
		if err == nil {
			t.Errorf("HexColor(%q) should return error", hex)
		}
	}
}

func TestColor_Equal(t *testing.T) {
	tests := []struct {
		name  string
		a, b  Color
		equal bool
	}{
		{"default == default", DefaultColor(), DefaultColor(), true},
		{"ansi 0 == ansi 0", ANSIColor(0), ANSIColor(0), true},
		{"ansi 0 != ansi 1", ANSIColor(0), ANSIColor(1), false},
		{"rgb black == rgb black", RGBColor(0, 0, 0), RGBColor(0, 0, 0), true},
		{"rgb != rgb different", RGBColor(0, 0, 0), RGBColor(1, 0, 0), false},
		{"default != ansi", DefaultColor(), ANSIColor(0), false},
		{"default != rgb", DefaultColor(), RGBColor(0, 0, 0), false},
		{"ansi != rgb", ANSIColor(0), RGBColor(0, 0, 0), false},
	}
	for _, tt := range tests {
		if got := tt.a.Equal(tt.b); got != tt.equal {
			t.Errorf("%s: Equal() = %v, want %v", tt.name, got, tt.equal)
		}
		// Test symmetry
		if got := tt.b.Equal(tt.a); got != tt.equal {
			t.Errorf("%s (symmetric): Equal() = %v, want %v", tt.name, got, tt.equal)
		}
	}
}

func TestColor_ToANSI(t *testing.T) {
	// Default color should remain default
	t.Run("default unchanged", func(t *testing.T) {
		c := DefaultColor().ToANSI()
		if !c.IsDefault() {
			t.Error("DefaultColor().ToANSI() should remain default")
		}
	})

	// ANSI color should remain unchanged
	t.Run("ansi unchanged", func(t *testing.T) {
		c := ANSIColor(42).ToANSI()
		if c.Type() != ColorANSI || c.ANSI() != 42 {
			t.Errorf("ANSIColor(42).ToANSI() = %v, want ANSIColor(42)", c)
		}
	})

	// Pure red should map to red in color cube
	t.Run("pure red", func(t *testing.T) {
		c := RGBColor(255, 0, 0).ToANSI()
		if c.Type() != ColorANSI {
			t.Fatalf("ToANSI() type = %v, want ColorANSI", c.Type())
		}
		// Pure red (255, 0, 0) should map to color cube index 196 (5*36 + 0*6 + 0 + 16)
		expected := uint8(16 + 5*36 + 0*6 + 0)
		if c.ANSI() != expected {
			t.Errorf("RGBColor(255,0,0).ToANSI().ANSI() = %d, want %d", c.ANSI(), expected)
		}
	})

	// Pure green
	t.Run("pure green", func(t *testing.T) {
		c := RGBColor(0, 255, 0).ToANSI()
		if c.Type() != ColorANSI {
			t.Fatalf("ToANSI() type = %v, want ColorANSI", c.Type())
		}
		expected := uint8(16 + 0*36 + 5*6 + 0)
		if c.ANSI() != expected {
			t.Errorf("RGBColor(0,255,0).ToANSI().ANSI() = %d, want %d", c.ANSI(), expected)
		}
	})

	// Pure blue
	t.Run("pure blue", func(t *testing.T) {
		c := RGBColor(0, 0, 255).ToANSI()
		if c.Type() != ColorANSI {
			t.Fatalf("ToANSI() type = %v, want ColorANSI", c.Type())
		}
		expected := uint8(16 + 0*36 + 0*6 + 5)
		if c.ANSI() != expected {
			t.Errorf("RGBColor(0,0,255).ToANSI().ANSI() = %d, want %d", c.ANSI(), expected)
		}
	})

	// Gray values should use grayscale ramp
	t.Run("gray 128", func(t *testing.T) {
		c := RGBColor(128, 128, 128).ToANSI()
		if c.Type() != ColorANSI {
			t.Fatalf("ToANSI() type = %v, want ColorANSI", c.Type())
		}
		// Should be in the grayscale range 232-255
		idx := c.ANSI()
		if idx < 232 || idx > 255 {
			t.Errorf("Gray should map to grayscale range, got %d", idx)
		}
	})

	// Very dark gray should map to color cube black
	t.Run("very dark gray", func(t *testing.T) {
		c := RGBColor(4, 4, 4).ToANSI()
		if c.Type() != ColorANSI {
			t.Fatalf("ToANSI() type = %v, want ColorANSI", c.Type())
		}
		// Very dark should use color cube black (16)
		if c.ANSI() != 16 {
			t.Errorf("Very dark gray should map to 16, got %d", c.ANSI())
		}
	})

	// Very light gray should map to color cube white
	t.Run("very light gray", func(t *testing.T) {
		c := RGBColor(252, 252, 252).ToANSI()
		if c.Type() != ColorANSI {
			t.Fatalf("ToANSI() type = %v, want ColorANSI", c.Type())
		}
		// Very light should use color cube white (231)
		if c.ANSI() != 231 {
			t.Errorf("Very light gray should map to 231, got %d", c.ANSI())
		}
	})
}

func TestColor_ANSIPanicOnRGB(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Color.ANSI() on RGB color should panic")
		}
	}()
	RGBColor(255, 0, 0).ANSI()
}

func TestColor_ANSIPanicOnDefault(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Color.ANSI() on default color should panic")
		}
	}()
	DefaultColor().ANSI()
}

func TestColor_RGBPanicOnANSI(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Color.RGB() on ANSI color should panic")
		}
	}()
	ANSIColor(1).RGB()
}

func TestColor_RGBPanicOnDefault(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Color.RGB() on default color should panic")
		}
	}()
	DefaultColor().RGB()
}

func TestPredefinedColors(t *testing.T) {
	colors := []struct {
		name     string
		color    Color
		expected uint8
	}{
		{"Black", Black, 0},
		{"Red", Red, 1},
		{"Green", Green, 2},
		{"Yellow", Yellow, 3},
		{"Blue", Blue, 4},
		{"Magenta", Magenta, 5},
		{"Cyan", Cyan, 6},
		{"White", White, 7},
		{"BrightBlack", BrightBlack, 8},
		{"BrightRed", BrightRed, 9},
		{"BrightGreen", BrightGreen, 10},
		{"BrightYellow", BrightYellow, 11},
		{"BrightBlue", BrightBlue, 12},
		{"BrightMagenta", BrightMagenta, 13},
		{"BrightCyan", BrightCyan, 14},
		{"BrightWhite", BrightWhite, 15},
	}
	for _, tt := range colors {
		if tt.color.Type() != ColorANSI {
			t.Errorf("%s.Type() = %v, want ColorANSI", tt.name, tt.color.Type())
		}
		if tt.color.ANSI() != tt.expected {
			t.Errorf("%s.ANSI() = %d, want %d", tt.name, tt.color.ANSI(), tt.expected)
		}
	}
}
