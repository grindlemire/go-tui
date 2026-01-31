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

func TestLexer_HashToken(t *testing.T) {
	type tc struct {
		input    string
		expected []Token
	}

	tests := map[string]tc{
		"hash alone": {
			input: "#",
			expected: []Token{
				{Type: TokenHash, Literal: "#", Line: 1, Column: 1},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 2},
			},
		},
		"hash followed by identifier": {
			input: "#Content",
			expected: []Token{
				{Type: TokenHash, Literal: "#", Line: 1, Column: 1},
				{Type: TokenIdent, Literal: "Content", Line: 1, Column: 2},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 9},
			},
		},
		"hash in element context": {
			input: "<div #Content>",
			expected: []Token{
				{Type: TokenLAngle, Literal: "<", Line: 1, Column: 1},
				{Type: TokenIdent, Literal: "div", Line: 1, Column: 2},
				{Type: TokenHash, Literal: "#", Line: 1, Column: 6},
				{Type: TokenIdent, Literal: "Content", Line: 1, Column: 7},
				{Type: TokenRAngle, Literal: ">", Line: 1, Column: 14},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 15},
			},
		},
		"hash with attributes": {
			input: "<span #Title class=\"bold\">",
			expected: []Token{
				{Type: TokenLAngle, Literal: "<", Line: 1, Column: 1},
				{Type: TokenIdent, Literal: "span", Line: 1, Column: 2},
				{Type: TokenHash, Literal: "#", Line: 1, Column: 7},
				{Type: TokenIdent, Literal: "Title", Line: 1, Column: 8},
				{Type: TokenIdent, Literal: "class", Line: 1, Column: 14},
				{Type: TokenEquals, Literal: "=", Line: 1, Column: 19},
				{Type: TokenString, Literal: "bold", Line: 1, Column: 20},
				{Type: TokenRAngle, Literal: ">", Line: 1, Column: 26},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 27},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			for i, expected := range tt.expected {
				tok := l.Next()
				if tok.Type != expected.Type {
					t.Errorf("token %d: Type = %v, want %v", i, tok.Type, expected.Type)
				}
				if tok.Literal != expected.Literal {
					t.Errorf("token %d: Literal = %q, want %q", i, tok.Literal, expected.Literal)
				}
				if tok.Line != expected.Line {
					t.Errorf("token %d: Line = %d, want %d", i, tok.Line, expected.Line)
				}
				if tok.Column != expected.Column {
					t.Errorf("token %d: Column = %d, want %d", i, tok.Column, expected.Column)
				}
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
		"lowercase still keyword error": {
			input:    "@unknown",
			wantType: TokenError,
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
