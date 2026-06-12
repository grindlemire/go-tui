package lsp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/lsp/gopls"
)

// failAfterWriter fails every Write call after the first n successful ones.
type failAfterWriter struct {
	n int
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.n <= 0 {
		return 0, errors.New("write failed")
	}
	w.n--
	return len(p), nil
}

func TestServerRun_ContextCancelled(t *testing.T) {
	s := NewServer(strings.NewReader(""), new(bytes.Buffer))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Run = %v, want context.Canceled", err)
	}
}

func TestServerRun_ReadErrors(t *testing.T) {
	type tc struct {
		input   string
		wantSub string
	}

	tests := map[string]tc{
		"invalid content length": {
			input:   "Content-Length: abc\r\n\r\n",
			wantSub: "invalid Content-Length",
		},
		"missing content length": {
			input:   "Foo: bar\r\n\r\n",
			wantSub: "missing Content-Length",
		},
		"short content": {
			input:   "Content-Length: 100\r\n\r\n{}",
			wantSub: "reading content",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := NewServer(strings.NewReader(tt.input), new(bytes.Buffer))
			err := s.Run(context.Background())
			if err == nil || !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("Run = %v, want error containing %q", err, tt.wantSub)
			}
		})
	}
}

func TestServerRun_WriteErrors(t *testing.T) {
	type tc struct {
		successfulWrites int
	}

	tests := map[string]tc{
		"header write fails": {successfulWrites: 0},
		"body write fails":   {successfulWrites: 1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mock := newMockReadWriter()
			if err := mock.writeRequest(1, "initialize", InitializeParams{RootURI: "file:///x"}); err != nil {
				t.Fatal(err)
			}
			s := NewServer(mock.input, &failAfterWriter{n: tt.successfulWrites})
			err := s.Run(context.Background())
			if err == nil || !strings.Contains(err.Error(), "writing response") {
				t.Errorf("Run = %v, want writing response error", err)
			}
		})
	}
}

func TestServerRun_ParseErrorResponse(t *testing.T) {
	mock := newMockReadWriter()
	body := "this is not json"
	fmt.Fprintf(mock.input, "Content-Length: %d\r\n\r\n%s", len(body), body)

	s := NewServer(mock.input, mock.output)
	if err := s.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	resp, err := mock.readResponse()
	if err != nil {
		t.Fatalf("readResponse: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != CodeParseError {
		t.Errorf("error = %+v, want code %d", resp.Error, CodeParseError)
	}
}

func TestServerRun_MethodNotFoundResponse(t *testing.T) {
	mock := newMockReadWriter()
	if err := mock.writeRequest(7, "no/such/method", nil); err != nil {
		t.Fatal(err)
	}

	s := NewServer(mock.input, mock.output)
	if err := s.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	resp, err := mock.readResponse()
	if err != nil {
		t.Fatalf("readResponse: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != CodeMethodNotFound {
		t.Errorf("error = %+v, want code %d", resp.Error, CodeMethodNotFound)
	}
	if id, ok := resp.ID.(float64); !ok || id != 7 {
		t.Errorf("ID = %v, want 7", resp.ID)
	}
}

func TestSetLogFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lsp.log")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s := NewServer(strings.NewReader(""), new(bytes.Buffer))
	s.SetLogFile(f)
	defer s.SetLogFile(nil)

	// Trigger a server log line.
	s.router.Route(Request{JSONRPC: "2.0", Method: "definitely/unknown"})
	s.SetLogFile(nil)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[server]") {
		t.Errorf("log file %q missing server-prefixed entries", string(data))
	}
}

func TestLookupSourceMap(t *testing.T) {
	s := NewServer(strings.NewReader(""), new(bytes.Buffer))
	uri := "file:///map.gsx"

	// No virtual file cached.
	if fn := s.lookupSourceMap(uri); fn != nil {
		t.Error("expected nil lookup without a cached virtual file")
	}

	// Cached virtual file without a source map.
	s.virtualFiles.Put(uri, gopls.TuiURIToGoURI(uri), "package main", nil, 1)
	if fn := s.lookupSourceMap(uri); fn != nil {
		t.Error("expected nil lookup for cached file with nil source map")
	}

	// Cached virtual file with a real source map.
	doc := parseTestDoc(routerCoverageSrc)
	goContent, sourceMap := gopls.GenerateVirtualGo(doc.AST)
	s.virtualFiles.Put(uri, gopls.TuiURIToGoURI(uri), goContent, sourceMap, 2)
	fn := s.lookupSourceMap(uri)
	if fn == nil {
		t.Fatal("expected a translation function for cached source map")
	}
}

func TestHandleGoplsDiagnostics(t *testing.T) {
	out := new(bytes.Buffer)
	s := NewServer(strings.NewReader(""), out)
	uri := "file:///diag.gsx"
	s.docs.Open(uri, routerCoverageSrc, 1)

	diags := []gopls.GoplsDiagnostic{
		{
			Range: gopls.Range{
				Start: gopls.Position{Line: 4, Character: 2},
				End:   gopls.Position{Line: 4, Character: 7},
			},
			Severity: 1,
			Source:   "compiler",
			Message:  "undefined: title",
		},
	}

	s.handleGoplsDiagnostics(uri, diags)

	s.goplsDiagnosticsMu.RLock()
	stored := s.goplsDiagnostics[uri]
	s.goplsDiagnosticsMu.RUnlock()
	if len(stored) != 1 || stored[0].Message != "undefined: title" {
		t.Errorf("stored diagnostics = %+v", stored)
	}

	output := out.String()
	if !strings.Contains(output, "textDocument/publishDiagnostics") {
		t.Error("no publishDiagnostics notification was written")
	}
	if !strings.Contains(output, "undefined: title") {
		t.Error("gopls diagnostic message missing from published diagnostics")
	}

	// Diagnostics for an unopened document are stored but not published.
	out.Reset()
	s.handleGoplsDiagnostics("file:///ghost.gsx", diags)
	if out.Len() != 0 {
		t.Errorf("unexpected notification for unopened document: %s", out.String())
	}
}
