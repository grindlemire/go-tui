package tuigen

import (
	"strconv"
	"strings"
)

// parseComponentBody parses the body of a component.
func (p *Parser) parseComponentBody() []Node {
	nodes, _ := p.parseComponentBodyWithOrphans()
	return nodes
}

// parseComponentBodyWithOrphans parses the body of a component and also
// returns any orphan comments that weren't attached to nodes.
func (p *Parser) parseComponentBodyWithOrphans() ([]Node, []*CommentGroup) {
	var nodes []Node
	var orphanComments []*CommentGroup

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		p.skipNewlines()
		if p.current.Type == TokenRBrace || p.current.Type == TokenEOF {
			break
		}

		// Collect any pending comments before parsing the next node
		leadingComments := p.getLeadingCommentGroup()

		node := p.parseBodyNode()
		if node != nil {
			// Attach leading comments to the node
			p.attachLeadingComments(node, leadingComments)
			nodes = append(nodes, node)
		} else if leadingComments != nil {
			// No node was parsed, these comments are orphans
			orphanComments = append(orphanComments, leadingComments)
		}
	}

	// Collect any remaining orphan comments before the closing brace
	p.collectPendingComments()
	remaining := p.consumePendingComments()
	if len(remaining) > 0 {
		orphanComments = append(orphanComments, groupComments(remaining)...)
	}

	return nodes, orphanComments
}

// attachLeadingComments attaches leading comments to a node.
// This handles all node types that support LeadingComments.
func (p *Parser) attachLeadingComments(node Node, comments *CommentGroup) {
	if comments == nil {
		return
	}

	switch n := node.(type) {
	case *Element:
		n.LeadingComments = comments
	case *LetBinding:
		n.LeadingComments = comments
	case *ForLoop:
		n.LeadingComments = comments
	case *IfStmt:
		n.LeadingComments = comments
	case *ComponentCall:
		n.LeadingComments = comments
	case *GoCode:
		n.LeadingComments = comments
	case *GoExpr:
		n.LeadingComments = comments
	case *ChildrenSlot:
		n.LeadingComments = comments
	}
}

// parseBodyNode parses a single node in a component/element body.
// NOTE: We explicitly check for nil before returning to avoid the Go interface
// nil gotcha where a typed nil pointer (e.g., (*ComponentCall)(nil)) converted
// to an interface would pass `node != nil` checks in callers.
func (p *Parser) parseBodyNode() Node {
	switch p.current.Type {
	case TokenLAngle:
		if el := p.parseElement(); el != nil {
			return el
		}
	case TokenAtLet:
		if let := p.parseLet(); let != nil {
			return let
		}
	case TokenAtFor:
		if f := p.parseFor(); f != nil {
			return f
		}
	case TokenAtIf:
		if i := p.parseIf(); i != nil {
			return i
		}
	case TokenAtCall:
		if call := p.parseComponentCall(); call != nil {
			return call
		}
	case TokenAtExpr:
		if expr := p.parseComponentExpr(); expr != nil {
			return expr
		}
	case TokenLBrace:
		if node := p.parseGoExprOrChildrenSlot(); node != nil {
			return node
		}
	case TokenIdent, TokenIf, TokenFor, TokenFunc, TokenReturn:
		// Raw Go statement (e.g., fmt.Printf("x"), x := 1, if err != nil {...})
		// Note: go, defer, switch, select are lexed as TokenIdent
		if stmt := p.parseGoStatement(); stmt != nil {
			return stmt
		}
	default:
		p.errors.AddErrorf(p.position(), "unexpected token %s in body", p.current.Type)
		p.advance()
	}
	return nil
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

	// Parse attributes (ref={} and key={} are parsed as regular attributes, then extracted)
	elem.Attributes = p.parseAttributes()

	// Detect multi-line attributes from source positions
	if len(elem.Attributes) > 0 {
		lastAttr := elem.Attributes[len(elem.Attributes)-1]
		elem.MultiLineAttrs = lastAttr.Position.Line != pos.Line
	}

	// Extract ref={expr} attribute and move it to RefExpr
	for i, attr := range elem.Attributes {
		if attr.Name == "ref" {
			if expr, ok := attr.Value.(*GoExpr); ok {
				elem.RefExpr = expr
				// Remove ref from attributes
				elem.Attributes = append(elem.Attributes[:i], elem.Attributes[i+1:]...)
				break
			}
		}
	}

	// Check for key={expr} attribute and move it to RefKey
	for i, attr := range elem.Attributes {
		if attr.Name == "key" {
			if expr, ok := attr.Value.(*GoExpr); ok {
				elem.RefKey = expr
				// Remove key from attributes
				elem.Attributes = append(elem.Attributes[:i], elem.Attributes[i+1:]...)
				break
			}
		}
	}

	// Determine the line of the last thing before the closing bracket
	lastLineBeforeBracket := pos.Line
	if len(elem.Attributes) > 0 {
		lastLineBeforeBracket = elem.Attributes[len(elem.Attributes)-1].Position.Line
	}

	// Check for self-closing or opening tag
	if p.current.Type == TokenSlashAngle {
		// Self-closing: <tag />
		elem.SelfClose = true
		elem.ClosingBracketNewLine = p.current.Line > lastLineBeforeBracket
		closeLine := p.current.Line
		p.advanceSkipNewlines()
		// Check for trailing comment on same line as />
		elem.TrailingComments = p.getTrailingCommentOnLine(closeLine)
		return elem
	}

	// Detect closing bracket on new line for regular elements
	elem.ClosingBracketNewLine = p.current.Line > lastLineBeforeBracket
	openTagEndLine := p.current.Line // line of >

	if !p.expect(TokenRAngle) {
		return nil
	}

	// Parse children
	elem.Children, elem.OrphanComments = p.parseChildren(elem.Tag)

	// Detect inline children from source positions
	if len(elem.Children) > 0 {
		firstChild := elem.Children[0]
		elem.InlineChildren = firstChild.Pos().Line == openTagEndLine
	} else {
		// Empty elements are always inline: <div></div>
		elem.InlineChildren = true
	}

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

	closeLine := p.current.Line
	if !p.expect(TokenRAngle) {
		return elem
	}

	// Check for trailing comment on same line as closing >
	elem.TrailingComments = p.getTrailingCommentOnLine(closeLine)

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
			Name:          name,
			Value:         &BoolLit{Value: true, Position: pos},
			Position:      pos,
			ValuePosition: pos,
		}
	}

	p.advance() // consume =

	// Record the position of the value (after the '=')
	valuePos := p.position()

	// Parse value
	var value Node
	switch p.current.Type {
	case TokenString:
		value = &StringLit{Value: p.current.Literal, Position: p.position()}
		// For string literals, the value position should be inside the quotes
		// The current position is at the opening quote, so add 1 for the quote
		valuePos = p.position()
		valuePos.Column++ // Move past the opening quote
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
		Name:          name,
		Value:         value,
		Position:      pos,
		ValuePosition: valuePos,
	}
}

