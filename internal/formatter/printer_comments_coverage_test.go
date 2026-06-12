package formatter

import (
	"testing"
)

// TestFormatOrphanCommentGroups exercises printOrphanComments with multiple
// comment groups separated by blank lines. Groups must stay separated by a
// single blank line and the result must be idempotent.
func TestFormatOrphanCommentGroups(t *testing.T) {
	input := `package main

templ A() {
<div></div>
// orphan one

// orphan two
}
`
	want := `package main

templ A() {
	// orphan one

	// orphan two
	<div></div>
}
`

	fmtr := newTestFormatter()
	got, err := fmtr.Format("test.gsx", input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if got != want {
		t.Errorf("Format() mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}

	again, err := fmtr.Format("test.gsx", got)
	if err != nil {
		t.Fatalf("Format() second pass error = %v", err)
	}
	if again != got {
		t.Errorf("Format() not idempotent:\nfirst:\n%s\nsecond:\n%s", got, again)
	}
}

// TestPrintCommentGroupNil verifies the nil and empty guards in
// printCommentGroup produce no output.
func TestPrintCommentGroupNil(t *testing.T) {
	p := newPrinter("\t")
	p.printCommentGroup(nil)
	if got := p.buf.String(); got != "" {
		t.Errorf("printCommentGroup(nil) wrote %q, want empty", got)
	}
}

// TestEscapeStringCarriageReturn covers the \r escape branch, which the main
// TestEscapeString table does not exercise.
func TestEscapeStringCarriageReturn(t *testing.T) {
	got := escapeString("a\rb")
	if want := `a\rb`; got != want {
		t.Errorf("escapeString(%q) = %q, want %q", "a\rb", got, want)
	}
}

// TestFormatBlockCommentEdgeCases exercises formatBlockComment directly,
// including the passthrough for text that is not a block comment.
func TestFormatBlockCommentEdgeCases(t *testing.T) {
	type tc struct {
		in   string
		want string
	}

	tests := map[string]tc{
		"not a block comment passes through": {
			in:   "// line comment",
			want: "// line comment",
		},
		"missing terminator passes through": {
			in:   "/* unterminated",
			want: "/* unterminated",
		},
		"empty block comment": {
			in:   "/**/",
			want: "/* */",
		},
		"single line normalized": {
			in:   "/*text*/",
			want: "/* text */",
		},
		"multi line": {
			in:   "/* one\n   two */",
			want: "/*\none\ntwo\n*/",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := formatBlockComment(tt.in)
			if got != tt.want {
				t.Errorf("formatBlockComment(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestFormatLineCommentEdgeCases exercises formatLineComment directly,
// including the passthrough for text that is not a line comment.
func TestFormatLineCommentEdgeCases(t *testing.T) {
	type tc struct {
		in   string
		want string
	}

	tests := map[string]tc{
		"not a line comment passes through": {
			in:   "/* block */",
			want: "/* block */",
		},
		"empty content": {
			in:   "//",
			want: "//",
		},
		"already spaced": {
			in:   "// hi",
			want: "// hi",
		},
		"tab after slashes kept": {
			in:   "//\thi",
			want: "//\thi",
		},
		"space inserted": {
			in:   "//hi",
			want: "// hi",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := formatLineComment(tt.in)
			if got != tt.want {
				t.Errorf("formatLineComment(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestFormatInlineBlockComments exercises the inline scanner directly:
// block comment normalization plus string, escaped string, and raw string
// skipping.
func TestFormatInlineBlockComments(t *testing.T) {
	type tc struct {
		in   string
		want string
	}

	tests := map[string]tc{
		"plain code unchanged": {
			in:   `fmt.Sprintf("%d", n)`,
			want: `fmt.Sprintf("%d", n)`,
		},
		"block comment normalized": {
			in:   `fmt.Sprintf("> %s", /*item*/ item)`,
			want: `fmt.Sprintf("> %s", /* item */ item)`,
		},
		"comment-like text inside string skipped": {
			in:   `fn("/*not a comment*/")`,
			want: `fn("/*not a comment*/")`,
		},
		"escaped quote inside string": {
			in:   `fn("a\"b /*x*/")`,
			want: `fn("a\"b /*x*/")`,
		},
		"raw string skipped": {
			in:   "fn(`/*raw*/`)",
			want: "fn(`/*raw*/`)",
		},
		"unterminated raw string": {
			in:   "fn(`oops",
			want: "fn(`oops",
		},
		"unterminated block comment": {
			in:   "x /*dangling",
			want: "x /*dangling",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := formatInlineBlockComments(tt.in)
			if got != tt.want {
				t.Errorf("formatInlineBlockComments(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
