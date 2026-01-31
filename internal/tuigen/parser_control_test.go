package tuigen

import (
	"testing"
)

func TestParser_LetBinding(t *testing.T) {
	input := `package x
templ Test() {
	@let myText = <span>Hello</span>
	<div></div>
}`

	l := NewLexer("test.gsx", input)
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
		input     string
		wantIndex string
		wantValue string
		wantIter  string
	}

	tests := map[string]tc{
		"index and value": {
			input: `package x
templ Test() {
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
templ Test() {
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
templ Test() {
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
			l := NewLexer("test.gsx", tt.input)
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
templ Test() {
	@if showHeader {
		<span>Header</span>
	}
}`,
			wantCondition: "showHeader",
			wantElse:      false,
		},
		"if with else": {
			input: `package x
templ Test() {
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
templ Test() {
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
			l := NewLexer("test.gsx", tt.input)
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
templ Test() {
	@if a {
		<span>A</span>
	} @else @if b {
		<span>B</span>
	} @else {
		<span>C</span>
	}
}`

	l := NewLexer("test.gsx", input)
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

func TestParser_ControlFlowInChildren(t *testing.T) {
	input := `package x
templ Test(items []string) {
	<div>
		@for _, item := range items {
			<span>{item}</span>
		}
		@if len(items) == 0 {
			<span>No items</span>
		}
	</div>
}`

	l := NewLexer("test.gsx", input)
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

func TestParser_RawGoStatements(t *testing.T) {
	type tc struct {
		input     string
		wantCodes []string
	}

	tests := map[string]tc{
		"simple assignment": {
			input: `package x
templ Test() {
	x := 1
	<span>{x}</span>
}`,
			wantCodes: []string{"x := 1"},
		},
		"function call": {
			input: `package x
templ Test() {
	fmt.Println("hello")
	<span>world</span>
}`,
			wantCodes: []string{`fmt.Println("hello")`},
		},
		"multi-line statement": {
			input: `package x
templ Test() {
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
templ Test() {
	x := 1
	y := 2
	z := x + y
	<span>{z}</span>
}`,
			wantCodes: []string{"x := 1", "y := 2", "z := x + y"},
		},
		"inline if statement": {
			input: `package x
templ Test(err error) {
	if err != nil { log.Error(err) }
	<span>done</span>
}`,
			wantCodes: []string{"if err != nil { log.Error(err) }"},
		},
		"defer statement": {
			input: `package x
templ Test() {
	defer cleanup()
	<span>running</span>
}`,
			wantCodes: []string{"defer cleanup()"},
		},
		"go statement": {
			input: `package x
templ Test() {
	go doWork()
	<span>spawned</span>
}`,
			wantCodes: []string{"go doWork()"},
		},
		"for loop statement": {
			input: `package x
templ Test() {
	for i := 0; i < 10; i++ { sum += i }
	<span>{sum}</span>
}`,
			wantCodes: []string{"for i := 0; i < 10; i++ { sum += i }"},
		},
		"switch statement": {
			input: `package x
templ Test(x int) {
	switch x { case 1: y = "one"; case 2: y = "two" }
	<span>{y}</span>
}`,
			wantCodes: []string{`switch x { case 1: y = "one"; case 2: y = "two" }`},
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
templ Counter(count int) {
	formattedCount := fmt.Sprintf("%d", count)
	log.Printf("Rendering counter")
	<div>
		<span>{formattedCount}</span>
	</div>
}`

	l := NewLexer("test.gsx", input)
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
