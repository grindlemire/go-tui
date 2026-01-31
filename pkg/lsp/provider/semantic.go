package provider

import (
	"regexp"
	"sort"
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/lsp/schema"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// formatSpecifierRegex matches Go format specifiers like %s, %d, %v, %.2f, %#x, etc.
var formatSpecifierRegex = regexp.MustCompile(`%[-+# 0]*(\*|\d+)?(\.\*|\.\d+)?[vTtbcdoqxXUeEfFgGsp%]`)

// Semantic token types (must match the order in SemanticTokensLegend.TokenTypes).
const (
	TokenTypeNamespace = 0  // package
	TokenTypeType      = 1  // types
	TokenTypeClass     = 2  // components
	TokenTypeFunction  = 3  // functions
	TokenTypeParameter = 4  // parameters
	TokenTypeVariable  = 5  // variables
	TokenTypeProperty  = 6  // attributes
	TokenTypeKeyword   = 7  // keywords
	TokenTypeString    = 8  // strings
	TokenTypeNumber    = 9  // numbers
	TokenTypeOperator  = 10 // operators
	TokenTypeDecorator = 11 // @ prefix
	TokenTypeRegexp    = 12 // format specifiers (often purple)
	TokenTypeComment   = 13 // comments
)

// Semantic token modifiers (bit flags).
const (
	TokenModDeclaration  = 1 << 0 // where defined
	TokenModDefinition   = 1 << 1 // where defined
	TokenModReadonly     = 1 << 2 // const/let
	TokenModModification = 1 << 3 // where modified
)

// SemanticTokens represents the result of a semantic tokens request.
type SemanticTokens struct {
	Data []int `json:"data"`
}

// SemanticToken represents a single semantic token before encoding.
type SemanticToken struct {
	Line      int
	StartChar int
	Length    int
	TokenType int
	Modifiers int
}

// FunctionNameChecker checks if an identifier is a known function.
type FunctionNameChecker interface {
	IsFunctionName(name string) bool
}

// semanticTokensProvider implements SemanticTokensProvider.
type semanticTokensProvider struct {
	fnChecker FunctionNameChecker
}

// NewSemanticTokensProvider creates a new semantic tokens provider.
func NewSemanticTokensProvider(fnChecker FunctionNameChecker) SemanticTokensProvider {
	return &semanticTokensProvider{fnChecker: fnChecker}
}

func (s *semanticTokensProvider) SemanticTokensFull(doc *Document) (*SemanticTokens, error) {
	log.Server("=== SemanticTokens provider for %s ===", doc.URI)

	if doc.AST == nil {
		return &SemanticTokens{Data: []int{}}, nil
	}

	tokens := s.collectSemanticTokens(doc)

	// Sort tokens by position
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].Line != tokens[j].Line {
			return tokens[i].Line < tokens[j].Line
		}
		return tokens[i].StartChar < tokens[j].StartChar
	})

	// Encode tokens into delta format
	data := EncodeSemanticTokens(tokens)

	if len(data) > 0 {
		log.Server("Encoded %d tokens (%d ints). First 25 values: %v", len(tokens), len(data), data[:min(25, len(data))])
	}

	return &SemanticTokens{Data: data}, nil
}

