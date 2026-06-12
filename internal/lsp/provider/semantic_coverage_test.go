package provider

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func TestNewSemanticTokensProvider_Constructor(t *testing.T) {
	sp := NewSemanticTokensProvider(&stubFnChecker{names: map[string]bool{}}, &stubDocAccessor{})
	if sp == nil {
		t.Fatal("expected non-nil provider")
	}

	doc := &Document{URI: "file:///test.gsx", Content: "", Version: 1} // nil AST
	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected empty data for nil AST, got %v", result.Data)
	}
}

func TestSemanticTokens_PackageAndImports(t *testing.T) {
	content := `package main

import (
	"fmt"
	x "strings"
)

templ I() {
	<div>{fmt.Sprint(x.ToUpper("a"))}</div>
}
`
	doc := parseTestDoc(content)
	sp := newTestSemanticProvider()

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tokens := decodeTokens(result.Data)

	if !hasTokenAt(tokens, 0, 0, len("package"), TokenTypeKeyword) {
		t.Errorf("expected package keyword at 0:0, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 0, 8, len("main"), TokenTypeNamespace) {
		t.Errorf("expected package name at 0:8, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 2, 0, len("import"), TokenTypeKeyword) {
		t.Errorf("expected import keyword at 2:0, tokens: %+v", tokens)
	}
	// "fmt" with quotes: 5 characters at line 3 col 1.
	if !hasTokenAt(tokens, 3, 1, 5, TokenTypeString) {
		t.Errorf("expected quoted import path at 3:1, tokens: %+v", tokens)
	}
	// Alias x at line 4 col 1, then "strings" at col 3.
	if !hasTokenAt(tokens, 4, 1, 1, TokenTypeNamespace) {
		t.Errorf("expected import alias at 4:1, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 4, 3, len(`"strings"`), TokenTypeString) {
		t.Errorf("expected aliased import path at 4:3, tokens: %+v", tokens)
	}
}

func TestSemanticTokens_MethodTempl(t *testing.T) {
	content := `package main

templ (c *chat) Render() {
	<div>{c.title}</div>
}
`
	doc := parseTestDoc(content)
	sp := newTestSemanticProvider()

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tokens := decodeTokens(result.Data)

	if !hasTokenAt(tokens, 2, 0, len("templ"), TokenTypeKeyword) {
		t.Errorf("expected templ keyword at 2:0, tokens: %+v", tokens)
	}
	// Receiver name c at col 7.
	if !hasTokenAt(tokens, 2, 7, 1, TokenTypeParameter) {
		t.Errorf("expected receiver name token at 2:7, tokens: %+v", tokens)
	}
	// Receiver type *chat: pointer star at col 9, chat at col 10.
	if !hasTokenAt(tokens, 2, 9, 1, TokenTypeOperator) {
		t.Errorf("expected pointer operator at 2:9, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 2, 10, len("chat"), TokenTypeType) {
		t.Errorf("expected receiver type at 2:10, tokens: %+v", tokens)
	}
	// Component name Render at col 16.
	if !hasTokenAt(tokens, 2, 16, len("Render"), TokenTypeClass) {
		t.Errorf("expected component name at 2:16, tokens: %+v", tokens)
	}
}

func TestSemanticTokens_GenericFunction(t *testing.T) {
	content := `package main

func pick[T bool | string](a T, b T) T {
	return a
}

templ G() {
	<div>x</div>
}
`
	doc := parseTestDoc(content)
	sp := newTestSemanticProvider()

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tokens := decodeTokens(result.Data)

	if !hasTokenAt(tokens, 2, 0, len("func"), TokenTypeKeyword) {
		t.Errorf("expected func keyword at 2:0, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 2, 5, len("pick"), TokenTypeFunction) {
		t.Errorf("expected function name at 2:5, tokens: %+v", tokens)
	}
	// Generic type parameter list [T bool | string].
	if !hasTokenAt(tokens, 2, 10, 1, TokenTypeParameter) {
		t.Errorf("expected type param T at 2:10, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 2, 12, len("bool"), TokenTypeType) {
		t.Errorf("expected constraint bool at 2:12, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 2, 17, 1, TokenTypeOperator) {
		t.Errorf("expected union operator at 2:17, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 2, 19, len("string"), TokenTypeType) {
		t.Errorf("expected constraint string at 2:19, tokens: %+v", tokens)
	}
	// Parameters a and b with their T types.
	if !hasTokenAt(tokens, 2, 27, 1, TokenTypeParameter) {
		t.Errorf("expected param a at 2:27, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 2, 29, 1, TokenTypeType) {
		t.Errorf("expected param type T at 2:29, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 2, 32, 1, TokenTypeParameter) {
		t.Errorf("expected param b at 2:32, tokens: %+v", tokens)
	}
	// Return type T at col 37.
	if !hasTokenAt(tokens, 2, 37, 1, TokenTypeType) {
		t.Errorf("expected return type at 2:37, tokens: %+v", tokens)
	}
	// Body: return keyword and param usage.
	if !hasTokenAt(tokens, 3, 1, len("return"), TokenTypeKeyword) {
		t.Errorf("expected return keyword at 3:1, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 3, 8, 1, TokenTypeParameter) {
		t.Errorf("expected param usage at 3:8, tokens: %+v", tokens)
	}
}

func TestSemanticTokens_FullComponentFixture(t *testing.T) {
	content := `package main

type pair struct {
	a int
	b int
}

templ App(items []string) {
	count := tui.NewState(0)
	<div class="p-1" ref={panel}>
		{children...}
		for i, item := range items {
			<span>{item}</span>
		}
	</div>
	var footer = <span>end</span>
	{footer}
	@Card("x")
}
`
	doc := parseTestDoc(content)
	sp := newTestSemanticProvider()

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tokens := decodeTokens(result.Data)

	// GoDecl: type keyword, struct keyword, builtin field types.
	if !hasTokenAt(tokens, 2, 0, len("type"), TokenTypeKeyword) {
		t.Errorf("expected type keyword at 2:0")
	}
	if !hasTokenAt(tokens, 2, 10, len("struct"), TokenTypeKeyword) {
		t.Errorf("expected struct keyword at 2:10")
	}
	if !hasTokenAt(tokens, 3, 3, len("int"), TokenTypeType) {
		t.Errorf("expected int type at 3:3")
	}

	// State declaration: variable with declaration|readonly modifiers.
	found := false
	for _, tok := range tokens {
		if tok.Line == 8 && tok.StartChar == 1 && tok.Length == len("count") &&
			tok.TokenType == TokenTypeVariable &&
			tok.Modifiers == TokenModDeclaration|TokenModReadonly {
			found = true
		}
	}
	if !found {
		t.Errorf("expected state var token at 8:1 with declaration|readonly modifiers, tokens: %+v", tokens)
	}
	// := operator and NewState function call.
	if !hasTokenAt(tokens, 8, 7, 2, TokenTypeOperator) {
		t.Errorf("expected := operator at 8:7")
	}
	if !hasTokenAt(tokens, 8, 14, len("NewState"), TokenTypeFunction) {
		t.Errorf("expected NewState function token at 8:14")
	}
	if !hasTokenAt(tokens, 8, 23, 1, TokenTypeNumber) {
		t.Errorf("expected number token at 8:23")
	}

	// Element tag and attribute.
	if !hasTokenAt(tokens, 9, 2, len("div"), TokenTypeKeyword) {
		t.Errorf("expected div tag token at 9:2")
	}
	if !hasTokenAt(tokens, 9, 6, len("class"), TokenTypeFunction) {
		t.Errorf("expected class attribute token at 9:6")
	}

	// ref={panel}: ref as function, panel as variable declaration.
	if !hasTokenAt(tokens, 9, 18, len("ref"), TokenTypeFunction) {
		t.Errorf("expected ref token at 9:18")
	}
	if !hasTokenAt(tokens, 9, 23, len("panel"), TokenTypeVariable) {
		t.Errorf("expected panel ref value token at 9:23")
	}

	// Children slot: children keyword and ... operator.
	if !hasTokenAt(tokens, 10, 3, len("children"), TokenTypeKeyword) {
		t.Errorf("expected children keyword at 10:3")
	}
	if !hasTokenAt(tokens, 10, 11, 3, TokenTypeOperator) {
		t.Errorf("expected ... operator at 10:11")
	}

	// For loop: keyword, index, value, iterable (a component param).
	if !hasTokenAt(tokens, 11, 2, len("for"), TokenTypeKeyword) {
		t.Errorf("expected for keyword at 11:2")
	}
	if !hasTokenAt(tokens, 11, 6, 1, TokenTypeVariable) {
		t.Errorf("expected loop index token at 11:6")
	}
	if !hasTokenAt(tokens, 11, 9, len("item"), TokenTypeVariable) {
		t.Errorf("expected loop value token at 11:9")
	}
	if !hasTokenAt(tokens, 11, 23, len("items"), TokenTypeParameter) {
		t.Errorf("expected iterable param token at 11:23")
	}
	// Loop variable usage inside body.
	if !hasTokenAt(tokens, 12, 10, len("item"), TokenTypeVariable) {
		t.Errorf("expected loop var usage at 12:10")
	}

	// var-form let binding: var keyword and name.
	if !hasTokenAt(tokens, 15, 1, len("var"), TokenTypeKeyword) {
		t.Errorf("expected var keyword at 15:1")
	}
	if !hasTokenAt(tokens, 15, 5, len("footer"), TokenTypeVariable) {
		t.Errorf("expected footer binding token at 15:5")
	}

	// Component call: decorator, class name, string arg.
	if !hasTokenAt(tokens, 17, 1, 1, TokenTypeDecorator) {
		t.Errorf("expected @ decorator at 17:1")
	}
	if !hasTokenAt(tokens, 17, 2, len("Card"), TokenTypeClass) {
		t.Errorf("expected Card class token at 17:2")
	}
	if !hasTokenAt(tokens, 17, 7, 3, TokenTypeString) {
		t.Errorf("expected string arg token at 17:7")
	}
}

func TestCollectTokensInGoCode_Literals(t *testing.T) {
	type tc struct {
		code   string
		params map[string]bool
		locals map[string]bool
		want   []SemanticToken
	}

	tests := map[string]tc{
		"number bases": {
			code: "x := 0x1F + 0b101 + 0o17 + 1.5e+3",
			want: []SemanticToken{
				{Line: 0, StartChar: 2, Length: 2, TokenType: TokenTypeOperator},
				{Line: 0, StartChar: 5, Length: 4, TokenType: TokenTypeNumber},
				{Line: 0, StartChar: 12, Length: 5, TokenType: TokenTypeNumber},
				{Line: 0, StartChar: 20, Length: 4, TokenType: TokenTypeNumber},
				{Line: 0, StartChar: 27, Length: 6, TokenType: TokenTypeNumber},
			},
		},
		"block comment": {
			code: "/* hi */",
			want: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 8, TokenType: TokenTypeComment},
			},
		},
		"line comment": {
			code: "// note",
			want: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 7, TokenType: TokenTypeComment},
			},
		},
		"backtick string with format spec": {
			code: "`raw %d`",
			want: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 5, TokenType: TokenTypeString},
				{Line: 0, StartChar: 5, Length: 2, TokenType: TokenTypeRegexp},
				{Line: 0, StartChar: 7, Length: 1, TokenType: TokenTypeString},
			},
		},
		"rune literal with escape": {
			code: `'\n'`,
			want: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 4, TokenType: TokenTypeString},
			},
		},
		"booleans and nil": {
			code: "true false nil",
			want: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 4, TokenType: TokenTypeNumber},
				{Line: 0, StartChar: 5, Length: 5, TokenType: TokenTypeNumber},
				{Line: 0, StartChar: 11, Length: 3, TokenType: TokenTypeNumber},
			},
		},
		"param and local": {
			code:   "p + q",
			params: map[string]bool{"p": true},
			locals: map[string]bool{"q": true},
			want: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 1, TokenType: TokenTypeParameter},
				{Line: 0, StartChar: 4, Length: 1, TokenType: TokenTypeVariable},
			},
		},
		"package call": {
			code: "fmt.Println(v)",
			want: []SemanticToken{
				{Line: 0, StartChar: 4, Length: 7, TokenType: TokenTypeFunction},
			},
		},
		"known function via checker": {
			code: "Sprintf",
			want: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 7, TokenType: TokenTypeFunction},
			},
		},
		"generic type argument": {
			code: "State[bool]",
			want: []SemanticToken{
				{Line: 0, StartChar: 6, Length: 4, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
			},
		},
		"builtin outside brackets": {
			code: "var s string",
			want: []SemanticToken{
				{Line: 0, StartChar: 0, Length: 3, TokenType: TokenTypeKeyword},
				{Line: 0, StartChar: 6, Length: 6, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
			},
		},
	}

	sp := newTestSemanticProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := tt.params
			if params == nil {
				params = map[string]bool{}
			}
			locals := tt.locals
			if locals == nil {
				locals = map[string]bool{}
			}

			var tokens []SemanticToken
			sp.collectTokensInGoCode(tt.code, tuigen.Position{Line: 1, Column: 1}, 0, params, locals, &tokens)

			for _, want := range tt.want {
				matched := false
				for _, got := range tokens {
					if got == want {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("missing token %+v in %+v", want, tokens)
				}
			}
		})
	}
}

