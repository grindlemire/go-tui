package lsp

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// routerCoverageSrc is a valid .gsx fixture used by feature dispatch tests.
const routerCoverageSrc = `package main

templ Header(title string) {
	<div class="p-1">
		<span class="font-bold">{title}</span>
	</div>
}

templ App() {
	<div class="flex-col">
		@Header("hello")
	</div>
}

func shout(msg string) string {
	return msg
}
`

// newRoutedServer creates a server with an output buffer and opens the given
// document through the router so all handler paths run.
func newRoutedServer(t *testing.T, uri, content string) (*Server, *bytes.Buffer) {
	t.Helper()
	out := new(bytes.Buffer)
	s := NewServer(strings.NewReader(""), out)

	params, err := json.Marshal(DidOpenParams{
		TextDocument: TextDocumentItem{URI: uri, LanguageID: "tui", Version: 1, Text: content},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, rpcErr := s.router.Route(Request{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: params}); rpcErr != nil {
		t.Fatalf("didOpen: %v", rpcErr)
	}
	return s, out
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func positionalParams(t *testing.T, uri string, line, char int) json.RawMessage {
	t.Helper()
	return mustJSON(t, map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": line, "character": char},
	})
}

func TestRouterRoute_UnknownMethod(t *testing.T) {
	s, _ := newRoutedServer(t, "file:///r.gsx", routerCoverageSrc)
	result, rpcErr := s.router.Route(Request{JSONRPC: "2.0", ID: 1, Method: "bogus/method"})
	if result != nil {
		t.Errorf("result = %v, want nil", result)
	}
	if rpcErr == nil || rpcErr.Code != CodeMethodNotFound {
		t.Fatalf("error = %v, want code %d", rpcErr, CodeMethodNotFound)
	}
	if !strings.Contains(rpcErr.Message, "bogus/method") {
		t.Errorf("message %q does not name the method", rpcErr.Message)
	}
}

func TestRouterRoute_InvalidParams(t *testing.T) {
	type tc struct {
		method string
	}

	tests := map[string]tc{
		"hover":           {method: "textDocument/hover"},
		"completion":      {method: "textDocument/completion"},
		"definition":      {method: "textDocument/definition"},
		"references":      {method: "textDocument/references"},
		"document symbol": {method: "textDocument/documentSymbol"},
		"workspace symbol": {
			method: "workspace/symbol",
		},
		"formatting":      {method: "textDocument/formatting"},
		"semantic tokens": {method: "textDocument/semanticTokens/full"},
	}

	s, _ := newRoutedServer(t, "file:///r.gsx", routerCoverageSrc)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// An array does not unmarshal into the params struct.
			result, rpcErr := s.router.Route(Request{
				JSONRPC: "2.0", ID: 1, Method: tt.method, Params: json.RawMessage(`[]`),
			})
			if rpcErr == nil || rpcErr.Code != CodeInvalidParams {
				t.Fatalf("error = %v, want code %d", rpcErr, CodeInvalidParams)
			}
			if result != nil {
				t.Errorf("result = %v, want nil", result)
			}
		})
	}
}

func TestRouterRoute_UnknownDocument(t *testing.T) {
	type tc struct {
		method string
		params json.RawMessage
	}

	missing := "file:///missing.gsx"
	s, _ := newRoutedServer(t, "file:///r.gsx", routerCoverageSrc)

	tests := map[string]tc{
		"hover":           {method: "textDocument/hover", params: positionalParams(t, missing, 0, 0)},
		"completion":      {method: "textDocument/completion", params: positionalParams(t, missing, 0, 0)},
		"definition":      {method: "textDocument/definition", params: positionalParams(t, missing, 0, 0)},
		"references":      {method: "textDocument/references", params: positionalParams(t, missing, 0, 0)},
		"document symbol": {method: "textDocument/documentSymbol", params: mustJSON(t, DocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: missing}})},
		"formatting":      {method: "textDocument/formatting", params: mustJSON(t, DocumentFormattingParams{TextDocument: TextDocumentIdentifier{URI: missing}})},
		"semantic tokens": {method: "textDocument/semanticTokens/full", params: mustJSON(t, SemanticTokensParams{TextDocument: TextDocumentIdentifier{URI: missing}})},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, rpcErr := s.router.Route(Request{JSONRPC: "2.0", ID: 1, Method: tt.method, Params: tt.params})
			if rpcErr != nil {
				t.Fatalf("unexpected error: %v", rpcErr)
			}
			if result != nil {
				t.Errorf("result = %#v, want nil for unknown document", result)
			}
		})
	}
}

func TestRouterRoute_Hover(t *testing.T) {
	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	// Cursor on the "div" tag of Header (line 3, after the tab and '<').
	result, rpcErr := s.router.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "textDocument/hover",
		Params: positionalParams(t, uri, 3, 2),
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	hover, ok := result.(*Hover)
	if !ok || hover == nil {
		t.Fatalf("result = %T, want *Hover", result)
	}
	if !strings.Contains(hover.Contents.Value, "div") {
		t.Errorf("hover contents %q do not mention div", hover.Contents.Value)
	}
}

