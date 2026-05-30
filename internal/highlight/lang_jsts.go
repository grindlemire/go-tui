package highlight

import "unicode"

var jsKeywords = wordsSet(`await break case catch class const continue debugger default delete
	do else export extends finally for function if import in instanceof let new return super switch
	this throw try typeof var void while with yield async of as from get set static interface type
	enum implements public private protected readonly`)

var jsLiterals = wordsSet(`true false null undefined NaN Infinity`)

func lexJS(code string) []Token {
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
	expectName := false // previous keyword was function/class/new
	for i := 0; i < n; {
		r := rs[i]
		switch {
		case r == '/' && i+1 < n && rs[i+1] == '/':
			flush()
			expectName = false
			j := i
			for j < n && rs[j] != '\n' {
				j++
			}
			toks = append(toks, Token{KindComment, string(rs[i:j])})
			i = j
		case r == '/' && i+1 < n && rs[i+1] == '*':
			flush()
			expectName = false
			j := i + 2
			for j < n && !(rs[j] == '*' && j+1 < n && rs[j+1] == '/') {
				j++
			}
			if j < n {
				j += 2
			} else {
				j = n
			}
			toks = append(toks, Token{KindComment, string(rs[i:j])})
			i = j
		case r == '`':
			flush()
			expectName = false
			j := i + 1
			for j < n && rs[j] != '`' {
				if rs[j] == '\\' && j+1 < n {
					j += 2
				} else {
					j++
				}
			}
			if j < n {
				j++
			}
			toks = append(toks, Token{KindString, string(rs[i:j])})
			i = j
		case r == '"' || r == '\'':
			flush()
			expectName = false
			j := scanQuoted(rs, i)
			toks = append(toks, Token{KindString, string(rs[i:j])})
			i = j
		case isDigit(r) || (r == '.' && i+1 < n && isDigit(rs[i+1])):
			flush()
			expectName = false
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
			case jsKeywords[word]:
				kind = KindKeyword
			case jsLiterals[word]:
				kind = KindLiteral
			case expectName:
				kind = KindType
			case nextNonSpaceIs(rs, j, '('):
				kind = KindType
			case unicode.IsUpper(rs[i]):
				kind = KindType
			}
			expectName = (word == "function" || word == "class" || word == "new") && nextNonSpaceIsIdent(rs, j)
			toks = append(toks, Token{kind, word})
			i = j
		case isOperator(r):
			flush()
			expectName = false
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
