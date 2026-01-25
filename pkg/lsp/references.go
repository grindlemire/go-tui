package lsp

import (
	"encoding/json"
	"strings"

	"github.com/grindlemire/go-tui/pkg/tuigen"
)

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

// handleReferences handles textDocument/references requests.
func (s *Server) handleReferences(params json.RawMessage) (any, *Error) {
	var p ReferenceParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	s.log("References request at %s:%d:%d", p.TextDocument.URI, p.Position.Line, p.Position.Character)

	doc := s.docs.Get(p.TextDocument.URI)
	if doc == nil {
		return []Location{}, nil
	}

	// Find what's at the cursor position
	word := s.getWordAtPosition(doc, p.Position)
	if word == "" {
		return []Location{}, nil
	}

	s.log("Finding references for: %s", word)

	var refs []Location

	// Check if this is a component - find all usages
	componentName := strings.TrimPrefix(word, "@")
	if _, ok := s.index.Lookup(componentName); ok {
		refs = s.findComponentReferences(componentName, p.Context.IncludeDeclaration)
		return refs, nil
	}

	// Check if this is a function - find all usages
	if _, ok := s.index.LookupFunc(word); ok {
		refs = s.findFunctionReferences(word, p.Context.IncludeDeclaration)
		return refs, nil
	}

	// Check if this is a parameter - find usages within the component
	if componentCtx := s.findComponentAtPosition(doc, p.Position); componentCtx != "" {
		if _, ok := s.index.LookupParam(componentCtx, word); ok {
			refs = s.findParamReferences(doc, componentCtx, word, p.Context.IncludeDeclaration)
			return refs, nil
		}
	}

	// Check if this is a local variable (@let binding)
	if componentCtx := s.findComponentAtPosition(doc, p.Position); componentCtx != "" {
		refs = s.findLocalVariableReferences(doc, componentCtx, word, p.Position, p.Context.IncludeDeclaration)
		if len(refs) > 0 {
			return refs, nil
		}

		// Check if this is a for loop variable (e.g., @for i, item := range items)
		refs = s.findLoopVariableReferences(doc, componentCtx, word, p.Position, p.Context.IncludeDeclaration)
		if len(refs) > 0 {
			return refs, nil
		}

		// Check if this is a GoCode variable (e.g., x := 1)
		refs = s.findGoCodeVariableReferences(doc, componentCtx, word, p.Context.IncludeDeclaration)
		if len(refs) > 0 {
			return refs, nil
		}
	}

	return refs, nil
}

// findComponentReferences finds all references to a component.
func (s *Server) findComponentReferences(name string, includeDeclaration bool) []Location {
	var refs []Location

	// Include declaration if requested
	if includeDeclaration {
		if info, ok := s.index.Lookup(name); ok {
			refs = append(refs, info.Location)
		}
	}

	// Search all open documents for usages
	for _, doc := range s.docs.All() {
		if doc.AST == nil {
			continue
		}
		for _, comp := range doc.AST.Components {
			s.findComponentCallsInNodes(comp.Body, name, doc.URI, &refs)
		}
	}

	return refs
}

// findComponentCallsInNodes recursively finds component calls in nodes.
func (s *Server) findComponentCallsInNodes(nodes []tuigen.Node, name string, uri string, refs *[]Location) {
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
				s.findComponentCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.Element:
			if n != nil {
				s.findComponentCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				s.findComponentCallsInNodes(n.Body, name, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				s.findComponentCallsInNodes(n.Then, name, uri, refs)
				s.findComponentCallsInNodes(n.Else, name, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				s.findComponentCallsInNodes(n.Element.Children, name, uri, refs)
			}
		}
	}
}

// findFunctionReferences finds all references to a function.
func (s *Server) findFunctionReferences(name string, includeDeclaration bool) []Location {
	var refs []Location

	// Include declaration if requested
	if includeDeclaration {
		if info, ok := s.index.LookupFunc(name); ok {
			refs = append(refs, info.Location)
		}
	}

	// Search all open documents for usages in Go expressions
	for _, doc := range s.docs.All() {
		if doc.AST == nil {
			continue
		}
		for _, comp := range doc.AST.Components {
			s.findFunctionCallsInNodes(comp.Body, name, doc.URI, doc.Content, &refs)
		}
	}

	// Also search workspace files that aren't open
	s.searchWorkspaceForFunctionRefs(name, &refs)

	return refs
}

