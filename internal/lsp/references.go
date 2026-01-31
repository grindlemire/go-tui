package lsp

// ReferenceParams represents textDocument/references parameters.
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

// ReferenceContext contains additional context for references.
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

