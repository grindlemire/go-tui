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

	if !p.expect(TokenAtIf) {
		return nil
	}

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
