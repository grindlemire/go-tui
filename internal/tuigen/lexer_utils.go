package tuigen

import (
	"unicode"
	"unicode/utf8"
)

// skipWhitespaceAndCollectComments skips spaces, tabs, and collects comments (but not newlines).
func (l *Lexer) skipWhitespaceAndCollectComments() {
	for {
		switch l.ch {
		case ' ', '\t', '\r':
			l.readChar()
		case '/':
			if l.peekChar() == '/' {
				// Line comment: collect it
				l.collectLineComment()
			} else if l.peekChar() == '*' {
				// Block comment: collect it
				l.collectBlockComment()
			} else {
				return
			}
		default:
			return
		}
	}
}

// skipWhitespaceOnly skips spaces and tabs only, no comments.
func (l *Lexer) skipWhitespaceOnly() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// collectLineComment reads a // comment and adds it to pendingComments.
func (l *Lexer) collectLineComment() {
	startPos := l.pos
	startLine := l.line
	startCol := l.column

	// Check if there was a blank line before this comment
	blankLineBefore := l.hadBlankLineBefore(startLine)

	// Read until end of line or EOF
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	comment := &Comment{
		Text:            l.source[startPos:l.pos],
		Position:        Position{File: l.filename, Line: startLine, Column: startCol},
		EndLine:         l.line,
		EndCol:          l.column,
		IsBlock:         false,
		BlankLineBefore: blankLineBefore,
	}
	l.pendingComments = append(l.pendingComments, comment)
	l.lastCommentEndLine = l.line
}

// collectBlockComment reads a /* */ comment and adds it to pendingComments.
func (l *Lexer) collectBlockComment() {
	startPos := l.pos
	startLine := l.line
	startCol := l.column

	// Check if there was a blank line before this comment
	blankLineBefore := l.hadBlankLineBefore(startLine)

	l.readChar() // skip /
	l.readChar() // skip *

	for {
		if l.ch == 0 {
			l.errors.AddError(Position{File: l.filename, Line: startLine, Column: startCol}, "unterminated block comment")
			return
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip *
			l.readChar() // skip /
			break
		}
		l.readChar()
	}

	comment := &Comment{
		Text:            l.source[startPos:l.pos],
		Position:        Position{File: l.filename, Line: startLine, Column: startCol},
		EndLine:         l.line,
		EndCol:          l.column,
		IsBlock:         true,
		BlankLineBefore: blankLineBefore,
	}
	l.pendingComments = append(l.pendingComments, comment)
	l.lastCommentEndLine = l.line
}

// hadBlankLineBefore checks if there was a blank line before the given line.
// This only returns true if there's a previous comment (in pending list or recently consumed)
// and there's a blank line between that comment and the current line.
func (l *Lexer) hadBlankLineBefore(currentLine int) bool {
	// Check against the last pending comment first
	if len(l.pendingComments) > 0 {
		lastComment := l.pendingComments[len(l.pendingComments)-1]
		return currentLine > lastComment.EndLine+1
	}

	// Check against the last consumed comment (if any)
	if l.lastCommentEndLine > 0 {
		return currentLine > l.lastCommentEndLine+1
	}

	return false
}

// ConsumeComments returns and clears pending comments.
// Called by parser after each node is parsed.
func (l *Lexer) ConsumeComments() []*Comment {
	comments := l.pendingComments
	l.pendingComments = nil
	return comments
}

// readIdentifier reads an identifier or keyword.
func (l *Lexer) readIdentifier() Token {
	startPos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	literal := l.source[startPos:l.pos]
	tokenType := LookupIdent(literal)
	return l.makeToken(tokenType, literal)
}

// readAtKeyword reads a @-prefixed DSL keyword.
func (l *Lexer) readAtKeyword() Token {
	l.readChar() // consume @

	startPos := l.pos
	for isLetter(l.ch) {
		l.readChar()
	}
	keyword := l.source[startPos:l.pos]

	switch keyword {
	case "let":
		return l.makeToken(TokenAtLet, "@let")
	case "for":
		return l.makeToken(TokenAtFor, "@for")
	case "if":
		return l.makeToken(TokenAtIf, "@if")
	case "else":
		return l.makeToken(TokenAtElse, "@else")
	default:
		// Check if this is a component call (uppercase first letter)
		if len(keyword) > 0 {
			firstRune, _ := utf8.DecodeRuneInString(keyword)
			if unicode.IsUpper(firstRune) {
				return l.makeToken(TokenAtCall, keyword)
			}
		}
		l.errors.AddErrorf(l.position(), "unknown @ keyword: @%s", keyword)
		return l.makeToken(TokenError, "@"+keyword)
	}
}

// isLetter returns true if the rune is a letter or underscore.
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

// CurrentChar returns the current character being examined.
// Used by the parser to check if we're at a { for Go expressions.
func (l *Lexer) CurrentChar() rune {
	return l.ch
}

// SkipWhitespace is a public method for the parser to skip whitespace and collect comments.
func (l *Lexer) SkipWhitespace() {
	l.skipWhitespaceAndCollectComments()
}

// SourcePos returns the current position in the source string.
// Used by the parser to mark start positions for raw source capture.
func (l *Lexer) SourcePos() int {
	return l.pos
}

// SourceRange extracts a substring of the original source from start to end positions.
// Used by the parser to capture raw Go code without tokenization.
func (l *Lexer) SourceRange(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(l.source) {
		end = len(l.source)
	}
	if start >= end {
		return ""
	}
	return l.source[start:end]
}
