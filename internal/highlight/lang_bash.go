package highlight

var bashKeywords = wordsSet(`if then elif else fi for while until do done case esac in
	function select time return break continue local export readonly declare unset source`)

func lexBash(code string) []Token {
	rs := []rune(code)
	n := len(rs)
	var toks []Token
	var plain []rune
	flush := func() {
		if len(plain) > 0 {
			toks = append(toks, Token{KindPlain, string(plain)})
			plain = plain[:0]
		}
	}
	for i := 0; i < n; {
		r := rs[i]
		switch {
		case r == '#':
			flush()
			j := i
			for j < n && rs[j] != '\n' {
				j++
			}
			toks = append(toks, Token{KindComment, string(rs[i:j])})
			i = j
		case r == '"' || r == '\'':
			flush()
			j := scanQuoted(rs, i)
			toks = append(toks, Token{KindString, string(rs[i:j])})
			i = j
		case r == '$':
			flush()
			j := i + 1
			if j < n && rs[j] == '{' {
				for j < n && rs[j] != '}' {
					j++
				}
				if j < n {
					j++
				}
			} else {
				for j < n && isIdentPart(rs[j]) {
					j++
				}
			}
			toks = append(toks, Token{KindLiteral, string(rs[i:j])})
			i = j
		case isDigit(r):
			flush()
			j := scanNumber(rs, i)
			toks = append(toks, Token{KindNumber, string(rs[i:j])})
			i = j
		case isIdentStart(r):
			flush()
			j := i + 1
			for j < n && isIdentPart(rs[j]) {
				j++
			}
			word := string(rs[i:j])
			kind := KindPlain
			switch {
			case bashKeywords[word]:
				kind = KindKeyword
			case nextNonSpaceIs(rs, j, '('):
				kind = KindType // function definition or call: name()
			}
			toks = append(toks, Token{kind, word})
			i = j
		case isBashOperator(r):
			flush()
			toks = append(toks, Token{KindOperator, string(r)})
			i++
		default:
			plain = append(plain, r)
			i++
		}
	}
	flush()
	return toks
}

func isBashOperator(r rune) bool {
	switch r {
	case '|', '&', '>', '<', ';', '(', ')', '{', '}', '=':
		return true
	}
	return false
}
