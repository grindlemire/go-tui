// Package lsp provides a Language Server Protocol implementation for .tui files.
package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/grindlemire/go-tui/pkg/lsp/gopls"
	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// Server represents the TUI LSP server.
type Server struct {
	// Input/output for JSON-RPC communication
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex // protects writer

	// Document management
	docs *DocumentManager

	// Component index for workspace symbols and go-to-definition
	index *ComponentIndex

	// Workspace AST cache for files not open in editor
	workspaceASTs   map[string]*tuigen.File // URI -> AST
	workspaceASTsMu sync.RWMutex

	// gopls proxy for Go expression intelligence
	goplsProxy *gopls.GoplsProxy

	// Virtual file cache for gopls
	virtualFiles *gopls.VirtualFileCache

	// Server state
	initialized bool
	shutdown    bool
	rootURI     string

	// Context for gopls
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServer creates a new LSP server that communicates over the given reader/writer.
func NewServer(reader io.Reader, writer io.Writer) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		reader:        bufio.NewReader(reader),
		writer:        writer,
		docs:          NewDocumentManager(),
		index:         NewComponentIndex(),
		workspaceASTs: make(map[string]*tuigen.File),
		virtualFiles:  gopls.NewVirtualFileCache(),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// SetLogFile sets a file for debug logging.
func (s *Server) SetLogFile(f *os.File) {
	log.SetOutput(f)
}

// InitGopls initializes the gopls proxy. Call this after Initialize.
func (s *Server) InitGopls() error {
	if s.rootURI == "" {
		log.Server("Cannot init gopls without rootURI")
		return nil
	}

	proxy, err := gopls.NewGoplsProxy(s.ctx)
	if err != nil {
		log.Server("Failed to start gopls: %v", err)
		return nil // Non-fatal, continue without gopls
	}

	if err := proxy.Initialize(s.rootURI); err != nil {
		log.Server("Failed to initialize gopls: %v", err)
		proxy.Shutdown()
		return nil // Non-fatal
	}

	s.goplsProxy = proxy
	log.Server("gopls proxy initialized successfully")

	// Update virtual files for all already-open documents
	for _, doc := range s.docs.All() {
		s.UpdateVirtualFile(doc)
	}

	return nil
}

// ShutdownGopls shuts down the gopls proxy.
func (s *Server) ShutdownGopls() {
	if s.goplsProxy != nil {
		s.goplsProxy.Shutdown()
		s.goplsProxy = nil
	}
	if s.cancel != nil {
		s.cancel()
	}
}

// UpdateVirtualFile updates the virtual .go file for a .tui document.
func (s *Server) UpdateVirtualFile(doc *Document) {
	if s.goplsProxy == nil || doc.AST == nil {
		return
	}

	// Generate virtual Go file
	log.Server("=== Generating virtual Go file for %s ===", doc.URI)
	goContent, sourceMap := gopls.GenerateVirtualGo(doc.AST)
	goURI := gopls.TuiURIToGoURI(doc.URI)

	// Log the generated content
	log.Server("Generated Go content:\n%s", goContent)

	// Log all mappings
	log.Server("=== Source mappings (%d total) ===", sourceMap.Len())
	for i, m := range sourceMap.AllMappings() {
		log.Server("  [%d] TuiLine=%d TuiCol=%d -> GoLine=%d GoCol=%d Len=%d",
			i, m.TuiLine, m.TuiCol, m.GoLine, m.GoCol, m.Length)
	}
	log.Server("=== End mappings ===")

	// Check if we already have this file open in gopls
	cached := s.virtualFiles.Get(doc.URI)
	if cached != nil {
		// Update existing file
		if err := s.goplsProxy.UpdateVirtualFile(goURI, goContent, doc.Version); err != nil {
			log.Server("Failed to update virtual file: %v", err)
		}
	} else {
		// Open new file
		if err := s.goplsProxy.OpenVirtualFile(goURI, goContent, doc.Version); err != nil {
			log.Server("Failed to open virtual file: %v", err)
		}
	}

	// Update cache
	s.virtualFiles.Put(doc.URI, goURI, goContent, sourceMap, doc.Version)
}

// CloseVirtualFile closes the virtual .go file for a .tui document.
func (s *Server) CloseVirtualFile(uri string) {
	if s.goplsProxy == nil {
		return
	}

	cached := s.virtualFiles.Get(uri)
	if cached != nil {
		if err := s.goplsProxy.CloseVirtualFile(cached.GoURI); err != nil {
			log.Server("Failed to close virtual file: %v", err)
		}
		s.virtualFiles.Remove(uri)
	}
}

// Run starts the LSP server main loop.
func (s *Server) Run(ctx context.Context) error {
	log.Server("LSP server starting")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				log.Server("Connection closed")
				return nil
			}
			log.Server("Error reading message: %v", err)
			return fmt.Errorf("reading message: %w", err)
		}

		log.Server("Received: %s", string(msg))

		response, err := s.handleMessage(msg)
		if err != nil {
			log.Server("Error handling message: %v", err)
			// Send error response if we have an ID
			continue
		}

		if response != nil {
			if err := s.writeMessage(response); err != nil {
				log.Server("Error writing response: %v", err)
				return fmt.Errorf("writing response: %w", err)
			}
		}

		if s.shutdown {
			log.Server("Server shutdown requested")
			return nil
		}
	}
}

// readMessage reads a JSON-RPC message from the input.
// Messages are formatted as HTTP-like headers followed by content:
// Content-Length: <length>\r\n
// \r\n
// <content>
func (s *Server) readMessage() ([]byte, error) {
	// Read headers
	var contentLength int
	for {
		line, err := s.reader.ReadString('\n')
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
	_, err := io.ReadFull(s.reader, content)
	if err != nil {
		return nil, fmt.Errorf("reading content: %w", err)
	}

	return content, nil
}

// writeMessage writes a JSON-RPC message to the output.
func (s *Server) writeMessage(msg []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(msg))
	if _, err := s.writer.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := s.writer.Write(msg); err != nil {
		return err
	}

	log.Server("Sent: %s", string(msg))
	return nil
}

// sendNotification sends a notification (no response expected).
func (s *Server) sendNotification(method string, params any) error {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return s.writeMessage(data)
}

// Request represents a JSON-RPC request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"` // can be number or string
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result"`
	Error   *Error `json:"error,omitempty"`
}

// Error represents a JSON-RPC error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JSON-RPC error codes
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// handleMessage processes a single JSON-RPC message.
func (s *Server) handleMessage(msg []byte) ([]byte, error) {
	var req Request
	if err := json.Unmarshal(msg, &req); err != nil {
		return s.errorResponse(nil, CodeParseError, "Parse error")
	}

	log.Server("Handling method: %s", req.Method)

	// Route to appropriate handler
	result, rpcErr := s.route(req)

	// Notifications don't get responses
	if req.ID == nil {
		return nil, nil
	}

	if rpcErr != nil {
		return s.errorResponse(req.ID, rpcErr.Code, rpcErr.Message)
	}

	resp := Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
	return json.Marshal(resp)
}

// errorResponse creates an error response.
func (s *Server) errorResponse(id any, code int, message string) ([]byte, error) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
	return json.Marshal(resp)
}