// collectSemanticTokens collects all semantic tokens from a document.
func (s *semanticTokensProvider) collectSemanticTokens(doc *Document) []SemanticToken {
	var tokens []SemanticToken
	ast := doc.AST

	// Collect comment tokens
	s.collectAllCommentTokens(ast, &tokens)

	// Collect component-related tokens
	for _, comp := range ast.Components {
		// Component keyword (templ)
		tokens = append(tokens, SemanticToken{
			Line:      comp.Position.Line - 1,
			StartChar: comp.Position.Column - 1,
			Length:    len("templ"),
			TokenType: TokenTypeKeyword,
			Modifiers: 0,
		})

		// Component name (declaration)
		nameStart := comp.Position.Column - 1 + len("templ ")
		tokens = append(tokens, SemanticToken{
			Line:      comp.Position.Line - 1,
			StartChar: nameStart,
			Length:    len(comp.Name),
			TokenType: TokenTypeClass,
			Modifiers: TokenModDeclaration | TokenModDefinition,
		})

		// Parameters (declarations)
		for _, param := range comp.Params {
			tokens = append(tokens, SemanticToken{
				Line:      param.Position.Line - 1,
				StartChar: param.Position.Column - 1,
				Length:    len(param.Name),
				TokenType: TokenTypeParameter,
				Modifiers: TokenModDeclaration,
			})
		}

		// Collect tokens from body
		s.collectTokensFromNodes(comp.Body, comp.Params, &tokens)
	}

	// Collect function-related tokens
	for _, fn := range ast.Funcs {
		// func keyword
		tokens = append(tokens, SemanticToken{
			Line:      fn.Position.Line - 1,
			StartChar: fn.Position.Column - 1,
			Length:    len("func"),
			TokenType: TokenTypeKeyword,
			Modifiers: 0,
		})

		// Function name
		name, _, params, _ := parseFuncSignatureForTokens(fn.Code)
		if name != "" {
			nameStart := fn.Position.Column - 1 + len("func ")
			tokens = append(tokens, SemanticToken{
				Line:      fn.Position.Line - 1,
				StartChar: nameStart,
				Length:    len(name),
				TokenType: TokenTypeFunction,
				Modifiers: TokenModDeclaration | TokenModDefinition,
			})

			// Function parameters
			paramStart := nameStart + len(name) + 1 // +1 for '('
			for _, p := range params {
				tokens = append(tokens, SemanticToken{
					Line:      fn.Position.Line - 1,
					StartChar: paramStart,
					Length:    len(p.Name),
					TokenType: TokenTypeParameter,
					Modifiers: TokenModDeclaration,
				})
				paramStart += len(p.Name) + 1 + len(p.Type) + 2
			}

			// Build parameter names map for body tokenization
			paramNames := make(map[string]bool)
			for _, p := range params {
				paramNames[p.Name] = true
			}

			// Tokenize function body
			s.collectTokensFromFuncBody(fn.Code, fn.Position, paramNames, &tokens)
		}
	}

	return tokens
}

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
		// Named ref (#Name) — emit # as operator and ref name as variable declaration
		if n.NamedRef != "" {
			// Find the # position in the line by searching the document content
			// The # appears after the tag name on the element's position line
			line := n.Position.Line - 1
			tagEnd := n.Position.Column - 1 + len(n.Tag)
			// Emit # as operator token at approximated position after tag
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: tagEnd + 1, // space + #
				Length:    1,
				TokenType: TokenTypeOperator,
				Modifiers: 0,
			})
			// Emit ref name as variable with declaration modifier
			*tokens = append(*tokens, SemanticToken{
				Line:      line,
				StartChar: tagEnd + 2, // space + # + Name
				Length:    len(n.NamedRef),
				TokenType: TokenTypeVariable,
				Modifiers: TokenModDeclaration,
			})
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
		for _, child := range n.Else {
			s.collectTokensFromNode(child, paramNames, localVars, tokens)
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
		// Check if this is a state variable declaration for distinct highlighting
		isStateDecl := strings.Contains(n.Code, "tui.NewState(")
		varDecls := extractVarDeclarationsWithPositions(n.Code)
		for _, decl := range varDecls {
			localVars[decl.name] = true
			modifiers := TokenModDeclaration
			if isStateDecl {
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

// collectVariableTokensInCode finds and tokenizes variable references in Go code.
func (s *semanticTokensProvider) collectVariableTokensInCode(code string, pos tuigen.Position, paramNames map[string]bool, localVars map[string]bool, tokens *[]SemanticToken) {
	s.collectTokensInGoCode(code, pos, 1, paramNames, localVars, tokens)
}

// collectTokensInGoCodeDirect tokenizes Go code without any offset adjustment.
func (s *semanticTokensProvider) collectTokensInGoCodeDirect(code string, pos tuigen.Position, paramNames map[string]bool, localVars map[string]bool, tokens *[]SemanticToken) {
	s.collectTokensInGoCode(code, pos, 0, paramNames, localVars, tokens)
}

// collectTokensInGoCode tokenizes Go code for semantic highlighting.
func (s *semanticTokensProvider) collectTokensInGoCode(code string, pos tuigen.Position, startOffset int, paramNames map[string]bool, localVars map[string]bool, tokens *[]SemanticToken) {
	log.Server("collectTokensInGoCode: code=%q pos.Line=%d pos.Column=%d startOffset=%d", code, pos.Line, pos.Column, startOffset)
	i := 0
	for i < len(code) {
		ch := code[i]

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		// Block comment /* ... */
		if ch == '/' && i+1 < len(code) && code[i+1] == '*' {
			start := i
			i += 2
			for i+1 < len(code) && !(code[i] == '*' && code[i+1] == '/') {
				i++
			}
			if i+1 < len(code) {
				i += 2
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    i - start,
				TokenType: TokenTypeComment,
				Modifiers: 0,
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
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    i - start,
				TokenType: TokenTypeComment,
				Modifiers: 0,
			})
			continue
		}

		// String literal (double-quoted)
		if ch == '"' {
			start := i
			i++
			for i < len(code) && code[i] != '"' {
				if code[i] == '\\' && i+1 < len(code) {
					i += 2
				} else {
					i++
				}
			}
			if i < len(code) {
				i++
			}
			stringContent := code[start:i]
			charPos := pos.Column - 1 + startOffset + start
			log.Server("Found string in GoCode: code=%q stringContent=%q start=%d pos.Column=%d startOffset=%d charPos=%d",
				code, stringContent, start, pos.Column, startOffset, charPos)
			s.emitStringWithFormatSpecifiers(stringContent, pos.Line-1, charPos, tokens)
			continue
		}

		// Backtick string
		if ch == '`' {
			start := i
			i++
			for i < len(code) && code[i] != '`' {
				i++
			}
			if i < len(code) {
				i++
			}
			stringContent := code[start:i]
			charPos := pos.Column - 1 + startOffset + start
			s.emitStringWithFormatSpecifiers(stringContent, pos.Line-1, charPos, tokens)
			continue
		}

		// Rune literal
		if ch == '\'' {
			start := i
			i++
			for i < len(code) && code[i] != '\'' {
				if code[i] == '\\' && i+1 < len(code) {
					i += 2
				} else {
					i++
				}
			}
			if i < len(code) {
				i++
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    i - start,
				TokenType: TokenTypeString,
				Modifiers: 0,
			})
			continue
		}

		// Number literal
		if isDigit(ch) || (ch == '.' && i+1 < len(code) && isDigit(code[i+1])) {
			start := i
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
					for i < len(code) && (isDigit(code[i]) || code[i] == '.' || code[i] == 'e' || code[i] == 'E' || code[i] == '+' || code[i] == '-') {
						i++
					}
				}
			} else {
				for i < len(code) && (isDigit(code[i]) || code[i] == '.' || code[i] == 'e' || code[i] == 'E' || code[i] == '+' || code[i] == '-') {
					i++
				}
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    i - start,
				TokenType: TokenTypeNumber,
				Modifiers: 0,
			})
			continue
		}

		// Identifier or keyword
		if isWordStartChar(ch) {
			start := i
			for i < len(code) && isWordCharByte(code[i]) {
				i++
			}
			ident := code[start:i]
			charPos := pos.Column - 1 + startOffset + start

			// Boolean/nil literals
			if ident == "true" || ident == "false" || ident == "nil" {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeNumber,
					Modifiers: 0,
				})
				continue
			}

			if paramNames[ident] {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeParameter,
					Modifiers: 0,
				})
				continue
			}

			if localVars[ident] {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeVariable,
					Modifiers: 0,
				})
				continue
			}

			// Check what comes after the identifier
			peekIdx := i
			for peekIdx < len(code) && (code[peekIdx] == ' ' || code[peekIdx] == '\t') {
				peekIdx++
			}

			// If followed by '.', this is a package/namespace
			if peekIdx < len(code) && code[peekIdx] == '.' {
				continue
			}

			// If followed by '(' or preceded by '.', this is a function call
			isPrecededByDot := start > 0 && code[start-1] == '.'
			isFollowedByParen := peekIdx < len(code) && code[peekIdx] == '('
			if isPrecededByDot || isFollowedByParen {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeFunction,
					Modifiers: 0,
				})
				continue
			}

			// Check if it's a known function
			if s.fnChecker != nil && s.fnChecker.IsFunctionName(ident) {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeFunction,
					Modifiers: 0,
				})
				continue
			}

			continue
		}

		// := operator
		if ch == ':' && i+1 < len(code) && code[i+1] == '=' {
			charPos := pos.Column - 1 + startOffset + i
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    2,
				TokenType: TokenTypeOperator,
				Modifiers: 0,
			})
			i += 2
			continue
		}

		i++
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
		log.Server("  emit NUMBER token for format spec: line=%d startChar=%d length=%d (content=%q)",
			line, stringStartChar+match[0], match[1]-match[0], str[match[0]:match[1]])
		*tokens = append(*tokens, SemanticToken{
			Line:      line,
			StartChar: stringStartChar + match[0],
			Length:    match[1] - match[0],
			TokenType: TokenTypeNumber,
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

// collectTokensFromFuncBody tokenizes the body of a Go function.
func (s *semanticTokensProvider) collectTokensFromFuncBody(code string, pos tuigen.Position, paramNames map[string]bool, tokens *[]SemanticToken) {
	braceIdx := strings.Index(code, "{")
	if braceIdx == -1 {
		return
	}

	lines := strings.Split(code, "\n")
	if len(lines) == 0 {
		return
	}

	localVars := make(map[string]bool)

	charCount := 0
	bodyStartLine := 0
	for lineIdx, line := range lines {
		lineEnd := charCount + len(line)
		if braceIdx >= charCount && braceIdx < lineEnd+1 {
			bodyStartLine = lineIdx
			break
		}
		charCount = lineEnd + 1
	}

	for lineIdx := bodyStartLine; lineIdx < len(lines); lineIdx++ {
		line := lines[lineIdx]
		trimmed := strings.TrimSpace(line)
		if trimmed == "{" || trimmed == "}" {
			continue
		}

		docLine := pos.Line + lineIdx
		lineStartCol := 1
		if lineIdx == bodyStartLine {
			braceInLine := strings.Index(line, "{")
			if braceInLine != -1 {
				line = line[braceInLine+1:]
				lineStartCol = pos.Column + braceInLine + 1
			}
		}

		varDecls := extractVarDeclarationsWithPositions(line)
		for _, decl := range varDecls {
			localVars[decl.name] = true
			*tokens = append(*tokens, SemanticToken{
				Line:      docLine - 1,
				StartChar: lineStartCol - 1 + decl.offset,
				Length:    len(decl.name),
				TokenType: TokenTypeVariable,
				Modifiers: TokenModDeclaration,
			})
		}

		linePos := tuigen.Position{Line: docLine, Column: lineStartCol}
		s.collectTokensInGoCode(line, linePos, 0, paramNames, localVars, tokens)
	}
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

// --- Helper functions ---

// EncodeSemanticTokens encodes tokens into the LSP delta format.
func EncodeSemanticTokens(tokens []SemanticToken) []int {
	if len(tokens) == 0 {
		return []int{}
	}
	data := make([]int, 0, len(tokens)*5)
	prevLine := 0
	prevChar := 0
	for _, t := range tokens {
		deltaLine := t.Line - prevLine
		deltaChar := t.StartChar
		if deltaLine == 0 {
			deltaChar = t.StartChar - prevChar
		}
		data = append(data, deltaLine, deltaChar, t.Length, t.TokenType, t.Modifiers)
		prevLine = t.Line
		prevChar = t.StartChar
	}
	return data
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func isWordStartChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isWordCharByte checks if a byte is a valid word character (for identifiers).
func isWordCharByte(c byte) bool {
	return isWordStartChar(c) || isDigit(c)
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

// varDecl represents a variable declaration with its position.
type varDecl struct {
	name   string
	offset int
}

// extractVarDeclarationsWithPositions extracts variable names and their positions from Go code.
func extractVarDeclarationsWithPositions(code string) []varDecl {
	var decls []varDecl

	if idx := strings.Index(code, ":="); idx > 0 {
		lhs := code[:idx]
		parts := strings.Split(lhs, ",")
		pos := 0
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" && trimmed != "_" && isValidIdentifier(trimmed) {
				identStart := strings.Index(lhs[pos:], trimmed)
				if identStart >= 0 {
					decls = append(decls, varDecl{
						name:   trimmed,
						offset: pos + identStart,
					})
				}
			}
			pos += len(part) + 1
		}
		return decls
	}

	trimmed := strings.TrimSpace(code)
	if strings.HasPrefix(trimmed, "var ") {
		varIdx := strings.Index(code, "var ")
		rest := code[varIdx+4:]
		if eqIdx := strings.Index(rest, "="); eqIdx > 0 {
			lhs := rest[:eqIdx]
			parts := strings.Split(lhs, ",")
			pos := 0
			for _, part := range parts {
				part = strings.TrimSpace(part)
				fields := strings.Fields(part)
				if len(fields) > 0 {
					name := fields[0]
					if name != "_" && isValidIdentifier(name) {
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

// funcParam represents a function parameter for semantic tokenization.
type funcParam struct {
	Name string
	Type string
}

// parseFuncSignatureForTokens extracts function name, signature, params, and return type from code.
func parseFuncSignatureForTokens(code string) (name, signature string, params []funcParam, returns string) {
	code = strings.TrimSpace(code)
	if !strings.HasPrefix(code, "func ") {
		return "", "", nil, ""
	}
	code = code[5:]

	parenIdx := strings.Index(code, "(")
	if parenIdx == -1 {
		return "", "", nil, ""
	}
	name = strings.TrimSpace(code[:parenIdx])
	code = code[parenIdx:]

	// Find matching close paren
	depth := 0
	closeIdx := -1
	for i, c := range code {
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
			if depth == 0 {
				closeIdx = i
				break
			}
		}
	}
	if closeIdx == -1 {
		return name, "", nil, ""
	}

	paramStr := code[1:closeIdx]
	signature = "func " + name + code[:closeIdx+1]

	// Parse parameters
	if paramStr != "" {
		for _, p := range strings.Split(paramStr, ",") {
			p = strings.TrimSpace(p)
			parts := strings.SplitN(p, " ", 2)
			if len(parts) == 2 {
				params = append(params, funcParam{Name: parts[0], Type: parts[1]})
			}
		}
	}

	// Return type
	rest := strings.TrimSpace(code[closeIdx+1:])
	if braceIdx := strings.Index(rest, "{"); braceIdx > 0 {
		returns = strings.TrimSpace(rest[:braceIdx])
	}

	return name, signature, params, returns
}
