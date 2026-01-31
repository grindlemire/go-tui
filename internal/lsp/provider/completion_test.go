package provider

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func newTestCompletionProvider(index ComponentIndex) *completionProvider {
	return &completionProvider{
		index:        index,
		goplsProxy:   &nilGoplsProxy{},
		virtualFiles: &nilVirtualFiles{},
	}
}

func TestCompletion_ElementTag(t *testing.T) {
	src := `package test

templ Page() {
	<
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	// Cursor right after <
	ctx := makeCtx(doc, NodeKindUnknown, "")
	ctx.Position = Position{Line: 3, Character: 2}
	ctx.Offset = PositionToOffset(doc.Content, ctx.Position)

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil CompletionList")
	}

	// Should include element tags
	found := false
	for _, item := range result.Items {
		if item.Label == "div" {
			found = true
			if item.Kind != CompletionItemKindClass {
				t.Errorf("expected div to have kind Class, got %d", item.Kind)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'div' in element completions")
	}
}

func TestCompletion_DSLKeywords(t *testing.T) {
	src := `package test

templ Page() {
	@
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	// Cursor right after @
	ctx := makeCtx(doc, NodeKindUnknown, "")
	ctx.Position = Position{Line: 3, Character: 2}
	ctx.Offset = PositionToOffset(doc.Content, ctx.Position)

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil CompletionList")
	}

	// Should include DSL keywords
	keywords := map[string]bool{"for": false, "if": false, "let": false}
	for _, item := range result.Items {
		if _, ok := keywords[item.Label]; ok {
			keywords[item.Label] = true
		}
	}
	for kw, found := range keywords {
		if !found {
			t.Errorf("expected keyword %q in completions", kw)
		}
	}
}

func TestCompletion_ComponentCall(t *testing.T) {
	src := `package test

templ Page() {
	@
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	index.components["Header"] = &ComponentInfo{
		Name: "Header",
	}

	cp := newTestCompletionProvider(index)

	ctx := makeCtx(doc, NodeKindUnknown, "")
	ctx.Position = Position{Line: 3, Character: 2}
	ctx.Offset = PositionToOffset(doc.Content, ctx.Position)

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, item := range result.Items {
		if item.Label == "Header" {
			found = true
			if item.Kind != CompletionItemKindFunction {
				t.Errorf("expected Header to have kind Function, got %d", item.Kind)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'Header' in component completions")
	}
}

func TestCompletion_Attributes(t *testing.T) {
	src := `package test

templ Page() {
	<div >
	</div>
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	// Simulate cursor inside element tag — set InElement and AttrTag
	ctx := makeCtx(doc, NodeKindUnknown, "")
	ctx.Position = Position{Line: 3, Character: 6}
	ctx.Offset = PositionToOffset(doc.Content, ctx.Position)
	ctx.InElement = true
	ctx.AttrTag = "div"

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil CompletionList")
	}

	// Should include common attributes like "class", "id"
	foundClass := false
	foundID := false
	for _, item := range result.Items {
		if item.Label == "class" {
			foundClass = true
		}
		if item.Label == "id" {
			foundID = true
		}
	}
	if !foundClass {
		t.Error("expected 'class' in attribute completions")
	}
	if !foundID {
		t.Error("expected 'id' in attribute completions")
	}
}

func TestCompletion_TailwindClasses(t *testing.T) {
	src := `package test

templ Page() {
	<div class="flex-">
	</div>
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	// Cursor inside class attribute right after "flex-"
	ctx := makeCtx(doc, NodeKindTailwindClass, "flex-")
	ctx.Position = Position{Line: 3, Character: 18}
	ctx.Offset = PositionToOffset(doc.Content, ctx.Position)
	ctx.InClassAttr = true

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil CompletionList")
	}

	// Should include flex-col, flex-row
	foundCol := false
	for _, item := range result.Items {
		if item.Label == "flex-col" {
			foundCol = true
			if item.Kind != CompletionItemKindConstant {
				t.Errorf("expected flex-col to have kind Constant, got %d", item.Kind)
			}
			break
		}
	}
	if !foundCol {
		t.Error("expected 'flex-col' in tailwind completions")
	}
}

func TestCompletion_TailwindPrefixFilter(t *testing.T) {
	src := `package test

templ Page() {
	<div class="bord">
	</div>
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	ctx := makeCtx(doc, NodeKindTailwindClass, "bord")
	ctx.Position = Position{Line: 3, Character: 17}
	ctx.Offset = PositionToOffset(doc.Content, ctx.Position)
	ctx.InClassAttr = true

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil CompletionList")
	}

	// All items should start with "bord"
	for _, item := range result.Items {
		if !strings.HasPrefix(item.Label, "bord") {
			t.Errorf("expected item %q to start with 'bord'", item.Label)
		}
	}

	if len(result.Items) == 0 {
		t.Error("expected at least some border completions")
	}
}

