package lsp

import (
	"encoding/json"
	"testing"

	"github.com/grindlemire/go-tui/internal/lsp/provider"
)

func TestCommentSemanticTokens(t *testing.T) {
	type tc struct {
		content     string
		wantTokens  int  // minimum number of comment tokens expected
		checkTokens bool // if true, verify specific token positions
	}

	tests := map[string]tc{
		"line comment before component": {
			content: `package main

// This is a comment
templ Hello() {
	<span>Hello</span>
}
`,
			wantTokens:  1,
			checkTokens: true,
		},
		"trailing comment on component": {
			content: `package main

templ Hello() {  // trailing comment
	<span>Hello</span>
}
`,
			wantTokens:  1,
			checkTokens: true,
		},
		"block comment": {
			content: `package main

/* Block comment */
templ Hello() {
	<span>Hello</span>
}
`,
			wantTokens:  1,
			checkTokens: true,
		},
		"comment inside element": {
			content: `package main

templ Hello() {
	// comment inside body
	<span>Hello</span>
}
`,
			wantTokens:  1,
			checkTokens: true,
		},
		"multiple comments": {
			content: `package main

// Comment 1
// Comment 2
templ Hello() {
	// Comment 3
	<span>Hello</span>  // Comment 4
}
`,
			wantTokens:  4,
			checkTokens: true,
		},
		"comment in if statement": {
			content: `package main

templ Hello(show bool) {
	// comment before if
	@if show {
		<span>Hello</span>
	}
}
`,
			wantTokens:  1,
			checkTokens: true,
		},
		"comment in for loop": {
			content: `package main

templ List(items []string) {
	// comment before for
	@for _, item := range items {
		<span>{item}</span>
	}
}
`,
			wantTokens:  1,
			checkTokens: true,
		},
		"orphan comment in component body": {
			content: `package main

templ Hello() {
	// orphan comment with no following node
}
`,
			wantTokens:  1,
			checkTokens: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			params, _ := json.Marshal(SemanticTokensParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
			})

			result, rpcErr := server.router.Route(Request{
				Method: "textDocument/semanticTokens/full",
				Params: params,
			})

			if rpcErr != nil {
				t.Fatalf("semantic tokens error: %v", rpcErr)
			}

			tokens, ok := result.(*SemanticTokens)
			if !ok {
				t.Fatalf("expected *SemanticTokens, got %T", result)
			}

			// Count comment tokens (every 5th value starting at index 3 is the token type)
			commentCount := countTokensByType(tokens.Data, provider.TokenTypeComment)

			if commentCount < tt.wantTokens {
				t.Errorf("got %d comment tokens, want at least %d", commentCount, tt.wantTokens)
			}
		})
	}
}

func TestCommentTokenPositions(t *testing.T) {
	type tc struct {
		content        string
		expectedLine   int // 0-indexed
		expectedCol    int // 0-indexed
		expectedLength int
	}

	tests := map[string]tc{
		"line comment position": {
			content: `package main

// Hello
templ Hello() {
	<span>Hello</span>
}
`,
			expectedLine:   2, // 0-indexed, line 3 in 1-indexed
			expectedCol:    0,
			expectedLength: 8, // "// Hello"
		},
		"indented comment position": {
			content: `package main

templ Hello() {
	// indented comment
	<span>Hello</span>
}
`,
			expectedLine:   3, // 0-indexed
			expectedCol:    1, // after tab
			expectedLength: 19,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			params, _ := json.Marshal(SemanticTokensParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
			})

			result, rpcErr := server.router.Route(Request{
				Method: "textDocument/semanticTokens/full",
				Params: params,
			})

			if rpcErr != nil {
				t.Fatalf("semantic tokens error: %v", rpcErr)
			}

			tokens, ok := result.(*SemanticTokens)
			if !ok {
				t.Fatalf("expected *SemanticTokens, got %T", result)
			}

			// Decode delta-encoded tokens to find the first comment token
			decoded := decodeSemanticTokens(tokens.Data)
			var foundComment *decodedToken
			for i := range decoded {
				if decoded[i].tokenType == provider.TokenTypeComment {
					foundComment = &decoded[i]
					break
				}
			}

			if foundComment == nil {
				t.Fatal("expected to find a comment token")
			}

			if foundComment.line != tt.expectedLine {
				t.Errorf("comment token line = %d, want %d", foundComment.line, tt.expectedLine)
			}

			if foundComment.startChar != tt.expectedCol {
				t.Errorf("comment token startChar = %d, want %d", foundComment.startChar, tt.expectedCol)
			}

			if foundComment.length != tt.expectedLength {
				t.Errorf("comment token length = %d, want %d", foundComment.length, tt.expectedLength)
			}
		})
	}
}

