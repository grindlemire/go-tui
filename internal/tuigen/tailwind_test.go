package tuigen

import (
	"testing"
)

func TestParseTailwindClass_Layout(t *testing.T) {
	type tc struct {
		input       string
		wantOK      bool
		wantOption  string
		wantImport  string
	}

	tests := map[string]tc{
		"flex": {
			input:      "flex",
			wantOK:     true,
			wantOption: "tui.WithDirection(tui.Row)",
			wantImport: "tui",
		},
		"flex-row": {
			input:      "flex-row",
			wantOK:     true,
			wantOption: "tui.WithDirection(tui.Row)",
			wantImport: "tui",
		},
		"flex-col": {
			input:      "flex-col",
			wantOK:     true,
			wantOption: "tui.WithDirection(tui.Column)",
			wantImport: "tui",
		},
		"flex-grow": {
			input:      "flex-grow",
			wantOK:     true,
			wantOption: "tui.WithFlexGrow(1)",
		},
		"flex-shrink": {
			input:      "flex-shrink",
			wantOK:     true,
			wantOption: "tui.WithFlexShrink(1)",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if !ok {
				return
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
			if mapping.NeedsImport != tt.wantImport {
				t.Errorf("NeedsImport = %q, want %q", mapping.NeedsImport, tt.wantImport)
			}
		})
	}
}

func TestParseTailwindClass_Alignment(t *testing.T) {
	type tc struct {
		input       string
		wantOK      bool
		wantOption  string
	}

	tests := map[string]tc{
		"justify-start": {
			input:      "justify-start",
			wantOK:     true,
			wantOption: "tui.WithJustify(tui.JustifyStart)",
		},
		"justify-center": {
			input:      "justify-center",
			wantOK:     true,
			wantOption: "tui.WithJustify(tui.JustifyCenter)",
		},
		"justify-end": {
			input:      "justify-end",
			wantOK:     true,
			wantOption: "tui.WithJustify(tui.JustifyEnd)",
		},
		"justify-between": {
			input:      "justify-between",
			wantOK:     true,
			wantOption: "tui.WithJustify(tui.JustifySpaceBetween)",
		},
		"items-start": {
			input:      "items-start",
			wantOK:     true,
			wantOption: "tui.WithAlign(tui.AlignStart)",
		},
		"items-center": {
			input:      "items-center",
			wantOK:     true,
			wantOption: "tui.WithAlign(tui.AlignCenter)",
		},
		"items-end": {
			input:      "items-end",
			wantOK:     true,
			wantOption: "tui.WithAlign(tui.AlignEnd)",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
		})
	}
}

func TestParseTailwindClass_DynamicSpacing(t *testing.T) {
	type tc struct {
		input       string
		wantOK      bool
		wantOption  string
	}

	tests := map[string]tc{
		"gap-1": {
			input:      "gap-1",
			wantOK:     true,
			wantOption: "tui.WithGap(1)",
		},
		"gap-4": {
			input:      "gap-4",
			wantOK:     true,
			wantOption: "tui.WithGap(4)",
		},
		"p-2": {
			input:      "p-2",
			wantOK:     true,
			wantOption: "tui.WithPadding(2)",
		},
		"px-3": {
			input:      "px-3",
			wantOK:     true,
			wantOption: "tui.WithPaddingTRBL(0, 3, 0, 3)",
		},
		"py-5": {
			input:      "py-5",
			wantOK:     true,
			wantOption: "tui.WithPaddingTRBL(5, 0, 5, 0)",
		},
		"m-1": {
			input:      "m-1",
			wantOK:     true,
			wantOption: "tui.WithMargin(1)",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
		})
	}
}

