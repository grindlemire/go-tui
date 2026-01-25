package lsp

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// Semantic token types (must match the order in SemanticTokensLegend.TokenTypes)
const (
	tokenTypeNamespace = 0  // package
	tokenTypeType      = 1  // types
	tokenTypeClass     = 2  // components
	tokenTypeFunction  = 3  // functions
	tokenTypeParameter = 4  // parameters
	tokenTypeVariable  = 5  // variables
	tokenTypeProperty  = 6  // attributes
	tokenTypeKeyword   = 7  // keywords
	tokenTypeString    = 8  // strings
	tokenTypeNumber    = 9  // numbers
	tokenTypeOperator  = 10 // operators
	tokenTypeDecorator = 11 // @ prefix
)

// Semantic token modifiers (bit flags)
const (
	tokenModDeclaration  = 1 << 0 // where defined
	tokenModDefinition   = 1 << 1 // where defined
	tokenModReadonly     = 1 << 2 // const/let
	tokenModModification = 1 << 3 // where modified
)

// SemanticTokensParams represents textDocument/semanticTokens/full parameters.
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SemanticTokens represents the result of semantic tokens request.
type SemanticTokens struct {
	Data []int `json:"data"`
}

// semanticToken represents a single semantic token before encoding.
type semanticToken struct {
	line      int
	startChar int
	length    int
	tokenType int
	modifiers int
}

// handleSemanticTokensFull handles textDocument/semanticTokens/full requests.
func (s *Server) handleSemanticTokensFull(params json.RawMessage) (any, *Error) {
	var p SemanticTokensParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	s.log("Semantic tokens request for %s", p.TextDocument.URI)

	doc := s.docs.Get(p.TextDocument.URI)
	if doc == nil {
		return SemanticTokens{Data: []int{}}, nil
	}

	if doc.AST == nil {
		return SemanticTokens{Data: []int{}}, nil
	}

	tokens := s.collectSemanticTokens(doc)

	// Sort tokens by position
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].line != tokens[j].line {
			return tokens[i].line < tokens[j].line
		}
		return tokens[i].startChar < tokens[j].startChar
	})

	// Encode tokens into delta format
	data := encodeSemanticTokens(tokens)

	return SemanticTokens{Data: data}, nil
}

// collectSemanticTokens collects all semantic tokens from a document.
func (s *Server) collectSemanticTokens(doc *Document) []semanticToken {
	var tokens []semanticToken
	ast := doc.AST

	// Collect component-related tokens
	for _, comp := range ast.Components {
		// Component keyword (@component)
		tokens = append(tokens, semanticToken{
			line:      comp.Position.Line - 1,
			startChar: comp.Position.Column - 1,
			length:    len("@component"),
			tokenType: tokenTypeKeyword,
			modifiers: 0,
		})

		// Component name (declaration)
		nameStart := comp.Position.Column - 1 + len("@component ")
		tokens = append(tokens, semanticToken{
			line:      comp.Position.Line - 1,
			startChar: nameStart,
			length:    len(comp.Name),
			tokenType: tokenTypeClass,
			modifiers: tokenModDeclaration | tokenModDefinition,
		})

		// Parameters (declarations)
		for _, param := range comp.Params {
			tokens = append(tokens, semanticToken{
				line:      param.Position.Line - 1,
				startChar: param.Position.Column - 1,
				length:    len(param.Name),
				tokenType: tokenTypeParameter,
				modifiers: tokenModDeclaration,
			})
		}

		// Collect tokens from body
		s.collectTokensFromNodes(comp.Body, comp.Params, &tokens)
	}

	// Collect function-related tokens
	for _, fn := range ast.Funcs {
		// func keyword
		tokens = append(tokens, semanticToken{
			line:      fn.Position.Line - 1,
			startChar: fn.Position.Column - 1,
			length:    len("func"),
			tokenType: tokenTypeKeyword,
			modifiers: 0,
		})

		// Function name
		name, _, params, _ := parseFuncSignature(fn.Code)
		if name != "" {
			nameStart := fn.Position.Column - 1 + len("func ")
			tokens = append(tokens, semanticToken{
				line:      fn.Position.Line - 1,
				startChar: nameStart,
				length:    len(name),
				tokenType: tokenTypeFunction,
				modifiers: tokenModDeclaration | tokenModDefinition,
			})

			// Function parameters
			paramStart := nameStart + len(name) + 1 // +1 for '('
			for _, p := range params {
				tokens = append(tokens, semanticToken{
					line:      fn.Position.Line - 1,
					startChar: paramStart,
					length:    len(p.Name),
					tokenType: tokenTypeParameter,
					modifiers: tokenModDeclaration,
				})
				// Move past "name type, "
				paramStart += len(p.Name) + 1 + len(p.Type) + 2
			}
		}
	}

	return tokens
}

