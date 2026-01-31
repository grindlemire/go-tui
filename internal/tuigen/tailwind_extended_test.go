package tuigen

import (
	"testing"
)

func TestParseTailwindClass_WidthFractions(t *testing.T) {
	type tc struct {
		input      string
		wantOK     bool
		wantOption string
	}

	tests := map[string]tc{
		"w-1/2": {
			input:      "w-1/2",
			wantOK:     true,
			wantOption: "tui.WithWidthPercent(50.00)",
		},
		"w-1/3": {
			input:      "w-1/3",
			wantOK:     true,
			wantOption: "tui.WithWidthPercent(33.33)",
		},
		"w-2/3": {
			input:      "w-2/3",
			wantOK:     true,
			wantOption: "tui.WithWidthPercent(66.67)",
		},
		"w-1/4": {
			input:      "w-1/4",
			wantOK:     true,
			wantOption: "tui.WithWidthPercent(25.00)",
		},
		"w-3/4": {
			input:      "w-3/4",
			wantOK:     true,
			wantOption: "tui.WithWidthPercent(75.00)",
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

func TestParseTailwindClass_HeightFractions(t *testing.T) {
	type tc struct {
		input      string
		wantOK     bool
		wantOption string
	}

	tests := map[string]tc{
		"h-1/2": {
			input:      "h-1/2",
			wantOK:     true,
			wantOption: "tui.WithHeightPercent(50.00)",
		},
		"h-1/4": {
			input:      "h-1/4",
			wantOK:     true,
			wantOption: "tui.WithHeightPercent(25.00)",
		},
		"h-3/4": {
			input:      "h-3/4",
			wantOK:     true,
			wantOption: "tui.WithHeightPercent(75.00)",
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

func TestParseTailwindClass_WidthHeightKeywords(t *testing.T) {
	type tc struct {
		input      string
		wantOK     bool
		wantOption string
	}

	tests := map[string]tc{
		"w-full": {
			input:      "w-full",
			wantOK:     true,
			wantOption: "tui.WithWidthPercent(100.00)",
		},
		"w-auto": {
			input:      "w-auto",
			wantOK:     true,
			wantOption: "tui.WithWidthAuto()",
		},
		"h-full": {
			input:      "h-full",
			wantOK:     true,
			wantOption: "tui.WithHeightPercent(100.00)",
		},
		"h-auto": {
			input:      "h-auto",
			wantOK:     true,
			wantOption: "tui.WithHeightAuto()",
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

func TestParseTailwindClass_IndividualPadding(t *testing.T) {
	type tc struct {
		input  string
		wantOK bool
	}

	tests := map[string]tc{
		"pt-2": {input: "pt-2", wantOK: true},
		"pr-3": {input: "pr-3", wantOK: true},
		"pb-4": {input: "pb-4", wantOK: true},
		"pl-1": {input: "pl-1", wantOK: true},
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

func TestParseTailwindClass_IndividualMargin(t *testing.T) {
	type tc struct {
		input  string
		wantOK bool
	}

	tests := map[string]tc{
		"mt-2": {input: "mt-2", wantOK: true},
		"mr-3": {input: "mr-3", wantOK: true},
		"mb-4": {input: "mb-4", wantOK: true},
		"ml-1": {input: "ml-1", wantOK: true},
		"mx-2": {input: "mx-2", wantOK: true},
		"my-3": {input: "my-3", wantOK: true},
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

func TestParseTailwindClass_FlexUtilities(t *testing.T) {
	type tc struct {
		input      string
		wantOK     bool
		wantOption string
	}

	tests := map[string]tc{
		"self-start": {
			input:      "self-start",
			wantOK:     true,
			wantOption: "tui.WithAlignSelf(tui.AlignStart)",
		},
		"self-end": {
			input:      "self-end",
			wantOK:     true,
			wantOption: "tui.WithAlignSelf(tui.AlignEnd)",
		},
		"self-center": {
			input:      "self-center",
			wantOK:     true,
			wantOption: "tui.WithAlignSelf(tui.AlignCenter)",
		},
		"self-stretch": {
			input:      "self-stretch",
			wantOK:     true,
			wantOption: "tui.WithAlignSelf(tui.AlignStretch)",
		},
		"justify-evenly": {
			input:      "justify-evenly",
			wantOK:     true,
			wantOption: "tui.WithJustify(tui.JustifySpaceEvenly)",
		},
		"justify-around": {
			input:      "justify-around",
			wantOK:     true,
			wantOption: "tui.WithJustify(tui.JustifySpaceAround)",
		},
		"items-stretch": {
			input:      "items-stretch",
			wantOK:     true,
			wantOption: "tui.WithAlign(tui.AlignStretch)",
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

func TestParseTailwindClass_FlexGrowShrink(t *testing.T) {
	type tc struct {
		input      string
		wantOK     bool
		wantOption string
	}

	tests := map[string]tc{
		"flex-grow-0": {
			input:      "flex-grow-0",
			wantOK:     true,
			wantOption: "tui.WithFlexGrow(0)",
		},
		"flex-grow-2": {
			input:      "flex-grow-2",
			wantOK:     true,
			wantOption: "tui.WithFlexGrow(2)",
		},
		"flex-shrink-0": {
			input:      "flex-shrink-0",
			wantOK:     true,
			wantOption: "tui.WithFlexShrink(0)",
		},
		"flex-shrink-1": {
			input:      "flex-shrink-1",
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
			if mapping.Option != tt.wantOption {
				t.Errorf("Option = %q, want %q", mapping.Option, tt.wantOption)
			}
		})
	}
}

func TestParseTailwindClass_BorderColors(t *testing.T) {
	type tc struct {
		input      string
		wantOK     bool
		wantOption string
		wantImport string
	}

	tests := map[string]tc{
		"border-red": {
			input:      "border-red",
			wantOK:     true,
			wantOption: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Red))",
			wantImport: "tui",
		},
		"border-cyan": {
			input:      "border-cyan",
			wantOK:     true,
			wantOption: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan))",
			wantImport: "tui",
		},
		"border-green": {
			input:      "border-green",
			wantOK:     true,
			wantOption: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Green))",
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

func TestParseTailwindClass_TextAlignment(t *testing.T) {
	type tc struct {
		input      string
		wantOK     bool
		wantOption string
	}

	tests := map[string]tc{
		"text-left": {
			input:      "text-left",
			wantOK:     true,
			wantOption: "tui.WithTextAlign(tui.TextAlignLeft)",
		},
		"text-center": {
			input:      "text-center",
			wantOK:     true,
			wantOption: "tui.WithTextAlign(tui.TextAlignCenter)",
		},
		"text-right": {
			input:      "text-right",
			wantOK:     true,
			wantOption: "tui.WithTextAlign(tui.TextAlignRight)",
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
