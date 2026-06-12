package gopls

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func TestGenerateVirtualGo_Imports(t *testing.T) {
	type tc struct {
		imports      []tuigen.Import
		wantContains []string
	}

	tests := map[string]tc{
		"single import no alias": {
			imports:      []tuigen.Import{{Path: "fmt"}},
			wantContains: []string{"import \"fmt\"\n"},
		},
		"single import with alias": {
			imports:      []tuigen.Import{{Path: "github.com/grindlemire/go-tui", Alias: "tui"}},
			wantContains: []string{"import tui \"github.com/grindlemire/go-tui\"\n"},
		},
		"multiple imports mixed": {
			imports: []tuigen.Import{
				{Path: "fmt"},
				{Path: "github.com/grindlemire/go-tui", Alias: "tui"},
			},
			wantContains: []string{
				"import (\n\t\"fmt\"\n\ttui \"github.com/grindlemire/go-tui\"\n)\n",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			file := &tuigen.File{Package: "main", Imports: tt.imports}
			source, _ := GenerateVirtualGo(file)
			if !strings.HasPrefix(source, "package main\n") {
				t.Errorf("missing package clause:\n%s", source)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(source, want) {
					t.Errorf("generated source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestGenerateVirtualGo_TopLevelFunc(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Funcs: []*tuigen.GoFunc{
			{
				Code:     "func helper(s string) string {\n\treturn s\n}",
				Position: tuigen.Position{Line: 10, Column: 1},
			},
		},
	}

	source, sourceMap := GenerateVirtualGo(file)

	want := "func helper(s string) string {\n\treturn s\n}\n"
	if !strings.Contains(source, want) {
		t.Errorf("generated source missing function body:\n%s", source)
	}

	// Each of the 3 function lines gets a mapping anchored at column 0,
	// starting at gsx line 10 (0-indexed 9).
	if sourceMap.Len() != 3 {
		t.Fatalf("sourceMap.Len() = %d, want 3", sourceMap.Len())
	}
	goLine, goCol, found := sourceMap.TuiToGo(9, 5)
	if !found {
		t.Fatal("no mapping for first function line")
	}
	// "package main" + blank = 2 lines, so the func starts at go line 2.
	if goLine != 2 || goCol != 5 {
		t.Errorf("TuiToGo(9, 5) = (%d, %d), want (2, 5)", goLine, goCol)
	}
	if _, _, found := sourceMap.TuiToGo(11, 0); !found {
		t.Error("no mapping for closing brace line of function")
	}
}

func TestGenerateVirtualGo_LetBindings(t *testing.T) {
	type tc struct {
		binding     *tuigen.LetBinding
		wantDecl    string
		wantTuiCol  int // expected gsx column of the mapped variable name (0-indexed)
		wantMapping bool
	}

	tests := map[string]tc{
		"short form maps name at position": {
			binding: &tuigen.LetBinding{
				Name:        "label",
				IsShortForm: true,
				Position:    tuigen.Position{Line: 4, Column: 2},
			},
			wantDecl:    "\tvar label any\n",
			wantTuiCol:  1,
			wantMapping: true,
		},
		"var form maps name after var keyword": {
			binding: &tuigen.LetBinding{
				Name:      "footer",
				IsVarForm: true,
				Position:  tuigen.Position{Line: 4, Column: 2},
			},
			wantDecl:    "\tvar footer any\n",
			wantTuiCol:  5, // column 2 (1-indexed) -> 1 (0-indexed) + len("var ")
			wantMapping: true,
		},
		"fallback form": {
			binding: &tuigen.LetBinding{
				Name:     "x",
				Position: tuigen.Position{Line: 4, Column: 2},
			},
			wantDecl:    "\tvar x any\n",
			wantTuiCol:  1,
			wantMapping: true,
		},
		"binding with element generates child expressions": {
			binding: &tuigen.LetBinding{
				Name:        "row",
				IsShortForm: true,
				Position:    tuigen.Position{Line: 4, Column: 2},
				Element: &tuigen.Element{
					Tag:      "span",
					Position: tuigen.Position{Line: 4, Column: 9},
					Children: []tuigen.Node{
						&tuigen.GoExpr{Code: "row.Title", Position: tuigen.Position{Line: 4, Column: 15}},
					},
				},
			},
			wantDecl:    "\tvar row any\n",
			wantTuiCol:  1,
			wantMapping: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			file := &tuigen.File{
				Package: "main",
				Components: []*tuigen.Component{
					{
						Name:       "App",
						ReturnType: "*element.Element",
						Position:   tuigen.Position{Line: 3, Column: 1},
						Body:       []tuigen.Node{tt.binding},
					},
				},
			}

			source, sourceMap := GenerateVirtualGo(file)

			if !strings.Contains(source, tt.wantDecl) {
				t.Errorf("generated source missing %q:\n%s", tt.wantDecl, source)
			}
			if tt.binding.Element != nil && !strings.Contains(source, "_ = row.Title") {
				t.Errorf("generated source missing element child expression:\n%s", source)
			}

			if tt.wantMapping {
				m := sourceMap.FindMappingForTuiPosition(tt.binding.Position.Line-1, tt.wantTuiCol)
				if m == nil {
					t.Fatalf("no mapping at gsx %d:%d\nmappings: %+v",
						tt.binding.Position.Line-1, tt.wantTuiCol, sourceMap.AllMappings())
				}
				if m.Length != len(tt.binding.Name) {
					t.Errorf("mapping length = %d, want %d", m.Length, len(tt.binding.Name))
				}
				if m.GoCol != len("\t")+len("var ") {
					t.Errorf("mapping GoCol = %d, want %d", m.GoCol, len("\t")+len("var "))
				}
			}
		})
	}
}

func TestGenerateVirtualGo_RawGoExpr(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "App",
				ReturnType: "*element.Element",
				Position:   tuigen.Position{Line: 3, Column: 1},
				Body: []tuigen.Node{
					&tuigen.RawGoExpr{Code: "  ", Position: tuigen.Position{Line: 4, Column: 2}},
					&tuigen.RawGoExpr{Code: "myRef", Position: tuigen.Position{Line: 5, Column: 8}},
				},
			},
		},
	}

	source, sourceMap := GenerateVirtualGo(file)

	if !strings.Contains(source, "\t_ = myRef\n") {
		t.Errorf("generated source missing raw expression:\n%s", source)
	}

	// The blank expression is skipped; only the real one is mapped.
	// RawGoExpr column convention: Position.Column (1-indexed at '{') is the
	// 0-indexed expression start.
	m := sourceMap.FindMappingForTuiPosition(4, 8)
	if m == nil {
		t.Fatalf("no mapping for raw expr, mappings: %+v", sourceMap.AllMappings())
	}
	if m.Length != len("myRef") || m.GoCol != len("\t")+len("_ = ") {
		t.Errorf("mapping = %+v, want Length 5 GoCol 5", m)
	}
	if found := sourceMap.FindMappingForTuiPosition(3, 2); found != nil {
		t.Errorf("blank raw expr should not be mapped, got %+v", found)
	}
}

