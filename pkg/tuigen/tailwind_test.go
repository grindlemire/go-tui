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
