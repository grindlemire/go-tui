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

	type tc struct {
		name  string
		check func() bool
	}

	tests := []tc{
		{"TextDocumentSync is set", func() bool { return caps.TextDocumentSync != nil }},
		{"TextDocumentSync OpenClose", func() bool { return caps.TextDocumentSync.OpenClose }},
		{"TextDocumentSync FullSync", func() bool { return caps.TextDocumentSync.Change == TextDocumentSyncKindFull }},
		{"CompletionProvider is set", func() bool { return caps.CompletionProvider != nil }},
		{"HoverProvider", func() bool { return caps.HoverProvider }},
		{"DefinitionProvider", func() bool { return caps.DefinitionProvider }},
		{"ReferencesProvider", func() bool { return caps.ReferencesProvider }},
		{"DocumentSymbolProvider", func() bool { return caps.DocumentSymbolProvider }},
		{"WorkspaceSymbolProvider", func() bool { return caps.WorkspaceSymbolProvider }},
		{"DocumentFormattingProvider", func() bool { return caps.DocumentFormattingProvider }},
		{"SemanticTokensProvider is set", func() bool { return caps.SemanticTokensProvider != nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("capability check failed: %s", tt.name)
			}
		})
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
