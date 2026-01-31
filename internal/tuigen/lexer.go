package tuigen

import (
	"unicode/utf8"
)

// Lexer tokenizes .gsx source files.
type Lexer struct {
	filename string
	source   string
	pos      int  // current position in source
	readPos  int  // next position to read
	ch       rune // current character
	line     int  // current line (1-based)
	column   int  // current column (1-based)

	// Track the start position of current token
	tokenLine     int
	tokenColumn   int
	tokenStartPos int // byte offset where current token starts

	// Comments collected since last ConsumeComments() call
	pendingComments []*Comment

	// Track the end line of the last comment that was collected
	// (persists across ConsumeComments calls to detect blank lines between comment batches)
	lastCommentEndLine int

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
	l.tokenStartPos = l.pos
}

// makeToken creates a token with the current start position.
func (l *Lexer) makeToken(typ TokenType, literal string) Token {
	// Reset lastCommentEndLine for non-newline tokens
	// This ensures we don't detect false "blank lines" between comment groups
	// that are actually separated by code
	if typ != TokenNewline && typ != TokenEOF {
		l.lastCommentEndLine = 0
	}
	return Token{
		Type:     typ,
		Literal:  literal,
		Line:     l.tokenLine,
		Column:   l.tokenColumn,
		StartPos: l.tokenStartPos,
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
	// Skip whitespace and collect any comments
	l.skipWhitespaceAndCollectComments()

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

	case '\'':
		return l.readRune()

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
