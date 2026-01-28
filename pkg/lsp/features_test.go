package lsp

import (
	"encoding/json"
	"testing"
)

func TestComponentIndex(t *testing.T) {
	type tc struct {
		content      string
		wantComps    []string
		lookupName   string
		lookupExists bool
	}

	tests := map[string]tc{
		"single component": {
			content: `package main

templ Hello() {
	<span>Hello</span>
}
`,
			wantComps:    []string{"Hello"},
			lookupName:   "Hello",
			lookupExists: true,
		},
		"multiple components": {
			content: `package main

templ Header() {
	<span>Header</span>
}

templ Footer() {
	<span>Footer</span>
}
`,
			wantComps:    []string{"Header", "Footer"},
			lookupName:   "Footer",
			lookupExists: true,
		},
		"lookup nonexistent": {
			content: `package main

templ Hello() {
	<span>Hello</span>
}
`,
			wantComps:    []string{"Hello"},
			lookupName:   "NotExists",
			lookupExists: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dm := NewDocumentManager()
			idx := NewComponentIndex()

			uri := "file:///test.gsx"
			doc := dm.Open(uri, tt.content, 1)

			idx.IndexDocument(uri, doc.AST)

			// Check all expected components are indexed
			for _, compName := range tt.wantComps {
				if _, ok := idx.Lookup(compName); !ok {
					t.Errorf("expected component %s to be indexed", compName)
				}
			}

			// Test lookup
			_, exists := idx.Lookup(tt.lookupName)
			if exists != tt.lookupExists {
				t.Errorf("Lookup(%s) = _, %v; want _, %v", tt.lookupName, exists, tt.lookupExists)
			}
		})
	}
}

func TestComponentIndexRemove(t *testing.T) {
	dm := NewDocumentManager()
	idx := NewComponentIndex()

	uri := "file:///test.gsx"
	content := `package main

templ Hello() {
	<span>Hello</span>
}
`
	doc := dm.Open(uri, content, 1)
	idx.IndexDocument(uri, doc.AST)

	// Verify component is indexed
	if _, ok := idx.Lookup("Hello"); !ok {
		t.Fatal("expected Hello to be indexed")
	}

	// Remove the file
	idx.Remove(uri)

	// Verify component is removed
	if _, ok := idx.Lookup("Hello"); ok {
		t.Fatal("expected Hello to be removed from index")
	}
}

