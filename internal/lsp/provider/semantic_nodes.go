package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/lsp/schema"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// collectTokensFromNodes collects semantic tokens from AST nodes.
func (s *semanticTokensProvider) collectTokensFromNodes(nodes []tuigen.Node, params []*tuigen.Param, tokens *[]SemanticToken) {
	paramNames := make(map[string]bool)
	for _, p := range params {
		paramNames[p.Name] = true
	}
	localVars := make(map[string]bool)
	for _, node := range nodes {
		s.collectTokensFromNode(node, paramNames, localVars, tokens)
	}
}

// collectTokensFromNode collects semantic tokens from a single node.
func (s *semanticTokensProvider) collectTokensFromNode(node tuigen.Node, paramNames map[string]bool, localVars map[string]bool, tokens *[]SemanticToken) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *tuigen.Element:
		if n == nil {
			return
		}
		// ref={name} attribute — emit "ref" as function token and the value as variable
		if n.RefExpr != nil && s.currentContent != "" {
			// Search for "ref={" in the document content to find exact position
			docLines := strings.Split(s.currentContent, "\n")
			startLine := n.Position.Line - 1
			maxSearch := startLine + 20
			if maxSearch > len(docLines) {
				maxSearch = len(docLines)
			}
			for lineIdx := startLine; lineIdx < maxSearch; lineIdx++ {
				refIdx := strings.Index(docLines[lineIdx], "ref={"+n.RefExpr.Code+"}")
				if refIdx >= 0 {
					// Emit "ref" as function token (attribute name)
					*tokens = append(*tokens, SemanticToken{
						Line:      lineIdx,
						StartChar: refIdx,
						Length:    len("ref"),
						TokenType: TokenTypeFunction,
						Modifiers: 0,
					})
					// Emit ref value as variable with declaration modifier
					*tokens = append(*tokens, SemanticToken{
						Line:      lineIdx,
						StartChar: refIdx + len("ref={"),
						Length:    len(n.RefExpr.Code),
						TokenType: TokenTypeVariable,
						Modifiers: TokenModDeclaration,
					})
					break
				}
			}
		}
		// Attributes — distinguish event handlers from regular attributes
		for _, attr := range n.Attributes {
			tokenType := TokenTypeFunction
			if schema.IsEventHandler(attr.Name) {
				tokenType = TokenTypeDecorator // Use decorator type for event handlers
			}
			*tokens = append(*tokens, SemanticToken{
				Line:      attr.Position.Line - 1,
				StartChar: attr.Position.Column - 1,
				Length:    len(attr.Name),
				TokenType: tokenType,
				Modifiers: 0,
			})
			if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
				s.collectVariableTokensInCode(expr.Code, expr.Position, paramNames, localVars, tokens)
			}
		}
		for _, child := range n.Children {
			s.collectTokensFromNode(child, paramNames, localVars, tokens)
		}

	case *tuigen.GoExpr:
		if n == nil {
			return
		}
		log.Server("GoExpr node: Code=%q Position.Line=%d Position.Column=%d",
			n.Code, n.Position.Line, n.Position.Column)
		s.collectVariableTokensInCode(n.Code, n.Position, paramNames, localVars, tokens)

	case *tuigen.RawGoExpr:
		if n == nil {
			return
		}
		s.collectVariableTokensInCode(n.Code, n.Position, paramNames, localVars, tokens)

	case *tuigen.ForLoop:
		if n == nil {
			return
		}
		*tokens = append(*tokens, SemanticToken{
			Line:      n.Position.Line - 1,
			StartChar: n.Position.Column - 1,
			Length:    len("@for"),
			TokenType: TokenTypeKeyword,
			Modifiers: 0,
		})
		loopVars := make(map[string]bool)
		for k, v := range localVars {
			loopVars[k] = v
		}
		if n.Index != "" && n.Index != "_" {
			loopVars[n.Index] = true
			idxStart := n.Position.Column - 1 + len("@for ")
			*tokens = append(*tokens, SemanticToken{
				Line:      n.Position.Line - 1,
				StartChar: idxStart,
				Length:    len(n.Index),
				TokenType: TokenTypeVariable,
				Modifiers: TokenModDeclaration,
			})
		}
		if n.Value != "" {
			loopVars[n.Value] = true
			valStart := n.Position.Column - 1 + len("@for ")
			if n.Index != "" {
				valStart += len(n.Index) + 2
			}
			*tokens = append(*tokens, SemanticToken{
				Line:      n.Position.Line - 1,
				StartChar: valStart,
				Length:    len(n.Value),
				TokenType: TokenTypeVariable,
				Modifiers: TokenModDeclaration,
			})
		}
		// Iterable expression
		iterableOffset := len("@for ")
		if n.Index != "" {
			iterableOffset += len(n.Index) + 2
		}
		iterableOffset += len(n.Value) + len(" := range ")
		iterablePos := tuigen.Position{Line: n.Position.Line, Column: n.Position.Column + iterableOffset}
		s.collectTokensInGoCodeDirect(n.Iterable, iterablePos, paramNames, loopVars, tokens)
		for _, child := range n.Body {
			s.collectTokensFromNode(child, paramNames, loopVars, tokens)
		}

	case *tuigen.IfStmt:
		if n == nil {
			return
		}
		*tokens = append(*tokens, SemanticToken{
			Line:      n.Position.Line - 1,
			StartChar: n.Position.Column - 1,
			Length:    len("@if"),
			TokenType: TokenTypeKeyword,
			Modifiers: 0,
		})
		condPos := tuigen.Position{Line: n.Position.Line, Column: n.Position.Column + len("@if ")}
		s.collectTokensInGoCodeDirect(n.Condition, condPos, paramNames, localVars, tokens)
		for _, child := range n.Then {
			s.collectTokensFromNode(child, paramNames, localVars, tokens)
		}
		if len(n.Else) > 0 {
			// Emit @else keyword token by scanning source text
			elseKeywordLine, elseKeywordCol := s.findElseKeyword(n)
			if elseKeywordLine >= 0 {
				*tokens = append(*tokens, SemanticToken{
					Line:      elseKeywordLine,
					StartChar: elseKeywordCol,
					Length:    len("@else"),
					TokenType: TokenTypeKeyword,
					Modifiers: 0,
				})
			}
			for _, child := range n.Else {
				s.collectTokensFromNode(child, paramNames, localVars, tokens)
			}
		}

	case *tuigen.LetBinding:
		if n == nil {
			return
		}
		*tokens = append(*tokens, SemanticToken{
			Line:      n.Position.Line - 1,
			StartChar: n.Position.Column - 1,
			Length:    len("@let"),
			TokenType: TokenTypeKeyword,
			Modifiers: 0,
		})
		varStart := n.Position.Column - 1 + len("@let ")
		*tokens = append(*tokens, SemanticToken{
			Line:      n.Position.Line - 1,
			StartChar: varStart,
			Length:    len(n.Name),
			TokenType: TokenTypeVariable,
			Modifiers: TokenModDeclaration | TokenModReadonly,
		})
		localVars[n.Name] = true
		if n.Element != nil {
			s.collectTokensFromNode(n.Element, paramNames, localVars, tokens)
		}

	case *tuigen.GoCode:
		if n == nil {
			return
		}
		// Identify which specific variable (if any) is the state declaration.
		// Only that variable gets the readonly modifier, not all vars in the block.
		stateVarName := ""
		if matches := stateNewStateRegex.FindStringSubmatch(n.Code); len(matches) >= 2 {
			stateVarName = matches[1]
		}
		varDecls := extractVarDeclarationsWithPositions(n.Code)
		for _, decl := range varDecls {
			localVars[decl.name] = true
			modifiers := TokenModDeclaration
			if stateVarName != "" && decl.name == stateVarName {
				modifiers = TokenModDeclaration | TokenModReadonly // State vars use readonly modifier
			}
			*tokens = append(*tokens, SemanticToken{
				Line:      n.Position.Line - 1,
				StartChar: n.Position.Column - 1 + decl.offset,
				Length:    len(decl.name),
				TokenType: TokenTypeVariable,
				Modifiers: modifiers,
			})
		}
		s.collectTokensInGoCode(n.Code, n.Position, 0, paramNames, localVars, tokens)

	case *tuigen.ComponentCall:
		if n == nil {
			return
		}
		*tokens = append(*tokens, SemanticToken{
			Line:      n.Position.Line - 1,
			StartChar: n.Position.Column - 1,
			Length:    1,
			TokenType: TokenTypeDecorator,
			Modifiers: 0,
		})
		*tokens = append(*tokens, SemanticToken{
			Line:      n.Position.Line - 1,
			StartChar: n.Position.Column,
			Length:    len(n.Name),
			TokenType: TokenTypeClass,
			Modifiers: 0,
		})
		if n.Args != "" {
			argPos := tuigen.Position{Line: n.Position.Line, Column: n.Position.Column + len(n.Name) + 1}
			s.collectVariableTokensInCode(n.Args, argPos, paramNames, localVars, tokens)
		}
		for _, child := range n.Children {
			s.collectTokensFromNode(child, paramNames, localVars, tokens)
		}
	}
}