func TestCollectTokensInGoCode_MultiLine(t *testing.T) {
	sp := newTestSemanticProvider()

	var tokens []SemanticToken
	sp.collectTokensInGoCode("a := 1\nb := 2", tuigen.Position{Line: 5, Column: 3}, 2, map[string]bool{}, map[string]bool{}, &tokens)

	// First line keeps the position column and offset: := at 3-1+2+2 = 6.
	if !hasTokenAt(tokens, 4, 6, 2, TokenTypeOperator) {
		t.Errorf("expected := at 4:6, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 4, 9, 1, TokenTypeNumber) {
		t.Errorf("expected number at 4:9, tokens: %+v", tokens)
	}
	// Continuation line resets column and offset: := at 2.
	if !hasTokenAt(tokens, 5, 2, 2, TokenTypeOperator) {
		t.Errorf("expected := at 5:2, tokens: %+v", tokens)
	}
	if !hasTokenAt(tokens, 5, 5, 1, TokenTypeNumber) {
		t.Errorf("expected number at 5:5, tokens: %+v", tokens)
	}
}

func TestEmitStringWithFormatSpecifiers_Direct(t *testing.T) {
	type tc struct {
		str  string
		want []SemanticToken
	}

	tests := map[string]tc{
		"no specifiers": {
			str: `"plain"`,
			want: []SemanticToken{
				{Line: 0, StartChar: 10, Length: 7, TokenType: TokenTypeString},
			},
		},
		"specifier in middle": {
			str: `"a %d b"`,
			want: []SemanticToken{
				{Line: 0, StartChar: 10, Length: 3, TokenType: TokenTypeString},
				{Line: 0, StartChar: 13, Length: 2, TokenType: TokenTypeRegexp},
				{Line: 0, StartChar: 15, Length: 3, TokenType: TokenTypeString},
			},
		},
		"specifier at end of content": {
			str: `"items: %v"`,
			want: []SemanticToken{
				{Line: 0, StartChar: 10, Length: 8, TokenType: TokenTypeString},
				{Line: 0, StartChar: 18, Length: 2, TokenType: TokenTypeRegexp},
				{Line: 0, StartChar: 20, Length: 1, TokenType: TokenTypeString},
			},
		},
	}

	sp := newTestSemanticProvider()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var tokens []SemanticToken
			sp.emitStringWithFormatSpecifiers(tt.str, 0, 10, &tokens)

			if len(tokens) != len(tt.want) {
				t.Fatalf("got %d tokens, want %d: %+v", len(tokens), len(tt.want), tokens)
			}
			for i, want := range tt.want {
				if tokens[i] != want {
					t.Errorf("token %d = %+v, want %+v", i, tokens[i], want)
				}
			}
		})
	}
}

