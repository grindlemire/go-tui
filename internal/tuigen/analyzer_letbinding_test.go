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

// TestAnalyzer_LetBindingScopedToTempl verifies that := binding names only
// rewrite {name} references inside the templ that declares them. The binding
// registry used to be collected for the whole file first, so a binding in one
// templ turned a same-named string param in a sibling templ into AddChild.
func TestAnalyzer_LetBindingScopedToTempl(t *testing.T) {
	type tc struct {
		input             string
		wantAddChildCount int
		wantContains      []string
		wantNotContains   []string
	}

	tests := map[string]tc{
		"binding in one templ does not leak into a sibling templ": {
			wantAddChildCount: 1,
			input: `package demo

templ A() {
	label := <span>bound</span>
	<div>{label}</div>
}

templ B(label string) {
	<div>{label}</div>
}`,
			wantContains: []string{
				"label := tui.New(",
				".AddChild(label)",
				"tui.New(tui.WithText(label))",
			},
		},
		"binding declared after a sibling templ still resolves in its own templ": {
			wantAddChildCount: 1,
			input: `package demo

templ B(label string) {
	<div>{label}</div>
}

templ A() {
	label := <span>bound</span>
	<div>
		<span>first</span>
		{label}
	</div>
}`,
			wantContains: []string{
				"tui.New(tui.WithText(label))",
				".AddChild(label)",
			},
		},
		"binding inside an if body resolves in the same templ only": {
			wantAddChildCount: 1,
			input: `package demo

templ A(show bool) {
	if show {
		label := <span>bound</span>
		<div>{label}</div>
	} else {
		<div>none</div>
	}
}

templ B(label string) {
	<div>{label}</div>
}`,
			wantContains: []string{
				"if show {",
				".AddChild(label)",
				"tui.New(tui.WithText(label))",
			},
		},
		"two templs binding the same name each resolve their own": {
			wantAddChildCount: 2,
			input: `package demo

templ A() {
	label := <span>a</span>
	<div>{label}</div>
}

templ B() {
	label := <span>b</span>
	<div>{label}</div>
}`,
			wantContains: []string{
				`tui.WithText("a")`,
				`tui.WithText("b")`,
			},
			wantNotContains: []string{
				"tui.New(tui.WithText(label))",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			code := parseAnalyzeGenerate(t, tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, code)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(code, notWant) {
					t.Errorf("output contains unexpected %q\nGot:\n%s", notWant, code)
				}
			}
			if got := strings.Count(code, ".AddChild(label)"); got != tt.wantAddChildCount {
				t.Errorf("AddChild(label) count = %d, want %d\nGot:\n%s", got, tt.wantAddChildCount, code)
			}
		})
	}
}