// findFunctionCallsInNodes recursively finds function calls in nodes.
func (s *Server) findFunctionCallsInNodes(nodes []tuigen.Node, name string, uri string, content string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoCode:
			// Standalone Go statements like: myFunction("test")
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
				// Find the position of the function call within the expression
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
				// Check attributes for function calls
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
				s.findFunctionCallsInNodes(n.Children, name, uri, content, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				s.findFunctionCallsInNodes(n.Body, name, uri, content, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				s.findFunctionCallsInNodes(n.Then, name, uri, content, refs)
				s.findFunctionCallsInNodes(n.Else, name, uri, content, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				// Check args for function calls
				if strings.Contains(n.Args, name+"(") {
					idx := strings.Index(n.Args, name)
					if idx >= 0 {
						// Approximate position
						*refs = append(*refs, Location{
							URI: uri,
							Range: Range{
								Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column + len("@") + len(n.Name) + 1 + idx},
								End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column + len("@") + len(n.Name) + 1 + idx + len(name)},
							},
						})
					}
				}
				s.findFunctionCallsInNodes(n.Children, name, uri, content, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				s.findFunctionCallsInNodes(n.Element.Children, name, uri, content, refs)
			}
		}
	}
}

// findParamReferences finds all references to a parameter within a component.
func (s *Server) findParamReferences(doc *Document, componentName, paramName string, includeDeclaration bool) []Location {
	var refs []Location

	if doc.AST == nil {
		return refs
	}

	// Find the component
	for _, comp := range doc.AST.Components {
		if comp.Name != componentName {
			continue
		}

		// Include declaration if requested
		if includeDeclaration {
			for _, p := range comp.Params {
				if p.Name == paramName {
					refs = append(refs, Location{
						URI: doc.URI,
						Range: Range{
							Start: Position{Line: p.Position.Line - 1, Character: p.Position.Column - 1},
							End:   Position{Line: p.Position.Line - 1, Character: p.Position.Column - 1 + len(paramName)},
						},
					})
					break
				}
			}
		}

		// Find usages in component body
		s.findVariableUsagesInNodes(comp.Body, paramName, doc.URI, &refs)
		break
	}

	return refs
}

// findLocalVariableReferences finds all references to a local variable within a component.
func (s *Server) findLocalVariableReferences(doc *Document, componentName, varName string, pos Position, includeDeclaration bool) []Location {
	var refs []Location

	if doc.AST == nil {
		return refs
	}

	// Find the component
	for _, comp := range doc.AST.Components {
		if comp.Name != componentName {
			continue
		}

		// Find @let bindings with this name
		letBinding := s.findLetBindingInNodes(comp.Body, varName)
		if letBinding == nil {
			return refs
		}

		// Include declaration if requested
		if includeDeclaration {
			refs = append(refs, Location{
				URI: doc.URI,
				Range: Range{
					Start: Position{Line: letBinding.Position.Line - 1, Character: letBinding.Position.Column - 1 + len("@let ")},
					End:   Position{Line: letBinding.Position.Line - 1, Character: letBinding.Position.Column - 1 + len("@let ") + len(varName)},
				},
			})
		}

		// Find usages in component body (after the let binding)
		s.findVariableUsagesInNodes(comp.Body, varName, doc.URI, &refs)
		break
	}

	return refs
}

