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

	"github.com/grindlemire/go-tui/internal/lsp/log"
)

// DiagnosticCallback is called when gopls publishes diagnostics.
// The URI is the original .gsx file URI, and diagnostics have translated positions.
type DiagnosticCallback func(uri string, diagnostics []GoplsDiagnostic)

// SourceMapLookup is called to retrieve a source map for position translation.
// Returns goLine, goCol -> gsxLine, gsxCol translation function, or nil if not found.
type SourceMapLookup func(gsxURI string) func(goLine, goCol int) (gsxLine, gsxCol int, found bool)

// GoplsDiagnostic represents a diagnostic from gopls with translated positions.
type GoplsDiagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
	Source   string `json:"source"`
}

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

	// Diagnostic callback
	diagnosticCallback DiagnosticCallback
	diagnosticMu       sync.RWMutex

	// Source map lookup callback
	sourceMapLookup SourceMapLookup

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

// SetDiagnosticCallback sets the callback for diagnostic notifications.
func (p *GoplsProxy) SetDiagnosticCallback(cb DiagnosticCallback) {
	p.diagnosticMu.Lock()
	defer p.diagnosticMu.Unlock()
	p.diagnosticCallback = cb
}

// SetSourceMapLookup sets the callback for source map lookup.
func (p *GoplsProxy) SetSourceMapLookup(lookup SourceMapLookup) {
	p.sourceMapLookup = lookup
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

// Notification represents a JSON-RPC notification (no ID).
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// publishDiagnosticsParams represents gopls diagnostic notification params.
type publishDiagnosticsParams struct {
	URI         string            `json:"uri"`
	Diagnostics []goplsDiagnostic `json:"diagnostics"`
}

// goplsDiagnostic is the raw diagnostic from gopls.
type goplsDiagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
	Source   string `json:"source"`
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

		// First try to parse as a notification (no ID field or ID=0)
		var notif Notification
		if err := json.Unmarshal(msg, &notif); err == nil && notif.Method != "" {
			p.handleNotification(&notif)
			continue
		}

		var resp Response
		if err := json.Unmarshal(msg, &resp); err != nil {
			log.Gopls("Error parsing response: %v", err)
			continue
		}

		// If this is a notification (no ID), skip
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

// handleNotification processes a notification from gopls.
func (p *GoplsProxy) handleNotification(notif *Notification) {
	log.Gopls("Received notification: method=%s", notif.Method)

	if notif.Method != "textDocument/publishDiagnostics" {
		return
	}

	var params publishDiagnosticsParams
	if err := json.Unmarshal(notif.Params, &params); err != nil {
		log.Gopls("Error parsing diagnostics params: %v", err)
		return
	}

	log.Gopls("Diagnostics for URI=%s count=%d", params.URI, len(params.Diagnostics))
	for i, d := range params.Diagnostics {
		log.Gopls("  [%d] %s: %s", i, d.Range, d.Message)
	}

	// Only process diagnostics from real generated files (_gsx.go)
	// Skip virtual files (_gsx_generated.go) - they exist only for hover/completion
	// and having both can confuse gopls with duplicate type definitions
	if IsVirtualGoFile(params.URI) {
		log.Gopls("Skipping virtual file diagnostics: %s", params.URI)
		return
	}

	if !IsGeneratedGoFile(params.URI) {
		log.Gopls("Skipping - not a generated _gsx.go file: %s", params.URI)
		return
	}

	// Real generated file (counter_gsx.go) - needs goimports offset
	gsxURI := GeneratedGoURIToTuiURI(params.URI)
	lineOffset := 1 // goimports adds 1 blank line between import groups
	log.Gopls("Real file diagnostics for %s (offset=%d)", gsxURI, lineOffset)
	log.Gopls("Mapped to gsxURI=%s", gsxURI)

	// Get source map lookup function
	var translatePos func(goLine, goCol int) (gsxLine, gsxCol int, found bool)
	if p.sourceMapLookup != nil {
		translatePos = p.sourceMapLookup(gsxURI)
	}
	if translatePos == nil {
		log.Gopls("No source map available for %s", gsxURI)
		return
	}

	// Translate diagnostics
	var translated []GoplsDiagnostic
	for _, diag := range params.Diagnostics {
		// Skip errors caused by our virtual/generated file setup.
		// Gopls sees both virtual files and real _gsx.go files, causing conflicts.
		// Filter redeclaration and unknown field errors that reference generated files.
		isRedeclaration := strings.Contains(diag.Message, "redeclared") || strings.Contains(diag.Message, "already declared")
		referencesGeneratedFile := strings.Contains(diag.Message, "_gsx.go")
		isBlockRedeclaration := strings.Contains(diag.Message, "redeclared in this block")
		if isRedeclaration && (referencesGeneratedFile || isBlockRedeclaration) {
			log.Gopls("Skipping redeclaration error: %s", diag.Message)
			continue
		}
		if strings.Contains(diag.Message, "unknown field") && strings.Contains(diag.Message, "in struct literal") {
			log.Gopls("Skipping unknown field error: %s", diag.Message)
			continue
		}

		// Translate positions using source map
		// For real files, goimports adds blank lines that shift positions
		adjustedStartLine := diag.Range.Start.Line - lineOffset
		if adjustedStartLine < 0 {
			adjustedStartLine = 0
		}
		gsxStartLine, gsxStartCol, startFound := translatePos(adjustedStartLine, diag.Range.Start.Character)

		adjustedEndLine := diag.Range.End.Line - lineOffset
		if adjustedEndLine < 0 {
			adjustedEndLine = 0
		}
		gsxEndLine, gsxEndCol, endFound := translatePos(adjustedEndLine, diag.Range.End.Character)

		if !startFound || !endFound {
			log.Gopls("Could not translate diagnostic position: line=%d col=%d msg=%s",
				diag.Range.Start.Line, diag.Range.Start.Character, diag.Message)
			continue
		}

		log.Gopls("Translated: go=%d:%d -> gsx=%d:%d msg=%s",
			diag.Range.Start.Line, diag.Range.Start.Character,
			gsxStartLine, gsxStartCol, diag.Message)

		translated = append(translated, GoplsDiagnostic{
			Range: Range{
				Start: Position{Line: gsxStartLine, Character: gsxStartCol},
				End:   Position{Line: gsxEndLine, Character: gsxEndCol},
			},
			Severity: diag.Severity,
			Message:  diag.Message,
			Source:   "gopls",
		})
	}

	// Call the callback if registered
	p.diagnosticMu.RLock()
	cb := p.diagnosticCallback
	p.diagnosticMu.RUnlock()

	if cb != nil && len(translated) > 0 {
		cb(gsxURI, translated)
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

// IsGeneratedGoFile returns true if the URI is a real generated _gsx.go file (on disk).
func IsGeneratedGoFile(uri string) bool {
	// Match files like counter_gsx.go but NOT counter_gsx_generated.go (virtual)
	return strings.HasSuffix(uri, "_gsx.go") && !strings.HasSuffix(uri, "_gsx_generated.go")
}

// GeneratedGoURIToTuiURI converts a real generated _gsx.go file URI to the .gsx file URI.
func GeneratedGoURIToTuiURI(goURI string) string {
	if strings.HasSuffix(goURI, "_gsx.go") {
		return strings.TrimSuffix(goURI, "_gsx.go") + ".gsx"
	}
	return goURI
}
