package tuigen

import (
	"strconv"
	"strings"
)

// Parser parses .tui source files into an AST.
type Parser struct {
	lexer   *Lexer
	current Token
	peek    Token
	errors  *ErrorList
}

// NewParser creates a new Parser for the given lexer.
func NewParser(lexer *Lexer) *Parser {
	p := &Parser{
		lexer:  lexer,
		errors: NewErrorList(),
	}
	// Read two tokens to initialize current and peek
	p.advance()
	p.advance()
	return p
}

// Errors returns any errors encountered during parsing.
func (p *Parser) Errors() *ErrorList {
	return p.errors
}

// advance moves to the next token, skipping newlines where appropriate.
func (p *Parser) advance() {
	p.current = p.peek
	p.peek = p.lexer.Next()
}

// advanceSkipNewlines advances while skipping newline tokens.
func (p *Parser) advanceSkipNewlines() {
	p.advance()
	p.skipNewlines()
}

// skipNewlines consumes any newline tokens.
func (p *Parser) skipNewlines() {
	for p.current.Type == TokenNewline {
		p.advance()
	}
}

// position returns the current token's position.
func (p *Parser) position() Position {
	return Position{
		File:   p.lexer.filename,
		Line:   p.current.Line,
		Column: p.current.Column,
	}
}

// expect checks if the current token matches the expected type and advances.
// Returns true if matched, false otherwise (and records an error).
func (p *Parser) expect(typ TokenType) bool {
	if p.current.Type == typ {
		p.advance()
		return true
	}
	p.errors.AddErrorf(p.position(), "expected %s, got %s", typ, p.current.Type)
	return false
}

// expectSkipNewlines is like expect but skips newlines after advancing.
func (p *Parser) expectSkipNewlines(typ TokenType) bool {
	if !p.expect(typ) {
		return false
	}
	p.skipNewlines()
	return true
}

// ParseFile parses a complete .tui file into a File AST node.
func (p *Parser) ParseFile() (*File, error) {
	file := &File{
		Position: p.position(),
	}

	p.skipNewlines()

	// Parse package declaration
	file.Package = p.parsePackage()
	if file.Package == "" {
		return nil, p.errors.Err()
	}

	p.skipNewlines()

	// Parse imports
	file.Imports = p.parseImports()

	p.skipNewlines()

	// Parse components and top-level functions
	for p.current.Type != TokenEOF {
		p.skipNewlines()
		if p.current.Type == TokenEOF {
			break
		}

		switch p.current.Type {
		case TokenAtComponent:
			comp := p.parseComponent()
			if comp != nil {
				file.Components = append(file.Components, comp)
			}
		case TokenFunc:
			fn := p.parseGoFunc()
			if fn != nil {
				file.Funcs = append(file.Funcs, fn)
			}
		default:
			p.errors.AddErrorf(p.position(), "unexpected token %s, expected @component or func", p.current.Type)
			p.advance()
		}
	}

	// Merge lexer errors
	for _, err := range p.lexer.Errors().Errors() {
		p.errors.Add(err)
	}

	return file, p.errors.Err()
}

// parsePackage parses "package <name>".
func (p *Parser) parsePackage() string {
	if p.current.Type != TokenPackage {
		p.errors.AddError(p.position(), "expected 'package' declaration")
		return ""
	}
	p.advance()

	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected package name")
		return ""
	}
	name := p.current.Literal
	p.advanceSkipNewlines()
	return name
}

// parseImports parses import statements.
// Supports:
//   - import "path"
//   - import alias "path"
//   - import ( "path1"; "path2" )
//   - import ( alias "path" )
func (p *Parser) parseImports() []Import {
	var imports []Import

	for p.current.Type == TokenImport {
		p.advance() // consume 'import'
		p.skipNewlines()

		if p.current.Type == TokenLParen {
			// Grouped imports
			p.advance()
			p.skipNewlines()

			for p.current.Type != TokenRParen && p.current.Type != TokenEOF {
				imp := p.parseSingleImport()
				if imp != nil {
					imports = append(imports, *imp)
				}
				p.skipNewlines()
			}
			p.expect(TokenRParen)
		} else {
			// Single import
			imp := p.parseSingleImport()
			if imp != nil {
				imports = append(imports, *imp)
			}
		}
		p.skipNewlines()
	}

	return imports
}

