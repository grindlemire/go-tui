package tuigen

import (
	"strings"
	"testing"
)

func TestAnalyzer_DepsStringLiteralError(t *testing.T) {
	// Test that deps="string" produces an error
	input := `package x
templ Test(count *tui.State[int]) {
	<span deps="not-valid">{count.Get()}</span>
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

	errors := analyzer.Errors().Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0].Message, "must use expression syntax") {
		t.Errorf("error message = %q, want to contain 'must use expression syntax'", errors[0].Message)
	}
}

func TestAnalyzer_DepsMissingBracketsError(t *testing.T) {
	// Test that deps={count} (missing brackets) produces an error
	input := `package x
templ Test(count *tui.State[int]) {
	<span deps={count}>{count.Get()}</span>
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

	errors := analyzer.Errors().Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0].Message, "must be an array literal") {
		t.Errorf("error message = %q, want to contain 'must be an array literal'", errors[0].Message)
	}
}

func TestAnalyzer_DepsEmptyArrayWarning(t *testing.T) {
	// Test that deps={[]} (empty) produces a warning
	input := `package x
templ Test(count *tui.State[int]) {
	<span deps={[]}>{count.Get()}</span>
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

	errors := analyzer.Errors().Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0].Message, "empty deps attribute") {
		t.Errorf("error message = %q, want to contain 'empty deps attribute'", errors[0].Message)
	}
}

func TestAnalyzer_MultipleErrors(t *testing.T) {
	// Test that multiple errors are collected
	input := `package x
templ Test() {
	<unknownTag1 />
	<unknownTag2 />
}`

	_, err := AnalyzeFile("test.gsx", input)
	if err == nil {
		t.Fatal("expected errors, got nil")
	}

	errStr := err.Error()

	if !strings.Contains(errStr, "unknownTag1") {
		t.Error("missing error for unknownTag1")
	}

	if !strings.Contains(errStr, "unknownTag2") {
		t.Error("missing error for unknownTag2")
	}
}

func TestAnalyzer_ErrorHint(t *testing.T) {
	input := `package x
templ Test() {
	<div colour="red"></div>
}`

	_, err := AnalyzeFile("test.gsx", input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()

	// Should have hint about similar attribute
	if !strings.Contains(errStr, "did you mean") {
		t.Errorf("error should contain hint, got: %s", errStr)
	}
}
