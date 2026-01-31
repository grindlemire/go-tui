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
	if len(result) != 2 {
		t.Errorf("expected 2 references (decl + usage), got %d", len(result))
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
	if len(result) != 2 {
		t.Errorf("expected 2 references (decl + call), got %d", len(result))
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
	if len(result) != 2 {
		t.Errorf("expected 2 references (decl + usage), got %d", len(result))
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
	if len(result) != 2 {
		t.Errorf("expected 2 references (decl + usage), got %d", len(result))
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
	if len(result) != 2 {
		t.Errorf("expected 2 references (decl + usage) for named ref, got %d", len(result))
	}
}

func TestReferences_NamedRef_Multiline(t *testing.T) {
	src := `package test

templ Layout() {
	<div
		#Header
		class="p-1">title</div>
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
		{Name: "Header", Position: tuigen.Position{Line: 5, Column: 3}},
	}

	result, err := rp.References(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 references (decl + usage) for multiline named ref, got %d", len(result))
	}

	// The declaration should point to #Header on line 4 (0-indexed), not the <div line
	if len(result) > 0 {
		decl := result[0]
		if decl.Range.Start.Line != 4 {
			t.Errorf("expected decl on line 4, got %d", decl.Range.Start.Line)
		}
		if decl.Range.End.Character-decl.Range.Start.Character != len("#Header") {
			t.Errorf("expected decl range length %d, got %d", len("#Header"), decl.Range.End.Character-decl.Range.Start.Character)
		}
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
	if len(result) != 3 {
		t.Errorf("expected 3 references (decl + decl-line usage + expr usage) for state var, got %d", len(result))
	}
}

func TestReferences_StateVarWordBoundary(t *testing.T) {
	// Regression: "count" should not match "accountCount" as a state var reference.
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

	// All results should reference exactly "count", not a substring match
	for _, ref := range result {
		nameLen := ref.Range.End.Character - ref.Range.Start.Character
		if nameLen != len("count") {
			t.Errorf("reference range has width %d, expected %d (len 'count')", nameLen, len("count"))
		}
	}
}

func TestIndexWholeWord(t *testing.T) {
	type tc struct {
		s    string
		word string
		want int
	}

	tests := map[string]tc{
		"exact match": {
			s: "count := 1", word: "count", want: 0,
		},
		"no match substring": {
			s: "accountCount := 1", word: "count", want: -1,
		},
		"match after prefix": {
			s: "accountCount := 1; count := 2", word: "count", want: 19,
		},
		"no match at all": {
			s: "foo := 1", word: "count", want: -1,
		},
		"match with dot after": {
			s: "count.Get()", word: "count", want: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := indexWholeWord(tt.s, tt.word)
			if got != tt.want {
				t.Errorf("indexWholeWord(%q, %q) = %d, want %d", tt.s, tt.word, got, tt.want)
			}
		})
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
	if len(result) != 2 {
		t.Errorf("expected 2 references (decl + usage) for loop variable, got %d", len(result))
	}
}