func TestRouterRoute_Completion(t *testing.T) {
	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	// Cursor inside class="p-1" value on line 3.
	idx := strings.Index(strings.Split(routerCoverageSrc, "\n")[3], `p-1`)
	result, rpcErr := s.router.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "textDocument/completion",
		Params: positionalParams(t, uri, 3, idx+1),
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	list, ok := result.(*CompletionList)
	if !ok || list == nil {
		t.Fatalf("result = %T, want *CompletionList", result)
	}
	if len(list.Items) == 0 {
		t.Error("expected completion items inside class attribute")
	}
}

func TestRouterRoute_Definition(t *testing.T) {
	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	// Cursor on @Header call (line 10).
	idx := strings.Index(strings.Split(routerCoverageSrc, "\n")[10], "@Header")
	result, rpcErr := s.router.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "textDocument/definition",
		Params: positionalParams(t, uri, 10, idx+2),
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	locs, ok := result.([]Location)
	if !ok {
		t.Fatalf("result = %T, want []Location", result)
	}
	if len(locs) != 1 {
		t.Fatalf("got %d locations, want 1", len(locs))
	}
	if locs[0].URI != uri {
		t.Errorf("URI = %q, want %q", locs[0].URI, uri)
	}
	if locs[0].Range.Start.Line != 2 {
		t.Errorf("definition line = %d, want 2 (templ Header)", locs[0].Range.Start.Line)
	}
}

func TestRouterRoute_References(t *testing.T) {
	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	// Cursor on the Header component name (line 2).
	idx := strings.Index(strings.Split(routerCoverageSrc, "\n")[2], "Header")
	params := mustJSON(t, map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": 2, "character": idx + 1},
		"context":      map[string]any{"includeDeclaration": true},
	})
	result, rpcErr := s.router.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "textDocument/references", Params: params,
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	locs, ok := result.([]Location)
	if !ok {
		t.Fatalf("result = %T, want []Location", result)
	}
	if len(locs) == 0 {
		t.Fatal("expected at least one reference location")
	}
	foundCall := false
	for _, l := range locs {
		if l.Range.Start.Line == 10 {
			foundCall = true
		}
	}
	if !foundCall {
		t.Errorf("references %v do not include the @Header call on line 10", locs)
	}
}

func TestRouterRoute_DocumentSymbol(t *testing.T) {
	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	result, rpcErr := s.router.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "textDocument/documentSymbol",
		Params: mustJSON(t, DocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: uri}}),
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	symbols, ok := result.([]DocumentSymbol)
	if !ok {
		t.Fatalf("result = %T, want []DocumentSymbol", result)
	}
	names := map[string]bool{}
	for _, sym := range symbols {
		names[sym.Name] = true
	}
	if !names["Header"] || !names["App"] {
		t.Errorf("symbols = %v, want Header and App", names)
	}
}

func TestRouterRoute_WorkspaceSymbol(t *testing.T) {
	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)
	_ = uri

	result, rpcErr := s.router.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "workspace/symbol",
		Params: mustJSON(t, WorkspaceSymbolParams{Query: "Head"}),
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	symbols, ok := result.([]SymbolInformation)
	if !ok {
		t.Fatalf("result = %T, want []SymbolInformation", result)
	}
	found := false
	for _, sym := range symbols {
		if sym.Name == "Header" {
			found = true
		}
	}
	if !found {
		t.Errorf("symbols %v missing Header", symbols)
	}
}

func TestRouterRoute_Formatting(t *testing.T) {
	uri := "file:///fmt.gsx"
	// Sloppy spacing forces the formatter to produce an edit.
	sloppy := "package main\n\ntempl Messy() {\n\t<div    class=\"p-1\">\n\t\t<span>hi</span>\n\t</div>\n}\n"
	s, _ := newRoutedServer(t, uri, sloppy)

	result, rpcErr := s.router.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "textDocument/formatting",
		Params: mustJSON(t, DocumentFormattingParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Options:      FormattingOptions{TabSize: 4},
		}),
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	edits, ok := result.([]TextEdit)
	if !ok {
		t.Fatalf("result = %T, want []TextEdit", result)
	}
	if len(edits) != 1 {
		t.Fatalf("got %d edits, want 1 full-document edit", len(edits))
	}
	if !strings.Contains(edits[0].NewText, `<div class="p-1">`) {
		t.Errorf("formatted text did not collapse spaces:\n%s", edits[0].NewText)
	}
}