// parseChildren parses children inside an element until the closing tag.
// Returns the child nodes and any orphan comments not attached to a node.
func (p *Parser) parseChildren(parentTag string) ([]Node, []*CommentGroup) {
	var children []Node
	var orphanComments []*CommentGroup

	for {
		// Count newlines to detect blank lines between siblings
		nlCount := 0
		for p.current.Type == TokenNewline {
			nlCount++
			p.advance()
		}
		hadBlankLine := nlCount >= 2

		// Check for closing tag — consume any pending comments as orphans first
		if p.current.Type == TokenLAngleSlash || p.current.Type == TokenEOF {
			p.collectPendingComments()
			remaining := p.consumePendingComments()
			if len(remaining) > 0 {
				orphanComments = append(orphanComments, groupComments(remaining)...)
			}
			break
		}

		// Collect any pending comments before parsing the next child
		leadingComments := p.getLeadingCommentGroup()

		// Parse child based on token type
		var child Node
		switch p.current.Type {
		case TokenLAngle:
			// Nested element
			// Check concrete type before assigning to interface to avoid typed nil
			if elem := p.parseElement(); elem != nil {
				child = elem
			}
		case TokenLBrace:
			// Go expression or children slot
			// Note: parseGoExprOrChildrenSlot returns Node interface, so no typed nil issue
			child = p.parseGoExprOrChildrenSlot()
		case TokenAtLet:
			if let := p.parseLet(); let != nil {
				child = let
			}
		case TokenAtFor:
			if f := p.parseFor(); f != nil {
				child = f
			}
		case TokenAtIf:
			if i := p.parseIf(); i != nil {
				child = i
			}
		case TokenAtCall:
			if call := p.parseComponentCall(); call != nil {
				child = call
			}
		case TokenAtExpr:
			if expr := p.parseComponentExpr(); expr != nil {
				child = expr
			}
		default:
			// Coalesce consecutive text tokens into a single TextContent.
			// In element content, we treat identifiers and various punctuation as text
			// until we hit a special delimiter ({, <, @, newline, EOF, or closing tag).
			if isTextToken(p.current.Type) {
				var text strings.Builder
				textPos := p.position()
				prevWasWord := false
				prevWasSpaceAfterPunct := false
				for isTextToken(p.current.Type) {
					currIsWord := isWordToken(p.current.Type)
					// Add space between:
					// - consecutive word tokens (e.g., "Hello World")
					// - punctuation that should have trailing space followed by word (e.g., ", q")
					if text.Len() > 0 && currIsWord && (prevWasWord || prevWasSpaceAfterPunct) {
						text.WriteByte(' ')
					}
					text.WriteString(p.current.Literal)
					prevWasWord = currIsWord
					// Comma should have a space after it before the next word
					prevWasSpaceAfterPunct = p.current.Type == TokenComma || p.current.Type == TokenColon || p.current.Type == TokenSemicolon
					p.advance()
				}
				child = &TextContent{
					Text:     text.String(),
					Position: textPos,
				}
			} else if p.current.Type != TokenEOF && p.current.Type != TokenLAngleSlash {
				// Skip unknown tokens in children context
				p.advance()
			} else {
				break
			}
		}

		// Attach leading comments to the child if one was parsed
		if child != nil {
			// Mark blank line before non-first children
			if hadBlankLine && len(children) > 0 {
				setBlankLineBefore(child, true)
			}
			p.attachLeadingComments(child, leadingComments)
			children = append(children, child)
		} else if leadingComments != nil {
			// No node was parsed, these comments are orphans
			orphanComments = append(orphanComments, leadingComments)
		}
	}

	return children, orphanComments
}

// setBlankLineBefore sets the BlankLineBefore field on a node.
func setBlankLineBefore(node Node, v bool) {
	switch n := node.(type) {
	case *Element:
		n.BlankLineBefore = v
	case *GoExpr:
		n.BlankLineBefore = v
	case *TextContent:
		n.BlankLineBefore = v
	case *LetBinding:
		n.BlankLineBefore = v
	case *ForLoop:
		n.BlankLineBefore = v
	case *IfStmt:
		n.BlankLineBefore = v
	case *ComponentCall:
		n.BlankLineBefore = v
	case *ChildrenSlot:
		n.BlankLineBefore = v
	}
}
