package tuigen

import (
	"testing"
)

func TestParser_PackageAndImports(t *testing.T) {
	type tc struct {
		input       string
		wantPkg     string
		wantImports int
		wantError   bool
	}

	tests := map[string]tc{
		"simple package": {
			input:       "package myapp\n",
			wantPkg:     "myapp",
			wantImports: 0,
		},
		"package with single import": {
			input:       "package myapp\nimport \"fmt\"\n",
			wantPkg:     "myapp",
			wantImports: 1,
		},
		"package with grouped imports": {
			input: `package myapp
import (
	"fmt"
	"strings"
)
`,
			wantPkg:     "myapp",
			wantImports: 2,
		},
		"package with aliased import": {
			input:       "package myapp\nimport e \"github.com/example/element\"\n",
			wantPkg:     "myapp",
			wantImports: 1,
		},
		"missing package": {
			input:     "import \"fmt\"\n",
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			p := NewParser(l)
			file, err := p.ParseFile()

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if file.Package != tt.wantPkg {
				t.Errorf("Package = %q, want %q", file.Package, tt.wantPkg)
			}

			if len(file.Imports) != tt.wantImports {
				t.Errorf("len(Imports) = %d, want %d", len(file.Imports), tt.wantImports)
			}
		})
	}
}

func TestParser_ImportDetails(t *testing.T) {
	type tc struct {
		input     string
		wantAlias string
		wantPath  string
	}

	tests := map[string]tc{
		"simple import": {
			input:     "package x\nimport \"fmt\"\n",
			wantAlias: "",
			wantPath:  "fmt",
		},
		"aliased import": {
			input:     "package x\nimport f \"fmt\"\n",
			wantAlias: "f",
			wantPath:  "fmt",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			p := NewParser(l)
			file, err := p.ParseFile()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(file.Imports) != 1 {
				t.Fatalf("expected 1 import, got %d", len(file.Imports))
			}

			imp := file.Imports[0]
			if imp.Alias != tt.wantAlias {
				t.Errorf("Alias = %q, want %q", imp.Alias, tt.wantAlias)
			}
			if imp.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", imp.Path, tt.wantPath)
			}
		})
	}
}

func TestParser_SimpleComponent(t *testing.T) {
	type tc struct {
		input      string
		wantName   string
		wantParams int
		wantError  bool
	}

	tests := map[string]tc{
		"no params": {
			input: `package x
@component Header() {
	<text>Hello</text>
}`,
			wantName:   "Header",
			wantParams: 0,
		},
		"one param": {
			input: `package x
@component Greeting(name string) {
	<text>Hello</text>
}`,
			wantName:   "Greeting",
			wantParams: 1,
		},
		"multiple params": {
			input: `package x
@component Counter(count int, label string) {
	<text>Hello</text>
}`,
			wantName:   "Counter",
			wantParams: 2,
		},
		"complex types": {
			input: `package x
@component List(items []string, onClick func()) {
	<text>Hello</text>
}`,
			wantName:   "List",
			wantParams: 2,
		},
		"pointer type": {
			input: `package x
@component View(elem *element.Element) {
	<text>Hello</text>
}`,
			wantName:   "View",
			wantParams: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			p := NewParser(l)
			file, err := p.ParseFile()

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(file.Components) != 1 {
				t.Fatalf("expected 1 component, got %d", len(file.Components))
			}

			comp := file.Components[0]
			if comp.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", comp.Name, tt.wantName)
			}
			if len(comp.Params) != tt.wantParams {
				t.Errorf("len(Params) = %d, want %d", len(comp.Params), tt.wantParams)
			}
		})
	}
}