// emitStringWithFormatSpecifiers emits tokens for a string, splitting format specifiers.
func (s *semanticTokensProvider) emitStringWithFormatSpecifiers(str string, line int, stringStartChar int, tokens *[]SemanticToken) {
	log.Server("emitStringWithFormatSpecifiers: str=%q line=%d startChar=%d", str, line, stringStartChar)
	matches := formatSpecifierRegex.FindAllStringIndex(str, -1)
	log.Server("  format specifier matches: %v", matches)

	if len(matches) == 0 {
		log.Server("  emitting whole string token: line=%d startChar=%d len=%d type=string", line, stringStartChar, len(str))
		*tokens = append(*tokens, SemanticToken{
			Line:      line,
			StartChar: stringStartChar,
			Length:    len(str),
			TokenType: TokenTypeString,
			Modifiers: 0,
		})
		return
	}

	idx := 0
	for _, match := range matches {
		if match[0] > idx {
			log.Server("  emit STRING token: line=%d startChar=%d length=%d (content=%q)",
				line, stringStartChar+idx, match[0]-idx, str[idx:match[0]])
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: stringStartChar + idx,
				Length:    match[0] - idx,
				TokenType: TokenTypeString,
				Modifiers: 0,
			})
		}
		log.Server("  emit REGEXP token for format spec: line=%d startChar=%d length=%d (content=%q)",
			line, stringStartChar+match[0], match[1]-match[0], str[match[0]:match[1]])
		*tokens = append(*tokens, SemanticToken{
			Line:      line,
			StartChar: stringStartChar + match[0],
			Length:    match[1] - match[0],
			TokenType: TokenTypeRegexp,
			Modifiers: 0,
		})
		idx = match[1]
	}
	if idx < len(str) {
		log.Server("  emit STRING token (tail): line=%d startChar=%d length=%d (content=%q)",
			line, stringStartChar+idx, len(str)-idx, str[idx:])
		*tokens = append(*tokens, SemanticToken{
			Line:      line,
			StartChar: stringStartChar + idx,
			Length:    len(str) - idx,
			TokenType: TokenTypeString,
			Modifiers: 0,
		})
	}
}

