package gopls

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// fakeGoplsEnvVar marks a re-execution of the test binary as the fake gopls
// subprocess. NewGoplsProxy locates "gopls" via PATH; tests symlink the test
// binary under that name and set this variable so TestMain serves JSON-RPC
// over stdio instead of running the test suite.
const fakeGoplsEnvVar = "TUI_TEST_FAKE_GOPLS"

func TestMain(m *testing.M) {
	if os.Getenv(fakeGoplsEnvVar) == "1" {
		fakeGoplsServe(os.Stdin, os.Stdout)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// readFramed reads one Content-Length framed JSON-RPC message.
func readFramed(r *bufio.Reader) ([]byte, error) {
	contentLength := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if after, ok := strings.CutPrefix(line, "Content-Length:"); ok {
			if _, err := fmt.Sscanf(strings.TrimSpace(after), "%d", &contentLength); err != nil {
				return nil, err
			}
		}
	}
	buf := make([]byte, contentLength)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// writeFramed writes one Content-Length framed JSON-RPC message.
func writeFramed(w io.Writer, payload string) {
	fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(payload), payload)
}

// fakeGoplsServe speaks just enough LSP over stdio to exercise the proxy.
// Behavior is keyed off the request method and the document URI so tests can
// select response shapes (object vs array vs null vs malformed).
func fakeGoplsServe(in io.Reader, out io.Writer) {
	r := bufio.NewReader(in)

	// Log of document lifecycle notifications, queryable via "test/seen".
	var seen []string

	type docParams struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		Position Position `json:"position"`
	}

	for {
		msg, err := readFramed(r)
		if err != nil {
			return
		}

		var req struct {
			ID     int64           `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(msg, &req); err != nil {
			continue
		}

		reply := func(result string) {
			writeFramed(out, fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":%s}`, req.ID, result))
		}

		var doc docParams
		_ = json.Unmarshal(req.Params, &doc)
		uri := doc.TextDocument.URI

		switch req.Method {
		case "initialize":
			reply(`{"capabilities":{"hoverProvider":true}}`)
		case "initialized":
			// Notification, no response.
		case "shutdown":
			reply("null")
		case "exit":
			return
		case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose":
			seen = append(seen, req.Method+" "+string(req.Params))
		case "test/seen":
			data, _ := json.Marshal(seen)
			reply(string(data))
		case "test/error":
			writeFramed(out, fmt.Sprintf(
				`{"jsonrpc":"2.0","id":%d,"error":{"code":-32601,"message":"method not found"}}`, req.ID))
		case "test/publishDiagnostics":
			// Echo the params back as a server-initiated diagnostics notification.
			writeFramed(out, fmt.Sprintf(
				`{"jsonrpc":"2.0","method":"textDocument/publishDiagnostics","params":%s}`, string(req.Params)))
		case "test/sendBare":
			// Response with no id and no method: readResponses must skip it.
			writeFramed(out, `{"jsonrpc":"2.0","result":{}}`)
		case "test/sendGarbage":
			// Unparseable payload: readResponses must log and continue.
			writeFramed(out, `{this is not json`)
		case "textDocument/hover":
			switch {
			case strings.Contains(uri, "nullhover"):
				reply("null")
			case strings.Contains(uri, "badhover"):
				reply(`[1,2]`)
			default:
				reply(fmt.Sprintf(`{"contents":{"kind":"markdown","value":"hover:%s:%d:%d"}}`,
					uri, doc.Position.Line, doc.Position.Character))
			}
		case "textDocument/definition":
			switch {
			case strings.Contains(uri, "nulldef"):
				reply("null")
			case strings.Contains(uri, "singledef"):
				reply(`{"uri":"file:///single.go","range":{"start":{"line":1,"character":2},"end":{"line":1,"character":7}}}`)
			case strings.Contains(uri, "baddef"):
				reply(`5`)
			default:
				reply(`[{"uri":"file:///def.go","range":{"start":{"line":3,"character":4},"end":{"line":3,"character":9}}}]`)
			}
		case "textDocument/completion":
			switch {
			case strings.Contains(uri, "arraycomp"):
				reply(`[{"label":"ArrayItem","kind":6}]`)
			case strings.Contains(uri, "badcomp"):
				reply(`"oops"`)
			default:
				reply(`{"isIncomplete":false,"items":[{"label":"Println","kind":3,"detail":"func(a ...any) (n int, err error)"}]}`)
			}
		default:
			if req.ID != 0 {
				reply("null")
			}
		}
	}
}

