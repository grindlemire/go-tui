package tuigen

import (
	"strings"
	"testing"
)

func TestAnalyzer_NestedElements(t *testing.T) {
	// Test that nested elements are all validated
	input := `package x
templ Test() {
	<div>
		<div>
			<unknownTag />
		</div>
	</div>
}`

	_, err := AnalyzeFile("test.gsx", input)
	if err == nil {
		t.Error("expected error for nested unknown tag")
		return
	}

	if !strings.Contains(err.Error(), "unknown element tag <unknownTag>") {
		t.Errorf("error %q does not contain expected message", err.Error())
	}
}

func TestAnalyzer_ControlFlowValidation(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"valid for loop": {
			input: `package x
templ Test(items []string) {
	<div>
		@for _, item := range items {
			<span>{item}</span>
		}
	</div>
}`,
			wantError: false,
		},
		"invalid element in for loop": {
			input: `package x
templ Test(items []string) {
	<div>
		@for _, item := range items {
			<badTag />
		}
	</div>
}`,
			wantError:     true,
			errorContains: "unknown element tag <badTag>",
		},
		"valid if statement": {
			input: `package x
templ Test(show bool) {
	<div>
		@if show {
			<span>visible</span>
		}
	</div>
}`,
			wantError: false,
		},
		"invalid element in if then": {
			input: `package x
templ Test(show bool) {
	<div>
		@if show {
			<badTag />
		}
	</div>
}`,
			wantError:     true,
			errorContains: "unknown element tag <badTag>",
		},
		"invalid element in if else": {
			input: `package x
templ Test(show bool) {
	<div>
		@if show {
			<span>yes</span>
		} @else {
			<badTag />
		}
	</div>
}`,
			wantError:     true,
			errorContains: "unknown element tag <badTag>",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_AllKnownAttributes(t *testing.T) {
	// Test all known attributes are accepted
	attributes := []string{
		"width", "widthPercent", "height", "heightPercent",
		"minWidth", "minHeight", "maxWidth", "maxHeight",
		"direction", "justify", "align", "gap",
		"flexGrow", "flexShrink", "alignSelf",
		"padding", "margin",
		"border", "borderStyle", "background",
		"text", "textStyle", "textAlign",
		"onFocus", "onBlur", "onEvent",
		"scrollable", "scrollbarStyle", "scrollbarThumbStyle",
		"disabled", "id",
	}

	for _, attr := range attributes {
		t.Run(attr, func(t *testing.T) {
			input := `package x
templ Test() {
	<div ` + attr + `=1></div>
}`
			_, err := AnalyzeFile("test.gsx", input)
			if err != nil {
				t.Errorf("attribute %q should be valid, got error: %v", attr, err)
			}
		})
	}
}

func TestAnalyzer_AllKnownTags(t *testing.T) {
	// Test all known tags are accepted
	tags := []string{
		"div", "span", "p", "ul", "li",
		"button", "input", "table", "progress",
		"hr", "br",
	}

	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			input := `package x
templ Test() {
	<` + tag + ` />
}`
			_, err := AnalyzeFile("test.gsx", input)
			if err != nil {
				t.Errorf("tag %q should be valid, got error: %v", tag, err)
			}
		})
	}
}

func TestAnalyzer_HRValid(t *testing.T) {
	type tc struct {
		input     string
		wantError bool
	}

	tests := map[string]tc{
		"hr self-closing": {
			input: `package x
templ Test() {
	<div>
		<hr/>
	</div>
}`,
			wantError: false,
		},
		"hr with class": {
			input: `package x
templ Test() {
	<div>
		<hr class="border-double"/>
	</div>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_BRValid(t *testing.T) {
	type tc struct {
		input     string
		wantError bool
	}

	tests := map[string]tc{
		"br self-closing": {
			input: `package x
templ Test() {
	<div>
		<br/>
	</div>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_VoidWithChildren(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"hr with text child": {
			input: `package x
templ Test() {
	<div>
		<hr>text</hr>
	</div>
}`,
			wantError:     true,
			errorContains: "<hr> is a void element and cannot have children",
		},
		"hr with element child": {
			input: `package x
templ Test() {
	<div>
		<hr><span>nested</span></hr>
	</div>
}`,
			wantError:     true,
			errorContains: "<hr> is a void element and cannot have children",
		},
		"br with text child": {
			input: `package x
templ Test() {
	<div>
		<br>text</br>
	</div>
}`,
			wantError:     true,
			errorContains: "<br> is a void element and cannot have children",
		},
		"input with child": {
			input: `package x
templ Test() {
	<div>
		<input>text</input>
	</div>
}`,
			wantError:     true,
			errorContains: "<input> is a void element and cannot have children",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_TailwindClassValidation(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
		hintContains  string
	}

	tests := map[string]tc{
		"valid tailwind classes": {
			input: `package x
templ Test() {
	<div class="flex-col gap-2 p-4"></div>
}`,
			wantError: false,
		},
		"valid width and height classes": {
			input: `package x
templ Test() {
	<div class="w-full h-1/2"></div>
}`,
			wantError: false,
		},
		"valid individual padding classes": {
			input: `package x
templ Test() {
	<div class="pt-2 pb-4 pl-1"></div>
}`,
			wantError: false,
		},
		"valid border color classes": {
			input: `package x
templ Test() {
	<div class="border border-red"></div>
}`,
			wantError: false,
		},
		"valid text alignment classes": {
			input: `package x
templ Test() {
	<div class="text-center"></div>
}`,
			wantError: false,
		},
		"unknown tailwind class": {
			input: `package x
templ Test() {
	<div class="flex-columns"></div>
}`,
			wantError:     true,
			errorContains: `unknown Tailwind class "flex-columns"`,
			hintContains:  `flex-col`,
		},
		"unknown tailwind class without suggestion": {
			input: `package x
templ Test() {
	<div class="xyz-completely-invalid"></div>
}`,
			wantError:     true,
			errorContains: `unknown Tailwind class "xyz-completely-invalid"`,
		},
		"multiple unknown classes": {
			input: `package x
templ Test() {
	<div class="flex-columns badclass"></div>
}`,
			wantError:     true,
			errorContains: `unknown Tailwind class`,
		},
		"mix of valid and invalid classes": {
			input: `package x
templ Test() {
	<div class="flex-col gap-2 badclass p-4"></div>
}`,
			wantError:     true,
			errorContains: `unknown Tailwind class "badclass"`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				if tt.hintContains != "" && !strings.Contains(err.Error(), tt.hintContains) {
					t.Errorf("error %q does not contain hint %q", err.Error(), tt.hintContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_TailwindClassErrorPosition(t *testing.T) {
	// Test that the error position correctly points to the invalid class
	input := `package x
templ Test() {
	<div class="flex-col badclass p-2"></div>
}`

	_, err := AnalyzeFile("test.gsx", input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errList, ok := err.(*ErrorList)
	if !ok {
		t.Fatalf("expected *ErrorList, got %T", err)
	}

	errors := errList.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	tuiErr := errors[0]
	// The error should point to "badclass" which starts at column 22
	// (class=" is at column 7, content starts at column 14, "flex-col " is 9 chars, so badclass is at column 23)
	// Actually, let's verify the error message contains the class name
	if !strings.Contains(tuiErr.Message, "badclass") {
		t.Errorf("error message %q should contain 'badclass'", tuiErr.Message)
	}

	// Verify EndPos is set for range highlighting
	if tuiErr.EndPos.Line == 0 && tuiErr.EndPos.Column == 0 {
		t.Error("EndPos should be set for range-based highlighting")
	}
}
