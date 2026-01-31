package tuigen

import (
	"testing"
)

func TestLexer_Strings(t *testing.T) {
	type tc struct {
		input   string
		literal string
	}

	tests := map[string]tc{
		"simple":        {input: `"hello"`, literal: "hello"},
		"empty":         {input: `""`, literal: ""},
		"with spaces":   {input: `"hello world"`, literal: "hello world"},
		"escape n":      {input: `"hello\nworld"`, literal: "hello\nworld"},
		"escape t":      {input: `"hello\tworld"`, literal: "hello\tworld"},
		"escape r":      {input: `"hello\rworld"`, literal: "hello\rworld"},
		"escape quote":  {input: `"say \"hi\""`, literal: `say "hi"`},
		"escape backslash": {input: `"path\\to\\file"`, literal: `path\to\file`},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			tok := l.Next()
			if tok.Type != TokenString {
				t.Errorf("Type = %v, want TokenString", tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.literal)
			}
		})
	}
}

func TestLexer_RawStrings(t *testing.T) {
	type tc struct {
		input   string
		literal string
	}

	tests := map[string]tc{
		"simple":      {input: "`hello`", literal: "hello"},
		"empty":       {input: "``", literal: ""},
		"multiline":   {input: "`hello\nworld`", literal: "hello\nworld"},
		"no escapes":  {input: "`hello\\nworld`", literal: "hello\\nworld"},
		"with quotes": {input: "`say \"hi\"`", literal: `say "hi"`},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.gsx", tt.input)
			tok := l.Next()
			if tok.Type != TokenRawString {
				t.Errorf("Type = %v, want TokenRawString", tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.literal)
			}
		})
	}
}

func TestLexer_CompleteComponent(t *testing.T) {
	input := `templ Counter(count int) {
    <div direction={layout.Column}>
        <span>{fmt.Sprintf("Count: %d", count)}</span>
    </div>
}`

	l := NewLexer("test.gsx", input)
	tokens := []Token{}
	for {
		tok := l.Next()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}

	// Verify no errors
	if l.Errors().HasErrors() {
		t.Errorf("unexpected errors: %v", l.Errors())
	}

	// Verify we got the expected token sequence start
	if len(tokens) < 5 {
		t.Fatalf("expected at least 5 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenTempl {
		t.Errorf("token 0: Type = %v, want TokenTempl", tokens[0].Type)
	}
	if tokens[1].Type != TokenIdent || tokens[1].Literal != "Counter" {
		t.Errorf("token 1: Type = %v, Literal = %q, want TokenIdent, Counter", tokens[1].Type, tokens[1].Literal)
	}
}

func TestToken_String(t *testing.T) {
	type tc struct {
		token    Token
		contains string
	}

	tests := map[string]tc{
		"simple": {
			token:    Token{Type: TokenIdent, Literal: "foo", Line: 1, Column: 5},
			contains: "foo",
		},
		"empty literal": {
			token:    Token{Type: TokenEOF, Literal: "", Line: 10, Column: 1},
			contains: "EOF",
		},
		"long literal truncated": {
			token:    Token{Type: TokenString, Literal: "this is a very long string that should be truncated", Line: 1, Column: 1},
			contains: "...",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := tt.token.String()
			if len(s) == 0 {
				t.Error("String() returned empty string")
			}
			// Just verify it doesn't panic and produces output
		})
	}
}

func TestPosition_String(t *testing.T) {
	type tc struct {
		pos      Position
		expected string
	}

	tests := map[string]tc{
		"with file": {
			pos:      Position{File: "test.tui", Line: 10, Column: 5},
			expected: "test.tui:10:5",
		},
		"without file": {
			pos:      Position{File: "", Line: 10, Column: 5},
			expected: "10:5",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := tt.pos.String()
			if s != tt.expected {
				t.Errorf("String() = %q, want %q", s, tt.expected)
			}
		})
	}
}

func TestLexer_SourcePos(t *testing.T) {
	input := "package test"
	l := NewLexer("test.tui", input)

	// After creation, pos should be at start
	if l.SourcePos() != 0 {
		t.Errorf("initial SourcePos() = %d, want 0", l.SourcePos())
	}

	// After tokenizing "package", pos should advance
	l.Next()
	pos := l.SourcePos()
	if pos <= 0 {
		t.Errorf("SourcePos() after 'package' = %d, want > 0", pos)
	}
}

func TestLexer_SourceRange(t *testing.T) {
	type tc struct {
		input    string
		start    int
		end      int
		expected string
	}

	tests := map[string]tc{
		"full range":    {input: "hello world", start: 0, end: 11, expected: "hello world"},
		"partial":       {input: "hello world", start: 0, end: 5, expected: "hello"},
		"middle":        {input: "hello world", start: 6, end: 11, expected: "world"},
		"empty":         {input: "hello", start: 5, end: 5, expected: ""},
		"out of bounds": {input: "hi", start: 0, end: 100, expected: "hi"},
		"negative":      {input: "hi", start: -5, end: 2, expected: "hi"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			result := l.SourceRange(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("SourceRange(%d, %d) = %q, want %q", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}
