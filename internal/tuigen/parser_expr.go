package tuigen

import "strings"

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
		case TokenLParen, TokenLBracket:
			depth++
		case TokenLBrace:
			if isForLoop {
				inForHeader = false // entered for body, semicolons can now terminate
			}
			depth++
		case TokenRParen, TokenRBracket:
			depth--
		case TokenRBrace:
			if depth == 0 {
				// Closing brace at depth 0 belongs to the parent (e.g., if/for body).
				// Stop before consuming it so the parent parser can handle it.
				code := strings.TrimSpace(p.lexer.SourceRange(startPos, p.current.StartPos))
				if code == "" {
					return nil
				}
				return &GoCode{Code: code, Position: pos}
			}
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

// isRangeForLoop reports whether the current `for` opens a `:= range` loop,
// judged from the header tokens (for x[, y] := range). Go 1.22 forms like
// `for range ch {` and C-style loops return false and fall through to GoCode.
// A source scan for ":= range" used to run into a later loop when prose
// containing "for" preceded it on the same line.
func (p *Parser) isRangeForLoop() bool {
	saved := p.saveState()
	defer p.restoreState(saved)
	p.advance() // for
	if !p.isLoopVar() {
		return false
	}
	p.advance()
	if p.current.Type == TokenComma {
		p.advance()
		if !p.isLoopVar() {
			return false
		}
		p.advance()
	}
	if p.current.Type != TokenColonEquals {
		return false
	}
	p.advance()
	return p.current.Type == TokenRange
}

func (p *Parser) isLoopVar() bool {
	return p.current.Type == TokenIdent || p.current.Type == TokenUnderscore
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
	openLine := p.current.Line
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
	rawArgs := p.lexer.SourceRange(argsStart, p.current.StartPos)
	args := strings.TrimSpace(rawArgs)
	multiLineArgs := p.current.Line != openLine

	// Compute the source position of the first character of the trimmed args.
	var argsPos Position
	if args != "" {
		leading := len(rawArgs) - len(strings.TrimLeft(rawArgs, " \t\n\r"))
		argsPos = p.lexer.PositionAt(argsStart + leading)
	}

	if !p.expect(TokenRParen) {
		return nil
	}

	// A children block must open on the same line as the closing paren;
	// a brace on the next line belongs to a sibling node.
	hasBlock := p.current.Type == TokenLBrace
	p.skipNewlines()

	call := &ComponentCall{
		Name:          name,
		Args:          args,
		ArgsPosition:  argsPos,
		MultiLineArgs: multiLineArgs,
		IsStructMount: p.inMethodTempl,
		Position:      pos,
	}

	// Optional children block
	if hasBlock {
		p.advance()
		p.skipNewlines()
		call.Children = p.parseComponentBody()
		if !p.expect(TokenRBrace) {
			return nil
		}
	}

	return call
}

// parseComponentExpr parses @expr where expr is a Component field/variable.
// The lexer has already captured the full expression (e.g., "c.textarea").
func (p *Parser) parseComponentExpr() *ComponentExpr {
	pos := p.position()

	// Current token is TokenAtExpr with the expression as Literal
	expr := p.current.Literal
	p.advance()

	return &ComponentExpr{
		Expr:     expr,
		Position: pos,
	}
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

// isTextOrKeywordToken reports whether typ can begin or continue element text. Go
// keywords count too: control flow is tried before text, so a keyword reaching
// the text path is prose ("wait for it", "Press Escape to return").
func isTextOrKeywordToken(typ TokenType) bool {
	if isTextToken(typ) {
		return true
	}
	switch typ {
	case TokenFor, TokenIf, TokenElse, TokenFunc, TokenReturn, TokenVar,
		TokenRange, TokenTypeKw, TokenConst, TokenPackage, TokenImport, TokenTempl:
		return true
	default:
		return false
	}
}

// isTextToken returns true if the token type is part of text content inside elements.
// Text content can include identifiers and various punctuation that might appear in
// user-facing text like "Use j/k to scroll, q to quit".
func isTextToken(typ TokenType) bool {
	switch typ {
	case TokenIdent, TokenInt, TokenFloat, TokenSymbol:
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
