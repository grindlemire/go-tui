package lsp

import (
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/schema"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// NodeKind classifies what language construct the cursor is on.
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

// Scope holds the enclosing scope information for a cursor position.
type Scope struct {
	Component *tuigen.Component  // Enclosing component (nil if at file level)
	Function  *tuigen.GoFunc     // Enclosing function (nil if not in a function)
	ForLoop   *tuigen.ForLoop    // Enclosing for loop (nil if not in a loop)
	IfStmt    *tuigen.IfStmt     // Enclosing if statement (nil if not in conditional)
	NamedRefs []tuigen.NamedRef  // Named refs in scope
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

// resolveFromAST walks the parsed AST to find the node at the cursor position
// and populate scope information.
func resolveFromAST(ctx *CursorContext, file *tuigen.File) {
	// Convert LSP 0-indexed position to tuigen 1-indexed
	line := ctx.Position.Line + 1
	col := ctx.Position.Character + 1

	// Check if cursor is on a component declaration line (name or parameter).
	// We check all components for exact line match first because it's a
	// precise match regardless of component ordering.
	for _, comp := range file.Components {
		if line == comp.Position.Line {
			// Check if cursor is on the component name
			nameStart := comp.Position.Column
			nameEnd := nameStart + len(comp.Name)
			if col >= nameStart && col <= nameEnd {
				ctx.Node = comp
				ctx.NodeKind = NodeKindComponent
				ctx.Scope.Component = comp
				ctx.Scope.Params = comp.Params
				collectScopeFromBody(ctx, comp.Body, comp)
				return
			}

			// Check if cursor is on a parameter
			for _, p := range comp.Params {
				if p.Position.Line == line {
					pStart := p.Position.Column
					pEnd := pStart + len(p.Name)
					if col >= pStart && col <= pEnd {
						ctx.Node = p
						ctx.NodeKind = NodeKindParameter
						ctx.Scope.Component = comp
						ctx.Scope.Params = comp.Params
						collectScopeFromBody(ctx, comp.Body, comp)
						return
					}
				}
			}
		}
	}

	// Find the enclosing component by selecting the last component whose
	// declaration line is <= the cursor line. Components are ordered by
	// position, so the last one that starts before the cursor is the
	// innermost enclosing component.
	var enclosingComp *tuigen.Component
	for _, comp := range file.Components {
		if line >= comp.Position.Line {
			enclosingComp = comp
		}
	}

	if enclosingComp != nil {
		ctx.Scope.Component = enclosingComp
		ctx.Scope.Params = enclosingComp.Params
		collectScopeFromBody(ctx, enclosingComp.Body, enclosingComp)

		if found := resolveInNodes(ctx, enclosingComp.Body, line, col); found {
			return
		}
	}

	// Check if cursor is on a function
	for _, fn := range file.Funcs {
		if line == fn.Position.Line {
			ctx.Node = fn
			ctx.NodeKind = NodeKindFunction
			ctx.Scope.Function = fn
			return
		}
	}

	// Fall back to text-based classification
	ctx.NodeKind = classifyFromText(ctx)
}

// resolveInNodes walks a list of AST nodes to find the one at the cursor.
func resolveInNodes(ctx *CursorContext, nodes []tuigen.Node, line, col int) bool {
	for _, node := range nodes {
		if found := resolveInNode(ctx, node, line, col); found {
			return true
		}
	}
	return false
}

// resolveInNode checks a single AST node and its children.
func resolveInNode(ctx *CursorContext, node tuigen.Node, line, col int) bool {
	switch n := node.(type) {
	case *tuigen.Element:
		return resolveInElement(ctx, n, line, col)
	case *tuigen.ForLoop:
		return resolveInForLoop(ctx, n, line, col)
	case *tuigen.IfStmt:
		return resolveInIfStmt(ctx, n, line, col)
	case *tuigen.LetBinding:
		return resolveInLetBinding(ctx, n, line, col)
	case *tuigen.ComponentCall:
		return resolveInComponentCall(ctx, n, line, col)
	case *tuigen.GoExpr:
		if n != nil && n.Position.Line == line {
			ctx.Node = n
			ctx.NodeKind = NodeKindGoExpr
			return true
		}
	case *tuigen.GoCode:
		if n != nil && n.Position.Line == line {
			ctx.Node = n
			ctx.NodeKind = classifyGoCode(ctx, n)
			return true
		}
	case *tuigen.TextContent:
		if n != nil && n.Position.Line == line {
			ctx.Node = n
			ctx.NodeKind = NodeKindText
			return true
		}
	}
	return false
}

// resolveInElement checks if cursor is within an element.
func resolveInElement(ctx *CursorContext, elem *tuigen.Element, line, col int) bool {
	if elem == nil {
		return false
	}

	pos := elem.Position

	// Check if cursor is on the element's tag line
	if pos.Line == line {
		// Check named ref (#Name)
		if elem.NamedRef != "" {
			// The # appears after the tag name. Search for it in the line text.
			hashIdx := strings.Index(ctx.Line, "#"+elem.NamedRef)
			if hashIdx >= 0 {
				refColStart := hashIdx + 1 // 0-indexed column of the ref name
				refColEnd := refColStart + len(elem.NamedRef)
				cursorCol := ctx.Position.Character
				if cursorCol >= hashIdx && cursorCol <= refColEnd {
					ctx.Node = elem
					ctx.NodeKind = NodeKindNamedRef
					return true
				}
			}
		}

		// Check if cursor is on the tag name
		tagStart := pos.Column
		tagEnd := tagStart + len(elem.Tag)
		if col >= tagStart && col <= tagEnd {
			ctx.Node = elem
			ctx.NodeKind = NodeKindElement
			return true
		}
	}

	// Check attributes
	for _, attr := range elem.Attributes {
		if attr.Position.Line == line {
			attrStart := attr.Position.Column
			attrEnd := attrStart + len(attr.Name)
			if col >= attrStart && col <= attrEnd {
				ctx.Node = attr
				ctx.NodeKind = NodeKindAttribute
				ctx.AttrTag = elem.Tag
				ctx.AttrName = attr.Name

				// Check if this is an event handler attribute
				if schema.IsEventHandler(attr.Name) {
					ctx.NodeKind = NodeKindEventHandler
				}
				return true
			}
		}
	}

	// Search children
	return resolveInNodes(ctx, elem.Children, line, col)
}

// resolveInForLoop checks if cursor is within a for loop.
func resolveInForLoop(ctx *CursorContext, loop *tuigen.ForLoop, line, col int) bool {
	if loop == nil {
		return false
	}

	if loop.Position.Line == line {
		ctx.Node = loop
		ctx.NodeKind = NodeKindForLoop
		ctx.Scope.ForLoop = loop
		return true
	}

	// Check body
	prevLoop := ctx.Scope.ForLoop
	ctx.Scope.ForLoop = loop
	if found := resolveInNodes(ctx, loop.Body, line, col); found {
		return true
	}
	ctx.Scope.ForLoop = prevLoop
	return false
}

// resolveInIfStmt checks if cursor is within an if statement.
func resolveInIfStmt(ctx *CursorContext, stmt *tuigen.IfStmt, line, col int) bool {
	if stmt == nil {
		return false
	}

	if stmt.Position.Line == line {
		ctx.Node = stmt
		ctx.NodeKind = NodeKindIfStmt
		ctx.Scope.IfStmt = stmt
		return true
	}

	// Check then/else branches
	prevIf := ctx.Scope.IfStmt
	ctx.Scope.IfStmt = stmt
	if found := resolveInNodes(ctx, stmt.Then, line, col); found {
		return true
	}
	if found := resolveInNodes(ctx, stmt.Else, line, col); found {
		return true
	}
	ctx.Scope.IfStmt = prevIf
	return false
}

// resolveInLetBinding checks if cursor is on a let binding.
func resolveInLetBinding(ctx *CursorContext, let *tuigen.LetBinding, line, col int) bool {
	if let == nil {
		return false
	}

	if let.Position.Line == line {
		// Check if cursor is on the variable name
		nameStart := let.Position.Column
		nameEnd := nameStart + len(let.Name)
		if col >= nameStart && col <= nameEnd {
			ctx.Node = let
			ctx.NodeKind = NodeKindLetBinding
			return true
		}
	}

	// Check element within let binding
	if let.Element != nil {
		return resolveInElement(ctx, let.Element, line, col)
	}
	return false
}

// resolveInComponentCall checks if cursor is on a component call.
func resolveInComponentCall(ctx *CursorContext, call *tuigen.ComponentCall, line, col int) bool {
	if call == nil {
		return false
	}

	if call.Position.Line == line {
		ctx.Node = call
		ctx.NodeKind = NodeKindComponentCall
		return true
	}

	// Check children
	return resolveInNodes(ctx, call.Children, line, col)
}

// classifyGoCode determines the NodeKind for a GoCode node.
// Detects state declarations (tui.NewState).
func classifyGoCode(ctx *CursorContext, code *tuigen.GoCode) NodeKind {
	if code == nil {
		return NodeKindGoExpr
	}

	trimmed := strings.TrimSpace(code.Code)

	// Check for state declaration: varName := tui.NewState(...)
	if strings.Contains(trimmed, "tui.NewState(") {
		return NodeKindStateDecl
	}

	// Check for state access: .Get(), .Set(), .Update(), .Bind(), .Batch()
	if strings.Contains(trimmed, ".Get()") ||
		strings.Contains(trimmed, ".Set(") ||
		strings.Contains(trimmed, ".Update(") ||
		strings.Contains(trimmed, ".Bind(") ||
		strings.Contains(trimmed, ".Batch(") {
		return NodeKindStateAccess
	}

	return NodeKindGoExpr
}

// classifyFromText classifies the cursor position using text heuristics
// when no AST node was found.
func classifyFromText(ctx *CursorContext) NodeKind {
	word := ctx.Word

	if ctx.InClassAttr {
		return NodeKindTailwindClass
	}
	if ctx.InGoExpr {
		return NodeKindGoExpr
	}

	// Check if word is a keyword
	if schema.GetKeyword(word) != nil {
		return NodeKindKeyword
	}

	// Check if word is an element tag
	if schema.IsElementTag(word) && ctx.InElement {
		return NodeKindElement
	}

	// Check if word starts with @ (component call)
	if strings.HasPrefix(word, "@") {
		return NodeKindComponentCall
	}

	return NodeKindUnknown
}

// collectScopeFromBody collects named refs, state vars, and let bindings from component body.
func collectScopeFromBody(ctx *CursorContext, nodes []tuigen.Node, comp *tuigen.Component) {
	// stateVarsCollected tracks whether DetectStateVars has already been called
	// for this component. DetectStateVars scans the entire component body, so it
	// only needs to be invoked once regardless of how many GoCode nodes exist.
	stateVarsCollected := false
	collectScopeFromBodyInner(ctx, nodes, comp, &stateVarsCollected)
}

func collectScopeFromBodyInner(ctx *CursorContext, nodes []tuigen.Node, comp *tuigen.Component, stateVarsCollected *bool) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.Element:
			if n.NamedRef != "" {
				ref := tuigen.NamedRef{
					Name:    n.NamedRef,
					Element: n,
				}
				if ctx.Scope.ForLoop != nil {
					ref.InLoop = true
				}
				if ctx.Scope.IfStmt != nil {
					ref.InConditional = true
				}
				if n.RefKey != nil {
					ref.KeyExpr = n.RefKey.Code
				}
				ctx.Scope.NamedRefs = append(ctx.Scope.NamedRefs, ref)
			}
			collectScopeFromBodyInner(ctx, n.Children, comp, stateVarsCollected)
		case *tuigen.GoCode:
			// Detect state variables via tui.NewState pattern. DetectStateVars
			// scans the entire component, so we only call it once per component.
			if n != nil && !*stateVarsCollected && strings.Contains(n.Code, "tui.NewState(") {
				*stateVarsCollected = true
				analyzer := tuigen.NewAnalyzer()
				stateVars := analyzer.DetectStateVars(comp)
				ctx.Scope.StateVars = append(ctx.Scope.StateVars, stateVars...)
			}
		case *tuigen.LetBinding:
			ctx.Scope.LetBinds = append(ctx.Scope.LetBinds, n)
			if n.Element != nil {
				collectScopeFromBodyInner(ctx, []tuigen.Node{n.Element}, comp, stateVarsCollected)
			}
		case *tuigen.ForLoop:
			collectScopeFromBodyInner(ctx, n.Body, comp, stateVarsCollected)
		case *tuigen.IfStmt:
			collectScopeFromBodyInner(ctx, n.Then, comp, stateVarsCollected)
			collectScopeFromBodyInner(ctx, n.Else, comp, stateVarsCollected)
		case *tuigen.ComponentCall:
			collectScopeFromBodyInner(ctx, n.Children, comp, stateVarsCollected)
		}
	}
}

