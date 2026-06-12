package tuigen

import (
	"strings"
	"testing"
)

func TestParser_GenericFuncCapturedRaw(t *testing.T) {
	input := `package x

func identity[T any](v T) T {
	return v
}

templ App() {
	<div></div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Funcs) != 1 {
		t.Fatalf("expected 1 raw Go func, got %d", len(file.Funcs))
	}
	if !strings.Contains(file.Funcs[0].Code, "func identity[T any](v T) T") {
		t.Errorf("raw func code missing generic signature, got:\n%s", file.Funcs[0].Code)
	}
	if len(file.Components) != 1 || file.Components[0].Name != "App" {
		t.Errorf("expected App component to still parse, got %+v", file.Components)
	}
}

func TestParser_MethodFuncCapturedRaw(t *testing.T) {
	input := `package x

func (s *store) get(key string) string {
	return s.data[key]
}

templ App() {
	<div></div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Funcs) != 1 {
		t.Fatalf("expected 1 raw Go func, got %d", len(file.Funcs))
	}
	if !strings.Contains(file.Funcs[0].Code, "func (s *store) get(key string) string") {
		t.Errorf("raw func code missing method signature, got:\n%s", file.Funcs[0].Code)
	}
}

func TestParser_OldStyleElementComponent(t *testing.T) {
	// func Name() Element parses as a component with a DSL body.
	input := `package x

func Card() Element {
	<div>
		<span>hi</span>
	</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(file.Components))
	}
	comp := file.Components[0]
	if comp.Name != "Card" {
		t.Errorf("component name = %q, want %q", comp.Name, "Card")
	}
	if comp.ReturnType != "*element.Element" {
		t.Errorf("return type = %q, want %q", comp.ReturnType, "*element.Element")
	}
	if len(comp.Body) == 0 {
		t.Error("component body is empty, want parsed element tree")
	}
}

func TestParser_FuncAndTemplErrors(t *testing.T) {
	type tc struct {
		input   string
		wantErr string
	}

	tests := map[string]tc{
		"func without name": {
			input: `package x

func 123() {}
`,
			wantErr: "expected function name",
		},
		"unterminated function": {
			input: `package x

func helper() {
	x := 1`,
			wantErr: "unterminated function definition",
		},
		"unterminated method func": {
			input: `package x

func (s *store) helper() {
	x := 1`,
			wantErr: "unterminated function definition",
		},
		"templ without name": {
			input: `package x

templ 123() {
	<div></div>
}`,
			wantErr: "expected component name or method receiver",
		},
		"method templ wrong method name": {
			input: `package x

templ (c *shell) Draw() {
	<div></div>
}`,
			wantErr: "method templ name must be 'Render'",
		},
		"method templ with parameters": {
			input: `package x

templ (c *shell) Render(w int) {
	<div></div>
}`,
			wantErr: "method templ Render() must not have parameters",
		},
		"method templ missing receiver name": {
			input: `package x

templ (*shell) Render() {
	<div></div>
}`,
			wantErr: "expected receiver name",
		},
		"method templ non-ident method name": {
			input: `package x

templ (c *shell) 123() {
	<div></div>
}`,
			wantErr: "expected method name after receiver",
		},
		"param without name": {
			input: `package x

templ App(123 int) {
	<div></div>
}`,
			wantErr: "expected parameter name",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			p := NewParser(l)
			_, err := p.ParseFile()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParser_GoDeclForms(t *testing.T) {
	type tc struct {
		input        string
		wantKind     string
		wantContains string
	}

	tests := map[string]tc{
		"simple var": {
			input: `package x

var count = 42

templ App() {
	<div></div>
}`,
			wantKind:     "var",
			wantContains: "var count = 42",
		},
		"grouped const": {
			input: `package x

const (
	A = 1
	B = 2
)

templ App() {
	<div></div>
}`,
			wantKind:     "const",
			wantContains: "B = 2",
		},
		"grouped var": {
			input: `package x

var (
	x = "a"
	y = "b"
)

templ App() {
	<div></div>
}`,
			wantKind:     "var",
			wantContains: `y = "b"`,
		},
		"type struct": {
			input: `package x

type model struct {
	name string
}

templ App() {
	<div></div>
}`,
			wantKind:     "type",
			wantContains: "type model struct {",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			p := NewParser(l)
			file, err := p.ParseFile()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(file.Decls) != 1 {
				t.Fatalf("expected 1 decl, got %d", len(file.Decls))
			}
			decl := file.Decls[0]
			if decl.Kind != tt.wantKind {
				t.Errorf("decl kind = %q, want %q", decl.Kind, tt.wantKind)
			}
			if !strings.Contains(decl.Code, tt.wantContains) {
				t.Errorf("decl code missing %q, got:\n%s", tt.wantContains, decl.Code)
			}
		})
	}
}

func TestParser_GoDeclAtEOF(t *testing.T) {
	input := `package x

var trailing = 1`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(file.Decls))
	}
	if !strings.Contains(file.Decls[0].Code, "var trailing = 1") {
		t.Errorf("decl code = %q, want it to contain %q", file.Decls[0].Code, "var trailing = 1")
	}
}

func TestParser_TemplParamsTrailingComma(t *testing.T) {
	input := `package x

templ App(
	title string,
	count int,
) {
	<div></div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(file.Components))
	}
	params := file.Components[0].Params
	if len(params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(params))
	}
	if params[0].Name != "title" || params[0].Type != "string" {
		t.Errorf("param 0 = (%q, %q), want (title, string)", params[0].Name, params[0].Type)
	}
	if params[1].Name != "count" || params[1].Type != "int" {
		t.Errorf("param 1 = (%q, %q), want (count, int)", params[1].Name, params[1].Type)
	}
}

func TestParser_MethodTemplReceiverWithParens(t *testing.T) {
	// Receiver types containing parentheses exercise the depth tracking in
	// parseMethodTempl's receiver capture.
	input := `package x

templ (c (shell)) Render() {
	<div></div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(file.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(file.Components))
	}
	comp := file.Components[0]
	if comp.ReceiverType != "(shell)" {
		t.Errorf("receiver type = %q, want %q", comp.ReceiverType, "(shell)")
	}
	if comp.ReceiverName != "c" {
		t.Errorf("receiver name = %q, want %q", comp.ReceiverName, "c")
	}
}
