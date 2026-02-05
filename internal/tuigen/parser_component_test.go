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

func TestParser_MethodTempl(t *testing.T) {
	type tc struct {
		input            string
		wantName         string
		wantReceiver     string
		wantReceiverName string
		wantReceiverType string
		wantBodyLen      int
	}

	tests := map[string]tc{
		"pointer receiver": {
			input: `package x
templ (s *sidebar) Render() {
	<span>Hello</span>
}`,
			wantName:         "Render",
			wantReceiver:     "s *sidebar",
			wantReceiverName: "s",
			wantReceiverType: "*sidebar",
			wantBodyLen:      1,
		},
		"value receiver": {
			input: `package x
templ (v myView) Render() {
	<span>Hello</span>
}`,
			wantName:         "Render",
			wantReceiver:     "v myView",
			wantReceiverName: "v",
			wantReceiverType: "myView",
			wantBodyLen:      1,
		},
		"receiver with qualified type": {
			input: `package x
templ (a *pkg.App) Render() {
	<div>
		<span>Content</span>
	</div>
}`,
			wantName:         "Render",
			wantReceiver:     "a *pkg.App",
			wantReceiverName: "a",
			wantReceiverType: "*pkg.App",
			wantBodyLen:      1,
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

			comp := file.Components[0]
			if comp.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", comp.Name, tt.wantName)
			}
			if comp.Receiver != tt.wantReceiver {
				t.Errorf("Receiver = %q, want %q", comp.Receiver, tt.wantReceiver)
			}
			if comp.ReceiverName != tt.wantReceiverName {
				t.Errorf("ReceiverName = %q, want %q", comp.ReceiverName, tt.wantReceiverName)
			}
			if comp.ReceiverType != tt.wantReceiverType {
				t.Errorf("ReceiverType = %q, want %q", comp.ReceiverType, tt.wantReceiverType)
			}
			if len(comp.Params) != 0 {
				t.Errorf("method templ should have no params, got %d", len(comp.Params))
			}
			if len(comp.Body) != tt.wantBodyLen {
				t.Errorf("len(Body) = %d, want %d", len(comp.Body), tt.wantBodyLen)
			}
		})
	}
}

func TestParser_MethodTemplErrors(t *testing.T) {
	type tc struct {
		input       string
		wantErrText string
	}

	tests := map[string]tc{
		"method name must be Render": {
			input: `package x
templ (s *sidebar) NotRender() {
	<span>Hello</span>
}`,
			wantErrText: "method templ name must be 'Render'",
		},
		"method templ no params allowed": {
			input: `package x
templ (s *sidebar) Render(title string) {
	<span>Hello</span>
}`,
			wantErrText: "method templ Render() must not have parameters",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			p := NewParser(l)
			_, err := p.ParseFile()

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			errStr := err.Error()
			if !containsSubstring(errStr, tt.wantErrText) {
				t.Errorf("error = %q, want to contain %q", errStr, tt.wantErrText)
			}
		})
	}
}

func TestParser_FunctionTemplReceiverFieldsEmpty(t *testing.T) {
	input := `package x
templ Header(title string) {
	<span>{title}</span>
}`
	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	if comp.Receiver != "" {
		t.Errorf("function templ Receiver = %q, want empty", comp.Receiver)
	}
	if comp.ReceiverName != "" {
		t.Errorf("function templ ReceiverName = %q, want empty", comp.ReceiverName)
	}
	if comp.ReceiverType != "" {
		t.Errorf("function templ ReceiverType = %q, want empty", comp.ReceiverType)
	}
}

