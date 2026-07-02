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
