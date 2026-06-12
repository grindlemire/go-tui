package tuigen

import (
	"testing"
)

func TestParseTailwindClass_HexColors(t *testing.T) {
	type tc struct {
		class       string
		wantOption  string
		wantText    string
		isTextStyle bool
	}

	tests := map[string]tc{
		"text hex six digits": {
			class:       "text-[#ff0000]",
			isTextStyle: true,
			wantText:    "Foreground(tui.RGBColor(255, 0, 0))",
		},
		"text hex shorthand": {
			class:       "text-[#abc]",
			isTextStyle: true,
			wantText:    "Foreground(tui.RGBColor(170, 187, 204))",
		},
		"bg hex": {
			class:      "bg-[#00ff00]",
			wantOption: "tui.WithBackground(tui.NewStyle().Background(tui.RGBColor(0, 255, 0)))",
		},
		"border hex": {
			class:      "border-[#0000ff]",
			wantOption: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.RGBColor(0, 0, 255)))",
		},
		"scrollbar hex": {
			class:      "scrollbar-[#102030]",
			wantOption: "tui.WithScrollbarStyle(tui.NewStyle().Foreground(tui.RGBColor(16, 32, 48)))",
		},
		"scrollbar thumb hex": {
			class:      "scrollbar-thumb-[#a0b0c0]",
			wantOption: "tui.WithScrollbarThumbStyle(tui.NewStyle().Foreground(tui.RGBColor(160, 176, 192)))",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.class)
			if !ok {
				t.Fatalf("ParseTailwindClass(%q) not recognized", tt.class)
			}
			if mapping.IsTextStyle != tt.isTextStyle {
				t.Errorf("IsTextStyle = %v, want %v", mapping.IsTextStyle, tt.isTextStyle)
			}
			if tt.isTextStyle {
				if mapping.TextMethod != tt.wantText {
					t.Errorf("TextMethod = %q, want %q", mapping.TextMethod, tt.wantText)
				}
			} else if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
		})
	}
}

func TestParseTailwindClass_GradientDirections(t *testing.T) {
	type tc struct {
		class      string
		wantOption string
	}

	tests := map[string]tc{
		"bg gradient vertical": {
			class:      "bg-gradient-red-blue-v",
			wantOption: "tui.WithBackgroundGradient(tui.NewGradient(tui.Red, tui.Blue).WithDirection(tui.GradientVertical))",
		},
		"bg gradient diagonal down": {
			class:      "bg-gradient-green-yellow-dd",
			wantOption: "tui.WithBackgroundGradient(tui.NewGradient(tui.Green, tui.Yellow).WithDirection(tui.GradientDiagonalDown))",
		},
		"bg gradient diagonal up": {
			class:      "bg-gradient-cyan-magenta-du",
			wantOption: "tui.WithBackgroundGradient(tui.NewGradient(tui.Cyan, tui.Magenta).WithDirection(tui.GradientDiagonalUp))",
		},
		"border gradient vertical": {
			class:      "border-gradient-white-black-v",
			wantOption: "tui.WithBorderGradient(tui.NewGradient(tui.White, tui.Black).WithDirection(tui.GradientVertical))",
		},
		"border gradient diagonal down": {
			class:      "border-gradient-red-cyan-dd",
			wantOption: "tui.WithBorderGradient(tui.NewGradient(tui.Red, tui.Cyan).WithDirection(tui.GradientDiagonalDown))",
		},
		"border gradient diagonal up": {
			class:      "border-gradient-blue-green-du",
			wantOption: "tui.WithBorderGradient(tui.NewGradient(tui.Blue, tui.Green).WithDirection(tui.GradientDiagonalUp))",
		},
		"text gradient diagonal down": {
			class:      "text-gradient-yellow-magenta-dd",
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.Yellow, tui.Magenta).WithDirection(tui.GradientDiagonalDown))",
		},
		"text gradient diagonal up": {
			class:      "text-gradient-black-white-du",
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.Black, tui.White).WithDirection(tui.GradientDiagonalUp))",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.class)
			if !ok {
				t.Fatalf("ParseTailwindClass(%q) not recognized", tt.class)
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
		})
	}
}