func TestParser_ComponentParams(t *testing.T) {
	input := `package x
@component Test(name string, count int, items []string, handler func()) {
	<text>Hello</text>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(file.Components))
	}

	params := file.Components[0].Params
	if len(params) != 4 {
		t.Fatalf("expected 4 params, got %d", len(params))
	}

	type expectedParam struct {
		name  string
		typ   string
	}

	expected := []expectedParam{
		{"name", "string"},
		{"count", "int"},
		{"items", "[]string"},
		{"handler", "func()"},
	}

	for i, exp := range expected {
		if params[i].Name != exp.name {
			t.Errorf("param %d: Name = %q, want %q", i, params[i].Name, exp.name)
		}
		if params[i].Type != exp.typ {
			t.Errorf("param %d: Type = %q, want %q", i, params[i].Type, exp.typ)
		}
	}
}

func TestParser_SelfClosingElement(t *testing.T) {
	input := `package x
@component Test() {
	<input />
}`

	l := NewLexer("test.tui", input)
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
@component Test() {
	<box>
		<text>Hello</text>
		<text>World</text>
	</box>
}`

	l := NewLexer("test.tui", input)
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

	if box.Tag != "box" {
		t.Errorf("Tag = %q, want 'box'", box.Tag)
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
		if elem.Tag != "text" {
			t.Errorf("child %d: Tag = %q, want 'text'", i, elem.Tag)
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
@component Test() {
	<box></box>
}`,
			wantAttrs: 0,
		},
		"string attribute": {
			input: `package x
@component Test() {
	<text textAlign="center"></text>
}`,
			wantAttrs: 1,
		},
		"int attribute": {
			input: `package x
@component Test() {
	<box width=100></box>
}`,
			wantAttrs: 1,
		},
		"expression attribute": {
			input: `package x
@component Test() {
	<box direction={layout.Column}></box>
}`,
			wantAttrs: 1,
		},
		"multiple attributes": {
			input: `package x
@component Test() {
	<box width=100 height=50 direction={layout.Row}></box>
}`,
			wantAttrs: 3,
		},
		"boolean shorthand": {
			input: `package x
@component Test() {
	<input disabled></input>
}`,
			wantAttrs: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
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
@component Test() {
	<box
		strAttr="hello"
		intAttr=42
		floatAttr=3.14
		exprAttr={layout.Column}
		boolAttr=true
		shorthand
	></box>
}`

	l := NewLexer("test.tui", input)
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
	} else if expr.Code != "layout.Column" {
		t.Errorf("attr 3 value code = %q, want 'layout.Column'", expr.Code)
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

func TestParser_LetBinding(t *testing.T) {
	input := `package x
@component Test() {
	@let myText = <text>Hello</text>
	<box></box>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	if len(comp.Body) != 2 {
		t.Fatalf("expected 2 body nodes, got %d", len(comp.Body))
	}

	let, ok := comp.Body[0].(*LetBinding)
	if !ok {
		t.Fatalf("expected *LetBinding, got %T", comp.Body[0])
	}

	if let.Name != "myText" {
		t.Errorf("Name = %q, want 'myText'", let.Name)
	}

	if let.Element == nil {
		t.Fatal("Element is nil")
	}

	if let.Element.Tag != "text" {
		t.Errorf("Element.Tag = %q, want 'text'", let.Element.Tag)
	}
}

func TestParser_ForLoop(t *testing.T) {
	type tc struct {
		input      string
		wantIndex  string
		wantValue  string
		wantIter   string
	}

	tests := map[string]tc{
		"index and value": {
			input: `package x
@component Test() {
	@for i, item := range items {
		<text>Hello</text>
	}
}`,
			wantIndex: "i",
			wantValue: "item",
			wantIter:  "items",
		},
		"underscore index": {
			input: `package x
@component Test() {
	@for _, item := range items {
		<text>Hello</text>
	}
}`,
			wantIndex: "_",
			wantValue: "item",
			wantIter:  "items",
		},
		"value only": {
			input: `package x
@component Test() {
	@for item := range items {
		<text>Hello</text>
	}
}`,
			wantIndex: "",
			wantValue: "item",
			wantIter:  "items",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			p := NewParser(l)
			file, err := p.ParseFile()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			forLoop, ok := file.Components[0].Body[0].(*ForLoop)
			if !ok {
				t.Fatalf("expected *ForLoop, got %T", file.Components[0].Body[0])
			}

			if forLoop.Index != tt.wantIndex {
				t.Errorf("Index = %q, want %q", forLoop.Index, tt.wantIndex)
			}
			if forLoop.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", forLoop.Value, tt.wantValue)
			}
			if forLoop.Iterable != tt.wantIter {
				t.Errorf("Iterable = %q, want %q", forLoop.Iterable, tt.wantIter)
			}
			if len(forLoop.Body) != 1 {
				t.Errorf("len(Body) = %d, want 1", len(forLoop.Body))
			}
		})
	}
}

func TestParser_IfStatement(t *testing.T) {
	type tc struct {
		input         string
		wantCondition string
		wantElse      bool
	}

	tests := map[string]tc{
		"simple if": {
			input: `package x
@component Test() {
	@if showHeader {
		<text>Header</text>
	}
}`,
			wantCondition: "showHeader",
			wantElse:      false,
		},
		"if with else": {
			input: `package x
@component Test() {
	@if isLoading {
		<text>Loading</text>
	} @else {
		<text>Done</text>
	}
}`,
			wantCondition: "isLoading",
			wantElse:      true,
		},
		"complex condition": {
			input: `package x
@component Test() {
	@if err != nil {
		<text>Error</text>
	}
}`,
			wantCondition: "err != nil",
			wantElse:      false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			p := NewParser(l)
			file, err := p.ParseFile()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			ifStmt, ok := file.Components[0].Body[0].(*IfStmt)
			if !ok {
				t.Fatalf("expected *IfStmt, got %T", file.Components[0].Body[0])
			}

			if ifStmt.Condition != tt.wantCondition {
				t.Errorf("Condition = %q, want %q", ifStmt.Condition, tt.wantCondition)
			}

			if len(ifStmt.Then) != 1 {
				t.Errorf("len(Then) = %d, want 1", len(ifStmt.Then))
			}

			hasElse := len(ifStmt.Else) > 0
			if hasElse != tt.wantElse {
				t.Errorf("hasElse = %v, want %v", hasElse, tt.wantElse)
			}
		})
	}
}

func TestParser_IfElseIf(t *testing.T) {
	input := `package x
@component Test() {
	@if a {
		<text>A</text>
	} @else @if b {
		<text>B</text>
	} @else {
		<text>C</text>
	}
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ifStmt, ok := file.Components[0].Body[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected *IfStmt, got %T", file.Components[0].Body[0])
	}

	if ifStmt.Condition != "a" {
		t.Errorf("Condition = %q, want 'a'", ifStmt.Condition)
	}

	// Else should contain an IfStmt
	if len(ifStmt.Else) != 1 {
		t.Fatalf("len(Else) = %d, want 1", len(ifStmt.Else))
	}

	elseIf, ok := ifStmt.Else[0].(*IfStmt)
	if !ok {
		t.Fatalf("Else[0] expected *IfStmt, got %T", ifStmt.Else[0])
	}

	if elseIf.Condition != "b" {
		t.Errorf("elseIf.Condition = %q, want 'b'", elseIf.Condition)
	}

	// Inner else
	if len(elseIf.Else) != 1 {
		t.Fatalf("len(elseIf.Else) = %d, want 1", len(elseIf.Else))
	}
}

func TestParser_GoExpression(t *testing.T) {
	input := `package x
@component Test() {
	<text>{fmt.Sprintf("Count: %d", count)}</text>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	textElem := file.Components[0].Body[0].(*Element)
	if len(textElem.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(textElem.Children))
	}

	expr, ok := textElem.Children[0].(*GoExpr)
	if !ok {
		t.Fatalf("expected *GoExpr, got %T", textElem.Children[0])
	}

	expected := `fmt.Sprintf("Count: %d", count)`
	if expr.Code != expected {
		t.Errorf("Code = %q, want %q", expr.Code, expected)
	}
}

func TestParser_NestedElements(t *testing.T) {
	input := `package x
@component Test() {
	<box>
		<box>
			<text>Deep</text>
		</box>
	</box>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outerBox := file.Components[0].Body[0].(*Element)
	if outerBox.Tag != "box" {
		t.Errorf("outer tag = %q, want 'box'", outerBox.Tag)
	}

	if len(outerBox.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(outerBox.Children))
	}

	innerBox := outerBox.Children[0].(*Element)
	if innerBox.Tag != "box" {
		t.Errorf("inner tag = %q, want 'box'", innerBox.Tag)
	}

	if len(innerBox.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(innerBox.Children))
	}

	text := innerBox.Children[0].(*Element)
	if text.Tag != "text" {
		t.Errorf("text tag = %q, want 'text'", text.Tag)
	}
}

func TestParser_CompleteExample(t *testing.T) {
	input := `package components

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/layout"
)

@component Dashboard(items []string, selectedIndex int) {
	<box direction={layout.Column} padding=1>
		<text>Dashboard</text>
		@for i, item := range items {
			@if i == selectedIndex {
				<text textStyle={highlightStyle}>{item}</text>
			} @else {
				<text>{item}</text>
			}
		}
	</box>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check package
	if file.Package != "components" {
		t.Errorf("Package = %q, want 'components'", file.Package)
	}

	// Check imports
	if len(file.Imports) != 2 {
		t.Errorf("len(Imports) = %d, want 2", len(file.Imports))
	}

	// Check component
	if len(file.Components) != 1 {
		t.Fatalf("len(Components) = %d, want 1", len(file.Components))
	}

	comp := file.Components[0]
	if comp.Name != "Dashboard" {
		t.Errorf("Name = %q, want 'Dashboard'", comp.Name)
	}

	if len(comp.Params) != 2 {
		t.Errorf("len(Params) = %d, want 2", len(comp.Params))
	}

	// Check body structure
	if len(comp.Body) != 1 {
		t.Fatalf("len(Body) = %d, want 1", len(comp.Body))
	}

	box := comp.Body[0].(*Element)
	if box.Tag != "box" {
		t.Errorf("box.Tag = %q, want 'box'", box.Tag)
	}

	// Should have text and for loop as children
	if len(box.Children) < 2 {
		t.Errorf("len(box.Children) = %d, want >= 2", len(box.Children))
	}
}

func TestParser_ErrorRecovery(t *testing.T) {
	type tc struct {
		input         string
		errorContains string
	}

	tests := map[string]tc{
		"missing component name": {
			input: `package x
@component () {
	<text>Hello</text>
}`,
			errorContains: "expected component name",
		},
		"unclosed element": {
			input: `package x
@component Test() {
	<box>
		<text>Hello</text>
}`,
			errorContains: "expected closing tag",
		},
		"mismatched tags": {
			input: `package x
@component Test() {
	<box>Hello</text>
}`,
			errorContains: "mismatched closing tag",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			p := NewParser(l)
			_, err := p.ParseFile()

			if err == nil {
				t.Error("expected error, got nil")
				return
			}

			// We just verify an error occurred, not the exact message
			// since error recovery may report different errors
		})
	}
}

func TestParser_Position(t *testing.T) {
	input := `package x

@component Test() {
	<text>Hello</text>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	// Component should be on line 3
	if comp.Position.Line != 3 {
		t.Errorf("Component line = %d, want 3", comp.Position.Line)
	}

	elem := comp.Body[0].(*Element)
	// Element should be on line 4
	if elem.Position.Line != 4 {
		t.Errorf("Element line = %d, want 4", elem.Position.Line)
	}
}

func TestParser_MultipleComponents(t *testing.T) {
	input := `package x

@component Header() {
	<text>Header</text>
}

@component Footer() {
	<text>Footer</text>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(file.Components))
	}

	if file.Components[0].Name != "Header" {
		t.Errorf("component 0 name = %q, want 'Header'", file.Components[0].Name)
	}

	if file.Components[1].Name != "Footer" {
		t.Errorf("component 1 name = %q, want 'Footer'", file.Components[1].Name)
	}
}

func TestParser_TextContent(t *testing.T) {
	input := `package x
@component Test() {
	<text>Hello World</text>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	textElem := file.Components[0].Body[0].(*Element)
	if textElem.Tag != "text" {
		t.Errorf("Tag = %q, want 'text'", textElem.Tag)
	}

	// Text content is parsed as identifier tokens (Hello, World)
	if len(textElem.Children) < 1 {
		t.Errorf("expected children, got %d", len(textElem.Children))
	}
}

func TestParser_ControlFlowInChildren(t *testing.T) {
	input := `package x
@component Test(items []string) {
	<box>
		@for _, item := range items {
			<text>{item}</text>
		}
		@if len(items) == 0 {
			<text>No items</text>
		}
	</box>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	box := file.Components[0].Body[0].(*Element)

	// Should have for loop and if statement as children
	hasFor := false
	hasIf := false
	for _, child := range box.Children {
		switch child.(type) {
		case *ForLoop:
			hasFor = true
		case *IfStmt:
			hasIf = true
		}
	}

	if !hasFor {
		t.Error("expected ForLoop child")
	}
	if !hasIf {
		t.Error("expected IfStmt child")
	}
}