// collectTokensFromNodes collects semantic tokens from AST nodes.
func (s *Server) collectTokensFromNodes(nodes []tuigen.Node, params []*tuigen.Param, tokens *[]semanticToken) {
	// Build a set of parameter names for quick lookup
	paramNames := make(map[string]bool)
	for _, p := range params {
		paramNames[p.Name] = true
	}

	// Track local variables from @let bindings
	localVars := make(map[string]bool)

	for _, node := range nodes {
		s.collectTokensFromNode(node, paramNames, localVars, tokens)
	}
}

// collectTokensFromNode collects semantic tokens from a single node.
func (s *Server) collectTokensFromNode(node tuigen.Node, paramNames map[string]bool, localVars map[string]bool, tokens *[]semanticToken) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *tuigen.Element:
		if n == nil {
			return
		}
		// Note: We don't tokenize element tags - let syntax highlighting handle them
		// This avoids overriding the default pink color with semantic token colors

		// Attributes
		for _, attr := range n.Attributes {
			*tokens = append(*tokens, semanticToken{
				line:      attr.Position.Line - 1,
				startChar: attr.Position.Column - 1,
				length:    len(attr.Name),
				tokenType: tokenTypeProperty,
				modifiers: 0,
			})

			// Check for variable usages in attribute values
			if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
				s.collectVariableTokensInCode(expr.Code, expr.Position, paramNames, localVars, tokens)
			}
		}

		// Children
		for _, child := range n.Children {
			s.collectTokensFromNode(child, paramNames, localVars, tokens)
		}

	case *tuigen.GoExpr:
		if n == nil {
			return
		}
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
		// @for keyword
		*tokens = append(*tokens, semanticToken{
			line:      n.Position.Line - 1,
			startChar: n.Position.Column - 1,
			length:    len("@for"),
			tokenType: tokenTypeKeyword,
			modifiers: 0,
		})

		// Loop variables are local to the loop
		loopVars := make(map[string]bool)
		for k, v := range localVars {
			loopVars[k] = v
		}
		if n.Index != "" && n.Index != "_" {
			loopVars[n.Index] = true
			// Index variable token
			idxStart := n.Position.Column - 1 + len("@for ")
			*tokens = append(*tokens, semanticToken{
				line:      n.Position.Line - 1,
				startChar: idxStart,
				length:    len(n.Index),
				tokenType: tokenTypeVariable,
				modifiers: tokenModDeclaration,
			})
		}
		if n.Value != "" {
			loopVars[n.Value] = true
			// Value variable token - approximate position
			valStart := n.Position.Column - 1 + len("@for ")
			if n.Index != "" {
				valStart += len(n.Index) + 2 // ", "
			}
			*tokens = append(*tokens, semanticToken{
				line:      n.Position.Line - 1,
				startChar: valStart,
				length:    len(n.Value),
				tokenType: tokenTypeVariable,
				modifiers: tokenModDeclaration,
			})
		}

		// Iterable expression
		s.collectVariableTokensInCode(n.Iterable, n.Position, paramNames, loopVars, tokens)

		// Body with loop variables in scope
		for _, child := range n.Body {
			s.collectTokensFromNode(child, paramNames, loopVars, tokens)
		}

	case *tuigen.IfStmt:
		if n == nil {
			return
		}
		// @if keyword
		*tokens = append(*tokens, semanticToken{
			line:      n.Position.Line - 1,
			startChar: n.Position.Column - 1,
			length:    len("@if"),
			tokenType: tokenTypeKeyword,
			modifiers: 0,
		})

		// Condition
		condPos := tuigen.Position{Line: n.Position.Line, Column: n.Position.Column + len("@if ")}
		s.collectVariableTokensInCode(n.Condition, condPos, paramNames, localVars, tokens)

		// Then branch
		for _, child := range n.Then {
			s.collectTokensFromNode(child, paramNames, localVars, tokens)
		}

		// Else branch
		for _, child := range n.Else {
			s.collectTokensFromNode(child, paramNames, localVars, tokens)
		}

	case *tuigen.LetBinding:
		if n == nil {
			return
		}
		// @let keyword
		*tokens = append(*tokens, semanticToken{
			line:      n.Position.Line - 1,
			startChar: n.Position.Column - 1,
			length:    len("@let"),
			tokenType: tokenTypeKeyword,
			modifiers: 0,
		})

		// Variable name (declaration)
		varStart := n.Position.Column - 1 + len("@let ")
		*tokens = append(*tokens, semanticToken{
			line:      n.Position.Line - 1,
			startChar: varStart,
			length:    len(n.Name),
			tokenType: tokenTypeVariable,
			modifiers: tokenModDeclaration | tokenModReadonly,
		})

		// Add to local vars for subsequent nodes
		localVars[n.Name] = true

		// Process the entire element (including tag, attributes, and children)
		if n.Element != nil {
			s.collectTokensFromNode(n.Element, paramNames, localVars, tokens)
		}

	case *tuigen.GoCode:
		if n == nil {
			return
		}
		// Extract variable declarations from Go code (e.g., "x := 1" or "var x = 1")
		// and add them to localVars for subsequent highlighting
		varNames := extractVarDeclarations(n.Code)
		for _, varName := range varNames {
			localVars[varName] = true
		}
		// Generate tokens for variables in the Go code
		// Use offset 0 because GoCode position points to the first char (no braces)
		s.collectVariableTokensInCodeWithOffset(n.Code, n.Position, 0, paramNames, localVars, tokens)

	case *tuigen.ComponentCall:
		if n == nil {
			return
		}
		// @ decorator
		*tokens = append(*tokens, semanticToken{
			line:      n.Position.Line - 1,
			startChar: n.Position.Column - 1,
			length:    1,
			tokenType: tokenTypeDecorator,
			modifiers: 0,
		})

		// Component name
		*tokens = append(*tokens, semanticToken{
			line:      n.Position.Line - 1,
			startChar: n.Position.Column, // After @
			length:    len(n.Name),
			tokenType: tokenTypeClass,
			modifiers: 0,
		})

		// Arguments
		if n.Args != "" {
			argPos := tuigen.Position{Line: n.Position.Line, Column: n.Position.Column + len(n.Name) + 1}
			s.collectVariableTokensInCode(n.Args, argPos, paramNames, localVars, tokens)
		}

		// Children
		for _, child := range n.Children {
			s.collectTokensFromNode(child, paramNames, localVars, tokens)
		}
	}
}

