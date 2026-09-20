package tuigen

import (
	"strings"
	"testing"
)

// A templ renders only its first top-level element, but the generator still
// emitted the rest, so a trailing <modal> mounted and drew as an overlay while
// never joining the element tree (issue #169). The analyzer now rejects a body
// with more than one top-level element.
func TestAnalyzer_SingleRootElement(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
		hintContains  string
		posContains   string
	}

	tests := map[string]tc{
		"element followed by modal errors": {
			input: `package x
type app struct{}
templ (a *app) Render() {
	<div>
		<span>hi</span>
	</div>
	<modal open={a.show}>
		<span>dialog</span>
	</modal>
}`,
			wantError:     true,
			errorContains: "templ Render has more than one top-level element",
			hintContains:  "wrap them in a single root element",
			posContains:   "test.gsx:7:",
		},
		"element followed by component call errors": {
			input: `package x
templ Page() {
	<div>hi</div>
	@Footer()
}`,
			wantError:     true,
			errorContains: "templ Page has more than one top-level element",
			posContains:   "test.gsx:4:",
		},
		"element followed by component expr errors": {
			input: `package x
type app struct{}
templ (a *app) Render() {
	<div>hi</div>
	@a.footer
}`,
			wantError:     true,
			errorContains: "templ Render has more than one top-level element",
		},
		"element followed by for loop errors": {
			input: `package x
templ List(items []string) {
	<div>hi</div>
	for _, item := range items {
		<span>{item}</span>
	}
}`,
			wantError:     true,
			errorContains: "templ List has more than one top-level element",
		},
		"if statement followed by element errors": {
			input: `package x
templ Page(show bool) {
	if show {
		<span>on</span>
	}
	<div>hi</div>
}`,
			wantError:     true,
			errorContains: "templ Page has more than one top-level element",
		},
		"three elements report once at the second": {
			input: `package x
templ Page() {
	<div>a</div>
	<div>b</div>
	<div>c</div>
}`,
			wantError:     true,
			errorContains: "test.gsx:4:",
		},
		"let binding before the root is allowed": {
			input: `package x
templ Page() {
	label := <span>hi</span>
	<div>{label}</div>
}`,
			wantError: false,
		},
		"modal nested inside the root is allowed": {
			input: `package x
type app struct{}
templ (a *app) Render() {
	<div>
		<span>hi</span>
		<modal open={a.show}>
			<span>dialog</span>
		</modal>
	</div>
}`,
			wantError: false,
		},
		"modal as the only root is allowed": {
			input: `package x
type app struct{}
templ (a *app) Render() {
	<modal open={a.show}>
		<span>dialog</span>
	</modal>
}`,
			wantError: false,
		},
		"if else as the only root is allowed": {
			input: `package x
templ Page(show bool) {
	if show {
		<span>on</span>
	} else {
		<span>off</span>
	}
}`,
			wantError: false,
		},
		"for loop as the only root is allowed": {
			input: `package x
templ List(items []string) {
	for _, item := range items {
		<span>{item}</span>
	}
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)
			if !tt.wantError {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error, got nil")
			}
			msg := err.Error()
			if !strings.Contains(msg, tt.errorContains) {
				t.Errorf("error %q does not contain %q", msg, tt.errorContains)
			}
			if tt.hintContains != "" && !strings.Contains(msg, tt.hintContains) {
				t.Errorf("error %q does not contain hint %q", msg, tt.hintContains)
			}
			if tt.posContains != "" && !strings.Contains(msg, tt.posContains) {
				t.Errorf("error %q is not anchored at %q", msg, tt.posContains)
			}
			if n := strings.Count(msg, "more than one top-level element"); n != 1 {
				t.Errorf("expected exactly one report, got %d in %q", n, msg)
			}
		})
	}
}