// findElseKeyword scans the document content to find the @else keyword position for an IfStmt.
// Returns 0-indexed (line, col), or (-1, -1) if not found.
func (s *semanticTokensProvider) findElseKeyword(ifStmt *tuigen.IfStmt) (int, int) {
	if s.docs == nil {
		return -1, -1
	}
	doc := s.docs.GetDocument(s.currentURI)
	if doc == nil {
		return -1, -1
	}
	lines := strings.Split(doc.Content, "\n")
	// Scan from the @if line downward looking for @else
	startLine := ifStmt.Position.Line - 1 // 0-indexed
	for i := startLine; i < len(lines); i++ {
		idx := strings.Index(lines[i], "@else")
		if idx >= 0 {
			return i, idx
		}
	}
	return -1, -1
}

// --- Comment collection ---

func (s *semanticTokensProvider) collectAllCommentTokens(file *tuigen.File, tokens *[]SemanticToken) {
	if file == nil {
		return
	}
	s.collectCommentGroupTokens(file.LeadingComments, tokens)
	for _, cg := range file.OrphanComments {
		s.collectCommentGroupTokens(cg, tokens)
	}
	for _, imp := range file.Imports {
		s.collectCommentGroupTokens(imp.TrailingComments, tokens)
	}
	for _, comp := range file.Components {
		s.collectComponentCommentTokens(comp, tokens)
	}
	for _, fn := range file.Funcs {
		s.collectCommentGroupTokens(fn.LeadingComments, tokens)
		s.collectCommentGroupTokens(fn.TrailingComments, tokens)
	}
}

