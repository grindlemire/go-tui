package lsp

import (
	"github.com/grindlemire/go-tui/internal/lsp/provider"
	"github.com/grindlemire/go-tui/internal/lsp/schema"
)

// HoverParams represents textDocument/hover parameters.
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Hover and MarkupContent are type aliases for the canonical definitions
// in the provider package, eliminating duplicate type definitions.
type Hover = provider.Hover
type MarkupContent = provider.MarkupContent

// getElementAttributes returns attribute documentation for an element tag.
// Delegates to the centralized schema. Used by features_test.go.
func getElementAttributes(tag string) []schema.AttributeDef {
	elem := schema.GetElement(tag)
	if elem == nil {
		return nil
	}
	return elem.Attributes
}

// isElementTag returns true if the word is a known element tag.
// Delegates to the centralized schema. Used by features_test.go.
func isElementTag(word string) bool {
	return schema.GetElement(word) != nil
}
