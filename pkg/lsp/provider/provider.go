// Package provider implements LSP feature providers that the router dispatches to.
// Each provider handles a specific LSP capability (hover, definition, references, etc.)
// using the CursorContext for position resolution and the schema for language knowledge.
//
// Providers depend on abstractions (interfaces) from the parent lsp package to avoid
// circular imports. The lsp.Server injects concrete implementations when constructing
// providers.
package provider

import (
	"github.com/grindlemire/go-tui/pkg/lsp/gopls"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// --- Type aliases for parent-package types ---
// These let provider code reference types without importing the parent lsp package.

// CursorContext is the resolved cursor context, mirroring lsp.CursorContext.
// This type is duplicated here to avoid circular imports between lsp and provider.
// The lsp package's router converts its CursorContext to this type before dispatch.
// Changes to either CursorContext must be reflected in both packages.
type CursorContext struct {
	Document *Document
	Position Position
	Offset   int

	Node        tuigen.Node
	NodeKind    NodeKind
	Scope       *Scope
	ParentChain []tuigen.Node

	Word        string
	Line        string
	InGoExpr    bool
	InClassAttr bool
	InElement   bool

	AttrTag  string
	AttrName string
}

// NodeKind classifies the cursor position. Mirrors lsp.NodeKind — both enums
// must stay in sync. See CursorContext comment above for the duplication rationale.
type NodeKind int

const (
	NodeKindUnknown NodeKind = iota
	NodeKindComponent
	NodeKindElement
	NodeKindAttribute
	NodeKindNamedRef
	NodeKindGoExpr
	NodeKindForLoop
	NodeKindIfStmt
	NodeKindLetBinding
	NodeKindStateDecl
	NodeKindStateAccess
	NodeKindParameter
	NodeKindFunction
	NodeKindComponentCall
	NodeKindEventHandler
	NodeKindText
	NodeKindKeyword
	NodeKindTailwindClass
)

// String returns a human-readable name for the NodeKind.
func (k NodeKind) String() string {
	switch k {
	case NodeKindComponent:
		return "Component"
	case NodeKindElement:
		return "Element"
	case NodeKindAttribute:
		return "Attribute"
	case NodeKindNamedRef:
		return "NamedRef"
	case NodeKindGoExpr:
		return "GoExpr"
	case NodeKindForLoop:
		return "ForLoop"
	case NodeKindIfStmt:
		return "IfStmt"
	case NodeKindLetBinding:
		return "LetBinding"
	case NodeKindStateDecl:
		return "StateDecl"
	case NodeKindStateAccess:
		return "StateAccess"
	case NodeKindParameter:
		return "Parameter"
	case NodeKindFunction:
		return "Function"
	case NodeKindComponentCall:
		return "ComponentCall"
	case NodeKindEventHandler:
		return "EventHandler"
	case NodeKindText:
		return "Text"
	case NodeKindKeyword:
		return "Keyword"
	case NodeKindTailwindClass:
		return "TailwindClass"
	default:
		return "Unknown"
	}
}

// Scope holds enclosing scope information. Mirrors lsp.Scope.
type Scope struct {
	Component *tuigen.Component
	Function  *tuigen.GoFunc
	ForLoop   *tuigen.ForLoop
	IfStmt    *tuigen.IfStmt
	NamedRefs []tuigen.NamedRef
	StateVars []tuigen.StateVar
	LetBinds  []*tuigen.LetBinding
	Params    []*tuigen.Param
}

// Document represents an open document in the editor.
type Document struct {
	URI     string
	Content string
	Version int
	AST     *tuigen.File
	Errors  []*tuigen.Error
}

// --- LSP protocol types ---

// Position is a 0-indexed line and character in a document.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a span in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location is a position in a specific document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Hover is the result of a hover request.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents markup content.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// --- Dependency interfaces ---
// These let providers access server state without importing the lsp package.

// ComponentIndex provides lookup of workspace-wide component and function definitions.
type ComponentIndex interface {
	Lookup(name string) (*ComponentInfo, bool)
	LookupFunc(name string) (*FuncInfo, bool)
	LookupParam(componentName, paramName string) (*ParamInfo, bool)
	LookupFuncParam(funcName, paramName string) (*FuncParamInfo, bool)
	All() []string
	AllFunctions() []string
}

// ComponentInfo stores information about a component.
type ComponentInfo struct {
	Name     string
	Location Location
	Params   []*tuigen.Param
}

// FuncInfo stores information about a function.
type FuncInfo struct {
	Name      string
	Location  Location
	Signature string
	Returns   string
}

// ParamInfo stores information about a component parameter.
type ParamInfo struct {
	Name          string
	Type          string
	ComponentName string
	Location      Location
}

// FuncParamInfo stores information about a function parameter.
type FuncParamInfo struct {
	Name     string
	Type     string
	FuncName string
	Location Location
}

// GoplsProxyAccessor provides access to the gopls proxy.
type GoplsProxyAccessor interface {
	GetProxy() *gopls.GoplsProxy
}

// VirtualFileAccessor provides access to the virtual file cache.
type VirtualFileAccessor interface {
	GetVirtualFile(uri string) *gopls.CachedVirtualFile
}

// DocumentAccessor provides access to open documents and workspace ASTs.
type DocumentAccessor interface {
	GetDocument(uri string) *Document
	AllDocuments() []*Document
}

// WorkspaceASTAccessor provides access to cached workspace ASTs.
type WorkspaceASTAccessor interface {
	GetWorkspaceAST(uri string) *tuigen.File
	AllWorkspaceASTs() map[string]*tuigen.File
}

// --- Provider interfaces ---

// HoverProvider produces hover documentation.
type HoverProvider interface {
	Hover(ctx *CursorContext) (*Hover, error)
}

// DefinitionProvider resolves go-to-definition.
type DefinitionProvider interface {
	Definition(ctx *CursorContext) ([]Location, error)
}

// ReferencesProvider finds all references to a symbol.
type ReferencesProvider interface {
	References(ctx *CursorContext, includeDecl bool) ([]Location, error)
}

// CompletionProvider produces completion suggestions.
type CompletionProvider interface {
	Complete(ctx *CursorContext) (*CompletionList, error)
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

// --- Symbol types ---

// DocumentSymbol represents a symbol found in a document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation represents a symbol for workspace symbols.
type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
}

// SymbolKind represents the kind of symbol.
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// --- Helper functions ---

// PositionToOffset converts a 0-indexed line/character to a byte offset in the content.
func PositionToOffset(content string, pos Position) int {
	line := 0
	for i, ch := range content {
		if line == pos.Line {
			return i + pos.Character
		}
		if ch == '\n' {
			line++
		}
	}
	if line == pos.Line {
		return len(content) + pos.Character
	}
	return len(content)
}

// IsWordChar returns true if c is a valid identifier character.
func IsWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
