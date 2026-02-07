package lsp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// InitializeParams represents the parameters for the initialize request.
type InitializeParams struct {
	ProcessID             *int               `json:"processId"`
	RootURI               string             `json:"rootUri"`
	RootPath              string             `json:"rootPath"`
	Capabilities          ClientCapabilities `json:"capabilities"`
	InitializationOptions json.RawMessage    `json:"initializationOptions,omitempty"`
}

// ClientCapabilities represents client capabilities.
type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

// TextDocumentClientCapabilities represents text document capabilities.
type TextDocumentClientCapabilities struct {
	Synchronization    *SynchronizationCapabilities `json:"synchronization,omitempty"`
	Completion         *CompletionCapabilities      `json:"completion,omitempty"`
	Hover              *HoverCapabilities           `json:"hover,omitempty"`
	Definition         *DefinitionCapabilities      `json:"definition,omitempty"`
	PublishDiagnostics *PublishDiagnostics          `json:"publishDiagnostics,omitempty"`
}

// SynchronizationCapabilities represents synchronization capabilities.
type SynchronizationCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	WillSave            bool `json:"willSave,omitempty"`
	WillSaveWaitUntil   bool `json:"willSaveWaitUntil,omitempty"`
	DidSave             bool `json:"didSave,omitempty"`
}

// CompletionCapabilities represents completion capabilities.
type CompletionCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

// HoverCapabilities represents hover capabilities.
type HoverCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

// DefinitionCapabilities represents definition capabilities.
type DefinitionCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

// PublishDiagnostics represents publish diagnostics capabilities.
type PublishDiagnostics struct {
	RelatedInformation bool `json:"relatedInformation,omitempty"`
}

