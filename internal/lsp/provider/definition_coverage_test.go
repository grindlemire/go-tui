package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func TestNewDefinitionProvider_Constructor(t *testing.T) {
	index := newStubIndex()
	index.functions["helper"] = &FuncInfo{
		Name:     "helper",
		Location: Location{URI: "file:///lib.gsx", Range: Range{Start: Position{Line: 9, Character: 0}}},
	}
	dp := NewDefinitionProvider(index, &nilGoplsProxy{}, &nilVirtualFiles{}, &stubDocAccessor{})
	if dp == nil {
		t.Fatal("expected non-nil provider")
	}

	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindUnknown, "helper")
	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].URI != "file:///lib.gsx" || result[0].Range.Start.Line != 9 {
		t.Errorf("expected function location at file:///lib.gsx:9, got %+v", result)
	}
}

func TestDefinition_EmptyWordInGoExpr(t *testing.T) {
	dp := newTestDefinitionProvider(newStubIndex())
	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindGoExpr, "")
	ctx.InGoExpr = true

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil without gopls, got %+v", result)
	}
}

func TestDefinition_ComponentCallVariants(t *testing.T) {
	src := `package test

templ Page() {
	@Sidebar()
}
`
	doc := parseTestDoc(src)
	call := doc.AST.Components[0].Body[0].(*tuigen.ComponentCall)

	t.Run("struct mount resolves via function index", func(t *testing.T) {
		index := newStubIndex()
		index.functions["Sidebar"] = &FuncInfo{
			Name:     "Sidebar",
			Location: Location{URI: "file:///side.gsx", Range: Range{Start: Position{Line: 12, Character: 0}}},
		}
		dp := newTestDefinitionProvider(index)

		ctx := makeCtx(doc, NodeKindComponentCall, "@Sidebar")
		ctx.InGoExpr = true // skip the early function-index shortcut so the call handler runs
		ctx.Node = call

		result, err := dp.Definition(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].URI != "file:///side.gsx" || result[0].Range.Start.Line != 12 {
			t.Errorf("expected constructor location, got %+v", result)
		}
	})

	t.Run("unknown call returns nil", func(t *testing.T) {
		dp := newTestDefinitionProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindComponentCall, "@Sidebar")
		ctx.Node = call

		result, err := dp.Definition(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})

	t.Run("wrong node type returns nil", func(t *testing.T) {
		dp := newTestDefinitionProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindComponentCall, "@Sidebar")
		ctx.Node = &tuigen.Element{Tag: "div"}

		result, err := dp.Definition(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})
}

func TestDefinition_RefAttrDottedFieldName(t *testing.T) {
	src := `package test

type chat struct {
	textareaRef *tui.Ref
}

templ (c *chat) Render() {
	<div ref={c.textareaRef} class="p-1">x</div>
}
`
	doc := parseTestDoc(src)
	comp := doc.AST.Components[0]
	elem := comp.Body[0].(*tuigen.Element)
	if elem.RefExpr == nil {
		t.Fatal("fixture did not parse ref expression")
	}

	dp := newTestDefinitionProvider(newStubIndex())
	ctx := makeCtx(doc, NodeKindRefAttr, "textareaRef")
	ctx.Node = elem
	ctx.Scope.Component = comp

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 location, got %d", len(result))
	}
	// Struct field "textareaRef" is on source line 4 (0-indexed line 3),
	// starting after the leading tab (column 1).
	loc := result[0]
	if loc.Range.Start.Line != 3 {
		t.Errorf("expected line 3, got %d", loc.Range.Start.Line)
	}
	if loc.Range.Start.Character != 1 {
		t.Errorf("expected character 1, got %d", loc.Range.Start.Character)
	}
	if loc.Range.End.Character-loc.Range.Start.Character != len("textareaRef") {
		t.Errorf("expected range width %d, got %d", len("textareaRef"), loc.Range.End.Character-loc.Range.Start.Character)
	}
}

