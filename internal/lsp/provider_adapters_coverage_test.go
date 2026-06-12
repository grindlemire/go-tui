package lsp

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/lsp/gopls"
	"github.com/grindlemire/go-tui/internal/lsp/provider"
)

func newAdapterServer(t *testing.T) (*Server, string) {
	t.Helper()
	s := NewServer(strings.NewReader(""), new(bytes.Buffer))
	uri := "file:///adapter.gsx"
	doc := s.docs.Open(uri, indexCoverageSrc, 1)
	s.index.IndexDocument(uri, doc.AST)
	return s, uri
}

func TestComponentIndexAdapter(t *testing.T) {
	s, uri := newAdapterServer(t)
	a := &componentIndexAdapter{index: s.index}

	t.Run("Lookup", func(t *testing.T) {
		info, ok := a.Lookup("Card")
		if !ok || info == nil {
			t.Fatal("Card not found")
		}
		if info.Name != "Card" || info.Location.URI != uri {
			t.Errorf("info = %+v", info)
		}
		if _, ok := a.Lookup("Nope"); ok {
			t.Error("Lookup(Nope) should not be found")
		}
	})

	t.Run("LookupFunc", func(t *testing.T) {
		info, ok := a.LookupFunc("format")
		if !ok || info == nil {
			t.Fatal("format not found")
		}
		if info.Returns != "string" {
			t.Errorf("Returns = %q, want string", info.Returns)
		}
		if _, ok := a.LookupFunc("nope"); ok {
			t.Error("LookupFunc(nope) should not be found")
		}
	})

	t.Run("LookupParam", func(t *testing.T) {
		info, ok := a.LookupParam("Card", "title")
		if !ok || info == nil {
			t.Fatal("Card.title not found")
		}
		if info.Type != "string" || info.ComponentName != "Card" {
			t.Errorf("info = %+v", info)
		}
		if _, ok := a.LookupParam("Card", "nope"); ok {
			t.Error("LookupParam(Card, nope) should not be found")
		}
	})

	t.Run("LookupFuncParam", func(t *testing.T) {
		info, ok := a.LookupFuncParam("format", "prefix")
		if !ok || info == nil {
			t.Fatal("format.prefix not found")
		}
		if info.FuncName != "format" || info.Type != "string" {
			t.Errorf("info = %+v", info)
		}
		if info.Location.URI != uri {
			t.Errorf("URI = %q, want %q", info.Location.URI, uri)
		}
		wantEnd := info.Location.Range.Start.Character + len("prefix")
		if info.Location.Range.End.Character != wantEnd {
			t.Errorf("end character = %d, want %d", info.Location.Range.End.Character, wantEnd)
		}
		if _, ok := a.LookupFuncParam("format", "nope"); ok {
			t.Error("LookupFuncParam(format, nope) should not be found")
		}
	})

	t.Run("All and AllFunctions", func(t *testing.T) {
		if got := a.All(); len(got) != 1 || got[0] != "Card" {
			t.Errorf("All = %v, want [Card]", got)
		}
		if got := a.AllFunctions(); len(got) != 1 || got[0] != "format" {
			t.Errorf("AllFunctions = %v, want [format]", got)
		}
	})
}

func TestDocumentAdapter(t *testing.T) {
	s, uri := newAdapterServer(t)
	a := &documentAdapter{server: s}

	doc := a.GetDocument(uri)
	if doc == nil {
		t.Fatal("GetDocument returned nil for open document")
	}
	if doc.URI != uri || doc.Content != indexCoverageSrc || doc.AST == nil {
		t.Errorf("converted document = %+v", doc)
	}

	if a.GetDocument("file:///nope.gsx") != nil {
		t.Error("GetDocument for unopened URI should be nil")
	}

	all := a.AllDocuments()
	if len(all) != 1 || all[0].URI != uri {
		t.Errorf("AllDocuments = %+v, want one entry for %s", all, uri)
	}
}

func TestWorkspaceASTAdapter(t *testing.T) {
	s, _ := newAdapterServer(t)
	a := &workspaceASTAdapter{server: s}

	wsURI := "file:///workspace-only.gsx"
	wsDoc := parseTestDoc(routerCoverageSrc)
	s.workspaceASTsMu.Lock()
	s.workspaceASTs[wsURI] = wsDoc.AST
	s.workspaceASTsMu.Unlock()

	if got := a.GetWorkspaceAST(wsURI); got != wsDoc.AST {
		t.Error("GetWorkspaceAST did not return the cached AST")
	}
	if a.GetWorkspaceAST("file:///nope.gsx") != nil {
		t.Error("GetWorkspaceAST for unknown URI should be nil")
	}

	all := a.AllWorkspaceASTs()
	if len(all) != 1 || all[wsURI] != wsDoc.AST {
		t.Errorf("AllWorkspaceASTs = %v, want copy containing %s", all, wsURI)
	}
	// The returned map is a copy; mutating it must not affect the server.
	delete(all, wsURI)
	if a.GetWorkspaceAST(wsURI) == nil {
		t.Error("mutating the returned map changed server state")
	}
}

func TestVirtualFileAdapter(t *testing.T) {
	s, uri := newAdapterServer(t)
	a := &virtualFileAdapter{server: s}

	if a.GetVirtualFile(uri) != nil {
		t.Error("expected nil virtual file before caching")
	}
	s.virtualFiles.Put(uri, gopls.TuiURIToGoURI(uri), "package main", nil, 1)
	cached := a.GetVirtualFile(uri)
	if cached == nil {
		t.Fatal("expected cached virtual file")
	}
	if cached.Content != "package main" {
		t.Errorf("Content = %q", cached.Content)
	}
}

func TestFunctionNameCheckerAdapter(t *testing.T) {
	type tc struct {
		name string
		want bool
	}

	tests := map[string]tc{
		"indexed function": {name: "format", want: true},
		"go builtin":       {name: "len", want: true},
		"unknown":          {name: "mystery", want: false},
	}

	s, _ := newAdapterServer(t)
	a := &functionNameCheckerAdapter{server: s}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := a.IsFunctionName(tt.name); got != tt.want {
				t.Errorf("IsFunctionName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// stubSemanticProvider lets the adapter tests control the inner provider result.
type stubSemanticProvider struct {
	result *provider.SemanticTokens
	err    error
}

func (s *stubSemanticProvider) SemanticTokensFull(*provider.Document) (*provider.SemanticTokens, error) {
	return s.result, s.err
}

func TestSemanticTokensProviderAdapter(t *testing.T) {
	doc := &Document{URI: "file:///s.gsx", Content: "package main"}

	t.Run("nil result becomes empty tokens", func(t *testing.T) {
		a := newSemanticTokensProviderAdapter(&stubSemanticProvider{result: nil})
		got, err := a.SemanticTokensFull(doc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || got.Data == nil || len(got.Data) != 0 {
			t.Errorf("got %+v, want empty token data", got)
		}
	})

	t.Run("error is propagated", func(t *testing.T) {
		a := newSemanticTokensProviderAdapter(&stubSemanticProvider{err: errors.New("sem boom")})
		got, err := a.SemanticTokensFull(doc)
		if err == nil || !strings.Contains(err.Error(), "sem boom") {
			t.Fatalf("error = %v, want sem boom", err)
		}
		if got != nil {
			t.Errorf("result = %+v, want nil on error", got)
		}
	})
}
