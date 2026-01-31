package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
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

func TestDefinition_RefAttr(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	src := `package test

templ Layout() {
	<div ref={header} class="p-1">title</div>
}
`
	doc := parseTestDoc(src)
	elem := doc.AST.Components[0].Body[0].(*tuigen.Element)

	ctx := makeCtx(doc, NodeKindRefAttr, "header")
	ctx.Node = elem

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for ref attr")
	}
	// Should point to ref={header}, not the element tag
	loc := result[0]
	line := loc.Range.Start.Line
	char := loc.Range.Start.Character
	endChar := loc.Range.End.Character
	if line != 3 {
		t.Errorf("expected line 3, got %d", line)
	}
	if endChar-char != len("ref={header}") {
		t.Errorf("expected range length %d, got %d", len("ref={header}"), endChar-char)
	}
}

func TestDefinition_RefAttr_Usage(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	src := `package test

templ Layout() {
	<div
		ref={content}
		class="p-1">title</div>
	<span>{content}</span>
}
`
	doc := parseTestDoc(src)
	elem := doc.AST.Components[0].Body[0].(*tuigen.Element)

	// Simulate cursor on "content" inside {content} — a Go expression usage
	ctx := makeCtx(doc, NodeKindGoExpr, "content")
	ctx.InGoExpr = true
	ctx.Scope.Component = doc.AST.Components[0]
	ctx.Scope.Refs = []tuigen.RefInfo{
		{Name: "content", Element: elem},
	}

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for ref usage in Go expr")
	}
	loc := result[0]
	// Should point to ref={content} on line 4, not the <div on line 3
	if loc.Range.Start.Line != 4 {
		t.Errorf("expected line 4, got %d", loc.Range.Start.Line)
	}
	if loc.Range.End.Character-loc.Range.Start.Character != len("ref={content}") {
		t.Errorf("expected range length %d, got %d", len("ref={content}"), loc.Range.End.Character-loc.Range.Start.Character)
	}
}

func TestDefinition_RefAttr_WithDeclaration(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	src := `package test

templ StreamApp() {
	content := tui.NewRef()
	<div ref={content} class="p-1">title</div>
}
`
	doc := parseTestDoc(src)
	elem := doc.AST.Components[0].Body[1].(*tuigen.Element)

	ctx := makeCtx(doc, NodeKindRefAttr, "content")
	ctx.Node = elem
	ctx.Scope.Component = doc.AST.Components[0]

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for ref attr with declaration")
	}
	loc := result[0]
	// Should point to content := tui.NewRef() on line 3 (0-indexed), not ref={content}
	if loc.Range.Start.Line != 3 {
		t.Errorf("expected line 3 (declaration), got %d", loc.Range.Start.Line)
	}
	if loc.Range.End.Character-loc.Range.Start.Character != len("content") {
		t.Errorf("expected range length %d, got %d", len("content"), loc.Range.End.Character-loc.Range.Start.Character)
	}
}

func TestDefinition_RefAttr_Multiline(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	src := `package test

templ Layout() {
	<div
		ref={header}
		class="p-1">title</div>
}
`
	doc := parseTestDoc(src)
	elem := doc.AST.Components[0].Body[0].(*tuigen.Element)

	ctx := makeCtx(doc, NodeKindRefAttr, "header")
	ctx.Node = elem

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for multiline ref attr")
	}
	loc := result[0]
	// ref={header} is on line 4 (0-indexed), not line 3 (the <div line)
	if loc.Range.Start.Line != 4 {
		t.Errorf("expected line 4, got %d", loc.Range.Start.Line)
	}
	if loc.Range.End.Character-loc.Range.Start.Character != len("ref={header}") {
		t.Errorf("expected range length %d, got %d", len("ref={header}"), loc.Range.End.Character-loc.Range.Start.Character)
	}
}

func TestDefinition_StateDecl(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}
`
	doc := parseTestDoc(src)

	ctx := makeCtx(doc, NodeKindStateDecl, "count")
	ctx.Scope.Component = doc.AST.Components[0]
	ctx.Scope.StateVars = []tuigen.StateVar{
		{Name: "count", Type: "int", InitExpr: "0", Position: tuigen.Position{Line: 4, Column: 2}},
	}

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for state decl")
	}
	if result[0].Range.Start.Line != 3 {
		t.Errorf("expected line 3, got %d", result[0].Range.Start.Line)
	}
}

func TestDefinition_StateAccess(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}
`
	doc := parseTestDoc(src)

	ctx := makeCtx(doc, NodeKindStateAccess, "count")
	ctx.Scope.Component = doc.AST.Components[0]
	ctx.Scope.StateVars = []tuigen.StateVar{
		{Name: "count", Type: "int", InitExpr: "0", Position: tuigen.Position{Line: 4, Column: 2}},
	}

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected definition location for state access")
	}
	if result[0].Range.Start.Line != 3 {
		t.Errorf("expected line 3, got %d", result[0].Range.Start.Line)
	}
}

func TestDefinition_EventHandler(t *testing.T) {
	index := newStubIndex()
	dp := newTestDefinitionProvider(index)

	doc := parseTestDoc("package test")

	ctx := makeCtx(doc, NodeKindEventHandler, "handleClick")
	ctx.InGoExpr = true

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Without a real gopls proxy, this should return nil
	if result != nil && len(result) > 0 {
		t.Errorf("expected nil result without gopls, got %v", result)
	}
}

func TestFindVarDeclPosition_WordBoundary(t *testing.T) {
	type tc struct {
		code    string
		varName string
		want    int
	}

	tests := map[string]tc{
		"simple decl": {
			code: "count := 1", varName: "count", want: 0,
		},
		"no substring match": {
			code: "accountCount := 1", varName: "count", want: -1,
		},
		"multi var decl": {
			code: "a, count := 1, 2", varName: "count", want: 3,
		},
		"var keyword": {
			code: "var count = 1", varName: "count", want: 4,
		},
		"var keyword no substring": {
			code: "var accountCount = 1", varName: "count", want: -1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := findVarDeclPosition(tt.code, tt.varName)
			if got != tt.want {
				t.Errorf("findVarDeclPosition(%q, %q) = %d, want %d", tt.code, tt.varName, got, tt.want)
			}
		})
	}
}

func TestContainsVarDecl(t *testing.T) {
	type tc struct {
		code    string
		varName string
		want    bool
	}

	tests := map[string]tc{
		"simple match": {
			code: "x := 1", varName: "x", want: true,
		},
		"no match": {
			code: "y := 1", varName: "x", want: false,
		},
		"multi var": {
			code: "a, b := 1, 2", varName: "b", want: true,
		},
		"var keyword": {
			code: "var x = 1", varName: "x", want: true,
		},
		"no assignment": {
			code: "fmt.Println(x)", varName: "x", want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := containsVarDecl(tt.code, tt.varName)
			if got != tt.want {
				t.Errorf("containsVarDecl(%q, %q) = %v, want %v", tt.code, tt.varName, got, tt.want)
			}
		})
	}
}
