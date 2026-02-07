package lsp

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/schema"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// resolveFromAST walks the parsed AST to find the node at the cursor position
// and populate scope information.
func resolveFromAST(ctx *CursorContext, file *tuigen.File) {
	// Convert LSP 0-indexed position to tuigen 1-indexed
	line := ctx.Position.Line + 1
	col := ctx.Position.Character + 1

	// Check if cursor is on an import line.
	for i := range file.Imports {
		imp := &file.Imports[i]
		if imp.Position.Line != line {
			continue
		}
		// Cursor is on this import's line. Check if it's within the
		// import content (alias or path). Search line text for the quoted
		// path to handle both aliased and non-aliased imports uniformly.
		quotedPath := `"` + imp.Path + `"`
		pathIdx := strings.Index(ctx.Line, quotedPath)
		if pathIdx < 0 {
			continue
		}
		// Determine the start of actionable content on this line.
		// For aliased imports the alias precedes the path.
		contentStart := pathIdx
		if imp.Alias != "" {
			aliasIdx := strings.Index(ctx.Line, imp.Alias)
			if aliasIdx >= 0 && aliasIdx < pathIdx {
				contentStart = aliasIdx
			}
		}
		pathEnd := pathIdx + len(quotedPath)
		if ctx.Position.Character >= contentStart && ctx.Position.Character < pathEnd {
			ctx.NodeKind = NodeKindImportPath
			ctx.Node = imp
			ctx.ImportPath = imp.Path
			ctx.Word = imp.Path
			return
		}
	}

	// Check if cursor is on a top-level declaration (var, type, const).
	// These are passed to gopls for type resolution.
	for _, decl := range file.Decls {
		declLines := strings.Count(decl.Code, "\n") + 1
		declEndLine := decl.Position.Line + declLines - 1
		if line >= decl.Position.Line && line <= declEndLine {
			ctx.Node = decl
			ctx.NodeKind = NodeKindGoDecl
			return
		}
	}

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

			// Cursor is on the declaration line but not on the name or a param name
			// (e.g., on a parameter type). Set NodeKindComponent so hover/definition
			// providers can delegate to gopls for type resolution.
			ctx.Node = comp
			ctx.NodeKind = NodeKindComponent
			ctx.Scope.Component = comp
			ctx.Scope.Params = comp.Params
			collectScopeFromBody(ctx, comp.Body, comp)
			return
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

	// Verify cursor is actually inside the component body (#20)
	if enclosingComp != nil {
		endLine := findComponentEndLine(ctx.Document.Content, enclosingComp)
		if line-1 > endLine { // line is 1-indexed, endLine is 0-indexed
			enclosingComp = nil
		}
	}

	if enclosingComp != nil {
		ctx.Scope.Component = enclosingComp
		ctx.Scope.Params = enclosingComp.Params
		ctx.ParentChain = append(ctx.ParentChain, enclosingComp)
		collectScopeFromBody(ctx, enclosingComp.Body, enclosingComp)

		if found := resolveInNodes(ctx, enclosingComp.Body, line, col); found {
			return
		}
		ctx.ParentChain = ctx.ParentChain[:len(ctx.ParentChain)-1]
	}

	// Check if cursor is on a function declaration line
	for _, fn := range file.Funcs {
		if line == fn.Position.Line {
			// Check if cursor is on a parameter name
			if paramName := findFuncParamAtColumn(fn, col); paramName != "" {
				ctx.Node = fn
				ctx.NodeKind = NodeKindParameter
				ctx.Scope.Function = fn
				return
			}
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
// Manages the ParentChain: pushes the node before checking, pops if not found.
func resolveInNode(ctx *CursorContext, node tuigen.Node, line, col int) bool {
	ctx.ParentChain = append(ctx.ParentChain, node)
	found := resolveInNodeInner(ctx, node, line, col)
	if !found {
		ctx.ParentChain = ctx.ParentChain[:len(ctx.ParentChain)-1]
	}
	return found
}

func resolveInNodeInner(ctx *CursorContext, node tuigen.Node, line, col int) bool {
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
			// For single-line expressions, also verify column range (#26)
			if !strings.Contains(n.Code, "\n") {
				start := n.Position.Column
				end := start + len(n.Code)
				if col < start || col > end {
					return false
				}
			}
			ctx.Node = n
			ctx.NodeKind = classifyGoExpr(n)
			return true
		}
	case *tuigen.GoCode:
		if n != nil && n.Position.Line == line {
			// For single-line code blocks, also verify column range (#26)
			if !strings.Contains(n.Code, "\n") {
				start := n.Position.Column
				end := start + len(n.Code)
				if col < start || col > end {
					return false
				}
			}
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

	// Check if cursor is on the element's tag name (always on the opening tag line)
	if pos.Line == line {
		tagStart := pos.Column
		tagEnd := tagStart + len(elem.Tag)
		if col >= tagStart && col <= tagEnd {
			ctx.Node = elem
			ctx.NodeKind = NodeKindElement
			return true
		}
	}

	// Check ref={} attribute — detect cursor on ref attribute name or value.
	// The ref attribute is stored in elem.RefExpr (extracted from attributes by the analyzer).
	if elem.RefExpr != nil && ctx.InElement {
		// Look for ref={...} in the source line
		refIdx := strings.Index(ctx.Line, "ref={")
		if refIdx >= 0 {
			cursorCol := ctx.Position.Character
			// ref={expr} — cover from "ref" through the closing "}"
			refEnd := refIdx + len("ref={") + len(elem.RefExpr.Code) + 1 // +1 for }
			if cursorCol >= refIdx && cursorCol <= refEnd {
				ctx.Node = elem
				ctx.NodeKind = NodeKindRefAttr
				return true
			}
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

// classifyGoExpr determines the NodeKind for a GoExpr node.
// Detects state access patterns (.Get(), .Set(), etc.).
func classifyGoExpr(expr *tuigen.GoExpr) NodeKind {
	if expr == nil {
		return NodeKindGoExpr
	}
	trimmed := strings.TrimSpace(expr.Code)
	if strings.Contains(trimmed, ".Get()") ||
		strings.Contains(trimmed, ".Set(") ||
		strings.Contains(trimmed, ".Update(") ||
		strings.Contains(trimmed, ".Bind(") ||
		strings.Contains(trimmed, ".Batch(") {
		return NodeKindStateAccess
	}
	return NodeKindGoExpr
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
			if n.RefExpr != nil {
				ref := tuigen.RefInfo{
					Name:    n.RefExpr.Code,
					Element: n,
				}
				// Capitalize first letter for export name
				if len(ref.Name) > 0 {
					ref.ExportName = strings.ToUpper(ref.Name[:1]) + ref.Name[1:]
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
				// Determine ref kind
				if ref.InLoop {
					if ref.KeyExpr != "" {
						ref.RefKind = tuigen.RefMap
					} else {
						ref.RefKind = tuigen.RefList
					}
				} else {
					ref.RefKind = tuigen.RefSingle
				}
				ref.Position = n.RefExpr.Position
				ctx.Scope.Refs = append(ctx.Scope.Refs, ref)
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
			prevLoop := ctx.Scope.ForLoop
			ctx.Scope.ForLoop = n
			collectScopeFromBodyInner(ctx, n.Body, comp, stateVarsCollected)
			ctx.Scope.ForLoop = prevLoop
		case *tuigen.IfStmt:
			prevIf := ctx.Scope.IfStmt
			ctx.Scope.IfStmt = n
			collectScopeFromBodyInner(ctx, n.Then, comp, stateVarsCollected)
			collectScopeFromBodyInner(ctx, n.Else, comp, stateVarsCollected)
			ctx.Scope.IfStmt = prevIf
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
	// Note: # prefix no longer needed (refs now use ref={} syntax)

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

// findComponentEndLine finds the 0-indexed line number of the closing '}' for a component.
// Uses brace counting from the component declaration line.
func findComponentEndLine(content string, comp *tuigen.Component) int {
	lines := strings.Split(content, "\n")
	startLine := comp.Position.Line - 1 // convert to 0-indexed
	depth := 0
	for i := startLine; i < len(lines); i++ {
		for _, ch := range lines[i] {
			if ch == '{' {
				depth++
			} else if ch == '}' {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}
	return len(lines) - 1
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

// findFuncParamAtColumn checks if the cursor column (1-indexed) is on a parameter
// name in a function declaration. Returns the parameter name if found, empty string otherwise.
func findFuncParamAtColumn(fn *tuigen.GoFunc, col int) string {
	code := fn.Code
	if !strings.HasPrefix(strings.TrimSpace(code), "func ") {
		return ""
	}

	parenIdx := strings.Index(code, "(")
	if parenIdx < 0 {
		return ""
	}

	// Find matching close paren (depth-aware for nested parens in types)
	depth := 0
	closeIdx := -1
	for i := parenIdx; i < len(code); i++ {
		switch code[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				closeIdx = i
			}
		}
		if closeIdx >= 0 {
			break
		}
	}
	if closeIdx < 0 {
		return ""
	}

	paramStr := code[parenIdx+1 : closeIdx]
	// Column where param content starts (1-indexed, matching col)
	paramStartCol := fn.Position.Column + parenIdx + 1

	// Split params at top level (depth-aware for nested parens/brackets in types)
	depth = 0
	paramBegin := 0
	for i := 0; i <= len(paramStr); i++ {
		if i < len(paramStr) {
			switch paramStr[i] {
			case '(', '[':
				depth++
			case ')', ']':
				depth--
			}
		}

		if (i == len(paramStr)) || (paramStr[i] == ',' && depth == 0) {
			param := paramStr[paramBegin:i]
			trimmed := strings.TrimSpace(param)
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				paramName := fields[0]
				// Find name position within the raw param substring
				nameInParam := strings.Index(param, paramName)
				nameCol := paramStartCol + paramBegin + nameInParam
				if col >= nameCol && col < nameCol+len(paramName) {
					return paramName
				}
			}
			paramBegin = i + 1
		}
	}

	return ""
}
