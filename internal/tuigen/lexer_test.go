package tuigen

import (
	"testing"
)

func TestLexer_BasicTokens(t *testing.T) {
	type tc struct {
		input    string
		expected []Token
	}

	tests := map[string]tc{
		"empty": {
			input:    "",
			expected: []Token{{Type: TokenEOF, Literal: "", Line: 1, Column: 1}},
		},
		"single paren": {
			input: "()",
			expected: []Token{
				{Type: TokenLParen, Literal: "(", Line: 1, Column: 1},
				{Type: TokenRParen, Literal: ")", Line: 1, Column: 2},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 3},
			},
		},
		"braces": {
			input: "{}",
			expected: []Token{
				{Type: TokenLBrace, Literal: "{", Line: 1, Column: 1},
				{Type: TokenRBrace, Literal: "}", Line: 1, Column: 2},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 3},
			},
		},
		"brackets": {
			input: "[]",
			expected: []Token{
				{Type: TokenLBracket, Literal: "[", Line: 1, Column: 1},
				{Type: TokenRBracket, Literal: "]", Line: 1, Column: 2},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 3},
			},
		},
		"angles": {
			input: "<>",
			expected: []Token{
				{Type: TokenLAngle, Literal: "<", Line: 1, Column: 1},
				{Type: TokenRAngle, Literal: ">", Line: 1, Column: 2},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 3},
			},
		},
		"closing tag": {
			input: "</",
			expected: []Token{
				{Type: TokenLAngleSlash, Literal: "</", Line: 1, Column: 1},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 3},
			},
		},
		"self close": {
			input: "/>",
			expected: []Token{
				{Type: TokenSlashAngle, Literal: "/>", Line: 1, Column: 1},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 3},
			},
		},
		"colon equals": {
			input: ":=",
			expected: []Token{
				{Type: TokenColonEquals, Literal: ":=", Line: 1, Column: 1},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 3},
			},
		},
		"single colon": {
			input: ":",
			expected: []Token{
				{Type: TokenColon, Literal: ":", Line: 1, Column: 1},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 2},
			},
		},
		"operators": {
			input: "+-*/&|!",
			expected: []Token{
				{Type: TokenPlus, Literal: "+", Line: 1, Column: 1},
				{Type: TokenMinus, Literal: "-", Line: 1, Column: 2},
				{Type: TokenStar, Literal: "*", Line: 1, Column: 3},
				{Type: TokenSlash, Literal: "/", Line: 1, Column: 4},
				{Type: TokenAmpersand, Literal: "&", Line: 1, Column: 5},
				{Type: TokenPipe, Literal: "|", Line: 1, Column: 6},
				{Type: TokenBang, Literal: "!", Line: 1, Column: 7},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 8},
			},
		},
		"punctuation": {
			input: ",.;=",
			expected: []Token{
				{Type: TokenComma, Literal: ",", Line: 1, Column: 1},
				{Type: TokenDot, Literal: ".", Line: 1, Column: 2},
				{Type: TokenSemicolon, Literal: ";", Line: 1, Column: 3},
				{Type: TokenEquals, Literal: "=", Line: 1, Column: 4},
				{Type: TokenEOF, Literal: "", Line: 1, Column: 5},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
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

func TestLexer_Keywords(t *testing.T) {
	type tc struct {
		input        string
		expectedType TokenType
	}

	tests := map[string]tc{
		"package": {input: "package", expectedType: TokenPackage},
		"import":  {input: "import", expectedType: TokenImport},
		"func":    {input: "func", expectedType: TokenFunc},
		"return":  {input: "return", expectedType: TokenReturn},
		"if":      {input: "if", expectedType: TokenIf},
		"else":    {input: "else", expectedType: TokenElse},
		"for":     {input: "for", expectedType: TokenFor},
		"range":   {input: "range", expectedType: TokenRange},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			tok := l.Next()
			if tok.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", tok.Type, tt.expectedType)
			}
			if tok.Literal != tt.input {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.input)
			}
		})
	}
}

