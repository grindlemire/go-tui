package tuigen

import (
	"strings"
	"testing"
)

func TestLexer_ReadGoExpr_CharLiteralsAndErrors(t *testing.T) {
	type tc struct {
		input       string
		wantLiteral string
		wantType    TokenType
		wantErr     string
	}

	tests := map[string]tc{
		"char literal": {
			input:       "{x == 'a'}",
			wantLiteral: "x == 'a'",
			wantType:    TokenGoExpr,
		},
		"escaped char literal": {
			input:       `{x == '\''}`,
			wantLiteral: `x == '\''`,
			wantType:    TokenGoExpr,
		},
		"newline escape char": {
			input:       `{sep == '\n'}`,
			wantLiteral: `sep == '\n'`,
			wantType:    TokenGoExpr,
		},
		"string with escapes": {
			input:       `{s == "a\"b"}`,
			wantLiteral: `s == "a\"b"`,
			wantType:    TokenGoExpr,
		},
		"braces inside string ignored": {
			input:       `{f("}")}`,
			wantLiteral: `f("}")`,
			wantType:    TokenGoExpr,
		},
		"braces inside raw string ignored": {
			input:       "{f(`}`)}",
			wantLiteral: "f(`}`)",
			wantType:    TokenGoExpr,
		},
		"unterminated expression": {
			input:    "{x + y",
			wantType: TokenError,
			wantErr:  "unterminated Go expression: unmatched '{'",
		},
		"unmatched parentheses warning": {
			input:       "{f(}",
			wantLiteral: "f(",
			wantType:    TokenGoExpr,
			wantErr:     "unmatched parentheses",
		},
		"unmatched brackets warning": {
			input:       "{a[}",
			wantLiteral: "a[",
			wantType:    TokenGoExpr,
			wantErr:     "unmatched brackets",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			brace := l.Next()
			if brace.Type != TokenLBrace {
				t.Fatalf("expected TokenLBrace, got %v", brace.Type)
			}
			tok := l.ReadGoExpr()
			if tok.Type != tt.wantType {
				t.Errorf("token type = %v, want %v", tok.Type, tt.wantType)
			}
			if tt.wantType == TokenGoExpr && tok.Literal != tt.wantLiteral {
				t.Errorf("literal = %q, want %q", tok.Literal, tt.wantLiteral)
			}
			if tt.wantErr != "" {
				err := l.Errors().Err()
				if err == nil {
					t.Fatalf("expected lexer error containing %q, got none", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestLexer_PeekToken(t *testing.T) {
	l := NewLexer("test.gsx", "templ Foo()")

	peeked := l.PeekToken()
	if peeked.Type != TokenTempl {
		t.Fatalf("PeekToken type = %v, want TokenTempl", peeked.Type)
	}

	// Peeking must not consume: Next should return the same token.
	next := l.Next()
	if next.Type != peeked.Type || next.Literal != peeked.Literal {
		t.Errorf("Next after Peek = (%v, %q), want (%v, %q)",
			next.Type, next.Literal, peeked.Type, peeked.Literal)
	}

	// The following token should be the identifier.
	ident := l.Next()
	if ident.Type != TokenIdent || ident.Literal != "Foo" {
		t.Errorf("second token = (%v, %q), want (TokenIdent, \"Foo\")", ident.Type, ident.Literal)
	}
}

func TestLexer_ReadBalancedBraces(t *testing.T) {
	type tc struct {
		input    string
		want     string
		wantErr  string
		hasError bool
	}

	tests := map[string]tc{
		"simple": {
			input: "{x + y}",
			want:  "x + y",
		},
		"nested braces": {
			input: "{if a { b } else { c }}",
			want:  "if a { b } else { c }",
		},
		"string with brace": {
			input: `{"}"}`,
			want:  `"}"`,
		},
		"raw string with brace": {
			input: "{`}`}",
			want:  "`}`",
		},
		"char literal with brace": {
			input: "{'}'}",
			want:  "'}'",
		},
		"not at brace": {
			input:    "x{y}",
			hasError: true,
			wantErr:  "expected '{' at start of balanced braces",
		},
		"unterminated": {
			input:    "{x",
			hasError: true,
			wantErr:  "unterminated braces: unmatched '{'",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			got, err := l.ReadBalancedBraces()
			if tt.hasError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (content %q)", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("content = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLexer_ReadUntilBrace(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"stops at brace": {
			input: "x > 5 {",
			want:  "x > 5 ",
		},
		"stops at newline": {
			input: "cond\n{",
			want:  "cond",
		},
		"stops at EOF": {
			input: "a && b",
			want:  "a && b",
		},
		"skips leading whitespace": {
			input: "   y == 2 {",
			want:  "y == 2 ",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			got := l.ReadUntilBrace()
			if got != tt.want {
				t.Errorf("ReadUntilBrace() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLexer_ReadBalancedBracesFrom_Errors(t *testing.T) {
	type tc struct {
		input    string
		startPos int
		wantErr  string
	}

	tests := map[string]tc{
		"negative start": {
			input:    "{x}",
			startPos: -1,
			wantErr:  "invalid start position",
		},
		"start beyond source": {
			input:    "{x}",
			startPos: 10,
			wantErr:  "invalid start position",
		},
		"start not at brace": {
			input:    "a{x}",
			startPos: 0,
			wantErr:  "invalid start position",
		},
		"unterminated braces": {
			input:    "{x",
			startPos: 0,
			wantErr:  "unterminated braces",
		},
		"unterminated with escape in string": {
			input:    `{"a\"`,
			startPos: 0,
			wantErr:  "unterminated braces",
		},
		"unterminated with char escape": {
			input:    `{'\n`,
			startPos: 0,
			wantErr:  "unterminated braces",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			_, err := l.ReadBalancedBracesFrom(tt.startPos)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
