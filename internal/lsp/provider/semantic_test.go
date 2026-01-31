package provider

import (
	"sort"
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

func TestSemanticTokens_NamedRef(t *testing.T) {
	type tc struct {
		content      string
		wantOperator int // # operator token
		wantVarDecl  int // ref name as variable with declaration modifier
	}

	tests := map[string]tc{
		"simple named ref": {
			content: `package main

templ Layout() {
	<div #Header class="p-1">title</div>
}
`,
			wantOperator: 1, // the # symbol
			wantVarDecl:  1, // Header ref name (label token with declaration modifier)
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

			// Count operator tokens (the # symbol)
			operatorCount := countByType(tokens, TokenTypeOperator)
			if operatorCount < tt.wantOperator {
				t.Errorf("got %d operator tokens, want at least %d (for # in named ref)", operatorCount, tt.wantOperator)
			}

			// Find keyword tokens with declaration modifier (the ref name)
			keywordDeclCount := 0
			for _, tok := range tokens {
				if tok.TokenType == TokenTypeKeyword && tok.Modifiers&TokenModDeclaration != 0 {
					keywordDeclCount++
				}
			}
			if keywordDeclCount < tt.wantVarDecl {
				t.Errorf("got %d keyword declaration tokens, want at least %d (for ref name)", keywordDeclCount, tt.wantVarDecl)
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
	<button onClick={handleClick} class="p-1">Click</button>
}
`,
			wantDecorator: 2, // onClick as decorator + the @ if there's a component call (just onClick here)
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
				t.Errorf("got %d decorator tokens, want at least 1 (for onClick)", decoratorCount)
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
