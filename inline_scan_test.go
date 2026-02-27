package tui

import (
	"testing"
)

func TestScanStyledBytes(t *testing.T) {
	type token struct {
		kind  styledTokenKind
		bytes []byte
		runeW int // only meaningful for tokenRune
	}

	type tc struct {
		input  []byte
		tokens []token
	}

	tests := map[string]tc{
		"plain ascii": {
			input: []byte("abc"),
			tokens: []token{
				{kind: tokenRune, bytes: []byte("a"), runeW: 1},
				{kind: tokenRune, bytes: []byte("b"), runeW: 1},
				{kind: tokenRune, bytes: []byte("c"), runeW: 1},
			},
		},
		"newline": {
			input: []byte("a\nb"),
			tokens: []token{
				{kind: tokenRune, bytes: []byte("a"), runeW: 1},
				{kind: tokenNewline, bytes: []byte("\n")},
				{kind: tokenRune, bytes: []byte("b"), runeW: 1},
			},
		},
		"CSI sequence": {
			input: []byte("\x1b[31mhi"),
			tokens: []token{
				{kind: tokenANSI, bytes: []byte("\x1b[31m")},
				{kind: tokenRune, bytes: []byte("h"), runeW: 1},
				{kind: tokenRune, bytes: []byte("i"), runeW: 1},
			},
		},
		"wide rune": {
			input: []byte("好"),
			tokens: []token{
				{kind: tokenRune, bytes: []byte("好"), runeW: 2},
			},
		},
		"tab becomes space": {
			input: []byte("\t"),
			tokens: []token{
				{kind: tokenRune, bytes: []byte(" "), runeW: 1},
			},
		},
		"control chars dropped": {
			input: []byte("a\x01b"),
			tokens: []token{
				{kind: tokenRune, bytes: []byte("a"), runeW: 1},
				{kind: tokenRune, bytes: []byte("b"), runeW: 1},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var scanner styledByteScanner
			scanner.reset(tt.input)

			for i, want := range tt.tokens {
				if !scanner.next() {
					t.Fatalf("token %d: scanner exhausted early", i)
				}
				if scanner.kind != want.kind {
					t.Fatalf("token %d: kind = %d, want %d", i, scanner.kind, want.kind)
				}
				if string(scanner.bytes()) != string(want.bytes) {
					t.Fatalf("token %d: bytes = %q, want %q", i, scanner.bytes(), want.bytes)
				}
				if want.kind == tokenRune && scanner.runeWidth != want.runeW {
					t.Fatalf("token %d: runeWidth = %d, want %d", i, scanner.runeWidth, want.runeW)
				}
			}

			if scanner.next() {
				t.Fatalf("expected scanner exhausted, got kind=%d bytes=%q", scanner.kind, scanner.bytes())
			}
		})
	}
}
