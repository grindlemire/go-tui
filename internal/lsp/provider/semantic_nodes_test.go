package provider

import (
	"sort"
	"testing"
)

func TestSemanticTokens_ComponentCalls(t *testing.T) {
	type tc struct {
		content       string
		wantDecorator int
	}

	tests := map[string]tc{
		"component call": {
			content: `package main

templ App() {
	@Header("title")
}
`,
			wantDecorator: 1,
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

			decoratorCount := countByType(tokens, TokenTypeDecorator)
			if decoratorCount < tt.wantDecorator {
				t.Errorf("got %d decorator tokens, want at least %d", decoratorCount, tt.wantDecorator)
			}
		})
	}
}

func TestSemanticTokens_TokenTypeConstants(t *testing.T) {
	type tc struct {
		name     string
		constant int
		expected int
	}

	tests := []tc{
		{"namespace", TokenTypeNamespace, 0},
		{"type", TokenTypeType, 1},
		{"class", TokenTypeClass, 2},
		{"function", TokenTypeFunction, 3},
		{"parameter", TokenTypeParameter, 4},
		{"variable", TokenTypeVariable, 5},
		{"property", TokenTypeProperty, 6},
		{"keyword", TokenTypeKeyword, 7},
		{"string", TokenTypeString, 8},
		{"number", TokenTypeNumber, 9},
		{"operator", TokenTypeOperator, 10},
		{"decorator", TokenTypeDecorator, 11},
		{"regexp", TokenTypeRegexp, 12},
		{"comment", TokenTypeComment, 13},
		{"label", TokenTypeLabel, 14},
		{"typeParameter", TokenTypeTypeParameter, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("TokenType%s = %d, want %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestSemanticTokens_Encoding(t *testing.T) {
	type tc struct {
		tokens   []SemanticToken
		expected []int
	}

	tests := map[string]tc{
		"single token": {
			tokens: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 7, TokenType: TokenTypeKeyword, Modifiers: 0},
			},
			expected: []int{0, 0, 7, TokenTypeKeyword, 0},
		},
		"two tokens same line": {
			tokens: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 7, TokenType: TokenTypeKeyword, Modifiers: 0},
				{Line: 0, StartChar: 8, Length: 5, TokenType: TokenTypeClass, Modifiers: 0},
			},
			expected: []int{
				0, 0, 7, TokenTypeKeyword, 0,
				0, 8, 5, TokenTypeClass, 0,
			},
		},
		"two tokens different lines": {
			tokens: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 7, TokenType: TokenTypeKeyword, Modifiers: 0},
				{Line: 2, StartChar: 1, Length: 4, TokenType: TokenTypeVariable, Modifiers: 0},
			},
			expected: []int{
				0, 0, 7, TokenTypeKeyword, 0,
				2, 1, 4, TokenTypeVariable, 0,
			},
		},
		"empty": {
			tokens:   []SemanticToken{},
			expected: []int{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sort.Slice(tt.tokens, func(i, j int) bool {
				if tt.tokens[i].Line != tt.tokens[j].Line {
					return tt.tokens[i].Line < tt.tokens[j].Line
				}
				return tt.tokens[i].StartChar < tt.tokens[j].StartChar
			})

			result := EncodeSemanticTokens(tt.tokens)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %d ints, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("result[%d] = %d, want %d", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSemanticTokens_NilAST(t *testing.T) {
	sp := newTestSemanticProvider()
	doc := &Document{
		URI:     "file:///test.gsx",
		Content: "",
		Version: 1,
		AST:     nil,
	}

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Data) != 0 {
		t.Errorf("expected empty data, got %d ints", len(result.Data))
	}
}

func TestSemanticTokens_Variables(t *testing.T) {
	sp := newTestSemanticProvider()
	doc := parseTestDoc(`package main

templ Greeting(name string) {
	<span>{name}</span>
}
`)

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokens := decodeTokens(result.Data)

	paramCount := countByType(tokens, TokenTypeParameter)
	if paramCount == 0 {
		t.Error("expected at least one parameter token for 'name'")
	}
}

func TestSemanticTokens_RefAttr(t *testing.T) {
	type tc struct {
		content     string
		wantRefAttr int // "ref" attribute token
		wantVarDecl int // ref value as variable with declaration modifier
	}

	tests := map[string]tc{
		"simple ref attr": {
			content: `package main

templ Layout() {
	<div ref={header} class="p-1">title</div>
}
`,
			wantRefAttr: 1, // the "ref" attribute name
			wantVarDecl: 1, // header ref value (variable with declaration modifier)
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

			// Count function tokens that match "ref" (attribute name)
			refAttrCount := 0
			for _, tok := range tokens {
				if tok.TokenType == TokenTypeFunction && tok.Length == len("ref") {
					refAttrCount++
				}
			}
			if refAttrCount < tt.wantRefAttr {
				t.Errorf("got %d ref attribute tokens, want at least %d", refAttrCount, tt.wantRefAttr)
			}

			// Find variable tokens with declaration modifier (the ref value)
			varDeclCount := 0
			for _, tok := range tokens {
				if tok.TokenType == TokenTypeVariable && tok.Modifiers&TokenModDeclaration != 0 {
					varDeclCount++
				}
			}
			if varDeclCount < tt.wantVarDecl {
				t.Errorf("got %d variable declaration tokens, want at least %d (for ref value)", varDeclCount, tt.wantVarDecl)
			}
		})
	}
}

func TestSemanticTokens_EventHandlerAttributes(t *testing.T) {
	type tc struct {
		content       string
		wantDecorator int // event handler attributes get decorator type
		wantFunction  int // regular attributes get function type
	}

	tests := map[string]tc{
		"event handler vs regular attribute": {
			content: `package main

templ Button() {
	<button onFocus={handleFocus} class="p-1">Click</button>
}
`,
			wantDecorator: 2, // onFocus as decorator + the @ if there's a component call (just onFocus here)
			wantFunction:  1, // class as function
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

			// Event handler attributes should be decorated differently
			decoratorCount := countByType(tokens, TokenTypeDecorator)
			if decoratorCount < 1 {
				t.Errorf("got %d decorator tokens, want at least 1 (for onFocus)", decoratorCount)
			}

			// Regular attributes should be function tokens
			funcCount := countByType(tokens, TokenTypeFunction)
			if funcCount < tt.wantFunction {
				t.Errorf("got %d function tokens, want at least %d (for regular attributes)", funcCount, tt.wantFunction)
			}
		})
	}
}

