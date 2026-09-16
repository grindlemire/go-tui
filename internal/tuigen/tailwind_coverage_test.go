package tuigen

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tailwind"
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

// Every class the shared table knows must render to Go source without hitting
// the renderer's unhandled-op panic. Parameterized forms are sampled.
func TestRenderCoversEveryClass(t *testing.T) {
	classes := append(tailwind.StaticClasses(),
		"gap-1", "p-2", "px-1", "py-1", "pt-1", "pr-1", "pb-1", "pl-1",
		"m-2", "mx-1", "my-1", "mt-1", "mr-1", "mb-1", "ml-1",
		"w-3", "h-3", "min-w-1", "max-w-1", "min-h-1", "max-h-1",
		"w-1/2", "h-1/3", "w-full", "w-auto", "h-full", "h-auto",
		"flex-grow-2", "flex-shrink-2",
		"text-[#abc]", "bg-[#abcdef]", "border-[#123]", "scrollbar-[#123]", "scrollbar-thumb-[#123]",
		"text-gradient-red-blue", "bg-gradient-red-blue-v", "border-gradient-bright-red-bright-blue-dd",
	)
	for _, class := range classes {
		m, ok := ParseTailwindClass(class)
		if !ok {
			t.Errorf("ParseTailwindClass(%q) not ok", class)
			continue
		}
		code := m.Option
		if m.IsTextStyle {
			code = m.TextMethod
		}
		if code == "" || !strings.Contains(code, "(") {
			t.Errorf("ParseTailwindClass(%q) rendered %q", class, code)
		}
	}
}

func TestColorExpr(t *testing.T) {
	type tc struct {
		color tailwind.Color
		want  string
	}

	tests := map[string]tc{
		"black":        {color: tailwind.Named("black"), want: "tui.Black"},
		"red":          {color: tailwind.Named("red"), want: "tui.Red"},
		"bright-white": {color: tailwind.Named("bright-white"), want: "tui.BrightWhite"},
		"unknown":      {color: tailwind.Named("chartreuse"), want: "tui.Black"},
		"rgb":          {color: tailwind.RGB(255, 128, 0), want: "tui.RGBColor(255, 128, 0)"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := colorExpr(tt.color); got != tt.want {
				t.Errorf("colorExpr(%+v) = %q, want %q", tt.color, got, tt.want)
			}
		})
	}
}
