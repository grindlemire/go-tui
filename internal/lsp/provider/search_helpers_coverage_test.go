package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func TestNewCompletionProvider_Constructor(t *testing.T) {
	cp := NewCompletionProvider(newStubIndex(), &nilGoplsProxy{}, &nilVirtualFiles{})
	if cp == nil {
		t.Fatal("expected non-nil provider")
	}

	// Tailwind completion inside class attribute exercises the provider
	// without gopls.
	content := `<div class="fle`
	doc := &Document{URI: "file:///test.gsx", Content: content, Version: 1}
	ctx := &CursorContext{
		Document:    doc,
		Position:    Position{Line: 0, Character: len(content)},
		Offset:      len(content),
		Word:        "fle",
		InClassAttr: true,
		Scope:       &Scope{},
	}

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Items) == 0 {
		t.Fatal("expected tailwind completion items")
	}
	foundFlex := false
	for _, item := range result.Items {
		if item.Label == "flex" {
			foundFlex = true
			break
		}
	}
	if !foundFlex {
		t.Errorf("expected flex completion among items")
	}
}

func TestFindLetBindingInNodes_Recursion(t *testing.T) {
	binding := &tuigen.LetBinding{Name: "header", IsShortForm: true, Position: tuigen.Position{Line: 5, Column: 2}}

	type tc struct {
		nodes []tuigen.Node
		found bool
	}

	tests := map[string]tc{
		"inside element": {
			nodes: []tuigen.Node{&tuigen.Element{Tag: "div", Children: []tuigen.Node{binding}}},
			found: true,
		},
		"inside for loop": {
			nodes: []tuigen.Node{&tuigen.ForLoop{Body: []tuigen.Node{binding}}},
			found: true,
		},
		"inside if then": {
			nodes: []tuigen.Node{&tuigen.IfStmt{Then: []tuigen.Node{binding}}},
			found: true,
		},
		"inside if else": {
			nodes: []tuigen.Node{&tuigen.IfStmt{Else: []tuigen.Node{binding}}},
			found: true,
		},
		"inside component call": {
			nodes: []tuigen.Node{&tuigen.ComponentCall{Name: "Card", Children: []tuigen.Node{binding}}},
			found: true,
		},
		"name mismatch": {
			nodes: []tuigen.Node{&tuigen.LetBinding{Name: "other", IsShortForm: true}},
			found: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := findLetBindingInNodes(tt.nodes, "header")
			if tt.found && got != binding {
				t.Errorf("expected to find binding, got %+v", got)
			}
			if !tt.found && got != nil {
				t.Errorf("expected nil, got %+v", got)
			}
		})
	}
}

func TestFindForLoopWithVariable_Recursion(t *testing.T) {
	loop := &tuigen.ForLoop{Index: "i", Value: "item", Position: tuigen.Position{Line: 5, Column: 2}}

	type tc struct {
		nodes   []tuigen.Node
		varName string
		found   bool
	}

	tests := map[string]tc{
		"index variable direct": {
			nodes:   []tuigen.Node{loop},
			varName: "i",
			found:   true,
		},
		"nested loop in body": {
			nodes:   []tuigen.Node{&tuigen.ForLoop{Index: "x", Value: "y", Body: []tuigen.Node{loop}}},
			varName: "item",
			found:   true,
		},
		"inside element": {
			nodes:   []tuigen.Node{&tuigen.Element{Tag: "div", Children: []tuigen.Node{loop}}},
			varName: "item",
			found:   true,
		},
		"inside if then": {
			nodes:   []tuigen.Node{&tuigen.IfStmt{Then: []tuigen.Node{loop}}},
			varName: "item",
			found:   true,
		},
		"inside if else": {
			nodes:   []tuigen.Node{&tuigen.IfStmt{Else: []tuigen.Node{loop}}},
			varName: "item",
			found:   true,
		},
		"inside component call": {
			nodes:   []tuigen.Node{&tuigen.ComponentCall{Name: "Card", Children: []tuigen.Node{loop}}},
			varName: "item",
			found:   true,
		},
		"no match": {
			nodes:   []tuigen.Node{loop},
			varName: "missing",
			found:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := findForLoopWithVariable(tt.nodes, tt.varName)
			if tt.found && got != loop {
				t.Errorf("expected to find loop, got %+v", got)
			}
			if !tt.found && got != nil {
				t.Errorf("expected nil, got %+v", got)
			}
		})
	}
}

