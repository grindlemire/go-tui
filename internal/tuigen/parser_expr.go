package tuigen

import (
	"strings"
)

// parseGoExprNode parses a Go expression {expr} as a node.
func (p *Parser) parseGoExprNode() *GoExpr {
	pos := p.position()

	if p.current.Type != TokenLBrace {
		p.errors.AddError(p.position(), "expected '{'")
		return nil
	}

	// Use the token's source position to capture raw content between braces
	bracePos := p.current.StartPos
	code, err := p.lexer.ReadBalancedBracesFrom(bracePos)
	if err != nil {
		p.errors.AddError(pos, err.Error())
		return nil
	}

	// Re-sync parser state after lexer advanced
	p.current = p.lexer.Next()
	p.peek = p.lexer.Next()

	return &GoExpr{
		Code:     strings.TrimSpace(code),
		Position: pos,
	}
}

// parseGoStatement parses a raw Go statement (e.g., fmt.Printf("x"), x := 1).
// Called when we see an identifier or Go keyword at the start of a body line.
// Captures the statement as raw source until newline or semicolon at bracket depth 0.
func (p *Parser) parseGoStatement() *GoCode {
	pos := p.position()
	startPos := p.current.StartPos

	// Track bracket depth to handle multi-line statements
	depth := 0

	// Special handling for 'for' loops: semicolons in the for clause (before the body {)
	// are NOT statement terminators. e.g., "for i := 0; i < 10; i++ { ... }"
	isForLoop := p.current.Type == TokenFor || (p.current.Type == TokenIdent && p.current.Literal == "for")
	inForHeader := isForLoop

	for p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenLParen, TokenLBracket, TokenLBrace:
			if p.current.Type == TokenLBrace && isForLoop {
				inForHeader = false // entered for body, semicolons can now terminate
			}
			depth++
		case TokenRParen, TokenRBracket, TokenRBrace:
			depth--
		case TokenNewline:
			if depth == 0 {
				// End of statement
				code := strings.TrimSpace(p.lexer.SourceRange(startPos, p.current.StartPos))
				p.advance() // consume newline
				return &GoCode{Code: code, Position: pos}
			}
		case TokenSemicolon:
			// Don't terminate on semicolons inside a for-loop header
			if depth == 0 && !inForHeader {
				// End of statement - capture up to but not including semicolon
				code := strings.TrimSpace(p.lexer.SourceRange(startPos, p.current.StartPos))
				p.advance() // consume semicolon
				return &GoCode{Code: code, Position: pos}
			}
		}
		p.advance()
	}

	// EOF reached
	code := strings.TrimSpace(p.lexer.SourceRange(startPos, p.lexer.SourcePos()))
	return &GoCode{Code: code, Position: pos}
}

// parseComponentCall parses @ComponentName(args) or @ComponentName(args) { children }
func (p *Parser) parseComponentCall() *ComponentCall {
	pos := p.position()

	// Current token is TokenAtCall with the component name as Literal
	name := p.current.Literal
	p.advance()

	// Parse arguments (required parentheses, may be empty)
	if !p.expect(TokenLParen) {
		return nil
	}

	// Capture arguments as raw source until matching )
	argsStart := p.current.StartPos
	parenDepth := 1
	for parenDepth > 0 && p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenLParen:
			parenDepth++
		case TokenRParen:
			parenDepth--
			if parenDepth == 0 {
				continue // don't advance, we'll handle below
			}
		}
		if parenDepth > 0 {
			p.advance()
		}
	}
	args := strings.TrimSpace(p.lexer.SourceRange(argsStart, p.current.StartPos))

	if !p.expect(TokenRParen) {
		return nil
	}

	p.skipNewlines()

	call := &ComponentCall{
		Name:          name,
		Args:          args,
		IsStructMount: p.inMethodTempl,
		Position:      pos,
	}

	// Optional children block
	if p.current.Type == TokenLBrace {
		p.advance()
		p.skipNewlines()
		call.Children = p.parseComponentBody()
		if !p.expect(TokenRBrace) {
			return nil
		}
	}

	return call
}

// parseGoExprOrChildrenSlot parses either a Go expression {expr} or children slot {children...}
func (p *Parser) parseGoExprOrChildrenSlot() Node {
	pos := p.position()

	if p.current.Type != TokenLBrace {
		p.errors.AddError(p.position(), "expected '{'")
		return nil
	}

	// Use the token's source position to capture raw content between braces
	bracePos := p.current.StartPos
	code, err := p.lexer.ReadBalancedBracesFrom(bracePos)
	if err != nil {
		p.errors.AddError(pos, err.Error())
		return nil
	}

	// Re-sync parser state after lexer advanced
	p.current = p.lexer.Next()
	p.peek = p.lexer.Next()

	trimmed := strings.TrimSpace(code)

	// Check for children slot syntax
	if trimmed == "children..." {
		return &ChildrenSlot{Position: pos}
	}

	return &GoExpr{
		Code:     trimmed,
		Position: pos,
	}
}

// isTextToken returns true if the token type is part of text content inside elements.
// Text content can include identifiers and various punctuation that might appear in
// user-facing text like "Use j/k to scroll, q to quit".
func isTextToken(typ TokenType) bool {
	switch typ {
	case TokenIdent, TokenInt, TokenFloat:
		return true
	// Punctuation commonly found in text
	case TokenComma, TokenSlash, TokenDot, TokenColon, TokenSemicolon:
		return true
	case TokenBang, TokenMinus, TokenPlus, TokenStar:
		return true
	case TokenPipe, TokenAmpersand, TokenEquals:
		return true
	case TokenLParen, TokenRParen, TokenLBracket, TokenRBracket:
		return true
	case TokenUnderscore:
		return true
	default:
		return false
	}
}

// isWordToken returns true if the token is a "word" that should have spaces between
// consecutive instances. Identifiers and numbers are words; punctuation is not.
func isWordToken(typ TokenType) bool {
	switch typ {
	case TokenIdent, TokenInt, TokenFloat:
		return true
	default:
		return false
	}
}
