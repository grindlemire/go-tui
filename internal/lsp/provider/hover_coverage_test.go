package provider

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// covIndex extends stubIndex with function parameter lookups.
type covIndex struct {
	*stubIndex
	funcParams map[string]*FuncParamInfo
}

func newCovIndex() *covIndex {
	return &covIndex{
		stubIndex:  newStubIndex(),
		funcParams: make(map[string]*FuncParamInfo),
	}
}

func (c *covIndex) LookupFuncParam(funcName, paramName string) (*FuncParamInfo, bool) {
	info, ok := c.funcParams[funcName+"."+paramName]
	return info, ok
}

func TestNewHoverProvider_Constructor(t *testing.T) {
	hp := NewHoverProvider(newStubIndex(), &nilGoplsProxy{}, &nilVirtualFiles{})
	if hp == nil {
		t.Fatal("expected non-nil provider")
	}

	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindUnknown, "div")
	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !strings.Contains(result.Contents.Value, "<div>") {
		t.Errorf("expected element hover for div, got %+v", result)
	}
}

func TestHover_ComponentNode(t *testing.T) {
	src := `package test

templ Header(title string, count int) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)
	comp := doc.AST.Components[0]

	t.Run("fallback builds signature from AST", func(t *testing.T) {
		hp := newTestHoverProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindComponent, "Header")
		ctx.Node = comp

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected hover result")
		}
		if !strings.Contains(result.Contents.Value, "func Header(title string, count int) *element.Element") {
			t.Errorf("expected AST-built signature, got: %s", result.Contents.Value)
		}
		if !strings.Contains(result.Contents.Value, "TUI Component") {
			t.Errorf("expected TUI Component label, got: %s", result.Contents.Value)
		}
	})

	t.Run("index hit uses component info", func(t *testing.T) {
		index := newStubIndex()
		index.components["Header"] = &ComponentInfo{
			Name:   "Header",
			Params: []*tuigen.Param{{Name: "title", Type: "string"}},
		}
		hp := newTestHoverProvider(index)
		ctx := makeCtx(doc, NodeKindComponent, "Header")
		ctx.Node = comp

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil || !strings.Contains(result.Contents.Value, "func Header(title string)") {
			t.Errorf("expected index-built signature, got: %+v", result)
		}
	})

	t.Run("wrong node type returns nil", func(t *testing.T) {
		hp := newTestHoverProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindComponent, "Header")
		ctx.Node = &tuigen.Element{Tag: "div"}

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil for non-component node, got %+v", result)
		}
	})
}

func TestHover_ElementNode(t *testing.T) {
	type tc struct {
		tag     string
		wantNil bool
		wantIn  string
	}

	tests := map[string]tc{
		"known div tag": {
			tag:    "div",
			wantIn: "## `<div>`",
		},
		"known span tag": {
			tag:    "span",
			wantIn: "## `<span>`",
		},
		"unknown tag": {
			tag:     "widget",
			wantNil: true,
		},
	}

	hp := newTestHoverProvider(newStubIndex())
	doc := parseTestDoc("package test")

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := makeCtx(doc, NodeKindElement, tt.tag)
			ctx.Node = &tuigen.Element{Tag: tt.tag}

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
				t.Fatal("expected hover result")
			}
			if !strings.Contains(result.Contents.Value, tt.wantIn) {
				t.Errorf("hover %q does not contain %q", result.Contents.Value, tt.wantIn)
			}
			if !strings.Contains(result.Contents.Value, "**Available attributes:**") {
				t.Errorf("expected attribute list in element hover, got: %s", result.Contents.Value)
			}
		})
	}

	t.Run("wrong node type returns nil", func(t *testing.T) {
		ctx := makeCtx(doc, NodeKindElement, "div")
		ctx.Node = &tuigen.GoExpr{Code: "x"}
		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil for non-element node, got %+v", result)
		}
	})
}

func TestHover_AttributeNode(t *testing.T) {
	type tc struct {
		attrTag  string
		attrName string
		wantNil  bool
		wantIn   string
	}

	tests := map[string]tc{
		"known attribute": {
			attrTag:  "div",
			attrName: "class",
			wantIn:   "**class**",
		},
		"unknown attribute falls back": {
			attrTag:  "div",
			attrName: "zzz",
			wantIn:   "**zzz** attribute on `<div>`",
		},
		"missing tag returns nil": {
			attrTag:  "",
			attrName: "class",
			wantNil:  true,
		},
		"missing name returns nil": {
			attrTag:  "div",
			attrName: "",
			wantNil:  true,
		},
	}

	hp := newTestHoverProvider(newStubIndex())
	doc := parseTestDoc("package test")

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := makeCtx(doc, NodeKindAttribute, tt.attrName)
			ctx.AttrTag = tt.attrTag
			ctx.AttrName = tt.attrName

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
				t.Fatal("expected hover result")
			}
			if !strings.Contains(result.Contents.Value, tt.wantIn) {
				t.Errorf("hover %q does not contain %q", result.Contents.Value, tt.wantIn)
			}
		})
	}
}

func TestHover_EventHandlerUnknown(t *testing.T) {
	hp := newTestHoverProvider(newStubIndex())
	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindEventHandler, "onWarp")
	ctx.AttrName = "onWarp"

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for unknown event handler, got %+v", result)
	}
}

func TestHover_ParameterNode(t *testing.T) {
	src := `package test

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)
	comp := doc.AST.Components[0]
	param := comp.Params[0]
	hp := newTestHoverProvider(newStubIndex())

	t.Run("with component scope", func(t *testing.T) {
		ctx := makeCtx(doc, NodeKindParameter, "title")
		ctx.Node = param
		ctx.Scope.Component = comp

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected hover result")
		}
		if !strings.Contains(result.Contents.Value, "title string") {
			t.Errorf("expected param signature, got: %s", result.Contents.Value)
		}
		if !strings.Contains(result.Contents.Value, "**Parameter** of component `Header`") {
			t.Errorf("expected component name, got: %s", result.Contents.Value)
		}
	})

	t.Run("without component scope", func(t *testing.T) {
		ctx := makeCtx(doc, NodeKindParameter, "title")
		ctx.Node = param

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected hover result")
		}
		if !strings.Contains(result.Contents.Value, "**Parameter** of component ``") {
			t.Errorf("expected empty component name, got: %s", result.Contents.Value)
		}
	})

	t.Run("wrong node type returns nil", func(t *testing.T) {
		ctx := makeCtx(doc, NodeKindParameter, "title")
		ctx.Node = &tuigen.Element{Tag: "div"}

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil for non-param node, got %+v", result)
		}
	})
}

