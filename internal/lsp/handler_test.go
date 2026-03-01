package lsp

import (
	"bytes"
	"encoding/json"
	"testing"
)

func newTestServer() *Server {
	return NewServer(bytes.NewReader(nil), &bytes.Buffer{})
}

func TestHandleInitialize_ReturnsCapabilities(t *testing.T) {
	s := newTestServer()

	params, _ := json.Marshal(InitializeParams{
		RootURI: "file:///tmp/test",
	})

	result, lspErr := s.handleInitialize(params)
	if lspErr != nil {
		t.Fatalf("handleInitialize returned error: %v", lspErr)
	}

	initResult, ok := result.(InitializeResult)
	if !ok {
		t.Fatalf("expected InitializeResult, got %T", result)
	}

	caps := initResult.Capabilities

	if caps.TextDocumentSync == nil {
		t.Fatal("TextDocumentSync should be set")
	}
	if !caps.TextDocumentSync.OpenClose {
		t.Error("TextDocumentSync.OpenClose should be true")
	}
	if caps.TextDocumentSync.Change != TextDocumentSyncKindFull {
		t.Errorf("TextDocumentSync.Change = %d, want %d", caps.TextDocumentSync.Change, TextDocumentSyncKindFull)
	}
	if caps.CompletionProvider == nil {
		t.Error("CompletionProvider should be set")
	}

	type tc struct {
		got  bool
		want bool
	}

	tests := map[string]tc{
		"HoverProvider":              {got: caps.HoverProvider, want: true},
		"DefinitionProvider":         {got: caps.DefinitionProvider, want: true},
		"ReferencesProvider":         {got: caps.ReferencesProvider, want: true},
		"DocumentSymbolProvider":     {got: caps.DocumentSymbolProvider, want: true},
		"WorkspaceSymbolProvider":    {got: caps.WorkspaceSymbolProvider, want: true},
		"DocumentFormattingProvider": {got: caps.DocumentFormattingProvider, want: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", name, tt.got, tt.want)
			}
		})
	}

	if caps.SemanticTokensProvider == nil {
		t.Error("SemanticTokensProvider should be set")
	}
}

func TestHandleInitialize_SetsRootURI(t *testing.T) {
	s := newTestServer()

	params, _ := json.Marshal(InitializeParams{
		RootURI: "file:///workspace/myproject",
	})

	_, lspErr := s.handleInitialize(params)
	if lspErr != nil {
		t.Fatalf("handleInitialize returned error: %v", lspErr)
	}

	if s.rootURI != "file:///workspace/myproject" {
		t.Errorf("rootURI = %q, want %q", s.rootURI, "file:///workspace/myproject")
	}
}

func TestHandleInitialize_InvalidParams(t *testing.T) {
	s := newTestServer()

	_, lspErr := s.handleInitialize([]byte(`{invalid json`))
	if lspErr == nil {
		t.Fatal("expected error for invalid params")
	}
	if lspErr.Code != CodeInvalidParams {
		t.Errorf("error code = %d, want %d", lspErr.Code, CodeInvalidParams)
	}
}

func TestHandleShutdown(t *testing.T) {
	s := newTestServer()

	if s.shutdown {
		t.Fatal("server should not be shut down initially")
	}

	_, lspErr := s.handleShutdown()
	if lspErr != nil {
		t.Fatalf("handleShutdown returned error: %v", lspErr)
	}

	if !s.shutdown {
		t.Error("server should be shut down after handleShutdown")
	}
}

func TestHandleInitialized(t *testing.T) {
	s := newTestServer()

	if s.initialized {
		t.Fatal("server should not be initialized initially")
	}

	_, lspErr := s.handleInitialized()
	if lspErr != nil {
		t.Fatalf("handleInitialized returned error: %v", lspErr)
	}

	if !s.initialized {
		t.Error("server should be initialized after handleInitialized")
	}
}
