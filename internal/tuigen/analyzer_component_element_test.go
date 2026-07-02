package tuigen

import (
	"strings"
	"testing"
)

// Component elements (textarea/input/modal/markdown) mount via app.Mount against
// the host component's receiver. A function templ has no receiver, so using one
// there generated invalid Go (issue #111). The analyzer now rejects it with a
// message that points at the struct-component pattern.
func TestAnalyzer_ComponentElementRequiresReceiver(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
		hintContains  string
	}

	tests := map[string]tc{
		"textarea in function templ errors": {
			input: `package x
templ MyForm() {
	<textarea placeholder="Type here..." />
}`,
			wantError:     true,
			errorContains: "<textarea> can only be used inside a struct component",
			hintContains:  "give MyForm a receiver",
		},
		"input in function templ errors": {
			input: `package x
templ MyForm() {
	<input placeholder="Name" />
}`,
			wantError:     true,
			errorContains: "<input> can only be used inside a struct component",
		},
		"markdown in function templ errors": {
			input: `package x
templ Doc(src string) {
	<markdown source={src} />
}`,
			wantError:     true,
			errorContains: "<markdown> can only be used inside a struct component",
		},
		"modal in function templ errors": {
			input: `package x
templ Dialog(show bool) {
	<modal open={show}>
		<span>hi</span>
	</modal>
}`,
			wantError:     true,
			errorContains: "<modal> can only be used inside a struct component",
		},
		"nested in for loop in function templ errors": {
			input: `package x
templ List(items []string) {
	<div>
		for _, item := range items {
			<textarea placeholder={item} />
		}
	</div>
}`,
			wantError:     true,
			errorContains: "<textarea> can only be used inside a struct component",
		},
		"nested in if in function templ errors": {
			input: `package x
templ Maybe(show bool) {
	<div>
		if show {
			<input />
		}
	</div>
}`,
			wantError:     true,
			errorContains: "<input> can only be used inside a struct component",
		},
		"in let binding in function templ errors": {
			input: `package x
templ MyForm() {
	editor := <textarea />
	<div>{editor}</div>
}`,
			wantError:     true,
			errorContains: "<textarea> can only be used inside a struct component",
		},
		"in component-call children in function templ errors": {
			input: `package x
templ MyForm() {
	@Card() {
		<textarea />
	}
}`,
			wantError:     true,
			errorContains: "<textarea> can only be used inside a struct component",
		},
		"textarea in method templ is allowed": {
			input: `package x
type myForm struct{}
templ (c *myForm) Render() {
	<textarea placeholder="Type here..." />
}`,
			wantError: false,
		},
		"modal in method templ is allowed": {
			input: `package x
type myForm struct{}
templ (c *myForm) Render() {
	<modal open={c.show}>
		<span>hi</span>
	</modal>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				if tt.hintContains != "" && !strings.Contains(err.Error(), tt.hintContains) {
					t.Errorf("error %q does not contain hint %q", err.Error(), tt.hintContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// Component expressions (@expr) render via expr.Render(app). A function templ
// body has no app in scope, so the analyzer rejects them there while allowing
// them in method templs, whose Render(app) receiver supplies app.
func TestAnalyzer_ComponentExprRequiresReceiver(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
		hintContains  string
	}

	tests := map[string]tc{
		"component expr in function templ errors": {
			input: `package x
templ Foo(child tui.Component) {
	<div>@child</div>
}`,
			wantError:     true,
			errorContains: "component expression @child can only be used inside a struct component",
			hintContains:  "give Foo a receiver",
		},
		"component expr nested in for loop in function templ errors": {
			input: `package x
templ List(kids []tui.Component) {
	<div>
		for _, kid := range kids {
			@kid
		}
	</div>
}`,
			wantError:     true,
			errorContains: "component expression @kid can only be used inside a struct component",
		},
		"component expr in let binding in function templ errors": {
			input: `package x
templ Foo(child tui.Component) {
	thing := @child
	<div>{thing}</div>
}`,
			wantError:     true,
			errorContains: "component expression @child can only be used inside a struct component",
		},
		"component expr nested in if in function templ errors": {
			input: `package x
templ Maybe(show bool, child tui.Component) {
	<div>
		if show {
			@child
		}
	</div>
}`,
			wantError:     true,
			errorContains: "component expression @child can only be used inside a struct component",
		},
		"component expr in method templ is allowed": {
			input: `package x
type host struct{ child tui.Component }
templ (h *host) Render() {
	<div>@h.child</div>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				if tt.hintContains != "" && !strings.Contains(err.Error(), tt.hintContains) {
					t.Errorf("error %q does not contain hint %q", err.Error(), tt.hintContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// A @Factory() call that returns a struct component mounts via app.Mount against
// the caller's receiver. In a function templ the generator instead emits a plain
// call and reads .Root on a type that has none. The analyzer rejects the call in
// a function templ, but leaves it (and plain function-component calls) alone in a
// method templ.
func TestAnalyzer_StructComponentCallRequiresReceiver(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
		hintContains  string
	}

	// Shared preamble: a struct component `widget` with a factory `Widget()`.
	const structComp = `package x
type widget struct{}
func Widget() *widget { return &widget{} }
templ (w *widget) Render() {
	<span>hi</span>
}
`

	tests := map[string]tc{
		"struct factory call in function templ errors": {
			input: structComp + `templ Page() {
	<div>@Widget()</div>
}`,
			wantError:     true,
			errorContains: "@Widget() mounts a struct component and can only be used inside a struct component",
			hintContains:  "give Page a receiver",
		},
		"struct factory with params call in function templ errors": {
			input: `package x
type widget struct{}
func Widget(id string) *widget { return &widget{} }
templ (w *widget) Render() {
	<span>hi</span>
}
templ Page() {
	<div>@Widget("a")</div>
}`,
			wantError:     true,
			errorContains: "@Widget() mounts a struct component and can only be used inside a struct component",
		},
		"struct factory call in method templ is allowed": {
			input: structComp + `type host struct{}
templ (h *host) Render() {
	<div>@Widget()</div>
}`,
			wantError: false,
		},
		"struct factory used in a go expression is not flagged": {
			input: structComp + `templ Page() {
	<div>{Widget()}</div>
}`,
			wantError: false,
		},
		"function component call in function templ is allowed": {
			input: `package x
templ Badge() {
	<span>b</span>
}
templ Page() {
	<div>@Badge()</div>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				if tt.hintContains != "" && !strings.Contains(err.Error(), tt.hintContains) {
					t.Errorf("error %q does not contain hint %q", err.Error(), tt.hintContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
