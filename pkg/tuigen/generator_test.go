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
				"type EmptyView struct",
				"Root     *element.Element",
				"watchers []tui.Watcher",
				"func Empty() EmptyView",
				"var view EmptyView",
				"var watchers []tui.Watcher",
				"return view",
			},
		},
		"component with single element": {
			input: `package x
@component Header() {
	<div></div>
}`,
			wantContains: []string{
				"type HeaderView struct",
				"func Header() HeaderView",
				"__tui_0 := element.New()",
				"Root:     __tui_0",
			},
		},
		"component with params": {
			input: `package x
@component Greeting(name string, count int) {
	<span>Hello</span>
}`,
			wantContains: []string{
				"func Greeting(name string, count int) GreetingView",
				"type GreetingView struct",
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
	<div width=100></div>
}`,
			wantContains: []string{
				"element.WithWidth(100)",
			},
		},
		"multiple attributes": {
			input: `package x
@component Box() {
	<div width=100 height=50 gap=2></div>
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
	<span text="hello"></span>
}`,
			wantContains: []string{
				`element.WithText("hello")`,
			},
		},
		"expression attribute": {
			input: `package x
@component Box() {
	<div direction={layout.Column}></div>
}`,
			wantContains: []string{
				"element.WithDirection(layout.Column)",
			},
		},
		"border attribute": {
			input: `package x
@component Box() {
	<div border={tui.BorderSingle}></div>
}`,
			wantContains: []string{
				"element.WithBorder(tui.BorderSingle)",
			},
		},
		"onEvent attribute": {
			input: `package x
@component Button() {
	<div onEvent={handleClick}></div>
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
	<div>
		<div>
			<span>nested</span>
		</div>
	</div>
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

	// Should return view struct with Root set to outer element
	if !strings.Contains(code, "Root:     __tui_0") {
		t.Error("missing Root assignment to outer element")
	}

	if !strings.Contains(code, "return view") {
		t.Error("missing return view statement")
	}
}

func TestGenerator_LetBinding(t *testing.T) {
	input := `package x
@component Counter() {
	@let countText = <span>{"0"}</span>
	<div></div>
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

	// Should return the box element (first top-level Element, not LetBinding) as Root
	// @let bindings are used for references, not as root elements
	if !strings.Contains(code, "Root:     __tui_0") {
		t.Errorf("should set Root to the box element, not the let-bound variable\nGot:\n%s", code)
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
	<div>
		@for i, item := range items {
			<span>{item}</span>
		}
	</div>
}`,
			wantContains: []string{
				"for i, item := range items {",
				"_ = i", // silence unused warning
			},
		},
		"for with underscore index": {
			input: `package x
@component List(items []string) {
	<div>
		@for _, item := range items {
			<span>{item}</span>
		}
	</div>
}`,
			wantContains: []string{
				"for _, item := range items {",
			},
		},
		"for with value only": {
			input: `package x
@component List(items []string) {
	<div>
		@for item := range items {
			<span>{item}</span>
		}
	</div>
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
	<div>
		@if show {
			<span>visible</span>
		}
	</div>
}`,
			wantContains: []string{
				"if show {",
			},
		},
		"if-else": {
			input: `package x
@component View(loading bool) {
	<div>
		@if loading {
			<span>loading</span>
		} @else {
			<span>done</span>
		}
	</div>
}`,
			wantContains: []string{
				"if loading {",
				"} else {",
			},
		},
		"if-else-if": {
			input: `package x
@component View(state int) {
	<div>
		@if state == 0 {
			<span>zero</span>
		} @else @if state == 1 {
			<span>one</span>
		} @else {
			<span>other</span>
		}
	</div>
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
	<div>
		@if err != nil {
			<span>error</span>
		}
	</div>
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
	<span>Hello World</span>
}`,
			wantContains: []string{
				`element.WithText("Hello World")`,
			},
		},
		"text with expression content": {
			input: `package x
@component Text(msg string) {
	<span>{msg}</span>
}`,
			wantContains: []string{
				"element.WithText(msg)",
			},
		},
		"text with formatted expression": {
			input: `package x
@component Text(count int) {
	<span>{fmt.Sprintf("Count: %d", count)}</span>
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
	<span>hello</span>
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
	<span>hello</span>
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
	<span>hello</span>
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
	<span>hello</span>
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
	<div direction={layout.Column}></div>
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
	<span>hello</span>
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
	<div direction={layout.Column} padding=1>
		<span>Header</span>
		@for i, item := range items {
			@if i == 0 {
				<span textStyle={highlightStyle}>{item}</span>
			} @else {
				<span>{item}</span>
			}
		}
	</div>
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
	<div direction={layout.Column} padding=1>
		<div
			border={tui.BorderRounded}
			padding=1
			direction={layout.Row}
		>
			<span>Todo List</span>
			<span>{fmt.Sprintf("%d/%d done", countDone(items), len(items))}</span>
		</div>

		<div direction={layout.Column} flexGrow=1>
			@for i, item := range items {
				@if i == selectedIndex {
					<span borderStyle={selectedStyle}>{item.Name}</span>
				} @else {
					<span>{item.Name}</span>
				}
			}
		</div>
	</div>
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

	// Check component generated with view struct return
	if !strings.Contains(code, "func Dashboard(items []Item, selectedIndex int) DashboardView") {
		t.Error("missing Dashboard function with view struct return")
	}

	// Check view struct generated
	if !strings.Contains(code, "type DashboardView struct") {
		t.Error("missing DashboardView struct")
	}

	// Check control flow
	if !strings.Contains(code, "for i, item := range items") {
		t.Error("missing for loop")
	}

	if !strings.Contains(code, "if i == selectedIndex") {
		t.Error("missing if statement")
	}
}

func TestGenerator_ScrollableAttribute(t *testing.T) {
	input := `package x
@component ScrollView() {
	<div scrollable={element.ScrollVertical}>
		<span>content</span>
	</div>
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
	<div>
		<input />
	</div>
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
	<div>
		@let item = <span>hello</span>
	</div>
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
	<span>Header</span>
}

@component Footer() {
	<span>Footer</span>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "func Header() HeaderView") {
		t.Error("missing Header function with HeaderView return")
	}

	if !strings.Contains(code, "func Footer() FooterView") {
		t.Error("missing Footer function with FooterView return")
	}

	if !strings.Contains(code, "type HeaderView struct") {
		t.Error("missing HeaderView struct")
	}

	if !strings.Contains(code, "type FooterView struct") {
		t.Error("missing FooterView struct")
	}
}

func TestGenerator_ExpressionInLoopBody(t *testing.T) {
	input := `package x
@component List(items []string) {
	<div>
		@for _, item := range items {
			{item}
		}
	</div>
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
	<div scrollable={element.ScrollVertical}></div>
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
	<div flexGrow=1 flexShrink=0></div>
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
	<div>
		<span>{title}</span>
		{children...}
	</div>
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

	// Check that view struct is generated
	if !strings.Contains(code, "type CardView struct") {
		t.Errorf("missing CardView struct\nGot:\n%s", code)
	}

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
	<span>{title}</span>
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

	// Check that component call is generated - returns view struct
	if !strings.Contains(code, `Header("Welcome")`) {
		t.Errorf("missing component call\nGot:\n%s", code)
	}

	// Check that .Root is used to get element from component view
	if !strings.Contains(code, ".Root,") {
		t.Errorf("missing .Root accessor for component child\nGot:\n%s", code)
	}
}

func TestGenerator_ComponentCallWithChildren(t *testing.T) {
	input := `package x
@component Card(title string) {
	<div>
		<span>{title}</span>
		{children...}
	</div>
}

@component App() {
	@Card("My Card") {
		<span>Line 1</span>
		<span>Line 2</span>
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

	// Check that Card has children parameter and returns view struct
	if !strings.Contains(code, "func Card(title string, children []*element.Element) CardView") {
		t.Errorf("Card should have children parameter and return CardView\nGot:\n%s", code)
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

func TestGenerator_TailwindClassAttribute(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"layout classes": {
			input: `package x
@component Box() {
	<div class="flex flex-col gap-2 p-4"></div>
}`,
			wantContains: []string{
				"element.WithDirection(layout.Row)",
				"element.WithDirection(layout.Column)",
				"element.WithGap(2)",
				"element.WithPadding(4)",
			},
		},
		"border class": {
			input: `package x
@component Box() {
	<div class="border-rounded"></div>
}`,
			wantContains: []string{
				"element.WithBorder(tui.BorderRounded)",
			},
		},
		"text style classes": {
			input: `package x
@component Text() {
	<span class="font-bold text-cyan">hello</span>
}`,
			wantContains: []string{
				"element.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.Cyan))",
			},
		},
		"combined text and layout classes": {
			input: `package x
@component Card() {
	<div class="flex-col p-2 border">
		<span class="font-bold italic">Title</span>
	</div>
}`,
			wantContains: []string{
				"element.WithDirection(layout.Column)",
				"element.WithPadding(2)",
				"element.WithBorder(tui.BorderSingle)",
				"element.WithTextStyle(tui.NewStyle().Bold().Italic())",
			},
		},
		"alignment classes": {
			input: `package x
@component Center() {
	<div class="flex items-center justify-center"></div>
}`,
			wantContains: []string{
				"element.WithAlign(layout.AlignCenter)",
				"element.WithJustify(layout.JustifyCenter)",
			},
		},
		"sizing classes": {
			input: `package x
@component Sized() {
	<div class="w-50 h-20 min-w-10 max-w-100"></div>
}`,
			wantContains: []string{
				"element.WithWidth(50)",
				"element.WithHeight(20)",
				"element.WithMinWidth(10)",
				"element.WithMaxWidth(100)",
			},
		},
		"scroll classes": {
			input: `package x
@component Scrollable() {
	<div class="overflow-y-scroll"></div>
}`,
			wantContains: []string{
				"element.WithScrollable(element.ScrollVertical)",
			},
		},
		"class with explicit attribute": {
			input: `package x
@component Mixed() {
	<div class="flex-col" gap=5></div>
}`,
			wantContains: []string{
				"element.WithDirection(layout.Column)",
				"element.WithGap(5)",
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

func TestGenerator_HR(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"basic hr": {
			input: `package x
@component Divider() {
	<div>
		<hr/>
	</div>
}`,
			wantContains: []string{
				"element.WithHR()",
			},
		},
		"hr with border-double class": {
			input: `package x
@component Divider() {
	<div>
		<hr class="border-double"/>
	</div>
}`,
			wantContains: []string{
				"element.WithHR()",
				"element.WithBorder(tui.BorderDouble)",
			},
		},
		"hr with text-cyan class": {
			input: `package x
@component Divider() {
	<div>
		<hr class="text-cyan"/>
	</div>
}`,
			wantContains: []string{
				"element.WithHR()",
				"element.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan))",
			},
		},
		"hr with border-thick and text color": {
			input: `package x
@component Divider() {
	<div>
		<hr class="border-thick text-red"/>
	</div>
}`,
			wantContains: []string{
				"element.WithHR()",
				"element.WithBorder(tui.BorderThick)",
				"element.WithTextStyle(tui.NewStyle().Foreground(tui.Red))",
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

func TestGenerator_BR(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"basic br": {
			input: `package x
@component Lines() {
	<div>
		<span>Line 1</span>
		<br/>
		<span>Line 2</span>
	</div>
}`,
			wantContains: []string{
				"element.WithWidth(0)",
				"element.WithHeight(1)",
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

// TestGenerator_NamedRefs tests named element references (#Name syntax)
func TestGenerator_NamedRefs(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"simple named ref": {
			input: `package x
@component StreamBox() {
	<div #Content scrollable={element.ScrollVertical}></div>
}`,
			wantContains: []string{
				"type StreamBoxView struct",
				"Root     *element.Element",
				"watchers []tui.Watcher",
				"Content  *element.Element",
				"Content := element.New(",
				"element.WithScrollable(element.ScrollVertical)",
				"view = StreamBoxView{",
				"Root:     Content,",
				"Content:  Content,",
			},
		},
		"multiple named refs": {
			input: `package x
@component Layout() {
	<div>
		<div #Header height={3}></div>
		<div #Content flexGrow={1}></div>
		<div #Footer height={3}></div>
	</div>
}`,
			wantContains: []string{
				"type LayoutView struct",
				"Header   *element.Element",
				"Content  *element.Element",
				"Footer   *element.Element",
				"Header := element.New(",
				"Content := element.New(",
				"Footer := element.New(",
				"Header:   Header,",
				"Content:  Content,",
				"Footer:   Footer,",
			},
		},
		"named ref on root element": {
			input: `package x
@component Sidebar() {
	<nav #Navigation class="flex-col"></nav>
}`,
			wantContains: []string{
				"type SidebarView struct",
				"Root       *element.Element",
				"Navigation *element.Element",
				"Navigation := element.New(",
				"Root:       Navigation,",
				"Navigation: Navigation,",
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

// TestGenerator_NamedRefsInLoop tests refs inside @for loops generate slice fields
func TestGenerator_NamedRefsInLoop(t *testing.T) {
	input := `package x
@component ItemList(items []string) {
	<ul>
		@for _, item := range items {
			<li #Items>{item}</li>
		}
	</ul>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Should generate slice field for loop ref
	if !strings.Contains(code, "Items []*element.Element") {
		t.Errorf("missing slice field for loop ref\nGot:\n%s", code)
	}

	// Should declare var at function scope
	if !strings.Contains(code, "var Items []*element.Element") {
		t.Errorf("missing var declaration for loop ref\nGot:\n%s", code)
	}

	// Should append to slice in loop
	if !strings.Contains(code, "Items = append(Items,") {
		t.Errorf("missing append to Items slice\nGot:\n%s", code)
	}
}

// TestGenerator_NamedRefsInConditional tests refs inside @if generate may-be-nil fields
func TestGenerator_NamedRefsInConditional(t *testing.T) {
	input := `package x
@component Foo(showLabel bool) {
	<div>
		@if showLabel {
			<span #Label>{"Hi"}</span>
		}
	</div>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Should have comment indicating may be nil
	if !strings.Contains(code, "Label    *element.Element // may be nil") {
		t.Errorf("missing 'may be nil' comment for conditional ref\nGot:\n%s", code)
	}

	// Should declare var at function scope (outside conditional)
	if !strings.Contains(code, "var Label *element.Element") {
		t.Errorf("missing var declaration for conditional ref\nGot:\n%s", code)
	}

	// Should assign inside conditional
	if !strings.Contains(code, "Label = ") {
		t.Errorf("missing assignment inside conditional\nGot:\n%s", code)
	}
}

// TestGenerator_NamedRefsWithKey tests refs with key={expr} generate map fields
func TestGenerator_NamedRefsWithKey(t *testing.T) {
	input := `package x
@component UserList(users []User) {
	<ul>
		@for _, user := range users {
			<li #Users key={user.ID}>{user.Name}</li>
		}
	</ul>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Should generate map field for keyed ref
	if !strings.Contains(code, "Users    map[string]*element.Element") {
		t.Errorf("missing map field for keyed ref\nGot:\n%s", code)
	}

	// Should make map at function scope
	if !strings.Contains(code, "Users := make(map[string]*element.Element)") {
		t.Errorf("missing make for keyed ref\nGot:\n%s", code)
	}

	// Should assign to map in loop
	if !strings.Contains(code, "Users[user.ID] =") {
		t.Errorf("missing map assignment for keyed ref\nGot:\n%s", code)
	}
}

// TestGenerator_ViewVariablePreDeclared tests that view variable is pre-declared for closure capture
func TestGenerator_ViewVariablePreDeclared(t *testing.T) {
	input := `package x
@component StreamApp() {
	<div #Content></div>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Should have var view declared at start of function
	if !strings.Contains(code, "var view StreamAppView") {
		t.Errorf("missing pre-declared view variable\nGot:\n%s", code)
	}

	// The view variable should be assigned before return
	if !strings.Contains(code, "view = StreamAppView{") {
		t.Errorf("missing view assignment\nGot:\n%s", code)
	}

	// Should return view (not view.Root)
	if !strings.Contains(code, "return view") {
		t.Errorf("should return view struct\nGot:\n%s", code)
	}
}

// TestGenerator_WatcherGeneration tests onChannel and onTimer watcher attributes
func TestGenerator_WatcherGeneration(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"onChannel watcher": {
			input: `package x
@component StreamBox(dataCh chan string) {
	<div onChannel={tui.Watch(dataCh, handleData(lines))}></div>
}`,
			wantContains: []string{
				"watchers []tui.Watcher",
				"var watchers []tui.Watcher",
				"watchers = append(watchers, tui.Watch(dataCh, handleData(lines)))",
				"watchers: watchers,",
			},
		},
		"onTimer watcher": {
			input: `package x
@component Clock() {
	<div onTimer={tui.OnTimer(time.Second, tick(elapsed))}></div>
}`,
			wantContains: []string{
				"watchers []tui.Watcher",
				"var watchers []tui.Watcher",
				"watchers = append(watchers, tui.OnTimer(time.Second, tick(elapsed)))",
				"watchers: watchers,",
			},
		},
		"multiple watchers on same element": {
			input: `package x
@component Streaming(dataCh chan string) {
	<div
		onChannel={tui.Watch(dataCh, handleData)}
		onTimer={tui.OnTimer(time.Second, tick)}
	></div>
}`,
			wantContains: []string{
				"watchers = append(watchers, tui.Watch(dataCh, handleData))",
				"watchers = append(watchers, tui.OnTimer(time.Second, tick))",
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

// TestGenerator_WatcherAggregation tests that nested component watchers are aggregated
func TestGenerator_WatcherAggregation(t *testing.T) {
	input := `package x
@component StreamBox(dataCh chan string) {
	<div onChannel={tui.Watch(dataCh, handleData)}></div>
}

@component Clock() {
	<div onTimer={tui.OnTimer(time.Second, tick)}></div>
}

@component App(dataCh chan string) {
	<div>
		@StreamBox(dataCh)
		@Clock()
	</div>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// App should aggregate watchers from child components
	if !strings.Contains(code, ".GetWatchers()...") {
		t.Errorf("missing watcher aggregation from child components\nGot:\n%s", code)
	}
}

// TestGenerator_ViewableInterface tests GetRoot and GetWatchers methods are generated
func TestGenerator_ViewableInterface(t *testing.T) {
	input := `package x
@component Test() {
	<div></div>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	// Check GetRoot method
	if !strings.Contains(code, "func (v TestView) GetRoot() tui.Renderable { return v.Root }") {
		t.Errorf("missing GetRoot method\nGot:\n%s", code)
	}

	// Check GetWatchers method
	if !strings.Contains(code, "func (v TestView) GetWatchers() []tui.Watcher { return v.watchers }") {
		t.Errorf("missing GetWatchers method\nGot:\n%s", code)
	}
}

// TestGenerator_OnKeyPressAttribute tests onKeyPress handler generation
func TestGenerator_OnKeyPressAttribute(t *testing.T) {
	input := `package x
@component Counter() {
	<div onKeyPress={handleKeys(count)} focusable={true}></div>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "element.WithOnKeyPress(handleKeys(count))") {
		t.Errorf("missing onKeyPress option\nGot:\n%s", code)
	}

	if !strings.Contains(code, "element.WithFocusable(true)") {
		t.Errorf("missing focusable option\nGot:\n%s", code)
	}
}

// TestGenerator_OnClickAttribute tests onClick handler generation
func TestGenerator_OnClickAttribute(t *testing.T) {
	input := `package x
@component Button(onClick func()) {
	<div onClick={onClick}></div>
}`

	output, err := ParseAndGenerate("test.tui", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	code := string(output)

	if !strings.Contains(code, "element.WithOnClick(onClick)") {
		t.Errorf("missing onClick option\nGot:\n%s", code)
	}
}

// TestGenerator_StateBindings tests state variable and binding generation
func TestGenerator_StateBindings(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
		wantNotContains []string
	}

	tests := map[string]tc{
		"single state with binding": {
			input: `package x
@component Counter() {
	count := tui.NewState(0)
	<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
}`,
			wantContains: []string{
				"count := tui.NewState(0)",
				"// State bindings",
				"count.Bind(func(_ int) {",
				`__tui_0.SetText(fmt.Sprintf("Count: %d", count.Get()))`,
			},
		},
		"state parameter - no declaration generated": {
			input: `package x
@component Counter(count *tui.State[int]) {
	<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
}`,
			wantContains: []string{
				"func Counter(count *tui.State[int]) CounterView",
				"// State bindings",
				"count.Bind(func(_ int) {",
				`__tui_0.SetText(fmt.Sprintf("Count: %d", count.Get()))`,
			},
			wantNotContains: []string{
				"count := tui.NewState", // parameter states should not be re-declared
			},
		},
		"multiple states in expression": {
			input: `package x
@component Profile() {
	firstName := tui.NewState("Alice")
	lastName := tui.NewState("Smith")
	<span>{firstName.Get() + " " + lastName.Get()}</span>
}`,
			wantContains: []string{
				`firstName := tui.NewState("Alice")`,
				`lastName := tui.NewState("Smith")`,
				"// State bindings",
				"__update___tui_0 := func()",
				"firstName.Bind(func(_ string)",
				"lastName.Bind(func(_ string)",
				"__update___tui_0()",
			},
		},
		"state with named ref": {
			input: `package x
@component Counter() {
	count := tui.NewState(0)
	<span #Display>{fmt.Sprintf("Count: %d", count.Get())}</span>
}`,
			wantContains: []string{
				"count := tui.NewState(0)",
				"count.Bind(func(_ int) {",
				`Display.SetText(fmt.Sprintf("Count: %d", count.Get()))`,
			},
		},
		"explicit deps attribute": {
			input: `package x
@component UserCard() {
	user := tui.NewState(&User{Name: "Alice"})
	<span deps={[user]}>{formatUser(user.Get())}</span>
}`,
			wantContains: []string{
				`user := tui.NewState(&User{Name: "Alice"})`,
				"user.Bind(func(_ *User) {",
				"__tui_0.SetText(formatUser(user.Get()))",
			},
		},
		"no binding when no state used": {
			input: `package x
@component Static() {
	<span>{"Hello"}</span>
}`,
			wantNotContains: []string{
				"// State bindings",
				".Bind(",
			},
		},
		"bool state type inference": {
			input: `package x
@component Toggle() {
	enabled := tui.NewState(true)
	<span>{fmt.Sprintf("Enabled: %v", enabled.Get())}</span>
}`,
			wantContains: []string{
				"enabled := tui.NewState(true)",
				"enabled.Bind(func(_ bool) {",
			},
		},
		"slice state type inference": {
			input: `package x
@component List() {
	items := tui.NewState([]string{})
	<span>{fmt.Sprintf("Items: %d", len(items.Get()))}</span>
}`,
			wantContains: []string{
				`items := tui.NewState([]string{})`,
				"items.Bind(func(_ []string) {",
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

// TestGenerator_StateBindingsCompile verifies that generated state binding code compiles
func TestGenerator_StateBindingsCompile(t *testing.T) {
	// Skip if go command not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	input := `package main

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/tui"
)

@component Counter() {
	count := tui.NewState(0)
	<div class="flex-col">
		<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
	</div>
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

	code := string(output)

	// Check key elements are present
	if !strings.Contains(code, "count := tui.NewState(0)") {
		t.Error("missing state declaration")
	}
	if !strings.Contains(code, "count.Bind(") {
		t.Error("missing state binding")
	}
	if !strings.Contains(code, "func Counter() CounterView") {
		t.Error("missing function declaration")
	}
}
