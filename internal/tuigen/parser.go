package tuigen

// Parser parses .gsx source files into an AST.
type Parser struct {
	lexer           *Lexer
	current         Token
	peek            Token
	errors          *ErrorList
	pendingComments []*Comment // Comments collected since last attachment
	inMethodTempl   bool       // true when parsing inside a method templ (has receiver)
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

// clearPendingComments discards all pending comments from both the lexer and parser.
// This is used after capturing raw Go code (functions, declarations) where
// comments inside the code are already part of the captured source string.
func (p *Parser) clearPendingComments() {
	p.lexer.ConsumeComments()
	p.pendingComments = nil
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
