package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// hasLocation checks for a location with the given URI and exact start position and width.
func hasLocation(locs []Location, uri string, line, char, width int) bool {
	for _, l := range locs {
		if l.URI == uri && l.Range.Start.Line == line && l.Range.Start.Character == char &&
			l.Range.End.Character-l.Range.Start.Character == width {
			return true
		}
	}
	return false
}

func TestNewReferencesProvider_Constructor(t *testing.T) {
	doc := parseTestDoc("package test")
	rp := NewReferencesProvider(newStubIndex(), &stubDocAccessor{}, &stubWorkspaceAST{asts: map[string]*tuigen.File{}})
	if rp == nil {
		t.Fatal("expected non-nil provider")
	}

	ctx := makeCtx(doc, NodeKindUnknown, "")
	result, err := rp.References(ctx, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %+v", result)
	}
}

func TestReferences_FuncParam(t *testing.T) {
	src := `package test

templ Display() {
	<span>x</span>
}

func format(msg string) string {
	return msg + msg
}
`
	doc := parseTestDoc(src)
	fn := doc.AST.Funcs[0]

	index := newCovIndex()
	index.funcParams["format.msg"] = &FuncParamInfo{
		Name:     "msg",
		Type:     "string",
		FuncName: "format",
		Location: Location{URI: doc.URI, Range: Range{
			Start: Position{Line: 6, Character: 12},
			End:   Position{Line: 6, Character: 15},
		}},
	}

	rp := newTestReferencesProvider(index, &stubDocAccessor{docs: []*Document{doc}})

	ctx := makeCtx(doc, NodeKindParameter, "msg")
	ctx.Scope.Function = fn

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 references (decl + 2 body usages), got %d: %+v", len(result), result)
	}
	if !hasLocation(result, doc.URI, 6, 12, 3) {
		t.Errorf("missing declaration at 6:12, got %+v", result)
	}
	// "	return msg + msg" on line 7: msg at char 8 and char 14.
	if !hasLocation(result, doc.URI, 7, 8, 3) {
		t.Errorf("missing first body usage at 7:8, got %+v", result)
	}
	if !hasLocation(result, doc.URI, 7, 14, 3) {
		t.Errorf("missing second body usage at 7:14, got %+v", result)
	}
}

func TestReferences_GoCodeVariable(t *testing.T) {
	src := `package test

templ App() {
	msg := tui.NewRef()
	<span>{msg}</span>
}
`
	doc := parseTestDoc(src)

	rp := newTestReferencesProvider(newStubIndex(), &stubDocAccessor{docs: []*Document{doc}})

	ctx := makeCtx(doc, NodeKindUnknown, "msg")
	ctx.Scope.Component = doc.AST.Components[0]

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 references (decl + usage), got %d: %+v", len(result), result)
	}
	// Declaration "msg :=" on line 3 at char 1 (after the tab).
	if !hasLocation(result, doc.URI, 3, 1, 3) {
		t.Errorf("missing declaration at 3:1, got %+v", result)
	}
	// Usage inside {msg} on line 4: "{" is char 7, so msg starts at char 8.
	if !hasLocation(result, doc.URI, 4, 8, 3) {
		t.Errorf("missing usage at 4:8, got %+v", result)
	}
}

func TestReferences_WorkspaceComponentSearch(t *testing.T) {
	openSrc := `package test

templ Page() {
	@Header("a")
}
`
	wsSrc := `package test

templ Other() {
	<div>
		@Header("b")
	</div>
}
`
	openDoc := parseTestDoc(openSrc)
	openDoc.URI = "file:///open.gsx"
	wsDoc := parseTestDoc(wsSrc)

	index := newStubIndex()
	index.components["Header"] = &ComponentInfo{
		Name: "Header",
		Location: Location{URI: "file:///header.gsx", Range: Range{
			Start: Position{Line: 2, Character: 0},
			End:   Position{Line: 2, Character: 12},
		}},
	}

	rp := &referencesProvider{
		index: index,
		docs:  &stubDocAccessor{docs: []*Document{openDoc}},
		workspace: &stubWorkspaceAST{asts: map[string]*tuigen.File{
			"file:///open.gsx": openDoc.AST, // skipped: already open
			"file:///nil.gsx":  nil,         // skipped: nil AST
			"file:///ws.gsx":   wsDoc.AST,
		}},
	}

	ctx := makeCtx(openDoc, NodeKindComponentCall, "@Header")

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 references (decl + open + workspace), got %d: %+v", len(result), result)
	}
	if !hasLocation(result, "file:///header.gsx", 2, 0, 12) {
		t.Errorf("missing declaration location, got %+v", result)
	}
	if !hasLocation(result, "file:///open.gsx", 3, 1, len("@Header")) {
		t.Errorf("missing open-document call at 3:1, got %+v", result)
	}
	if !hasLocation(result, "file:///ws.gsx", 4, 2, len("@Header")) {
		t.Errorf("missing workspace call at 4:2, got %+v", result)
	}
}