// startFakeGopls installs the test binary as "gopls" into a temp dir,
// prepends that dir to PATH, and starts a proxy against it. It symlinks
// where possible and falls back to copying the binary, since symlink
// creation on Windows requires elevated privileges. The executable name
// needs the .exe suffix on Windows for exec.LookPath to resolve it.
func startFakeGopls(t *testing.T) *GoplsProxy {
	t.Helper()

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	name := "gopls"
	if runtime.GOOS == "windows" {
		name = "gopls.exe"
	}
	dir := t.TempDir()
	target := filepath.Join(dir, name)
	if err := os.Symlink(exe, target); err != nil {
		data, readErr := os.ReadFile(exe)
		if readErr != nil {
			t.Fatalf("reading test binary for fake gopls: %v", readErr)
		}
		if writeErr := os.WriteFile(target, data, 0o755); writeErr != nil {
			t.Fatalf("copying fake gopls: %v", writeErr)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv(fakeGoplsEnvVar, "1")

	p, err := NewGoplsProxy(context.Background())
	if err != nil {
		t.Fatalf("NewGoplsProxy: %v", err)
	}
	return p
}

// shutdownProxy shuts the proxy down. A fake gopls that received the exit
// notification exits cleanly, so Shutdown must return nil; anything else is
// a real failure (historically Shutdown canceled the command context before
// cmd.Wait, racing a kill against the clean exit on every platform).
func shutdownProxy(t *testing.T, p *GoplsProxy) {
	t.Helper()
	if err := p.Shutdown(); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}

func TestGoplsProxyLifecycle(t *testing.T) {
	p := startFakeGopls(t)

	if err := p.Initialize("file:///workspace"); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if p.rootURI != "file:///workspace" {
		t.Errorf("rootURI = %q, want %q", p.rootURI, "file:///workspace")
	}

	// Hover round trip proves request marshaling and response routing.
	hover, err := p.Hover("file:///main.go", Position{Line: 7, Character: 12})
	if err != nil {
		t.Fatalf("Hover: %v", err)
	}
	if hover == nil {
		t.Fatal("Hover returned nil, want hover content")
	}
	want := "hover:file:///main.go:7:12"
	if hover.Contents.Value != want {
		t.Errorf("hover value = %q, want %q", hover.Contents.Value, want)
	}
	if hover.Contents.Kind != "markdown" {
		t.Errorf("hover kind = %q, want markdown", hover.Contents.Kind)
	}

	shutdownProxy(t, p)
}

func TestGoplsProxyRequests(t *testing.T) {
	p := startFakeGopls(t)
	defer shutdownProxy(t, p)

	if err := p.Initialize("file:///ws"); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	t.Run("virtual file lifecycle notifications", func(t *testing.T) {
		if err := p.OpenVirtualFile("file:///ws/a_gsx_generated.go", "package a", 1); err != nil {
			t.Fatalf("OpenVirtualFile: %v", err)
		}
		if vf := p.virtualFiles["file:///ws/a_gsx_generated.go"]; vf == nil || vf.Content != "package a" || vf.Version != 1 {
			t.Errorf("virtual file not tracked after open: %+v", vf)
		}

		if err := p.UpdateVirtualFile("file:///ws/a_gsx_generated.go", "package a // v2", 2); err != nil {
			t.Fatalf("UpdateVirtualFile: %v", err)
		}
		if vf := p.virtualFiles["file:///ws/a_gsx_generated.go"]; vf == nil || vf.Content != "package a // v2" || vf.Version != 2 {
			t.Errorf("virtual file not updated: %+v", vf)
		}

		// Updating an untracked URI still notifies but does not create an entry.
		if err := p.UpdateVirtualFile("file:///ws/untracked.go", "x", 1); err != nil {
			t.Fatalf("UpdateVirtualFile untracked: %v", err)
		}
		if _, ok := p.virtualFiles["file:///ws/untracked.go"]; ok {
			t.Error("update created an entry for an untracked file")
		}

		if err := p.CloseVirtualFile("file:///ws/a_gsx_generated.go"); err != nil {
			t.Fatalf("CloseVirtualFile: %v", err)
		}
		if _, ok := p.virtualFiles["file:///ws/a_gsx_generated.go"]; ok {
			t.Error("virtual file still tracked after close")
		}

		// Ask the fake server what it received and assert the payloads.
		result, err := p.call("test/seen", nil)
		if err != nil {
			t.Fatalf("test/seen: %v", err)
		}
		var seen []string
		if err := json.Unmarshal(result, &seen); err != nil {
			t.Fatalf("parsing seen log: %v", err)
		}
		if len(seen) != 4 {
			t.Fatalf("seen %d notifications, want 4: %v", len(seen), seen)
		}
		checks := []struct{ method, substr string }{
			{"textDocument/didOpen", `"text":"package a"`},
			{"textDocument/didChange", `"text":"package a // v2"`},
			{"textDocument/didChange", `"uri":"file:///ws/untracked.go"`},
			{"textDocument/didClose", `"uri":"file:///ws/a_gsx_generated.go"`},
		}
		for i, c := range checks {
			if !strings.HasPrefix(seen[i], c.method) {
				t.Errorf("seen[%d] = %q, want method %q", i, seen[i], c.method)
			}
			if !strings.Contains(seen[i], c.substr) {
				t.Errorf("seen[%d] = %q, want substring %q", i, seen[i], c.substr)
			}
		}
		if !strings.Contains(seen[0], `"version":1`) || !strings.Contains(seen[1], `"version":2`) {
			t.Errorf("versions not propagated in notifications: %v", seen[:2])
		}
	})

	t.Run("completion list form", func(t *testing.T) {
		items, err := p.Completion("file:///ws/x.go", Position{Line: 1, Character: 1})
		if err != nil {
			t.Fatalf("Completion: %v", err)
		}
		if len(items) != 1 || items[0].Label != "Println" || items[0].Kind != 3 {
			t.Errorf("items = %+v, want one Println item with kind 3", items)
		}
		if items[0].Detail != "func(a ...any) (n int, err error)" {
			t.Errorf("detail = %q", items[0].Detail)
		}
	})

	t.Run("completion array form", func(t *testing.T) {
		items, err := p.Completion("file:///ws/arraycomp.go", Position{})
		if err != nil {
			t.Fatalf("Completion: %v", err)
		}
		if len(items) != 1 || items[0].Label != "ArrayItem" || items[0].Kind != 6 {
			t.Errorf("items = %+v, want one ArrayItem with kind 6", items)
		}
	})

	t.Run("completion malformed result", func(t *testing.T) {
		_, err := p.Completion("file:///ws/badcomp.go", Position{})
		if err == nil || !strings.Contains(err.Error(), "parsing completion result") {
			t.Errorf("err = %v, want parsing completion result error", err)
		}
	})

	t.Run("hover null result", func(t *testing.T) {
		hover, err := p.Hover("file:///ws/nullhover.go", Position{})
		if err != nil {
			t.Fatalf("Hover: %v", err)
		}
		if hover != nil {
			t.Errorf("hover = %+v, want nil", hover)
		}
	})

	t.Run("hover malformed result", func(t *testing.T) {
		_, err := p.Hover("file:///ws/badhover.go", Position{})
		if err == nil || !strings.Contains(err.Error(), "parsing hover result") {
			t.Errorf("err = %v, want parsing hover result error", err)
		}
	})

	t.Run("definition array form", func(t *testing.T) {
		locs, err := p.Definition("file:///ws/x.go", Position{Line: 3, Character: 5})
		if err != nil {
			t.Fatalf("Definition: %v", err)
		}
		want := []Location{{
			URI: "file:///def.go",
			Range: Range{
				Start: Position{Line: 3, Character: 4},
				End:   Position{Line: 3, Character: 9},
			},
		}}
		if len(locs) != 1 || locs[0] != want[0] {
			t.Errorf("locs = %+v, want %+v", locs, want)
		}
	})

	t.Run("definition single object form", func(t *testing.T) {
		locs, err := p.Definition("file:///ws/singledef.go", Position{})
		if err != nil {
			t.Fatalf("Definition: %v", err)
		}
		if len(locs) != 1 || locs[0].URI != "file:///single.go" || locs[0].Range.Start.Character != 2 {
			t.Errorf("locs = %+v, want single location for file:///single.go", locs)
		}
	})

	t.Run("definition null result", func(t *testing.T) {
		locs, err := p.Definition("file:///ws/nulldef.go", Position{})
		if err != nil {
			t.Fatalf("Definition: %v", err)
		}
		if locs != nil {
			t.Errorf("locs = %+v, want nil", locs)
		}
	})

	t.Run("definition malformed result", func(t *testing.T) {
		_, err := p.Definition("file:///ws/baddef.go", Position{})
		if err == nil || !strings.Contains(err.Error(), "parsing definition result") {
			t.Errorf("err = %v, want parsing definition result error", err)
		}
	})

	t.Run("error response", func(t *testing.T) {
		_, err := p.call("test/error", nil)
		if err == nil || !strings.Contains(err.Error(), "gopls error -32601: method not found") {
			t.Errorf("err = %v, want gopls error -32601", err)
		}
	})

	t.Run("reader skips bare and garbage messages", func(t *testing.T) {
		// The fake emits a response with no id, then an unparseable payload.
		// readResponses must skip both; the next request proves the read loop
		// is still alive and ordering is preserved.
		if err := p.notify("test/sendBare", nil); err != nil {
			t.Fatalf("notify sendBare: %v", err)
		}
		if err := p.notify("test/sendGarbage", nil); err != nil {
			t.Fatalf("notify sendGarbage: %v", err)
		}
		hover, err := p.Hover("file:///ws/alive.go", Position{Line: 1, Character: 2})
		if err != nil {
			t.Fatalf("Hover after junk messages: %v", err)
		}
		if hover == nil || hover.Contents.Value != "hover:file:///ws/alive.go:1:2" {
			t.Errorf("hover = %+v, reader did not survive junk messages", hover)
		}
	})

	t.Run("diagnostics round trip with translation", func(t *testing.T) {
		// Lookup translates go positions to gsx positions with a fixed shift,
		// except go line 50 which is unmapped.
		p.SetSourceMapLookup(func(gsxURI string) func(int, int) (int, int, bool) {
			if gsxURI != "file:///ws/counter.gsx" {
				return nil
			}
			return func(goLine, goCol int) (int, int, bool) {
				if goLine == 50 {
					return 0, 0, false
				}
				return goLine + 100, goCol + 200, true
			}
		})

		got := make(chan struct {
			uri   string
			diags []GoplsDiagnostic
		}, 1)
		p.SetDiagnosticCallback(func(uri string, diags []GoplsDiagnostic) {
			got <- struct {
				uri   string
				diags []GoplsDiagnostic
			}{uri, diags}
		})

		diagRange := func(line, ch int) Range {
			return Range{Start: Position{Line: line, Character: ch}, End: Position{Line: line, Character: ch + 3}}
		}

		// First two notifications must be silently skipped (virtual file,
		// non-generated file). The third carries one reportable diagnostic
		// among several that must be filtered; receiving its callback proves
		// the earlier notifications were processed and skipped in order.
		send := func(uri string, diags []goplsDiagnostic) {
			params := publishDiagnosticsParams{URI: uri, Diagnostics: diags}
			if err := p.notify("test/publishDiagnostics", params); err != nil {
				t.Fatalf("notify publishDiagnostics: %v", err)
			}
		}

		send("file:///ws/counter_gsx_generated.go", []goplsDiagnostic{
			{Range: diagRange(5, 1), Severity: 1, Message: "virtual file error"},
		})
		send("file:///ws/regular.go", []goplsDiagnostic{
			{Range: diagRange(5, 1), Severity: 1, Message: "regular file error"},
		})
		send("file:///ws/counter_gsx.go", []goplsDiagnostic{
			{Range: diagRange(5, 1), Severity: 1, Message: "Counter redeclared in this block"},
			{Range: diagRange(6, 1), Severity: 1, Message: "Counter redeclared, see counter_gsx.go"},
			{Range: diagRange(7, 1), Severity: 1, Message: "unknown field Foo in struct literal"},
			{Range: diagRange(51, 1), Severity: 1, Message: "unmapped position error"},
			{Range: diagRange(10, 4), Severity: 2, Message: "undefined: frobnicate"},
		})

		select {
		case res := <-got:
			if res.uri != "file:///ws/counter.gsx" {
				t.Errorf("callback uri = %q, want file:///ws/counter.gsx", res.uri)
			}
			if len(res.diags) != 1 {
				t.Fatalf("callback got %d diagnostics, want 1 (filtering failed): %+v", len(res.diags), res.diags)
			}
			d := res.diags[0]
			if d.Message != "undefined: frobnicate" || d.Severity != 2 || d.Source != "gopls" {
				t.Errorf("diag = %+v, want undefined: frobnicate severity 2 source gopls", d)
			}
			// Go line 10 minus the goimports offset of 1, then +100/+200.
			wantRange := Range{
				Start: Position{Line: 109, Character: 204},
				End:   Position{Line: 109, Character: 207},
			}
			if d.Range != wantRange {
				t.Errorf("diag range = %+v, want %+v", d.Range, wantRange)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for diagnostic callback")
		}
	})
}

func TestNewGoplsProxyNotFound(t *testing.T) {
	// PATH with only an empty directory: LookPath must fail.
	t.Setenv("PATH", t.TempDir())

	_, err := NewGoplsProxy(context.Background())
	if err == nil || !strings.Contains(err.Error(), "gopls not found in PATH") {
		t.Errorf("err = %v, want gopls not found in PATH", err)
	}
}

// nopWriteCloser is a stdin stub that accepts all writes.
type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

// failWriteCloser is a stdin stub that rejects all writes.
type failWriteCloser struct{}

func (failWriteCloser) Write([]byte) (int, error) { return 0, errors.New("stdin closed") }
func (failWriteCloser) Close() error              { return nil }

func TestCallContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &GoplsProxy{
		stdin:   nopWriteCloser{io.Discard},
		pending: make(map[int64]chan *Response),
		ctx:     ctx,
		cancel:  cancel,
	}

	_, err := p.call("textDocument/hover", nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestSendErrors(t *testing.T) {
	t.Run("write failure", func(t *testing.T) {
		p := &GoplsProxy{stdin: failWriteCloser{}}
		err := p.send(Request{JSONRPC: "2.0", Method: "x"})
		if err == nil || !strings.Contains(err.Error(), "stdin closed") {
			t.Errorf("err = %v, want stdin closed", err)
		}
	})

	t.Run("marshal failure", func(t *testing.T) {
		p := &GoplsProxy{stdin: nopWriteCloser{io.Discard}}
		err := p.send(Request{JSONRPC: "2.0", Method: "x", Params: make(chan int)})
		if err == nil {
			t.Error("err = nil, want JSON marshal error for channel param")
		}
	})

	t.Run("call surfaces send failure", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		p := &GoplsProxy{
			stdin:   failWriteCloser{},
			pending: make(map[int64]chan *Response),
			ctx:     ctx,
			cancel:  cancel,
		}
		_, err := p.call("textDocument/hover", nil)
		if err == nil || !strings.Contains(err.Error(), "stdin closed") {
			t.Errorf("err = %v, want stdin closed", err)
		}
	})
}

func TestReadMessageErrors(t *testing.T) {
	type tc struct {
		input   string
		wantErr string
	}

	tests := map[string]tc{
		"invalid content length": {
			input:   "Content-Length: abc\r\n\r\n",
			wantErr: "invalid Content-Length",
		},
		"missing content length": {
			input:   "\r\n",
			wantErr: "missing Content-Length header",
		},
		"truncated content": {
			input:   "Content-Length: 100\r\n\r\nshort",
			wantErr: "reading content",
		},
		"eof on headers": {
			input:   "",
			wantErr: "EOF",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := &GoplsProxy{stdout: bufio.NewReader(strings.NewReader(tt.input))}
			_, err := p.readMessage()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestReadMessageSuccess(t *testing.T) {
	payload := `{"jsonrpc":"2.0","id":1,"result":null}`
	input := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/json\r\n\r\n%s", len(payload), payload)
	p := &GoplsProxy{stdout: bufio.NewReader(strings.NewReader(input))}

	msg, err := p.readMessage()
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}
	if string(msg) != payload {
		t.Errorf("msg = %q, want %q", msg, payload)
	}
}

func TestHandleNotificationSkipBranches(t *testing.T) {
	type tc struct {
		method string
		params string
		lookup SourceMapLookup
	}

	// Every case must complete without invoking the diagnostic callback.
	tests := map[string]tc{
		"non diagnostic method": {
			method: "window/showMessage",
			params: `{}`,
		},
		"malformed params": {
			method: "textDocument/publishDiagnostics",
			params: `{not json`,
		},
		"nil lookup": {
			method: "textDocument/publishDiagnostics",
			params: `{"uri":"file:///a_gsx.go","diagnostics":[{"range":{"start":{"line":1,"character":0},"end":{"line":1,"character":1}},"severity":1,"message":"x"}]}`,
			lookup: nil,
		},
		"lookup returns nil translator": {
			method: "textDocument/publishDiagnostics",
			params: `{"uri":"file:///a_gsx.go","diagnostics":[{"range":{"start":{"line":1,"character":0},"end":{"line":1,"character":1}},"severity":1,"message":"x"}]}`,
			lookup: func(string) func(int, int) (int, int, bool) { return nil },
		},
		"all diagnostics filtered yields no callback": {
			method: "textDocument/publishDiagnostics",
			params: `{"uri":"file:///a_gsx.go","diagnostics":[{"range":{"start":{"line":1,"character":0},"end":{"line":1,"character":1}},"severity":1,"message":"foo redeclared in this block"}]}`,
			lookup: func(string) func(int, int) (int, int, bool) {
				return func(l, c int) (int, int, bool) { return l, c, true }
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := &GoplsProxy{}
			p.SetSourceMapLookup(tt.lookup)
			called := false
			p.SetDiagnosticCallback(func(string, []GoplsDiagnostic) { called = true })

			p.handleNotification(&Notification{
				JSONRPC: "2.0",
				Method:  tt.method,
				Params:  json.RawMessage(tt.params),
			})

			if called {
				t.Error("diagnostic callback invoked, want skipped")
			}
		})
	}
}

func TestURIHelpers(t *testing.T) {
	type tc struct {
		fn   func(string) string
		in   string
		want string
	}

	tests := map[string]tc{
		"TuiURIToGoURI gsx file": {
			fn:   TuiURIToGoURI,
			in:   "file:///ws/counter.gsx",
			want: "file:///ws/counter_gsx_generated.go",
		},
		"TuiURIToGoURI non gsx file": {
			fn:   TuiURIToGoURI,
			in:   "file:///ws/counter.txt",
			want: "file:///ws/counter.txt_generated.go",
		},
		"GoURIToTuiURI virtual file": {
			fn:   GoURIToTuiURI,
			in:   "file:///ws/counter_gsx_generated.go",
			want: "file:///ws/counter.gsx",
		},
		"GoURIToTuiURI passthrough": {
			fn:   GoURIToTuiURI,
			in:   "file:///ws/other.go",
			want: "file:///ws/other.go",
		},
		"GetVirtualFilePath gsx file": {
			fn:   GetVirtualFilePath,
			in:   "/ws/app/counter.gsx",
			want: "/ws/app/counter_gsx_generated.go",
		},
		"GetVirtualFilePath non gsx file": {
			fn:   GetVirtualFilePath,
			in:   "/ws/app/counter.txt",
			want: "/ws/app/counter.txt_generated.go",
		},
		"GeneratedGoURIToTuiURI generated file": {
			fn:   GeneratedGoURIToTuiURI,
			in:   "file:///ws/counter_gsx.go",
			want: "file:///ws/counter.gsx",
		},
		"GeneratedGoURIToTuiURI passthrough": {
			fn:   GeneratedGoURIToTuiURI,
			in:   "file:///ws/other.go",
			want: "file:///ws/other.go",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// GetVirtualFilePath goes through filepath.Join, which emits
			// backslashes on Windows; normalize before comparing.
			if got := filepath.ToSlash(tt.fn(tt.in)); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFileKindPredicates(t *testing.T) {
	type tc struct {
		fn   func(string) bool
		in   string
		want bool
	}

	tests := map[string]tc{
		"IsVirtualGoFile true": {
			fn:   IsVirtualGoFile,
			in:   "file:///ws/counter_gsx_generated.go",
			want: true,
		},
		"IsVirtualGoFile false": {
			fn:   IsVirtualGoFile,
			in:   "file:///ws/counter_gsx.go",
			want: false,
		},
		"IsGeneratedGoFile true": {
			fn:   IsGeneratedGoFile,
			in:   "file:///ws/counter_gsx.go",
			want: true,
		},
		"IsGeneratedGoFile false for virtual": {
			fn:   IsGeneratedGoFile,
			in:   "file:///ws/counter_gsx_generated.go",
			want: false,
		},
		"IsGeneratedGoFile false for plain go": {
			fn:   IsGeneratedGoFile,
			in:   "file:///ws/main.go",
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.fn(tt.in); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