func TestHover_KeywordNodeKinds(t *testing.T) {
	type tc struct {
		kind NodeKind
		word string
	}

	tests := map[string]tc{
		"keyword kind":     {kind: NodeKindKeyword, word: "templ"},
		"for loop kind":    {kind: NodeKindForLoop, word: "for"},
		"if stmt kind":     {kind: NodeKindIfStmt, word: "if"},
		"let binding kind": {kind: NodeKindLetBinding, word: "else"},
	}

	hp := newTestHoverProvider(newStubIndex())
	doc := parseTestDoc("package test")

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := makeCtx(doc, tt.kind, tt.word)
			result, err := hp.Hover(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatalf("expected keyword hover for %q", tt.word)
			}
			if result.Contents.Kind != "markdown" {
				t.Errorf("expected markdown, got %q", result.Contents.Kind)
			}
			if result.Contents.Value == "" {
				t.Error("expected non-empty documentation")
			}
		})
	}
}

func TestHover_FunctionNode(t *testing.T) {
	src := `package test

func helper(s string) string {
	return s
}

templ App() {
	<div>{helper("x")}</div>
}
`
	doc := parseTestDoc(src)
	fn := doc.AST.Funcs[0]

	t.Run("function in index", func(t *testing.T) {
		index := newStubIndex()
		index.functions["helper"] = &FuncInfo{
			Name:      "helper",
			Signature: "func helper(s string) string",
		}
		hp := newTestHoverProvider(index)

		ctx := makeCtx(doc, NodeKindFunction, "helper")
		ctx.Node = fn

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected hover result")
		}
		if !strings.Contains(result.Contents.Value, "func helper(s string) string") {
			t.Errorf("expected signature, got: %s", result.Contents.Value)
		}
		if !strings.Contains(result.Contents.Value, "**Helper Function**") {
			t.Errorf("expected helper label, got: %s", result.Contents.Value)
		}
	})

	t.Run("function not in index returns nil", func(t *testing.T) {
		hp := newTestHoverProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindFunction, "helper")
		ctx.Node = fn

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})

	t.Run("wrong node type returns nil", func(t *testing.T) {
		hp := newTestHoverProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindFunction, "helper")
		ctx.Node = &tuigen.Element{Tag: "div"}

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})
}

