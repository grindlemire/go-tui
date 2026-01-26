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
			wantOption: "element.WithDirection(layout.Row)",
			wantImport: "layout",
		},
		"flex-row": {
			input:      "flex-row",
			wantOK:     true,
			wantOption: "element.WithDirection(layout.Row)",
			wantImport: "layout",
		},
		"flex-col": {
			input:      "flex-col",
			wantOK:     true,
			wantOption: "element.WithDirection(layout.Column)",
			wantImport: "layout",
		},
		"flex-grow": {
			input:      "flex-grow",
			wantOK:     true,
			wantOption: "element.WithFlexGrow(1)",
		},
		"flex-shrink": {
			input:      "flex-shrink",
			wantOK:     true,
			wantOption: "element.WithFlexShrink(1)",
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
			wantOption: "element.WithJustify(layout.JustifyStart)",
		},
		"justify-center": {
			input:      "justify-center",
			wantOK:     true,
			wantOption: "element.WithJustify(layout.JustifyCenter)",
		},
		"justify-end": {
			input:      "justify-end",
			wantOK:     true,
			wantOption: "element.WithJustify(layout.JustifyEnd)",
		},
		"justify-between": {
			input:      "justify-between",
			wantOK:     true,
			wantOption: "element.WithJustify(layout.JustifySpaceBetween)",
		},
		"items-start": {
			input:      "items-start",
			wantOK:     true,
			wantOption: "element.WithAlign(layout.AlignStart)",
		},
		"items-center": {
			input:      "items-center",
			wantOK:     true,
			wantOption: "element.WithAlign(layout.AlignCenter)",
		},
		"items-end": {
			input:      "items-end",
			wantOK:     true,
			wantOption: "element.WithAlign(layout.AlignEnd)",
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
			wantOption: "element.WithGap(1)",
		},
		"gap-4": {
			input:      "gap-4",
			wantOK:     true,
			wantOption: "element.WithGap(4)",
		},
		"p-2": {
			input:      "p-2",
			wantOK:     true,
			wantOption: "element.WithPadding(2)",
		},
		"px-3": {
			input:      "px-3",
			wantOK:     true,
			wantOption: "element.WithPaddingTRBL(0, 3, 0, 3)",
		},
		"py-5": {
			input:      "py-5",
			wantOK:     true,
			wantOption: "element.WithPaddingTRBL(5, 0, 5, 0)",
		},
		"m-1": {
			input:      "m-1",
			wantOK:     true,
			wantOption: "element.WithMargin(1)",
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
			wantOption: "element.WithWidth(10)",
		},
		"h-20": {
			input:      "h-20",
			wantOK:     true,
			wantOption: "element.WithHeight(20)",
		},
		"min-w-5": {
			input:      "min-w-5",
			wantOK:     true,
			wantOption: "element.WithMinWidth(5)",
		},
		"max-w-100": {
			input:      "max-w-100",
			wantOK:     true,
			wantOption: "element.WithMaxWidth(100)",
		},
		"min-h-3": {
			input:      "min-h-3",
			wantOK:     true,
			wantOption: "element.WithMinHeight(3)",
		},
		"max-h-50": {
			input:      "max-h-50",
			wantOK:     true,
			wantOption: "element.WithMaxHeight(50)",
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
			wantOption: "element.WithBorder(tui.BorderSingle)",
			wantImport: "tui",
		},
		"border-rounded": {
			input:      "border-rounded",
			wantOK:     true,
			wantOption: "element.WithBorder(tui.BorderRounded)",
			wantImport: "tui",
		},
		"border-double": {
			input:      "border-double",
			wantOK:     true,
			wantOption: "element.WithBorder(tui.BorderDouble)",
			wantImport: "tui",
		},
		"border-thick": {
			input:      "border-thick",
			wantOK:     true,
			wantOption: "element.WithBorder(tui.BorderThick)",
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
			input:          "bg-blue",
			wantOK:         true,
			wantTextMethod: "Background(tui.Blue)",
			wantImport:     "tui",
		},
		"bg-yellow": {
			input:          "bg-yellow",
			wantOK:         true,
			wantTextMethod: "Background(tui.Yellow)",
			wantImport:     "tui",
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
			wantOption: "element.WithScrollable(element.ScrollBoth)",
		},
		"overflow-y-scroll": {
			input:      "overflow-y-scroll",
			wantOK:     true,
			wantOption: "element.WithScrollable(element.ScrollVertical)",
		},
		"overflow-x-scroll": {
			input:      "overflow-x-scroll",
			wantOK:     true,
			wantOption: "element.WithScrollable(element.ScrollHorizontal)",
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

func TestParseTailwindClasses_Multiple(t *testing.T) {
	type tc struct {
		input           string
		wantOptions     []string
		wantTextMethods []string
		wantImports     []string
	}

	tests := map[string]tc{
		"layout classes": {
			input:       "flex flex-col gap-2 p-4",
			wantOptions: []string{
				"element.WithDirection(layout.Row)",
				"element.WithDirection(layout.Column)",
				"element.WithGap(2)",
				"element.WithPadding(4)",
			},
			wantImports: []string{"layout"},
		},
		"text styles": {
			input:           "font-bold text-cyan",
			wantTextMethods: []string{"Bold()", "Foreground(tui.Cyan)"},
			wantImports:     []string{"tui"},
		},
		"mixed classes": {
			input:       "flex-col border-rounded font-bold text-red",
			wantOptions: []string{
				"element.WithDirection(layout.Column)",
				"element.WithBorder(tui.BorderRounded)",
			},
			wantTextMethods: []string{"Bold()", "Foreground(tui.Red)"},
			wantImports:     []string{"layout", "tui"},
		},
		"with unknown classes": {
			input:       "flex unknown-class gap-1",
			wantOptions: []string{
				"element.WithDirection(layout.Row)",
				"element.WithGap(1)",
			},
			wantImports: []string{"layout"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ParseTailwindClasses(tt.input)

			// Check options
			if len(result.Options) != len(tt.wantOptions) {
				t.Errorf("Options count = %d, want %d", len(result.Options), len(tt.wantOptions))
			} else {
				for i, opt := range tt.wantOptions {
					if result.Options[i] != opt {
						t.Errorf("Options[%d] = %q, want %q", i, result.Options[i], opt)
					}
				}
			}

			// Check text methods
			if len(result.TextMethods) != len(tt.wantTextMethods) {
				t.Errorf("TextMethods count = %d, want %d", len(result.TextMethods), len(tt.wantTextMethods))
			} else {
				for i, method := range tt.wantTextMethods {
					if result.TextMethods[i] != method {
						t.Errorf("TextMethods[%d] = %q, want %q", i, result.TextMethods[i], method)
					}
				}
			}

			// Check imports
			for _, imp := range tt.wantImports {
				if !result.NeedsImports[imp] {
					t.Errorf("missing import %q", imp)
				}
			}
		})
	}
}

func TestBuildTextStyleOption(t *testing.T) {
	type tc struct {
		methods []string
		want    string
	}

	tests := map[string]tc{
		"empty": {
			methods: nil,
			want:    "",
		},
		"single method": {
			methods: []string{"Bold()"},
			want:    "element.WithTextStyle(tui.NewStyle().Bold())",
		},
		"multiple methods": {
			methods: []string{"Bold()", "Foreground(tui.Cyan)", "Italic()"},
			want:    "element.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.Cyan).Italic())",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := BuildTextStyleOption(tt.methods)
			if got != tt.want {
				t.Errorf("BuildTextStyleOption() = %q, want %q", got, tt.want)
			}
		})
	}
}

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
			wantOption: "element.WithWidthPercent(50.00)",
		},
		"w-1/3": {
			input:      "w-1/3",
			wantOK:     true,
			wantOption: "element.WithWidthPercent(33.33)",
		},
		"w-2/3": {
			input:      "w-2/3",
			wantOK:     true,
			wantOption: "element.WithWidthPercent(66.67)",
		},
		"w-1/4": {
			input:      "w-1/4",
			wantOK:     true,
			wantOption: "element.WithWidthPercent(25.00)",
		},
		"w-3/4": {
			input:      "w-3/4",
			wantOK:     true,
			wantOption: "element.WithWidthPercent(75.00)",
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
			wantOption: "element.WithHeightPercent(50.00)",
		},
		"h-1/4": {
			input:      "h-1/4",
			wantOK:     true,
			wantOption: "element.WithHeightPercent(25.00)",
		},
		"h-3/4": {
			input:      "h-3/4",
			wantOK:     true,
			wantOption: "element.WithHeightPercent(75.00)",
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
			wantOption: "element.WithWidthPercent(100.00)",
		},
		"w-auto": {
			input:      "w-auto",
			wantOK:     true,
			wantOption: "element.WithWidthAuto()",
		},
		"h-full": {
			input:      "h-full",
			wantOK:     true,
			wantOption: "element.WithHeightPercent(100.00)",
		},
		"h-auto": {
			input:      "h-auto",
			wantOK:     true,
			wantOption: "element.WithHeightAuto()",
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
			wantOption: "element.WithAlignSelf(layout.AlignStart)",
		},
		"self-end": {
			input:      "self-end",
			wantOK:     true,
			wantOption: "element.WithAlignSelf(layout.AlignEnd)",
		},
		"self-center": {
			input:      "self-center",
			wantOK:     true,
			wantOption: "element.WithAlignSelf(layout.AlignCenter)",
		},
		"self-stretch": {
			input:      "self-stretch",
			wantOK:     true,
			wantOption: "element.WithAlignSelf(layout.AlignStretch)",
		},
		"justify-evenly": {
			input:      "justify-evenly",
			wantOK:     true,
			wantOption: "element.WithJustify(layout.JustifySpaceEvenly)",
		},
		"justify-around": {
			input:      "justify-around",
			wantOK:     true,
			wantOption: "element.WithJustify(layout.JustifySpaceAround)",
		},
		"items-stretch": {
			input:      "items-stretch",
			wantOK:     true,
			wantOption: "element.WithAlign(layout.AlignStretch)",
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
			wantOption: "element.WithFlexGrow(0)",
		},
		"flex-grow-2": {
			input:      "flex-grow-2",
			wantOK:     true,
			wantOption: "element.WithFlexGrow(2)",
		},
		"flex-shrink-0": {
			input:      "flex-shrink-0",
			wantOK:     true,
			wantOption: "element.WithFlexShrink(0)",
		},
		"flex-shrink-1": {
			input:      "flex-shrink-1",
			wantOK:     true,
			wantOption: "element.WithFlexShrink(1)",
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
			wantOption: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Red))",
			wantImport: "tui",
		},
		"border-cyan": {
			input:      "border-cyan",
			wantOK:     true,
			wantOption: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan))",
			wantImport: "tui",
		},
		"border-green": {
			input:      "border-green",
			wantOK:     true,
			wantOption: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Green))",
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
			wantOption: "element.WithTextAlign(element.TextAlignLeft)",
		},
		"text-center": {
			input:      "text-center",
			wantOK:     true,
			wantOption: "element.WithTextAlign(element.TextAlignCenter)",
		},
		"text-right": {
			input:      "text-right",
			wantOK:     true,
			wantOption: "element.WithTextAlign(element.TextAlignRight)",
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

func TestParseTailwindClasses_PaddingAccumulation(t *testing.T) {
	type tc struct {
		input       string
		wantOptions []string
	}

	tests := map[string]tc{
		"single padding top": {
			input:       "pt-2",
			wantOptions: []string{"element.WithPaddingTRBL(2, 0, 0, 0)"},
		},
		"padding top and bottom": {
			input:       "pt-2 pb-4",
			wantOptions: []string{"element.WithPaddingTRBL(2, 0, 4, 0)"},
		},
		"all padding sides": {
			input:       "pt-1 pr-2 pb-3 pl-4",
			wantOptions: []string{"element.WithPaddingTRBL(1, 2, 3, 4)"},
		},
		"padding with other classes": {
			input:       "flex pt-2 pb-4 gap-1",
			wantOptions: []string{
				"element.WithDirection(layout.Row)",
				"element.WithGap(1)",
				"element.WithPaddingTRBL(2, 0, 4, 0)",
			},
		},
		"padding horizontal": {
			input:       "px-3",
			wantOptions: []string{"element.WithPaddingTRBL(0, 3, 0, 3)"},
		},
		"padding vertical": {
			input:       "py-5",
			wantOptions: []string{"element.WithPaddingTRBL(5, 0, 5, 0)"},
		},
		"padding horizontal and top": {
			input:       "px-3 pt-2",
			wantOptions: []string{"element.WithPaddingTRBL(2, 3, 0, 3)"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ParseTailwindClasses(tt.input)

			if len(result.Options) != len(tt.wantOptions) {
				t.Errorf("Options count = %d, want %d. Got: %v", len(result.Options), len(tt.wantOptions), result.Options)
				return
			}

			for i, opt := range tt.wantOptions {
				if result.Options[i] != opt {
					t.Errorf("Options[%d] = %q, want %q", i, result.Options[i], opt)
				}
			}
		})
	}
}

func TestParseTailwindClasses_MarginAccumulation(t *testing.T) {
	type tc struct {
		input       string
		wantOptions []string
	}

	tests := map[string]tc{
		"single margin top": {
			input:       "mt-2",
			wantOptions: []string{"element.WithMarginTRBL(2, 0, 0, 0)"},
		},
		"margin horizontal": {
			input:       "mx-3",
			wantOptions: []string{"element.WithMarginTRBL(0, 3, 0, 3)"},
		},
		"margin vertical": {
			input:       "my-2",
			wantOptions: []string{"element.WithMarginTRBL(2, 0, 2, 0)"},
		},
		"margin all sides": {
			input:       "mt-1 mr-2 mb-3 ml-4",
			wantOptions: []string{"element.WithMarginTRBL(1, 2, 3, 4)"},
		},
		"margin with other classes": {
			input:       "flex mt-2 mb-4 gap-1",
			wantOptions: []string{
				"element.WithDirection(layout.Row)",
				"element.WithGap(1)",
				"element.WithMarginTRBL(2, 0, 4, 0)",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ParseTailwindClasses(tt.input)

			if len(result.Options) != len(tt.wantOptions) {
				t.Errorf("Options count = %d, want %d. Got: %v", len(result.Options), len(tt.wantOptions), result.Options)
				return
			}

			for i, opt := range tt.wantOptions {
				if result.Options[i] != opt {
					t.Errorf("Options[%d] = %q, want %q", i, result.Options[i], opt)
				}
			}
		})
	}
}

func TestParseTailwindClasses_PaddingAndMarginCombined(t *testing.T) {
	result := ParseTailwindClasses("pt-1 pb-2 mt-3 mb-4")

	expected := []string{
		"element.WithPaddingTRBL(1, 0, 2, 0)",
		"element.WithMarginTRBL(3, 0, 4, 0)",
	}

	if len(result.Options) != len(expected) {
		t.Errorf("Options count = %d, want %d. Got: %v", len(result.Options), len(expected), result.Options)
		return
	}

	for i, opt := range expected {
		if result.Options[i] != opt {
			t.Errorf("Options[%d] = %q, want %q", i, result.Options[i], opt)
		}
	}
}

func TestPaddingAccumulator(t *testing.T) {
	t.Run("merge and toOption", func(t *testing.T) {
		var acc PaddingAccumulator
		acc.Merge("top", 1)
		acc.Merge("right", 2)
		acc.Merge("bottom", 3)
		acc.Merge("left", 4)

		want := "element.WithPaddingTRBL(1, 2, 3, 4)"
		got := acc.ToOption()
		if got != want {
			t.Errorf("ToOption() = %q, want %q", got, want)
		}
	})

	t.Run("partial sides", func(t *testing.T) {
		var acc PaddingAccumulator
		acc.Merge("top", 5)
		acc.Merge("bottom", 10)

		want := "element.WithPaddingTRBL(5, 0, 10, 0)"
		got := acc.ToOption()
		if got != want {
			t.Errorf("ToOption() = %q, want %q", got, want)
		}
	})

	t.Run("merge x sets left and right", func(t *testing.T) {
		var acc PaddingAccumulator
		acc.Merge("x", 5)

		want := "element.WithPaddingTRBL(0, 5, 0, 5)"
		got := acc.ToOption()
		if got != want {
			t.Errorf("ToOption() = %q, want %q", got, want)
		}
	})

	t.Run("merge y sets top and bottom", func(t *testing.T) {
		var acc PaddingAccumulator
		acc.Merge("y", 3)

		want := "element.WithPaddingTRBL(3, 0, 3, 0)"
		got := acc.ToOption()
		if got != want {
			t.Errorf("ToOption() = %q, want %q", got, want)
		}
	})

	t.Run("empty returns empty string", func(t *testing.T) {
		var acc PaddingAccumulator
		if got := acc.ToOption(); got != "" {
			t.Errorf("empty ToOption() = %q, want empty", got)
		}
	})
}

func TestMarginAccumulator(t *testing.T) {
	t.Run("merge individual sides", func(t *testing.T) {
		var acc MarginAccumulator
		acc.Merge("top", 1)
		acc.Merge("right", 2)
		acc.Merge("bottom", 3)
		acc.Merge("left", 4)

		want := "element.WithMarginTRBL(1, 2, 3, 4)"
		got := acc.ToOption()
		if got != want {
			t.Errorf("ToOption() = %q, want %q", got, want)
		}
	})

	t.Run("merge x sets left and right", func(t *testing.T) {
		var acc MarginAccumulator
		acc.Merge("x", 5)

		want := "element.WithMarginTRBL(0, 5, 0, 5)"
		got := acc.ToOption()
		if got != want {
			t.Errorf("ToOption() = %q, want %q", got, want)
		}
	})

	t.Run("merge y sets top and bottom", func(t *testing.T) {
		var acc MarginAccumulator
		acc.Merge("y", 3)

		want := "element.WithMarginTRBL(3, 0, 3, 0)"
		got := acc.ToOption()
		if got != want {
			t.Errorf("ToOption() = %q, want %q", got, want)
		}
	})

	t.Run("empty returns empty string", func(t *testing.T) {
		var acc MarginAccumulator
		if got := acc.ToOption(); got != "" {
			t.Errorf("empty ToOption() = %q, want empty", got)
		}
	})
}

func TestLevenshteinDistance(t *testing.T) {
	type tc struct {
		a        string
		b        string
		expected int
	}

	tests := map[string]tc{
		"identical strings": {
			a:        "flex",
			b:        "flex",
			expected: 0,
		},
		"one character different": {
			a:        "flex",
			b:        "fles",
			expected: 1,
		},
		"one character added": {
			a:        "flex",
			b:        "flexs",
			expected: 1,
		},
		"one character removed": {
			a:        "flex",
			b:        "fle",
			expected: 1,
		},
		"completely different": {
			a:        "flex",
			b:        "border",
			expected: 5, // flex→blex→bolex→borex→borde→border or equivalent 5-edit path
		},
		"empty first string": {
			a:        "",
			b:        "flex",
			expected: 4,
		},
		"empty second string": {
			a:        "flex",
			b:        "",
			expected: 4,
		},
		"both empty": {
			a:        "",
			b:        "",
			expected: 0,
		},
		"flex-col vs flex-column": {
			a:        "flex-col",
			b:        "flex-column",
			expected: 3,
		},
		"flex-columns vs flex-col": {
			a:        "flex-columns",
			b:        "flex-col",
			expected: 4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := levenshteinDistance(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestFindSimilarClass(t *testing.T) {
	type tc struct {
		input    string
		expected string
	}

	tests := map[string]tc{
		"exact match in similarClasses map": {
			input:    "flex-columns",
			expected: "flex-col",
		},
		"flex-column typo": {
			input:    "flex-column",
			expected: "flex-col",
		},
		"bold to font-bold": {
			input:    "bold",
			expected: "font-bold",
		},
		"center to text-center": {
			input:    "center",
			expected: "text-center",
		},
		"grow to flex-grow": {
			input:    "grow",
			expected: "flex-grow",
		},
		"shrink to flex-shrink": {
			input:    "shrink",
			expected: "flex-shrink",
		},
		"no-grow to flex-grow-0": {
			input:    "no-grow",
			expected: "flex-grow-0",
		},
		"padding-top to pt-1": {
			input:    "padding-top",
			expected: "pt-1",
		},
		"fuzzy match - fex to flex": {
			input:    "fex",
			expected: "flex",
		},
		"fuzzy match - border-rounde to border-rounded": {
			input:    "border-rounde",
			expected: "border-rounded",
		},
		"no match for very different string": {
			input:    "xyzabc123",
			expected: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := findSimilarClass(tt.input)
			if got != tt.expected {
				t.Errorf("findSimilarClass(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateTailwindClass_Valid(t *testing.T) {
	type tc struct {
		input string
	}

	tests := map[string]tc{
		"flex":           {input: "flex"},
		"flex-col":       {input: "flex-col"},
		"flex-row":       {input: "flex-row"},
		"gap-2":          {input: "gap-2"},
		"p-4":            {input: "p-4"},
		"pt-2":           {input: "pt-2"},
		"m-1":            {input: "m-1"},
		"mt-3":           {input: "mt-3"},
		"w-full":         {input: "w-full"},
		"w-1/2":          {input: "w-1/2"},
		"h-auto":         {input: "h-auto"},
		"border":         {input: "border"},
		"border-rounded": {input: "border-rounded"},
		"font-bold":      {input: "font-bold"},
		"text-cyan":      {input: "text-cyan"},
		"bg-red":         {input: "bg-red"},
		"justify-center": {input: "justify-center"},
		"items-center":   {input: "items-center"},
		"self-start":     {input: "self-start"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ValidateTailwindClass(tt.input)
			if !result.Valid {
				t.Errorf("ValidateTailwindClass(%q) Valid = false, want true", tt.input)
			}
			if result.Class != tt.input {
				t.Errorf("ValidateTailwindClass(%q) Class = %q, want %q", tt.input, result.Class, tt.input)
			}
		})
	}
}

func TestValidateTailwindClass_Invalid(t *testing.T) {
	type tc struct {
		input          string
		wantSuggestion string
	}

	tests := map[string]tc{
		"flex-columns typo": {
			input:          "flex-columns",
			wantSuggestion: "flex-col",
		},
		"bold typo": {
			input:          "bold",
			wantSuggestion: "font-bold",
		},
		"center typo": {
			input:          "center",
			wantSuggestion: "text-center",
		},
		"completely unknown": {
			input:          "xyzabc123",
			wantSuggestion: "",
		},
		"empty string": {
			input:          "",
			wantSuggestion: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ValidateTailwindClass(tt.input)
			if result.Valid {
				t.Errorf("ValidateTailwindClass(%q) Valid = true, want false", tt.input)
			}
			if result.Suggestion != tt.wantSuggestion {
				t.Errorf("ValidateTailwindClass(%q) Suggestion = %q, want %q", tt.input, result.Suggestion, tt.wantSuggestion)
			}
		})
	}
}

func TestParseTailwindClassesWithPositions(t *testing.T) {
	type tc struct {
		input        string
		attrStartCol int
		wantCount    int
		checkClass   int    // index of class to check
		wantClass    string
		wantStartCol int
		wantEndCol   int
		wantValid    bool
	}

	tests := map[string]tc{
		"single valid class": {
			input:        "flex",
			attrStartCol: 10,
			wantCount:    1,
			checkClass:   0,
			wantClass:    "flex",
			wantStartCol: 10,
			wantEndCol:   14,
			wantValid:    true,
		},
		"multiple classes": {
			input:        "flex gap-2",
			attrStartCol: 5,
			wantCount:    2,
			checkClass:   1,
			wantClass:    "gap-2",
			wantStartCol: 10,
			wantEndCol:   15,
			wantValid:    true,
		},
		"invalid class": {
			input:        "flex-columns",
			attrStartCol: 0,
			wantCount:    1,
			checkClass:   0,
			wantClass:    "flex-columns",
			wantStartCol: 0,
			wantEndCol:   12,
			wantValid:    false,
		},
		"mixed valid and invalid": {
			input:        "flex flex-columns gap-2",
			attrStartCol: 0,
			wantCount:    3,
			checkClass:   1,
			wantClass:    "flex-columns",
			wantStartCol: 5,
			wantEndCol:   17,
			wantValid:    false,
		},
		"with extra whitespace": {
			input:        "  flex   gap-2  ",
			attrStartCol: 0,
			wantCount:    2,
			checkClass:   0,
			wantClass:    "flex",
			wantStartCol: 2,
			wantEndCol:   6,
			wantValid:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ParseTailwindClassesWithPositions(tt.input, tt.attrStartCol)

			if len(result) != tt.wantCount {
				t.Errorf("ParseTailwindClassesWithPositions count = %d, want %d", len(result), tt.wantCount)
				return
			}

			if tt.checkClass >= len(result) {
				t.Fatalf("checkClass index %d out of range", tt.checkClass)
			}

			class := result[tt.checkClass]
			if class.Class != tt.wantClass {
				t.Errorf("Class = %q, want %q", class.Class, tt.wantClass)
			}
			if class.StartCol != tt.wantStartCol {
				t.Errorf("StartCol = %d, want %d", class.StartCol, tt.wantStartCol)
			}
			if class.EndCol != tt.wantEndCol {
				t.Errorf("EndCol = %d, want %d", class.EndCol, tt.wantEndCol)
			}
			if class.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", class.Valid, tt.wantValid)
			}
		})
	}
}

func TestParseTailwindClassesWithPositions_Suggestions(t *testing.T) {
	result := ParseTailwindClassesWithPositions("flex-columns", 0)

	if len(result) != 1 {
		t.Fatalf("expected 1 class, got %d", len(result))
	}

	if result[0].Suggestion != "flex-col" {
		t.Errorf("Suggestion = %q, want %q", result[0].Suggestion, "flex-col")
	}
}

func TestAllTailwindClasses(t *testing.T) {
	classes := AllTailwindClasses()

	// Should return a non-empty list
	if len(classes) == 0 {
		t.Error("AllTailwindClasses() returned empty list")
	}

	// Check that we have classes from different categories
	categories := make(map[string]int)
	for _, c := range classes {
		categories[c.Category]++
	}

	expectedCategories := []string{"layout", "flex", "spacing", "typography", "visual"}
	for _, cat := range expectedCategories {
		if categories[cat] == 0 {
			t.Errorf("expected category %q to have classes, got 0", cat)
		}
	}

	// Check that all classes have required fields
	for _, c := range classes {
		if c.Name == "" {
			t.Error("found class with empty Name")
		}
		if c.Category == "" {
			t.Errorf("class %q has empty Category", c.Name)
		}
		if c.Description == "" {
			t.Errorf("class %q has empty Description", c.Name)
		}
		if c.Example == "" {
			t.Errorf("class %q has empty Example", c.Name)
		}
	}
}

func TestAllTailwindClasses_SpecificClasses(t *testing.T) {
	classes := AllTailwindClasses()

	// Build a map for easy lookup
	classMap := make(map[string]TailwindClassInfo)
	for _, c := range classes {
		classMap[c.Name] = c
	}

	// Check for specific classes that should exist
	expectedClasses := []string{
		"flex", "flex-col", "flex-row",
		"flex-grow", "flex-shrink", "flex-grow-0", "flex-shrink-0",
		"justify-start", "justify-center", "justify-end", "justify-between", "justify-around", "justify-evenly",
		"items-start", "items-center", "items-end", "items-stretch",
		"self-start", "self-center", "self-end", "self-stretch",
		"gap-1", "gap-2",
		"p-1", "p-2", "px-1", "py-1", "pt-1", "pr-1", "pb-1", "pl-1",
		"m-1", "m-2", "mx-1", "my-1", "mt-1", "mr-1", "mb-1", "ml-1",
		"w-full", "w-auto", "w-1/2",
		"h-full", "h-auto", "h-1/2",
		"border", "border-rounded", "border-double", "border-thick",
		"border-red", "border-green", "border-blue", "border-cyan",
		"font-bold", "font-dim", "italic", "underline",
		"text-left", "text-center", "text-right",
		"text-red", "text-green", "text-cyan",
		"bg-red", "bg-green", "bg-blue",
		"overflow-scroll", "overflow-y-scroll", "overflow-x-scroll",
	}

	for _, name := range expectedClasses {
		if _, ok := classMap[name]; !ok {
			t.Errorf("expected class %q not found in AllTailwindClasses()", name)
		}
	}
}
