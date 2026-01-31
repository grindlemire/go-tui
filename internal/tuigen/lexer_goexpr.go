package tuigen

import (
	"unicode/utf8"
)

// ReadGoExpr reads a Go expression inside {}, handling nested braces.
// This is called by the parser when it encounters { in an attribute or content context.
// Note: The opening '{' has already been consumed by the lexer when producing TokenLBrace,
// so we start reading from the current position (after the '{').
func (l *Lexer) ReadGoExpr() Token {
	l.startToken()

	// The '{' was already consumed when TokenLBrace was created,
	// so we start with braceDepth=1 and read from current position.
	startPos := l.pos
	braceDepth := 1
	parenDepth := 0
	bracketDepth := 0

	for braceDepth > 0 && l.ch != 0 {
		switch l.ch {
		case '{':
			braceDepth++
		case '}':
			braceDepth--
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		case '"':
			l.skipStringInExpr()
			continue
		case '`':
			l.skipRawStringInExpr()
			continue
		case '\'':
			l.skipCharInExpr()
			continue
		}

		if braceDepth > 0 {
			l.readChar()
		}
	}

	if braceDepth != 0 {
		l.errors.AddError(l.position(), "unterminated Go expression: unmatched '{'")
		return l.makeToken(TokenError, l.source[startPos:l.pos])
	}

	if parenDepth != 0 {
		l.errors.AddErrorf(l.position(), "unterminated Go expression: unmatched parentheses")
	}
	if bracketDepth != 0 {
		l.errors.AddErrorf(l.position(), "unterminated Go expression: unmatched brackets")
	}

	expr := l.source[startPos:l.pos]
	l.readChar() // consume closing }

	return l.makeToken(TokenGoExpr, expr)
}

// skipStringInExpr skips a string literal inside a Go expression.
func (l *Lexer) skipStringInExpr() {
	l.readChar() // consume opening "
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // skip escape
		}
		l.readChar()
	}
	if l.ch == '"' {
		l.readChar() // consume closing "
	}
}

// skipRawStringInExpr skips a raw string literal inside a Go expression.
func (l *Lexer) skipRawStringInExpr() {
	l.readChar() // consume opening `
	for l.ch != '`' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '`' {
		l.readChar() // consume closing `
	}
}

// skipCharInExpr skips a character literal inside a Go expression.
func (l *Lexer) skipCharInExpr() {
	l.readChar() // consume opening '
	if l.ch == '\\' {
		l.readChar() // skip escape
	}
	l.readChar() // skip character
	if l.ch == '\'' {
		l.readChar() // consume closing '
	}
}

// PeekToken returns the next token without consuming it.
// This is used by the parser to look ahead.
func (l *Lexer) PeekToken() Token {
	// Save current state
	pos := l.pos
	readPos := l.readPos
	ch := l.ch
	line := l.line
	column := l.column
	tokenLine := l.tokenLine
	tokenColumn := l.tokenColumn
	pendingComments := l.pendingComments

	// Get next token
	tok := l.Next()

	// Restore state
	l.pos = pos
	l.readPos = readPos
	l.ch = ch
	l.line = line
	l.column = column
	l.tokenLine = tokenLine
	l.tokenColumn = tokenColumn
	l.pendingComments = pendingComments

	return tok
}

// ReadBalancedBraces reads content inside balanced braces, handling nested braces,
// strings, raw strings, and character literals. The opening '{' should NOT have been
// consumed yet - this method reads from the current position expecting to see '{'.
// Returns the content between braces (excluding the braces themselves).
func (l *Lexer) ReadBalancedBraces() (string, error) {
	// Expect opening brace
	if l.ch != '{' {
		return "", NewError(l.position(), "expected '{' at start of balanced braces")
	}
	l.readChar() // consume opening {

	startPos := l.pos
	braceDepth := 1

	for braceDepth > 0 && l.ch != 0 {
		switch l.ch {
		case '{':
			braceDepth++
			l.readChar()
		case '}':
			braceDepth--
			if braceDepth > 0 {
				l.readChar()
			}
		case '"':
			l.skipStringInExpr()
		case '`':
			l.skipRawStringInExpr()
		case '\'':
			l.skipCharInExpr()
		default:
			l.readChar()
		}
	}

	if braceDepth != 0 {
		return "", NewError(l.position(), "unterminated braces: unmatched '{'")
	}

	content := l.source[startPos:l.pos]
	l.readChar() // consume closing }

	return content, nil
}

