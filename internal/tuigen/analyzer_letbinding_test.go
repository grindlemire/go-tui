package tuigen

import (
	"strings"
	"testing"
)

// TestAnalyzer_LetBindingNonElementRHS verifies that the analyzer handles
// let bindings whose right-hand side is a component call or component
// expression. The ref walk used to dereference LetBinding.Element without a
// nil check, so `badge := @Badge("hi")` in a function templ crashed
// tui generate with a nil pointer panic.
func TestAnalyzer_LetBindingNonElementRHS(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"call RHS at top level of function templ": {
			input: `package x

templ Badge(label string) {
	<span>{label}</span>
}

templ App() {
	badge := @Badge("hi")
	<div>{badge}</div>
}`,
			wantContains: []string{
				"__tui_0 := Badge(\"hi\")",
				"badge := __tui_0.Root",
			},
		},
		"call RHS nested in component call children": {
			input: `package x

templ Badge(label string) {
	<span>{label}</span>
}

templ Wrapper() {
	<div>{children...}</div>
}

templ App() {
	@Wrapper() {
		badge := @Badge("hi")
		<div>{badge}</div>
	}
}`,
			wantContains: []string{
				":= Badge(\"hi\")",
				"badge := ",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}
