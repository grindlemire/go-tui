package lsp

import (
	"encoding/json"
	"testing"
)

func TestDefinitionDirect(t *testing.T) {
	type tc struct {
		content     string
		line        int
		character   int
		wantDefined bool
	}

	tests := map[string]tc{
		"component definition from call": {
			content: `package main

templ Header() {
	<span>Header</span>
}

templ Main() {
	@Header()
}
`,
			line:        7, // @Header() call (0-indexed)
			character:   2,
			wantDefined: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a server and test via router
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			params, _ := json.Marshal(DefinitionParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
				Position:     Position{Line: tt.line, Character: tt.character},
			})

			result, rpcErr := server.router.Route(Request{
				Method: "textDocument/definition",
				Params: params,
			})

			if rpcErr != nil {
				t.Fatalf("definition error: %v", rpcErr)
			}

			if tt.wantDefined {
				if result == nil {
					t.Error("expected definition result, got nil")
				}
			}
		})
	}
}

func TestHoverDirect(t *testing.T) {
	type tc struct {
		content   string
		line      int
		character int
		wantHover bool
	}

	tests := map[string]tc{
		"hover on component call": {
			content: `package main

templ Header(title string) {
	<span>{title}</span>
}

templ Main() {
	@Header("test")
}
`,
			line:      7, // @Header("test") (0-indexed)
			character: 2,
			wantHover: true,
		},
		"hover on element tag": {
			content: `package main

templ Hello() {
	<div padding={1}>
		<span>Hello</span>
	</div>
}
`,
			line:      3, // <div> (0-indexed)
			character: 2,
			wantHover: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			params, _ := json.Marshal(HoverParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
				Position:     Position{Line: tt.line, Character: tt.character},
			})

			result, rpcErr := server.router.Route(Request{
				Method: "textDocument/hover",
				Params: params,
			})

			if rpcErr != nil {
				t.Fatalf("hover error: %v", rpcErr)
			}

			if tt.wantHover {
				if result == nil {
					t.Error("expected hover result, got nil")
				}
			}
		})
	}
}

func TestCompletionDirect(t *testing.T) {
	type tc struct {
		content   string
		line      int
		character int
		trigger   string
		wantItems bool
	}

	tests := map[string]tc{
		"after @": {
			content: `package main

templ Hello() {
	<span>Hello</span>
}

templ Main() {
	@
}
`,
			line:      7,
			character: 2,
			trigger:   "@",
			wantItems: true,
		},
		"after <": {
			content: `package main

templ Hello() {
	<
}
`,
			line:      3,
			character: 2,
			trigger:   "<",
			wantItems: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			completionParams := CompletionParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
				Position:     Position{Line: tt.line, Character: tt.character},
			}
			if tt.trigger != "" {
				completionParams.Context = &CompletionContext{
					TriggerKind:      2,
					TriggerCharacter: tt.trigger,
				}
			}

			params, _ := json.Marshal(completionParams)
			result, rpcErr := server.router.Route(Request{Method: "textDocument/completion", Params: params})

			if rpcErr != nil {
				t.Fatalf("handleCompletion error: %v", rpcErr)
			}

			if tt.wantItems {
				list, ok := result.(*CompletionList)
				if !ok {
					t.Fatalf("expected CompletionList, got %T", result)
				}
				if len(list.Items) == 0 {
					t.Error("expected completion items, got none")
				}
			}
		})
	}
}

func TestDocumentSymbolDirect(t *testing.T) {
	type tc struct {
		content     string
		wantSymbols int
	}

	tests := map[string]tc{
		"single component": {
			content: `package main

templ Hello() {
	<span>Hello</span>
}
`,
			wantSymbols: 1,
		},
		"multiple components": {
			content: `package main

templ Header() {
	<span>Header</span>
}

templ Footer() {
	<span>Footer</span>
}

templ Main() {
	@Header()
	@Footer()
}
`,
			wantSymbols: 3,
		},
		"component with go func": {
			content: `package main

templ Hello() {
	<span>Hello</span>
}

func helper() string {
	return "test"
}
`,
			wantSymbols: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			params, _ := json.Marshal(DocumentSymbolParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
			})

			result, rpcErr := server.router.Route(Request{Method: "textDocument/documentSymbol", Params: params})

			if rpcErr != nil {
				t.Fatalf("handleDocumentSymbol error: %v", rpcErr)
			}

			symbols, ok := result.([]DocumentSymbol)
			if !ok {
				t.Fatalf("expected []DocumentSymbol, got %T", result)
			}

			if len(symbols) != tt.wantSymbols {
				t.Errorf("got %d symbols, want %d", len(symbols), tt.wantSymbols)
			}
		})
	}
}
