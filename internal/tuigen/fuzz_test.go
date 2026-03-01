package tuigen

import (
	"testing"
)

func FuzzLexer(f *testing.F) {
	// Seed with valid .gsx snippets
	f.Add([]byte(`package test
templ Hello() {
	<div>Hello</div>
}
`))
	f.Add([]byte(`package test
templ X(a int, b string) {
	<span class="font-bold">{a}</span>
}
`))
	f.Add([]byte(`package test
templ Loop(items []string) {
	@for _, item := range items {
		<span>{item}</span>
	}
}
`))

	f.Fuzz(func(t *testing.T, data []byte) {
		lexer := NewLexer("fuzz.gsx", string(data))
		// Should not panic
		for {
			tok := lexer.Next()
			if tok.Type == TokenEOF || tok.Type == TokenError {
				break
			}
		}
	})
}

func FuzzParser(f *testing.F) {
	f.Add([]byte(`package test
templ Hello() {
	<div>Hello</div>
}
`))

	f.Fuzz(func(t *testing.T, data []byte) {
		lexer := NewLexer("fuzz.gsx", string(data))
		parser := NewParser(lexer)
		// Should not panic — may return errors, that's fine
		_, _ = parser.ParseFile()
	})
}