func TestEmitGoTypeTokens_Direct(t *testing.T) {
	type tc struct {
		typeStr string
		want    []SemanticToken
	}

	tests := map[string]tc{
		"pointer to generic state": {
			typeStr: "*tui.State[int]",
			want: []SemanticToken{
				{Line: 1, StartChar: 4, Length: 1, TokenType: TokenTypeOperator},
				{Line: 1, StartChar: 9, Length: 5, TokenType: TokenTypeType},
				{Line: 1, StartChar: 14, Length: 1, TokenType: TokenTypeOperator},
				{Line: 1, StartChar: 15, Length: 3, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
				{Line: 1, StartChar: 18, Length: 1, TokenType: TokenTypeOperator},
			},
		},
		"function type": {
			typeStr: "func(string) error",
			want: []SemanticToken{
				{Line: 1, StartChar: 4, Length: 4, TokenType: TokenTypeKeyword},
				{Line: 1, StartChar: 9, Length: 6, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
				{Line: 1, StartChar: 17, Length: 5, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
			},
		},
		"map type": {
			typeStr: "map[string]int",
			want: []SemanticToken{
				{Line: 1, StartChar: 4, Length: 3, TokenType: TokenTypeKeyword},
				{Line: 1, StartChar: 7, Length: 1, TokenType: TokenTypeOperator},
				{Line: 1, StartChar: 8, Length: 6, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
				{Line: 1, StartChar: 14, Length: 1, TokenType: TokenTypeOperator},
				{Line: 1, StartChar: 15, Length: 3, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
			},
		},
		"receive channel": {
			typeStr: "<-chan int",
			want: []SemanticToken{
				{Line: 1, StartChar: 4, Length: 2, TokenType: TokenTypeOperator},
				{Line: 1, StartChar: 6, Length: 4, TokenType: TokenTypeKeyword},
				{Line: 1, StartChar: 11, Length: 3, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
			},
		},
		"tuple return": {
			typeStr: "(int, bool)",
			want: []SemanticToken{
				{Line: 1, StartChar: 5, Length: 3, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
				{Line: 1, StartChar: 10, Length: 4, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var tokens []SemanticToken
			emitGoTypeTokens(tt.typeStr, 1, 4, &tokens)

			if len(tokens) != len(tt.want) {
				t.Fatalf("got %d tokens, want %d: %+v", len(tokens), len(tt.want), tokens)
			}
			for i, want := range tt.want {
				if tokens[i] != want {
					t.Errorf("token %d = %+v, want %+v", i, tokens[i], want)
				}
			}
		})
	}
}

func TestEmitGenericTypeParamTokens_Direct(t *testing.T) {
	type tc struct {
		typeParams string
		want       []SemanticToken
	}

	tests := map[string]tc{
		"two groups": {
			typeParams: "[K comparable, V any]",
			want: []SemanticToken{
				{Line: 0, StartChar: 1, Length: 1, TokenType: TokenTypeParameter, Modifiers: TokenModDeclaration},
				{Line: 0, StartChar: 3, Length: 10, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
				{Line: 0, StartChar: 15, Length: 1, TokenType: TokenTypeParameter, Modifiers: TokenModDeclaration},
				{Line: 0, StartChar: 17, Length: 3, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
			},
		},
		"approximation constraint": {
			typeParams: "[T ~int]",
			want: []SemanticToken{
				{Line: 0, StartChar: 1, Length: 1, TokenType: TokenTypeParameter, Modifiers: TokenModDeclaration},
				{Line: 0, StartChar: 3, Length: 1, TokenType: TokenTypeOperator},
				{Line: 0, StartChar: 4, Length: 3, TokenType: TokenTypeType, Modifiers: TokenModDefaultLibrary},
			},
		},
		"pointer and package constraint": {
			typeParams: "[P *pkg.Type]",
			want: []SemanticToken{
				{Line: 0, StartChar: 1, Length: 1, TokenType: TokenTypeParameter, Modifiers: TokenModDeclaration},
				{Line: 0, StartChar: 3, Length: 1, TokenType: TokenTypeOperator},
				{Line: 0, StartChar: 8, Length: 4, TokenType: TokenTypeType},
			},
		},
		"interface constraint keyword": {
			typeParams: "[T interface{ Foo() }]",
			want: []SemanticToken{
				{Line: 0, StartChar: 1, Length: 1, TokenType: TokenTypeParameter, Modifiers: TokenModDeclaration},
				{Line: 0, StartChar: 3, Length: 9, TokenType: TokenTypeKeyword},
			},
		},
		"malformed input": {
			typeParams: "x",
			want:       nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var tokens []SemanticToken
			emitGenericTypeParamTokens(tt.typeParams, 0, 0, &tokens)

			for _, want := range tt.want {
				matched := false
				for _, got := range tokens {
					if got == want {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("missing token %+v in %+v", want, tokens)
				}
			}
			if tt.want == nil && len(tokens) != 0 {
				t.Errorf("expected no tokens, got %+v", tokens)
			}
		})
	}
}

func TestParseFuncSignatureForTokens(t *testing.T) {
	type tc struct {
		code           string
		wantName       string
		wantReceiver   string
		wantTypeParams string
		wantParams     []funcParam
		wantReturns    string
	}

	tests := map[string]tc{
		"plain function": {
			code:        "func helper(s string) string {\n\treturn s\n}",
			wantName:    "helper",
			wantParams:  []funcParam{{Name: "s", Type: "string"}},
			wantReturns: "string",
		},
		"method with receiver": {
			code:         "func (c *chat) update(h int) int {\n\treturn h\n}",
			wantName:     "update",
			wantReceiver: "c *chat",
			wantParams:   []funcParam{{Name: "h", Type: "int"}},
			wantReturns:  "int",
		},
		"generic function": {
			code:           "func pick[T any](a T) T {\n\treturn a\n}",
			wantName:       "pick",
			wantTypeParams: "[T any]",
			wantParams:     []funcParam{{Name: "a", Type: "T"}},
			wantReturns:    "T",
		},
		"not a function": {
			code: "var x = 1",
		},
		"unclosed receiver": {
			code: "func (c *chat",
		},
		"no parameter list": {
			code: "func broken",
		},
		"unclosed type params": {
			code:     "func f[T any(",
			wantName: "f",
		},
		"unclosed params": {
			code:     "func f(a int",
			wantName: "f",
		},
		"no params no returns": {
			code:     "func f() {}",
			wantName: "f",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotName, gotRecv, gotTP, gotParams, gotReturns := parseFuncSignatureForTokens(tt.code)
			if gotName != tt.wantName {
				t.Errorf("name = %q, want %q", gotName, tt.wantName)
			}
			if gotRecv != tt.wantReceiver {
				t.Errorf("receiver = %q, want %q", gotRecv, tt.wantReceiver)
			}
			if gotTP != tt.wantTypeParams {
				t.Errorf("typeParams = %q, want %q", gotTP, tt.wantTypeParams)
			}
			if len(gotParams) != len(tt.wantParams) {
				t.Fatalf("params = %+v, want %+v", gotParams, tt.wantParams)
			}
			for i := range gotParams {
				if gotParams[i] != tt.wantParams[i] {
					t.Errorf("param %d = %+v, want %+v", i, gotParams[i], tt.wantParams[i])
				}
			}
			if gotReturns != tt.wantReturns {
				t.Errorf("returns = %q, want %q", gotReturns, tt.wantReturns)
			}
		})
	}
}

func TestExtractVarDeclarationsWithPositions(t *testing.T) {
	type tc struct {
		code string
		want []varDecl
	}

	tests := map[string]tc{
		"short form multi": {
			code: "a, b := 1, 2",
			want: []varDecl{{name: "a", offset: 0}, {name: "b", offset: 3}},
		},
		"short form with blank": {
			code: "_, v := f()",
			want: []varDecl{{name: "v", offset: 3}},
		},
		"var form multi": {
			code: "var x, y = 1, 2",
			want: []varDecl{{name: "x", offset: 4}, {name: "y", offset: 7}},
		},
		"var form without assignment": {
			code: "var x int",
			want: nil,
		},
		"no declaration": {
			code: "f(x)",
			want: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := extractVarDeclarationsWithPositions(tt.code)
			if len(got) != len(tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("decl %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFindElseKeyword_EdgeCases(t *testing.T) {
	ifStmt := &tuigen.IfStmt{Position: tuigen.Position{Line: 1, Column: 1}}

	t.Run("nil docs accessor", func(t *testing.T) {
		sp := &semanticTokensProvider{fnChecker: &stubFnChecker{names: map[string]bool{}}}
		line, col := sp.findElseKeyword(ifStmt)
		if line != -1 || col != -1 {
			t.Errorf("expected (-1, -1), got (%d, %d)", line, col)
		}
	})

	t.Run("document not found", func(t *testing.T) {
		sp := &semanticTokensProvider{
			fnChecker:  &stubFnChecker{names: map[string]bool{}},
			docs:       &stubDocAccessor{},
			currentURI: "file:///missing.gsx",
		}
		line, col := sp.findElseKeyword(ifStmt)
		if line != -1 || col != -1 {
			t.Errorf("expected (-1, -1), got (%d, %d)", line, col)
		}
	})
}

func TestIsHexDigit(t *testing.T) {
	type tc struct {
		c    byte
		want bool
	}

	tests := map[string]tc{
		"digit":            {c: '7', want: true},
		"lowercase hex":    {c: 'f', want: true},
		"uppercase hex":    {c: 'B', want: true},
		"out of range low": {c: 'g', want: false},
		"symbol":           {c: '-', want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := isHexDigit(tt.c); got != tt.want {
				t.Errorf("isHexDigit(%q) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	type tc struct {
		s    string
		want bool
	}

	tests := map[string]tc{
		"empty":              {s: "", want: false},
		"simple":             {s: "abc", want: true},
		"with digits":        {s: "a1b2", want: true},
		"leading digit":      {s: "1abc", want: false},
		"leading underscore": {s: "_x", want: true},
		"hyphenated":         {s: "x-y", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := isValidIdentifier(tt.s); got != tt.want {
				t.Errorf("isValidIdentifier(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestSemanticTokens_BlockCommentMultiline(t *testing.T) {
	content := `package main

/* first line
   second line */
templ Hello() {
	<span>Hi</span>
}
`
	doc := parseTestDoc(content)
	sp := newTestSemanticProvider()

	result, err := sp.SemanticTokensFull(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tokens := decodeTokens(result.Data)

	commentCount := countByType(tokens, TokenTypeComment)
	if commentCount < 2 {
		t.Errorf("expected at least 2 comment tokens for multi-line block comment, got %d: %+v", commentCount, tokens)
	}
	if !hasTokenAt(tokens, 2, 0, len("/* first line"), TokenTypeComment) {
		t.Errorf("expected first comment line token at 2:0, tokens: %+v", tokens)
	}
	// Continuation lines keep the full line length (including the leading
	// whitespace already counted in StartChar), so the length is 17 here.
	if !hasTokenAt(tokens, 3, 3, len("   second line */"), TokenTypeComment) {
		t.Errorf("expected second comment line token at 3:3, tokens: %+v", tokens)
	}
}

func TestNodeKindString(t *testing.T) {
	type tc struct {
		kind NodeKind
		want string
	}

	tests := map[string]tc{
		"component":      {kind: NodeKindComponent, want: "Component"},
		"element":        {kind: NodeKindElement, want: "Element"},
		"attribute":      {kind: NodeKindAttribute, want: "Attribute"},
		"ref attr":       {kind: NodeKindRefAttr, want: "RefAttr"},
		"go expr":        {kind: NodeKindGoExpr, want: "GoExpr"},
		"for loop":       {kind: NodeKindForLoop, want: "ForLoop"},
		"if stmt":        {kind: NodeKindIfStmt, want: "IfStmt"},
		"let binding":    {kind: NodeKindLetBinding, want: "LetBinding"},
		"state decl":     {kind: NodeKindStateDecl, want: "StateDecl"},
		"state access":   {kind: NodeKindStateAccess, want: "StateAccess"},
		"parameter":      {kind: NodeKindParameter, want: "Parameter"},
		"function":       {kind: NodeKindFunction, want: "Function"},
		"go decl":        {kind: NodeKindGoDecl, want: "GoDecl"},
		"component call": {kind: NodeKindComponentCall, want: "ComponentCall"},
		"event handler":  {kind: NodeKindEventHandler, want: "EventHandler"},
		"text":           {kind: NodeKindText, want: "Text"},
		"keyword":        {kind: NodeKindKeyword, want: "Keyword"},
		"tailwind class": {kind: NodeKindTailwindClass, want: "TailwindClass"},
		"import path":    {kind: NodeKindImportPath, want: "ImportPath"},
		"unknown":        {kind: NodeKindUnknown, want: "Unknown"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPositionToOffset_EndOfContent(t *testing.T) {
	type tc struct {
		content string
		pos     Position
		want    int
	}

	tests := map[string]tc{
		"line after final newline": {
			content: "a\n",
			pos:     Position{Line: 1, Character: 0},
			want:    2,
		},
		"line beyond content": {
			content: "a",
			pos:     Position{Line: 5, Character: 0},
			want:    1,
		},
		"within first line": {
			content: "abc",
			pos:     Position{Line: 0, Character: 2},
			want:    2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := PositionToOffset(tt.content, tt.pos); got != tt.want {
				t.Errorf("PositionToOffset(%q, %+v) = %d, want %d", tt.content, tt.pos, got, tt.want)
			}
		})
	}
}
