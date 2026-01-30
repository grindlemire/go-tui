package tuigen

import (
	"strconv"
	"strings"
)

// Parser parses .gsx source files into an AST.
type Parser struct {
	lexer           *Lexer
	current         Token
	peek            Token
	errors          *ErrorList
	pendingComments []*Comment // Comments collected since last attachment
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

// synchronize skips tokens until a synchronization point is found.
// Synchronization points are top-level declarations: func, templ, type, const, var.
// This allows the parser to recover from errors and continue parsing.
func (p *Parser) synchronize() {
	for p.current.Type != TokenEOF {
		// Sync points: top-level declarations
		switch p.current.Type {
		case TokenFunc, TokenTempl, TokenTypeKw, TokenConst, TokenVar:
			return
		}
		p.advance()
	}
}

// collectPendingComments collects comments from the lexer and adds them to
// the parser's pending comments buffer.
func (p *Parser) collectPendingComments() {
	comments := p.lexer.ConsumeComments()
	p.pendingComments = append(p.pendingComments, comments...)
}

// consumePendingComments returns all pending comments and clears the buffer.
func (p *Parser) consumePendingComments() []*Comment {
	comments := p.pendingComments
	p.pendingComments = nil
	return comments
}

// groupComments groups comments into CommentGroups based on blank lines.
// Adjacent comments (no blank line between) form a group.
// A blank line (2+ newlines) starts a new group.
func groupComments(comments []*Comment) []*CommentGroup {
	if len(comments) == 0 {
		return nil
	}

	var groups []*CommentGroup
	var current []*Comment

	for i, c := range comments {
		if i == 0 {
			current = append(current, c)
			continue
		}

		// Check if there's a blank line between this comment and the previous one
		prev := comments[i-1]
		// A blank line means the previous comment's EndLine is at least 2 less than
		// this comment's start line
		if c.Position.Line > prev.EndLine+1 {
			// Blank line - start a new group
			if len(current) > 0 {
				groups = append(groups, &CommentGroup{List: current})
			}
			current = []*Comment{c}
		} else {
			// Adjacent - add to current group
			current = append(current, c)
		}
	}

	// Don't forget the last group
	if len(current) > 0 {
		groups = append(groups, &CommentGroup{List: current})
	}

	return groups
}

// getLeadingCommentGroup returns a single CommentGroup containing all pending
// comments, or nil if no comments are pending.
func (p *Parser) getLeadingCommentGroup() *CommentGroup {
	p.collectPendingComments()
	comments := p.consumePendingComments()
	if len(comments) == 0 {
		return nil
	}
	return &CommentGroup{List: comments}
}

// getTrailingCommentOnLine checks if there's a comment on the same line as the given position.
// If so, returns it as a CommentGroup. Otherwise returns nil.
// The comment must start on the same line to be considered trailing.
func (p *Parser) getTrailingCommentOnLine(line int) *CommentGroup {
	p.collectPendingComments()
	if len(p.pendingComments) == 0 {
		return nil
	}

	// Check if first pending comment is on the same line
	first := p.pendingComments[0]
	if first.Position.Line == line {
		// This is a trailing comment
		trailing := p.pendingComments[0]
		p.pendingComments = p.pendingComments[1:]
		return &CommentGroup{List: []*Comment{trailing}}
	}

	return nil
}

// ParseFile parses a complete .gsx file into a File AST node.
func (p *Parser) ParseFile() (*File, error) {
	file := &File{
		Position: p.position(),
	}

	p.skipNewlines()

	// Collect any leading comments before package
	file.LeadingComments = p.getLeadingCommentGroup()

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
	// Components are defined with `templ`: templ Name(params) { ... }
	// Helper functions are all other func declarations
	for p.current.Type != TokenEOF {
		p.skipNewlines()
		if p.current.Type == TokenEOF {
			break
		}

		// Collect comments before the next declaration
		leadingComments := p.getLeadingCommentGroup()

		switch p.current.Type {
		case TokenTempl:
			// Parse templ - always a component, no return type
			comp := p.parseTempl()
			if comp != nil {
				comp.LeadingComments = leadingComments
				file.Components = append(file.Components, comp)
			} else {
				// Error recovery: skip to next top-level declaration
				p.synchronize()
			}
		case TokenFunc:
			// Parse func - could be either a component (returns Element) or helper function
			result := p.parseFuncOrComponent()
			if result != nil {
				switch r := result.(type) {
				case *Component:
					r.LeadingComments = leadingComments
					file.Components = append(file.Components, r)
				case *GoFunc:
					r.LeadingComments = leadingComments
					file.Funcs = append(file.Funcs, r)
				}
			} else {
				// Error recovery: skip to next top-level declaration
				p.synchronize()
			}
		case TokenTypeKw, TokenConst, TokenVar:
			// Parse type, const, or var declaration
			decl := p.parseGoDecl()
			if decl != nil {
				decl.LeadingComments = leadingComments
				file.Decls = append(file.Decls, decl)
			} else {
				// Error recovery: skip to next top-level declaration
				p.synchronize()
			}
		default:
			p.errors.AddErrorf(p.position(), "unexpected token %s, expected func, templ, type, const, or var", p.current.Type)
			// Error recovery: skip to next top-level declaration
			p.synchronize()
		}
	}

	// Collect any orphan comments at end of file
	p.collectPendingComments()
	orphanComments := p.consumePendingComments()
	if len(orphanComments) > 0 {
		file.OrphanComments = groupComments(orphanComments)
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

	// Check for #Name (named ref)
	if p.current.Type == TokenHash {
		p.advance() // consume #
		if p.current.Type != TokenIdent {
			p.errors.AddError(p.position(), "expected identifier after '#' for named ref")
			return nil
		}
		elem.NamedRef = p.current.Literal
		p.advance()
	}

	// Parse attributes
	elem.Attributes = p.parseAttributes()

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

	// Check for self-closing or opening tag
	if p.current.Type == TokenSlashAngle {
		// Self-closing: <tag />
		elem.SelfClose = true
		closeLine := p.current.Line
		p.advanceSkipNewlines()
		// Check for trailing comment on same line as />
		elem.TrailingComments = p.getTrailingCommentOnLine(closeLine)
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

	closeLine := p.current.Line
	if !p.expect(TokenRAngle) {
		return elem
	}

	// Check for trailing comment on same line as closing >
	elem.TrailingComments = p.getTrailingCommentOnLine(closeLine)

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
func (p *Parser) parseChildren(parentTag string) []Node {
	var children []Node

	for {
		// Skip whitespace-only newlines
		p.skipNewlines()

		// Check for closing tag
		if p.current.Type == TokenLAngleSlash || p.current.Type == TokenEOF {
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
			p.attachLeadingComments(child, leadingComments)
			children = append(children, child)
		}
		// Note: if child is nil and we had leading comments, they become orphans
		// but we don't track orphan comments in element children for simplicity
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
			// Note: else orphan comments are stored in the nested IfStmt or lost
			// This is a simplification - full support would require a separate field
			stmt.Else = p.parseComponentBody()

			if !p.expectSkipNewlines(TokenRBrace) {
				return stmt
			}
		}
	}

	return stmt
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
		Name:     name,
		Args:     args,
		Position: pos,
	}

	// Optional children block
	if p.current.Type == TokenLBrace {
		p.advance()
		p.skipNewlines()
		call.Children = p.parseComponentBody()
		if !p.expectSkipNewlines(TokenRBrace) {
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
	case TokenUnderscore, TokenHash:
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