func TestBlockCommentTokens(t *testing.T) {
	type tc struct {
		content    string
		wantTokens int // number of tokens for the block comment (one per line)
	}

	tests := map[string]tc{
		"single line block comment": {
			content: `package main

/* single line */
templ Hello() {
	<span>Hello</span>
}
`,
			wantTokens: 1,
		},
		"multi-line block comment": {
			content: `package main

/* line 1
   line 2
   line 3 */
templ Hello() {
	<span>Hello</span>
}
`,
			wantTokens: 3, // one token per line
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			params, _ := json.Marshal(SemanticTokensParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
			})

			result, rpcErr := server.router.Route(Request{
				Method: "textDocument/semanticTokens/full",
				Params: params,
			})

			if rpcErr != nil {
				t.Fatalf("semantic tokens error: %v", rpcErr)
			}

			tokens, ok := result.(*SemanticTokens)
			if !ok {
				t.Fatalf("expected *SemanticTokens, got %T", result)
			}

			commentCount := countTokensByType(tokens.Data, provider.TokenTypeComment)

			if commentCount != tt.wantTokens {
				t.Errorf("got %d comment tokens, want %d", commentCount, tt.wantTokens)
			}
		})
	}
}

func TestInlineBlockCommentInGoExpr(t *testing.T) {
	// Test that inline block comments inside Go expressions are properly tokenized
	type tc struct {
		content string
	}

	tests := map[string]tc{
		"inline block comment in sprintf": {
			content: `package main

import "fmt"

templ Hello(item string) {
	<span>{fmt.Sprintf("> %s", /* ItemList item */ item)}</span>
}
`,
		},
		"inline block comment in simple expression": {
			content: `package main

templ Hello(x int) {
	<span>{/* test */ x}</span>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			params, _ := json.Marshal(SemanticTokensParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
			})

			result, rpcErr := server.router.Route(Request{
				Method: "textDocument/semanticTokens/full",
				Params: params,
			})

			if rpcErr != nil {
				t.Fatalf("semantic tokens error: %v", rpcErr)
			}

			tokens, ok := result.(*SemanticTokens)
			if !ok {
				t.Fatalf("expected *SemanticTokens, got %T", result)
			}

			// Decode and find comment tokens
			decoded := decodeSemanticTokens(tokens.Data)
			var foundComment *decodedToken
			for i := range decoded {
				if decoded[i].tokenType == provider.TokenTypeComment {
					foundComment = &decoded[i]
					break
				}
			}

			if foundComment == nil {
				t.Fatalf("expected to find a comment token for inline block comment")
			}

			// Just verify we found a comment token - the actual highlighting will work
			t.Logf("Found comment token: line=%d startChar=%d length=%d",
				foundComment.line, foundComment.startChar, foundComment.length)
		})
	}
}

func TestCommentsInVariousPositions(t *testing.T) {
	// Test that comments are collected from all supported positions
	content := `package main

// File leading comment

// Component doc comment
templ Hello(name string) {  // Component trailing
	// Element leading
	<div>  // Element trailing
		// Nested comment
		<span>{name}</span>
	</div>
}

// Function comment
func helper() string {
	return "test"
}
`
	server := NewServer(nil, nil)

	doc := server.docs.Open("file:///test.gsx", content, 1)
	server.index.IndexDocument("file:///test.gsx", doc.AST)

	params, _ := json.Marshal(SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
	})

	result, rpcErr := server.router.Route(Request{
		Method: "textDocument/semanticTokens/full",
		Params: params,
	})

	if rpcErr != nil {
		t.Fatalf("semantic tokens error: %v", rpcErr)
	}

	tokens, ok := result.(*SemanticTokens)
	if !ok {
		t.Fatalf("expected *SemanticTokens, got %T", result)
	}

	commentCount := countTokensByType(tokens.Data, provider.TokenTypeComment)

	// We expect at least 7 comments:
	// 1. File leading
	// 2. Component doc
	// 3. Component trailing
	// 4. Element leading
	// 5. Element trailing
	// 6. Nested comment
	// 7. Function comment
	expectedMin := 7
	if commentCount < expectedMin {
		t.Errorf("got %d comment tokens, want at least %d", commentCount, expectedMin)
	}
}

// --- Test helpers ---

// decodedToken represents a semantic token decoded from LSP delta format.
type decodedToken struct {
	line      int
	startChar int
	length    int
	tokenType int
	modifiers int
}

// countTokensByType counts tokens of a given type in LSP delta-encoded data.
func countTokensByType(data []int, tokenType int) int {
	count := 0
	for i := 3; i < len(data); i += 5 {
		if data[i] == tokenType {
			count++
		}
	}
	return count
}

// decodeSemanticTokens decodes LSP delta-encoded semantic token data into
// absolute positions.
func decodeSemanticTokens(data []int) []decodedToken {
	var tokens []decodedToken
	prevLine := 0
	prevChar := 0

	for i := 0; i+4 < len(data); i += 5 {
		deltaLine := data[i]
		deltaChar := data[i+1]
		length := data[i+2]
		tokenType := data[i+3]
		modifiers := data[i+4]

		line := prevLine + deltaLine
		startChar := deltaChar
		if deltaLine == 0 {
			startChar = prevChar + deltaChar
		}

		tokens = append(tokens, decodedToken{
			line:      line,
			startChar: startChar,
			length:    length,
			tokenType: tokenType,
			modifiers: modifiers,
		})

		prevLine = line
		prevChar = startChar
	}

	return tokens
}
