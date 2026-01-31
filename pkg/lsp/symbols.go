package lsp

import "github.com/grindlemire/go-tui/pkg/lsp/provider"

// DocumentSymbolParams represents textDocument/documentSymbol parameters.
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// WorkspaceSymbolParams represents workspace/symbol parameters.
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

// DocumentSymbol, SymbolInformation, and SymbolKind are type aliases for the
// canonical definitions in the provider package, eliminating duplicate type definitions.
type DocumentSymbol = provider.DocumentSymbol
type SymbolInformation = provider.SymbolInformation
type SymbolKind = provider.SymbolKind

// Re-export SymbolKind constants so existing lsp package code compiles unchanged.
const (
	SymbolKindFile          = provider.SymbolKindFile
	SymbolKindModule        = provider.SymbolKindModule
	SymbolKindNamespace     = provider.SymbolKindNamespace
	SymbolKindPackage       = provider.SymbolKindPackage
	SymbolKindClass         = provider.SymbolKindClass
	SymbolKindMethod        = provider.SymbolKindMethod
	SymbolKindProperty      = provider.SymbolKindProperty
	SymbolKindField         = provider.SymbolKindField
	SymbolKindConstructor   = provider.SymbolKindConstructor
	SymbolKindEnum          = provider.SymbolKindEnum
	SymbolKindInterface     = provider.SymbolKindInterface
	SymbolKindFunction      = provider.SymbolKindFunction
	SymbolKindVariable      = provider.SymbolKindVariable
	SymbolKindConstant      = provider.SymbolKindConstant
	SymbolKindString        = provider.SymbolKindString
	SymbolKindNumber        = provider.SymbolKindNumber
	SymbolKindBoolean       = provider.SymbolKindBoolean
	SymbolKindArray         = provider.SymbolKindArray
	SymbolKindObject        = provider.SymbolKindObject
	SymbolKindKey           = provider.SymbolKindKey
	SymbolKindNull          = provider.SymbolKindNull
	SymbolKindEnumMember    = provider.SymbolKindEnumMember
	SymbolKindStruct        = provider.SymbolKindStruct
	SymbolKindEvent         = provider.SymbolKindEvent
	SymbolKindOperator      = provider.SymbolKindOperator
	SymbolKindTypeParameter = provider.SymbolKindTypeParameter
)