func TestParseTailwindClass_DynamicSizing(t *testing.T) {
	type tc struct {
		input       string
		wantOK      bool
		wantOption  string
	}

	tests := map[string]tc{
		"w-10": {
			input:      "w-10",
			wantOK:     true,
			wantOption: "tui.WithWidth(10)",
		},
		"h-20": {
			input:      "h-20",
			wantOK:     true,
			wantOption: "tui.WithHeight(20)",
		},
		"min-w-5": {
			input:      "min-w-5",
			wantOK:     true,
			wantOption: "tui.WithMinWidth(5)",
		},
		"max-w-100": {
			input:      "max-w-100",
			wantOK:     true,
			wantOption: "tui.WithMaxWidth(100)",
		},
		"min-h-3": {
			input:      "min-h-3",
			wantOK:     true,
			wantOption: "tui.WithMinHeight(3)",
		},
		"max-h-50": {
			input:      "max-h-50",
			wantOK:     true,
			wantOption: "tui.WithMaxHeight(50)",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
		})
	}
}

func TestParseTailwindClass_Borders(t *testing.T) {
	type tc struct {
		input       string
		wantOK      bool
		wantOption  string
		wantImport  string
	}

	tests := map[string]tc{
		"border": {
			input:      "border",
			wantOK:     true,
			wantOption: "tui.WithBorder(tui.BorderSingle)",
			wantImport: "tui",
		},
		"border-rounded": {
			input:      "border-rounded",
			wantOK:     true,
			wantOption: "tui.WithBorder(tui.BorderRounded)",
			wantImport: "tui",
		},
		"border-double": {
			input:      "border-double",
			wantOK:     true,
			wantOption: "tui.WithBorder(tui.BorderDouble)",
			wantImport: "tui",
		},
		"border-thick": {
			input:      "border-thick",
			wantOK:     true,
			wantOption: "tui.WithBorder(tui.BorderThick)",
			wantImport: "tui",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
			if mapping.NeedsImport != tt.wantImport {
				t.Errorf("NeedsImport = %q, want %q", mapping.NeedsImport, tt.wantImport)
			}
		})
	}
}

func TestParseTailwindClass_TextStyles(t *testing.T) {
	type tc struct {
		input          string
		wantOK         bool
		wantIsTextStyle bool
		wantTextMethod string
	}

	tests := map[string]tc{
		"font-bold": {
			input:          "font-bold",
			wantOK:         true,
			wantIsTextStyle: true,
			wantTextMethod: "Bold()",
		},
		"font-dim": {
			input:          "font-dim",
			wantOK:         true,
			wantIsTextStyle: true,
			wantTextMethod: "Dim()",
		},
		"italic": {
			input:          "italic",
			wantOK:         true,
			wantIsTextStyle: true,
			wantTextMethod: "Italic()",
		},
		"underline": {
			input:          "underline",
			wantOK:         true,
			wantIsTextStyle: true,
			wantTextMethod: "Underline()",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if mapping.IsTextStyle != tt.wantIsTextStyle {
				t.Errorf("IsTextStyle = %v, want %v", mapping.IsTextStyle, tt.wantIsTextStyle)
			}
			if mapping.TextMethod != tt.wantTextMethod {
				t.Errorf("TextMethod = %q, want %q", mapping.TextMethod, tt.wantTextMethod)
			}
		})
	}
}

func TestParseTailwindClass_Colors(t *testing.T) {
	type tc struct {
		input          string
		wantOK         bool
		wantTextMethod string
		wantOption     string
		wantImport     string
	}

	tests := map[string]tc{
		"text-red": {
			input:          "text-red",
			wantOK:         true,
			wantTextMethod: "Foreground(tui.Red)",
			wantImport:     "tui",
		},
		"text-cyan": {
			input:          "text-cyan",
			wantOK:         true,
			wantTextMethod: "Foreground(tui.Cyan)",
			wantImport:     "tui",
		},
		"bg-blue": {
			input:      "bg-blue",
			wantOK:     true,
			wantOption: "tui.WithBackground(tui.NewStyle().Background(tui.Blue))",
			wantImport: "tui",
		},
		"bg-yellow": {
			input:      "bg-yellow",
			wantOK:     true,
			wantOption: "tui.WithBackground(tui.NewStyle().Background(tui.Yellow))",
			wantImport: "tui",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if mapping.TextMethod != tt.wantTextMethod {
				t.Errorf("TextMethod = %q, want %q", mapping.TextMethod, tt.wantTextMethod)
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
			if mapping.NeedsImport != tt.wantImport {
				t.Errorf("NeedsImport = %q, want %q", mapping.NeedsImport, tt.wantImport)
			}
		})
	}
}

