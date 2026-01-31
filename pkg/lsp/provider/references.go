package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// referencesProvider implements ReferencesProvider.
type referencesProvider struct {
	index     ComponentIndex
	docs      DocumentAccessor
	workspace WorkspaceASTAccessor
}

// NewReferencesProvider creates a new references provider.
func NewReferencesProvider(index ComponentIndex, docs DocumentAccessor, workspace WorkspaceASTAccessor) ReferencesProvider {
	return &referencesProvider{
		index:     index,
		docs:      docs,
		workspace: workspace,
	}
}

func (r *referencesProvider) References(ctx *CursorContext, includeDecl bool) ([]Location, error) {
	log.Server("References provider: NodeKind=%s, Word=%q", ctx.NodeKind, ctx.Word)

	word := ctx.Word
	if word == "" {
		return []Location{}, nil
	}

	// Dispatch based on what we can identify
	componentName := strings.TrimPrefix(word, "@")

	// Check if this is a component
	if _, ok := r.index.Lookup(componentName); ok {
		return r.findComponentReferences(componentName, includeDecl), nil
	}

	// Check if this is a function
	if _, ok := r.index.LookupFunc(word); ok {
		return r.findFunctionReferences(word, includeDecl), nil
	}

	// Check within component scope
	if ctx.Scope != nil && ctx.Scope.Component != nil {
		compName := ctx.Scope.Component.Name

		// Named ref — check if the word matches a named ref in scope
		refWord := strings.TrimPrefix(word, "#")
		for _, ref := range ctx.Scope.NamedRefs {
			if ref.Name == refWord {
				return r.findNamedRefReferences(ctx, refWord, includeDecl), nil
			}
		}

		// State variable — check if the word matches a state var in scope
		for _, sv := range ctx.Scope.StateVars {
			if sv.Name == word {
				return r.findStateVarReferences(ctx, word, includeDecl), nil
			}
		}

		// Parameter
		if _, ok := r.index.LookupParam(compName, word); ok {
			return r.findParamReferences(ctx, word, includeDecl), nil
		}

		// Let binding
		refs := r.findLocalVariableReferences(ctx, word, includeDecl)
		if len(refs) > 0 {
			return refs, nil
		}

		// For loop variable
		refs = r.findLoopVariableReferences(ctx, word, includeDecl)
		if len(refs) > 0 {
			return refs, nil
		}

		// GoCode variable
		refs = r.findGoCodeVariableReferences(ctx, word, includeDecl)
		if len(refs) > 0 {
			return refs, nil
		}
	}

	return []Location{}, nil
}

// --- Component references ---

func (r *referencesProvider) findComponentReferences(name string, includeDecl bool) []Location {
	var refs []Location

	if includeDecl {
		if info, ok := r.index.Lookup(name); ok {
			refs = append(refs, info.Location)
		}
	}

	// Search all open documents
	for _, doc := range r.docs.AllDocuments() {
		if doc.AST == nil {
			continue
		}
		for _, comp := range doc.AST.Components {
			findComponentCallsInNodes(comp.Body, name, doc.URI, &refs)
		}
	}

	return refs
}

// --- Function references ---

func (r *referencesProvider) findFunctionReferences(name string, includeDecl bool) []Location {
	var refs []Location

	if includeDecl {
		if info, ok := r.index.LookupFunc(name); ok {
			refs = append(refs, info.Location)
		}
	}

	// Search open documents
	for _, doc := range r.docs.AllDocuments() {
		if doc.AST == nil {
			continue
		}
		for _, comp := range doc.AST.Components {
			findFunctionCallsInNodes(comp.Body, name, doc.URI, &refs)
		}
	}

	// Search workspace ASTs for files not open in editor
	r.searchWorkspaceForFunctionRefs(name, &refs)

	return refs
}

// --- Parameter references ---

func (r *referencesProvider) findParamReferences(ctx *CursorContext, paramName string, includeDecl bool) []Location {
	var refs []Location

	if ctx.Document.AST == nil {
		return refs
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		if includeDecl {
			for _, p := range comp.Params {
				if p.Name == paramName {
					refs = append(refs, Location{
						URI: ctx.Document.URI,
						Range: Range{
							Start: Position{Line: p.Position.Line - 1, Character: p.Position.Column - 1},
							End:   Position{Line: p.Position.Line - 1, Character: p.Position.Column - 1 + len(paramName)},
						},
					})
					break
				}
			}
		}

		findVariableUsagesInNodes(comp.Body, paramName, ctx.Document.URI, &refs)
		break
	}

	return refs
}

// --- Local variable references ---

