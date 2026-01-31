package tuigen

import (
	"testing"
)

func TestAnalyzer_DetectStateVars_IntLiteral(t *testing.T) {
	// Since GoCode blocks are handled specially, we need to test with
	// the actual parsing. For now, test the type inference separately.
	type tc struct {
		expr     string
		wantType string
	}

	tests := map[string]tc{
		"integer 0":       {expr: "0", wantType: "int"},
		"integer 42":      {expr: "42", wantType: "int"},
		"negative int":    {expr: "-5", wantType: "int"},
		"float":           {expr: "3.14", wantType: "float64"},
		"negative float":  {expr: "-2.5", wantType: "float64"},
		"bool true":       {expr: "true", wantType: "bool"},
		"bool false":      {expr: "false", wantType: "bool"},
		"string double":   {expr: `"hello"`, wantType: "string"},
		"string backtick": {expr: "`raw`", wantType: "string"},
		"slice literal":   {expr: "[]string{}", wantType: "[]string"},
		"slice with pkg":  {expr: "[]pkg.Type{}", wantType: "[]pkg.Type"},
		"map literal":     {expr: "map[string]int{}", wantType: "map[string]int"},
		"pointer struct":  {expr: "&User{}", wantType: "*User"},
		"pointer pkg":     {expr: "&pkg.User{}", wantType: "*pkg.User"},
		"struct literal":  {expr: "User{}", wantType: "User"},
		"nil":             {expr: "nil", wantType: "any"},
		"function call":   {expr: "someFunc()", wantType: "any"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := inferTypeFromExpr(tt.expr)
			if result != tt.wantType {
				t.Errorf("inferTypeFromExpr(%q) = %q, want %q", tt.expr, result, tt.wantType)
			}
		})
	}
}

func TestAnalyzer_DetectStateVars_Parameter(t *testing.T) {
	input := `package x
templ Counter(count *tui.State[int]) {
	<span>{count.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "count" {
		t.Errorf("Name = %q, want 'count'", sv.Name)
	}
	if sv.Type != "int" {
		t.Errorf("Type = %q, want 'int'", sv.Type)
	}
	if !sv.IsParameter {
		t.Error("expected IsParameter to be true")
	}
	if sv.InitExpr != "" {
		t.Errorf("InitExpr = %q, want empty for parameter", sv.InitExpr)
	}
}

func TestAnalyzer_DetectStateVars_StringParameter(t *testing.T) {
	input := `package x
templ Greeting(name *tui.State[string]) {
	<span>{name.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "name" {
		t.Errorf("Name = %q, want 'name'", sv.Name)
	}
	if sv.Type != "string" {
		t.Errorf("Type = %q, want 'string'", sv.Type)
	}
}

func TestAnalyzer_DetectStateVars_SliceParameter(t *testing.T) {
	input := `package x
templ TodoList(items *tui.State[[]string]) {
	<div>{items.Get()}</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Type != "[]string" {
		t.Errorf("Type = %q, want '[]string'", sv.Type)
	}
}

func TestAnalyzer_DetectStateVars_PointerParameter(t *testing.T) {
	input := `package x
templ UserProfile(user *tui.State[*User]) {
	<div>{user.Get()}</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Type != "*User" {
		t.Errorf("Type = %q, want '*User'", sv.Type)
	}
}

func TestAnalyzer_DetectStateVars_GoCodeDeclaration(t *testing.T) {
	// Test detection of tui.NewState in component body (GoCode block)
	input := `package x
templ Counter() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "count" {
		t.Errorf("Name = %q, want 'count'", sv.Name)
	}
	if sv.Type != "int" {
		t.Errorf("Type = %q, want 'int'", sv.Type)
	}
	if sv.IsParameter {
		t.Error("expected IsParameter to be false for GoCode declaration")
	}
	if sv.InitExpr != "0" {
		t.Errorf("InitExpr = %q, want '0'", sv.InitExpr)
	}
}

func TestAnalyzer_DetectStateVars_GoCodeDeclarationString(t *testing.T) {
	// Test detection of tui.NewState with string literal
	input := `package x
templ Greeting() {
	name := tui.NewState("Alice")
	<span>{name.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "name" {
		t.Errorf("Name = %q, want 'name'", sv.Name)
	}
	if sv.Type != "string" {
		t.Errorf("Type = %q, want 'string'", sv.Type)
	}
	if sv.InitExpr != `"Alice"` {
		t.Errorf("InitExpr = %q, want '\"Alice\"'", sv.InitExpr)
	}
}

func TestAnalyzer_DetectStateVars_GoCodeDeclarationSlice(t *testing.T) {
	// Test detection of tui.NewState with slice literal (matching plan spec)
	input := `package x
templ TodoList() {
	items := tui.NewState([]string{})
	<div>{items.Get()}</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "items" {
		t.Errorf("Name = %q, want 'items'", sv.Name)
	}
	if sv.Type != "[]string" {
		t.Errorf("Type = %q, want '[]string'", sv.Type)
	}
}

func TestAnalyzer_DetectStateVars_GoCodeDeclarationBool(t *testing.T) {
	// Test detection of tui.NewState with boolean literal
	input := `package x
templ Toggle() {
	enabled := tui.NewState(true)
	<span>{enabled.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Type != "bool" {
		t.Errorf("Type = %q, want 'bool'", sv.Type)
	}
	if sv.InitExpr != "true" {
		t.Errorf("InitExpr = %q, want 'true'", sv.InitExpr)
	}
}

func TestAnalyzer_DetectStateVars_MultipleDeclarations(t *testing.T) {
	// Test detection of multiple tui.NewState declarations
	input := `package x
templ Profile() {
	firstName := tui.NewState("Alice")
	lastName := tui.NewState("Smith")
	age := tui.NewState(30)
	<span>{firstName.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 3 {
		t.Fatalf("expected 3 state vars, got %d", len(stateVars))
	}

	// Check that all are detected
	names := make(map[string]string)
	for _, sv := range stateVars {
		names[sv.Name] = sv.Type
	}

	if names["firstName"] != "string" {
		t.Errorf("firstName type = %q, want 'string'", names["firstName"])
	}
	if names["lastName"] != "string" {
		t.Errorf("lastName type = %q, want 'string'", names["lastName"])
	}
	if names["age"] != "int" {
		t.Errorf("age type = %q, want 'int'", names["age"])
	}
}

func TestAnalyzer_DetectStateVars_MixedParamsAndDeclarations(t *testing.T) {
	// Test detection of both parameter states and GoCode declarations
	input := `package x
templ Counter(initialCount *tui.State[int]) {
	label := tui.NewState("Count: ")
	<span>{label.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 2 {
		t.Fatalf("expected 2 state vars, got %d", len(stateVars))
	}

	// Find each by name
	var param, decl *StateVar
	for i := range stateVars {
		if stateVars[i].Name == "initialCount" {
			param = &stateVars[i]
		}
		if stateVars[i].Name == "label" {
			decl = &stateVars[i]
		}
	}

	if param == nil {
		t.Fatal("parameter state 'initialCount' not found")
	}
	if !param.IsParameter {
		t.Error("initialCount should be marked as parameter")
	}
	if param.Type != "int" {
		t.Errorf("initialCount type = %q, want 'int'", param.Type)
	}

	if decl == nil {
		t.Fatal("declared state 'label' not found")
	}
	if decl.IsParameter {
		t.Error("label should not be marked as parameter")
	}
	if decl.Type != "string" {
		t.Errorf("label type = %q, want 'string'", decl.Type)
	}
}
