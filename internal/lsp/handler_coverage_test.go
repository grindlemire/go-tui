package lsp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleExit(t *testing.T) {
	s := NewServer(strings.NewReader(""), new(bytes.Buffer))
	result, rpcErr := s.router.Route(Request{JSONRPC: "2.0", Method: "exit"})
	if result != nil || rpcErr != nil {
		t.Errorf("exit: result=%v err=%v, want nil/nil", result, rpcErr)
	}
}

func TestDocumentSync_InvalidParams(t *testing.T) {
	type tc struct {
		method string
	}

	tests := map[string]tc{
		"didOpen":   {method: "textDocument/didOpen"},
		"didChange": {method: "textDocument/didChange"},
		"didClose":  {method: "textDocument/didClose"},
		"didSave":   {method: "textDocument/didSave"},
	}

	s := NewServer(strings.NewReader(""), new(bytes.Buffer))

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, rpcErr := s.router.Route(Request{
				JSONRPC: "2.0", Method: tt.method, Params: json.RawMessage(`[]`),
			})
			if rpcErr == nil || rpcErr.Code != CodeInvalidParams {
				t.Fatalf("error = %v, want code %d", rpcErr, CodeInvalidParams)
			}
		})
	}
}

func TestHandleDidChange_NoContentChanges(t *testing.T) {
	uri := "file:///nc.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	params := mustJSON(t, DidChangeParams{
		TextDocument:   VersionedTextDocumentIdentifier{URI: uri, Version: 2},
		ContentChanges: nil,
	})
	result, rpcErr := s.router.Route(Request{JSONRPC: "2.0", Method: "textDocument/didChange", Params: params})
	if result != nil || rpcErr != nil {
		t.Fatalf("result=%v err=%v, want nil/nil", result, rpcErr)
	}
	// Document is untouched.
	doc := s.docs.Get(uri)
	if doc == nil || doc.Version != 1 {
		t.Errorf("document version changed unexpectedly: %+v", doc)
	}
}

func TestHandleDidClose_UnopenedDocument(t *testing.T) {
	s := NewServer(strings.NewReader(""), new(bytes.Buffer))
	uri := "file:///never-opened.gsx"

	params := mustJSON(t, DidCloseParams{TextDocument: TextDocumentIdentifier{URI: uri}})
	result, rpcErr := s.router.Route(Request{JSONRPC: "2.0", Method: "textDocument/didClose", Params: params})
	if result != nil || rpcErr != nil {
		t.Fatalf("result=%v err=%v, want nil/nil", result, rpcErr)
	}
	// Nothing should be cached for the URI.
	s.workspaceASTsMu.RLock()
	_, cached := s.workspaceASTs[uri]
	s.workspaceASTsMu.RUnlock()
	if cached {
		t.Error("unexpected workspace AST cache entry for unopened document")
	}
}

func TestHandleDidSave(t *testing.T) {
	uri := "file:///save.gsx"
	s, _ := newRoutedServer(t, uri, routerCoverageSrc)

	updated := strings.Replace(routerCoverageSrc, "templ Header", "templ Banner", 1)

	// Save with text updates the document and re-indexes.
	params := mustJSON(t, DidSaveParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Text:         &updated,
	})
	if _, rpcErr := s.router.Route(Request{JSONRPC: "2.0", Method: "textDocument/didSave", Params: params}); rpcErr != nil {
		t.Fatalf("didSave: %v", rpcErr)
	}

	doc := s.docs.Get(uri)
	if doc == nil {
		t.Fatal("document missing after save")
	}
	if doc.Version != 2 {
		t.Errorf("version = %d, want 2", doc.Version)
	}
	if !strings.Contains(doc.Content, "templ Banner") {
		t.Error("document content was not updated from save text")
	}
	if _, ok := s.index.Lookup("Banner"); !ok {
		t.Error("index was not refreshed with renamed component")
	}

	// Save without text is a no-op.
	noText := mustJSON(t, DidSaveParams{TextDocument: TextDocumentIdentifier{URI: uri}})
	if _, rpcErr := s.router.Route(Request{JSONRPC: "2.0", Method: "textDocument/didSave", Params: noText}); rpcErr != nil {
		t.Fatalf("didSave without text: %v", rpcErr)
	}
	if got := s.docs.Get(uri).Version; got != 2 {
		t.Errorf("version after textless save = %d, want 2", got)
	}

	// Save with text for a document that is not open does nothing.
	other := "package main"
	closed := mustJSON(t, DidSaveParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///closed.gsx"},
		Text:         &other,
	})
	if _, rpcErr := s.router.Route(Request{JSONRPC: "2.0", Method: "textDocument/didSave", Params: closed}); rpcErr != nil {
		t.Fatalf("didSave closed doc: %v", rpcErr)
	}
	if s.docs.Get("file:///closed.gsx") != nil {
		t.Error("save must not open a closed document")
	}
}

func TestIndexWorkspace(t *testing.T) {
	tmp := t.TempDir()

	writeFile := func(rel, content string) string {
		t.Helper()
		path := filepath.Join(tmp, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	writeFile("good.gsx", "package main\n\ntempl Good() {\n\t<span>ok</span>\n}\n")
	writeFile("broken.gsx", "package main\n\ntempl Broken( {\n")
	writeFile("notes.txt", "not a gsx file")
	writeFile(".hidden/skipped.gsx", "package main\n\ntempl Hidden() {\n\t<span>no</span>\n}\n")
	writeFile("vendor/v.gsx", "package main\n\ntempl Vendored() {\n\t<span>no</span>\n}\n")
	writeFile("node_modules/n.gsx", "package main\n\ntempl NodeMod() {\n\t<span>no</span>\n}\n")
	openPath := writeFile("open.gsx", "package main\n\ntempl AlreadyOpen() {\n\t<span>open</span>\n}\n")

	s := NewServer(strings.NewReader(""), new(bytes.Buffer))
	s.rootURI = "file://" + tmp

	// Open one file through the document manager so the walker skips it.
	openURI := "file://" + openPath
	s.docs.Open(openURI, "package main\n\ntempl AlreadyOpen() {\n\t<span>open</span>\n}\n", 1)

	s.indexWorkspace()

	if _, ok := s.index.Lookup("Good"); !ok {
		t.Error("Good component was not indexed")
	}
	for _, name := range []string{"Hidden", "Vendored", "NodeMod"} {
		if _, ok := s.index.Lookup(name); ok {
			t.Errorf("%s should have been skipped by directory filters", name)
		}
	}

	s.workspaceASTsMu.RLock()
	_, goodCached := s.workspaceASTs["file://"+filepath.Join(tmp, "good.gsx")]
	_, openCached := s.workspaceASTs[openURI]
	s.workspaceASTsMu.RUnlock()
	if !goodCached {
		t.Error("good.gsx AST missing from workspace cache")
	}
	if openCached {
		t.Error("open document must not be re-cached by the workspace walk")
	}
}

func TestIndexWorkspace_NoRoot(t *testing.T) {
	s := NewServer(strings.NewReader(""), new(bytes.Buffer))
	s.rootURI = ""
	s.indexWorkspace() // must not panic or index anything
	if got := len(s.index.All()); got != 0 {
		t.Errorf("indexed %d components without a root", got)
	}
}