func (r *referencesProvider) findLocalVariableReferences(ctx *CursorContext, varName string, includeDecl bool) []Location {
	var refs []Location

	if ctx.Document.AST == nil {
		return refs
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		letBinding := findLetBindingInNodes(comp.Body, varName)
		if letBinding == nil {
			return refs
		}

		if includeDecl {
			refs = append(refs, Location{
				URI: ctx.Document.URI,
				Range: Range{
					Start: Position{Line: letBinding.Position.Line - 1, Character: letBinding.Position.Column - 1 + len("@let ")},
					End:   Position{Line: letBinding.Position.Line - 1, Character: letBinding.Position.Column - 1 + len("@let ") + len(varName)},
				},
			})
		}

		findVariableUsagesInNodes(comp.Body, varName, ctx.Document.URI, &refs)
		break
	}

	return refs
}

// --- Loop variable references ---

func (r *referencesProvider) findLoopVariableReferences(ctx *CursorContext, varName string, includeDecl bool) []Location {
	var refs []Location

	if ctx.Document.AST == nil {
		return refs
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		loop := findForLoopWithVariable(comp.Body, varName)
		if loop == nil {
			return refs
		}

		// Calculate declaration position
		declLine := loop.Position.Line - 1
		var declCharStart, declCharEnd int
		if loop.Index == varName {
			declCharStart = loop.Position.Column - 1 + len("@for ")
			declCharEnd = declCharStart + len(varName)
		} else if loop.Value == varName {
			offset := len("@for ")
			if loop.Index != "" {
				offset += len(loop.Index) + 2
			}
			declCharStart = loop.Position.Column - 1 + offset
			declCharEnd = declCharStart + len(varName)
		}

		cursorOnDecl := ctx.Position.Line == declLine &&
			ctx.Position.Character >= declCharStart &&
			ctx.Position.Character <= declCharEnd

		if includeDecl && !cursorOnDecl {
			refs = append(refs, Location{
				URI: ctx.Document.URI,
				Range: Range{
					Start: Position{Line: declLine, Character: declCharStart},
					End:   Position{Line: declLine, Character: declCharEnd},
				},
			})
		}

		// Find usages in loop body only
		findVariableUsagesInNodes(loop.Body, varName, ctx.Document.URI, &refs)
		break
	}

	return refs
}

// --- GoCode variable references ---

func (r *referencesProvider) findGoCodeVariableReferences(ctx *CursorContext, varName string, includeDecl bool) []Location {
	var refs []Location

	if ctx.Document.AST == nil {
		return refs
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		goCode := findGoCodeWithVariable(comp.Body, varName)
		if goCode == nil {
			return refs
		}

		var declLine, declCharStart, declCharEnd int
		declIdx := findVarDeclPosition(goCode.Code, varName)
		if declIdx >= 0 {
			declLine = goCode.Position.Line - 1
			declCharStart = goCode.Position.Column - 1 + declIdx
			declCharEnd = declCharStart + len(varName)

			if includeDecl {
				refs = append(refs, Location{
					URI: ctx.Document.URI,
					Range: Range{
						Start: Position{Line: declLine, Character: declCharStart},
						End:   Position{Line: declLine, Character: declCharEnd},
					},
				})
			}
		}

		findVariableUsagesInNodesExcluding(comp.Body, varName, ctx.Document.URI, declLine, declCharStart, declCharEnd, &refs)
		break
	}

	return refs
}

// --- Named ref references ---

func (r *referencesProvider) findNamedRefReferences(ctx *CursorContext, refName string, includeDecl bool) []Location {
	var refs []Location

	if ctx.Document.AST == nil {
		return refs
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		// Find the #Name declaration on the element
		if includeDecl {
			findNamedRefDeclInNodes(comp.Body, refName, ctx.Document.URI, &refs)
		}

		// Find all usages of the ref name in Go expressions and handler arguments
		findVariableUsagesInNodes(comp.Body, refName, ctx.Document.URI, &refs)
		break
	}

	return refs
}