func TestHover_ComponentCallNode(t *testing.T) {
	src := `package test

templ Page() {
	@Header("x")
}
`
	doc := parseTestDoc(src)
	call := doc.AST.Components[0].Body[0].(*tuigen.ComponentCall)

	t.Run("call in index", func(t *testing.T) {
		index := newStubIndex()
		index.components["Header"] = &ComponentInfo{
			Name:   "Header",
			Params: []*tuigen.Param{{Name: "title", Type: "string"}},
		}
		hp := newTestHoverProvider(index)

		ctx := makeCtx(doc, NodeKindComponentCall, "@Header")
		ctx.Node = call

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected hover result")
		}
		if !strings.Contains(result.Contents.Value, "func Header(title string)") {
			t.Errorf("expected signature, got: %s", result.Contents.Value)
		}
	})

	t.Run("call not in index returns nil", func(t *testing.T) {
		hp := newTestHoverProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindComponentCall, "@Header")
		ctx.Node = call

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})

	t.Run("wrong node type returns nil", func(t *testing.T) {
		hp := newTestHoverProvider(newStubIndex())
		ctx := makeCtx(doc, NodeKindComponentCall, "@Header")
		ctx.Node = &tuigen.Element{Tag: "div"}

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})
}

func TestHover_RefAttrContexts(t *testing.T) {
	type tc struct {
		ref      tuigen.RefInfo
		wantType string
		wantCtx  string
	}

	tests := map[string]tc{
		"keyed loop ref": {
			ref:      tuigen.RefInfo{Name: "row", InLoop: true, KeyExpr: "item.ID"},
			wantType: "`map[KeyType]*tui.Element`",
			wantCtx:  "Keyed (map access)",
		},
		"loop ref without key": {
			ref:      tuigen.RefInfo{Name: "row", InLoop: true},
			wantType: "`[]*tui.Element`",
			wantCtx:  "Loop (slice access)",
		},
		"conditional ref": {
			ref:      tuigen.RefInfo{Name: "row", InConditional: true},
			wantType: "`*tui.Element`",
			wantCtx:  "Simple (direct access) (nullable)",
		},
	}

	src := `package test

templ Layout() {
	<div ref={row}>x</div>
}
`
	doc := parseTestDoc(src)
	elem := doc.AST.Components[0].Body[0].(*tuigen.Element)
	hp := newTestHoverProvider(newStubIndex())

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := makeCtx(doc, NodeKindRefAttr, "row")
			ctx.Node = elem
			ctx.Scope.Refs = []tuigen.RefInfo{tt.ref}

			result, err := hp.Hover(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected hover result")
			}
			if !strings.Contains(result.Contents.Value, "Type: "+tt.wantType) {
				t.Errorf("hover %q does not contain type %q", result.Contents.Value, tt.wantType)
			}
			if !strings.Contains(result.Contents.Value, "Context: "+tt.wantCtx) {
				t.Errorf("hover %q does not contain context %q", result.Contents.Value, tt.wantCtx)
			}
		})
	}

	t.Run("element without ref returns nil", func(t *testing.T) {
		ctx := makeCtx(doc, NodeKindRefAttr, "row")
		ctx.Node = &tuigen.Element{Tag: "div"}

		result, err := hp.Hover(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})
}

func TestHover_StateDeclFallback(t *testing.T) {
	hp := newTestHoverProvider(newStubIndex())
	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindStateDecl, "missing")

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected fallback hover")
	}
	if !strings.Contains(result.Contents.Value, "**State Declaration**") {
		t.Errorf("expected state declaration fallback, got: %s", result.Contents.Value)
	}
}

