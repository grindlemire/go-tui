package tuigen

import (
	"strings"
	"testing"
)

// TestAnalyzer_NameCollisions verifies that declarations colliding with
// generated code are rejected at analysis time with an error pointing at the
// .gsx source, instead of surfacing as a raw Go "redeclared" error inside the
// generated file.
func TestAnalyzer_NameCollisions(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"templ conflicts with Go function of the same name": {
			input: `package x

func Foo() string { return "x" }

templ Foo() {
	<span>hi</span>
}`,
			wantError:     true,
			errorContains: "conflicts with a Go function",
		},
		"duplicate function templ names": {
			input: `package x

templ Foo() {
	<span>one</span>
}

templ Foo() {
	<span>two</span>
}`,
			wantError:     true,
			errorContains: "duplicate templ",
		},
		"handwritten Render method alongside a method templ": {
			input: `package x

import tui "github.com/grindlemire/go-tui"

type row struct{ v string }

func (r *row) Render(app *tui.App) *tui.Element { return nil }

templ (r *row) Render() {
	<span>{r.v}</span>
}`,
			wantError:     true,
			errorContains: "already declares a Render method",
		},
		"user type collides with generated view struct": {
			input: `package x

type FooView struct{ n int }

templ Foo() {
	<span>hi</span>
}`,
			wantError:     true,
			errorContains: "conflicts with the view struct",
		},
		"duplicate Render templs on the same receiver": {
			input: `package x

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}

templ (r *row) Render() {
	<span>again</span>
}`,
			wantError:     true,
			errorContains: "duplicate Render templ",
		},
		"unrelated Go function name is fine": {
			input: `package x

func helper() string { return "x" }

templ Foo() {
	<span>hi</span>
}`,
			wantError: false,
		},
		"Render method on a different type is fine": {
			input: `package x

import tui "github.com/grindlemire/go-tui"

type other struct{}

func (o *other) Render(app *tui.App) *tui.Element { return nil }

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}`,
			wantError: false,
		},
		"user type without the View suffix is fine": {
			input: `package x

type FooModel struct{ n int }

templ Foo() {
	<span>hi</span>
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
