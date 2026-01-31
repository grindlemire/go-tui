package lsp

// DefinitionParams represents textDocument/definition parameters.
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// isWordChar returns true if c is a valid word character.
// Used by context.go for cursor word extraction.
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