func TestGenerateVirtualGo_GoCodeVariants(t *testing.T) {
	type tc struct {
		code        string
		wantEmitted string // empty means the code must not be emitted as GoCode
		wantMapping bool
	}

	tests := map[string]tc{
		"plain statement emitted with mapping": {
			code:        "x := compute()",
			wantEmitted: "\tx := compute()\n",
			wantMapping: true,
		},
		"empty code skipped": {
			code:        "   ",
			wantEmitted: "",
			wantMapping: false,
		},
		"state declaration skipped by generateGoCode": {
			// Emitted once by emitStateVarDeclarations, then skipped by
			// generateGoCode to avoid a duplicate.
			code:        "count := tui.NewState(0)",
			wantEmitted: "\tcount := tui.NewState(0)\n",
			wantMapping: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			file := &tuigen.File{
				Package: "main",
				Components: []*tuigen.Component{
					{
						Name:       "App",
						ReturnType: "*element.Element",
						Position:   tuigen.Position{Line: 3, Column: 1},
						Body: []tuigen.Node{
							&tuigen.GoCode{Code: tt.code, Position: tuigen.Position{Line: 4, Column: 2}},
						},
					},
				},
			}

			source, sourceMap := GenerateVirtualGo(file)

			if tt.wantEmitted != "" {
				if got := strings.Count(source, tt.wantEmitted); got != 1 {
					t.Errorf("code emitted %d times, want exactly 1:\n%s", got, source)
				}
			} else if strings.Contains(source, strings.TrimSpace(tt.code)) && strings.TrimSpace(tt.code) != "" {
				t.Errorf("empty code should not be emitted:\n%s", source)
			}

			if tt.wantMapping && sourceMap.Len() == 0 {
				t.Error("expected at least one mapping")
			}
			if !tt.wantMapping && sourceMap.Len() != 0 {
				t.Errorf("expected no mappings, got %+v", sourceMap.AllMappings())
			}
		})
	}
}

