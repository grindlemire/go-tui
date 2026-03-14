package tuigen

import (
	"strings"
)

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

	if p.current.Type != TokenFor {
		p.errors.AddError(p.position(), "expected 'for'")
		return nil
	}
	p.advance()

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

	// Capture iterable as raw source from current position until {
	iterStart := p.current.StartPos
	for p.current.Type != TokenLBrace && p.current.Type != TokenEOF && p.current.Type != TokenNewline {
		p.advance()
	}
	loop.Iterable = strings.TrimSpace(p.lexer.SourceRange(iterStart, p.current.StartPos))

	p.skipNewlines()

	// Parse body
	openBraceLine := p.current.Line
	if !p.expect(TokenLBrace) {
		return nil
	}

	// Check for trailing comment on same line as opening brace
	loop.TrailingComments = p.getTrailingCommentOnLine(openBraceLine)

	p.skipNewlines()
	loop.Body, loop.OrphanComments = p.parseComponentBodyWithOrphans()

	if !p.expect(TokenRBrace) {
		return nil
	}

	return loop
}

// parseIf parses @if condition { ... } @else { ... }
func (p *Parser) parseIf() *IfStmt {
	pos := p.position()

	if p.current.Type != TokenIf {
		p.errors.AddError(p.position(), "expected 'if'")
		return nil
	}
	p.advance()

	stmt := &IfStmt{Position: pos}

	// Capture condition as raw source from current position until {
	condStart := p.current.StartPos
	for p.current.Type != TokenLBrace && p.current.Type != TokenEOF && p.current.Type != TokenNewline {
		p.advance()
	}
	stmt.Condition = strings.TrimSpace(p.lexer.SourceRange(condStart, p.current.StartPos))

	p.skipNewlines()

	// Parse then body
	openBraceLine := p.current.Line
	if !p.expect(TokenLBrace) {
		return nil
	}

	// Check for trailing comment on same line as opening brace
	stmt.TrailingComments = p.getTrailingCommentOnLine(openBraceLine)

	p.skipNewlines()
	stmt.Then, stmt.OrphanComments = p.parseComponentBodyWithOrphans()

	if !p.expect(TokenRBrace) {
		return nil
	}

	// Skip newlines before checking for @else
	p.skipNewlines()

	// Check for else
	if p.current.Type == TokenElse {
		p.advance()
		p.skipNewlines()

		// Check for else-if
		if p.current.Type == TokenIf {
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
			// Note: else orphan comments are stored in the nested IfStmt or lost
			// This is a simplification - full support would require a separate field
			stmt.Else = p.parseComponentBody()

			if !p.expect(TokenRBrace) {
				return stmt
			}
		}
	}

	return stmt
}

// parseShortBinding parses name := <element> or name := @Component().
// Called when parser sees TokenIdent followed by TokenColonEquals followed by
// TokenLAngle (element) or TokenAtCall/TokenAtExpr (component call).
// The identifier and := have already been consumed; name and pos are passed in.
func (p *Parser) parseShortBinding(name string, pos Position) *LetBinding {
	// := was already consumed by caller
	p.skipNewlines()

	binding := &LetBinding{
		Name:        name,
		IsShortForm: true,
		Position:    pos,
	}

	switch p.current.Type {
	case TokenLAngle:
		elem := p.parseElement()
		if elem == nil {
			return nil
		}
		binding.Element = elem
	case TokenAtCall:
		call := p.parseComponentCall()
		if call == nil {
			return nil
		}
		binding.Call = call
	case TokenAtExpr:
		expr := p.parseComponentExpr()
		if expr == nil {
			return nil
		}
		binding.Expr = expr.Expr
	default:
		p.errors.AddErrorf(p.position(), "expected element or component after :=")
		return nil
	}

	return binding
}

// parseVarBinding parses var name = <element> or var name = @Component().
// Called when parser sees TokenVar followed by TokenIdent followed by TokenEquals
// followed by TokenLAngle or TokenAtCall/TokenAtExpr.
// Sets IsVarForm=true to distinguish from @let bindings in semantic highlighting.
func (p *Parser) parseVarBinding() *LetBinding {
	pos := p.position()
	p.advance() // consume "var"

	if p.current.Type != TokenIdent {
		p.errors.AddError(p.position(), "expected variable name after var")
		return nil
	}
	name := p.current.Literal
	p.advance()

	if p.current.Type != TokenEquals {
		return nil // not a var-binding form; caller will restore state
	}
	p.advance()
	p.skipNewlines()

	binding := &LetBinding{
		Name:        name,
		IsShortForm: false,
		IsVarForm:   true,
		Position:    pos,
	}

	switch p.current.Type {
	case TokenLAngle:
		elem := p.parseElement()
		if elem == nil {
			return nil
		}
		binding.Element = elem
	case TokenAtCall:
		call := p.parseComponentCall()
		if call == nil {
			return nil
		}
		binding.Call = call
	case TokenAtExpr:
		expr := p.parseComponentExpr()
		if expr == nil {
			return nil
		}
		binding.Expr = expr.Expr
	default:
		return nil
	}

	return binding
}
