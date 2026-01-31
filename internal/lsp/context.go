package lsp

import (
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// NodeKind classifies what language construct the cursor is on.
type NodeKind int

const (
	NodeKindUnknown NodeKind = iota
	NodeKindComponent
	NodeKindElement
	NodeKindAttribute
	NodeKindRefAttr // Cursor on ref={} attribute value
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
	case NodeKindRefAttr:
		return "RefAttr"
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

// Scope holds the enclosing scope information for a cursor position.
type Scope struct {
	Component *tuigen.Component  // Enclosing component (nil if at file level)
	Function  *tuigen.GoFunc     // Enclosing function (nil if not in a function)
	ForLoop   *tuigen.ForLoop    // Enclosing for loop (nil if not in a loop)
	IfStmt    *tuigen.IfStmt     // Enclosing if statement (nil if not in conditional)
	Refs      []tuigen.RefInfo   // Element refs in scope (from ref={} attributes)
	StateVars []tuigen.StateVar  // State variables in scope
	LetBinds  []*tuigen.LetBinding
	Params    []*tuigen.Param
}

// CursorContext contains resolved information about the cursor position.
// Providers receive this instead of raw positions, centralizing all
// "what is under the cursor" logic.
type CursorContext struct {
	Document *Document
	Position Position
	Offset   int

	// Resolved AST information
	Node        tuigen.Node   // The AST node at the cursor (may be nil)
	NodeKind    NodeKind      // Classification of the cursor position
	Scope       *Scope        // Enclosing scope information
	ParentChain []tuigen.Node // Path from root to current node

	// Convenience fields
	Word        string // Word under cursor
	Line        string // Full line text
	InGoExpr    bool   // Inside a Go expression {..}
	InClassAttr bool   // Inside class="..."
	InElement   bool   // Inside an element tag (between < and >)

	// For attribute context
	AttrTag  string // Element tag when cursor is on an attribute
	AttrName string // Attribute name when cursor is on an attribute
}

// ResolveCursorContext resolves the cursor context for a document position.
// This walks the AST and surrounding text to determine what the cursor is
// pointing at and what scope it's in.
func ResolveCursorContext(doc *Document, pos Position) *CursorContext {
	ctx := &CursorContext{
		Document: doc,
		Position: pos,
		Offset:   PositionToOffset(doc.Content, pos),
		Scope:    &Scope{},
	}

	// Extract line text
	ctx.Line = getLineText(doc.Content, pos.Line)

	// Extract word under cursor
	ctx.Word = getWordAtOffset(doc.Content, ctx.Offset)

	// Check text-level context flags
	ctx.InGoExpr = isOffsetInGoExpr(doc.Content, ctx.Offset)
	ctx.InClassAttr = isOffsetInClassAttr(doc.Content, ctx.Offset)
	ctx.InElement = isOffsetInElementTag(doc.Content, ctx.Offset)

	// If no AST, classify from text heuristics only
	if doc.AST == nil {
		ctx.NodeKind = classifyFromText(ctx)
		return ctx
	}

	// Walk AST to resolve node, scope, and kind
	resolveFromAST(ctx, doc.AST)

	return ctx
}
