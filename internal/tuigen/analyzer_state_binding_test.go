package tuigen

import (
	"strings"
	"testing"
)

func TestAnalyzer_DetectStateBindings_SimpleGet(t *testing.T) {
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
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 1 || b.StateVars[0] != "count" {
		t.Errorf("StateVars = %v, want [count]", b.StateVars)
	}
	if b.Attribute != "text" {
		t.Errorf("Attribute = %q, want 'text'", b.Attribute)
	}
	if b.ExplicitDeps {
		t.Error("expected ExplicitDeps to be false")
	}
}

func TestAnalyzer_DetectStateBindings_FormatString(t *testing.T) {
	input := `package x
templ Counter(count *tui.State[int]) {
	<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 1 || b.StateVars[0] != "count" {
		t.Errorf("StateVars = %v, want [count]", b.StateVars)
	}
	if !strings.Contains(b.Expr, "fmt.Sprintf") {
		t.Errorf("Expr = %q, should contain 'fmt.Sprintf'", b.Expr)
	}
}

func TestAnalyzer_DetectStateBindings_MultipleStates(t *testing.T) {
	input := `package x
templ Profile(firstName *tui.State[string], lastName *tui.State[string]) {
	<span>{fmt.Sprintf("%s %s", firstName.Get(), lastName.Get())}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 2 {
		t.Fatalf("expected 2 state vars, got %d: %v", len(b.StateVars), b.StateVars)
	}
	// Check both states are detected (order may vary)
	hasFirst := false
	hasLast := false
	for _, sv := range b.StateVars {
		if sv == "firstName" {
			hasFirst = true
		}
		if sv == "lastName" {
			hasLast = true
		}
	}
	if !hasFirst || !hasLast {
		t.Errorf("StateVars = %v, want [firstName, lastName]", b.StateVars)
	}
}

func TestAnalyzer_DetectStateBindings_ExplicitDeps(t *testing.T) {
	input := `package x
templ UserCard(user *tui.State[*User]) {
	<span deps={[user]}>{formatUser(user.Get())}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 1 || b.StateVars[0] != "user" {
		t.Errorf("StateVars = %v, want [user]", b.StateVars)
	}
	if !b.ExplicitDeps {
		t.Error("expected ExplicitDeps to be true")
	}
}

func TestAnalyzer_DetectStateBindings_ExplicitDepsMultiple(t *testing.T) {
	input := `package x
templ Combined(count *tui.State[int], name *tui.State[string]) {
	<span deps={[count, name]}>{compute(count, name)}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 2 {
		t.Fatalf("expected 2 state vars, got %d: %v", len(b.StateVars), b.StateVars)
	}
	if !b.ExplicitDeps {
		t.Error("expected ExplicitDeps to be true")
	}
}

func TestAnalyzer_DetectStateBindings_UnknownStateInDeps(t *testing.T) {
	input := `package x
templ Test(count *tui.State[int]) {
	<span deps={[unknown]}>{count.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	_ = analyzer.DetectStateBindings(file.Components[0], stateVars)

	// Check that an error was recorded
	errors := analyzer.Errors().Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0].Message, "unknown state variable") {
		t.Errorf("error message = %q, want to contain 'unknown state variable'", errors[0].Message)
	}
}

func TestAnalyzer_DetectStateBindings_DynamicClass(t *testing.T) {
	input := `package x
templ Toggle(enabled *tui.State[bool]) {
	<span class={enabled.Get() ? "text-green" : "text-red"}>Status</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if b.Attribute != "class" {
		t.Errorf("Attribute = %q, want 'class'", b.Attribute)
	}
	if len(b.StateVars) != 1 || b.StateVars[0] != "enabled" {
		t.Errorf("StateVars = %v, want [enabled]", b.StateVars)
	}
}

func TestAnalyzer_DetectStateBindings_NoStateUsage(t *testing.T) {
	input := `package x
templ Static() {
	<span>{"Hello, World!"}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 0 {
		t.Errorf("expected 0 bindings, got %d", len(bindings))
	}
}

func TestAnalyzer_DetectStateBindings_WithRef(t *testing.T) {
	input := `package x
templ Counter(count *tui.State[int]) {
	<span ref={label}>{count.Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	// All elements use auto-generated variable names now
	if b.ElementName != "__tui_0" {
		t.Errorf("ElementName = %q, want '__tui_0'", b.ElementName)
	}
}

func TestAnalyzer_DepsAttributeValid(t *testing.T) {
	// Test that deps attribute is recognized as valid
	input := `package x
templ Test(count *tui.State[int]) {
	<span deps={[count]}>{count.Get()}</span>
}`

	_, err := AnalyzeFile("test.gsx", input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnalyzer_DetectStateBindings_ReactiveIfSkipsChildren(t *testing.T) {
	// Elements inside a reactive @if should NOT generate separate text bindings
	// because the reactive update function rebuilds them entirely.
	input := `package x
templ Toggle() {
	count := tui.NewState(0)
	<div>
		@if count.Get() == 0 {
			<span>{fmt.Sprintf("zero: %d", count.Get())}</span>
		} @else {
			<span>{fmt.Sprintf("nonzero: %d", count.Get())}</span>
		}
	</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	// Should have NO text bindings - elements inside reactive @if are skipped
	if len(bindings) != 0 {
		t.Errorf("expected 0 bindings for reactive @if children, got %d", len(bindings))
		for i, b := range bindings {
			t.Errorf("  binding[%d]: element=%s attr=%s expr=%s", i, b.ElementName, b.Attribute, b.Expr)
		}
	}
}

func TestAnalyzer_DetectStateBindings_NonReactiveIfKeepsBindings(t *testing.T) {
	// Elements inside a non-reactive @if (no state in condition) should still
	// generate text bindings normally.
	input := `package x
templ Toggle(count *tui.State[int]) {
	<div>
		@if true {
			<span>{count.Get()}</span>
		}
	</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	// Should have 1 text binding since @if condition doesn't reference state
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding for non-reactive @if, got %d", len(bindings))
	}

	b := bindings[0]
	if b.Attribute != "text" {
		t.Errorf("Attribute = %q, want %q", b.Attribute, "text")
	}
	if len(b.StateVars) != 1 || b.StateVars[0] != "count" {
		t.Errorf("StateVars = %v, want [count]", b.StateVars)
	}
}

func TestAnalyzer_DetectStateBindings_DereferencedPointer(t *testing.T) {
	// Test that (*count).Get() pattern is detected
	input := `package x
templ Counter(count *tui.State[int]) {
	<span>{(*count).Get()}</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 1 || b.StateVars[0] != "count" {
		t.Errorf("StateVars = %v, want [count]", b.StateVars)
	}
}