// findNamedRefDeclInNodes finds the element with #Name declaration.
func findNamedRefDeclInNodes(nodes []tuigen.Node, refName string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.Element:
			if n != nil && n.NamedRef == refName {
				*refs = append(*refs, Location{
					URI: uri,
					Range: Range{
						Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1},
						End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + len(n.Tag)},
					},
				})
			}
			if n != nil {
				findNamedRefDeclInNodes(n.Children, refName, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				findNamedRefDeclInNodes(n.Body, refName, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findNamedRefDeclInNodes(n.Then, refName, uri, refs)
				findNamedRefDeclInNodes(n.Else, refName, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findNamedRefDeclInNodes([]tuigen.Node{n.Element}, refName, uri, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				findNamedRefDeclInNodes(n.Children, refName, uri, refs)
			}
		}
	}
}

// --- State variable references ---

func (r *referencesProvider) findStateVarReferences(ctx *CursorContext, varName string, includeDecl bool) []Location {
	var refs []Location

	if ctx.Document.AST == nil {
		return refs
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		// Find the declaration (tui.NewState line)
		if includeDecl {
			findStateVarDeclInNodes(comp.Body, varName, ctx.Document.URI, &refs)
		}

		// Find all usages (.Get(), .Set(), handler arguments, etc.)
		findVariableUsagesInNodes(comp.Body, varName, ctx.Document.URI, &refs)
		break
	}

	return refs
}

// findStateVarDeclInNodes finds the GoCode node that declares the state variable.
func findStateVarDeclInNodes(nodes []tuigen.Node, varName string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoCode:
			if n != nil && strings.Contains(n.Code, varName) && strings.Contains(n.Code, "tui.NewState(") {
				idx := strings.Index(n.Code, varName)
				if idx >= 0 {
					*refs = append(*refs, Location{
						URI: uri,
						Range: Range{
							Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + idx},
							End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + idx + len(varName)},
						},
					})
				}
			}
		case *tuigen.Element:
			if n != nil {
				findStateVarDeclInNodes(n.Children, varName, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				findStateVarDeclInNodes(n.Body, varName, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findStateVarDeclInNodes(n.Then, varName, uri, refs)
				findStateVarDeclInNodes(n.Else, varName, uri, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				findStateVarDeclInNodes(n.Children, varName, uri, refs)
			}
		}
	}
}

// --- Workspace search ---

func (r *referencesProvider) searchWorkspaceForFunctionRefs(name string, refs *[]Location) {
	for uri, ast := range r.workspace.AllWorkspaceASTs() {
		// Skip if file is already open (already searched above)
		if r.docs.GetDocument(uri) != nil {
			continue
		}
		if ast == nil {
			continue
		}
		for _, comp := range ast.Components {
			findFunctionCallsInNodes(comp.Body, name, uri, refs)
		}
	}
}

// --- AST traversal helpers ---

// findComponentCallsInNodes finds component calls recursively.
func findComponentCallsInNodes(nodes []tuigen.Node, name string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.ComponentCall:
			if n != nil && n.Name == name {
				*refs = append(*refs, Location{
					URI: uri,
					Range: Range{
						Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1},
						End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + len("@") + len(name)},
					},
				})
			}
			if n != nil {
				findComponentCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.Element:
			if n != nil {
				findComponentCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				findComponentCallsInNodes(n.Body, name, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findComponentCallsInNodes(n.Then, name, uri, refs)
				findComponentCallsInNodes(n.Else, name, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findComponentCallsInNodes(n.Element.Children, name, uri, refs)
			}
		}
	}
}

// findFunctionCallsInNodes finds function calls in Go expressions.
func findFunctionCallsInNodes(nodes []tuigen.Node, name string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoCode:
			if n != nil && strings.Contains(n.Code, name+"(") {
				idx := strings.Index(n.Code, name)
				if idx >= 0 {
					*refs = append(*refs, Location{
						URI: uri,
						Range: Range{
							Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + idx},
							End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + idx + len(name)},
						},
					})
				}
			}
		case *tuigen.GoExpr:
			if n != nil && strings.Contains(n.Code, name+"(") {
				idx := strings.Index(n.Code, name)
				if idx >= 0 {
					*refs = append(*refs, Location{
						URI: uri,
						Range: Range{
							Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column + idx},
							End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column + idx + len(name)},
						},
					})
				}
			}
		case *tuigen.Element:
			if n != nil {
				for _, attr := range n.Attributes {
					if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
						if strings.Contains(expr.Code, name+"(") {
							idx := strings.Index(expr.Code, name)
							if idx >= 0 {
								*refs = append(*refs, Location{
									URI: uri,
									Range: Range{
										Start: Position{Line: expr.Position.Line - 1, Character: expr.Position.Column + idx},
										End:   Position{Line: expr.Position.Line - 1, Character: expr.Position.Column + idx + len(name)},
									},
								})
							}
						}
					}
				}
				findFunctionCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				findFunctionCallsInNodes(n.Body, name, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findFunctionCallsInNodes(n.Then, name, uri, refs)
				findFunctionCallsInNodes(n.Else, name, uri, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				if strings.Contains(n.Args, name+"(") {
					idx := strings.Index(n.Args, name)
					if idx >= 0 {
						*refs = append(*refs, Location{
							URI: uri,
							Range: Range{
								Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column + len("@") + len(n.Name) + 1 + idx},
								End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column + len("@") + len(n.Name) + 1 + idx + len(name)},
							},
						})
					}
				}
				findFunctionCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findFunctionCallsInNodes(n.Element.Children, name, uri, refs)
			}
		}
	}
}