// --- Text helper functions ---

// getLineText returns the text of the given 0-indexed line.
func getLineText(content string, line int) string {
	currentLine := 0
	start := 0
	for i, ch := range content {
		if currentLine == line {
			start = i
			end := strings.IndexByte(content[i:], '\n')
			if end == -1 {
				return content[start:]
			}
			return content[start : start+end]
		}
		if ch == '\n' {
			currentLine++
		}
	}
	return ""
}

// getWordAtOffset extracts the word at the given byte offset.
// Includes hyphens in words (for Tailwind class names like "flex-col"),
// and includes @ or # prefixes for keywords/refs.
func getWordAtOffset(content string, offset int) string {
	if offset < 0 || offset >= len(content) {
		return ""
	}

	// isWordOrHyphen extends the existing isWordChar to also include hyphens
	// so that Tailwind classes like "flex-col" are treated as single words.
	isWordOrHyphen := func(b byte) bool {
		return isWordChar(b) || b == '-'
	}

	// Find word start
	start := offset
	for start > 0 && isWordOrHyphen(content[start-1]) {
		start--
	}
	// Include @ prefix for keywords/component calls
	if start > 0 && content[start-1] == '@' {
		start--
	}
	// Include # prefix for named refs
	if start > 0 && content[start-1] == '#' {
		start--
	}

	// Find word end
	end := offset
	for end < len(content) && isWordOrHyphen(content[end]) {
		end++
	}

	if start == end {
		return ""
	}
	return content[start:end]
}

