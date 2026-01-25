package lsp

import (
	"encoding/json"
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/gopls"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// DefinitionParams represents textDocument/definition parameters.
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// handleDefinition handles textDocument/definition requests.
func (s *Server) handleDefinition(params json.RawMessage) (any, *Error) {
	var p DefinitionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	s.log("Definition request at %s:%d:%d", p.TextDocument.URI, p.Position.Line, p.Position.Character)

	doc := s.docs.Get(p.TextDocument.URI)
	if doc == nil {
		return nil, nil
	}

	// First, check if the word at cursor is a locally-defined function.
	// This prevents gopls from returning the generated .go file instead of .tui.
	if word := s.getWordAtPosition(doc, p.Position); word != "" {
		s.log("Looking up function '%s' in index (all functions: %v)", word, s.index.AllFunctions())
		// Check if this is a helper function defined in a .tui file
		if funcInfo, ok := s.index.LookupFunc(word); ok {
			s.log("Found local function %s at %s (before gopls)", word, funcInfo.Location.URI)
			return funcInfo.Location, nil
		}
		s.log("Function '%s' not found in index", word)

		// Check if this is a for loop variable (defined in .tui DSL, not Go)
		// Must check BEFORE gopls since gopls doesn't understand .tui for loop syntax
		if componentCtx := s.findComponentAtPosition(doc, p.Position); componentCtx != "" {
			if loc := s.findLoopVariableDefinition(doc, componentCtx, word, p.Position); loc != nil {
				s.log("Found loop variable %s in %s (before gopls)", word, componentCtx)
				return loc, nil
			}
		}
	}

	// Check if we're inside a Go expression - use gopls if available
	if s.isInGoExpression(doc, p.Position) {
		locs, err := s.getGoplsDefinition(doc, p.Position)
		if err != nil {
			s.log("gopls definition error: %v", err)
			// Fall through to TUI definition
		} else if len(locs) > 0 {
			if len(locs) == 1 {
				loc := locs[0]
				s.log("FINAL RESPONSE (single): URI=%s Range=(%d:%d)-(%d:%d)",
					loc.URI, loc.Range.Start.Line, loc.Range.Start.Character,
					loc.Range.End.Line, loc.Range.End.Character)
				return loc, nil
			}
			s.log("FINAL RESPONSE (multiple): %d locations", len(locs))
			return locs, nil
		}
	}

	// Find what's at the cursor position
	word := s.getWordAtPosition(doc, p.Position)
	if word == "" {
		return nil, nil
	}

	s.log("Word at position: %s", word)

	// Check if this is a component call (starts with @ or is a known component)
	componentName := word
	if strings.HasPrefix(word, "@") {
		componentName = strings.TrimPrefix(word, "@")
	}

	// Look up component in index
	info, ok := s.index.Lookup(componentName)
	if ok {
		s.log("Found component %s at %s", componentName, info.Location.URI)
		return info.Location, nil
	}

	// Check if this is a helper function
	if funcInfo, ok := s.index.LookupFunc(word); ok {
		s.log("Found function %s at %s", word, funcInfo.Location.URI)
		return funcInfo.Location, nil
	}

	// Check if this is a component parameter
	if componentCtx := s.findComponentAtPosition(doc, p.Position); componentCtx != "" {
		if paramInfo, ok := s.index.LookupParam(componentCtx, word); ok {
			s.log("Found param %s.%s at %s", componentCtx, word, paramInfo.Location.URI)
			return paramInfo.Location, nil
		}

		// Check if this is a local variable (@let binding)
		if loc := s.findLocalVariableDefinition(doc, componentCtx, word); loc != nil {
			s.log("Found local variable %s in %s", word, componentCtx)
			return loc, nil
		}

		// Check if this is a for loop variable
		if loc := s.findLoopVariableDefinition(doc, componentCtx, word, p.Position); loc != nil {
			s.log("Found loop variable %s in %s", word, componentCtx)
			return loc, nil
		}

		// Check if this is a GoCode variable (e.g., x := 1)
		if loc := s.findGoCodeVariableDefinition(doc, componentCtx, word); loc != nil {
			s.log("Found GoCode variable %s in %s", word, componentCtx)
			return loc, nil
		}
	}

	// Also check if this is a component call in the AST
	if doc.AST != nil {
		if loc := s.findComponentDefinition(doc.AST, componentName, p.TextDocument.URI); loc != nil {
			return loc, nil
		}
	}

	// Check for function definition in AST
	if doc.AST != nil {
		if loc := s.findFuncDefinition(doc.AST, word, p.TextDocument.URI); loc != nil {
			return loc, nil
		}
	}

	return nil, nil
}

// getWordAtPosition extracts the word at the given position.
func (s *Server) getWordAtPosition(doc *Document, pos Position) string {
	offset := PositionToOffset(doc.Content, pos)
	if offset >= len(doc.Content) {
		return ""
	}

	// Find word boundaries
	start := offset
	end := offset

	// Expand backwards to find word start
	for start > 0 && isWordChar(doc.Content[start-1]) {
		start--
	}

	// Check for @ prefix
	if start > 0 && doc.Content[start-1] == '@' {
		start--
	}

	// Expand forwards to find word end
	for end < len(doc.Content) && isWordChar(doc.Content[end]) {
		end++
	}

	if start >= end {
		return ""
	}

	return doc.Content[start:end]
}

// isWordChar returns true if c is a valid word character.
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// findComponentDefinition looks for a component definition in the AST.
func (s *Server) findComponentDefinition(ast *tuigen.File, name string, uri string) *Location {
	for _, comp := range ast.Components {
		if comp.Name == name {
			return &Location{
				URI: uri,
				Range: Range{
					Start: Position{
						Line:      comp.Position.Line - 1,
						Character: comp.Position.Column - 1,
					},
					End: Position{
						Line:      comp.Position.Line - 1,
						Character: comp.Position.Column - 1 + len("@component") + 1 + len(comp.Name),
					},
				},
			}
		}
	}
	return nil
}

// findFuncDefinition looks for a function definition in the AST.
func (s *Server) findFuncDefinition(ast *tuigen.File, name string, uri string) *Location {
	for _, fn := range ast.Funcs {
		// Parse function name from the code
		fnName, _, _, _ := parseFuncSignature(fn.Code)
		if fnName == name {
			return &Location{
				URI: uri,
				Range: Range{
					Start: Position{
						Line:      fn.Position.Line - 1,
						Character: fn.Position.Column - 1,
					},
					End: Position{
						Line:      fn.Position.Line - 1,
						Character: fn.Position.Column - 1 + len("func") + 1 + len(fnName),
					},
				},
			}
		}
	}
	return nil
}


// findLocalVariableDefinition finds a @let binding definition.
func (s *Server) findLocalVariableDefinition(doc *Document, componentName, varName string) *Location {
	if doc.AST == nil {
		return nil
	}

	for _, comp := range doc.AST.Components {
		if comp.Name != componentName {
			continue
		}

		// Find @let binding with this name
		if binding := s.findLetBindingInNodesForDef(comp.Body, varName); binding != nil {
			return &Location{
				URI: doc.URI,
				Range: Range{
					Start: Position{
						Line:      binding.Position.Line - 1,
						Character: binding.Position.Column - 1 + len("@let "),
					},
					End: Position{
						Line:      binding.Position.Line - 1,
						Character: binding.Position.Column - 1 + len("@let ") + len(varName),
					},
				},
			}
		}
	}

	return nil
}

// findLetBindingInNodesForDef finds a @let binding by name in AST nodes.
func (s *Server) findLetBindingInNodesForDef(nodes []tuigen.Node, name string) *tuigen.LetBinding {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.LetBinding:
			if n != nil && n.Name == name {
				return n
			}
		case *tuigen.Element:
			if n != nil {
				if found := s.findLetBindingInNodesForDef(n.Children, name); found != nil {
					return found
				}
			}
		case *tuigen.ForLoop:
			if n != nil {
				if found := s.findLetBindingInNodesForDef(n.Body, name); found != nil {
					return found
				}
			}
		case *tuigen.IfStmt:
			if n != nil {
				if found := s.findLetBindingInNodesForDef(n.Then, name); found != nil {
					return found
				}
				if found := s.findLetBindingInNodesForDef(n.Else, name); found != nil {
					return found
				}
			}
		case *tuigen.ComponentCall:
			if n != nil {
				if found := s.findLetBindingInNodesForDef(n.Children, name); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// findLoopVariableDefinition finds a for loop variable definition.
func (s *Server) findLoopVariableDefinition(doc *Document, componentName, varName string, pos Position) *Location {
	if doc.AST == nil {
		return nil
	}

	for _, comp := range doc.AST.Components {
		if comp.Name != componentName {
			continue
		}

		// Find for loop containing this position with the variable
		if loop := s.findForLoopWithVariable(comp.Body, varName, pos); loop != nil {
			// Determine if it's the index or value variable
			if loop.Index == varName {
				return &Location{
					URI: doc.URI,
					Range: Range{
						Start: Position{
							Line:      loop.Position.Line - 1,
							Character: loop.Position.Column - 1 + len("@for "),
						},
						End: Position{
							Line:      loop.Position.Line - 1,
							Character: loop.Position.Column - 1 + len("@for ") + len(varName),
						},
					},
				}
			} else if loop.Value == varName {
				offset := len("@for ")
				if loop.Index != "" {
					offset += len(loop.Index) + 2 // ", "
				}
				return &Location{
					URI: doc.URI,
					Range: Range{
						Start: Position{
							Line:      loop.Position.Line - 1,
							Character: loop.Position.Column - 1 + offset,
						},
						End: Position{
							Line:      loop.Position.Line - 1,
							Character: loop.Position.Column - 1 + offset + len(varName),
						},
					},
				}
			}
		}
	}

	return nil
}

// findForLoopWithVariable finds a for loop that declares the given variable and contains the position.
func (s *Server) findForLoopWithVariable(nodes []tuigen.Node, varName string, pos Position) *tuigen.ForLoop {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.ForLoop:
			if n != nil && (n.Index == varName || n.Value == varName) {
				// Check if the position is within this loop's body (rough check)
				// We're being lenient here - if we find a loop with the variable, return it
				return n
			}
			// Also check nested loops
			if n != nil {
				if found := s.findForLoopWithVariable(n.Body, varName, pos); found != nil {
					return found
				}
			}
		case *tuigen.Element:
			if n != nil {
				if found := s.findForLoopWithVariable(n.Children, varName, pos); found != nil {
					return found
				}
			}
		case *tuigen.IfStmt:
			if n != nil {
				if found := s.findForLoopWithVariable(n.Then, varName, pos); found != nil {
					return found
				}
				if found := s.findForLoopWithVariable(n.Else, varName, pos); found != nil {
					return found
				}
			}
		case *tuigen.ComponentCall:
			if n != nil {
				if found := s.findForLoopWithVariable(n.Children, varName, pos); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// findGoCodeVariableDefinition finds a variable declared in a GoCode statement.
func (s *Server) findGoCodeVariableDefinition(doc *Document, componentName, varName string) *Location {
	if doc.AST == nil {
		return nil
	}

	for _, comp := range doc.AST.Components {
		if comp.Name != componentName {
			continue
		}

		// Find GoCode statement that declares this variable
		if goCode := s.findGoCodeWithVariable(comp.Body, varName); goCode != nil {
			// Find the position of the variable name within the code
			idx := findVarDeclPosition(goCode.Code, varName)
			if idx >= 0 {
				return &Location{
					URI: doc.URI,
					Range: Range{
						Start: Position{
							Line:      goCode.Position.Line - 1,
							Character: goCode.Position.Column - 1 + idx,
						},
						End: Position{
							Line:      goCode.Position.Line - 1,
							Character: goCode.Position.Column - 1 + idx + len(varName),
						},
					},
				}
			}
		}
	}

	return nil
}

// findGoCodeWithVariable finds a GoCode node that declares the given variable.
func (s *Server) findGoCodeWithVariable(nodes []tuigen.Node, varName string) *tuigen.GoCode {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoCode:
			if n != nil && containsVarDecl(n.Code, varName) {
				return n
			}
		case *tuigen.Element:
			if n != nil {
				if found := s.findGoCodeWithVariable(n.Children, varName); found != nil {
					return found
				}
			}
		case *tuigen.ForLoop:
			if n != nil {
				if found := s.findGoCodeWithVariable(n.Body, varName); found != nil {
					return found
				}
			}
		case *tuigen.IfStmt:
			if n != nil {
				if found := s.findGoCodeWithVariable(n.Then, varName); found != nil {
					return found
				}
				if found := s.findGoCodeWithVariable(n.Else, varName); found != nil {
					return found
				}
			}
		case *tuigen.ComponentCall:
			if n != nil {
				if found := s.findGoCodeWithVariable(n.Children, varName); found != nil {
					return found
				}
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				if found := s.findGoCodeWithVariable(n.Element.Children, varName); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// containsVarDecl checks if the code declares the given variable.
func containsVarDecl(code, varName string) bool {
	// Handle short variable declaration: name := or name, other :=
	if idx := strings.Index(code, ":="); idx > 0 {
		lhs := code[:idx]
		parts := strings.Split(lhs, ",")
		for _, part := range parts {
			if strings.TrimSpace(part) == varName {
				return true
			}
		}
	}

	// Handle var declaration
	if strings.HasPrefix(strings.TrimSpace(code), "var ") {
		rest := strings.TrimPrefix(strings.TrimSpace(code), "var ")
		if idx := strings.Index(rest, "="); idx > 0 {
			lhs := rest[:idx]
			parts := strings.Split(lhs, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				fields := strings.Fields(part)
				if len(fields) > 0 && fields[0] == varName {
					return true
				}
			}
		}
	}

	return false
}

// findVarDeclPosition finds the position of a variable declaration in code.
func findVarDeclPosition(code, varName string) int {
	// Handle short variable declaration
	if idx := strings.Index(code, ":="); idx > 0 {
		lhs := code[:idx]
		parts := strings.Split(lhs, ",")
		pos := 0
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == varName {
				// Find the actual position accounting for leading whitespace
				partStart := strings.Index(lhs[pos:], trimmed)
				if partStart >= 0 {
					return pos + partStart
				}
			}
			pos += len(part) + 1 // +1 for comma
		}
	}

	// Handle var declaration
	if strings.HasPrefix(strings.TrimSpace(code), "var ") {
		varIdx := strings.Index(code, "var ")
		rest := code[varIdx+4:]
		if idx := strings.Index(rest, "="); idx > 0 {
			lhs := rest[:idx]
			partStart := strings.Index(lhs, varName)
			if partStart >= 0 {
				return varIdx + 4 + partStart
			}
		}
	}

	return -1
}

// getGoplsDefinition gets definition locations from gopls for Go expressions.
func (s *Server) getGoplsDefinition(doc *Document, pos Position) ([]Location, error) {
	if s.goplsProxy == nil {
		return nil, nil
	}

	// Get the cached virtual file
	cached := s.virtualFiles.Get(doc.URI)
	if cached == nil || cached.SourceMap == nil {
		return nil, nil
	}

	// Translate position from .tui to .go
	goLine, goCol, found := cached.SourceMap.TuiToGo(pos.Line, pos.Character)
	if !found {
		s.log("No mapping found for definition position %d:%d", pos.Line, pos.Character)
		return nil, nil
	}

	s.log("Translated definition position %d:%d -> %d:%d", pos.Line, pos.Character, goLine, goCol)

	// Call gopls for definition
	goplsLocs, err := s.goplsProxy.Definition(cached.GoURI, gopls.Position{
		Line:      goLine,
		Character: goCol,
	})
	if err != nil {
		return nil, err
	}

	if len(goplsLocs) == 0 {
		s.log("gopls returned no definition locations")
		return nil, nil
	}

	s.log("gopls returned %d definition location(s)", len(goplsLocs))

	// Convert gopls locations to our Location format
	var locs []Location
	for i, gl := range goplsLocs {
		s.log("gopls location[%d]: URI=%s Range=(%d:%d)-(%d:%d)",
			i, gl.URI, gl.Range.Start.Line, gl.Range.Start.Character, gl.Range.End.Line, gl.Range.End.Character)

		// Check if this is a virtual file - if so, translate back to .tui
		if gopls.IsVirtualGoFile(gl.URI) {
			tuiURI := gopls.GoURIToTuiURI(gl.URI)
			cachedFile := s.virtualFiles.Get(tuiURI)
			if cachedFile != nil && cachedFile.SourceMap != nil {
				s.log("Translating virtual file range back to .tui")
				tuiStartLine, tuiStartCol, startFound := cachedFile.SourceMap.GoToTui(gl.Range.Start.Line, gl.Range.Start.Character)
				tuiEndLine, tuiEndCol, endFound := cachedFile.SourceMap.GoToTui(gl.Range.End.Line, gl.Range.End.Character)
				s.log("GoToTui translation: start(%d:%d->%d:%d, found=%v) end(%d:%d->%d:%d, found=%v)",
					gl.Range.Start.Line, gl.Range.Start.Character, tuiStartLine, tuiStartCol, startFound,
					gl.Range.End.Line, gl.Range.End.Character, tuiEndLine, tuiEndCol, endFound)
				finalLoc := Location{
					URI: tuiURI,
					Range: Range{
						Start: Position{Line: tuiStartLine, Character: tuiStartCol},
						End:   Position{Line: tuiEndLine, Character: tuiEndCol},
					},
				}
				s.log("RETURNING definition location: URI=%s Range=(%d:%d)-(%d:%d)",
					finalLoc.URI, finalLoc.Range.Start.Line, finalLoc.Range.Start.Character,
					finalLoc.Range.End.Line, finalLoc.Range.End.Character)
				locs = append(locs, finalLoc)
				continue
			}
		}

		// For external files (like standard library), return as-is
		locs = append(locs, Location{
			URI: gl.URI,
			Range: Range{
				Start: Position{Line: gl.Range.Start.Line, Character: gl.Range.Start.Character},
				End:   Position{Line: gl.Range.End.Line, Character: gl.Range.End.Character},
			},
		})
	}

	return locs, nil
}