func TestParser_ComponentCallIsStructMount(t *testing.T) {
	type tc struct {
		input           string
		wantStructMount bool
	}

	tests := map[string]tc{
		"component call in method templ is struct mount": {
			input: `package x
templ (a *myApp) Render() {
	@Sidebar(a.query)
}`,
			wantStructMount: true,
		},
		"component call in function templ is not struct mount": {
			input: `package x
templ App() {
	@Header()
}`,
			wantStructMount: false,
		},
		"nested component call in method templ element": {
			input: `package x
templ (a *myApp) Render() {
	<div>
		@SearchInput(a.active, a.query)
	</div>
}`,
			wantStructMount: true,
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

			// Find the ComponentCall in the body (may be top-level or nested in element)
			call := findComponentCall(file.Components[0].Body)
			if call == nil {
				t.Fatal("expected to find a ComponentCall in body")
			}

			if call.IsStructMount != tt.wantStructMount {
				t.Errorf("IsStructMount = %v, want %v", call.IsStructMount, tt.wantStructMount)
			}
		})
	}
}

func TestParser_BothTemplFormsCoexist(t *testing.T) {
	input := `package x

templ Header(title string) {
	<span>{title}</span>
}

templ (s *sidebar) Render() {
	<div>
		@Header("Welcome")
	</div>
}
`
	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(file.Components))
	}

	// First: function templ
	header := file.Components[0]
	if header.Name != "Header" {
		t.Errorf("component 0 Name = %q, want 'Header'", header.Name)
	}
	if header.Receiver != "" {
		t.Errorf("component 0 should be function templ (no receiver), got %q", header.Receiver)
	}

	// Second: method templ
	sidebar := file.Components[1]
	if sidebar.Name != "Render" {
		t.Errorf("component 1 Name = %q, want 'Render'", sidebar.Name)
	}
	if sidebar.Receiver != "s *sidebar" {
		t.Errorf("component 1 Receiver = %q, want 's *sidebar'", sidebar.Receiver)
	}

	// @Header inside method templ should be struct mount
	call := findComponentCall(sidebar.Body)
	if call == nil {
		t.Fatal("expected ComponentCall in method templ body")
	}
	if !call.IsStructMount {
		t.Error("ComponentCall inside method templ should have IsStructMount=true")
	}
}

func TestParser_MethodTemplWithControlFlow(t *testing.T) {
	input := `package x
templ (s *sidebar) Render() {
	@if s.expanded.Get() {
		<div>
			@ChildComponent(s.query)
		</div>
	}
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

	comp := file.Components[0]
	if comp.Receiver != "s *sidebar" {
		t.Errorf("Receiver = %q, want 's *sidebar'", comp.Receiver)
	}

	// Body should have an @if
	if len(comp.Body) != 1 {
		t.Fatalf("expected 1 body node, got %d", len(comp.Body))
	}

	ifStmt, ok := comp.Body[0].(*IfStmt)
	if !ok {
		t.Fatalf("body[0]: expected *IfStmt, got %T", comp.Body[0])
	}

	// The @if should contain a <div> which contains @ChildComponent
	if len(ifStmt.Then) != 1 {
		t.Fatalf("expected 1 then node, got %d", len(ifStmt.Then))
	}
	elem, ok := ifStmt.Then[0].(*Element)
	if !ok {
		t.Fatalf("then[0]: expected *Element, got %T", ifStmt.Then[0])
	}

	call := findComponentCall([]Node{elem})
	if call == nil {
		t.Fatal("expected ComponentCall inside @if body")
	}
	if call.Name != "ChildComponent" {
		t.Errorf("call.Name = %q, want 'ChildComponent'", call.Name)
	}
	if !call.IsStructMount {
		t.Error("ComponentCall inside method templ @if should have IsStructMount=true")
	}
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// findComponentCall recursively searches nodes for the first ComponentCall.
func findComponentCall(nodes []Node) *ComponentCall {
	for _, n := range nodes {
		switch v := n.(type) {
		case *ComponentCall:
			return v
		case *Element:
			if call := findComponentCall(v.Children); call != nil {
				return call
			}
		case *IfStmt:
			if call := findComponentCall(v.Then); call != nil {
				return call
			}
			if call := findComponentCall(v.Else); call != nil {
				return call
			}
		case *ForLoop:
			if call := findComponentCall(v.Body); call != nil {
				return call
			}
		}
	}
	return nil
}
