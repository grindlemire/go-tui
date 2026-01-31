package provider

import (
	"testing"
)

// stubFnChecker implements FunctionNameChecker for testing.
type stubFnChecker struct {
	names map[string]bool
}

func (s *stubFnChecker) IsFunctionName(name string) bool {
	return s.names[name]
}

func newTestSemanticProvider() *semanticTokensProvider {
	return &semanticTokensProvider{
		fnChecker: &stubFnChecker{names: map[string]bool{
			"Sprintf": true,
			"len":     true,
		}},
	}
}

// decodeTokens decodes LSP delta-encoded semantic token data into absolute positions.
func decodeTokens(data []int) []SemanticToken {
	var tokens []SemanticToken
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

		tokens = append(tokens, SemanticToken{
			Line:      line,
			StartChar: startChar,
			Length:    length,
			TokenType: tokenType,
			Modifiers: modifiers,
		})

		prevLine = line
		prevChar = startChar
	}

	return tokens
}

// countByType counts tokens of a given type in decoded tokens.
func countByType(tokens []SemanticToken, tokenType int) int {
	count := 0
	for _, t := range tokens {
		if t.TokenType == tokenType {
			count++
		}
	}
	return count
}

// findFirstByType returns the first token of a given type.
func findFirstByType(tokens []SemanticToken, tokenType int) *SemanticToken {
	for i := range tokens {
		if tokens[i].TokenType == tokenType {
			return &tokens[i]
		}
	}
	return nil
}

// hasTokenAt checks if a token exists at the given position with the given type and length.
func hasTokenAt(tokens []SemanticToken, line, col, length, tokenType int) bool {
	for _, tok := range tokens {
		if tok.Line == line && tok.StartChar == col && tok.Length == length && tok.TokenType == tokenType {
			return true
		}
	}
	return false
}

func TestSemanticTokens_ComponentDecl(t *testing.T) {
	type tc struct {
		content       string
		wantKeyword   bool // should find "templ" keyword token
		wantClassName bool // should find component name token
		wantParams    int  // number of parameter tokens expected
	}

	tests := map[string]tc{
		"simple component": {
			content: `package main

templ Hello() {
	<span>Hello</span>
}
`,
			wantKeyword:   true,
			wantClassName: true,
			wantParams:    0,
		},
		"component with params": {
			content: `package main

templ Greeting(name string, count int) {
	<span>{name}</span>
}
`,
			wantKeyword:   true,
			wantClassName: true,
			wantParams:    2,
		},
	}

	sp := newTestSemanticProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := parseTestDoc(tt.content)
			result, err := sp.SemanticTokensFull(doc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokens := decodeTokens(result.Data)

			keywordCount := countByType(tokens, TokenTypeKeyword)
			if tt.wantKeyword && keywordCount == 0 {
				t.Error("expected at least one keyword token (templ)")
			}

			classCount := countByType(tokens, TokenTypeClass)
			if tt.wantClassName && classCount == 0 {
				t.Error("expected at least one class token (component name)")
			}

			paramCount := countByType(tokens, TokenTypeParameter)
			if paramCount < tt.wantParams {
				t.Errorf("got %d parameter tokens, want at least %d", paramCount, tt.wantParams)
			}
		})
	}

	// Position assertions for "simple component": templ keyword at line 2, col 0
	t.Run("simple component positions", func(t *testing.T) {
		doc := parseTestDoc(tests["simple component"].content)
		result, err := sp.SemanticTokensFull(doc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tokens := decodeTokens(result.Data)

		if !hasTokenAt(tokens, 2, 0, 5, TokenTypeKeyword) {
			t.Error("expected templ keyword token at 2:0 with length 5")
		}
		if !hasTokenAt(tokens, 2, 6, 5, TokenTypeClass) {
			t.Error("expected Hello class token at 2:6 with length 5")
		}
	})

	// Position assertions for "component with params": params at expected positions
	t.Run("component with params positions", func(t *testing.T) {
		doc := parseTestDoc(tests["component with params"].content)
		result, err := sp.SemanticTokensFull(doc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tokens := decodeTokens(result.Data)

		if !hasTokenAt(tokens, 2, 0, 5, TokenTypeKeyword) {
			t.Error("expected templ keyword token at 2:0 with length 5")
		}
		if !hasTokenAt(tokens, 2, 6, 8, TokenTypeClass) {
			t.Error("expected Greeting class token at 2:6 with length 8")
		}
	})
}

func TestSemanticTokens_FunctionDecl(t *testing.T) {
	type tc struct {
		content  string
		wantFunc bool
	}

	tests := map[string]tc{
		"helper function": {
			content: `package main

func helper(s string) string {
	return s
}
`,
			wantFunc: true,
		},
	}

	sp := newTestSemanticProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := parseTestDoc(tt.content)
			result, err := sp.SemanticTokensFull(doc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokens := decodeTokens(result.Data)

			funcCount := countByType(tokens, TokenTypeFunction)
			if tt.wantFunc && funcCount == 0 {
				t.Error("expected at least one function token")
			}
		})
	}
}

