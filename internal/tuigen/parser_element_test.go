package tuigen

import (
	"testing"
)

func TestParser_SelfClosingElement(t *testing.T) {
	input := `package x
templ Test() {
	<input />
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	if len(comp.Body) != 1 {
		t.Fatalf("expected 1 body node, got %d", len(comp.Body))
	}

	elem, ok := comp.Body[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element, got %T", comp.Body[0])
	}

	if elem.Tag != "input" {
		t.Errorf("Tag = %q, want 'input'", elem.Tag)
	}

	if !elem.SelfClose {
		t.Error("SelfClose = false, want true")
	}
}

func TestParser_ElementWithChildren(t *testing.T) {
	input := `package x
templ Test() {
	<div>
		<span>Hello</span>
		<span>World</span>
	</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	if len(comp.Body) != 1 {
		t.Fatalf("expected 1 body node, got %d", len(comp.Body))
	}

	box, ok := comp.Body[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element, got %T", comp.Body[0])
	}

	if box.Tag != "div" {
		t.Errorf("Tag = %q, want 'div'", box.Tag)
	}

	if len(box.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(box.Children))
	}

	for i, child := range box.Children {
		elem, ok := child.(*Element)
		if !ok {
			t.Errorf("child %d: expected *Element, got %T", i, child)
			continue
		}
		if elem.Tag != "span" {
			t.Errorf("child %d: Tag = %q, want 'span'", i, elem.Tag)
		}
	}
}

func TestParser_ElementWithAttributes(t *testing.T) {
	type tc struct {
		input     string
		wantAttrs int
	}

	tests := map[string]tc{
		"no attributes": {
			input: `package x
templ Test() {
	<div></div>
}`,
			wantAttrs: 0,
		},
		"string attribute": {
			input: `package x
templ Test() {
	<span textAlign="center"></span>
}`,
			wantAttrs: 1,
		},
		"int attribute": {
			input: `package x
templ Test() {
	<div width=100></div>
}`,
			wantAttrs: 1,
		},
		"expression attribute": {
			input: `package x
templ Test() {
	<div direction={tui.Column}></div>
}`,
			wantAttrs: 1,
		},
		"multiple attributes": {
			input: `package x
templ Test() {
	<div width=100 height=50 direction={tui.Row}></div>
}`,
			wantAttrs: 3,
		},
		"boolean shorthand": {
			input: `package x
templ Test() {
	<input disabled></input>
}`,
			wantAttrs: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			p := NewParser(l)
			file, err := p.ParseFile()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			elem := file.Components[0].Body[0].(*Element)
			if len(elem.Attributes) != tt.wantAttrs {
				t.Errorf("len(Attributes) = %d, want %d", len(elem.Attributes), tt.wantAttrs)
			}
		})
	}
}

func TestParser_AttributeValues(t *testing.T) {
	input := `package x
templ Test() {
	<div
		strAttr="hello"
		intAttr=42
		floatAttr=3.14
		exprAttr={tui.Column}
		boolAttr=true
		shorthand
	></div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	elem := file.Components[0].Body[0].(*Element)

	if len(elem.Attributes) != 6 {
		t.Fatalf("expected 6 attributes, got %d", len(elem.Attributes))
	}

	// Check string attribute
	strAttr := elem.Attributes[0]
	if strAttr.Name != "strAttr" {
		t.Errorf("attr 0 name = %q, want 'strAttr'", strAttr.Name)
	}
	if str, ok := strAttr.Value.(*StringLit); !ok || str.Value != "hello" {
		t.Errorf("attr 0 value = %v, want StringLit{hello}", strAttr.Value)
	}

	// Check int attribute
	intAttr := elem.Attributes[1]
	if intAttr.Name != "intAttr" {
		t.Errorf("attr 1 name = %q, want 'intAttr'", intAttr.Name)
	}
	if num, ok := intAttr.Value.(*IntLit); !ok || num.Value != 42 {
		t.Errorf("attr 1 value = %v, want IntLit{42}", intAttr.Value)
	}

	// Check float attribute
	floatAttr := elem.Attributes[2]
	if floatAttr.Name != "floatAttr" {
		t.Errorf("attr 2 name = %q, want 'floatAttr'", floatAttr.Name)
	}
	if num, ok := floatAttr.Value.(*FloatLit); !ok || num.Value != 3.14 {
		t.Errorf("attr 2 value = %v, want FloatLit{3.14}", floatAttr.Value)
	}

	// Check expression attribute
	exprAttr := elem.Attributes[3]
	if exprAttr.Name != "exprAttr" {
		t.Errorf("attr 3 name = %q, want 'exprAttr'", exprAttr.Name)
	}
	if expr, ok := exprAttr.Value.(*GoExpr); !ok {
		t.Errorf("attr 3 value = %T, want *GoExpr", exprAttr.Value)
	} else if expr.Code != "tui.Column" {
		t.Errorf("attr 3 value code = %q, want 'tui.Column'", expr.Code)
	}

	// Check bool attribute
	boolAttr := elem.Attributes[4]
	if boolAttr.Name != "boolAttr" {
		t.Errorf("attr 4 name = %q, want 'boolAttr'", boolAttr.Name)
	}
	if b, ok := boolAttr.Value.(*BoolLit); !ok || b.Value != true {
		t.Errorf("attr 4 value = %v, want BoolLit{true}", boolAttr.Value)
	}

	// Check shorthand bool attribute
	shorthand := elem.Attributes[5]
	if shorthand.Name != "shorthand" {
		t.Errorf("attr 5 name = %q, want 'shorthand'", shorthand.Name)
	}
	if b, ok := shorthand.Value.(*BoolLit); !ok || b.Value != true {
		t.Errorf("attr 5 value = %v, want BoolLit{true}", shorthand.Value)
	}
}

func TestParser_NestedElements(t *testing.T) {
	input := `package x
templ Test() {
	<div>
		<div>
			<span>Deep</span>
		</div>
	</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outerBox := file.Components[0].Body[0].(*Element)
	if outerBox.Tag != "div" {
		t.Errorf("outer tag = %q, want 'div'", outerBox.Tag)
	}

	if len(outerBox.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(outerBox.Children))
	}

	innerBox := outerBox.Children[0].(*Element)
	if innerBox.Tag != "div" {
		t.Errorf("inner tag = %q, want 'div'", innerBox.Tag)
	}

	if len(innerBox.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(innerBox.Children))
	}

	text := innerBox.Children[0].(*Element)
	if text.Tag != "span" {
		t.Errorf("text tag = %q, want 'span'", text.Tag)
	}
}