func TestReferences_WorkspaceFunctionSearch(t *testing.T) {
	openSrc := `package test

templ Page() {
	<span>{helper("a")}</span>
}
`
	wsSrc := `package test

templ Other() {
	<span>{helper("b")}</span>
}
`
	openDoc := parseTestDoc(openSrc)
	openDoc.URI = "file:///open.gsx"
	wsDoc := parseTestDoc(wsSrc)

	index := newStubIndex()
	index.functions["helper"] = &FuncInfo{
		Name: "helper",
		Location: Location{URI: "file:///lib.gsx", Range: Range{
			Start: Position{Line: 8, Character: 5},
			End:   Position{Line: 8, Character: 11},
		}},
	}

	rp := &referencesProvider{
		index: index,
		docs:  &stubDocAccessor{docs: []*Document{openDoc}},
		workspace: &stubWorkspaceAST{asts: map[string]*tuigen.File{
			"file:///open.gsx": openDoc.AST,
			"file:///nil.gsx":  nil,
			"file:///ws.gsx":   wsDoc.AST,
		}},
	}

	ctx := makeCtx(openDoc, NodeKindFunction, "helper")

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 references (decl + open + workspace), got %d: %+v", len(result), result)
	}
	if !hasLocation(result, "file:///lib.gsx", 8, 5, 6) {
		t.Errorf("missing declaration location, got %+v", result)
	}
	// {helper("a")}: "{" at char 7, helper starts at char 8.
	if !hasLocation(result, "file:///open.gsx", 3, 8, 6) {
		t.Errorf("missing open-document call at 3:8, got %+v", result)
	}
	if !hasLocation(result, "file:///ws.gsx", 3, 8, 6) {
		t.Errorf("missing workspace call at 3:8, got %+v", result)
	}
}

func TestReferences_FallbackDispatch(t *testing.T) {
	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<div ref={panel}>
		<span>{count.Get()}</span>
		<span>{panel}</span>
	</div>
}
`
	doc := parseTestDoc(src)
	comp := doc.AST.Components[0]
	elem := comp.Body[1].(*tuigen.Element)

	t.Run("scope ref match", func(t *testing.T) {
		rp := newTestReferencesProvider(newStubIndex(), &stubDocAccessor{docs: []*Document{doc}})
		ctx := makeCtx(doc, NodeKindUnknown, "panel")
		ctx.Scope.Component = comp
		ctx.Scope.Refs = []tuigen.RefInfo{{Name: "panel", Element: elem}}

		result, err := rp.References(ctx, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 references (ref attr + usage), got %d: %+v", len(result), result)
		}
		// ref={panel} declaration on line 4 at char 6.
		if !hasLocation(result, doc.URI, 4, 6, len("ref={panel}")) {
			t.Errorf("missing ref attr declaration, got %+v", result)
		}
		// {panel} usage on line 6: "{" at char 8, panel at char 9.
		if !hasLocation(result, doc.URI, 6, 9, 5) {
			t.Errorf("missing ref usage at 6:9, got %+v", result)
		}
	})

	t.Run("scope state var match", func(t *testing.T) {
		rp := newTestReferencesProvider(newStubIndex(), &stubDocAccessor{docs: []*Document{doc}})
		ctx := makeCtx(doc, NodeKindUnknown, "count")
		ctx.Scope.Component = comp
		ctx.Scope.StateVars = []tuigen.StateVar{{Name: "count", Type: "int"}}

		result, err := rp.References(ctx, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Declaration + decl-line usage + {count.Get()} usage.
		if len(result) != 3 {
			t.Fatalf("expected 3 references for state var, got %d: %+v", len(result), result)
		}
		if !hasLocation(result, doc.URI, 3, 1, 5) {
			t.Errorf("missing state declaration at 3:1, got %+v", result)
		}
		if !hasLocation(result, doc.URI, 5, 9, 5) {
			t.Errorf("missing usage at 5:9, got %+v", result)
		}
	})

	t.Run("component via index lookup", func(t *testing.T) {
		callDoc := parseTestDoc(`package test

