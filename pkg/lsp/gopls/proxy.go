// Package gopls provides a proxy layer to gopls for Go expression intelligence.
package gopls

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/grindlemire/go-tui/pkg/lsp/log"
)

// GoplsProxy manages communication with a gopls subprocess.
type GoplsProxy struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex // protects writes to stdin

	// Request ID counter
	nextID atomic.Int64

	// Pending requests waiting for responses
	pending   map[int64]chan *Response
	pendingMu sync.Mutex

	// Virtual file management
	virtualFiles   map[string]*VirtualFile
	virtualFilesMu sync.RWMutex

	// Root URI for the workspace
	rootURI string

	// Context for managing shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// VirtualFile represents a generated .go file from a .gsx file.
type VirtualFile struct {
	URI       string
	Content   string
	SourceMap *SourceMap
	Version   int
}

// Request represents a JSON-RPC request.
type Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// Response represents a JSON-RPC response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents a JSON-RPC error.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Position represents a position in a document (0-indexed).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier identifies a text document.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// TextDocumentPositionParams contains position parameters.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionParams represents completion request parameters.
type CompletionParams struct {
	TextDocumentPositionParams
	Context *CompletionContext `json:"context,omitempty"`
}

// CompletionContext provides additional context for completion.
type CompletionContext struct {
	TriggerKind      int    `json:"triggerKind"`
	TriggerCharacter string `json:"triggerCharacter,omitempty"`
}

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label         string         `json:"label"`
	Kind          int            `json:"kind,omitempty"`
	Detail        string         `json:"detail,omitempty"`
	Documentation *MarkupContent `json:"documentation,omitempty"`
	InsertText    string         `json:"insertText,omitempty"`
	FilterText    string         `json:"filterText,omitempty"`
}

// CompletionList represents a list of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// Hover represents hover information.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents markup content.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// NewGoplsProxy creates and starts a new gopls proxy.
func NewGoplsProxy(ctx context.Context) (*GoplsProxy, error) {
	// Find gopls binary
	goplsPath, err := exec.LookPath("gopls")
	if err != nil {
		return nil, fmt.Errorf("gopls not found in PATH: %w", err)
	}

	proxyCtx, cancel := context.WithCancel(ctx)

	cmd := exec.CommandContext(proxyCtx, goplsPath, "-mode=stdio")
	cmd.Env = os.Environ()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("creating stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	// Redirect stderr to discard or to log file
	cmd.Stderr = nil

	p := &GoplsProxy{
		cmd:          cmd,
		stdin:        stdin,
		stdout:       bufio.NewReader(stdout),
		pending:      make(map[int64]chan *Response),
		virtualFiles: make(map[string]*VirtualFile),
		ctx:          proxyCtx,
		cancel:       cancel,
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("starting gopls: %w", err)
	}

	// Start response reader goroutine
	go p.readResponses()

	return p, nil
}

// Initialize initializes the gopls server with the workspace root.
func (p *GoplsProxy) Initialize(rootURI string) error {
	p.rootURI = rootURI

	initParams := map[string]any{
		"processId": os.Getpid(),
		"rootUri":   rootURI,
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"completion": map[string]any{
					"completionItem": map[string]any{
						"snippetSupport": false,
					},
				},
				"hover": map[string]any{
					"contentFormat": []string{"markdown", "plaintext"},
				},
			},
		},
	}

	result, err := p.call("initialize", initParams)
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	log.Gopls("Initialize result: %s", string(result))

	// Send initialized notification
	if err := p.notify("initialized", map[string]any{}); err != nil {
		return fmt.Errorf("initialized notification: %w", err)
	}

	return nil
}

// Shutdown shuts down the gopls server.
func (p *GoplsProxy) Shutdown() error {
	_, err := p.call("shutdown", nil)
	if err != nil {
		return err
	}

	if err := p.notify("exit", nil); err != nil {
		return err
	}

	p.cancel()
	return p.cmd.Wait()
}

// OpenVirtualFile opens a virtual .go file in gopls.
func (p *GoplsProxy) OpenVirtualFile(uri, content string, version int) error {
	p.virtualFilesMu.Lock()
	p.virtualFiles[uri] = &VirtualFile{
		URI:     uri,
		Content: content,
		Version: version,
	}
	p.virtualFilesMu.Unlock()

	params := map[string]any{
		"textDocument": map[string]any{
			"uri":        uri,
			"languageId": "go",
			"version":    version,
			"text":       content,
		},
	}

	return p.notify("textDocument/didOpen", params)
}

// UpdateVirtualFile updates a virtual .go file in gopls.
func (p *GoplsProxy) UpdateVirtualFile(uri, content string, version int) error {
	p.virtualFilesMu.Lock()
	if vf, ok := p.virtualFiles[uri]; ok {
		vf.Content = content
		vf.Version = version
	}
	p.virtualFilesMu.Unlock()

	params := map[string]any{
		"textDocument": map[string]any{
			"uri":     uri,
			"version": version,
		},
		"contentChanges": []map[string]any{
			{"text": content},
		},
	}

	return p.notify("textDocument/didChange", params)
}

// CloseVirtualFile closes a virtual .go file in gopls.
func (p *GoplsProxy) CloseVirtualFile(uri string) error {
	p.virtualFilesMu.Lock()
	delete(p.virtualFiles, uri)
	p.virtualFilesMu.Unlock()

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}

	return p.notify("textDocument/didClose", params)
}