func TestParser_RefAttribute(t *testing.T) {
	type tc struct {
		input     string
		wantRef   string
		wantTag   string
		wantAttrs int
	}

	tests := map[string]tc{
		"simple ref attribute": {
			input: `package x
templ Test() {
	<div ref={content}></div>
}`,
			wantRef:   "content",
			wantTag:   "div",
			wantAttrs: 0,
		},
		"ref attribute with attributes": {
			input: `package x
templ Test() {
	<span ref={title} class="bold">hello</span>
}`,
			wantRef:   "title",
			wantTag:   "span",
			wantAttrs: 1,
		},
		"ref attribute self-closing": {
			input: `package x
templ Test() {
	<div ref={spacer} />
}`,
			wantRef:   "spacer",
			wantTag:   "div",
			wantAttrs: 0,
		},
		"ref attribute with multiple attributes": {
			input: `package x
templ Test() {
	<div ref={content} width=100 height=50></div>
}`,
			wantRef:   "content",
			wantTag:   "div",
			wantAttrs: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			p := NewParser(l)
			file, err := p.ParseFile()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(file.Components) != 1 {
				t.Fatalf("expected 1 component, got %d", len(file.Components))
			}

			body := file.Components[0].Body
			if len(body) != 1 {
				t.Fatalf("expected 1 body node, got %d", len(body))
			}

			elem, ok := body[0].(*Element)
			if !ok {
				t.Fatalf("body[0]: expected *Element, got %T", body[0])
			}

			if elem.RefExpr == nil {
				t.Fatal("RefExpr should not be nil")
			}
			if elem.RefExpr.Code != tt.wantRef {
				t.Errorf("RefExpr.Code = %q, want %q", elem.RefExpr.Code, tt.wantRef)
			}
			if elem.Tag != tt.wantTag {
				t.Errorf("Tag = %q, want %q", elem.Tag, tt.wantTag)
			}
			if len(elem.Attributes) != tt.wantAttrs {
				t.Errorf("len(Attributes) = %d, want %d", len(elem.Attributes), tt.wantAttrs)
			}

			// ref should be removed from attributes
			for _, attr := range elem.Attributes {
				if attr.Name == "ref" {
					t.Error("ref attribute should be moved to RefExpr, not remain in Attributes")
				}
			}
		})
	}
}

func TestParser_RefAttributeWithKey(t *testing.T) {
	input := `package x
templ Test(items []Item) {
	<ul>
		@for _, item := range items {
			<li ref={items} key={item.ID}>{item.Name}</li>
		}
	</ul>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	ul := comp.Body[0].(*Element)
	forLoop := ul.Children[0].(*ForLoop)
	li := forLoop.Body[0].(*Element)

	if li.RefExpr == nil {
		t.Fatal("RefExpr should not be nil")
	}

	if li.RefExpr.Code != "items" {
		t.Errorf("RefExpr.Code = %q, want 'items'", li.RefExpr.Code)
	}

	if li.RefKey == nil {
		t.Fatal("RefKey should not be nil")
	}

	if li.RefKey.Code != "item.ID" {
		t.Errorf("RefKey.Code = %q, want 'item.ID'", li.RefKey.Code)
	}

	// ref and key should be removed from attributes
	for _, attr := range li.Attributes {
		if attr.Name == "ref" {
			t.Error("ref attribute should be moved to RefExpr, not remain in Attributes")
		}
		if attr.Name == "key" {
			t.Error("key attribute should be moved to RefKey, not remain in Attributes")
		}
	}
}

func TestParser_MultipleRefAttributes(t *testing.T) {
	input := `package x
templ Test() {
	<div>
		<div ref={header} height=3></div>
		<div ref={content}></div>
		<div ref={footer} height=3></div>
	</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	container := comp.Body[0].(*Element)

	if len(container.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(container.Children))
	}

	expectedRefs := []string{"header", "content", "footer"}
	for i, child := range container.Children {
		elem, ok := child.(*Element)
		if !ok {
			t.Fatalf("child %d: expected *Element, got %T", i, child)
		}
		if elem.RefExpr == nil {
			t.Errorf("child %d: RefExpr should not be nil", i)
			continue
		}
		if elem.RefExpr.Code != expectedRefs[i] {
			t.Errorf("child %d: RefExpr.Code = %q, want %q", i, elem.RefExpr.Code, expectedRefs[i])
		}
	}
}
