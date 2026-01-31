package lsp

import (
	"testing"
)

func TestResolveCursorContext_ForLoop(t *testing.T) {
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

	// Cursor on "@for" keyword line (line 4)
	// "\t\t@for" — '@' at col 2
	ctx := ResolveCursorContext(doc, Position{Line: 4, Character: 3})

	if ctx.NodeKind != NodeKindForLoop {
		t.Errorf("expected NodeKindForLoop, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_IfStmt(t *testing.T) {
	src := `package test

templ Conditional(show bool) {
	<div>
		@if show {
			<span>visible</span>
		}
	</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "@if" line (line 4)
	ctx := ResolveCursorContext(doc, Position{Line: 4, Character: 3})

	if ctx.NodeKind != NodeKindIfStmt {
		t.Errorf("expected NodeKindIfStmt, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_LetBinding(t *testing.T) {
	src := `package test

templ Example() {
	@let header = <div>title</div>
	{header}
}
`
	doc := parseTestDoc(src)

	// Cursor on "header" after @let — "\t@let header =" — '@let ' is 5 chars starting at col 1, 'header' starts at col 6
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 7})

	if ctx.NodeKind != NodeKindLetBinding {
		t.Errorf("expected NodeKindLetBinding, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_ComponentCall(t *testing.T) {
	src := `package test

templ Page() {
	@Header("My Title")
}

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "@Header" (line 3)
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 2})

	if ctx.NodeKind != NodeKindComponentCall {
		t.Errorf("expected NodeKindComponentCall, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_Function(t *testing.T) {
	src := `package test

func helper(s string) string {
	return s
}
`
	doc := parseTestDoc(src)

	// Cursor on "func" line (line 2)
	ctx := ResolveCursorContext(doc, Position{Line: 2, Character: 5})

	if ctx.NodeKind != NodeKindFunction {
		t.Errorf("expected NodeKindFunction, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_GoExpr(t *testing.T) {
	src := `package test

templ Display(count int) {
	<span>{fmt.Sprintf("%d", count)}</span>
}
`
	doc := parseTestDoc(src)

	// Cursor inside {fmt.Sprintf...} — inside the braces
	// "\t<span>{fmt.Sprintf(..." — '{' is around col 7
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 10})

	if !ctx.InGoExpr {
		t.Error("expected InGoExpr to be true")
	}
}

func TestResolveCursorContext_TailwindClass(t *testing.T) {
	src := `package test

templ Box() {
	<div class="flex-col p-2">content</div>
}
`
	doc := parseTestDoc(src)

	// Cursor inside class="flex-col" — on "flex-col"
	// "\t<div class=\"flex-col p-2\">" — need position inside the class value
	line := getLineText(doc.Content, 3)
	classIdx := -1
	for i := 0; i < len(line)-6; i++ {
		if line[i:i+6] == `class=` {
			classIdx = i
			break
		}
	}
	if classIdx < 0 {
		t.Fatal("could not find class= in line")
	}
	// Position right after class=" — cursor on 'f' of flex-col
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: classIdx + 7})

	if !ctx.InClassAttr {
		t.Error("expected InClassAttr to be true")
	}
}

func TestResolveCursorContext_InElement(t *testing.T) {
	src := `package test

templ Box() {
	<div class="p-1">content</div>
}
`
	doc := parseTestDoc(src)

	// Cursor inside element tag (between < and >), on "class"
	line := getLineText(doc.Content, 3)
	classIdx := -1
	for i := 0; i < len(line)-5; i++ {
		if line[i:i+5] == "class" {
			classIdx = i
			break
		}
	}
	if classIdx < 0 {
		t.Fatal("could not find 'class' in line")
	}
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: classIdx})

	if !ctx.InElement {
		t.Error("expected InElement to be true")
	}
}

func TestResolveCursorContext_Keyword(t *testing.T) {
	src := `package test
`
	doc := parseTestDoc(src)

	// Cursor on "package" (line 0, character 3)
	ctx := ResolveCursorContext(doc, Position{Line: 0, Character: 3})

	if ctx.NodeKind != NodeKindKeyword {
		t.Errorf("expected NodeKindKeyword, got %s", ctx.NodeKind)
	}
	if ctx.Word != "package" {
		t.Errorf("expected Word 'package', got %q", ctx.Word)
	}
}

func TestResolveCursorContext_ScopeCollectsNamedRefs(t *testing.T) {
	src := `package test

templ Layout() {
	<div #Header class="p-1">title</div>
	<div #Footer class="p-1">footer</div>
}
`
	doc := parseTestDoc(src)

	// Cursor anywhere inside the component body
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 2})

	if ctx.Scope.Component == nil {
		t.Fatal("expected Scope.Component to be set")
	}
	if len(ctx.Scope.NamedRefs) < 2 {
		t.Errorf("expected at least 2 named refs in scope, got %d", len(ctx.Scope.NamedRefs))
	}

	// Verify ref names
	refNames := make(map[string]bool)
	for _, ref := range ctx.Scope.NamedRefs {
		refNames[ref.Name] = true
	}
	if !refNames["Header"] {
		t.Error("expected 'Header' in named refs")
	}
	if !refNames["Footer"] {
		t.Error("expected 'Footer' in named refs")
	}
}