func TestFindGoCodeWithVariable_Recursion(t *testing.T) {
	goCode := &tuigen.GoCode{Code: "msg := tui.NewRef()", Position: tuigen.Position{Line: 5, Column: 2}}

	type tc struct {
		nodes []tuigen.Node
		found bool
	}

	tests := map[string]tc{
		"inside element": {
			nodes: []tuigen.Node{&tuigen.Element{Tag: "div", Children: []tuigen.Node{goCode}}},
			found: true,
		},
		"inside for loop": {
			nodes: []tuigen.Node{&tuigen.ForLoop{Body: []tuigen.Node{goCode}}},
			found: true,
		},
		"inside if then": {
			nodes: []tuigen.Node{&tuigen.IfStmt{Then: []tuigen.Node{goCode}}},
			found: true,
		},
		"inside if else": {
			nodes: []tuigen.Node{&tuigen.IfStmt{Else: []tuigen.Node{goCode}}},
			found: true,
		},
		"inside component call": {
			nodes: []tuigen.Node{&tuigen.ComponentCall{Name: "Card", Children: []tuigen.Node{goCode}}},
			found: true,
		},
		"inside let binding element": {
			nodes: []tuigen.Node{&tuigen.LetBinding{Element: &tuigen.Element{
				Tag:      "div",
				Children: []tuigen.Node{goCode},
			}}},
			found: true,
		},
		"no declaration": {
			nodes: []tuigen.Node{&tuigen.GoCode{Code: "fmt.Println(msg)"}},
			found: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := findGoCodeWithVariable(tt.nodes, "msg")
			if tt.found && got != goCode {
				t.Errorf("expected to find GoCode, got %+v", got)
			}
			if !tt.found && got != nil {
				t.Errorf("expected nil, got %+v", got)
			}
		})
	}
}

func TestFindVariableUsagesInNodesExcluding_Recursion(t *testing.T) {
	uri := "file:///test.gsx"

	nodes := []tuigen.Node{
		// This GoCode contains the excluded declaration occurrence plus one usage.
		&tuigen.GoCode{Code: "msg := transform(msg)", Position: tuigen.Position{Line: 10, Column: 2}},
		&tuigen.RawGoExpr{Code: "msg", Position: tuigen.Position{Line: 11, Column: 7}},
		&tuigen.Element{
			Tag: "div",
			Attributes: []*tuigen.Attribute{
				{Name: "id", Value: &tuigen.GoExpr{Code: "msg", Position: tuigen.Position{Line: 12, Column: 10}}},
			},
			Children: []tuigen.Node{
				&tuigen.GoExpr{Code: "msg", Position: tuigen.Position{Line: 13, Column: 7}},
			},
		},
		&tuigen.ForLoop{
			Index:    "i",
			Value:    "v",
			Iterable: "msg",
			Position: tuigen.Position{Line: 14, Column: 2},
		},
		&tuigen.IfStmt{
			Condition: "msg != nil",
			Position:  tuigen.Position{Line: 15, Column: 2},
		},
		&tuigen.ComponentCall{
			Name:     "Card",
			Args:     "msg",
			Position: tuigen.Position{Line: 16, Column: 2},
		},
		&tuigen.LetBinding{Element: &tuigen.Element{Tag: "div", Children: []tuigen.Node{
			&tuigen.GoExpr{Code: "msg", Position: tuigen.Position{Line: 17, Column: 7}},
		}}},
	}

	var refs []Location
	// Exclude the declaration at line 9 (0-indexed), chars 1-4.
	findVariableUsagesInNodesExcluding(nodes, "msg", uri, 9, 1, 4, &refs)

	if len(refs) != 8 {
		t.Fatalf("expected 8 usages (decl excluded), got %d: %+v", len(refs), refs)
	}
	// The excluded declaration position must not appear.
	if hasLocation(refs, uri, 9, 1, 3) {
		t.Errorf("declaration at 9:1 should be excluded, got %+v", refs)
	}
	// The same-line argument usage survives: "msg := transform(msg)", msg at char 18.
	if !hasLocation(refs, uri, 9, 18, 3) {
		t.Errorf("missing same-line usage at 9:18, got %+v", refs)
	}
	if !hasLocation(refs, uri, 10, 7, 3) {
		t.Errorf("missing RawGoExpr usage at 10:7, got %+v", refs)
	}
	if !hasLocation(refs, uri, 11, 10, 3) {
		t.Errorf("missing attribute usage at 11:10, got %+v", refs)
	}
	if !hasLocation(refs, uri, 12, 7, 3) {
		t.Errorf("missing child expr usage at 12:7, got %+v", refs)
	}
	// ForLoop iterable: col 2-1+18 = 19.
	if !hasLocation(refs, uri, 13, 19, 3) {
		t.Errorf("missing iterable usage at 13:19, got %+v", refs)
	}
	// IfStmt condition: col 2-1+3 = 4.
	if !hasLocation(refs, uri, 14, 4, 3) {
		t.Errorf("missing condition usage at 14:4, got %+v", refs)
	}
	// ComponentCall args: col 2-1+6 = 7.
	if !hasLocation(refs, uri, 15, 7, 3) {
		t.Errorf("missing args usage at 15:7, got %+v", refs)
	}
	// LetBinding element child GoExpr: col 7-1+1 = 7.
	if !hasLocation(refs, uri, 16, 7, 3) {
		t.Errorf("missing let-binding usage at 16:7, got %+v", refs)
	}
}