// findVariableUsagesInNodes finds usages of a variable in AST nodes.
func findVariableUsagesInNodes(nodes []tuigen.Node, varName string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoExpr:
			if n != nil {
				findVariableInCode(n.Code, varName, n.Position, 1, uri, refs)
			}
		case *tuigen.RawGoExpr:
			if n != nil {
				findVariableInCode(n.Code, varName, n.Position, 1, uri, refs)
			}
		case *tuigen.GoCode:
			if n != nil {
				findVariableInCode(n.Code, varName, n.Position, 0, uri, refs)
			}
		case *tuigen.Element:
			if n != nil {
				for _, attr := range n.Attributes {
					if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
						findVariableInCode(expr.Code, varName, expr.Position, 1, uri, refs)
					}
				}
				findVariableUsagesInNodes(n.Children, varName, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				iterableOffset := len("@for ") + len(n.Index) + len(", ") + len(n.Value) + len(" := range ")
				findVariableInCode(n.Iterable, varName, n.Position, iterableOffset, uri, refs)
				findVariableUsagesInNodes(n.Body, varName, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findVariableInCode(n.Condition, varName, n.Position, len("@if "), uri, refs)
				findVariableUsagesInNodes(n.Then, varName, uri, refs)
				findVariableUsagesInNodes(n.Else, varName, uri, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				argsOffset := len("@") + len(n.Name) + 1
				findVariableInCode(n.Args, varName, n.Position, argsOffset, uri, refs)
				findVariableUsagesInNodes(n.Children, varName, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findVariableUsagesInNodes(n.Element.Children, varName, uri, refs)
			}
		}
	}
}

// findVariableUsagesInNodesExcluding finds usages excluding a specific location.
func findVariableUsagesInNodesExcluding(nodes []tuigen.Node, varName string, uri string, exclLine, exclCharStart, exclCharEnd int, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoExpr:
			if n != nil {
				findVariableInCodeExcluding(n.Code, varName, n.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.RawGoExpr:
			if n != nil {
				findVariableInCodeExcluding(n.Code, varName, n.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.GoCode:
			if n != nil {
				findVariableInCodeExcluding(n.Code, varName, n.Position, 0, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.Element:
			if n != nil {
				for _, attr := range n.Attributes {
					if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
						findVariableInCodeExcluding(expr.Code, varName, expr.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
					}
				}
				findVariableUsagesInNodesExcluding(n.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				iterableOffset := len("@for ") + len(n.Index) + len(", ") + len(n.Value) + len(" := range ")
				findVariableInCodeExcluding(n.Iterable, varName, n.Position, iterableOffset, uri, exclLine, exclCharStart, exclCharEnd, refs)
				findVariableUsagesInNodesExcluding(n.Body, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findVariableInCodeExcluding(n.Condition, varName, n.Position, len("@if "), uri, exclLine, exclCharStart, exclCharEnd, refs)
				findVariableUsagesInNodesExcluding(n.Then, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
				findVariableUsagesInNodesExcluding(n.Else, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				argsOffset := len("@") + len(n.Name) + 1
				findVariableInCodeExcluding(n.Args, varName, n.Position, argsOffset, uri, exclLine, exclCharStart, exclCharEnd, refs)
				findVariableUsagesInNodesExcluding(n.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findVariableUsagesInNodesExcluding(n.Element.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		}
	}
}

// findVariableInCode finds variable occurrences with a custom offset.
func findVariableInCode(code, varName string, pos tuigen.Position, startOffset int, uri string, refs *[]Location) {
	findVariableInCodeExcluding(code, varName, pos, startOffset, uri, -1, -1, -1, refs)
}

// findVariableInCodeExcluding finds variable occurrences, excluding a specific location.
func findVariableInCodeExcluding(code, varName string, pos tuigen.Position, startOffset int, uri string, exclLine, exclCharStart, exclCharEnd int, refs *[]Location) {
	idx := 0
	for {
		i := strings.Index(code[idx:], varName)
		if i < 0 {
			break
		}
		absIdx := idx + i

		before := absIdx == 0 || !IsWordChar(code[absIdx-1])
		after := absIdx+len(varName) >= len(code) || !IsWordChar(code[absIdx+len(varName)])

		if before && after {
			charPos := pos.Column - 1 + startOffset + absIdx
			line := pos.Line - 1

			if line == exclLine && charPos == exclCharStart && charPos+len(varName) == exclCharEnd {
				idx = absIdx + len(varName)
				continue
			}

			*refs = append(*refs, Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: line, Character: charPos},
					End:   Position{Line: line, Character: charPos + len(varName)},
				},
			})
		}

		idx = absIdx + len(varName)
	}
}
