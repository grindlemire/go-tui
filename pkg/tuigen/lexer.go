package tuigen

import (
	"unicode"
	"unicode/utf8"
)

// Lexer tokenizes .tui source files.
type Lexer struct {
	filename string
	source   string
	pos      int  // current position in source
	readPos  int  // next position to read
	ch       rune // current character
	line     int  // current line (1-based)
	column   int  // current column (1-based)

	// Track the start position of current token
	tokenLine   int
	tokenColumn int

	errors *ErrorList
}

// NewLexer creates a new Lexer for the given source.
func NewLexer(filename, source string) *Lexer {
	l := &Lexer{
		filename: filename,
		source:   source,
		line:     1,
		column:   0,
		errors:   NewErrorList(),
	}
	l.readChar()
	return l
}

// Errors returns any errors encountered during lexing.
func (l *Lexer) Errors() *ErrorList {
	return l.errors
}

// readChar advances to the next character in the source.
func (l *Lexer) readChar() {
	// Track if previous char was a newline for line counting
	prevWasNewline := l.ch == '\n'

	if l.readPos >= len(l.source) {
		l.ch = 0 // EOF
		l.pos = l.readPos
		if prevWasNewline {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		return
	}

	r, size := utf8.DecodeRuneInString(l.source[l.readPos:])
	l.ch = r
	l.pos = l.readPos
	l.readPos += size

	if prevWasNewline {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

// peekChar returns the next character without advancing.
func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.source) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.readPos:])
	return r
}

// startToken marks the beginning of a new token.
func (l *Lexer) startToken() {
	l.tokenLine = l.line
	l.tokenColumn = l.column
}

// makeToken creates a token with the current start position.
func (l *Lexer) makeToken(typ TokenType, literal string) Token {
	return Token{
		Type:    typ,
		Literal: literal,
		Line:    l.tokenLine,
		Column:  l.tokenColumn,
	}
}

// position returns the current Position for error reporting.
func (l *Lexer) position() Position {
	return Position{
		File:   l.filename,
		Line:   l.tokenLine,
		Column: l.tokenColumn,
	}
}

// Next returns the next token from the source.
func (l *Lexer) Next() Token {
	l.skipWhitespaceAndComments()

	l.startToken()

	switch l.ch {
	case 0:
		return l.makeToken(TokenEOF, "")

	case '\n':
		l.readChar()
		return l.makeToken(TokenNewline, "\n")

	case '(':
		l.readChar()
		return l.makeToken(TokenLParen, "(")

	case ')':
		l.readChar()
		return l.makeToken(TokenRParen, ")")

	case '{':
		l.readChar()
		return l.makeToken(TokenLBrace, "{")

	case '}':
		l.readChar()
		return l.makeToken(TokenRBrace, "}")

	case '[':
		l.readChar()
		return l.makeToken(TokenLBracket, "[")

	case ']':
		l.readChar()
		return l.makeToken(TokenRBracket, "]")

	case '<':
		if l.peekChar() == '/' {
			l.readChar() // consume <
			l.readChar() // consume /
			return l.makeToken(TokenLAngleSlash, "</")
		}
		l.readChar()
		return l.makeToken(TokenLAngle, "<")

	case '>':
		l.readChar()
		return l.makeToken(TokenRAngle, ">")

	case '/':
		if l.peekChar() == '>' {
			l.readChar() // consume /
			l.readChar() // consume >
			return l.makeToken(TokenSlashAngle, "/>")
		}
		l.readChar()
		return l.makeToken(TokenSlash, "/")

	case '=':
		l.readChar()
		return l.makeToken(TokenEquals, "=")

	case ',':
		l.readChar()
		return l.makeToken(TokenComma, ",")

	case '.':
		// Could be . or ... or a number like .5
		if isDigit(l.peekChar()) {
			return l.readNumber()
		}
		l.readChar()
		return l.makeToken(TokenDot, ".")

	case ':':
		if l.peekChar() == '=' {
			l.readChar() // consume :
			l.readChar() // consume =
			return l.makeToken(TokenColonEquals, ":=")
		}
		l.readChar()
		return l.makeToken(TokenColon, ":")

	case ';':
		l.readChar()
		return l.makeToken(TokenSemicolon, ";")

	case '&':
		l.readChar()
		return l.makeToken(TokenAmpersand, "&")

	case '|':
		l.readChar()
		return l.makeToken(TokenPipe, "|")

	case '*':
		l.readChar()
		return l.makeToken(TokenStar, "*")

	case '+':
		l.readChar()
		return l.makeToken(TokenPlus, "+")

	case '-':
		l.readChar()
		return l.makeToken(TokenMinus, "-")

	case '!':
		l.readChar()
		return l.makeToken(TokenBang, "!")

	case '_':
		// Could be _ or an identifier starting with _
		if isLetter(l.peekChar()) || isDigit(l.peekChar()) {
			return l.readIdentifier()
		}
		l.readChar()
		return l.makeToken(TokenUnderscore, "_")

	case '@':
		return l.readAtKeyword()

	case '"':
		return l.readString()

	case '`':
		return l.readRawString()

	default:
		if isLetter(l.ch) {
			return l.readIdentifier()
		}
		if isDigit(l.ch) {
			return l.readNumber()
		}

		// Unknown character
		ch := l.ch
		l.readChar()
		l.errors.AddErrorf(l.position(), "unexpected character %q", ch)
		return l.makeToken(TokenError, string(ch))
	}
}