templ Page() {
	@Header("x")
}
`)
		index := newStubIndex()
		index.components["Header"] = &ComponentInfo{
			Name:     "Header",
			Location: Location{URI: callDoc.URI, Range: Range{Start: Position{Line: 9, Character: 0}}},
		}
		rp := newTestReferencesProvider(index, &stubDocAccessor{docs: []*Document{callDoc}})
		ctx := makeCtx(callDoc, NodeKindUnknown, "Header")

		result, err := rp.References(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 call reference, got %d: %+v", len(result), result)
		}
		if !hasLocation(result, callDoc.URI, 3, 1, len("@Header")) {
			t.Errorf("missing call at 3:1, got %+v", result)
		}
	})

	t.Run("function via index lookup", func(t *testing.T) {
		callDoc := parseTestDoc(`package test

templ Page() {
	<span>{helper("x")}</span>
}
`)
		index := newStubIndex()
		index.functions["helper"] = &FuncInfo{
			Name:     "helper",
			Location: Location{URI: callDoc.URI, Range: Range{Start: Position{Line: 9, Character: 0}}},
		}
		rp := newTestReferencesProvider(index, &stubDocAccessor{docs: []*Document{callDoc}})
		ctx := makeCtx(callDoc, NodeKindUnknown, "helper")

		result, err := rp.References(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 call reference, got %d: %+v", len(result), result)
		}
		if !hasLocation(result, callDoc.URI, 3, 8, 6) {
			t.Errorf("missing call at 3:8, got %+v", result)
		}
	})
}

func TestReferences_VarFormLetBinding(t *testing.T) {
	src := `package test

