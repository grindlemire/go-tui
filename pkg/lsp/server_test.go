package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
)

// mockReadWriter provides a mock for testing LSP communication.
type mockReadWriter struct {
	input  *bytes.Buffer
	output *bytes.Buffer
}

func newMockReadWriter() *mockReadWriter {
	return &mockReadWriter{
		input:  new(bytes.Buffer),
		output: new(bytes.Buffer),
	}
}

func (m *mockReadWriter) Read(p []byte) (n int, err error) {
	return m.input.Read(p)
}

func (m *mockReadWriter) Write(p []byte) (n int, err error) {
	return m.output.Write(p)
}

// writeRequest writes a JSON-RPC request to the mock input.
func (m *mockReadWriter) writeRequest(id any, method string, params any) error {
	req := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if id != nil {
		req["id"] = id
	}
	if params != nil {
		req["params"] = params
	}

	content, err := json.Marshal(req)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))
	m.input.WriteString(header)
	m.input.Write(content)
	return nil
}

// readResponse reads a JSON-RPC response from the mock output.
func (m *mockReadWriter) readResponse() (*Response, error) {
	// Read header
	var contentLength int
	for {
		line, err := m.output.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			lenStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			if _, err := fmt.Sscanf(lenStr, "%d", &contentLength); err != nil {
				return nil, err
			}
		}
	}

	if contentLength == 0 {
		return nil, io.EOF
	}

	content := make([]byte, contentLength)
	if _, err := io.ReadFull(m.output, content); err != nil {
		return nil, err
	}

	var resp Response
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func TestServerInitialize(t *testing.T) {
	type tc struct {
		params    InitializeParams
		wantError bool
	}

	tests := map[string]tc{
		"basic initialize": {
			params: InitializeParams{
				RootURI: "file:///test",
			},
			wantError: false,
		},
		"with capabilities": {
			params: InitializeParams{
				RootURI: "file:///test",
				Capabilities: ClientCapabilities{
					TextDocument: TextDocumentClientCapabilities{
						Synchronization: &SynchronizationCapabilities{
							DidSave: true,
						},
					},
				},
			},
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mock := newMockReadWriter()

			// Write initialize request
			if err := mock.writeRequest(1, "initialize", tt.params); err != nil {
				t.Fatalf("writeRequest: %v", err)
			}

			// Write shutdown request to end server loop
			if err := mock.writeRequest(2, "shutdown", nil); err != nil {
				t.Fatalf("writeRequest: %v", err)
			}

			// Write exit notification
			if err := mock.writeRequest(nil, "exit", nil); err != nil {
				t.Fatalf("writeRequest: %v", err)
			}

			server := NewServer(mock.input, mock.output)
			if err := server.Run(t.Context()); err != nil {
				t.Fatalf("Run: %v", err)
			}

			// Read initialize response
			resp, err := mock.readResponse()
			if err != nil {
				t.Fatalf("readResponse: %v", err)
			}

			if tt.wantError {
				if resp.Error == nil {
					t.Error("expected error response, got success")
				}
				return
			}

			if resp.Error != nil {
				t.Errorf("unexpected error: %v", resp.Error)
				return
			}

			// Check result has capabilities
			resultMap, ok := resp.Result.(map[string]any)
			if !ok {
				t.Fatalf("expected map result, got %T", resp.Result)
			}

			caps, ok := resultMap["capabilities"].(map[string]any)
			if !ok {
				t.Fatal("missing capabilities in result")
			}

			// Verify text document sync is enabled
			if _, ok := caps["textDocumentSync"]; !ok {
				t.Error("missing textDocumentSync capability")
			}
		})
	}
}