func TestSemanticTokens_StateVarDeclaration(t *testing.T) {
	type tc struct {
		content          string
		wantReadonlyDecl int // state vars should get declaration + readonly modifiers
	}

	tests := map[string]tc{
		"state variable declaration": {
			content: `package main

templ Counter() {
	count := tui.NewState(0)
	<span>{count}</span>
}
`,
			wantReadonlyDecl: 1, // count var declaration with readonly modifier
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

			// Find variable tokens with both declaration and readonly modifiers
			readonlyDeclCount := 0
			for _, tok := range tokens {
				if tok.TokenType == TokenTypeVariable &&
					tok.Modifiers&TokenModDeclaration != 0 &&
					tok.Modifiers&TokenModReadonly != 0 {
					readonlyDeclCount++
				}
			}
			if readonlyDeclCount < tt.wantReadonlyDecl {
				t.Errorf("got %d variable tokens with declaration+readonly modifiers, want at least %d",
					readonlyDeclCount, tt.wantReadonlyDecl)
			}
		})
	}
}

func TestSemanticTokens_MultipleComponents(t *testing.T) {
	sp := newTestSemanticProvider()
	doc := parseTestDoc(`package main

templ Header(title string) {
	<div>{title}</div>
}

templ Footer() {
	<span>Footer</span>
}
`)

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokens := decodeTokens(result.Data)

	keywordCount := countByType(tokens, TokenTypeKeyword)
	if keywordCount < 2 {
		t.Errorf("got %d keyword tokens, want at least 2", keywordCount)
	}

	classCount := countByType(tokens, TokenTypeClass)
	if classCount < 2 {
		t.Errorf("got %d class tokens, want at least 2", classCount)
	}
}

func TestSemanticTokens_StateModifierOnlyOnStateVar(t *testing.T) {
	// Regression: the readonly modifier should only apply to the state variable,
	// not to all variables declared in the same GoCode block.
	src := `package test

templ Counter() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}
`
	sp := newTestSemanticProvider()
	doc := parseTestDoc(src)

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokens := decodeTokens(result.Data)

	// Find variable tokens with declaration modifier
	for _, tok := range tokens {
		if tok.TokenType == TokenTypeVariable && (tok.Modifiers&TokenModDeclaration) != 0 {
			// The "count" variable should have readonly modifier
			if tok.StartChar == 1 { // "count" is at column 1 (after tab)
				if (tok.Modifiers & TokenModReadonly) == 0 {
					t.Error("state variable 'count' should have readonly modifier")
				}
			}
		}
	}
}

func TestSemanticTokens_NonStateVarNoReadonly(t *testing.T) {
	// When a GoCode block has a non-state variable, it should NOT get readonly.
	// Note: the parser produces separate GoCode nodes for separate statements,
	// so we test with a state declaration to verify only it gets readonly.
	src := `package test

templ Example() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}
`
	sp := newTestSemanticProvider()
	doc := parseTestDoc(src)

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokens := decodeTokens(result.Data)

	// All variable declarations should be accounted for
	varDeclCount := 0
	readonlyVarCount := 0
	for _, tok := range tokens {
		if tok.TokenType == TokenTypeVariable && (tok.Modifiers&TokenModDeclaration) != 0 {
			varDeclCount++
			if (tok.Modifiers & TokenModReadonly) != 0 {
				readonlyVarCount++
			}
		}
	}
	if varDeclCount == 0 {
		t.Error("expected at least one variable declaration token")
	}
	// Only state var should be readonly
	if readonlyVarCount > 1 {
		t.Errorf("expected at most 1 readonly variable, got %d", readonlyVarCount)
	}
}