// ReadUntilBrace reads raw source from the current position until a '{' is encountered.
// Used for capturing @if conditions and @for iterables as raw Go code.
// Does not consume the '{'.
func (l *Lexer) ReadUntilBrace() string {
	l.skipWhitespaceAndCollectComments()
	startPos := l.pos

	for l.ch != '{' && l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}

	return l.source[startPos:l.pos]
}

// ReadBalancedBracesFrom reads balanced brace content starting from a given source position.
// The startPos should point to the opening '{'. Returns the content between braces
// (excluding the braces themselves) and updates the lexer position to after the closing '}'.
// This is used by the parser when it has a TokenLBrace and needs to capture raw Go code.
func (l *Lexer) ReadBalancedBracesFrom(startPos int) (string, error) {
	// Validate startPos points to '{'
	if startPos < 0 || startPos >= len(l.source) || l.source[startPos] != '{' {
		return "", NewError(l.position(), "invalid start position for balanced braces")
	}

	// Scan forward from startPos+1 to find matching '}'
	contentStart := startPos + 1
	pos := contentStart
	braceDepth := 1

	for pos < len(l.source) && braceDepth > 0 {
		ch := l.source[pos]

		switch ch {
		case '{':
			braceDepth++
			pos++
		case '}':
			braceDepth--
			if braceDepth > 0 {
				pos++
			}
		case '"':
			// Skip string literal
			pos++
			for pos < len(l.source) && l.source[pos] != '"' {
				if l.source[pos] == '\\' && pos+1 < len(l.source) {
					pos += 2 // skip escape sequence
				} else {
					pos++
				}
			}
			if pos < len(l.source) {
				pos++ // skip closing "
			}
		case '`':
			// Skip raw string literal
			pos++
			for pos < len(l.source) && l.source[pos] != '`' {
				pos++
			}
			if pos < len(l.source) {
				pos++ // skip closing `
			}
		case '\'':
			// Skip character literal
			pos++
			if pos < len(l.source) && l.source[pos] == '\\' {
				pos += 2 // skip escape
			} else if pos < len(l.source) {
				pos++ // skip char
			}
			if pos < len(l.source) && l.source[pos] == '\'' {
				pos++ // skip closing '
			}
		default:
			pos++
		}
	}

	if braceDepth != 0 {
		return "", NewError(l.position(), "unterminated braces: unmatched '{'")
	}

	content := l.source[contentStart:pos]

	// Update lexer position to after closing '}'
	l.pos = pos
	l.readPos = pos + 1
	if l.readPos <= len(l.source) {
		r, _ := utf8.DecodeRuneInString(l.source[pos:])
		l.ch = r
	} else {
		l.ch = 0
	}

	// Calculate correct line/column from the start of the source.
	// We need to recalculate from scratch because the lexer may have peeked ahead
	// and l.column could be in an inconsistent state.
	lineStart := 0
	lineNum := 1
	for i := 0; i < startPos; i++ {
		if l.source[i] == '\n' {
			lineStart = i + 1
			lineNum++
		}
	}

	// Scan from lineStart to pos to get correct line and column
	l.line = lineNum
	l.column = 1
	for i := lineStart; i < pos && i < len(l.source); i++ {
		if l.source[i] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
	}

	l.readChar() // advance past '}' and update column

	return content, nil
}
