package tuigen

import (
	"testing"
)

func TestLexer_UnicodeSymbols(t *testing.T) {
	type tc struct {
		source   string
		expected []TokenType
	}

	tests := map[string]tc{
		"arrow symbols in text": {
			source: `<span>Use ↑/↓ to navigate</span>`,
			expected: []TokenType{
				TokenLAngle, TokenIdent, TokenRAngle, // <span>
				TokenIdent, TokenSymbol, TokenSlash, TokenSymbol, TokenIdent, TokenIdent, // Use ↑/↓ to navigate
				TokenLAngleSlash, TokenIdent, TokenRAngle, // </span>
				TokenEOF,
			},
		},
		"various unicode arrows": {
			source: `<div>← → ↑ ↓</div>`,
			expected: []TokenType{
				TokenLAngle, TokenIdent, TokenRAngle, // <div>
				TokenSymbol, TokenSymbol, TokenSymbol, TokenSymbol, // ← → ↑ ↓
				TokenLAngleSlash, TokenIdent, TokenRAngle, // </div>
				TokenEOF,
			},
		},
		"emoji in text": {
			source: `<span>Hello 👋 World 🌍</span>`,
			expected: []TokenType{
				TokenLAngle, TokenIdent, TokenRAngle, // <span>
				TokenIdent, TokenSymbol, TokenIdent, TokenSymbol, // Hello 👋 World 🌍
				TokenLAngleSlash, TokenIdent, TokenRAngle, // </span>
				TokenEOF,
			},
		},
		"mixed unicode and ascii": {
			source: `<span>Count: 5 ↑</span>`,
			expected: []TokenType{
				TokenLAngle, TokenIdent, TokenRAngle, // <span>
				TokenIdent, TokenColon, TokenInt, TokenSymbol, // Count: 5 ↑
				TokenLAngleSlash, TokenIdent, TokenRAngle, // </span>
				TokenEOF,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := NewLexer("test.gsx", tt.source)

			var tokens []TokenType
			for {
				tok := lexer.Next()
				tokens = append(tokens, tok.Type)
				if tok.Type == TokenEOF {
					break
				}
			}

			if len(tokens) != len(tt.expected) {
				t.Errorf("expected %d tokens, got %d", len(tt.expected), len(tokens))
				t.Logf("expected: %v", tt.expected)
				t.Logf("got:      %v", tokens)
				return
			}

			for i, expected := range tt.expected {
				if tokens[i] != expected {
					t.Errorf("token %d: expected %s, got %s", i, expected, tokens[i])
				}
			}

			// Check for errors
			if lexer.Errors().HasErrors() {
				t.Errorf("unexpected lexer errors: %v", lexer.Errors())
			}
		})
	}
}

func TestParser_UnicodeTextContent(t *testing.T) {
	type tc struct {
		source      string
		expectError bool
	}

	tests := map[string]tc{
		"arrows in span text": {
			source: `package test
templ Test() {
	<span>Use ↑/↓ to navigate</span>
}`,
			expectError: false,
		},
		"arrows in attribute string": {
			source: `package test
templ Test() {
	<span text="Use ↑/↓ to navigate"></span>
}`,
			expectError: false,
		},
		"emoji in text": {
			source: `package test
templ Test() {
	<div>
		<span>Hello 👋</span>
		<span>World 🌍</span>
	</div>
}`,
			expectError: false,
		},
		"unicode in multiple elements": {
			source: `package test
templ Test() {
	<div>
		<span>← Previous</span>
		<span>Next →</span>
	</div>
}`,
			expectError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := NewLexer("test.gsx", tt.source)
			parser := NewParser(lexer)
			_, _ = parser.ParseFile()

			hasErrors := parser.Errors().HasErrors()
			if hasErrors != tt.expectError {
				if tt.expectError {
					t.Errorf("expected errors but got none")
				} else {
					t.Errorf("unexpected errors: %v", parser.Errors())
				}
			}
		})
	}
}
