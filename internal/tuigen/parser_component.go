package tuigen

import (
	"strings"
)

// parseFuncOrComponent parses a func definition and determines if it's a component or helper function.
// If the return type is exactly "Element", it's parsed as a component with DSL body.
// Otherwise, it's captured as raw Go code.
func (p *Parser) parseFuncOrComponent() Node {
	pos := p.position()
	startPos := p.current.StartPos

	if !p.expect(TokenFunc) {
		return nil
	}

	// Check for method receiver: func (receiver Type) Name()
	// If we see '(' after 'func', this is a method - capture as raw Go
	if p.current.Type == TokenLParen {
		return p.captureRawGoFunc(startPos, pos)
	}

	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected function name")
		return nil
	}

	name := p.current.Literal
	p.advance()

	// Parse parameters
	if !p.expect(TokenLParen) {
		return nil
	}

	params := p.parseParams()

	if !p.expect(TokenRParen) {
		return nil
	}

	// Check for return type
	// Skip newlines but preserve position tracking
	p.skipNewlines()

	returnType := ""
	if p.current.Type == TokenIdent {
		returnType = p.current.Literal
		p.advance()
	}

	p.skipNewlines()

	// Decision: if return type is exactly "Element", parse as component with DSL body
	if returnType == "Element" {
		comp := &Component{
			Name:       name,
			Params:     params,
			ReturnType: "*element.Element", // Internal representation stays the same
			Position:   pos,
		}

		// Parse body as DSL
		openBraceLine := p.current.Line
		if !p.expect(TokenLBrace) {
			return nil
		}

		// Check for trailing comment on the same line as opening brace
		comp.TrailingComments = p.getTrailingCommentOnLine(openBraceLine)

		p.skipNewlines()
		comp.Body, comp.OrphanComments = p.parseComponentBodyWithOrphans()

		if !p.expectSkipNewlines(TokenRBrace) {
			return nil
		}

		return comp
	}

	// Not a component - capture as raw Go function
	// We need to continue from where we left off to capture the full function
	// Skip to matching closing brace
	braceDepth := 0
	started := false

	for p.current.Type != TokenEOF {
		if p.current.Type == TokenLBrace {
			braceDepth++
			started = true
		} else if p.current.Type == TokenRBrace {
			braceDepth--
			if started && braceDepth == 0 {
				// Capture raw source from func to after closing brace
				endPos := p.current.StartPos + 1 // +1 to include the '}'
				code := p.lexer.SourceRange(startPos, endPos)
				p.clearPendingComments()
				p.advance() // move past '}'
				p.skipNewlines()
				return &GoFunc{
					Code:     code,
					Position: pos,
				}
			}
		}
		p.advance()
	}

	// If we reach here, function was not properly closed
	p.errors.AddError(pos, "unterminated function definition")
	code := p.lexer.SourceRange(startPos, p.lexer.SourcePos())
	p.skipNewlines()

	return &GoFunc{
		Code:     code,
		Position: pos,
	}
}

// parseGoDecl parses a top-level Go declaration (type, const, or var).
// These are captured as raw Go code and passed through unchanged.
func (p *Parser) parseGoDecl() *GoDecl {
	pos := p.position()
	startPos := p.current.StartPos
	kind := p.current.Literal // "type", "const", or "var"

	// Track brace/paren depth to find end of declaration
	braceDepth := 0
	parenDepth := 0

	for p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenLBrace:
			braceDepth++
		case TokenRBrace:
			braceDepth--
			if braceDepth == 0 && parenDepth == 0 {
				// End of braced declaration (type struct{}, const/var block)
				endPos := p.current.StartPos + 1
				code := p.lexer.SourceRange(startPos, endPos)
				p.clearPendingComments()
				p.advance()
				p.skipNewlines()
				return &GoDecl{Kind: kind, Code: code, Position: pos}
			}
		case TokenLParen:
			parenDepth++
		case TokenRParen:
			parenDepth--
			if braceDepth == 0 && parenDepth == 0 {
				// End of grouped declaration: const (...) or var (...)
				endPos := p.current.StartPos + 1
				code := p.lexer.SourceRange(startPos, endPos)
				p.clearPendingComments()
				p.advance()
				p.skipNewlines()
				return &GoDecl{Kind: kind, Code: code, Position: pos}
			}
		case TokenNewline:
			// Simple declaration ends at newline (if not inside braces/parens)
			if braceDepth == 0 && parenDepth == 0 {
				endPos := p.current.StartPos
				code := p.lexer.SourceRange(startPos, endPos)
				p.skipNewlines()
				return &GoDecl{Kind: kind, Code: code, Position: pos}
			}
		}
		p.advance()
	}

	// Handle EOF
	code := p.lexer.SourceRange(startPos, p.lexer.SourcePos())
	return &GoDecl{Kind: kind, Code: code, Position: pos}
}