func TestCompletion_EventHandlerAttributes(t *testing.T) {
	src := `package test

templ Page() {
	<div >
	</div>
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	ctx := makeCtx(doc, NodeKindUnknown, "")
	ctx.InElement = true
	ctx.AttrTag = "div"

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that event handler attributes are included
	foundOnClick := false
	for _, item := range result.Items {
		if item.Label == "onClick" {
			foundOnClick = true
			break
		}
	}
	// onClick should already be in div's attribute set (it's in eventAttrs())
	// so it appears via getAttributeCompletions
	if !foundOnClick {
		t.Error("expected 'onClick' in attribute completions for div")
	}
}

func TestCompletion_StateMethodCompletions(t *testing.T) {
	// Component with a state variable; cursor is after "count." inside a Go expression
	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<span>{count.}</span>
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	// Build a context with state vars in scope and cursor in a Go expression
	ctx := makeCtx(doc, NodeKindGoExpr, "count")
	ctx.InGoExpr = true
	ctx.Scope = &Scope{
		Component: doc.AST.Components[0],
		StateVars: []tuigen.StateVar{
			{Name: "count", Type: "int", InitExpr: "0"},
		},
	}
	// Position cursor after "count." — line 4 (0-indexed), character on the dot area
	ctx.Position = Position{Line: 4, Character: 14}
	ctx.Offset = PositionToOffset(doc.Content, ctx.Position)

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil CompletionList")
	}

	// Should include state methods
	expectedMethods := map[string]bool{
		"Get()":       false,
		"Set(value)":  false,
		"Update(fn)":  false,
		"Bind(fn)":    false,
		"Batch(fn)":   false,
	}
	for _, item := range result.Items {
		if _, ok := expectedMethods[item.Label]; ok {
			expectedMethods[item.Label] = true
			if item.Kind != CompletionItemKindMethod {
				t.Errorf("expected state method %q to have kind Method, got %d", item.Label, item.Kind)
			}
		}
	}
	for method, found := range expectedMethods {
		if !found {
			t.Errorf("expected state method %q in completions", method)
		}
	}
}

func TestCompletion_StateMethodCompletions_NonStateVar(t *testing.T) {
	// Verify that non-state variables don't trigger state method completions
	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<span>{other.}</span>
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	ctx := makeCtx(doc, NodeKindGoExpr, "other")
	ctx.InGoExpr = true
	ctx.Scope = &Scope{
		Component: doc.AST.Components[0],
		StateVars: []tuigen.StateVar{
			{Name: "count", Type: "int", InitExpr: "0"},
		},
	}
	ctx.Position = Position{Line: 4, Character: 15}
	ctx.Offset = PositionToOffset(doc.Content, ctx.Position)

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT return state method completions for "other." since it's not a state var
	for _, item := range result.Items {
		if item.Label == "Get()" || item.Label == "Set(value)" {
			t.Errorf("should not offer state methods for non-state var, but got %q", item.Label)
		}
	}
}

func TestCompletion_GoplsFallback(t *testing.T) {
	src := `package test

templ Page() {
	<span>{fmt.Sprintf("hello")}</span>
}
`
	doc := parseTestDoc(src)
	index := newStubIndex()
	cp := newTestCompletionProvider(index)

	// Cursor inside Go expression — gopls is nil so should fall through gracefully
	ctx := makeCtx(doc, NodeKindGoExpr, "fmt")
	ctx.InGoExpr = true

	result, err := cp.Complete(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With nil gopls proxy, should return empty or fallback completions
	if result == nil {
		t.Fatal("expected non-nil CompletionList even with nil gopls")
	}
}
