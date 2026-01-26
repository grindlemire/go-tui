package lsp

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// formatSpecifierRegex matches Go format specifiers like %s, %d, %v, %.2f, %#x, etc.
var formatSpecifierRegex = regexp.MustCompile(`%[-+# 0]*(\*|\d+)?(\.\*|\.\d+)?[vTtbcdoqxXUeEfFgGsp%]`)

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
	tokenTypeRegexp    = 12 // format specifiers (often purple)
	tokenTypeComment   = 13 // comments
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

	log.Server("=== Semantic tokens request for %s ===", p.TextDocument.URI)

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

	// Log a sample of the encoded data for debugging
	if len(data) > 0 {
		log.Server("Encoded %d tokens (%d ints). First 25 values: %v", len(tokens), len(data), data[:min(25, len(data))])
	}

	return SemanticTokens{Data: data}, nil
}

// collectSemanticTokens collects all semantic tokens from a document.
func (s *Server) collectSemanticTokens(doc *Document) []semanticToken {
	var tokens []semanticToken
	ast := doc.AST

	// Collect comment tokens from the entire AST
	s.collectAllCommentTokens(ast, &tokens)

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

			// Build parameter names map for body tokenization
			paramNames := make(map[string]bool)
			for _, p := range params {
				paramNames[p.Name] = true
			}

			// Tokenize function body (line by line for multi-line support)
			s.collectTokensFromFuncBody(fn.Code, fn.Position, paramNames, &tokens)
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

		// Attributes - using tokenTypeFunction which is typically green
		for _, attr := range n.Attributes {
			*tokens = append(*tokens, semanticToken{
				line:      attr.Position.Line - 1,
				startChar: attr.Position.Column - 1,
				length:    len(attr.Name),
				tokenType: tokenTypeFunction,
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
		// Log position and verify against expected. The position should point to '{',
		// so if we add startOffset=1, we should land on the first char of Code
		log.Server("GoExpr node: Code=%q Position.Line=%d Position.Column=%d (expect Column to point to '{')",
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

		// Iterable expression - calculate position after "@for index, value := range "
		iterableOffset := len("@for ")
		if n.Index != "" {
			iterableOffset += len(n.Index) + 2 // ", "
		}
		iterableOffset += len(n.Value) + len(" := range ")
		iterablePos := tuigen.Position{Line: n.Position.Line, Column: n.Position.Column + iterableOffset}
		s.collectTokensInGoCodeDirect(n.Iterable, iterablePos, paramNames, loopVars, tokens)

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

		// Condition - position points directly to start of condition
		condPos := tuigen.Position{Line: n.Position.Line, Column: n.Position.Column + len("@if ")}
		s.collectTokensInGoCodeDirect(n.Condition, condPos, paramNames, localVars, tokens)

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
		// and emit declaration tokens for them
		varDecls := extractVarDeclarationsWithPositions(n.Code)
		for _, decl := range varDecls {
			localVars[decl.name] = true
			// Emit token for the variable declaration
			*tokens = append(*tokens, semanticToken{
				line:      n.Position.Line - 1,
				startChar: n.Position.Column - 1 + decl.offset,
				length:    len(decl.name),
				tokenType: tokenTypeVariable,
				modifiers: tokenModDeclaration,
			})
		}
		// Generate tokens for other elements in the Go code (strings, booleans, etc.)
		// Use offset 0 because GoCode position points to the first char (no braces)
		s.collectTokensInGoCode(n.Code, n.Position, 0, paramNames, localVars, tokens)

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
// GoExpr.Position points to '{', so we use startOffset=1 to skip past it.
func (s *Server) collectVariableTokensInCode(code string, pos tuigen.Position, paramNames map[string]bool, localVars map[string]bool, tokens *[]semanticToken) {
	s.collectTokensInGoCode(code, pos, 1, paramNames, localVars, tokens)
}

// collectTokensInGoCodeDirect tokenizes Go code without any offset adjustment.
// Use this when the position already points to the exact start of the code.
func (s *Server) collectTokensInGoCodeDirect(code string, pos tuigen.Position, paramNames map[string]bool, localVars map[string]bool, tokens *[]semanticToken) {
	s.collectTokensInGoCode(code, pos, 0, paramNames, localVars, tokens)
}

// collectTokensInGoCode tokenizes Go code for semantic highlighting.
// Handles identifiers (variables, parameters, functions), string literals, boolean literals, numbers, and comments.
func (s *Server) collectTokensInGoCode(code string, pos tuigen.Position, startOffset int, paramNames map[string]bool, localVars map[string]bool, tokens *[]semanticToken) {
	log.Server("collectTokensInGoCode: code=%q pos.Line=%d pos.Column=%d startOffset=%d", code, pos.Line, pos.Column, startOffset)
	i := 0
	for i < len(code) {
		ch := code[i]

		// Skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		// Block comment /* ... */
		if ch == '/' && i+1 < len(code) && code[i+1] == '*' {
			start := i
			i += 2 // skip /*
			for i+1 < len(code) && !(code[i] == '*' && code[i+1] == '/') {
				i++
			}
			if i+1 < len(code) {
				i += 2 // skip */
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, semanticToken{
				line:      pos.Line - 1,
				startChar: charPos,
				length:    i - start,
				tokenType: tokenTypeComment,
				modifiers: 0,
			})
			continue
		}

		// Line comment // ...
		if ch == '/' && i+1 < len(code) && code[i+1] == '/' {
			start := i
			for i < len(code) && code[i] != '\n' {
				i++
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, semanticToken{
				line:      pos.Line - 1,
				startChar: charPos,
				length:    i - start,
				tokenType: tokenTypeComment,
				modifiers: 0,
			})
			continue
		}

		// String literal (double-quoted)
		if ch == '"' {
			start := i
			i++ // skip opening quote
			for i < len(code) && code[i] != '"' {
				if code[i] == '\\' && i+1 < len(code) {
					i += 2 // skip escaped char
				} else {
					i++
				}
			}
			if i < len(code) {
				i++ // skip closing quote
			}
			stringContent := code[start:i]
			charPos := pos.Column - 1 + startOffset + start
			log.Server("Found string in GoCode: code=%q stringContent=%q start=%d pos.Column=%d startOffset=%d charPos=%d",
				code, stringContent, start, pos.Column, startOffset, charPos)
			// Emit string tokens with format specifiers highlighted separately
			s.emitStringWithFormatSpecifiers(stringContent, pos.Line-1, charPos, tokens)
			continue
		}

		// String literal (backtick/raw string)
		if ch == '`' {
			start := i
			i++ // skip opening backtick
			for i < len(code) && code[i] != '`' {
				i++
			}
			if i < len(code) {
				i++ // skip closing backtick
			}
			stringContent := code[start:i]
			charPos := pos.Column - 1 + startOffset + start
			// Emit string tokens with format specifiers highlighted separately
			s.emitStringWithFormatSpecifiers(stringContent, pos.Line-1, charPos, tokens)
			continue
		}

		// Rune literal (single-quoted)
		if ch == '\'' {
			start := i
			i++ // skip opening quote
			for i < len(code) && code[i] != '\'' {
				if code[i] == '\\' && i+1 < len(code) {
					i += 2 // skip escaped char
				} else {
					i++
				}
			}
			if i < len(code) {
				i++ // skip closing quote
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, semanticToken{
				line:      pos.Line - 1,
				startChar: charPos,
				length:    i - start,
				tokenType: tokenTypeString,
				modifiers: 0,
			})
			continue
		}

		// Number literal
		if isDigit(ch) || (ch == '.' && i+1 < len(code) && isDigit(code[i+1])) {
			start := i
			// Handle hex, octal, binary prefixes
			if ch == '0' && i+1 < len(code) {
				next := code[i+1]
				if next == 'x' || next == 'X' {
					i += 2
					for i < len(code) && isHexDigit(code[i]) {
						i++
					}
				} else if next == 'b' || next == 'B' {
					i += 2
					for i < len(code) && (code[i] == '0' || code[i] == '1') {
						i++
					}
				} else if next == 'o' || next == 'O' {
					i += 2
					for i < len(code) && code[i] >= '0' && code[i] <= '7' {
						i++
					}
				} else {
					// Regular number or octal
					for i < len(code) && (isDigit(code[i]) || code[i] == '.' || code[i] == 'e' || code[i] == 'E' || code[i] == '+' || code[i] == '-') {
						i++
					}
				}
			} else {
				// Regular decimal or float
				for i < len(code) && (isDigit(code[i]) || code[i] == '.' || code[i] == 'e' || code[i] == 'E' || code[i] == '+' || code[i] == '-') {
					i++
				}
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, semanticToken{
				line:      pos.Line - 1,
				startChar: charPos,
				length:    i - start,
				tokenType: tokenTypeNumber,
				modifiers: 0,
			})
			continue
		}

		// Identifier or keyword
		if isWordStartChar(ch) {
			start := i
			for i < len(code) && isWordChar(code[i]) {
				i++
			}
			ident := code[start:i]
			charPos := pos.Column - 1 + startOffset + start

			// Check for boolean literals and nil (use number type for literal coloring)
			if ident == "true" || ident == "false" || ident == "nil" {
				*tokens = append(*tokens, semanticToken{
					line:      pos.Line - 1,
					startChar: charPos,
					length:    len(ident),
					tokenType: tokenTypeNumber, // Use number type for consistent literal coloring
					modifiers: 0,
				})
				continue
			}

			// Check if it's a parameter
			if paramNames[ident] {
				*tokens = append(*tokens, semanticToken{
					line:      pos.Line - 1,
					startChar: charPos,
					length:    len(ident),
					tokenType: tokenTypeParameter,
					modifiers: 0,
				})
				continue
			}

			// Check if it's a local variable
			if localVars[ident] {
				*tokens = append(*tokens, semanticToken{
					line:      pos.Line - 1,
					startChar: charPos,
					length:    len(ident),
					tokenType: tokenTypeVariable,
					modifiers: 0,
				})
				continue
			}

			// Check what comes after the identifier (skip whitespace)
			peekIdx := i
			for peekIdx < len(code) && (code[peekIdx] == ' ' || code[peekIdx] == '\t') {
				peekIdx++
			}

			// If followed by '.', this is a package/namespace - leave white (no token)
			if peekIdx < len(code) && code[peekIdx] == '.' {
				continue
			}

			// If followed by '(' or preceded by '.', this is a function call
			isPrecededByDot := start > 0 && code[start-1] == '.'
			isFollowedByParen := peekIdx < len(code) && code[peekIdx] == '('
			if isPrecededByDot || isFollowedByParen {
				*tokens = append(*tokens, semanticToken{
					line:      pos.Line - 1,
					startChar: charPos,
					length:    len(ident),
					tokenType: tokenTypeFunction,
					modifiers: 0,
				})
				continue
			}

			// Check if it's a known function (built-in or indexed)
			if s.isFunctionName(ident) {
				*tokens = append(*tokens, semanticToken{
					line:      pos.Line - 1,
					startChar: charPos,
					length:    len(ident),
					tokenType: tokenTypeFunction,
					modifiers: 0,
				})
				continue
			}

			// Skip other identifiers (types, packages, etc. - let syntax highlighting handle them)
			continue
		}

		// Check for := operator
		if ch == ':' && i+1 < len(code) && code[i+1] == '=' {
			charPos := pos.Column - 1 + startOffset + i
			*tokens = append(*tokens, semanticToken{
				line:      pos.Line - 1,
				startChar: charPos,
				length:    2,
				tokenType: tokenTypeOperator,
				modifiers: 0,
			})
			i += 2
			continue
		}

		// Skip other characters (operators, punctuation, etc.)
		i++
	}
}

// isDigit returns true if c is a decimal digit.
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// safeCharAt returns the character at index i, or "OOB" if out of bounds.
func safeCharAt(s string, i int) string {
	if i < 0 || i >= len(s) {
		return "OOB"
	}
	return string(s[i])
}

// isHexDigit returns true if c is a hexadecimal digit.
func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// isWordStartChar returns true if c can start an identifier.
func isWordStartChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// emitStringWithFormatSpecifiers emits tokens for a string, splitting it into
// string parts and format specifier parts (since VSCode doesn't support overlapping tokens).
func (s *Server) emitStringWithFormatSpecifiers(str string, line int, stringStartChar int, tokens *[]semanticToken) {
	log.Server("emitStringWithFormatSpecifiers: str=%q line=%d startChar=%d", str, line, stringStartChar)
	matches := formatSpecifierRegex.FindAllStringIndex(str, -1)
	log.Server("  format specifier matches: %v", matches)

	if len(matches) == 0 {
		// No format specifiers, emit the whole string as one token
		log.Server("  emitting whole string token: line=%d startChar=%d len=%d type=string", line, stringStartChar, len(str))
		*tokens = append(*tokens, semanticToken{
			line:      line,
			startChar: stringStartChar,
			length:    len(str),
			tokenType: tokenTypeString,
			modifiers: 0,
		})
		return
	}

	// Emit tokens for parts between format specifiers
	idx := 0
	for _, match := range matches {
		// Emit string part before this format specifier
		if match[0] > idx {
			log.Server("  emit STRING token: line=%d startChar=%d length=%d (content=%q)",
				line, stringStartChar+idx, match[0]-idx, str[idx:match[0]])
			*tokens = append(*tokens, semanticToken{
				line:      line,
				startChar: stringStartChar + idx,
				length:    match[0] - idx,
				tokenType: tokenTypeString,
				modifiers: 0,
			})
		}
		// Emit format specifier as number (regexp doesn't work in embedded contexts)
		log.Server("  emit NUMBER token for format spec: line=%d startChar=%d length=%d (content=%q)",
			line, stringStartChar+match[0], match[1]-match[0], str[match[0]:match[1]])
		*tokens = append(*tokens, semanticToken{
			line:      line,
			startChar: stringStartChar + match[0],
			length:    match[1] - match[0],
			tokenType: tokenTypeNumber,
			modifiers: 0,
		})
		idx = match[1]
	}
	// Emit remaining string part after last format specifier
	if idx < len(str) {
		log.Server("  emit STRING token (tail): line=%d startChar=%d length=%d (content=%q)",
			line, stringStartChar+idx, len(str)-idx, str[idx:])
		*tokens = append(*tokens, semanticToken{
			line:      line,
			startChar: stringStartChar + idx,
			length:    len(str) - idx,
			tokenType: tokenTypeString,
			modifiers: 0,
		})
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

// varDecl represents a variable declaration with its position.
type varDecl struct {
	name   string
	offset int // offset from start of code string
}

// extractVarDeclarationsWithPositions extracts variable names and their positions from Go code.
// Handles patterns like "x := 1", "x, y := 1, 2", "var x = 1", "var x, y = 1, 2"
func extractVarDeclarationsWithPositions(code string) []varDecl {
	var decls []varDecl

	// Handle short variable declaration: x := or x, y :=
	if idx := strings.Index(code, ":="); idx > 0 {
		lhs := code[:idx]
		// Split by comma for multiple declarations
		parts := strings.Split(lhs, ",")
		pos := 0
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" && trimmed != "_" && isValidIdentifier(trimmed) {
				// Find actual position of identifier in the original string
				identStart := strings.Index(lhs[pos:], trimmed)
				if identStart >= 0 {
					decls = append(decls, varDecl{
						name:   trimmed,
						offset: pos + identStart,
					})
				}
			}
			pos += len(part) + 1 // +1 for comma
		}
		return decls
	}

	// Handle var declaration: var x = or var x, y =
	trimmed := strings.TrimSpace(code)
	if strings.HasPrefix(trimmed, "var ") {
		varIdx := strings.Index(code, "var ")
		rest := code[varIdx+4:]
		// Find = sign
		if eqIdx := strings.Index(rest, "="); eqIdx > 0 {
			lhs := rest[:eqIdx]
			// Split by comma for multiple declarations
			parts := strings.Split(lhs, ",")
			pos := 0
			for _, part := range parts {
				part = strings.TrimSpace(part)
				// Take first word (variable name before type)
				fields := strings.Fields(part)
				if len(fields) > 0 {
					name := fields[0]
					if name != "_" && isValidIdentifier(name) {
						// Find actual position
						identStart := strings.Index(lhs[pos:], name)
						if identStart >= 0 {
							decls = append(decls, varDecl{
								name:   name,
								offset: varIdx + 4 + pos + identStart,
							})
						}
					}
				}
				pos += len(part) + 1
			}
		}
	}

	return decls
}

// extractVarDeclarations extracts variable names from Go code declarations (legacy).
func extractVarDeclarations(code string) []string {
	decls := extractVarDeclarationsWithPositions(code)
	names := make([]string, len(decls))
	for i, d := range decls {
		names[i] = d.name
	}
	return names
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

// collectTokensFromFuncBody tokenizes the body of a Go function.
// It parses line by line to handle multi-line function bodies correctly.
func (s *Server) collectTokensFromFuncBody(code string, pos tuigen.Position, paramNames map[string]bool, tokens *[]semanticToken) {
	// Find the opening brace of the function body
	braceIdx := strings.Index(code, "{")
	if braceIdx == -1 {
		return
	}

	// Split the code into lines to track line positions
	lines := strings.Split(code, "\n")
	if len(lines) == 0 {
		return
	}

	// Track local variables declared in the function body
	localVars := make(map[string]bool)

	// Find which line and column the opening brace is on
	charCount := 0
	bodyStartLine := 0
	for lineIdx, line := range lines {
		lineEnd := charCount + len(line)
		if braceIdx >= charCount && braceIdx < lineEnd+1 { // +1 for newline
			bodyStartLine = lineIdx
			break
		}
		charCount = lineEnd + 1 // +1 for newline
	}

	// Process each line after the opening brace
	for lineIdx := bodyStartLine; lineIdx < len(lines); lineIdx++ {
		line := lines[lineIdx]

		// Skip if this is just the opening or closing brace line with no content
		trimmed := strings.TrimSpace(line)
		if trimmed == "{" || trimmed == "}" {
			continue
		}

		// Calculate the actual line number in the document
		docLine := pos.Line + lineIdx

		// For the first line of the body (same line as signature), adjust start position
		lineStartCol := 1
		if lineIdx == bodyStartLine {
			// Find position after the opening brace on this line
			braceInLine := strings.Index(line, "{")
			if braceInLine != -1 {
				line = line[braceInLine+1:]
				lineStartCol = pos.Column + braceInLine + 1
			}
		}

		// Extract variable declarations from this line
		varDecls := extractVarDeclarationsWithPositions(line)
		for _, decl := range varDecls {
			localVars[decl.name] = true
			// Emit token for the variable declaration
			*tokens = append(*tokens, semanticToken{
				line:      docLine - 1,
				startChar: lineStartCol - 1 + decl.offset,
				length:    len(decl.name),
				tokenType: tokenTypeVariable,
				modifiers: tokenModDeclaration,
			})
		}

		// Tokenize the line for function calls, parameters, etc.
		linePos := tuigen.Position{Line: docLine, Column: lineStartCol}
		s.collectTokensInGoCode(line, linePos, 0, paramNames, localVars, tokens)
	}
}

// collectCommentGroupTokens adds semantic tokens for all comments in a comment group.
func (s *Server) collectCommentGroupTokens(cg *tuigen.CommentGroup, tokens *[]semanticToken) {
	if cg == nil {
		return
	}
	for _, c := range cg.List {
		s.collectCommentToken(c, tokens)
	}
}

// collectCommentToken adds a semantic token for a single comment.
// For multi-line block comments, we emit a token for each line.
func (s *Server) collectCommentToken(c *tuigen.Comment, tokens *[]semanticToken) {
	if c == nil {
		return
	}

	if !c.IsBlock {
		// Line comment: single token
		*tokens = append(*tokens, semanticToken{
			line:      c.Position.Line - 1,
			startChar: c.Position.Column - 1,
			length:    len(c.Text),
			tokenType: tokenTypeComment,
			modifiers: 0,
		})
		return
	}

	// Block comment: may span multiple lines
	// Split by newlines and emit a token for each line
	lines := strings.Split(c.Text, "\n")
	for i, line := range lines {
		lineNum := c.Position.Line - 1 + i
		var startChar int
		if i == 0 {
			// First line starts at the comment's column
			startChar = c.Position.Column - 1
		} else {
			// Subsequent lines: find the actual start column
			// For multi-line block comments, the text includes leading whitespace
			// We need to calculate based on the original text position
			startChar = 0
			for j := 0; j < len(line) && (line[j] == ' ' || line[j] == '\t'); j++ {
				startChar++
			}
		}
		if len(line) > 0 {
			*tokens = append(*tokens, semanticToken{
				line:      lineNum,
				startChar: startChar,
				length:    len(line),
				tokenType: tokenTypeComment,
				modifiers: 0,
			})
		}
	}
}

// collectAllCommentTokens walks the AST and collects semantic tokens for all comments.
func (s *Server) collectAllCommentTokens(file *tuigen.File, tokens *[]semanticToken) {
	if file == nil {
		return
	}

	// File-level comments
	s.collectCommentGroupTokens(file.LeadingComments, tokens)
	for _, cg := range file.OrphanComments {
		s.collectCommentGroupTokens(cg, tokens)
	}

	// Import trailing comments
	for _, imp := range file.Imports {
		s.collectCommentGroupTokens(imp.TrailingComments, tokens)
	}

	// Component comments
	for _, comp := range file.Components {
		s.collectComponentCommentTokens(comp, tokens)
	}

	// Function comments
	for _, fn := range file.Funcs {
		s.collectCommentGroupTokens(fn.LeadingComments, tokens)
		s.collectCommentGroupTokens(fn.TrailingComments, tokens)
	}
}

// collectComponentCommentTokens collects comment tokens from a component and its body.
func (s *Server) collectComponentCommentTokens(comp *tuigen.Component, tokens *[]semanticToken) {
	if comp == nil {
		return
	}

	s.collectCommentGroupTokens(comp.LeadingComments, tokens)
	s.collectCommentGroupTokens(comp.TrailingComments, tokens)
	for _, cg := range comp.OrphanComments {
		s.collectCommentGroupTokens(cg, tokens)
	}

	// Collect from body nodes
	for _, node := range comp.Body {
		s.collectNodeCommentTokens(node, tokens)
	}
}

// collectNodeCommentTokens collects comment tokens from any AST node.
func (s *Server) collectNodeCommentTokens(node tuigen.Node, tokens *[]semanticToken) {
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