func TestParseTailwindClass_GradientColorParsing(t *testing.T) {
	type tc struct {
		class      string
		wantOption string
	}

	tests := map[string]tc{
		"text gradient with bright colors no direction": {
			class:      "text-gradient-bright-red-bright-blue",
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.BrightRed, tui.BrightBlue).WithDirection(tui.GradientHorizontal))",
		},
		"bg gradient bright color suffix match": {
			class:      "bg-gradient-bright-green-bright-magenta",
			wantOption: "tui.WithBackgroundGradient(tui.NewGradient(tui.BrightGreen, tui.BrightMagenta).WithDirection(tui.GradientHorizontal))",
		},
		"border gradient bright color suffix match": {
			class:      "border-gradient-bright-cyan-bright-yellow",
			wantOption: "tui.WithBorderGradient(tui.NewGradient(tui.BrightCyan, tui.BrightYellow).WithDirection(tui.GradientHorizontal))",
		},
		// Unknown color names fall back to splitting on the last hyphen and
		// then default to tui.Black for unrecognized names.
		"bg gradient unknown colors fall back to black": {
			class:      "bg-gradient-orange-pink",
			wantOption: "tui.WithBackgroundGradient(tui.NewGradient(tui.Black, tui.Black).WithDirection(tui.GradientHorizontal))",
		},
		"border gradient unknown colors fall back to black": {
			class:      "border-gradient-orange-pink",
			wantOption: "tui.WithBorderGradient(tui.NewGradient(tui.Black, tui.Black).WithDirection(tui.GradientHorizontal))",
		},
		"text gradient unknown colors fall back to black": {
			class:      "text-gradient-orange-pink",
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.Black, tui.Black).WithDirection(tui.GradientHorizontal))",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.class)
			if !ok {
				t.Fatalf("ParseTailwindClass(%q) not recognized", tt.class)
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
		})
	}
}

func TestColorNameToColor(t *testing.T) {
	type tc struct {
		name string
		want string
	}

	tests := map[string]tc{
		"red":            {name: "red", want: "tui.Red"},
		"green":          {name: "green", want: "tui.Green"},
		"blue":           {name: "blue", want: "tui.Blue"},
		"cyan":           {name: "cyan", want: "tui.Cyan"},
		"magenta":        {name: "magenta", want: "tui.Magenta"},
		"yellow":         {name: "yellow", want: "tui.Yellow"},
		"white":          {name: "white", want: "tui.White"},
		"black":          {name: "black", want: "tui.Black"},
		"bright-red":     {name: "bright-red", want: "tui.BrightRed"},
		"bright-green":   {name: "bright-green", want: "tui.BrightGreen"},
		"bright-blue":    {name: "bright-blue", want: "tui.BrightBlue"},
		"bright-cyan":    {name: "bright-cyan", want: "tui.BrightCyan"},
		"bright-magenta": {name: "bright-magenta", want: "tui.BrightMagenta"},
		"bright-yellow":  {name: "bright-yellow", want: "tui.BrightYellow"},
		"bright-white":   {name: "bright-white", want: "tui.BrightWhite"},
		"bright-black":   {name: "bright-black", want: "tui.BrightBlack"},
		"unknown":        {name: "chartreuse", want: "tui.Black"},
		"empty":          {name: "", want: "tui.Black"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := colorNameToColor(tt.name); got != tt.want {
				t.Errorf("colorNameToColor(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestParseHexToRGB(t *testing.T) {
	type tc struct {
		hex    string
		wantR  uint8
		wantG  uint8
		wantB  uint8
		wantOK bool
	}

	tests := map[string]tc{
		"six digit":            {hex: "ff8000", wantR: 255, wantG: 128, wantB: 0, wantOK: true},
		"three digit expanded": {hex: "f80", wantR: 255, wantG: 136, wantB: 0, wantOK: true},
		"wrong length":         {hex: "ffff", wantOK: false},
		"empty":                {hex: "", wantOK: false},
		"bad red component":    {hex: "zzff00", wantOK: false},
		"bad green component":  {hex: "ffzz00", wantOK: false},
		"bad blue component":   {hex: "ff00zz", wantOK: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r, g, b, ok := parseHexToRGB(tt.hex)
			if ok != tt.wantOK {
				t.Fatalf("parseHexToRGB(%q) ok = %v, want %v", tt.hex, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("parseHexToRGB(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.hex, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}
