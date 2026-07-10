package tuigen

import (
	"testing"
)

func TestPackageContext_AddGoSource(t *testing.T) {
	type tc struct {
		src            string
		wantMethods    map[string][]string // type -> methods present
		wantNotMethods map[string][]string // type -> methods absent
		wantFuncs      []string
		wantTypes      []string
	}

	tests := map[string]tc{
		"pointer receiver method": {
			src: `package x

import tui "github.com/grindlemire/go-tui"

func (r *row) UpdateProps(fresh tui.Component) {}`,
			wantMethods: map[string][]string{"row": {"UpdateProps"}},
		},
		"value receiver method": {
			src: `package x

func (r row) BindApp(app *App) {}`,
			wantMethods: map[string][]string{"row": {"BindApp"}},
		},
		"unnamed receiver method": {
			src: `package x

func (*row) UnbindApp() {}`,
			wantMethods: map[string][]string{"row": {"UnbindApp"}},
		},
		"generic receiver method": {
			src: `package x

func (r *box[T]) UpdateProps(fresh Component) {}`,
			wantMethods: map[string][]string{"box": {"UpdateProps"}},
		},
		"plain function is not a method": {
			src: `package x

func UpdateProps() {}`,
			wantFuncs:      []string{"UpdateProps"},
			wantNotMethods: map[string][]string{"row": {"UpdateProps"}},
		},
		"type declarations including grouped": {
			src: `package x

type FooView struct{ n int }

type (
	BarView int
	other   string
)`,
			wantTypes: []string{"FooView", "BarView", "other"},
		},
		"method mentioned only in a comment is ignored": {
			src: `package x

// func (r *row) UpdateProps(fresh tui.Component) {}
func helper() {}`,
			wantFuncs:      []string{"helper"},
			wantNotMethods: map[string][]string{"row": {"UpdateProps"}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := NewPackageContext()
			if err := ctx.AddGoSource("sibling.go", tt.src); err != nil {
				t.Fatalf("AddGoSource failed: %v", err)
			}

			for typeName, methods := range tt.wantMethods {
				for _, m := range methods {
					if !ctx.HasMethod(typeName, m) {
						t.Errorf("HasMethod(%q, %q) = false, want true", typeName, m)
					}
				}
			}
			for typeName, methods := range tt.wantNotMethods {
				for _, m := range methods {
					if ctx.HasMethod(typeName, m) {
						t.Errorf("HasMethod(%q, %q) = true, want false", typeName, m)
					}
				}
			}
			for _, f := range tt.wantFuncs {
				if !ctx.HasFunc(f) {
					t.Errorf("HasFunc(%q) = false, want true", f)
				}
			}
			for _, ty := range tt.wantTypes {
				if !ctx.HasType(ty) {
					t.Errorf("HasType(%q) = false, want true", ty)
				}
			}
		})
	}
}

func TestPackageContext_AddGoSource_ParseError(t *testing.T) {
	ctx := NewPackageContext()
	if err := ctx.AddGoSource("bad.go", "package x\nfunc ("); err == nil {
		t.Error("expected parse error, got nil")
	}
}

func TestPackageContext_AddGSXFile(t *testing.T) {
	input := `package x

import tui "github.com/grindlemire/go-tui"

type helper struct{}

func topLevel() string { return "x" }

func (h *helper) UpdateProps(fresh tui.Component) {}

templ FnTempl() {
	<span>hi</span>
}

templ (h *helper) Render() {
	<span>hi</span>
}`

	lexer := NewLexer("sibling.gsx", input)
	parser := NewParser(lexer)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	ctx := NewPackageContext()
	ctx.AddGSXFile(file)

	if !ctx.HasFunc("topLevel") {
		t.Error("HasFunc(topLevel) = false, want true")
	}
	if !ctx.HasType("helper") {
		t.Error("HasType(helper) = false, want true")
	}
	if !ctx.HasMethod("helper", "UpdateProps") {
		t.Error("HasMethod(helper, UpdateProps) = false, want true")
	}
	if !ctx.HasTempl("FnTempl") {
		t.Error("HasTempl(FnTempl) = false, want true")
	}
	if !ctx.HasRenderTempl("helper") {
		t.Error("HasRenderTempl(helper) = false, want true")
	}
	if ctx.HasTempl("helper") {
		t.Error("HasTempl(helper) = true, want false (method templ, not function templ)")
	}
}