func TestDocumentLifecycle(t *testing.T) {
	type tc struct {
		content    string
		wantErrors int
	}

	tests := map[string]tc{
		"valid document": {
			content: `package main

func Hello() Element {
	<span>Hello</span>
}
`,
			wantErrors: 0,
		},
		"document with parse error": {
			content: `package main

@component Hello( {
	<span>Hello</span>
}
`,
			wantErrors: 2, // Parser generates multiple errors for this malformed input
		},
		"missing package": {
			content: `func Hello() Element {
	<span>Hello</span>
}
`,
			wantErrors: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mock := newMockReadWriter()
			uri := "file:///test.gsx"

			// Initialize
			if err := mock.writeRequest(1, "initialize", InitializeParams{RootURI: "file:///"}); err != nil {
				t.Fatal(err)
			}

			// Open document
			if err := mock.writeRequest(nil, "textDocument/didOpen", DidOpenParams{
				TextDocument: TextDocumentItem{
					URI:        uri,
					LanguageID: "tui",
					Version:    1,
					Text:       tt.content,
				},
			}); err != nil {
				t.Fatal(err)
			}

			// Shutdown
			if err := mock.writeRequest(2, "shutdown", nil); err != nil {
				t.Fatal(err)
			}

			// Exit
			if err := mock.writeRequest(nil, "exit", nil); err != nil {
				t.Fatal(err)
			}

			server := NewServer(mock.input, mock.output)
			if err := server.Run(t.Context()); err != nil {
				t.Fatalf("Run: %v", err)
			}

			// Check that document was opened
			doc := server.docs.Get(uri)
			if doc == nil {
				t.Fatal("document not found after open")
			}

			if len(doc.Errors) != tt.wantErrors {
				t.Errorf("got %d errors, want %d", len(doc.Errors), tt.wantErrors)
				for _, err := range doc.Errors {
					t.Logf("  error: %v", err)
				}
			}
		})
	}
}

func TestDocumentUpdate(t *testing.T) {
	type tc struct {
		initial       string
		updated       string
		wantInitErrs  int
		wantFinalErrs int
	}

	tests := map[string]tc{
		"fix error": {
			initial: `package main

@component Hello( {
	<span>Hello</span>
}
`,
			updated: `package main

func Hello() Element {
	<span>Hello</span>
}
`,
			wantInitErrs:  2, // Parser generates multiple errors for this malformed input
			wantFinalErrs: 0,
		},
		"introduce error": {
			initial: `package main

func Hello() Element {
	<span>Hello</span>
}
`,
			updated: `package main

@component Hello( {
	<span>Hello</span>
}
`,
			wantInitErrs:  0,
			wantFinalErrs: 2, // Parser generates multiple errors for this malformed input
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mock := newMockReadWriter()
			uri := "file:///test.gsx"

			// Initialize
			if err := mock.writeRequest(1, "initialize", InitializeParams{RootURI: "file:///"}); err != nil {
				t.Fatal(err)
			}

			// Open document
			if err := mock.writeRequest(nil, "textDocument/didOpen", DidOpenParams{
				TextDocument: TextDocumentItem{
					URI:        uri,
					LanguageID: "tui",
					Version:    1,
					Text:       tt.initial,
				},
			}); err != nil {
				t.Fatal(err)
			}

			// Update document
			if err := mock.writeRequest(nil, "textDocument/didChange", DidChangeParams{
				TextDocument: VersionedTextDocumentIdentifier{
					URI:     uri,
					Version: 2,
				},
				ContentChanges: []TextDocumentContentChangeEvent{
					{Text: tt.updated},
				},
			}); err != nil {
				t.Fatal(err)
			}

			// Shutdown
			if err := mock.writeRequest(2, "shutdown", nil); err != nil {
				t.Fatal(err)
			}

			// Exit
			if err := mock.writeRequest(nil, "exit", nil); err != nil {
				t.Fatal(err)
			}

			server := NewServer(mock.input, mock.output)
			if err := server.Run(t.Context()); err != nil {
				t.Fatalf("Run: %v", err)
			}

			// Check final state
			doc := server.docs.Get(uri)
			if doc == nil {
				t.Fatal("document not found")
			}

			if len(doc.Errors) != tt.wantFinalErrs {
				t.Errorf("got %d errors, want %d", len(doc.Errors), tt.wantFinalErrs)
			}

			if doc.Version != 2 {
				t.Errorf("version = %d, want 2", doc.Version)
			}
		})
	}
}

