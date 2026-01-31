package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func TestDiagnostics_NoErrors(t *testing.T) {
	dp := NewDiagnosticsProvider()
	doc := parseTestDoc(`package main

templ Hello() {
	<span>Hello</span>
}
`)

	result, err := dp.Diagnose(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(result))
	}
}

func TestDiagnostics_NilAST(t *testing.T) {
	dp := NewDiagnosticsProvider()
	doc := &Document{
		URI:     "file:///test.gsx",
		Content: "",
		Version: 1,
		AST:     nil,
		Errors:  nil,
	}

	result, err := dp.Diagnose(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(result))
	}
}

func TestDiagnostics_SingleError(t *testing.T) {
	type tc struct {
		errors   []*tuigen.Error
		wantDiag int
		wantMsg  string
	}

	tests := map[string]tc{
		"basic error": {
			errors: []*tuigen.Error{
				{
					Pos:     tuigen.Position{Line: 3, Column: 5},
					Message: "unexpected token",
				},
			},
			wantDiag: 1,
			wantMsg:  "unexpected token",
		},
		"error with hint": {
			errors: []*tuigen.Error{
				{
					Pos:     tuigen.Position{Line: 3, Column: 5},
					Message: "unknown element",
					Hint:    "did you mean <div>?",
				},
			},
			wantDiag: 1,
			wantMsg:  "did you mean",
		},
	}

	dp := NewDiagnosticsProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := &Document{
				URI:     "file:///test.gsx",
				Content: "package main\n",
				Version: 1,
				AST:     nil,
				Errors:  tt.errors,
			}

			result, err := dp.Diagnose(doc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.wantDiag {
				t.Fatalf("expected %d diagnostics, got %d", tt.wantDiag, len(result))
			}

			diag := result[0]
			if diag.Severity != DiagnosticSeverityError {
				t.Errorf("expected severity Error, got %d", diag.Severity)
			}
			if diag.Source != "gsx" {
				t.Errorf("expected source 'gsx', got %q", diag.Source)
			}
			if tt.wantMsg != "" && !containsStr(diag.Message, tt.wantMsg) {
				t.Errorf("message %q does not contain %q", diag.Message, tt.wantMsg)
			}
		})
	}
}

func TestDiagnostics_PositionMapping(t *testing.T) {
	type tc struct {
		err      *tuigen.Error
		wantLine int // 0-indexed LSP line
		wantChar int // 0-indexed LSP character
	}

	tests := map[string]tc{
		"1-indexed to 0-indexed": {
			err: &tuigen.Error{
				Pos:     tuigen.Position{Line: 3, Column: 5},
				Message: "error",
			},
			wantLine: 2, // 3-1
			wantChar: 4, // 5-1
		},
		"line 1 col 1": {
			err: &tuigen.Error{
				Pos:     tuigen.Position{Line: 1, Column: 1},
				Message: "error",
			},
			wantLine: 0,
			wantChar: 0,
		},
	}

	dp := NewDiagnosticsProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := &Document{
				URI:     "file:///test.gsx",
				Content: "package main\n",
				Version: 1,
				AST:     nil,
				Errors:  []*tuigen.Error{tt.err},
			}

			result, err := dp.Diagnose(doc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != 1 {
				t.Fatalf("expected 1 diagnostic, got %d", len(result))
			}

			diag := result[0]
			if diag.Range.Start.Line != tt.wantLine {
				t.Errorf("range start line = %d, want %d", diag.Range.Start.Line, tt.wantLine)
			}
			if diag.Range.Start.Character != tt.wantChar {
				t.Errorf("range start char = %d, want %d", diag.Range.Start.Character, tt.wantChar)
			}
		})
	}
}

func TestDiagnostics_RangeBasedHighlighting(t *testing.T) {
	dp := NewDiagnosticsProvider()

	doc := &Document{
		URI:     "file:///test.gsx",
		Content: "package main\n",
		Version: 1,
		AST:     nil,
		Errors: []*tuigen.Error{
			{
				Pos:     tuigen.Position{Line: 3, Column: 5},
				EndPos:  tuigen.Position{Line: 3, Column: 15},
				Message: "range error",
			},
		},
	}

	result, err := dp.Diagnose(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(result))
	}

	diag := result[0]
	// Start: Line 3, Col 5 -> 0-indexed: Line 2, Char 4
	if diag.Range.Start.Line != 2 || diag.Range.Start.Character != 4 {
		t.Errorf("range start = (%d, %d), want (2, 4)",
			diag.Range.Start.Line, diag.Range.Start.Character)
	}
	// End: Line 3, Col 15 -> 0-indexed: Line 2, Char 14
	if diag.Range.End.Line != 2 || diag.Range.End.Character != 14 {
		t.Errorf("range end = (%d, %d), want (2, 14)",
			diag.Range.End.Line, diag.Range.End.Character)
	}
}

func TestDiagnostics_MultipleErrors(t *testing.T) {
	dp := NewDiagnosticsProvider()

	doc := &Document{
		URI:     "file:///test.gsx",
		Content: "package main\n",
		Version: 1,
		AST:     nil,
		Errors: []*tuigen.Error{
			{
				Pos:     tuigen.Position{Line: 1, Column: 1},
				Message: "error one",
			},
			{
				Pos:     tuigen.Position{Line: 5, Column: 10},
				Message: "error two",
			},
			{
				Pos:     tuigen.Position{Line: 8, Column: 3},
				Message: "error three",
			},
		},
	}

	result, err := dp.Diagnose(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 diagnostics, got %d", len(result))
	}

	// All should be errors
	for i, d := range result {
		if d.Severity != DiagnosticSeverityError {
			t.Errorf("diagnostic[%d] severity = %d, want Error", i, d.Severity)
		}
	}
}

func TestEstimateErrorLength(t *testing.T) {
	type tc struct {
		message string
		want    int
	}

	tests := map[string]tc{
		"quoted token single": {
			message: "unexpected 'foo'",
			want:    3,
		},
		"quoted token double": {
			message: `expected "bar"`,
			want:    3,
		},
		"quoted token backtick": {
			message: "unknown `element`",
			want:    7,
		},
		"no quotes, last word": {
			message: "unexpected token",
			want:    5, // len("token")
		},
		"trailing punctuation": {
			message: "missing semicolon.",
			want:    9, // len("semicolon")
		},
		"empty message": {
			message: "",
			want:    1, // minimum
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := estimateErrorLength(tt.message)
			if got != tt.want {
				t.Errorf("estimateErrorLength(%q) = %d, want %d", tt.message, got, tt.want)
			}
		})
	}
}

func TestDiagnostics_ErrorHighlightWidth(t *testing.T) {
	// Regression: error highlights should use message content, not hardcoded width.
	dp := NewDiagnosticsProvider()
	doc := &Document{
		URI:     "file:///test.gsx",
		Content: "package main\n",
		Version: 1,
		AST:     nil,
		Errors: []*tuigen.Error{
			{
				Pos:     tuigen.Position{Line: 1, Column: 1},
				Message: "unexpected 'xyz'",
			},
		},
	}

	result, err := dp.Diagnose(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(result))
	}

	width := result[0].Range.End.Character - result[0].Range.Start.Character
	if width == 10 {
		t.Error("error highlight width should not be hardcoded to 10")
	}
	if width != 3 {
		t.Errorf("error highlight width = %d, want 3 (length of 'xyz')", width)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