// skipWhitespaceAndComments skips spaces, tabs, and comments (but not newlines).
func (l *Lexer) skipWhitespaceAndComments() {
	for {
		switch l.ch {
		case ' ', '\t', '\r':
			l.readChar()
		case '/':
			if l.peekChar() == '/' {
				// Line comment: skip until newline
				l.skipLineComment()
			} else if l.peekChar() == '*' {
				// Block comment: skip until */
				l.skipBlockComment()
			} else {
				return
			}
		default:
			return
		}
	}
}

// skipLineComment skips a // comment until end of line.
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipBlockComment skips a /* */ comment.
func (l *Lexer) skipBlockComment() {
	l.readChar() // skip /
	l.readChar() // skip *

	for {
		if l.ch == 0 {
			l.errors.AddError(l.position(), "unterminated block comment")
			return
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip *
			l.readChar() // skip /
			return
		}
		l.readChar()
	}
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
	case "component":
		return l.makeToken(TokenAtComponent, "@component")
	case "let":
		return l.makeToken(TokenAtLet, "@let")
	case "for":
		return l.makeToken(TokenAtFor, "@for")
	case "if":
		return l.makeToken(TokenAtIf, "@if")
	case "else":
		return l.makeToken(TokenAtElse, "@else")
	default:
		l.errors.AddErrorf(l.position(), "unknown @ keyword: @%s", keyword)
		return l.makeToken(TokenError, "@"+keyword)
	}
}

// readString reads a double-quoted string with escape sequences.
func (l *Lexer) readString() Token {
	l.readChar() // consume opening "

	var result []rune
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\n' {
			l.errors.AddError(l.position(), "unterminated string literal")
			return l.makeToken(TokenError, string(result))
		}
		if l.ch == '\\' {
			l.readChar() // consume backslash
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			case '0':
				result = append(result, '\000')
			default:
				// Keep the backslash and character as-is
				result = append(result, '\\', l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}

	if l.ch == 0 {
		l.errors.AddError(l.position(), "unterminated string literal")
		return l.makeToken(TokenError, string(result))
	}

	l.readChar() // consume closing "
	return l.makeToken(TokenString, string(result))
}

// readRawString reads a backtick-quoted raw string.
func (l *Lexer) readRawString() Token {
	l.readChar() // consume opening `

	startPos := l.pos
	for l.ch != '`' && l.ch != 0 {
		l.readChar()
	}

	if l.ch == 0 {
		l.errors.AddError(l.position(), "unterminated raw string literal")
		return l.makeToken(TokenError, l.source[startPos:l.pos])
	}

	literal := l.source[startPos:l.pos]
	l.readChar() // consume closing `
	return l.makeToken(TokenRawString, literal)
}

// readNumber reads an integer or float literal.
func (l *Lexer) readNumber() Token {
	startPos := l.pos
	isFloat := false

	// Handle leading dot for floats like .5
	if l.ch == '.' {
		isFloat = true
		l.readChar()
	}

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && !isFloat {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Check for exponent
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.source[startPos:l.pos]
	if isFloat {
		return l.makeToken(TokenFloat, literal)
	}
	return l.makeToken(TokenInt, literal)
}

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

// isLetter returns true if the rune is a letter or underscore.
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
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

	return tok
}

// CurrentChar returns the current character being examined.
// Used by the parser to check if we're at a { for Go expressions.
func (l *Lexer) CurrentChar() rune {
	return l.ch
}

// SkipWhitespace is a public method for the parser to skip whitespace.
func (l *Lexer) SkipWhitespace() {
	l.skipWhitespaceAndComments()
}
