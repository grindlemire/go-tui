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
		"reserved helper name updatePropsFields is rejected": {
			input: `package x

import tui "github.com/grindlemire/go-tui"

type row struct{ v string }

func (r *row) updatePropsFields(fresh tui.Component) {}

templ (r *row) Render() {
	<span>{r.v}</span>
}`,
			wantError:     true,
			errorContains: "reserved generated helper name",
		},
		"reserved helper name bindAppFields is rejected": {
			input: `package x

import tui "github.com/grindlemire/go-tui"

type row struct{ v string }

func (r *row) bindAppFields(app *tui.App) {}

templ (r *row) Render() {
	<span>{r.v}</span>
}`,
			wantError:     true,
			errorContains: "reserved generated helper name",
		},
		"reserved helper name on a type without a templ is fine": {
			input: `package x

import tui "github.com/grindlemire/go-tui"

type other struct{}

func (o *other) updatePropsFields(fresh tui.Component) {}

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}`,
			wantError: false,
		},
		"calling the helper from an override is fine": {
			input: `package x

import tui "github.com/grindlemire/go-tui"

type row struct{ v string }

func (r *row) UpdateProps(fresh tui.Component) {
	r.updatePropsFields(fresh)
}

templ (r *row) Render() {
	<span>{r.v}</span>
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

// TestAnalyzer_CrossFileNameCollisions verifies the same collisions are
// caught when the conflicting declaration lives in a sibling file of the
// package, supplied via PackageContext.
func TestAnalyzer_CrossFileNameCollisions(t *testing.T) {
	type tc struct {
		input         string
		siblingGo     string // parsed as a sibling .go file
		siblingGSX    string // parsed as a sibling .gsx file
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"templ conflicts with function in sibling file": {
			input: `package x

templ Foo() {
	<span>hi</span>
}`,
			siblingGo:     "package x\n\nfunc Foo() string { return \"x\" }",
			wantError:     true,
			errorContains: "conflicts with a Go function",
		},
		"templ duplicated in sibling gsx file": {
			input: `package x

templ Foo() {
	<span>hi</span>
}`,
			siblingGSX: `package x

templ Foo() {
	<span>other</span>
}`,
			wantError:     true,
			errorContains: "another file",
		},
		"view struct conflicts with type in sibling file": {
			input: `package x

templ Foo() {
	<span>hi</span>
}`,
			siblingGo:     "package x\n\ntype FooView struct{ n int }",
			wantError:     true,
			errorContains: "conflicts with the view struct",
		},
		"handwritten Render in sibling file": {
			input: `package x

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}`,
			siblingGo:     "package x\n\nimport tui \"github.com/grindlemire/go-tui\"\n\nfunc (r *row) Render(app *tui.App) *tui.Element { return nil }",
			wantError:     true,
			errorContains: "already declares a Render method",
		},
		"Render templ duplicated in sibling gsx file": {
			input: `package x

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}`,
			siblingGSX: `package x

templ (r *row) Render() {
	<span>other</span>
}`,
			wantError:     true,
			errorContains: "another file",
		},
		"unrelated sibling declarations are fine": {
			input: `package x

templ Foo() {
	<span>hi</span>
}`,
			siblingGo: "package x\n\nfunc Bar() {}\n\ntype BarView struct{}",
			wantError: false,
		},
		"reserved helper name in sibling file is rejected": {
			input: `package x

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}`,
			siblingGo:     "package x\n\nimport tui \"github.com/grindlemire/go-tui\"\n\nfunc (r *row) unbindAppFields() {}",
			wantError:     true,
			errorContains: "reserved generated helper name",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := NewPackageContext()
			if tt.siblingGo != "" {
				if err := ctx.AddGoSource("sibling.go", tt.siblingGo); err != nil {
					t.Fatalf("AddGoSource failed: %v", err)
				}
			}
			if tt.siblingGSX != "" {
				lexer := NewLexer("sibling.gsx", tt.siblingGSX)
				parser := NewParser(lexer)
				sib, err := parser.ParseFile()
				if err != nil {
					t.Fatalf("sibling parse failed: %v", err)
				}
				ctx.AddGSXFile(sib)
			}

			lexer := NewLexer("test.gsx", tt.input)
			parser := NewParser(lexer)
			file, err := parser.ParseFile()
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			analyzer := NewAnalyzer()
			analyzer.SetPackageContext(ctx)
			err = analyzer.Analyze(file)

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