// parseSingleImport parses a single import: [alias] "path"
func (p *Parser) parseSingleImport() *Import {
	pos := p.position()
	var alias string

	// Check for alias
	if p.current.Type == TokenIdent {
		alias = p.current.Literal
		p.advance()
	}

	// Expect string path
	if p.current.Type != TokenString {
		p.errors.AddError(p.position(), "expected import path string")
		return nil
	}

	path := p.current.Literal
	p.advance()

	return &Import{
		Alias:    alias,
		Path:     path,
		Position: pos,
	}
}

// parseComponent parses a @component definition.
func (p *Parser) parseComponent() *Component {
	pos := p.position()

	if !p.expect(TokenAtComponent) {
		return nil
	}

	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected component name")
		return nil
	}

	comp := &Component{
		Name:       p.current.Literal,
		ReturnType: "*element.Element",
		Position:   pos,
	}
	p.advance()

	// Parse parameters
	if !p.expect(TokenLParen) {
		return nil
	}

	comp.Params = p.parseParams()

	if !p.expect(TokenRParen) {
		return nil
	}

	p.skipNewlines()

	// Parse body
	if !p.expect(TokenLBrace) {
		return nil
	}

	p.skipNewlines()
	comp.Body = p.parseComponentBody()

	if !p.expectSkipNewlines(TokenRBrace) {
		return nil
	}

	return comp
}

// parseParams parses function parameters.
func (p *Parser) parseParams() []*Param {
	var params []*Param

	for p.current.Type != TokenRParen && p.current.Type != TokenEOF {
		param := p.parseParam()
		if param != nil {
			params = append(params, param)
		}

		if p.current.Type == TokenComma {
			p.advance()
		} else {
			break
		}
	}

	return params
}

// parseParam parses a single parameter: name Type
func (p *Parser) parseParam() *Param {
	pos := p.position()

	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected parameter name")
		return nil
	}

	name := p.current.Literal
	p.advance()

	// Parse type (could be complex like *element.Element, []string, func())
	typeStr := p.parseType()
	if typeStr == "" {
		return nil
	}

	return &Param{
		Name:     name,
		Type:     typeStr,
		Position: pos,
	}
}

// parseType parses a Go type expression.
func (p *Parser) parseType() string {
	var sb strings.Builder

	// Handle pointer
	for p.current.Type == TokenStar {
		sb.WriteString("*")
		p.advance()
	}

	// Handle slice
	if p.current.Type == TokenLBracket {
		sb.WriteString("[")
		p.advance()
		if p.current.Type == TokenRBracket {
			sb.WriteString("]")
			p.advance()
		} else {
			// Could be [N]Type or map[K]V
			sb.WriteString(p.current.Literal)
			p.advance()
			if p.current.Type == TokenRBracket {
				sb.WriteString("]")
				p.advance()
			}
		}
	}

	// Handle func type
	if p.current.Type == TokenFunc {
		sb.WriteString("func")
		p.advance()
		if p.current.Type == TokenLParen {
			sb.WriteString("(")
			p.advance()
			depth := 1
			for depth > 0 && p.current.Type != TokenEOF {
				if p.current.Type == TokenLParen {
					depth++
				} else if p.current.Type == TokenRParen {
					depth--
				}
				sb.WriteString(p.current.Literal)
				p.advance()
			}
		}
		return sb.String()
	}

	// Handle map type
	if p.current.Type == TokenIdent && p.current.Literal == "map" {
		sb.WriteString("map")
		p.advance()
		if p.current.Type == TokenLBracket {
			sb.WriteString("[")
			p.advance()
			// Parse key type
			keyType := p.parseType()
			sb.WriteString(keyType)
			if p.current.Type == TokenRBracket {
				sb.WriteString("]")
				p.advance()
			}
			// Parse value type
			valType := p.parseType()
			sb.WriteString(valType)
		}
		return sb.String()
	}

	// Base type (identifier, possibly qualified like pkg.Type)
	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected type")
		return sb.String()
	}

	sb.WriteString(p.current.Literal)
	p.advance()

	// Check for qualified type (pkg.Type)
	for p.current.Type == TokenDot {
		sb.WriteString(".")
		p.advance()
		if p.current.Type == TokenIdent {
			sb.WriteString(p.current.Literal)
			p.advance()
		}
	}

	return sb.String()
}

