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
			l := NewLexer("test.tui", tt.input)
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
		"@component": {input: "@component", expectedType: TokenAtComponent, literal: "@component"},
		"@let":       {input: "@let", expectedType: TokenAtLet, literal: "@let"},
		"@for":       {input: "@for", expectedType: TokenAtFor, literal: "@for"},
		"@if":        {input: "@if", expectedType: TokenAtIf, literal: "@if"},
		"@else":      {input: "@else", expectedType: TokenAtElse, literal: "@else"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
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
			l := NewLexer("test.tui", tt.input)
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
			l := NewLexer("test.tui", tt.input)
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
			l := NewLexer("test.tui", tt.input)
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
			l := NewLexer("test.tui", tt.input)
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
			l := NewLexer("test.tui", tt.input)
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
			l := NewLexer("test.tui", tt.input)
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
			l := NewLexer("test.tui", tt.input)
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

@component Foo() {
    <text>Hello</text>
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
		{TokenAtComponent, 5},
		{TokenIdent, 5}, // Foo
		{TokenLParen, 5},
		{TokenRParen, 5},
		{TokenLBrace, 5},
		{TokenNewline, 5},
		{TokenLAngle, 6},
		{TokenIdent, 6}, // text
		{TokenRAngle, 6},
		{TokenIdent, 6}, // Hello
		{TokenLAngleSlash, 6},
		{TokenIdent, 6}, // text
		{TokenRAngle, 6},
		{TokenNewline, 6},
		{TokenRBrace, 7},
		{TokenEOF, 7},
	}

	l := NewLexer("test.tui", input)
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
	input := `<box direction={layout.Column}>
    <text>Hello</text>
    <input />
</box>`

	type tc struct {
		expectedType TokenType
		literal      string
	}

	expected := []tc{
		{TokenLAngle, "<"},
		{TokenIdent, "box"},
		{TokenIdent, "direction"},
		{TokenEquals, "="},
		// Note: { starts a Go expression context, parser handles this
		{TokenLBrace, "{"},
		// Parser would call ReadGoExpr here
	}

	l := NewLexer("test.tui", input)
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
			l := NewLexer("test.tui", tt.input)
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
			l := NewLexer("test.tui", tt.input)
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

func TestLexer_CompleteComponent(t *testing.T) {
	input := `@component Counter(count int) {
    <box direction={layout.Column}>
        <text>{fmt.Sprintf("Count: %d", count)}</text>
    </box>
}`

	l := NewLexer("test.tui", input)
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
	if tokens[0].Type != TokenAtComponent {
		t.Errorf("token 0: Type = %v, want TokenAtComponent", tokens[0].Type)
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
		"component keyword": {
			input:       "@component",
			wantType:    TokenAtComponent,
			wantLiteral: "@component",
		},
		"lowercase still keyword error": {
			input:    "@unknown",
			wantType: TokenError,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
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
