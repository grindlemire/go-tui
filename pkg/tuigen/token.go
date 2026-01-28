// Package tuigen provides a DSL compiler that transforms .tui files into Go source code.
// The DSL provides a templ-inspired syntax for building go-tui element trees.
package tuigen

import "fmt"

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Special tokens
	TokenEOF        TokenType = iota // end of file
	TokenError                       // lexer error
	TokenNewline                     // newline
	TokenWhitespace                  // spaces/tabs (usually skipped)

	// Keywords
	TokenPackage // package
	TokenImport  // import
	TokenFunc    // func
	TokenReturn  // return
	TokenIf      // if
	TokenElse    // else
	TokenFor     // for
	TokenRange   // range

	// DSL keywords (@ prefixed)
	TokenAtComponent // @component
	TokenAtLet       // @let
	TokenAtFor       // @for
	TokenAtIf        // @if
	TokenAtElse      // @else
	TokenAtCall      // @ComponentName (uppercase, component call)

	// Literals
	TokenIdent     // identifier
	TokenInt       // integer literal: 123
	TokenFloat     // float literal: 1.23
	TokenString    // string literal: "..."
	TokenRawString // raw string literal: `...`
	TokenRune      // rune literal: 'x'

	// Operators and Punctuation
	TokenLParen      // (
	TokenRParen      // )
	TokenLBrace      // {
	TokenRBrace      // }
	TokenLAngle      // <
	TokenRAngle      // >
	TokenLBracket    // [
	TokenRBracket    // ]
	TokenSlash       // /
	TokenEquals      // =
	TokenComma       // ,
	TokenDot         // .
	TokenColon       // :
	TokenSemicolon   // ;
	TokenColonEquals // :=
	TokenAmpersand   // &
	TokenPipe        // |
	TokenStar        // *
	TokenPlus        // +
	TokenMinus       // -
	TokenBang        // !
	TokenUnderscore  // _
	TokenHash        // #

	// Composite tokens
	TokenGoExpr      // Go expression inside {}
	TokenSlashAngle  // />  (self-closing tag end)
	TokenLAngleSlash // </ (closing tag start)

	// Comment tokens (collected but not emitted by lexer)
	TokenLineComment  // // comment
	TokenBlockComment // /* comment */
)

// tokenNames maps token types to their string names for debugging.
var tokenNames = map[TokenType]string{
	TokenEOF:         "EOF",
	TokenError:       "Error",
	TokenNewline:     "Newline",
	TokenWhitespace:  "Whitespace",
	TokenPackage:     "package",
	TokenImport:      "import",
	TokenFunc:        "func",
	TokenReturn:      "return",
	TokenIf:          "if",
	TokenElse:        "else",
	TokenFor:         "for",
	TokenRange:       "range",
	TokenAtComponent: "@component",
	TokenAtLet:       "@let",
	TokenAtFor:       "@for",
	TokenAtIf:        "@if",
	TokenAtElse:      "@else",
	TokenAtCall:      "@Call",
	TokenIdent:       "Ident",
	TokenInt:         "Int",
	TokenFloat:       "Float",
	TokenString:      "String",
	TokenRawString:   "RawString",
	TokenRune:        "Rune",
	TokenLParen:      "(",
	TokenRParen:      ")",
	TokenLBrace:      "{",
	TokenRBrace:      "}",
	TokenLAngle:      "<",
	TokenRAngle:      ">",
	TokenLBracket:    "[",
	TokenRBracket:    "]",
	TokenSlash:       "/",
	TokenEquals:      "=",
	TokenComma:       ",",
	TokenDot:         ".",
	TokenColon:       ":",
	TokenSemicolon:   ";",
	TokenColonEquals: ":=",
	TokenAmpersand:   "&",
	TokenPipe:        "|",
	TokenStar:        "*",
	TokenPlus:        "+",
	TokenMinus:       "-",
	TokenBang:        "!",
	TokenUnderscore:  "_",
	TokenHash:        "#",
	TokenGoExpr:       "GoExpr",
	TokenSlashAngle:   "/>",
	TokenLAngleSlash:  "</",
	TokenLineComment:  "LineComment",
	TokenBlockComment: "BlockComment",
}

// String returns a human-readable name for the token type.
func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", t)
}

// Token represents a lexical token with its type, literal value, and source position.
type Token struct {
	Type     TokenType
	Literal  string
	Line     int
	Column   int
	StartPos int // byte offset in source where token starts
}

// String returns a debug representation of the token.
func (t Token) String() string {
	if t.Literal == "" {
		return fmt.Sprintf("%s at %d:%d", t.Type, t.Line, t.Column)
	}
	// Truncate long literals for readability
	lit := t.Literal
	if len(lit) > 20 {
		lit = lit[:17] + "..."
	}
	return fmt.Sprintf("%s(%q) at %d:%d", t.Type, lit, t.Line, t.Column)
}

// Position represents a source code location for error reporting.
type Position struct {
	File   string
	Line   int
	Column int
}

// String returns a formatted position string.
func (p Position) String() string {
	if p.File == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}

// keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
	"package": TokenPackage,
	"import":  TokenImport,
	"func":    TokenFunc,
	"return":  TokenReturn,
	"if":      TokenIf,
	"else":    TokenElse,
	"for":     TokenFor,
	"range":   TokenRange,
}

// LookupIdent returns the token type for an identifier,
// checking if it's a keyword first.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}
