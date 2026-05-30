// Package highlight is a small, dependency-free syntax tokenizer for fenced code
// blocks. It does not import the tui package; the root package maps its tokens to
// colored spans. Output is "syntax coloring", not a full parser.
package highlight

import (
	"strings"
	"unicode"
)

// Kind is the small token taxonomy emitted by the lexers.
type Kind int

const (
	KindPlain    Kind = iota // whitespace, punctuation-free text, uncolored
	KindKeyword              // language keywords
	KindString               // string/char/template literals
	KindComment              // line and block comments
	KindNumber               // numeric literals
	KindLiteral              // true/false/null/nil, builtins, shell $VARs
	KindType                 // type and function/method names (heuristic)
	KindOperator             // operators and punctuation
	KindKey                  // JSON object keys
)

// Token is a contiguous run of source text classified as one Kind.
type Token struct {
	Kind Kind
	Text string
}

// Tokenize returns one []Token per line of code. Newlines are line separators
// and never appear in token Text. Lexer state carries across lines, so multi-line
// strings and comments tokenize correctly. Unknown or empty lang yields each line
// as a single KindPlain token. The concatenation of a line's token Text always
// equals the original line.
func Tokenize(lang, code string) [][]Token {
	lex := lexerFor(lang)
	if lex == nil {
		return plainLines(code)
	}
	return splitLines(lex(code))
}

func lexerFor(lang string) func(string) []Token {
	switch normalizeLang(lang) {
	case "go":
		return lexGo
	case "json":
		return lexJSON
	case "bash":
		return lexBash
	case "js":
		return lexJS
	}
	return nil
}

func normalizeLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "go", "golang":
		return "go"
	case "json":
		return "json"
	case "bash", "sh", "shell", "zsh":
		return "bash"
	case "js", "javascript", "ts", "typescript", "jsx", "tsx":
		return "js"
	}
	return ""
}

// plainLines wraps each line of code as a single plain token (fallback path).
func plainLines(code string) [][]Token {
	lines := strings.Split(code, "\n")
	out := make([][]Token, len(lines))
	for i, ln := range lines {
		if ln == "" {
			out[i] = []Token{}
		} else {
			out[i] = []Token{{Kind: KindPlain, Text: ln}}
		}
	}
	return out
}

// splitLines turns a flat token stream (whose Text may contain newlines) into one
// []Token per line, splitting any newline-spanning token into same-kind pieces.
func splitLines(toks []Token) [][]Token {
	out := [][]Token{}
	cur := []Token{}
	for _, t := range toks {
		parts := strings.Split(t.Text, "\n")
		for k, p := range parts {
			if k > 0 {
				out = append(out, cur)
				cur = []Token{}
			}
			if p != "" {
				cur = append(cur, Token{Kind: t.Kind, Text: p})
			}
		}
	}
	out = append(out, cur)
	return out
}

// --- shared scanner helpers (used by the per-language lexers) ---

func wordsSet(s string) map[string]bool {
	m := make(map[string]bool)
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

func isDigit(r rune) bool      { return r >= '0' && r <= '9' }
func isIdentStart(r rune) bool { return r == '_' || unicode.IsLetter(r) }
func isIdentPart(r rune) bool  { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }

// scanQuoted returns the index just past a single-line quoted run beginning at
// rs[start] (the opening quote). Honors backslash escapes; stops at the matching
// quote, a newline, or end of input.
func scanQuoted(rs []rune, start int) int {
	q := rs[start]
	j := start + 1
	for j < len(rs) {
		if rs[j] == '\\' && j+1 < len(rs) && rs[j+1] != '\n' {
			j += 2
			continue
		}
		if rs[j] == q || rs[j] == '\n' {
			break
		}
		j++
	}
	if j < len(rs) && rs[j] == q {
		j++
	}
	return j
}

// scanNumber returns the index just past a numeric literal beginning at rs[start]
// (a digit). Accepts hex/octal/binary prefixes, digit separators, and a dot.
func scanNumber(rs []rune, start int) int {
	j := start
	for j < len(rs) {
		r := rs[j]
		if isDigit(r) || r == '.' || r == '_' ||
			(r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') ||
			r == 'x' || r == 'X' || r == 'o' || r == 'O' || r == 'b' || r == 'B' {
			j++
			continue
		}
		break
	}
	return j
}

// nextNonSpaceIs reports whether the next non-space, non-tab rune at or after i
// equals c. Used for the "identifier followed by (" function heuristic.
func nextNonSpaceIs(rs []rune, i int, c rune) bool {
	for i < len(rs) {
		if rs[i] == ' ' || rs[i] == '\t' {
			i++
			continue
		}
		return rs[i] == c
	}
	return false
}

// nextNonSpaceIsIdent reports whether the next non-space, non-tab rune at or
// after i begins an identifier. Used to arm the "name after func/type" type
// heuristic only when a name actually follows (not '(' or '{').
func nextNonSpaceIsIdent(rs []rune, i int) bool {
	for i < len(rs) {
		if rs[i] == ' ' || rs[i] == '\t' {
			i++
			continue
		}
		return isIdentStart(rs[i])
	}
	return false
}

const operatorRunes = "{}()[]<>+-*/%=&|^~!?:;,."

func isOperator(r rune) bool { return strings.ContainsRune(operatorRunes, r) }
