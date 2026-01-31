package tuigen

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

// readRune reads a single-quoted rune literal with escape sequences.
func (l *Lexer) readRune() Token {
	l.readChar() // consume opening '

	var r rune
	if l.ch == '\\' {
		l.readChar() // consume backslash
		switch l.ch {
		case 'n':
			r = '\n'
		case 't':
			r = '\t'
		case 'r':
			r = '\r'
		case '\\':
			r = '\\'
		case '\'':
			r = '\''
		case '0':
			r = '\000'
		default:
			// For other escapes, keep the character as-is
			r = l.ch
		}
		l.readChar()
	} else if l.ch == '\'' || l.ch == 0 {
		l.errors.AddError(l.position(), "empty rune literal")
		return l.makeToken(TokenError, "")
	} else {
		r = l.ch
		l.readChar()
	}

	if l.ch != '\'' {
		l.errors.AddError(l.position(), "unterminated rune literal")
		return l.makeToken(TokenError, string(r))
	}

	l.readChar() // consume closing '
	return l.makeToken(TokenRune, string(r))
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