// testServer runs a server with the given requests and returns responses by ID.
func testServer(t *testing.T, requests func(m *mockReadWriter, uri string) int) (map[int]*Response, *Server) {
	t.Helper()

	mock := newMockReadWriter()
	uri := "file:///test.gsx"

	// Send requests
	maxID := requests(mock, uri)

	server := NewServer(mock.input, mock.output)
	if err := server.Run(t.Context()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Read all responses
	responses := make(map[int]*Response)
	for i := 0; i <= maxID; i++ {
		resp, err := mock.readResponse()
		if err != nil {
			break
		}
		if resp.ID != nil {
			switch id := resp.ID.(type) {
			case float64:
				responses[int(id)] = resp
			case int:
				responses[id] = resp
			}
		}
		// Skip notifications
	}

	return responses, server
}

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
			// Create a server and test directly via handlers
			server := NewServer(nil, nil)

			doc := server.docs.Open("file:///test.gsx", tt.content, 1)
			server.index.IndexDocument("file:///test.gsx", doc.AST)

			params, _ := json.Marshal(DefinitionParams{
				TextDocument: TextDocumentIdentifier{URI: "file:///test.gsx"},
				Position:     Position{Line: tt.line, Character: tt.character},
			})

			result, rpcErr := server.handleDefinition(params)

			if rpcErr != nil {
				t.Fatalf("handleDefinition error: %v", rpcErr)
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

			result, rpcErr := server.handleHover(params)

			if rpcErr != nil {
				t.Fatalf("handleHover error: %v", rpcErr)
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
			result, rpcErr := server.handleCompletion(params)

			if rpcErr != nil {
				t.Fatalf("handleCompletion error: %v", rpcErr)
			}

			if tt.wantItems {
				list, ok := result.(CompletionList)
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

			result, rpcErr := server.handleDocumentSymbol(params)

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

func TestWorkspaceSymbolDirect(t *testing.T) {
	type tc struct {
		contents    map[string]string
		query       string
		wantSymbols int
	}

	tests := map[string]tc{
		"empty query returns all": {
			contents: map[string]string{
				"file:///a.gsx": `package main

templ Hello() {
	<span>Hello</span>
}
`,
				"file:///b.gsx": `package main

templ World() {
	<span>World</span>
}
`,
			},
			query:       "",
			wantSymbols: 2,
		},
		"filter by query": {
			contents: map[string]string{
				"file:///a.gsx": `package main

templ Hello() {
	<span>Hello</span>
}
`,
				"file:///b.gsx": `package main

templ World() {
	<span>World</span>
}
`,
			},
			query:       "Hello",
			wantSymbols: 1,
		},
		"case insensitive query": {
			contents: map[string]string{
				"file:///a.gsx": `package main

templ HelloWorld() {
	<span>Hello</span>
}
`,
			},
			query:       "hello",
			wantSymbols: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)

			for uri, content := range tt.contents {
				doc := server.docs.Open(uri, content, 1)
				server.index.IndexDocument(uri, doc.AST)
			}

			params, _ := json.Marshal(WorkspaceSymbolParams{
				Query: tt.query,
			})

			result, rpcErr := server.handleWorkspaceSymbol(params)

			if rpcErr != nil {
				t.Fatalf("handleWorkspaceSymbol error: %v", rpcErr)
			}

			symbols, ok := result.([]SymbolInformation)
			if !ok {
				t.Fatalf("expected []SymbolInformation, got %T", result)
			}

			if len(symbols) != tt.wantSymbols {
				t.Errorf("got %d symbols, want %d", len(symbols), tt.wantSymbols)
			}
		})
	}
}

func TestGetElementAttributes(t *testing.T) {
	type tc struct {
		tag       string
		wantAttrs bool
	}

	tests := map[string]tc{
		"div element": {
			tag:       "div",
			wantAttrs: true,
		},
		"span element": {
			tag:       "span",
			wantAttrs: true,
		},
		"input element": {
			tag:       "input",
			wantAttrs: true,
		},
		"button element": {
			tag:       "button",
			wantAttrs: true,
		},
		"unknown element": {
			tag:       "unknown",
			wantAttrs: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			attrs := getElementAttributes(tt.tag)
			if tt.wantAttrs && len(attrs) == 0 {
				t.Error("expected attributes, got none")
			}
			if !tt.wantAttrs && len(attrs) > 0 {
				t.Errorf("expected no attributes, got %d", len(attrs))
			}
		})
	}
}

func TestIsElementTag(t *testing.T) {
	type tc struct {
		word string
		want bool
	}

	tests := map[string]tc{
		"div":      {word: "div", want: true},
		"span":     {word: "span", want: true},
		"p":        {word: "p", want: true},
		"ul":       {word: "ul", want: true},
		"li":       {word: "li", want: true},
		"button":   {word: "button", want: true},
		"input":    {word: "input", want: true},
		"table":    {word: "table", want: true},
		"progress": {word: "progress", want: true},
		"unknown":  {word: "unknown", want: false},
		"empty":    {word: "", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := isElementTag(tt.word)
			if got != tt.want {
				t.Errorf("isElementTag(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func TestIsInClassAttribute(t *testing.T) {
	type tc struct {
		content    string
		line       int
		character  int
		wantInAttr bool
		wantPrefix string
	}

	tests := map[string]tc{
		"inside class attribute empty": {
			content: `package main

templ Hello() {
	<div class="">
	</div>
}
`,
			line:       3,
			character:  13, // Right after the opening quote
			wantInAttr: true,
			wantPrefix: "",
		},
		"inside class attribute with prefix": {
			content: `package main

templ Hello() {
	<div class="flex">
	</div>
}
`,
			line:       3,
			character:  17, // After "flex"
			wantInAttr: true,
			wantPrefix: "flex",
		},
		"inside class attribute partial class after space": {
			content: `package main

templ Hello() {
	<div class="flex-col gap">
	</div>
}
`,
			line:       3,
			character:  25, // After "gap"
			wantInAttr: true,
			wantPrefix: "gap",
		},
		"inside class attribute at space": {
			content: `package main

templ Hello() {
	<div class="flex-col ">
	</div>
}
`,
			line:       3,
			character:  22, // After space after "flex-col "
			wantInAttr: true,
			wantPrefix: "",
		},
		"not in class attribute - in id": {
			content: `package main

templ Hello() {
	<div id="test">
	</div>
}
`,
			line:       3,
			character:  13, // Inside id attribute
			wantInAttr: false,
			wantPrefix: "",
		},
		"not in class attribute - outside quotes": {
			content: `package main

templ Hello() {
	<div class="flex">
	</div>
}
`,
			line:       3,
			character:  6, // On "div"
			wantInAttr: false,
			wantPrefix: "",
		},
		"not in class attribute - different line": {
			content: `package main

templ Hello() {
	<div class="flex">
		<span>Hello</span>
	</div>
}
`,
			line:       4,
			character:  10, // Inside span
			wantInAttr: false,
			wantPrefix: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)
			server.docs.Open("file:///test.gsx", tt.content, 1)
			doc := server.docs.Get("file:///test.gsx")

			gotInAttr, gotPrefix := server.isInClassAttribute(doc, Position{Line: tt.line, Character: tt.character})

			if gotInAttr != tt.wantInAttr {
				t.Errorf("isInClassAttribute() inAttr = %v, want %v", gotInAttr, tt.wantInAttr)
			}
			if gotPrefix != tt.wantPrefix {
				t.Errorf("isInClassAttribute() prefix = %q, want %q", gotPrefix, tt.wantPrefix)
			}
		})
	}
}

func TestGetTailwindCompletions(t *testing.T) {
	type tc struct {
		prefix    string
		wantCount int // -1 means we just check > 0
		wantFirst string
		checkAll  bool // if true, check that all returned items have the prefix
	}

	tests := map[string]tc{
		"empty prefix returns all classes": {
			prefix:    "",
			wantCount: -1, // We don't know exact count, just check > 0
			checkAll:  false,
		},
		"flex prefix filters correctly": {
			prefix:    "flex",
			wantCount: -1,
			checkAll:  true,
		},
		"gap prefix filters correctly": {
			prefix:    "gap",
			wantCount: -1,
			checkAll:  true,
		},
		"no matches": {
			prefix:    "zzzznotaclass",
			wantCount: 0,
			checkAll:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewServer(nil, nil)
			items := server.getTailwindCompletions(tt.prefix)

			if tt.wantCount == -1 {
				if len(items) == 0 {
					t.Error("expected completion items, got none")
				}
			} else if len(items) != tt.wantCount {
				t.Errorf("got %d items, want %d", len(items), tt.wantCount)
			}

			// Check that all items have the prefix if requested
			if tt.checkAll {
				for _, item := range items {
					if len(item.Label) < len(tt.prefix) || item.Label[:len(tt.prefix)] != tt.prefix {
						t.Errorf("item %q does not have prefix %q", item.Label, tt.prefix)
					}
				}
			}

			// Check that completion items have documentation
			for _, item := range items {
				if item.Documentation == nil || item.Documentation.Value == "" {
					t.Errorf("item %q missing documentation", item.Label)
				}
				if item.Detail == "" {
					t.Errorf("item %q missing detail (category)", item.Label)
				}
			}
		})
	}
}

func TestTailwindCompletionInCompletion(t *testing.T) {
	type tc struct {
		content      string
		line         int
		character    int
		wantTailwind bool // true if we expect Tailwind completions
	}

	tests := map[string]tc{
		"inside class attribute": {
			content: `package main

templ Hello() {
	<div class="flex">
	</div>
}
`,
			line:         3,
			character:    13, // Inside class=""
			wantTailwind: true,
		},
		"not inside class attribute": {
			content: `package main

templ Hello() {
	<div id="test">
	</div>
}
`,
			line:         3,
			character:    13, // Inside id=""
			wantTailwind: false,
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

			params, _ := json.Marshal(completionParams)
			result, rpcErr := server.handleCompletion(params)

			if rpcErr != nil {
				t.Fatalf("handleCompletion error: %v", rpcErr)
			}

			list, ok := result.(CompletionList)
			if !ok {
				t.Fatalf("expected CompletionList, got %T", result)
			}

			if tt.wantTailwind {
				if len(list.Items) == 0 {
					t.Error("expected Tailwind completion items, got none")
					return
				}
				// Check that we got Tailwind-style completions
				hasTailwindClass := false
				for _, item := range list.Items {
					if item.Label == "flex" || item.Label == "flex-col" || item.Label == "gap-1" {
						hasTailwindClass = true
						break
					}
				}
				if !hasTailwindClass {
					t.Error("expected Tailwind class completions, but didn't find expected classes")
				}
			}
		})
	}
}
