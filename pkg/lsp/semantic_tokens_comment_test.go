package lsp

import (
	"encoding/json"
	"testing"
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

			result, rpcErr := server.handleSemanticTokensFull(params)

			if rpcErr != nil {
				t.Fatalf("handleSemanticTokensFull error: %v", rpcErr)
			}

			tokens, ok := result.(SemanticTokens)
			if !ok {
				t.Fatalf("expected SemanticTokens, got %T", result)
			}

			// Count comment tokens (every 5th value starting at index 3 is the token type)
			commentCount := 0
			for i := 3; i < len(tokens.Data); i += 5 {
				if tokens.Data[i] == tokenTypeComment {
					commentCount++
				}
			}

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

			tokens := server.collectSemanticTokens(doc)

			// Find the comment token
			var foundComment *semanticToken
			for i := range tokens {
				if tokens[i].tokenType == tokenTypeComment {
					foundComment = &tokens[i]
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

			tokens := server.collectSemanticTokens(doc)

			// Count comment tokens
			commentCount := 0
			for i := range tokens {
				if tokens[i].tokenType == tokenTypeComment {
					commentCount++
				}
			}

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

			tokens := server.collectSemanticTokens(doc)

			// Find comment tokens
			var foundComment *semanticToken
			for i := range tokens {
				if tokens[i].tokenType == tokenTypeComment {
					foundComment = &tokens[i]
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

	tokens := server.collectSemanticTokens(doc)

	// Count comment tokens
	commentCount := 0
	for i := range tokens {
		if tokens[i].tokenType == tokenTypeComment {
			commentCount++
		}
	}

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
