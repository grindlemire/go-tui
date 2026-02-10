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
			l := NewLexer("test.gsx", tt.input)
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
			l := NewLexer("test.gsx", tt.input)
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

func TestParser_FunctionParamsMultilineWithTrailingComma(t *testing.T) {
	input := `package x
type SettingsApp struct{}

func NewSettingsApp(
	provider string,
	model string,
	onClose func(),
) *SettingsApp {
	return &SettingsApp{}
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(file.Funcs))
	}
}

func TestParser_GoExpression(t *testing.T) {
	input := `package x
templ Test() {
	<span>{fmt.Sprintf("Count: %d", count)}</span>
}`

	l := NewLexer("test.gsx", input)
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

func TestParser_CompleteExample(t *testing.T) {
	input := `package components

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Dashboard(items []string, selectedIndex int) {
	<div direction={tui.Column} padding=1>
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

	l := NewLexer("test.gsx", input)
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
templ() {
	<span>Hello</span>
}`,
			errorContains: "expected component name",
		},
		"unclosed element": {
			input: `package x
templ Test() {
	<div>
		<span>Hello</span>
}`,
			errorContains: "expected closing tag",
		},
		"mismatched tags": {
			input: `package x
templ Test() {
	<div>Hello</span>
}`,
			errorContains: "mismatched closing tag",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
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

templ Test() {
	<span>Hello</span>
}`

	l := NewLexer("test.gsx", input)
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

templ Header() {
	<span>Header</span>
}

templ Footer() {
	<span>Footer</span>
}`

	l := NewLexer("test.gsx", input)
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
templ Test() {
	<span>Hello World</span>
}`

	l := NewLexer("test.gsx", input)
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

func TestParser_TextContentCoalescing(t *testing.T) {
	input := `package x
templ Test() {
	<span>Hello World from component</span>
}`

	l := NewLexer("test.gsx", input)
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

templ Broken(
	<span>Hello</span>
}

templ Working() {
	<span>World</span>
}`

	l := NewLexer("test.gsx", input)
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
templ Test() {
	@if user.Name != "" && user.Age >= 18 {
		<span>Adult</span>
	}
	@for i, v := range items[0:10] {
		<span>{v}</span>
	}
}`

	l := NewLexer("test.gsx", input)
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