// parseComponentBody parses the body of a component.
func (p *Parser) parseComponentBody() []Node {
	var nodes []Node

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		p.skipNewlines()
		if p.current.Type == TokenRBrace || p.current.Type == TokenEOF {
			break
		}

		node := p.parseBodyNode()
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// parseBodyNode parses a single node in a component/element body.
func (p *Parser) parseBodyNode() Node {
	switch p.current.Type {
	case TokenLAngle:
		return p.parseElement()
	case TokenAtLet:
		return p.parseLet()
	case TokenAtFor:
		return p.parseFor()
	case TokenAtIf:
		return p.parseIf()
	case TokenLBrace:
		return p.parseGoExprNode()
	default:
		p.errors.AddErrorf(p.position(), "unexpected token %s in body", p.current.Type)
		p.advance()
		return nil
	}
}

// parseElement parses an XML-like element.
func (p *Parser) parseElement() *Element {
	pos := p.position()

	if !p.expect(TokenLAngle) {
		return nil
	}

	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected element tag name")
		return nil
	}

	elem := &Element{
		Tag:      p.current.Literal,
		Position: pos,
	}
	p.advance()

	// Parse attributes
	elem.Attributes = p.parseAttributes()

	// Check for self-closing or opening tag
	if p.current.Type == TokenSlashAngle {
		// Self-closing: <tag />
		elem.SelfClose = true
		p.advanceSkipNewlines()
		return elem
	}

	if !p.expect(TokenRAngle) {
		return nil
	}

	// Parse children
	elem.Children = p.parseChildren(elem.Tag)

	// Expect closing tag </tag>
	if p.current.Type != TokenLAngleSlash {
		p.errors.AddErrorf(p.position(), "expected closing tag </%s>", elem.Tag)
		return elem
	}
	p.advance()

	if p.current.Type != TokenIdent || p.current.Literal != elem.Tag {
		p.errors.AddErrorf(p.position(), "mismatched closing tag: expected </%s>, got </%s>", elem.Tag, p.current.Literal)
	}
	if p.current.Type == TokenIdent {
		p.advance()
	}

	if !p.expect(TokenRAngle) {
		return elem
	}

	p.skipNewlines()
	return elem
}

// parseAttributes parses element attributes.
func (p *Parser) parseAttributes() []*Attribute {
	var attrs []*Attribute

	for {
		// Skip newlines between attributes (for multi-line attribute lists)
		p.skipNewlines()

		// Stop if we hit end of attributes (> or /> or EOF)
		if p.current.Type != TokenIdent {
			break
		}

		attr := p.parseAttribute()
		if attr != nil {
			attrs = append(attrs, attr)
		}
	}

	return attrs
}

// parseAttribute parses a single attribute: name=value or name={expr} or just name (boolean)
func (p *Parser) parseAttribute() *Attribute {
	pos := p.position()

	if p.current.Type != TokenIdent {
		return nil
	}

	name := p.current.Literal
	p.advance()

	// Check for shorthand boolean attribute (no =)
	if p.current.Type != TokenEquals {
		return &Attribute{
			Name:     name,
			Value:    &BoolLit{Value: true, Position: pos},
			Position: pos,
		}
	}

	p.advance() // consume =

	// Parse value
	var value Node
	switch p.current.Type {
	case TokenString:
		value = &StringLit{Value: p.current.Literal, Position: p.position()}
		p.advance()
	case TokenInt:
		v, _ := strconv.ParseInt(p.current.Literal, 10, 64)
		value = &IntLit{Value: v, Position: p.position()}
		p.advance()
	case TokenFloat:
		v, _ := strconv.ParseFloat(p.current.Literal, 64)
		value = &FloatLit{Value: v, Position: p.position()}
		p.advance()
	case TokenLBrace:
		value = p.parseGoExprNode()
	case TokenIdent:
		// Could be true, false, or other identifier
		if p.current.Literal == "true" {
			value = &BoolLit{Value: true, Position: p.position()}
			p.advance()
		} else if p.current.Literal == "false" {
			value = &BoolLit{Value: false, Position: p.position()}
			p.advance()
		} else {
			// Treat as identifier expression
			value = &GoExpr{Code: p.current.Literal, Position: p.position()}
			p.advance()
		}
	default:
		p.errors.AddErrorf(p.position(), "expected attribute value, got %s", p.current.Type)
		return nil
	}

	return &Attribute{
		Name:     name,
		Value:    value,
		Position: pos,
	}
}