func TestDefinition_RefAttrFallbackToElementPosition(t *testing.T) {
	// When the ref attribute text cannot be found in the content (e.g. stale
	// content), the location falls back to the element tag position.
	elem := &tuigen.Element{
		Tag:      "div",
		RefExpr:  &tuigen.GoExpr{Code: "panel"},
		Position: tuigen.Position{Line: 1, Column: 1},
	}
	doc := &Document{URI: "file:///test.gsx", Content: "nothing here", Version: 1}

	dp := newTestDefinitionProvider(newStubIndex())
	ctx := makeCtx(doc, NodeKindRefAttr, "panel")
	ctx.Node = elem

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 location, got %d", len(result))
	}
	loc := result[0]
	if loc.Range.Start.Line != 0 || loc.Range.Start.Character != 0 {
		t.Errorf("expected fallback to element position 0:0, got %d:%d", loc.Range.Start.Line, loc.Range.Start.Character)
	}
	if loc.Range.End.Character != len("ref={panel}") {
		t.Errorf("expected end character %d, got %d", len("ref={panel}"), loc.Range.End.Character)
	}
}

func TestDefinition_ParameterNodeKind(t *testing.T) {
	src := `package test

func format(msg string) string {
	return msg
}

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)
	comp := doc.AST.Components[0]
	fn := doc.AST.Funcs[0]

	t.Run("function parameter", func(t *testing.T) {
		index := newCovIndex()
		index.funcParams["format.msg"] = &FuncParamInfo{
			Name:     "msg",
			Type:     "string",
			FuncName: "format",
			Location: Location{URI: doc.URI, Range: Range{Start: Position{Line: 2, Character: 12}}},
		}
		dp := newTestDefinitionProvider(index)

		ctx := makeCtx(doc, NodeKindParameter, "msg")
		ctx.Scope.Function = fn

		result, err := dp.Definition(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].Range.Start.Line != 2 || result[0].Range.Start.Character != 12 {
			t.Errorf("expected function param location at 2:12, got %+v", result)
		}
	})

	t.Run("component parameter", func(t *testing.T) {
		index := newStubIndex()
		index.params["Header.title"] = &ParamInfo{
			Name:          "title",
			Type:          "string",
			ComponentName: "Header",
			Location:      Location{URI: doc.URI, Range: Range{Start: Position{Line: 6, Character: 13}}},
		}
		dp := newTestDefinitionProvider(index)

		ctx := makeCtx(doc, NodeKindParameter, "title")
		ctx.Scope.Component = comp

		result, err := dp.Definition(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].Range.Start.Line != 6 || result[0].Range.Start.Character != 13 {
			t.Errorf("expected component param location at 6:13, got %+v", result)
		}
	})

	t.Run("unresolved parameter returns nil", func(t *testing.T) {
		dp := newTestDefinitionProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindParameter, "missing")

		result, err := dp.Definition(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})
}

func TestDefinition_ImportPath(t *testing.T) {
	type tc struct {
		importPath string
	}

	tests := map[string]tc{
		"empty import path": {importPath: ""},
		"no gopls proxy":    {importPath: "fmt"},
	}

	dp := newTestDefinitionProvider(newStubIndex())
	doc := parseTestDoc("package test\n\nimport \"fmt\"\n")

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := makeCtx(doc, NodeKindImportPath, "fmt")
			ctx.ImportPath = tt.importPath

			result, err := dp.Definition(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != nil {
				t.Errorf("expected nil without gopls, got %+v", result)
			}
		})
	}
}

func TestDefinition_ComponentFromAST(t *testing.T) {
	src := `package test

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)
	dp := newTestDefinitionProvider(newStubIndex())

	ctx := makeCtx(doc, NodeKindUnknown, "Header")
	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 location, got %d", len(result))
	}
	loc := result[0]
	if loc.Range.Start.Line != 2 || loc.Range.Start.Character != 0 {
		t.Errorf("expected templ position 2:0, got %d:%d", loc.Range.Start.Line, loc.Range.Start.Character)
	}
	if loc.Range.End.Character != len("templ Header") {
		t.Errorf("expected end character %d, got %d", len("templ Header"), loc.Range.End.Character)
	}
}

func TestDefinition_FuncFromAST(t *testing.T) {
	src := `package test

func (c *chat) updateHeight(h int) int {
	return h
}

templ App() {
	<div>x</div>
}
`
	doc := parseTestDoc(src)
	dp := newTestDefinitionProvider(newStubIndex())

	ctx := makeCtx(doc, NodeKindUnknown, "updateHeight")
	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 location, got %d", len(result))
	}
	loc := result[0]
	// "updateHeight" starts at column 15 (0-indexed) on line 2.
	if loc.Range.Start.Line != 2 || loc.Range.Start.Character != 15 {
		t.Errorf("expected method name at 2:15, got %d:%d", loc.Range.Start.Line, loc.Range.Start.Character)
	}
	if loc.Range.End.Character-loc.Range.Start.Character != len("updateHeight") {
		t.Errorf("expected range width %d, got %d", len("updateHeight"), loc.Range.End.Character-loc.Range.Start.Character)
	}
}