func TestDocumentClose(t *testing.T) {
	mock := newMockReadWriter()
	uri := "file:///test.gsx"
	content := `package main

func Hello() Element {
	<span>Hello</span>
}
`

	// Initialize
	if err := mock.writeRequest(1, "initialize", InitializeParams{RootURI: "file:///"}); err != nil {
		t.Fatal(err)
	}

	// Open document
	if err := mock.writeRequest(nil, "textDocument/didOpen", DidOpenParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: "tui",
			Version:    1,
			Text:       content,
		},
	}); err != nil {
		t.Fatal(err)
	}

	// Close document
	if err := mock.writeRequest(nil, "textDocument/didClose", DidCloseParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	}); err != nil {
		t.Fatal(err)
	}

	// Shutdown
	if err := mock.writeRequest(2, "shutdown", nil); err != nil {
		t.Fatal(err)
	}

	// Exit
	if err := mock.writeRequest(nil, "exit", nil); err != nil {
		t.Fatal(err)
	}

	server := NewServer(mock.input, mock.output)
	if err := server.Run(t.Context()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Document should be gone
	if doc := server.docs.Get(uri); doc != nil {
		t.Error("document still exists after close")
	}
}

func TestDocumentManager(t *testing.T) {
	type tc struct {
		operations []func(dm *DocumentManager)
		wantDocs   int
	}

	tests := map[string]tc{
		"open single": {
			operations: []func(dm *DocumentManager){
				func(dm *DocumentManager) {
					dm.Open("file:///a.gsx", "package main", 1)
				},
			},
			wantDocs: 1,
		},
		"open multiple": {
			operations: []func(dm *DocumentManager){
				func(dm *DocumentManager) {
					dm.Open("file:///a.gsx", "package main", 1)
				},
				func(dm *DocumentManager) {
					dm.Open("file:///b.gsx", "package main", 1)
				},
			},
			wantDocs: 2,
		},
		"open and close": {
			operations: []func(dm *DocumentManager){
				func(dm *DocumentManager) {
					dm.Open("file:///a.gsx", "package main", 1)
				},
				func(dm *DocumentManager) {
					dm.Close("file:///a.gsx")
				},
			},
			wantDocs: 0,
		},
		"update": {
			operations: []func(dm *DocumentManager){
				func(dm *DocumentManager) {
					dm.Open("file:///a.gsx", "package main", 1)
				},
				func(dm *DocumentManager) {
					dm.Update("file:///a.gsx", "package updated", 2)
				},
			},
			wantDocs: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dm := NewDocumentManager()

			for _, op := range tt.operations {
				op(dm)
			}

			docs := dm.All()
			if len(docs) != tt.wantDocs {
				t.Errorf("got %d documents, want %d", len(docs), tt.wantDocs)
			}
		})
	}
}

func TestPositionConversion(t *testing.T) {
	type tc struct {
		content  string
		pos      Position
		wantOff  int
		wantBack Position
	}

	tests := map[string]tc{
		"start of file": {
			content:  "hello\nworld",
			pos:      Position{Line: 0, Character: 0},
			wantOff:  0,
			wantBack: Position{Line: 0, Character: 0},
		},
		"middle of first line": {
			content:  "hello\nworld",
			pos:      Position{Line: 0, Character: 3},
			wantOff:  3,
			wantBack: Position{Line: 0, Character: 3},
		},
		"start of second line": {
			content:  "hello\nworld",
			pos:      Position{Line: 1, Character: 0},
			wantOff:  6,
			wantBack: Position{Line: 1, Character: 0},
		},
		"middle of second line": {
			content:  "hello\nworld",
			pos:      Position{Line: 1, Character: 2},
			wantOff:  8,
			wantBack: Position{Line: 1, Character: 2},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			offset := PositionToOffset(tt.content, tt.pos)
			if offset != tt.wantOff {
				t.Errorf("PositionToOffset = %d, want %d", offset, tt.wantOff)
			}

			back := OffsetToPosition(tt.content, offset)
			if back != tt.wantBack {
				t.Errorf("OffsetToPosition = %+v, want %+v", back, tt.wantBack)
			}
		})
	}
}
