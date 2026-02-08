package tuigen

import (
	"testing"
)

func TestLexer_GoExpressions(t *testing.T) {
	type tc struct {
		input   string
		literal string
	}

	tests := map[string]tc{
		"simple":           {input: "{x}", literal: "x"},
		"with spaces":      {input: "{ x + y }", literal: " x + y "},
		"nested braces":    {input: "{map[string]int{}}", literal: "map[string]int{}"},
		"deeply nested":    {input: "{func() { if true { x } }()}", literal: "func() { if true { x } }()"},
		"with string":      {input: `{fmt.Sprintf("%d", x)}`, literal: `fmt.Sprintf("%d", x)`},
		"with raw string":  {input: "{`hello`}", literal: "`hello`"},
		"function call":    {input: "{onClick()}", literal: "onClick()"},
		"method call":      {input: "{s.Method(a, b)}", literal: "s.Method(a, b)"},
		"struct literal":   {input: "{Point{X: 1, Y: 2}}", literal: "Point{X: 1, Y: 2}"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			// First, consume the opening brace via Next()
			// ReadGoExpr expects to be called after { was tokenized
			brace := l.Next()
			if brace.Type != TokenLBrace {
				t.Fatalf("expected TokenLBrace, got %v", brace.Type)
			}
			tok := l.ReadGoExpr()
			if tok.Type != TokenGoExpr {
				t.Errorf("Type = %v, want TokenGoExpr", tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.literal)
			}
		})
	}
}

func TestLexer_ReadBalancedBracesFrom(t *testing.T) {
	type tc struct {
		input    string
		startPos int
		expected string
		hasError bool
	}

	tests := map[string]tc{
		"simple":          {input: "{x}", startPos: 0, expected: "x"},
		"with spaces":     {input: "{ x + y }", startPos: 0, expected: " x + y "},
		"nested braces":   {input: "{map[string]int{}}", startPos: 0, expected: "map[string]int{}"},
		"with string":     {input: `{fmt.Sprintf("%d", x)}`, startPos: 0, expected: `fmt.Sprintf("%d", x)`},
		"with raw string": {input: "{`hello`}", startPos: 0, expected: "`hello`"},
		"deeply nested":   {input: "{func() { if true { x } }()}", startPos: 0, expected: "func() { if true { x } }()"},
		"not at brace":    {input: "abc{x}", startPos: 0, hasError: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			result, err := l.ReadBalancedBracesFrom(tt.startPos)

			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ReadBalancedBracesFrom(%d) = %q, want %q", tt.startPos, result, tt.expected)
			}
		})
	}
}

func TestLexer_ComponentCall(t *testing.T) {
	type tc struct {
		input       string
		wantType    TokenType
		wantLiteral string
	}

	tests := map[string]tc{
		"simple call": {
			input:       "@Card",
			wantType:    TokenAtCall,
			wantLiteral: "Card",
		},
		"multi-word name": {
			input:       "@MyCustomComponent",
			wantType:    TokenAtCall,
			wantLiteral: "MyCustomComponent",
		},
		"header component": {
			input:       "@Header",
			wantType:    TokenAtCall,
			wantLiteral: "Header",
		},
		"lowercase is component expr": {
			input:       "@unknown",
			wantType:    TokenAtExpr,
			wantLiteral: "unknown",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			tok := l.Next()

			if tok.Type != tt.wantType {
				t.Errorf("token type = %v, want %v", tok.Type, tt.wantType)
			}
			if tt.wantLiteral != "" && tok.Literal != tt.wantLiteral {
				t.Errorf("token literal = %q, want %q", tok.Literal, tt.wantLiteral)
			}
		})
	}
}