func TestHover_StateAccess(t *testing.T) {
	type tc struct {
		word   string
		wantIn string
	}

	tests := map[string]tc{
		"get":     {word: "Get", wantIn: "**State.Get()**"},
		"set":     {word: "Set", wantIn: "**State.Set(value)**"},
		"update":  {word: "Update", wantIn: "**State.Update(fn)**"},
		"bind":    {word: "Bind", wantIn: "**State.Bind(fn)**"},
		"batch":   {word: "Batch", wantIn: "**State.Batch(fn)**"},
		"unknown": {word: "Frobnicate", wantIn: "**State Access**"},
	}

	hp := newTestHoverProvider(newStubIndex())
	doc := parseTestDoc("package test")

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := makeCtx(doc, NodeKindStateAccess, tt.word)
			result, err := hp.Hover(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected hover result")
			}
			if !strings.Contains(result.Contents.Value, tt.wantIn) {
				t.Errorf("hover %q does not contain %q", result.Contents.Value, tt.wantIn)
			}
		})
	}
}

func TestHover_TailwindWordEdgeCases(t *testing.T) {
	type tc struct {
		content string
		offset  int
		word    string
		wantNil bool
		wantIn  string
	}

	flexContent := `<div class="flex-col p-2">`

	tests := map[string]tc{
		"no class attr before cursor": {
			content: `<div id="main">`,
			offset:  10,
			word:    "main",
			wantNil: true,
		},
		"cursor past closing quote": {
			content: `<div class="p-1" id="x">`,
			offset:  21, // inside id value, after the class value closed
			word:    "x",
			wantNil: true,
		},
		"unknown class": {
			content: `<div class="zz-bogus">`,
			offset:  15,
			word:    "zz-bogus",
			wantNil: true,
		},
		"cursor mid class extends forward": {
			content: flexContent,
			offset:  strings.Index(flexContent, "flex-col") + 4,
			word:    "flex-col",
			wantIn:  "flex-col",
		},
	}

	hp := newTestHoverProvider(newStubIndex())

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := &Document{URI: "file:///test.gsx", Content: tt.content, Version: 1}
			ctx := &CursorContext{
				Document:    doc,
				Position:    Position{Line: 0, Character: tt.offset},
				Offset:      tt.offset,
				NodeKind:    NodeKindTailwindClass,
				Word:        tt.word,
				InClassAttr: true,
				Scope:       &Scope{},
			}

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
				t.Fatal("expected hover result")
			}
			if !strings.Contains(result.Contents.Value, tt.wantIn) {
				t.Errorf("hover %q does not contain %q", result.Contents.Value, tt.wantIn)
			}
		})
	}
}

func TestHover_AttributeWordFallback(t *testing.T) {
	// When the AST does not resolve the node but the cursor is inside an
	// element, the word-based fallback consults the attribute schema.
	hp := newTestHoverProvider(newStubIndex())
	doc := parseTestDoc("package test")

	ctx := makeCtx(doc, NodeKindUnknown, "scrollable")
	ctx.InElement = true
	ctx.AttrTag = "div"

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result for attribute fallback")
	}
	if !strings.Contains(result.Contents.Value, "**scrollable**") {
		t.Errorf("expected scrollable attribute doc, got: %s", result.Contents.Value)
	}
	if !strings.Contains(result.Contents.Value, "`<div>`") {
		t.Errorf("expected tag reference, got: %s", result.Contents.Value)
	}
}

func TestHover_GoExprWordFallback(t *testing.T) {
	// NodeKindGoExpr tries gopls (nil here) then falls through to the
	// word-based component lookup.
	index := newStubIndex()
	index.components["Card"] = &ComponentInfo{Name: "Card"}
	hp := newTestHoverProvider(index)

	doc := parseTestDoc("package test")
	ctx := makeCtx(doc, NodeKindGoExpr, "Card")
	ctx.InGoExpr = true

	result, err := hp.Hover(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected hover result")
	}
	if !strings.Contains(result.Contents.Value, "func Card() *element.Element") {
		t.Errorf("expected component signature, got: %s", result.Contents.Value)
	}
}