// findLetBindingInNodes finds a @let binding by name.
func (s *Server) findLetBindingInNodes(nodes []tuigen.Node, name string) *tuigen.LetBinding {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.LetBinding:
			if n != nil && n.Name == name {
				return n
			}
		case *tuigen.Element:
			if n != nil {
				if found := s.findLetBindingInNodes(n.Children, name); found != nil {
					return found
				}
			}
		case *tuigen.ForLoop:
			if n != nil {
				if found := s.findLetBindingInNodes(n.Body, name); found != nil {
					return found
				}
			}
		case *tuigen.IfStmt:
			if n != nil {
				if found := s.findLetBindingInNodes(n.Then, name); found != nil {
					return found
				}
				if found := s.findLetBindingInNodes(n.Else, name); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// findVariableUsagesInNodesExcluding finds all usages of a variable in nodes, excluding a specific location.
func (s *Server) findVariableUsagesInNodesExcluding(nodes []tuigen.Node, varName string, uri string, exclLine, exclCharStart, exclCharEnd int, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoExpr:
			if n != nil {
				s.findVariableInCodeWithOffsetExcluding(n.Code, varName, n.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.RawGoExpr:
			if n != nil {
				s.findVariableInCodeWithOffsetExcluding(n.Code, varName, n.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.GoCode:
			if n != nil {
				// Use offset 0 because GoCode position points to first char (no braces)
				s.findVariableInCodeWithOffsetExcluding(n.Code, varName, n.Position, 0, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.Element:
			if n != nil {
				for _, attr := range n.Attributes {
					if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
						s.findVariableInCodeWithOffsetExcluding(expr.Code, varName, expr.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
					}
				}
				s.findVariableUsagesInNodesExcluding(n.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				// Check iterable - offset is "@for index, value := range " to get to the iterable
				iterableOffset := len("@for ") + len(n.Index) + len(", ") + len(n.Value) + len(" := range ")
				s.findVariableInCodeWithOffsetExcluding(n.Iterable, varName, n.Position, iterableOffset, uri, exclLine, exclCharStart, exclCharEnd, refs)
				s.findVariableUsagesInNodesExcluding(n.Body, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				// Check condition - offset is len("@if ") = 4 to skip past the keyword
				s.findVariableInCodeWithOffsetExcluding(n.Condition, varName, n.Position, len("@if "), uri, exclLine, exclCharStart, exclCharEnd, refs)
				s.findVariableUsagesInNodesExcluding(n.Then, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
				s.findVariableUsagesInNodesExcluding(n.Else, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				// Check args - offset is "@Name(" to get to the args
				argsOffset := len("@") + len(n.Name) + 1 // +1 for (
				s.findVariableInCodeWithOffsetExcluding(n.Args, varName, n.Position, argsOffset, uri, exclLine, exclCharStart, exclCharEnd, refs)
				s.findVariableUsagesInNodesExcluding(n.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.LetBinding:
			if n != nil {
				// Check the expression part (after =)
				// The expression is captured somewhere, need to find it
				if n.Element != nil {
					s.findVariableUsagesInNodesExcluding(n.Element.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
				}
			}
		}
	}
}

// findVariableUsagesInNodes finds all usages of a variable in nodes.
func (s *Server) findVariableUsagesInNodes(nodes []tuigen.Node, varName string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoExpr:
			if n != nil {
				s.findVariableInCode(n.Code, varName, n.Position, uri, refs)
			}
		case *tuigen.RawGoExpr:
			if n != nil {
				s.findVariableInCode(n.Code, varName, n.Position, uri, refs)
			}
		case *tuigen.GoCode:
			if n != nil {
				// Use offset 0 because GoCode position points to first char (no braces)
				s.findVariableInCodeWithOffset(n.Code, varName, n.Position, 0, uri, refs)
			}
		case *tuigen.Element:
			if n != nil {
				for _, attr := range n.Attributes {
					if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
						s.findVariableInCode(expr.Code, varName, expr.Position, uri, refs)
					}
				}
				s.findVariableUsagesInNodes(n.Children, varName, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				// Check iterable - offset is "@for index, value := range " to get to the iterable
				iterableOffset := len("@for ") + len(n.Index) + len(", ") + len(n.Value) + len(" := range ")
				s.findVariableInCodeWithOffset(n.Iterable, varName, n.Position, iterableOffset, uri, refs)
				s.findVariableUsagesInNodes(n.Body, varName, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				// Check condition - offset is len("@if ") = 4 to skip past the keyword
				s.findVariableInCodeWithOffset(n.Condition, varName, n.Position, len("@if "), uri, refs)
				s.findVariableUsagesInNodes(n.Then, varName, uri, refs)
				s.findVariableUsagesInNodes(n.Else, varName, uri, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				// Check args - offset is "@Name(" to get to the args
				argsOffset := len("@") + len(n.Name) + 1 // +1 for (
				s.findVariableInCodeWithOffset(n.Args, varName, n.Position, argsOffset, uri, refs)
				s.findVariableUsagesInNodes(n.Children, varName, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil {
				// Check the expression part (after =)
				// The expression is captured somewhere, need to find it
				if n.Element != nil {
					s.findVariableUsagesInNodes(n.Element.Children, varName, uri, refs)
				}
			}
		}
	}
}

// findVariableInCode finds occurrences of a variable in Go code.
// The offset parameter adjusts for leading characters (1 for GoExpr inside {}, 0 for GoCode).
func (s *Server) findVariableInCode(code, varName string, pos tuigen.Position, uri string, refs *[]Location) {
	s.findVariableInCodeWithOffset(code, varName, pos, 1, uri, refs)
}

// findVariableInCodeWithOffset finds variable occurrences with a custom offset.
func (s *Server) findVariableInCodeWithOffset(code, varName string, pos tuigen.Position, startOffset int, uri string, refs *[]Location) {
	s.findVariableInCodeWithOffsetExcluding(code, varName, pos, startOffset, uri, -1, -1, -1, refs)
}

// findVariableInCodeWithOffsetExcluding finds variable occurrences, excluding a specific location.
func (s *Server) findVariableInCodeWithOffsetExcluding(code, varName string, pos tuigen.Position, startOffset int, uri string, exclLine, exclCharStart, exclCharEnd int, refs *[]Location) {
	// Simple token-based search for the variable
	idx := 0
	for {
		i := strings.Index(code[idx:], varName)
		if i < 0 {
			break
		}
		absIdx := idx + i

		// Check that it's a whole word (not part of another identifier)
		before := absIdx == 0 || !isWordChar(code[absIdx-1])
		after := absIdx+len(varName) >= len(code) || !isWordChar(code[absIdx+len(varName)])

		if before && after {
			// Calculate character position: pos.Column is 1-indexed, LSP wants 0-indexed
			charPos := pos.Column - 1 + startOffset + absIdx
			line := pos.Line - 1

			// Skip if this matches the excluded location (the declaration)
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

// findGoCodeVariableReferences finds all references to a variable declared in GoCode.
func (s *Server) findGoCodeVariableReferences(doc *Document, componentName, varName string, includeDeclaration bool) []Location {
	var refs []Location

	if doc.AST == nil {
		return refs
	}

	// Find the component
	for _, comp := range doc.AST.Components {
		if comp.Name != componentName {
			continue
		}

		// Find GoCode that declares this variable
		goCode := s.findGoCodeWithVariable(comp.Body, varName)
		if goCode == nil {
			return refs
		}

		// Calculate the declaration location to exclude it from usages
		var declLine, declCharStart, declCharEnd int
		declIdx := findVarDeclPosition(goCode.Code, varName)
		if declIdx >= 0 {
			declLine = goCode.Position.Line - 1
			declCharStart = goCode.Position.Column - 1 + declIdx
			declCharEnd = declCharStart + len(varName)

			// Include declaration if requested
			if includeDeclaration {
				refs = append(refs, Location{
					URI: doc.URI,
					Range: Range{
						Start: Position{Line: declLine, Character: declCharStart},
						End:   Position{Line: declLine, Character: declCharEnd},
					},
				})
			}
		}

		// Find usages in component body, excluding the declaration
		s.findVariableUsagesInNodesExcluding(comp.Body, varName, doc.URI, declLine, declCharStart, declCharEnd, &refs)
		break
	}

	return refs
}

// findLoopVariableReferences finds all references to a for loop variable within its scope.
func (s *Server) findLoopVariableReferences(doc *Document, componentName, varName string, pos Position, includeDeclaration bool) []Location {
	var refs []Location

	if doc.AST == nil {
		return refs
	}

	// Find the component
	for _, comp := range doc.AST.Components {
		if comp.Name != componentName {
			continue
		}

		// Find the for loop that declares this variable and contains the position
		loop := s.findForLoopWithVariable(comp.Body, varName, pos)
		if loop == nil {
			return refs
		}

		// Calculate the declaration position
		var declLine, declCharStart, declCharEnd int
		declLine = loop.Position.Line - 1
		if loop.Index == varName {
			declCharStart = loop.Position.Column - 1 + len("@for ")
			declCharEnd = declCharStart + len(varName)
		} else if loop.Value == varName {
			offset := len("@for ")
			if loop.Index != "" {
				offset += len(loop.Index) + 2 // ", "
			}
			declCharStart = loop.Position.Column - 1 + offset
			declCharEnd = declCharStart + len(varName)
		}

		// Check if cursor is on the declaration - if so, don't include it
		cursorOnDecl := pos.Line == declLine && pos.Character >= declCharStart && pos.Character <= declCharEnd

		// Include declaration if requested and cursor is not on it
		if includeDeclaration && !cursorOnDecl {
			refs = append(refs, Location{
				URI: doc.URI,
				Range: Range{
					Start: Position{Line: declLine, Character: declCharStart},
					End:   Position{Line: declLine, Character: declCharEnd},
				},
			})
		}

		// Find usages in the loop body only
		s.findVariableUsagesInNodes(loop.Body, varName, doc.URI, &refs)
		break
	}

	return refs
}

// searchWorkspaceForFunctionRefs searches cached workspace ASTs for function references.
func (s *Server) searchWorkspaceForFunctionRefs(name string, refs *[]Location) {
	s.workspaceASTsMu.RLock()
	defer s.workspaceASTsMu.RUnlock()

	s.log("searchWorkspaceForFunctionRefs: searching for '%s', workspace has %d cached ASTs", name, len(s.workspaceASTs))

	for uri, ast := range s.workspaceASTs {
		s.log("  checking cached file: %s", uri)
		// Skip if file is already open (already searched in findFunctionReferences)
		if s.docs.Get(uri) != nil {
			s.log("    skipping (file is open in editor)")
			continue
		}

		if ast == nil {
			s.log("    skipping (nil AST)")
			continue
		}

		s.log("    searching %d components", len(ast.Components))
		beforeCount := len(*refs)
		// Search for function calls in this file
		for _, comp := range ast.Components {
			s.findFunctionCallsInNodes(comp.Body, name, uri, "", refs)
		}
		s.log("    found %d refs in this file", len(*refs)-beforeCount)
	}
}