// isOffsetInGoExpr checks if the offset is inside a Go expression ({...}).
//
// Known limitation: This is a heuristic based on brace counting. It may
// false-positive inside Go struct literals, map literals, or when braces
// appear inside string literals. This is acceptable for Phase 1 as a
// best-effort heuristic; more accurate detection would require full
// lexer-aware parsing.
func isOffsetInGoExpr(content string, offset int) bool {
	if offset <= 0 || offset >= len(content) {
		return false
	}

	// Search backwards for unmatched {
	braceDepth := 0
	for i := offset - 1; i >= 0; i-- {
		switch content[i] {
		case '{':
			if braceDepth == 0 {
				return true
			}
			braceDepth--
		case '}':
			braceDepth++
		}
	}
	return false
}

// maxClassAttrSearchDistance is the maximum number of bytes to search backwards
// when looking for a class="..." attribute opening. This should be large enough to
// handle elements with many attributes before the class attribute.
const maxClassAttrSearchDistance = 500

// isOffsetInClassAttr checks if the offset is inside a class="..." attribute value.
func isOffsetInClassAttr(content string, offset int) bool {
	if offset <= 0 || offset >= len(content) {
		return false
	}

	// Search backwards for class="
	searchStart := offset - maxClassAttrSearchDistance
	if searchStart < 0 {
		searchStart = 0
	}

	segment := content[searchStart:offset]
	classIdx := strings.LastIndex(segment, `class="`)
	if classIdx == -1 {
		return false
	}

	// Check we haven't passed the closing quote
	afterClass := segment[classIdx+7:]
	return !strings.Contains(afterClass, `"`)
}

// isOffsetInElementTag checks if the offset is inside an element tag (between < and >).
// Supports multi-line element tags where attributes span multiple lines.
func isOffsetInElementTag(content string, offset int) bool {
	if offset <= 0 || offset >= len(content) {
		return false
	}

	// Search backwards for < or >, allowing newlines (multi-line tags).
	// Limit search to avoid scanning the entire file for very large documents.
	minOffset := offset - 500
	if minOffset < 0 {
		minOffset = 0
	}
	for i := offset - 1; i >= minOffset; i-- {
		switch content[i] {
		case '<':
			return true
		case '>':
			return false
		}
	}
	return false
}