func TestRouterRoute_SemanticTokensFull(t *testing.T) {
	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	result, rpcErr := s.router.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "textDocument/semanticTokens/full",
		Params: mustJSON(t, SemanticTokensParams{TextDocument: TextDocumentIdentifier{URI: uri}}),
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	tokens, ok := result.(*SemanticTokens)
	if !ok || tokens == nil {
		t.Fatalf("result = %T, want *SemanticTokens", result)
	}
	if len(tokens.Data) == 0 {
		t.Fatal("expected semantic token data")
	}
	if len(tokens.Data)%5 != 0 {
		t.Errorf("token data length %d is not a multiple of 5", len(tokens.Data))
	}
}

func TestRouterRoute_NilRegistryFallbacks(t *testing.T) {
	type tc struct {
		method string
	}

	tests := map[string]tc{
		"hover":            {method: "textDocument/hover"},
		"completion":       {method: "textDocument/completion"},
		"definition":       {method: "textDocument/definition"},
		"references":       {method: "textDocument/references"},
		"document symbol":  {method: "textDocument/documentSymbol"},
		"workspace symbol": {method: "workspace/symbol"},
		"formatting":       {method: "textDocument/formatting"},
	}

	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)
	r := NewRouter(s, nil)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, rpcErr := r.Route(Request{
				JSONRPC: "2.0", ID: 1, Method: tt.method,
				Params: positionalParams(t, uri, 0, 0),
			})
			if rpcErr != nil {
				t.Fatalf("unexpected error: %v", rpcErr)
			}
			if result != nil {
				t.Errorf("result = %#v, want nil without registry", result)
			}
		})
	}

	// Semantic tokens returns an empty token set instead of nil.
	result, rpcErr := r.Route(Request{
		JSONRPC: "2.0", ID: 1, Method: "textDocument/semanticTokens/full",
		Params: mustJSON(t, SemanticTokensParams{TextDocument: TextDocumentIdentifier{URI: uri}}),
	})
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	tokens, ok := result.(*SemanticTokens)
	if !ok || tokens == nil {
		t.Fatalf("result = %T, want *SemanticTokens", result)
	}
	if len(tokens.Data) != 0 {
		t.Errorf("token data = %v, want empty", tokens.Data)
	}
}

// errProviders implements every lsp provider interface and always fails,
// exercising the router's internal-error paths.
type errProviders struct{}

var errProvider = errors.New("provider boom")

func (errProviders) Hover(*CursorContext) (*Hover, error)             { return nil, errProvider }
func (errProviders) Complete(*CursorContext) (*CompletionList, error) { return nil, errProvider }
func (errProviders) Definition(*CursorContext) ([]Location, error)    { return nil, errProvider }
func (errProviders) References(*CursorContext, bool) ([]Location, error) {
	return nil, errProvider
}
func (errProviders) DocumentSymbols(*Document) ([]DocumentSymbol, error) { return nil, errProvider }
func (errProviders) WorkspaceSymbols(string) ([]SymbolInformation, error) {
	return nil, errProvider
}
func (errProviders) Diagnose(*Document) ([]Diagnostic, error) { return nil, errProvider }
func (errProviders) Format(*Document, FormattingOptions) ([]TextEdit, error) {
	return nil, errProvider
}

func (errProviders) SemanticTokensFull(*Document) (*SemanticTokens, error) {
	return nil, errProvider
}

func TestRouterRoute_ProviderErrors(t *testing.T) {
	type tc struct {
		method string
		params json.RawMessage
	}

	uri := "file:///r.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	ep := errProviders{}
	r := NewRouter(s, &Registry{
		Hover:           ep,
		Completion:      ep,
		Definition:      ep,
		References:      ep,
		DocumentSymbol:  ep,
		WorkspaceSymbol: ep,
		Formatting:      ep,
		SemanticTokens:  ep,
	})

	tests := map[string]tc{
		"hover":            {method: "textDocument/hover", params: positionalParams(t, uri, 0, 0)},
		"references":       {method: "textDocument/references", params: positionalParams(t, uri, 0, 0)},
		"document symbol":  {method: "textDocument/documentSymbol", params: mustJSON(t, DocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: uri}})},
		"workspace symbol": {method: "workspace/symbol", params: mustJSON(t, WorkspaceSymbolParams{Query: "x"})},
		"formatting":       {method: "textDocument/formatting", params: mustJSON(t, DocumentFormattingParams{TextDocument: TextDocumentIdentifier{URI: uri}})},
		"semantic tokens":  {method: "textDocument/semanticTokens/full", params: mustJSON(t, SemanticTokensParams{TextDocument: TextDocumentIdentifier{URI: uri}})},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, rpcErr := r.Route(Request{JSONRPC: "2.0", ID: 1, Method: tt.method, Params: tt.params})
			if rpcErr == nil || rpcErr.Code != CodeInternalError {
				t.Fatalf("error = %v, want code %d", rpcErr, CodeInternalError)
			}
			if !strings.Contains(rpcErr.Message, "provider boom") {
				t.Errorf("message = %q, want provider error text", rpcErr.Message)
			}
			if result != nil {
				t.Errorf("result = %#v, want nil", result)
			}
		})
	}
}