// collectVariableTokensInCode finds and tokenizes variable references in Go code.
// The startOffset parameter adjusts for the leading character(s) before the code
// (e.g., 1 for GoExpr inside {}, 0 for raw GoCode).
func (s *Server) collectVariableTokensInCode(code string, pos tuigen.Position, paramNames map[string]bool, localVars map[string]bool, tokens *[]semanticToken) {
	s.collectVariableTokensInCodeWithOffset(code, pos, 1, paramNames, localVars, tokens)
}

// collectVariableTokensInCodeWithOffset is like collectVariableTokensInCode but allows
// specifying the offset from the position to the start of the code.
func (s *Server) collectVariableTokensInCodeWithOffset(code string, pos tuigen.Position, startOffset int, paramNames map[string]bool, localVars map[string]bool, tokens *[]semanticToken) {
	// Simple tokenization to find identifiers
	i := 0
	for i < len(code) {
		// Skip non-identifier characters
		if !isWordChar(code[i]) {
			i++
			continue
		}

		// Find identifier
		start := i
		for i < len(code) && isWordChar(code[i]) {
			i++
		}
		ident := code[start:i]

		// Calculate the character position:
		// pos.Column is 1-indexed, LSP wants 0-indexed
		// startOffset accounts for leading chars (1 for '{' in GoExpr, 0 for GoCode)
		charPos := pos.Column - 1 + startOffset + start

		// Check if it's a parameter or local variable
		if paramNames[ident] {
			*tokens = append(*tokens, semanticToken{
				line:      pos.Line - 1,
				startChar: charPos,
				length:    len(ident),
				tokenType: tokenTypeParameter,
				modifiers: 0,
			})
		} else if localVars[ident] {
			*tokens = append(*tokens, semanticToken{
				line:      pos.Line - 1,
				startChar: charPos,
				length:    len(ident),
				tokenType: tokenTypeVariable,
				modifiers: 0,
			})
		} else if s.isFunctionName(ident) {
			*tokens = append(*tokens, semanticToken{
				line:      pos.Line - 1,
				startChar: charPos,
				length:    len(ident),
				tokenType: tokenTypeFunction,
				modifiers: 0,
			})
		}
	}
}

