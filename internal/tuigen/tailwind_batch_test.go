package tuigen

import (
	"testing"
)

func TestParseTailwindClasses_Multiple(t *testing.T) {
	type tc struct {
		input           string
		wantOptions     []string
		wantTextMethods []string
		wantImports     []string
	}

	tests := map[string]tc{
		"layout classes": {
			input: "flex flex-col gap-2 p-4",
			wantOptions: []string{
				"tui.WithDisplay(tui.DisplayFlex), tui.WithDirection(tui.Row)",
				"tui.WithDisplay(tui.DisplayFlex), tui.WithDirection(tui.Column)",
				"tui.WithGap(2)",
				"tui.WithPadding(4)",
			},
			wantImports: []string{"tui"},
		},
		"text styles": {
			input:           "font-bold text-cyan",
			wantTextMethods: []string{"Bold()", "Foreground(tui.Cyan)"},
			wantImports:     []string{"tui"},
		},
		"mixed classes": {
			input: "flex-col border-rounded font-bold text-red",
			wantOptions: []string{
				"tui.WithDisplay(tui.DisplayFlex), tui.WithDirection(tui.Column)",
				"tui.WithBorder(tui.BorderRounded)",
			},
			wantTextMethods: []string{"Bold()", "Foreground(tui.Red)"},
			wantImports:     []string{"tui"},
		},
		"with unknown classes": {
			input: "flex unknown-class gap-1",
			wantOptions: []string{
				"tui.WithDisplay(tui.DisplayFlex), tui.WithDirection(tui.Row)",
				"tui.WithGap(1)",
			},
			wantImports: []string{"tui"},
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
			want:    "tui.WithTextStyle(tui.NewStyle().Bold())",
		},
		"multiple methods": {
			methods: []string{"Bold()", "Foreground(tui.Cyan)", "Italic()"},
			want:    "tui.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.Cyan).Italic())",
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

func TestParseTailwindClasses_PaddingAccumulation(t *testing.T) {
	type tc struct {
		input       string
		wantOptions []string
	}

	tests := map[string]tc{
		"single padding top": {
			input:       "pt-2",
			wantOptions: []string{"tui.WithPaddingTRBL(2, 0, 0, 0)"},
		},
		"padding top and bottom": {
			input:       "pt-2 pb-4",
			wantOptions: []string{"tui.WithPaddingTRBL(2, 0, 4, 0)"},
		},
		"all padding sides": {
			input:       "pt-1 pr-2 pb-3 pl-4",
			wantOptions: []string{"tui.WithPaddingTRBL(1, 2, 3, 4)"},
		},
		"padding with other classes": {
			input: "flex pt-2 pb-4 gap-1",
			wantOptions: []string{
				"tui.WithDisplay(tui.DisplayFlex), tui.WithDirection(tui.Row)",
				"tui.WithGap(1)",
				"tui.WithPaddingTRBL(2, 0, 4, 0)",
			},
		},
		"padding horizontal": {
			input:       "px-3",
			wantOptions: []string{"tui.WithPaddingTRBL(0, 3, 0, 3)"},
		},
		"padding vertical": {
			input:       "py-5",
			wantOptions: []string{"tui.WithPaddingTRBL(5, 0, 5, 0)"},
		},
		"padding horizontal and top": {
			input:       "px-3 pt-2",
			wantOptions: []string{"tui.WithPaddingTRBL(2, 3, 0, 3)"},
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
			wantOptions: []string{"tui.WithMarginTRBL(2, 0, 0, 0)"},
		},
		"margin horizontal": {
			input:       "mx-3",
			wantOptions: []string{"tui.WithMarginTRBL(0, 3, 0, 3)"},
		},
		"margin vertical": {
			input:       "my-2",
			wantOptions: []string{"tui.WithMarginTRBL(2, 0, 2, 0)"},
		},
		"margin all sides": {
			input:       "mt-1 mr-2 mb-3 ml-4",
			wantOptions: []string{"tui.WithMarginTRBL(1, 2, 3, 4)"},
		},
		"margin with other classes": {
			input: "flex mt-2 mb-4 gap-1",
			wantOptions: []string{
				"tui.WithDisplay(tui.DisplayFlex), tui.WithDirection(tui.Row)",
				"tui.WithGap(1)",
				"tui.WithMarginTRBL(2, 0, 4, 0)",
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
		"tui.WithPaddingTRBL(1, 0, 2, 0)",
		"tui.WithMarginTRBL(3, 0, 4, 0)",
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

func TestParseTailwindClasses_BorderAccumulation(t *testing.T) {
	type tc struct {
		input       string
		wantOptions []string
	}

	tests := map[string]tc{
		"no border (substractive)": {
			input:       "border border-none",
			wantOptions: []string{"tui.WithBorder(tui.BorderNone)"},
		},
		"no border (substractive, symmetrical classes)": {
			input:       "border border-x-none border-y-none",
			wantOptions: []string{"tui.WithBorder(tui.BorderNone)"},
		},
		"border in all sides (individual classes)": {
			input:       "border-t-double border-r-thick border-b-rounded border-l-single",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderDouble, tui.BorderThick, tui.BorderRounded, tui.BorderSingle)"},
		},
		"border in all sides (symmetric classes)": {
			input:       "border-x-double border-y-thick",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderThick, tui.BorderDouble, tui.BorderThick, tui.BorderDouble)"},
		},
		"border in all sides (mixed classes)": {
			input:       "border-b-thick border-x-double border-t-thick",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderThick, tui.BorderDouble, tui.BorderThick, tui.BorderDouble)"},
		},
		"border in all sides but top (substractive)": {
			input:       "border-single border-t-none",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderNone, tui.BorderSingle, tui.BorderSingle, tui.BorderSingle)"},
		},
		"border in all sides but top (additive)": {
			input:       "border-r-single border-b-single border-l-single",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderNone, tui.BorderSingle, tui.BorderSingle, tui.BorderSingle)"},
		},
		"border in all sides but right (substractive)": {
			input:       "border-x-double border-y-thick border-r-none",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderThick, tui.BorderNone, tui.BorderThick, tui.BorderDouble)"},
		},
		"border in all sides but right (additive)": {
			input:       "border-y-rounded border-l-single",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderRounded, tui.BorderNone, tui.BorderRounded, tui.BorderSingle)"},
		},
		"border in all sides but bottom (substractive)": {
			input:       "border-rounded border-b-none",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderRounded, tui.BorderRounded, tui.BorderNone, tui.BorderRounded)"},
		},
		"border in all sides but bottom (additive)": {
			input:       "border-t-rounded border-x-rounded",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderRounded, tui.BorderRounded, tui.BorderNone, tui.BorderRounded)"},
		},
		"border in all sides but left (substractive)": {
			input:       "border-x-rounded border-l-none border-t-rounded border-b-rounded",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderRounded, tui.BorderRounded, tui.BorderRounded, tui.BorderNone)"},
		},
		"border in all sides but left (additive)": {
			input:       "border-y-thick border-r-double",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderThick, tui.BorderDouble, tui.BorderThick, tui.BorderNone)"},
		},
		"border in all sides but left and right (substractive, symmetric classes)": {
			input:       "border-double border-x-none",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderDouble, tui.BorderNone, tui.BorderDouble, tui.BorderNone)"},
		},
		"border in all sides but top and bottom (substractive, symmetric classes)": {
			input:       "border-double border-y-none",
			wantOptions: []string{"tui.WithBorderTRBL(tui.BorderNone, tui.BorderDouble, tui.BorderNone, tui.BorderDouble)"},
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
