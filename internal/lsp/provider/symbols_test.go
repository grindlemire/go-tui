package provider

import (
	"testing"
)

func TestDocumentSymbols_Components(t *testing.T) {
	src := `package test

templ Header(title string) {
	<div>{title}</div>
}

templ Footer() {
	<span>footer</span>
}
`
	doc := parseTestDoc(src)
	sp := NewDocumentSymbolProvider()

	result, err := sp.DocumentSymbols(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 symbols (Header, Footer), got %d", len(result))
	}

	if result[0].Name != "Header" {
		t.Errorf("expected first symbol name 'Header', got %q", result[0].Name)
	}
	if result[0].Kind != SymbolKindFunction {
		t.Errorf("expected kind Function, got %d", result[0].Kind)
	}
	if result[0].Detail != "(title string)" {
		t.Errorf("expected detail '(title string)', got %q", result[0].Detail)
	}

	if result[1].Name != "Footer" {
		t.Errorf("expected second symbol name 'Footer', got %q", result[1].Name)
	}
	if result[1].Detail != "()" {
		t.Errorf("expected detail '()', got %q", result[1].Detail)
	}
}

func TestDocumentSymbols_Functions(t *testing.T) {
	src := `package test

func helper(s string) string {
	return s
}
`
	doc := parseTestDoc(src)
	sp := NewDocumentSymbolProvider()

	result, err := sp.DocumentSymbols(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 symbol (helper), got %d", len(result))
	}

	if result[0].Name != "helper" {
		t.Errorf("expected symbol name 'helper', got %q", result[0].Name)
	}
	if result[0].Kind != SymbolKindFunction {
		t.Errorf("expected kind Function, got %d", result[0].Kind)
	}
	if result[0].Detail != "func" {
		t.Errorf("expected detail 'func', got %q", result[0].Detail)
	}
}

func TestDocumentSymbols_LetBindingChildren(t *testing.T) {
	src := `package test

templ Page() {
	@let header = <div>Header</div>
	{header}
}
`
	doc := parseTestDoc(src)
	sp := NewDocumentSymbolProvider()

	result, err := sp.DocumentSymbols(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 top-level symbol (Page), got %d", len(result))
	}

	page := result[0]
	if page.Name != "Page" {
		t.Errorf("expected 'Page', got %q", page.Name)
	}

	// Should have child for let binding
	if len(page.Children) < 1 {
		t.Errorf("expected at least 1 child symbol (let binding), got %d", len(page.Children))
	} else {
		found := false
		for _, child := range page.Children {
			if child.Name == "header" && child.Kind == SymbolKindVariable {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected child symbol 'header' with kind Variable")
		}
	}
}

func TestDocumentSymbols_NilAST(t *testing.T) {
	doc := &Document{
		URI:     "file:///test.gsx",
		Content: "invalid content",
	}
	sp := NewDocumentSymbolProvider()

	result, err := sp.DocumentSymbols(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 symbols for nil AST, got %d", len(result))
	}
}

func TestWorkspaceSymbols_NameMatching(t *testing.T) {
	index := newStubIndex()
	index.components["Header"] = &ComponentInfo{
		Name: "Header",
		Location: Location{
			URI: "file:///test.gsx",
			Range: Range{
				Start: Position{Line: 2, Character: 0},
				End:   Position{Line: 2, Character: 20},
			},
		},
	}
	index.components["Footer"] = &ComponentInfo{
		Name: "Footer",
		Location: Location{
			URI: "file:///test.gsx",
			Range: Range{
				Start: Position{Line: 10, Character: 0},
				End:   Position{Line: 10, Character: 20},
			},
		},
	}

	wp := NewWorkspaceSymbolProvider(index)

	// Empty query — returns all
	result, err := wp.WorkspaceSymbols("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 2 {
		t.Errorf("expected at least 2 symbols for empty query, got %d", len(result))
	}

	// Specific query — case insensitive
	result, err = wp.WorkspaceSymbols("head")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 symbol for 'head' query, got %d", len(result))
	}
	if len(result) > 0 && result[0].Name != "Header" {
		t.Errorf("expected 'Header', got %q", result[0].Name)
	}
}

func TestWorkspaceSymbols_FunctionSearch(t *testing.T) {
	index := newStubIndex()
	index.functions["formatLabel"] = &FuncInfo{
		Name:      "formatLabel",
		Signature: "func formatLabel(s string) string",
		Location: Location{
			URI: "file:///test.gsx",
			Range: Range{
				Start: Position{Line: 5, Character: 0},
				End:   Position{Line: 5, Character: 30},
			},
		},
	}

	wp := NewWorkspaceSymbolProvider(index)

	result, err := wp.WorkspaceSymbols("format")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 symbol for 'format' query, got %d", len(result))
	}
	if len(result) > 0 && result[0].Name != "formatLabel" {
		t.Errorf("expected 'formatLabel', got %q", result[0].Name)
	}
}

func TestWorkspaceSymbols_NoMatch(t *testing.T) {
	index := newStubIndex()
	index.components["Header"] = &ComponentInfo{Name: "Header"}

	wp := NewWorkspaceSymbolProvider(index)

	result, err := wp.WorkspaceSymbols("xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 symbols for 'xyz' query, got %d", len(result))
	}
}
