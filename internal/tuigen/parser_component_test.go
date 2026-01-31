package tuigen

import (
	"testing"
)

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
templ Header() {
	<span>Hello</span>
}`,
			wantName:   "Header",
			wantParams: 0,
		},
		"one param": {
			input: `package x
templ Greeting(name string) {
	<span>Hello</span>
}`,
			wantName:   "Greeting",
			wantParams: 1,
		},
		"multiple params": {
			input: `package x
templ Counter(count int, label string) {
	<span>Hello</span>
}`,
			wantName:   "Counter",
			wantParams: 2,
		},
		"complex types": {
			input: `package x
templ List(items []string, onClick func()) {
	<span>Hello</span>
}`,
			wantName:   "List",
			wantParams: 2,
		},
		"pointer type": {
			input: `package x
templ View(elem *element.Element) {
	<span>Hello</span>
}`,
			wantName:   "View",
			wantParams: 1,
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
templ Test(name string, count int, items []string, handler func()) {
	<span>Hello</span>
}`

	l := NewLexer("test.gsx", input)
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
		name string
		typ  string
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

func TestParser_ComplexTypeSignatures(t *testing.T) {
	type tc struct {
		input     string
		wantTypes []string
	}

	tests := map[string]tc{
		"channel type": {
			input: `package x
templ Test(ch chan int) {
	<span>Hello</span>
}`,
			wantTypes: []string{"chan int"},
		},
		"receive channel": {
			input: `package x
templ Test(ch <-chan string) {
	<span>Hello</span>
}`,
			wantTypes: []string{"<-chan string"},
		},
		"complex map": {
			input: `package x
templ Test(m map[string][]int) {
	<span>Hello</span>
}`,
			wantTypes: []string{"map[string][]int"},
		},
		"function with return": {
			input: `package x
templ Test(fn func(a, b int) (string, error)) {
	<span>Hello</span>
}`,
			wantTypes: []string{"func(a, b int) (string, error)"},
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
templ App() {
	@Header()
}`,
			wantName:     "Header",
			wantArgs:     "",
			wantChildren: 0,
		},
		"call with args no children": {
			input: `package x
templ App() {
	@Header("Welcome", true)
}`,
			wantName:     "Header",
			wantArgs:     `"Welcome", true`,
			wantChildren: 0,
		},
		"call with children": {
			input: `package x
templ App() {
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
templ App() {
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
templ Card(title string) {
	<div>
		<span>{title}</span>
		{children...}
	</div>
}`
	l := NewLexer("test.gsx", input)
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
templ App() {
	<div>
		@Header("Title")
		@Footer()
	</div>
}`
	l := NewLexer("test.gsx", input)
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
