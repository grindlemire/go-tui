package highlight

func lexJSON(code string) []Token {
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
		case r == '"':
			flush()
			j := scanQuoted(rs, i)
			kind := KindString
			if nextNonSpaceIs(rs, j, ':') {
				kind = KindKey
			}
			toks = append(toks, Token{kind, string(rs[i:j])})
			i = j
		case isDigit(r) || (r == '-' && i+1 < n && isDigit(rs[i+1])):
			flush()
			j := scanNumber(rs, i+1) // skip a leading '-' or first digit; rs[i:j] keeps it
			if j <= i {
				j = i + 1
			}
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
			if word == "true" || word == "false" || word == "null" {
				kind = KindLiteral
			}
			toks = append(toks, Token{kind, word})
			i = j
		case r == '{' || r == '}' || r == '[' || r == ']' || r == ':' || r == ',':
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