func TestDefinition_GoDeclFromAST(t *testing.T) {
	src := `package test

type chat struct {
	showSettings bool
}

var limit = 10

templ App() {
	<div>x</div>
}
`
	doc := parseTestDoc(src)

	type tc struct {
		word     string
		nodeKind NodeKind
		wantLine int
		wantChar int
	}

	tests := map[string]tc{
		"type name via GoDecl kind": {
			word:     "chat",
			nodeKind: NodeKindGoDecl,
			wantLine: 2,
			wantChar: 5,
		},
		"struct field via fallback": {
			word:     "showSettings",
			nodeKind: NodeKindUnknown,
			wantLine: 3,
			wantChar: 1,
		},
		"var declaration": {
			word:     "limit",
			nodeKind: NodeKindUnknown,
			wantLine: 6,
			wantChar: 4,
		},
	}

	dp := newTestDefinitionProvider(newStubIndex())

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := makeCtx(doc, tt.nodeKind, tt.word)

			result, err := dp.Definition(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != 1 {
				t.Fatalf("expected 1 location, got %d", len(result))
			}
			loc := result[0]
			if loc.Range.Start.Line != tt.wantLine || loc.Range.Start.Character != tt.wantChar {
				t.Errorf("expected %d:%d, got %d:%d", tt.wantLine, tt.wantChar, loc.Range.Start.Line, loc.Range.Start.Character)
			}
			if loc.Range.End.Character-loc.Range.Start.Character != len(tt.word) {
				t.Errorf("expected range width %d, got %d", len(tt.word), loc.Range.End.Character-loc.Range.Start.Character)
			}
		})
	}
}

func TestDefinition_StateVarNoMatch(t *testing.T) {
	dp := newTestDefinitionProvider(newStubIndex())
	doc := parseTestDoc("package test")

	ctx := makeCtx(doc, NodeKindStateAccess, "count")
	ctx.Scope.StateVars = []tuigen.StateVar{
		{Name: "other", Type: "int", Position: tuigen.Position{Line: 4, Column: 2}},
	}

	result, err := dp.Definition(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for unmatched state var, got %+v", result)
	}
}

func TestOffsetToLineChar(t *testing.T) {
	type tc struct {
		content  string
		offset   int
		wantLine int
		wantChar int
	}

	tests := map[string]tc{
		"start of content": {
			content: "abc\ndef", offset: 0, wantLine: 0, wantChar: 0,
		},
		"middle of first line": {
			content: "abc\ndef", offset: 2, wantLine: 0, wantChar: 2,
		},
		"start of second line": {
			content: "abc\ndef", offset: 4, wantLine: 1, wantChar: 0,
		},
		"middle of second line": {
			content: "abc\ndef", offset: 6, wantLine: 1, wantChar: 2,
		},
		"offset beyond content": {
			content: "ab", offset: 10, wantLine: 0, wantChar: 10,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			line, char := offsetToLineChar(tt.content, tt.offset)
			if line != tt.wantLine || char != tt.wantChar {
				t.Errorf("offsetToLineChar(%q, %d) = (%d, %d), want (%d, %d)",
					tt.content, tt.offset, line, char, tt.wantLine, tt.wantChar)
			}
		})
	}
}

func TestParseFuncName(t *testing.T) {
	type tc struct {
		code string
		want string
	}

	tests := map[string]tc{
		"plain function": {
			code: "func helper(s string) string { return s }",
			want: "helper",
		},
		"method with receiver": {
			code: "func (c *chat) updateHeight(h int) int { return h }",
			want: "updateHeight",
		},
		"no func keyword": {
			code: "var x = 1",
			want: "",
		},
		"receiver without close paren": {
			code: "func (c *chat",
			want: "",
		},
		"no parameter list": {
			code: "func broken",
			want: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := parseFuncName(tt.code)
			if got != tt.want {
				t.Errorf("parseFuncName(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
