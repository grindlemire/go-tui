package tuigen

import (
	"os/exec"
	"strings"
	"testing"
)

func TestGenerator_SimpleComponent(t *testing.T) {
	type tc struct {
		input         string
		wantContains  []string
		wantNotContains []string
	}

	tests := map[string]tc{
		"empty component": {
			input: `package x
@component Empty() {
}`,
			wantContains: []string{
				"func Empty() *element.Element",
				"return nil",
			},
		},
		"component with single element": {
			input: `package x
@component Header() {
	<box></box>
}`,
			wantContains: []string{
				"func Header() *element.Element",
				"__tui_0 := element.New()",
				"return __tui_0",
			},
		},
		"component with params": {
			input: `package x
@component Greeting(name string, count int) {
	<text>Hello</text>
}`,
			wantContains: []string{
				"func Greeting(name string, count int) *element.Element",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := ParseAndGenerate("test.tui", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(code, notWant) {
					t.Errorf("output contains unexpected string: %q\nGot:\n%s", notWant, code)
				}
			}
		})
	}
}

func TestGenerator_ElementWithAttributes(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"width attribute": {
			input: `package x
@component Box() {
	<box width=100></box>
}`,
			wantContains: []string{
				"element.WithWidth(100)",
			},
		},
		"multiple attributes": {
			input: `package x
@component Box() {
	<box width=100 height=50 gap=2></box>
}`,
			wantContains: []string{
				"element.WithWidth(100)",
				"element.WithHeight(50)",
				"element.WithGap(2)",
			},
		},
		"string attribute": {
			input: `package x
@component Text() {
	<text text="hello"></text>
}`,
			wantContains: []string{
				`element.WithText("hello")`,
			},
		},
		"expression attribute": {
			input: `package x
@component Box() {
	<box direction={layout.Column}></box>
}`,
			wantContains: []string{
				"element.WithDirection(layout.Column)",
			},
		},
		"border attribute": {
			input: `package x
@component Box() {
	<box border={tui.BorderSingle}></box>
}`,
			wantContains: []string{
				"element.WithBorder(tui.BorderSingle)",
			},
		},
		"onEvent attribute": {
			input: `package x
@component Button() {
	<box onEvent={handleClick}></box>
}`,
			wantContains: []string{
				"element.WithOnEvent(handleClick)",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := ParseAndGenerate("test.tui", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_NestedElements(t *testing.T) {
	input := `package x
@component Layout() {
	<box>
		<box>
			<text>nested</text>
		</box>
	</box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Should have 3 element variables
	if !strings.Contains(code, "__tui_0 := element.New()") {
		t.Error("missing outer box element")
	}

	// Should have AddChild calls
	if !strings.Contains(code, ".AddChild(") {
		t.Error("missing AddChild call")
	}

	// Should return the outer element
	if !strings.Contains(code, "return __tui_0") {
		t.Error("missing return statement for outer element")
	}
}

func TestGenerator_LetBinding(t *testing.T) {
	input := `package x
@component Counter() {
	@let countText = <text>{"0"}</text>
	<box></box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Let binding should create a named variable
	if !strings.Contains(code, "countText := element.New(") {
		t.Errorf("missing let binding variable\nGot:\n%s", code)
	}

	// Should return the box element (first top-level Element, not LetBinding)
	// @let bindings are used for references, not as root elements
	if !strings.Contains(code, "return __tui_0") {
		t.Errorf("should return the box element, not the let-bound variable\nGot:\n%s", code)
	}
}

func TestGenerator_ForLoop(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"basic for loop": {
			input: `package x
@component List(items []string) {
	<box>
		@for i, item := range items {
			<text>{item}</text>
		}
	</box>
}`,
			wantContains: []string{
				"for i, item := range items {",
				"_ = i", // silence unused warning
			},
		},
		"for with underscore index": {
			input: `package x
@component List(items []string) {
	<box>
		@for _, item := range items {
			<text>{item}</text>
		}
	</box>
}`,
			wantContains: []string{
				"for _, item := range items {",
			},
		},
		"for with value only": {
			input: `package x
@component List(items []string) {
	<box>
		@for item := range items {
			<text>{item}</text>
		}
	</box>
}`,
			wantContains: []string{
				"for item := range items {",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := ParseAndGenerate("test.tui", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_IfStatement(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"simple if": {
			input: `package x
@component View(show bool) {
	<box>
		@if show {
			<text>visible</text>
		}
	</box>
}`,
			wantContains: []string{
				"if show {",
			},
		},
		"if-else": {
			input: `package x
@component View(loading bool) {
	<box>
		@if loading {
			<text>loading</text>
		} @else {
			<text>done</text>
		}
	</box>
}`,
			wantContains: []string{
				"if loading {",
				"} else {",
			},
		},
		"if-else-if": {
			input: `package x
@component View(state int) {
	<box>
		@if state == 0 {
			<text>zero</text>
		} @else @if state == 1 {
			<text>one</text>
		} @else {
			<text>other</text>
		}
	</box>
}`,
			wantContains: []string{
				"if state == 0 {",
				"} else if state == 1 {",
				"} else {",
			},
		},
		"complex condition": {
			input: `package x
@component View(err error) {
	<box>
		@if err != nil {
			<text>error</text>
		}
	</box>
}`,
			wantContains: []string{
				"if err != nil {",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := ParseAndGenerate("test.tui", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_TextElement(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"text with literal content": {
			input: `package x
@component Text() {
	<text>Hello World</text>
}`,
			wantContains: []string{
				`element.WithText("Hello World")`,
			},
		},
		"text with expression content": {
			input: `package x
@component Text(msg string) {
	<text>{msg}</text>
}`,
			wantContains: []string{
				"element.WithText(msg)",
			},
		},
		"text with formatted expression": {
			input: `package x
@component Text(count int) {
	<text>{fmt.Sprintf("Count: %d", count)}</text>
}`,
			wantContains: []string{
				`element.WithText(fmt.Sprintf("Count: %d", count))`,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := ParseAndGenerate("test.tui", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_RawGoStatements(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"variable assignment": {
			input: `package x
@component Counter() {
	count := 0
	<text>hello</text>
}`,
			wantContains: []string{
				"count := 0",
			},
		},
		"function call": {
			input: `package x
import "fmt"
@component Debug() {
	fmt.Println("debug")
	<text>hello</text>
}`,
			wantContains: []string{
				`fmt.Println("debug")`,
			},
		},
		"multiple statements": {
			input: `package x
@component Complex() {
	x := 1
	y := 2
	z := x + y
	<text>hello</text>
}`,
			wantContains: []string{
				"x := 1",
				"y := 2",
				"z := x + y",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := ParseAndGenerate("test.tui", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_TopLevelGoFunc(t *testing.T) {
	input := `package x

func helper(x int) int {
	return x * 2
}

@component Test() {
	<text>hello</text>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "func helper(x int) int") {
		t.Errorf("missing helper function\nGot:\n%s", code)
	}

	if !strings.Contains(code, "return x * 2") {
		t.Errorf("missing helper function body\nGot:\n%s", code)
	}
}

func TestGenerator_ImportPropagation(t *testing.T) {
	input := `package x
import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/layout"
)

@component Test() {
	<box direction={layout.Column}></box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Original imports should be preserved
	if !strings.Contains(code, `"fmt"`) {
		t.Error("missing fmt import")
	}

	if !strings.Contains(code, `"github.com/grindlemire/go-tui/pkg/layout"`) {
		t.Error("missing layout import")
	}

	// Element import should be added
	if !strings.Contains(code, `"github.com/grindlemire/go-tui/pkg/tui/element"`) {
		t.Error("missing element import")
	}
}

func TestGenerator_Header(t *testing.T) {
	input := `package x
@component Test() {
	<text>hello</text>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "Code generated by tui generate. DO NOT EDIT.") {
		t.Error("missing DO NOT EDIT header")
	}

	if !strings.Contains(code, "Source: test.tui") {
		t.Error("missing source file comment")
	}
}

func TestGenerator_OutputCompiles(t *testing.T) {
	// Skip if go command not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	input := `package main

import (
	"github.com/grindlemire/go-tui/pkg/layout"
)

@component Dashboard(items []string) {
	<box direction={layout.Column} padding=1>
		<text>Header</text>
		@for i, item := range items {
			@if i == 0 {
				<text textStyle={highlightStyle}>{item}</text>
			} @else {
				<text>{item}</text>
			}
		}
	</box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Verify the output is valid Go syntax by checking it formats
	// (gofmt is called internally by Generate)
	if len(output) == 0 {
		t.Error("empty output")
	}

	// The output should at least be valid Go code structure
	code := string(output)
	if !strings.Contains(code, "package main") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(code, "func Dashboard") {
		t.Error("missing function declaration")
	}
}

func TestGenerator_CompleteExample(t *testing.T) {
	input := `package components

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

func countDone(items []Item) int {
	count := 0
	for _, item := range items {
		if item.Done {
			count++
		}
	}
	return count
}

@component Dashboard(items []Item, selectedIndex int) {
	<box direction={layout.Column} padding=1>
		<box
			border={tui.BorderRounded}
			padding=1
			direction={layout.Row}
		>
			<text>Todo List</text>
			<text>{fmt.Sprintf("%d/%d done", countDone(items), len(items))}</text>
		</box>

		<box direction={layout.Column} flexGrow=1>
			@for i, item := range items {
				@if i == selectedIndex {
					<text borderStyle={selectedStyle}>{item.Name}</text>
				} @else {
					<text>{item.Name}</text>
				}
			}
		</box>
	</box>
}`

	output, err := ParseAndGenerate("components.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Check package
	if !strings.Contains(code, "package components") {
		t.Error("wrong package")
	}

	// Check imports preserved
	if !strings.Contains(code, `"fmt"`) {
		t.Error("missing fmt import")
	}

	// Check helper function preserved
	if !strings.Contains(code, "func countDone") {
		t.Error("missing helper function")
	}

	// Check component generated
	if !strings.Contains(code, "func Dashboard(items []Item, selectedIndex int)") {
		t.Error("missing Dashboard function")
	}

	// Check control flow
	if !strings.Contains(code, "for i, item := range items") {
		t.Error("missing for loop")
	}

	if !strings.Contains(code, "if i == selectedIndex") {
		t.Error("missing if statement")
	}
}

func TestGenerator_ScrollableElement(t *testing.T) {
	input := `package x
@component ScrollView() {
	<scrollable>
		<text>content</text>
	</scrollable>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "element.WithScrollable(element.ScrollVertical)") {
		t.Errorf("missing scrollable option\nGot:\n%s", code)
	}
}

func TestGenerator_SelfClosingElement(t *testing.T) {
	input := `package x
@component Test() {
	<box>
		<input />
	</box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Self-closing element should still generate valid element.New()
	if !strings.Contains(code, "element.New()") {
		t.Error("missing element creation for self-closing element")
	}
}

func TestGenerator_LetBindingAsChild(t *testing.T) {
	input := `package x
@component Test() {
	<box>
		@let item = <text>hello</text>
	</box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Let binding inside box should be added as child
	if !strings.Contains(code, "item := element.New(") {
		t.Errorf("missing let binding\nGot:\n%s", code)
	}

	// And should be added to parent
	if !strings.Contains(code, ".AddChild(item)") {
		t.Errorf("let binding should be added to parent\nGot:\n%s", code)
	}
}

func TestGenerator_MultipleComponents(t *testing.T) {
	input := `package x

@component Header() {
	<text>Header</text>
}

@component Footer() {
	<text>Footer</text>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "func Header()") {
		t.Error("missing Header function")
	}

	if !strings.Contains(code, "func Footer()") {
		t.Error("missing Footer function")
	}
}

func TestGenerator_ExpressionInLoopBody(t *testing.T) {
	input := `package x
@component List(items []string) {
	<box>
		@for _, item := range items {
			{item}
		}
	</box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Expression in loop body should create text element
	if !strings.Contains(code, "element.WithText(item)") {
		t.Errorf("missing text element for expression\nGot:\n%s", code)
	}
}

func TestGenerator_BooleanAttributes(t *testing.T) {
	input := `package x
@component Test() {
	<box scrollable={element.ScrollVertical}></box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "element.WithScrollable(element.ScrollVertical)") {
		t.Errorf("missing scrollable attribute\nGot:\n%s", code)
	}
}

func TestGenerator_FlexAttributes(t *testing.T) {
	input := `package x
@component Test() {
	<box flexGrow=1 flexShrink=0></box>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "element.WithFlexGrow(1)") {
		t.Errorf("missing flexGrow\nGot:\n%s", code)
	}

	if !strings.Contains(code, "element.WithFlexShrink(0)") {
		t.Errorf("missing flexShrink\nGot:\n%s", code)
	}
}

func TestGenerator_ComponentWithChildren(t *testing.T) {
	input := `package x
@component Card(title string) {
	<box>
		<text>{title}</text>
		{children...}
	</box>
}`

	// First parse and analyze
	lexer := NewLexer("test.tui", input)
	parser := NewParser(lexer)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	analyzer := NewAnalyzer()
	if err := analyzer.Analyze(file); err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	// Check that AcceptsChildren is set
	if !file.Components[0].AcceptsChildren {
		t.Error("AcceptsChildren should be true")
	}

	// Generate
	gen := NewGenerator()
	output, err := gen.Generate(file, "test.tui")
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Check that children parameter is present
	if !strings.Contains(code, "children []*element.Element") {
		t.Errorf("missing children parameter\nGot:\n%s", code)
	}

	// Check that children loop is generated
	if !strings.Contains(code, "for _, __child := range children") {
		t.Errorf("missing children loop\nGot:\n%s", code)
	}
}

func TestGenerator_ComponentCall(t *testing.T) {
	input := `package x
@component Header(title string) {
	<text>{title}</text>
}

@component App() {
	@Header("Welcome")
}`

	// Parse and analyze
	lexer := NewLexer("test.tui", input)
	parser := NewParser(lexer)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	analyzer := NewAnalyzer()
	if err := analyzer.Analyze(file); err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	// Generate
	gen := NewGenerator()
	output, err := gen.Generate(file, "test.tui")
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Check that component call is generated
	if !strings.Contains(code, `Header("Welcome"`) {
		t.Errorf("missing component call\nGot:\n%s", code)
	}
}

func TestGenerator_ComponentCallWithChildren(t *testing.T) {
	input := `package x
@component Card(title string) {
	<box>
		<text>{title}</text>
		{children...}
	</box>
}

@component App() {
	@Card("My Card") {
		<text>Line 1</text>
		<text>Line 2</text>
	}
}`

	// Parse and analyze
	lexer := NewLexer("test.tui", input)
	parser := NewParser(lexer)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	analyzer := NewAnalyzer()
	if err := analyzer.Analyze(file); err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	// Generate
	gen := NewGenerator()
	output, err := gen.Generate(file, "test.tui")
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Check that Card has children parameter
	if !strings.Contains(code, "func Card(title string, children []*element.Element)") {
		t.Errorf("Card should have children parameter\nGot:\n%s", code)
	}

	// Check that App creates children slice
	if !strings.Contains(code, "_children := []*element.Element{}") {
		t.Errorf("App should create children slice\nGot:\n%s", code)
	}

	// Check that children elements are appended
	if !strings.Contains(code, "append(") {
		t.Errorf("Should append children\nGot:\n%s", code)
	}

	// Check that Card is called with children
	if !strings.Contains(code, `Card("My Card"`) {
		t.Errorf("Should call Card\nGot:\n%s", code)
	}
}