// captureRawGoFunc captures a function definition as raw Go code.
// Used for methods with receivers and helper functions.
func (p *Parser) captureRawGoFunc(startPos int, pos Position) *GoFunc {
	braceDepth := 0
	started := false

	for p.current.Type != TokenEOF {
		if p.current.Type == TokenLBrace {
			braceDepth++
			started = true
		} else if p.current.Type == TokenRBrace {
			braceDepth--
			if started && braceDepth == 0 {
				endPos := p.current.StartPos + 1
				code := p.lexer.SourceRange(startPos, endPos)
				p.clearPendingComments()
				p.advance()
				p.skipNewlines()
				return &GoFunc{Code: code, Position: pos}
			}
		}
		p.advance()
	}

	p.errors.AddError(pos, "unterminated function definition")
	code := p.lexer.SourceRange(startPos, p.lexer.SourcePos())
	return &GoFunc{Code: code, Position: pos}
}

// parseTempl parses a templ definition which is always a component.
// Syntax: templ Name(params) { body }
// No return type is specified - it's always generated as *element.Element.
func (p *Parser) parseTempl() *Component {
	pos := p.position()

	if !p.expect(TokenTempl) {
		return nil
	}

	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected component name")
		return nil
	}

	name := p.current.Literal
	p.advance()

	// Parse parameters
	if !p.expect(TokenLParen) {
		return nil
	}

	params := p.parseParams()

	if !p.expect(TokenRParen) {
		return nil
	}

	p.skipNewlines()

	comp := &Component{
		Name:       name,
		Params:     params,
		ReturnType: "*element.Element",
		Position:   pos,
	}

	// Parse body as DSL
	openBraceLine := p.current.Line
	if !p.expect(TokenLBrace) {
		return nil
	}

	// Check for trailing comment on the same line as opening brace
	comp.TrailingComments = p.getTrailingCommentOnLine(openBraceLine)

	p.skipNewlines()
	comp.Body, comp.OrphanComments = p.parseComponentBodyWithOrphans()

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

// parseType parses a Go type expression by capturing raw source.
// This handles all Go types including generics, channels, and complex function signatures.
func (p *Parser) parseType() string {
	startPos := p.current.StartPos
	depth := 0 // track [], (), {}

	for p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenComma:
			// Comma at depth 0 means end of this type
			if depth == 0 {
				return strings.TrimSpace(p.lexer.SourceRange(startPos, p.current.StartPos))
			}
			p.advance()
		case TokenRParen:
			// Right paren at depth 0 means end of parameter list
			if depth == 0 {
				return strings.TrimSpace(p.lexer.SourceRange(startPos, p.current.StartPos))
			}
			depth--
			p.advance()
		case TokenLBracket, TokenLParen, TokenLBrace:
			depth++
			p.advance()
		case TokenRBracket, TokenRBrace:
			depth--
			p.advance()
		default:
			p.advance()
		}
	}

	return strings.TrimSpace(p.lexer.SourceRange(startPos, p.lexer.SourcePos()))
}