// parseChildren parses children inside an element until the closing tag.
func (p *Parser) parseChildren(parentTag string) []Node {
	var children []Node

	for {
		// Skip whitespace-only newlines
		p.skipNewlines()

		// Check for closing tag
		if p.current.Type == TokenLAngleSlash || p.current.Type == TokenEOF {
			break
		}

		// Parse child based on token type
		switch p.current.Type {
		case TokenLAngle:
			// Nested element
			elem := p.parseElement()
			if elem != nil {
				children = append(children, elem)
			}
		case TokenLBrace:
			// Go expression
			expr := p.parseGoExprNode()
			if expr != nil {
				children = append(children, expr)
			}
		case TokenAtLet:
			let := p.parseLet()
			if let != nil {
				children = append(children, let)
			}
		case TokenAtFor:
			f := p.parseFor()
			if f != nil {
				children = append(children, f)
			}
		case TokenAtIf:
			i := p.parseIf()
			if i != nil {
				children = append(children, i)
			}
		case TokenIdent:
			// Text content
			text := &TextContent{
				Text:     p.current.Literal,
				Position: p.position(),
			}
			children = append(children, text)
			p.advance()
		default:
			// Try to collect text content
			if p.current.Type != TokenEOF && p.current.Type != TokenLAngleSlash {
				// Skip unknown tokens in children context
				p.advance()
			} else {
				break
			}
		}
	}

	return children
}

// parseGoExprNode parses a Go expression {expr} as a node.
func (p *Parser) parseGoExprNode() *GoExpr {
	pos := p.position()

	if p.current.Type != TokenLBrace {
		p.errors.AddError(p.position(), "expected '{'")
		return nil
	}

	// The parser has already peeked ahead, consuming some characters via Next().
	// We need to reconstruct the expression by:
	// 1. Including any token that was peeked (the first token after {)
	// 2. Using ReadGoExpr to get the rest

	var exprParts []string

	// If peek isn't } or EOF, it's part of the expression
	if p.peek.Type != TokenRBrace && p.peek.Type != TokenEOF {
		exprParts = append(exprParts, p.peek.Literal)
	}

	// Now read the rest of the expression
	tok := p.lexer.ReadGoExpr()
	if tok.Type == TokenError {
		return nil
	}

	exprParts = append(exprParts, tok.Literal)
	code := strings.Join(exprParts, "")

	// Update parser state after lexer advanced
	p.peek = p.lexer.Next()
	p.advance()

	return &GoExpr{
		Code:     code,
		Position: pos,
	}
}

// parseLet parses @let name = <element>
func (p *Parser) parseLet() *LetBinding {
	pos := p.position()

	if !p.expect(TokenAtLet) {
		return nil
	}

	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected variable name after @let")
		return nil
	}

	name := p.current.Literal
	p.advance()

	if !p.expect(TokenEquals) {
		return nil
	}

	p.skipNewlines()

	if p.current.Type != TokenLAngle {
		p.errors.AddError(p.position(), "expected element after @let =")
		return nil
	}

	elem := p.parseElement()
	if elem == nil {
		return nil
	}

	return &LetBinding{
		Name:     name,
		Element:  elem,
		Position: pos,
	}
}

// parseFor parses @for i, v := range items { ... }
func (p *Parser) parseFor() *ForLoop {
	pos := p.position()

	if !p.expect(TokenAtFor) {
		return nil
	}

	loop := &ForLoop{Position: pos}

	// Parse loop variables
	// Could be: @for i, v := range items
	//       or: @for _, v := range items
	//       or: @for i := range items

	var firstVar string
	if p.current.Type == TokenUnderscore {
		firstVar = "_"
		p.advance()
	} else if p.current.Type == TokenIdent {
		firstVar = p.current.Literal
		p.advance()
	} else {
		p.errors.AddError(p.position(), "expected loop variable")
		return nil
	}

	// Check for comma (two variables)
	if p.current.Type == TokenComma {
		p.advance()
		loop.Index = firstVar

		if p.current.Type == TokenUnderscore {
			loop.Value = "_"
			p.advance()
		} else if p.current.Type == TokenIdent {
			loop.Value = p.current.Literal
			p.advance()
		} else {
			p.errors.AddError(p.position(), "expected second loop variable")
			return nil
		}
	} else {
		// Single variable - it's the value
		loop.Index = ""
		loop.Value = firstVar
	}

	// Expect :=
	if !p.expect(TokenColonEquals) {
		return nil
	}

	// Expect range
	if p.current.Type != TokenRange {
		p.errors.AddError(p.position(), "expected 'range'")
		return nil
	}
	p.advance()

	// Parse iterable expression until {
	var iterParts []string
	for p.current.Type != TokenLBrace && p.current.Type != TokenEOF && p.current.Type != TokenNewline {
		iterParts = append(iterParts, p.current.Literal)
		p.advance()
	}
	loop.Iterable = strings.TrimSpace(strings.Join(iterParts, ""))

	p.skipNewlines()

	// Parse body
	if !p.expect(TokenLBrace) {
		return nil
	}

	p.skipNewlines()
	loop.Body = p.parseComponentBody()

	if !p.expectSkipNewlines(TokenRBrace) {
		return nil
	}

	return loop
}

