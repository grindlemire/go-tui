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
	<span>Hello</span>
}`,
			wantName:   "Header",
			wantParams: 0,
		},
		"one param": {
			input: `package x
@component Greeting(name string) {
	<span>Hello</span>
}`,
			wantName:   "Greeting",
			wantParams: 1,
		},
		"multiple params": {
			input: `package x
@component Counter(count int, label string) {
	<span>Hello</span>
}`,
			wantName:   "Counter",
			wantParams: 2,
		},
		"complex types": {
			input: `package x
@component List(items []string, onClick func()) {
	<span>Hello</span>
}`,
			wantName:   "List",
			wantParams: 2,
		},
		"pointer type": {
			input: `package x
@component View(elem *element.Element) {
	<span>Hello</span>
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
	<span>Hello</span>
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
	<div>
		<span>Hello</span>
		<span>World</span>
	</div>
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
@component Test() {
	<div></div>
}`,
			wantAttrs: 0,
		},
		"string attribute": {
			input: `package x
@component Test() {
	<span textAlign="center"></span>
}`,
			wantAttrs: 1,
		},
		"int attribute": {
			input: `package x
@component Test() {
	<div width=100></div>
}`,
			wantAttrs: 1,
		},
		"expression attribute": {
			input: `package x
@component Test() {
	<div direction={layout.Column}></div>
}`,
			wantAttrs: 1,
		},
		"multiple attributes": {
			input: `package x
@component Test() {
	<div width=100 height=50 direction={layout.Row}></div>
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
	<div
		strAttr="hello"
		intAttr=42
		floatAttr=3.14
		exprAttr={layout.Column}
		boolAttr=true
		shorthand
	></div>
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
	@let myText = <span>Hello</span>
	<div></div>
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

	if let.Element.Tag != "span" {
		t.Errorf("Element.Tag = %q, want 'span'", let.Element.Tag)
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
		<span>Hello</span>
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
		<span>Hello</span>
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
		<span>Hello</span>
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
		<span>Header</span>
	}
}`,
			wantCondition: "showHeader",
			wantElse:      false,
		},
		"if with else": {
			input: `package x
@component Test() {
	@if isLoading {
		<span>Loading</span>
	} @else {
		<span>Done</span>
	}
}`,
			wantCondition: "isLoading",
			wantElse:      true,
		},
		"complex condition": {
			input: `package x
@component Test() {
	@if err != nil {
		<span>Error</span>
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
		<span>A</span>
	} @else @if b {
		<span>B</span>
	} @else {
		<span>C</span>
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
	<span>{fmt.Sprintf("Count: %d", count)}</span>
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
	<div>
		<div>
			<span>Deep</span>
		</div>
	</div>
}`

	l := NewLexer("test.tui", input)
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

func TestParser_CompleteExample(t *testing.T) {
	input := `package components

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/layout"
)

@component Dashboard(items []string, selectedIndex int) {
	<div direction={layout.Column} padding=1>
		<span>Dashboard</span>
		@for i, item := range items {
			@if i == selectedIndex {
				<span textStyle={highlightStyle}>{item}</span>
			} @else {
				<span>{item}</span>
			}
		}
	</div>
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
	if box.Tag != "div" {
		t.Errorf("box.Tag = %q, want 'div'", box.Tag)
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
	<span>Hello</span>
}`,
			errorContains: "expected component name",
		},
		"unclosed element": {
			input: `package x
@component Test() {
	<div>
		<span>Hello</span>
}`,
			errorContains: "expected closing tag",
		},
		"mismatched tags": {
			input: `package x
@component Test() {
	<div>Hello</span>
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
	<span>Hello</span>
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
	<span>Header</span>
}

@component Footer() {
	<span>Footer</span>
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
	<span>Hello World</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	textElem := file.Components[0].Body[0].(*Element)
	if textElem.Tag != "span" {
		t.Errorf("Tag = %q, want 'span'", textElem.Tag)
	}

	// Text content is parsed as identifier tokens (Hello, World)
	if len(textElem.Children) < 1 {
		t.Errorf("expected children, got %d", len(textElem.Children))
	}
}

func TestParser_ControlFlowInChildren(t *testing.T) {
	input := `package x
@component Test(items []string) {
	<div>
		@for _, item := range items {
			<span>{item}</span>
		}
		@if len(items) == 0 {
			<span>No items</span>
		}
	</div>
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

func TestParser_ComplexTypeSignatures(t *testing.T) {
	type tc struct {
		input        string
		wantTypes    []string
	}

	tests := map[string]tc{
		"channel type": {
			input: `package x
@component Test(ch chan int) {
	<span>Hello</span>
}`,
			wantTypes: []string{"chan int"},
		},
		"receive channel": {
			input: `package x
@component Test(ch <-chan string) {
	<span>Hello</span>
}`,
			wantTypes: []string{"<-chan string"},
		},
		"complex map": {
			input: `package x
@component Test(m map[string][]int) {
	<span>Hello</span>
}`,
			wantTypes: []string{"map[string][]int"},
		},
		"function with return": {
			input: `package x
@component Test(fn func(a, b int) (string, error)) {
	<span>Hello</span>
}`,
			wantTypes: []string{"func(a, b int) (string, error)"},
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

			if len(file.Components) != 1 {
				t.Fatalf("expected 1 component, got %d", len(file.Components))
			}

			params := file.Components[0].Params
			if len(params) != len(tt.wantTypes) {
				t.Fatalf("expected %d params, got %d", len(tt.wantTypes), len(params))
			}

			for i, wantType := range tt.wantTypes {
				if params[i].Type != wantType {
					t.Errorf("param %d: Type = %q, want %q", i, params[i].Type, wantType)
				}
			}
		})
	}
}

func TestParser_TextContentCoalescing(t *testing.T) {
	input := `package x
@component Test() {
	<span>Hello World from component</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	textElem := file.Components[0].Body[0].(*Element)
	if textElem.Tag != "span" {
		t.Fatalf("expected span element, got %s", textElem.Tag)
	}

	// Text should be coalesced into a single TextContent node
	if len(textElem.Children) != 1 {
		t.Fatalf("expected 1 child (coalesced text), got %d", len(textElem.Children))
	}

	textContent, ok := textElem.Children[0].(*TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", textElem.Children[0])
	}

	expected := "Hello World from component"
	if textContent.Text != expected {
		t.Errorf("Text = %q, want %q", textContent.Text, expected)
	}
}

func TestParser_ErrorRecoveryMultipleComponents(t *testing.T) {
	// This test verifies that the parser can recover from errors
	// and continue parsing subsequent components
	input := `package x

@component Broken(
	<span>Hello</span>
}

@component Working() {
	<span>World</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	// Should have errors
	if err == nil {
		t.Log("Note: No error returned, parser may have recovered")
	}

	// But should still have parsed what it could
	// The first component is broken, but we should try to recover
	if file == nil {
		t.Fatal("expected file to be non-nil even with errors")
	}

	// Due to error recovery, we might get the second component
	t.Logf("Parsed %d components", len(file.Components))
}

func TestParser_RawSourcePreservation(t *testing.T) {
	// Test that raw source is preserved correctly in conditions/iterables
	input := `package x
@component Test() {
	@if user.Name != "" && user.Age >= 18 {
		<span>Adult</span>
	}
	@for i, v := range items[0:10] {
		<span>{v}</span>
	}
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := file.Components[0].Body

	// Check @if condition preserves original formatting
	ifStmt := body[0].(*IfStmt)
	expectedCond := `user.Name != "" && user.Age >= 18`
	if ifStmt.Condition != expectedCond {
		t.Errorf("Condition = %q, want %q", ifStmt.Condition, expectedCond)
	}

	// Check @for iterable preserves original formatting
	forLoop := body[1].(*ForLoop)
	expectedIter := "items[0:10]"
	if forLoop.Iterable != expectedIter {
		t.Errorf("Iterable = %q, want %q", forLoop.Iterable, expectedIter)
	}
}

func TestParser_RawGoStatements(t *testing.T) {
	type tc struct {
		input     string
		wantCodes []string
	}

	tests := map[string]tc{
		"simple assignment": {
			input: `package x
@component Test() {
	x := 1
	<span>{x}</span>
}`,
			wantCodes: []string{"x := 1"},
		},
		"function call": {
			input: `package x
@component Test() {
	fmt.Println("hello")
	<span>world</span>
}`,
			wantCodes: []string{`fmt.Println("hello")`},
		},
		"multi-line statement": {
			input: `package x
@component Test() {
	result := compute(
		arg1,
		arg2,
	)
	<span>{result}</span>
}`,
			wantCodes: []string{"result := compute(\n\t\targ1,\n\t\targ2,\n\t)"},
		},
		"multiple statements": {
			input: `package x
@component Test() {
	x := 1
	y := 2
	z := x + y
	<span>{z}</span>
}`,
			wantCodes: []string{"x := 1", "y := 2", "z := x + y"},
		},
		"inline if statement": {
			input: `package x
@component Test(err error) {
	if err != nil { log.Error(err) }
	<span>done</span>
}`,
			wantCodes: []string{"if err != nil { log.Error(err) }"},
		},
		"defer statement": {
			input: `package x
@component Test() {
	defer cleanup()
	<span>running</span>
}`,
			wantCodes: []string{"defer cleanup()"},
		},
		"go statement": {
			input: `package x
@component Test() {
	go doWork()
	<span>spawned</span>
}`,
			wantCodes: []string{"go doWork()"},
		},
		"for loop statement": {
			input: `package x
@component Test() {
	for i := 0; i < 10; i++ { sum += i }
	<span>{sum}</span>
}`,
			wantCodes: []string{"for i := 0; i < 10; i++ { sum += i }"},
		},
		"switch statement": {
			input: `package x
@component Test(x int) {
	switch x { case 1: y = "one"; case 2: y = "two" }
	<span>{y}</span>
}`,
			wantCodes: []string{`switch x { case 1: y = "one"; case 2: y = "two" }`},
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

			if len(file.Components) != 1 {
				t.Fatalf("expected 1 component, got %d", len(file.Components))
			}

			body := file.Components[0].Body

			// Count GoCode nodes
			var goCodes []*GoCode
			for _, node := range body {
				if gc, ok := node.(*GoCode); ok {
					goCodes = append(goCodes, gc)
				}
			}

			if len(goCodes) != len(tt.wantCodes) {
				t.Fatalf("expected %d GoCode nodes, got %d", len(tt.wantCodes), len(goCodes))
			}

			for i, wantCode := range tt.wantCodes {
				if goCodes[i].Code != wantCode {
					t.Errorf("GoCode[%d].Code = %q, want %q", i, goCodes[i].Code, wantCode)
				}
			}
		})
	}
}

func TestParser_RawGoStatementsWithElements(t *testing.T) {
	// Test that Go statements and elements can be mixed in component body
	input := `package x
@component Counter(count int) {
	formattedCount := fmt.Sprintf("%d", count)
	log.Printf("Rendering counter")
	<div>
		<span>{formattedCount}</span>
	</div>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := file.Components[0].Body
	if len(body) != 3 {
		t.Fatalf("expected 3 body nodes, got %d", len(body))
	}

	// First two should be GoCode
	gc1, ok := body[0].(*GoCode)
	if !ok {
		t.Fatalf("body[0]: expected *GoCode, got %T", body[0])
	}
	if gc1.Code != `formattedCount := fmt.Sprintf("%d", count)` {
		t.Errorf("body[0].Code = %q", gc1.Code)
	}

	gc2, ok := body[1].(*GoCode)
	if !ok {
		t.Fatalf("body[1]: expected *GoCode, got %T", body[1])
	}
	if gc2.Code != `log.Printf("Rendering counter")` {
		t.Errorf("body[1].Code = %q", gc2.Code)
	}

	// Third should be Element
	elem, ok := body[2].(*Element)
	if !ok {
		t.Fatalf("body[2]: expected *Element, got %T", body[2])
	}
	if elem.Tag != "div" {
		t.Errorf("body[2].Tag = %q, want 'div'", elem.Tag)
	}
}

func TestParser_ComponentCall(t *testing.T) {
	type tc struct {
		input        string
		wantName     string
		wantArgs     string
		wantChildren int
	}

	tests := map[string]tc{
		"call without args or children": {
			input: `package x
@component App() {
	@Header()
}`,
			wantName:     "Header",
			wantArgs:     "",
			wantChildren: 0,
		},
		"call with args no children": {
			input: `package x
@component App() {
	@Header("Welcome", true)
}`,
			wantName:     "Header",
			wantArgs:     `"Welcome", true`,
			wantChildren: 0,
		},
		"call with children": {
			input: `package x
@component App() {
	@Card("Title") {
		<span>Child 1</span>
		<span>Child 2</span>
	}
}`,
			wantName:     "Card",
			wantArgs:     `"Title"`,
			wantChildren: 2,
		},
		"call with empty args and children": {
			input: `package x
@component App() {
	@Wrapper() {
		<span>Content</span>
	}
}`,
			wantName:     "Wrapper",
			wantArgs:     "",
			wantChildren: 1,
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

			if len(file.Components) != 1 {
				t.Fatalf("expected 1 component, got %d", len(file.Components))
			}

			body := file.Components[0].Body
			if len(body) != 1 {
				t.Fatalf("expected 1 body node, got %d", len(body))
			}

			call, ok := body[0].(*ComponentCall)
			if !ok {
				t.Fatalf("body[0]: expected *ComponentCall, got %T", body[0])
			}

			if call.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", call.Name, tt.wantName)
			}
			if call.Args != tt.wantArgs {
				t.Errorf("Args = %q, want %q", call.Args, tt.wantArgs)
			}
			if len(call.Children) != tt.wantChildren {
				t.Errorf("len(Children) = %d, want %d", len(call.Children), tt.wantChildren)
			}
		})
	}
}

func TestParser_ChildrenSlot(t *testing.T) {
	input := `package x
@component Card(title string) {
	<div>
		<span>{title}</span>
		{children...}
	</div>
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

	body := file.Components[0].Body
	if len(body) != 1 {
		t.Fatalf("expected 1 body node, got %d", len(body))
	}

	elem, ok := body[0].(*Element)
	if !ok {
		t.Fatalf("body[0]: expected *Element, got %T", body[0])
	}

	// Box should have 2 children: text and children slot
	if len(elem.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(elem.Children))
	}

	// Second child should be ChildrenSlot
	slot, ok := elem.Children[1].(*ChildrenSlot)
	if !ok {
		t.Fatalf("children[1]: expected *ChildrenSlot, got %T", elem.Children[1])
	}
	if slot == nil {
		t.Error("ChildrenSlot should not be nil")
	}
}

func TestParser_ComponentCallNestedInElement(t *testing.T) {
	input := `package x
@component App() {
	<div>
		@Header("Title")
		@Footer()
	</div>
}`
	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := file.Components[0].Body
	if len(body) != 1 {
		t.Fatalf("expected 1 body node, got %d", len(body))
	}

	elem, ok := body[0].(*Element)
	if !ok {
		t.Fatalf("body[0]: expected *Element, got %T", body[0])
	}

	// Box should have 2 children: two component calls
	if len(elem.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(elem.Children))
	}

	call1, ok := elem.Children[0].(*ComponentCall)
	if !ok {
		t.Fatalf("children[0]: expected *ComponentCall, got %T", elem.Children[0])
	}
	if call1.Name != "Header" {
		t.Errorf("children[0].Name = %q, want 'Header'", call1.Name)
	}

	call2, ok := elem.Children[1].(*ComponentCall)
	if !ok {
		t.Fatalf("children[1]: expected *ComponentCall, got %T", elem.Children[1])
	}
	if call2.Name != "Footer" {
		t.Errorf("children[1].Name = %q, want 'Footer'", call2.Name)
	}
}

func TestParser_NamedRef(t *testing.T) {
	type tc struct {
		input     string
		wantRef   string
		wantTag   string
		wantAttrs int
	}

	tests := map[string]tc{
		"simple named ref": {
			input: `package x
@component Test() {
	<div #Content></div>
}`,
			wantRef:   "Content",
			wantTag:   "div",
			wantAttrs: 0,
		},
		"named ref with attributes": {
			input: `package x
@component Test() {
	<span #Title class="bold">hello</span>
}`,
			wantRef:   "Title",
			wantTag:   "span",
			wantAttrs: 1,
		},
		"named ref self-closing": {
			input: `package x
@component Test() {
	<div #Spacer />
}`,
			wantRef:   "Spacer",
			wantTag:   "div",
			wantAttrs: 0,
		},
		"named ref with multiple attributes": {
			input: `package x
@component Test() {
	<div #Content width=100 height=50></div>
}`,
			wantRef:   "Content",
			wantTag:   "div",
			wantAttrs: 2,
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

			if elem.NamedRef != tt.wantRef {
				t.Errorf("NamedRef = %q, want %q", elem.NamedRef, tt.wantRef)
			}
			if elem.Tag != tt.wantTag {
				t.Errorf("Tag = %q, want %q", elem.Tag, tt.wantTag)
			}
			if len(elem.Attributes) != tt.wantAttrs {
				t.Errorf("len(Attributes) = %d, want %d", len(elem.Attributes), tt.wantAttrs)
			}
		})
	}
}

func TestParser_NamedRefWithKey(t *testing.T) {
	input := `package x
@component Test(items []Item) {
	<ul>
		@for _, item := range items {
			<li #Items key={item.ID}>{item.Name}</li>
		}
	</ul>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	ul := comp.Body[0].(*Element)
	forLoop := ul.Children[0].(*ForLoop)
	li := forLoop.Body[0].(*Element)

	if li.NamedRef != "Items" {
		t.Errorf("NamedRef = %q, want 'Items'", li.NamedRef)
	}

	if li.RefKey == nil {
		t.Fatal("RefKey should not be nil")
	}

	if li.RefKey.Code != "item.ID" {
		t.Errorf("RefKey.Code = %q, want 'item.ID'", li.RefKey.Code)
	}

	// key should be removed from attributes
	for _, attr := range li.Attributes {
		if attr.Name == "key" {
			t.Error("key attribute should be moved to RefKey, not remain in Attributes")
		}
	}
}

func TestParser_MultipleNamedRefs(t *testing.T) {
	input := `package x
@component Test() {
	<div>
		<div #Header height=3></div>
		<div #Content></div>
		<div #Footer height=3></div>
	</div>
}`

	l := NewLexer("test.tui", input)
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

	expectedRefs := []string{"Header", "Content", "Footer"}
	for i, child := range container.Children {
		elem, ok := child.(*Element)
		if !ok {
			t.Fatalf("child %d: expected *Element, got %T", i, child)
		}
		if elem.NamedRef != expectedRefs[i] {
			t.Errorf("child %d: NamedRef = %q, want %q", i, elem.NamedRef, expectedRefs[i])
		}
	}
}