// Completion requests completion items at a position.
func (p *GoplsProxy) Completion(uri string, pos Position) ([]CompletionItem, error) {
	params := CompletionParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     pos,
		},
	}

	result, err := p.call("textDocument/completion", params)
	if err != nil {
		return nil, err
	}

	// gopls returns either CompletionList or []CompletionItem
	var list CompletionList
	if err := json.Unmarshal(result, &list); err != nil {
		// Try as array
		var items []CompletionItem
		if err := json.Unmarshal(result, &items); err != nil {
			return nil, fmt.Errorf("parsing completion result: %w", err)
		}
		return items, nil
	}

	return list.Items, nil
}

// Hover requests hover information at a position.
func (p *GoplsProxy) Hover(uri string, pos Position) (*Hover, error) {
	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
	}

	result, err := p.call("textDocument/hover", params)
	if err != nil {
		return nil, err
	}

	if result == nil || string(result) == "null" {
		return nil, nil
	}

	var hover Hover
	if err := json.Unmarshal(result, &hover); err != nil {
		return nil, fmt.Errorf("parsing hover result: %w", err)
	}

	return &hover, nil
}

// Definition requests the definition location at a position.
func (p *GoplsProxy) Definition(uri string, pos Position) ([]Location, error) {
	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
	}

	result, err := p.call("textDocument/definition", params)
	if err != nil {
		return nil, err
	}

	if result == nil || string(result) == "null" {
		return nil, nil
	}

	// gopls returns either Location or []Location
	var locs []Location
	if err := json.Unmarshal(result, &locs); err != nil {
		var loc Location
		if err := json.Unmarshal(result, &loc); err != nil {
			return nil, fmt.Errorf("parsing definition result: %w", err)
		}
		return []Location{loc}, nil
	}

	return locs, nil
}

// call sends a request and waits for the response.
func (p *GoplsProxy) call(method string, params any) (json.RawMessage, error) {
	id := p.nextID.Add(1)

	// Create response channel
	respChan := make(chan *Response, 1)
	p.pendingMu.Lock()
	p.pending[id] = respChan
	p.pendingMu.Unlock()

	defer func() {
		p.pendingMu.Lock()
		delete(p.pending, id)
		p.pendingMu.Unlock()
	}()

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := p.send(req); err != nil {
		return nil, err
	}

	select {
	case <-p.ctx.Done():
		return nil, p.ctx.Err()
	case resp := <-respChan:
		if resp.Error != nil {
			return nil, fmt.Errorf("gopls error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// notify sends a notification (no response expected).
func (p *GoplsProxy) notify(method string, params any) error {
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	return p.send(req)
}

// send sends a request to gopls.
func (p *GoplsProxy) send(req Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	log.Gopls("Sending: %s", string(data))

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, err := p.stdin.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := p.stdin.Write(data); err != nil {
		return err
	}

	return nil
}

// readResponses reads responses from gopls in a loop.
func (p *GoplsProxy) readResponses() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		msg, err := p.readMessage()
		if err != nil {
			log.Gopls("Error reading message: %v", err)
			return
		}

		log.Gopls("Received: %s", string(msg))

		var resp Response
		if err := json.Unmarshal(msg, &resp); err != nil {
			log.Gopls("Error parsing response: %v", err)
			continue
		}

		// If this is a notification (no ID), ignore it
		if resp.ID == 0 {
			continue
		}

		// Route to waiting caller
		p.pendingMu.Lock()
		if ch, ok := p.pending[resp.ID]; ok {
			ch <- &resp
		}
		p.pendingMu.Unlock()
	}
}

// readMessage reads a single JSON-RPC message from gopls.
func (p *GoplsProxy) readMessage() ([]byte, error) {
	// Read headers
	var contentLength int
	for {
		line, err := p.stdout.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			lenStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, err = strconv.Atoi(lenStr)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %w", err)
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	// Read content
	content := make([]byte, contentLength)
	_, err := io.ReadFull(p.stdout, content)
	if err != nil {
		return nil, fmt.Errorf("reading content: %w", err)
	}

	return content, nil
}

// TuiURIToGoURI converts a .gsx file URI to a virtual .go file URI.
func TuiURIToGoURI(tuiURI string) string {
	// Replace .gsx extension with _gsx_generated.go
	if strings.HasSuffix(tuiURI, ".gsx") {
		return strings.TrimSuffix(tuiURI, ".gsx") + "_gsx_generated.go"
	}
	return tuiURI + "_generated.go"
}

// GoURIToTuiURI converts a virtual .go file URI back to the .gsx file URI.
func GoURIToTuiURI(goURI string) string {
	if strings.HasSuffix(goURI, "_gsx_generated.go") {
		return strings.TrimSuffix(goURI, "_gsx_generated.go") + ".gsx"
	}
	return goURI
}

// IsVirtualGoFile returns true if the URI is a virtual .go file.
func IsVirtualGoFile(uri string) bool {
	return strings.HasSuffix(uri, "_gsx_generated.go")
}

// GetVirtualFilePath returns the path where a virtual .go file would be created.
func GetVirtualFilePath(gsxPath string) string {
	dir := filepath.Dir(gsxPath)
	base := filepath.Base(gsxPath)
	if strings.HasSuffix(base, ".gsx") {
		base = strings.TrimSuffix(base, ".gsx") + "_gsx_generated.go"
	} else {
		base = base + "_generated.go"
	}
	return filepath.Join(dir, base)
}
