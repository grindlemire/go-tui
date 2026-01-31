package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/pkg/tuigen"
)

func newTestDefinitionProvider(index ComponentIndex) *definitionProvider {
	return &definitionProvider{
		index:        index,
		goplsProxy:   &nilGoplsProxy{},
		virtualFiles: &nilVirtualFiles{},
		docs:         &stubDocAccessor{},
	}
}

func TestDefinition_ComponentCall(t *testing.T) {
	index := newStubIndex()
	index.components["Header"] = &ComponentInfo{
		Name: "Header",
		Location: Location{
			URI: "file:///test.gsx",
			Range: Range{
				Start: Position{Line: 2, Character: 0},
				End:   Position{Line: 2, Character: 17},
			},
		},
	}

	dp := newTestDefinitionProvider(index)

	src := `package test

templ Page() {
	@Header("title")
}

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)
	call := doc.AST.Components[0].Body[0].(*tuigen.ComponentCall)

	ctx := makeCtx(doc, NodeKindComponentCall, "@Header")
	ctx.Node = call

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location")
	}
	if result[0].URI != "file:///test.gsx" {
		t.Errorf("expected URI file:///test.gsx, got %s", result[0].URI)
	}
}

func TestDefinition_FunctionByWord(t *testing.T) {
	index := newStubIndex()
	index.functions["helper"] = &FuncInfo{
		Name:      "helper",
		Signature: "func helper(s string) string",
		Location: Location{
			URI: "file:///test.gsx",
			Range: Range{
				Start: Position{Line: 5, Character: 0},
				End:   Position{Line: 5, Character: 15},
			},
		},
	}

	dp := newTestDefinitionProvider(index)

	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindUnknown, "helper")

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for function")
	}
	if result[0].Range.Start.Line != 5 {
		t.Errorf("expected line 5, got %d", result[0].Range.Start.Line)
	}
}

func TestDefinition_Parameter(t *testing.T) {
	index := newStubIndex()
	index.params["Header.title"] = &ParamInfo{
		Name:          "title",
		Type:          "string",
		ComponentName: "Header",
		Location: Location{
			URI: "file:///test.gsx",
			Range: Range{
				Start: Position{Line: 2, Character: 13},
				End:   Position{Line: 2, Character: 18},
			},
		},
	}

	dp := newTestDefinitionProvider(index)

	src := `package test

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)

	ctx := makeCtx(doc, NodeKindUnknown, "title")
	ctx.Scope.Component = doc.AST.Components[0]

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for parameter")
	}
	if result[0].Range.Start.Line != 2 {
		t.Errorf("expected line 2, got %d", result[0].Range.Start.Line)
	}
}

func TestDefinition_LetBinding(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	src := `package test

templ Example() {
	@let header = <div>title</div>
	{header}
}
`
	doc := parseTestDoc(src)

	ctx := makeCtx(doc, NodeKindUnknown, "header")
	ctx.Scope.Component = doc.AST.Components[0]

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for let binding")
	}
}

func TestDefinition_ForLoopVariable(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	src := `package test

templ List(items []string) {
	<div>
		@for _, item := range items {
			<span>{item}</span>
		}
	</div>
}
`
	doc := parseTestDoc(src)

	ctx := makeCtx(doc, NodeKindUnknown, "item")
	ctx.Scope.Component = doc.AST.Components[0]

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for loop variable")
	}
}

func TestDefinition_NilDocument(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	doc := &Document{URI: "file:///test.gsx", Content: "", Version: 1}
	ctx := makeCtx(doc, NodeKindUnknown, "")

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty word, got %v", result)
	}
}
