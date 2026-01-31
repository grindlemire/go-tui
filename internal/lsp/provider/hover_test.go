package provider

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// newTestHoverProvider creates a hover provider with nil gopls/virtual-file stubs.
func newTestHoverProvider(index ComponentIndex) *hoverProvider {
	return &hoverProvider{
		index:        index,
		goplsProxy:   &nilGoplsProxy{},
		virtualFiles: &nilVirtualFiles{},
	}
}

// --- Tests ---

func TestHover_Element(t *testing.T) {
	type tc struct {
		word    string
		wantNil bool
		wantIn  string
	}

	tests := map[string]tc{
		"div element": {
			word:   "div",
			wantIn: "<div>",
		},
		"span element": {
			word:   "span",
			wantIn: "<span>",
		},
		"unknown element": {
			word:    "foobar",
			wantNil: true,
		},
	}

	index := newStubIndex()
	hp := newTestHoverProvider(index)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := parseTestDoc("package test")
			ctx := makeCtx(doc, NodeKindUnknown, tt.word)

			result, err := hp.Hover(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil hover result")
			}
			if !strings.Contains(result.Contents.Value, tt.wantIn) {
				t.Errorf("hover content %q does not contain %q", result.Contents.Value, tt.wantIn)
			}
		})
	}
}

func TestHover_Keyword(t *testing.T) {
	type tc struct {
		word    string
		wantNil bool
		wantIn  string
	}

	tests := map[string]tc{
		"package keyword": {
			word:   "package",
			wantIn: "package",
		},
		"for keyword": {
			word:   "@for",
			wantIn: "@for",
		},
		"if keyword": {
			word:   "@if",
			wantIn: "@if",
		},
		"let keyword": {
			word:   "@let",
			wantIn: "@let",
		},
		"unknown word": {
			word:    "foobar",
			wantNil: true,
		},
	}

	index := newStubIndex()
	hp := newTestHoverProvider(index)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := parseTestDoc("package test")
			ctx := makeCtx(doc, NodeKindUnknown, tt.word)

			result, err := hp.Hover(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil hover for %q, got content", tt.word)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil hover result")
			}
			if result.Contents.Kind != "markdown" {
				t.Errorf("expected markdown, got %q", result.Contents.Kind)
			}
		})
	}
}

func TestHover_ComponentFromIndex(t *testing.T) {
	index := newStubIndex()
	index.components["Header"] = &ComponentInfo{
		Name: "Header",
		Params: []*tuigen.Param{
			{Name: "title", Type: "string"},
		},
	}

	hp := newTestHoverProvider(index)

	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindUnknown, "Header")

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result for component")
	}
	if !strings.Contains(result.Contents.Value, "Header") {
		t.Errorf("hover should mention component name, got: %s", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "title string") {
		t.Errorf("hover should mention param signature, got: %s", result.Contents.Value)
	}
}

func TestHover_FunctionFromIndex(t *testing.T) {
	index := newStubIndex()
	index.functions["helper"] = &FuncInfo{
		Name:      "helper",
		Signature: "func helper(s string) string",
	}

	hp := newTestHoverProvider(index)

	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindUnknown, "helper")

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result for function")
	}
	if !strings.Contains(result.Contents.Value, "helper") {
		t.Errorf("hover should mention function name, got: %s", result.Contents.Value)
	}
}

func TestHover_Parameter(t *testing.T) {
	index := newStubIndex()
	index.params["Header.title"] = &ParamInfo{
		Name:          "title",
		Type:          "string",
		ComponentName: "Header",
	}

	hp := newTestHoverProvider(index)

	src := `package test

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)
	comp := doc.AST.Components[0]

	ctx := makeCtx(doc, NodeKindUnknown, "title")
	ctx.Scope.Component = comp

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result for parameter")
	}
	if !strings.Contains(result.Contents.Value, "title") {
		t.Errorf("hover should mention param name, got: %s", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "string") {
		t.Errorf("hover should mention param type, got: %s", result.Contents.Value)
	}
}

func TestHover_EventHandler(t *testing.T) {
	index := newStubIndex()
	hp := newTestHoverProvider(index)

	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindEventHandler, "onClick")
	ctx.AttrName = "onClick"

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result for event handler")
	}
	if !strings.Contains(result.Contents.Value, "onClick") {
		t.Errorf("hover should mention event handler name, got: %s", result.Contents.Value)
	}
}

func TestHover_TailwindInClass(t *testing.T) {
	index := newStubIndex()
	hp := newTestHoverProvider(index)

	content := `<div class="flex-col p-2">`
	doc := &Document{URI: "file:///test.gsx", Content: content, Version: 1}

	// Position cursor inside class value, on "flex-col"
	classIdx := strings.Index(content, "flex-col")
	ctx := &CursorContext{
		Document:    doc,
		Position:    Position{Line: 0, Character: classIdx + 2},
		Offset:      classIdx + 2,
		NodeKind:    NodeKindTailwindClass,
		Word:        "flex-col",
		InClassAttr: true,
		Scope:       &Scope{},
	}

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result for tailwind class")
	}
	if !strings.Contains(result.Contents.Value, "flex-col") {
		t.Errorf("hover should mention class name, got: %s", result.Contents.Value)
	}
}

func TestHover_RefAttr(t *testing.T) {
	src := `package test

templ Layout() {
	<div ref={header} class="p-1">content</div>
}
`
	doc := parseTestDoc(src)
	elem := doc.AST.Components[0].Body[0].(*tuigen.Element)

	index := newStubIndex()
	hp := newTestHoverProvider(index)

	ctx := makeCtx(doc, NodeKindRefAttr, "header")
	ctx.Node = elem
	ctx.Scope.Refs = []tuigen.RefInfo{
		{Name: "header", ExportName: "Header", Element: elem},
	}

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result for ref attr")
	}
	if !strings.Contains(result.Contents.Value, "Element Ref") {
		t.Errorf("hover should mention 'Element Ref', got: %s", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "header") {
		t.Errorf("hover should mention ref name, got: %s", result.Contents.Value)
	}
}

func TestHover_StateDecl(t *testing.T) {
	index := newStubIndex()
	hp := newTestHoverProvider(index)

	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindStateDecl, "count")
	ctx.Scope.StateVars = []tuigen.StateVar{
		{Name: "count", Type: "int", InitExpr: "0"},
	}

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result for state decl")
	}
	if !strings.Contains(result.Contents.Value, "State Variable") {
		t.Errorf("hover should mention 'State Variable', got: %s", result.Contents.Value)
	}
}
