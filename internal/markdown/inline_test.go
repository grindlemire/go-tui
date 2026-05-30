package markdown

import "testing"

func TestParseInline(t *testing.T) {
	type tc struct {
		in   string
		want []Inline
	}
	tests := map[string]tc{
		"plain": {
			in:   "hello world",
			want: []Inline{{Text: "hello world"}},
		},
		"bold": {
			in:   "**hi**",
			want: []Inline{{Text: "hi", Bold: true}},
		},
		"bold underscore": {
			in:   "__hi__",
			want: []Inline{{Text: "hi", Bold: true}},
		},
		"italic": {
			in:   "*hi*",
			want: []Inline{{Text: "hi", Italic: true}},
		},
		"code": {
			in:   "`x := 1`",
			want: []Inline{{Text: "x := 1", Code: true}},
		},
		"link": {
			in:   "[Go](https://go.dev)",
			want: []Inline{{Text: "Go", Link: "https://go.dev"}},
		},
		"mixed": {
			in: "a **b** c",
			want: []Inline{
				{Text: "a "},
				{Text: "b", Bold: true},
				{Text: " c"},
			},
		},
		"unmatched bracket is literal": {
			in:   "[oops",
			want: []Inline{{Text: "[oops"}},
		},
		"unmatched bold is literal": {
			in:   "see **docs",
			want: []Inline{{Text: "see **docs"}},
		},
		"unmatched italic is literal": {
			in:   "foo *bar baz",
			want: []Inline{{Text: "foo *bar baz"}},
		},
		"closed bold then stray italic stays literal": {
			in: "**a** *b",
			want: []Inline{
				{Text: "a", Bold: true},
				{Text: " *b"},
			},
		},
		"single star inside bold stays literal": {
			in: "**bold *text**",
			want: []Inline{
				{Text: "bold *text", Bold: true},
			},
		},
		"closing double does not leak italic past it": {
			in: "**bold *text** more",
			want: []Inline{
				{Text: "bold *text", Bold: true},
				{Text: " more"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := parseInline(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("parseInline(%q) = %+v (len %d), want %+v (len %d)", tt.in, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("segment %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
