package lsp

import "github.com/grindlemire/go-tui/internal/lsp/provider"

// DocumentFormattingParams represents textDocument/formatting parameters.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions and TextEdit are type aliases for the canonical definitions
// in the provider package, eliminating duplicate type definitions.
type FormattingOptions = provider.FormattingOptions
type TextEdit = provider.TextEdit
