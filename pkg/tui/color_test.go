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
		"zero":    {idx: 0},
		"one":     {idx: 1},
		"mid":     {idx: 127},
		"max":     {idx: 255},
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
		"black":   {r: 0, g: 0, b: 0},
		"white":   {r: 255, g: 255, b: 255},
		"red":     {r: 255, g: 0, b: 0},
		"green":   {r: 0, g: 255, b: 0},
		"blue":    {r: 0, g: 0, b: 255},
		"mixed":   {r: 128, g: 64, b: 32},
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
		"black":            {hex: "#000000", r: 0, g: 0, b: 0},
		"white uppercase":  {hex: "#FFFFFF", r: 255, g: 255, b: 255},
		"white lowercase":  {hex: "#ffffff", r: 255, g: 255, b: 255},
		"red":              {hex: "#FF0000", r: 255, g: 0, b: 0},
		"green":            {hex: "#00FF00", r: 0, g: 255, b: 0},
		"blue":             {hex: "#0000FF", r: 0, g: 0, b: 255},
		"mixed":            {hex: "#1A2B3C", r: 26, g: 43, b: 60},
		"without hash":     {hex: "1A2B3C", r: 26, g: 43, b: 60},
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
		"empty":            {hex: ""},
		"hash only":        {hex: "#"},
		"one digit":        {hex: "#1"},
		"two digits":       {hex: "#12"},
		"four digits":      {hex: "#1234"},
		"five digits":      {hex: "#12345"},
		"seven digits":     {hex: "#1234567"},
		"invalid 3 digit":  {hex: "#GGG"},
		"invalid 6 digit":  {hex: "#GGGGGG"},
		"partial invalid":  {hex: "#12345G"},
		"not a color":      {hex: "not-a-color"},
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
		"default == default":    {a: DefaultColor(), b: DefaultColor(), equal: true},
		"ansi 0 == ansi 0":      {a: ANSIColor(0), b: ANSIColor(0), equal: true},
		"ansi 0 != ansi 1":      {a: ANSIColor(0), b: ANSIColor(1), equal: false},
		"rgb black == rgb black": {a: RGBColor(0, 0, 0), b: RGBColor(0, 0, 0), equal: true},
		"rgb != rgb different":  {a: RGBColor(0, 0, 0), b: RGBColor(1, 0, 0), equal: false},
		"default != ansi":       {a: DefaultColor(), b: ANSIColor(0), equal: false},
		"default != rgb":        {a: DefaultColor(), b: RGBColor(0, 0, 0), equal: false},
		"ansi != rgb":           {a: ANSIColor(0), b: RGBColor(0, 0, 0), equal: false},
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
