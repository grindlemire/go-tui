package highlight

import (
	"reflect"
	"testing"
)

func TestLexBash(t *testing.T) {
	type tc struct {
		code string
		want [][]Token
	}
	tests := map[string]tc{
		"comment and variable": {
			code: `echo $HOME # hi`,
			want: [][]Token{{
				{KindPlain, "echo"},
				{KindPlain, " "},
				{KindLiteral, "$HOME"},
				{KindPlain, " "},
				{KindComment, "# hi"},
			}},
		},
		"keyword and string": {
			code: `if "x"`,
			want: [][]Token{{
				{KindKeyword, "if"},
				{KindPlain, " "},
				{KindString, `"x"`},
			}},
		},
		"braced variable": {
			code: `${FOO}`,
			want: [][]Token{{{KindLiteral, "${FOO}"}}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Tokenize("bash", tt.code)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