func TestGenerateVirtualGo_ControlFlowAndCalls(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "App",
				ReturnType: "*element.Element",
				Position:   tuigen.Position{Line: 3, Column: 1},
				Body: []tuigen.Node{
					// For loop with empty index: generated as "_".
					&tuigen.ForLoop{
						Value:    "item",
						Iterable: "items",
						Position: tuigen.Position{Line: 4, Column: 2},
						Body: []tuigen.Node{
							&tuigen.GoExpr{Code: "item", Position: tuigen.Position{Line: 5, Column: 9}},
						},
					},
					// If with else branch.
					&tuigen.IfStmt{
						Condition: "visible",
						Position:  tuigen.Position{Line: 7, Column: 2},
						Then: []tuigen.Node{
							&tuigen.GoExpr{Code: "shownLabel", Position: tuigen.Position{Line: 8, Column: 9}},
						},
						Else: []tuigen.Node{
							&tuigen.GoExpr{Code: "hiddenLabel", Position: tuigen.Position{Line: 10, Column: 9}},
						},
					},
					// Component call without args but with children.
					&tuigen.ComponentCall{
						Name:     "Header",
						Position: tuigen.Position{Line: 12, Column: 2},
						Children: []tuigen.Node{
							&tuigen.GoExpr{Code: "title", Position: tuigen.Position{Line: 13, Column: 9}},
						},
					},
					// Component call with args.
					&tuigen.ComponentCall{
						Name:     "Footer",
						Args:     `"hi", 2`,
						Position: tuigen.Position{Line: 15, Column: 2},
					},
					// Ignored node kinds.
					&tuigen.TextContent{Text: "plain text", Position: tuigen.Position{Line: 16, Column: 2}},
					&tuigen.ChildrenSlot{Position: tuigen.Position{Line: 17, Column: 2}},
				},
			},
		},
	}

	source, sourceMap := GenerateVirtualGo(file)

	wantLines := []string{
		"\tfor _, item := range items {\n",
		"\t\t_ = item\n",
		"\tif visible {\n",
		"\t\t_ = shownLabel\n",
		"\t} else {\n",
		"\t\t_ = hiddenLabel\n",
		"\t_ = Header()\n",
		"\t\t_ = title\n",
		"\t_ = Footer(\"hi\", 2)\n",
	}
	for _, want := range wantLines {
		if !strings.Contains(source, want) {
			t.Errorf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "plain text") {
		t.Errorf("text content leaked into generated source:\n%s", source)
	}

	// Footer args are mapped; Header (no args) is not.
	argsCol := 2 - 1 + len("@") + len("Footer") + 1 // parser column to args start
	if m := sourceMap.FindMappingForTuiPosition(14, argsCol); m == nil || m.Length != len(`"hi", 2`) {
		t.Errorf("no args mapping for Footer at 14:%d, mappings: %+v", argsCol, sourceMap.AllMappings())
	}
}

func TestGenerateVirtualGo_ElementAttributesAndReceiver(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:         "Render",
				Receiver:     "s *sidebar",
				ReceiverName: "s",
				ReceiverType: "*sidebar",
				ReturnType:   "*element.Element",
				Position:     tuigen.Position{Line: 3, Column: 1},
				Body: []tuigen.Node{
					&tuigen.Element{
						Tag:      "div",
						Position: tuigen.Position{Line: 4, Column: 2},
						Attributes: []*tuigen.Attribute{
							{
								Name:  "onActivate",
								Value: &tuigen.GoExpr{Code: "s.activate", Position: tuigen.Position{Line: 4, Column: 20}},
							},
							{
								Name:  "class",
								Value: &tuigen.StringLit{Value: "flex-col"},
							},
						},
						Children: []tuigen.Node{
							&tuigen.GoExpr{Code: "", Position: tuigen.Position{Line: 5, Column: 3}},
							&tuigen.GoExpr{Code: "s.title", Position: tuigen.Position{Line: 6, Column: 9}},
						},
					},
				},
			},
		},
	}

	source, sourceMap := GenerateVirtualGo(file)

	if !strings.Contains(source, "func (s *sidebar) Render() *element.Element {\n") {
		t.Errorf("missing method signature:\n%s", source)
	}
	if !strings.Contains(source, "\t_ = s.activate\n") {
		t.Errorf("missing attribute expression:\n%s", source)
	}
	if !strings.Contains(source, "\t_ = s.title\n") {
		t.Errorf("missing child expression:\n%s", source)
	}
	if strings.Contains(source, "flex-col") {
		t.Errorf("string literal attribute leaked into generated source:\n%s", source)
	}
	// Empty GoExpr child contributes no mapping; attribute + child do.
	if sourceMap.Len() != 2 {
		t.Errorf("sourceMap.Len() = %d, want 2: %+v", sourceMap.Len(), sourceMap.AllMappings())
	}
}