// parseIf parses @if condition { ... } @else { ... }
func (p *Parser) parseIf() *IfStmt {
	pos := p.position()

	if !p.expect(TokenAtIf) {
		return nil
	}

	stmt := &IfStmt{Position: pos}

	// Parse condition until {
	// We need to be smart about spacing - don't add spaces around operators
	var condParts []string
	for p.current.Type != TokenLBrace && p.current.Type != TokenEOF && p.current.Type != TokenNewline {
		condParts = append(condParts, p.current.Literal)
		p.advance()
	}
	// Join without spaces, but we need to handle this differently
	// Actually, we need to join smartly based on token types
	stmt.Condition = joinConditionTokens(condParts)

	p.skipNewlines()

	// Parse then body
	if !p.expect(TokenLBrace) {
		return nil
	}

	p.skipNewlines()
	stmt.Then = p.parseComponentBody()

	if !p.expectSkipNewlines(TokenRBrace) {
		return nil
	}

	// Check for @else
	if p.current.Type == TokenAtElse {
		p.advance()
		p.skipNewlines()

		// Check for else-if
		if p.current.Type == TokenAtIf {
			elseIf := p.parseIf()
			if elseIf != nil {
				stmt.Else = []Node{elseIf}
			}
		} else {
			// Regular else
			if !p.expect(TokenLBrace) {
				return stmt
			}

			p.skipNewlines()
			stmt.Else = p.parseComponentBody()

			if !p.expectSkipNewlines(TokenRBrace) {
				return stmt
			}
		}
	}

	return stmt
}

// parseGoFunc parses a top-level Go function.
func (p *Parser) parseGoFunc() *GoFunc {
	pos := p.position()

	var code strings.Builder
	code.WriteString("func")
	p.advance()

	// Collect the entire function definition
	braceDepth := 0
	started := false

	for p.current.Type != TokenEOF {
		if p.current.Type == TokenLBrace {
			braceDepth++
			started = true
		} else if p.current.Type == TokenRBrace {
			braceDepth--
		}

		if p.current.Type == TokenNewline {
			code.WriteString("\n")
		} else {
			code.WriteString(p.current.Literal)
		}
		p.advance()

		// Check if function body is complete
		if started && braceDepth == 0 {
			break
		}

		// Add spacing between tokens (rough approximation)
		if p.current.Type != TokenEOF && p.current.Type != TokenNewline &&
			p.current.Type != TokenLParen && p.current.Type != TokenRParen &&
			p.current.Type != TokenLBrace && p.current.Type != TokenRBrace {
			code.WriteString(" ")
		}
	}

	p.skipNewlines()

	return &GoFunc{
		Code:     code.String(),
		Position: pos,
	}
}

// joinConditionTokens joins condition tokens smartly, adding spaces only where needed.
// Operators like !=, ==, <=, >= should not have spaces inserted between their parts.
func joinConditionTokens(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	var result strings.Builder
	for i, part := range parts {
		if i > 0 {
			prev := parts[i-1]
			// Don't add space if current or previous is an operator character
			// that should be adjacent (like ! before =, or = after !)
			needsSpace := true

			// Check if we're building a compound operator
			if (prev == "!" || prev == "=" || prev == "<" || prev == ">" || prev == ":" || prev == "&" || prev == "|") &&
				(part == "=" || part == "&" || part == "|") {
				needsSpace = false
			}
			// Don't add space before/after dots (for qualified names like pkg.Type)
			if prev == "." || part == "." {
				needsSpace = false
			}
			// Don't add space around parens
			if prev == "(" || part == "(" || prev == ")" || part == ")" {
				needsSpace = false
			}
			// Don't add space around brackets
			if prev == "[" || part == "[" || prev == "]" || part == "]" {
				needsSpace = false
			}

			if needsSpace {
				result.WriteString(" ")
			}
		}
		result.WriteString(part)
	}
	return result.String()
}