// InitializeResult represents the result of the initialize request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ServerCapabilities represents server capabilities.
type ServerCapabilities struct {
	TextDocumentSync           *TextDocumentSyncOptions  `json:"textDocumentSync,omitempty"`
	CompletionProvider         *CompletionOptions        `json:"completionProvider,omitempty"`
	HoverProvider              bool                      `json:"hoverProvider,omitempty"`
	DefinitionProvider         bool                      `json:"definitionProvider,omitempty"`
	ReferencesProvider         bool                      `json:"referencesProvider,omitempty"`
	DocumentSymbolProvider     bool                      `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider    bool                      `json:"workspaceSymbolProvider,omitempty"`
	DocumentFormattingProvider bool                      `json:"documentFormattingProvider,omitempty"`
	SemanticTokensProvider     *SemanticTokensOptions    `json:"semanticTokensProvider,omitempty"`
}

// SemanticTokensOptions represents semantic tokens capabilities.
type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
}

// SemanticTokensLegend describes the token types and modifiers.
type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

// TextDocumentSyncOptions represents text document sync options.
type TextDocumentSyncOptions struct {
	OpenClose bool                 `json:"openClose"`
	Change    TextDocumentSyncKind `json:"change"`
	Save      *SaveOptions         `json:"save,omitempty"`
}

// TextDocumentSyncKind represents how documents are synced.
type TextDocumentSyncKind int

const (
	// TextDocumentSyncKindNone means documents should not be synced.
	TextDocumentSyncKindNone TextDocumentSyncKind = 0
	// TextDocumentSyncKindFull means full documents are synced.
	TextDocumentSyncKindFull TextDocumentSyncKind = 1
	// TextDocumentSyncKindIncremental means incremental updates are sent.
	TextDocumentSyncKindIncremental TextDocumentSyncKind = 2
)

// SaveOptions represents save options.
type SaveOptions struct {
	IncludeText bool `json:"includeText,omitempty"`
}

// CompletionOptions represents completion options.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
}

// handleInitialize handles the initialize request.
func (s *Server) handleInitialize(params json.RawMessage) (any, *Error) {
	var p InitializeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	s.rootURI = p.RootURI
	log.Server("Initialize with root: %s", s.rootURI)

	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    TextDocumentSyncKindFull,
				Save: &SaveOptions{
					IncludeText: true,
				},
			},
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{"@", "<", "{", "."},
			},
			HoverProvider:              true,
			DefinitionProvider:         true,
			ReferencesProvider:         true,
			DocumentSymbolProvider:     true,
			WorkspaceSymbolProvider:    true,
			DocumentFormattingProvider: true,
			SemanticTokensProvider: &SemanticTokensOptions{
				Legend: SemanticTokensLegend{
					TokenTypes: []string{
						"namespace",   // 0: package
						"type",        // 1: types
						"class",       // 2: components
						"function",    // 3: functions
						"parameter",   // 4: parameters
						"variable",    // 5: variables
						"property",    // 6: attributes
						"keyword",     // 7: keywords (templ, @for, etc.)
						"string",      // 8: strings
						"number",      // 9: numbers
						"operator",    // 10: operators
						"decorator",   // 11: @ prefix
						"regexp",      // 12: format specifiers (often purple)
						"comment",     // 13: comments
					"label",         // 14: named refs (#Name)
						"typeParameter", // 15: generic type arguments
					},
					TokenModifiers: []string{
						"declaration",  // 0: where defined
						"definition",   // 1: where defined
						"readonly",     // 2: const/let
						"modification", // 3: where modified
					},
				},
				Full: true,
			},
		},
	}

	return result, nil
}

// handleInitialized handles the initialized notification.
func (s *Server) handleInitialized() (any, *Error) {
	s.initialized = true
	log.Server("Server initialized")

	// Index all .gsx files in the workspace for cross-file references
	go s.indexWorkspace()

	// Start gopls proxy in the background
	go s.InitGopls()

	return nil, nil
}

// handleShutdown handles the shutdown request.
func (s *Server) handleShutdown() (any, *Error) {
	log.Server("Shutdown requested")
	s.shutdown = true

	// Shutdown gopls proxy
	s.ShutdownGopls()

	return nil, nil
}

// handleExit handles the exit notification.
func (s *Server) handleExit() {
	log.Server("Exit requested")
	// The server will exit after returning from the handler
}

// DidOpenParams represents textDocument/didOpen parameters.
type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentItem represents an item passed in didOpen.
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// handleDidOpen handles textDocument/didOpen.
func (s *Server) handleDidOpen(params json.RawMessage) (any, *Error) {
	var p DidOpenParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("Document opened: %s", p.TextDocument.URI)

	doc := s.docs.Open(p.TextDocument.URI, p.TextDocument.Text, p.TextDocument.Version)

	// Index components from this document
	s.index.IndexDocument(p.TextDocument.URI, doc.AST)

	// Remove from workspace cache (now managed by docs)
	s.workspaceASTsMu.Lock()
	delete(s.workspaceASTs, p.TextDocument.URI)
	s.workspaceASTsMu.Unlock()

	// Update virtual Go file for gopls
	s.UpdateVirtualFile(doc)

	// Publish diagnostics
	s.publishDiagnostics(doc)

	return nil, nil
}

// DidChangeParams represents textDocument/didChange parameters.
type DidChangeParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// VersionedTextDocumentIdentifier represents a versioned document ID.
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// TextDocumentContentChangeEvent represents a content change.
type TextDocumentContentChangeEvent struct {
	// Full text sync: Text contains the whole document
	Text string `json:"text"`
}

// handleDidChange handles textDocument/didChange.
func (s *Server) handleDidChange(params json.RawMessage) (any, *Error) {
	var p DidChangeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("Document changed: %s", p.TextDocument.URI)

	if len(p.ContentChanges) == 0 {
		return nil, nil
	}

	// We use full document sync, so take the last change
	newContent := p.ContentChanges[len(p.ContentChanges)-1].Text
	doc := s.docs.Update(p.TextDocument.URI, newContent, p.TextDocument.Version)

	if doc != nil {
		// Re-index components from this document
		s.index.IndexDocument(p.TextDocument.URI, doc.AST)

		// Update virtual Go file for gopls
		s.UpdateVirtualFile(doc)

		// Publish diagnostics
		s.publishDiagnostics(doc)
	}

	return nil, nil
}

// DidCloseParams represents textDocument/didClose parameters.
type DidCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// TextDocumentIdentifier represents a document identifier.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// handleDidClose handles textDocument/didClose.
func (s *Server) handleDidClose(params json.RawMessage) (any, *Error) {
	var p DidCloseParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("Document closed: %s", p.TextDocument.URI)

	// Get the AST before closing so we can cache it
	doc := s.docs.Get(p.TextDocument.URI)
	var ast *tuigen.File
	if doc != nil {
		ast = doc.AST
	}

	s.docs.Close(p.TextDocument.URI)

	// Re-add to workspace cache if we have an AST, and re-index
	if ast != nil {
		s.workspaceASTsMu.Lock()
		s.workspaceASTs[p.TextDocument.URI] = ast
		s.workspaceASTsMu.Unlock()
		// Re-index so lookups still work
		s.index.IndexDocument(p.TextDocument.URI, ast)
	} else {
		// Remove components from index
		s.index.Remove(p.TextDocument.URI)
	}

	// Close virtual Go file in gopls
	s.CloseVirtualFile(p.TextDocument.URI)

	// Clear diagnostics for closed document
	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         p.TextDocument.URI,
		Diagnostics: []Diagnostic{},
	})

	return nil, nil
}

// DidSaveParams represents textDocument/didSave parameters.
type DidSaveParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         *string                `json:"text,omitempty"`
}

// handleDidSave handles textDocument/didSave.
func (s *Server) handleDidSave(params json.RawMessage) (any, *Error) {
	var p DidSaveParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("Document saved: %s", p.TextDocument.URI)

	// Invalidate source map cache so it reloads from disk after regeneration
	if s.sourceMapCache != nil {
		s.sourceMapCache.Invalidate(p.TextDocument.URI)
		log.Server("Invalidated source map cache for: %s", p.TextDocument.URI)
	}

	// If text is provided, update the document
	if p.Text != nil {
		doc := s.docs.Get(p.TextDocument.URI)
		if doc != nil {
			doc = s.docs.Update(p.TextDocument.URI, *p.Text, doc.Version+1)
			s.index.IndexDocument(p.TextDocument.URI, doc.AST)
			s.publishDiagnostics(doc)
		}
	}

	return nil, nil
}

// indexWorkspace scans the workspace for .gsx files and indexes them.
// This enables cross-file go-to-definition for components and functions.
func (s *Server) indexWorkspace() {
	if s.rootURI == "" {
		log.Server("Cannot index workspace: no rootURI")
		return
	}

	// Convert file:// URI to path
	rootPath := strings.TrimPrefix(s.rootURI, "file://")
	log.Server("=== WORKSPACE INDEXING START ===")
	log.Server("rootURI: %s", s.rootURI)
	log.Server("rootPath: %s", rootPath)

	count := 0
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Skip hidden directories and common non-source directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .gsx files
		if !strings.HasSuffix(path, ".gsx") {
			return nil
		}

		// Skip if file is already open (will be indexed via didOpen)
		uri := "file://" + path
		if s.docs.Get(uri) != nil {
			return nil
		}

		// Read and parse the file
		content, err := os.ReadFile(path)
		if err != nil {
			log.Server("Failed to read %s: %v", path, err)
			return nil
		}

		// Parse the file
		lexer := tuigen.NewLexer(path, string(content))
		parser := tuigen.NewParser(lexer)
		ast, err := parser.ParseFile()
		if err != nil {
			log.Server("Parse error in %s: %v", path, err)
			// Still index what we can
		}

		if ast != nil {
			s.index.IndexDocument(uri, ast)
			// Cache the AST for later use (e.g., find references)
			s.workspaceASTsMu.Lock()
			s.workspaceASTs[uri] = ast
			s.workspaceASTsMu.Unlock()
			count++
			log.Server("Indexed %s: %d components, %d functions",
				path, len(ast.Components), len(ast.Funcs))
		}

		return nil
	})

	if err != nil {
		log.Server("Error walking workspace: %v", err)
	}

	log.Server("Workspace indexing complete: %d files indexed", count)
	log.Server("All indexed functions: %v", s.index.AllFunctions())
	log.Server("All indexed components: %v", s.index.All())
	log.Server("=== WORKSPACE INDEXING END ===")
}