func TestLexer_DSLKeywords(t *testing.T) {
	type tc struct {
		input        string
		expectedType TokenType
		literal      string
	}

	tests := map[string]tc{
		"@let":  {input: "@let", expectedType: TokenAtLet, literal: "@let"},
		"@for":  {input: "@for", expectedType: TokenAtFor, literal: "@for"},
		"@if":   {input: "@if", expectedType: TokenAtIf, literal: "@if"},
		"@else": {input: "@else", expectedType: TokenAtElse, literal: "@else"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			tok := l.Next()
			if tok.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", tok.Type, tt.expectedType)
			}
			if tok.Literal != tt.literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.literal)
			}
		})
	}
}

func TestLexer_Identifiers(t *testing.T) {
	type tc struct {
		input   string
		literal string
	}

	tests := map[string]tc{
		"simple":         {input: "foo", literal: "foo"},
		"with numbers":   {input: "foo123", literal: "foo123"},
		"with underscore": {input: "foo_bar", literal: "foo_bar"},
		"starts underscore": {input: "_foo", literal: "_foo"},
		"uppercase":      {input: "FooBar", literal: "FooBar"},
		"mixed":          {input: "myVar2_test", literal: "myVar2_test"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			tok := l.Next()
			if tok.Type != TokenIdent {
				t.Errorf("Type = %v, want TokenIdent", tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.literal)
			}
		})
	}
}

func TestLexer_Numbers(t *testing.T) {
	type tc struct {
		input        string
		expectedType TokenType
		literal      string
	}

	tests := map[string]tc{
		"integer":       {input: "123", expectedType: TokenInt, literal: "123"},
		"zero":          {input: "0", expectedType: TokenInt, literal: "0"},
		"large":         {input: "999999", expectedType: TokenInt, literal: "999999"},
		"float":         {input: "1.23", expectedType: TokenFloat, literal: "1.23"},
		"float no int":  {input: ".5", expectedType: TokenFloat, literal: ".5"},
		"exponent":      {input: "1e10", expectedType: TokenFloat, literal: "1e10"},
		"exponent neg":  {input: "1e-5", expectedType: TokenFloat, literal: "1e-5"},
		"exponent pos":  {input: "1E+5", expectedType: TokenFloat, literal: "1E+5"},
		"full float":    {input: "3.14e-10", expectedType: TokenFloat, literal: "3.14e-10"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			tok := l.Next()
			if tok.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", tok.Type, tt.expectedType)
			}
			if tok.Literal != tt.literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.literal)
			}
		})
	}
}

func TestLexer_Whitespace(t *testing.T) {
	type tc struct {
		input    string
		expected []TokenType
	}

	tests := map[string]tc{
		"spaces ignored": {
			input:    "a   b",
			expected: []TokenType{TokenIdent, TokenIdent, TokenEOF},
		},
		"tabs ignored": {
			input:    "a\t\tb",
			expected: []TokenType{TokenIdent, TokenIdent, TokenEOF},
		},
		"newlines kept": {
			input:    "a\nb",
			expected: []TokenType{TokenIdent, TokenNewline, TokenIdent, TokenEOF},
		},
		"mixed": {
			input:    "a \t\n  b",
			expected: []TokenType{TokenIdent, TokenNewline, TokenIdent, TokenEOF},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			for i, expectedType := range tt.expected {
				tok := l.Next()
				if tok.Type != expectedType {
					t.Errorf("token %d: Type = %v, want %v", i, tok.Type, expectedType)
				}
			}
		})
	}
}

func TestLexer_Comments(t *testing.T) {
	type tc struct {
		input    string
		expected []TokenType
	}

	tests := map[string]tc{
		"line comment": {
			input:    "a // comment\nb",
			expected: []TokenType{TokenIdent, TokenNewline, TokenIdent, TokenEOF},
		},
		"block comment": {
			input:    "a /* comment */ b",
			expected: []TokenType{TokenIdent, TokenIdent, TokenEOF},
		},
		"multiline block": {
			input:    "a /* multi\nline */ b",
			expected: []TokenType{TokenIdent, TokenIdent, TokenEOF},
		},
		"comment at end": {
			input:    "a // comment",
			expected: []TokenType{TokenIdent, TokenEOF},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			for i, expectedType := range tt.expected {
				tok := l.Next()
				if tok.Type != expectedType {
					t.Errorf("token %d: Type = %v, want %v", i, tok.Type, expectedType)
				}
			}
		})
	}
}