func TestResolveCursorContext_ScopeCollectsLetBindings(t *testing.T) {
	src := `package test

templ Example() {
	@let title = <span>Hello</span>
	{title}
}
`
	doc := parseTestDoc(src)

	// Cursor inside the component body
	ctx := ResolveCursorContext(doc, Position{Line: 4, Character: 2})

	if ctx.Scope.Component == nil {
		t.Fatal("expected Scope.Component to be set")
	}
	if len(ctx.Scope.LetBinds) < 1 {
		t.Errorf("expected at least 1 let binding in scope, got %d", len(ctx.Scope.LetBinds))
	}
}

func TestResolveCursorContext_ScopeCollectsParams(t *testing.T) {
	src := `package test

templ Header(title string, subtitle string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)

	// Cursor inside the component body
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 2})

	if len(ctx.Scope.Params) != 2 {
		t.Errorf("expected 2 params in scope, got %d", len(ctx.Scope.Params))
	}
}

func TestResolveCursorContext_NilAST(t *testing.T) {
	// Test graceful handling when AST is nil (unparseable file)
	doc := &Document{
		URI:     "file:///broken.gsx",
		Content: "this is not valid gsx at all {{{{",
		Version: 1,
		AST:     nil,
	}

	ctx := ResolveCursorContext(doc, Position{Line: 0, Character: 0})

	// Should not panic, should return a valid context
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if ctx.Document != doc {
		t.Error("expected Document to be set")
	}
}

func TestResolveCursorContext_EmptyDocument(t *testing.T) {
	doc := &Document{
		URI:     "file:///empty.gsx",
		Content: "",
		Version: 1,
		AST:     nil,
	}

	ctx := ResolveCursorContext(doc, Position{Line: 0, Character: 0})

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if ctx.NodeKind != NodeKindUnknown {
		t.Errorf("expected NodeKindUnknown for empty document, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_StateDecl(t *testing.T) {
	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}
`
	doc := parseTestDoc(src)

	// Cursor on the state declaration line (line 3)
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 5})

	if ctx.NodeKind != NodeKindStateDecl {
		t.Errorf("expected NodeKindStateDecl, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_StateAccess(t *testing.T) {
	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}
`
	doc := parseTestDoc(src)

	// Cursor on "count" inside {count.Get()} expression (line 4)
	// "\t<span>{count.Get()}</span>" — '{' is at col 7, 'count' starts at col 8
	ctx := ResolveCursorContext(doc, Position{Line: 4, Character: 9})

	if ctx.NodeKind != NodeKindStateAccess {
		t.Errorf("expected NodeKindStateAccess, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_TextContent(t *testing.T) {
	src := `package test

templ Simple() {
	<div>Hello World</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "Hello World" text content (line 3)
	// This is tricky because text content and element tag are on the same line.
	// The inline text may be parsed as part of the element's children.
	// Let's test the word extraction instead.
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 8})

	if ctx.Word == "" {
		t.Error("expected Word to be non-empty on text content")
	}
}

func TestNodeKind_String(t *testing.T) {
	type tc struct {
		kind NodeKind
		want string
	}

	tests := map[string]tc{
		"component":     {kind: NodeKindComponent, want: "Component"},
		"element":       {kind: NodeKindElement, want: "Element"},
		"attribute":     {kind: NodeKindAttribute, want: "Attribute"},
		"named ref":     {kind: NodeKindNamedRef, want: "NamedRef"},
		"go expr":       {kind: NodeKindGoExpr, want: "GoExpr"},
		"for loop":      {kind: NodeKindForLoop, want: "ForLoop"},
		"if stmt":       {kind: NodeKindIfStmt, want: "IfStmt"},
		"let binding":   {kind: NodeKindLetBinding, want: "LetBinding"},
		"state decl":    {kind: NodeKindStateDecl, want: "StateDecl"},
		"state access":  {kind: NodeKindStateAccess, want: "StateAccess"},
		"parameter":     {kind: NodeKindParameter, want: "Parameter"},
		"function":      {kind: NodeKindFunction, want: "Function"},
		"component call":{kind: NodeKindComponentCall, want: "ComponentCall"},
		"event handler": {kind: NodeKindEventHandler, want: "EventHandler"},
		"text":          {kind: NodeKindText, want: "Text"},
		"keyword":       {kind: NodeKindKeyword, want: "Keyword"},
		"tailwind class":{kind: NodeKindTailwindClass, want: "TailwindClass"},
		"unknown":       {kind: NodeKindUnknown, want: "Unknown"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.kind.String()
			if got != tt.want {
				t.Errorf("NodeKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
			}
		})
	}
}