func TestParseTailwindClass_Scroll(t *testing.T) {
	type tc struct {
		input      string
		wantOK     bool
		wantOption string
	}

	tests := map[string]tc{
		"overflow-scroll": {
			input:      "overflow-scroll",
			wantOK:     true,
			wantOption: "tui.WithScrollable(tui.ScrollBoth)",
		},
		"overflow-y-scroll": {
			input:      "overflow-y-scroll",
			wantOK:     true,
			wantOption: "tui.WithScrollable(tui.ScrollVertical)",
		},
		"overflow-x-scroll": {
			input:      "overflow-x-scroll",
			wantOK:     true,
			wantOption: "tui.WithScrollable(tui.ScrollHorizontal)",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
		})
	}
}

func TestParseTailwindClass_Unknown(t *testing.T) {
	type tc struct {
		input  string
		wantOK bool
	}

	tests := map[string]tc{
		"unknown-class": {
			input:  "unknown-class",
			wantOK: false,
		},
		"random": {
			input:  "random",
			wantOK: false,
		},
		"empty": {
			input:  "",
			wantOK: false,
		},
		"whitespace": {
			input:  "  ",
			wantOK: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
		})
	}
}

func TestParseTailwindClass_Gradients(t *testing.T) {
	type tc struct {
		input       string
		wantOK      bool
		wantOption  string
		wantImport  string
	}

	tests := map[string]tc{
		"text-gradient horizontal": {
			input:      "text-gradient-red-blue",
			wantOK:     true,
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.Red, tui.Blue).WithDirection(tui.GradientHorizontal))",
			wantImport: "tui",
		},
		"text-gradient vertical": {
			input:      "text-gradient-red-blue-v",
			wantOK:     true,
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.Red, tui.Blue).WithDirection(tui.GradientVertical))",
			wantImport: "tui",
		},
		"text-gradient diagonal down": {
			input:      "text-gradient-cyan-magenta-dd",
			wantOK:     true,
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.Cyan, tui.Magenta).WithDirection(tui.GradientDiagonalDown))",
			wantImport: "tui",
		},
		"text-gradient diagonal up": {
			input:      "text-gradient-yellow-red-du",
			wantOK:     true,
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.Yellow, tui.Red).WithDirection(tui.GradientDiagonalUp))",
			wantImport: "tui",
		},
		"bg-gradient horizontal": {
			input:      "bg-gradient-green-blue",
			wantOK:     true,
			wantOption: "tui.WithBackgroundGradient(tui.NewGradient(tui.Green, tui.Blue).WithDirection(tui.GradientHorizontal))",
			wantImport: "tui",
		},
		"bg-gradient vertical": {
			input:      "bg-gradient-red-blue-v",
			wantOK:     true,
			wantOption: "tui.WithBackgroundGradient(tui.NewGradient(tui.Red, tui.Blue).WithDirection(tui.GradientVertical))",
			wantImport: "tui",
		},
		"border-gradient horizontal": {
			input:      "border-gradient-yellow-red",
			wantOK:     true,
			wantOption: "tui.WithBorderGradient(tui.NewGradient(tui.Yellow, tui.Red).WithDirection(tui.GradientHorizontal))",
			wantImport: "tui",
		},
		"border-gradient diagonal": {
			input:      "border-gradient-white-black-dd",
			wantOK:     true,
			wantOption: "tui.WithBorderGradient(tui.NewGradient(tui.White, tui.Black).WithDirection(tui.GradientDiagonalDown))",
			wantImport: "tui",
		},
		"bright colors": {
			input:      "text-gradient-bright-red-bright-blue",
			wantOK:     true,
			wantOption: "tui.WithTextGradient(tui.NewGradient(tui.BrightRed, tui.BrightBlue).WithDirection(tui.GradientHorizontal))",
			wantImport: "tui",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mapping, ok := ParseTailwindClass(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ParseTailwindClass(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if !ok {
				return
			}
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
			if mapping.NeedsImport != tt.wantImport {
				t.Errorf("NeedsImport = %q, want %q", mapping.NeedsImport, tt.wantImport)
			}
		})
	}
}