func (s *semanticTokensProvider) collectComponentCommentTokens(comp *tuigen.Component, tokens *[]SemanticToken) {
	if comp == nil {
		return
	}
	s.collectCommentGroupTokens(comp.LeadingComments, tokens)
	s.collectCommentGroupTokens(comp.TrailingComments, tokens)
	for _, cg := range comp.OrphanComments {
		s.collectCommentGroupTokens(cg, tokens)
	}
	for _, node := range comp.Body {
		s.collectNodeCommentTokens(node, tokens)
	}
}

func (s *semanticTokensProvider) collectNodeCommentTokens(node tuigen.Node, tokens *[]SemanticToken) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *tuigen.Element:
		if n == nil {
			return
		}
		s.collectCommentGroupTokens(n.LeadingComments, tokens)
		s.collectCommentGroupTokens(n.TrailingComments, tokens)
		for _, child := range n.Children {
			s.collectNodeCommentTokens(child, tokens)
		}
	case *tuigen.GoExpr:
		if n == nil {
			return
		}
		s.collectCommentGroupTokens(n.LeadingComments, tokens)
		s.collectCommentGroupTokens(n.TrailingComments, tokens)
	case *tuigen.GoCode:
		if n == nil {
			return
		}
		s.collectCommentGroupTokens(n.LeadingComments, tokens)
		s.collectCommentGroupTokens(n.TrailingComments, tokens)
	case *tuigen.ForLoop:
		if n == nil {
			return
		}
		s.collectCommentGroupTokens(n.LeadingComments, tokens)
		s.collectCommentGroupTokens(n.TrailingComments, tokens)
		for _, cg := range n.OrphanComments {
			s.collectCommentGroupTokens(cg, tokens)
		}
		for _, child := range n.Body {
			s.collectNodeCommentTokens(child, tokens)
		}
	case *tuigen.IfStmt:
		if n == nil {
			return
		}
		s.collectCommentGroupTokens(n.LeadingComments, tokens)
		s.collectCommentGroupTokens(n.TrailingComments, tokens)
		for _, cg := range n.OrphanComments {
			s.collectCommentGroupTokens(cg, tokens)
		}
		for _, child := range n.Then {
			s.collectNodeCommentTokens(child, tokens)
		}
		for _, child := range n.Else {
			s.collectNodeCommentTokens(child, tokens)
		}
	case *tuigen.LetBinding:
		if n == nil {
			return
		}
		s.collectCommentGroupTokens(n.LeadingComments, tokens)
		s.collectCommentGroupTokens(n.TrailingComments, tokens)
		if n.Element != nil {
			s.collectNodeCommentTokens(n.Element, tokens)
		}
	case *tuigen.ComponentCall:
		if n == nil {
			return
		}
		s.collectCommentGroupTokens(n.LeadingComments, tokens)
		s.collectCommentGroupTokens(n.TrailingComments, tokens)
		for _, child := range n.Children {
			s.collectNodeCommentTokens(child, tokens)
		}
	case *tuigen.ChildrenSlot:
		if n == nil {
			return
		}
		s.collectCommentGroupTokens(n.LeadingComments, tokens)
		s.collectCommentGroupTokens(n.TrailingComments, tokens)
	}
}

func (s *semanticTokensProvider) collectCommentGroupTokens(cg *tuigen.CommentGroup, tokens *[]SemanticToken) {
	if cg == nil {
		return
	}
	for _, c := range cg.List {
		s.collectCommentToken(c, tokens)
	}
}

