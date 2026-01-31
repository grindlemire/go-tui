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

func TestGenerateVirtualGo_NamedRefSimple(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "Layout",
				Position:   tuigen.Position{Line: 3, Column: 1},
				ReturnType: "*element.Element",
				Body: []tuigen.Node{
					&tuigen.Element{
						Tag:      "div",
						NamedRef: "Header",
						Position: tuigen.Position{Line: 4, Column: 2},
					},
				},
			},
		},
	}

	source, _ := GenerateVirtualGo(file)

	// Simple ref should be *element.Element
	if !strings.Contains(source, "var Header *element.Element") {
		t.Errorf("expected virtual Go to contain simple ref declaration, got:\n%s", source)
	}
}

func TestGenerateVirtualGo_NamedRefInLoop(t *testing.T) {
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
								Tag:      "span",
								NamedRef: "Items",
								Position: tuigen.Position{Line: 5, Column: 3},
							},
						},
					},
				},
			},
		},
	}

	source, _ := GenerateVirtualGo(file)

	// Loop ref should be []*element.Element
	if !strings.Contains(source, "var Items []*element.Element") {
		t.Errorf("expected virtual Go to contain loop ref slice declaration, got:\n%s", source)
	}
}

func TestGenerateVirtualGo_NamedRefKeyed(t *testing.T) {
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
								Tag:      "span",
								NamedRef: "Users",
								RefKey:   &tuigen.GoExpr{Code: "user.ID", Position: tuigen.Position{Line: 5, Column: 20}},
								Position: tuigen.Position{Line: 5, Column: 3},
							},
						},
					},
				},
			},
		},
	}

	source, _ := GenerateVirtualGo(file)

	// Keyed ref should be map[string]*element.Element
	if !strings.Contains(source, "var Users map[string]*element.Element") {
		t.Errorf("expected virtual Go to contain keyed ref map declaration, got:\n%s", source)
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