func TestGenerateVirtualGo_DefaultReturnType(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{Name: "App", Position: tuigen.Position{Line: 3, Column: 1}},
		},
	}

	source, _ := GenerateVirtualGo(file)

	if !strings.Contains(source, "func App() *element.Element {\n") {
		t.Errorf("expected default return type *element.Element:\n%s", source)
	}
}

func TestGenerateVirtualGo_NilNodesAreSkipped(t *testing.T) {
	// Typed nil nodes must not panic and must produce no output beyond the
	// component scaffold.
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "App",
				ReturnType: "*element.Element",
				Position:   tuigen.Position{Line: 3, Column: 1},
				Body: []tuigen.Node{
					(*tuigen.Element)(nil),
					(*tuigen.GoExpr)(nil),
					(*tuigen.ForLoop)(nil),
					(*tuigen.IfStmt)(nil),
					(*tuigen.LetBinding)(nil),
					(*tuigen.ComponentCall)(nil),
					(*tuigen.GoCode)(nil),
					(*tuigen.RawGoExpr)(nil),
				},
			},
		},
	}

	source, sourceMap := GenerateVirtualGo(file)

	want := "package main\n\nfunc App() *element.Element {\n\treturn nil\n}\n\n"
	if source != want {
		t.Errorf("source = %q, want %q", source, want)
	}
	if sourceMap.Len() != 0 {
		t.Errorf("expected no mappings, got %+v", sourceMap.AllMappings())
	}
}

func TestGenerateVirtualGo_RefsInNestedNodes(t *testing.T) {
	// Refs nested under if/else, let bindings, and component calls are all
	// hoisted into "_ =" statements by emitRefDeclarations.
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name:       "App",
				ReturnType: "*element.Element",
				Position:   tuigen.Position{Line: 3, Column: 1},
				Body: []tuigen.Node{
					&tuigen.IfStmt{
						Condition: "cond",
						Position:  tuigen.Position{Line: 4, Column: 2},
						Then: []tuigen.Node{
							&tuigen.Element{
								Tag:      "div",
								RefExpr:  &tuigen.GoExpr{Code: "thenRef", Position: tuigen.Position{Line: 5, Column: 12}},
								Position: tuigen.Position{Line: 5, Column: 3},
							},
						},
						Else: []tuigen.Node{
							&tuigen.Element{
								Tag:      "div",
								RefExpr:  &tuigen.GoExpr{Code: "elseRef", Position: tuigen.Position{Line: 7, Column: 12}},
								Position: tuigen.Position{Line: 7, Column: 3},
							},
						},
					},
					&tuigen.LetBinding{
						Name:        "bound",
						IsShortForm: true,
						Position:    tuigen.Position{Line: 9, Column: 2},
						Element: &tuigen.Element{
							Tag:      "span",
							RefExpr:  &tuigen.GoExpr{Code: "letRef", Position: tuigen.Position{Line: 9, Column: 22}},
							Position: tuigen.Position{Line: 9, Column: 11},
						},
					},
					&tuigen.ComponentCall{
						Name:     "Card",
						Position: tuigen.Position{Line: 11, Column: 2},
						Children: []tuigen.Node{
							&tuigen.Element{
								Tag:      "p",
								RefExpr:  &tuigen.GoExpr{Code: "callRef", Position: tuigen.Position{Line: 12, Column: 10}},
								Position: tuigen.Position{Line: 12, Column: 3},
							},
						},
					},
				},
			},
		},
	}

	source, _ := GenerateVirtualGo(file)

	for _, ref := range []string{"thenRef", "elseRef", "letRef", "callRef"} {
		if !strings.Contains(source, "\t_ = "+ref+"\n") {
			t.Errorf("generated source missing ref hoist for %s:\n%s", ref, source)
		}
	}
}
