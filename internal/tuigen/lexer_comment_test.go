package tuigen

import (
	"testing"
)

func TestLexer_CommentCollection_LineComment(t *testing.T) {
	type tc struct {
		input        string
		wantText     string
		wantLine     int
		wantEndLine  int
		wantIsBlock  bool
	}

	tests := map[string]tc{
		"simple line comment": {
			input:       "// hello",
			wantText:    "// hello",
			wantLine:    1,
			wantEndLine: 1,
			wantIsBlock: false,
		},
		"line comment with spaces": {
			input:       "//   spaced   ",
			wantText:    "//   spaced   ",
			wantLine:    1,
			wantEndLine: 1,
			wantIsBlock: false,
		},
		"empty line comment": {
			input:       "//",
			wantText:    "//",
			wantLine:    1,
			wantEndLine: 1,
			wantIsBlock: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			// Trigger comment collection by calling Next
			tok := l.Next()
			if tok.Type != TokenEOF {
				t.Errorf("expected EOF after comment, got %v", tok.Type)
			}

			comments := l.ConsumeComments()
			if len(comments) != 1 {
				t.Fatalf("expected 1 comment, got %d", len(comments))
			}

			c := comments[0]
			if c.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", c.Text, tt.wantText)
			}
			if c.Position.Line != tt.wantLine {
				t.Errorf("Position.Line = %d, want %d", c.Position.Line, tt.wantLine)
			}
			if c.EndLine != tt.wantEndLine {
				t.Errorf("EndLine = %d, want %d", c.EndLine, tt.wantEndLine)
			}
			if c.IsBlock != tt.wantIsBlock {
				t.Errorf("IsBlock = %v, want %v", c.IsBlock, tt.wantIsBlock)
			}
		})
	}
}

func TestLexer_CommentCollection_BlockComment(t *testing.T) {
	type tc struct {
		input        string
		wantText     string
		wantLine     int
		wantEndLine  int
		wantIsBlock  bool
	}

	tests := map[string]tc{
		"simple block comment": {
			input:       "/* hello */",
			wantText:    "/* hello */",
			wantLine:    1,
			wantEndLine: 1,
			wantIsBlock: true,
		},
		"empty block comment": {
			input:       "/**/",
			wantText:    "/**/",
			wantLine:    1,
			wantEndLine: 1,
			wantIsBlock: true,
		},
		"block comment with asterisks": {
			input:       "/* * ** *** */",
			wantText:    "/* * ** *** */",
			wantLine:    1,
			wantEndLine: 1,
			wantIsBlock: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := NewLexer("test.tui", tt.input)
			// Trigger comment collection by calling Next
			tok := l.Next()
			if tok.Type != TokenEOF {
				t.Errorf("expected EOF after comment, got %v", tok.Type)
			}

			comments := l.ConsumeComments()
			if len(comments) != 1 {
				t.Fatalf("expected 1 comment, got %d", len(comments))
			}

			c := comments[0]
			if c.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", c.Text, tt.wantText)
			}
			if c.Position.Line != tt.wantLine {
				t.Errorf("Position.Line = %d, want %d", c.Position.Line, tt.wantLine)
			}
			if c.EndLine != tt.wantEndLine {
				t.Errorf("EndLine = %d, want %d", c.EndLine, tt.wantEndLine)
			}
			if c.IsBlock != tt.wantIsBlock {
				t.Errorf("IsBlock = %v, want %v", c.IsBlock, tt.wantIsBlock)
			}
		})
	}
}

func TestLexer_CommentCollection_MultilineBlock(t *testing.T) {
	input := `/* line 1
line 2
line 3 */`

	l := NewLexer("test.tui", input)
	tok := l.Next()
	if tok.Type != TokenEOF {
		t.Errorf("expected EOF after comment, got %v", tok.Type)
	}

	comments := l.ConsumeComments()
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	c := comments[0]
	if c.Text != input {
		t.Errorf("Text = %q, want %q", c.Text, input)
	}
	if c.Position.Line != 1 {
		t.Errorf("Position.Line = %d, want 1", c.Position.Line)
	}
	if c.EndLine != 3 {
		t.Errorf("EndLine = %d, want 3", c.EndLine)
	}
	if !c.IsBlock {
		t.Error("IsBlock should be true for block comment")
	}
}

func TestLexer_CommentCollection_UnterminatedBlockComment(t *testing.T) {
	input := "/* unclosed"

	l := NewLexer("test.tui", input)
	// Consume all tokens
	for {
		tok := l.Next()
		if tok.Type == TokenEOF {
			break
		}
	}

	if !l.Errors().HasErrors() {
		t.Error("expected error for unterminated block comment")
	}
}

