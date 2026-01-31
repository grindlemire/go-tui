package lsp

// NOTE: These provider interfaces mirror the ones in provider/provider.go.
// Both sets exist to avoid circular imports: the lsp package uses these interfaces
// (with lsp-package types like *CursorContext, *Document), while the provider package
// defines its own copies (with provider-package types). The adapters in
// provider_adapters.go bridge the two. Changes to a provider interface must be
// made in both files.

// HoverProvider produces hover documentation for a cursor position.
type HoverProvider interface {
	Hover(ctx *CursorContext) (*Hover, error)
}

// CompletionProvider produces completion suggestions for a cursor position.
type CompletionProvider interface {
	Complete(ctx *CursorContext) (*CompletionList, error)
}

// DefinitionProvider resolves go-to-definition for a cursor position.
type DefinitionProvider interface {
	Definition(ctx *CursorContext) ([]Location, error)
}

// ReferencesProvider finds all references to the symbol at the cursor.
type ReferencesProvider interface {
	References(ctx *CursorContext, includeDecl bool) ([]Location, error)
}

// DocumentSymbolProvider returns the symbol hierarchy for a document.
type DocumentSymbolProvider interface {
	DocumentSymbols(doc *Document) ([]DocumentSymbol, error)
}

// WorkspaceSymbolProvider searches for symbols across the workspace.
type WorkspaceSymbolProvider interface {
	WorkspaceSymbols(query string) ([]SymbolInformation, error)
}

// DiagnosticsProvider produces diagnostics for a document.
type DiagnosticsProvider interface {
	Diagnose(doc *Document) ([]Diagnostic, error)
}

// FormattingProvider formats a document.
type FormattingProvider interface {
	Format(doc *Document, opts FormattingOptions) ([]TextEdit, error)
}

// SemanticTokensProvider produces semantic tokens for syntax highlighting.
type SemanticTokensProvider interface {
	SemanticTokensFull(doc *Document) (*SemanticTokens, error)
}

// Registry holds all registered LSP providers.
// The router dispatches to these providers when handling requests.
type Registry struct {
	Hover           HoverProvider
	Completion      CompletionProvider
	Definition      DefinitionProvider
	References      ReferencesProvider
	DocumentSymbol  DocumentSymbolProvider
	WorkspaceSymbol WorkspaceSymbolProvider
	Diagnostics     DiagnosticsProvider
	Formatting      FormattingProvider
	SemanticTokens  SemanticTokensProvider
}