func (s *semanticTokensProvider) collectCommentToken(c *tuigen.Comment, tokens *[]SemanticToken) {
	if c == nil {
		return
	}
	if !c.IsBlock {
		*tokens = append(*tokens, SemanticToken{
			Line:      c.Position.Line - 1,
			StartChar: c.Position.Column - 1,
			Length:    len(c.Text),
			TokenType: TokenTypeComment,
			Modifiers: 0,
		})
		return
	}
	lines := strings.Split(c.Text, "\n")
	for i, line := range lines {
		lineNum := c.Position.Line - 1 + i
		var startChar int
		if i == 0 {
			startChar = c.Position.Column - 1
		} else {
			startChar = 0
			for j := 0; j < len(line) && (line[j] == ' ' || line[j] == '\t'); j++ {
				startChar++
			}
		}
		if len(line) > 0 {
			*tokens = append(*tokens, SemanticToken{
				Line:      lineNum,
				StartChar: startChar,
				Length:    len(line),
				TokenType: TokenTypeComment,
				Modifiers: 0,
			})
		}
	}
}

// --- Go type expression tokenizer ---

// emitGoTypeTokens tokenizes a Go type expression and emits semantic tokens.
// Handles: *Type, pkg.Type, Type[Generic], func(...) ReturnType, (T1, T2).
// Types inside brackets (generic type arguments) use TokenTypeTypeParameter.
func emitGoTypeTokens(typeStr string, line int, startCol int, tokens *[]SemanticToken) {
	i := 0
	bracketDepth := 0
	for i < len(typeStr) {
		ch := typeStr[i]

		// Skip whitespace
		if ch == ' ' || ch == '\t' {
			i++
			continue
		}

		// Pointer star
		if ch == '*' {
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: startCol + i,
				Length:    1,
				TokenType: TokenTypeOperator,
				Modifiers: 0,
			})
			i++
			continue
		}

		// Brackets — track depth for generic type argument coloring
		if ch == '[' {
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: startCol + i,
				Length:    1,
				TokenType: TokenTypeOperator,
				Modifiers: 0,
			})
			bracketDepth++
			i++
			continue
		}
		if ch == ']' {
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: startCol + i,
				Length:    1,
				TokenType: TokenTypeOperator,
				Modifiers: 0,
			})
			bracketDepth--
			i++
			continue
		}

		// Commas, dots, parens — skip
		if ch == ',' || ch == '.' || ch == '(' || ch == ')' {
			i++
			continue
		}

		// Variadic ...
		if ch == '.' && i+2 < len(typeStr) && typeStr[i:i+3] == "..." {
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: startCol + i,
				Length:    3,
				TokenType: TokenTypeOperator,
				Modifiers: 0,
			})
			i += 3
			continue
		}

		// Identifier
		if isWordStartChar(ch) {
			start := i
			for i < len(typeStr) && isWordCharByte(typeStr[i]) {
				i++
			}
			ident := typeStr[start:i]

			// Check what follows: if '.', this is a package prefix — emit as namespace
			if i < len(typeStr) && typeStr[i] == '.' {
				// Package prefix (e.g. "tui" in "tui.State")
				// Skip — don't emit token, keeping it uncolored like Go convention
				i++ // skip the dot
				continue
			}

			// "func" keyword in function types
			if ident == "func" || ident == "map" || ident == "chan" || ident == "interface" {
				*tokens = append(*tokens, SemanticToken{
					Line:      line,
					StartChar: startCol + start,
					Length:    len(ident),
					TokenType: TokenTypeKeyword,
					Modifiers: 0,
				})
				continue
			}

			// Type name — use typeParameter for types inside brackets (generic args)
			tokenType := TokenTypeType
			if bracketDepth > 0 {
				tokenType = TokenTypeTypeParameter
			}
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: startCol + start,
				Length:    len(ident),
				TokenType: tokenType,
				Modifiers: 0,
			})
			continue
		}

		// Channel direction operator <-
		if ch == '<' && i+1 < len(typeStr) && typeStr[i+1] == '-' {
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: startCol + i,
				Length:    2,
				TokenType: TokenTypeOperator,
				Modifiers: 0,
			})
			i += 2
			continue
		}

		// Skip anything else
		i++
	}
}