func TestLexer_LineTracking(t *testing.T) {
	input := `package test

import "fmt"

templ Foo() {
    <span>Hello</span>
}`

	type tc struct {
		expectedType TokenType
		line         int
	}

	expected := []tc{
		{TokenPackage, 1},
		{TokenIdent, 1}, // test
		{TokenNewline, 1},
		{TokenNewline, 2},
		{TokenImport, 3},
		{TokenString, 3}, // "fmt"
		{TokenNewline, 3},
		{TokenNewline, 4},
		{TokenTempl, 5},
		{TokenIdent, 5}, // Foo
		{TokenLParen, 5},
		{TokenRParen, 5},
		{TokenLBrace, 5},
		{TokenNewline, 5},
		{TokenLAngle, 6},
		{TokenIdent, 6}, // span
		{TokenRAngle, 6},
		{TokenIdent, 6}, // Hello
		{TokenLAngleSlash, 6},
		{TokenIdent, 6}, // span
		{TokenRAngle, 6},
		{TokenNewline, 6},
		{TokenRBrace, 7},
		{TokenEOF, 7},
	}

	l := NewLexer("test.gsx", input)
	for i, tt := range expected {
		tok := l.Next()
		if tok.Type != tt.expectedType {
			t.Errorf("token %d: Type = %v, want %v", i, tok.Type, tt.expectedType)
		}
		if tok.Line != tt.line {
			t.Errorf("token %d (%v): Line = %d, want %d", i, tok.Type, tok.Line, tt.line)
		}
	}
}

func TestLexer_XMLLikeTokens(t *testing.T) {
	input := `<div direction={layout.Column}>
    <span>Hello</span>
    <input />
</div>`

	type tc struct {
		expectedType TokenType
		literal      string
	}

	expected := []tc{
		{TokenLAngle, "<"},
		{TokenIdent, "div"},
		{TokenIdent, "direction"},
		{TokenEquals, "="},
		// Note: { starts a Go expression context, parser handles this
		{TokenLBrace, "{"},
		// Parser would call ReadGoExpr here
	}

	l := NewLexer("test.gsx", input)
	for i, tt := range expected {
		tok := l.Next()
		if tok.Type != tt.expectedType {
			t.Errorf("token %d: Type = %v, want %v", i, tok.Type, tt.expectedType)
		}
		if tok.Literal != tt.literal {
			t.Errorf("token %d: Literal = %q, want %q", i, tok.Literal, tt.literal)
		}
	}
}

func TestLexer_ErrorCases(t *testing.T) {
	type tc struct {
		input         string
		hasError      bool
		errorContains string
	}

	tests := map[string]tc{
		"unclosed string": {
			input:         `"hello`,
			hasError:      true,
			errorContains: "unterminated string",
		},
		"unclosed raw string": {
			input:         "`hello",
			hasError:      true,
			errorContains: "unterminated raw string",
		},
		"unknown @ keyword": {
			input:         "@unknown",
			hasError:      true,
			errorContains: "unknown @ keyword",
		},
		"unclosed block comment": {
			input:         "/* unclosed",
			hasError:      true,
			errorContains: "unterminated block comment",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			// Consume all tokens
			for {
				tok := l.Next()
				if tok.Type == TokenEOF {
					break
				}
			}

			if tt.hasError && !l.Errors().HasErrors() {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestLexer_Underscore(t *testing.T) {
	type tc struct {
		input        string
		expectedType TokenType
		literal      string
	}

	tests := map[string]tc{
		"standalone underscore": {input: "_", expectedType: TokenUnderscore, literal: "_"},
		"underscore identifier": {input: "_foo", expectedType: TokenIdent, literal: "_foo"},
		"underscore in range":   {input: "_, x := range items", expectedType: TokenUnderscore, literal: "_"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			tok := l.Next()
			if tok.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", tok.Type, tt.expectedType)
			}
			if tok.Literal != tt.literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.literal)
			}
		})
	}
}