// isFunctionName checks if an identifier is a known function.
func (s *Server) isFunctionName(name string) bool {
	// Check indexed functions
	if _, ok := s.index.LookupFunc(name); ok {
		return true
	}

	// Check if followed by '(' in the call context would be better,
	// but for now we check common Go built-ins and functions
	builtins := map[string]bool{
		"len": true, "cap": true, "make": true, "new": true,
		"append": true, "copy": true, "delete": true,
		"close": true, "panic": true, "recover": true,
		"print": true, "println": true,
		"real": true, "imag": true, "complex": true,
	}
	return builtins[name]
}

// encodeSemanticTokens encodes tokens into the LSP delta format.
func encodeSemanticTokens(tokens []semanticToken) []int {
	if len(tokens) == 0 {
		return []int{}
	}

	// Encode as [deltaLine, deltaStartChar, length, tokenType, tokenModifiers]
	data := make([]int, 0, len(tokens)*5)

	prevLine := 0
	prevChar := 0

	for _, t := range tokens {
		deltaLine := t.line - prevLine
		deltaChar := t.startChar
		if deltaLine == 0 {
			deltaChar = t.startChar - prevChar
		}

		data = append(data, deltaLine, deltaChar, t.length, t.tokenType, t.modifiers)

		prevLine = t.line
		prevChar = t.startChar
	}

	return data
}

// Helper to check if char is a Go keyword
func isGoKeyword(word string) bool {
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true,
		"continue": true, "default": true, "defer": true, "else": true,
		"fallthrough": true, "for": true, "func": true, "go": true,
		"goto": true, "if": true, "import": true, "interface": true,
		"map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true,
		"var": true,
	}
	return keywords[strings.ToLower(word)]
}

// extractVarDeclarations extracts variable names from Go code declarations.
// Handles patterns like "x := 1", "x, y := 1, 2", "var x = 1", "var x, y = 1, 2"
func extractVarDeclarations(code string) []string {
	var varNames []string

	// Handle short variable declaration: x := or x, y :=
	if idx := strings.Index(code, ":="); idx > 0 {
		lhs := strings.TrimSpace(code[:idx])
		// Split by comma for multiple declarations
		parts := strings.Split(lhs, ",")
		for _, part := range parts {
			name := strings.TrimSpace(part)
			if name != "" && name != "_" && isValidIdentifier(name) {
				varNames = append(varNames, name)
			}
		}
		return varNames
	}

	// Handle var declaration: var x = or var x, y =
	if strings.HasPrefix(strings.TrimSpace(code), "var ") {
		rest := strings.TrimPrefix(strings.TrimSpace(code), "var ")
		// Find = sign
		if idx := strings.Index(rest, "="); idx > 0 {
			lhs := strings.TrimSpace(rest[:idx])
			// Remove type annotation if present (e.g., "x int" -> "x")
			parts := strings.Split(lhs, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				// Take first word (variable name before type)
				fields := strings.Fields(part)
				if len(fields) > 0 {
					name := fields[0]
					if name != "_" && isValidIdentifier(name) {
						varNames = append(varNames, name)
					}
				}
			}
		}
	}

	return varNames
}

// isValidIdentifier checks if a string is a valid Go identifier.
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
				return false
			}
		} else {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
	}
	return true
}
