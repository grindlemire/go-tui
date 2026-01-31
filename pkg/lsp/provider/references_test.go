package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/pkg/tuigen"
)

func newTestReferencesProvider(index ComponentIndex, docs DocumentAccessor) *referencesProvider {
	return &referencesProvider{
		index:     index,
		docs:      docs,
		workspace: &stubWorkspaceAST{asts: make(map[string]*tuigen.File)},
	}
}

func TestReferences_Component(t *testing.T) {
	src := `package test

templ Page() {
	@Header("title")
}

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)

	index := newStubIndex()
	index.components["Header"] = &ComponentInfo{
		Name: "Header",
		Location: Location{
			URI: doc.URI,
			Range: Range{
				Start: Position{Line: 6, Character: 0},
				End:   Position{Line: 6, Character: 17},
			},
		},
	}

	docsAccessor := &stubDocAccessor{docs: []*Document{doc}}
	rp := newTestReferencesProvider(index, docsAccessor)

	ctx := makeCtx(doc, NodeKindComponentCall, "@Header")

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 2 {
		t.Errorf("expected at least 2 references (decl + usage), got %d", len(result))
	}
}

func TestReferences_Function(t *testing.T) {
	src := `package test

templ Display() {
	<span>{helper("test")}</span>
}

func helper(s string) string {
	return s
}
`
	doc := parseTestDoc(src)

	index := newStubIndex()
	index.functions["helper"] = &FuncInfo{
		Name:      "helper",
		Signature: "func helper(s string) string",
		Location: Location{
			URI: doc.URI,
			Range: Range{
				Start: Position{Line: 6, Character: 0},
				End:   Position{Line: 6, Character: 16},
			},
		},
	}

	docsAccessor := &stubDocAccessor{docs: []*Document{doc}}
	rp := newTestReferencesProvider(index, docsAccessor)

	ctx := makeCtx(doc, NodeKindFunction, "helper")

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 2 {
		t.Errorf("expected at least 2 references (decl + call), got %d", len(result))
	}
}

func TestReferences_Parameter(t *testing.T) {
	src := `package test

templ Header(title string) {
	<span>{title}</span>
}
`
	doc := parseTestDoc(src)

	index := newStubIndex()
	index.components["Header"] = &ComponentInfo{Name: "Header"}
	index.params["Header.title"] = &ParamInfo{
		Name:          "title",
		Type:          "string",
		ComponentName: "Header",
		Location: Location{
			URI: doc.URI,
			Range: Range{
				Start: Position{Line: 2, Character: 13},
				End:   Position{Line: 2, Character: 18},
			},
		},
	}

	docsAccessor := &stubDocAccessor{docs: []*Document{doc}}
	rp := newTestReferencesProvider(index, docsAccessor)

	ctx := makeCtx(doc, NodeKindParameter, "title")
	ctx.Scope.Component = doc.AST.Components[0]

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 2 {
		t.Errorf("expected at least 2 references (decl + usage), got %d", len(result))
	}
}

func TestReferences_LetBinding(t *testing.T) {
	src := `package test

templ Example() {
	@let header = <div>title</div>
	{header}
}
`
	doc := parseTestDoc(src)

	index := newStubIndex()
	docsAccessor := &stubDocAccessor{docs: []*Document{doc}}
	rp := newTestReferencesProvider(index, docsAccessor)

	ctx := makeCtx(doc, NodeKindLetBinding, "header")
	ctx.Scope.Component = doc.AST.Components[0]

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 2 {
		t.Errorf("expected at least 2 references (decl + usage), got %d", len(result))
	}
}

func TestReferences_EmptyWord(t *testing.T) {
	doc := parseTestDoc("package test")
	index := newStubIndex()
	docsAccessor := &stubDocAccessor{docs: []*Document{doc}}
	rp := newTestReferencesProvider(index, docsAccessor)

	ctx := makeCtx(doc, NodeKindUnknown, "")

	result, err := rp.References(ctx, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 references for empty word, got %d", len(result))
	}
}

func TestReferences_NamedRef(t *testing.T) {
	src := `package test

templ Layout() {
	<div #Header class="p-1">title</div>
	<span>{Header}</span>
}
`
	doc := parseTestDoc(src)

	index := newStubIndex()
	docsAccessor := &stubDocAccessor{docs: []*Document{doc}}
	rp := newTestReferencesProvider(index, docsAccessor)

	ctx := makeCtx(doc, NodeKindNamedRef, "Header")
	ctx.Scope.Component = doc.AST.Components[0]
	ctx.Scope.NamedRefs = []tuigen.NamedRef{
		{Name: "Header", Position: tuigen.Position{Line: 4, Column: 7}},
	}

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should find at least the declaration and the usage in {Header}
	if len(result) < 2 {
		t.Errorf("expected at least 2 references (decl + usage) for named ref, got %d", len(result))
	}
}

func TestReferences_StateVar(t *testing.T) {
	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}
`
	doc := parseTestDoc(src)

	index := newStubIndex()
	docsAccessor := &stubDocAccessor{docs: []*Document{doc}}
	rp := newTestReferencesProvider(index, docsAccessor)

	ctx := makeCtx(doc, NodeKindStateDecl, "count")
	ctx.Scope.Component = doc.AST.Components[0]
	ctx.Scope.StateVars = []tuigen.StateVar{
		{Name: "count", Type: "int", InitExpr: "0", Position: tuigen.Position{Line: 4, Column: 2}},
	}

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should find at least the declaration and the usage in {count.Get()}
	if len(result) < 2 {
		t.Errorf("expected at least 2 references (decl + usage) for state var, got %d", len(result))
	}
}

func TestReferences_LoopVariable(t *testing.T) {
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

	index := newStubIndex()
	docsAccessor := &stubDocAccessor{docs: []*Document{doc}}
	rp := newTestReferencesProvider(index, docsAccessor)

	ctx := makeCtx(doc, NodeKindUnknown, "item")
	ctx.Scope.Component = doc.AST.Components[0]
	ctx.Position = Position{Line: 5, Character: 12}

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 1 {
		t.Errorf("expected at least 1 reference for loop variable, got %d", len(result))
	}
}