func TestSemanticTokens_Keywords(t *testing.T) {
	type tc struct {
		content     string
		wantKeyword int // minimum number of keyword tokens
	}

	tests := map[string]tc{
		"for loop keyword": {
			content: `package main

templ List(items []string) {
	@for _, item := range items {
		<span>{item}</span>
	}
}
`,
			wantKeyword: 2, // templ + @for
		},
		"if/else keywords": {
			content: `package main

templ Cond(show bool) {
	@if show {
		<span>Yes</span>
	} @else {
		<span>No</span>
	}
}
`,
			wantKeyword: 2, // templ + @if (at minimum)
		},
	}

	sp := newTestSemanticProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := parseTestDoc(tt.content)
			result, err := sp.SemanticTokensFull(doc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokens := decodeTokens(result.Data)

			keywordCount := countByType(tokens, TokenTypeKeyword)
			if keywordCount < tt.wantKeyword {
				t.Errorf("got %d keyword tokens, want at least %d", keywordCount, tt.wantKeyword)
			}
		})
	}

	// Position assertions for "for loop keyword"
	t.Run("for loop keyword positions", func(t *testing.T) {
		doc := parseTestDoc(tests["for loop keyword"].content)
		result, err := sp.SemanticTokensFull(doc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tokens := decodeTokens(result.Data)

		if !hasTokenAt(tokens, 2, 0, 5, TokenTypeKeyword) {
			t.Error("expected templ keyword at 2:0 with length 5")
		}
		if !hasTokenAt(tokens, 3, 1, 4, TokenTypeKeyword) {
			t.Error("expected @for keyword at 3:1 with length 4")
		}
	})

	// Position assertions for "if/else keywords" — needs docs accessor for @else
	t.Run("if/else keyword positions", func(t *testing.T) {
		doc := parseTestDoc(tests["if/else keywords"].content)
		spWithDocs := &semanticTokensProvider{
			fnChecker: &stubFnChecker{names: map[string]bool{}},
			docs:      &stubDocAccessor{docs: []*Document{doc}},
		}
		result, err := spWithDocs.SemanticTokensFull(doc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tokens := decodeTokens(result.Data)

		if !hasTokenAt(tokens, 2, 0, 5, TokenTypeKeyword) {
			t.Error("expected templ keyword at 2:0 with length 5")
		}
		if !hasTokenAt(tokens, 3, 1, 3, TokenTypeKeyword) {
			t.Error("expected @if keyword at 3:1 with length 3")
		}
		if !hasTokenAt(tokens, 5, 3, 5, TokenTypeKeyword) {
			t.Error("expected @else keyword at 5:3 with length 5")
		}
	})
}

func TestSemanticTokens_ElementAttributes(t *testing.T) {
	type tc struct {
		content  string
		wantAttr int // minimum number of attribute tokens (emitted as TokenTypeFunction)
	}

	tests := map[string]tc{
		"element with attributes": {
			content: `package main

templ Hello() {
	<div class="border-single" id="main">
		<span>Hello</span>
	</div>
}
`,
			wantAttr: 2, // class, id
		},
	}

	sp := newTestSemanticProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := parseTestDoc(tt.content)
			result, err := sp.SemanticTokensFull(doc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokens := decodeTokens(result.Data)

			// The provider emits attribute names as TokenTypeFunction
			funcCount := countByType(tokens, TokenTypeFunction)
			if funcCount < tt.wantAttr {
				t.Errorf("got %d function/attribute tokens, want at least %d", funcCount, tt.wantAttr)
			}
		})
	}
}

func TestSemanticTokens_Comments(t *testing.T) {
	type tc struct {
		content     string
		wantComment int
	}

	tests := map[string]tc{
		"line comment": {
			content: `package main

// A comment
templ Hello() {
	<span>Hello</span>
}
`,
			wantComment: 1,
		},
		"block comment": {
			content: `package main

/* Block comment */
templ Hello() {
	<span>Hello</span>
}
`,
			wantComment: 1,
		},
		"comment inside component": {
			content: `package main

templ Hello() {
	// inner comment
	<span>Hello</span>
}
`,
			wantComment: 1,
		},
	}

	sp := newTestSemanticProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := parseTestDoc(tt.content)
			result, err := sp.SemanticTokensFull(doc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokens := decodeTokens(result.Data)

			commentCount := countByType(tokens, TokenTypeComment)
			if commentCount < tt.wantComment {
				t.Errorf("got %d comment tokens, want at least %d", commentCount, tt.wantComment)
			}
		})
	}
}
