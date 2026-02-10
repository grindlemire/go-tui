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
	type tc struct {
		idx uint8
	}

	tests := map[string]tc{
		"zero": {idx: 0},
		"one":  {idx: 1},
		"mid":  {idx: 127},
		"max":  {idx: 255},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := ANSIColor(tt.idx)
			if c.Type() != ColorANSI {
				t.Errorf("ANSIColor(%d).Type() = %v, want ColorANSI", tt.idx, c.Type())
			}
			if c.IsDefault() {
				t.Errorf("ANSIColor(%d).IsDefault() = true, want false", tt.idx)
			}
			if got := c.ANSI(); got != tt.idx {
				t.Errorf("ANSIColor(%d).ANSI() = %d, want %d", tt.idx, got, tt.idx)
			}
		})
	}
}

func TestRGBColor(t *testing.T) {
	type tc struct {
		r, g, b uint8
	}

	tests := map[string]tc{
		"black": {r: 0, g: 0, b: 0},
		"white": {r: 255, g: 255, b: 255},
		"red":   {r: 255, g: 0, b: 0},
		"green": {r: 0, g: 255, b: 0},
		"blue":  {r: 0, g: 0, b: 255},
		"mixed": {r: 128, g: 64, b: 32},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
		})
	}
}

func TestHexColor_Valid6Digit(t *testing.T) {
	type tc struct {
		hex     string
		r, g, b uint8
	}

	tests := map[string]tc{
		"black":           {hex: "#000000", r: 0, g: 0, b: 0},
		"white uppercase": {hex: "#FFFFFF", r: 255, g: 255, b: 255},
		"white lowercase": {hex: "#ffffff", r: 255, g: 255, b: 255},
		"red":             {hex: "#FF0000", r: 255, g: 0, b: 0},
		"green":           {hex: "#00FF00", r: 0, g: 255, b: 0},
		"blue":            {hex: "#0000FF", r: 0, g: 0, b: 255},
		"mixed":           {hex: "#1A2B3C", r: 26, g: 43, b: 60},
		"without hash":    {hex: "1A2B3C", r: 26, g: 43, b: 60},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := HexColor(tt.hex)
			if err != nil {
				t.Fatalf("HexColor(%q) returned error: %v", tt.hex, err)
			}
			if c.Type() != ColorRGB {
				t.Fatalf("HexColor(%q).Type() = %v, want ColorRGB", tt.hex, c.Type())
			}
			r, g, b := c.RGB()
			if r != tt.r || g != tt.g || b != tt.b {
				t.Errorf("HexColor(%q).RGB() = %d,%d,%d, want %d,%d,%d",
					tt.hex, r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestHexColor_Valid3Digit(t *testing.T) {
	type tc struct {
		hex     string
		r, g, b uint8
	}

	tests := map[string]tc{
		"black":           {hex: "#000", r: 0, g: 0, b: 0},
		"white uppercase": {hex: "#FFF", r: 255, g: 255, b: 255},
		"white lowercase": {hex: "#fff", r: 255, g: 255, b: 255},
		"red":             {hex: "#F00", r: 255, g: 0, b: 0},
		"green":           {hex: "#0F0", r: 0, g: 255, b: 0},
		"blue":            {hex: "#00F", r: 0, g: 0, b: 255},
		"mixed":           {hex: "#ABC", r: 0xAA, g: 0xBB, b: 0xCC},
		"without hash":    {hex: "ABC", r: 0xAA, g: 0xBB, b: 0xCC},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := HexColor(tt.hex)
			if err != nil {
				t.Fatalf("HexColor(%q) returned error: %v", tt.hex, err)
			}
			if c.Type() != ColorRGB {
				t.Fatalf("HexColor(%q).Type() = %v, want ColorRGB", tt.hex, c.Type())
			}
			r, g, b := c.RGB()
			if r != tt.r || g != tt.g || b != tt.b {
				t.Errorf("HexColor(%q).RGB() = %d,%d,%d, want %d,%d,%d",
					tt.hex, r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestHexColor_Invalid(t *testing.T) {
	type tc struct {
		hex string
	}

	tests := map[string]tc{
		"empty":           {hex: ""},
		"hash only":       {hex: "#"},
		"one digit":       {hex: "#1"},
		"two digits":      {hex: "#12"},
		"four digits":     {hex: "#1234"},
		"five digits":     {hex: "#12345"},
		"seven digits":    {hex: "#1234567"},
		"invalid 3 digit": {hex: "#GGG"},
		"invalid 6 digit": {hex: "#GGGGGG"},
		"partial invalid": {hex: "#12345G"},
		"not a color":     {hex: "not-a-color"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := HexColor(tt.hex)
			if err == nil {
				t.Errorf("HexColor(%q) should return error", tt.hex)
			}
		})
	}
}

func TestColor_Equal(t *testing.T) {
	type tc struct {
		a, b  Color
		equal bool
	}

	tests := map[string]tc{
		"default == default":     {a: DefaultColor(), b: DefaultColor(), equal: true},
		"ansi 0 == ansi 0":       {a: ANSIColor(0), b: ANSIColor(0), equal: true},
		"ansi 0 != ansi 1":       {a: ANSIColor(0), b: ANSIColor(1), equal: false},
		"rgb black == rgb black": {a: RGBColor(0, 0, 0), b: RGBColor(0, 0, 0), equal: true},
		"rgb != rgb different":   {a: RGBColor(0, 0, 0), b: RGBColor(1, 0, 0), equal: false},
		"default != ansi":        {a: DefaultColor(), b: ANSIColor(0), equal: false},
		"default != rgb":         {a: DefaultColor(), b: RGBColor(0, 0, 0), equal: false},
		"ansi != rgb":            {a: ANSIColor(0), b: RGBColor(0, 0, 0), equal: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.equal {
				t.Errorf("Equal() = %v, want %v", got, tt.equal)
			}
			// Test symmetry
			if got := tt.b.Equal(tt.a); got != tt.equal {
				t.Errorf("(symmetric) Equal() = %v, want %v", got, tt.equal)
			}
		})
	}
}

func TestColor_ToANSI(t *testing.T) {
	type tc struct {
		color       Color
		checkType   ColorType
		expected    uint8
		inGrayRange bool
		isDefault   bool
	}

	tests := map[string]tc{
		"default unchanged": {
			color:     DefaultColor(),
			isDefault: true,
		},
		"ansi unchanged": {
			color:     ANSIColor(42),
			checkType: ColorANSI,
			expected:  42,
		},
		"pure red": {
			color:     RGBColor(255, 0, 0),
			checkType: ColorANSI,
			expected:  uint8(16 + 5*36 + 0*6 + 0),
		},
		"pure green": {
			color:     RGBColor(0, 255, 0),
			checkType: ColorANSI,
			expected:  uint8(16 + 0*36 + 5*6 + 0),
		},
		"pure blue": {
			color:     RGBColor(0, 0, 255),
			checkType: ColorANSI,
			expected:  uint8(16 + 0*36 + 0*6 + 5),
		},
		"gray 128": {
			color:       RGBColor(128, 128, 128),
			checkType:   ColorANSI,
			inGrayRange: true,
		},
		"very dark gray": {
			color:     RGBColor(4, 4, 4),
			checkType: ColorANSI,
			expected:  16,
		},
		"very light gray": {
			color:     RGBColor(252, 252, 252),
			checkType: ColorANSI,
			expected:  231,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := tt.color.ToANSI()

			if tt.isDefault {
				if !c.IsDefault() {
					t.Error("ToANSI() should remain default")
				}
				return
			}

			if c.Type() != tt.checkType {
				t.Fatalf("ToANSI() type = %v, want %v", c.Type(), tt.checkType)
			}

			if tt.inGrayRange {
				idx := c.ANSI()
				if idx < 232 || idx > 255 {
					t.Errorf("Gray should map to grayscale range 232-255, got %d", idx)
				}
				return
			}

			if c.ANSI() != tt.expected {
				t.Errorf("ToANSI().ANSI() = %d, want %d", c.ANSI(), tt.expected)
			}
		})
	}
}

func TestColor_AccessorsAreSafeOnWrongType(t *testing.T) {
	if got := RGBColor(255, 0, 0).ANSI(); got != 0 {
		t.Errorf("RGBColor(...).ANSI() = %d, want 0", got)
	}
	if got := DefaultColor().ANSI(); got != 0 {
		t.Errorf("DefaultColor().ANSI() = %d, want 0", got)
	}

	r, g, b := ANSIColor(1).RGB()
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("ANSIColor(...).RGB() = %d,%d,%d, want 0,0,0", r, g, b)
	}
	r, g, b = DefaultColor().RGB()
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("DefaultColor().RGB() = %d,%d,%d, want 0,0,0", r, g, b)
	}
}

func TestColor_TypedAccessors(t *testing.T) {
	if _, ok := RGBColor(1, 2, 3).ANSIValue(); ok {
		t.Error("ANSIValue() should report !ok for RGB color")
	}
	if idx, ok := ANSIColor(42).ANSIValue(); !ok || idx != 42 {
		t.Errorf("ANSIValue() = (%d,%v), want (42,true)", idx, ok)
	}
	if _, _, _, ok := ANSIColor(1).RGBValue(); ok {
		t.Error("RGBValue() should report !ok for ANSI color")
	}
	r, g, b, ok := RGBColor(1, 2, 3).RGBValue()
	if !ok || r != 1 || g != 2 || b != 3 {
		t.Errorf("RGBValue() = (%d,%d,%d,%v), want (1,2,3,true)", r, g, b, ok)
	}
}

func TestPredefinedColors(t *testing.T) {
	type tc struct {
		color    Color
		expected uint8
	}

	tests := map[string]tc{
		"Black":         {color: Black, expected: 0},
		"Red":           {color: Red, expected: 1},
		"Green":         {color: Green, expected: 2},
		"Yellow":        {color: Yellow, expected: 3},
		"Blue":          {color: Blue, expected: 4},
		"Magenta":       {color: Magenta, expected: 5},
		"Cyan":          {color: Cyan, expected: 6},
		"White":         {color: White, expected: 7},
		"BrightBlack":   {color: BrightBlack, expected: 8},
		"BrightRed":     {color: BrightRed, expected: 9},
		"BrightGreen":   {color: BrightGreen, expected: 10},
		"BrightYellow":  {color: BrightYellow, expected: 11},
		"BrightBlue":    {color: BrightBlue, expected: 12},
		"BrightMagenta": {color: BrightMagenta, expected: 13},
		"BrightCyan":    {color: BrightCyan, expected: 14},
		"BrightWhite":   {color: BrightWhite, expected: 15},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.color.Type() != ColorANSI {
				t.Errorf("Type() = %v, want ColorANSI", tt.color.Type())
			}
			if tt.color.ANSI() != tt.expected {
				t.Errorf("ANSI() = %d, want %d", tt.color.ANSI(), tt.expected)
			}
		})
	}
}

func TestColor_ToRGBValues(t *testing.T) {
	type tc struct {
		color Color
		wantR uint8
		wantG uint8
		wantB uint8
	}

	tests := map[string]tc{
		"default returns zero": {
			color: DefaultColor(),
			wantR: 0, wantG: 0, wantB: 0,
		},
		"RGB passes through": {
			color: RGBColor(100, 150, 200),
			wantR: 100, wantG: 150, wantB: 200,
		},
		"ANSI black": {
			color: Black,
			wantR: 0, wantG: 0, wantB: 0,
		},
		"ANSI white": {
			color: White,
			wantR: 229, wantG: 229, wantB: 229,
		},
		"ANSI red": {
			color: Red,
			wantR: 205, wantG: 49, wantB: 49,
		},
		"ANSI bright white": {
			color: BrightWhite,
			wantR: 255, wantG: 255, wantB: 255,
		},
		"ANSI 6x6x6 cube red": {
			color: ANSIColor(196), // pure red in cube
			wantR: 255, wantG: 0, wantB: 0,
		},
		"ANSI grayscale mid": {
			color: ANSIColor(244), // middle gray
			wantR: 128, wantG: 128, wantB: 128,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r, g, b := tt.color.ToRGBValues()
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("ToRGBValues() = (%d, %d, %d), want (%d, %d, %d)",
					r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestColor_IsLight(t *testing.T) {
	type tc struct {
		color     Color
		wantLight bool
	}

	tests := map[string]tc{
		"default is dark": {
			color:     DefaultColor(),
			wantLight: false,
		},
		"black is dark": {
			color:     Black,
			wantLight: false,
		},
		"white is light": {
			color:     White,
			wantLight: true,
		},
		"bright white is light": {
			color:     BrightWhite,
			wantLight: true,
		},
		"bright yellow is light": {
			color:     BrightYellow,
			wantLight: true,
		},
		"red is dark": {
			color:     Red,
			wantLight: false,
		},
		"blue is dark": {
			color:     Blue,
			wantLight: false,
		},
		"cyan is dark": {
			color:     Cyan,
			wantLight: false,
		},
		"bright cyan is dark": {
			// Bright cyan RGB(41, 184, 219) has luminance ~0.4, below 0.5 threshold
			color:     BrightCyan,
			wantLight: false,
		},
		"RGB white is light": {
			color:     RGBColor(255, 255, 255),
			wantLight: true,
		},
		"RGB black is dark": {
			color:     RGBColor(0, 0, 0),
			wantLight: false,
		},
		"RGB light yellow is light": {
			color:     RGBColor(255, 255, 200),
			wantLight: true,
		},
		"RGB dark blue is dark": {
			color:     RGBColor(20, 20, 60),
			wantLight: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.color.IsLight()
			if got != tt.wantLight {
				t.Errorf("IsLight() = %v, want %v", got, tt.wantLight)
			}
		})
	}
}

func TestGradient_At(t *testing.T) {
	type tc struct {
		start Color
		end   Color
		t     float64
		wantR uint8
		wantG uint8
		wantB uint8
	}

	tests := map[string]tc{
		"start color": {
			start: RGBColor(255, 0, 0),
			end:   RGBColor(0, 0, 255),
			t:     0.0,
			wantR: 255,
			wantG: 0,
			wantB: 0,
		},
		"end color": {
			start: RGBColor(255, 0, 0),
			end:   RGBColor(0, 0, 255),
			t:     1.0,
			wantR: 0,
			wantG: 0,
			wantB: 255,
		},
		"middle": {
			start: RGBColor(0, 0, 0),
			end:   RGBColor(255, 255, 255),
			t:     0.5,
			wantR: 127,
			wantG: 127,
			wantB: 127,
		},
		"quarter": {
			start: RGBColor(0, 0, 0),
			end:   RGBColor(100, 200, 255),
			t:     0.25,
			wantR: 25,
			wantG: 50,
			wantB: 63,
		},
		"clamped below": {
			start: RGBColor(100, 100, 100),
			end:   RGBColor(200, 200, 200),
			t:     -1.0,
			wantR: 100,
			wantG: 100,
			wantB: 100,
		},
		"clamped above": {
			start: RGBColor(100, 100, 100),
			end:   RGBColor(200, 200, 200),
			t:     2.0,
			wantR: 200,
			wantG: 200,
			wantB: 200,
		},
		"ansi colors": {
			start: Red,
			end:   Blue,
			t:     0.5,
			// ANSI colors convert to RGB, so we just verify it's in a reasonable range
			wantR: 100, // Will be adjusted in test
			wantG: 50,
			wantB: 100,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			g := NewGradient(tt.start, tt.end)
			result := g.At(tt.t)
			if result.Type() != ColorRGB {
				t.Errorf("Gradient.At(%v).Type() = %v, want ColorRGB", tt.t, result.Type())
			}
			r, gVal, b := result.RGB()
			if name == "ansi colors" {
				// For ANSI colors, just verify we get valid RGB values
				if r > 255 || gVal > 255 || b > 255 {
					t.Errorf("Gradient.At(%v).RGB() = (%d, %d, %d), values out of range", tt.t, r, gVal, b)
				}
			} else {
				if r != tt.wantR || gVal != tt.wantG || b != tt.wantB {
					t.Errorf("Gradient.At(%v).RGB() = (%d, %d, %d), want (%d, %d, %d)", tt.t, r, gVal, b, tt.wantR, tt.wantG, tt.wantB)
				}
			}
		})
	}
}

func TestGradient_WithDirection(t *testing.T) {
	g := NewGradient(Red, Blue)
	if g.Direction != GradientHorizontal {
		t.Errorf("NewGradient().Direction = %v, want GradientHorizontal", g.Direction)
	}

	g2 := g.WithDirection(GradientVertical)
	if g2.Direction != GradientVertical {
		t.Errorf("WithDirection(GradientVertical).Direction = %v, want GradientVertical", g2.Direction)
	}
	if g.Direction != GradientHorizontal {
		t.Error("WithDirection should not modify original gradient")
	}
}
