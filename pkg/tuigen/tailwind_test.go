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
			wantOption: "element.WithPaddingX(3)",
		},
		"py-5": {
			input:      "py-5",
			wantOK:     true,
			wantOption: "element.WithPaddingY(5)",
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