templ V() {
	var header = <div>title</div>
	{header}
}
`
	doc := parseTestDoc(src)

	rp := newTestReferencesProvider(newStubIndex(), &stubDocAccessor{docs: []*Document{doc}})

	ctx := makeCtx(doc, NodeKindLetBinding, "header")
	ctx.Scope.Component = doc.AST.Components[0]

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 references (decl + usage), got %d: %+v", len(result), result)
	}
	// Declaration: "var " is 4 chars, binding starts at char 1, so name at char 5.
	if !hasLocation(result, doc.URI, 3, 5, len("header")) {
		t.Errorf("missing var-form declaration at 3:5, got %+v", result)
	}
	// Usage {header} on line 4: "{" at char 1, header at char 2.
	if !hasLocation(result, doc.URI, 4, 2, len("header")) {
		t.Errorf("missing usage at 4:2, got %+v", result)
	}
}

func TestFindComponentCallsInNodes_Recursion(t *testing.T) {
	uri := "file:///test.gsx"
	mkCall := func(line int) *tuigen.ComponentCall {
		return &tuigen.ComponentCall{Name: "Header", Position: tuigen.Position{Line: line, Column: 3}}
	}

	nodes := []tuigen.Node{
		&tuigen.Element{Tag: "div", Children: []tuigen.Node{mkCall(10)}},
		&tuigen.ForLoop{Body: []tuigen.Node{mkCall(11)}},
		&tuigen.IfStmt{
			Then: []tuigen.Node{mkCall(12)},
			Else: []tuigen.Node{mkCall(13)},
		},
		&tuigen.LetBinding{Element: &tuigen.Element{Tag: "div", Children: []tuigen.Node{mkCall(14)}}},
		&tuigen.ComponentCall{
			Name:     "Other",
			Position: tuigen.Position{Line: 15, Column: 3},
			Children: []tuigen.Node{mkCall(16)},
		},
	}

	var refs []Location
	findComponentCallsInNodes(nodes, "Header", uri, &refs)

	if len(refs) != 6 {
		t.Fatalf("expected 6 call references, got %d: %+v", len(refs), refs)
	}
	for _, line := range []int{9, 10, 11, 12, 13, 15} {
		if !hasLocation(refs, uri, line, 2, len("@Header")) {
			t.Errorf("missing call at %d:2, got %+v", line, refs)
		}
	}
}

func TestFindFunctionCallsInNodes_Recursion(t *testing.T) {
	uri := "file:///test.gsx"

	nodes := []tuigen.Node{
		&tuigen.GoCode{Code: `x := helper("a")`, Position: tuigen.Position{Line: 10, Column: 2}},
		&tuigen.Element{
			Tag: "div",
			Attributes: []*tuigen.Attribute{
				{Name: "id", Value: &tuigen.GoExpr{Code: `helper("b")`, Position: tuigen.Position{Line: 11, Column: 10}}},
			},
			Children: []tuigen.Node{
				&tuigen.GoExpr{Code: `helper("c")`, Position: tuigen.Position{Line: 12, Column: 7}},
			},
		},
		&tuigen.ForLoop{Body: []tuigen.Node{
			&tuigen.GoExpr{Code: `helper("d")`, Position: tuigen.Position{Line: 13, Column: 7}},
		}},
		&tuigen.IfStmt{
			Then: []tuigen.Node{&tuigen.GoExpr{Code: `helper("e")`, Position: tuigen.Position{Line: 14, Column: 7}}},
			Else: []tuigen.Node{&tuigen.GoExpr{Code: `helper("f")`, Position: tuigen.Position{Line: 15, Column: 7}}},
		},
		&tuigen.ComponentCall{
			Name:     "Card",
			Args:     `helper("g")`,
			Position: tuigen.Position{Line: 16, Column: 2},
		},
		&tuigen.LetBinding{Element: &tuigen.Element{Tag: "div", Children: []tuigen.Node{
			&tuigen.GoExpr{Code: `helper("h")`, Position: tuigen.Position{Line: 17, Column: 7}},
		}}},
	}

	var refs []Location
	findFunctionCallsInNodes(nodes, "helper", uri, &refs)

	if len(refs) != 8 {
		t.Fatalf("expected 8 call references, got %d: %+v", len(refs), refs)
	}
	// GoCode: column base is Position.Column-1=1, "helper" at offset 5.
	if !hasLocation(refs, uri, 9, 6, 6) {
		t.Errorf("missing GoCode call at 9:6, got %+v", refs)
	}
	// Attribute expr: column base is Position.Column=10.
	if !hasLocation(refs, uri, 10, 10, 6) {
		t.Errorf("missing attribute call at 10:10, got %+v", refs)
	}
	// GoExpr child: column base is Position.Column=7.
	if !hasLocation(refs, uri, 11, 7, 6) {
		t.Errorf("missing child call at 11:7, got %+v", refs)
	}
	// ComponentCall args: column base is Column + len("@Card") + 1 = 2+5+1 = 8.
	if !hasLocation(refs, uri, 15, 8, 6) {
		t.Errorf("missing args call at 15:8, got %+v", refs)
	}
}

func TestFindVariableUsagesInNodes_Recursion(t *testing.T) {
	uri := "file:///test.gsx"

	nodes := []tuigen.Node{
		&tuigen.RawGoExpr{Code: "item", Position: tuigen.Position{Line: 10, Column: 7}},
		&tuigen.ForLoop{
			Index:    "i",
			Value:    "v",
			Iterable: "item",
			Position: tuigen.Position{Line: 11, Column: 2},
			Body:     []tuigen.Node{},
		},
		&tuigen.IfStmt{
			Condition: "item > 0",
			Position:  tuigen.Position{Line: 12, Column: 2},
		},
		&tuigen.ComponentCall{
			Name:     "Card",
			Args:     "item",
			Position: tuigen.Position{Line: 13, Column: 2},
		},
		&tuigen.LetBinding{Element: &tuigen.Element{Tag: "div", Children: []tuigen.Node{
			&tuigen.GoExpr{Code: "item", Position: tuigen.Position{Line: 14, Column: 7}},
		}}},
	}

	var refs []Location
	findVariableUsagesInNodes(nodes, "item", uri, &refs)

	if len(refs) != 5 {
		t.Fatalf("expected 5 usages, got %d: %+v", len(refs), refs)
	}
	// RawGoExpr: startOffset 1, column base 7-1+1 = 7.
	if !hasLocation(refs, uri, 9, 7, 4) {
		t.Errorf("missing RawGoExpr usage at 9:7, got %+v", refs)
	}
	// ForLoop iterable: offset = len("for ")+len("i")+2+len("v")+len(" := range ") = 4+1+2+1+10 = 18; col 2-1+18 = 19.
	if !hasLocation(refs, uri, 10, 19, 4) {
		t.Errorf("missing iterable usage at 10:19, got %+v", refs)
	}
	// IfStmt condition: col 2-1+len("if ") = 4.
	if !hasLocation(refs, uri, 11, 4, 4) {
		t.Errorf("missing condition usage at 11:4, got %+v", refs)
	}
	// ComponentCall args: col 2-1+len("@Card")+1 = 7.
	if !hasLocation(refs, uri, 12, 7, 4) {
		t.Errorf("missing args usage at 12:7, got %+v", refs)
	}
	// LetBinding element child GoExpr: col 7-1+1 = 7.
	if !hasLocation(refs, uri, 13, 7, 4) {
		t.Errorf("missing let-binding usage at 13:7, got %+v", refs)
	}
}

func TestFindRefAttrDeclInNodes_Recursion(t *testing.T) {
	uri := "file:///test.gsx"
	content := "x\n  <div ref={panel}>\n"

	mkElem := func() *tuigen.Element {
		return &tuigen.Element{
			Tag:      "div",
			RefExpr:  &tuigen.GoExpr{Code: "panel"},
			Position: tuigen.Position{Line: 2, Column: 3},
		}
	}

	type tc struct {
		nodes []tuigen.Node
	}

	tests := map[string]tc{
		"inside for loop": {
			nodes: []tuigen.Node{&tuigen.ForLoop{Body: []tuigen.Node{mkElem()}}},
		},
		"inside if then": {
			nodes: []tuigen.Node{&tuigen.IfStmt{Then: []tuigen.Node{mkElem()}}},
		},
		"inside if else": {
			nodes: []tuigen.Node{&tuigen.IfStmt{Else: []tuigen.Node{mkElem()}}},
		},
		"inside let binding": {
			nodes: []tuigen.Node{&tuigen.LetBinding{Element: &tuigen.Element{
				Tag:      "div",
				Children: []tuigen.Node{mkElem()},
			}}},
		},
		"inside component call": {
			nodes: []tuigen.Node{&tuigen.ComponentCall{Name: "Card", Children: []tuigen.Node{mkElem()}}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var refs []Location
			findRefAttrDeclInNodes(tt.nodes, "panel", content, uri, &refs)
			if len(refs) != 1 {
				t.Fatalf("expected 1 ref declaration, got %d: %+v", len(refs), refs)
			}
			// "ref={panel}" is at line 1 (0-indexed), char 7.
			if !hasLocation(refs, uri, 1, 7, len("ref={panel}")) {
				t.Errorf("missing ref decl at 1:7, got %+v", refs)
			}
		})
	}

	t.Run("falls back to element position when attr not in content", func(t *testing.T) {
		var refs []Location
		findRefAttrDeclInNodes([]tuigen.Node{mkElem()}, "panel", "no ref here", uri, &refs)
		if len(refs) != 1 {
			t.Fatalf("expected 1 ref declaration, got %d: %+v", len(refs), refs)
		}
		if !hasLocation(refs, uri, 1, 2, len("ref={panel}")) {
			t.Errorf("expected fallback at element position 1:2, got %+v", refs)
		}
	})
}

func TestFindStateVarDeclInNodes_Recursion(t *testing.T) {
	uri := "file:///test.gsx"
	mkDecl := func(line int) *tuigen.GoCode {
		return &tuigen.GoCode{
			Code:     "count := tui.NewState(0)",
			Position: tuigen.Position{Line: line, Column: 2},
		}
	}

	type tc struct {
		nodes    []tuigen.Node
		wantLine int
	}

	tests := map[string]tc{
		"inside element": {
			nodes:    []tuigen.Node{&tuigen.Element{Tag: "div", Children: []tuigen.Node{mkDecl(10)}}},
			wantLine: 9,
		},
		"inside for loop": {
			nodes:    []tuigen.Node{&tuigen.ForLoop{Body: []tuigen.Node{mkDecl(11)}}},
			wantLine: 10,
		},
		"inside if branches": {
			nodes: []tuigen.Node{&tuigen.IfStmt{
				Then: []tuigen.Node{mkDecl(12)},
			}},
			wantLine: 11,
		},
		"inside component call": {
			nodes:    []tuigen.Node{&tuigen.ComponentCall{Name: "Card", Children: []tuigen.Node{mkDecl(13)}}},
			wantLine: 12,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var refs []Location
			findStateVarDeclInNodes(tt.nodes, "count", uri, &refs)
			if len(refs) != 1 {
				t.Fatalf("expected 1 state declaration, got %d: %+v", len(refs), refs)
			}
			if !hasLocation(refs, uri, tt.wantLine, 1, len("count")) {
				t.Errorf("missing declaration at %d:1, got %+v", tt.wantLine, refs)
			}
		})
	}
}
