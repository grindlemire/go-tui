package tui

import "github.com/grindlemire/go-tui/internal/highlight"

// TokenKind is the syntax category of a highlighted code run. It mirrors the
// internal lexer taxonomy.
type TokenKind int

// The token kinds produced by the built-in tokenizer. A Palette maps each
// kind to a color.
const (
	TokenPlain    TokenKind = iota // uncolored
	TokenKeyword                   // language keywords
	TokenString                    // string/char/template literals
	TokenComment                   // comments
	TokenNumber                    // numeric literals
	TokenLiteral                   // true/false/null/nil, builtins
	TokenType                      // type and function/method names
	TokenOperator                  // operators and punctuation
	TokenKey                       // JSON object keys
)

// CodeHighlighter turns a fenced code block into per-line styled spans.
// Implementations receive the whole block so they can track multi-line
// constructs (raw strings, block comments). They must return one []TextSpan per
// input line, and the concatenated text of each line's spans must equal the
// input line: a highlighter colorizes, it never rewrites the code.
type CodeHighlighter interface {
	Highlight(lang, code string) [][]TextSpan
}

// Palette maps token kinds to foreground colors. A missing entry (or TokenPlain)
// means "no color", so the element's base CodeBlockText style shows through.
type Palette map[TokenKind]Color

// DefaultPalette returns a dark-friendly syntax color scheme (One Dark).
func DefaultPalette() Palette {
	return Palette{
		TokenKeyword:  hexOrDefault("#c678dd"),
		TokenString:   hexOrDefault("#98c379"),
		TokenComment:  hexOrDefault("#7f848e"),
		TokenNumber:   hexOrDefault("#d19a66"),
		TokenLiteral:  hexOrDefault("#d19a66"),
		TokenType:     hexOrDefault("#61afef"),
		TokenOperator: hexOrDefault("#56b6c2"),
		TokenKey:      hexOrDefault("#e06c75"),
	}
}

// hexOrDefault parses a literal hex color, returning the terminal default if the
// string is malformed. The palette uses only valid literals, so the error path is
// unreachable in practice.
func hexOrDefault(hex string) Color {
	c, err := HexColor(hex)
	if err != nil {
		return DefaultColor()
	}
	return c
}

// NewHighlighter returns the built-in zero-dependency highlighter using palette p.
func NewHighlighter(p Palette) CodeHighlighter {
	return &builtinHighlighter{palette: p}
}

type builtinHighlighter struct {
	palette Palette
}

func (h *builtinHighlighter) Highlight(lang, code string) [][]TextSpan {
	lines := highlight.Tokenize(lang, code)
	out := make([][]TextSpan, len(lines))
	for i, line := range lines {
		spans := make([]TextSpan, 0, len(line))
		for _, t := range line {
			s := TextSpan{Text: t.Text}
			if c, ok := h.palette[mapKind(t.Kind)]; ok {
				s.Style = NewStyle().Foreground(c)
			}
			spans = append(spans, s)
		}
		out[i] = spans
	}
	return out
}

func mapKind(k highlight.Kind) TokenKind {
	switch k {
	case highlight.KindKeyword:
		return TokenKeyword
	case highlight.KindString:
		return TokenString
	case highlight.KindComment:
		return TokenComment
	case highlight.KindNumber:
		return TokenNumber
	case highlight.KindLiteral:
		return TokenLiteral
	case highlight.KindType:
		return TokenType
	case highlight.KindOperator:
		return TokenOperator
	case highlight.KindKey:
		return TokenKey
	}
	return TokenPlain
}