func TestLexer_CommentCollection_BetweenTokens(t *testing.T) {
	input := "a // comment\nb"

	l := NewLexer("test.tui", input)

	// First token: identifier 'a'
	tok := l.Next()
	if tok.Type != TokenIdent || tok.Literal != "a" {
		t.Errorf("token 0: got %v %q, want Ident 'a'", tok.Type, tok.Literal)
	}

	// No comments collected yet (comment is after 'a', not before)
	comments := l.ConsumeComments()
	if len(comments) != 0 {
		t.Errorf("expected 0 comments before 'a', got %d", len(comments))
	}

	// Newline token
	tok = l.Next()
	if tok.Type != TokenNewline {
		t.Errorf("token 1: got %v, want Newline", tok.Type)
	}

	// Comment should be collected before the newline is returned
	// Actually, comments are collected during skipWhitespaceAndCollectComments
	// which happens at the START of Next(), so the comment will be collected
	// when we call Next() after 'a' (which gives us newline)
	// Let's re-check: Next() first calls skipWhitespaceAndCollectComments
	// After 'a', we're at ' ', so we skip space, then see '//' which collects comment
	// So comment IS collected before returning newline

	// Now get identifier 'b'
	tok = l.Next()
	if tok.Type != TokenIdent || tok.Literal != "b" {
		t.Errorf("token 2: got %v %q, want Ident 'b'", tok.Type, tok.Literal)
	}

	// The comment was collected before 'b'
	comments = l.ConsumeComments()
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment before 'b', got %d", len(comments))
	}
	if comments[0].Text != "// comment" {
		t.Errorf("comment text = %q, want %q", comments[0].Text, "// comment")
	}
}

func TestLexer_CommentCollection_ConsumeCommentsClearsBuffer(t *testing.T) {
	input := "// comment\na"

	l := NewLexer("test.tui", input)

	// First call to Next() collects comment before returning token
	tok := l.Next() // returns newline, collects comment
	if tok.Type != TokenNewline {
		t.Errorf("expected newline, got %v", tok.Type)
	}

	// Get identifier
	tok = l.Next()
	if tok.Type != TokenIdent {
		t.Errorf("expected ident, got %v", tok.Type)
	}

	// Consume comments - should get the comment
	comments := l.ConsumeComments()
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	// Consume again - buffer should be cleared
	comments = l.ConsumeComments()
	if len(comments) != 0 {
		t.Errorf("expected 0 comments after clear, got %d", len(comments))
	}
}

func TestLexer_CommentCollection_MultipleComments(t *testing.T) {
	input := `// comment 1
// comment 2
a`

	l := NewLexer("test.tui", input)

	// First newline
	tok := l.Next()
	if tok.Type != TokenNewline {
		t.Errorf("token 0: got %v, want Newline", tok.Type)
	}

	// Second newline
	tok = l.Next()
	if tok.Type != TokenNewline {
		t.Errorf("token 1: got %v, want Newline", tok.Type)
	}

	// Identifier 'a'
	tok = l.Next()
	if tok.Type != TokenIdent {
		t.Errorf("token 2: got %v, want Ident", tok.Type)
	}

	// Should have collected both comments
	comments := l.ConsumeComments()
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
	if comments[0].Text != "// comment 1" {
		t.Errorf("comment 0 text = %q, want %q", comments[0].Text, "// comment 1")
	}
	if comments[1].Text != "// comment 2" {
		t.Errorf("comment 1 text = %q, want %q", comments[1].Text, "// comment 2")
	}
}

func TestLexer_CommentCollection_BlockAndLine(t *testing.T) {
	input := `/* block */ // line
a`

	l := NewLexer("test.tui", input)

	// Newline
	tok := l.Next()
	if tok.Type != TokenNewline {
		t.Errorf("token 0: got %v, want Newline", tok.Type)
	}

	// Identifier 'a'
	tok = l.Next()
	if tok.Type != TokenIdent {
		t.Errorf("token 1: got %v, want Ident", tok.Type)
	}

	// Should have collected both comments
	comments := l.ConsumeComments()
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
	if !comments[0].IsBlock {
		t.Error("comment 0 should be block comment")
	}
	if comments[1].IsBlock {
		t.Error("comment 1 should be line comment")
	}
}

func TestCommentGroup_Text(t *testing.T) {
	type tc struct {
		group    *CommentGroup
		expected string
	}

	tests := map[string]tc{
		"nil group": {
			group:    nil,
			expected: "",
		},
		"empty group": {
			group:    &CommentGroup{List: []*Comment{}},
			expected: "",
		},
		"single line comment": {
			group: &CommentGroup{
				List: []*Comment{
					{Text: "// hello world", IsBlock: false},
				},
			},
			expected: "hello world",
		},
		"single block comment": {
			group: &CommentGroup{
				List: []*Comment{
					{Text: "/* hello world */", IsBlock: true},
				},
			},
			expected: "hello world",
		},
		"multiple line comments": {
			group: &CommentGroup{
				List: []*Comment{
					{Text: "// line 1", IsBlock: false},
					{Text: "// line 2", IsBlock: false},
				},
			},
			expected: "line 1\nline 2",
		},
		"line comment no space after slashes": {
			group: &CommentGroup{
				List: []*Comment{
					{Text: "//nospace", IsBlock: false},
				},
			},
			expected: "nospace",
		},
		"block comment extra whitespace": {
			group: &CommentGroup{
				List: []*Comment{
					{Text: "/*   padded   */", IsBlock: true},
				},
			},
			expected: "padded",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.group.Text()
			if result != tt.expected {
				t.Errorf("Text() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLexer_CommentCollection_PositionTracking(t *testing.T) {
	input := `   // comment at column 4`

	l := NewLexer("test.tui", input)
	tok := l.Next()
	if tok.Type != TokenEOF {
		t.Errorf("expected EOF, got %v", tok.Type)
	}

	comments := l.ConsumeComments()
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	c := comments[0]
	if c.Position.Column != 4 {
		t.Errorf("Position.Column = %d, want 4", c.Position.Column)
	}
}

func TestLexer_ExistingCommentBehavior(t *testing.T) {
	// Verify that existing tests still pass - comments should be collected but
	// tokens should still be returned correctly

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
