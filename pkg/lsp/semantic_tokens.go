package lsp

import "github.com/grindlemire/go-tui/pkg/lsp/provider"

// SemanticTokensParams represents textDocument/semanticTokens/full parameters.
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SemanticTokens is a type alias for the canonical definition
// in the provider package, eliminating duplicate type definitions.
type SemanticTokens = provider.SemanticTokens
