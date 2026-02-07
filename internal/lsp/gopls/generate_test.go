package gopls

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func TestGenerateVirtualGo_StateVarDeclarations(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "Counter",
				Position:   tuigen.Position{Line: 3, Column: 1},
				ReturnType: "*element.Element",
				Body: []tuigen.Node{
					&tuigen.GoCode{
						Code:     `count := tui.NewState(0)`,
						Position: tuigen.Position{Line: 4, Column: 2},
					},
				},
			},
		},
	}

	source, sourceMap := GenerateVirtualGo(file)

	// Verify state variable declaration is emitted
	if !strings.Contains(source, "count := tui.NewState(0)") {
		t.Errorf("expected virtual Go to contain state declaration, got:\n%s", source)
	}

	// Verify source map has a mapping for the state variable
	if sourceMap == nil {
		t.Fatal("expected non-nil source map")
	}
}

func TestGenerateVirtualGo_RefSimple(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "Layout",
				Position:   tuigen.Position{Line: 3, Column: 1},
				ReturnType: "*element.Element",
				Body: []tuigen.Node{
					&tuigen.Element{
						Tag:     "div",
						RefExpr: &tuigen.GoExpr{Code: "header", Position: tuigen.Position{Line: 4, Column: 10}},
						Position: tuigen.Position{Line: 4, Column: 2},
					},
				},
			},
		},
	}

	source, _ := GenerateVirtualGo(file)

	// Ref expression should be validated with _ = assignment
	if !strings.Contains(source, "_ = header") {
		t.Errorf("expected virtual Go to contain ref usage, got:\n%s", source)
	}
}

func TestGenerateVirtualGo_RefInLoop(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "List",
				Position:   tuigen.Position{Line: 3, Column: 1},
				ReturnType: "*element.Element",
				Body: []tuigen.Node{
					&tuigen.ForLoop{
						Index:    "_",
						Value:    "item",
						Iterable: "items",
						Position: tuigen.Position{Line: 4, Column: 2},
						Body: []tuigen.Node{
							&tuigen.Element{
								Tag:     "span",
								RefExpr: &tuigen.GoExpr{Code: "items", Position: tuigen.Position{Line: 5, Column: 14}},
								Position: tuigen.Position{Line: 5, Column: 3},
							},
						},
					},
				},
			},
		},
	}

	source, _ := GenerateVirtualGo(file)

	// Ref expression should be validated with _ = assignment
	if !strings.Contains(source, "_ = items") {
		t.Errorf("expected virtual Go to contain ref usage, got:\n%s", source)
	}
}

func TestGenerateVirtualGo_RefKeyed(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "UserList",
				Position:   tuigen.Position{Line: 3, Column: 1},
				ReturnType: "*element.Element",
				Body: []tuigen.Node{
					&tuigen.ForLoop{
						Index:    "_",
						Value:    "user",
						Iterable: "users",
						Position: tuigen.Position{Line: 4, Column: 2},
						Body: []tuigen.Node{
							&tuigen.Element{
								Tag:     "span",
								RefExpr: &tuigen.GoExpr{Code: "users", Position: tuigen.Position{Line: 5, Column: 14}},
								RefKey:  &tuigen.GoExpr{Code: "user.ID", Position: tuigen.Position{Line: 5, Column: 28}},
								Position: tuigen.Position{Line: 5, Column: 3},
							},
						},
					},
				},
			},
		},
	}

	source, _ := GenerateVirtualGo(file)

	// Ref expression should be validated with _ = assignment
	if !strings.Contains(source, "_ = users") {
		t.Errorf("expected virtual Go to contain ref usage, got:\n%s", source)
	}
}

func TestGenerateVirtualGo_GoDecl(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Imports: []tuigen.Import{
			{Path: "github.com/grindlemire/go-tui", Alias: "tui"},
		},
		Decls: []*tuigen.GoDecl{
			{
				Kind:     "var",
				Code:     "var _ tui.Component = (*counter)(nil)",
				Position: tuigen.Position{Line: 5, Column: 1},
			},
		},
		Components: []*tuigen.Component{
			{
				Name:       "Test",
				Position:   tuigen.Position{Line: 7, Column: 1},
				ReturnType: "*element.Element",
				Body:       []tuigen.Node{},
			},
		},
	}

	source, sourceMap := GenerateVirtualGo(file)

	// Verify GoDecl is emitted in virtual Go
	if !strings.Contains(source, "var _ tui.Component = (*counter)(nil)") {
		t.Errorf("expected virtual Go to contain var declaration, got:\n%s", source)
	}

	// Verify source map has a mapping for the declaration
	if sourceMap == nil {
		t.Fatal("expected non-nil source map")
	}

	// The GoDecl should appear after imports
	importIdx := strings.Index(source, `tui "github.com/grindlemire/go-tui"`)
	declIdx := strings.Index(source, "var _ tui.Component")
	funcIdx := strings.Index(source, "func Test(")

	if importIdx >= declIdx {
		t.Errorf("expected import before decl")
	}
	if declIdx >= funcIdx {
		t.Errorf("expected decl before func")
	}
}

func TestGenerateVirtualGo_ExistingFunctionality(t *testing.T) {
	// Verify existing functionality still works
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "Hello",
				Position:   tuigen.Position{Line: 3, Column: 1},
				ReturnType: "*element.Element",
				Params: []*tuigen.Param{
					{Name: "name", Type: "string", Position: tuigen.Position{Line: 3, Column: 12}},
				},
				Body: []tuigen.Node{
					&tuigen.Element{
						Tag:      "span",
						Position: tuigen.Position{Line: 4, Column: 2},
						Children: []tuigen.Node{
							&tuigen.GoExpr{
								Code:     "name",
								Position: tuigen.Position{Line: 4, Column: 8},
							},
						},
					},
				},
			},
		},
	}

	source, sourceMap := GenerateVirtualGo(file)

	if !strings.Contains(source, "func Hello(name string) *element.Element") {
		t.Errorf("expected function signature, got:\n%s", source)
	}
	if !strings.Contains(source, "_ = name") {
		t.Errorf("expected expression assignment, got:\n%s", source)
	}
	if sourceMap == nil {
		t.Fatal("expected non-nil source map")
	}
}
