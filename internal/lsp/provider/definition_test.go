package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/lsp/gopls"

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
	header := <div>title</div>
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
		for _, item := range items {
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
	if len(result) > 0 {
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

// A gopls hit inside a generated _gsx.go file is translated to the templ or
// func declaration in the sibling .gsx via the workspace AST.
func TestDefinition_LocateInWorkspaceGsx(t *testing.T) {
	type tc struct {
		uri      string
		name     string
		wantLine int // -1 means no location
		wantChar int // 0-indexed start of the name; the range must span exactly the name
	}

	const widgets = "file:///w/widgets.gsx"
	ws := &stubWorkspaceAST{asts: map[string]*tuigen.File{
		widgets: {
			Components: []*tuigen.Component{
				{Name: "Header", Position: tuigen.Position{Line: 5, Column: 1}, NamePos: tuigen.Position{Line: 5, Column: 7}},
			},
			Funcs: []*tuigen.GoFunc{
				{Code: "func NewCard(label string) *Card {\n\treturn &Card{}\n}", Position: tuigen.Position{Line: 12, Column: 1}},
				{Code: "func (c *Card) helper() {}", Position: tuigen.Position{Line: 20, Column: 1}},
				{Code: "func NewList[T any](items []T) *List[T] {\n\treturn nil\n}", Position: tuigen.Position{Line: 30, Column: 1}},
			},
		},
	}}

	tests := map[string]tc{
		"function templ":        {uri: widgets, name: "Header", wantLine: 4, wantChar: 6},
		"struct factory func":   {uri: widgets, name: "NewCard", wantLine: 11, wantChar: 5},
		"generic factory func":  {uri: widgets, name: "NewList", wantLine: 29, wantChar: 5},
		"method is not a match": {uri: widgets, name: "helper", wantLine: -1},
		"unknown name":          {uri: widgets, name: "Nope", wantLine: -1},
		"file not in workspace": {uri: "file:///w/other.gsx", name: "Header", wantLine: -1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dp := newTestDefinitionProvider(newStubIndex())
			dp.workspace = ws

			loc := dp.locateInWorkspaceGsx(tt.uri, tt.name)
			if tt.wantLine < 0 {
				if loc != nil {
					t.Fatalf("expected no location, got %+v", *loc)
				}
				return
			}
			if loc == nil {
				t.Fatal("expected a location, got nil")
			}
			if loc.URI != tt.uri || loc.Range.Start.Line != tt.wantLine {
				t.Errorf("got %s:%d, want %s:%d", loc.URI, loc.Range.Start.Line, tt.uri, tt.wantLine)
			}
			if loc.Range.Start.Character != tt.wantChar || loc.Range.End.Character != tt.wantChar+len(tt.name) {
				t.Errorf("range = %d..%d, want %d..%d (the name itself)",
					loc.Range.Start.Character, loc.Range.End.Character, tt.wantChar, tt.wantChar+len(tt.name))
			}
		})
	}
}

// gopls locations inside generated _gsx.go files map to the sibling .gsx when
// that file is in the workspace, and pass through untouched when it is not
// (a dependency in the module cache).
func TestDefinition_TranslateGoplsLocations(t *testing.T) {
	type tc struct {
		goURI    string
		word     string
		wantURI  string
		wantLine int
		wantNone bool
	}

	ws := &stubWorkspaceAST{asts: map[string]*tuigen.File{
		"file:///w/widgets.gsx": {
			Components: []*tuigen.Component{{Name: "Header", Position: tuigen.Position{Line: 5, Column: 1}}},
		},
	}}

	tests := map[string]tc{
		"workspace generated file maps to its .gsx": {
			goURI: "file:///w/widgets_gsx.go", word: "Header",
			wantURI: "file:///w/widgets.gsx", wantLine: 4,
		},
		"workspace generated file with unknown name is skipped": {
			goURI: "file:///w/widgets_gsx.go", word: "Nope",
			wantNone: true,
		},
		"module cache generated file passes through": {
			goURI: "file:///go/pkg/mod/example.com/dep@v1.0.0/dep_gsx.go", word: "Header",
			wantURI: "file:///go/pkg/mod/example.com/dep@v1.0.0/dep_gsx.go", wantLine: 57,
		},
		"plain go file passes through": {
			goURI: "file:///usr/local/go/src/fmt/print.go", word: "Println",
			wantURI: "file:///usr/local/go/src/fmt/print.go", wantLine: 57,
		},
		"open document maps to its .gsx even though the workspace cache dropped it": {
			goURI: "file:///w/open_gsx.go", word: "Panel",
			wantURI: "file:///w/open.gsx", wantLine: 2,
		},
	}

	// Opening a file moves its AST from the workspace cache to the document
	// manager, so the provider must consult both.
	openDocs := &stubDocAccessor{docs: []*Document{{
		URI: "file:///w/open.gsx",
		AST: &tuigen.File{Components: []*tuigen.Component{{Name: "Panel", Position: tuigen.Position{Line: 3, Column: 1}}}},
	}}}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dp := newTestDefinitionProvider(newStubIndex())
			dp.workspace = ws
			dp.docs = openDocs
			ctx := makeCtx(parseTestDoc("package test"), NodeKindComponentCall, tt.word)

			got := dp.translateGoplsLocations(ctx, []gopls.Location{{
				URI:   tt.goURI,
				Range: gopls.Range{Start: gopls.Position{Line: 57, Character: 5}, End: gopls.Position{Line: 57, Character: 11}},
			}})

			if tt.wantNone {
				if len(got) != 0 {
					t.Fatalf("expected no locations, got %+v", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("expected 1 location, got %d: %+v", len(got), got)
			}
			if got[0].URI != tt.wantURI || got[0].Range.Start.Line != tt.wantLine {
				t.Errorf("got %s:%d, want %s:%d", got[0].URI, got[0].Range.Start.Line, tt.wantURI, tt.wantLine)
			}
		})
	}
}

// A local func that happens to share the last segment of a qualified call
// must not shadow the call's real target.
func TestDefinition_QualifiedCallNotShadowedByLocalFunc(t *testing.T) {
	index := newStubIndex()
	index.functions["Header"] = &FuncInfo{
		Name:     "Header",
		Location: Location{URI: "file:///w/local.gsx", Range: Range{Start: Position{Line: 9}}},
	}
	dp := newTestDefinitionProvider(index)

	ctx := makeCtx(parseTestDoc("package test"), NodeKindComponentCall, "Header")
	ctx.Node = &tuigen.ComponentCall{Name: "widgets.Header"}

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, loc := range result {
		if loc.URI == "file:///w/local.gsx" {
			t.Fatalf("qualified call resolved to the local func: %+v", loc)
		}
	}
}
